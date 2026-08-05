package production

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

// applyFoundationTransition owns the post-action part of the closed replay
// hook chain. V3 evidence predates executable foundation hooks and remains
// replayable under its pinned semantics; every current active command is v4.
func applyFoundationTransition(bundle CatalogBundle, before, state, founder *save.State, revision save.Revision, request IntentRequest, now time.Time, contributions []multiplier.Contribution, actionDebits map[string]string, terminal bool, events *[]save.EventWrite) error {
	if !bundle.foundationsActive() || before == nil || state == nil || founder == nil || events == nil {
		return ErrInvalidEngineState
	}
	attendedBefore, err := prestigecore.AttendedMS(before, before.EvaluatedThrough)
	if err != nil {
		return ErrInvalidEngineState
	}
	attendedAfter, err := prestigecore.AttendedMS(state, state.EvaluatedThrough)
	if err != nil || attendedAfter < attendedBefore {
		return ErrInvalidEngineState
	}
	attendedMS := attendedAfter - attendedBefore
	if attendedMS > meters.MaximumAttendedStep {
		return ErrInvalidEngineState
	}
	newFacts := map[string]bool{}
	for fact, present := range state.LedgerFactKinds {
		if present && !before.LedgerFactKinds[fact] {
			newFacts[fact] = true
		}
	}
	activeContributions := map[string]bool{}
	for _, contribution := range contributions {
		if !contribution.Factor.Eq(decimal.One) {
			activeContributions[meters.ContributionKey(string(contribution.Slot), contribution.SourceID)] = true
		}
	}
	meterState := meters.State{Values: state.MeterValues, DecayRemainders: state.MeterDecayRemainders, InputRemainders: state.MeterInputRemainders}
	changes, err := meters.Advance(bundle.Meters, meterState, meters.AdvanceContext{AttendedMS: attendedMS, NewFactKinds: newFacts, ActiveContributions: activeContributions})
	if err != nil {
		return err
	}
	runID := map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq}
	for _, change := range changes {
		payload, _ := json.Marshal(map[string]any{"run_id": runID, "meter_id": change.MeterID, "from_band": change.FromBand,
			"to_band": change.ToBand, "direction": change.Direction, "value_before": change.ValueBefore, "value_after": change.ValueAfter})
		*events = append(*events, save.EventWrite{Kind: save.EventMeterBandChanged, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	}
	run, career, err := achievementObservations(state, founder, terminal, now)
	if err != nil {
		return err
	}
	newlyEarned, err := bundle.Achievements.NewlyEarned(state.AchievementsEarnedRun, founder.AchievementsEarnedLifetime, run, career)
	if err != nil {
		return err
	}
	staged := make([]achievements.Definition, 0, len(newlyEarned))
	for _, definition := range newlyEarned {
		if achievementProofSatisfied(definition, actionDebits, *events) {
			staged = append(staged, definition)
		}
	}
	for _, definition := range staged {
		state.AchievementsEarnedRun[definition.ID] = true
	}
	score, err := bundle.Achievements.Score(state.AchievementsEarnedRun)
	if err != nil {
		return err
	}
	state.AchievementScoreRun = score
	for _, definition := range staged {
		payload, _ := json.Marshal(map[string]any{"run_id": runID, "achievement_id": definition.ID,
			"condition_scope": definition.ConditionScope, "score_grant": definition.ScoreGrant})
		*events = append(*events, save.EventWrite{Kind: save.EventAchievementEarned, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	}
	return bundle.ValidateFoundationState(state)
}

func achievementObservations(company, founder *save.State, terminal bool, now time.Time) (achievements.Observation, achievements.Observation, error) {
	runFacts := cloneBools(company.LedgerFactKinds)
	runGenerators := make(map[string]int64, len(company.GeneratorCounts))
	for id, count := range company.GeneratorCounts {
		runGenerators[id] = count
	}
	run := achievements.Observation{Facts: runFacts, Counters: map[string]int64{
		"generators_purchased_total": company.GeneratorPurchasedTotal,
		"tier":                       company.Tier,
	}, ExitCount: 0, Generators: runGenerators}
	careerFacts := cloneBools(founder.LedgerFactKinds)
	ageMS, exitCount := founder.AgeMS, int64(len(founder.ExitHistory))
	if terminal {
		attended, err := prestigecore.AttendedMS(company, save.CanonicalServerTime(now))
		if err != nil || ageMS > decimal.MaxExactInteger-attended {
			return achievements.Observation{}, achievements.Observation{}, ErrInvalidEngineState
		}
		ageMS += attended
		exitCount++
		for fact, present := range company.LedgerFactKinds {
			if present {
				careerFacts[fact] = true
			}
		}
	}
	career := achievements.Observation{Facts: careerFacts, Counters: map[string]int64{
		"age_ms": ageMS, "notoriety": founder.Notoriety,
	}, ExitCount: exitCount, Generators: map[string]int64{}}
	return run, career, nil
}

func achievementProofSatisfied(definition achievements.Definition, actionDebits map[string]string, events []save.EventWrite) bool {
	if definition.Proof.Kind != achievements.ProofBurn {
		return true
	}
	matchedEvent := false
	for _, event := range events {
		if string(event.Kind) == definition.Proof.EventKind {
			matchedEvent = true
			break
		}
	}
	if !matchedEvent || actionDebits == nil {
		return false
	}
	debitRaw, ok := actionDebits[definition.Proof.ResourceID]
	minimum, err := decimal.ParseCanonical(definition.Proof.Minimum)
	debit, debitErr := decimal.ParseCanonical(debitRaw)
	if !ok || err != nil || debitErr != nil {
		return false
	}
	return debit.Gte(minimum)
}

func validateFoundationHookInputs(bundle CatalogBundle, state, founder *save.State) error {
	if !bundle.foundationsActive() || state == nil || founder == nil {
		return fmt.Errorf("%w: foundation hook inputs", ErrInvalidReplayInputs)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		return err
	}
	return validateFounderCarryFoundationState(bundle, founder)
}
