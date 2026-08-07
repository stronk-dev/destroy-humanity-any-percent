package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

func recoverySoulBundle(t *testing.T) CatalogBundle {
	t.Helper()
	_, foundations := foundationTestBundles(t)
	_, pets := founderFeatureBundles(t, foundations)
	bundle := soulFeatureBundle(t, fiscalFeatureBundle(t, pets))
	var root map[string]any
	if err := json.Unmarshal(bundle.Artifacts["soul"], &root); err != nil {
		t.Fatal(err)
	}
	root["recovery_activities"] = []any{map[string]any{
		"activity_id": "touch_grass.fixture", "duration_attended_ms": float64(5_000),
		"recovery_amount": float64(15), "reason_key": "category.any_percent",
	}}
	artifact, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	catalog, err := soul.LoadCatalog(artifact, soul.Declarations{CopyKeys: keys, EpochSeeded: true,
		CatchupCeilingMS: bundle.Prestige.CatchupCeilingMS})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	bundle.Artifacts["soul"] = artifact
	bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Soul = catalog
	if !bundle.valid(bundle.ConstantsHash) {
		t.Fatal("Soul recovery fixture bundle invalid")
	}
	return bundle
}

func TestSoulRecoveryIntegrationAtomicSuppressionReplayAndExclusivity(t *testing.T) {
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
	bundle := recoverySoulBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{
		economy:  map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes:   map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes},
		prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction},
	}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 6, 22, 0, 0, 0, time.UTC)
	const accountID = "01986666-b100-4000-8000-000000000001"
	const founderID = "01986666-b100-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	company.WireVersion = 16
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterBands = nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	company.GeneratorCounts["generator.beige_tower"] = 10
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "soul.recovery.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion = 20
	founder.FiscalPeriodOpenedWallMS = now.UnixMilli()
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{}
	founder.Soul = 50
	founder.SoulExhaustedSourceIDs = []string{}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "soul.recovery.integration-founder"})
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
	repository, err := soul.NewRecoveryRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}),
		WithGuildSettlements(emptyGuildSettlements{}), WithSoulRecovery(repository))
	if err != nil {
		t.Fatal(err)
	}
	const sessionID = "01986666-b101-7000-8000-000000000001"
	start, err := service.StartSoulRecovery(ctx, StartSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, ActivityID: "touch_grass.fixture"}, now)
	var startReceipt soulRecoveryStartReceipt
	var startKeys map[string]json.RawMessage
	if err != nil || json.Unmarshal(start.Receipt, &startReceipt) != nil || json.Unmarshal(start.Receipt, &startKeys) != nil || len(startKeys) != 7 || startReceipt.SessionID != sessionID ||
		startReceipt.ProgressToken == "" || startReceipt.AttendedProgressMS != 0 || startReceipt.LastProgressServerMS != now.UnixMilli() {
		t.Fatalf("start receipt=%s err=%v", start.Receipt, err)
	}
	retryStart, err := service.StartSoulRecovery(ctx, StartSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, ActivityID: "touch_grass.fixture"}, now)
	var reconnectReceipt soulRecoveryStartReceipt
	if err != nil || json.Unmarshal(retryStart.Receipt, &reconnectReceipt) != nil ||
		reconnectReceipt.SessionID != sessionID || reconnectReceipt.ProgressToken == startReceipt.ProgressToken {
		t.Fatalf("start retry receipt=%s err=%v", retryStart.Receipt, err)
	}
	if _, err := service.ProgressSoulRecovery(ctx, ProgressSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID,
		ProgressToken: startReceipt.ProgressToken}, now.Add(time.Second), nil); err == nil || !strings.Contains(err.Error(), "recovery_token") {
		t.Fatalf("stale progress token error=%v", err)
	}
	ordinary := []byte(`{"intent_id":"01986666-b101-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`)
	blocked, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(time.Second), ordinary)
	if err != nil || !strings.Contains(string(blocked.Receipt), `"detail":"exclusive_activity"`) {
		t.Fatalf("ordinary command not blocked receipt=%s err=%v", blocked.Receipt, err)
	}
	minigameRepository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tenantRegistry, err := minigame.NewTenantRegistry(resolutionFixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	platform, err := minigame.NewService(minigameRepository, tenantRegistry)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := platform.Start(ctx, minigame.StartRequest{SessionID: "01986666-b101-7000-8000-000000000003",
		MinigameID: "fixture.counter", FounderID: founderID, CompanyStreamID: companyRevision.StreamID,
		RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0", ConstantsHash: bundle.ConstantsHash,
		ScalingInputs: map[string]int64{"option.count": 1}, Seed: "1", Mode: minigame.ModeSolo}); !errors.Is(err, minigame.ErrExclusiveActivity) {
		t.Fatalf("minigame start during Soul recovery error=%v", err)
	}
	if _, err := service.ResolveSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID}, now.Add(time.Second), nil); err == nil {
		t.Fatal("early Soul recovery resolved")
	}
	active, err := repository.Load(ctx, founderID, sessionID)
	if err != nil || active.Status != soul.RecoveryActive {
		t.Fatalf("early resolution mutated session=%+v err=%v", active, err)
	}
	progressTimes := []time.Duration{time.Second, time.Second, 10 * time.Second, 14 * time.Second}
	wantProgress := []int64{1000, 1000, 1000, 5000}
	for index, offset := range progressTimes {
		progress, progressErr := service.ProgressSoulRecovery(ctx, ProgressSoulRecoveryRequest{SessionID: sessionID,
			FounderID: founderID, ProgressToken: reconnectReceipt.ProgressToken}, now.Add(offset), nil)
		var receipt soulRecoveryProgressReceipt
		var progressKeys map[string]json.RawMessage
		if progressErr != nil || json.Unmarshal(progress.Receipt, &receipt) != nil || json.Unmarshal(progress.Receipt, &progressKeys) != nil || len(progressKeys) != 5 || receipt.AttendedProgressMS != wantProgress[index] ||
			receipt.Eligible != (wantProgress[index] >= 5000) {
			t.Fatalf("progress[%d] receipt=%s err=%v", index, progress.Receipt, progressErr)
		}
	}
	result, err := service.ResolveSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID}, now.Add(14*time.Second), nil)
	if err != nil || result.Replay || !strings.Contains(string(result.Receipt), `"soul_after":65`) {
		t.Fatalf("resolve receipt=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	retry, err := service.ResolveSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID}, now.Add(14*time.Second), nil)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("resolve retry receipt=%s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	loadedCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	loadedFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
	cash, _ := loadedCompany.State.Ledger.Balance("company.cash")
	if loadedCompany.Revision.Number != 2 || loadedFounder.Revision.Number != 2 || cash.String() != "0" ||
		loadedFounder.State.Soul != 65 || !loadedCompany.State.EvaluatedThrough.Equal(now.Add(14*time.Second)) {
		t.Fatalf("company=%d cash=%s evaluated=%s founder=%d soul=%d", loadedCompany.Revision.Number, cash,
			loadedCompany.State.EvaluatedThrough, loadedFounder.Revision.Number, loadedFounder.State.Soul)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}) != ReplayVerified {
		t.Fatalf("Founder Soul history failed verification: %v", err)
	}
	const watchdogSessionID = "01986666-b101-7000-8000-000000000004"
	watchdogStart, err := service.StartSoulRecovery(ctx, StartSoulRecoveryRequest{SessionID: watchdogSessionID,
		FounderID: founderID, CompanyStreamID: companyRevision.StreamID, ActivityID: "touch_grass.fixture"}, now.Add(20*time.Second))
	var watchdogStartReceipt soulRecoveryStartReceipt
	if err != nil || json.Unmarshal(watchdogStart.Receipt, &watchdogStartReceipt) != nil {
		t.Fatalf("watchdog start=%s err=%v", watchdogStart.Receipt, err)
	}
	expiredBlocked, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(25*time.Hour), ordinary)
	if err != nil || !strings.Contains(string(expiredBlocked.Receipt), `"session_expired":true`) {
		t.Fatalf("expired ordinary command receipt=%s err=%v", expiredBlocked.Receipt, err)
	}
	watchdogResult, err := service.ProgressSoulRecovery(ctx, ProgressSoulRecoveryRequest{SessionID: watchdogSessionID,
		FounderID: founderID, ProgressToken: watchdogStartReceipt.ProgressToken}, now.Add(25*time.Hour), nil)
	if err != nil || !strings.Contains(string(watchdogResult.Receipt), `"cancelled_by":"watchdog"`) {
		t.Fatalf("watchdog receipt=%s err=%v", watchdogResult.Receipt, err)
	}
	watchdogCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	watchdogFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
	if !watchdogCompany.State.EvaluatedThrough.Equal(now.Add(20*time.Second)) || watchdogFounder.State.Soul != 65 {
		t.Fatalf("watchdog terminal coordinate=%s soul=%d", watchdogCompany.State.EvaluatedThrough, watchdogFounder.State.Soul)
	}
	var companyLogs, founderLogs, startedEvents, recoveredEvents int
	queries := []struct {
		target *int
		query  string
		args   []any
	}{
		{&companyLogs, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND intent_id=$2`, []any{companyRevision.StreamID, sessionID}},
		{&founderLogs, `SELECT count(*) FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2`, []any{founderRevision.StreamID, sessionID}},
		{&startedEvents, `SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2 AND kind='soul_recovery_started.v1'`, []any{founderRevision.StreamID, sessionID}},
		{&recoveredEvents, `SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2 AND kind='soul_recovered.v1'`, []any{founderRevision.StreamID, sessionID}},
	}
	for _, query := range queries {
		if err := db.QueryRowContext(ctx, query.query, query.args...).Scan(query.target); err != nil {
			t.Fatal(err)
		}
	}
	if companyLogs != 1 || founderLogs != 1 || startedEvents != 1 || recoveredEvents != 1 {
		t.Fatalf("logs=%d/%d events=%d/%d", companyLogs, founderLogs, startedEvents, recoveredEvents)
	}
	var loggedPayload, loggedInputs, loggedReceipt []byte
	if err := db.QueryRowContext(ctx, `SELECT canonical_payload,replay_inputs::text,receipt::text
		FROM run_log WHERE company_stream_id=$1 AND intent_id=$2`, companyRevision.StreamID, sessionID).
		Scan(&loggedPayload, &loggedInputs, &loggedReceipt); err != nil {
		t.Fatal(err)
	}
	replayState, err := save.RestoreState(genesis, save.VersionForState(company), bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	replayedCompany, err := ApplyLogged(replayState, loggedPayload, bundle, loggedInputs)
	if err != nil || !jsonSemanticallyEqual(replayedCompany.Receipt, loggedReceipt) {
		t.Fatalf("live Company receipt is not replayable: live=%s replay=%s err=%v", loggedReceipt, replayedCompany.Receipt, err)
	}

	faultSteps := []string{"soul_session_terminal", "company_revision", "company_events", "run_log", "founder_revision", "founder_events", "founder_log", "intent_record", "retention"}
	for index, step := range faultSteps {
		sessionID := fmt.Sprintf("01986666-b2%02x-7000-8000-%012d", index, index+1)
		at := now.Add(25*time.Hour + time.Duration(index+10)*time.Second)
		if _, err := service.StartSoulRecovery(ctx, StartSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID,
			CompanyStreamID: companyRevision.StreamID, ActivityID: "touch_grass.fixture"}, at); err != nil {
			t.Fatal(err)
		}
		beforeCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
		beforeFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
		_, cancelErr := service.CancelSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID}, at, func(actual string) error {
			if actual == step {
				return errors.New("injected Soul recovery fault")
			}
			return nil
		})
		if cancelErr == nil {
			t.Fatalf("fault %s committed", step)
		}
		afterCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
		afterFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
		session, loadErr := repository.Load(ctx, founderID, sessionID)
		if loadErr != nil || session.Status != soul.RecoveryActive || afterCompany.Revision.Number != beforeCompany.Revision.Number ||
			afterFounder.Revision.Number != beforeFounder.Revision.Number {
			t.Fatalf("fault %s leaked company=%d/%d founder=%d/%d session=%+v err=%v", step,
				beforeCompany.Revision.Number, afterCompany.Revision.Number, beforeFounder.Revision.Number,
				afterFounder.Revision.Number, session, loadErr)
		}
		cleanup, err := service.CancelSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: sessionID, FounderID: founderID}, at, nil)
		if err != nil || !strings.Contains(string(cleanup.Receipt), `"cancelled_by":"player"`) {
			t.Fatalf("cleanup cancel after %s: %v", step, err)
		}
	}
}
