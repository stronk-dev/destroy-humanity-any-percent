package production

import (
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Manual RP-664 comparison on actual callers, not a performance acceptance gate.
// Run the same benchmark on active source and an unapplied compiler overlay.
// Setup/catalog parsing is excluded. Evaluate includes fresh save restoration;
// it must not coast on a cursor or balance left by an earlier iteration.
func BenchmarkAggregationCallerResearch(b *testing.B) {
	data, err := os.ReadFile("../../balance/testdata/epoch5/economy.json")
	if err != nil {
		b.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		b.Fatal(err)
	}
	counts := map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID] = 8
	}
	rates, err := Rates(catalog, counts, nil)
	if err != nil || len(rates["company.cash"]) == 0 {
		b.Fatal("missing actual cash rates", err)
	}
	b.Logf("epoch5: %d purchased classes, %d cash source rates", len(counts), len(rates["company.cash"]))
	axisCatalog, contributions := rateTraversalFixture(b)
	axisCounts := map[string]int64{}
	for _, generator := range axisCatalog.GeneratorClassesForScope(economy.ScopeCompany) {
		axisCounts[generator.ID] = 8
	}
	axisRates, err := Rates(axisCatalog, axisCounts, contributions)
	if err != nil || len(axisRates["company.cash"]) < 2 {
		b.Fatal("missing multi-source/all-slot rate population", err)
	}
	b.Logf("axis fixture: %d purchased classes, %d contributions, %d cash source rates", len(axisCounts), len(contributions), len(axisRates["company.cash"]))
	b.Run("accrue_axis_all_slots", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			value, err := AccrueConstant(axisRates["company.cash"], 1000, decimal.One)
			if err != nil || !value.IsStateValue() || !value.Gt(decimal.Zero) {
				b.Fatal("multi-source accrual did not produce a valid positive payout", value, err)
			}
		}
	})
	for _, elapsed := range []struct {
		name string
		ms   int64
	}{
		{"one_second", 1000},
		{"offline_cap", catalog.OfflinePolicy().AccrualCapMS},
	} {
		b.Run("accrue_"+elapsed.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				value, err := AccrueConstant(rates["company.cash"], elapsed.ms, decimal.One)
				if err != nil || !value.IsStateValue() || !value.Gt(decimal.Zero) {
					b.Fatal("accrual did not produce a valid positive payout", value, err)
				}
			}
		})
		b.Run("evaluate_"+elapsed.name, func(b *testing.B) {
			mode := ModeOnline
			if elapsed.name == "offline_cap" {
				mode = ModeOffline
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				state, err := save.RestoreState([]byte(`{"balances":{"company.cash":"0"}}`), 1,
					catalog, economy.ScopeCompany, engineCursor)
				if err != nil {
					b.Fatal(err)
				}
				for id, count := range counts {
					state.GeneratorCounts[id] = count
					state.GeneratorPurchasedTotal += count
				}
				result, err := Evaluate(state, catalog, engineCursor.Add(time.Duration(elapsed.ms)*time.Millisecond), mode, nil)
				cash, exists := state.Ledger.Balance("company.cash")
				if err != nil || result.ProductionMS != elapsed.ms || !exists || !cash.Gt(decimal.Zero) || len(result.Receipt.Changes) == 0 {
					b.Fatal("evaluation did not execute its full interval and ledger credit", result, cash, err)
				}
			}
		})
	}
}
