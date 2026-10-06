package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const currentReputationPrefix = "planning/reputation-tree-v1/"

var currentReputationPaths = []string{
	currentReputationPrefix + "first-hour-reputation.2026-10-06.v1.json",
	currentReputationPrefix + "threshold-measurement.2026-10-06.v1.json",
	currentReputationPrefix + "measurement-lineage.2026-10-06.v1.json",
}

var currentReputationMode = flag.String("reputation-current-measurement", "off", "off, record, or verify: full 97-run H1/H2 measurement, not balance adoption")

// Private research companion, not a production wire/schema authority. Git trees
// cover all server dependencies and balance inputs, not just the harness file.
type reputationMeasurementProducer struct {
	Commit        string `json:"commit"`
	ServerTree    string `json:"server_tree"`
	BalanceTree   string `json:"balance_tree"`
	KernelVersion string `json:"kernel_version"`
}

type reputationMeasurementLineage struct {
	SchemaVersion int                           `json:"schema_version"`
	Producer      reputationMeasurementProducer `json:"producer"`
	GoVersion     string                        `json:"go_version"`
	GOOS          string                        `json:"goos"`
	GOARCH        string                        `json:"goarch"`
	H1SHA256      string                        `json:"h1_sha256"`
	H2SHA256      string                        `json:"h2_sha256"`
	RunCount      int                           `json:"run_count"`
	ClockCount    int                           `json:"clock_count"`
}

func reputationMeasurementSHA(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func reputationMeasurementJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return errors.New("measurement has trailing JSON")
	}
	return nil
}

func reputationMeasurementGit(args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = repositoryRootForReputation
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v: %w: %s", args, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func reputationMeasurementIdentity(ref string) (reputationMeasurementProducer, error) {
	var producer reputationMeasurementProducer
	var err error
	producer.Commit, err = reputationMeasurementGit("rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return producer, err
	}
	producer.ServerTree, err = reputationMeasurementGit("rev-parse", producer.Commit+":server")
	if err != nil {
		return producer, err
	}
	producer.BalanceTree, err = reputationMeasurementGit("rev-parse", producer.Commit+":balance")
	if err != nil {
		return producer, err
	}
	producer.KernelVersion, err = reputationMeasurementGit("show", producer.Commit+":kernel/VERSION")
	return producer, err
}

func reputationMeasurementSameInputs(recorded, current reputationMeasurementProducer) error {
	// Commit can advance for records/docs; source/data trees and kernel may not.
	if recorded.ServerTree != current.ServerTree || recorded.BalanceTree != current.BalanceTree || recorded.KernelVersion != current.KernelVersion {
		return errors.New("current producer/data trees differ from recorded measurement")
	}
	return nil
}

func reputationMeasurementCleanInputs() error {
	status, err := reputationMeasurementGit("status", "--porcelain", "--untracked-files=all", "--", "server", "balance", "kernel/VERSION")
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("measurement inputs are not committed and clean: %s", status)
	}
	return nil
}

