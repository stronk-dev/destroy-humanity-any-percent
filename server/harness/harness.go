package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

const SchemaVersion = 1

var Epoch = time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC)

type Scenario struct {
	SchemaVersion      int         `json:"schema_version"`
	ID                 string      `json:"id"`
	Version            int         `json:"version"`
	Catalog            string      `json:"catalog"`
	RoutesCatalog      string      `json:"routes_catalog"`
	CommonsCatalog     string      `json:"commons_catalog"`
	Runs               []RunSpec   `json:"runs"`
	Milestones         []Milestone `json:"milestones"`
	Envelopes          []Envelope  `json:"envelopes"`
	RequiredInvariants []string    `json:"required_invariants"`
}

type RunSpec struct {
	PolicyID      string `json:"policy_id"`
	PolicyVersion int    `json:"policy_version"`
	SeedStart     string `json:"seed_start"`
	SeedCount     int    `json:"seed_count"`
	HorizonMS     int64  `json:"horizon_ms"`
}

type Milestone struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	MustReach  bool   `json:"must_reach"`
	IntentKind string `json:"intent_kind,omitempty"`
	EventKind  string `json:"event_kind,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Amount     string `json:"amount,omitempty"`
	Generator  string `json:"generator_id,omitempty"`
	Count      int64  `json:"count,omitempty"`
	Tier       int    `json:"tier,omitempty"`
}

type Envelope struct {
	PolicyID  string `json:"policy_id"`
	Milestone string `json:"milestone_id"`
	Statistic string `json:"statistic"`
	MinimumMS *int64 `json:"minimum_ms,omitempty"`
	MaximumMS *int64 `json:"maximum_ms,omitempty"`
}

type RunKey struct {
	HarnessSchemaVersion int    `json:"harness_schema_version"`
	ScenarioID           string `json:"scenario_id"`
	ScenarioVersion      int    `json:"scenario_version"`
	ScenarioHash         string `json:"scenario_hash"`
	PolicyID             string `json:"policy_id"`
	PolicyVersion        int    `json:"policy_version"`
	Seed                 string `json:"seed"`
	ConstantsHash        string `json:"constants_hash"`
}

type TimedMilestone struct {
	ID      string `json:"id"`
	FirstMS *int64 `json:"first_ms"`
}

type NamedCount struct {
	ID    string `json:"id"`
	Count int64  `json:"count"`
}

type RoleActivationCount struct {
	GeneratorID string                    `json:"generator_id"`
	Kind        economy.GeneratorRoleKind `json:"kind"`
	TargetID    string                    `json:"target_id"`
	Count       int64                     `json:"count"`
}

type ResourceAmount struct {
	ResourceID string `json:"resource_id"`
	Amount     string `json:"amount"`
}

type RunReport struct {
	Key                  RunKey                `json:"key"`
	Outcome              string                `json:"outcome"`
	FinalVirtualMS       int64                 `json:"final_virtual_ms"`
	FinalStateHash       string                `json:"final_state_hash"`
	Milestones           []TimedMilestone      `json:"milestones"`
	Applied              []NamedCount          `json:"applied"`
	Rejected             []NamedCount          `json:"rejected"`
	RoleActivations      []RoleActivationCount `json:"role_activations,omitempty"`
	SourceTotals         []ResourceAmount      `json:"source_totals"`
	SinkTotals           []ResourceAmount      `json:"sink_totals"`
	FinalBalances        []ResourceAmount      `json:"final_balances"`
	MaximumProgressGapMS int64                 `json:"maximum_progress_gap_ms"`
	InvariantFailures    []string              `json:"invariant_failures"`
}

type AggregateValue struct {
	PolicyID  string `json:"policy_id"`
	Milestone string `json:"milestone_id"`
	Statistic string `json:"statistic"`
	ValueMS   int64  `json:"value_ms"`
}

type AggregateReport struct {
	SchemaVersion int              `json:"schema_version"`
	ScenarioID    string           `json:"scenario_id"`
	ScenarioHash  string           `json:"scenario_hash"`
	ConstantsHash string           `json:"constants_hash"`
	RunCount      int              `json:"run_count"`
	Values        []AggregateValue `json:"values"`
	Warnings      []string         `json:"warnings"`
	Failures      []string         `json:"failures"`
}

type GoldenReport struct {
	SchemaVersion int         `json:"schema_version"`
	Runs          []RunReport `json:"runs"`
}

type SuiteReport struct {
	SchemaVersion int             `json:"schema_version"`
	Runs          []RunReport     `json:"runs"`
	Aggregate     AggregateReport `json:"aggregate"`
}

type Suite struct {
	Scenario            Scenario
	ScenarioBytes       []byte
	Catalog             *economy.Catalog
	CatalogBytes        []byte
	RoutesCatalog       *routes.Catalog
	RoutesCatalogBytes  []byte
	CommonsCatalog      *commons.Catalog
	CommonsCatalogBytes []byte
	ScenarioHash        string
	ConstantsHash       string
}

type runTask struct {
	spec RunSpec
	seed uint64
	key  RunKey
}

func LoadSuite(repositoryRoot, scenarioPath string) (*Suite, error) {
	scenarioBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenarioPath)))
	if err != nil {
		return nil, err
	}
	var scenario Scenario
	decoder := json.NewDecoder(bytes.NewReader(scenarioBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scenario); err != nil {
		return nil, fmt.Errorf("scenario: %w", err)
	}
	if scenario.SchemaVersion != 1 || scenario.ID == "" || scenario.Version < 1 || scenario.Catalog == "" || scenario.RoutesCatalog == "" || scenario.CommonsCatalog == "" || len(scenario.Runs) == 0 {
		return nil, errors.New("invalid scenario envelope")
	}
	knownInvariants := map[string]bool{"state_encodes": true, "numeric_domain": true, "resource_bounds": true,
		"ledger_reconciles": true, "revision_monotone": true, "must_reach": true}
	seenInvariants := make(map[string]bool)
	for _, invariant := range scenario.RequiredInvariants {
		if !knownInvariants[invariant] || seenInvariants[invariant] {
			return nil, fmt.Errorf("invalid required invariant %q", invariant)
		}
		seenInvariants[invariant] = true
	}
	if len(seenInvariants) != len(knownInvariants) {
		return nil, errors.New("scenario must require the complete v1 invariant registry")
	}
	if err := validateMilestones(scenario.Milestones); err != nil {
		return nil, err
	}
	if err := validateObservationMatrix(scenario.Runs, scenario.Milestones, scenario.Envelopes); err != nil {
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
	commonsCatalogBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenario.CommonsCatalog)))
	if err != nil {
		return nil, err
	}
	commonsCatalog, err := commons.LoadCatalog(commonsCatalogBytes)
	if err != nil {
		return nil, err
	}
	routesCatalogBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenario.RoutesCatalog)))
	if err != nil {
		return nil, err
	}
	routesCatalog, err := routes.LoadCatalog(routesCatalogBytes)
	if err != nil {
		return nil, err
	}
	bundle, err := epochseed.Load(repositoryRoot)
	if err != nil {
		return nil, err
	}
	for name, scenarioPath := range map[string]string{"commons": scenario.CommonsCatalog, "economy": scenario.Catalog, "routes": scenario.RoutesCatalog} {
		manifestPath, ok := epochseed.ArtifactPath(bundle.Seed, name)
		if !ok || manifestPath != scenarioPath {
			return nil, fmt.Errorf("scenario %s path %q differs from epoch manifest %q", name, scenarioPath, manifestPath)
		}
	}
	scenarioDigest := sha256.Sum256(scenarioBytes)
	return &Suite{Scenario: scenario, ScenarioBytes: scenarioBytes, Catalog: catalog, CatalogBytes: catalogBytes,
		RoutesCatalog: routesCatalog, RoutesCatalogBytes: routesCatalogBytes,
		CommonsCatalog: commonsCatalog, CommonsCatalogBytes: commonsCatalogBytes,
		ScenarioHash: "sha256:" + hex.EncodeToString(scenarioDigest[:]), ConstantsHash: bundle.Hash}, nil
}

func (suite *Suite) RunAll() ([]RunReport, AggregateReport, error) {
	var tasks []runTask
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil || spec.SeedCount < 1 || spec.HorizonMS < 1 || uint64(spec.SeedCount-1) > ^uint64(0)-start {
			return nil, AggregateReport{}, errors.New("invalid run specification")
		}
		for offset := 0; offset < spec.SeedCount; offset++ {
			seed := start + uint64(offset)
			tasks = append(tasks, runTask{spec: spec, seed: seed, key: suite.runKey(spec, seed)})
		}
	}
	reports, err := dispatchRunTasks(tasks, 4, func(task runTask) RunReport {
		return suite.run(task.spec, task.seed)
	})
	if err != nil {
		return nil, AggregateReport{}, err
	}
	sort.Slice(reports, func(left, right int) bool { return lessKey(reports[left].Key, reports[right].Key) })
	aggregate := suite.aggregate(reports)
	return reports, aggregate, nil
}

func dispatchRunTasks(tasks []runTask, workerLimit int, execute func(runTask) RunReport) ([]RunReport, error) {
	if len(tasks) == 0 || workerLimit < 1 || execute == nil {
		return nil, errors.New("invalid harness task dispatch")
	}
	reports := make([]RunReport, len(tasks))
	var wait sync.WaitGroup
	work := make(chan int)
	workers := workerLimit
	if len(tasks) < workers {
		workers = len(tasks)
	}
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for index := range work {
				reports[index] = execute(tasks[index])
			}
		}()
	}
	for index := range tasks {
		work <- index
	}
	close(work)
	wait.Wait()
	for index := range tasks {
		if reports[index].Key != tasks[index].key {
			return nil, fmt.Errorf("harness task %d returned mismatched run key", index)
		}
	}
	return reports, nil
}

func (suite *Suite) run(spec RunSpec, seed uint64) RunReport {
	key := suite.runKey(spec, seed)
	report := RunReport{Key: key, Outcome: "completed", FinalVirtualMS: spec.HorizonMS, InvariantFailures: []string{}}
	ledger, err := economy.NewLedger(suite.Catalog, economy.ScopeCompany)
	if err != nil {
		return failed(report, err)
	}
	counts := make(map[string]int64)
	provisioned := make(map[string]int64)
	remainders := make(map[string]int64)
	for _, generator := range suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID] = 0
		provisioned[generator.ID] = 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	state := &save.State{Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{}, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, EvaluatedThrough: Epoch, RunStartedAt: Epoch, RunSeq: 1,
		ManualTokenMilli: suite.Catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: Epoch,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}}
	revision := int64(1)
	random := NewSplitMix64(seed)
	uuids := NewUUIDStream(seed)
	milestones := make(map[string]*int64)
	applied, rejected := map[string]int64{}, map[string]int64{}
	roleActivations := map[string]RoleActivationCount{}
	sources, sinks := map[string][]decimal.Decimal{}, map[string][]decimal.Decimal{}
	times := actionTimes(spec.PolicyID, spec.HorizonMS)
	if times == nil || (spec.PolicyID == "casual.phase0" && spec.PolicyVersion != 1) ||
		(spec.PolicyID == "chaos.phase0" && spec.PolicyVersion != 1) {
		return failed(report, fmt.Errorf("unknown policy %s v%d", spec.PolicyID, spec.PolicyVersion))
	}
	previousSession := int64(-1)
	for _, offsetMS := range times {
		now := Epoch.Add(time.Duration(offsetMS) * time.Millisecond)
		mode := production.ModeOnline
		if spec.PolicyID == "casual.phase0" {
			session := casualSession(offsetMS)
			if previousSession >= 0 && session != previousSession {
				mode = production.ModeOffline
			}
			previousSession = session
		}
		intentBytes, kind, intentErr := suite.intentBytes(spec.PolicyID, state, revision, now, random, uuids)
		if intentErr != nil {
			return failed(report, intentErr)
		}
		request, err := production.ParseIntent(intentBytes)
		if err != nil {
			return failed(report, err)
		}
		candidate, err := cloneState(suite.Catalog, state)
		if err != nil {
			return failed(report, err)
		}
		beforeBalances := state.Ledger.Snapshot()
		beforeRevision := revision
		simulation, err := production.SimulateTransition(request, candidate, suite.Catalog, production.SimulationDependencies{Routes: suite.RoutesCatalog}, save.Revision{Number: revision,
			ConstantsHash: suite.ConstantsHash}, mode, now, nil, nil, production.AblationMask{})
		if err != nil {
			return failed(report, err)
		}
		decision := simulation.Decision
		if decision.Outcome == save.IntentApplied {
			state = candidate
			revision++
			applied[kind]++
			for _, activation := range simulation.RoleActivations {
				key := activation.GeneratorID + "\x00" + string(activation.Kind) + "\x00" + activation.TargetID
				entry := roleActivations[key]
				entry.GeneratorID, entry.Kind, entry.TargetID, entry.Count = activation.GeneratorID, activation.Kind, activation.TargetID, entry.Count+1
				roleActivations[key] = entry
			}
			if revision != beforeRevision+1 {
				report.InvariantFailures = append(report.InvariantFailures, "revision_monotone")
				break
			}
			if err := collectReceipt(decision.Receipt, beforeBalances, state.Ledger.Snapshot(), sources, sinks); err != nil {
				report.InvariantFailures = append(report.InvariantFailures, "ledger_reconciles:"+err.Error())
				break
			}
			if _, err := save.EncodeState(state); err != nil {
				report.InvariantFailures = append(report.InvariantFailures, "state_encodes:"+err.Error())
				break
			}
			if err := validateStateDomain(suite.Catalog, state); err != nil {
				report.InvariantFailures = append(report.InvariantFailures, err.Error())
				break
			}
			suite.observeMilestones(milestones, offsetMS, kind, decision, state)
		} else {
			rejected[rejectionCategory(decision.Receipt)]++
			if revision != beforeRevision {
				report.InvariantFailures = append(report.InvariantFailures, "revision_monotone")
				break
			}
		}
	}
	report.Milestones = milestoneReport(suite.Scenario.Milestones, milestones, &report.InvariantFailures)
	report.Applied = namedCounts(applied)
	report.Rejected = namedCounts(rejected)
	report.RoleActivations = sortedRoleActivations(roleActivations)
	report.SourceTotals = resourceTotals(sources)
	report.SinkTotals = resourceTotals(sinks)
	report.FinalBalances = balancesReport(state.Ledger.Snapshot())
	report.MaximumProgressGapMS = maximumGap(report.Milestones)
	encoded, err := save.EncodeState(state)
	if err != nil {
		report.InvariantFailures = append(report.InvariantFailures, "state_encodes:"+err.Error())
	} else {
		digest := sha256.Sum256(encoded)
		report.FinalStateHash = "sha256:" + hex.EncodeToString(digest[:])
	}
	if len(report.InvariantFailures) > 0 {
		report.Outcome = "failed"
	}
	return report
}

func (suite *Suite) runKey(spec RunSpec, seed uint64) RunKey {
	return RunKey{HarnessSchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioVersion: suite.Scenario.Version,
		ScenarioHash: suite.ScenarioHash, PolicyID: spec.PolicyID, PolicyVersion: spec.PolicyVersion,
		Seed: strconv.FormatUint(seed, 10), ConstantsHash: suite.ConstantsHash}
}

func (suite *Suite) intentBytes(policy string, state *save.State, revision int64, now time.Time, random *SplitMix64, uuids *UUIDStream) ([]byte, string, error) {
	id, err := uuids.Next(now.UnixMilli())
	if err != nil {
		return nil, "", err
	}
	manual := suite.Catalog.ManualActions()
	sort.Slice(manual, func(i, j int) bool { return manual[i].ID < manual[j].ID })
	generators := suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany)
	sort.Slice(generators, func(i, j int) bool { return generators[i].ID < generators[j].ID })
	if policy == "casual.phase0" {
		chosenID, chosenCost := "", decimal.NaN
		for _, generator := range generators {
			cash, _ := state.Ledger.Balance(generator.Price.ResourceID)
			cost, quoteErr := suite.Catalog.BulkCost(generator.ID, state.GeneratorCounts[generator.ID], 1)
			if quoteErr == nil && cost.Lte(cash) && (chosenID == "" || cost.Lt(chosenCost) || cost.Eq(chosenCost) && generator.ID < chosenID) {
				chosenID, chosenCost = generator.ID, cost
			}
		}
		if chosenID != "" {
			return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"buy_generator","expected_revision":%d,"generator_id":%q,"count":{"mode":"exact","value":1}}`, id, revision, chosenID)), production.IntentBuyGenerator, nil
		}
		return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"perform_manual_batch","expected_revision":%d,"action_id":%q,"count":1,"window_ms":1000}`, id, revision, manual[0].ID)), production.IntentPerformManualBatch, nil
	}
	candidateCount := len(manual) + len(generators)
	choice := int(random.Bound(uint64(candidateCount)))
	if choice < len(manual) {
		count := int64(random.Bound(80) + 1)
		return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"perform_manual_batch","expected_revision":%d,"action_id":%q,"count":%d,"window_ms":300000}`, id, revision, manual[choice].ID, count)), production.IntentPerformManualBatch, nil
	}
	generator := generators[choice-len(manual)]
	if random.Bound(2) == 0 {
		return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"buy_generator","expected_revision":%d,"generator_id":%q,"count":{"mode":"max"}}`, id, revision, generator.ID)), production.IntentBuyGenerator, nil
	}
	count := int64(random.Bound(80) + 1)
	return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":"buy_generator","expected_revision":%d,"generator_id":%q,"count":{"mode":"exact","value":%d}}`, id, revision, generator.ID, count)), production.IntentBuyGenerator, nil
}

