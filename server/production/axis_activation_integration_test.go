package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// Actual Service.Handle is the success path; faults use the existing Store
// callback with the LIVE service resolver, not a second Exit implementation.
func TestAxisActivationWriteFaultsIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, url)
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
	current, next := activeContentBundle(t), axisContentBundle(t)
	seedProductionEpoch(t, db, current.ConstantsHash, current.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{current.ConstantsHash: current.Economy, next.ConstantsHash: next.Economy},
		routes:   map[string]*routes.Catalog{current.ConstantsHash: current.Routes, next.ConstantsHash: next.Routes},
		prestige: map[string]*prestigecore.Policy{current.ConstantsHash: current.Prestige, next.ConstantsHash: next.Prestige},
		factions: map[string]*faction.Catalog{current.ConstantsHash: current.Faction, next.ConstantsHash: next.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	const account = "01986666-0f00-4000-8000-000000000001"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, account); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, account, activationFounderID); err != nil {
		t.Fatal(err)
	}
	// Founder Fiscal uses the database command clock, not the simulated Company
	// action clock. An admitted fixture must open its period before that clock.
	var databaseNow time.Time
	if err := db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
		t.Fatal(err)
	}
	start := save.CanonicalServerTime(databaseNow.Add(-time.Hour))
	company, founder := activationInitial(t, current, start, true)
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: activationFounderID, Scope: economy.ScopeCompany}, current.ConstantsHash, company, save.WriteContext{Cause: "axis.activation"})
	if err != nil {
		t.Fatal(err)
	}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: activationFounderID, Scope: economy.ScopeFounder}, current.ConstantsHash, founder, save.WriteContext{Cause: "axis.activation"})
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFounderContributions(current, founder)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, activationFounderID, 2, current.ConstantsHash, 18, mustEncodeState(t, company)); err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 2, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// Close the fixture's old current epoch, without changing its immutable pin.
	if _, err := db.ExecContext(ctx, `UPDATE epochs SET ended_at=clock_timestamp() WHERE ended_at IS NULL`); err != nil {
		t.Fatal(err)
	}
	seedProductionEpoch(t, db, next.ConstantsHash, next.Artifacts)
	activity, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{current.ConstantsHash: current, next.ConstantsHash: next}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithProgressionRuntime(resolver), WithCurrentConstantsHash(next.ConstantsHash), WithReplayCatalogs(set),
		WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(activity), WithRouteCatalogs(resolver))
	if err != nil {
		t.Fatal(err)
	}
	manual := []byte(`{"intent_id":"01986666-0f11-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1000}`)
	assertApplied := func(result HandleResult, err error) {
		t.Helper()
		var receipt struct {
			Outcome string `json:"outcome"`
		}
		if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != "applied" {
			t.Fatalf("Handle receipt=%s replay=%v err=%v", result.Receipt, result.Replay, err)
		}
	}
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, start.Add(time.Second), manual)
	assertApplied(result, err)
	loaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || loaded.Revision.ConstantsHash != current.ConstantsHash || loaded.State.WireVersion != 18 || loaded.State.AchievementsAttainedRun != nil {
		t.Fatalf("old pinned action differs: %v", err)
	}
	beforeCompany := mustEncodeState(t, loaded.State)
	beforeFounder := mustEncodeState(t, founder)
	exitTime := start.Add(4114 * time.Millisecond)
	exitBytes := []byte(`{"intent_id":"01986666-0f12-7000-8000-000000000002","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`)
	request, err := ParseIntent(exitBytes)
	if err != nil || request.InvalidDetail != "" {
		t.Fatal("invalid integration Exit request")
	}
	before := reputationPlanDBSnapshot(t, ctx, db)
	stages := []string{"founder_genesis", "founder_revision", "founder_events", "company_final_revision", "company_ended_events", "company_started_revision", "run_epoch", "run_frozen_contributions", "run_genesis", "company_started_events", "founder_log", "run_log", "intent_record", "retention"}
	firedCount := 0
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			injected := errors.New("injected pinned activation " + stage)
			fired := false
			_, err := store.ApplyExitTransactionLogged(ctx, companyRevision.StreamID, 2, 1, request.IntentID, request.RequestHash, request.CanonicalPayload,
				func(f *save.State, fr save.Revision, c *save.State, cr save.Revision, command save.ReplayCommand, fc save.FounderReplayCommand) (save.ExitDecision, json.RawMessage, error) {
					return service.applyLoggedExit(ctx, request, f, fr, c, cr, command, fc, ModeOnline, exitTime, []string{})
				}, func(actual string) error {
					if actual == stage {
						fired = true
						firedCount++
						return injected
					}
					return nil
				})
			if !fired || !errors.Is(err, injected) {
				t.Fatalf("fault not reached/exact: %v %v", fired, err)
			}
			after := reputationPlanDBSnapshot(t, ctx, db)
			for table, rows := range before {
				if after[table] != rows {
					t.Errorf("fault changed complete rows in %s", table)
				}
			}
			c, err := store.LoadLatest(ctx, companyRevision.StreamID)
			if err != nil || c.Revision.Number != 2 || c.Revision.ConstantsHash != current.ConstantsHash || !bytes.Equal(mustEncodeState(t, c.State), beforeCompany) {
				t.Fatalf("Company rollback differs: %v", err)
			}
			f, err := store.LoadLatest(ctx, founderRevision.StreamID)
			if err != nil || f.Revision.Number != 1 || f.Revision.ConstantsHash != current.ConstantsHash || !bytes.Equal(mustEncodeState(t, f.State), beforeFounder) {
				t.Fatalf("Founder rollback differs: %v", err)
			}
		})
	}
	if firedCount != 14 {
		t.Errorf("fault population fired=%d want14", firedCount)
	}
	if t.Failed() {
		return
	}
	result, err = service.Handle(ctx, companyRevision.StreamID, ModeOnline, exitTime, exitBytes)
	assertApplied(result, err)
	committed := reputationPlanDBSnapshot(t, ctx, db)
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, exitTime.Add(time.Second), exitBytes)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("Exit retry differs: %v", err)
	}
	retried := reputationPlanDBSnapshot(t, ctx, db)
	for table, rows := range committed {
		if retried[table] != rows {
			t.Fatalf("retry changed %s", table)
		}
	}
	loaded, err = store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loaded.Revision.Number != 4 || loaded.Revision.ConstantsHash != next.ConstantsHash || loaded.State.WireVersion != 19 || loaded.State.RunSeq != 3 || len(loaded.State.AchievementsAttainedRun) != 0 || loaded.State.AttainmentScoreRun != 0 {
		t.Fatalf("next run initial reset differs: %v", err)
	}
	if loaded.State.PendingOpportunity != nil || len(loaded.State.ActiveBuffs) != 0 || loaded.State.OpportunitySpawnSeq != 0 || loaded.State.NextOpportunityAttendedMS <= 0 {
		t.Fatal("next run scheduler not fresh")
	}
	newInitial := mustEncodeState(t, loaded.State)
	loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loadedFounder.Revision.Number != 2 || loadedFounder.Revision.ConstantsHash != next.ConstantsHash || loadedFounder.State.WireVersion != 21 || loadedFounder.State.AchievementScoreLifetime != founder.AchievementScoreLifetime || !bytes.Equal([]byte(canonicalFixtureValue(t, loadedFounder.State.AchievementsEarnedLifetime)), []byte(canonicalFixtureValue(t, founder.AchievementsEarnedLifetime))) {
		t.Fatalf("Founder floor/lifetime axis changed: %v", err)
	}
	for seq, bundle := range map[int64]CatalogBundle{2: current, 3: next} {
		genesis, err := store.LoadRunGenesis(ctx, companyRevision.StreamID, seq)
		wantVersion := 18
		if seq == 3 {
			wantVersion = 19
		}
		if err != nil || genesis.ConstantsHash != bundle.ConstantsHash || genesis.Version != wantVersion {
			t.Fatalf("run%d pin/genesis differs: %v", seq, err)
		}
		if seq == 3 && !canonicalJSONEqual(genesis.State, newInitial) {
			t.Fatal("new immutable genesis differs from actual initial persisted head")
		}
		var pin string
		if err := db.QueryRowContext(ctx, `SELECT constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2`, companyRevision.StreamID, seq).Scan(&pin); err != nil || pin != bundle.ConstantsHash {
			t.Fatalf("run%d actual pin differs: %v", seq, err)
		}
	}
	for _, action := range []struct {
		at  time.Duration
		raw string
	}{
		{time.Second, `{"intent_id":"01986666-0f13-7000-8000-000000000003","kind":"perform_manual_batch","expected_revision":4,"action_id":"manual.click","count":10,"window_ms":1000}`},
		{1001 * time.Millisecond, `{"intent_id":"01986666-0f14-7000-8000-000000000004","kind":"buy_generator","expected_revision":5,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`},
	} {
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, exitTime.Add(action.at), []byte(action.raw))
		assertApplied(result, err)
	}
	loaded, err = store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loaded.Revision.Number != 6 || loaded.State.AttainmentScoreRun != 2 || len(loaded.State.AchievementsAttainedRun) != 1 || !loaded.State.AchievementsAttainedRun["achievement.generators_purchased_1"] || loaded.State.AchievementScoreRun != 0 || len(loaded.State.AchievementsEarnedRun) != 0 {
		t.Fatalf("persisted re-attainment differs: %v", err)
	}
	// This existing reader separates automatic Fiscal prefixes (Founder-owned)
	// from Company replay events. The independent Founder verifier checks them.
	oldGenesis, oldVersion, oldEntries := reputationCareerReplay(t, db, companyRevision.StreamID, founderRevision.StreamID, 2, &next)
	newGenesis, newVersion, newEntries := reputationCareerReplay(t, db, companyRevision.StreamID, founderRevision.StreamID, 3, nil)
	verdict, oldFinal := verifyReplayRunDetailed(oldGenesis, oldVersion, current, oldEntries, current.ConstantsHash, false)
	if len(oldEntries) != 2 || oldVersion != 18 || verdict != ReplayVerified {
		t.Fatalf("old pinned complete history does not replay: %s", verdict)
	}
	var oldPersisted []byte
	var oldPersistedVersion int
	var oldHash string
	if err := db.QueryRowContext(ctx, `SELECT state,version,constants_hash FROM save_revisions WHERE stream_id=$1 AND revision=3`, companyRevision.StreamID).Scan(&oldPersisted, &oldPersistedVersion, &oldHash); err != nil || oldPersistedVersion != 18 || oldHash != current.ConstantsHash || !canonicalJSONEqual(oldPersisted, mustEncodeState(t, oldFinal)) {
		t.Fatalf("old complete terminal head differs: %v", err)
	}
	// Run3 is deliberately open: completed-run VerifyReplayRun must return
	// log_gap, not be misrepresented as a completed-run proof. Replay every
	// actual transition and compare its full persisted receipt/events/head.
	if len(newEntries) != 2 || newVersion != 19 || VerifyReplayRun(newGenesis, newVersion, next, newEntries, next.ConstantsHash, false) != ReplayLogGap {
		t.Fatal("new open-run population differs")
	}
	working, err := save.RestoreState(newGenesis, newVersion, next.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	for index, entry := range newEntries {
		if entry.Sequence != int64(index+1) || entry.Terminal {
			t.Fatal("new open-run log ordering differs")
		}
		transition, err := ApplyLogged(working, entry.CanonicalPayload, next, entry.ReplayInputs)
		if err != nil || transition.Outcome != save.IntentApplied || !canonicalJSONEqual(transition.Receipt, entry.ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(transition.Events), entry.EventsJSON) {
			t.Fatalf("new logged transition%d differs: %v", index, err)
		}
		working = transition.State
	}
	if !bytes.Equal(mustEncodeState(t, working), mustEncodeState(t, loaded.State)) {
		t.Fatal("new replay full persisted head differs")
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || len(history.Entries) != 1 || VerifyFounderHistory(history, set) != ReplayVerified {
		t.Fatalf("independent Founder history differs: %v", err)
	}
	t.Log("actual two-epoch Handle sequence; all14 write faults exact/full-row rollback; retry unchanged; old completed/Founder histories verified, new open-run full replay head exact; not retention pruning")
}
