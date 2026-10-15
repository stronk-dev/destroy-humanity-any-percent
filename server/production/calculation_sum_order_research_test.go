package production

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// A retained counterexample, not an assertion that decreasing production is
// correct. It prevents a general monotonicity claim from relying on sampled
// arithmetic. The context changes by one purchased unit; time is fixed.
func TestCalculationSumOrderResearch(t *testing.T) {
	var fixture struct {
		Version  int
		Variable struct {
			ID         string `json:"id"`
			BaseRate   string `json:"base_rate"`
			LowerCount int64  `json:"lower_count"`
			UpperCount int64  `json:"upper_count"`
			LowerRate  string `json:"lower_rate"`
			UpperRate  string `json:"upper_rate"`
		}
		Fixed []struct {
			ID       string `json:"id"`
			BaseRate string `json:"base_rate"`
			Count    int64  `json:"count"`
			Rate     string `json:"rate"`
		}
		LowerPayout string `json:"lower_payout"`
		UpperPayout string `json:"upper_payout"`
	}
	source, err := os.ReadFile("../../testdata/numeric-sum-order-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(source, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Fixed) != 4 || fixture.Variable.UpperCount != fixture.Variable.LowerCount+1 {
		t.Fatal("changed counterexample population")
	}
	data, err := os.ReadFile("../../balance/testdata/epoch5/economy.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{}
	addGenerator := func(id, base string) {
		if _, err := decimal.ParseCanonical(base); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, map[string]any{"id": id,
			"price":      map[string]any{"resource_id": "company.cash", "base": "1e0", "curve": map[string]any{"kind": "constant"}},
			"production": map[string]any{"resource_id": "company.cash", "base_rate": base}})
	}
	addGenerator(fixture.Variable.ID, fixture.Variable.BaseRate)
	for _, row := range fixture.Fixed {
		addGenerator(row.ID, row.BaseRate)
	}
	document["generator_classes"], err = json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	amounts := []decimal.Decimal{}
	for index, count := range []int64{fixture.Variable.LowerCount, fixture.Variable.UpperCount} {
		state, err := save.RestoreState([]byte(`{"balances":{"company.cash":"0"}}`), 1, catalog, economy.ScopeCompany, engineCursor)
		if err != nil {
			t.Fatal(err)
		}
		state.GeneratorCounts[fixture.Variable.ID], state.GeneratorPurchasedTotal = count, count
		for _, row := range fixture.Fixed {
			state.GeneratorCounts[row.ID] = row.Count
			state.GeneratorPurchasedTotal += row.Count
		}
		encoded, err := save.EncodeState(state)
		if err != nil {
			t.Fatal(err)
		}
		state, err = save.RestoreState(encoded, save.VersionForState(state), catalog, economy.ScopeCompany, engineCursor)
		if err != nil {
			t.Fatal(err)
		}
		origin := append([]byte(nil), encoded...)
		wantRates := []decimal.Decimal{}
		for _, row := range fixture.Fixed {
			want := decimal.FromString(row.Rate)
			actual := decimal.FromString(row.BaseRate).Mul(decimal.FromString(strconv.FormatInt(row.Count, 10)))
			if !actual.IsStateValue() || !actual.Eq(want) {
				t.Fatalf("fixed producer differs: %s = (%g,e%d)", row.ID, actual.Mantissa(), actual.Exponent())
			}
			wantRates = append(wantRates, want)
		}
		variable := decimal.FromString(fixture.Variable.BaseRate).Mul(decimal.FromString(strconv.FormatInt(count, 10)))
		wantVariable, wantPayout := fixture.Variable.LowerRate, fixture.LowerPayout
		if index == 1 {
			wantVariable, wantPayout = fixture.Variable.UpperRate, fixture.UpperPayout
		}
		if !variable.Eq(decimal.FromString(wantVariable)) {
			t.Fatal("variable rate does not reproduce the literal counterexample")
		}
		gotRates, err := Rates(catalog, state.GeneratorCounts, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw := decimal.SumDeterministic(append(wantRates, variable))
		actualRaw := decimal.SumDeterministic(gotRates["company.cash"])
		if len(gotRates) != 1 || len(gotRates["company.cash"]) != 5 || !actualRaw.Eq(raw) {
			t.Fatal("actual production rates differ from counterexample")
		}
		before, err := save.EncodeState(state)
		if err != nil || !bytes.Equal(before, origin) {
			t.Fatal("rate reads changed the origin")
		}
		result, err := Evaluate(state, catalog, engineCursor.Add(time.Second), ModeOnline, nil)
		if err != nil {
			t.Fatal(err)
		}
		cash, _ := state.Ledger.Balance("company.cash")
		if cash.String() != wantPayout || result.ProductionMS != 1000 || len(result.Receipt.Changes) != 1 || result.Receipt.Changes[0].After != wantPayout {
			t.Fatalf("observed evaluator differs: count%d cash%s expected%s receipt%+v", count, cash, wantPayout, result.Receipt)
		}
		amounts = append(amounts, cash)
		t.Logf("count%d variable(%g,e%d) raw sum(%g,e%d) payout%s", count, variable.Mantissa(), variable.Exponent(), raw.Mantissa(), raw.Exponent(), cash)
	}
	if !amounts[0].Gt(amounts[1]) {
		t.Fatal("recorded decreasing payout no longer reproduces; investigate, do not restamp")
	}
}
