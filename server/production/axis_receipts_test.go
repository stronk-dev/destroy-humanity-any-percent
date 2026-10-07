package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

// Independent literal oracle: no AxisInput/AxisFactors/AxisProduct invocation.
func requireAxisReceipt(t *testing.T, raw []byte, expected string) {
	t.Helper()
	var receipt struct {
		Snapshot map[string]json.RawMessage `json:"snapshot"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	axis, ok := receipt.Snapshot["axis_stack"]
	if !ok || !canonicalJSONEqual(axis, []byte(expected)) {
		t.Fatalf("actual applied receipt axis_stack missing/wrong: %s", axis)
	}
}

const emptyAxisReceipt = `{"input_kind":"achievement_attainment_run","input_value":0,"input_cap":44,"cap_reason_key":"cap.axis_stack_input","saturated":false,"contributions":[],"product":"1e0"}`
const ownedAxisReceipt8 = `{"input_kind":"achievement_attainment_run","input_value":8,"input_cap":44,"cap_reason_key":"cap.axis_stack_input","saturated":false,"contributions":[{"source_id":"upgrade.pr_intern_1.axis","upgrade_id":"upgrade.pr_intern_1","factor":"1.2e0"}],"product":"1.2e0"}`

func TestAxisReceiptActualBoundaries(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	t.Run("ordinary_then_refreshed", func(t *testing.T) {
		state := axisTimingState(t, bundle, start)
		transition := axisTimingPurchase(t, bundle, state, start.Add(time.Second), ModeOnline)
		requireAxisReceipt(t, transition.Receipt, ownedAxisReceipt8)
		decision := save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt, Events: transition.Events}
		// A noop refresh must fail: start with a stale, valid-shaped empty
		// projection rather than reasserting the already-correct ordinary one.
		var stale map[string]any
		if err := json.Unmarshal(decision.Receipt, &stale); err != nil {
			t.Fatal(err)
		}
		stale["snapshot"].(map[string]any)["axis_stack"] = json.RawMessage(emptyAxisReceipt)
		decision.Receipt = mustJSON(stale)
		if err := refreshAppliedSnapshot(&decision, transition.State, bundle.Economy); err != nil {
			t.Fatal(err)
		}
		requireAxisReceipt(t, decision.Receipt, ownedAxisReceipt8)
	})
	current := activeContentBundle(t)
	for _, veteran := range []bool{false, true} {
		for _, gap := range []int64{0, 3114, 90000000} {
			row := activationResearchCase(t, current, bundle, veteran, gap, start)
			t.Run(row.ID, func(t *testing.T) {
				requireAxisReceipt(t, row.Exit.Receipt, emptyAxisReceipt)
				var old struct {
					Snapshot map[string]json.RawMessage `json:"snapshot"`
				}
				if err := json.Unmarshal(row.Old.Receipt, &old); err != nil {
					t.Fatal(err)
				}
				if _, exists := old.Snapshot["axis_stack"]; exists {
					t.Fatal("legacy receipt retrofitted axis")
				}
			})
		}
	}
}

func TestAxisReceiptProjectionBoundsAndOwnership(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, row := range []struct {
		name          string
		input         int64
		owned         int
		saturated     bool
		first, second string
		product       string
	}{
		{"empty", 0, 0, false, "1e0", "1e0", "1e0"},
		{"one", 8, 1, false, "1.2e0", "1.16e0", "1.2e0"},
		{"two", 6, 2, false, "1.15e0", "1.12e0", "1.288e0"},
		{"at_cap", 44, 2, false, "2.1e0", "1.88e0", "3.948e0"},
		{"above_cap_projection_only", 45, 2, true, "2.1e0", "1.88e0", "3.948e0"},
	} {
		t.Run(row.name, func(t *testing.T) {
			state := axisTimingState(t, bundle, start)
			// Direct read projection only: this table does not claim these
			// patched latched-score states passed full pinned save admission.
			state.AttainmentScoreRun = row.input
			state.UpgradesOwned = map[string]bool{}
			state.UpgradesOwned["upgrade.pr_intern_2"] = row.owned == 2
			state.UpgradesOwned["upgrade.pr_intern_1"] = row.owned >= 1
			contributions := "[]"
			if row.owned >= 1 {
				contributions = fmt.Sprintf(`[{"source_id":"upgrade.pr_intern_1.axis","upgrade_id":"upgrade.pr_intern_1","factor":%q}]`, row.first)
			}
			if row.owned == 2 {
				contributions = contributions[:len(contributions)-1] + fmt.Sprintf(`,{"source_id":"upgrade.pr_intern_2.axis","upgrade_id":"upgrade.pr_intern_2","factor":%q}]`, row.second)
			}
			expected := fmt.Sprintf(`{"input_kind":"achievement_attainment_run","input_value":%d,"input_cap":44,"cap_reason_key":"cap.axis_stack_input","saturated":%t,"contributions":%s,"product":%q}`, row.input, row.saturated, contributions, row.product)
			projection, err := wireSnapshot(state, bundle.Economy)
			if err != nil {
				t.Fatal(err)
			}
			requireAxisReceipt(t, mustJSON(map[string]any{"snapshot": projection}), expected)
		})
	}
	contrast := axisContentBundleWithInput(t, "achievement_score_run")
	state := axisTimingState(t, contrast, start)
	state.AchievementScoreRun = 8
	projection, err := wireSnapshot(state, contrast.Economy)
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string]any
	if err := json.Unmarshal([]byte(ownedAxisReceipt8), &expected); err != nil {
		t.Fatal(err)
	}
	expected["input_kind"] = "achievement_score_run"
	requireAxisReceipt(t, mustJSON(map[string]any{"snapshot": projection}), string(mustJSON(expected)))

	state = axisTimingState(t, bundle, start)
	state.AttainmentScoreRun = -1
	if projected, err := wireSnapshot(state, bundle.Economy); !errors.Is(err, ErrInvalidEngineState) || projected != nil {
		t.Fatalf("invalid projection not refused: %v %v", projected, err)
	}
	decision, err := appliedDecision(IntentRequest{IntentID: "01986666-0901-7000-8000-000000000901"}, state, bundle.Economy, 2, 1, state.Ledger.Snapshot(), nil, nil)
	if !errors.Is(err, ErrInvalidEngineState) || decision.Receipt != nil {
		t.Fatalf("ordinary snapshot error degraded: %s %v", decision.Receipt, err)
	}
	decision.Receipt = []byte(`{"snapshot":{},"outcome":"applied"}`)
	before := bytes.Clone(decision.Receipt)
	if err := refreshAppliedSnapshot(&decision, state, bundle.Economy); !errors.Is(err, ErrInvalidEngineState) || !bytes.Equal(before, decision.Receipt) {
		t.Fatalf("refresh error degraded/mutated receipt: %s %v", decision.Receipt, err)
	}
}
