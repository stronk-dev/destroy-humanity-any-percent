package production

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/commonsprojection"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type integrationAssignment struct{ serverID string }

func (value integrationAssignment) ResolveAssignment(string) (commonsprojection.AssignmentContext, bool) {
	return commonsprojection.AssignmentContext{ServerID: value.serverID, ActivityBracket: "activity.standard"}, true
}

type integrationWeight int64

func (value integrationWeight) CompactWeightPPM(string) (int64, bool) { return int64(value), true }

type integrationCatalogs struct {
	economy  map[string]*economy.Catalog
	routes   map[string]*routes.Catalog
	prestige map[string]*prestigecore.Policy
	factions map[string]*faction.Catalog
}

func (catalogs integrationCatalogs) ResolvePrestige(hash string) (*prestigecore.Policy, bool) {
	policy, ok := catalogs.prestige[hash]
	return policy, ok
}

func (catalogs integrationCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs.economy[hash]
	return catalog, ok
}

func (catalogs integrationCatalogs) ResolveRoutes(hash string) (*routes.Catalog, bool) {
	catalog, ok := catalogs.routes[hash]
	return catalog, ok
}

func (catalogs integrationCatalogs) ResolveFaction(hash string) (*faction.Catalog, bool) {
	catalog, ok := catalogs.factions[hash]
	return catalog, ok
}

