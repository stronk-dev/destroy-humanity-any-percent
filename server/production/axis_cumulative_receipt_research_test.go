package production

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Research objects only, not a proposed wire codec or production mutation API.
// Receipt is cumulative FROM THE ORIGIN; its Delta is deliberately not claimed
// to add to the preceding visible projection. That distinction needs adoption.
type cumulativeReceiptResearch struct {
	OriginSHA256  string
	ConstantsHash string
	BeforeSHA256  string
	AfterSHA256   string
	Mode          EvaluationMode
	Through       time.Time
	Evaluation    EvaluationResult
	Company       []byte
}

func cumulativeResearchHash(data []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(data))
}

func cumulativeResearchProjection(bundle CatalogBundle, origin []byte, through time.Time, mode EvaluationMode) (*save.State, EvaluationResult, error) {
	state, err := save.RestoreState(origin, 19, bundle.Economy, economy.ScopeCompany, through)
	if err != nil {
		return nil, EvaluationResult{}, err
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		return nil, EvaluationResult{}, err
	}
	contributions, err := assembleContributions(state, bundle.Economy, nil)
	if err != nil {
		return nil, EvaluationResult{}, err
	}
	result, err := Evaluate(state, bundle.Economy, through, mode, contributions)
	return state, result, err
}

func cumulativeResearchProof(t *testing.T, bundle CatalogBundle, origin []byte, before *save.State, through time.Time, mode EvaluationMode) cumulativeReceiptResearch {
	t.Helper()
	projected, result, err := cumulativeResearchProjection(bundle, origin, through, mode)
	if err != nil {
		t.Fatal(err)
	}
	encoded := mustEncodeState(t, projected)
	return cumulativeReceiptResearch{
		OriginSHA256: cumulativeResearchHash(origin), ConstantsHash: bundle.ConstantsHash,
		BeforeSHA256: cumulativeResearchHash(mustEncodeState(t, before)),
		AfterSHA256:  cumulativeResearchHash(encoded), Mode: mode, Through: through,
		Evaluation: result, Company: encoded,
	}
}

// Return a candidate only after checking the complete prior head, pinned origin
// and complete next state/result. Never install it in the caller's live state.
// Hashes identify the trusted inputs; they are not signatures or proof of how
// those inputs were obtained. This is expressiveness, not Ledger.ApplyAccrual.
func cumulativeResearchVerify(t *testing.T, bundle CatalogBundle, origin []byte, before *save.State, through time.Time, mode EvaluationMode, proof cumulativeReceiptResearch) (*save.State, error) {
	t.Helper()
	invalid := errors.New("invalid cumulative research proof")
	if proof.OriginSHA256 != cumulativeResearchHash(origin) || proof.ConstantsHash != bundle.ConstantsHash ||
		proof.BeforeSHA256 != cumulativeResearchHash(mustEncodeState(t, before)) ||
		proof.Mode != mode || !proof.Through.Equal(through) || !through.After(before.EvaluatedThrough) ||
		proof.AfterSHA256 != cumulativeResearchHash(proof.Company) {
		return nil, invalid
	}
	previous, _, err := cumulativeResearchProjection(bundle, origin, before.EvaluatedThrough, mode)
	if err != nil || !bytes.Equal(mustEncodeState(t, previous), mustEncodeState(t, before)) {
		return nil, invalid // A rehashed head after a debit cannot reuse this origin.
	}
	projected, result, err := cumulativeResearchProjection(bundle, origin, through, mode)
	if err != nil || !bytes.Equal(mustEncodeState(t, projected), proof.Company) || !reflect.DeepEqual(result, proof.Evaluation) {
		return nil, invalid
	}
	return projected, nil
}

func assertCumulativeResearchReceipt(t *testing.T, origin, projected *save.State, result EvaluationResult) {
	t.Helper()
	before, after := origin.Ledger.Snapshot(), projected.Ledger.Snapshot()
	seen := map[string]bool{}
	for _, change := range result.Receipt.Changes {
		if seen[change.ResourceID] || change.Before != before[change.ResourceID] || change.After != after[change.ResourceID] ||
			mustDecimal(t, change.Delta).Lt(decimal.Zero) ||
			!mustDecimal(t, change.Before).Add(mustDecimal(t, change.Delta)).Quantize(decimal.CanonicalSignificantDigits).Eq(mustDecimal(t, change.After)) {
			t.Fatalf("origin-relative receipt does not reproduce the exact target: %+v", change)
		}
		seen[change.ResourceID] = true
	}
	for id, value := range before {
		if (value != after[id]) != seen[id] {
			t.Fatalf("cumulative receipt misses or invents resource change %s", id)
		}
	}
}

