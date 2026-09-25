package gameui

import (
	"sort"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// axisStackArm is the Clout v1 CV5/CV9 producer block: the run-local axis
// input, the owned axis contributions and their product, the attainment set
// (with an earned-this-run flag), and every axis-scaled upgrade's progress and
// factor at the current input. Every factor is server-derived; the surface
// never computes one. Null unless the pinned economy declares axis_stack.
type axisStackArm struct {
	InputKind     string             `json:"input_kind"`
	InputValue    int64              `json:"input_value"`
	InputCap      int64              `json:"input_cap"`
	CapReasonKey  string             `json:"cap_reason_key"`
	Saturated     bool               `json:"saturated"`
	Product       string             `json:"product"`
	Contributions []axisContribution `json:"contributions"`
	Attained      []axisAttainedRow  `json:"attained"`
	Interns       []axisInternRow    `json:"interns"`
}

type axisContribution struct {
	SourceID  string `json:"source_id"`
	UpgradeID string `json:"upgrade_id"`
	Factor    string `json:"factor"`
}

type axisAttainedRow struct {
	AchievementID string `json:"achievement_id"`
	EarnedThisRun bool   `json:"earned_this_run"`
}

type axisInternRow struct {
	UpgradeID string `json:"upgrade_id"`
	Minimum   int64  `json:"minimum"`
	FactorPPM int64  `json:"factor_ppm"`
	Owned     bool   `json:"owned"`
	Factor    string `json:"factor"`
}

func projectAxisStack(bundle production.CatalogBundle, state *save.State) (*axisStackArm, error) {
	axis, declared := bundle.Economy.AxisStack()
	if !declared {
		return nil, nil
	}
	_, saturated, err := production.AxisInput(state, bundle.Economy)
	if err != nil {
		return nil, err
	}
	input := state.AttainmentScoreRun
	if axis.Input == economy.AxisInputScoreRun {
		input = state.AchievementScoreRun
	}
	factors, err := production.AxisFactors(state, bundle.Economy)
	if err != nil {
		return nil, err
	}
	arm := &axisStackArm{InputKind: string(axis.Input), InputValue: input, InputCap: axis.InputCap, CapReasonKey: axis.CapReasonKey,
		Saturated: saturated, Contributions: []axisContribution{}, Attained: []axisAttainedRow{}, Interns: []axisInternRow{}}
	for _, upgrade := range bundle.Economy.Upgrades() {
		for _, effect := range upgrade.Effects {
			if effect.Slot != economy.SlotAxisStack {
				continue
			}
			factor := factors[effect.SourceID]
			owned := state.UpgradesOwned[upgrade.ID]
			arm.Interns = append(arm.Interns, axisInternRow{UpgradeID: upgrade.ID, Minimum: upgrade.AxisMinimum, FactorPPM: effect.FactorPPM, Owned: owned, Factor: factor.String()})
			if owned {
				arm.Contributions = append(arm.Contributions, axisContribution{SourceID: effect.SourceID, UpgradeID: upgrade.ID, Factor: factor.String()})
			}
		}
	}
	sort.Slice(arm.Contributions, func(l, r int) bool { return arm.Contributions[l].SourceID < arm.Contributions[r].SourceID })
	sort.Slice(arm.Interns, func(l, r int) bool { return arm.Interns[l].UpgradeID < arm.Interns[r].UpgradeID })
	product, err := production.AxisProduct(state, bundle.Economy)
	if err != nil {
		return nil, err
	}
	arm.Product = product.String()
	for id := range state.AchievementsAttainedRun {
		arm.Attained = append(arm.Attained, axisAttainedRow{AchievementID: id, EarnedThisRun: state.AchievementsEarnedRun[id]})
	}
	sort.Slice(arm.Attained, func(l, r int) bool { return arm.Attained[l].AchievementID < arm.Attained[r].AchievementID })
	return arm, nil
}
