package replaycatalog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

func TestLoadRequiresExactSevenArtifactBundle(t *testing.T) {
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(bundle.Hash, bundle.Artifacts)
	if err != nil || loaded.ConstantsHash != bundle.Hash || loaded.Commons == nil {
		t.Fatalf("bundle=%+v err=%v", loaded, err)
	}
	missing := make(map[string][]byte, len(bundle.Artifacts)-1)
	for name, data := range bundle.Artifacts {
		if name != "guilds" {
			missing[name] = data
		}
	}
	if _, err := Load(bundle.Hash, missing); err == nil {
		t.Fatal("catalog bundle missing guilds artifact was accepted")
	}
	extra := make(map[string][]byte, len(bundle.Artifacts)+1)
	for name, data := range bundle.Artifacts {
		extra[name] = data
	}
	extra["future"] = []byte(`{}`)
	if _, err := Load(bundle.Hash, extra); err == nil {
		t.Fatal("catalog bundle with unregistered artifact was accepted")
	}
	relabeled := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := Load(relabeled, bundle.Artifacts); err == nil {
		t.Fatal("catalog bundle bytes were accepted under a false constants hash")
	}
	invalidCategories := make(map[string][]byte, len(bundle.Artifacts))
	for name, data := range bundle.Artifacts {
		invalidCategories[name] = append([]byte(nil), data...)
	}
	invalidCategories["categories"] = []byte(`{"schema_version":1}`)
	invalidHash, err := save.ConstantsHashArtifacts(invalidCategories)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(invalidHash, invalidCategories); err == nil {
		t.Fatal("invalid pinned category artifact was accepted")
	}
	loaded.Artifacts["economy"] = append(loaded.Artifacts["economy"], '\n')
	if _, ok := loaded.ResolvePrestige(bundle.Hash); ok {
		t.Fatal("mutated catalog bundle remained valid under its old constants hash")
	}
}

func TestLoadAcceptsOnlyPairedFoundationArtifacts(t *testing.T) {
	seed, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "meters-catalog-parity-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(fixture, &envelope); err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+2)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["meters"] = envelope.Baseline
	artifacts["achievements"] = []byte(`{"schema_version":1,"achievements":[{"id":"achievement.first_gate","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"provenance","event_kinds":["gate_crossed"]},"score_grant":4,"copy_key":"category.any_percent"}]}`)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(hash, artifacts)
	if err != nil || loaded.Meters == nil || loaded.Achievements == nil {
		t.Fatalf("active bundle=%+v err=%v", loaded, err)
	}
	delete(artifacts, "achievements")
	oneSidedHash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(oneSidedHash, artifacts); err == nil {
		t.Fatal("one-sided foundation artifact set was accepted")
	}
}

func TestLoadActivatesDoctrineOnlyAbovePairedFoundations(t *testing.T) {
	seed, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	meterFixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "meters-catalog-parity-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var meterEnvelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(meterFixture, &meterEnvelope); err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+3)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["meters"] = meterEnvelope.Baseline
	artifacts["achievements"] = []byte(`{"schema_version":1,"achievements":[{"id":"achievement.first_gate","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"provenance","event_kinds":["gate_crossed"]},"score_grant":4,"copy_key":"category.any_percent"}]}`)
	artifacts["doctrines"] = []byte(`{"schema_version":1,"transitions":[{"transition_id":"transition.t3_to_t4","source_tier":3,"gate_id":"gate.t3_to_t4","doctrine_ids":["doctrine.capture","doctrine.ethical"]}]}`)
	var routeRoot map[string]any
	if err := json.Unmarshal(artifacts["routes"], &routeRoot); err != nil {
		t.Fatal(err)
	}
	routeRoot["gates"] = append(routeRoot["gates"].([]any), map[string]any{"gate_id": "gate.t3_to_t4", "requirement": []any{map[string]any{"resource_id": "company.cash", "amount": "1e12"}}, "routes": []any{}})
	artifacts["routes"], _ = json.Marshal(routeRoot)
	var categories map[string]any
	if err := json.Unmarshal(artifacts["categories"], &categories); err != nil {
		t.Fatal(err)
	}
	set := append(categories["full_gate_set"].([]any), "gate.t3_to_t4")
	sort.Slice(set, func(i, j int) bool { return set[i].(string) < set[j].(string) })
	categories["full_gate_set"] = set
	artifacts["categories"], _ = json.Marshal(categories)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(hash, artifacts)
	if err != nil || loaded.Doctrines == nil {
		t.Fatalf("doctrine bundle=%+v err=%v", loaded, err)
	}
	if _, ok := loaded.ResolvePrestige(hash); !ok {
		t.Fatal("loaded doctrine bundle did not satisfy catalog validity")
	}
	delete(artifacts, "meters")
	delete(artifacts, "achievements")
	orphanHash, _ := save.ConstantsHashArtifacts(artifacts)
	if _, err := Load(orphanHash, artifacts); err == nil {
		t.Fatal("doctrine artifact activated without paired foundations")
	}
}

