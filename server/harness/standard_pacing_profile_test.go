package harness

import (
	"crypto/sha256"
	"runtime"
	"strconv"
	"testing"
)

// Explicit measurement only: ordinary tests do not run benchmarks. Use
// -benchtime=1x and Go's -cpuprofile on this exact standard phase. This cannot
// substitute for the full check, registered relevance or hosted timing.
func BenchmarkStandardPacingProfile(b *testing.B) {
	if b.N != 1 {
		b.Fatal("standard pacing profile requires -benchtime=1x")
	}
	suite, err := LoadSuite("../..", "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		b.Fatal(err)
	}
	expected := map[RunKey]bool{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil {
			b.Fatal(err)
		}
		for index := range spec.SeedCount {
			key := suite.runKey(spec, start+uint64(index))
			if expected[key] {
				b.Fatal("duplicate declared standard run", key)
			}
			expected[key] = true
		}
	}
	if len(expected) != 300 {
		b.Fatal("standard population changed; reconcile the declared profile", len(expected))
	}
	b.Logf("standard phase only; workers=12 go=%s os=%s arch=%s logical_cpus=%d scenario=%s constants=%s",
		runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), suite.ScenarioHash, suite.ConstantsHash)
	b.ResetTimer()
	runs, aggregate, err := suite.RunAllWithWorkers(12)
	b.StopTimer()
	if err != nil || len(runs) != len(expected) || aggregate.RunCount != len(expected) || len(aggregate.Failures) != 0 {
		b.Fatalf("incomplete/failed phase: runs=%d aggregate=%+v err=%v", len(runs), aggregate, err)
	}
	for _, report := range runs {
		if !expected[report.Key] || report.Outcome != "completed" || len(report.InvariantFailures) != 0 {
			b.Fatal("missing/repeated/failed governed run", report)
		}
		delete(expected, report.Key)
	}
	if len(expected) != 0 {
		b.Fatal("declared standard runs remain unexecuted", len(expected))
	}
	data, err := CanonicalJSON(SuiteReport{SchemaVersion: 1, Runs: runs, Aggregate: aggregate})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportMetric(float64(len(runs)), "runs/op")
	b.Logf("complete standard population=%d report_sha256=%x; not a full-check verdict", len(runs), sha256.Sum256(data))
}
