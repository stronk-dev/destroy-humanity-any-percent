package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

const RelevanceReportSchemaVersion = 1

type RelevanceRunSpec struct {
	PolicyID  string `json:"policy_id"`
	SeedStart string `json:"seed_start"`
	SeedCount int64  `json:"seed_count"`
	Reference bool   `json:"reference"`
}

type RelevanceMilestone struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	ResourceID string `json:"resource_id"`
	Amount     string `json:"amount"`
	Required   bool   `json:"required"`
}

type RelevanceSegment struct {
	MilestoneID string  `json:"milestone_id"`
	FromGate    string  `json:"from_gate"`
	ToGate      *string `json:"to_gate"`
}

type RelevanceScenario struct {
	SchemaVersion                 int                `json:"schema_version"`
	ID                            string             `json:"id"`
	Catalog                       string             `json:"catalog"`
	RoutesCatalog                 string             `json:"routes_catalog"`
	Policy                        string             `json:"relevance_policy"`
	Runs                          []RelevanceRunSpec `json:"runs"`
	Milestone                     RelevanceMilestone `json:"optimization_milestone"`
	Segments                      []RelevanceSegment `json:"segments"`
	Reducer                       string             `json:"reducer"`
	HorizonMS                     int64              `json:"horizon_ms"`
	DecisionHorizonsMS            []int64            `json:"decision_horizons_ms"`
	MaxDecisions                  int64              `json:"max_decisions"`
	BeamWidth                     int64              `json:"beam_width"`
	GreedyGapMaximumPPM           int64              `json:"greedy_gap_maximum_ppm"`
	RelevanceBudgetMaxRuns        int64              `json:"relevance_budget_max_runs"`
	RelevanceBudgetMaxTransitions int64              `json:"relevance_budget_max_transitions"`
}

type RelevanceDelta struct {
	MilestoneID string `json:"milestone_id"`
	Status      string `json:"status"`
	BaselineMS  *int64 `json:"baseline_ms"`
	AblatedMS   *int64 `json:"ablated_ms"`
	DeltaMS     *int64 `json:"delta_ms"`
}

type RelevanceRunBudget struct {
	DeclaredRuns        int64 `json:"declared_runs"`
	ExecutedRuns        int64 `json:"executed_runs"`
	DeclaredTransitions int64 `json:"declared_transitions"`
	ExecutedTransitions int64 `json:"executed_transitions"`
}

type RelevanceGreedyOracle struct {
	MilestoneID string `json:"milestone_id"`
	GreedyMS    int64  `json:"greedy_ms"`
	BeamMS      int64  `json:"beam_ms"`
	GapPPM      int64  `json:"gap_ppm"`
	MaximumPPM  int64  `json:"maximum_ppm"`
	Passed      bool   `json:"passed"`
}

type RelevanceItemReport struct {
	PurchasableID           string           `json:"purchasable_id"`
	AvailabilityWindow      RelevanceWindow  `json:"availability_window"`
	EpsilonMS               int64            `json:"epsilon_ms"`
	TrapExempt              bool             `json:"trap_exempt"`
	JustificationKey        *string          `json:"justification_key"`
	BaselinePurchaseCount   int64            `json:"baseline_purchase_count"`
	IndividualDeltas        []RelevanceDelta `json:"individual_deltas"`
	ActionRemovalDeltas     []RelevanceDelta `json:"action_removal_deltas"`
	Support                 string           `json:"support"`
	SupportingGroupID       *string          `json:"supporting_group_id"`
	NearestPassingEpsilonMS int64            `json:"nearest_passing_epsilon_ms"`
	RelevancePassed         bool             `json:"relevance_passed"`
	TrapPassed              bool             `json:"trap_passed"`
}

type RelevanceGroupReport struct {
	GroupID string           `json:"group_id"`
	Axis    string           `json:"axis"`
	Deltas  []RelevanceDelta `json:"deltas"`
	Passed  bool             `json:"passed"`
}

type RelevanceTierContribution struct {
	GroupID string           `json:"group_id"`
	Deltas  []RelevanceDelta `json:"deltas"`
}

