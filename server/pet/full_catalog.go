package pet

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
)

var ErrInvalidCatalog = errors.New("invalid pet catalog")

type StatPolicyRow struct {
	StatID          StatID `json:"stat_id"`
	InitialPPM      int64  `json:"initial_ppm"`
	FloorPPM        int64  `json:"floor_ppm"`
	DecayPPMPerGrid int64  `json:"decay_ppm_per_grid"`
}

type StatPolicy struct {
	GridMS                  int64           `json:"grid_ms"`
	Stats                   []StatPolicyRow `json:"stats"`
	DiminishingThresholdPPM int64           `json:"diminishing_threshold_ppm"`
	DiminishingFactorPPM    int64           `json:"diminishing_factor_ppm"`
}

type ActionPolicy struct {
	ActionID           string `json:"action_id"`
	StatID             StatID `json:"stat_id"`
	DeltaPPM           int64  `json:"delta_ppm"`
	CooldownAttendedMS int64  `json:"cooldown_attended_ms"`
	MinEligiblePPM     int64  `json:"min_eligible_ppm"`
	SoulGate           string `json:"soul_gate"`
}

type TrustPolicy struct {
	InitialPPM                int64 `json:"initial_ppm"`
	NeutralPPM                int64 `json:"neutral_ppm"`
	FloorPPM                  int64 `json:"floor_ppm"`
	CapPPM                    int64 `json:"cap_ppm"`
	GainPPMPerEffectiveAction int64 `json:"gain_ppm_per_effective_action"`
	DecayPPMPerGrid           int64 `json:"decay_ppm_per_grid"`
}

