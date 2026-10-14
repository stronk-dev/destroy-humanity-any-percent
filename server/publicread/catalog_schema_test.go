package publicread

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	"cloud-clicker/server/prestige"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/soul"
)

func currentCatalogSchemaOwners() []CatalogSchemaOwner {
	return []CatalogSchemaOwner{
		{"categories", leaderboard.CategoryCatalogSchemaName, leaderboard.CategorySchemas()},
		{"achievements", achievements.AchievementCatalogSchemaName, achievements.AchievementSchemas()},
		{"meters", meters.MeterCatalogSchemaName, meters.MeterSchemas()},
		{"opportunities", activeplay.ActivePlayCatalogSchemaName, activeplay.ActivePlaySchemas()},
		{"doctrines", doctrine.DoctrineCatalogSchemaName, doctrine.DoctrineSchemas()},
		{"fiscal", fiscal.FiscalCatalogSchemaName, fiscal.FiscalSchemas()},
		{"soul", soul.SoulCatalogSchemaName, soul.SoulSchemas()},
		{"curriculum", curriculum.CurriculumCatalogSchemaName, curriculum.CurriculumSchemas()},
		{"prestige", prestige.PrestigePolicySchemaName, prestige.PrestigeSchemas()},
		{"pitch", pitch.PitchCatalogSchemaName, pitch.PitchSchemas()},
		{"pets", pet.PetCatalogSchemaName, pet.PetSchemas()},
		{"factions", faction.FactionCatalogSchemaName, faction.FactionSchemas()},
		{"guilds", guild.GuildCatalogSchemaName, guild.GuildSchemas()},
		{"commons", commons.CommonsCatalogSchemaName, commons.CommonsSchemas()},
		{"minigames", minigame.MinigameCatalogSchemaName, minigame.MinigameSchemas()},
	}
}

// C18 construction diagnostic, not an implemented catalog transport. Exact
// descriptor validation does not prove the claimed original-byte digest survives
// embedding. Use a real formatted artifact, not a conveniently compact fixture.
func TestCatalogSchemaCompositionDoesNotProveArtifactByteIdentity(t *testing.T) {
	data, err := os.ReadFile("../../balance/epochs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := epochseed.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	path, found := epochseed.ArtifactPath(seed, "categories")
	if !found {
		t.Fatal("missing actual category source")
	}
	original, err := os.ReadFile(filepath.Join("../..", path))
	if err != nil {
		t.Fatal(err)
	}
	owners := []CatalogSchemaOwner{}
	for _, owner := range currentCatalogSchemaOwners() {
		if owner.Name == "categories" || owner.Name == "meters" {
			owners = append(owners, owner)
		}
	}
	rows, err := ComposeCatalogSchemas([]string{"categories", "meters"}, owners)
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := publicapi.ValidateSchemaDefinitions(rows)
	if err != nil {
		t.Fatal(err)
	}
	originalHash := sha256.Sum256(original)
	body, err := json.Marshal(map[string]any{"name": "categories", "sha256": fmt.Sprintf("sha256:%x", originalHash), "json": json.RawMessage(original)})
	if err != nil {
		t.Fatal(err)
	}
	if err := publicapi.ValidateJSON(CatalogArtifactSchemaName, body, definitions); err != nil {
		t.Fatal("real source envelope does not validate", err)
	}
	var received map[string]json.RawMessage
	if err := json.Unmarshal(body, &received); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(received["json"], original) || sha256.Sum256(received["json"]) == originalHash {
		t.Fatal("diagnostic failed to expose RawMessage compaction of the formatted source")
	}
	// Do not rewrite the source/digest or claim schema-valid reserialization is
	// third-party verification. A byte-preserving writer/consumer is still needed.
}

func TestCatalogSchemaCompositionRefusesActualIncompleteEpochSet(t *testing.T) {
	data, err := os.ReadFile("../../balance/epochs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := epochseed.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	families := make([]string, 0, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		families = append(families, artifact.Name)
	}
	owners := currentCatalogSchemaOwners()
	if len(families) != 19 || len(owners) != 15 {
		t.Fatal("update actual served/provider population with its accepted changes")
	}
	result, err := ComposeCatalogSchemas(families, owners)
	var coverage *CatalogSchemaCoverageError
	if result != nil || !errors.Is(err, ErrCatalogSchemaComposition) || !errors.As(err, &coverage) {
		t.Fatalf("partial schemas escaped: %v / %v", result, err)
	}
	if !reflect.DeepEqual(coverage.Missing, []string{"economy", "minigame_api", "relevance", "routes"}) {
		t.Fatal(coverage.Missing)
	}
	// Verify the existing exports coexist and admit their actual source bytes.
	// This is not a fabricated four-owner completion or a public response union.
	definitions := []publicapi.NamedSchema{}
	for _, owner := range owners {
		definitions = append(definitions, owner.Definitions...)
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	validated, err := publicapi.ValidateSchemaDefinitions(definitions)
	if err != nil {
		t.Fatal("real owner namespaces do not compose", err)
	}
	for _, owner := range owners {
		path, found := epochseed.ArtifactPath(seed, owner.Name)
		if !found {
			t.Fatal("provider outside real manifest", owner.Name)
		}
		data, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(owner.Root, data, validated); err != nil {
			t.Fatal(owner.Name, err)
		}
		if !bytes.Equal(data, before) {
			t.Fatal("composition rewrote original artifact")
		}
	}
}

// A small explicit unit context, never the production nineteen-family set.
// Meters declares SharedFlag once; the other root references that definition.
func fixtureCatalogSchemaOwners() []CatalogSchemaOwner {
	return []CatalogSchemaOwner{
		{Name: "categories", Root: "FixtureCategories", Definitions: []publicapi.NamedSchema{{Name: "FixtureCategories", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("flag", &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "SharedFlag"}),
		}}}}},
		{Name: "meters", Root: "FixtureMeters", Definitions: []publicapi.NamedSchema{
			{Name: "FixtureMeters", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{field("quantity", integer(0, 10))}}},
			{Name: "SharedFlag", Schema: &publicapi.Schema{Kind: publicapi.SchemaBoolean}},
		}},
	}
}

