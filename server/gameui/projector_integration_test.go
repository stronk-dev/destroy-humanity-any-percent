package gameui

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type integrationCatalogs struct {
	hash    string
	economy *economy.Catalog
	bundle  production.CatalogBundle
}

func (catalogs integrationCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	return catalogs.economy, hash == catalogs.hash
}

func (catalogs integrationCatalogs) ResolveReplayCatalogs(hash string) (production.CatalogBundle, bool) {
	return catalogs.bundle, hash == catalogs.hash
}

func TestGameUISnapshotProjectsStoredSchemaV4CompanyV18RatesIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	economyBytes, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	opportunityBytes, err := os.ReadFile("../../balance/testdata/t0-t1/opportunities-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	opportunities, err := activeplay.LoadCatalog(opportunityBytes, catalog)
	if err != nil {
		t.Fatal(err)
	}
	routeBytes, err := os.ReadFile("../../balance/testdata/t0-t1/routes-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	hash := save.ConstantsHash(economyBytes)
	catalogs := integrationCatalogs{hash: hash, economy: catalog, bundle: production.CatalogBundle{
		ConstantsHash: hash, Economy: catalog, Routes: routeCatalog, Opportunities: opportunities,
		Prestige: &prestigecore.Policy{CatchupCeilingMS: 86_400_000},
	}}
	store, err := save.NewStore(db, catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.UnixMilli(1_800_000_000_000).UTC()
	company := candidateState(t, catalog)
	company.WireVersion = 18
	company.RunStartedAt = now.Add(-time.Minute)
	company.EvaluatedThrough = now
	company.ManualTokenRefilledAt = now
	company.GeneratorCounts["generator.beige_tower"] = 25
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = map[string]int{}, map[string]int64{}, map[string]int64{}
	company.AchievementsEarnedRun, company.AchievementsEarnedLifetime = map[string]bool{}, map[string]bool{}
	company.ComputeBurstRemainingMS = 0
	company.OpportunitySpawnSeq, company.NextOpportunityAttendedMS = 1, 120_000
	company.PendingOpportunity = nil
	company.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "018f6b7c-9abc-7def-8abc-0123456789ab",
		EffectRowID: "active.production", ActivatedAttendedMS: 1, ExpiresAttendedMS: 120_000}}
	ownerID := "018f6b7c-9abc-7def-8abc-012345678901"
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany},
		hash, company, save.WriteContext{Cause: "game-ui-v18-integration"})
	if err != nil {
		t.Fatal(err)
	}
	founderLedger, err := economy.NewLedger(catalog, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	founder := &save.State{Ledger: founderLedger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: now,
		ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
		CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero, OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeFounder},
		hash, founder, save.WriteContext{Cause: "game-ui-v18-integration-founder"})
	if err != nil {
		t.Fatal(err)
	}
	projector, err := New(store, catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := projector.GameUISnapshot(ctx, companyRevision.StreamID, now)
	if err != nil {
		t.Fatal(err)
	}
	var projected struct {
		FounderRevision int64 `json:"founder_revision"`
		SchemaVersion   int   `json:"schema_version"`
		Generators      []struct {
			GeneratorID      string `json:"generator_id"`
			RateContribution string `json:"rate_contribution"`
		} `json:"generators"`
		Resources []struct {
			ResourceID    string `json:"resource_id"`
			RatePerSecond string `json:"rate_per_second"`
		} `json:"resources"`
		Transitions struct {
			CrossGate *struct {
				Eligible bool   `json:"eligible"`
				GateID   string `json:"gate_id"`
			} `json:"cross_gate"`
			WindDown struct {
				Eligible bool `json:"eligible"`
			} `json:"wind_down"`
		} `json:"transitions"`
	}
	if json.Unmarshal(encoded, &projected) != nil {
		t.Fatal("snapshot JSON")
	}
	if projected.SchemaVersion != 3 || projected.FounderRevision != founderRevision.Number ||
		len(projected.Generators) != 9 || projected.Generators[1].GeneratorID != "generator.beige_tower" ||
		projected.Generators[1].RateContribution != "4.018e2" || len(projected.Resources) != 2 ||
		projected.Resources[0].ResourceID != "company.cash" || projected.Resources[0].RatePerSecond != "4.018e2" ||
		projected.Transitions.CrossGate == nil || !projected.Transitions.CrossGate.Eligible ||
		projected.Transitions.CrossGate.GateID != "gate.t0_to_t1" || projected.Transitions.WindDown.Eligible {
		t.Fatalf("v4/v18 projection=%+v", projected)
	}
}