type RelevanceReport struct {
	SchemaVersion       int                         `json:"schema_version"`
	ScenarioID          string                      `json:"scenario_id"`
	ScenarioHash        string                      `json:"scenario_hash"`
	ConstantsHash       string                      `json:"constants_hash"`
	RelevancePolicyHash string                      `json:"relevance_policy_hash"`
	RunBudget           RelevanceRunBudget          `json:"run_budget"`
	GreedyOracle        *RelevanceGreedyOracle      `json:"greedy_oracle"`
	Items               []RelevanceItemReport       `json:"items"`
	Groups              []RelevanceGroupReport      `json:"groups"`
	TierContributions   []RelevanceTierContribution `json:"tier_contributions"`
	RoleActivations     []RoleActivationCount       `json:"role_activations"`
	Failures            []string                    `json:"failures"`
}

type RelevanceSuite struct {
	Scenario      RelevanceScenario
	ScenarioHash  string
	Catalog       *economy.Catalog
	Routes        *routes.Catalog
	Policy        *RelevancePolicy
	ConstantsHash string
}

type relevanceRunResult struct {
	MilestoneMS    *int64
	Purchases      map[string]int64
	Roles          map[string]RoleActivationCount
	Transitions    int64
	FinalState     *save.State
	FinalVirtualMS int64
}

func LoadRelevanceSuite(repositoryRoot, scenarioPath string) (*RelevanceSuite, error) {
	scenarioBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenarioPath)))
	if err != nil {
		return nil, err
	}
	var scenario RelevanceScenario
	if err := decodeRelevanceStrict(scenarioBytes, &scenario); err != nil {
		return nil, fmt.Errorf("relevance scenario: %w", err)
	}
	if err := validateRelevanceScenario(scenario); err != nil {
		return nil, err
	}
	catalogBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenario.Catalog)))
	if err != nil {
		return nil, err
	}
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		return nil, err
	}
	routesBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenario.RoutesCatalog)))
	if err != nil {
		return nil, err
	}
	routeCatalog, err := routes.LoadCatalog(routesBytes)
	if err != nil {
		return nil, err
	}
	policyBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenario.Policy)))
	if err != nil {
		return nil, err
	}
	policy, err := LoadRelevancePolicy(policyBytes, catalog, routeCatalog)
	if err != nil {
		return nil, err
	}
	if err := validateRelevanceSegments(scenario, routeCatalog); err != nil {
		return nil, err
	}
	constantsHash, err := save.ConstantsHashArtifacts(map[string][]byte{"economy": catalogBytes, "relevance_policy": policyBytes, "routes": routesBytes})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(scenarioBytes)
	return &RelevanceSuite{Scenario: scenario, ScenarioHash: "sha256:" + hex.EncodeToString(digest[:]), Catalog: catalog,
		Routes: routeCatalog, Policy: policy, ConstantsHash: constantsHash}, nil
}

func validateRelevanceScenario(scenario RelevanceScenario) error {
	if scenario.SchemaVersion != 1 || !relevanceIDPattern.MatchString(scenario.ID) || scenario.Catalog == "" || scenario.RoutesCatalog == "" || scenario.Policy == "" ||
		len(scenario.Runs) == 0 || scenario.Reducer != "worst" && scenario.Reducer != "p05" || scenario.HorizonMS < 1 ||
		scenario.HorizonMS > relevanceMaxSafeInteger || scenario.MaxDecisions < 1 || scenario.BeamWidth != 8 || scenario.GreedyGapMaximumPPM < 0 ||
		scenario.GreedyGapMaximumPPM > 1_000_000 || scenario.RelevanceBudgetMaxRuns < 1 || scenario.RelevanceBudgetMaxTransitions < 1 {
		return errors.New("invalid relevance scenario envelope")
	}
	if scenario.Milestone.Kind != "resource_at_least" || !scenario.Milestone.Required || !relevanceIDPattern.MatchString(scenario.Milestone.ID) ||
		!relevanceIDPattern.MatchString(scenario.Milestone.ResourceID) {
		return errors.New("relevance scenario requires one resource_at_least optimization milestone")
	}
	target, err := decimal.ParseCanonical(scenario.Milestone.Amount)
	if err != nil || !target.Gt(decimal.Zero) {
		return errors.New("invalid relevance milestone amount")
	}
	prior := int64(-1)
	for _, horizon := range scenario.DecisionHorizonsMS {
		if horizon <= prior || horizon > scenario.HorizonMS {
			return errors.New("decision horizons must be sorted unique and within horizon")
		}
		prior = horizon
	}
	if len(scenario.DecisionHorizonsMS) == 0 || scenario.DecisionHorizonsMS[len(scenario.DecisionHorizonsMS)-1] != scenario.HorizonMS {
		return errors.New("decision horizons must end at scenario horizon")
	}
	referenceCount := 0
	for _, run := range scenario.Runs {
		if !relevanceIDPattern.MatchString(run.PolicyID) || run.SeedCount < 1 || run.SeedCount > relevanceMaxSafeInteger {
			return errors.New("invalid relevance run")
		}
		if _, err := strconv.ParseUint(run.SeedStart, 10, 64); err != nil {
			return errors.New("invalid relevance seed")
		}
		if run.Reference {
			referenceCount++
			if run.PolicyID != "reference.greedy" {
				return errors.New("reference run must use reference.greedy")
			}
		}
	}
	if referenceCount != 1 {
		return errors.New("relevance scenario requires exactly one reference run spec")
	}
	return nil
}