func TestLoadActivatesFiscalOnlyAfterPets(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "replay", "apply-logged-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		PetArtifacts map[string]string `json:"pet_founder_artifacts"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(fixture.PetArtifacts)+1)
	for name, value := range fixture.PetArtifacts {
		artifacts[name] = []byte(value)
	}
	var economyRoot map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economyRoot); err != nil {
		t.Fatal(err)
	}
	sources := economyRoot["multiplier_sources"].([]any)
	sources = append(sources,
		map[string]any{"id": "fiscal.generator.beige_tower", "slot": "prestige", "target": "generator.beige_tower", "provider": "fiscal"},
		map[string]any{"id": "fiscal.hoard", "slot": "prestige", "target": "all", "provider": "fiscal"},
	)
	economyRoot["multiplier_sources"] = sources
	artifacts["economy"], _ = json.Marshal(economyRoot)
	fiscalFixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "fiscal-foundation-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fiscalEnvelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(fiscalFixture, &fiscalEnvelope); err != nil {
		t.Fatal(err)
	}
	artifacts["fiscal"] = fiscalEnvelope.Baseline
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(hash, artifacts)
	if err != nil || loaded.Fiscal == nil {
		t.Fatalf("fiscal bundle=%+v err=%v", loaded, err)
	}
	delete(artifacts, "pets")
	orphanHash, _ := save.ConstantsHashArtifacts(artifacts)
	if _, err := Load(orphanHash, artifacts); err == nil {
		t.Fatal("fiscal artifact activated without pets")
	}
}

func TestLoadActivatesPitchOnlyOnCompleteSoulChain(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "replay", "apply-logged-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		PetArtifacts map[string]string `json:"pet_founder_artifacts"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(fixture.PetArtifacts)+4)
	for name, value := range fixture.PetArtifacts {
		artifacts[name] = []byte(value)
	}
	var economyRoot map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economyRoot); err != nil {
		t.Fatal(err)
	}
	economyRoot["multiplier_sources"] = append(economyRoot["multiplier_sources"].([]any),
		map[string]any{"id": "fiscal.generator.beige_tower", "slot": "prestige", "target": "generator.beige_tower", "provider": "fiscal"},
		map[string]any{"id": "fiscal.hoard", "slot": "prestige", "target": "all", "provider": "fiscal"})
	artifacts["economy"], _ = json.Marshal(economyRoot)
	fiscalFixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "fiscal-foundation-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fiscalEnvelope struct {
		Baseline map[string]any `json:"baseline"`
	}
	if err := json.Unmarshal(fiscalFixture, &fiscalEnvelope); err != nil {
		t.Fatal(err)
	}
	rows := fiscalEnvelope.Baseline["unlock_rows"].([]any)
	fiscalEnvelope.Baseline["unlock_rows"] = append([]any{map[string]any{"unlock_id": "minigame.pitch", "cost": float64(3)}}, rows...)
	artifacts["fiscal"], _ = json.Marshal(fiscalEnvelope.Baseline)
	artifacts["minigames"], err = os.ReadFile(filepath.Join("..", "..", "testdata", "minigame", "pitch-v3.json"))
	if err != nil {
		t.Fatal(err)
	}
	var petRoot map[string]any
	if err := json.Unmarshal(artifacts["pets"], &petRoot); err != nil {
		t.Fatal(err)
	}
	petRoot["schema_version"] = float64(2)
	for _, row := range petRoot["actions"].([]any) {
		row.(map[string]any)["soul_gate"] = "ordinary"
	}
	artifacts["pets"], _ = json.Marshal(petRoot)
	artifacts["soul"] = []byte(`{"schema_version":1,"policy":{"soul_floor":0,"soul_initial":100,"soul_max":100,"recovery_beat_ceiling_ms":5000,"max_session_wall_ms":86400000},"bands":[{"band_member":"near_zero","min_inclusive":0,"max_inclusive":9,"human_content_locked":true,"reason_key":"category.low_percent"},{"band_member":"hollow","min_inclusive":10,"max_inclusive":39,"human_content_locked":false,"reason_key":"category.ethical_percent"},{"band_member":"dimming","min_inclusive":40,"max_inclusive":74,"human_content_locked":false,"reason_key":"category.hundred_percent"},{"band_member":"whole","min_inclusive":75,"max_inclusive":100,"human_content_locked":false,"reason_key":"category.any_percent"}],"debit_sources":[],"recovery_activities":[],"ending_policy":{"whole_variant":"earnest_ascension","depleted_variant":"training_data"}}`)
	artifacts["pitch"], err = os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "pitch-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Load(hash, artifacts)
	if err != nil || bundle.Pitch == nil {
		t.Fatalf("Pitch bundle=%+v err=%v", bundle, err)
	}
	content, ok := production.ReplayCatalogSet{hash: bundle}.ResolveTenantContent(hash, "pitch", "1.0.0")
	if !ok || !bytes.Equal(content.Bytes, artifacts["pitch"]) || content.SchemaVersion != 1 {
		t.Fatalf("content=%+v ok=%v", content, ok)
	}
	delete(artifacts, "soul")
	orphanHash, _ := save.ConstantsHashArtifacts(artifacts)
	if _, err := Load(orphanHash, artifacts); err == nil {
		t.Fatal("Pitch artifact activated without Soul")
	}
}
