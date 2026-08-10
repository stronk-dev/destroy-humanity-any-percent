package production

import (
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

func TestRateProjectionUsesCanonicalContentAndActivePlayAssemblyWithoutMutation(t *testing.T) {
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
	var opportunityRoot map[string]any
	if json.Unmarshal(opportunityBytes, &opportunityRoot) != nil {
		t.Fatal("opportunity fixture")
	}
	opportunityRoot["combo_policy"].(map[string]any)["combo_cap"] = "1e1"
	opportunityBytes, _ = json.Marshal(opportunityRoot)
	opportunities, err := activeplay.LoadCatalog(opportunityBytes, catalog)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	counts, provisioned := map[string]int64{}, map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
	}
	counts["generator.beige_tower"] = 2
	provisioned["generator.beige_tower"] = 3
	state := &save.State{WireVersion: 18, Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned,
		UpgradesOwned: map[string]bool{}, ActiveBuffs: []save.ActiveBuff{
			{BuffInstanceID: "018f6b7c-9abc-7def-8abc-0123456789ab", EffectRowID: "active.production", ActivatedAttendedMS: 1, ExpiresAttendedMS: 10_000},
			{BuffInstanceID: "018f6b7c-9abc-7def-8abc-0123456789ac", EffectRowID: "active.production", ActivatedAttendedMS: 2, ExpiresAttendedMS: 10_000},
		}}
	bundle := CatalogBundle{Economy: catalog, Opportunities: opportunities}
	beforeCounts, beforeProvisioned := maps.Clone(state.GeneratorCounts), maps.Clone(state.GeneratorProvisioned)
	beforeBuffs, beforeLedger := append([]save.ActiveBuff(nil), state.ActiveBuffs...), state.Ledger.Snapshot()
	active, err := ProjectRates(bundle, state, nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	withoutBuffs := *state
	withoutBuffs.ActiveBuffs = []save.ActiveBuff{}
	base, err := ProjectRates(bundle, &withoutBuffs, nil, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(active.Generators) != 9 || active.Generators[0].GeneratorID != "generator.answering_machine" ||
		active.Generators[1].GeneratorID != "generator.beige_tower" || len(active.Resources) != 2 ||
		active.Resources[0].ResourceID != "company.cash" || active.Resources[1].ResourceID != "company.permits" {
		t.Fatalf("unsorted projection=%+v", active)
	}
	ratio := active.Generators[1].Rate.Div(base.Generators[1].Rate).Quantize(decimal.CanonicalSignificantDigits)
	if ratio.Gt(decimal.FromFloat64(10)) || ratio.Lt(decimal.FromFloat64(9.99)) {
		t.Fatalf("active clamp rate=%s base=%s", active.Generators[1].Rate, base.Generators[1].Rate)
	}
	if !reflect.DeepEqual(state.GeneratorCounts, beforeCounts) || !reflect.DeepEqual(state.GeneratorProvisioned, beforeProvisioned) ||
		!reflect.DeepEqual(state.ActiveBuffs, beforeBuffs) || !reflect.DeepEqual(state.Ledger.Snapshot(), beforeLedger) {
		t.Fatal("projection mutated replay-owned state")
	}
}

func TestResolveRateProjectionAttendedMSClassifiesLongGapOnClone(t *testing.T) {
	started := time.UnixMilli(1_800_000_000_000).UTC()
	ledger, err := economy.NewLedger(rateProjectionMinimalCatalog(t), economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	state := &save.State{WireVersion: 18, Ledger: ledger, RunStartedAt: started, EvaluatedThrough: started,
		OfflineSpans: []save.OfflineSpan{}, ActiveBuffs: []save.ActiveBuff{}}
	bundle := CatalogBundle{Opportunities: &activeplay.Catalog{}, Prestige: &prestigecore.Policy{CatchupCeilingMS: 60_000}}
	attended, err := ResolveRateProjectionAttendedMS(bundle, state, started.Add(time.Hour))
	if err != nil || attended != 0 || len(state.OfflineSpans) != 0 {
		t.Fatalf("attended=%d offline=%v err=%v", attended, state.OfflineSpans, err)
	}
}

func rateProjectionMinimalCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/epoch5/economy.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
