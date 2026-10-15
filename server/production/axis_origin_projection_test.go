package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

// Test-only whole-Company origin. This deliberately uses the existing codec,
// not a proposed production accrual schema. No production caller uses it.
func originProjectionState(t *testing.T, bundle CatalogBundle, origin []byte, now time.Time, mode EvaluationMode) *save.State {
	t.Helper()
	return originProjectionStateWithExternal(t, bundle, origin, now, mode, nil)
}

func originProjectionStateWithExternal(t *testing.T, bundle CatalogBundle, origin []byte, now time.Time, mode EvaluationMode, external []multiplier.Contribution) *save.State {
	t.Helper()
	state, err := save.RestoreState(origin, 19, bundle.Economy, economy.ScopeCompany, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatal(err)
	}
	contributions, err := assembleContributions(state, bundle.Economy, external)
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

// This constrains an origin-carrying search proposal; the current beam does not
// carry these origins. It is not an execution of the beam or a live-player loss.
func TestAxisOriginProjectionSearchIdentityResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name                    string
		endMS, cutMS            int64
		wantOriginal, wantReset string
		wantEqual               bool
	}{
		{"retained_rounding_counterexample", 3114, 1553, "1.00035846811e4", "1.00035846812e4", false},
		{"whole_second_control", 2000, 1000, "1.00023023e4", "1.00023023e4", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := mustEncodeState(t, axisTimingState(t, bundle, start))
			originalBefore := bytes.Clone(original)
			cut := start.Add(time.Duration(tc.cutMS) * time.Millisecond)
			visible := originProjectionState(t, bundle, original, cut, ModeOnline)
			reset := mustEncodeState(t, visible)
			resetBefore := bytes.Clone(reset)
			if bytes.Equal(original, reset) {
				t.Fatal("population does not contain two distinct origins")
			}
			resetVisible := originProjectionState(t, bundle, reset, cut, ModeOnline)
			if !bytes.Equal(mustEncodeState(t, visible), mustEncodeState(t, resetVisible)) ||
				!visible.EvaluatedThrough.Equal(cut) || !resetVisible.EvaluatedThrough.Equal(cut) {
				t.Fatal("origins do not project to the same complete visible state and time")
			}
			end := start.Add(time.Duration(tc.endMS) * time.Millisecond)
			fromOriginal := originProjectionState(t, bundle, original, end, ModeOnline)
			fromReset := originProjectionState(t, bundle, reset, end, ModeOnline)
			if got := fromOriginal.Ledger.Snapshot()["company.cash"]; got != tc.wantOriginal {
				t.Fatalf("original-origin cash=%s want=%s", got, tc.wantOriginal)
			}
			if got := fromReset.Ledger.Snapshot()["company.cash"]; got != tc.wantReset {
				t.Fatalf("reset-origin cash=%s want=%s", got, tc.wantReset)
			}
			equal := bytes.Equal(mustEncodeState(t, fromOriginal), mustEncodeState(t, fromReset))
			if equal != tc.wantEqual {
				t.Fatalf("future full-state equality=%t want=%t", equal, tc.wantEqual)
			}
			if !bytes.Equal(original, originalBefore) || !bytes.Equal(reset, resetBefore) {
				t.Fatal("projection mutated an origin")
			}
			t.Logf("same visible state/time, distinct origins: future cash %s versus %s; equal=%t",
				tc.wantOriginal, tc.wantReset, equal)
		})
	}
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

