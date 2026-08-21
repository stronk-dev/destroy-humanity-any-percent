package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cloud-clicker/server/harness"
)

func runRegisteredRelevance(root, output, selector string) error {
	if output == "" || selector == "" {
		return fmt.Errorf("-output and -relevance-entry are required in relevance-registered mode")
	}
	registry, err := harness.LoadRelevanceRegistry(root)
	if err != nil {
		return err
	}
	matchedIndex := -1
	var entry harness.RelevanceRegistryEntry
	for index, candidate := range registry {
		if candidate.Scenario != selector {
			continue
		}
		if matchedIndex != -1 {
			return fmt.Errorf("ambiguous registered relevance selector %q", selector)
		}
		matchedIndex, entry = index, candidate
	}
	if matchedIndex == -1 {
		return fmt.Errorf("unregistered relevance selector %q", selector)
	}
	suite, err := harness.LoadRegisteredRelevanceSuite(root, entry)
	if err != nil {
		return err
	}
	active := entry.Active
	identity := harness.HarnessObservationIdentity{RegistryIndex: &matchedIndex, ScenarioPath: entry.Scenario,
		EconomyCatalogPath: entry.EconomyCatalog, RelevancePolicyPath: entry.RelevancePolicy,
		GoldenReportPath: entry.GoldenReport, ScenarioHash: suite.ScenarioHash,
		RelevancePolicyHash: suite.Policy.Hash, ConstantsHash: suite.ConstantsHash, Active: &active}
	if err := observationRecorder.StartObjective(harness.HarnessObservationObjectiveSpec{
		ID: fmt.Sprintf("relevance:%d:%s", matchedIndex, entry.Scenario), Kind: "registered_relevance", Identity: identity}); err != nil {
		return err
	}
	report, err := suite.RunRelevanceObserved(recordRelevanceProgress)
	if err != nil {
		return err
	}
	data, err := harness.CanonicalJSON(report)
	if err != nil {
		return err
	}
	want, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.GoldenReport)))
	if err != nil {
		return err
	}
	if string(want) != string(data) {
		return fmt.Errorf("relevance golden drift for %q", entry.Scenario)
	}
	if entry.Active {
		branch, loadErr := harness.LoadRegisteredRelevanceBranchReport(root, entry)
		if loadErr != nil {
			return loadErr
		}
		if err := harness.ValidateActiveRelevanceEvidence(entry, report, branch); err != nil {
			return err
		}
	} else if err := harness.ValidateActiveRelevanceReport(entry, report); err != nil {
		return err
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return err
	}
	progress := clearObservationProgress()
	progress.Work = observationWorkFromRelevance(report)
	progress.InstrumentExcluded = report.InstrumentExcludedIDs
	return observationRecorder.CompleteObjective(progress)
}

func recordRelevanceProgress(progress harness.RelevanceProgress) error {
	work := harness.HarnessObservationWork{DeclaredRuns: int64Address(progress.DeclaredRuns),
		ExecutedRuns: int64Address(progress.ExecutedRuns), ExecutedTransitions: int64Address(progress.ExecutedTransitions)}
	if progress.Complete {
		work.DeclaredTransitions = int64Address(progress.ExecutedTransitions)
	}
	return observationRecorder.Progress(harness.HarnessObservationProgress{Work: work})
}

func observationWorkFromRelevance(report harness.RelevanceReport) harness.HarnessObservationWork {
	return harness.HarnessObservationWork{
		DeclaredRuns: int64Address(report.RunBudget.DeclaredRuns), ExecutedRuns: int64Address(report.RunBudget.ExecutedRuns),
		DeclaredTransitions: int64Address(report.RunBudget.DeclaredTransitions), ExecutedTransitions: int64Address(report.RunBudget.ExecutedTransitions),
	}
}

func clearObservationProgress() harness.HarnessObservationProgress {
	return harness.HarnessObservationProgress{GuardState: harness.ObservationConditionClear,
		PopulationExclusions: harness.ObservationConditionClear, TruncationState: harness.ObservationConditionClear}
}

func observationGuardState(fired bool) string {
	if fired {
		return harness.ObservationConditionFired
	}
	return harness.ObservationConditionClear
}

func declaredPacingRuns(suite *harness.Suite) int64 {
	var count int64
	for _, run := range suite.Scenario.Runs {
		count += int64(run.SeedCount)
	}
	return count
}

func int64Address(value int64) *int64 { return &value }

func installObservationSignalHandler() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		received := <-signals
		err := fmt.Errorf("received signal %s", received)
		_ = observationRecorder.Fail("signal", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}()
}
