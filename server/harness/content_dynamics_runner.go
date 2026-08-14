package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
)

const ContentDynamicsScenarioSchemaVersion = 1
const ContentDynamicsReportSchemaVersion = 1

const (
	contentDynamicsActivePlay = "active_play_window"
	contentDynamicsFiscal     = "fiscal_harvest"
	contentDynamicsPitch      = "pitch_payout"
	contentDynamicsPermits    = "permit_accrual"
	pitchPolicyVersion        = "pitch.first_four_cheapest.v1"
)

type ContentDynamicsScenario struct {
	SchemaVersion    int                  `json:"schema_version"`
	ID               string               `json:"id"`
	Version          int                  `json:"version"`
	Runs             []ContentDynamicsRun `json:"runs"`
	TransitionBudget int64                `json:"transition_budget"`
}

type ContentDynamicsRun struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	SeedStart string          `json:"seed_start"`
	SeedCount int64           `json:"seed_count"`
	HorizonMS int64           `json:"horizon_ms"`
	Policy    json.RawMessage `json:"policy"`
}

type ContentDynamicsObservation struct {
	RunID     string `json:"run_id"`
	MetricID  string `json:"metric_id"`
	Statistic string `json:"statistic"`
	Value     string `json:"value"`
}

type ContentDynamicsReport struct {
	SchemaVersion       int                          `json:"schema_version"`
	ScenarioID          string                       `json:"scenario_id"`
	ScenarioHash        string                       `json:"scenario_hash"`
	ConstantsHash       string                       `json:"constants_hash"`
	ManifestPath        string                       `json:"manifest_path"`
	DeclaredRuns        int64                        `json:"declared_runs"`
	ExecutedRuns        int64                        `json:"executed_runs"`
	DeclaredTransitions int64                        `json:"declared_transitions"`
	ExecutedTransitions int64                        `json:"executed_transitions"`
	Observations        []ContentDynamicsObservation `json:"observations"`
	InvariantFailures   []string                     `json:"invariant_failures"`
}

type contentDynamicsActivePolicy struct {
	FounderID         string `json:"founder_id"`
	RunSeq            int64  `json:"run_seq"`
	EffectRowID       string `json:"effect_row_id"`
	TargetGeneratorID string `json:"target_generator_id"`
}

type contentDynamicsFiscalPolicy struct {
	Periods int64 `json:"periods"`
}

type contentDynamicsPitchPolicy struct {
	StrategyVersion string `json:"strategy_version"`
}

type contentDynamicsPermitPolicy struct {
	FiscalCredit int64 `json:"fiscal_credit"`
}

type ContentDynamicsSuite struct {
	Scenario      ContentDynamicsScenario
	ScenarioBytes []byte
	ScenarioHash  string
	ManifestPath  string
	Bundle        production.CatalogBundle
}

func LoadCandidateContentDynamicsSuite(repositoryRoot, scenarioPath, manifestPath string) (*ContentDynamicsSuite, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(scenarioPath)))
	if err != nil {
		return nil, err
	}
	scenario, err := loadContentDynamicsScenario(data)
	if err != nil {
		return nil, err
	}
	manifest, artifacts, _, err := loadCandidateManifest(repositoryRoot, manifestPath)
	if err != nil {
		return nil, err
	}
	bundle, err := replaycatalog.Load(manifest.ConstantsHash, artifacts)
	if err != nil {
		return nil, fmt.Errorf("content-dynamics candidate bundle: %w", err)
	}
	resolved, ok := (production.ReplayCatalogSet{manifest.ConstantsHash: bundle}).ResolveReplayCatalogs(manifest.ConstantsHash)
	if !ok {
		return nil, errors.New("content-dynamics candidate bundle is not live-replay valid")
	}
	bundle = resolved
	if bundle.Opportunities == nil || bundle.Relevance == nil || bundle.Fiscal == nil || bundle.Pitch == nil || bundle.Minigames == nil {
		return nil, errors.New("content-dynamics candidate bundle omits a required owner")
	}
	digest := sha256.Sum256(data)
	return &ContentDynamicsSuite{Scenario: scenario, ScenarioBytes: append([]byte(nil), data...),
		ScenarioHash: "sha256:" + hex.EncodeToString(digest[:]), ManifestPath: manifestPath, Bundle: bundle}, nil
}