func validateCurrentReputationMeasurement(suite *FirstHourSuite, experiment FirstHourExperiment, h1, h2 []byte,
	lineage reputationMeasurementLineage, producer reputationMeasurementProducer) error {
	if lineage.SchemaVersion != 1 || lineage.Producer != producer || lineage.GoVersion == "" || lineage.GOOS == "" || lineage.GOARCH == "" {
		return errors.New("measurement lineage identity mismatch")
	}
	if lineage.H1SHA256 != reputationMeasurementSHA(h1) || lineage.H2SHA256 != reputationMeasurementSHA(h2) {
		return errors.New("measurement raw report binding mismatch")
	}
	var report FirstHourExperimentReport
	if err := reputationMeasurementJSON(h1, &report); err != nil {
		return err
	}
	if err := validateReputationFirstHourPopulation(suite, experiment, report); err != nil {
		return err
	}
	if lineage.RunCount != len(report.Runs) || lineage.ClockCount != len(report.Runs)*len(suite.Scenario.Milestones) {
		return errors.New("measurement lineage census mismatch")
	}
	aggregate, err := CanonicalJSON(suite.aggregateFirstHour(report.Runs))
	if err != nil {
		return err
	}
	retainedAggregate, err := CanonicalJSON(report.Aggregate)
	if err != nil {
		return err
	}
	if !bytes.Equal(aggregate, retainedAggregate) || len(report.Aggregate.Failures) != 0 {
		return errors.New("measurement aggregate drift or fired first-hour criterion")
	}
	measured, err := measureReputationThresholdStudy(suite, experiment, report, suite.Bundle.Prestige, reputationThresholdGrid())
	if err != nil {
		return err
	}
	expectedH2, err := CanonicalJSON(measured)
	if err != nil {
		return err
	}
	if !bytes.Equal(h2, expectedH2) {
		return errors.New("measurement H2 result differs from admitted H1 and pinned policy/grid")
	}
	foundLive := false
	for _, row := range measured.Rows {
		if row.Threshold == suite.Bundle.Prestige.Threshold {
			foundLive = true
			if row.Satisfies {
				return errors.New("live threshold unexpectedly satisfies envelope")
			}
			for _, persona := range row.Personas {
				if persona.Max != 0 {
					return errors.New("live threshold unexpectedly pays Reputation")
				}
			}
		}
	}
	if !foundLive {
		return errors.New("measurement omits live threshold negative control")
	}
	return nil
}

func currentReputationInputs(t *testing.T) (*FirstHourSuite, FirstHourExperiment) {
	t.Helper()
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	return suite, FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
}

func currentReputationArtifacts(t *testing.T) ([]byte, []byte, reputationMeasurementLineage) {
	t.Helper()
	var data [3][]byte
	for index, path := range currentReputationPaths {
		var err error
		data[index], err = os.ReadFile(filepath.Join(repositoryRootForReputation, path))
		if err != nil {
			t.Fatal(err)
		}
	}
	var lineage reputationMeasurementLineage
	if err := reputationMeasurementJSON(data[2], &lineage); err != nil {
		t.Fatal(err)
	}
	return data[0], data[1], lineage
}

// This fast gate validates retained artifacts only. It never claims a new run.
func TestReputationCurrentMeasurementArtifacts(t *testing.T) {
	suite, experiment := currentReputationInputs(t)
	h1, h2, lineage := currentReputationArtifacts(t)
	producer, err := reputationMeasurementIdentity(lineage.Producer.Commit)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCurrentReputationMeasurement(suite, experiment, h1, h2, lineage, producer); err != nil {
		t.Fatal(err)
	}
	t.Logf("retained artifact validation only: producer=%s runs=%d clocks=%d", producer.Commit, lineage.RunCount, lineage.ClockCount)
}

