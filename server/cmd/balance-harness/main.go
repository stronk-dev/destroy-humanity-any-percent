package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/harness"
)

func main() {
	mode := flag.String("mode", "check", "run, check, update, guard, candidate, content-candidate, content, relevance, relevance-branches, relevance-beam, first-hour, or epoch-hash")
	output := flag.String("output", "", "explicit output path for run mode")
	root := flag.String("root", "..", "repository root")
	candidateManifest := flag.String("candidate-manifest", "", "repository-relative ratified candidate manifest for candidate mode")
	scenario := flag.String("scenario", "", "repository-relative scenario for relevance mode")
	relevanceReport := flag.String("relevance-report", "", "validated relevance report used to derive branch rows")
	workers := flag.Int("workers", 4, "parallel workers for the standard pacing suite")
	firstHourPolicy := flag.String("first-hour-policy", "", "repository-relative ratified first-hour policy for first-hour mode")
	acquihireMinimum := flag.Int64("acquihire-minimum", 0, "measurement-only first-hour acquihire purchased-generator threshold")
	burnoutFactor := flag.String("burnout-factor", "", "measurement-only canonical burnout price factor")
	routeKnowledgeBonus := flag.Int64("route-knowledge-bonus", 0, "measurement-only first-hour Route Knowledge bonus")
	seedCapital := flag.String("seed-capital", "", "measurement-only canonical seed-capital amount")
	generatedTowers := flag.Int64("generated-towers", 0, "measurement-only generated Beige Tower count")
	flag.Parse()
	if *mode == "epoch-hash" {
		hash, err := harness.ComputeEpochSeedHash(*root)
		if err != nil {
			fail(err)
		}
		fmt.Println(hash)
		return
	}
	if *mode == "candidate" {
		runCandidate(*root, *output, *candidateManifest, *workers)
		return
	}
	if *mode == "content-candidate" {
		runContentCandidate(*root, *output, *scenario, *candidateManifest)
		return
	}
	if *mode == "relevance" {
		if err := runRelevance(*root, *output, *scenario); err != nil {
			fail(err)
		}
		return
	}
	if *mode == "relevance-branches" {
		if err := runRelevanceBranches(*root, *output, *scenario, *relevanceReport); err != nil {
			fail(err)
		}
		return
	}
	if *mode == "relevance-beam" {
		if err := runRelevanceBeam(*root, *output, *scenario); err != nil {
			fail(err)
		}
		return
	}
	if *mode == "first-hour" {
		if err := runFirstHour(*root, *output, *scenario, *firstHourPolicy, *workers, harness.FirstHourExperiment{
			AcquihirePurchasedMinimum: *acquihireMinimum, BurnoutPriceFactor: *burnoutFactor,
			RouteKnowledgeBonus: *routeKnowledgeBonus, SeedCapital: *seedCapital, GeneratedBeigeTowers: *generatedTowers,
		}); err != nil {
			fail(err)
		}
		return
	}
	if *mode == "content" {
		if err := harness.GenerateRegisteredContentBaselines(*root); err != nil {
			fail(err)
		}
		return
	}
	if *mode == "guard" {
		if _, err := harness.LoadRelevanceRegistry(*root); err != nil {
			fail(err)
		}
		if err := validateHarnessRepositoryGuards(*root); err != nil {
			fail(err)
		}
		return
	}
	var registry []harness.RelevanceRegistryEntry
	var err error
	if *mode == "update" {
		registry, err = harness.LoadRelevanceRegistryForUpdate(*root)
	} else {
		registry, err = harness.LoadRelevanceRegistry(*root)
	}
	if err != nil {
		fail(err)
	}
	if *mode == "check" {
		if err := validateHarnessRepositoryGuards(*root); err != nil {
			fail(err)
		}
	}
	suite, err := harness.LoadSuite(*root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		fail(err)
	}
	runs, aggregate, err := suite.RunAllWithWorkers(*workers)
	if err != nil {
		fail(err)
	}
	type relevanceResult struct {
		entry       harness.RelevanceRegistryEntry
		report      harness.RelevanceReport
		bytes       []byte
		branch      *harness.RelevanceBranchReport
		branchBytes []byte
	}
	relevanceResults := make([]relevanceResult, 0, len(registry))
	for _, entry := range registry {
		relevanceSuite, loadErr := harness.LoadRegisteredRelevanceSuite(*root, entry)
		if loadErr != nil {
			fail(loadErr)
		}
		relevance, runErr := relevanceSuite.RunRelevance()
		if runErr != nil {
			fail(runErr)
		}
		relevanceBytes, encodeErr := harness.CanonicalJSON(relevance)
		if encodeErr != nil {
			fail(encodeErr)
		}
		result := relevanceResult{entry: entry, report: relevance, bytes: relevanceBytes}
		if entry.BranchReport != "" && (*mode == "update" || *mode == "check") {
			var branch harness.RelevanceBranchReport
			var branchErr error
			if *mode == "update" {
				branch, branchErr = relevanceSuite.RunRelevanceBranchProofs(&relevance)
			} else {
				branch, branchErr = harness.LoadRegisteredRelevanceBranchReport(*root, entry)
			}
			if branchErr != nil {
				fail(branchErr)
			}
			branchBytes, encodeErr := harness.CanonicalJSON(branch)
			if encodeErr != nil {
				fail(encodeErr)
			}
			result.branch, result.branchBytes = &branch, branchBytes
		}
		relevanceResults = append(relevanceResults, result)
	}
	golden := harness.GoldenReport{SchemaVersion: 1}
	for _, report := range runs {
		if report.Key.Seed == "0" {
			golden.Runs = append(golden.Runs, report)
		}
	}
	goldenBytes, _ := harness.CanonicalJSON(golden)
	aggregateBytes, _ := harness.CanonicalJSON(aggregate)
	suiteBytes, _ := harness.CanonicalJSON(harness.SuiteReport{SchemaVersion: 1, Runs: runs, Aggregate: aggregate})
	goldenPath := filepath.Join(*root, "testdata/harness/golden-seed.json")
	baselinePath := filepath.Join(*root, "testdata/harness/pacing-baseline.json")
	switch *mode {
	case "update":
		if err := os.WriteFile(goldenPath, goldenBytes, 0o644); err != nil {
			fail(err)
		}
		if err := os.WriteFile(baselinePath, aggregateBytes, 0o644); err != nil {
			fail(err)
		}
		for _, result := range relevanceResults {
			if err := os.WriteFile(filepath.Join(*root, filepath.FromSlash(result.entry.GoldenReport)), result.bytes, 0o644); err != nil {
				fail(err)
			}
			if result.branch != nil {
				if err := os.WriteFile(filepath.Join(*root, filepath.FromSlash(result.entry.BranchReport)), result.branchBytes, 0o644); err != nil {
					fail(err)
				}
			}
		}
	case "run":
		if *output == "" {
			fail(fmt.Errorf("-output is required in run mode"))
		}
		if err := os.WriteFile(*output, suiteBytes, 0o644); err != nil {
			fail(err)
		}
	case "check":
		wantGolden, err := os.ReadFile(goldenPath)
		if err != nil {
			fail(err)
		}
		wantBaseline, err := os.ReadFile(baselinePath)
		if err != nil {
			fail(err)
		}
		var baseline harness.AggregateReport
		if err := json.Unmarshal(wantBaseline, &baseline); err != nil {
			fail(err)
		}
		warnings, driftFailures := harness.CompareBaseline(aggregate, baseline)
		for _, warning := range warnings {
			fmt.Fprintln(os.Stderr, "warning:", warning)
		}
		if string(wantGolden) != string(goldenBytes) {
			fail(fmt.Errorf("golden seed drift; run make harness-update and review"))
		}
		for _, result := range relevanceResults {
			wantRelevance, readErr := os.ReadFile(filepath.Join(*root, filepath.FromSlash(result.entry.GoldenReport)))
			if readErr != nil {
				fail(readErr)
			}
			if string(wantRelevance) != string(result.bytes) {
				fail(fmt.Errorf("relevance golden drift for %q; run make harness-update and review", result.entry.Scenario))
			}
			if result.entry.Active {
				wantBranch, readErr := os.ReadFile(filepath.Join(*root, filepath.FromSlash(result.entry.BranchReport)))
				if readErr != nil {
					fail(readErr)
				}
				if result.branch == nil || string(wantBranch) != string(result.branchBytes) {
					fail(fmt.Errorf("relevance branch golden drift for %q; run make harness-update and review", result.entry.Scenario))
				}
				if gateErr := harness.ValidateActiveRelevanceEvidence(result.entry, result.report, *result.branch); gateErr != nil {
					fail(gateErr)
				}
			} else if gateErr := harness.ValidateActiveRelevanceReport(result.entry, result.report); gateErr != nil {
				fail(gateErr)
			}
		}
		if aggregate.ScenarioHash != baseline.ScenarioHash || aggregate.ConstantsHash != baseline.ConstantsHash {
			fail(fmt.Errorf("baseline hashes do not match scenario/catalog"))
		}
		if len(aggregate.Failures) > 0 || len(driftFailures) > 0 {
			fail(fmt.Errorf("harness failures: %v %v", aggregate.Failures, driftFailures))
		}
		_ = wantBaseline
	default:
		fail(fmt.Errorf("unsupported mode %q", *mode))
	}
}

