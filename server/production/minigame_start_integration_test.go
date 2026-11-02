package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestStartMinigameAPISessionAtomicSequenceIdempotencyAndReplayIntegration(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	for _, row := range []struct {
		name  string
		modes []minigame.Mode
		gate  string
	}{
		{"solo", []minigame.Mode{minigame.ModeSolo}, "unchanged"},
		{"founder_source_scaling", []minigame.Mode{minigame.ModeSolo}, "founder_source_scaling"},
		{"founder_source_empty", []minigame.Mode{minigame.ModeSolo}, "founder_source_empty"},
		{"dual_mode", []minigame.Mode{minigame.ModeAsyncSnapshot, minigame.ModeSolo}, "unchanged"},
		{"async_only", []minigame.Mode{minigame.ModeAsyncSnapshot}, "unchanged"},
		{"missing_fiscal_unlock", []minigame.Mode{minigame.ModeSolo}, "fiscal_locked"},
		{"human_hobby_locked", []minigame.Mode{minigame.ModeSolo}, "soul_locked"},
		{"human_hobby_first_unlocked_value", []minigame.Mode{minigame.ModeSolo}, "soul_unlocked"},
	} {
		t.Run(row.name, func(t *testing.T) { testStartMinigameAPISession(t, row.modes, row.gate) })
	}
}

