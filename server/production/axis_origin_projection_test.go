package production

import (
	"bytes"
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