// Research of the closed transition boundary, not an implementation of AC6.
// Scratch work must not consume the interval that ApplyLogged and its real
// accrual hooks still need. Nonzero cash and a genuine purchase/reattainment
// distinguish this population from the earlier zero-origin ledger experiment.
func TestAxisOriginDeferredLoggedPurchaseResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(307))
	profiles, preadvancedLosses := 0, 0
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
				profiles++
				initial := axisTimingState(t, bundle, start)
				origin := mustEncodeState(t, initial)
				at := start.Add(time.Duration(end) * time.Millisecond)
				baseline := activationRestore(t, bundle, origin, 19, start)
				ordinary := axisTimingPurchase(t, bundle, baseline, at, mode)
				if !baseline.LifetimeValue.Gt(initial.LifetimeValue) ||
					baseline.AttainmentScoreRun != 8 || baseline.GeneratorCounts["generator.beige_tower"] != 2 ||
					!hasEventKind(ordinary.Events, save.EventAchievementReattained) {
					t.Fatal("baseline did not execute positive lifetime accrual and purchase reattainment")
				}

				deferred := activationRestore(t, bundle, origin, 19, start)
				for _, offset := range []int64{cut, end} {
					projection := originProjectionState(t, bundle, origin, start.Add(time.Duration(offset)*time.Millisecond), mode)
					if !projection.EvaluatedThrough.Equal(start.Add(time.Duration(offset)*time.Millisecond)) ||
						!mustDecimal(t, projection.Ledger.Snapshot()["company.cash"]).Gt(mustDecimal(t, initial.Ledger.Snapshot()["company.cash"])) {
						t.Fatal("scratch projection did not execute the declared nonzero interval")
					}
					if !bytes.Equal(origin, mustEncodeState(t, deferred)) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatal("internal projection consumed the committed origin")
					}
				}
				settled := axisTimingPurchase(t, bundle, deferred, at, mode)
				ordinaryEvents, err := json.Marshal(ordinary.Events)
				if err != nil {
					t.Fatal(err)
				}
				settledEvents, err := json.Marshal(settled.Events)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(mustEncodeState(t, baseline), mustEncodeState(t, deferred)) ||
					!bytes.Equal(ordinary.Receipt, settled.Receipt) || !bytes.Equal(ordinaryEvents, settledEvents) {
					t.Fatal("scratch cuts changed the complete logged state, receipt or ordered events")
				}

				// Deliberately wrong integration: install the final projected
				// state before the real transition, without running its hooks.
				// An applied purchase and matching cash are not sufficient.
				preadvanced := originProjectionState(t, bundle, origin, at, mode)
				wrong := axisTimingPurchase(t, bundle, preadvanced, at, mode)
				if wrong.Outcome != save.IntentApplied || preadvanced.Ledger.Snapshot()["company.cash"] != baseline.Ledger.Snapshot()["company.cash"] {
					t.Fatal("preadvance control failed before demonstrating the hidden hook loss")
				}
				if !preadvanced.LifetimeValue.Eq(initial.LifetimeValue) || !baseline.LifetimeValue.Gt(preadvanced.LifetimeValue) {
					t.Fatal("full transition oracle failed to distinguish missing prestige accrual")
				}
				preadvancedLosses++
			})
		}
	}
	if profiles != 128 || preadvancedLosses != profiles {
		t.Fatalf("incomplete logged-origin population: profiles=%d hook-loss controls=%d", profiles, preadvancedLosses)
	}
	t.Log("128 scratch-cut logged purchases retain full state/receipts/events; 128 live-preadvance controls lose prestige accrual despite matching cash; AC6 NOT_PROVEN")
}

// The Company alone is not a complete producer origin: these contributions
// come from outside it. Use existing replay-owned context, not rounded rates
// or a proposed save field, to test the next bounded composition requirement.
func TestAxisOriginDeferredExternalContextResearch(t *testing.T) {
	if len(originExternalReplayActions(t)) != 6 {
		t.Fatal("external replay action census incomplete")
	}
}