func actionTimes(policy string, horizon int64) []int64 {
	var result []int64
	if policy == "chaos.phase0" {
		for at := int64(300_000); at <= horizon; at += 300_000 {
			result = append(result, at)
		}
		return result
	}
	if policy != "casual.phase0" {
		return nil
	}
	for day := int64(0); day*86_400_000 <= horizon; day++ {
		for _, start := range []int64{0, 18_000_000, 39_600_000} {
			for within := int64(0); within < 480_000; within += 1_000 {
				at := day*86_400_000 + start + within
				if at <= horizon {
					result = append(result, at)
				}
			}
		}
	}
	return result
}

func casualSession(offset int64) int64 {
	day, within := offset/86_400_000, offset%86_400_000
	index := int64(0)
	if within >= 39_600_000 {
		index = 2
	} else if within >= 18_000_000 {
		index = 1
	}
	return day*3 + index
}

func (suite *Suite) observeMilestones(found map[string]*int64, at int64, intentKind string, decision save.IntentDecision, state *save.State) {
	for _, milestone := range suite.Scenario.Milestones {
		if found[milestone.ID] != nil {
			continue
		}
		reached := false
		switch milestone.Kind {
		case "intent_applied":
			reached = milestone.IntentKind == intentKind
		case "event_seen":
			for _, event := range decision.Events {
				reached = reached || string(event.Kind) == milestone.EventKind
			}
		case "resource_at_least":
			value, ok := state.Ledger.Balance(milestone.ResourceID)
			target, err := decimal.ParseCanonical(milestone.Amount)
			reached = ok && err == nil && value.Gte(target)
		case "generator_count_at_least":
			reached = state.GeneratorCounts[milestone.Generator] >= milestone.Count
		case "progress_at_least":
			value, err := production.SubProgressValue(suite.Catalog, state, milestone.Tier)
			target, parseErr := decimal.ParseCanonical(milestone.Amount)
			reached = err == nil && parseErr == nil && value.Gte(target)
		}
		if reached {
			copy := at
			found[milestone.ID] = &copy
		}
	}
}