func testStartMinigameAPISession(t *testing.T, modes []minigame.Mode, gate string) {
	t.Helper()
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
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}

	bundle := pitchFeatureBundle(t)
	apiBytes, err := os.ReadFile("../../balance/testdata/minigame-api-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	// Diagnostic metadata only: this attempts the existing solo operation with
	// the real solo Pitch engine. It does not implement an async Pitch engine or
	// make the public request select a mode. Both-mode lists are byte-sorted.
	var minigames map[string]json.RawMessage
	if err := json.Unmarshal(bundle.Artifacts["minigames"], &minigames); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(minigames["minigames"], &rows); err != nil || len(rows) != 1 || string(rows[0]["minigame_id"]) != `"pitch"` {
		t.Fatalf("expected exactly the Pitch policy: %v", err)
	}
	rows[0]["modes"], err = json.Marshal(modes)
	if err != nil {
		t.Fatal(err)
	}
	if gate == "founder_source_scaling" || gate == "founder_source_empty" {
		// Controlled genesis funding, not a natural-progression claim. One real
		// hint purchase will consume the balance while the session is active.
		cost := bundle.Routes.KnowledgePolicy().HintCost
		if cost <= 0 {
			t.Fatal("Founder scaling fixture needs a positive hint cost")
		}
		rows[0]["scaling"] = json.RawMessage(fmt.Sprintf(`{"schema_version":1,"scaling_inputs":[{"destination":"minigame.pitch","destination_class":"breadth","source_kind":"founder_carry_counter","source_ref":"route_knowledge_balance","op":"floordiv","operand":%d,"clamp_min":0,"clamp_max":1}]}`, cost))
	}
	minigames["minigames"], err = json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Artifacts["minigames"], err = json.Marshal(minigames)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Minigames, err = minigame.LoadCatalog(bundle.Artifacts["minigames"])
	if err != nil {
		t.Fatal(err)
	}
	bundle.Artifacts["minigame_api"] = apiBytes
	bundle.MinigameAPI, err = minigameapi.LoadCatalog(apiBytes)
	if err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil || !bundle.valid(bundle.ConstantsHash) {
		t.Fatalf("v21 bundle err=%v", err)
	}
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes:   map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes},
		prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	const accountID = "01986666-ca00-4000-8000-000000000001"
	const founderID = "01986666-ca00-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	company.WireVersion, company.MeterBands = 16, nil
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders =
		meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "minigame-api-start.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion, founder.MinigameSessionSeq = 21, 0
	founder.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	founder.Pets = map[string]pet.CareState{}
	// Guarantee that start_minigame_session performs its lazy Fiscal sweep. The
	// sweep event belongs to the server-authored Founder intent, not the public
	// session coordinate; keeping those IDs distinct pins that persistence law.
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, now.Add(-time.Minute).UnixMilli(), 0
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{"minigame.pitch": true}
	founder.Soul, founder.SoulExhaustedSourceIDs = 50, []string{}
	var admissionErr error
	if len(modes) == 1 && modes[0] == minigame.ModeAsyncSnapshot {
		admissionErr = ErrInvalidIntent
	}
	switch gate {
	case "unchanged":
	case "founder_source_scaling":
		founder.RouteKnowledgeBalance = bundle.Routes.KnowledgePolicy().HintCost
	case "founder_source_empty":
		admissionErr = minigame.ErrTenantDivergence
	case "fiscal_locked":
		delete(founder.FiscalUnlocks, "minigame.pitch")
		admissionErr = ErrMinigameFiscalUnlockRequired
	case "soul_locked", "soul_unlocked":
		founder.Soul = bundle.Soul.Bands[0].MaxInclusive
		if gate == "soul_unlocked" {
			founder.Soul++
		} else {
			admissionErr = ErrMinigameHumanContentLocked
		}
		locked, err := bundle.Soul.HumanContentLocked(founder.Soul)
		if err != nil || locked != (gate == "soul_locked") {
			t.Fatalf("invalid live Soul boundary: %v", err)
		}
	default:
		t.Fatal("unknown live admission gate")
	}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "minigame-api-start.integration"})
	if err != nil {
		t.Fatal(err)
	}
	genesis, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil {
		t.Fatal(err)
	}
	pinTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, founderID, 1,
		bundle.ConstantsHash, save.VersionForState(company), genesis); err == nil {
		err = save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 1, frozen)
	}
	if err == nil {
		err = pinTx.Commit()
	} else {
		_ = pinTx.Rollback()
	}
	if err != nil {
		t.Fatal(err)
	}

	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := minigame.NewTenantRegistry(pitch.NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	platform, err := minigame.NewService(repository, registry, set)
	if err != nil {
		t.Fatal(err)
	}
	options := []ServiceOption{WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}),
		WithMinigameActivity(repository)}
	if gate == "founder_source_scaling" {
		projector, err := routeprojection.New(db, resolver)
		if err != nil {
			t.Fatal(err)
		}
		// Match the controlled genesis balance in the real projection. This
		// setup is not proof of earning Knowledge; all in-session changes below
		// must run through the actual hint transition and event projector.
		// These projections intentionally have no account FK. Reset and clean
		// only this fixture's rows so repeating the population cannot inherit
		// a previous hint event or balance after the save tables are truncated.
		cleanProjection := func() {
			for _, table := range []string{"route_hint_projection_events", "founder_route_state"} {
				if _, err := db.ExecContext(ctx, "DELETE FROM "+table+" WHERE founder_id=$1", founderID); err != nil {
					t.Errorf("clean fixture projection %s: %v", table, err)
				}
			}
		}
		cleanProjection()
		defer cleanProjection()
		if t.Failed() {
			t.FailNow()
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance,route_knowledge_debt) VALUES($1,$2,0)`,
			founderID, founder.RouteKnowledgeBalance); err != nil {
			t.Fatal(err)
		}
		options = append(options, WithRouteCatalogs(resolver), WithRouteProjector(projector), WithEventProjector(projector))
	}
	production, err := NewService(store, resolver, nil, nil, nil, options...)
	if err != nil {
		t.Fatal(err)
	}

	if admissionErr != nil {
		beforeFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		beforeCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = production.StartMinigameAPISession(ctx, platform, StartMinigameAPIRequest{
			SessionID: "01986666-ca01-7000-8000-000000000010", IntentID: "01986666-ca02-7000-8000-000000000010",
			FounderID: founderID, CompanyStreamID: companyRevision.StreamID, MinigameID: "pitch", IdempotencyKey: "async-only-refusal",
		}, now, nil)
		if !errors.Is(err, admissionErr) || gate != "founder_source_empty" && !errors.Is(err, ErrInvalidIntent) {
			t.Fatalf("ineligible solo start: err=%v want=%v", err, admissionErr)
		}
		afterFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil || !reflect.DeepEqual(afterFounder, beforeFounder) {
			t.Fatalf("admission refusal changed Founder: %v", err)
		}
		afterCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil || !reflect.DeepEqual(afterCompany, beforeCompany) {
			t.Fatalf("admission refusal changed Company: %v", err)
		}
		for _, table := range []string{"minigame_sessions", "minigame_create_receipts"} {
			var count int
			if err := db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE founder_id=$1", founderID).Scan(&count); err != nil || count != 0 {
				t.Fatalf("admission refusal left %s rows=%d err=%v", table, count, err)
			}
		}
		return
	}

	for index, step := range []string{"founder_genesis", "founder_revision", "founder_events", "founder_log", "retention"} {
		sessionID := "01986666-ca01-7000-8000-00000000000" + string(rune('1'+index))
		key := "fault-" + step
		_, startErr := production.StartMinigameAPISession(ctx, platform, StartMinigameAPIRequest{
			SessionID: sessionID, IntentID: "01986666-ca02-7000-8000-00000000000" + string(rune('1'+index)), FounderID: founderID, CompanyStreamID: companyRevision.StreamID,
			MinigameID: "pitch", IdempotencyKey: key,
		}, now, func(actual string) error {
			if actual == step {
				return errors.New("injected " + step)
			}
			return nil
		})
		if startErr == nil {
			t.Fatalf("fault %s committed", step)
		}
		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_sessions WHERE founder_id=$1`, founderID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("fault %s left sessions=%d err=%v", step, count, err)
		}
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_create_receipts WHERE founder_id=$1`, founderID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("fault %s left receipts=%d err=%v", step, count, err)
		}
		if latest, loadErr := store.LoadLatest(ctx, founderRevision.StreamID); loadErr != nil || latest.Revision.Number != 1 || latest.State.MinigameSessionSeq != 0 {
			t.Fatalf("fault %s founder=%+v err=%v", step, latest, loadErr)
		}
	}

	request := StartMinigameAPIRequest{SessionID: "01986666-ca01-7000-8000-000000000010", IntentID: "01986666-ca02-7000-8000-000000000010", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, MinigameID: "pitch", IdempotencyKey: "create-pitch-1"}
	created, err := production.StartMinigameAPISession(ctx, platform, request, now, nil)
	if err != nil || created.Replay {
		t.Fatalf("create receipt=%s replay=%v err=%v", created.Receipt, created.Replay, err)
	}
	if len(created.Events) != 1 || created.Events[0].Kind != save.EventFiscalPeriodHarvested ||
		created.Events[0].IntentID != request.IntentID || created.Events[0].IntentID == request.SessionID {
		t.Fatalf("Fiscal sweep event identity=%+v intent=%s session=%s", created.Events, request.IntentID, request.SessionID)
	}
	var response struct {
		SessionID string          `json:"session_id"`
		Revision  int64           `json:"revision"`
		Status    minigame.Status `json:"status"`
		Snapshot  json.RawMessage `json:"snapshot"`
	}
	var responseKeys map[string]json.RawMessage
	if json.Unmarshal(created.Receipt, &response) != nil || json.Unmarshal(created.Receipt, &responseKeys) != nil ||
		len(responseKeys) != 9 || response.SessionID != request.SessionID || response.Revision != 1 ||
		response.Status != minigame.StatusActive || len(response.Snapshot) == 0 {
		t.Fatalf("create response=%s", created.Receipt)
	}
	loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loadedFounder.Revision.Number != 2 || loadedFounder.State.MinigameSessionSeq != 1 {
		t.Fatalf("Founder after create=%+v err=%v", loadedFounder, err)
	}
	loadedCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loadedCompany.Revision.Number != 1 {
		t.Fatalf("Company changed on start=%+v err=%v", loadedCompany, err)
	}
	session, err := repository.Load(ctx, founderID, request.SessionID)
	if err != nil || session.Seed != minigameSessionSeed(founderID, 1, 1) {
		t.Fatalf("session=%+v err=%v", session, err)
	}

	retryRequest := request
	retryRequest.SessionID = "01986666-ca01-7000-8000-000000000011"
	retried, err := production.StartMinigameAPISession(ctx, platform, retryRequest, now.Add(time.Second), nil)
	if err != nil || !retried.Replay || !bytes.Equal(retried.Receipt, created.Receipt) {
		t.Fatalf("retry receipt=%s replay=%v err=%v", retried.Receipt, retried.Replay, err)
	}
	loadedFounder, _ = store.LoadLatest(ctx, founderRevision.StreamID)
	if loadedFounder.Revision.Number != 2 || loadedFounder.State.MinigameSessionSeq != 1 {
		t.Fatalf("retry advanced Founder=%+v", loadedFounder)
	}
	var sessionCount, receiptCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_sessions WHERE founder_id=$1`, founderID).Scan(&sessionCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_create_receipts WHERE founder_id=$1`, founderID).Scan(&receiptCount); err != nil {
		t.Fatal(err)
	}
	if sessionCount != 1 || receiptCount != 1 {
		t.Fatalf("idempotent cardinality sessions=%d receipts=%d", sessionCount, receiptCount)
	}
	_, err = production.StartMinigameAPISession(ctx, platform, StartMinigameAPIRequest{
		SessionID: "01986666-ca01-7000-8000-000000000012", IntentID: "01986666-ca02-7000-8000-000000000012", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, MinigameID: "pitch", IdempotencyKey: "create-pitch-2",
	}, now.Add(2*time.Second), nil)
	if !errors.Is(err, minigame.ErrExclusiveActivity) {
		t.Fatalf("second active session err=%v", err)
	}
	loadedFounder, _ = store.LoadLatest(ctx, founderRevision.StreamID)
	if loadedFounder.Revision.Number != 2 || loadedFounder.State.MinigameSessionSeq != 1 {
		t.Fatalf("rejected second session advanced Founder=%+v", loadedFounder)
	}
	exit := []byte(`{"intent_id":"01986666-ca01-7000-8000-000000000013","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`)
	exitResult, err := production.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(3*time.Second), exit)
	if err != nil || !bytes.Contains(exitResult.Receipt, []byte(`"detail":"minigame_session_active"`)) {
		t.Fatalf("active-session Exit receipt=%s err=%v", exitResult.Receipt, err)
	}
	loadedCompany, _ = store.LoadLatest(ctx, companyRevision.StreamID)
	loadedFounder, _ = store.LoadLatest(ctx, founderRevision.StreamID)
	if loadedCompany.Revision.Number != 1 || loadedFounder.Revision.Number != 2 || loadedFounder.State.MinigameSessionSeq != 1 {
		t.Fatalf("active-session Exit mutated Company/Founder: company=%d founder=%+v", loadedCompany.Revision.Number, loadedFounder)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || VerifyFounderHistory(history, set) != ReplayVerified {
		t.Fatalf("Founder replay verdict=%s err=%v", VerifyFounderHistory(history, set), err)
	}
	if gate == "founder_source_scaling" {
		testMinigameFrozenFounderSource(t, ctx, db, production, platform, store, set, session, founderRevision.StreamID, now)
	}
}

// Extend the real start transaction above, rather than passing caller-invented
// scaling straight to platform.Start. MP1/C28 must freeze the locked Founder
// source, not just the Company source covered by the public HTTP population.
func testMinigameFrozenFounderSource(t *testing.T, ctx context.Context, db *sql.DB, production *Service, platform *minigame.Service,
	store *save.Store, catalogs ReplayCatalogSet, session minigame.Session, founderStreamID string, now time.Time,
) {
	t.Helper()
	if string(session.ScalingInputs) != `{"minigame.pitch":1}` {
		t.Fatalf("start did not resolve the real Founder source: %s", session.ScalingInputs)
	}
	hint := []byte(`{"intent_id":"01986666-ca01-7000-8000-000000000014","kind":"buy_route_hint","expected_revision":2,"route_id":"route.nonprofit_wrapper_zip"}`)
	purchased, err := production.Handle(ctx, founderStreamID, ModeOnline, now.Add(4*time.Second), hint)
	if err != nil || purchased.Replay || !bytes.Contains(purchased.Receipt, []byte(`"outcome":"applied"`)) {
		t.Fatalf("mid-session Founder purchase receipt=%s err=%v", purchased.Receipt, err)
	}
	founder, err := store.LoadLatest(ctx, founderStreamID)
	if err != nil || founder.Revision.Number != 3 || founder.State.RouteKnowledgeBalance != 0 ||
		!founder.State.HintsUnlocked["route.nonprofit_wrapper_zip"] {
		t.Fatalf("hint did not persist the changed Founder source: founder=%+v err=%v", founder, err)
	}
	var projectedBalance int64
	if err := db.QueryRowContext(ctx, `SELECT route_knowledge_balance FROM founder_route_state WHERE founder_id=$1`, session.FounderID).Scan(&projectedBalance); err != nil || projectedBalance != 0 {
		t.Fatalf("hint did not update the real Knowledge projection: balance=%d err=%v", projectedBalance, err)
	}
	bundle := catalogs[session.ConstantsHash]
	definition, ok := bundle.Minigames.Definition("pitch")
	if !ok {
		t.Fatal("missing pinned Pitch definition")
	}
	live, err := definition.Scaling.Resolve(minigame.ScalingContext{
		PurchasedGeneratorCounts: map[string]int64{}, AttendedQualityGrades: map[string]int64{},
		FounderCarryCounters: map[string]int64{"route_knowledge_balance": founder.State.RouteKnowledgeBalance},
	})
	if err != nil || !reflect.DeepEqual(live, map[string]int64{"minigame.pitch": 0}) {
		t.Fatalf("Founder source did not diverge: live=%v err=%v", live, err)
	}
	content, ok := bundle.TenantContent(pitch.EngineRef, pitch.EngineVersion)
	seed, err := strconv.ParseUint(session.Seed, 10, 64)
	if !ok || err != nil {
		t.Fatalf("missing pinned tenant context: content=%v seed=%v", ok, err)
	}
	probe := minigame.CreateInput{Mode: session.Mode, Seed: seed, ScalingInputs: live,
		Content: content.Bytes, ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion}
	if _, err := pitch.NewTenant().Create(probe); !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Fatalf("live zero input was not rejected: %v", err)
	}
	probe.ScalingInputs = map[string]int64{"minigame.pitch": 1}
	if _, err := pitch.NewTenant().Create(probe); err != nil {
		t.Fatalf("same tenant context rejected frozen input: %v", err)
	}
	beforeCompany, err := store.LoadLatest(ctx, session.CompanyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	beforeCash, ok := beforeCompany.State.Ledger.Balance("company.cash")
	if !ok {
		t.Fatal("missing Company cash")
	}
	// Continue using the real tenant and persisted command chain, reconstruct
	// the platform after the first nonterminal command, then certify/payout.
	var resolution *minigame.CertifiedResolution
	for step := 0; step < 64 && resolution == nil; step++ {
		var snapshot pitch.Snapshot
		if json.Unmarshal(session.State, &snapshot) != nil {
			t.Fatal("invalid persisted Pitch snapshot")
		}
		var command any
		switch snapshot.Phase {
		case "playing":
			if len(snapshot.Hand) == 0 {
				t.Fatal("active Pitch hand is empty")
			}
			command = map[string]any{"kind": "play_hand", "card_ids": snapshot.Hand[:min(4, len(snapshot.Hand))]}
		case "shop":
			command = map[string]any{"kind": "end_shop"}
		default:
			t.Fatalf("unexpected active Pitch phase %q", snapshot.Phase)
		}
		payload, err := json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
		if step == 0 {
			// An unchanged JSON field alone is insufficient: this exact next
			// command must reject the changed source and accept the frozen one.
			apply := minigame.ApplyInput{Mode: session.Mode, Seed: seed, Revision: session.Revision,
				Snapshot: session.State, Command: payload, ScalingInputs: live,
				Content: content.Bytes, ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion}
			if _, err := pitch.NewTenant().Apply(apply); !errors.Is(err, minigame.ErrInvalidTenant) {
				t.Fatalf("next command tolerated changed live input: %v", err)
			}
			apply.ScalingInputs = probe.ScalingInputs
			if _, err := pitch.NewTenant().Apply(apply); err != nil {
				t.Fatalf("same next command rejected frozen input: %v", err)
			}
		}
		played, err := platform.Play(ctx, minigame.PlayRequest{FounderID: session.FounderID, SessionID: session.SessionID,
			ExpectedRevision: session.Revision, Command: payload})
		if err != nil {
			t.Fatalf("continue after Founder source change: %v", err)
		}
		if !bytes.Equal(played.Session.ScalingInputs, session.ScalingInputs) ||
			!bytes.Equal(played.Session.Genesis, session.Genesis) || played.Session.Seed != session.Seed {
			t.Fatal("Founder purchase changed frozen session coordinates")
		}
		session, resolution = played.Session, played.Resolution
		if step == 0 {
			if resolution != nil {
				t.Fatal("fixture did not exercise nonterminal recovery")
			}
			repository, err := minigame.NewRepository(db)
			if err != nil {
				t.Fatal(err)
			}
			registry, err := minigame.NewTenantRegistry(pitch.NewTenant())
			if err != nil {
				t.Fatal(err)
			}
			platform, err = minigame.NewService(repository, registry, catalogs)
			if err != nil {
				t.Fatal(err)
			}
			resumed, err := platform.Load(ctx, session.FounderID, session.SessionID)
			if err != nil || !reflect.DeepEqual(resumed, session) {
				t.Fatalf("reconstructed platform lost persisted session: err=%v", err)
			}
			session = resumed
		}
	}
	if resolution == nil {
		t.Fatal("Pitch never reached certification within the diagnostic guard")
	}
	var terminal pitch.Snapshot
	if json.Unmarshal(session.State, &terminal) != nil || terminal.Round < 2 {
		t.Fatal("fixture must reach a terminal round with a nonzero payout")
	}
	result, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(5*time.Second), nil)
	if err != nil || result.Replay {
		t.Fatalf("resolve after Founder source change: receipt=%s err=%v", result.Receipt, err)
	}
	retry, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(6*time.Second), nil)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("resolution retry changed receipt: receipt=%s err=%v", retry.Receipt, err)
	}
	resolved, err := platform.Load(ctx, session.FounderID, session.SessionID)
	if err != nil || resolved.Status != minigame.StatusResolved ||
		!bytes.Equal(resolved.ScalingInputs, session.ScalingInputs) ||
		!bytes.Equal(resolved.Genesis, session.Genesis) || resolved.Seed != session.Seed ||
		!bytes.Equal(resolved.ResolutionReceipt, result.Receipt) {
		t.Fatalf("terminal persistence lost frozen inputs or receipt: err=%v", err)
	}
	afterCompany, err := store.LoadLatest(ctx, session.CompanyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	afterCash, ok := afterCompany.State.Ledger.Balance("company.cash")
	// Independent arithmetic for this existing fixture's first send: round/2,
	// below its 300-unit cap. No Company action or passive settlement intervenes.
	wantCredit := decimal.FromString(fmt.Sprint(terminal.Round / 2))
	if !ok || afterCompany.Revision.Number != beforeCompany.Revision.Number+1 ||
		afterCash.Sub(beforeCash).String() != wantCredit.String() {
		t.Fatalf("Founder change lost/duplicated payout: revision=%d credit=%s want=%s",
			afterCompany.Revision.Number, afterCash.Sub(beforeCash), wantCredit)
	}
	var quota int
	if err := db.QueryRowContext(ctx, `SELECT quota_used FROM minigame_faucet_window WHERE founder_id=$1 AND minigame_id='pitch'`, session.FounderID).Scan(&quota); err != nil || quota != 1 {
		t.Fatalf("payout/retry quota=%d err=%v", quota, err)
	}
	founder, err = store.LoadLatest(ctx, founderStreamID)
	if err != nil || founder.Revision.Number != 4 || founder.State.RouteKnowledgeBalance != 0 ||
		!founder.State.HintsUnlocked["route.nonprofit_wrapper_zip"] {
		t.Fatalf("payout/retry overwrote or duplicated Founder mutation: founder=%+v err=%v", founder, err)
	}
	history, err := store.LoadFounderHistory(ctx, founderStreamID)
	if err != nil || VerifyFounderHistory(history, catalogs) != ReplayVerified {
		t.Fatalf("start/hint/resolution history did not verify: %v", err)
	}
	t.Logf("Founder Knowledge spent during play; frozen input=1 live=0; terminal round=%d credit=%s; payout quota=%d",
		terminal.Round, wantCredit, quota)
}