func TestReputationCurrentMeasurementAdmission(t *testing.T) {
	suite, experiment := currentReputationInputs(t)
	report, _ := reputationMeasurementInputs(t)
	measured, err := measureReputationThresholdStudy(suite, experiment, report, suite.Bundle.Prestige, reputationThresholdGrid())
	if err != nil {
		t.Fatal(err)
	}
	encode := func(value any) []byte {
		data, err := CanonicalJSON(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	h1, h2 := encode(report), encode(measured)
	producer := reputationMeasurementProducer{Commit: strings.Repeat("1", 40), ServerTree: strings.Repeat("2", 40), BalanceTree: strings.Repeat("3", 40), KernelVersion: "0.3.161"}
	lineage := reputationMeasurementLineage{SchemaVersion: 1, Producer: producer, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		H1SHA256: reputationMeasurementSHA(h1), H2SHA256: reputationMeasurementSHA(h2), RunCount: len(report.Runs), ClockCount: len(report.Runs) * len(suite.Scenario.Milestones)}
	if err := validateCurrentReputationMeasurement(suite, experiment, h1, h2, lineage, producer); err != nil {
		t.Fatalf("historical healthy admission control: %v", err)
	}
	for name, mutate := range map[string]func(*reputationMeasurementLineage){
		"schema":       func(l *reputationMeasurementLineage) { l.SchemaVersion++ },
		"commit":       func(l *reputationMeasurementLineage) { l.Producer.Commit = strings.Repeat("4", 40) },
		"server-tree":  func(l *reputationMeasurementLineage) { l.Producer.ServerTree = strings.Repeat("4", 40) },
		"balance-tree": func(l *reputationMeasurementLineage) { l.Producer.BalanceTree = strings.Repeat("4", 40) },
		"kernel":       func(l *reputationMeasurementLineage) { l.Producer.KernelVersion = "0.3.160" },
		"runtime":      func(l *reputationMeasurementLineage) { l.GoVersion = "" },
		"os":           func(l *reputationMeasurementLineage) { l.GOOS = "" },
		"arch":         func(l *reputationMeasurementLineage) { l.GOARCH = "" },
		"h1-hash":      func(l *reputationMeasurementLineage) { l.H1SHA256 = "invented" },
		"h2-hash":      func(l *reputationMeasurementLineage) { l.H2SHA256 = "invented" },
		"runs":         func(l *reputationMeasurementLineage) { l.RunCount-- },
		"clocks":       func(l *reputationMeasurementLineage) { l.ClockCount-- },
	} {
		t.Run(name, func(t *testing.T) {
			copy := lineage
			mutate(&copy)
			if err := validateCurrentReputationMeasurement(suite, experiment, h1, h2, copy, producer); err == nil {
				t.Fatal("corrupt lineage admitted")
			}
		})
	}
	for _, name := range []string{"raw-h1", "raw-h2", "aggregate", "aggregate-failure", "missing-run", "duplicate-run", "result", "envelope", "grid"} {
		t.Run(name, func(t *testing.T) {
			copy := cloneReputationThresholdSource(t, report)
			var threshold ReputationThresholdReport
			if err := json.Unmarshal(h2, &threshold); err != nil {
				t.Fatal(err)
			}
			left, right, metadata := h1, h2, lineage
			switch name {
			case "raw-h1":
				left = append(append([]byte{}, h1...), ' ')
			case "raw-h2":
				right = append(append([]byte{}, h2...), ' ')
			case "aggregate":
				copy.Aggregate.Values = nil
			case "aggregate-failure":
				copy.Aggregate.Failures = []string{"guard exhaustion"}
			case "missing-run":
				copy.Runs = copy.Runs[1:]
			case "duplicate-run":
				copy.Runs[0] = copy.Runs[1]
			case "result":
				threshold.Rows[0].Personas[0].P50++
			case "envelope":
				threshold.Envelope.Maximum++
			case "grid":
				threshold.Rows = threshold.Rows[1:]
			}
			if name != "raw-h1" && name != "raw-h2" {
				left, right = encode(copy), encode(threshold)
				// Rebind hashes: these cases must discriminate semantic corruption,
				// not just a stale hash. This is no claim of producer authenticity.
				metadata.H1SHA256, metadata.H2SHA256 = reputationMeasurementSHA(left), reputationMeasurementSHA(right)
			}
			if err := validateCurrentReputationMeasurement(suite, experiment, left, right, metadata, producer); err == nil {
				t.Fatal("corrupt report admitted")
			}
		})
	}
	advanced := producer
	advanced.Commit = strings.Repeat("4", 40)
	if err := reputationMeasurementSameInputs(producer, advanced); err != nil {
		t.Fatal("record-only commit advance changed producer trees")
	}
	for _, field := range []string{"server", "balance", "kernel"} {
		t.Run("changed-"+field, func(t *testing.T) {
			copy := producer
			switch field {
			case "server":
				copy.ServerTree = strings.Repeat("4", 40)
			case "balance":
				copy.BalanceTree = strings.Repeat("4", 40)
			case "kernel":
				copy.KernelVersion = "0.3.160"
			}
			if err := reputationMeasurementSameInputs(producer, copy); err == nil {
				t.Fatal("changed current producer inputs admitted")
			}
		})
	}
}

func TestReputationCurrentMeasurement(t *testing.T) {
	mode := *currentReputationMode
	if mode == "off" {
		t.Skip("fresh 97-run measurement requires -reputation-current-measurement=record or verify; fast artifact validation is not freshness")
	}
	if mode != "record" && mode != "verify" {
		t.Fatalf("unknown measurement selector %q", mode)
	}
	if err := reputationMeasurementCleanInputs(); err != nil {
		t.Fatal(err)
	}
	producer, err := reputationMeasurementIdentity("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	var retainedH1, retainedH2 []byte
	var retainedLineage reputationMeasurementLineage
	if mode == "record" {
		for _, path := range currentReputationPaths {
			if _, err := os.Stat(filepath.Join(repositoryRootForReputation, path)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("record refuses existing/unresolved output %s: %v", path, err)
			}
		}
	} else {
		retainedH1, retainedH2, retainedLineage = currentReputationArtifacts(t)
		if err := reputationMeasurementSameInputs(retainedLineage.Producer, producer); err != nil {
			t.Fatal(err)
		}
	}
	suite, experiment := currentReputationInputs(t)
	if mode == "verify" {
		recordedProducer, err := reputationMeasurementIdentity(retainedLineage.Producer.Commit)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateCurrentReputationMeasurement(suite, experiment, retainedH1, retainedH2, retainedLineage, recordedProducer); err != nil {
			t.Fatal(err)
		}
	}
	report, err := suite.RunAllExperiments(experiment, 8)
	if err != nil {
		t.Fatal(err)
	}
	measured, err := measureReputationThresholdStudy(suite, experiment, report, suite.Bundle.Prestige, reputationThresholdGrid())
	if err != nil {
		t.Fatal(err)
	}
	h1, err := CanonicalJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := CanonicalJSON(measured)
	if err != nil {
		t.Fatal(err)
	}
	lineage := reputationMeasurementLineage{SchemaVersion: 1, Producer: producer, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		H1SHA256: reputationMeasurementSHA(h1), H2SHA256: reputationMeasurementSHA(h2), RunCount: len(report.Runs), ClockCount: len(report.Runs) * len(suite.Scenario.Milestones)}
	if err := validateCurrentReputationMeasurement(suite, experiment, h1, h2, lineage, producer); err != nil {
		t.Fatal(err)
	}
	// Recheck after production: no caller may change inputs under a live run.
	if err := reputationMeasurementCleanInputs(); err != nil {
		t.Fatal(err)
	}
	after, err := reputationMeasurementIdentity("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := reputationMeasurementSameInputs(producer, after); err != nil {
		t.Fatal(err)
	}
	if mode == "verify" {
		if !bytes.Equal(retainedH1, h1) || !bytes.Equal(retainedH2, h2) {
			t.Fatal("fresh full producer differs from retained H1/H2 bytes")
		}
	} else {
		companion, err := CanonicalJSON(lineage)
		if err != nil {
			t.Fatal(err)
		}
		for index, data := range [][]byte{h1, h2, companion} {
			path := filepath.Join(repositoryRootForReputation, currentReputationPaths[index])
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				t.Fatal(err)
			}
			_, writeErr := file.Write(data)
			closeErr := file.Close()
			if writeErr != nil || closeErr != nil {
				t.Fatalf("incomplete measurement write: %v / %v", writeErr, closeErr)
			}
		}
	}
	t.Logf("fresh %s: producer=%s source=%s runs=%d clocks=%d H1=%s H2=%s candidates=%v (not adopted)", mode, producer.Commit,
		suite.ConstantsHash, len(report.Runs), lineage.ClockCount, lineage.H1SHA256, lineage.H2SHA256, measured.Satisfying)
}