func validateMilestones(milestones []Milestone) error {
	if len(milestones) == 0 {
		return errors.New("scenario must define milestones")
	}
	seen := make(map[string]bool, len(milestones))
	for _, milestone := range milestones {
		if milestone.ID == "" || seen[milestone.ID] {
			return fmt.Errorf("invalid or duplicate milestone id %q", milestone.ID)
		}
		seen[milestone.ID] = true
		switch milestone.Kind {
		case "intent_applied":
			if milestone.IntentKind == "" {
				return fmt.Errorf("milestone %q requires intent_kind", milestone.ID)
			}
		case "event_seen":
			if milestone.EventKind == "" {
				return fmt.Errorf("milestone %q requires event_kind", milestone.ID)
			}
		case "resource_at_least":
			amount, err := decimal.ParseCanonical(milestone.Amount)
			if milestone.ResourceID == "" || err != nil || !amount.Gt(decimal.Zero) {
				return fmt.Errorf("milestone %q has invalid resource target", milestone.ID)
			}
		case "generator_count_at_least":
			if milestone.Generator == "" || milestone.Count < 1 {
				return fmt.Errorf("milestone %q has invalid generator target", milestone.ID)
			}
		case "progress_at_least":
			amount, err := decimal.ParseCanonical(milestone.Amount)
			if milestone.Tier < 0 || err != nil || !amount.Gt(decimal.Zero) {
				return fmt.Errorf("milestone %q has invalid progress target", milestone.ID)
			}
		default:
			return fmt.Errorf("unknown milestone kind %q", milestone.Kind)
		}
	}
	return nil
}