func originExternalReplayActions(t *testing.T) []axisSequenceAction {
	t.Helper()
	var actions []axisSequenceAction
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	at := start.Add(2001 * time.Millisecond)
	commonsCatalog, err := commons.LoadCatalog(bundle.Artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	commonsValues, err := commons.Contribution(commonsCatalog, true, 800_000, 500_000)
	if err != nil || len(commonsValues) != 1 || !commonsValues[0].Factor.Gt(decimal.One) {
		t.Fatalf("Commons producer did not supply a positive modifier: values=%v err=%v", commonsValues, err)
	}
	carry := tier2FounderCarry(t, bundle, start)
	founder, err := stateFromFounderCarry(carry, bundle)
	if err != nil {
		t.Fatal(err)
	}
	founder.FiscalCredit = 9
	founder.FiscalGeneratorLevels["generator.beige_tower"] = 2
	frozen, err := FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil {
		t.Fatal(err)
	}
	fiscalValues, err := ResolveFrozenContributions(bundle.Economy, frozen)
	if err != nil || len(fiscalValues) != 2 {
		t.Fatalf("Fiscal producer context=%v err=%v", fiscalValues, err)
	}
	combined := append(append([]multiplier.Contribution{}, commonsValues...), fiscalValues...)
	combined = append(combined, multiplier.Contribution{SourceID: guild.StockConsumptionSourceID,
		Slot: multiplier.SlotFaction, Target: "all", Factor: decimal.One})
	cells := 0
	for _, profile := range []struct {
		name     string
		external []multiplier.Contribution
		member   bool
		guild    bool
	}{
		{"commons", commonsValues, true, false},
		{"fiscal", fiscalValues, false, false},
		{"commons_fiscal_guild", combined, true, true},
	} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(profile.name+"_"+string(mode), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				var weight *int64
				if profile.member {
					initial.CompactMember, initial.CompactTithePPM = true, commonsCatalog.DefaultTithePPM
					value := int64(500_000)
					weight = &value
				}
				if profile.guild {
					initial.GuildTitheCarryPPM = 999_999
				}
				if err := bundle.ValidateFoundationState(initial); err != nil {
					t.Fatalf("external-origin Company admission: %v", err)
				}
				origin := mustEncodeState(t, initial)
				accrual, err := makeReplayAccrual(profile.external, weight, guild.SettlementBatch{}, bundle.Routes.ContextVersion())
				if err != nil {
					t.Fatal(err)
				}
				encodedContext, err := json.Marshal(accrual)
				if err != nil {
					t.Fatal(err)
				}
				var retained replayAccrual
				if err := json.Unmarshal(encodedContext, &retained); err != nil {
					t.Fatal(err)
				}
				external, err := contributionsFromReplay(retained)
				if err != nil {
					t.Fatal(err)
				}
				liveContributions, err := assembleContributions(initial, bundle.Economy, profile.external)
				if err != nil {
					t.Fatal(err)
				}
				retainedContributions, err := assembleContributions(initial, bundle.Economy, external)
				if err != nil {
					t.Fatal(err)
				}
				liveRates, err := Rates(bundle.Economy, initial.GeneratorCounts, liveContributions)
				if err != nil {
					t.Fatal(err)
				}
				retainedRates, err := Rates(bundle.Economy, initial.GeneratorCounts, retainedContributions)
				if err != nil || len(liveRates) != len(retainedRates) {
					t.Fatalf("retained raw-rate resource population: %v", err)
				}
				for resource, rates := range liveRates {
					if len(rates) != len(retainedRates[resource]) {
						t.Fatalf("retained raw-rate source population for %s", resource)
					}
					for index, rate := range rates {
						if !rate.Eq(retainedRates[resource][index]) {
							t.Fatalf("external context changed raw rate for %s source%d", resource, index)
						}
					}
				}
				baseline := activationRestore(t, bundle, origin, 19, start)
				ordinary, request, inputs := originLoggedManualEvidence(t, bundle, baseline, at, mode, accrual)
				if !baseline.LifetimeValue.Gt(initial.LifetimeValue) ||
					profile.member && (!hasEventKind(ordinary.Events, save.EventCompactSampled) || len(baseline.CompactSamples) == 0) ||
					profile.guild && (!hasEventKind(ordinary.Events, save.EventGuildTitheAccrued) || baseline.GuildTitheCarryPPM == initial.GuildTitheCarryPPM) {
					t.Fatal("ordinary manual action did not execute the declared accrual hooks")
				}
				deferred := activationRestore(t, bundle, origin, 19, start)
				for _, offset := range []int64{1, 1000, 2000, 2001} {
					now := start.Add(time.Duration(offset) * time.Millisecond)
					expected := originProjectionStateWithExternal(t, bundle, origin, now, mode, profile.external)
					projected := originProjectionStateWithExternal(t, bundle, origin, now, mode, external)
					if !bytes.Equal(mustEncodeState(t, expected), mustEncodeState(t, projected)) ||
						!bytes.Equal(origin, mustEncodeState(t, deferred)) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatal("retained external context changed projection or consumed the committed origin")
					}
				}
				settled := originLoggedManual(t, bundle, deferred, at, mode, retained)
				ordinaryEvents, err := json.Marshal(ordinary.Events)
				if err != nil {
					t.Fatal(err)
				}
				settledEvents, err := json.Marshal(settled.Events)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(mustEncodeState(t, baseline), mustEncodeState(t, deferred)) ||
					!bytes.Equal(ordinary.Receipt, settled.Receipt) || !bytes.Equal(ordinaryEvents, settledEvents) {
					t.Fatal("retained producer context lost the closed manual transition state/receipt/events")
				}
				complete := originProjectionStateWithExternal(t, bundle, origin, at, mode, external)
				missing := originProjectionState(t, bundle, origin, at, mode)
				if complete.Ledger.Snapshot()["company.cash"] == missing.Ledger.Snapshot()["company.cash"] {
					t.Fatal("omitted external inputs escaped the projection control")
				}
				wrong := originLoggedManual(t, bundle, complete, at, mode, retained)
				if wrong.Outcome != save.IntentApplied || complete.Ledger.Snapshot()["company.cash"] != baseline.Ledger.Snapshot()["company.cash"] ||
					!complete.LifetimeValue.Eq(initial.LifetimeValue) ||
					profile.member && (hasEventKind(wrong.Events, save.EventCompactSampled) || len(complete.CompactSamples) != 0) ||
					profile.guild && (hasEventKind(wrong.Events, save.EventGuildTitheAccrued) || complete.GuildTitheCarryPPM != initial.GuildTitheCarryPPM) {
					t.Fatal("pre-advanced control failed to expose lost hooks despite equal cash")
				}
				actions = append(actions, originReplayAction(t, profile.name+"/"+string(mode), origin, request, inputs, baseline, ordinary))
				cells++
			})
		}
	}
	if cells != 6 {
		t.Fatalf("external-context census incomplete: %d", cells)
	}
	t.Log("6 external-context/manual cells preserve complete transitions; omission loses rates and pre-advance loses hooks; AC6 NOT_PROVEN")
	return actions
}

func originLoggedManual(t *testing.T, bundle CatalogBundle, state *save.State, now time.Time, mode EvaluationMode, accrual replayAccrual) LoggedTransition {
	transition, _, _ := originLoggedManualEvidence(t, bundle, state, now, mode, accrual)
	return transition
}

func originLoggedManualEvidence(t *testing.T, bundle CatalogBundle, state *save.State, now time.Time, mode EvaluationMode, accrual replayAccrual) (LoggedTransition, IntentRequest, []byte) {
	t.Helper()
	request, inputs := originManualInputs(t, bundle, state, now, mode, accrual, "manual.click")
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatalf("external-origin manual action outcome=%s error=%v", transition.Outcome, err)
	}
	return transition, request, inputs
}

