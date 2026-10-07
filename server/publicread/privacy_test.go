package publicread

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

func walkPublicFields(name string, schema *publicapi.Schema, visit func(path string)) {
	if schema == nil {
		return
	}
	for _, field := range schema.Fields {
		visit(name + "." + field.Name)
		walkPublicFields(name, field.Schema, visit)
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
	for _, operation := range registry.Operations() {
		if !operation.Public || operation.Auth != publicapi.AuthNone || operation.Method != "GET" || !strings.HasPrefix(operation.Path, "/api/public/v1/") || operation.Request != "" {
			t.Fatalf("public registry operation %s is not a closed unauthenticated GET read", operation.ID)
		}
	}
	seen := map[string]bool{}
	for _, named := range registry.Schemas() {
		if named.Name == "APIError" {
			continue
		}
		walkPublicFields(named.Name, named.Schema, func(path string) {
			field := path[strings.LastIndex(path, ".")+1:]
			for _, fragment := range forbiddenFieldFragments {
				if strings.Contains(field, fragment) {
					t.Fatalf("public schema field %s exposes a private identity (%q)", path, fragment)
				}
			}
			if strings.Contains(field, "founder") {
				if _, ok := publicIdentityFields[path]; !ok {
					t.Fatalf("public schema field %s carries a founder identity outside the ruled public board identity", path)
				}
				seen[path] = true
			}
		})
	}
	// The allow-list must not rot: every allowed identity is actually present.
	for path := range publicIdentityFields {
		if !seen[path] {
			t.Fatalf("allow-listed identity %s no longer exists; remove it", path)
		}
	}
}
