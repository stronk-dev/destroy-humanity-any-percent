package commonsbinding

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type fixedWeight int64

func (weight fixedWeight) CompactWeightPPM(string) (int64, bool) { return int64(weight), true }

func TestHookAdvancesRollingSolidarityAndCapacity(t *testing.T) {
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := commons.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const hash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hook := Hook{Catalogs: commons.CatalogSet{hash: catalog}, Weights: fixedWeight(1_000_000)}
	end := time.Date(2026, 7, 29, 15, 0, 0, 0, time.UTC)
	state := &save.State{CompactMember: true, CompactTithePPM: 100_000, RunSeq: 1, EvaluatedThrough: end}
	result := accrualhook.Result{ElapsedMS: 3_600_000, ProductionMS: 3_600_000, Receipt: economy.Receipt{Changes: []economy.Change{{ResourceID: "company.cash", Before: "0", Delta: "1e1", After: "1e1"}}}}
	events, err := hook.AfterAccrual(state, nil, save.Revision{StreamID: "11111111-1111-4111-8111-111111111111", OwnerID: "22222222-2222-4222-8222-222222222222", ConstantsHash: hash}, result, nil)
	if err != nil || len(events) != 1 || state.CompactSolidarityPPM != 1_388 || len(state.CompactSamples) != 1 {
		t.Fatalf("events=%+v solidarity=%d samples=%d err=%v", events, state.CompactSolidarityPPM, len(state.CompactSamples), err)
	}
	var payload struct {
		CompliancePPM int64  `json:"compliance_ppm"`
		Capacity      string `json:"capacity"`
	}
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil || payload.CompliancePPM != 1_000_000 || payload.Capacity != "1e0" {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
	state.CompactSamples = nil
	state.EvaluatedThrough = end.Add(30 * 24 * time.Hour)
	result.ElapsedMS = catalog.SolidarityWindowMS
	result.ProductionMS = catalog.SolidarityWindowMS
	if _, err := hook.AfterAccrual(state, nil, save.Revision{StreamID: "11111111-1111-4111-8111-111111111111", OwnerID: "22222222-2222-4222-8222-222222222222", ConstantsHash: hash}, result, nil); err != nil || state.CompactSolidarityPPM != 1_000_000 || len(state.CompactSamples) != 720 {
		t.Fatalf("full window solidarity=%d samples=%d err=%v", state.CompactSolidarityPPM, len(state.CompactSamples), err)
	}
}
