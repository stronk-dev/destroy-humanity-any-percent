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
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestStartMinigameAPISessionAtomicSequenceIdempotencyAndReplay(t *testing.T) {
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
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, now.UnixMilli(), 0
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{"minigame.pitch": true}
	founder.Soul, founder.SoulExhaustedSourceIDs = 50, []string{}
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
	production, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}

	for index, step := range []string{"founder_genesis", "founder_revision", "founder_events", "founder_log", "retention"} {
		sessionID := "01986666-ca01-7000-8000-00000000000" + string(rune('1'+index))
		key := "fault-" + step
		_, startErr := production.StartMinigameAPISession(ctx, platform, StartMinigameAPIRequest{
			SessionID: sessionID, FounderID: founderID, CompanyStreamID: companyRevision.StreamID,
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

	request := StartMinigameAPIRequest{SessionID: "01986666-ca01-7000-8000-000000000010", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, MinigameID: "pitch", IdempotencyKey: "create-pitch-1"}
	created, err := production.StartMinigameAPISession(ctx, platform, request, now, nil)
	if err != nil || created.Replay {
		t.Fatalf("create receipt=%s replay=%v err=%v", created.Receipt, created.Replay, err)
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
		SessionID: "01986666-ca01-7000-8000-000000000012", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, MinigameID: "pitch", IdempotencyKey: "create-pitch-2",
	}, now.Add(2*time.Second), nil)
	if !errors.Is(err, minigame.ErrExclusiveActivity) {
		t.Fatalf("second active session err=%v", err)
	}
	loadedFounder, _ = store.LoadLatest(ctx, founderRevision.StreamID)
	if loadedFounder.Revision.Number != 2 || loadedFounder.State.MinigameSessionSeq != 1 {
		t.Fatalf("rejected second session advanced Founder=%+v", loadedFounder)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || VerifyFounderHistory(history, set) != ReplayVerified {
		t.Fatalf("Founder replay verdict=%s err=%v", VerifyFounderHistory(history, set), err)
	}
}
