package publicread

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"cloud-clicker/server/publicapi"
)

// The public packages may consume narrow reader interfaces but must not import
// private account/session/save repositories (C10). Inspect production imports,
// not test fixtures or comments. Transitive reader dependencies and runtime
// disclosure remain separate acceptance boundaries.
func privateRepositoryImport(source []byte) error {
	file, err := parser.ParseFile(token.NewFileSet(), "public.go", source, parser.ImportsOnly)
	if err != nil {
		return err
	}
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return err
		}
		for _, private := range []string{"cloud-clicker/server/account", "cloud-clicker/server/session", "cloud-clicker/server/save", "cloud-clicker/server/production"} {
			if path == private || strings.HasPrefix(path, private+"/") {
				return fmt.Errorf("public API imports private repository owner %s", path)
			}
		}
	}
	return nil
}

func TestPublicPackagesDoNotImportPrivateRepositories(t *testing.T) {
	for _, directory := range []string{".", "../publicapi"} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		checked := 0
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(directory, entry.Name())
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := privateRepositoryImport(source); err != nil {
				t.Errorf("%s: %v", path, err)
			}
			checked++
		}
		if checked == 0 {
			t.Fatalf("no production source checked in %s", directory)
		}
	}
}

func TestPrivateRepositoryImportGuardDiscriminatesAliasesAndComments(t *testing.T) {
	for _, source := range []string{
		`package publicread; import "cloud-clicker/server/account"`,
		`package publicread; import renamed "cloud-clicker/server/account"`,
		`package publicread; import . "cloud-clicker/server/save"`,
		`package publicread; import _ "cloud-clicker/server/session"`,
		`package publicread; import "cloud-clicker/server/save/nested"`,
		`package publicread; import "cloud-clicker/server/production"`,
	} {
		if err := privateRepositoryImport([]byte(source)); err == nil {
			t.Fatalf("private import admitted: %s", source)
		}
	}
	for _, source := range []string{
		`package publicread // import "cloud-clicker/server/account"`,
		`package publicread; import "cloud-clicker/server/publicapi"`,
		`package publicread; import "cloud-clicker/server/leaderboard"`,
		`package publicread; import "cloud-clicker/server/accounting"`,
	} {
		if err := privateRepositoryImport([]byte(source)); err != nil {
			t.Fatalf("non-private source refused: %s: %v", source, err)
		}
	}
	if err := privateRepositoryImport([]byte(`package publicread; import (`)); err == nil {
		t.Fatal("unparseable import list admitted")
	}
}

// publicIdentityFields are the only founder-identifying fields the public
// surface may carry (AC5: "public board identity"): a verified board row's
// founder and a route's credited first executor (nullable after anonymization).
var publicIdentityFields = map[string]string{
	"PublicBoardItem.founder_id":            "verified board rows are the public record (A3)",
	"PublicRoute.first_executor_founder_id": "Route Registry credit, withheld after anonymization (C12)",
}

// forbiddenFieldFragments are founder-private or account-level identities the
// public surface must never expose, whatever the schema.
var forbiddenFieldFragments = []string{"account", "email", "token", "session", "recovery", "password", "secret", "stream", "save", "ip_", "device", "presence"}

// Independent publication expectations from C12/C14, not generated from the
// registry under test. This is a privacy gate, not another runtime schema.
// A new public DTO or field needs its owning contract and review here too.
var publicSchemaFields = map[string][]string{
	"APIError":             {"category", "detail"},
	"PublicBoardCursorKey": {"exponent", "key", "quantized_mantissa", "run_id"},
	"PublicBoardItem":      {"founder_id", "key", "rank", "run_id", "verified_at", "world_first"},
	"PublicBoardKey":       {"exponent", "kind", "quantized_mantissa", "value"},
	"PublicBoardPage":      {"category_id", "epoch_id", "items", "mandate_level", "next_cursor", "ranking_kind", "variables"},
	"PublicBoardVariables": {"advisor", "commons", "faction", "glitched"},
	"PublicEpoch":          {"accepted_hashes", "changelog_markdown", "changelog_ref", "ended_at", "epoch_id", "name", "started_at"},
	"PublicEpochCursorKey": {"epoch_id"},
	"PublicEpochPage":      {"items", "next_cursor"},
	"PublicRoute":          {"adoption_count", "credited_at", "first_executor_founder_id", "naming_deadline", "naming_status", "public_name", "route_id"},
	"PublicRouteCursorKey": {"route_id"},
	"PublicRoutePage":      {"items", "next_cursor"},
	"PublicRunVerdict":     {"catalog_url", "constants_hash", "engine_version", "genesis_sha256", "genesis_url", "replay_log_sha256", "replay_log_url", "run_id", "verdict"},
}

func walkPublicFields(name string, schema *publicapi.Schema, visit func(path string)) {
	if schema == nil {
		return
	}
	for _, field := range schema.Fields {
		path := name + "." + field.Name
		visit(path)
		walkPublicFields(path, field.Schema, visit)
	}
	walkPublicFields(name, schema.Items, visit)
	for _, alternate := range schema.Alternates {
		walkPublicFields(name, alternate, visit)
	}
}