func validateObservationMatrix(runs []RunSpec, milestones []Milestone, envelopes []Envelope) error {
	policies := make(map[string]bool, len(runs))
	for _, run := range runs {
		if run.PolicyID == "" {
			return errors.New("scenario run requires policy_id")
		}
		policies[run.PolicyID] = true
	}
	milestoneIDs := make(map[string]bool, len(milestones))
	for _, milestone := range milestones {
		milestoneIDs[milestone.ID] = true
	}
	statistics := map[string]bool{"best": true, "p05": true, "p50": true, "p95": true, "worst": true}
	seen := make(map[string]bool, len(envelopes))
	for _, envelope := range envelopes {
		if !policies[envelope.PolicyID] {
			return fmt.Errorf("envelope references unknown policy %q", envelope.PolicyID)
		}
		if !milestoneIDs[envelope.Milestone] {
			return fmt.Errorf("envelope references unknown milestone %q", envelope.Milestone)
		}
		if !statistics[envelope.Statistic] {
			return fmt.Errorf("envelope uses unknown statistic %q", envelope.Statistic)
		}
		if envelope.MinimumMS != nil && *envelope.MinimumMS < 0 || envelope.MaximumMS != nil && *envelope.MaximumMS < 0 ||
			envelope.MinimumMS != nil && envelope.MaximumMS != nil && *envelope.MinimumMS > *envelope.MaximumMS {
			return fmt.Errorf("envelope %s/%s/%s has invalid bounds", envelope.PolicyID, envelope.Milestone, envelope.Statistic)
		}
		key := envelope.PolicyID + "\x00" + envelope.Milestone + "\x00" + envelope.Statistic
		if seen[key] {
			return fmt.Errorf("duplicate envelope %s/%s/%s", envelope.PolicyID, envelope.Milestone, envelope.Statistic)
		}
		seen[key] = true
	}
	for policy := range policies {
		for milestone := range milestoneIDs {
			for _, statistic := range []string{"p50", "p95"} {
				key := policy + "\x00" + milestone + "\x00" + statistic
				if !seen[key] {
					return fmt.Errorf("missing pacing observation %s/%s/%s", policy, milestone, statistic)
				}
			}
		}
	}
	return nil
}