func runFirstHour(root, output, scenarioPath, policyPath string, workers int, experiment harness.FirstHourExperiment) error {
	if output == "" || scenarioPath == "" || policyPath == "" {
		return fmt.Errorf("-output, -scenario, and -first-hour-policy are required in first-hour mode")
	}
	suite, err := harness.LoadFirstHourSuite(root, scenarioPath, policyPath)
	if err != nil {
		return err
	}
	report, err := suite.RunAllExperiments(experiment, workers)
	if err != nil {
		return err
	}
	data, err := harness.CanonicalJSON(report)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return err
	}
	if len(report.Aggregate.Failures) != 0 {
		return fmt.Errorf("first-hour harness findings: %v", report.Aggregate.Failures)
	}
	return nil
}

func validateHarnessRepositoryGuards(root string) error {
	if err := harness.ValidateRepositoryBaselineChange(root); err != nil {
		return err
	}
	if err := harness.ValidateContentDynamicsRegistry(root); err != nil {
		return err
	}
	return harness.ValidateRepositoryEpochChanges(root)
}

func runContentCandidate(root, output, scenarioPath, manifestPath string) {
	if output == "" || scenarioPath == "" || manifestPath == "" {
		fail(fmt.Errorf("-output, -scenario, and -candidate-manifest are required in content-candidate mode"))
	}
	suite, err := harness.LoadCandidateContentDynamicsSuite(root, scenarioPath, manifestPath)
	if err != nil {
		fail(err)
	}
	report, err := suite.Run()
	if err != nil {
		fail(err)
	}
	reportBytes, err := harness.CanonicalJSON(report)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(output, reportBytes, 0o644); err != nil {
		fail(err)
	}
}