func validateRelevanceSegments(scenario RelevanceScenario, catalog *routes.Catalog) error {
	if len(scenario.Segments) != 1 || scenario.Segments[0].MilestoneID != scenario.Milestone.ID {
		return errors.New("relevance segments must completely bind the optimization milestone")
	}
	order := map[string]int{}
	for index, gate := range catalog.Gates() {
		order[gate.ID] = index
	}
	segment := scenario.Segments[0]
	from, ok := order[segment.FromGate]
	if !ok {
		return errors.New("relevance segment references unknown from_gate")
	}
	if segment.ToGate != nil {
		to, ok := order[*segment.ToGate]
		if !ok || from >= to {
			return errors.New("relevance segment has invalid to_gate")
		}
	}
	return nil
}

func ComputeRelevanceRunBudget(nonReferenceSeeds, referenceSeeds, items, groups int64, beam bool) (int64, error) {
	if nonReferenceSeeds < 0 || referenceSeeds < 1 || items < 1 || groups < 0 {
		return 0, errors.New("invalid relevance run budget inputs")
	}
	checkedMul := func(left, right int64) (int64, error) {
		if left != 0 && right > relevanceMaxSafeInteger/left {
			return 0, errors.New("relevance run budget overflow")
		}
		return left * right, nil
	}
	base, err := checkedMul(nonReferenceSeeds+referenceSeeds, 1+items+groups)
	if err != nil {
		return 0, err
	}
	removals, err := checkedMul(referenceSeeds, items)
	if err != nil || base > relevanceMaxSafeInteger-removals {
		return 0, errors.New("relevance run budget overflow")
	}
	total := base + removals
	if beam {
		if total > relevanceMaxSafeInteger-referenceSeeds {
			return 0, errors.New("relevance run budget overflow")
		}
		total += referenceSeeds
	}
	return total, nil
}

func MakeRelevanceDelta(milestoneID string, baseline, ablated *int64, horizonMS int64) RelevanceDelta {
	result := RelevanceDelta{MilestoneID: milestoneID, BaselineMS: cloneInt64(baseline), AblatedMS: cloneInt64(ablated)}
	switch {
	case baseline != nil && ablated != nil:
		delta := *ablated - *baseline
		result.Status, result.DeltaMS = "both_reached", &delta
	case baseline != nil:
		delta := horizonMS - *baseline
		result.Status, result.DeltaMS = "ablated_unreached", &delta
	case ablated != nil:
		result.Status = "baseline_unreached"
	default:
		result.Status = "both_unreached"
	}
	return result
}

func ValidateRelevanceReport(report RelevanceReport) error {
	if report.SchemaVersion != RelevanceReportSchemaVersion || !relevanceIDPattern.MatchString(report.ScenarioID) ||
		len(report.ScenarioHash) != 71 || len(report.ConstantsHash) != 71 || len(report.RelevancePolicyHash) != 71 ||
		report.Items == nil || report.Groups == nil || report.TierContributions == nil || report.RoleActivations == nil || report.Failures == nil ||
		report.RunBudget.DeclaredRuns < 1 || report.RunBudget.DeclaredRuns != report.RunBudget.ExecutedRuns ||
		report.RunBudget.DeclaredTransitions != report.RunBudget.ExecutedTransitions {
		return errors.New("invalid relevance report envelope")
	}
	prior := ""
	for _, item := range report.Items {
		if !relevanceIDPattern.MatchString(item.PurchasableID) || prior != "" && prior >= item.PurchasableID || item.IndividualDeltas == nil ||
			item.ActionRemovalDeltas == nil || item.Support != "individual" && item.Support != "group_supported" && item.Support != "failed" ||
			item.SupportingGroupID != nil != (item.Support == "group_supported") {
			return fmt.Errorf("invalid relevance item report %q", item.PurchasableID)
		}
		for _, rows := range [][]RelevanceDelta{item.IndividualDeltas, item.ActionRemovalDeltas} {
			for _, row := range rows {
				if err := validateRelevanceDelta(row); err != nil {
					return err
				}
			}
		}
		prior = item.PurchasableID
	}
	prior = ""
	for _, group := range report.Groups {
		if !relevanceIDPattern.MatchString(group.GroupID) || prior != "" && prior >= group.GroupID || group.Deltas == nil {
			return fmt.Errorf("invalid relevance group report %q", group.GroupID)
		}
		for _, row := range group.Deltas {
			if err := validateRelevanceDelta(row); err != nil {
				return err
			}
		}
		prior = group.GroupID
	}
	for index, failure := range report.Failures {
		if failure == "" || index > 0 && report.Failures[index-1] >= failure {
			return errors.New("relevance failures must be sorted unique")
		}
	}
	return nil
}

