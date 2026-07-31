package guild

import (
	"testing"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

func TestAccrualTitheCarryAndHealth(t *testing.T) {
	catalog, err := LoadCatalog([]byte(phase0Catalog))
	if err != nil {
		t.Fatal(err)
	}
	hash := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hook := AccrualHook{Catalogs: CatalogSet{hash: catalog}}
	state := &save.State{RunSeq: 2}
	revision := save.Revision{StreamID: "018f0000-0000-7000-8000-000000000001", OwnerID: "018f0000-0000-4000-8000-000000000001", ConstantsHash: hash}
	contributions := []multiplier.Contribution{{Slot: multiplier.SlotFaction, SourceID: StockConsumptionSourceID, Target: "all", Factor: decimal.One}}
	for index := 0; index < 2; index++ {
		events, err := hook.AfterAccrual(state, nil, revision, accrualhook.Result{ProgressDeltaPPM: 25}, contributions)
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 && (len(events) != 0 || state.GuildTitheCarryPPM != 500_000) {
			t.Fatalf("first tithe events=%v carry=%d", events, state.GuildTitheCarryPPM)
		}
		if index == 1 && (len(events) != 1 || events[0].Kind != save.EventGuildTitheAccrued || state.GuildTitheCarryPPM != 0) {
			t.Fatalf("second tithe events=%v carry=%d", events, state.GuildTitheCarryPPM)
		}
	}
	if health, err := HealthPPM(125_000, 2, catalog.GuildXPTargetPerFounder); err != nil || health != 250_000 {
		t.Fatalf("health=%d err=%v", health, err)
	}
	if health, err := HealthPPM(0, 0, catalog.GuildXPTargetPerFounder); err != nil || health != 0 {
		t.Fatalf("inactive health=%d err=%v", health, err)
	}
}

func TestApplySettlementsWatermark(t *testing.T) {
	state := &save.State{StockUnits: 20, ConsumedStockUnits: 3}
	settlements := []Settlement{{BoundarySeq: 1, DebitUnits: 5, CreditUnits: 7}, {BoundarySeq: 2, DebitUnits: 2, CreditUnits: 1}}
	if err := ApplySettlements(state, settlements, 100); err != nil {
		t.Fatal(err)
	}
	if state.StockUnits != 13 || state.ConsumedStockUnits != 11 || state.GuildBoundarySeq != 2 || state.GuildConsumedWindow != 1 {
		t.Fatalf("state=%+v", state)
	}
	if err := ApplySettlements(state, []Settlement{{BoundarySeq: 2}}, 100); err == nil {
		t.Fatal("duplicate boundary applied")
	}
}
