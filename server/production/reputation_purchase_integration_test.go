package production

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// TestReputationPurchaseIntegrationRecordsReplayableFounderLog is AC3's
// real-Postgres half: applied and rejected purchases go through
// Service.Handle and ApplyFounderLogged, persist their event through the
// store's strict validator, replay idempotently, and verify as a history.
func TestReputationPurchaseIntegrationRecordsReplayableFounderLog(t *testing.T) {
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
	bundle := reputationContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	const accountID = "01986666-5f00-4000-8000-000000000001"
	const founderID = "01986666-5f00-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "reputation.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 22, now, 3)
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "reputation.integration"})
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	purchase := []byte(`{"intent_id":"01986666-5f01-7000-8000-000000000001","kind":"purchase_reputation_node","expected_revision":1,"node_id":"reputation.unlock.p05"}`)
	applied, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, purchase)
	if err != nil || !bytes.Contains(applied.Receipt, []byte(`"outcome":"applied"`)) || !bytes.Contains(applied.Receipt, []byte(`"effective_from":"next_run"`)) {
		t.Fatalf("purchase receipt=%s err=%v", applied.Receipt, err)
	}
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, purchase)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, applied.Receipt) {
		t.Fatalf("idempotent retry receipt=%s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	owned := []byte(`{"intent_id":"01986666-5f01-7000-8000-000000000002","kind":"purchase_reputation_node","expected_revision":2,"node_id":"reputation.unlock.p05"}`)
	rejected, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, owned)
	if err != nil || !bytes.Contains(rejected.Receipt, []byte(`"detail":"owned"`)) {
		t.Fatalf("owned rejection receipt=%s err=%v", rejected.Receipt, err)
	}
	loaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || loaded.State.ReputationSpent != 1 || loaded.State.ReputationUnlockPPM != 50_000 ||
		len(loaded.State.ReputationNodesOwned) != 1 {
		t.Fatalf("persisted Founder revision=%d state=%+v err=%v", loaded.Revision.Number, loaded.State, err)
	}
	var events int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind='reputation_node_purchased.v1'`).Scan(&events); err != nil || events != 1 {
		t.Fatalf("persisted purchase events=%d err=%v", events, err)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || len(history.Entries) != 2 {
		t.Fatalf("Founder log entries=%d err=%v (the rejection must be a recorded row)", len(history.Entries), err)
	}
	if verdict := VerifyFounderHistory(history, set); verdict != ReplayVerified {
		t.Fatalf("Founder history verdict=%s", verdict)
	}
}