func validateRelevanceDelta(row RelevanceDelta) error {
	if !relevanceIDPattern.MatchString(row.MilestoneID) {
		return errors.New("invalid relevance delta milestone")
	}
	valid := false
	switch row.Status {
	case "both_reached":
		valid = row.BaselineMS != nil && row.AblatedMS != nil && row.DeltaMS != nil
	case "ablated_unreached":
		valid = row.BaselineMS != nil && row.AblatedMS == nil && row.DeltaMS != nil
	case "baseline_unreached":
		valid = row.BaselineMS == nil && row.AblatedMS != nil && row.DeltaMS == nil
	case "both_unreached":
		valid = row.BaselineMS == nil && row.AblatedMS == nil && row.DeltaMS == nil
	}
	if !valid {
		return errors.New("invalid relevance delta union")
	}
	return nil
}

func reduceRelevanceDeltas(rows []RelevanceDelta, reducer string) (RelevanceDelta, error) {
	if len(rows) == 0 || reducer != "worst" && reducer != "p05" {
		return RelevanceDelta{}, errors.New("invalid relevance delta reduction")
	}
	valid := make([]RelevanceDelta, 0, len(rows))
	for _, row := range rows {
		if row.DeltaMS != nil {
			valid = append(valid, row)
		}
	}
	if len(valid) == 0 {
		return rows[0], nil
	}
	sort.Slice(valid, func(left, right int) bool { return *valid[left].DeltaMS < *valid[right].DeltaMS })
	index := 0
	if reducer == "p05" {
		index = (len(valid) - 1) * 5 / 100
	}
	return valid[index], nil
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func ceilDecimalRatio(numerator, denominator decimal.Decimal) (int64, error) {
	if !numerator.IsStateValue() || !denominator.IsStateValue() || numerator.Lt(decimal.Zero) || !denominator.Gt(decimal.Zero) {
		return 0, errors.New("invalid relevance ratio")
	}
	numeratorCoefficient, numeratorExponent := decimalCoefficient(numerator)
	denominatorCoefficient, denominatorExponent := decimalCoefficient(denominator)
	ratio := new(big.Rat).Quo(numeratorCoefficient, denominatorCoefficient)
	exponent := numeratorExponent - denominatorExponent
	if exponent > 16 {
		return 0, errors.New("relevance ratio outside exact integer domain")
	}
	if exponent < -20 {
		return 1, nil
	}
	power := new(big.Int).Exp(big.NewInt(10), big.NewInt(absInt64(exponent)), nil)
	if exponent >= 0 {
		ratio.Mul(ratio, new(big.Rat).SetInt(power))
	} else {
		ratio.Quo(ratio, new(big.Rat).SetInt(power))
	}
	quotient, remainder := new(big.Int).QuoRem(ratio.Num(), ratio.Denom(), new(big.Int))
	if remainder.Sign() != 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsInt64() || quotient.Sign() < 0 || quotient.Int64() > relevanceMaxSafeInteger {
		return 0, errors.New("relevance ratio outside exact integer domain")
	}
	return quotient.Int64(), nil
}

func decimalCoefficient(value decimal.Decimal) (*big.Rat, int64) {
	canonical := value.String()
	parts := bytes.Split([]byte(canonical), []byte("e"))
	coefficient := new(big.Rat)
	coefficient.SetString(string(parts[0]))
	exponent, _ := strconv.ParseInt(string(parts[1]), 10, 64)
	return coefficient, exponent
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func relevanceNow(offsetMS int64) time.Time {
	return Epoch.Add(time.Duration(offsetMS) * time.Millisecond)
}