func originReplayAction(t *testing.T, name string, before []byte, request IntentRequest, inputs []byte, state *save.State, transition LoggedTransition) axisSequenceAction {
	t.Helper()
	return axisSequenceAction{Name: name, PreStateJSON: canonicalFixtureJSON(t, before),
		Payload: request.CanonicalPayload, Inputs: json.RawMessage(inputs), Outcome: string(transition.Outcome),
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)),
		PostStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, state))}
}

func TestAxisExternalReplayEvidenceCapture(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	actions := append(originExternalReplayActions(t), originGuildReplayActions(t)...)
	expected := []string{}
	for _, profile := range []string{"commons", "fiscal", "commons_fiscal_guild", "single", "ordered", "nonzero_base"} {
		for _, mode := range []string{"online", "offline"} {
			expected = append(expected, profile+"/"+mode)
		}
	}
	if len(actions) != len(expected) {
		t.Fatal("external/Guild capture is not the complete twelve-cell population")
	}
	for index, action := range actions {
		t.Run("captured/"+action.Name, func(t *testing.T) {
			if action.Name != expected[index] || action.Outcome != string(save.IntentApplied) {
				t.Fatal("captured action identity/outcome mismatch")
			}
			state := activationRestore(t, bundle, []byte(action.PreStateJSON), 19, start)
			if canonicalFixtureJSON(t, mustEncodeState(t, state)) != action.PreStateJSON {
				t.Fatal("capture lost complete admitted pre-state")
			}
			result, err := ApplyLogged(state, action.Payload, bundle, action.Inputs)
			if err != nil || result.Outcome != save.IntentApplied ||
				canonicalFixtureJSON(t, result.Receipt) != action.ReceiptJSON ||
				canonicalFixtureValue(t, fixtureEvents(result.Events)) != action.EventsJSON ||
				canonicalFixtureJSON(t, mustEncodeState(t, state)) != action.PostStateJSON {
				t.Fatalf("captured complete transition does not replay: %v", err)
			}
		})
	}
}

func originManualInputs(t *testing.T, bundle CatalogBundle, state *save.State, now time.Time, mode EvaluationMode, accrual replayAccrual, actionID string) (IntentRequest, []byte) {
	t.Helper()
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":"01986666-0901-7000-8000-000000000902","kind":"perform_manual_batch","expected_revision":1,"action_id":%q,"count":1,"window_ms":1000}`, actionID)))
	if err != nil {
		t.Fatal(err)
	}
	return request, originRequestInputs(t, bundle, state, now, mode, accrual, request)
}

func originRequestInputs(t *testing.T, bundle CatalogBundle, state *save.State, now time.Time, mode EvaluationMode, accrual replayAccrual, request IntentRequest) []byte {
	t.Helper()
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001",
		FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: state.RunSeq, RunLogSeq: 1}
	carry := tier2FounderCarry(t, bundle, state.RunStartedAt)
	active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	external, err := contributionsFromReplay(accrual)
	if err != nil {
		t.Fatal(err)
	}
	encodedBatch := accrual.GuildSettlementBatch
	batch := guild.SettlementBatch{GuildID: encodedBatch.GuildID, BaseSeq: encodedBatch.BaseSeq}
	for _, row := range encodedBatch.Settlements {
		batch.Settlements = append(batch.Settlements, guild.Settlement{GuildID: batch.GuildID,
			BoundarySeq: row.BoundarySeq, DebitUnits: row.DebitUnits, CreditUnits: row.CreditUnits})
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: mode, Now: now, IntentKind: request.Kind,
		Contributions: external, CommonsWeightPPM: accrual.CommonsWeightPPM,
		GuildSettlementBatch: batch,
		RouteContextVersion:  accrual.RouteContextVersion, FounderCarry: &carry, ActivePlay: &active})
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

type originGuildMember struct{}

type originTemporalReplayRow struct {
	ID             string             `json:"id"`
	Profile        string             `json:"profile"`
	Mode           EvaluationMode     `json:"mode"`
	OffsetMS       int64              `json:"offset_ms"`
	AttendedMS     int64              `json:"attended_ms"`
	FrozenChanged  bool               `json:"frozen_changed"`
	OmittedChanged bool               `json:"omitted_changed"`
	Action         axisSequenceAction `json:"action"`
}

type originTemporalReplayReport struct {
	Version    int                       `json:"version"`
	Acceptance string                    `json:"acceptance_status"`
	Sources    map[string]string         `json:"source_sha256"`
	Bundle     reputationCorpusBundle    `json:"bundle"`
	Rows       []originTemporalReplayRow `json:"rows"`
	Negatives  int                       `json:"invalid_cases"`
}

