package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/save"
)

func TestEpochGuardRejectsUnregisteredArtifactChange(t *testing.T) {
	root, _, artifacts := newEpochGuardRepository(t)
	writeGuardCommit(t, root, "economy: silent retune", map[string]string{
		artifacts["commons"]: `{"changed":true}`,
	})
	if err := ValidateRepositoryEpochChanges(root); err == nil || !strings.Contains(err.Error(), "without "+epochSeedPath) {
		t.Fatalf("unregistered artifact err=%v", err)
	}
}

func TestEpochGuardAcceptsRegisteredHotfix(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	commons := `{"corrected":true}`
	seed.Epochs[0].AcceptedHashes = append(seed.Epochs[0].AcceptedHashes,
		epochHash(t, root, seed, map[string]string{artifacts["commons"]: commons}))
	sort.Strings(seed.Epochs[0].AcceptedHashes)
	writeGuardCommit(t, root, "commons: correct catalog typo", map[string]string{
		artifacts["commons"]: commons,
		epochSeedPath:        encodeEpochSeed(t, seed),
	})
	if err := ValidateRepositoryEpochChanges(root); err != nil {
		t.Fatal(err)
	}
}

func TestConstantsIdentityPinsArtifactBytesAtPreviousBaseline(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	initialHash := seed.Epochs[0].AcceptedHashes[0]
	baseline := AggregateReport{SchemaVersion: 1, ConstantsHash: initialHash}
	golden := GoldenReport{SchemaVersion: 1, Runs: []RunReport{{Key: RunKey{ConstantsHash: initialHash}}}}
	baselineBytes, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	goldenBytes, err := json.Marshal(golden)
	if err != nil {
		t.Fatal(err)
	}
	writeGuardCommit(t, root, "harness: initial baseline", map[string]string{
		baselinePath: string(baselineBytes),
		goldenPath:   string(goldenBytes),
	})
	previousBaselineBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	previousBaseline := string(previousBaselineBytes)

	if err := validateConstantsIdentityArtifactBytes(root, previousBaseline, previousBaseline, seed); err != nil {
		t.Fatalf("unchanged identity bytes failed: %v", err)
	}

	prestige := `{"prestige_cost":"halved"}`
	resultingHash := epochHash(t, root, seed, map[string]string{artifacts["prestige"]: prestige})
	seed.Epochs[0].AcceptedHashes = append(seed.Epochs[0].AcceptedHashes,
		resultingHash)
	sort.Strings(seed.Epochs[0].AcceptedHashes)
	writeGuardCommit(t, root, "prestige: hotfix cost", map[string]string{
		artifacts["prestige"]: prestige,
		epochSeedPath:         encodeEpochSeed(t, seed),
	})
	baseline.ConstantsHash = resultingHash
	golden.Runs[0].Key.ConstantsHash = resultingHash
	baselineBytes, err = json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	goldenBytes, err = json.Marshal(golden)
	if err != nil {
		t.Fatal(err)
	}
	writeGuardCommit(t, root, "CONSTANTS-IDENTITY: hide prestige retune", map[string]string{
		baselinePath: string(baselineBytes),
		goldenPath:   string(goldenBytes),
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil ||
		!strings.Contains(err.Error(), "prestige") {
		t.Fatalf("identity guard accepted changed prestige bytes: %v", err)
	}
}

func TestConstantsIdentityCannotRewriteInactiveRelevanceBehavior(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	initialHash := seed.Epochs[0].AcceptedHashes[0]
	baselineBytes, _ := json.Marshal(AggregateReport{SchemaVersion: 1, ConstantsHash: initialHash})
	goldenBytes, _ := json.Marshal(GoldenReport{SchemaVersion: 1, Runs: []RunReport{{Key: RunKey{ConstantsHash: initialHash}}}})
	relevancePath := "testdata/harness/relevance/identity-report.json"
	relevanceBytes, _ := json.Marshal(RelevanceReport{SchemaVersion: 1, ConstantsHash: "sha256:local-fixture"})
	registry := `{"schema_version":1,"entries":[{"economy_catalog":"` + artifacts["economy"] + `","scenario":"testdata/harness/relevance/scenario-v1.json","relevance_policy":"testdata/harness/relevance/policy-v1.json","golden_report":"` + relevancePath + `","justification_changelog":"changelog/epoch-1.md"}]}`
	writeGuardCommit(t, root, "harness: initial baseline", map[string]string{
		baselinePath:          string(baselineBytes),
		goldenPath:            string(goldenBytes),
		relevancePath:         string(relevanceBytes),
		relevanceRegistryPath: registry,
	})
	rewrittenBytes, _ := json.Marshal(RelevanceReport{SchemaVersion: 1, ConstantsHash: "sha256:local-fixture", Failures: []string{"hidden.rewrite"}})
	writeGuardCommit(t, root, "CONSTANTS-IDENTITY: rewrite relevance behavior", map[string]string{
		relevancePath: string(rewrittenBytes),
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "inactive relevance report") {
		t.Fatalf("identity guard accepted inactive relevance rewrite: %v", err)
	}
}

func TestConstantsIdentityAllowsOnlyActiveRelevanceHashRepair(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	economyBytes, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	policyBytes, err := os.ReadFile("../../testdata/harness/relevance/policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	policyPath := "balance/relevance/phase0.json"
	seed.Artifacts = []epochArtifact{
		{Name: "commons", Path: artifacts["commons"]},
		{Name: "economy", Path: artifacts["economy"]},
		{Name: "prestige", Path: artifacts["prestige"]},
		{Name: "relevance_policy", Path: policyPath},
		{Name: "routes", Path: artifacts["routes"]},
	}
	expectedHash := epochHash(t, root, seed, map[string]string{
		artifacts["economy"]: string(economyBytes),
		policyPath:           string(policyBytes),
	})
	seed.Epochs[0].AcceptedHashes = []string{expectedHash}
	baselineBytes, _ := json.Marshal(AggregateReport{SchemaVersion: 1, ConstantsHash: expectedHash})
	goldenBytes, _ := json.Marshal(GoldenReport{SchemaVersion: 1, Runs: []RunReport{{Key: RunKey{ConstantsHash: expectedHash}}}})
	relevancePath := "testdata/harness/relevance/identity-report.json"
	oldHash := "sha256:" + strings.Repeat("0", 64)
	relevanceBytes, _ := json.Marshal(RelevanceReport{SchemaVersion: 1, ConstantsHash: oldHash})
	scenarioPath := "testdata/harness/relevance/scenario-v1.json"
	scenario := `{"catalog":"` + artifacts["economy"] + `","routes_catalog":"` + artifacts["routes"] + `","relevance_policy":"` + policyPath + `"}`
	registry := `{"schema_version":1,"entries":[{"economy_catalog":"` + artifacts["economy"] + `","scenario":"` + scenarioPath + `","relevance_policy":"` + policyPath + `","golden_report":"` + relevancePath + `","justification_changelog":"changelog/epoch-1.md"}]}`
	writeGuardCommit(t, root, "harness: initial active relevance baseline", map[string]string{
		artifacts["economy"]:  string(economyBytes),
		policyPath:            string(policyBytes),
		epochSeedPath:         encodeEpochSeed(t, seed),
		baselinePath:          string(baselineBytes),
		goldenPath:            string(goldenBytes),
		relevancePath:         string(relevanceBytes),
		relevanceRegistryPath: registry,
		scenarioPath:          scenario,
	})
	repairedBytes, _ := json.Marshal(RelevanceReport{SchemaVersion: 1, ConstantsHash: expectedHash})
	writeGuardCommit(t, root, "CONSTANTS-IDENTITY: repair active relevance hash", map[string]string{
		relevancePath: string(repairedBytes),
	})
	if err := ValidateRepositoryBaselineChange(root); err != nil {
		t.Fatal(err)
	}
}

func TestEpochGuardRejectsHardcapLoweringAsHotfix(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	economyBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifacts["economy"])))
	if err != nil {
		t.Fatal(err)
	}
	lowered := strings.Replace(string(economyBytes), `"1e1000"`, `"1e999"`, 1)
	seed.Epochs[0].AcceptedHashes = append(seed.Epochs[0].AcceptedHashes,
		epochHash(t, root, seed, map[string]string{artifacts["economy"]: lowered}))
	sort.Strings(seed.Epochs[0].AcceptedHashes)
	writeGuardCommit(t, root, "economy: correctness-only cap change", map[string]string{
		artifacts["economy"]: lowered,
		epochSeedPath:        encodeEpochSeed(t, seed),
	})
	if err := ValidateRepositoryEpochChanges(root); err == nil || !strings.Contains(err.Error(), "hotfix lowers hardcaps company.cash") {
		t.Fatalf("lowered hotfix err=%v", err)
	}
}

func TestEpochGuardAcceptsMintWithCapMigrationPolicy(t *testing.T) {
	root, seed, artifacts := newEpochGuardRepository(t)
	economyBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifacts["economy"])))
	if err != nil {
		t.Fatal(err)
	}
	lowered := strings.Replace(string(economyBytes), `"1e1000"`, `"1e999"`, 1)
	resultingHash := epochHash(t, root, seed, map[string]string{artifacts["economy"]: lowered})
	seed.CurrentEpochID = 2
	seed.Epochs = append(seed.Epochs, epochRecord{
		ID: 2, Name: "Phase 0.1", ChangelogRef: "changelog/epoch-2.md", AcceptedHashes: []string{resultingHash},
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: lower Phase-0 cap", map[string]string{
		artifacts["economy"]:   lowered,
		epochSeedPath:          encodeEpochSeed(t, seed),
		"changelog/epoch-2.md": "# Epoch 2\n\nCap migration: balances above the new cap saturate once during save migration.\n",
	})
	if err := ValidateRepositoryEpochChanges(root); err != nil {
		t.Fatal(err)
	}
}

