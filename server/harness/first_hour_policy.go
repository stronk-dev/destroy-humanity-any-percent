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
	"slices"
	"sort"
)

const FirstHourPolicyPath = "balance/testdata/t0-t1/first-hour-policy-v1.json"

type FirstHourPolicyRegistry struct {
	SchemaVersion    int                         `json:"schema_version"`
	Policies         []FirstHourPersonaPolicy    `json:"policies"`
	GateCrossing     string                      `json:"gate_crossing"`
	Run2Continuation string                      `json:"run_2_continuation"`
	ElectiveExit     FirstHourElectiveExitPolicy `json:"elective_exit"`
	ScriptedFailure  FirstHourFailurePolicy      `json:"scripted_failure"`
	Hash             string                      `json:"-"`
}

type FirstHourPersonaPolicy struct {
	PolicyID        string `json:"policy_id"`
	PolicyVersion   int    `json:"policy_version"`
	ActionCadenceMS int64  `json:"action_cadence_ms"`
	Selection       string `json:"selection"`
	TopK            int64  `json:"top_k"`
}

type FirstHourElectiveExitPolicy struct {
	MinimumFounderAttendedMS   int64  `json:"minimum_founder_attended_ms"`
	RequiresGateID             string `json:"requires_gate_id"`
	RequiresAnyPersistentValue bool   `json:"requires_any_persistent_value"`
}

type FirstHourFailurePolicy struct {
	MinimumRunAttendedMS            int64  `json:"minimum_run_attended_ms"`
	RequiresGateID                  string `json:"requires_gate_id"`
	AcquihireMinGeneratorPurchases  int64  `json:"acquihire_min_generator_purchases"`
	BurnoutCashToCheapestUnownedPPM int64  `json:"burnout_cash_to_cheapest_unowned_ppm"`
	BurnoutRouteKnowledgeBonus      int64  `json:"burnout_route_knowledge_bonus"`
}

func LoadFirstHourPolicyRegistry(repositoryRoot string) (FirstHourPolicyRegistry, []byte, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(FirstHourPolicyPath)))
	if err != nil {
		return FirstHourPolicyRegistry{}, nil, err
	}
	registry, err := DecodeFirstHourPolicyRegistry(data)
	return registry, data, err
}

func DecodeFirstHourPolicyRegistry(data []byte) (FirstHourPolicyRegistry, error) {
	var registry FirstHourPolicyRegistry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return FirstHourPolicyRegistry{}, fmt.Errorf("first-hour policy: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return FirstHourPolicyRegistry{}, errors.New("first-hour policy must contain exactly one JSON value")
	}
	if registry.SchemaVersion != 1 || registry.GateCrossing != "immediate_when_reachable" ||
		registry.Run2Continuation != "same_policy" || len(registry.Policies) != 3 {
		return FirstHourPolicyRegistry{}, errors.New("invalid first-hour policy envelope")
	}
	wantIDs := []string{"casual.t0_t1", "chaos.t0_t1", "reference.greedy"}
	seen := make(map[string]bool, len(registry.Policies))
	for _, policy := range registry.Policies {
		if policy.PolicyVersion != 1 || policy.ActionCadenceMS < 1 || policy.TopK < 1 || seen[policy.PolicyID] {
			return FirstHourPolicyRegistry{}, fmt.Errorf("invalid first-hour persona %q", policy.PolicyID)
		}
		seen[policy.PolicyID] = true
		switch policy.Selection {
		case "cheapest_affordable", "projected_time":
			if policy.TopK != 1 {
				return FirstHourPolicyRegistry{}, fmt.Errorf("first-hour persona %q must use top_k 1", policy.PolicyID)
			}
		case "seeded_projected_top_k":
			if policy.TopK < 2 {
				return FirstHourPolicyRegistry{}, fmt.Errorf("first-hour persona %q requires top_k >= 2", policy.PolicyID)
			}
		default:
			return FirstHourPolicyRegistry{}, fmt.Errorf("unknown first-hour selection %q", policy.Selection)
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if !slices.Equal(ids, wantIDs) {
		return FirstHourPolicyRegistry{}, fmt.Errorf("first-hour persona set=%v", ids)
	}
	elective := registry.ElectiveExit
	failure := registry.ScriptedFailure
	if elective.MinimumFounderAttendedMS < 1 || elective.RequiresGateID == "" || !elective.RequiresAnyPersistentValue ||
		failure.MinimumRunAttendedMS < 1 || failure.RequiresGateID == "" || failure.AcquihireMinGeneratorPurchases < 1 ||
		failure.BurnoutCashToCheapestUnownedPPM < 1 || failure.BurnoutCashToCheapestUnownedPPM > 1_000_000 ||
		failure.BurnoutRouteKnowledgeBonus < 0 || failure.BurnoutRouteKnowledgeBonus%2 != 0 {
		return FirstHourPolicyRegistry{}, errors.New("invalid first-hour terminal policy")
	}
	digest := sha256.Sum256(data)
	registry.Hash = "sha256:" + hex.EncodeToString(digest[:])
	return registry, nil
}

func (registry FirstHourPolicyRegistry) Policy(id string, version int) (FirstHourPersonaPolicy, bool) {
	for _, policy := range registry.Policies {
		if policy.PolicyID == id && policy.PolicyVersion == version {
			return policy, true
		}
	}
	return FirstHourPersonaPolicy{}, false
}