type Catalog struct {
	SchemaVersion  int                 `json:"schema_version"`
	StatPolicy     StatPolicy          `json:"stat_policy"`
	Actions        []ActionPolicy      `json:"actions"`
	TrustPolicy    TrustPolicy         `json:"trust_policy"`
	MoodPolicy     []MoodThreshold     `json:"mood_policy"`
	BehaviorPolicy []BehaviorCandidate `json:"behavior_policy"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	if !uniqueStateJSONKeys(data) {
		return nil, ErrInvalidCatalog
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var catalog Catalog
	if decoder.Decode(&catalog) != nil || catalog.SchemaVersion != 1 && catalog.SchemaVersion != 2 || !exactFullCatalogKeys(data, catalog.SchemaVersion) {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || validateFullCatalog(&catalog) != nil {
		return nil, ErrInvalidCatalog
	}
	return &catalog, nil
}

func exactFullCatalogKeys(data []byte, schemaVersion int) bool {
	var root map[string]json.RawMessage
	if json.Unmarshal(data, &root) != nil || !hasRawKeys(root, []string{
		"schema_version", "stat_policy", "actions", "trust_policy", "mood_policy", "behavior_policy",
	}) {
		return false
	}
	var statPolicy, trustPolicy map[string]json.RawMessage
	if json.Unmarshal(root["stat_policy"], &statPolicy) != nil || !hasRawKeys(statPolicy, []string{
		"grid_ms", "stats", "diminishing_threshold_ppm", "diminishing_factor_ppm",
	}) || json.Unmarshal(root["trust_policy"], &trustPolicy) != nil || !hasRawKeys(trustPolicy, []string{
		"initial_ppm", "neutral_ppm", "floor_ppm", "cap_ppm", "gain_ppm_per_effective_action", "decay_ppm_per_grid",
	}) {
		return false
	}
	actionKeys := []string{"action_id", "stat_id", "delta_ppm", "cooldown_attended_ms", "min_eligible_ppm"}
	if schemaVersion >= 2 {
		actionKeys = append(actionKeys, "soul_gate")
	}
	return exactCatalogRows(statPolicy["stats"], []string{"stat_id", "initial_ppm", "floor_ppm", "decay_ppm_per_grid"}) &&
		exactCatalogRows(root["actions"], actionKeys) &&
		exactCatalogRows(root["mood_policy"], []string{"mood_member", "floor_ppm"}) &&
		exactCatalogRows(root["behavior_policy"], []string{"from_state", "event", "to_state", "duration_grid_ticks"})
}

func validateFullCatalog(catalog *Catalog) error {
	if catalog == nil || catalog.StatPolicy.GridMS < 1 || !exactNonnegative(catalog.StatPolicy.GridMS) ||
		len(catalog.StatPolicy.Stats) != len(statIDs) || catalog.StatPolicy.DiminishingThresholdPPM < 0 ||
		catalog.StatPolicy.DiminishingThresholdPPM > 1_000_000 || catalog.StatPolicy.DiminishingFactorPPM < 0 || catalog.StatPolicy.DiminishingFactorPPM > 1_000_000 ||
		catalog.Actions == nil || len(catalog.MoodPolicy) != len(moods) || catalog.BehaviorPolicy == nil {
		return ErrInvalidCatalog
	}
	seenStats := map[StatID]bool{}
	statFloors := map[StatID]int64{}
	for _, row := range catalog.StatPolicy.Stats {
		if !ValidStatID(row.StatID) || seenStats[row.StatID] || row.InitialPPM < row.FloorPPM || row.InitialPPM > 1_000_000 ||
			row.FloorPPM < 0 || row.DecayPPMPerGrid < 0 || row.DecayPPMPerGrid > 1_000_000 {
			return ErrInvalidCatalog
		}
		seenStats[row.StatID] = true
		statFloors[row.StatID] = row.FloorPPM
	}
	lastAction := ""
	for _, row := range catalog.Actions {
		if !mechanicalIDPattern.MatchString(row.ActionID) || row.ActionID <= lastAction || !ValidStatID(row.StatID) || row.DeltaPPM < 1 ||
			row.DeltaPPM > 1_000_000 || !exactNonnegative(row.CooldownAttendedMS) || row.MinEligiblePPM < 0 ||
			row.MinEligiblePPM > statFloors[row.StatID] || catalog.SchemaVersion == 1 && row.SoulGate != "" || catalog.SchemaVersion == 2 && !validSoulGate(row.SoulGate) {
			return ErrInvalidCatalog
		}
		lastAction = row.ActionID
	}
	trust := catalog.TrustPolicy
	if trust.FloorPPM < 0 || trust.FloorPPM > trust.NeutralPPM || trust.NeutralPPM > trust.InitialPPM || trust.InitialPPM > trust.CapPPM || trust.CapPPM > 1_000_000 ||
		trust.GainPPMPerEffectiveAction < 0 || trust.GainPPMPerEffectiveAction > 1_000_000 || trust.DecayPPMPerGrid < 0 || trust.DecayPPMPerGrid > 1_000_000 {
		return ErrInvalidCatalog
	}
	grammar := CatalogGrammar{MoodThresholds: catalog.MoodPolicy, BehaviorCandidates: catalog.BehaviorPolicy}
	if validateCatalogGrammar(grammar) != nil {
		return ErrInvalidCatalog
	}
	seenTransitions := map[[2]string]bool{}
	for _, row := range catalog.BehaviorPolicy {
		key := [2]string{string(row.FromState), string(row.Event)}
		if seenTransitions[key] {
			return ErrInvalidCatalog
		}
		seenTransitions[key] = true
	}
	return nil
}

func validSoulGate(value string) bool {
	return value == "essential" || value == "recovery" || value == "ordinary"
}

func (catalog *Catalog) SchemaSupportsSoul() bool {
	if catalog == nil || catalog.SchemaVersion < 2 {
		return false
	}
	for _, row := range catalog.Actions {
		if !validSoulGate(row.SoulGate) {
			return false
		}
	}
	return true
}

func (catalog *Catalog) StateDeclarations() StateDeclarations {
	if catalog == nil {
		return StateDeclarations{}
	}
	actions := make([]string, len(catalog.Actions))
	for index, row := range catalog.Actions {
		actions[index] = row.ActionID
	}
	behaviors := make([]string, len(behaviorStates))
	for index, value := range behaviorStates {
		behaviors[index] = string(value)
	}
	sort.Strings(behaviors)
	return StateDeclarations{ActionIDs: actions, BehaviorIDs: behaviors}
}