func TestEpochGuardAcceptsArtifactAdditionOnlyInMint(t *testing.T) {
	root, seed, _ := newEpochGuardRepository(t)
	factionsPath := "balance/factions/phase0.json"
	factions := `{"schema_version":1,"factions":[]}`
	seed.Artifacts = append(seed.Artifacts, epochArtifact{Name: "factions", Path: factionsPath})
	resultingHash := epochHash(t, root, seed, map[string]string{factionsPath: factions})
	seed.CurrentEpochID = 2
	seed.Epochs = append(seed.Epochs, epochRecord{
		ID: 2, Name: "Faction foundation", ChangelogRef: "changelog/epoch-2.md", AcceptedHashes: []string{resultingHash},
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: register faction catalog", map[string]string{
		factionsPath:           factions,
		epochSeedPath:          encodeEpochSeed(t, seed),
		"changelog/epoch-2.md": "# Epoch 2\n\nRegisters the faction catalog.\n",
	})
	if err := ValidateRepositoryEpochChanges(root); err != nil {
		t.Fatal(err)
	}
}

func TestEpochGuardRejectsArtifactAdditionAsHotfix(t *testing.T) {
	root, seed, _ := newEpochGuardRepository(t)
	factionsPath := "balance/factions/phase0.json"
	factions := `{"schema_version":1,"factions":[]}`
	seed.Artifacts = append(seed.Artifacts, epochArtifact{Name: "factions", Path: factionsPath})
	seed.Epochs[0].AcceptedHashes = append(seed.Epochs[0].AcceptedHashes,
		epochHash(t, root, seed, map[string]string{factionsPath: factions}))
	sort.Strings(seed.Epochs[0].AcceptedHashes)
	writeGuardCommit(t, root, "factions: register catalog", map[string]string{
		factionsPath:  factions,
		epochSeedPath: encodeEpochSeed(t, seed),
	})
	if err := ValidateRepositoryEpochChanges(root); err == nil || !strings.Contains(err.Error(), "BALANCE-CHANGE successor mint") {
		t.Fatalf("artifact-addition hotfix err=%v", err)
	}
}

