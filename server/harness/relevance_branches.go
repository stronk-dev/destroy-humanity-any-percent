package harness

import (
	"errors"
	"fmt"
	"sort"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
)

const RelevanceBranchReportSchemaVersion = 1

// RelevanceBranchProof is a candidate-owned proof for an upgrade that the
// unmasked reference does not buy. The prefix is produced from genesis by the
// same ranked policy as the reference, constrained only to the declared
// generator branch and the one upgrade under proof. Both completions start
// from the exact real-engine purchase state; the masked twin disables only the
// purchased upgrade's effect.
type RelevanceBranchProof struct {
	UpgradeID           string   `json:"upgrade_id"`
	EffectTargetID      string   `json:"effect_target_id"`
	RemovedGeneratorIDs []string `json:"removed_generator_ids"`
	RemovedUpgradeIDs   []string `json:"removed_upgrade_ids"`
	SelectedAtMS        *int64   `json:"selected_at_ms"`
	BaselineMS          *int64   `json:"baseline_ms"`
	MaskedMS            *int64   `json:"masked_ms"`
	DeltaMS             *int64   `json:"delta_ms"`
	EpsilonMS           int64    `json:"epsilon_ms"`
	Passed              bool     `json:"passed"`
}

type RelevanceBranchReport struct {
	SchemaVersion       int                    `json:"schema_version"`
	ScenarioID          string                 `json:"scenario_id"`
	ScenarioHash        string                 `json:"scenario_hash"`
	ConstantsHash       string                 `json:"constants_hash"`
	RelevancePolicyHash string                 `json:"relevance_policy_hash"`
	MainReferenceMS     int64                  `json:"main_reference_ms"`
	ExecutedTransitions int64                  `json:"executed_transitions"`
	Proofs              []RelevanceBranchProof `json:"proofs"`
	Failures            []string               `json:"failures"`
}

func (suite *RelevanceSuite) RunUpgradeBranchProofs() (RelevanceBranchReport, error) {
	if suite == nil || suite.Catalog == nil || suite.Policy == nil || suite.Routes == nil {
		return RelevanceBranchReport{}, errors.New("upgrade branch proof requires a complete relevance suite")
	}
	items, _, err := suite.measuredPolicyRows()
	if err != nil {
		return RelevanceBranchReport{}, err
	}
	report := RelevanceBranchReport{SchemaVersion: RelevanceBranchReportSchemaVersion,
		ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.ConstantsHash,
		RelevancePolicyHash: suite.Policy.Hash, Proofs: []RelevanceBranchProof{}, Failures: []string{}}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	opportunityMask, _, err := suite.opportunityAwareMask(production.AblationMask{}, counter)
	if err != nil {
		return RelevanceBranchReport{}, fmt.Errorf("upgrade branch opportunity preflight: %w", err)
	}
	main, err := suite.runReferenceWithOpportunity(production.AblationMask{}, opportunityMask, counter)
	if err != nil {
		return RelevanceBranchReport{}, fmt.Errorf("upgrade branch main reference: %w", err)
	}
	if main.DecisionStarved || main.MilestoneMS == nil {
		return RelevanceBranchReport{}, errors.New("upgrade branch main reference did not reach the milestone")
	}
	report.MainReferenceMS = *main.MilestoneMS

	for _, item := range items {
		upgrade, ok := suite.Catalog.Upgrade(item.PurchasableID)
		if !ok || main.Purchases[item.PurchasableID] > 0 {
			continue
		}
		proof, proofErr := suite.runUpgradeBranchProof(upgrade, item.EpsilonMS, counter)
		if proofErr != nil {
			return RelevanceBranchReport{}, fmt.Errorf("upgrade branch %s: %w", upgrade.ID, proofErr)
		}
		report.Proofs = append(report.Proofs, proof)
		if !proof.Passed {
			kind := "branch_floor"
			if proof.SelectedAtMS == nil {
				kind = "branch_unselected"
			} else if proof.BaselineMS == nil || proof.MaskedMS == nil {
				kind = "branch_unreached"
			}
			report.Failures = append(report.Failures, kind+":"+proof.UpgradeID)
		}
	}
	report.ExecutedTransitions = counter.value
	report.Failures = sortedUniqueStrings(report.Failures)
	if err := ValidateRelevanceBranchReport(report); err != nil {
		return RelevanceBranchReport{}, err
	}
	return report, nil
}