// A controlled, already-active history, not a claim that these buffs were
// acquired naturally. The pinned duration and target remain authoritative.
func TestAxisOriginDeferredTemporalContextResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	report := originTemporalReplayReport{
		Version: 1, Acceptance: "NOT_PROVEN: bounded temporal replay compatibility; production AC6 remains red",
		Bundle: reputationCorpusBundle{ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)},
		Sources: sourceResearchHashes(t, []string{
			"server/production/axis_origin_projection_test.go", "client/test/axis-temporal-replay.test.ts",
			"server/production/axis_timing_test.go", "server/production/axis_stack_test.go",
			"server/production/replay.go", "server/production/active_play.go", "server/production/engine.go",
			"server/save/state.go", "client/src/replay.ts", "client/src/economy.ts",
			"balance/testdata/axis-stack/economy-v5-fixture.json", "balance/opportunities/t0-t1.json",
		}),
	}
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	founderID := "01986666-2000-7000-8000-000000000001"
	completed, frozenDifferences, omittedDifferences := 0, 0, 0
	for _, profile := range []struct {
		name                 string
		production, building bool
	}{
		{"production", true, false}, {"building", false, true}, {"combined", true, true},
	} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			for _, offset := range []int64{999, 1000, 1001, 1499, 1500, 1501, 6000} {
				t.Run(fmt.Sprintf("%s_%s_%d", profile.name, mode, offset), func(t *testing.T) {
					initial := axisTimingState(t, bundle, start)
					initial.RunStartedAt = start.Add(-4000 * time.Millisecond)
					if profile.production {
						effect, ok := bundle.Opportunities.Effect("active.production")
						if !ok || effect.DurationMS != 5000 {
							t.Fatal("pinned production duration changed")
						}
						initial.ActiveBuffs = append(initial.ActiveBuffs, save.ActiveBuff{
							BuffInstanceID: "01985555-0000-7000-8000-000000000001", EffectRowID: effect.ID,
							ActivatedAttendedMS: 0, ExpiresAttendedMS: effect.DurationMS})
					}
					if profile.building {
						effect, ok := bundle.Opportunities.Effect("active.building")
						if !ok || effect.DurationMS != 2000 {
							t.Fatal("pinned building duration changed")
						}
						target := "generator.beige_tower"
						initial.ActiveBuffs = append(initial.ActiveBuffs, save.ActiveBuff{
							BuffInstanceID: "01985555-0000-7000-8000-000000000002", EffectRowID: effect.ID, SelectedTarget: &target,
							ActivatedAttendedMS: 3500, ExpiresAttendedMS: 3500 + effect.DurationMS})
					}
					if err := bundle.ValidateFoundationState(initial); err != nil {
						t.Fatal(err)
					}
					origin := mustEncodeState(t, initial)
					at := start.Add(time.Duration(offset) * time.Millisecond)
					accrual, err := makeReplayAccrual(nil, nil, guild.SettlementBatch{}, bundle.Routes.ContextVersion())
					if err != nil {
						t.Fatal(err)
					}
					request, inputs := originManualInputs(t, bundle, initial, at, mode, accrual, "manual.click")
					var wire replayInputsWire
					if err := json.Unmarshal(inputs, &wire); err != nil {
						t.Fatal(err)
					}
					var resolved replayAccrualResolved
					if err := json.Unmarshal(wire.Resolved, &resolved); err != nil || resolved.ActivePlay == nil {
						t.Fatalf("active schedule missing: %v", err)
					}
					wire.Resolved, err = json.Marshal(resolved)
					if err != nil {
						t.Fatal(err)
					}
					retained, err := json.Marshal(wire)
					if err != nil || !bytes.Equal(inputs, retained) {
						t.Fatalf("existing typed replay codec changed retained inputs: %v", err)
					}
					attended := int64(4000) + offset
					if offset > bundle.Prestige.CatchupCeilingMS {
						attended = 4000
					}
					if resolved.ActivePlay.AttendedNowMS != attended {
						t.Fatalf("attended coordinate: got%d want%d", resolved.ActivePlay.AttendedNowMS, attended)
					}
					expired := []activePlayExpiredBuff{}
					kept := []save.ActiveBuff{}
					for _, buff := range initial.ActiveBuffs {
						if buff.ExpiresAttendedMS <= attended {
							expired = append(expired, activePlayExpiredBuff{BuffInstanceID: buff.BuffInstanceID})
						} else {
							kept = append(kept, buff)
						}
					}
					if canonicalFixtureValue(t, expired) != canonicalFixtureValue(t, resolved.ActivePlay.ExpiredBuffs) {
						t.Fatal("wrong literal expiry identities")
					}
					// Scratch schedule and evaluator consume a fresh copy, never the
					// committed origin. This is not an intermediate saved head.
					projection := activationRestore(t, bundle, origin, 19, start)
					if _, err := applyActivePlaySchedule(projection, bundle.Opportunities, bundle.Prestige, founderID, at, *resolved.ActivePlay); err != nil {
						t.Fatal(err)
					}
					active, err := activePlayContributions(projection, bundle.Opportunities, attended)
					if err != nil {
						t.Fatal(err)
					}
					projected := originProjectionStateWithExternal(t, bundle, mustEncodeState(t, projection), at, mode, active)
					frozen, err := activePlayContributions(initial, bundle.Opportunities, 4000)
					if err != nil {
						t.Fatal(err)
					}
					wrongFrozen := originProjectionStateWithExternal(t, bundle, origin, at, mode, frozen)
					wrongMissing := originProjectionState(t, bundle, origin, at, mode)
					cash := projected.Ledger.Snapshot()["company.cash"]
					frozenChanged := cash != wrongFrozen.Ledger.Snapshot()["company.cash"]
					omittedChanged := cash != wrongMissing.Ledger.Snapshot()["company.cash"]
					if frozenChanged != (len(expired) > 0) || omittedChanged != (len(active) > 0) {
						t.Fatal("frozen/omitted controls did not discriminate exact active context")
					}
					ordinaryState := activationRestore(t, bundle, origin, 19, start)
					ordinary, err := ApplyLogged(ordinaryState, request.CanonicalPayload, bundle, inputs)
					if err != nil || ordinary.Outcome != save.IntentApplied {
						t.Fatalf("ordinary transition: %v", err)
					}
					// Bind the scratch accrual to the real action's ledger, not merely
					// to a second execution of the same replay input. Only the one
					// accepted manual action is added through the existing ledger.
					action, ok := bundle.Economy.ManualAction("manual.click")
					if !ok {
						t.Fatal("pinned manual action missing")
					}
					contributions, err := assembleContributions(projection, bundle.Economy, active)
					if err != nil {
						t.Fatal(err)
					}
					factor, err := contributionFactorForTarget(bundle.Economy, action.ID, contributions)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := projected.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{
						ResourceID: action.Output.ResourceID, Delta: action.Output.AmountPerAction.Mul(factor),
					}}}); err != nil {
						t.Fatal(err)
					}
					if canonicalFixtureValue(t, projected.Ledger.Snapshot()) != canonicalFixtureValue(t, ordinaryState.Ledger.Snapshot()) ||
						canonicalFixtureValue(t, ordinaryState.ActiveBuffs) != canonicalFixtureValue(t, kept) {
						t.Fatal("scheduled scratch accrual plus manual action missed actual ledger or kept buff identities")
					}
					if offset > bundle.Prestige.CatchupCeilingMS {
						if len(ordinaryState.OfflineSpans) != 1 || !ordinaryState.OfflineSpans[0].From.Equal(start) || !ordinaryState.OfflineSpans[0].To.Equal(at) {
							t.Fatal("long interval did not retain the exact absence span")
						}
					} else if len(ordinaryState.OfflineSpans) != 0 {
						t.Fatal("short interval unexpectedly recorded an absence")
					}
					deferred := activationRestore(t, bundle, origin, 19, start)
					if !bytes.Equal(origin, mustEncodeState(t, initial)) || !bytes.Equal(origin, mustEncodeState(t, deferred)) {
						t.Fatal("scratch projection consumed committed origin")
					}
					settled, err := ApplyLogged(deferred, request.CanonicalPayload, bundle, retained)
					if err != nil || settled.Outcome != save.IntentApplied || !bytes.Equal(mustEncodeState(t, deferred), mustEncodeState(t, ordinaryState)) ||
						!bytes.Equal(settled.Receipt, ordinary.Receipt) || canonicalFixtureValue(t, settled.Events) != canonicalFixtureValue(t, ordinary.Events) ||
						!deferred.LifetimeValue.Gt(initial.LifetimeValue) || len(deferred.ActiveBuffs) != len(initial.ActiveBuffs)-len(expired) {
						t.Fatalf("retained temporal context lost complete transition: %v", err)
					}
					for _, missing := range []bool{false, true} {
						bad := resolved
						if missing {
							bad.ActivePlay = nil
						} else {
							changed := *resolved.ActivePlay
							changed.AttendedNowMS++
							bad.ActivePlay = &changed
						}
						badWire := wire
						badWire.Resolved, err = json.Marshal(bad)
						if err != nil {
							t.Fatal(err)
						}
						badInputs, err := json.Marshal(badWire)
						if err != nil {
							t.Fatal(err)
						}
						refused := activationRestore(t, bundle, origin, 19, start)
						if _, err := ApplyLogged(refused, request.CanonicalPayload, bundle, badInputs); err == nil || !bytes.Equal(origin, mustEncodeState(t, refused)) {
							t.Fatalf("forged/missing schedule did not refuse unchanged (missing=%v): %v", missing, err)
						}
						report.Negatives++
					}
					id := fmt.Sprintf("%s/%s/%d", profile.name, mode, offset)
					report.Rows = append(report.Rows, originTemporalReplayRow{
						ID: id, Profile: profile.name, Mode: mode, OffsetMS: offset, AttendedMS: attended,
						FrozenChanged: frozenChanged, OmittedChanged: omittedChanged,
						Action: axisSequenceAction{Name: id + "/perform_manual_batch", PreStateJSON: canonicalFixtureJSON(t, origin),
							Payload: request.CanonicalPayload, Inputs: retained, Outcome: string(settled.Outcome),
							ReceiptJSON: canonicalFixtureJSON(t, settled.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(settled.Events)),
							PostStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, deferred))},
					})
					if frozenChanged {
						frozenDifferences++
					}
					if omittedChanged {
						omittedDifferences++
					}
					completed++
				})
			}
		}
	}
	if completed != 42 || frozenDifferences != 24 || omittedDifferences != 24 {
		t.Fatalf("incomplete temporal population: cells %d/42 frozen %d/24 omitted %d/24", completed, frozenDifferences, omittedDifferences)
	}
	if len(report.Rows) != 42 || report.Negatives != 84 {
		t.Fatalf("incomplete temporal replay corpus: %d/42, negatives %d/84", len(report.Rows), report.Negatives)
	}
	sourceResearchRegressionArtifact(t, "../../testdata/axis-stack/temporal-replay-research-v1.json", "UPDATE_TEMPORAL_REPLAY_RESEARCH", &report, &report.Sources)
	t.Log("42 temporal cells: closed state/receipt/events retained, 24 frozen and 24 omitted cash differences, forged/missing schedule refused; AC6 NOT_PROVEN")
}