func fixtureCatalogBundle() map[string]any {
	return map[string]any{"constants_hash": "sha256:" + strings.Repeat("a", 64), "artifacts": []any{
		map[string]any{"name": "categories", "sha256": "sha256:" + strings.Repeat("b", 64), "json": map[string]any{"flag": false}},
		map[string]any{"name": "meters", "sha256": "sha256:" + strings.Repeat("c", 64), "json": map[string]any{"quantity": 0}},
	}}
}

func TestCatalogSchemaCompositionLiteralBindingsAndClosedPayloads(t *testing.T) {
	rows, err := ComposeCatalogSchemas([]string{"meters", "categories"}, fixtureCatalogSchemaOwners())
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := publicapi.ValidateSchemaDefinitions(rows)
	if err != nil {
		t.Fatal(err)
	}
	validate := func(body map[string]any) error {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		return publicapi.ValidateJSON(CatalogBundleSchemaName, data, definitions)
	}
	if err := validate(fixtureCatalogBundle()); err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]func(map[string]any){
		"unknown family": func(v map[string]any) { v["artifacts"].([]any)[0].(map[string]any)["name"] = "unserved" },
		"wrong family payload": func(v map[string]any) {
			v["artifacts"].([]any)[0].(map[string]any)["json"] = map[string]any{"quantity": 0}
		},
		"private envelope": func(v map[string]any) { v["account_id"] = "private" },
		"private artifact": func(v map[string]any) { v["artifacts"].([]any)[0].(map[string]any)["private"] = true },
		"private payload": func(v map[string]any) {
			v["artifacts"].([]any)[0].(map[string]any)["json"].(map[string]any)["private"] = true
		},
		"null payload":           func(v map[string]any) { v["artifacts"].([]any)[0].(map[string]any)["json"] = nil },
		"string payload":         func(v map[string]any) { v["artifacts"].([]any)[0].(map[string]any)["json"] = "{}" },
		"missing digest":         func(v map[string]any) { delete(v["artifacts"].([]any)[0].(map[string]any), "sha256") },
		"invalid digest":         func(v map[string]any) { v["artifacts"].([]any)[0].(map[string]any)["sha256"] = "a" },
		"invalid constants hash": func(v map[string]any) { v["constants_hash"] = "a" },
		"null array":             func(v map[string]any) { v["artifacts"] = nil },
		"bound": func(v map[string]any) {
			v["artifacts"].([]any)[1].(map[string]any)["json"].(map[string]any)["quantity"] = 11
		},
	} {
		t.Run(name, func(t *testing.T) {
			body := fixtureCatalogBundle()
			edit(body)
			if err := validate(body); err == nil {
				t.Fatal("closed union admitted malformed/misbound payload")
			}
		})
	}
	// Historic bundles are allowed to contain a subset, not all current arms.
	// Ordering, exact membership and hash integrity are still reader semantics.
	historical := fixtureCatalogBundle()
	historical["artifacts"] = historical["artifacts"].([]any)[:1]
	if err := validate(historical); err != nil {
		t.Fatal("full union required every arm in a historical bundle", err)
	}
	artifacts := definitions[CatalogArtifactSchemaName].Alternates
	for index, want := range []string{"categories", "meters"} {
		if got := artifacts[index].Fields[1].Schema.Enum; !reflect.DeepEqual(got, []string{want}) {
			t.Fatal("arms not name ordered", got)
		}
	}
}

