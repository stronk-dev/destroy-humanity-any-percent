package gameui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type reputationProjectionCatalogs struct{ production.ReplayCatalogSet }

func (catalogs reputationProjectionCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	bundle, ok := catalogs.ResolveReplayCatalogs(hash)
	return bundle.Economy, ok
}

func (catalogs reputationProjectionCatalogs) ResolveRoutes(hash string) (*routes.Catalog, bool) {
	bundle, ok := catalogs.ResolveReplayCatalogs(hash)
	return bundle.Routes, ok
}

func (catalogs reputationProjectionCatalogs) ValidateState(hash string, state *save.State) error {
	bundle, ok := catalogs.ResolveReplayCatalogs(hash)
	if !ok {
		return production.ErrInvalidEngineState
	}
	return bundle.ValidateFoundationState(state)
}

type reputationProjectionSettlements struct{}

func (reputationProjectionSettlements) PendingSettlements(context.Context, string, string, int64) (guild.SettlementBatch, error) {
	return guild.SettlementBatch{}, nil
}

type reputationProjectionWeight struct{}

func (reputationProjectionWeight) CompactWeightPPM(context.Context, string, string, string) (int64, bool, error) {
	return 1000000, true, nil
}

// RP-302/AC7: actual committed states/public reads, not isolated header inputs.
// Initial earned6/generated5/tier1 are diagnostic, not natural progression.
func TestReputationCurrentAndNextPublicProjectionIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; persisted public projection NOT EXECUTED")
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
	bundle, company, founder, _ := reputationProjectionSource(t)
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1)`, bundle.ConstantsHash); err != nil {
		t.Fatal(err)
	}
	for name, data := range bundle.Artifacts {
		if _, err := db.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, bundle.ConstantsHash, name, data); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := db.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('AC7 diagnostic',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, bundle.ConstantsHash); err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	founder.ReputationLevel, founder.ReputationSpent, founder.ReputationUnlockPPM, founder.ReputationNodesOwned = 6, 0, 0, []string{}
	founder.EvaluatedThrough, founder.ManualTokenRefilledAt, founder.FiscalPeriodOpenedWallMS = now, now, now.UnixMilli()
	founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "scripted_first", OccurredAt: now.Add(-2 * time.Minute), ReputationDelta: 0}}
	company.EvaluatedThrough, company.ManualTokenRefilledAt, company.RunStartedAt = now, now, now
	company.GeneratorProvisioned["generator.beige_tower"] = 5
	const accountID = "01986666-e400-4000-8000-000000000001"
	const ownerID = "01986666-e400-4000-8000-000000000002"
	spawn, err := bundle.Opportunities.Spawn(ownerID, company.RunSeq, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.OpportunitySpawnSeq, company.NextOpportunityAttendedMS = 0, spawn.SpawnedAttendedMS
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, ownerID); err != nil {
		t.Fatal(err)
	}
	catalogs := reputationProjectionCatalogs{production.ReplayCatalogSet{bundle.ConstantsHash: bundle}}
	store, err := save.NewStore(db, catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "reputation.projection.integration"})
	if err != nil {
		t.Fatal(err)
	}
	player, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "reputation.projection.integration"})
	if err != nil {
		t.Fatal(err)
	}
	encode := func(value any) []byte {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	encodeState := func(state *save.State) []byte {
		t.Helper()
		data, err := save.EncodeState(state)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	appliedOutcome := func(result production.HandleResult) bool {
		var receipt struct {
			Outcome string `json:"outcome"`
		}
		return json.Unmarshal(result.Receipt, &receipt) == nil && receipt.Outcome == "applied"
	}
	genesis, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := production.FrozenFounderContributions(bundle, founder)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := save.PinRunWithGenesisTx(ctx, tx, player.StreamID, ownerID, company.RunSeq, bundle.ConstantsHash, player.Version, genesis); err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, player.StreamID, company.RunSeq, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	routeProjector, err := routeprojection.New(db, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	provider := production.FrozenContributionProvider{DB: db}
	service, err := production.NewService(store, catalogs, provider, nil, nil,
		production.WithProgressionRuntime(catalogs), production.WithRouteCatalogs(catalogs), production.WithRouteProjector(routeProjector),
		production.WithCurrentConstantsHash(bundle.ConstantsHash), production.WithReplayCatalogs(catalogs), production.WithGuildSettlements(reputationProjectionSettlements{}),
		production.WithMinigameActivity(repository), production.WithCompactPolicies(commons.CatalogSet{bundle.ConstantsHash: bundle.Commons.(commonsbinding.ReplayPolicy).Catalog}), production.WithCommonsWeightResolver(reputationProjectionWeight{}))
	if err != nil {
		t.Fatal(err)
	}
	projector, err := New(store, catalogs, provider, WithMinigameActivity(repository))
	if err != nil {
		t.Fatal(err)
	}
	tables := func() map[string]string {
		t.Helper()
		result := map[string]string{}
		for _, table := range []string{"save_streams", "save_revisions", "events", "run_epochs", "run_genesis", "run_frozen_contributions", "run_log", "founder_genesis", "founder_log", "intent_records", "transport_player_outbox", "verification_queue"} {
			var rows string
			query := fmt.Sprintf(`SELECT coalesce(jsonb_agg(to_jsonb(row) ORDER BY to_jsonb(row)::text),'[]'::jsonb)::text FROM %s row`, table)
			if err := db.QueryRowContext(ctx, query).Scan(&rows); err != nil {
				t.Fatal(err)
			}
			result[table] = rows
		}
		return result
	}
	read := func(at time.Time) snapshot {
		t.Helper()
		before := tables()
		data, err := projector.GameUISnapshot(ctx, player.StreamID, at)
		var value snapshot
		if err != nil || json.Unmarshal(data, &value) != nil {
			t.Fatalf("public persisted projection: %v", err)
		}
		if !bytes.Equal(encode(before), encode(tables())) {
			t.Fatal("public projection mutated recorded tables")
		}
		return value
	}
	assert := func(value snapshot, spent, available int64, current, next, rate string) {
		t.Helper()
		arm := value.Features.Reputation
		if arm == nil || arm.Level != 6 || arm.Spent != spent || arm.Available != available || arm.BonusFactorThisRun == nil || *arm.BonusFactorThisRun != current || arm.BonusFactorNextRun != next {
			t.Fatalf("public current/next header=%+v want spent%d available%d current%s next%s", arm, spent, available, current, next)
		}
		cashRows, towerRows := 0, 0
		for _, row := range value.Resources {
			if row.ResourceID == "company.cash" {
				cashRows++
				if row.RatePerSecond != rate {
					t.Fatalf("public cash rate=%s want%s", row.RatePerSecond, rate)
				}
			}
		}
		for _, row := range value.Generators {
			if row.GeneratorID == "generator.beige_tower" {
				towerRows++
				if row.RateContribution != rate || row.Provisioned != 5 {
					t.Fatalf("public tower row=%+v want rate%s/generated5", row, rate)
				}
			}
		}
		if cashRows != 1 || towerRows != 1 {
			t.Fatal("public rate rows missing or duplicated")
		}
	}
	retry := func(at time.Time, request []byte, result production.HandleResult) {
		t.Helper()
		before := tables()
		got, err := service.Handle(ctx, player.StreamID, production.ModeOnline, at, request)
		if err != nil || !got.Replay || !bytes.Equal(got.Receipt, result.Receipt) || !bytes.Equal(encode(before), encode(tables())) {
			t.Fatalf("exact retry changed receipt/tables: %v", err)
		}
	}
	before := read(now)
	assert(before, 0, 6, "1e0", "1e0", "5e0")
	companyBefore, err := store.LoadLatest(ctx, player.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	purchase := encode(map[string]any{"intent_id": "01986666-e401-7000-8000-000000000001", "kind": production.IntentPurchaseReputationNode, "expected_revision": owner.Number, "node_id": "reputation.unlock.p05"})
	applied, err := service.Handle(ctx, player.StreamID, production.ModeOnline, now, purchase)
	if err != nil || applied.Replay || !appliedOutcome(applied) {
		t.Fatalf("actual purchase: %s %v", applied.Receipt, err)
	}
	after := read(now)
	assert(after, 1, 5, "1e0", "1.003e0", "5e0")
	if !bytes.Equal(encode(before.Resources), encode(after.Resources)) || !bytes.Equal(encode(before.Generators), encode(after.Generators)) || after.Revision != before.Revision || after.FounderRevision != before.FounderRevision+1 {
		t.Fatal("mid-run purchase changed complete current rate rows or wrong revisions")
	}
	companyAfter, err := store.LoadLatest(ctx, player.StreamID)
	frozenAfter, frozenErr := save.LoadRunFrozenContributions(ctx, db, player.StreamID, company.RunSeq)
	if err != nil || frozenErr != nil || companyAfter.Revision != companyBefore.Revision || !bytes.Equal(encodeState(companyAfter.State), encodeState(companyBefore.State)) || !bytes.Equal(encode(frozenAfter), encode(frozen)) {
		t.Fatal("mid-run purchase changed complete Company head or frozen rows")
	}
	retry(now.Add(time.Second), purchase, applied)
	exit := encode(map[string]any{"intent_id": "01986666-e401-7000-8000-000000000002", "kind": production.IntentWindDown, "expected_revision": player.Number, "expected_founder_revision": owner.Number + 1, "reputation_plan": []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}})
	ended, err := service.Handle(ctx, player.StreamID, production.ModeOnline, now.Add(2*time.Second), exit)
	if err != nil || ended.Replay || !appliedOutcome(ended) {
		t.Fatalf("actual Exit: %s %v", ended.Receipt, err)
	}
	next := read(now.Add(2 * time.Second))
	assert(next, 6, 0, "1.003e0", "1.003e0", "5.015e0")
	newCompany, err := store.LoadLatest(ctx, player.StreamID)
	newFrozen, frozenErr := save.LoadRunFrozenContributions(ctx, db, player.StreamID, 3)
	if err != nil || frozenErr != nil || newCompany.State.RunSeq != 3 || newCompany.State.GeneratorPurchasedTotal != 0 || newCompany.State.GeneratorCounts["generator.beige_tower"] != 0 {
		t.Fatal("actual next run differs")
	}
	cash, ok := newCompany.State.Ledger.Balance("company.cash")
	if !ok || cash.String() != "1e3" {
		t.Fatal("next starter cash differs")
	}
	reputationRows := 0
	for _, row := range newFrozen {
		if row.SourceID == "reputation.founder_bonus" {
			reputationRows++
			if row.Factor != "1.003e0" {
				t.Fatal("actual next frozen factor differs")
			}
		}
	}
	if reputationRows != 1 {
		t.Fatal("actual next Reputation frozen population differs")
	}
	retry(now.Add(3*time.Second), exit, ended)
	history, err := store.LoadFounderHistory(ctx, owner.StreamID)
	if err != nil || len(history.Entries) != 2 || production.VerifyFounderHistory(history, catalogs) != production.ReplayVerified {
		t.Fatalf("actual purchase/Exit Founder history: %v", err)
	}
	t.Log("persisted public projection: actual mid-run purchase keeps complete current rows/factor; actual Exit changes next frozen/rates; all reads and retries preserve twelve tables")
}