func (suite *Suite) aggregate(reports []RunReport) AggregateReport {
	aggregate := AggregateReport{SchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash,
		ConstantsHash: suite.ConstantsHash, RunCount: len(reports), Warnings: []string{}, Failures: []string{}}
	for _, report := range reports {
		for _, failure := range report.InvariantFailures {
			aggregate.Failures = append(aggregate.Failures, formatRunKey(report.Key)+":"+failure)
		}
	}
	for _, envelope := range suite.Scenario.Envelopes {
		var values []int64
		for _, report := range reports {
			if report.Key.PolicyID != envelope.PolicyID {
				continue
			}
			for _, milestone := range report.Milestones {
				if milestone.ID == envelope.Milestone && milestone.FirstMS != nil {
					values = append(values, *milestone.FirstMS)
				}
			}
		}
		if len(values) == 0 {
			aggregate.Failures = append(aggregate.Failures, "envelope has no reached values: "+envelope.Milestone)
			continue
		}
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		value := statistic(values, envelope.Statistic)
		aggregate.Values = append(aggregate.Values, AggregateValue{PolicyID: envelope.PolicyID,
			Milestone: envelope.Milestone, Statistic: envelope.Statistic, ValueMS: value})
		if envelope.MinimumMS != nil && value < *envelope.MinimumMS || envelope.MaximumMS != nil && value > *envelope.MaximumMS {
			aggregate.Failures = append(aggregate.Failures, fmt.Sprintf("envelope %s/%s/%s=%d outside bounds", envelope.PolicyID, envelope.Milestone, envelope.Statistic, value))
		}
	}
	return aggregate
}

