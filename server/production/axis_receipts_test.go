package production

import (
	"encoding/json"
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
