package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
)

const (
	RatifiedFirstHourScenarioHash = "sha256:8798abb885db89bde52349d3de87cb6687381cb5f2c33bbd92e561c07cb2029c"
	RatifiedFirstHourPolicyHash   = "sha256:e5e5de7051beb0340e54f7013ce7d4a48c35bfcc3343220310290478445d10c3"
)

type FirstHourScenario struct {
	SchemaVersion      int                  `json:"schema_version"`
	ID                 string               `json:"id"`
	Version            int                  `json:"version"`
	EpochSeed          string               `json:"epoch_seed"`
	Runs               []RunSpec            `json:"runs"`
	Milestones         []FirstHourMilestone `json:"milestones"`
	Envelopes          []Envelope           `json:"envelopes"`
	Relations          []FirstHourRelation  `json:"relations"`
	RequiredInvariants []string             `json:"required_invariants"`
	TransitionBudget   int64                `json:"transition_budget"`
}

type FirstHourMilestone struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	IntentKind  string `json:"intent_kind,omitempty"`
	EventKind   string `json:"event_kind,omitempty"`
	GateID      string `json:"gate_id,omitempty"`
	ExitType    string `json:"exit_type,omitempty"`
	ExitTypeNot string `json:"exit_type_not,omitempty"`
	RunSeq      int64  `json:"run_seq,omitempty"`
	Clock       string `json:"clock"`
	MustReach   bool   `json:"must_reach"`
}

type FirstHourRelation struct {
	PolicyID       string `json:"policy_id"`
	Kind           string `json:"kind"`
	LeftMilestone  string `json:"left_milestone_id"`
	RightMilestone string `json:"right_milestone_id"`
	SameSeed       bool   `json:"same_seed"`
}

type FirstHourSuite struct {
	Scenario      FirstHourScenario
	ScenarioBytes []byte
	Policy        FirstHourPolicyRegistry
	PolicyBytes   []byte
	ScenarioHash  string
	PolicyHash    string
	ConstantsHash string
	Bundle        production.CatalogBundle
}

func LoadFirstHourSuite(repositoryRoot, scenarioPath, policyPath string) (*FirstHourSuite, error) {
	scenario, scenarioBytes, err := loadFirstHourScenario(repositoryRoot, scenarioPath)
	if err != nil {
		return nil, err
	}
	policy, policyBytes, policyHash, err := LoadFirstHourPolicy(repositoryRoot, policyPath)
	if err != nil {
		return nil, err
	}
	if scenario.EpochSeed != epochseed.Path {
		return nil, fmt.Errorf("first-hour scenario epoch seed %q differs from authority %q", scenario.EpochSeed, epochseed.Path)
	}
	scenarioDigest := sha256.Sum256(scenarioBytes)
	if "sha256:"+hex.EncodeToString(scenarioDigest[:]) != RatifiedFirstHourScenarioHash || policyHash != RatifiedFirstHourPolicyHash {
		return nil, errors.New("first-hour scenario/policy bytes are not owner-ratified")
	}
	epoch, err := epochseed.Load(repositoryRoot)
	if err != nil {
		return nil, err
	}
	bundle, err := replaycatalog.Load(epoch.Hash, epoch.Artifacts)
	if err != nil {
		return nil, err
	}
	for _, run := range scenario.Runs {
		if _, ok := policy.Policy(run.PolicyID, run.PolicyVersion); !ok {
			return nil, fmt.Errorf("scenario references unknown first-hour policy %s v%d", run.PolicyID, run.PolicyVersion)
		}
	}
	identity := sha256.New()
	identity.Write([]byte("first_hour_scenario.v1\x1f"))
	identity.Write(scenarioBytes)
	identity.Write([]byte{0x1f})
	identity.Write(policyBytes)
	return &FirstHourSuite{Scenario: scenario, ScenarioBytes: scenarioBytes, Policy: policy, PolicyBytes: policyBytes,
		ScenarioHash: "sha256:" + hex.EncodeToString(identity.Sum(nil)), PolicyHash: policyHash,
		ConstantsHash: epoch.Hash, Bundle: bundle}, nil
}

func loadFirstHourScenario(repositoryRoot, scenarioPath string) (FirstHourScenario, []byte, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenarioPath)))
	if err != nil {
		return FirstHourScenario{}, nil, err
	}
	var scenario FirstHourScenario
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scenario); err != nil {
		return FirstHourScenario{}, nil, fmt.Errorf("first-hour scenario: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return FirstHourScenario{}, nil, errors.New("first-hour scenario must contain exactly one JSON value")
	}
	if err := validateFirstHourScenario(scenario); err != nil {
		return FirstHourScenario{}, nil, err
	}
	return scenario, data, nil
}

