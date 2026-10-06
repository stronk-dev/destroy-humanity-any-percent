package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
)

// R-012 research only. This bounded codec/anchor is NOT a production save field.
type anchorResearchSnapshot struct {
	Version    int      `json:"version"`
	Initial    string   `json:"initial"`
	Rates      []string `json:"rates"`
	Efficiency string   `json:"efficiency"`
	Cap        string   `json:"cap"`
	Elapsed    int64    `json:"elapsed_ms"`
	Wire       string   `json:"wire"`
}

func anchorResearchValue(raw string) (decimal.Decimal, error) {
	value, err := decimal.ParseCanonical(raw)
	if err != nil || !value.IsStateValue() || value.Lt(decimal.Zero) {
		return decimal.NaN, fmt.Errorf("invalid research numeric value %q", raw)
	}
	return value, nil
}

func anchorResearchBuild(initial string, rates []string, efficiency, cap string, elapsed int64) (anchorResearchSnapshot, error) {
	result := anchorResearchSnapshot{1, initial, rates, efficiency, cap, elapsed, ""}
	if rates == nil || len(rates) > 3 {
		return result, fmt.Errorf("unsupported research sources")
	}
	base, err := anchorResearchValue(initial)
	if err != nil {
		return result, err
	}
	ceiling, err := anchorResearchValue(cap)
	if err != nil || base.Gt(ceiling) {
		return result, fmt.Errorf("invalid initial/cap")
	}
	eff, err := anchorResearchValue(efficiency)
	if err != nil {
		return result, err
	}
	parsed := make([]decimal.Decimal, len(rates))
	for i, rate := range rates {
		parsed[i], err = anchorResearchValue(rate)
		if err != nil {
			return result, err
		}
	}
	delta, err := AccrueConstant(parsed, elapsed, eff)
	if err != nil {
		return result, err
	}
	value := base.Add(delta).Quantize(decimal.CanonicalSignificantDigits)
	if !value.IsStateValue() {
		return result, fmt.Errorf("invalid research sum before cap")
	}
	result.Wire = value.Min(ceiling).String()
	return result, nil
}

func anchorResearchRestore(encoded []byte) (anchorResearchSnapshot, error) {
	var snapshot anchorResearchSnapshot
	if len(encoded) > 512 {
		return snapshot, fmt.Errorf("unsupported research JSON size")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return snapshot, err
	}
	if decoder.Decode(new(any)) != io.EOF || snapshot.Version != 1 {
		return snapshot, fmt.Errorf("invalid research framing/version")
	}
	reconstructed, err := anchorResearchBuild(snapshot.Initial, snapshot.Rates, snapshot.Efficiency, snapshot.Cap, snapshot.Elapsed)
	if err != nil || reconstructed.Wire != snapshot.Wire {
		return snapshot, fmt.Errorf("invalid research reconstruction: %v", err)
	}
	canonical, err := json.Marshal(reconstructed)
	if err != nil || !bytes.Equal(canonical, encoded) {
		return snapshot, fmt.Errorf("noncanonical research JSON")
	}
	return reconstructed, nil
}

func anchorResearchRun(initial string, rates []string, efficiency, cap string, cuts []int64, rebase bool) (anchorResearchSnapshot, error) {
	snapshot, err := anchorResearchBuild(initial, rates, efficiency, cap, 0)
	previous := int64(0)
	for _, at := range cuts {
		if err != nil {
			return snapshot, err
		}
		encoded, encodeErr := json.Marshal(snapshot)
		if encodeErr != nil {
			return snapshot, encodeErr
		}
		snapshot, err = anchorResearchRestore(encoded)
		if err != nil || at < previous || at > decimal.MaxExactInteger {
			return snapshot, fmt.Errorf("invalid research interval/restore: %v", err)
		}
		start, elapsed := snapshot.Initial, at
		if rebase {
			start, elapsed = snapshot.Wire, at-previous
		}
		snapshot, err = anchorResearchBuild(start, rates, efficiency, cap, elapsed)
		previous = at
	}
	if err != nil {
		return snapshot, err
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}
	return anchorResearchRestore(encoded)
}

