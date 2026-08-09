package replaycatalog

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/save"
)

func TestFirstContentCandidateBundleLoadsFromLiteralArtifacts(t *testing.T) {
	artifacts := firstContentCandidateArtifacts(t)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	manifest := firstContentCandidateManifest(t)
	if hash != manifest.ConstantsHash {
		t.Fatalf("candidate constants hash=%s manifest=%s", hash, manifest.ConstantsHash)
	}
	t.Logf("first-content candidate constants hash: %s", hash)
	bundle, err := Load(hash, artifacts)
	if err != nil {
		t.Fatalf("load first-content candidate %s: %v", hash, err)
	}
	if bundle.Meters == nil || bundle.Achievements == nil || bundle.Doctrines == nil || bundle.Minigames == nil ||
		bundle.Pets == nil || bundle.Fiscal == nil || bundle.Soul == nil || bundle.Pitch == nil || bundle.MinigameAPI == nil {
		t.Fatalf("candidate bundle omitted an activated catalog: %+v", bundle)
	}
}

func TestFirstContentCandidateCategorySetIsLoadBearing(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "first-content", "categories-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var candidate map[string]any
	if err := json.Unmarshal(data, &candidate); err != nil {
		t.Fatal(err)
	}
	candidate["full_gate_set"] = []string{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"}
	stale, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := firstContentCandidateArtifacts(t)
	artifacts["categories"] = stale
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(hash, artifacts); err == nil {
		t.Fatal("candidate bundle accepted the pre-permits leaderboard gate set")
	}
}

func TestFirstContentCandidateRejectsMeterResourceCollision(t *testing.T) {
	artifacts := firstContentCandidateArtifacts(t)
	var economy map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economy); err != nil {
		t.Fatal(err)
	}
	economy["resources"] = append(economy["resources"].([]any), map[string]any{
		"id": "trust.users.standing", "scope": "company", "numeric_kind": "decimal", "initial": "0", "minimum": "0",
		"hardcap": map[string]any{"amount": "1e2", "reason_key": "resource.trust_users_standing.cap.fixture"},
	})
	var err error
	artifacts["economy"], err = json.Marshal(economy)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(hash, artifacts); err == nil {
		t.Fatal("Go replay loader accepted an economy resource that collides with a meter ID")
	}
}

func firstContentCandidateArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	manifest := firstContentCandidateManifest(t)
	root := filepath.Join("..", "..")
	artifacts := make(map[string][]byte, len(manifest.Artifacts))
	prior := ""
	for _, row := range manifest.Artifacts {
		if row.Name <= prior || row.Name == "" || row.SourcePath == "" || row.ProductionPath == "" || row.SchemaVersion < 1 || row.ContentGate == "" || row.ConsumedVerdict == "" {
			t.Fatalf("invalid promotion manifest row for %q", row.Name)
		}
		data, err := os.ReadFile(filepath.Join(root, row.SourcePath))
		if err != nil {
			t.Fatalf("read %s: %v", row.Name, err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(data))
		if actual != row.SHA256 {
			t.Fatalf("%s SHA-256=%s manifest=%s", row.Name, actual, row.SHA256)
		}
		artifacts[row.Name] = data
		prior = row.Name
	}
	return artifacts
}

type firstContentManifest struct {
	SchemaVersion int                       `json:"schema_version"`
	Status        string                    `json:"status"`
	ConstantsHash string                    `json:"constants_hash"`
	Artifacts     []firstContentManifestRow `json:"artifacts"`
}

type firstContentManifestRow struct {
	Name            string `json:"name"`
	SchemaVersion   int    `json:"schema_version"`
	SourcePath      string `json:"source_path"`
	ProductionPath  string `json:"production_path"`
	SHA256          string `json:"sha256"`
	ContentGate     string `json:"content_gate"`
	ConsumedVerdict string `json:"consumed_verdict"`
}

func firstContentCandidateManifest(t *testing.T) firstContentManifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "planning", "first-content-epoch", "promotion-manifest.candidate.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest firstContentManifest
	if err := json.Unmarshal(data, &manifest); err != nil || manifest.SchemaVersion != 1 || manifest.Status != "ratified" || len(manifest.Artifacts) != 16 {
		t.Fatalf("invalid first-content promotion manifest: %v", err)
	}
	return manifest
}
