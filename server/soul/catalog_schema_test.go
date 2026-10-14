package soul

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/publicapi"
)

func soulSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(SoulSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestSoulOwnerSchemaAdmitsCurrentStoredAndRecoveryFixtureWithoutChangingBytes(t *testing.T) {
	definitions := soulSchemaDefinitions(t)
	declarations := Declarations{EpochSeeded: true, CatchupCeilingMS: 5000, CopyKeys: map[string]struct{}{}}
	for _, key := range copykeys.All() {
		declarations.CopyKeys[key] = struct{}{}
	}
	for _, path := range []string{
		"../../balance/soul/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/soul.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/soul.json",
		"../../testdata/soul/recovery-activities-v1.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(SoulCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		decl := declarations
		if path == "../../testdata/soul/recovery-activities-v1.json" {
			decl.EpochSeeded = false
		}
		if _, err := LoadCatalog(data, decl); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	if err := publicapi.ValidateJSON(SoulCatalogSchemaName, fixtureCatalogBytes(t), definitions); err != nil {
		t.Fatal(err)
	}
}

func TestSoulOwnerSchemaClosedShapesAndEnums(t *testing.T) {
	type object = map[string]any
	policy := func(v object) object { return v["policy"].(object) }
	band := func(v object) object { return v["bands"].([]any)[0].(object) }
	debit := func(v object) object { return v["debit_sources"].([]any)[0].(object) }
	activity := func(v object) object { return v["recovery_activities"].([]any)[0].(object) }
	ending := func(v object) object { return v["ending_policy"].(object) }
	definitions := soulSchemaDefinitions(t)
	for name, edit := range map[string]func(object){
		"root missing":             func(v object) { delete(v, "policy") },
		"root private":             func(v object) { v["account_id"] = "private" },
		"version":                  func(v object) { v["schema_version"] = 2 },
		"policy null":              func(v object) { policy(v)["soul_initial"] = nil },
		"policy negative":          func(v object) { policy(v)["soul_floor"] = -1 },
		"policy unsafe":            func(v object) { policy(v)["soul_max"] = json.Number("9007199254740992") },
		"policy fractional":        func(v object) { policy(v)["soul_initial"] = .5 },
		"policy missing heartbeat": func(v object) { delete(policy(v), "recovery_beat_ceiling_ms") },
		"policy wall zero":         func(v object) { policy(v)["max_session_wall_ms"] = 0 },
		"band missing":             func(v object) { delete(band(v), "reason_key") },
		"band member":              func(v object) { band(v)["band_member"] = "fictional" },
		"band null lock":           func(v object) { band(v)["human_content_locked"] = nil },
		"band unknown":             func(v object) { band(v)["script"] = "arbitrary" },
		"band ID":                  func(v object) { band(v)["reason_key"] = "bad-id" },
		"debit source":             func(v object) { debit(v)["source_id"] = "bad-id" },
		"debit owner":              func(v object) { debit(v)["owner_kind"] = "fictional" },
		"debit zero":               func(v object) { debit(v)["amount"] = 0 },
		"debit null":               func(v object) { debit(v)["may_exhaust"] = nil },
		"activity missing":         func(v object) { delete(activity(v), "title_copy_key") },
		"activity toy":             func(v object) { activity(v)["toy_kind"] = "fictional" },
		"activity duration":        func(v object) { activity(v)["duration_attended_ms"] = 0 },
		"activity amount":          func(v object) { activity(v)["recovery_amount"] = 0 },
		"activity copy":            func(v object) { activity(v)["disclosure_copy_key"] = "bad-id" },
		"ending whole":             func(v object) { ending(v)["whole_variant"] = "training_data" },
		"ending depleted":          func(v object) { ending(v)["depleted_variant"] = "earnest_ascension" },
	} {
		t.Run(name, func(t *testing.T) {
			data := editSoulFixture(t, edit)
			if err := publicapi.ValidateJSON(SoulCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadCatalog(data, fixtureDeclarations(false)); err == nil {
				t.Fatal("owner admitted invalid shape")
			}
		})
	}
	for _, owner := range []OwnerKind{OwnerEvent, OwnerLongevity, OwnerContract, OwnerFixture} {
		for _, toy := range []ToyKind{ToyDefrag, ToyRepot, ToyServerRoom} {
			data := editSoulFixture(t, func(v object) { debit(v)["owner_kind"] = string(owner); activity(v)["toy_kind"] = string(toy) })
			if err := publicapi.ValidateJSON(SoulCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, fixtureDeclarations(false)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSoulOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	definitions := soulSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"missing band":         func(v map[string]any) { v["bands"] = v["bands"].([]any)[:3] },
		"unbound copy":         func(v map[string]any) { v["bands"].([]any)[0].(map[string]any)["reason_key"] = "soul.absent" },
		"initial relationship": func(v map[string]any) { v["policy"].(map[string]any)["soul_initial"] = 101 },
		"heartbeat binding":    func(v map[string]any) { v["policy"].(map[string]any)["recovery_beat_ceiling_ms"] = 5001 },
		"watchdog": func(v map[string]any) {
			v["recovery_activities"].([]any)[0].(map[string]any)["duration_attended_ms"] = 86400000
		},
		"exhaust relationship": func(v map[string]any) { v["debit_sources"].([]any)[0].(map[string]any)["single_use"] = false },
	} {
		t.Run(name, func(t *testing.T) {
			data := editSoulFixture(t, edit)
			if err := publicapi.ValidateJSON(SoulCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, fixtureDeclarations(false)); err == nil {
				t.Fatal("structural descriptor replaced semantic owner")
			}
		})
	}
	data := editSoulFixture(t, func(v map[string]any) { v["debit_sources"], v["recovery_activities"] = []any{}, []any{} })
	if err := publicapi.ValidateJSON(SoulCatalogSchemaName, data, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, fixtureDeclarations(true)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, fixtureDeclarations(false)); err == nil {
		t.Fatal("empty non-epoch fixture content admitted")
	}
}

func TestSoulOwnerSchemaFreshExportsAndOriginalUniqueness(t *testing.T) {
	before, fresh := SoulSchemas(), SoulSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic export")
	}
	fresh[0].Schema.Fields[0].Schema.Items.Fields[0].Schema.Enum[0] = "changed"
	if !reflect.DeepEqual(before, SoulSchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	definitions := soulSchemaDefinitions(t)
	source := fixtureCatalogBytes(t)
	escaped := bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1)
	if bytes.Equal(source, escaped) {
		t.Fatal("escape control did not change source")
	}
	if err := publicapi.ValidateJSON(SoulCatalogSchemaName, escaped, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(escaped, fixtureDeclarations(false)); err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Replace(source, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1)
	if bytes.Equal(source, duplicate) {
		t.Fatal("duplicate control did not change source")
	}
	// Artifact descriptor map validation is not Soul's original-byte uniqueness
	// gate. Do not alter the shared validator's last-value contract to hide this.
	if err := publicapi.ValidateJSON(SoulCatalogSchemaName, duplicate, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(duplicate, fixtureDeclarations(false)); err == nil {
		t.Fatal("Soul's existing duplicate-member refusal was lost")
	}
}
