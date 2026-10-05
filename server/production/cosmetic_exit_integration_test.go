package production

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// AC8: both named Exit intents preserve actual acquired/equipped state in
// Postgres, and the same transition remains replayable on both history axes.
func TestCosmeticExitCarryIntegration(t *testing.T) {
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
	bundle := cosmeticsContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	minigames, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(minigames))
	if err != nil {
		t.Fatal(err)
	}
	for index, kind := range []string{IntentWindDown, IntentAcceptExitOffer} {
		t.Run(kind, func(t *testing.T) {
			var databaseNow time.Time
			if err := db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
				t.Fatal(err)
			}
			now := save.CanonicalServerTime(databaseNow)
			accountID := fmt.Sprintf("01986666-9f%02d-4000-8000-000000000001", index)
			founderID := fmt.Sprintf("01986666-9f%02d-7000-8000-000000000002", index)
			if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
				t.Fatal(err)
			}
			company := replayFixtureState(t, bundle.Economy, now.Add(-time.Minute))
			company.WireVersion, company.Tier, company.RunSeq = 18, 1, 2
			company.GatesCrossed["gate.t0_to_t1"] = true
			company.MeterBands = nil
			meterState, err := meters.NewRunState(bundle.Meters, 0)
			if err != nil {
				t.Fatal(err)
			}
			company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
			company.AchievementsEarnedRun = map[string]bool{}
			if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
				t.Fatal(err)
			}
			if kind == IntentAcceptExitOffer {
				seedCosmeticExitOffer(company, now)
			}
			companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
				bundle.ConstantsHash, company, save.WriteContext{Cause: "cosmetic.exit.integration"})
			if err != nil {
				t.Fatal(err)
			}
			founder := reputationFounderState(t, bundle, 24, now, 4)
			founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
			seedCosmeticExitWearer(t, bundle, founder, now)
			// Start empty: the owned/equipped input must be produced by the real
			// service commands below, rather than by the seeded pet fixture.
			founder.Cosmetics = cosmetic.NewState()
			founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
				bundle.ConstantsHash, founder, save.WriteContext{Cause: "cosmetic.exit.integration"})
			if err != nil {
				t.Fatal(err)
			}
			frozen, err := FrozenFounderContributions(bundle, founder)
			if err != nil {
				t.Fatal(err)
			}
			genesis := mustEncodeState(t, company)
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, founderID, 2, bundle.ConstantsHash, save.VersionForState(company), genesis); err != nil {
				t.Fatal(err)
			}
			if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 2, frozen); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			for step, body := range []string{
				`"kind":"acquire_cosmetic","expected_revision":1,"cosmetic_id":"horse_armor"`,
				`"kind":"equip_cosmetic","expected_revision":2,"cosmetic_id":"horse_armor","pet_id":"` + adoptedPetID + `"`,
			} {
				request := fmt.Sprintf(`{"intent_id":"01986666-9f%02d-7000-8000-%012d",%s}`, index, step+10, body)
				applied, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, []byte(request))
				if err != nil || !strings.Contains(string(applied.Receipt), `"outcome":"applied"`) {
					t.Fatalf("%s preparation: %s err=%v", kind, applied.Receipt, err)
				}
			}
			before, err := store.LoadLatest(ctx, founderRevision.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			if before.Revision.Number != 3 || len(before.State.Cosmetics.Owned) != 1 || before.State.Cosmetics.Equipped[adoptedPetID] != "horse_armor" {
				t.Fatalf("%s nonempty pre-Exit state: %+v", kind, before.State)
			}
			body := fmt.Sprintf(`{"intent_id":"01986666-9f%02d-7000-8000-000000000020","kind":%q,"expected_revision":1,"expected_founder_revision":3`, index, kind)
			if kind == IntentAcceptExitOffer {
				body += `,"offer_id":"` + company.OfferState.OfferID + `"`
			}
			request := []byte(body + "}")
			at := now.Add(5 * time.Second)
			result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, at, request)
			if err != nil || result.Replay || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
				t.Fatalf("%s Exit receipt=%s replay=%v err=%v", kind, result.Receipt, result.Replay, err)
			}
			after, err := store.LoadLatest(ctx, founderRevision.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			if after.Revision.Number != 4 || !before.State.Cosmetics.Equal(after.State.Cosmetics) {
				t.Fatalf("%s persisted Exit changed cosmetics: %+v", kind, after.State)
			}
			nextCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			if nextCompany.State.RunSeq != 3 || nextCompany.Revision.Number != 3 {
				t.Fatalf("%s did not advance to next Company: %+v", kind, nextCompany.State)
			}
			retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, at.Add(time.Second), request)
			if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) {
				t.Fatalf("%s Exit retry=%s replay=%v err=%v", kind, retry.Receipt, retry.Replay, err)
			}
			history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			if verdict := VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}); verdict != ReplayVerified {
				t.Fatalf("%s persisted Founder history: %s", kind, verdict)
			}
			if len(history.Entries) != 3 {
				t.Fatalf("%s acquisition/equip/Exit logged %d commands, want 3 despite retry", kind, len(history.Entries))
			}
			genesis, version, entries := persistedPrestigeReplay(t, db, companyRevision.StreamID, founderRevision.StreamID, 2, &bundle)
			if len(entries) != 1 || !entries[0].Terminal {
				t.Fatalf("%s Company history lacks its unique terminal: %+v", kind, entries)
			}
			if verdict := VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false); verdict != ReplayVerified {
				t.Fatalf("%s persisted Company history: %s", kind, verdict)
			}
		})
	}
}
