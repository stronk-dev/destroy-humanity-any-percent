package curriculum

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/routes"
)

func curriculumSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(CurriculumSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestCurriculumOwnerSchemaAdmitsCurrentStoredAndCandidateWithoutChangingBytes(t *testing.T) {
	definitions := curriculumSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/curriculum/t0-t1.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/curriculum.json",
		"../../balance/testdata/t0-t1/curriculum-v2.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		declarations := curriculumDeclarations(t)
		routePath := "../../balance/routes/phase0.json"
		if filepath.Base(filepath.Dir(path)) == "artifacts" {
			economyBytes, err := os.ReadFile(filepath.Join(filepath.Dir(path), "economy.json"))
			if err != nil {
				t.Fatal(err)
			}
			declarations.Economy, err = economy.LoadCatalog(economyBytes)
			if err != nil {
				t.Fatal(err)
			}
			routePath = filepath.Join(filepath.Dir(path), "routes.json")
		}
		routeBytes, err := os.ReadFile(routePath)
		if err != nil {
			t.Fatal(err)
		}
		routeCatalog, err := routes.LoadCatalog(routeBytes)
		if err != nil {
			t.Fatal(err)
		}
		declarations.GateIDs = map[string]struct{}{}
		for _, gate := range routeCatalog.Gates() {
			declarations.GateIDs[gate.ID] = struct{}{}
		}
		if _, err := Load(data, declarations); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed original artifact bytes")
		}
	}
}

func TestCurriculumOwnerSchemaClosedShapesAndBranchPairing(t *testing.T) {
	type object = map[string]any
	failure := func(v object) object { return v["first_failure"].(object) }
	branch := func(v object, index int) object { return failure(v)["branches"].([]any)[index].(object) }
	starter := func(v object, index int) object { return branch(v, index)["starter_package"].(object) }
	definitions := curriculumSchemaDefinitions(t)
	declarations := curriculumDeclarations(t)
	for name, edit := range map[string]func(object){
		"root private":         func(v object) { v["account_id"] = "private" },
		"version":              func(v object) { v["schema_version"] = 2 },
		"failure missing zero": func(v object) { delete(failure(v), "founder_exit_count") },
		"failure null zero":    func(v object) { failure(v)["founder_exit_count"] = nil },
		"failure count":        func(v object) { failure(v)["founder_exit_count"] = 1 },
		"failure time":         func(v object) { failure(v)["attended_ms"] = 899999 },
		"failure gate":         func(v object) { failure(v)["gate_id"] = "gate.t1_to_t2" },
		"failure effect":       func(v object) { failure(v)["requested_command_effect"] = "client_decides" },
		"null branches":        func(v object) { failure(v)["branches"] = nil },
		"branch unknown":       func(v object) { branch(v, 0)["branch"] = "fictional" },
		"branch wrong field":   func(v object) { branch(v, 1)["minimum_owned_upgrades"] = 0 },
		"branch missing bonus": func(v object) { delete(branch(v, 0), "route_knowledge_bonus") },
		"branch unsafe bonus":  func(v object) { branch(v, 0)["route_knowledge_bonus"] = json.Number("9007199254740992") },
		"acquihire zero":       func(v object) { branch(v, 0)["minimum_purchased_generators"] = 0 },
		"acquihire amount":     func(v object) { starter(v, 0)["amount"] = "10000" },
		"burnout factor":       func(v object) { branch(v, 1)["cheapest_price_factor"] = 2 },
		"burnout count":        func(v object) { starter(v, 1)["count"] = 0 },
		"pivot ID":             func(v object) { starter(v, 2)["upgrade_id"] = "bad-id" },
	} {
		t.Run(name, func(t *testing.T) {
			data := editCurriculumFixture(t, edit)
			if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := Load(data, declarations); err == nil {
				t.Fatal("owner admitted invalid shape")
			}
		})
	}
	for index := range 3 {
		data := editCurriculumFixture(t, func(v object) { branch(v, index)["starter_package"] = starter(v, (index+1)%3) })
		if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, definitions); err == nil {
			t.Fatal("descriptor admitted a valid starter on the wrong branch")
		}
	}
}

func TestCurriculumOwnerSchemaLeavesPayoffLiteralsAsData(t *testing.T) {
	data := editCurriculumFixture(t, func(v map[string]any) {
		rows := v["first_failure"].(map[string]any)["branches"].([]any)
		acquihire := rows[0].(map[string]any)
		acquihire["minimum_purchased_generators"], acquihire["minimum_owned_upgrades"] = 201, 2
		acquihire["route_knowledge_bonus"] = 1
		acquihire["starter_package"].(map[string]any)["amount"] = "2e4"
		burnout := rows[1].(map[string]any)
		burnout["cheapest_price_factor"] = "3e0"
		burnout["route_knowledge_bonus"] = 51
		burnout["starter_package"].(map[string]any)["count"] = 11
	})
	if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, curriculumSchemaDefinitions(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(data, curriculumDeclarations(t)); err != nil {
		t.Fatal(err)
	}
}

func TestCurriculumOwnerSchemaDoesNotReplaceSemanticLoaderOrNarrowGenericStarters(t *testing.T) {
	definitions := curriculumSchemaDefinitions(t)
	declarations := curriculumDeclarations(t)
	for name, edit := range map[string]func(map[string]any){
		"copy binding":   func(v map[string]any) { v["first_failure"].(map[string]any)["next_run_key"] = "curriculum.absent" },
		"empty branches": func(v map[string]any) { v["first_failure"].(map[string]any)["branches"] = []any{} },
		"branch ordering": func(v map[string]any) {
			rows := v["first_failure"].(map[string]any)["branches"].([]any)
			rows[0], rows[1] = rows[1], rows[0]
		},
		"starter amount": func(v map[string]any) {
			v["first_failure"].(map[string]any)["branches"].([]any)[0].(map[string]any)["starter_package"].(map[string]any)["amount"] = "0"
		},
		"upgrade binding": func(v map[string]any) {
			v["first_failure"].(map[string]any)["branches"].([]any)[2].(map[string]any)["starter_package"].(map[string]any)["upgrade_id"] = "upgrade.absent"
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editCurriculumFixture(t, edit)
			if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(data, declarations); err == nil {
				t.Fatal("structural descriptor replaced semantic owner")
			}
		})
	}
	catalog, err := Load(curriculumFixture(t), declarations)
	if err != nil {
		t.Fatal(err)
	}
	for _, branch := range catalog.FirstFailure.Branches {
		if err := ValidateStarter(branch.StarterPackage, declarations.Economy); err != nil {
			t.Fatalf("generic starter %s was narrowed: %v", branch.StarterPackage.Kind, err)
		}
	}
}

func TestCurriculumOwnerSchemaFreshExportsAndOriginalMemberControls(t *testing.T) {
	before, fresh := CurriculumSchemas(), CurriculumSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic export")
	}
	*fresh[0].Schema.Fields[0].Schema.Fields[0].Schema.Minimum = 1
	if !reflect.DeepEqual(before, CurriculumSchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	definitions := curriculumSchemaDefinitions(t)
	source := bytes.TrimSpace(curriculumFixture(t))
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("member control did not change source")
		}
		if err := publicapi.ValidateJSON(CurriculumCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(data, curriculumDeclarations(t)); err != nil {
			t.Fatal(err)
		}
	}
}