func validateFirstHourScenario(scenario FirstHourScenario) error {
	if scenario.SchemaVersion != 1 || scenario.ID != "scenario.t0_t1_first_hour" || scenario.Version != 1 ||
		scenario.EpochSeed == "" || len(scenario.Runs) != 3 || len(scenario.Milestones) != 7 ||
		len(scenario.Envelopes) != 7 || len(scenario.Relations) != 1 || scenario.TransitionBudget != 2_000_000 {
		return errors.New("invalid first-hour scenario envelope")
	}
	knownPolicies := map[string]int{"chaos.t0_t1": 64, "casual.t0_t1": 32, "reference.greedy": 1}
	seenPolicies := map[string]bool{}
	for _, run := range scenario.Runs {
		seed, err := strconv.ParseUint(run.SeedStart, 10, 64)
		if err != nil || seed != 0 || run.PolicyVersion != 1 || knownPolicies[run.PolicyID] != run.SeedCount || run.HorizonMS != 7_200_000 || seenPolicies[run.PolicyID] {
			return fmt.Errorf("invalid first-hour run %q", run.PolicyID)
		}
		seenPolicies[run.PolicyID] = true
	}
	if len(seenPolicies) != len(knownPolicies) {
		return errors.New("first-hour scenario omits a required policy")
	}
	knownKinds := map[string]bool{"intent_applied": true, "event_seen": true, "gate_crossed": true, "run_ended": true}
	milestones := map[string]FirstHourMilestone{}
	for _, milestone := range scenario.Milestones {
		if milestone.ID == "" || milestones[milestone.ID].ID != "" || !knownKinds[milestone.Kind] || !milestone.MustReach ||
			(milestone.Clock != "run_attended_ms" && milestone.Clock != "founder_attended_ms") {
			return fmt.Errorf("invalid first-hour milestone %q", milestone.ID)
		}
		switch milestone.Kind {
		case "intent_applied":
			if milestone.IntentKind == "" {
				return fmt.Errorf("first-hour milestone %q omits intent kind", milestone.ID)
			}
		case "event_seen":
			if milestone.EventKind == "" {
				return fmt.Errorf("first-hour milestone %q omits event kind", milestone.ID)
			}
		case "gate_crossed":
			if milestone.GateID == "" || milestone.RunSeq < 1 {
				return fmt.Errorf("first-hour milestone %q has invalid gate coordinate", milestone.ID)
			}
		case "run_ended":
			if (milestone.ExitType == "") == (milestone.ExitTypeNot == "") {
				return fmt.Errorf("first-hour milestone %q must bind exactly one exit selector", milestone.ID)
			}
		}
		milestones[milestone.ID] = milestone
	}
	for _, envelope := range scenario.Envelopes {
		if !seenPolicies[envelope.PolicyID] || milestones[envelope.Milestone].ID == "" ||
			(envelope.Statistic != "p50" && envelope.Statistic != "p95") || envelope.MinimumMS == nil && envelope.MaximumMS == nil {
			return fmt.Errorf("invalid first-hour envelope %s/%s", envelope.PolicyID, envelope.Milestone)
		}
	}
	relation := scenario.Relations[0]
	if relation.PolicyID != "chaos.t0_t1" || relation.Kind != "less_than" || !relation.SameSeed ||
		milestones[relation.LeftMilestone].ID == "" || milestones[relation.RightMilestone].ID == "" {
		return errors.New("invalid first-hour milestone relation")
	}
	wantInvariants := []string{"artifact_identity", "ledger_reconciles", "must_reach", "numeric_domain", "replay_parity", "resource_bounds", "revision_monotone", "role_activation", "state_encodes"}
	gotInvariants := append([]string(nil), scenario.RequiredInvariants...)
	sort.Strings(gotInvariants)
	if !equalStrings(gotInvariants, wantInvariants) {
		return errors.New("first-hour scenario must require the complete invariant set")
	}
	return nil
}

func (suite *FirstHourSuite) RunKey(spec RunSpec, seed uint64) RunKey {
	return RunKey{HarnessSchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioVersion: suite.Scenario.Version,
		ScenarioHash: suite.ScenarioHash, PolicyID: spec.PolicyID, PolicyVersion: spec.PolicyVersion,
		Seed: strconv.FormatUint(seed, 10), ConstantsHash: suite.ConstantsHash}
}

func (suite *FirstHourSuite) GateIDs() []string {
	seen := map[string]bool{}
	for _, milestone := range suite.Scenario.Milestones {
		if milestone.Kind == "gate_crossed" {
			seen[milestone.GateID] = true
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func FirstHourScenarioIdentity(scenarioBytes, policyBytes []byte) string {
	hash := sha256.New()
	hash.Write([]byte("first_hour_scenario.v1\x1f"))
	hash.Write(scenarioBytes)
	hash.Write([]byte{0x1f})
	hash.Write(policyBytes)
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func firstHourMilestoneByID(scenario FirstHourScenario, id string) (FirstHourMilestone, bool) {
	for _, milestone := range scenario.Milestones {
		if milestone.ID == id {
			return milestone, true
		}
	}
	return FirstHourMilestone{}, false
}

func firstHourPolicyIDs(scenario FirstHourScenario) string {
	ids := make([]string, 0, len(scenario.Runs))
	for _, run := range scenario.Runs {
		ids = append(ids, run.PolicyID)
	}
	sort.Strings(ids)
	return strings.Join(ids, ",")
}