func formatRunKey(key RunKey) string {
	return fmt.Sprintf("schema=%d/scenario=%s@%d/scenario_hash=%s/policy=%s@%d/seed=%s/constants_hash=%s",
		key.HarnessSchemaVersion, key.ScenarioID, key.ScenarioVersion, key.ScenarioHash,
		key.PolicyID, key.PolicyVersion, key.Seed, key.ConstantsHash)
}

func statistic(values []int64, name string) int64 {
	switch name {
	case "best":
		return values[0]
	case "worst":
		return values[len(values)-1]
	case "p05":
		return values[(5*len(values)+99)/100-1]
	case "p50":
		return values[(50*len(values)+99)/100-1]
	case "p95":
		return values[(95*len(values)+99)/100-1]
	default:
		return -1
	}
}

func CompareBaseline(current, baseline AggregateReport) (warnings, failures []string) {
	baselineByKey := make(map[string]int64)
	for _, value := range baseline.Values {
		baselineByKey[value.PolicyID+"\x00"+value.Milestone+"\x00"+value.Statistic] = value.ValueMS
	}
	for _, value := range current.Values {
		key := value.PolicyID + "\x00" + value.Milestone + "\x00" + value.Statistic
		prior, ok := baselineByKey[key]
		if !ok {
			failures = append(failures, "baseline missing "+key)
			continue
		}
		delta := value.ValueMS - prior
		if delta < 0 {
			delta = -delta
		}
		denominator := prior
		if denominator < 1 {
			denominator = 1
		}
		if delta*100 > denominator*25 {
			failures = append(failures, "drift over 25%: "+key)
		} else if delta*100 > denominator*10 {
			warnings = append(warnings, "drift over 10%: "+key)
		}
		delete(baselineByKey, key)
	}
	for key := range baselineByKey {
		failures = append(failures, "current report missing baseline key "+key)
	}
	sort.Strings(failures)
	return warnings, failures
}

func CanonicalJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func cloneState(catalog *economy.Catalog, source *save.State) (*save.State, error) {
	encoded, err := save.EncodeState(source)
	if err != nil {
		return nil, err
	}
	return save.RestoreState(encoded, save.CurrentVersion, catalog, economy.ScopeCompany, time.Time{})
}

func collectReceipt(data []byte, before, after map[string]string, sources, sinks map[string][]decimal.Decimal) error {
	var receipt struct {
		Receipt struct {
			Changes []struct {
				ResourceID string `json:"resource_id"`
				Before     string `json:"before"`
				Delta      string `json:"delta"`
				After      string `json:"after"`
			} `json:"changes"`
		} `json:"receipt"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, change := range receipt.Receipt.Changes {
		if seen[change.ResourceID] || before[change.ResourceID] == "" || after[change.ResourceID] == "" ||
			change.Before != before[change.ResourceID] || change.After != after[change.ResourceID] {
			return fmt.Errorf("invalid change resource %q", change.ResourceID)
		}
		seen[change.ResourceID] = true
		value, err := decimal.ParseCanonical(change.Delta)
		if err != nil {
			return err
		}
		if value.Lt(decimal.Zero) {
			sinks[change.ResourceID] = append(sinks[change.ResourceID], value.Abs())
		} else {
			sources[change.ResourceID] = append(sources[change.ResourceID], value)
		}
	}
	for id, beforeValue := range before {
		if beforeValue != after[id] && !seen[id] {
			return fmt.Errorf("changed resource %q missing from receipt", id)
		}
	}
	return nil
}

func validateStateDomain(catalog *economy.Catalog, state *save.State) error {
	for id, encoded := range state.Ledger.Snapshot() {
		value, err := decimal.ParseCanonical(encoded)
		if err != nil || !value.IsStateValue() {
			return fmt.Errorf("numeric_domain:%s", id)
		}
		definition, ok := catalog.Resource(id)
		if !ok || value.Lt(definition.Minimum) || definition.Hardcap != nil && value.Gt(definition.Hardcap.Amount) {
			return fmt.Errorf("resource_bounds:%s", id)
		}
	}
	return nil
}

func rejectionCategory(data []byte) string {
	var receipt struct {
		Rejection struct {
			Category string `json:"category"`
		} `json:"rejection"`
	}
	if json.Unmarshal(data, &receipt) != nil || receipt.Rejection.Category == "" {
		return "invalid_receipt"
	}
	return receipt.Rejection.Category
}

func failed(report RunReport, err error) RunReport {
	report.Outcome = "failed"
	report.Milestones = []TimedMilestone{}
	report.Applied = []NamedCount{}
	report.Rejected = []NamedCount{}
	report.SourceTotals = []ResourceAmount{}
	report.SinkTotals = []ResourceAmount{}
	report.FinalBalances = []ResourceAmount{}
	report.InvariantFailures = []string{err.Error()}
	return report
}

func lessKey(left, right RunKey) bool {
	leftParts := []string{left.ScenarioID, strconv.Itoa(left.ScenarioVersion), left.ScenarioHash, left.PolicyID,
		strconv.Itoa(left.PolicyVersion), left.Seed, left.ConstantsHash}
	rightParts := []string{right.ScenarioID, strconv.Itoa(right.ScenarioVersion), right.ScenarioHash, right.PolicyID,
		strconv.Itoa(right.PolicyVersion), right.Seed, right.ConstantsHash}
	for index := range leftParts {
		if leftParts[index] != rightParts[index] {
			return leftParts[index] < rightParts[index]
		}
	}
	return false
}

func namedCounts(source map[string]int64) []NamedCount {
	ids := make([]string, 0, len(source))
	for id := range source {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]NamedCount, 0, len(ids))
	for _, id := range ids {
		result = append(result, NamedCount{ID: id, Count: source[id]})
	}
	return result
}

func sortedRoleActivations(source map[string]RoleActivationCount) []RoleActivationCount {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]RoleActivationCount, 0, len(keys))
	for _, key := range keys {
		result = append(result, source[key])
	}
	return result
}

func resourceTotals(source map[string][]decimal.Decimal) []ResourceAmount {
	ids := make([]string, 0, len(source))
	for id := range source {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]ResourceAmount, 0, len(ids))
	for _, id := range ids {
		result = append(result, ResourceAmount{ResourceID: id, Amount: decimal.SumDeterministic(source[id]).Quantize(decimal.CanonicalSignificantDigits).String()})
	}
	return result
}

func balancesReport(source map[string]string) []ResourceAmount {
	ids := make([]string, 0, len(source))
	for id := range source {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]ResourceAmount, 0, len(ids))
	for _, id := range ids {
		result = append(result, ResourceAmount{ResourceID: id, Amount: source[id]})
	}
	return result
}

func milestoneReport(definitions []Milestone, found map[string]*int64, failures *[]string) []TimedMilestone {
	result := make([]TimedMilestone, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, TimedMilestone{ID: definition.ID, FirstMS: found[definition.ID]})
		if definition.MustReach && found[definition.ID] == nil {
			*failures = append(*failures, "must_reach:"+definition.ID)
		}
	}
	return result
}

func maximumGap(milestones []TimedMilestone) int64 {
	var values []int64
	for _, milestone := range milestones {
		if milestone.FirstMS != nil {
			values = append(values, *milestone.FirstMS)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	maximum, prior := int64(0), int64(0)
	for _, value := range values {
		if value-prior > maximum {
			maximum = value - prior
		}
		prior = value
	}
	return maximum
}
