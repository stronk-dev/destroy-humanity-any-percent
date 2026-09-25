package gameui

import (
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

// reputationArm is Reputation Tree v1 R9's server-derived block, carried as a
// nullable v4 `features.reputation` arm: null exactly when the pinned bundle
// has no tree or the Founder predates v22. The client never recomputes
// eligibility; node state is derived here from the pinned tree.
type reputationArm struct {
	Available          int64           `json:"available"`
	BonusFactorNextRun string          `json:"bonus_factor_next_run"`
	BonusFactorThisRun *string         `json:"bonus_factor_this_run"`
	Level              int64           `json:"level"`
	Nodes              []reputationRow `json:"nodes"`
	PerLevelPPM        int64           `json:"per_level_ppm"`
	Spent              int64           `json:"spent"`
	UnlockPPM          int64           `json:"unlock_ppm"`
}

type reputationRow struct {
	BodyKey  string   `json:"body_key"`
	Cost     int64    `json:"cost"`
	Kind     string   `json:"kind"`
	NodeID   string   `json:"node_id"`
	Requires []string `json:"requires"`
	State    string   `json:"state"`
	TitleKey string   `json:"title_key"`
}

func projectReputation(bundle production.CatalogBundle, founder *save.State, contributions []multiplier.Contribution) (*reputationArm, error) {
	tree := bundle.ReputationTree
	if tree == nil || save.VersionForState(founder) < 22 {
		return nil, nil
	}
	available, err := reputation.Available(founder.ReputationLevel, founder.ReputationSpent)
	if err != nil {
		return nil, ErrInvalidProjection
	}
	next, err := tree.BonusFactor(founder.ReputationLevel, founder.ReputationSpent, founder.ReputationUnlockPPM)
	if err != nil {
		return nil, ErrInvalidProjection
	}
	arm := &reputationArm{Available: available, BonusFactorNextRun: next.String(), Level: founder.ReputationLevel,
		PerLevelPPM: tree.Bonus.PerLevelPPM, Spent: founder.ReputationSpent, UnlockPPM: founder.ReputationUnlockPPM, Nodes: []reputationRow{}}
	for _, contribution := range contributions {
		if contribution.SourceID == tree.Bonus.SourceID {
			value := contribution.Factor.String()
			arm.BonusFactorThisRun = &value
		}
	}
	owned := map[string]bool{}
	for _, id := range founder.ReputationNodesOwned {
		owned[id] = true
	}
	for _, node := range tree.Nodes() {
		state := "available"
		switch {
		case owned[node.NodeID]:
			state = "owned"
		default:
			for _, requirement := range node.Requires {
				if !owned[requirement] {
					state = "locked"
				}
			}
			if state == "available" && node.Cost > available {
				state = "unaffordable"
			}
		}
		arm.Nodes = append(arm.Nodes, reputationRow{BodyKey: node.BodyKey, Cost: node.Cost, Kind: node.Kind, NodeID: node.NodeID,
			Requires: node.Requires, State: state, TitleKey: node.TitleKey})
	}
	return arm, nil
}
