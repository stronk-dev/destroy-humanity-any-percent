package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// Diagnostic eligible state, not naturally earned progression. The service,
// migrations, epoch reconciliation, save transactions and stored replay are real.
func TestFirstContentEpochPersistedBoundaryIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	old, next := epoch5TestBundle(t), firstContentEpochBundle(t)
	resolver := integrationCatalogs{
		economy:  map[string]*economy.Catalog{old.ConstantsHash: old.Economy, next.ConstantsHash: next.Economy},
		routes:   map[string]*routes.Catalog{old.ConstantsHash: old.Routes, next.ConstantsHash: next.Routes},
		prestige: map[string]*prestigecore.Policy{old.ConstantsHash: old.Prestige, next.ConstantsHash: next.Prestige},
		factions: map[string]*faction.Catalog{old.ConstantsHash: old.Faction, next.ConstantsHash: next.Faction},
	}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	var databaseNow time.Time
	if err := db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(databaseNow)
	epochs, err := leaderboard.NewRepository(db, "../..")
	if err != nil {
		t.Fatal(err)
	}
	reconcileHistoricalContentEpoch(t, ctx, epochs, old, 5, now.Add(-time.Hour))
	const owner = "01986666-0601-7000-8000-000000000001"
	founder := initialPrestigeWitnessState(t, old.Economy, economy.ScopeFounder, now)
	founder.ReputationLevel, founder.Notoriety = 1, 17
	founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
	company := initialPrestigeWitnessState(t, old.Economy, economy.ScopeCompany, now)
	company.RunSeq, company.Tier = 2, 3
	company.RunStartedAt, company.LifetimeValue = now.Add(-20*time.Minute), decimal.New(27, 12)
	create := func(owner string, bundle CatalogBundle, founder, company *save.State) (save.Revision, save.Revision) {
		t.Helper()
		founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder,
			OwnerID: owner, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "first-content.integration"})
		if err != nil {
			t.Fatal(err)
		}
		companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder,
			OwnerID: owner, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "first-content.integration"})
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, owner, company.RunSeq,
			bundle.ConstantsHash, companyRevision.Version, mustEncodeState(t, company)); err != nil {
			t.Fatal(err)
		}
		frozen, err := FrozenFounderContributions(bundle, founder)
		if err != nil {
			t.Fatal(err)
		}
		if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, company.RunSeq, frozen); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		return founderRevision, companyRevision
	}
	founderRevision, companyRevision := create(owner, old, founder, company)
	reconcileHistoricalContentEpoch(t, ctx, epochs, next, 6, now)
	newService := func(db *sql.DB, store *save.Store) *Service {
		t.Helper()
		activity, err := minigame.NewRepository(db)
		if err != nil {
			t.Fatal(err)
		}
		service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
			WithProgressionRuntime(resolver), WithCurrentConstantsHash(next.ConstantsHash),
			WithReplayCatalogs(ReplayCatalogSet{old.ConstantsHash: old, next.ConstantsHash: next}),
			WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(activity))
		if err != nil {
			t.Fatal(err)
		}
		return service
	}
	request := []byte(`{"intent_id":"01986666-0601-7000-8000-000000000002","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
	result, err := newService(db, store).Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
	var receipt struct {
		Outcome string `json:"outcome"`
	}
	if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != "applied" {
		t.Fatalf("Exit receipt=%s replay=%t err=%v", result.Receipt, result.Replay, err)
	}
	// Use another connection pool and Store/Service; no in-memory instance can
	// supply the reloads or duplicate receipt.
	reopenedDB, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopenedDB.Close()
	reopened, err := save.NewStore(reopenedDB, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	loadedCompany, err := reopened.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	loadedFounder, err := reopened.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if loadedCompany.Revision.Number != 3 || loadedCompany.Revision.Version != 17 || loadedCompany.Revision.ConstantsHash != next.ConstantsHash ||
		loadedCompany.State.RunSeq != 3 || loadedFounder.Revision.Number != 2 || loadedFounder.Revision.Version != 21 ||
		loadedFounder.Revision.ConstantsHash != next.ConstantsHash || len(loadedFounder.State.ExitHistory) != 2 ||
		loadedFounder.State.ExitHistory[1].ExitType != "collapse" {
		t.Fatalf("persisted boundary company=%+v founder=%+v", loadedCompany, loadedFounder)
	}
	history, err := reopened.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || len(history.Entries) != 1 || history.Genesis.Version != 14 || history.Genesis.ConstantsHash != old.ConstantsHash ||
		loadedFounder.State.FiscalPeriodOpenedWallMS != history.Entries[0].ServerTSMS ||
		VerifyFounderHistory(history, ReplayCatalogSet{old.ConstantsHash: old, next.ConstantsHash: next}) != ReplayVerified {
		t.Fatalf("persisted Founder activation/history failed: %+v err=%v", history, err)
	}
	assertNewRunPermits(t, loadedCompany.State)
	for _, state := range []*save.State{loadedCompany.State, loadedFounder.State} {
		if err := next.ValidateFoundationState(state); err != nil {
			t.Fatal(err)
		}
	}
	if len(loadedCompany.State.AchievementsEarnedRun) != 0 || len(loadedFounder.State.AchievementsEarnedLifetime) != 0 ||
		len(loadedFounder.State.Pets) != 0 || len(loadedCompany.State.MeterValues) != 11 || loadedCompany.State.ActiveBuffs != nil {
		t.Fatal("activation granted historical rewards or imported later content")
	}
	for runSeq, expected := range map[int64]struct {
		epoch int64
		hash  string
	}{2: {5, old.ConstantsHash}, 3: {6, next.ConstantsHash}} {
		var epoch int64
		var hash string
		if err := reopenedDB.QueryRowContext(ctx, `SELECT epoch_id,constants_hash FROM run_epochs WHERE company_stream_id=$1 AND run_seq=$2`,
			companyRevision.StreamID, runSeq).Scan(&epoch, &hash); err != nil || epoch != expected.epoch || hash != expected.hash {
			t.Fatalf("run %d pin=(%d,%s) want=(%d,%s) err=%v", runSeq, epoch, hash, expected.epoch, expected.hash, err)
		}
	}
	var terminal []byte
	var terminalVersion int
	var terminalHash string
	if err := reopenedDB.QueryRowContext(ctx, `SELECT state,version,constants_hash FROM save_revisions WHERE stream_id=$1 AND revision=2`,
		companyRevision.StreamID).Scan(&terminal, &terminalVersion, &terminalHash); err != nil {
		t.Fatal(err)
	}
	ending, err := save.RestoreState(terminal, terminalVersion, old.Economy, economy.ScopeCompany, time.Time{})
	if err != nil || terminalVersion != 14 || terminalHash != old.ConstantsHash {
		t.Fatalf("old terminal version=%d hash=%s err=%v", terminalVersion, terminalHash, err)
	}
	if _, exists := ending.Ledger.Balance("company.permits"); exists {
		t.Fatal("persisted ending run gained permits")
	}
	if _, err := save.RestoreState(terminal, terminalVersion, next.Economy, economy.ScopeCompany, time.Time{}); err == nil {
		t.Fatal("old terminal silently restored against new resource universe")
	}
	genesis, version, entries := persistedPrestigeReplay(t, reopenedDB, companyRevision.StreamID, founderRevision.StreamID, 2, &next)
	if len(entries) != 1 || !entries[0].Terminal || VerifyReplayRun(genesis, version, old, entries, old.ConstantsHash, false) != ReplayVerified {
		t.Fatal("stored old run no longer replays with its exact catalog and next-run boundary")
	}
	if verdict := VerifyReplayRun(genesis, version, old, entries, next.ConstantsHash, false); verdict != ReplayConstantsMismatch {
		t.Fatalf("wrong-history hash accepted: %s", verdict)
	}
	forged := append([]ReplayLogEntry(nil), entries...)
	forged[0].ReceiptJSON = []byte(`{"outcome":"rejected"}`)
	if verdict := VerifyReplayRun(genesis, version, old, forged, old.ConstantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("forged stored receipt accepted: %s", verdict)
	}
	nextGenesis, err := reopened.LoadRunGenesis(ctx, companyRevision.StreamID, 3)
	if err != nil || nextGenesis.Version != 17 || nextGenesis.ConstantsHash != next.ConstantsHash {
		t.Fatalf("new genesis=%+v err=%v", nextGenesis, err)
	}
	if canonicalFixtureValue(t, json.RawMessage(nextGenesis.State)) != canonicalFixtureValue(t, json.RawMessage(mustEncodeState(t, loadedCompany.State))) {
		t.Fatal("new run genesis differs from committed initial state")
	}
	contributions, err := save.LoadRunFrozenContributions(ctx, reopenedDB, companyRevision.StreamID, 3)
	expectedContributions, expectedErr := FrozenFounderContributions(next, loadedFounder.State)
	if err != nil || expectedErr != nil || len(contributions) != len(next.Fiscal.GeneratorLevelRows())+1 ||
		canonicalFixtureValue(t, contributions) != canonicalFixtureValue(t, expectedContributions) {
		t.Fatalf("persisted new-run contributions=%+v err=%v expected_err=%v", contributions, err, expectedErr)
	}
	beforeCompany, beforeFounder := mustEncodeState(t, loadedCompany.State), mustEncodeState(t, loadedFounder.State)
	retry, err := newService(reopenedDB, reopened).Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(time.Second), request)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("persisted duplicate receipt=%s replay=%t err=%v", retry.Receipt, retry.Replay, err)
	}
	afterCompany, err := reopened.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	afterFounder, err := reopened.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || afterCompany.Revision.Number != 3 || afterFounder.Revision.Number != 2 ||
		!bytes.Equal(beforeCompany, mustEncodeState(t, afterCompany.State)) || !bytes.Equal(beforeFounder, mustEncodeState(t, afterFounder.State)) {
		t.Fatal("duplicate Exit changed committed states/revisions")
	}
}

// The real registry's history and artifact declarations, with only its current
// epoch shortened. Bytes are the exact checked historical bundles, not today's
// deploy-current files. This reconstructs a test deployment without minting files.
func reconcileHistoricalContentEpoch(t *testing.T, ctx context.Context, repository *leaderboard.Repository, bundle CatalogBundle, epochID int64, at time.Time) {
	t.Helper()
	deployed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	seed := deployed.Seed
	seed.CurrentEpochID, seed.Epochs = epochID, seed.Epochs[:epochID]
	seed.Artifacts = nil
	for _, declaration := range deployed.Seed.Artifacts {
		if _, present := bundle.Artifacts[declaration.Name]; present {
			seed.Artifacts = append(seed.Artifacts, declaration)
		}
	}
	if err := repository.ReconcileSeed(ctx, epochseed.Bundle{Seed: seed, Hash: bundle.ConstantsHash, Artifacts: bundle.Artifacts}, at); err != nil {
		t.Fatal(err)
	}
}