func (catalogs integrationCatalogs) ValidateState(hash string, state *save.State) error {
	if state != nil && state.FactionID == "" {
		return nil
	}
	catalog, ok := catalogs.ResolveFaction(hash)
	if !ok {
		return faction.ErrInvalidStockState
	}
	return catalog.ValidateState(state)
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
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes, route_hint_projection_events, founder_route_state, founder_route_executions, route_projection_events, events, intent_records, save_revisions, save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	replayBundle := loadReplayTestBundle(t, bundle.Hash, bundle.Artifacts)
	catalogBytes := bundle.Artifacts["economy"]
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	routeBytes := bundle.Artifacts["routes"]
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	commonsBytes := bundle.Artifacts["commons"]
	commonsCatalog, err := commons.LoadCatalog(commonsBytes)
	if err != nil {
		t.Fatal(err)
	}
	factionCatalog, err := faction.LoadCatalog(bundle.Artifacts["factions"], faction.CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM,
		DefaultPPM: commonsCatalog.DefaultTithePPM,
		MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	prestigePolicy, err := prestigecore.LoadPolicy(bundle.Artifacts["prestige"])
	if err != nil {
		t.Fatal(err)
	}
	hash := bundle.Hash
	seedProductionEpoch(t, db, hash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{hash: catalog}, routes: map[string]*routes.Catalog{hash: routeCatalog}, prestige: map[string]*prestigecore.Policy{hash: prestigePolicy}, factions: map[string]*faction.Catalog{hash: factionCatalog}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	projector, err := routeprojection.New(db, resolver)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalogs := commons.CatalogSet{hash: commonsCatalog}
	commonsProjector, err := commonsprojection.New(db, integrationAssignment{serverID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}, commonsCatalogs)
	if err != nil {
		t.Fatal(err)
	}
	commonsProvider := commonsbinding.Provider{Catalogs: commonsCatalogs, Snapshots: commonsProjector}
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e2"})
	if err != nil {
		t.Fatal(err)
	}
	cursor := time.Date(2026, 7, 28, 8, 0, 0, 0, time.UTC)
	state := &save.State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0},
		EvaluatedThrough: cursor, ManualTokenMilli: 50_000, ManualTokenRefilledAt: cursor,
		StructureID: "structure.nonprofit",
	}
	revision, err := store.CreateStream(ctx, save.StreamKey{
		OwnerKind: save.OwnerFounder, OwnerID: "66666666-6666-4666-8666-666666666666", Scope: economy.ScopeCompany,
	}, hash, state, save.WriteContext{Cause: "production.integration"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PinRunToCurrentEpoch(ctx, revision.StreamID, "66666666-6666-4666-8666-666666666666", 1, hash); err != nil {
		t.Fatal(err)
	}
	founderLedger, err := economy.RestoreLedger(catalog, economy.ScopeFounder, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{
		OwnerKind: save.OwnerFounder, OwnerID: "66666666-6666-4666-8666-666666666666", Scope: economy.ScopeFounder,
	}, hash, &save.State{
		Ledger: founderLedger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: cursor, ManualTokenRefilledAt: cursor,
	}, save.WriteContext{Cause: "routes.integration"})
	if err != nil {
		t.Fatal(err)
	}
	metrics := fakeInvariantMetrics{}
	service, err := NewService(store, resolver, commonsProvider, metrics, nil, WithRouteCatalogs(resolver), WithRouteProjector(projector), WithCompactPolicies(commonsCatalogs), WithProgressionRuntime(resolver), WithCurrentConstantsHash(hash), WithCommonsWeightResolver(integrationWeight(1_000_000)), WithReplayCatalogs(ReplayCatalogSet{hash: replayBundle}), WithGuildSettlements(emptyGuildSettlements{}), WithEventProjector(commonsProjector))
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

	cross := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-999999999999","kind":"cross_gate","expected_revision":5,"gate_id":"gate.t4_to_t5","route_id":"route.nonprofit_wrapper_zip"}`)
	crossed, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), cross)
	if err != nil || crossed.Replay {
		t.Fatalf("cross=%s replay=%v err=%v", crossed.Receipt, crossed.Replay, err)
	}
	var crossReceipt struct {
		NewRevision int64 `json:"new_revision"`
		Snapshot    struct {
			Gates map[string]bool `json:"gates_crossed"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(crossed.Receipt, &crossReceipt); err != nil || crossReceipt.NewRevision != 6 || !crossReceipt.Snapshot.Gates["gate.t4_to_t5"] {
		t.Fatalf("cross receipt=%+v raw=%s err=%v", crossReceipt, crossed.Receipt, err)
	}
	crossReplay, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), cross)
	if err != nil || !crossReplay.Replay || string(crossReplay.Receipt) != string(crossed.Receipt) {
		t.Fatalf("cross replay=%+v err=%v", crossReplay, err)
	}
	incorporate := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-abababababab","kind":"incorporate","expected_revision":6,"faction_id":"bootstrapper"}`)
	incorporated, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), incorporate)
	if err != nil || incorporated.Replay {
		t.Fatalf("incorporate=%s replay=%v err=%v", incorporated.Receipt, incorporated.Replay, err)
	}
	var incorporatedReceipt struct {
		NewRevision int64 `json:"new_revision"`
		Snapshot    struct {
			FactionID     *string `json:"faction_id"`
			StockResource *string `json:"stock_resource"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(incorporated.Receipt, &incorporatedReceipt); err != nil || incorporatedReceipt.NewRevision != 7 || incorporatedReceipt.Snapshot.FactionID == nil || *incorporatedReceipt.Snapshot.FactionID != "bootstrapper" || incorporatedReceipt.Snapshot.StockResource == nil || *incorporatedReceipt.Snapshot.StockResource != "revenue" {
		t.Fatalf("incorporate receipt=%+v raw=%s err=%v", incorporatedReceipt, incorporated.Receipt, err)
	}
	incorporateReplay, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Second), incorporate)
	if err != nil || !incorporateReplay.Replay || string(incorporateReplay.Receipt) != string(incorporated.Receipt) {
		t.Fatalf("incorporate replay=%+v err=%v", incorporateReplay, err)
	}
	sign := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-bbbbbbbbbbbb","kind":"sign_compact","expected_revision":7,"tithe_ppm":100000}`)
	signed, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(2*time.Second), sign)
	if err != nil || signed.Replay {
		t.Fatalf("sign=%s replay=%v err=%v", signed.Receipt, signed.Replay, err)
	}
	memberManual := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-cccccccccccc","kind":"perform_manual_batch","expected_revision":8,"action_id":"manual.click","count":1,"window_ms":3600000}`)
	sampled, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(time.Hour+2*time.Second), memberManual)
	if err != nil || sampled.Replay {
		t.Fatalf("sample=%s replay=%v err=%v", sampled.Receipt, sampled.Replay, err)
	}
	commonsSnapshot, err := commonsProjector.Snapshot(ctx, "66666666-6666-4666-8666-666666666666", hash)
	if err != nil || commonsSnapshot.HealthPPM <= 0 || commonsSnapshot.CohortCapacity == "0" {
		t.Fatalf("commons snapshot=%+v err=%v", commonsSnapshot, err)
	}
	leave := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-dddddddddddd","kind":"leave_compact","expected_revision":9}`)
	leftCompact, err := service.Handle(ctx, revision.StreamID, ModeOnline, cursor.Add(2*time.Hour+2*time.Second), leave)
	if err != nil || leftCompact.Replay {
		t.Fatalf("leave=%s replay=%v err=%v", leftCompact.Receipt, leftCompact.Replay, err)
	}
	var projectedMember bool
	if err := db.QueryRowContext(ctx, `SELECT member FROM company_compact_memberships WHERE company_stream_id=$1`, revision.StreamID).Scan(&projectedMember); err != nil || projectedMember {
		t.Fatalf("projected member=%v err=%v", projectedMember, err)
	}

	existingOwner := "77777777-7777-4777-8777-777777777777"
	existingLedger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e2"})
	if err != nil {
		t.Fatal(err)
	}
	existingRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: existingOwner, Scope: economy.ScopeCompany}, hash, &save.State{
		Ledger: existingLedger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: cursor,
		ManualTokenRefilledAt: cursor, Tier: 2, RunSeq: 1, GatesCrossed: map[string]bool{},
		DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}, RunStartedAt: cursor,
	}, save.WriteContext{Cause: "production.integration.open_source"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PinRunToCurrentEpoch(ctx, existingRevision.StreamID, existingOwner, 1, hash); err != nil {
		t.Fatal(err)
	}
	existingSign := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-eeeeeeeeeeee","kind":"sign_compact","expected_revision":1,"tithe_ppm":100000}`)
	if _, err := service.Handle(ctx, existingRevision.StreamID, ModeOnline, cursor, existingSign); err != nil {
		t.Fatal(err)
	}
	existingIncorporate := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-ffffffffffff","kind":"incorporate","expected_revision":2,"faction_id":"open_source"}`)
	if _, err := service.Handle(ctx, existingRevision.StreamID, ModeOnline, cursor.Add(time.Second), existingIncorporate); err != nil {
		t.Fatal(err)
	}
	var projectedTithe, projectedRevision int64
	if err := db.QueryRowContext(ctx, `SELECT tithe_ppm,projected_revision FROM company_compact_memberships WHERE company_stream_id=$1`, existingRevision.StreamID).Scan(&projectedTithe, &projectedRevision); err != nil || projectedTithe != 130_000 || projectedRevision != 3 {
		t.Fatalf("continued membership tithe=%d revision=%d err=%v", projectedTithe, projectedRevision, err)
	}
	hint := []byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-aaaaaaaaaaaa","kind":"buy_route_hint","expected_revision":1,"route_id":"route.nonprofit_wrapper_zip"}`)
	hintResult, err := service.Handle(ctx, founderRevision.StreamID, ModeOnline, cursor.Add(time.Second), hint)
	if err != nil || hintResult.Replay {
		t.Fatalf("hint=%s replay=%v err=%v", hintResult.Receipt, hintResult.Replay, err)
	}
	var hintReceipt struct {
		Snapshot struct {
			Balance int64    `json:"route_knowledge_balance"`
			Hints   []string `json:"hints_unlocked"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(hintResult.Receipt, &hintReceipt); err != nil || hintReceipt.Snapshot.Balance != 75 || len(hintReceipt.Snapshot.Hints) != 1 {
		t.Fatalf("hint receipt=%+v raw=%s err=%v", hintReceipt, hintResult.Receipt, err)
	}
	hintReplay, err := service.Handle(ctx, founderRevision.StreamID, ModeOnline, cursor.Add(time.Second), hint)
	if err != nil || !hintReplay.Replay || string(hintReplay.Receipt) != string(hintResult.Receipt) {
		t.Fatalf("hint replay=%+v err=%v", hintReplay, err)
	}

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
	if revisions != 5 || events != 11 || intents != 11 || invariantEvents != 1 {
		t.Fatalf("rows revisions=%d events=%d intents=%d invariant_events=%d", revisions, events, intents, invariantEvents)
	}
	var runLogCount int
	var firstSequence, firstAppliedRevision int64
	var firstCanonical, firstReplayInputs []byte
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID).Scan(&runLogCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT seq,canonical_payload,replay_inputs,applied_revision FROM run_log WHERE company_stream_id=$1 AND run_seq=1 ORDER BY seq LIMIT 1`, revision.StreamID).Scan(&firstSequence, &firstCanonical, &firstReplayInputs, &firstAppliedRevision); err != nil {
		t.Fatal(err)
	}
	parsedBuy, err := ParseIntent(buy)
	parsedReplay, replayErr := parseReplayInputs(firstReplayInputs)
	if err != nil || replayErr != nil || runLogCount != 9 || firstSequence != 1 || firstAppliedRevision != 2 || string(firstCanonical) != string(parsedBuy.CanonicalPayload) ||
		parsedReplay.Command.IntentID != parsedBuy.IntentID || parsedReplay.Command.CompanyStreamID != revision.StreamID || parsedReplay.Command.RunLogSeq != 1 || parsedReplay.EvaluationMode != ModeOnline {
		t.Fatalf("run log count=%d seq=%d revision=%d canonical=%s want=%s err=%v", runLogCount, firstSequence, firstAppliedRevision, firstCanonical, parsedBuy.CanonicalPayload, err)
	}
	var rejectedRevision *int64
	if err := db.QueryRowContext(ctx, `SELECT applied_revision FROM run_log WHERE company_stream_id=$1 AND intent_id=$2`, revision.StreamID, "018f6b7c-9abc-7def-8abc-444444444444").Scan(&rejectedRevision); err != nil || rejectedRevision != nil {
		t.Fatalf("rejected run-log revision=%v err=%v", rejectedRevision, err)
	}
	if metrics[string(InvariantAffordFallback)] != 1 || metrics[string(InvariantResidualClamp)] != 1 ||
		metrics[string(InvariantResidualAbort)] != 1 {
		t.Fatalf("invariant metrics=%+v", metrics)
	}
}

func seedProductionEpoch(t *testing.T, db *sql.DB, hash string, artifacts map[string][]byte) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	for name, data := range artifacts {
		if _, err := db.Exec(`INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, hash, name, data); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := db.QueryRow(`INSERT INTO epochs(name,started_at,changelog_ref) VALUES('Phase 0',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, hash); err != nil {
		t.Fatal(err)
	}
}