func (suite *RelevanceSuite) runUpgradeBranchProof(upgrade economy.UpgradeDefinition, epsilonMS int64,
	counter *relevanceCounter) (RelevanceBranchProof, error) {
	targetID, ok := singleUpgradeEffectTarget(upgrade)
	proof := RelevanceBranchProof{UpgradeID: upgrade.ID, EffectTargetID: targetID, EpsilonMS: epsilonMS,
		RemovedGeneratorIDs: []string{}, RemovedUpgradeIDs: []string{}}
	if !ok {
		return proof, nil
	}
	branchMask := suite.upgradeBranchMask(upgrade.ID, targetID)
	proof.RemovedGeneratorIDs = append(proof.RemovedGeneratorIDs, branchMask.RemovedGeneratorIDs...)
	proof.RemovedUpgradeIDs = append(proof.RemovedUpgradeIDs, branchMask.RemovedUpgradeIDs...)
	traces := []relevanceDecisionTrace{}
	branch, err := suite.runReferenceWithOpportunityTrace(branchMask, production.AblationMask{}, counter, &traces)
	if err != nil {
		return proof, err
	}
	if branch.DecisionStarved {
		return proof, errReferenceDecisionStarved
	}
	var chosen *relevanceDecisionTrace
	for index := range traces {
		if traces[index].ReferenceArm == upgrade.ID && traces[index].ReferenceChoice != nil {
			chosen = &traces[index]
			break
		}
	}
	if chosen == nil {
		return proof, nil
	}
	selectedAt := chosen.ReferenceChoice.AtMS
	proof.SelectedAtMS = &selectedAt
	if !chosen.ReferenceChoice.State.UpgradesOwned[upgrade.ID] {
		return proof, errors.New("ranked branch choice did not apply the upgrade through the engine")
	}
	remaining := suite.Scenario.MaxDecisions - chosen.DecisionOrdinal - 1
	if remaining < 1 {
		return proof, nil
	}
	normalState, err := cloneState(suite.Catalog, chosen.ReferenceChoice.State)
	if err != nil {
		return proof, err
	}
	maskedState, err := cloneState(suite.Catalog, chosen.ReferenceChoice.State)
	if err != nil {
		return proof, err
	}
	normal, err := suite.runReferenceFromRanked(normalState, chosen.ReferenceChoice.Revision,
		selectedAt, remaining, branchMask, counter)
	if err != nil {
		return proof, err
	}
	masked := mergeAblationMasks(branchMask, production.AblationMask{UpgradeIDs: []string{upgrade.ID}})
	control, err := suite.runReferenceFromRanked(maskedState, chosen.ReferenceChoice.Revision,
		selectedAt, remaining, masked, counter)
	if err != nil {
		return proof, err
	}
	proof.BaselineMS, proof.MaskedMS = cloneInt64(normal.MilestoneMS), cloneInt64(control.MilestoneMS)
	if proof.BaselineMS != nil && proof.MaskedMS != nil {
		delta := *proof.MaskedMS - *proof.BaselineMS
		proof.DeltaMS = &delta
		proof.Passed = delta >= epsilonMS
	}
	return proof, nil
}

func singleUpgradeEffectTarget(upgrade economy.UpgradeDefinition) (string, bool) {
	if len(upgrade.Effects) == 0 {
		return "", false
	}
	target := upgrade.Effects[0].Target
	for _, effect := range upgrade.Effects[1:] {
		if effect.Target != target {
			return "", false
		}
	}
	return target, target != ""
}

func (suite *RelevanceSuite) upgradeBranchMask(upgradeID, targetID string) production.AblationMask {
	mask := production.AblationMask{RemovedGeneratorIDs: []string{}, RemovedUpgradeIDs: []string{}}
	if target, ok := suite.Catalog.GeneratorClass(targetID); ok {
		for _, generator := range suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany) {
			if generator.Tier > target.Tier || generator.Tier == target.Tier && generator.Price.Base.Gt(target.Price.Base) {
				mask.RemovedGeneratorIDs = append(mask.RemovedGeneratorIDs, generator.ID)
			}
		}
	}
	for _, upgrade := range suite.Catalog.Upgrades() {
		if upgrade.ID != upgradeID {
			mask.RemovedUpgradeIDs = append(mask.RemovedUpgradeIDs, upgrade.ID)
		}
	}
	sort.Strings(mask.RemovedGeneratorIDs)
	sort.Strings(mask.RemovedUpgradeIDs)
	return mask
}

func ValidateRelevanceBranchReport(report RelevanceBranchReport) error {
	if report.SchemaVersion != RelevanceBranchReportSchemaVersion || !relevanceIDPattern.MatchString(report.ScenarioID) ||
		!relevanceHashPattern.MatchString(report.ScenarioHash) || !relevanceHashPattern.MatchString(report.ConstantsHash) ||
		!relevanceHashPattern.MatchString(report.RelevancePolicyHash) || !relevanceSafePositive(report.MainReferenceMS) ||
		!relevanceSafe(report.ExecutedTransitions) || report.Proofs == nil || report.Failures == nil {
		return errors.New("invalid relevance branch report envelope")
	}
	prior := ""
	expectedFailures := []string{}
	for _, proof := range report.Proofs {
		if !relevanceIDPattern.MatchString(proof.UpgradeID) || !relevanceIDPattern.MatchString(proof.EffectTargetID) ||
			prior != "" && prior >= proof.UpgradeID || sortedUniqueIDs(proof.RemovedGeneratorIDs) != nil ||
			sortedUniqueIDs(proof.RemovedUpgradeIDs) != nil || !relevanceSafePositive(proof.EpsilonMS) {
			return fmt.Errorf("invalid relevance branch proof %q", proof.UpgradeID)
		}
		passed := proof.SelectedAtMS != nil && proof.BaselineMS != nil && proof.MaskedMS != nil && proof.DeltaMS != nil &&
			*proof.DeltaMS == *proof.MaskedMS-*proof.BaselineMS && *proof.DeltaMS >= proof.EpsilonMS
		if proof.Passed != passed {
			return fmt.Errorf("relevance branch proof %q does not reconcile", proof.UpgradeID)
		}
		for _, value := range []*int64{proof.SelectedAtMS, proof.BaselineMS, proof.MaskedMS} {
			if value != nil && !relevanceSafe(*value) {
				return fmt.Errorf("relevance branch proof %q time is unsafe", proof.UpgradeID)
			}
		}
		if !proof.Passed {
			kind := "branch_floor"
			if proof.SelectedAtMS == nil {
				kind = "branch_unselected"
			} else if proof.BaselineMS == nil || proof.MaskedMS == nil {
				kind = "branch_unreached"
			}
			expectedFailures = append(expectedFailures, kind+":"+proof.UpgradeID)
		}
		prior = proof.UpgradeID
	}
	if fmt.Sprint(sortedUniqueStrings(expectedFailures)) != fmt.Sprint(report.Failures) {
		return errors.New("relevance branch failures do not reconcile")
	}
	return nil
}