func (originGuildMember) FounderGuildMember(context.Context, string) (bool, error) { return true, nil }

// Observes the real producer's input without replacing its rate calculation.
type originSettlementProvider struct {
	provider faction.StockConsumptionProvider
	observed []byte
}

func (provider *originSettlementProvider) Contributions(ctx context.Context, state *save.State, catalog *economy.Catalog, revision save.Revision) ([]multiplier.Contribution, error) {
	var err error
	provider.observed, err = save.EncodeState(state)
	if err != nil {
		return nil, err
	}
	return provider.provider.Contributions(ctx, state, catalog, revision)
}

func TestAxisOriginDeferredGuildSettlementResearch(t *testing.T) {
	if len(originGuildReplayActions(t)) != 6 {
		t.Fatal("Guild replay action census incomplete")
	}
}

func originGuildReplayActions(t *testing.T) []axisSequenceAction {
	t.Helper()
	var actions []axisSequenceAction
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	at := start.Add(2001 * time.Millisecond)
	guildID := "01986666-3000-7000-8000-000000000001"
	completed := 0
	for _, profile := range []struct {
		name string
		base int64
		rows []guild.Settlement
	}{
		{"single", 0, []guild.Settlement{{GuildID: guildID, BoundarySeq: 1, DebitUnits: 3, CreditUnits: 7}}},
		{"ordered", 0, []guild.Settlement{{GuildID: guildID, BoundarySeq: 1, DebitUnits: 3, CreditUnits: 7}, {GuildID: guildID, BoundarySeq: 2, DebitUnits: 2, CreditUnits: 1}}},
		{"nonzero_base", 11, []guild.Settlement{{GuildID: guildID, BoundarySeq: 12, DebitUnits: 5, CreditUnits: 2}}},
	} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(profile.name+"_"+string(mode), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				initial.FactionID, initial.IncorporatedAt = "bootstrapper", start
				setCash(t, initial, "0")
				initial.StockUnits, initial.ConsumedStockUnits = 20, 3
				initial.GuildBoundaryGuildID, initial.GuildBoundarySeq = guildID, profile.base
				initial.GuildConsumedWindow, initial.GuildTitheCarryPPM = 1, 999_999
				origin := mustEncodeState(t, initial)
				batch := guild.SettlementBatch{GuildID: guildID, BaseSeq: profile.base, Settlements: profile.rows}
				expectedSettled := activationRestore(t, bundle, origin, 19, start)
				if err := guild.ApplySettlements(expectedSettled, batch, bundle.Faction.StockCap); err != nil {
					t.Fatal(err)
				}
				provider := &originSettlementProvider{provider: faction.StockConsumptionProvider{Catalogs: bundle, Members: originGuildMember{}}}
				service := Service{contributions: provider, guildSettlements: fixedGuildSettlements{batch: batch}}
				values, resolvedBatch, err := service.resolveReplayAccrual(context.Background(), initial,
					save.Revision{OwnerID: "01986666-2000-7000-8000-000000000001", ConstantsHash: bundle.ConstantsHash},
					bundle.Economy, bundle.Faction.StockCap, IntentRequest{Kind: IntentPerformManualBatch})
				if err != nil || !bytes.Equal(provider.observed, mustEncodeState(t, expectedSettled)) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
					t.Fatalf("live resolution did not observe settled inputs and restore origin: %v", err)
				}
				// Epoch-8 Guild data explicitly has no output bonus. Do not invent one
				// merely to make cash an oracle for clearing state.
				if len(values) != 1 || values[0].SourceID != guild.StockConsumptionSourceID || !values[0].Factor.Eq(decimal.One) {
					t.Fatalf("actual pinned Guild producer: %+v", values)
				}
				accrual, err := makeReplayAccrual(values, nil, resolvedBatch, bundle.Routes.ContextVersion())
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := json.Marshal(accrual)
				if err != nil {
					t.Fatal(err)
				}
				var retained replayAccrual
				if err := json.Unmarshal(encoded, &retained); err != nil {
					t.Fatal(err)
				}
				retainedJSON, err := json.Marshal(retained)
				if err != nil || !bytes.Equal(encoded, retainedJSON) {
					t.Fatalf("existing codec lost exact settlement context: %v", err)
				}
				external, err := contributionsFromReplay(retained)
				if err != nil {
					t.Fatal(err)
				}
				ordinaryState := activationRestore(t, bundle, origin, 19, start)
				ordinary, ordinaryRequest, ordinaryInputs := originLoggedManualEvidence(t, bundle, ordinaryState, at, mode, accrual)
				if ordinaryState.StockUnits != expectedSettled.StockUnits || ordinaryState.ConsumedStockUnits != expectedSettled.ConsumedStockUnits ||
					ordinaryState.GuildBoundarySeq != profile.rows[len(profile.rows)-1].BoundarySeq || ordinaryState.GuildBoundaryGuildID != guildID ||
					ordinaryState.GuildConsumedWindow != expectedSettled.GuildConsumedWindow || !ordinaryState.LifetimeValue.Gt(initial.LifetimeValue) ||
					!hasEventKind(ordinary.Events, save.EventGuildTitheAccrued) {
					t.Fatal("logged action missed exact settlement coordinates or positive accrual hooks")
				}
				deferred := activationRestore(t, bundle, origin, 19, start)
				for _, offset := range []int64{1, 1000, 2000, 2001} {
					copy := activationRestore(t, bundle, origin, 19, start)
					if err := applyReplayGuildSettlements(copy, retained.GuildSettlementBatch, bundle.Faction.StockCap); err != nil {
						t.Fatal(err)
					}
					now := start.Add(time.Duration(offset) * time.Millisecond)
					projected := originProjectionStateWithExternal(t, bundle, mustEncodeState(t, copy), now, mode, external)
					expected := originProjectionStateWithExternal(t, bundle, mustEncodeState(t, expectedSettled), now, mode, values)
					if !bytes.Equal(mustEncodeState(t, projected), mustEncodeState(t, expected)) || !bytes.Equal(origin, mustEncodeState(t, deferred)) {
						t.Fatal("scratch clearing changed projection or consumed committed origin")
					}
				}
				settled := originLoggedManual(t, bundle, deferred, at, mode, retained)
				assertTransition := func(state *save.State, result LoggedTransition) {
					t.Helper()
					if !bytes.Equal(mustEncodeState(t, state), mustEncodeState(t, ordinaryState)) || !bytes.Equal(result.Receipt, ordinary.Receipt) ||
						canonicalFixtureValue(t, result.Events) != canonicalFixtureValue(t, ordinary.Events) {
						t.Fatal("closed settlement/manual state, exact receipt or ordered events changed")
					}
				}
				assertTransition(deferred, settled)
				missingBatch := retained
				missingBatch.GuildSettlementBatch = replayGuildSettlementBatch{Settlements: []replayGuildSettlement{}}
				omitted := activationRestore(t, bundle, origin, 19, start)
				originLoggedManual(t, bundle, omitted, at, mode, missingBatch)
				if bytes.Equal(mustEncodeState(t, omitted), mustEncodeState(t, ordinaryState)) || omitted.GuildBoundarySeq != profile.base ||
					omitted.Ledger.Snapshot()["company.cash"] != ordinaryState.Ledger.Snapshot()["company.cash"] {
					t.Fatal("omitted clearing was not exposed by state despite equal pinned neutral-rate cash")
				}
				beforeDuplicate := mustEncodeState(t, deferred)
				request, inputs := originManualInputs(t, bundle, deferred, at, mode, retained, "manual.click")
				if _, err := ApplyLogged(deferred, request.CanonicalPayload, bundle, inputs); err == nil || !bytes.Equal(beforeDuplicate, mustEncodeState(t, deferred)) {
					t.Fatalf("duplicate clearing accepted or mutated settled Company: %v", err)
				}
				rejectedState := activationRestore(t, bundle, origin, 19, start)
				request, err = ParseIntent([]byte(`{"intent_id":"01986666-0901-7000-8000-000000000902","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`))
				if err != nil {
					t.Fatal(err)
				}
				inputs = originRequestInputs(t, bundle, rejectedState, at, mode, retained, request)
				rejected, err := ApplyLogged(rejectedState, request.CanonicalPayload, bundle, inputs)
				if err != nil || rejected.Outcome != save.IntentRejected || len(rejected.Events) != 0 || !bytes.Equal(origin, mustEncodeState(t, rejectedState)) {
					t.Fatalf("late unaffordable purchase leaked settlement/accrual/hooks: outcome=%s err=%v", rejected.Outcome, err)
				}
				if !bytes.Equal(rejected.Receipt, marshalRejection(request.IntentID, 1, "unaffordable", "generator.beige_tower")) {
					t.Fatalf("rollback control did not reach the post-accrual affordability decision: %s", rejected.Receipt)
				}
				assertTransition(rejectedState, originLoggedManual(t, bundle, rejectedState, at, mode, retained))
				actions = append(actions, originReplayAction(t, profile.name+"/"+string(mode), origin, ordinaryRequest, ordinaryInputs, ordinaryState, ordinary))
				completed++
			})
		}
	}
	if completed != 6 {
		t.Fatalf("settlement population incomplete: %d/6", completed)
	}
	t.Log("6 nonempty Guild-context cells: exact closed transitions, settled producer inputs, omission/duplicate/rollback controls; AC6 NOT_PROVEN")
	return actions
}