func TestCatalogSchemaCompositionRejectsAmbiguousOrInvalidOwnership(t *testing.T) {
	for name, edit := range map[string]func(*[]string, *[]CatalogSchemaOwner){
		"empty set":        func(names *[]string, _ *[]CatalogSchemaOwner) { *names = nil },
		"invalid family":   func(names *[]string, _ *[]CatalogSchemaOwner) { (*names)[0] = "UPPER" },
		"duplicate family": func(names *[]string, _ *[]CatalogSchemaOwner) { (*names)[0] = (*names)[1] },
		"missing owner":    func(_ *[]string, owners *[]CatalogSchemaOwner) { *owners = (*owners)[:1] },
		"extra owner": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			*owners = append(*owners, CatalogSchemaOwner{Name: "unserved"})
		},
		"duplicate owner": func(_ *[]string, owners *[]CatalogSchemaOwner) { *owners = append(*owners, (*owners)[0]) },
		"foreign root":    func(_ *[]string, owners *[]CatalogSchemaOwner) { (*owners)[0].Root = (*owners)[1].Root },
		"nil root":        func(_ *[]string, owners *[]CatalogSchemaOwner) { (*owners)[0].Definitions[0].Schema = nil },
		"reserved root": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[0].Definitions[0].Name = CatalogBundleSchemaName
		},
		"identical shared redeclaration": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[0].Definitions = append((*owners)[0].Definitions, (*owners)[1].Definitions[1])
		},
		"conflicting redeclaration": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[0].Definitions = append((*owners)[0].Definitions, publicapi.NamedSchema{Name: "SharedFlag", Schema: stringSchema("")})
		},
		"unresolved reference": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[0].Definitions[0].Schema.Fields[0].Schema.Ref = "Absent"
		},
		"reference cycle": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[1].Definitions[1].Schema = &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "FixtureCategories"}
		},
		"open JSON escape": func(_ *[]string, owners *[]CatalogSchemaOwner) {
			(*owners)[0].Definitions[0].Schema = &publicapi.Schema{Kind: "json"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			names, owners := []string{"categories", "meters"}, fixtureCatalogSchemaOwners()
			edit(&names, &owners)
			rows, err := ComposeCatalogSchemas(names, owners)
			if rows != nil || !errors.Is(err, ErrCatalogSchemaComposition) {
				t.Fatal("invalid composition returned schemas", rows, err)
			}
		})
	}
}