func TestPublicRegistryIsStructurallyPrivate(t *testing.T) {
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	if err := publicRegistryPrivacyError(registry); err != nil {
		t.Fatal(err)
	}
}

func publicRegistryPrivacyError(registry *publicapi.Registry) error {
	for _, operation := range registry.Operations() {
		if !operation.Public || operation.Auth != publicapi.AuthNone || operation.Method != "GET" || !strings.HasPrefix(operation.Path, "/api/public/v1/") || operation.Request != "" {
			return fmt.Errorf("public registry operation %s is not a closed unauthenticated GET read", operation.ID)
		}
	}
	seen := map[string]bool{}
	for _, named := range registry.Schemas() {
		fields, allowed := publicSchemaFields[named.Name]
		if !allowed {
			return fmt.Errorf("public schema %s has no specified publication field set", named.Name)
		}
		allowedPaths := map[string]bool{}
		for _, name := range fields {
			allowedPaths[named.Name+"."+name] = true
		}
		var fieldError error
		walkPublicFields(named.Name, named.Schema, func(path string) {
			if fieldError != nil {
				return
			}
			if !allowedPaths[path] {
				fieldError = fmt.Errorf("public schema field %s is outside its specified publication field set", path)
				return
			}
			field := path[strings.LastIndex(path, ".")+1:]
			for _, fragment := range forbiddenFieldFragments {
				if strings.Contains(field, fragment) {
					fieldError = fmt.Errorf("public schema field %s exposes a private identity (%q)", path, fragment)
					return
				}
			}
			if strings.Contains(field, "founder") {
				if _, ok := publicIdentityFields[path]; !ok {
					fieldError = fmt.Errorf("public schema field %s carries a founder identity outside the ruled public board identity", path)
					return
				}
				seen[path] = true
			}
		})
		if fieldError != nil {
			return fieldError
		}
	}
	// The allow-list must not rot: every allowed identity is actually present.
	for path := range publicIdentityFields {
		if !seen[path] {
			return fmt.Errorf("allow-listed identity %s no longer exists; remove it", path)
		}
	}
	return nil
}

func TestPublicPrivacyGuardRejectsUndeclaredStateFields(t *testing.T) {
	registry, err := Registry()
	if err != nil {
		t.Fatal(err)
	}
	if err := publicRegistryPrivacyError(registry); err != nil {
		t.Fatalf("healthy registry refused: %v", err)
	}
	pin, err := publicapi.CanonicalOperationPins(registry)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name       string
		schema     string
		compatible bool
		mutate     func(*publicapi.Schema)
	}{
		{"plain private balance", "PublicRoute", true, func(schema *publicapi.Schema) {
			schema.Fields = append(schema.Fields, publicapi.Field{Name: "cash", Schema: stringSchema("canonical-decimal")})
		}},
		{"nested reuse of an allowed name", "PublicRoute", false, func(schema *publicapi.Schema) {
			for index := range schema.Fields {
				if schema.Fields[index].Name == "public_name" {
					schema.Fields[index].Schema = &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{field("route_id", stringSchema(""))}}
				}
			}
		}},
		{"extra error state", "APIError", true, func(schema *publicapi.Schema) {
			schema.Fields = append(schema.Fields, publicapi.Field{Name: "cash", Schema: stringSchema("canonical-decimal")})
		}},
		{"forbidden identity", "PublicRoute", true, func(schema *publicapi.Schema) {
			schema.Fields = append(schema.Fields, publicapi.Field{Name: "account_id", Schema: stringSchema("uuid")})
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			schemas := registry.Schemas()
			found := false
			for _, named := range schemas {
				if named.Name == row.schema {
					row.mutate(named.Schema)
					sort.Slice(named.Schema.Fields, func(i, j int) bool { return named.Schema.Fields[i].Name < named.Schema.Fields[j].Name })
					found = true
				}
			}
			if !found {
				t.Fatalf("missing mutation subject %s", row.schema)
			}
			candidate, err := publicapi.NewRegistry(schemas, registry.Operations())
			if err != nil {
				t.Fatalf("negative fixture is not a valid registry: %v", err)
			}
			if err := publicapi.CheckCompatibilityPin(pin, candidate); (err == nil) != row.compatible {
				t.Fatalf("fixture compatibility differs from the expected boundary: %v", err)
			}
			if err := publicRegistryPrivacyError(candidate); err == nil {
				t.Fatal("privacy guard admitted undeclared state")
			}
		})
	}
	t.Run("unknown publication schema", func(t *testing.T) {
		schemas := append(registry.Schemas(), publicapi.NamedSchema{Name: "UnspecifiedPublicState", Schema: stringSchema("canonical-decimal")})
		candidate, err := publicapi.NewRegistry(schemas, registry.Operations())
		if err != nil {
			t.Fatalf("negative fixture is not a valid registry: %v", err)
		}
		if err := publicRegistryPrivacyError(candidate); err == nil {
			t.Fatal("privacy guard admitted a schema without a specified publication field set")
		}
	})
	if err := publicRegistryPrivacyError(registry); err != nil {
		t.Fatalf("negative fixtures mutated the healthy registry: %v", err)
	}
}
