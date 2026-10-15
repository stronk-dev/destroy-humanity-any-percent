package production

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Adversarial fixed-origin search, not a universal ordered-arithmetic claim.
// All completed buckets are fixed; only elapsed ms in the last bucket changes.
func TestCalculationBurstTemporalOrderResearch(t *testing.T) {
	seed := uint64(666)
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	speed := mustDecimal(t, "1e10")
	bonus := speed.Sub(decimal.One)
	threshold := mustDecimal(t, "1e1")
	uncrossed, confirmations := 0, 0
	for profile := 0; profile < 20_000; profile++ {
		tick := int64(1000 + next()%299001)
		burst := tick/4 + int64(next()%uint64(tick/2))
		prefix := int64(1 + next()%287)
		rate := decimal.FromFloat64(10 / (float64(burst) / 1000 * 1e10)).Quantize(decimal.CanonicalSignificantDigits)
		rates := []decimal.Decimal{rate}
		partial := func(ms int64) decimal.Decimal {
			value, err := accrueBoostedConstant(rates, ms, burst, decimal.One, bonus)
			if err != nil {
				t.Fatal(err)
			}
			return value
		}
		if partial(burst).Gte(threshold) {
			rate = rate.Sub(decimal.New(1, rate.Exponent()-11)).Quantize(decimal.CanonicalSignificantDigits)
			rates[0] = rate
		}
		if partial(burst).Gte(threshold) || partial(tick-1).Lt(threshold) {
			uncrossed++
			continue
		}
		low, high := burst+1, tick-1
		for low < high {
			middle := low + (high-low)/2
			if partial(middle).Gte(threshold) {
				high = middle
			} else {
				low = middle + 1
			}
		}
		before, after := partial(low-1), partial(low)
		if !before.Lt(threshold) || !after.Gte(threshold) {
			t.Fatal("search did not find an adjacent exponent carry", profile, low, before, after)
		}
		full, err := accrueBoostedConstant(rates, tick, tick, decimal.One, bonus)
		if err != nil {
			t.Fatal(err)
		}
		terms := make([]decimal.Decimal, prefix+1)
		for i := range terms {
			terms[i] = full
		}
		terms[prefix] = before
		lower := decimal.SumDeterministic(terms).Quantize(decimal.CanonicalSignificantDigits)
		terms[prefix] = after
		upper := decimal.SumDeterministic(terms).Quantize(decimal.CanonicalSignificantDigits)
		if !lower.IsStateValue() || !upper.IsStateValue() || !lower.Gt(decimal.Zero) || !upper.Gt(decimal.Zero) {
			t.Fatal("invalid or zero aggregate in the declared positive profile", profile, lower, upper)
		}
		if profile == 0 || profile == 19_999 || upper.Lt(lower) {
			confirmBurstTemporalPairResearch(t, tick, burst, prefix, rate, low, lower, upper)
			confirmations++
		}
		if upper.Lt(lower) {
			t.Fatalf("fixed-origin decrease profile%d: tick%d burst%d prefix%d rate%s partialms%d partial%s→%s full%s total%s→%s uncrossed%d", profile, tick, burst, prefix, rate, low, before, after, full, lower, upper, uncrossed)
		}
	}
	if uncrossed != 0 || confirmations != 2 {
		t.Fatal("declared carry population or real-evaluator controls changed", uncrossed, confirmations)
	}
	t.Log("declared20000 profiles complete, uncrossed0; two real-Evaluate adjacent pairs: no decrease found, not a general monotonicity proof")
}

func confirmBurstTemporalPairResearch(t *testing.T, tick, burst, prefix int64, rate decimal.Decimal, atMS int64, lower, upper decimal.Decimal) {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if json.Unmarshal(data, &root) != nil {
		t.Fatal("invalid source catalog")
	}
	root["provision_tick_ms"] = tick
	policy := root["offline_policy"].(map[string]any)
	policy["burst_speed"] = "1e10"
	policy["burst_max_duration_ms"] = prefix*tick + burst
	found := false
	for _, value := range root["generator_classes"].([]any) {
		generator := value.(map[string]any)
		if generator["id"] == "generator.beige_tower" {
			generator["production"].(map[string]any)["base_rate"] = rate.String()
			found = true
		}
	}
	if !found {
		t.Fatal("missing controlled producer")
	}
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal("retuned research catalog is not admitted", err)
	}
	state, err := save.RestoreState([]byte(`{"balances":{"company.cash":"0","company.permits":"0"}}`), 1,
		catalog, economy.ScopeCompany, engineCursor)
	if err != nil {
		t.Fatal(err)
	}
	state.WireVersion, state.ComputeBurstRemainingMS = 17, prefix*tick+burst
	state.AchievementsEarnedRun = map[string]bool{}
	state.GeneratorCounts["generator.beige_tower"], state.GeneratorPurchasedTotal = 1, 1
	origin, err := save.EncodeState(state)
	if err != nil {
		t.Fatal("research save is not encodable", err)
	}
	for index, partialMS := range []int64{atMS - 1, atMS} {
		current, err := save.RestoreState(origin, 17, catalog, economy.ScopeCompany, engineCursor)
		if err != nil {
			t.Fatal("research save is not admitted", err)
		}
		elapsedMS := prefix*tick + partialMS
		result, err := Evaluate(current, catalog, engineCursor.Add(time.Duration(elapsedMS)*time.Millisecond), ModeOnline, nil)
		want := lower
		if index == 1 {
			want = upper
		}
		balance, ok := current.Ledger.Balance("company.cash")
		if err != nil || !ok || !balance.Eq(want) || result.ProductionMS != elapsedMS || len(result.Receipt.Changes) != 1 {
			t.Fatal("bucket reference differs from actual Evaluate/ledger", index, elapsedMS, balance, want, result, err)
		}
		after, err := save.EncodeState(state)
		if err != nil || !bytes.Equal(after, origin) {
			t.Fatal("independent projections consume the origin", err)
		}
	}
}
