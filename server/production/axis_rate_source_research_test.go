package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Diagnostic round-trip float text/bits, NOT a new permitted state wire format.
type sourceRateDiagnostic struct {
	Mantissa  string `json:"mantissa_diagnostic"`
	Exponent  string `json:"exponent_diagnostic"`
	Bits      string `json:"mantissa_ieee_hex"`
	Canonical string `json:"canonical"`
}
type sourceRateObservation struct {
	ID           string                 `json:"id"`
	Count        int64                  `json:"count"`
	Elapsed      int64                  `json:"elapsed_ms"`
	Efficiency   string                 `json:"efficiency"`
	Rates        []sourceRateDiagnostic `json:"rates"`
	RawWireEqual bool                   `json:"raw_wire_equal"`
	RawDelta     string                 `json:"raw_delta"`
	RoundedDelta string                 `json:"rounded_delta"`
	ContextDelta string                 `json:"restored_context_delta"`
	EngineCash   string                 `json:"actual_engine_cash"`
}
type sourceRateResearchReport struct {
	Version      int                     `json:"version"`
	Acceptance   string                  `json:"acceptance_status"`
	BundleHash   string                  `json:"fixture_bundle_hash"`
	Sources      map[string]string       `json:"source_sha256"`
	Rows         []sourceRateObservation `json:"rows"`
	BitsLost     int                     `json:"raw_bits_changed_profiles"`
	DeltaChanged int                     `json:"delta_changed_profiles"`
}

func sourceResearchHashes(t *testing.T, paths []string) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, path := range paths {
		data, err := os.ReadFile("../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		result[path] = hex.EncodeToString(sum[:])
	}
	return result
}

func sourceResearchArtifact(t *testing.T, path, switchName string, report any) {
	t.Helper()
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv(switchName) == "1" {
		if err := os.WriteFile(path, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	committed, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(encoded, committed) {
		t.Fatalf("research source/results drifted; explicit re-observation required: %v", err)
	}
}

const sourceRateResearchPath = "../../testdata/axis-stack/rate-source-research-v1.json"

func TestAxisRateSerializationResearch(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	report := sourceRateResearchReport{Version: 1, Acceptance: "NOT_PROVEN: production AC6 remains red", BundleHash: bundle.ConstantsHash,
		Sources: sourceResearchHashes(t, []string{"server/production/axis_rate_source_research_test.go", "client/test/rate-source-research.test.ts", "server/production/axis_timing_test.go", "server/production/axis_stack_test.go", "server/production/axis_partition_research_test.go", "balance/testdata/axis-stack/economy-v5-fixture.json", "server/production/engine.go", "server/production/content.go", "server/production/accrual.go", "server/decimal/decimal.go", "server/decimal/canonical.go", "server/save/state.go", "client/src/numeric.ts", "client/src/production.ts"})}
	for _, count := range []int64{1, 2, 9, 99, 12345, 123456789, 1234567890123, decimal.MaxExactInteger} {
		for _, efficiency := range []string{"1e0", "9e-1"} {
			for _, elapsed := range []int64{2000, 3114, 2045, 60000} {
				state := axisTimingState(t, bundle, start)
				state.GeneratorCounts["generator.beige_tower"], state.GeneratorPurchasedTotal = count, count
				setCash(t, state, "0")
				if err := bundle.ValidateFoundationState(state); err != nil {
					t.Fatalf("profile count%d admission: %v", count, err)
				}
				contributions, err := assembleContributions(state, bundle.Economy, nil)
				if err != nil {
					t.Fatal(err)
				}
				rates, err := Rates(bundle.Economy, state.GeneratorCounts, contributions)
				if err != nil || len(rates) != 1 || len(rates["company.cash"]) == 0 {
					t.Fatalf("actual rate population: %v", err)
				}
				live := rates["company.cash"]
				encoded := mustEncodeState(t, state)
				restored, err := save.RestoreState(encoded, save.VersionForState(state), bundle.Economy, economy.ScopeCompany, start)
				if err != nil {
					t.Fatalf("existing context restore: %v", err)
				}
				if err := bundle.ValidateFoundationState(restored); err != nil {
					t.Fatalf("restored context admission: %v", err)
				}
				restoredContributions, err := assembleContributions(restored, bundle.Economy, nil)
				if err != nil {
					t.Fatal(err)
				}
				contextRates, err := Rates(bundle.Economy, restored.GeneratorCounts, restoredContributions)
				if err != nil || len(contextRates["company.cash"]) != len(live) {
					t.Fatalf("context rate population: %v", err)
				}
				row := sourceRateObservation{ID: fmt.Sprintf("producer/%d/%s/%d", count, efficiency, elapsed), Count: count, Elapsed: elapsed, Efficiency: efficiency, RawWireEqual: true}
				rounded := make([]decimal.Decimal, len(live))
				for i, rate := range live {
					if !rate.Eq(contextRates["company.cash"][i]) {
						t.Fatal("existing frozen context changed raw rate")
					}
					rounded[i], err = decimal.ParseCanonical(rate.String())
					if err != nil {
						t.Fatal(err)
					}
					row.RawWireEqual = row.RawWireEqual && rate.Eq(rounded[i])
					row.Rates = append(row.Rates, sourceRateDiagnostic{strconv.FormatFloat(rate.Mantissa(), 'g', 17, 64), strconv.FormatInt(rate.Exponent(), 10), fmt.Sprintf("%016x", math.Float64bits(rate.Mantissa())), rate.String()})
				}
				eff := mustDecimal(t, efficiency)
				rawDelta, err := AccrueConstant(live, elapsed, eff)
				if err != nil {
					t.Fatal(err)
				}
				roundedDelta, err := AccrueConstant(rounded, elapsed, eff)
				if err != nil {
					t.Fatal(err)
				}
				contextDelta, err := AccrueConstant(contextRates["company.cash"], elapsed, eff)
				if err != nil || !rawDelta.Eq(contextDelta) {
					t.Fatalf("context accrual changed: %v", err)
				}
				mode := ModeOnline
				if efficiency == "9e-1" {
					mode = ModeOffline
				}
				if _, err := Evaluate(state, bundle.Economy, start.Add(time.Duration(elapsed)*time.Millisecond), mode, contributions); err != nil {
					t.Fatal(err)
				}
				row.RawDelta, row.RoundedDelta, row.ContextDelta, row.EngineCash = rawDelta.String(), roundedDelta.String(), contextDelta.String(), state.Ledger.Snapshot()["company.cash"]
				if row.EngineCash != row.RawDelta {
					t.Fatalf("actual engine differs from raw primitive: %+v", row)
				}
				if !row.RawWireEqual {
					report.BitsLost++
				}
				if row.RawDelta != row.RoundedDelta {
					report.DeltaChanged++
				}
				report.Rows = append(report.Rows, row)
			}
		}
	}
	if len(report.Rows) != 64 {
		t.Fatal("incomplete producer population")
	}
	sourceResearchArtifact(t, sourceRateResearchPath, "UPDATE_RATE_RESEARCH", report)
	t.Logf("R-012 actual producer observation:64 profiles; raw bits changed%d, delta changed%d, existing restored contexts exact64; AC6 NOT_PROVEN", report.BitsLost, report.DeltaChanged)
}
