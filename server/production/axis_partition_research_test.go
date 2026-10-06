package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// R-012: bounded research models only. Neither the rational accumulator nor
// this experimental serialization is a production/save implementation.
type researchProjection struct {
	Wire    string `json:"wire"`
	Residue string `json:"residue"`
}

func researchRat(raw string) (*big.Rat, error) {
	if len(raw) > 160 {
		return nil, fmt.Errorf("research literal too long")
	}
	if at := strings.IndexAny(raw, "eE"); at >= 0 {
		exponent, err := strconv.Atoi(raw[at+1:])
		if err != nil || exponent < -32 || exponent > 32 {
			return nil, fmt.Errorf("research exponent outside -32..32")
		}
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, fmt.Errorf("invalid research rational")
	}
	return value, nil
}

// Produce an exact, normalized terminating decimal before crossing the SAME
// float64 quantizer as the existing numeric goldens. Ideal rational half-even
// rounding would silently replace the shipped midpoint semantics.
func researchProject(value *big.Rat) (string, error) {
	denominator := new(big.Int).Set(value.Denom())
	counts := [2]int{}
	for i, factor := range []int64{2, 5} {
		for denominator.Sign() > 0 {
			quotient, remainder := new(big.Int), new(big.Int)
			quotient.QuoRem(denominator, big.NewInt(factor), remainder)
			if remainder.Sign() != 0 {
				break
			}
			denominator = quotient
			counts[i]++
			if counts[i] > 48 {
				return "", fmt.Errorf("research scale exceeds48")
			}
		}
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return "", fmt.Errorf("nonterminating research rational")
	}
	scale := max(counts[0], counts[1])
	plain := value.FloatString(scale)
	sign := ""
	if strings.HasPrefix(plain, "-") {
		sign, plain = "-", plain[1:]
	}
	integer := strings.Split(plain, ".")[0]
	digits := strings.ReplaceAll(plain, ".", "")
	first := 0
	for first < len(digits) && digits[first] == '0' {
		first++
	}
	if first == len(digits) {
		return "0", nil
	}
	exponent := len(integer) - first - 1
	if exponent < -32 || exponent > 32 {
		return "", fmt.Errorf("research magnitude outside -32..32")
	}
	digits = strings.TrimRight(digits[first:], "0")
	coefficient := digits[:1]
	if len(digits) > 1 {
		coefficient += "." + digits[1:]
	}
	projected := decimal.FromString(sign + coefficient + "e" + strconv.Itoa(exponent)).Quantize(decimal.CanonicalSignificantDigits)
	if !projected.IsStateValue() {
		return "", fmt.Errorf("invalid research projection")
	}
	return projected.String(), nil
}

func researchCommit(value, cap *big.Rat) (researchProjection, error) {
	if value.Sign() < 0 {
		return researchProjection{}, fmt.Errorf("negative research balance")
	}
	if value.Cmp(cap) >= 0 {
		wire, err := researchProject(cap)
		return researchProjection{wire, "0"}, err
	}
	wire, err := researchProject(value)
	if err != nil {
		return researchProjection{}, err
	}
	visible, err := researchRat(wire)
	if err != nil {
		return researchProjection{}, err
	}
	return researchProjection{wire, new(big.Rat).Sub(value, visible).RatString()}, nil
}

func researchRestore(encoded []byte, cap *big.Rat) (*big.Rat, error) {
	var snapshot researchProjection
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, err
	}
	visible, err := decimal.ParseCanonical(snapshot.Wire)
	if err != nil || !visible.IsStateValue() {
		return nil, fmt.Errorf("invalid research wire")
	}
	wire, err := researchRat(snapshot.Wire)
	if err != nil {
		return nil, err
	}
	residue, err := researchRat(snapshot.Residue)
	if err != nil || residue.RatString() != snapshot.Residue {
		return nil, fmt.Errorf("invalid research residue")
	}
	value := new(big.Rat).Add(wire, residue)
	if value.Sign() < 0 || value.Cmp(cap) > 0 || wire.Cmp(cap) == 0 && residue.Sign() != 0 {
		return nil, fmt.Errorf("invalid reconstructed balance")
	}
	if _, err := researchProject(value); err != nil {
		return nil, err
	}
	return value, nil
}