func TestEpochGuardRejectsSeedOnlyAndMintWithoutChangelog(t *testing.T) {
	root, seed, _ := newEpochGuardRepository(t)
	seed.Epochs[0].Name = "renamed"
	writeGuardCommit(t, root, "docs: rename epoch", map[string]string{epochSeedPath: encodeEpochSeed(t, seed)})
	if err := ValidateRepositoryEpochChanges(root); err == nil || !strings.Contains(err.Error(), "without a constants artifact") {
		t.Fatalf("seed-only err=%v", err)
	}

	root, seed, artifacts := newEpochGuardRepository(t)
	commons := `{"balance":2}`
	resultingHash := epochHash(t, root, seed, map[string]string{artifacts["commons"]: commons})
	seed.CurrentEpochID = 2
	seed.Epochs = append(seed.Epochs, epochRecord{ID: 2, Name: "Phase 0.1", ChangelogRef: "changelog/epoch-2.md", AcceptedHashes: []string{resultingHash}})
	writeGuardCommit(t, root, "BALANCE-CHANGE: unannounced epoch", map[string]string{
		artifacts["commons"]: commons,
		epochSeedPath:        encodeEpochSeed(t, seed),
	})
	if err := ValidateRepositoryEpochChanges(root); err == nil || !strings.Contains(err.Error(), "changelog/epoch-2.md") {
		t.Fatalf("missing changelog err=%v", err)
	}
}