type anchorResearchCase struct {
	ID            string                  `json:"id"`
	Initial       string                  `json:"initial"`
	Rates         []string                `json:"rates"`
	Efficiency    string                  `json:"efficiency"`
	Cap           string                  `json:"cap"`
	End           int64                   `json:"end_ms"`
	Cut           int64                   `json:"cut_ms"`
	ExpectedError bool                    `json:"expected_error"`
	ReferenceWire string                  `json:"reference_wire"`
	One           *anchorResearchSnapshot `json:"one"`
	Split         *anchorResearchSnapshot `json:"split"`
	RebasedWire   string                  `json:"rebased_wire"`
}

func observeResearchAnchor(t *testing.T, id, initial string, rates []string, efficiency, cap string, end, cut int64, expectedError bool, literal string) anchorResearchCase {
	t.Helper()
	row := anchorResearchCase{ID: id, Initial: initial, Rates: rates, Efficiency: efficiency, Cap: cap, End: end, Cut: cut, ExpectedError: expectedError}
	reference, refErr := anchorResearchBuild(initial, rates, efficiency, cap, end)
	one, oneErr := anchorResearchRun(initial, rates, efficiency, cap, []int64{end}, false)
	split, splitErr := anchorResearchRun(initial, rates, efficiency, cap, []int64{cut, end}, false)
	if (refErr != nil) != expectedError || (oneErr != nil) != expectedError || (splitErr != nil) != expectedError {
		t.Fatalf("%s unexpected refusal/admission: ref=%v one=%v split=%v", id, refErr, oneErr, splitErr)
	}
	if expectedError {
		return row
	}
	if literal != "" && reference.Wire != literal {
		t.Fatalf("%s reference=%s literal=%s", id, reference.Wire, literal)
	}
	a, _ := json.Marshal(one)
	b, _ := json.Marshal(split)
	if !bytes.Equal(a, b) || one.Wire != reference.Wire {
		t.Fatalf("%s anchor differs: one=%s split=%s", id, a, b)
	}
	row.ReferenceWire, row.One, row.Split = reference.Wire, &one, &split
	rebased, err := anchorResearchRun(initial, rates, efficiency, cap, []int64{cut, end}, true)
	if err != nil {
		t.Fatalf("%s rebased negative model invalid: %v", id, err)
	}
	row.RebasedWire = rebased.Wire
	return row
}

type anchorResearchBoundary struct {
	ID                string                 `json:"id"`
	Before            anchorResearchSnapshot `json:"before"`
	After             anchorResearchSnapshot `json:"after"`
	PriorReference    string                 `json:"prior_reference"`
	PriorAgrees       bool                   `json:"prior_agrees"`
	RetroactiveBefore string                 `json:"retroactive_before"`
}

func observeAnchorBoundary(t *testing.T, row researchBoundaryRow) anchorResearchBoundary {
	t.Helper()
	before, err := anchorResearchRun(row.Initial, []string{row.RateBefore}, row.Efficiency, row.Cap, []int64{row.CutBefore, row.EndBefore}, false)
	if err != nil {
		t.Fatal(err)
	}
	debit := decimal.FromString(before.Wire).Sub(decimal.FromString(row.Debit)).Quantize(decimal.CanonicalSignificantDigits)
	after, err := anchorResearchRun(debit.String(), []string{row.RateAfter}, row.Efficiency, row.Cap, []int64{row.CutAfter, row.EndAfter}, false)
	if err != nil {
		t.Fatal(err)
	}
	one, err := anchorResearchRun(debit.String(), []string{row.RateAfter}, row.Efficiency, row.Cap, []int64{row.EndAfter}, false)
	if err != nil || one.Wire != after.Wire {
		t.Fatal("boundary second phase differs")
	}
	retroactive, err := anchorResearchBuild(row.Initial, []string{row.RateAfter}, row.Efficiency, row.Cap, row.EndBefore)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(row.ID, "rate-debit/") && retroactive.Wire == before.Wire {
		t.Fatal("retroactive control cannot discriminate")
	}
	if strings.HasPrefix(row.ID, "cap-debit/") && (before.Wire != row.Cap || after.Initial != debit.String() || after.Initial == before.Initial) {
		t.Fatal("cap/debit did not reset visible anchor")
	}
	return anchorResearchBoundary{row.ID, before, after, row.ReferenceAfter.Wire, after.Wire == row.ReferenceAfter.Wire, retroactive.Wire}
}

