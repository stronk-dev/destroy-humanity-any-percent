package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
)

func TestSourceFingerprintTracksExecutableAuthoritiesOnly(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources := make(map[string][]byte)
	for _, authority := range formulaAuthorities {
		if _, exists := sources[authority.path]; exists {
			continue
		}
		source, err := os.ReadFile(filepath.Join(root, authority.path))
		if err != nil {
			t.Fatal(err)
		}
		sources[authority.path] = source
	}
	read := func(path string) ([]byte, error) { return sources[path], nil }
	baseline, err := sourceFingerprintFrom(read)
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline) != 64 {
		t.Fatalf("fingerprint = %q", baseline)
	}

	originalEngine := append([]byte(nil), sources["production/engine.go"]...)
	sources["production/engine.go"] = append([]byte("// formula provenance comment\n"), originalEngine...)
	commentOnly, err := sourceFingerprintFrom(read)
	if err != nil || commentOnly != baseline {
		t.Fatalf("comment changed fingerprint: got %s want %s err=%v", commentOnly, baseline, err)
	}
	sources["production/engine.go"] = bytes.Replace(originalEngine, []byte("func ratesWithProvisionedAndPolicy("), []byte("func  ratesWithProvisionedAndPolicy ("), 1)
	formatOnly, err := sourceFingerprintFrom(read)
	if err != nil || formatOnly != baseline {
		t.Fatalf("formatting changed fingerprint: got %s want %s err=%v", formatOnly, baseline, err)
	}

	sources["production/engine.go"] = bytes.Replace(originalEngine,
		[]byte("rate = rate.Mul(bySource[identity].Factor)"),
		[]byte("rate = rate.Add(bySource[identity].Factor)"), 1)
	if bytes.Equal(sources["production/engine.go"], originalEngine) {
		t.Fatal("executable mutation fixture did not modify Rates")
	}
	changed, err := sourceFingerprintFrom(read)
	if err != nil {
		t.Fatal(err)
	}
	if changed == baseline {
		t.Fatal("executable rate change did not alter fingerprint")
	}

	sources["production/engine.go"] = bytes.Replace(originalEngine,
		[]byte("nextBoundary += tickMS"), []byte("nextBoundary += tickMS + 1"), 1)
	boundaryChanged, err := sourceFingerprintFrom(read)
	if err != nil || boundaryChanged == baseline {
		t.Fatalf("provision boundary change was not fingerprinted: got %s baseline %s err=%v", boundaryChanged, baseline, err)
	}
	sources["production/engine.go"] = originalEngine

	originalFaction := append([]byte(nil), sources["faction/hook.go"]...)
	sources["faction/hook.go"] = bytes.Replace(originalFaction,
		[]byte("factorPPM.Add(factorPPM"), []byte("factorPPM.Sub(factorPPM"), 1)
	if bytes.Equal(sources["faction/hook.go"], originalFaction) {
		t.Fatal("stock-rate mutation fixture did not modify AfterAccrual")
	}
	stockChanged, err := sourceFingerprintFrom(read)
	if err != nil || stockChanged == baseline {
		t.Fatalf("stock-rate change was not fingerprinted: got %s baseline %s err=%v", stockChanged, baseline, err)
	}
	sources["faction/hook.go"] = originalFaction

	sources["production/engine.go"] = []byte(strings.ReplaceAll(string(originalEngine), "func ratesWithProvisionedAndPolicy(", "func renamedRatesWithProvisionedAndPolicy("))
	if _, err := sourceFingerprintFrom(read); err == nil {
		t.Fatal("missing authority did not fail closed")
	}
}

func TestPurchasableArtifactPublishesCatalogCompositionAndCapReasons(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "economy-foundation-v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	artifact := contentFormulaFor(catalog)
	if artifact.ProvisionTickMS != 60_000 || len(artifact.ProvisionedHardcaps) != 1 || artifact.ProvisionedHardcaps[0].ReasonKey != "generator.low.provisioned_cap" || len(artifact.SynergyPools) != 1 || artifact.SynergyPools[0].ID != "pool.operations" || len(artifact.SynergyPools[0].Sources) != 2 {
		t.Fatalf("purchasable artifact=%+v", artifact)
	}
}

func TestGeneratedAuthoritiesMatchPublishedRuntimeTokens(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sourceFingerprint(root); err != nil {
		t.Fatal(err)
	}
}

func TestCommonsArtifactPublishesEveryCatalogControl(t *testing.T) {
	artifactType := commonsFormula{}
	data, err := json.Marshal(artifactType)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{
		"source_weights", "default_tithe_ppm", "minimum_tithe_ppm", "maximum_tithe_ppm",
		"guild_health_weight_ppm", "cohort_health_weight_ppm", "server_health_weight_ppm",
		"collective_weight_ppm", "collective_exponent_ppm", "collapse_health_ppm", "healthy_health_ppm", "maximum_bonus",
		"health_recovery_ppm_per_hour", "health_decay_ppm_per_hour", "solidarity_window_ms",
		"cohort_target_size", "cohort_merge_floor", "npc_population_floor", "npc_weight_ppm",
		"npc_compliance_ppm", "population_tolerance_ppm",
	} {
		if !bytes.Contains(data, []byte(`"`+field+`"`)) {
			t.Fatalf("Commons artifact omitted %s", field)
		}
	}
}