func newEpochGuardRepository(t *testing.T) (string, epochSeed, map[string]string) {
	t.Helper()
	root := t.TempDir()
	runGuardGit(t, root, "init", "--quiet", "--initial-branch=main")
	runGuardGit(t, root, "config", "user.name", "Epoch Guard Test")
	runGuardGit(t, root, "config", "user.email", "epoch@example.invalid")
	artifacts := map[string]string{
		"commons":  "balance/commons/phase0.json",
		"economy":  "balance/catalogs/phase0.json",
		"prestige": "balance/prestige/phase0.json",
		"routes":   "balance/routes/phase0.json",
	}
	contents := map[string][]byte{}
	for name, path := range artifacts {
		data, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		contents[name] = data
		writeGuardFile(t, root, path, string(data))
	}
	hash, err := save.ConstantsHashArtifacts(contents)
	if err != nil {
		t.Fatal(err)
	}
	seed := epochSeed{
		SchemaVersion: 1, CurrentEpochID: 1,
		Artifacts: []epochArtifact{
			{Name: "commons", Path: artifacts["commons"]},
			{Name: "economy", Path: artifacts["economy"]},
			{Name: "prestige", Path: artifacts["prestige"]},
			{Name: "routes", Path: artifacts["routes"]},
		},
		Epochs: []epochRecord{{ID: 1, Name: "Phase 0", ChangelogRef: "changelog/epoch-1.md", AcceptedHashes: []string{hash}}},
	}
	writeGuardFile(t, root, epochSeedPath, encodeEpochSeed(t, seed))
	writeGuardFile(t, root, "changelog/epoch-1.md", "# Epoch 1\n")
	writeGuardFile(t, root, relevanceRegistryPath, `{"schema_version":1,"entries":[{"economy_catalog":"balance/catalogs/phase0.json","scenario":"testdata/harness/relevance/scenario-v1.json","relevance_policy":"testdata/harness/relevance/policy-v1.json","golden_report":"testdata/harness/pacing-baseline.json","justification_changelog":"changelog/epoch-1.md"}]}`)
	runGuardGit(t, root, "add", ".")
	runGuardGit(t, root, "commit", "--quiet", "-m", "epochs: seed Phase 0")
	return root, seed, artifacts
}

func epochHash(t *testing.T, root string, seed epochSeed, replacements map[string]string) string {
	t.Helper()
	contents := make(map[string][]byte, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		if replacement, ok := replacements[artifact.Path]; ok {
			contents[artifact.Name] = []byte(replacement)
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.Path)))
		if err != nil {
			t.Fatal(err)
		}
		contents[artifact.Name] = data
	}
	hash, err := save.ConstantsHashArtifacts(contents)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func encodeEpochSeed(t *testing.T, seed epochSeed) string {
	t.Helper()
	data, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(data) + "\n"
}
