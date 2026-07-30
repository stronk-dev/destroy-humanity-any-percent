package faction

import (
	"os"
	"testing"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/save"
)

const testHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

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
	hook := AccrualHook{Catalogs: CatalogSet{testHash: catalog}, CatchupCeilingMS: 30_000}
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
	hook := AccrualHook{Catalogs: CatalogSet{testHash: catalog}, CatchupCeilingMS: 120_000}
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
