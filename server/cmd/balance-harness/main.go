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
	relevanceSuite, err := harness.LoadRelevanceSuite(*root, "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		fail(err)
	}
	relevance, err := relevanceSuite.RunRelevance()
	if err != nil {
		fail(err)
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
	relevanceBytes, _ := harness.CanonicalJSON(relevance)
	goldenPath := filepath.Join(*root, "testdata/harness/golden-seed.json")
	baselinePath := filepath.Join(*root, "testdata/harness/pacing-baseline.json")
	relevancePath := filepath.Join(*root, "testdata/harness/relevance/golden-report-v1.json")
	switch *mode {
	case "update":
		if err := os.WriteFile(goldenPath, goldenBytes, 0o644); err != nil {
			fail(err)
		}
		if err := os.WriteFile(baselinePath, aggregateBytes, 0o644); err != nil {
			fail(err)
		}
		if err := os.WriteFile(relevancePath, relevanceBytes, 0o644); err != nil {
			fail(err)
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
		wantRelevance, err := os.ReadFile(relevancePath)
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
		if string(wantRelevance) != string(relevanceBytes) {
			fail(fmt.Errorf("relevance golden drift; run make harness-update and review"))
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