func loadContentDynamicsScenario(data []byte) (ContentDynamicsScenario, error) {
	if !hasExactContentDynamicsKeys(data, "id", "runs", "schema_version", "transition_budget", "version") {
		return ContentDynamicsScenario{}, errors.New("content-dynamics scenario has an inexact root")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var wire struct {
		SchemaVersion    *json.Number      `json:"schema_version"`
		ID               *string           `json:"id"`
		Version          *json.Number      `json:"version"`
		Runs             []json.RawMessage `json:"runs"`
		TransitionBudget *json.Number      `json:"transition_budget"`
	}
	if err := decoder.Decode(&wire); err != nil {
		return ContentDynamicsScenario{}, err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || wire.SchemaVersion == nil || wire.ID == nil || wire.Version == nil ||
		wire.Runs == nil || wire.TransitionBudget == nil {
		return ContentDynamicsScenario{}, errors.New("content-dynamics scenario missing required value")
	}
	schema, err := relevanceSafeInteger(wire.SchemaVersion)
	if err != nil || schema != ContentDynamicsScenarioSchemaVersion || !relevanceIDPattern.MatchString(*wire.ID) {
		return ContentDynamicsScenario{}, errors.New("content-dynamics scenario has an invalid identity")
	}
	version, err := relevanceSafeInteger(wire.Version)
	if err != nil || version < 1 {
		return ContentDynamicsScenario{}, errors.New("content-dynamics scenario has an invalid version")
	}
	budget, err := relevanceSafeInteger(wire.TransitionBudget)
	if err != nil || budget < 1 {
		return ContentDynamicsScenario{}, errors.New("content-dynamics scenario has an invalid transition budget")
	}
	scenario := ContentDynamicsScenario{SchemaVersion: int(schema), ID: *wire.ID, Version: int(version),
		Runs: make([]ContentDynamicsRun, len(wire.Runs)), TransitionBudget: budget}
	prior := ""
	kindCounts := map[string]int64{}
	declaredRuns := int64(0)
	for index, raw := range wire.Runs {
		run, parseErr := parseContentDynamicsRun(raw)
		if parseErr != nil {
			return ContentDynamicsScenario{}, fmt.Errorf("content-dynamics runs[%d]: %w", index, parseErr)
		}
		if run.ID <= prior {
			return ContentDynamicsScenario{}, errors.New("content-dynamics run ids must be raw-byte sorted and unique")
		}
		prior = run.ID
		scenario.Runs[index] = run
		kindCounts[run.Kind] += run.SeedCount
		if declaredRuns > relevanceMaxSafeInteger-run.SeedCount {
			return ContentDynamicsScenario{}, errors.New("content-dynamics run cardinality overflow")
		}
		declaredRuns += run.SeedCount
	}
	if declaredRuns != 69 || kindCounts[contentDynamicsActivePlay] != 1 || kindCounts[contentDynamicsFiscal] != 2 ||
		kindCounts[contentDynamicsPitch] != 64 || kindCounts[contentDynamicsPermits] != 2 || len(kindCounts) != 4 {
		return ContentDynamicsScenario{}, fmt.Errorf("content-dynamics cardinality active=%d fiscal=%d pitch=%d permits=%d total=%d",
			kindCounts[contentDynamicsActivePlay], kindCounts[contentDynamicsFiscal], kindCounts[contentDynamicsPitch], kindCounts[contentDynamicsPermits], declaredRuns)
	}
	derived, err := contentDynamicsTransitionCeiling(scenario.Runs)
	if err != nil || derived != scenario.TransitionBudget {
		return ContentDynamicsScenario{}, fmt.Errorf("content-dynamics transition budget=%d derived=%d: %w", scenario.TransitionBudget, derived, err)
	}
	return scenario, nil
}

func parseContentDynamicsRun(data []byte) (ContentDynamicsRun, error) {
	if !hasExactContentDynamicsKeys(data, "horizon_ms", "id", "kind", "policy", "seed_count", "seed_start") {
		return ContentDynamicsRun{}, errors.New("inexact run keys")
	}
	var wire struct {
		ID        *string         `json:"id"`
		Kind      *string         `json:"kind"`
		SeedStart *string         `json:"seed_start"`
		SeedCount *json.Number    `json:"seed_count"`
		HorizonMS *json.Number    `json:"horizon_ms"`
		Policy    json.RawMessage `json:"policy"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(&wire); err != nil || wire.ID == nil || wire.Kind == nil || wire.SeedStart == nil || wire.SeedCount == nil || wire.HorizonMS == nil || wire.Policy == nil {
		return ContentDynamicsRun{}, errors.New("missing run value")
	}
	count, err := relevanceSafeInteger(wire.SeedCount)
	if err != nil || count < 1 {
		return ContentDynamicsRun{}, errors.New("invalid seed_count")
	}
	horizon, err := relevanceSafeInteger(wire.HorizonMS)
	if err != nil || horizon < 1 {
		return ContentDynamicsRun{}, errors.New("invalid horizon_ms")
	}
	seed, err := strconv.ParseUint(*wire.SeedStart, 10, 64)
	if err != nil || count > 1 && seed > ^uint64(0)-uint64(count-1) || !relevanceIDPattern.MatchString(*wire.ID) {
		return ContentDynamicsRun{}, errors.New("invalid run identity or seed range")
	}
	run := ContentDynamicsRun{ID: *wire.ID, Kind: *wire.Kind, SeedStart: *wire.SeedStart, SeedCount: count,
		HorizonMS: horizon, Policy: append([]byte(nil), wire.Policy...)}
	switch run.Kind {
	case contentDynamicsActivePlay:
		var policy contentDynamicsActivePolicy
		if !hasExactContentDynamicsKeys(run.Policy, "effect_row_id", "founder_id", "run_seq", "target_generator_id") || decodeContentDynamicsPolicy(run.Policy, &policy) != nil ||
			policy.FounderID == "" || policy.RunSeq < 1 || !relevanceIDPattern.MatchString(policy.EffectRowID) || !relevanceIDPattern.MatchString(policy.TargetGeneratorID) || count != 1 {
			return ContentDynamicsRun{}, errors.New("invalid active_play_window policy")
		}
	case contentDynamicsFiscal:
		var policy contentDynamicsFiscalPolicy
		if !hasExactContentDynamicsKeys(run.Policy, "periods") || decodeContentDynamicsPolicy(run.Policy, &policy) != nil || policy.Periods != 1 && policy.Periods != 4 || count != 1 {
			return ContentDynamicsRun{}, errors.New("invalid fiscal_harvest policy")
		}
	case contentDynamicsPitch:
		var policy contentDynamicsPitchPolicy
		if !hasExactContentDynamicsKeys(run.Policy, "strategy_version") || decodeContentDynamicsPolicy(run.Policy, &policy) != nil || policy.StrategyVersion != pitchPolicyVersion || count != 64 {
			return ContentDynamicsRun{}, errors.New("invalid pitch_payout policy")
		}
	case contentDynamicsPermits:
		var policy contentDynamicsPermitPolicy
		if !hasExactContentDynamicsKeys(run.Policy, "fiscal_credit") || decodeContentDynamicsPolicy(run.Policy, &policy) != nil || policy.FiscalCredit != 0 && policy.FiscalCredit != 100 || count != 1 {
			return ContentDynamicsRun{}, errors.New("invalid permit_accrual policy")
		}
	default:
		return ContentDynamicsRun{}, errors.New("unknown content-dynamics kind")
	}
	return run, nil
}

func contentDynamicsTransitionCeiling(runs []ContentDynamicsRun) (int64, error) {
	total := int64(0)
	for _, run := range runs {
		var perSeed int64
		switch run.Kind {
		case contentDynamicsActivePlay:
			perSeed = 5
		case contentDynamicsFiscal:
			perSeed = 1
		case contentDynamicsPitch:
			perSeed = 64
		case contentDynamicsPermits:
			perSeed = int64(2 * (bits.Len64(uint64(run.HorizonMS)) + 1))
		default:
			return 0, errors.New("unknown content-dynamics kind")
		}
		if run.SeedCount > relevanceMaxSafeInteger/perSeed || total > relevanceMaxSafeInteger-run.SeedCount*perSeed {
			return 0, errors.New("content-dynamics transition ceiling overflow")
		}
		total += run.SeedCount * perSeed
	}
	return total, nil
}

func (suite *ContentDynamicsSuite) Run() (ContentDynamicsReport, error) {
	if suite == nil || suite.Bundle.ConstantsHash == "" || suite.ScenarioHash == "" || suite.ManifestPath == "" {
		return ContentDynamicsReport{}, errors.New("invalid content-dynamics suite")
	}
	observations := make([]ContentDynamicsObservation, 0)
	executedRuns, executedTransitions := int64(0), int64(0)
	for _, run := range suite.Scenario.Runs {
		seedStart, _ := strconv.ParseUint(run.SeedStart, 10, 64)
		switch run.Kind {
		case contentDynamicsActivePlay:
			var policy contentDynamicsActivePolicy
			_ = json.Unmarshal(run.Policy, &policy)
			result, err := production.SimulateContentDynamicsActivePlay(suite.Bundle, production.ContentDynamicsActivePlayInput{
				FounderID: policy.FounderID, RunSeq: policy.RunSeq, EffectRowID: policy.EffectRowID, TargetGeneratorID: policy.TargetGeneratorID})
			if err != nil || result.SpawnedAttendedMS+result.DurationMS > run.HorizonMS {
				return ContentDynamicsReport{}, fmt.Errorf("content-dynamics %s: %w", run.ID, firstError(err, errors.New("active-play window outside horizon")))
			}
			executedRuns++
			executedTransitions += result.Transitions
			observations = append(observations,
				contentObservation(run.ID, "active_play.bonus_output", "exact", result.BonusOutput.String()),
				contentObservation(run.ID, "active_play.claimed_output", "exact", result.ClaimedOutput.String()),
				contentObservation(run.ID, "active_play.control_output", "exact", result.ControlOutput.String()),
				contentObservation(run.ID, "active_play.duration_ms", "exact", strconv.FormatInt(result.DurationMS, 10)),
				contentObservation(run.ID, "active_play.spawned_attended_ms", "exact", strconv.FormatInt(result.SpawnedAttendedMS, 10)))
		case contentDynamicsFiscal:
			var policy contentDynamicsFiscalPolicy
			_ = json.Unmarshal(run.Policy, &policy)
			result, err := production.SimulateContentDynamicsFiscal(suite.Bundle, policy.Periods)
			if err != nil || policy.Periods*suite.Bundle.Fiscal.Clock.AutoMS > run.HorizonMS {
				return ContentDynamicsReport{}, fmt.Errorf("content-dynamics %s: %w", run.ID, firstError(err, errors.New("fiscal sweep outside horizon")))
			}
			executedRuns++
			executedTransitions += result.Transitions
			observations = append(observations,
				contentObservation(run.ID, "fiscal.credit_after", "exact", strconv.FormatInt(result.CreditAfter, 10)),
				contentObservation(run.ID, "fiscal.credited", "exact", strconv.FormatInt(result.Credited, 10)),
				contentObservation(run.ID, "fiscal.period_sequence_after", "exact", strconv.FormatInt(result.SequenceAfter, 10)),
				contentObservation(run.ID, "fiscal.saturated", "exact", boolString(result.Saturated)))
		case contentDynamicsPitch:
			finalRounds, payoutScores, cash := make([]int64, 0, run.SeedCount), make([]int64, 0, run.SeedCount), make([]int64, 0, run.SeedCount)
			for offset := int64(0); offset < run.SeedCount; offset++ {
				result, err := production.SimulateContentDynamicsPitch(suite.Bundle, seedStart+uint64(offset))
				if err != nil {
					return ContentDynamicsReport{}, fmt.Errorf("content-dynamics %s seed %d: %w", run.ID, seedStart+uint64(offset), err)
				}
				executedRuns++
				executedTransitions += result.Transitions
				finalRounds = append(finalRounds, result.FinalRound)
				payoutScores = append(payoutScores, result.PayoutScore)
				cash = append(cash, result.ConvertedCash)
			}
			observations = append(observations, percentileObservations(run.ID, "pitch.final_round", finalRounds)...)
			observations = append(observations, percentileObservations(run.ID, "pitch.payout_score", payoutScores)...)
			observations = append(observations, percentileObservations(run.ID, "pitch.converted_cash", cash)...)
		case contentDynamicsPermits:
			var policy contentDynamicsPermitPolicy
			_ = json.Unmarshal(run.Policy, &policy)
			result, err := production.SimulateContentDynamicsPermits(suite.Bundle, policy.FiscalCredit, run.HorizonMS)
			if err != nil {
				return ContentDynamicsReport{}, fmt.Errorf("content-dynamics %s: %w", run.ID, err)
			}
			executedRuns++
			executedTransitions += result.Transitions
			observations = append(observations,
				contentObservation(run.ID, "permits.time_to_12_ms", "p50", strconv.FormatInt(result.TimeToTwelveMS, 10)),
				contentObservation(run.ID, "permits.time_to_12_ms", "p95", strconv.FormatInt(result.TimeToTwelveMS, 10)),
				contentObservation(run.ID, "permits.time_to_cap_ms", "p50", strconv.FormatInt(result.TimeToCapMS, 10)),
				contentObservation(run.ID, "permits.time_to_cap_ms", "p95", strconv.FormatInt(result.TimeToCapMS, 10)))
		}
		if executedTransitions > suite.Scenario.TransitionBudget {
			return ContentDynamicsReport{}, fmt.Errorf("content-dynamics transition budget exceeded: %d > %d", executedTransitions, suite.Scenario.TransitionBudget)
		}
	}
	sort.Slice(observations, func(i, j int) bool {
		return contentObservationKey(observations[i]) < contentObservationKey(observations[j])
	})
	report := ContentDynamicsReport{SchemaVersion: ContentDynamicsReportSchemaVersion, ScenarioID: suite.Scenario.ID,
		ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.Bundle.ConstantsHash, ManifestPath: suite.ManifestPath,
		DeclaredRuns: 69, ExecutedRuns: executedRuns, DeclaredTransitions: suite.Scenario.TransitionBudget,
		ExecutedTransitions: executedTransitions, Observations: observations, InvariantFailures: []string{}}
	if err := ValidateContentDynamicsReport(report, suite.Scenario); err != nil {
		return ContentDynamicsReport{}, err
	}
	return report, nil
}

func ValidateContentDynamicsReport(report ContentDynamicsReport, scenario ContentDynamicsScenario) error {
	if report.SchemaVersion != ContentDynamicsReportSchemaVersion || !relevanceIDPattern.MatchString(report.ScenarioID) ||
		!relevanceHashPattern.MatchString(report.ScenarioHash) || !relevanceHashPattern.MatchString(report.ConstantsHash) || report.ManifestPath == "" ||
		report.ScenarioID != scenario.ID || report.DeclaredRuns != 69 || report.ExecutedRuns != report.DeclaredRuns ||
		report.DeclaredTransitions != scenario.TransitionBudget ||
		report.ExecutedTransitions < 1 || report.ExecutedTransitions > report.DeclaredTransitions || report.Observations == nil || report.InvariantFailures == nil || len(report.InvariantFailures) != 0 {
		return errors.New("invalid content-dynamics report envelope")
	}
	expected := make(map[string]bool)
	for _, run := range scenario.Runs {
		var rows [][2]string
		switch run.Kind {
		case contentDynamicsActivePlay:
			rows = [][2]string{{"active_play.bonus_output", "exact"}, {"active_play.claimed_output", "exact"},
				{"active_play.control_output", "exact"}, {"active_play.duration_ms", "exact"}, {"active_play.spawned_attended_ms", "exact"}}
		case contentDynamicsFiscal:
			rows = [][2]string{{"fiscal.credit_after", "exact"}, {"fiscal.credited", "exact"},
				{"fiscal.period_sequence_after", "exact"}, {"fiscal.saturated", "exact"}}
		case contentDynamicsPitch:
			for _, metric := range []string{"pitch.converted_cash", "pitch.final_round", "pitch.payout_score"} {
				rows = append(rows, [2]string{metric, "p50"}, [2]string{metric, "p95"})
			}
		case contentDynamicsPermits:
			for _, metric := range []string{"permits.time_to_12_ms", "permits.time_to_cap_ms"} {
				rows = append(rows, [2]string{metric, "p50"}, [2]string{metric, "p95"})
			}
		default:
			return errors.New("content-dynamics report references an unknown scenario kind")
		}
		for _, row := range rows {
			expected[run.ID+"\x00"+row[0]+"\x00"+row[1]] = true
		}
	}
	if len(report.Observations) != len(expected) {
		return errors.New("content-dynamics report observation set is incomplete")
	}
	prior := ""
	for _, observation := range report.Observations {
		key := contentObservationKey(observation)
		if !relevanceIDPattern.MatchString(observation.RunID) || !relevanceIDPattern.MatchString(observation.MetricID) ||
			!expected[key] || key <= prior || !validContentDynamicsValue(observation.Value) {
			return fmt.Errorf("invalid content-dynamics observation %+v", observation)
		}
		prior = key
	}
	return nil
}

func percentileObservations(runID, metric string, values []int64) []ContentDynamicsObservation {
	ordered := append([]int64(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	return []ContentDynamicsObservation{
		contentObservation(runID, metric, "p50", strconv.FormatInt(nearestRank(ordered, 50), 10)),
		contentObservation(runID, metric, "p95", strconv.FormatInt(nearestRank(ordered, 95), 10)),
	}
}

func nearestRank(values []int64, percentile int64) int64 {
	index := (int64(len(values))*percentile + 99) / 100
	if index < 1 {
		index = 1
	}
	return values[index-1]
}

func contentObservation(runID, metricID, statistic, value string) ContentDynamicsObservation {
	return ContentDynamicsObservation{RunID: runID, MetricID: metricID, Statistic: statistic, Value: value}
}

func contentObservationKey(value ContentDynamicsObservation) string {
	return value.RunID + "\x00" + value.MetricID + "\x00" + value.Statistic
}

func validContentDynamicsValue(value string) bool {
	if integer, integerErr := strconv.ParseInt(value, 10, 64); integerErr == nil {
		return integer >= -relevanceMaxSafeInteger && integer <= relevanceMaxSafeInteger
	}
	parsed, err := decimal.ParseCanonical(value)
	if err != nil || !parsed.IsStateValue() {
		return false
	}
	return true
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func decodeContentDynamicsPolicy(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return errors.New("multiple policy values")
	}
	return nil
}

func hasExactContentDynamicsKeys(data []byte, expected ...string) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil || len(object) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func firstError(primary, fallback error) error {
	if primary != nil {
		return primary
	}
	return fallback
}
