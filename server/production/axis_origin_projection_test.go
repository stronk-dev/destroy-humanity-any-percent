package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Test-only whole-Company origin. This deliberately uses the existing codec,
// not a proposed production accrual schema. No production caller uses it.
func originProjectionState(t *testing.T, bundle CatalogBundle, origin []byte, now time.Time, mode EvaluationMode) *save.State {
	t.Helper()
	state, err := save.RestoreState(origin, 19, bundle.Economy, economy.ScopeCompany, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatal(err)
	}
	contributions, err := assembleContributions(state, bundle.Economy, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(state, bundle.Economy, now, mode, contributions); err != nil {
		t.Fatal(err)
	}
	return state
}

// Advance through the real ledger rather than installing a projected ledger.
// The remaining assignments are precisely Evaluate's non-ledger write set;
// full encoded equality below catches an incomplete write-set assumption.
func originProjectionAdvance(t *testing.T, visible, projected *save.State) {
	t.Helper()
	before := visible.Ledger.Snapshot()
	want := projected.Ledger.Snapshot()
	ids := make([]string, 0, len(want))
	for id := range want {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	entries := make([]economy.Entry, 0, len(ids))
	for _, id := range ids {
		old := mustDecimal(t, before[id])
		next := mustDecimal(t, want[id])
		if next.Lt(old) {
			t.Fatalf("origin projection decreased %s: %s -> %s", id, old, next)
		}
		if !next.Eq(old) {
			entries = append(entries, economy.Entry{ResourceID: id, Delta: next.Sub(old).Quantize(decimal.CanonicalSignificantDigits)})
		}
	}
	receipt, err := visible.Ledger.ApplyAccrual(economy.Transaction{Entries: entries})
	if err != nil {
		t.Fatalf("projected balance cannot be committed through ledger: %v", err)
	}
	for _, change := range receipt.Changes {
		if change.Before != before[change.ResourceID] || change.After != want[change.ResourceID] ||
			!mustDecimal(t, change.Before).Add(mustDecimal(t, change.Delta)).Quantize(decimal.CanonicalSignificantDigits).Eq(mustDecimal(t, change.After)) {
			t.Fatalf("projection receipt does not bind actual before/target: %+v", change)
		}
	}
	for _, id := range ids {
		if got := visible.Ledger.Snapshot()[id]; got != want[id] {
			t.Fatalf("ledger projection missed %s: got %s want %s", id, got, want[id])
		}
	}
	visible.ComputeCreditMS = projected.ComputeCreditMS
	visible.ComputeBurstRemainingMS = projected.ComputeBurstRemainingMS
	visible.GeneratorProvisioned = cloneInt64Counts(projected.GeneratorProvisioned)
	visible.ProvisionRemaindersPPM = cloneInt64Counts(projected.ProvisionRemaindersPPM)
	visible.EvaluatedThrough = projected.EvaluatedThrough
	if !bytes.Equal(mustEncodeState(t, visible), mustEncodeState(t, projected)) {
		t.Fatal("ledger advance and Evaluate write-set do not reproduce full projected Company")
	}
}

func originProjectionCuts(t *testing.T, bundle CatalogBundle, initial *save.State, mode EvaluationMode, offsets []int64) *save.State {
	t.Helper()
	origin := mustEncodeState(t, initial)
	frozen := bytes.Clone(origin)
	start := initial.EvaluatedThrough
	visible := originProjectionState(t, bundle, origin, start, mode)
	for _, offset := range offsets {
		projected := originProjectionState(t, bundle, origin, start.Add(time.Duration(offset)*time.Millisecond), mode)
		originProjectionAdvance(t, visible, projected)
		// Restore the visible full state at every cut. The origin remains distinct.
		encoded := mustEncodeState(t, visible)
		var err error
		visible, err = save.RestoreState(encoded, 19, bundle.Economy, economy.ScopeCompany, start)
		if err != nil || !bytes.Equal(encoded, mustEncodeState(t, visible)) {
			t.Fatalf("visible state restoration changed projection: %v", err)
		}
		if err := bundle.ValidateFoundationState(visible); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(origin, frozen) {
		t.Fatal("evaluation changed the retained origin")
	}
	return visible
}

func TestAxisOriginProjectionOriginalCuts(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(307))
	rebasedDifferences := 0
	for policy := 0; policy < 64; policy++ {
		end := int64(2000 + random.Intn(2000))
		cut := int64(1 + random.Intn(int(end-1)))
		if policy == 0 {
			end, cut = 2000, 1000
		}
		if policy == 1 {
			end, cut = 2001, 1
		}
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(fmt.Sprintf("policy_%02d_%s_end%d_cut%d", policy, mode, end, cut), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				origin := mustEncodeState(t, initial)
				one := originProjectionState(t, bundle, origin, start.Add(time.Duration(end)*time.Millisecond), mode)
				projected := originProjectionCuts(t, bundle, initial, mode, []int64{cut, end})
				if !bytes.Equal(mustEncodeState(t, one), mustEncodeState(t, projected)) {
					t.Fatal("retained-origin projection changed original one-shot result")
				}
				rebased := axisTimingState(t, bundle, start)
				contributions, err := assembleContributions(rebased, bundle.Economy, nil)
				if err != nil {
					t.Fatal(err)
				}
				for _, at := range []int64{cut, end} {
					if _, err := Evaluate(rebased, bundle.Economy, start.Add(time.Duration(at)*time.Millisecond), mode, contributions); err != nil {
						t.Fatal(err)
					}
				}
				if !bytes.Equal(mustEncodeState(t, one), mustEncodeState(t, rebased)) {
					rebasedDifferences++
				}
			})
		}
	}
	if rebasedDifferences != 27 {
		t.Fatalf("original population's rebasing control: got %d differences, want 27", rebasedDifferences)
	}
	t.Log("128 original profiles project exactly through the ledger; ordinary rebasing differs in 27")
}

