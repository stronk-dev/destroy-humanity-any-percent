package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cloud-clicker/server/harness"
)

func main() {
	mode := flag.String("mode", "check", "run, check, update, or epoch-hash")
	output := flag.String("output", "", "explicit output path for run mode")
	root := flag.String("root", "..", "repository root")
	flag.Parse()
	if *mode == "epoch-hash" {
		hash, err := harness.ComputeEpochSeedHash(*root)
		if err != nil {
			fail(err)
		}
		fmt.Println(hash)
		return
	}
	suite, err := harness.LoadSuite(*root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		fail(err)
	}
	runs, aggregate, err := suite.RunAll()
	if err != nil {
		fail(err)
	}
	registry, err := harness.LoadRelevanceRegistry(*root)
	if err != nil {
		fail(err)
	}
	type relevanceResult struct {
		entry  harness.RelevanceRegistryEntry
		report harness.RelevanceReport
		bytes  []byte
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
		relevanceResults = append(relevanceResults, relevanceResult{entry: entry, report: relevance, bytes: relevanceBytes})
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
		}
	case "run":
		if *output == "" {
			fail(fmt.Errorf("-output is required in run mode"))
		}
		if err := os.WriteFile(*output, suiteBytes, 0o644); err != nil {
			fail(err)
		}
	case "check":
		if err := harness.ValidateRepositoryBaselineChange(*root); err != nil {
			fail(err)
		}
		if err := harness.ValidateRelevanceRegistry(*root); err != nil {
			fail(err)
		}
		if err := harness.ValidateRepositoryEpochChanges(*root); err != nil {
			fail(err)
		}
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
			if gateErr := harness.ValidateActiveRelevanceReport(result.entry, result.report); gateErr != nil {
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

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