func anchorResearchNegatives(t *testing.T) map[string][]byte {
	t.Helper()
	base, err := anchorResearchBuild("1e4", []string{"1.15115e0"}, "1e0", "1e12", 3114)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(base)
	mutate := func(change func(*anchorResearchSnapshot)) []byte {
		copy := base
		change(&copy)
		data, _ := json.Marshal(copy)
		return data
	}
	remove := func(part string) []byte {
		result := strings.Replace(string(encoded), part, "", 1)
		if result == string(encoded) {
			t.Fatal("missing-field control did not mutate")
		}
		return []byte(result)
	}
	unknown := append(append([]byte{}, encoded[:len(encoded)-1]...), []byte(`,"extra":0}`)...)
	duplicate := append(append([]byte{}, encoded[:len(encoded)-1]...), []byte(`,"wire":"`+base.Wire+`"}`)...)
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	reordered, _ := json.Marshal(fields)
	if bytes.Equal(reordered, encoded) {
		t.Fatal("field-order control did not mutate")
	}
	return map[string][]byte{
		"missing version":          remove(`"version":1,`),
		"version2":                 mutate(func(s *anchorResearchSnapshot) { s.Version = 2 }),
		"missing initial":          remove(`"initial":"1e4",`),
		"missing rates":            remove(`"rates":["1.15115e0"],`),
		"negative initial":         mutate(func(s *anchorResearchSnapshot) { s.Initial = "-1e0" }),
		"noncanonical initial":     mutate(func(s *anchorResearchSnapshot) { s.Initial = "10000" }),
		"negative efficiency":      mutate(func(s *anchorResearchSnapshot) { s.Efficiency = "-1e0" }),
		"wrong valid wire":         mutate(func(s *anchorResearchSnapshot) { s.Wire = "1e4" }),
		"wrong valid rate":         mutate(func(s *anchorResearchSnapshot) { s.Rates = []string{"2.4048e0"} }),
		"unknown field":            unknown,
		"trailing JSON":            append(append([]byte{}, encoded...), []byte(`{}`)...),
		"duplicate wire":           duplicate,
		"noncanonical field order": reordered,
		"unsafe elapsed":           mutate(func(s *anchorResearchSnapshot) { s.Elapsed = decimal.MaxExactInteger + 1 }),
		"length513":                []byte(string(encoded) + strings.Repeat(" ", 513-len(encoded))),
		"four sources":             mutate(func(s *anchorResearchSnapshot) { s.Rates = []string{"1e0", "1e0", "1e0", "1e0"} }),
	}
}

type anchorCarryDiagnostic struct {
	ID       string `json:"id"`
	Wire     string `json:"wire"`
	Residue  string `json:"residue"`
	Admitted bool   `json:"admitted"`
}
type anchorResearchReport struct {
	Version       int                      `json:"version"`
	Acceptance    string                   `json:"acceptance_status"`
	Sources       map[string]string        `json:"source_sha256"`
	Cases         []anchorResearchCase     `json:"cases"`
	Boundaries    []anchorResearchBoundary `json:"boundaries"`
	NearCapAfter  []anchorResearchSnapshot `json:"near_cap_after_debit"`
	Negatives     map[string]string        `json:"restore_negatives"`
	Carry         []anchorCarryDiagnostic  `json:"go_only_carry_diagnostics"`
	RebaseHits    int                      `json:"ordinary_rebase_hits"`
	DomainRefused int                      `json:"domain_refused"`
	GoldenRefused int                      `json:"golden_refused"`
	PriorDiffs    int                      `json:"prior_boundary_differences"`
	MaxBytes      int                      `json:"max_snapshot_bytes"`
}

