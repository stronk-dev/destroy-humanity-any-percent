package harness

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"testing"

	"cloud-clicker/server/reputation"
)

var careerMeasurementMode = flag.String("reputation-career-measurement", "off", "off, record, or verify: complete H4/H5 observation, not acceptance")

var careerMeasurementPaths = []string{
	currentReputationPrefix + "career-h4.2026-10-06.v1.json",
	currentReputationPrefix + "relevance-h5.2026-10-06.v1.json",
	currentReputationPrefix + "career-measurement-lineage.2026-10-06.v1.json",
}

// Private research provenance. A valid negative study is not H4/H5 acceptance.
type reputationCareerLineage struct {
	SchemaVersion int                           `json:"schema_version"`
	Producer      reputationMeasurementProducer `json:"producer"`
	GoVersion     string                        `json:"go_version"`
	GOOS          string                        `json:"goos"`
	GOARCH        string                        `json:"goarch"`
	H4SHA256      string                        `json:"h4_sha256"`
	H5SHA256      string                        `json:"h5_sha256"`
	Pairs         int                           `json:"h4_pairs"`
	H4Sources     int                           `json:"h4_sources"`
	H5Arms        int                           `json:"h5_arms"`
}

type reputationCareerDeclaration struct {
	treated, control ReputationCareerMeasurementSource
	seed             uint64
}

type reputationCareerMeasurementInputs struct {
	nodes []reputation.Node
	pairs []reputationCareerDeclaration
	arms  []ReputationCareerMeasurementSource
}

func declaredCareerMeasurementInputs(t *testing.T) reputationCareerMeasurementInputs {
	t.Helper()
	suite, experiment := currentReputationInputs(t)
	bundle := reputationCareerBundle(t, suite)
	inputs := reputationCareerMeasurementInputs{nodes: bundle.ReputationTree.Nodes()}
	population := map[string]int{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		policy := CareerCheapest
		if spec.PolicyID == "chaos.t0_t1" {
			policy = CareerSeededUniform
		}
		config := ReputationCareerConfig{Bundle: bundle, Threshold: reputationCareerFixtureThreshold, Policy: policy}
		prestige := *suite.Bundle.Prestige
		prestige.Threshold = config.Threshold
		for offset := 0; offset < spec.SeedCount; offset++ {
			seed := start + uint64(offset)
			source := func(config ReputationCareerConfig) ReputationCareerMeasurementSource {
				got, err := describeReputationCareerSource(suite, spec, seed, experiment, config, &prestige)
				if err != nil {
					t.Fatal(err)
				}
				return got
			}
			treated := source(config)
			controlConfig := config
			controlConfig.Policy = CareerNone
			inputs.pairs = append(inputs.pairs, reputationCareerDeclaration{treated, source(controlConfig), seed})
			inputs.arms = append(inputs.arms, treated)
			for _, node := range inputs.nodes {
				masked := config
				masked.Exclude = node.NodeID
				inputs.arms = append(inputs.arms, source(masked))
			}
			population[spec.PolicyID]++
		}
	}
	if !reflect.DeepEqual(population, map[string]int{"chaos.t0_t1": 64, "casual.t0_t1": 32, "reference.greedy": 1}) || len(inputs.nodes) != 9 {
		t.Fatal("declared career study population differs from predeclared 64/32/1 and nine masks")
	}
	return inputs
}

func validateCareerMeasurement(inputs reputationCareerMeasurementInputs, h4, h5 []byte, lineage reputationCareerLineage,
	producer reputationMeasurementProducer) error {
	if lineage.SchemaVersion != 1 || lineage.Producer != producer || lineage.GoVersion == "" || lineage.GOOS == "" || lineage.GOARCH == "" {
		return errors.New("career lineage identity mismatch")
	}
	if lineage.H4SHA256 != reputationMeasurementSHA(h4) || lineage.H5SHA256 != reputationMeasurementSHA(h5) {
		return errors.New("career raw report hash mismatch")
	}
	var career reputationCareerReport
	var relevance reputationRelevanceReport
	if err := reputationMeasurementJSON(h4, &career); err != nil {
		return err
	}
	if err := reputationMeasurementJSON(h5, &relevance); err != nil {
		return err
	}
	if len(inputs.pairs) == 0 || len(career.Seeds) != len(inputs.pairs) || len(relevance.Arms) != len(inputs.arms) ||
		lineage.Pairs != len(inputs.pairs) || lineage.H4Sources != len(inputs.pairs)*2 || lineage.H5Arms != len(inputs.arms) {
		return errors.New("career measurement census differs from declared complete population")
	}
	known := map[string]string{}
	for _, node := range inputs.nodes {
		known[node.NodeID] = node.Kind
	}
	for index, declared := range inputs.pairs {
		row := career.Seeds[index]
		if row.TreatedSource != declared.treated || row.ControlSource != declared.control || row.PolicyID != declared.treated.RunKey.PolicyID ||
			row.Seed != declared.seed || row.CareerPolicy != string(declared.treated.PurchasePolicy) {
			return fmt.Errorf("H4 pair %d differs from declared source/order", index)
		}
		starters := map[string]bool{}
		for _, id := range row.AppliedStarterIDs {
			if known[id] != reputation.KindStarter || starters[id] || !slices.Contains(row.PurchasedNodeIDs, id) {
				return errors.New("H4 applied starter is unknown, duplicated or unpurchased")
			}
			starters[id] = true
		}
		baseline := relevance.Arms[index*(len(inputs.nodes)+1)]
		if baseline.Source != row.TreatedSource || !reflect.DeepEqual(baseline.Gate, row.TreatedGateMS) ||
			!slices.Equal(baseline.Purchased, row.PurchasedNodeIDs) {
			return errors.New("H4 treatment contradicts its H5 baseline")
		}
	}
	for index, declared := range inputs.arms {
		if relevance.Arms[index].Source != declared {
			return fmt.Errorf("H5 arm %d differs from declared source/order", index)
		}
	}
	rebuilt, err := newReputationCareerReport(career.Seeds)
	if err != nil {
		return err
	}
	expected, err := CanonicalJSON(rebuilt)
	if err != nil {
		return err
	}
	actual, err := CanonicalJSON(career)
	if err != nil || !bytes.Equal(expected, actual) {
		return errors.New("H4 retained report differs from seed recomposition")
	}
	return validateReputationRelevanceRecomposition(inputs.nodes, relevance)
}