func runCandidate(root, output, manifestPath string, workers int) {
	if output == "" || manifestPath == "" {
		fail(fmt.Errorf("-output and -candidate-manifest are required in candidate mode"))
	}
	suite, identity, err := harness.LoadCandidateSuite(root, "testdata/harness/scenarios/phase0-production.json", manifestPath)
	if err != nil {
		fail(err)
	}
	_, aggregate, err := suite.RunAllWithWorkers(workers)
	if err != nil {
		fail(err)
	}
	baselineBytes, err := os.ReadFile(filepath.Join(root, "testdata", "harness", "pacing-baseline.json"))
	if err != nil {
		fail(err)
	}
	var baseline harness.AggregateReport
	if err := json.Unmarshal(baselineBytes, &baseline); err != nil {
		fail(err)
	}
	report := harness.BuildCandidatePacingReport(identity, aggregate, baseline)
	reportBytes, err := harness.CanonicalJSON(report)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(output, reportBytes, 0o644); err != nil {
		fail(err)
	}
	if len(report.InvariantFailures) > 0 {
		fail(fmt.Errorf("candidate harness invariant failures: %v", report.InvariantFailures))
	}
}

func runRelevance(root, output, scenarioPath string) error {
	if output == "" || scenarioPath == "" {
		return fmt.Errorf("-output and -scenario are required in relevance mode")
	}
	suite, err := harness.LoadRelevanceSuite(root, scenarioPath)
	if err != nil {
		return err
	}
	report, err := suite.RunRelevance()
	if err != nil {
		return err
	}
	return writeRelevanceReport(output, report)
}