func TestAxisCumulativeReceiptLowerDomainResearch(t *testing.T) {
	cells, additiveRefusals := 0, 0
	for _, exponent := range []int64{-8999999999999999, -8999999999999998, -8999999999999996, -8999999999999987, 0} {
		bundle, initial := originLowerDomainFixture(t, exponent)
		origin := mustEncodeState(t, initial)
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			for _, cut := range []int64{1000, 2000} {
				t.Run(fmt.Sprintf("exponent%d_%s_cut%d", exponent, mode, cut), func(t *testing.T) {
					cells++
					before := originProjectionState(t, bundle, origin, initial.EvaluatedThrough.Add(time.Duration(cut)*time.Millisecond), mode)
					through := before.EvaluatedThrough.Add(time.Millisecond)
					beforeBytes := mustEncodeState(t, before)
					proof := cumulativeResearchProof(t, bundle, origin, before, through, mode)
					got, err := cumulativeResearchVerify(t, bundle, origin, before, through, mode, proof)
					if err != nil || got == nil {
						t.Fatalf("cumulative proof refused: %v", err)
					}
					want := originProjectionState(t, bundle, origin, through, mode)
					if !bytes.Equal(mustEncodeState(t, got), mustEncodeState(t, want)) ||
						!bytes.Equal(beforeBytes, mustEncodeState(t, before)) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatal("candidate differs from one-shot state or mutated its inputs")
					}
					assertCumulativeResearchReceipt(t, initial, got, proof.Evaluation)
					// Preserve the old contract's limitation rather than relabel the
					// cumulative receipt as a valid previous-head additive receipt.
					delta := mustDecimal(t, got.Ledger.Snapshot()["company.cash"]).Sub(mustDecimal(t, before.Ledger.Snapshot()["company.cash"])).Quantize(decimal.CanonicalSignificantDigits)
					if !delta.IsStateValue() {
						additiveRefusals++
						if _, err := before.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: delta}}}); !errors.Is(err, economy.ErrInvalidTransaction) ||
							!bytes.Equal(beforeBytes, mustEncodeState(t, before)) {
							t.Fatalf("old additive refusal/rollback changed: %v", err)
						}
					}
				})
			}
		}
	}
	if cells != 20 || additiveRefusals != 9 {
		t.Fatalf("incomplete cumulative experiment: cells=%d additive-refusals=%d", cells, additiveRefusals)
	}
	t.Log("20 origin-relative proofs express exact targets; nine previous-head additive deltas remain inadmissible; ledger mutation/AC6 NOT_PROVEN")
}

