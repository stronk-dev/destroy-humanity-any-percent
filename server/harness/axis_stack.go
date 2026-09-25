package harness

import (
	"errors"
	"fmt"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// Clout v1 CV10. The harness runtimes (Phase-0 and first-hour) execute
// pre-foundation Company semantics: they never run the achievement hook, so
// they cannot evaluate the run-local attainment input an axis stack reads.
// A scenario whose economy declares axis_stack is therefore refused loudly
// rather than simulated with a silently neutral stack (evidence discipline 3).
var ErrAxisStackUnevaluated = errors.New("axis_stack economy requires attainment evaluation the harness runtime does not execute")

func refuseAxisStack(catalog *economy.Catalog) error {
	if catalog == nil {
		return nil
	}
	if axis, declared := catalog.AxisStack(); declared {
		return fmt.Errorf("%w (input %s)", ErrAxisStackUnevaluated, axis.Input)
	}
	return nil
}

// CheckAxisInputWithinCap is the axis_input_within_cap invariant: on a
// Company that carries attainment under an economy declaring the axis stack,
// the clamped input never exceeds the visible cap and the raw attainment
// score is a non-negative exact integer. It is inert on states without the
// Company v19 attainment set.
func CheckAxisInputWithinCap(catalog *economy.Catalog, state *save.State) error {
	axis, declared := catalog.AxisStack()
	if !declared || state == nil || state.AchievementsAttainedRun == nil {
		return nil
	}
	if state.AttainmentScoreRun < 0 {
		return fmt.Errorf("axis_input_within_cap: negative attainment score %d", state.AttainmentScoreRun)
	}
	x, saturated, err := production.AxisInput(state, catalog)
	if err != nil {
		return fmt.Errorf("axis_input_within_cap: %w", err)
	}
	raw := state.AttainmentScoreRun
	if axis.Input == economy.AxisInputScoreRun {
		raw = state.AchievementScoreRun
	}
	if x > axis.InputCap || x < 0 || saturated != (raw > axis.InputCap) || !saturated && x != raw {
		return fmt.Errorf("axis_input_within_cap: x=%d cap=%d saturated=%t", x, axis.InputCap, saturated)
	}
	return nil
}