func runRelevanceBeam(root, output, scenarioPath string) error {
	if scenarioPath == "" {
		return fmt.Errorf("-scenario is required in relevance-beam mode")
	}
	suite, err := harness.LoadRelevanceSuite(root, scenarioPath)
	if err != nil {
		return err
	}
	diagnostic, err := suite.RunBeamDiagnostic()
	if err != nil {
		return err
	}
	data, err := harness.CanonicalJSON(diagnostic)
	if err != nil {
		return err
	}
	if output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(output, data, 0o644)
	}
	if err != nil {
		return err
	}
	if !diagnostic.Oracle.Passed {
		return fmt.Errorf("manual relevance beam found greedy gap or search regression: %+v", diagnostic.Oracle)
	}
	return nil
}

func runRelevanceBranches(root, output, scenarioPath, reportPath string) error {
	if output == "" || scenarioPath == "" {
		return fmt.Errorf("-output and -scenario are required in relevance-branches mode")
	}
	suite, err := harness.LoadRelevanceSuite(root, scenarioPath)
	if err != nil {
		return err
	}
	var measurement *harness.RelevanceReport
	if reportPath != "" {
		data, readErr := os.ReadFile(reportPath)
		if readErr != nil {
			return readErr
		}
		var direct harness.RelevanceReport
		if unmarshalErr := json.Unmarshal(data, &direct); unmarshalErr != nil {
			return unmarshalErr
		}
		if direct.ScenarioID == "" {
			var diagnostic struct {
				Report harness.RelevanceReport `json:"report"`
			}
			if unmarshalErr := json.Unmarshal(data, &diagnostic); unmarshalErr != nil {
				return unmarshalErr
			}
			direct = diagnostic.Report
		}
		measurement = &direct
	}
	report, err := suite.RunRelevanceBranchProofs(measurement)
	if err != nil {
		return err
	}
	data, err := harness.CanonicalJSON(report)
	if err != nil {
		return err
	}
	if len(report.Failures) > 0 {
		_ = os.Remove(output)
		if err := os.WriteFile(relevanceDiagnosticPath(output), data, 0o644); err != nil {
			return err
		}
		return fmt.Errorf("relevance branch failures: %v; executed_transitions=%d", report.Failures, report.ExecutedTransitions)
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		return err
	}
	if err := os.Remove(relevanceDiagnosticPath(output)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func writeRelevanceReport(output string, report harness.RelevanceReport) error {
	if len(report.Failures) > 0 {
		diagnostic := struct {
			SchemaVersion int                     `json:"schema_version"`
			Kind          string                  `json:"kind"`
			Authoritative bool                    `json:"authoritative"`
			Report        harness.RelevanceReport `json:"report"`
		}{SchemaVersion: 1, Kind: "non_authoritative_relevance_diagnostic", Authoritative: false, Report: report}
		bytes, err := harness.CanonicalJSON(diagnostic)
		if err != nil {
			return err
		}
		_ = os.Remove(output)
		if err := os.WriteFile(relevanceDiagnosticPath(output), bytes, 0o644); err != nil {
			return err
		}
		return fmt.Errorf("relevance failures: %v; run_budget=%+v; deviation_oracle=%+v", report.Failures, report.RunBudget, report.DeviationOracle)
	}
	reportBytes, err := harness.CanonicalJSON(report)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, reportBytes, 0o644); err != nil {
		return err
	}
	if err := os.Remove(relevanceDiagnosticPath(output)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func relevanceDiagnosticPath(output string) string {
	return strings.TrimSuffix(output, ".json") + ".diagnostic.json"
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