func researchDelta(rate, efficiency string, elapsed int64) (*big.Rat, error) {
	if elapsed < 0 || elapsed > 60000 {
		return nil, fmt.Errorf("unsupported research interval")
	}
	r, err := researchRat(rate)
	if err != nil {
		return nil, err
	}
	e, err := researchRat(efficiency)
	if err != nil {
		return nil, err
	}
	if r.Sign() < 0 || e.Sign() < 0 {
		return nil, fmt.Errorf("negative research input")
	}
	return new(big.Rat).Mul(new(big.Rat).Mul(r, e), big.NewRat(elapsed, 1000)), nil
}

func mustResearchRat(t *testing.T, raw string) *big.Rat {
	t.Helper()
	value, err := researchRat(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
func mustResearchCommit(t *testing.T, value, cap *big.Rat) researchProjection {
	t.Helper()
	result, err := researchCommit(value, cap)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func researchAccumulate(t *testing.T, initial, rate, efficiency string, cuts []int64, cap *big.Rat, dropResidue bool) researchProjection {
	t.Helper()
	value := mustResearchRat(t, initial)
	previous := int64(0)
	var snapshot researchProjection
	for _, cut := range cuts {
		if cut <= previous {
			t.Fatal("unordered research cuts")
		}
		delta, err := researchDelta(rate, efficiency, cut-previous)
		if err != nil {
			t.Fatal(err)
		}
		snapshot = mustResearchCommit(t, new(big.Rat).Add(value, delta), cap)
		if dropResidue {
			snapshot.Residue = "0"
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		value, err = researchRestore(encoded, cap)
		if err != nil {
			t.Fatal(err)
		}
		previous = cut
	}
	return snapshot
}

type researchFrozenRow struct {
	ID                 string             `json:"id"`
	Rate               string             `json:"rate"`
	Initial            string             `json:"initial"`
	Efficiency         string             `json:"efficiency"`
	Cap                string             `json:"cap"`
	EndMS              int64              `json:"end_ms"`
	CutMS              int64              `json:"cut_ms"`
	CurrentOne         string             `json:"current_one"`
	CurrentSplit       string             `json:"current_split"`
	CurrentStatesEqual bool               `json:"current_states_equal"`
	Reference          researchProjection `json:"reference"`
	PrototypeOne       researchProjection `json:"prototype_one"`
	PrototypeSplit     researchProjection `json:"prototype_split"`
	Discarded          researchProjection `json:"discarded"`
}
type researchBoundaryRow struct {
	ID                string             `json:"id"`
	Initial           string             `json:"initial"`
	RateBefore        string             `json:"rate_before"`
	RateAfter         string             `json:"rate_after"`
	Efficiency        string             `json:"efficiency"`
	Cap               string             `json:"cap"`
	EndBefore         int64              `json:"end_before"`
	CutBefore         int64              `json:"cut_before"`
	EndAfter          int64              `json:"end_after"`
	CutAfter          int64              `json:"cut_after"`
	Debit             string             `json:"debit"`
	ReferenceBefore   researchProjection `json:"reference_before"`
	Before            researchProjection `json:"before"`
	AfterDebit        string             `json:"after_debit"`
	ReferenceAfter    researchProjection `json:"reference_after"`
	After             researchProjection `json:"after"`
	RetroactiveBefore researchProjection `json:"retroactive_before"`
}
type researchControl struct {
	Source   string `json:"source"`
	Expected string `json:"expected"`
}
type researchReport struct {
	SchemaVersion        int                   `json:"schema_version"`
	Seed                 uint32                `json:"seed"`
	Valid                bool                  `json:"measurement_valid"`
	Acceptance           string                `json:"acceptance_status"`
	BundleHash           string                `json:"bundle_hash"`
	Sources              map[string]string     `json:"source_sha256"`
	Exclusions           []string              `json:"excluded_domains"`
	Frozen               []researchFrozenRow   `json:"frozen"`
	Boundaries           []researchBoundaryRow `json:"boundaries"`
	Projection           []researchControl     `json:"projection_controls"`
	CurrentDivergences   int                   `json:"current_partition_divergences"`
	ReferenceDifferences int                   `json:"current_one_reference_differences"`
	DiscardHits          int                   `json:"discarded_residue_hits"`
	RetroactiveHits      int                   `json:"retroactive_hits"`
	CapHits              int                   `json:"retained_cap_excess_hits"`
	Refusals             int                   `json:"invalid_controls_rejected"`
}

func researchBundleCap(t *testing.T, bundle CatalogBundle, cap string) CatalogBundle {
	t.Helper()
	result := bundle
	result.Artifacts = cloneArtifactMap(bundle.Artifacts)
	var root map[string]any
	if err := json.Unmarshal(result.Artifacts["economy"], &root); err != nil {
		t.Fatal(err)
	}
	root["resources"].([]any)[0].(map[string]any)["hardcap"].(map[string]any)["amount"] = cap
	var err error
	result.Artifacts["economy"], err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	result.Economy, err = economy.LoadCatalog(result.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	result.ConstantsHash, err = save.ConstantsHashArtifacts(result.Artifacts)
	if err != nil || !result.valid(result.ConstantsHash) {
		t.Fatalf("research cap bundle admission: %v", err)
	}
	return result
}

func observeResearchFrozen(t *testing.T, bundle CatalogBundle, id, initial, rate, efficiency, capRaw string, end, cut int64) researchFrozenRow {
	t.Helper()
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	one, split := axisTimingState(t, bundle, start), axisTimingState(t, bundle, start)
	for _, state := range []*save.State{one, split} {
		setCash(t, state, initial)
		if rate == "2.4048e0" {
			state.GeneratorCounts["generator.beige_tower"] = 2
			state.AchievementsAttainedRun["achievement.generators_purchased_1"] = true
			state.AttainmentScoreRun = 8
		}
		if err := bundle.ValidateFoundationState(state); err != nil {
			t.Fatal(err)
		}
	}
	contributions, err := assembleContributions(one, bundle.Economy, nil)
	if err != nil {
		t.Fatal(err)
	}
	rates, err := Rates(bundle.Economy, one.GeneratorCounts, contributions)
	if err != nil || decimal.SumDeterministic(rates["company.cash"]).String() != rate {
		t.Fatalf("fixture rate=%v error=%v", rates, err)
	}
	mode := ModeOnline
	if efficiency == "9e-1" {
		mode = ModeOffline
	}
	if _, err := Evaluate(one, bundle.Economy, start.Add(time.Duration(end)*time.Millisecond), mode, contributions); err != nil {
		t.Fatal(err)
	}
	for _, at := range []int64{cut, end} {
		if _, err := Evaluate(split, bundle.Economy, start.Add(time.Duration(at)*time.Millisecond), mode, contributions); err != nil {
			t.Fatal(err)
		}
	}
	cap := mustResearchRat(t, capRaw)
	delta, err := researchDelta(rate, efficiency, end)
	if err != nil {
		t.Fatal(err)
	}
	reference := mustResearchCommit(t, new(big.Rat).Add(mustResearchRat(t, initial), delta), cap)
	row := researchFrozenRow{id, rate, initial, efficiency, capRaw, end, cut, one.Ledger.Snapshot()["company.cash"], split.Ledger.Snapshot()["company.cash"], bytes.Equal(mustEncodeState(t, one), mustEncodeState(t, split)), reference,
		researchAccumulate(t, initial, rate, efficiency, []int64{end}, cap, false), researchAccumulate(t, initial, rate, efficiency, []int64{cut, end}, cap, false), researchAccumulate(t, initial, rate, efficiency, []int64{cut, end}, cap, true)}
	if row.PrototypeOne != reference || row.PrototypeSplit != reference {
		t.Fatalf("research prototype differs: %+v", row)
	}
	if id == "ordinary/1.15115e0/1e4/1e0/00" && (!row.CurrentStatesEqual || row.CurrentOne != reference.Wire) {
		t.Fatal("whole-second control disagrees")
	}
	return row
}

func observeResearchBoundary(t *testing.T, id, initial, beforeRate, afterRate, efficiency, capRaw string, beforeEnd, beforeCut, afterEnd, afterCut int64) researchBoundaryRow {
	t.Helper()
	cap := mustResearchRat(t, capRaw)
	debit := "1.13e1"
	delta, _ := researchDelta(beforeRate, efficiency, beforeEnd)
	referenceBefore := mustResearchCommit(t, new(big.Rat).Add(mustResearchRat(t, initial), delta), cap)
	before := researchAccumulate(t, initial, beforeRate, efficiency, []int64{beforeCut, beforeEnd}, cap, false)
	// Explicit EXPERIMENTAL branch: an actual debit settles ordinary visible
	// K3 state and resets the provisional carry, rather than spending hidden precision.
	afterDebit := decimal.FromString(before.Wire).Sub(decimal.FromString(debit)).Quantize(decimal.CanonicalSignificantDigits).String()
	if afterDebit != decimal.FromString(referenceBefore.Wire).Sub(decimal.FromString(debit)).Quantize(decimal.CanonicalSignificantDigits).String() {
		t.Fatal("experimental debit differs")
	}
	delta, _ = researchDelta(afterRate, efficiency, afterEnd)
	referenceAfter := mustResearchCommit(t, new(big.Rat).Add(mustResearchRat(t, afterDebit), delta), cap)
	after := researchAccumulate(t, afterDebit, afterRate, efficiency, []int64{afterCut, afterEnd}, cap, false)
	retroDelta, _ := researchDelta(afterRate, efficiency, beforeEnd)
	retro := mustResearchCommit(t, new(big.Rat).Add(mustResearchRat(t, initial), retroDelta), cap)
	if before != referenceBefore || after != referenceAfter {
		t.Fatal("boundary prototype differs")
	}
	return researchBoundaryRow{id, initial, beforeRate, afterRate, efficiency, capRaw, beforeEnd, beforeCut, afterEnd, afterCut, debit, referenceBefore, before, afterDebit, referenceAfter, after, retro}
}

const partitionResearchPath = "../../testdata/axis-stack/partition-research-v1.json"

func TestAxisPartitionResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	report := researchReport{SchemaVersion: 1, Seed: 120307, Valid: true, Acceptance: "NOT_PROVEN: production AC6 remains red", BundleHash: bundle.ConstantsHash, Sources: map[string]string{}, Exclusions: []string{"exponents outside -32..32", "nonterminating or scale>48 rationals", "intervals>60000ms and offline-cap banking", "production save/replay representation", "served TS Company transitions, SQL and browsers"}}
	for _, path := range []string{"balance/testdata/axis-stack/economy-v5-fixture.json", "testdata/decimal-vectors.json", "server/production/axis_partition_research_test.go", "server/production/axis_timing_test.go", "server/production/engine.go", "server/production/content.go", "server/production/accrual.go", "server/economy/ledger.go", "server/decimal/decimal.go", "server/decimal/canonical.go", "client/src/numeric.ts", "client/src/production.ts", "client/test/partition-research.test.ts"} {
		data, err := os.ReadFile("../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		report.Sources[path] = hex.EncodeToString(sum[:])
	}
	state := uint32(120307)
	next := func() uint32 { state = 1664525*state + 1013904223; return state }
	type interval struct{ end, cut int64 }
	intervals := []interval{{2000, 1000}, {3114, 1553}, {60000, 1}}
	for len(intervals) < 32 {
		end := int64(2 + next()%59999)
		intervals = append(intervals, interval{end, 1 + int64(next()%uint32(end-1))})
	}
	normal := researchBundleCap(t, bundle, "1e12")
	capped := researchBundleCap(t, bundle, "1e4")
	for _, rate := range []string{"1.15115e0", "2.4048e0"} {
		for _, eff := range []string{"1e0", "9e-1"} {
			for _, initial := range []string{"0", "1e0", "1e4", "1e8"} {
				for i, span := range intervals {
					id := fmt.Sprintf("ordinary/%s/%s/%s/%02d", rate, initial, eff, i)
					report.Frozen = append(report.Frozen, observeResearchFrozen(t, normal, id, initial, rate, eff, "1e12", span.end, span.cut))
				}
			}
			for _, initial := range []string{"1e4", "9.99999999999e3"} {
				id := "cap/" + rate + "/" + initial + "/" + eff
				report.Frozen = append(report.Frozen, observeResearchFrozen(t, capped, id, initial, rate, eff, "1e4", 3114, 1553))
			}
		}
	}
	for _, eff := range []string{"1e0", "9e-1"} {
		for _, cut1 := range []int64{1, 1553} {
			for _, cut2 := range []int64{1, 2000} {
				report.Boundaries = append(report.Boundaries, observeResearchBoundary(t, fmt.Sprintf("rate-debit/%s/%d/%d", eff, cut1, cut2), "1e4", "1.15115e0", "2.4048e0", eff, "1e12", 3114, cut1, 4000, cut2))
			}
		}
	}
	for _, rate := range []string{"1.15115e0", "2.4048e0"} {
		for _, eff := range []string{"1e0", "9e-1"} {
			row := observeResearchBoundary(t, "cap-debit/"+rate+"/"+eff, "9.99999999999e3", rate, rate, eff, "1e4", 1000, 1, 1000, 1)
			if row.Before.Wire != "1e4" || row.Before.Residue != "0" {
				t.Fatal("cap retained overflow")
			}
			// A model conserving the pre-cap amount has a strictly positive hidden
			// excess. The zero-carry cap oracle rejects it before a later debit.
			delta, _ := researchDelta(rate, eff, 1000)
			excess := new(big.Rat).Sub(new(big.Rat).Add(mustResearchRat(t, row.Initial), delta), mustResearchRat(t, "1e4"))
			if excess.Sign() <= 0 {
				t.Fatal("cap control cannot discriminate")
			}
			report.CapHits++
			report.Boundaries = append(report.Boundaries, row)
		}
	}
	var golden struct {
		Vectors []struct{ Op, A, Expect string }
	}
	data, err := os.ReadFile("../../testdata/decimal-vectors.json")
	if err != nil || json.Unmarshal(data, &golden) != nil {
		t.Fatal("numeric controls unavailable")
	}
	for _, vector := range golden.Vectors {
		if vector.Op == "canonical" && len(report.Projection) < 9 {
			report.Projection = append(report.Projection, researchControl{vector.A, vector.Expect})
		}
	}
	report.Projection = append(report.Projection, researchControl{"9.999999999995e-4", "1e-3"})
	for _, control := range report.Projection {
		got, err := researchProject(mustResearchRat(t, control.Source))
		if err != nil || got != control.Expected {
			t.Fatalf("numeric projection %s got%s want%s err%v", control.Source, got, control.Expected, err)
		}
	}
	cap := mustResearchRat(t, "1e12")
	invalid := []func() error{
		func() error { _, err := researchRat("1e33"); return err }, func() error { _, err := researchRat("1e-33"); return err },
		func() error { _, err := researchProject(big.NewRat(1, 3)); return err }, func() error {
			_, err := researchProject(new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 49)))
			return err
		},
		func() error { _, err := researchRestore([]byte(`{"wire":"1e4"}`), cap); return err }, func() error { _, err := researchRestore([]byte(`{"wire":"0","residue":"-1"}`), cap); return err },
		func() error { _, err := researchRestore([]byte(`{"wire":"10000","residue":"0"}`), cap); return err }, func() error { _, err := researchDelta("1e0", "1e0", 60001); return err },
	}
	for i, control := range invalid {
		if control() == nil {
			t.Fatalf("invalid control%d admitted", i)
		}
		report.Refusals++
	}
	for _, row := range report.Frozen {
		if !row.CurrentStatesEqual {
			report.CurrentDivergences++
		}
		if row.CurrentOne != row.Reference.Wire {
			report.ReferenceDifferences++
		}
		if row.Discarded.Wire != row.Reference.Wire {
			report.DiscardHits++
		}
	}
	for _, row := range report.Boundaries {
		if strings.HasPrefix(row.ID, "rate-debit/") && row.RetroactiveBefore.Wire != row.ReferenceBefore.Wire {
			report.RetroactiveHits++
		}
	}
	if len(report.Frozen) != 520 || len(report.Boundaries) != 12 || len(report.Projection) != 10 || report.Refusals != 8 || report.CurrentDivergences == 0 || report.DiscardHits == 0 || report.RetroactiveHits != 8 || report.CapHits != 4 {
		t.Fatalf("incomplete/vacuous research: frozen%d boundary%d projection%d refusal%d current%d discard%d retro%d cap%d", len(report.Frozen), len(report.Boundaries), len(report.Projection), report.Refusals, report.CurrentDivergences, report.DiscardHits, report.RetroactiveHits, report.CapHits)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("UPDATE_PARTITION_RESEARCH") == "1" {
		if err := os.WriteFile(partitionResearchPath, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	committed, err := os.ReadFile(partitionResearchPath)
	if err != nil || !bytes.Equal(committed, encoded) {
		t.Fatalf("research source/results drifted; explicit re-observation required: %v", err)
	}
	t.Logf("R-012 valid bounded observation:542 cases/8 refusals; current divergences%d, one/reference%d, discard hits%d, retroactive8, cap4; AC6 NOT_PROVEN", report.CurrentDivergences, report.ReferenceDifferences, report.DiscardHits)
}