func TestAxisOriginProjectionPolicyIntersections(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		for _, provider := range []int64{0, 1, 10} {
			for _, burst := range []int64{0, 1500} {
				t.Run(fmt.Sprintf("%s_provider%d_burst%d", mode, provider, burst), func(t *testing.T) {
					initial := axisTimingState(t, bundle, start)
					initial.GeneratorCounts["generator.beige_tower_v2"] = provider
					initial.GeneratorCounts["generator.legal_dept"] = 1
					initial.GeneratorPurchasedTotal = 2 + provider
					initial.ComputeBurstRemainingMS = burst
					visible := originProjectionCuts(t, bundle, initial, mode, []int64{1, 59999, 60000, 60001, 120000, 120001})
					permits, ok := visible.Ledger.Balance("company.permits")
					if !ok || !permits.Gt(decimal.Zero) || visible.ComputeBurstRemainingMS != 0 ||
						visible.GeneratorProvisioned["generator.beige_tower"] != provider/5 ||
						visible.ProvisionRemaindersPPM["generator.beige_tower"] != provider*200000%1000000 {
						t.Fatal("declared permits/burst/provision intersection did not execute")
					}
				})
			}
		}
	}
}

func TestAxisOriginProjectionSingleOfflineEpisode(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	policy := bundle.Economy.OfflinePolicy()
	for _, end := range []int64{policy.AccrualCapMS + 2, 2 * policy.AccrualCapMS} {
		for _, credit := range []int64{0, policy.BankCapMS - 1} {
			t.Run(fmt.Sprintf("end%d_credit%d", end, credit), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				initial.ComputeCreditMS = credit
				visible := originProjectionCuts(t, bundle, initial, ModeOffline, []int64{1, policy.AccrualCapMS, policy.AccrualCapMS + 1, end})
				wantBank := min((end-policy.AccrualCapMS)/2, policy.BankCapMS-credit)
				if visible.ComputeCreditMS != credit+wantBank {
					t.Fatal("single-episode bank floor or saturation changed")
				}
			})
		}
	}
}