func TestAxisCumulativeReceiptRejectsForgeriesResearch(t *testing.T) {
	bundle, initial := originLowerDomainFixture(t, -8999999999999999)
	origin := mustEncodeState(t, initial)
	before := originProjectionState(t, bundle, origin, initial.EvaluatedThrough.Add(time.Second), ModeOnline)
	through := before.EvaluatedThrough.Add(time.Millisecond)
	beforeBytes := mustEncodeState(t, before)
	for name, alter := range map[string]func(*cumulativeReceiptResearch){
		"origin":        func(p *cumulativeReceiptResearch) { p.OriginSHA256 = cumulativeResearchHash(beforeBytes) },
		"catalog":       func(p *cumulativeReceiptResearch) { p.ConstantsHash = cumulativeResearchHash(nil) },
		"previous head": func(p *cumulativeReceiptResearch) { p.BeforeSHA256 = cumulativeResearchHash(origin) },
		"mode":          func(p *cumulativeReceiptResearch) { p.Mode = ModeOffline },
		"time":          func(p *cumulativeReceiptResearch) { p.Through = p.Through.Add(time.Millisecond) },
		"after hash":    func(p *cumulativeReceiptResearch) { p.AfterSHA256 = cumulativeResearchHash(nil) },
		"rehashed other projection": func(p *cumulativeReceiptResearch) {
			p.Company = bytes.Clone(beforeBytes)
			p.AfterSHA256 = cumulativeResearchHash(p.Company)
		},
		"omitted receipt": func(p *cumulativeReceiptResearch) { p.Evaluation.Receipt.Changes = nil },
		"delta":           func(p *cumulativeReceiptResearch) { p.Evaluation.Receipt.Changes[0].Delta = "0" },
		"receipt origin": func(p *cumulativeReceiptResearch) {
			p.Evaluation.Receipt.Changes[0].Before = before.Ledger.Snapshot()["company.cash"]
		},
		"elapsed":    func(p *cumulativeReceiptResearch) { p.Evaluation.ElapsedMS-- },
		"production": func(p *cumulativeReceiptResearch) { p.Evaluation.ProductionMS-- },
		"bank":       func(p *cumulativeReceiptResearch) { p.Evaluation.BankedCreditMS++ },
		"progress":   func(p *cumulativeReceiptResearch) { p.Evaluation.ProgressDeltaPPM++ },
	} {
		t.Run(name, func(t *testing.T) {
			proof := cumulativeResearchProof(t, bundle, origin, before, through, ModeOnline)
			alter(&proof)
			if got, err := cumulativeResearchVerify(t, bundle, origin, before, through, ModeOnline, proof); err == nil || got != nil {
				t.Fatal("forged proof returned a candidate")
			}
			if !bytes.Equal(beforeBytes, mustEncodeState(t, before)) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
				t.Fatal("forged proof changed the previous head or origin")
			}
		})
	}
	proof := cumulativeResearchProof(t, bundle, origin, before, through, ModeOnline)
	if got, err := cumulativeResearchVerify(t, bundle, origin, before, before.EvaluatedThrough, ModeOnline, proof); err == nil || got != nil {
		t.Fatal("proof with no forward interval returned a candidate")
	}
}

func TestAxisCumulativeReceiptCapDebitReanchorsResearch(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e4")
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		t.Run(string(mode), func(t *testing.T) {
			initial := axisTimingState(t, bundle, start)
			origin := mustEncodeState(t, initial)
			first := originProjectionState(t, bundle, origin, start.Add(1553*time.Millisecond), mode)
			through := start.Add(3114 * time.Millisecond)
			proof := cumulativeResearchProof(t, bundle, origin, first, through, mode)
			capped, err := cumulativeResearchVerify(t, bundle, origin, first, through, mode, proof)
			if err != nil || capped.Ledger.Snapshot()["company.cash"] != "1e4" {
				t.Fatalf("capped candidate: %v", err)
			}
			assertCumulativeResearchReceipt(t, initial, capped, proof.Evaluation)
			if _, err := capped.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: mustDecimal(t, "-1.13e1")}}}); err != nil {
				t.Fatal(err)
			}
			afterDebit := mustEncodeState(t, capped)
			next := through.Add(time.Second)
			stale := cumulativeResearchProof(t, bundle, origin, capped, next, mode)
			// Even a freshly rehashed BeforeSHA256 cannot hide the debit from
			// the full previous-head reconstruction against the old origin.
			if got, err := cumulativeResearchVerify(t, bundle, origin, capped, next, mode, stale); err == nil || got != nil {
				t.Fatal("stale origin reclaimed discarded cap overflow after debit")
			}
			if !bytes.Equal(afterDebit, mustEncodeState(t, capped)) {
				t.Fatal("stale-origin refusal mutated the debited state")
			}
			fresh := cumulativeResearchProof(t, bundle, afterDebit, capped, next, mode)
			got, err := cumulativeResearchVerify(t, bundle, afterDebit, capped, next, mode, fresh)
			if err != nil || got == nil {
				t.Fatalf("new post-debit origin refused: %v", err)
			}
			want := "9.98985115e3"
			if mode == ModeOffline {
				want = "9.989736035e3"
			}
			if got.Ledger.Snapshot()["company.cash"] != want {
				t.Fatalf("discarded overflow became spendable: got %s want %s", got.Ledger.Snapshot()["company.cash"], want)
			}
			assertCumulativeResearchReceipt(t, capped, got, fresh.Evaluation)
		})
	}
}
