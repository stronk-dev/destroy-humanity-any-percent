package harness

import (
	"errors"
	"fmt"
	"sort"

	"cloud-clicker/server/decimal"
	prestigecore "cloud-clicker/server/prestige"
)

// Reputation Tree v1 OD-2/H2: re-derive the Reputation paid at the first
// elective Exit under candidate thresholds from recorded lifetime values
// (H1). Runs 1–2 of the first-hour suite are threshold-independent (no tree
// node can be owned yet, so the Founder bonus is exactly 1), which makes this
// re-derivation exact rather than an approximation.

type ReputationThresholdEnvelope struct {
	Minimum int64 `json:"minimum"`
	Maximum int64 `json:"maximum"`
}

type ReputationThresholdPersona struct {
	PolicyID   string `json:"policy_id"`
	Runs       int    `json:"runs"`
	P05        int64  `json:"p05"`
	P50        int64  `json:"p50"`
	P95        int64  `json:"p95"`
	Min        int64  `json:"min"`
	Max        int64  `json:"max"`
	InEnvelope bool   `json:"p50_in_envelope"`
}

type ReputationThresholdRow struct {
	Threshold string                       `json:"threshold"`
	Personas  []ReputationThresholdPersona `json:"personas"`
	Satisfies bool                         `json:"satisfies_envelope"`
}

type ReputationThresholdReport struct {
	SchemaVersion   int                         `json:"schema_version"`
	SourceScenario  string                      `json:"source_scenario_hash"`
	SourceConstants string                      `json:"source_constants_hash"`
	Envelope        ReputationThresholdEnvelope `json:"envelope"`
	GatedPersonas   []string                    `json:"gated_personas"`
	Rows            []ReputationThresholdRow    `json:"rows"`
	Satisfying      []string                    `json:"satisfying_thresholds"`
	Note            string                      `json:"note"`
}

var ErrReputationMeasurement = errors.New("invalid reputation threshold measurement")

// PaidReputationAtFirstElectiveExit re-derives one run's first-elective-Exit
// payout: run 1's scripted_first credit sets the Founder level, and run 2's
// collapse pays against it.
func PaidReputationAtFirstElectiveExit(run FirstHourRunResult, policy *prestigecore.Policy, threshold decimal.Decimal) (int64, error) {
	var scripted, elective *FirstHourReputationSample
	for index := range run.ReputationExits {
		sample := &run.ReputationExits[index]
		switch {
		case sample.ExitType == "scripted_first" && scripted == nil:
			scripted = sample
		case sample.ExitType == "collapse" && elective == nil:
			elective = sample
		}
	}
	if scripted == nil || elective == nil || policy == nil {
		return 0, fmt.Errorf("%w: run %s lacks recorded scripted and elective Exits", ErrReputationMeasurement, run.Key.PolicyID)
	}
	firstLifetime, err := decimal.ParseCanonical(scripted.LifetimeValue)
	if err != nil {
		return 0, err
	}
	secondLifetime, err := decimal.ParseCanonical(elective.LifetimeValue)
	if err != nil {
		return 0, err
	}
	level, err := prestigecore.ReputationDelta(firstLifetime, threshold, 0, policy.ExitModifiersPPM["scripted_first"])
	if err != nil {
		return 0, err
	}
	return prestigecore.ReputationDelta(secondLifetime, threshold, level, policy.ExitModifiersPPM["collapse"])
}

// MeasureReputationThresholds evaluates every candidate threshold against the
// envelope for the gated personas' p50. A failed or unrecorded run is an
// invalid measurement, never silently excluded.
func MeasureReputationThresholds(report FirstHourExperimentReport, policy *prestigecore.Policy, thresholds []string,
	envelope ReputationThresholdEnvelope, gated []string) (ReputationThresholdReport, error) {
	if envelope.Minimum < 0 || envelope.Maximum < envelope.Minimum || len(thresholds) == 0 || len(gated) == 0 {
		return ReputationThresholdReport{}, ErrReputationMeasurement
	}
	result := ReputationThresholdReport{SchemaVersion: 1, SourceScenario: report.ScenarioHash, SourceConstants: report.ConstantsHash,
		Envelope: envelope, GatedPersonas: append([]string{}, gated...), Satisfying: []string{},
		Note: "measurement only: thresholds are proposals for owner SHA ratification (OD-2/OD-10); nothing is minted"}
	for _, run := range report.Runs {
		if run.Outcome != "completed" {
			return ReputationThresholdReport{}, fmt.Errorf("%w: run %s seed %s outcome %s", ErrReputationMeasurement, run.Key.PolicyID, run.Key.Seed, run.Outcome)
		}
	}
	for _, source := range thresholds {
		threshold, err := decimal.ParseCanonical(source)
		if err != nil {
			return ReputationThresholdReport{}, fmt.Errorf("%w: threshold %q", ErrReputationMeasurement, source)
		}
		paid := map[string][]int64{}
		for _, run := range report.Runs {
			value, err := PaidReputationAtFirstElectiveExit(run, policy, threshold)
			if err != nil {
				return ReputationThresholdReport{}, err
			}
			paid[run.Key.PolicyID] = append(paid[run.Key.PolicyID], value)
		}
		row := ReputationThresholdRow{Threshold: source, Satisfies: true}
		ids := make([]string, 0, len(paid))
		for id := range paid {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			values := paid[id]
			sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
			quantile := func(fraction float64) int64 { return values[int(fraction*float64(len(values)-1))] }
			persona := ReputationThresholdPersona{PolicyID: id, Runs: len(values), P05: quantile(0.05), P50: quantile(0.5), P95: quantile(0.95),
				Min: values[0], Max: values[len(values)-1]}
			persona.InEnvelope = persona.P50 >= envelope.Minimum && persona.P50 <= envelope.Maximum
			row.Personas = append(row.Personas, persona)
		}
		for _, id := range gated {
			found := false
			for _, persona := range row.Personas {
				if persona.PolicyID == id {
					found = true
					row.Satisfies = row.Satisfies && persona.InEnvelope
				}
			}
			if !found {
				return ReputationThresholdReport{}, fmt.Errorf("%w: gated persona %s has no runs", ErrReputationMeasurement, id)
			}
		}
		if row.Satisfies {
			result.Satisfying = append(result.Satisfying, source)
		}
		result.Rows = append(result.Rows, row)
	}
	return result, nil
}
