package faction

import (
	"os"
	"testing"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const testHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type testCatchupPolicies map[string]int64

func (policies testCatchupPolicies) ResolveCatchupCeilingMS(hash string) (int64, bool) {
	value, ok := policies[hash]
	return value, ok
}

func purchasableEconomyCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func productionCatalog(t *testing.T) *Catalog {
	t.Helper()
	commonsData, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalog, err := commons.LoadCatalog(commonsData)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/factions/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data, CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM,
		DefaultPPM: commonsCatalog.DefaultTithePPM,
		MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestStockAccrualCarriesRemainderAndSkipsOffline(t *testing.T) {
	catalog := productionCatalog(t)
	hook := AccrualHook{Catalogs: CatalogSet{testHash: catalog}, Policies: testCatchupPolicies{testHash: 30_000}}
	state := &save.State{FactionID: "bootstrapper", StockProgressMS: 50_000, ConsumedStockUnits: 17}
	revision := save.Revision{ConstantsHash: testHash}
	if events, err := hook.AfterAccrual(state, nil, revision, accrualhook.Result{ElapsedMS: 20_001}, nil); err != nil || len(events) != 0 {
		t.Fatalf("attended accrual events=%v err=%v", events, err)
	}
	if state.StockUnits != 1 || state.StockProgressMS != 10_001 || state.ConsumedStockUnits != 17 {
		t.Fatalf("stock=%d remainder=%d consumed=%d", state.StockUnits, state.StockProgressMS, state.ConsumedStockUnits)
	}
	if events, err := hook.AfterAccrual(state, nil, revision, accrualhook.Result{ElapsedMS: 30_001}, nil); err != nil || len(events) != 0 {
		t.Fatalf("offline accrual events=%v err=%v", events, err)
	}
	if state.StockUnits != 1 || state.StockProgressMS != 10_001 || state.ConsumedStockUnits != 17 {
		t.Fatalf("offline changed stock=%d remainder=%d consumed=%d", state.StockUnits, state.StockProgressMS, state.ConsumedStockUnits)
	}
}

func TestStockAccrualSaturatesOnceAndCarriesRemainder(t *testing.T) {
	catalog := productionCatalog(t)
	hook := AccrualHook{Catalogs: CatalogSet{testHash: catalog}, Policies: testCatchupPolicies{testHash: 120_000}}
	state := &save.State{FactionID: "open_source", StockUnits: catalog.StockCap - 1, StockProgressMS: 59_999}
	revision := save.Revision{ConstantsHash: testHash}
	events, err := hook.AfterAccrual(state, nil, revision, accrualhook.Result{ElapsedMS: 60_002}, nil)
	if err != nil || len(events) != 1 || events[0].Kind != save.EventFactionStockSaturated {
		t.Fatalf("events=%v err=%v", events, err)
	}
	if state.StockUnits != catalog.StockCap || state.StockProgressMS != 1 {
		t.Fatalf("stock=%d remainder=%d", state.StockUnits, state.StockProgressMS)
	}
	events, err = hook.AfterAccrual(state, nil, revision, accrualhook.Result{ElapsedMS: 1}, nil)
	if err != nil || len(events) != 0 || state.StockUnits != catalog.StockCap || state.StockProgressMS != 2 {
		t.Fatalf("second events=%v stock=%d remainder=%d err=%v", events, state.StockUnits, state.StockProgressMS, err)
	}
}

func TestStockRateRoleUsesPurchasedCountAndExactPPMRemainder(t *testing.T) {
	catalog := productionCatalog(t)
	economyCatalog := purchasableEconomyCatalog(t)
	hook := AccrualHook{Catalogs: CatalogSet{testHash: catalog}, Policies: testCatchupPolicies{testHash: 120_000}}
	state := &save.State{FactionID: "bootstrapper", GeneratorCounts: map[string]int64{"generator.high": 0, "generator.low": 10}, GeneratorProvisioned: map[string]int64{"generator.high": 0, "generator.low": 999}}
	revision := save.Revision{ConstantsHash: testHash}
	if events, err := hook.AfterAccrual(state, economyCatalog, revision, accrualhook.Result{ElapsedMS: 60_000}, nil); err != nil || len(events) != 0 {
		t.Fatalf("events=%v err=%v", events, err)
	}
	if state.StockUnits != 1 || state.StockProgressMS != 1_200 || state.StockRateRemainderPPM != 0 {
		t.Fatalf("stock=%d progress=%d ppm-remainder=%d", state.StockUnits, state.StockProgressMS, state.StockRateRemainderPPM)
	}
}