func TestAxisOriginProjectionCapDebitResetsOrigin(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e4")
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		t.Run(string(mode), func(t *testing.T) {
			initial := axisTimingState(t, bundle, start)
			visible := originProjectionCuts(t, bundle, initial, mode, []int64{1553, 3114})
			if visible.Ledger.Snapshot()["company.cash"] != "1e4" {
				t.Fatal("initial cap was not saturated")
			}
			if _, err := visible.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: mustDecimal(t, "-1.13e1")}}}); err != nil {
				t.Fatal(err)
			}
			// This is a test debit/re-anchor, not an ApplyLogged integration claim.
			after := originProjectionCuts(t, bundle, visible, mode, []int64{1, 1000})
			want := "9.98985115e3"
			if mode == ModeOffline {
				want = "9.989736035e3"
			}
			if got := after.Ledger.Snapshot()["company.cash"]; got != want {
				t.Fatalf("cap overflow became spendable after debit: got %s want %s", got, want)
			}
		})
	}
}

// Representation research only: an admitted one-shot target is not proof that
// subtracting the previous projection yields an admitted incremental ledger delta.
func TestAxisOriginProjectionLowerDomainResearch(t *testing.T) {
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	cells, unchanged, invalid, refused, exact, wrong := 0, 0, 0, 0, 0, 0
	for _, exponent := range []int64{-8999999999999999, -8999999999999998, -8999999999999996, -8999999999999987, 0} {
		bundle := axisContentBundle(t)
		bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
		var root map[string]any
		if err := json.Unmarshal(bundle.Artifacts["economy"], &root); err != nil {
			t.Fatal(err)
		}
		rate := fmt.Sprintf("1e%d", exponent)
		root["generator_classes"].([]any)[0].(map[string]any)["production"].(map[string]any)["base_rate"] = rate
		var err error
		bundle.Artifacts["economy"], err = json.Marshal(root)
		if err != nil {
			t.Fatal(err)
		}
		bundle.Economy, err = economy.LoadCatalog(bundle.Artifacts["economy"])
		if err != nil {
			t.Fatal(err)
		}
		bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
		if err != nil || !bundle.valid(bundle.ConstantsHash) {
			t.Fatalf("research catalog admission: %v", err)
		}
		initial := axisTimingState(t, bundle, start)
		delete(initial.UpgradesOwned, "upgrade.pr_intern_1")
		setCash(t, initial, "0")
		if err := bundle.ValidateFoundationState(initial); err != nil {
			t.Fatal(err)
		}
		origin := mustEncodeState(t, initial)
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			for _, cut := range []int64{1000, 2000} {
				cells++
				visible := originProjectionState(t, bundle, origin, start.Add(time.Duration(cut)*time.Millisecond), mode)
				projected := originProjectionState(t, bundle, origin, start.Add(time.Duration(cut+1)*time.Millisecond), mode)
				beforeWire, targetWire := visible.Ledger.Snapshot()["company.cash"], projected.Ledger.Snapshot()["company.cash"]
				if exponent == -8999999999999999 && mode == ModeOnline && cut == 1000 &&
					(beforeWire != "1.001e-8999999999999999" || targetWire != "1.002001e-8999999999999999") {
					t.Fatal("retained lower-domain counterexample changed")
				}
				before, target := mustDecimal(t, beforeWire), mustDecimal(t, targetWire)
				if target.Lt(before) {
					t.Fatal("one-shot projection decreased")
				}
				if target.Eq(before) {
					unchanged++
					t.Logf("rate=%s mode=%s cut=%d before=%s target=%s classification=unchanged", rate, mode, cut, beforeWire, targetWire)
					continue
				}
				delta := target.Sub(before).Quantize(decimal.CanonicalSignificantDigits)
				if !delta.IsStateValue() {
					invalid++
				}
				beforeState := mustEncodeState(t, visible)
				receipt, commitErr := visible.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: delta}}})
				classification := "exact"
				if commitErr != nil {
					refused++
					classification = "refused"
					if delta.IsStateValue() || !errors.Is(commitErr, economy.ErrInvalidTransaction) {
						t.Fatalf("refusal is not the measured non-state-delta case: %v", commitErr)
					}
					if !bytes.Equal(beforeState, mustEncodeState(t, visible)) || len(receipt.Changes) != 0 {
						t.Fatal("refused experiment mutated the visible state")
					}
				} else if visible.Ledger.Snapshot()["company.cash"] != targetWire {
					wrong++
					classification = "wrong-target"
				} else {
					exact++
					if len(receipt.Changes) != 1 || receipt.Changes[0].Before != beforeWire || receipt.Changes[0].After != targetWire ||
						!before.Add(mustDecimal(t, receipt.Changes[0].Delta)).Quantize(decimal.CanonicalSignificantDigits).Eq(target) {
						t.Fatal("successful experiment's receipt does not reproduce its target")
					}
				}
				t.Logf("rate=%s mode=%s cut=%d before=%s target=%s delta=%s classification=%s error=%v", rate, mode, cut, beforeWire, targetWire, delta, classification, commitErr)
			}
		}
	}
	if cells != 20 || unchanged+refused+exact+wrong != cells {
		t.Fatal("incomplete lower-domain research census")
	}
	// Retain both the negative finding and exact positive controls. Refusing all
	// transactions or silently committing a wrong target must not pass this census.
	if unchanged != 1 || invalid != 9 || refused != 9 || exact != 10 || wrong != 0 {
		t.Fatalf("lower-domain observation changed: unchanged=%d invalid=%d refused=%d exact=%d wrong=%d", unchanged, invalid, refused, exact, wrong)
	}
	t.Logf("cells=%d unchanged=%d invalid-delta=%d refused=%d exact=%d wrong-target=%d; production conservation NOT_PROVEN", cells, unchanged, invalid, refused, exact, wrong)
}

