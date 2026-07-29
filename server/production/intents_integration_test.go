package production

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type integrationCatalogs map[string]*economy.Catalog

func (catalogs integrationCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func TestIntentServiceIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE events, intent_records, save_revisions, save_streams`); err != nil {
		t.Fatal(err)
	}
	catalogBytes, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	hash := save.ConstantsHash(catalogBytes)
	resolver := integrationCatalogs{hash: catalog}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e2"})
	if err != nil {
		t.Fatal(err)
	}
	cursor := time.Date(2026, 7, 28, 8, 0, 0, 0, time.UTC)
	state := &save.State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0},
		EvaluatedThrough: cursor, ManualTokenMilli: 50_000, ManualTokenRefilledAt: cursor,
	}
	revision, err := store.CreateStream(ctx, save.StreamKey{
		OwnerKind: save.OwnerFounder, OwnerID: "66666666-6666-4666-8666-666666666666", Scope: economy.ScopeCompany,
	}, hash, state, save.WriteContext{Cause: "production.integration"})
	if err != nil {
		t.Fatal(err)
	}
	metrics := fakeInvariantMetrics{}
	service, err := NewService(store, resolver, nil, metrics, nil)
	if err != nil {
		t.Fatal(err)
	}

	buy := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-111111111111","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":2}}`)
	first, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor, buy)
	if err != nil || first.Replay {
		t.Fatalf("buy=%s replay=%v err=%v", first.Receipt, first.Replay, err)
	}
	var buyReceipt struct {
		Outcome      string `json:"outcome"`
		AppliedCount int64  `json:"applied_count"`
		NewRevision  int64  `json:"new_revision"`
		Snapshot     struct {
			Balances   map[string]string `json:"balances"`
			Generators map[string]int64  `json:"generators"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(first.Receipt, &buyReceipt); err != nil {
		t.Fatal(err)
	}
	if buyReceipt.Outcome != "applied" || buyReceipt.AppliedCount != 2 || buyReceipt.NewRevision != 2 ||
		buyReceipt.Snapshot.Balances["company.cash"] != "7.87e1" || buyReceipt.Snapshot.Generators["generator.beige_tower"] != 2 {
		t.Fatalf("buy receipt = %+v raw=%s", buyReceipt, first.Receipt)
	}
	replay, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor, buy)
	if err != nil || !replay.Replay || string(replay.Receipt) != string(first.Receipt) {
		t.Fatalf("buy replay=%s replay=%v err=%v", replay.Receipt, replay.Replay, err)
	}

	manual := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-222222222222","kind":"perform_manual_batch","expected_revision":2,"action_id":"manual.click","count":60,"window_ms":2400}`)
	manualResult, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), manual)
	if err != nil {
		t.Fatal(err)
	}
	var manualReceipt struct {
		AppliedCount int64 `json:"applied_count"`
		NewRevision  int64 `json:"new_revision"`
		Snapshot     struct {
			Balances         map[string]string `json:"balances"`
			ManualTokenMilli int64             `json:"manual_token_milli"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(manualResult.Receipt, &manualReceipt); err != nil {
		t.Fatal(err)
	}
	if manualReceipt.AppliedCount != 50 || manualReceipt.NewRevision != 3 ||
		manualReceipt.Snapshot.Balances["company.cash"] != "1.307e2" || manualReceipt.Snapshot.ManualTokenMilli != 0 {
		t.Fatalf("manual receipt=%+v raw=%s", manualReceipt, manualResult.Receipt)
	}

	manualEmpty := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-333333333333","kind":"perform_manual_batch","expected_revision":3,"action_id":"manual.click","count":10,"window_ms":10}`)
	emptyResult, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), manualEmpty)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(emptyResult.Receipt, &manualReceipt); err != nil || manualReceipt.AppliedCount != 0 || manualReceipt.NewRevision != 4 {
		t.Fatalf("empty manual=%+v raw=%s err=%v", manualReceipt, emptyResult.Receipt, err)
	}

	unaffordable := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-444444444444","kind":"buy_generator","expected_revision":4,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1000}}`)
	rejection, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), unaffordable)
	if err != nil {
		t.Fatal(err)
	}
	var rejected struct {
		Outcome         string `json:"outcome"`
		CurrentRevision int64  `json:"current_revision"`
		Rejection       struct {
			Category string `json:"category"`
		} `json:"rejection"`
	}
	if err := json.Unmarshal(rejection.Receipt, &rejected); err != nil || rejected.Outcome != "rejected" ||
		rejected.CurrentRevision != 4 || rejected.Rejection.Category != "unaffordable" {
		t.Fatalf("rejection=%+v raw=%s err=%v", rejected, rejection.Receipt, err)
	}
	correctedSameID := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-444444444444","kind":"buy_generator","expected_revision":4,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`)
	stickyConflict, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), correctedSameID)
	if err != nil || json.Unmarshal(stickyConflict.Receipt, &rejected) != nil ||
		rejected.Rejection.Category != "idempotency_conflict" || rejected.CurrentRevision != 4 {
		t.Fatalf("sticky idempotency conflict=%+v raw=%s err=%v", rejected, stickyConflict.Receipt, err)
	}

	conflictingRevision := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-555555555555","kind":"buy_generator","expected_revision":2,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`)
	conflict, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor, conflictingRevision)
	if err != nil || json.Unmarshal(conflict.Receipt, &rejected) != nil || rejected.Rejection.Category != "revision_conflict" || rejected.CurrentRevision != 4 {
		t.Fatalf("revision conflict=%+v raw=%s err=%v", rejected, conflict.Receipt, err)
	}

	invariantIntentID := "018f6b7c-9abc-7def-8abc-666666666666"
	invariantRequestHash := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	invariantReport := InvariantReport{
		Kind: InvariantAffordFallback, IntentID: invariantIntentID, Detail: "generator.beige_tower",
	}
	invariantResult, err := store.ApplyIntent(ctx, revision.StreamID, 4, invariantIntentID, invariantRequestHash,
		func(state *save.State, current save.Revision) (save.IntentDecision, error) {
			return appliedDecision(IntentRequest{IntentID: invariantIntentID}, state, current.Number+1, 0,
				state.Ledger.Snapshot(), nil, []InvariantReport{invariantReport})
		})
	if err != nil || invariantResult.Replay {
		t.Fatalf("invariant apply=%+v err=%v", invariantResult, err)
	}
	service.recordCommittedInvariants(invariantResult, []InvariantReport{invariantReport})
	invariantReplay, err := store.ApplyIntent(ctx, revision.StreamID, 4, invariantIntentID, invariantRequestHash,
		func(*save.State, save.Revision) (save.IntentDecision, error) {
			return save.IntentDecision{}, errors.New("invariant replay callback must not run")
		})
	if err != nil || !invariantReplay.Replay {
		t.Fatalf("invariant replay=%+v err=%v", invariantReplay, err)
	}
	service.recordCommittedInvariants(invariantReplay, []InvariantReport{invariantReport})

	rejectedReport := InvariantReport{
		Kind: InvariantResidualClamp, IntentID: "018f6b7c-9abc-7def-8abc-777777777777", Detail: "generator.beige_tower",
	}
	rejectedResult, err := store.ApplyIntent(ctx, revision.StreamID, 5, rejectedReport.IntentID,
		"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		func(*save.State, save.Revision) (save.IntentDecision, error) {
			return rejectedDecision(IntentRequest{IntentID: rejectedReport.IntentID}, 5, "unaffordable", "generator.beige_tower")
		})
	if err != nil || rejectedResult.Outcome != save.IntentRejected {
		t.Fatalf("reported rejection=%+v err=%v", rejectedResult, err)
	}
	service.recordCommittedInvariants(rejectedResult, []InvariantReport{rejectedReport})

	abortReport := InvariantReport{
		Kind: InvariantResidualAbort, IntentID: "018f6b7c-9abc-7def-8abc-888888888888", Detail: "generator.beige_tower",
	}
	_, err = store.ApplyIntent(ctx, revision.StreamID, 5, abortReport.IntentID,
		"sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		func(*save.State, save.Revision) (save.IntentDecision, error) {
			return save.IntentDecision{}, ErrInvalidEngineState
		})
	if !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("invariant abort error=%v", err)
	}
	service.recordAbortedInvariants([]InvariantReport{abortReport})

	var revisions, events, intents, invariantEvents int
	if err := db.QueryRowContext(ctx, `
		SELECT
		 (SELECT count(*) FROM save_revisions WHERE stream_id=$1),
		 (SELECT count(*) FROM events WHERE stream_id=$1),
		 (SELECT count(*) FROM intent_records WHERE stream_id=$1),
		 (SELECT count(*) FROM events WHERE stream_id=$1 AND kind='invariant_reported')`, revision.StreamID,
	).Scan(&revisions, &events, &intents, &invariantEvents); err != nil {
		t.Fatal(err)
	}
	if revisions != 5 || events != 2 || intents != 6 || invariantEvents != 1 {
		t.Fatalf("rows revisions=%d events=%d intents=%d invariant_events=%d", revisions, events, intents, invariantEvents)
	}
	if metrics[string(InvariantAffordFallback)] != 1 || metrics[string(InvariantResidualClamp)] != 1 ||
		metrics[string(InvariantResidualAbort)] != 1 {
		t.Fatalf("invariant metrics=%+v", metrics)
	}
}
