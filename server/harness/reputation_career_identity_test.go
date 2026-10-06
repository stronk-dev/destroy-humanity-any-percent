package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

func cloneCareerArtifacts(artifacts map[string][]byte) map[string][]byte {
	copy := make(map[string][]byte, len(artifacts))
	for name, data := range artifacts {
		copy[name] = bytes.Clone(data)
	}
	return copy
}

// Derive the expected paired fixture from retained sources, independently of
// reputationCareerBundle's parsed objects and its reported identity.
func expectedCareerArtifacts(t *testing.T, base map[string][]byte) map[string][]byte {
	t.Helper()
	artifacts := cloneCareerArtifacts(base)
	var economy map[string]any
	if err := json.Unmarshal(base["economy"], &economy); err != nil {
		t.Fatal(err)
	}
	economy["multiplier_sources"] = append(economy["multiplier_sources"].([]any), map[string]any{
		"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": "reputation_tree",
	})
	var err error
	artifacts["economy"], err = json.Marshal(economy)
	if err != nil {
		t.Fatal(err)
	}
	artifacts["reputation_tree"], err = os.ReadFile(filepath.Join(repositoryRootForReputation, "balance/testdata/reputation-tree/fixture-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return artifacts
}

func TestReputationCareerFixtureIdentity(t *testing.T) {
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	base := cloneCareerArtifacts(suite.Bundle.Artifacts)
	baseHash := suite.ConstantsHash
	if suite.Bundle.ConstantsHash != baseHash || suite.Bundle.ReputationTree != nil || production.ReputationDeclared(suite.Bundle.Economy) {
		t.Fatal("identity control is not the tree-less ratified base suite")
	}
	want := expectedCareerArtifacts(t, base)
	wantHash, err := save.ConstantsHashArtifacts(want)
	if err != nil || wantHash == baseHash {
		t.Fatalf("independent paired fixture has no distinct complete identity: %s %v", wantHash, err)
	}
	bundle := reputationCareerBundle(t, suite)
	t.Cleanup(func() {
		if suite.ConstantsHash != baseHash || suite.Bundle.ConstantsHash != baseHash || !reflect.DeepEqual(base, suite.Bundle.Artifacts) ||
			suite.Bundle.ReputationTree != nil || production.ReputationDeclared(suite.Bundle.Economy) {
			t.Error("career fixture mutated the ratified base suite")
		}
	})
	t.Run("retained-tree-bytes", func(t *testing.T) {
		if !bytes.Equal(bundle.Artifacts["reputation_tree"], want["reputation_tree"]) {
			t.Fatal("career fixture did not retain its exact tree source")
		}
	})
	t.Run("paired-economy-declaration", func(t *testing.T) {
		if !bytes.Equal(bundle.Artifacts["economy"], want["economy"]) {
			t.Fatal("career artifact does not contain the paired economy declaration")
		}
	})
	t.Run("complete-hash-and-public-loader", func(t *testing.T) {
		if !reflect.DeepEqual(bundle.Artifacts, want) || bundle.ConstantsHash != wantHash {
			t.Fatalf("career fixture identity does not name all actual sources: got %s want %s", bundle.ConstantsHash, wantHash)
		}
		loaded, err := replaycatalog.Load(wantHash, bundle.Artifacts)
		if err != nil || !reflect.DeepEqual(loaded, bundle) {
			t.Fatalf("career fixture differs from its public loader roundtrip: %v", err)
		}
	})
	t.Run("actual-career-run-key", func(t *testing.T) {
		_, spec, experiment, config := reputationCareerAdmissionInputs(t)
		config.Bundle, config.Policy = bundle, CareerNone
		result, err := suite.RunReputationCareer(spec, 0, experiment, config)
		if err != nil || result.Run.Outcome != "completed" || len(result.Run.ReputationExits) != 2 || result.RunThreeGateMS == nil {
			t.Fatalf("identity witness lost its completed Chaos seed0 career: err=%v outcome=%s exits=%d gate=%v", err, result.Run.Outcome, len(result.Run.ReputationExits), result.RunThreeGateMS)
		}
		if len(result.PurchasedNodeIDs) != 0 || len(result.AppliedStarterIDs) != 0 || result.BonusFactor != "1e0" {
			t.Fatal("identity control acquired a Reputation effect")
		}
		if result.Run.Key.ConstantsHash != wantHash {
			t.Fatalf("actual career run key names the wrong inputs: got %s want %s", result.Run.Key.ConstantsHash, wantHash)
		}
	})
	for _, mutation := range []string{"removed-tree", "undeclared-source", "false-hash"} {
		t.Run("refuses-"+mutation, func(t *testing.T) {
			artifacts := cloneCareerArtifacts(want)
			switch mutation {
			case "removed-tree":
				delete(artifacts, "reputation_tree")
			case "undeclared-source":
				artifacts["economy"] = bytes.Clone(base["economy"])
			}
			hash, err := save.ConstantsHashArtifacts(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			if mutation == "false-hash" {
				hash = baseHash
			}
			if _, err := replaycatalog.Load(hash, artifacts); err == nil {
				t.Fatal("public loader admitted an incomplete or falsely named career fixture")
			}
		})
	}
}
