package publicread

import (
	"strings"
	"testing"

	"cloud-clicker/server/publicapi"
)

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