func careerMeasurementArtifacts(t *testing.T) ([]byte, []byte, reputationCareerLineage) {
	t.Helper()
	var data [3][]byte
	for index, path := range careerMeasurementPaths {
		var err error
		data[index], err = os.ReadFile(filepath.Join(repositoryRootForReputation, path))
		if err != nil {
			t.Fatal(err)
		}
	}
	var lineage reputationCareerLineage
	if err := reputationMeasurementJSON(data[2], &lineage); err != nil {
		t.Fatal(err)
	}
	return data[0], data[1], lineage
}

func TestReputationCareerCurrentMeasurement(t *testing.T) {
	mode := *careerMeasurementMode
	if mode == "off" {
		t.Skip("fresh career measurement requires record or verify; artifact validation is not freshness")
	}
	if mode != "record" && mode != "verify" {
		t.Fatalf("unknown career measurement selector %q", mode)
	}
	if err := reputationMeasurementCleanInputs(); err != nil {
		t.Fatal(err)
	}
	producer, err := reputationMeasurementIdentity("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	var retainedH4, retainedH5 []byte
	var retainedLineage reputationCareerLineage
	if mode == "record" {
		for _, path := range careerMeasurementPaths {
			if _, err := os.Lstat(filepath.Join(repositoryRootForReputation, path)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("record refuses existing/unresolved output %s: %v", path, err)
			}
		}
	} else {
		retainedH4, retainedH5, retainedLineage = careerMeasurementArtifacts(t)
		if err := reputationMeasurementSameInputs(retainedLineage.Producer, producer); err != nil {
			t.Fatal(err)
		}
		if retainedLineage.GoVersion != runtime.Version() || retainedLineage.GOOS != runtime.GOOS || retainedLineage.GOARCH != runtime.GOARCH {
			t.Fatal("fresh replay runtime differs from recorded career measurement")
		}
	}
	inputs := declaredCareerMeasurementInputs(t)
	if mode == "verify" {
		recordedProducer, err := reputationMeasurementIdentity(retainedLineage.Producer.Commit)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateCareerMeasurement(inputs, retainedH4, retainedH5, retainedLineage, recordedProducer); err != nil {
			t.Fatal(err)
		}
	}
	h4Report := measureReputationCareerReport(t)
	h5Report := measureReputationRelevanceReport(t)
	h4, err := CanonicalJSON(h4Report)
	if err != nil {
		t.Fatal(err)
	}
	h5, err := CanonicalJSON(h5Report)
	if err != nil {
		t.Fatal(err)
	}
	lineage := reputationCareerLineage{SchemaVersion: 1, Producer: producer, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		H4SHA256: reputationMeasurementSHA(h4), H5SHA256: reputationMeasurementSHA(h5), Pairs: len(h4Report.Seeds), H4Sources: len(h4Report.Seeds) * 2, H5Arms: len(h5Report.Arms)}
	if err := validateCareerMeasurement(inputs, h4, h5, lineage, producer); err != nil {
		t.Fatal(err)
	}
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
		if !bytes.Equal(retainedH4, h4) || !bytes.Equal(retainedH5, h5) {
			t.Fatal("fresh complete career producer differs from retained H4/H5 bytes")
		}
	} else {
		companion, err := CanonicalJSON(lineage)
		if err != nil {
			t.Fatal(err)
		}
		for index, data := range [][]byte{h4, h5, companion} {
			file, err := os.OpenFile(filepath.Join(repositoryRootForReputation, careerMeasurementPaths[index]), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				t.Fatal(err)
			}
			written, writeErr := file.Write(data)
			closeErr := file.Close()
			if written != len(data) || writeErr != nil || closeErr != nil {
				t.Fatalf("incomplete career measurement write: %d/%d %v / %v", written, len(data), writeErr, closeErr)
			}
		}
	}
	t.Logf("fresh %s: producer=%s pairs=%d sources=%d H5arms=%d H4=%s H5=%s H4passed=%t violations=%d (not acceptance)",
		mode, producer.Commit, lineage.Pairs, lineage.H4Sources, lineage.H5Arms, lineage.H4SHA256, lineage.H5SHA256, h4Report.GatePassed, len(h4Report.Violations))
}