// The numeric domain has a minimum positive value, not merely a subtraction
// implementation that could be replaced. This probes the actual ledger on the
// retained pair; no production delta or domain rule is altered.
func TestAxisOriginFloorIncrementBoundaryResearch(t *testing.T) {
	bundle, initial := originLowerDomainFixture(t, -8999999999999999)
	origin := mustEncodeState(t, initial)
	start := initial.EvaluatedThrough
	first := originProjectionState(t, bundle, origin, start.Add(time.Second), ModeOnline)
	last := originProjectionState(t, bundle, origin, start.Add(1001*time.Millisecond), ModeOnline)
	before, target := first.Ledger.Snapshot()["company.cash"], last.Ledger.Snapshot()["company.cash"]
	if before != "1.001e-8999999999999999" || target != "1.002001e-8999999999999999" {
		t.Fatal("retained floor pair changed")
	}
	minimum := mustDecimal(t, "1e-8999999999999999")
	if decimal.New(0.99999999999, minimum.Exponent()).IsStateValue() ||
		decimal.New(1, minimum.Exponent()-1).IsStateValue() {
		t.Fatal("the declared domain no longer has the retained minimum positive value")
	}
	for _, row := range []struct{ delta, after string }{
		{"0", before},
		{"1e-8999999999999999", "2.001e-8999999999999999"},
		{"1.00000000001e-8999999999999999", "2.00100000001e-8999999999999999"},
		{"9.99999999999e-8999999999999999", "1.1001e-8999999999999998"},
		{"1e-8999999999999998", "1.1001e-8999999999999998"},
	} {
		t.Run(row.delta, func(t *testing.T) {
			state := originProjectionState(t, bundle, origin, start.Add(time.Second), ModeOnline)
			delta := mustDecimal(t, row.delta)
			receipt, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: delta}}})
			if err != nil {
				t.Fatal(err)
			}
			if got := state.Ledger.Snapshot()["company.cash"]; got != row.after || got == target {
				t.Fatalf("floor increment result: got %s want %s (target %s)", got, row.after, target)
			}
			if row.delta == "0" {
				if len(receipt.Changes) != 0 || !bytes.Equal(mustEncodeState(t, state), mustEncodeState(t, first)) {
					t.Fatal("zero increment changed state or created a receipt")
				}
			} else if len(receipt.Changes) != 1 || receipt.Changes[0].Before != before || receipt.Changes[0].After != row.after ||
				!mustDecimal(t, before).Add(mustDecimal(t, receipt.Changes[0].Delta)).Quantize(decimal.CanonicalSignificantDigits).Eq(mustDecimal(t, row.after)) {
				t.Fatal("actual increment receipt cannot reproduce its literal result")
			}
		})
	}
	t.Log("floor pair: zero stays below target; minimum positive accrual overshoots; no intermediate-cut policy is adopted")
}