const anchorResearchPath = "../../testdata/axis-stack/anchor-research-v1.json"

func TestAxisAnchorResearch(t *testing.T) {
	var previous researchReport
	data, err := os.ReadFile(partitionResearchPath)
	if err != nil || json.Unmarshal(data, &previous) != nil || len(previous.Frozen) != 520 || len(previous.Boundaries) != 12 {
		t.Fatal("first-wave population unavailable")
	}
	report := anchorResearchReport{Version: 1, Acceptance: "NOT_PROVEN: production AC6 remains red", Sources: map[string]string{}, Negatives: map[string]string{}}
	for _, path := range []string{"server/production/axis_anchor_research_test.go", "server/production/axis_partition_research_test.go", "client/test/anchor-research.test.ts", "testdata/axis-stack/partition-research-v1.json", "testdata/production-accrual.json", "server/production/accrual.go", "server/decimal/decimal.go", "server/decimal/canonical.go", "client/src/production.ts", "client/src/numeric.ts"} {
		bytes, err := os.ReadFile("../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(bytes)
		report.Sources[path] = hex.EncodeToString(sum[:])
	}
	for _, row := range previous.Frozen {
		result := observeResearchAnchor(t, row.ID, row.Initial, []string{row.Rate}, row.Efficiency, row.Cap, row.EndMS, row.CutMS, false, row.CurrentOne)
		if result.RebasedWire != result.ReferenceWire {
			report.RebaseHits++
		}
		report.Cases = append(report.Cases, result)
	}
	maxExponent := int64(8_999_999_999_999_999)
	maxCap := "1e" + strconv.FormatInt(maxExponent, 10)
	for _, exponent := range []int64{-maxExponent, -1000000, -33, 0, 33, 1000000, maxExponent - 14, maxExponent - 1} {
		e := strconv.FormatInt(exponent, 10)
		for set, rates := range [][]string{{"1.15115e" + e}, {"1.15115e" + e, "2e" + e}} {
			for _, eff := range []string{"1e0", "9e-1"} {
				for _, span := range [][2]int64{{2000, 1000}, {decimal.MaxExactInteger, decimal.MaxExactInteger / 2}} {
					refused := exponent == maxExponent-1 && span[0] == decimal.MaxExactInteger
					id := fmt.Sprintf("domain/%s/%d/%s/%d", e, set, eff, span[0])
					result := observeResearchAnchor(t, id, "1e"+e, rates, eff, maxCap, span[0], span[1], refused, "")
					if refused {
						report.DomainRefused++
					}
					report.Cases = append(report.Cases, result)
				}
			}
		}
	}
	var goldens struct {
		Vectors []struct {
			Name       string   `json:"name"`
			Rates      []string `json:"rates"`
			Elapsed    int64    `json:"elapsed_ms"`
			Efficiency string   `json:"efficiency"`
			Expect     string   `json:"expect"`
			Error      bool     `json:"expect_error"`
		}
	}
	data, err = os.ReadFile("../../testdata/production-accrual.json")
	if err != nil || json.Unmarshal(data, &goldens) != nil || len(goldens.Vectors) != 16 {
		t.Fatal("primitive goldens unavailable")
	}
	for _, row := range goldens.Vectors {
		result := observeResearchAnchor(t, "golden/"+row.Name, "0", row.Rates, row.Efficiency, maxCap, row.Elapsed, row.Elapsed/2, row.Error, row.Expect)
		if row.Error {
			report.GoldenRefused++
		}
		report.Cases = append(report.Cases, result)
	}
	for _, row := range previous.Boundaries {
		result := observeAnchorBoundary(t, row)
		if !result.PriorAgrees {
			report.PriorDiffs++
		}
		report.Boundaries = append(report.Boundaries, result)
	}
	for _, rate := range []string{"4e-9", "5e-9", "6e-9"} {
		row := observeResearchAnchor(t, "near-cap/"+rate, "9.99999999999e3", []string{rate}, "1e0", "1e4", 2000, 1000, false, "")
		report.Cases = append(report.Cases, row)
		debit := decimal.FromString(row.One.Wire).Sub(decimal.FromString("1.13e1")).Quantize(decimal.CanonicalSignificantDigits)
		after, err := anchorResearchRun(debit.String(), []string{rate}, "1e0", "1e4", []int64{500, 1000}, false)
		if err != nil {
			t.Fatal(err)
		}
		report.NearCapAfter = append(report.NearCapAfter, after)
	}
	for id, encoded := range anchorResearchNegatives(t) {
		if _, err := anchorResearchRestore(encoded); err == nil {
			t.Fatalf("restore control %q admitted", id)
		}
		report.Negatives[id] = string(encoded)
	}
	cap := mustResearchRat(t, "1e12")
	for _, entry := range []struct{ id, raw, wire, residue string }{{"inconsistent projection", `{"wire":"0","residue":"1"}`, "0", "1"}, {"trailing JSON", `{"wire":"1e4","residue":"0"}{}`, "1e4", "0"}} {
		_, err := researchRestore([]byte(entry.raw), cap)
		report.Carry = append(report.Carry, anchorCarryDiagnostic{entry.id, entry.wire, entry.residue, err == nil})
	}
	near, err := researchCommit(mustResearchRat(t, "9999.999999998"), mustResearchRat(t, "1e4"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(near)
	_, err = researchRestore(encoded, mustResearchRat(t, "1e4"))
	report.Carry = append(report.Carry, anchorCarryDiagnostic{"below cap rounded to cap", near.Wire, near.Residue, err == nil})
	measure := func(snapshots ...*anchorResearchSnapshot) {
		for _, snapshot := range snapshots {
			if snapshot != nil {
				encoded, _ := json.Marshal(snapshot)
				report.MaxBytes = max(report.MaxBytes, len(encoded))
			}
		}
	}
	for _, row := range report.Cases {
		measure(row.One, row.Split)
	}
	for _, row := range report.Boundaries {
		measure(&row.Before, &row.After)
	}
	for _, row := range report.NearCapAfter {
		measure(&row)
	}
	if len(report.Cases) != 603 || len(report.Boundaries) != 12 || len(report.NearCapAfter) != 3 || len(report.Negatives) != 16 || len(report.Carry) != 3 || report.RebaseHits < 45 || report.DomainRefused != 4 || report.GoldenRefused != 5 || report.MaxBytes > 512 {
		t.Fatalf("incomplete/invalid anchor observation: cases%d boundary%d near%d refusal%d carry%d rebase%d domain%d golden%d size%d", len(report.Cases), len(report.Boundaries), len(report.NearCapAfter), len(report.Negatives), len(report.Carry), report.RebaseHits, report.DomainRefused, report.GoldenRefused, report.MaxBytes)
	}
	encoded, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("UPDATE_ANCHOR_RESEARCH") == "1" {
		if err := os.WriteFile(anchorResearchPath, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	committed, err := os.ReadFile(anchorResearchPath)
	if err != nil || !bytes.Equal(encoded, committed) {
		t.Fatalf("anchor source/results drifted; explicit re-observation required: %v", err)
	}
	t.Logf("R-012 anchor observation:615 primary/16 restore refusals; rebase%d, domain refused4, goldens refused5, prior boundary differences%d, max JSON%d; AC6 NOT_PROVEN", report.RebaseHits, report.PriorDiffs, report.MaxBytes)
}