func TestCatalogSchemaCompositionDoesNotAliasOrDependOnInputOrder(t *testing.T) {
	names, owners := []string{"meters", "categories"}, fixtureCatalogSchemaOwners()
	beforeNames := append([]string(nil), names...)
	beforeOwners, err := json.Marshal(owners)
	if err != nil {
		t.Fatal(err)
	}
	first, err := ComposeCatalogSchemas(names, owners)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range first {
		if row.Name == "FixtureMeters" {
			*row.Schema.Fields[0].Schema.Maximum = 99
		}
		if row.Name == CatalogArtifactSchemaName {
			row.Schema.Alternates[0].Fields[1].Schema.Enum[0] = "changed"
		}
	}
	afterOwners, err := json.Marshal(owners)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeOwners, afterOwners) || !reflect.DeepEqual(names, beforeNames) {
		t.Fatal("constructor/output mutated inputs")
	}
	second, err := ComposeCatalogSchemas(names, owners)
	if err != nil {
		t.Fatal(err)
	}
	owners[0], owners[1] = owners[1], owners[0]
	owners[0].Definitions[0], owners[0].Definitions[1] = owners[0].Definitions[1], owners[0].Definitions[0]
	third, err := ComposeCatalogSchemas([]string{"categories", "meters"}, owners)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(second, third) {
		t.Fatal("composition depends on caller order")
	}
	if reflect.DeepEqual(first, second) {
		t.Fatal("output mutation did not exercise nested ownership")
	}
}

func TestCatalogSchemaCompositionUsesExistingGenerators(t *testing.T) {
	rows, err := ComposeCatalogSchemas([]string{"categories", "meters"}, fixtureCatalogSchemaOwners())
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, publicapi.APIErrorSchema())
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	// Fixture registry only. No operation is added to the real public router.
	registry, err := publicapi.NewRegistry(rows, []publicapi.Operation{{
		ID: "catalog_schema_fixture", Method: "GET", Path: "/api/public/v1/catalog-schema-fixture",
		Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
		Responses: []publicapi.Response{{Kind: publicapi.ResponseSchema, Status: 200, ContentType: publicapi.ContentJSON, SchemaRef: CatalogBundleSchemaName}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	document, err := publicapi.GenerateOpenAPI(registry, "Catalog schema fixture")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(document, &parsed); err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		OneOf []struct {
			AdditionalProperties *bool    `json:"additionalProperties"`
			Required             []string `json:"required"`
			Properties           map[string]struct {
				Ref  string   `json:"$ref"`
				Enum []string `json:"enum"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	if err := json.Unmarshal(parsed.Components.Schemas[CatalogArtifactSchemaName], &artifact); err != nil {
		t.Fatal(err)
	}
	if len(artifact.OneOf) != 2 {
		t.Fatal("generator dropped an arm")
	}
	for index, row := range []struct{ name, root string }{{"categories", "FixtureCategories"}, {"meters", "FixtureMeters"}} {
		arm := artifact.OneOf[index]
		if arm.AdditionalProperties == nil || *arm.AdditionalProperties || !reflect.DeepEqual(arm.Required, []string{"json", "name", "sha256"}) ||
			!reflect.DeepEqual(arm.Properties["name"].Enum, []string{row.name}) || arm.Properties["json"].Ref != "#/components/schemas/"+row.root {
			t.Fatal("generation lost exact family-to-owner binding", arm)
		}
	}
	client, err := publicapi.GenerateTypeScript(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"export type PublicCatalogArtifact =", "json: FixtureCategories;", "json: FixtureMeters;", `name: "categories";`, `name: "meters";`} {
		if !bytes.Contains(client, []byte(fragment)) {
			t.Fatal("typed generation omitted", fragment)
		}
	}
	// The ordinary registry validator, not a second validation implementation,
	// consumes the same constructed descriptor as OpenAPI and TypeScript.
	body, err := json.Marshal(fixtureCatalogBundle())
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.ValidateResponse("catalog_schema_fixture", 200, body); err != nil {
		t.Fatal(err)
	}
}
