package production

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Existing producer context only: no prototype/save-policy adoption or AC6 claim.
func TestAxisRateContextSQLIntegrationResearch(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("NOT_EXECUTED: real Postgres required; use make test-save-integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db, err := save.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal("declared database connection failed") // Do not print credentials.
	}
	defer db.Close()
	var version string
	if err := db.QueryRowContext(ctx, "SHOW server_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(version, "16.") {
		t.Fatal("invalid measurement: requires Postgres16")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "CREATE TEMP TABLE axis_context_research (ordinal integer PRIMARY KEY, snapshot jsonb NOT NULL) ON COMMIT DROP"); err != nil {
		t.Fatal(err)
	}
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	ordinal := 0
	roundTrip := func(t *testing.T, state *save.State) *save.State {
		t.Helper()
		if err := bundle.ValidateFoundationState(state); err != nil {
			t.Fatalf("input context admission: %v", err)
		}
		encoded := mustEncodeState(t, state)
		if _, err := tx.ExecContext(ctx, "INSERT INTO axis_context_research VALUES ($1,$2::jsonb)", ordinal, string(encoded)); err != nil {
			t.Fatal(err)
		}
		var stored []byte
		if err := tx.QueryRowContext(ctx, "SELECT snapshot::text FROM axis_context_research WHERE ordinal=$1", ordinal).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		ordinal++
		restored, err := save.RestoreState(stored, save.VersionForState(state), bundle.Economy, economy.ScopeCompany, start)
		if err != nil {
			t.Fatalf("SQL context restore: %v", err)
		}
		if err := bundle.ValidateFoundationState(restored); err != nil {
			t.Fatalf("SQL context admission: %v", err)
		}
		if !bytes.Equal(encoded, mustEncodeState(t, restored)) {
			t.Fatal("SQL changed existing context")
		}
		return restored
	}
	rawRates := func(t *testing.T, state *save.State) map[string][]decimal.Decimal {
		t.Helper()
		contributions, err := assembleContributions(state, bundle.Economy, nil)
		if err != nil {
			t.Fatal(err)
		}
		rates, err := Rates(bundle.Economy, state.GeneratorCounts, contributions)
		if err != nil || len(rates) != 1 || len(rates["company.cash"]) == 0 {
			t.Fatalf("actual rate population: %v", err)
		}
		return rates
	}
	advance := func(t *testing.T, state *save.State, elapsed int64, mode EvaluationMode) EvaluationResult {
		t.Helper()
		contributions, err := assembleContributions(state, bundle.Economy, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := Evaluate(state, bundle.Economy, start.Add(time.Duration(elapsed)*time.Millisecond), mode, contributions)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	profiles, controls := 0, 0
	for _, count := range []int64{1, 2, 9, 99, 12345, 123456789, 1234567890123, decimal.MaxExactInteger} {
		for _, efficiency := range []string{"1e0", "9e-1"} {
			for _, elapsed := range []int64{2000, 3114, 2045, 60000} {
				id := fmt.Sprintf("producer/%d/%s/%d", count, efficiency, elapsed)
				t.Run(id, func(t *testing.T) {
					state := axisTimingState(t, bundle, start)
					state.GeneratorCounts["generator.beige_tower"], state.GeneratorPurchasedTotal = count, count
					setCash(t, state, "0")
					original := mustEncodeState(t, state)
					restored := roundTrip(t, state)
					live, fromSQL := rawRates(t, state), rawRates(t, restored)
					if !axisContextRatesExact(live, fromSQL) {
						t.Fatal("SQL context changed raw rate bits/exponents")
					}
					altered := axisTimingState(t, bundle, start)
					alteredCount := int64(1)
					if count == 1 {
						alteredCount = 2
					}
					altered.GeneratorCounts["generator.beige_tower"], altered.GeneratorPurchasedTotal = alteredCount, alteredCount
					setCash(t, altered, "0")
					changed := roundTrip(t, altered)
					if bytes.Equal(original, mustEncodeState(t, changed)) || axisContextRatesExact(live, rawRates(t, changed)) {
						t.Fatal("valid changed-context control escaped state/rate comparison")
					}
					mode := ModeOnline
					if efficiency == "9e-1" {
						mode = ModeOffline
					}
					expectedCash, err := AccrueConstant(live["company.cash"], elapsed, mustDecimal(t, efficiency))
					if err != nil {
						t.Fatal(err)
					}
					one, sql := advance(t, state, elapsed, mode), advance(t, restored, elapsed, mode)
					if !reflect.DeepEqual(one, sql) || !bytes.Equal(mustEncodeState(t, state), mustEncodeState(t, restored)) {
						t.Fatal("SQL context changed full evaluation result/state")
					}
					if restored.Ledger.Snapshot()["company.cash"] != expectedCash.String() {
						t.Fatal("SQL context accrual differs from original raw primitive")
					}
					advance(t, changed, elapsed, mode)
					if changed.Ledger.Snapshot()["company.cash"] == expectedCash.String() {
						t.Fatal("valid changed-context control escaped payout comparison")
					}
					profiles++
					controls++
				})
			}
		}
	}
	if profiles != 64 || controls != 64 || ordinal != 128 {
		t.Fatalf("incomplete population: profiles%d controls%d SQLrows%d", profiles, controls, ordinal)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	t.Logf("existing producer SQL fidelity: profiles64 exact context/rate bits/full evaluation; changed-context controls64 discriminate; Postgres%s Go%s/%s; AC6 NOT_PROVEN", version, runtime.GOOS, runtime.GOARCH)
}

func axisContextRatesExact(a, b map[string][]decimal.Decimal) bool {
	if len(a) != len(b) {
		return false
	}
	for resource, rates := range a {
		other, ok := b[resource]
		if !ok || len(rates) != len(other) {
			return false
		}
		for i, rate := range rates {
			if math.Float64bits(rate.Mantissa()) != math.Float64bits(other[i].Mantissa()) || rate.Exponent() != other[i].Exponent() {
				return false
			}
		}
	}
	return true
}

func TestAxisContextRawRateComparator(t *testing.T) {
	rate := decimal.New(1.234567890123456, 10)
	original := map[string][]decimal.Decimal{"company.cash": {rate}}
	if !axisContextRatesExact(original, map[string][]decimal.Decimal{"company.cash": {rate}}) {
		t.Fatal("exact rates refused")
	}
	bitChanged := decimal.New(math.Nextafter(rate.Mantissa(), math.Inf(1)), rate.Exponent())
	if bitChanged.String() != rate.String() {
		t.Fatal("low-bit control must keep the same canonical rate text")
	}
	for name, changed := range map[string]map[string][]decimal.Decimal{
		"same-text-low-bit": {"company.cash": {bitChanged}},
		"exponent":          {"company.cash": {decimal.New(rate.Mantissa(), rate.Exponent()+1)}},
		"missing-resource":  {},
		"wrong-resource":    {"company.other": {rate}},
		"missing-rate":      {"company.cash": {}},
		"extra-rate":        {"company.cash": {rate, rate}},
	} {
		t.Run(name, func(t *testing.T) {
			if axisContextRatesExact(original, changed) {
				t.Fatal("changed raw-rate population accepted")
			}
		})
	}
}