func originLowerDomainFixture(t *testing.T, exponent int64) (CatalogBundle, *save.State) {
	t.Helper()
	bundle := axisContentBundle(t)
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	var root map[string]any
	if err := json.Unmarshal(bundle.Artifacts["economy"], &root); err != nil {
		t.Fatal(err)
	}
	root["generator_classes"].([]any)[0].(map[string]any)["production"].(map[string]any)["base_rate"] = fmt.Sprintf("1e%d", exponent)
	var err error
	bundle.Artifacts["economy"], err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Economy, err = economy.LoadCatalog(bundle.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil || !bundle.valid(bundle.ConstantsHash) {
		t.Fatalf("research catalog admission: %v", err)
	}
	initial := axisTimingState(t, bundle, time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC))
	delete(initial.UpgradesOwned, "upgrade.pr_intern_1")
	setCash(t, initial, "0")
	if err := bundle.ValidateFoundationState(initial); err != nil {
		t.Fatal(err)
	}
	return bundle, initial
}

func TestAxisOriginDeferredLowerDomainSettlementResearch(t *testing.T) {
	cells := 0
	for _, exponent := range []int64{-8999999999999999, -8999999999999998, -8999999999999996, -8999999999999987, 0} {
		bundle, initial := originLowerDomainFixture(t, exponent)
		origin := mustEncodeState(t, initial)
		start := initial.EvaluatedThrough
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			for _, cut := range []int64{1000, 2000} {
				t.Run(fmt.Sprintf("exponent%d_%s_cut%d", exponent, mode, cut), func(t *testing.T) {
					cells++
					// Work cuts are scratch projections, not extra saved heads.
					for _, offset := range []int64{cut - 1, cut} {
						projection := originProjectionState(t, bundle, origin, start.Add(time.Duration(offset)*time.Millisecond), mode)
						if err := bundle.ValidateFoundationState(projection); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(origin, mustEncodeState(t, initial)) {
							t.Fatal("scratch projection mutated the committed origin")
						}
					}
					projected := originProjectionState(t, bundle, origin, start.Add(time.Duration(cut+1)*time.Millisecond), mode)
					settled, err := save.RestoreState(origin, 19, bundle.Economy, economy.ScopeCompany, start)
					if err != nil {
						t.Fatal(err)
					}
					originProjectionAdvance(t, settled, projected)
					if !bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatal("settlement mutated the separate origin")
					}
					if !bytes.Equal(mustEncodeState(t, settled), mustEncodeState(t, projected)) {
						t.Fatal("deferred ledger settlement differs from complete one-shot projection")
					}
				})
			}
		}
	}
	if cells != 20 {
		t.Fatalf("incomplete deferred-settlement population: got %d want20", cells)
	}
	t.Log("20 deferred settlements match actual one-shot Company projections; ApplyLogged/hooks/episode authority NOT_PROVEN")
}
