package production

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
)

// Type-check each isolated source before observing the guard. These are source
// ownership regressions, not execution of a live Clout mint.
func TestGaiaWriterMutationForms(t *testing.T) {
	for _, row := range []struct {
		name, body string
		want       int
	}{
		{"assignment", "founder.CloutLifetime = 2", 1},
		{"increment", "founder.CloutLifetime++", 1},
		{"literal", "_ = struct{ CloutLifetime int64 }{CloutLifetime: 2}", 1},
		{"parenthesized-assignment", "(founder.CloutLifetime) += 2", 1},
		{"parenthesized-increment", "((founder.CloutLifetime))++", 1},
		{"address-alias", "p := &founder.CloutLifetime; *p += 2", 1},
		{"parenthesized-address-alias", "p := &(founder.CloutLifetime); *p = 2", 1},
		{"mutator-argument", "mutate(&founder.CloutLifetime)", 1},
		{"inline-indirect-write", "(*(&founder.CloutLifetime))++", 1},
		{"read-only", "println(founder.CloutLifetime)", 0},
		{"value-copy", "value := founder.CloutLifetime; println(value)", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			const name = "production/seeded_hook.go"
			source := []byte("package production\nfunc mutate(value *int64) { *value += 2 }\nfunc seeded(founder *struct{ CloutLifetime int64 }) { " + row.body + " }\n")
			set := token.NewFileSet()
			file, err := parser.ParseFile(set, name, source, 0)
			if err != nil {
				t.Fatal(err)
			}
			config := types.Config{}
			if _, err := config.Check("production", set, []*ast.File{file}, nil); err != nil {
				t.Fatal(err)
			}
			if routes := cloutWriters(t, map[string][]byte{name: source}); len(routes) != row.want {
				t.Fatalf("mutation routes = %v, want %d for %s", routes, row.want, row.body)
			}
		})
	}
}

func TestGaiaCodecRoundTripExceptions(t *testing.T) {
	const healthy = `package save
type State struct { CloutLifetime int64 }
func EncodeStateVersion(normalized State) State { return State{CloutLifetime: normalized.CloutLifetime} }
func RestoreState(source State) State { return State{CloutLifetime: source.CloutLifetime} }
`
	for _, row := range []struct {
		name, filename, from, to, extra string
		refuse                          bool
	}{
		{name: "two-round-trip-copies", filename: "save/state.go"},
		{name: "encode-value-changed", filename: "save/state.go", from: "CloutLifetime: normalized.CloutLifetime", to: "CloutLifetime: normalized.CloutLifetime + 2", refuse: true},
		{name: "restore-value-changed", filename: "save/state.go", from: "CloutLifetime: source.CloutLifetime", to: "CloutLifetime: source.CloutLifetime + 2", refuse: true},
		{name: "copy-removed", filename: "save/state.go", from: "CloutLifetime: source.CloutLifetime", to: "", refuse: true},
		{name: "copy-duplicated", filename: "save/state.go", from: "func RestoreState(source State) State {", to: "func RestoreState(source State) State { _ = State{CloutLifetime: source.CloutLifetime};", refuse: true},
		{name: "other-codec-writer", filename: "save/state.go", extra: "func seeded(state *State) { state.CloutLifetime += 2 }", refuse: true},
		{name: "codec-mutable-alias", filename: "save/state.go", extra: "func seeded(state *State) { p := &state.CloutLifetime; *p += 2 }", refuse: true},
		{name: "copy-shape-outside-codec", filename: "production/seeded_hook.go", refuse: true},
	} {
		t.Run(row.name, func(t *testing.T) {
			source := healthy
			if row.from != "" {
				if strings.Count(source, row.from) != 1 {
					t.Fatal("probe does not select exactly one codec expression")
				}
				source = strings.Replace(source, row.from, row.to, 1)
			}
			source += row.extra
			set := token.NewFileSet()
			file, err := parser.ParseFile(set, row.filename, source, 0)
			if err != nil {
				t.Fatal(err)
			}
			config := types.Config{}
			if _, err := config.Check("save", set, []*ast.File{file}, nil); err != nil {
				t.Fatal(err)
			}
			routes := cloutWriters(t, map[string][]byte{row.filename: []byte(source)})
			if row.refuse && len(routes) == 0 || !row.refuse && len(routes) != 0 {
				t.Fatalf("refuse=%t, mutation routes=%v", row.refuse, routes)
			}
		})
	}
}

// Exercise the actual repository collector, not a separate hand-built codec
// inventory. Mutations remain in memory; no production source file is edited.
func TestGaiaRepositoryCodecBoundary(t *testing.T) {
	sources := cloutRepositorySources(t)
	codec := sources["save/state.go"]
	if routes := cloutWriters(t, sources); len(routes) != 0 {
		t.Fatalf("healthy source inventory refuses: %v", routes)
	}
	for _, body := range []string{
		"state.CloutLifetime += 2",
		"(state.CloutLifetime)++",
		"p := &state.CloutLifetime; *p += 2",
	} {
		t.Run(body, func(t *testing.T) {
			sources["save/state.go"] = append(append([]byte(nil), codec...), []byte("\nfunc gaiaSeededCodecWriter(state *State) { "+body+" }\n")...)
			routes := cloutWriters(t, sources)
			if len(routes) != 1 || !strings.HasPrefix(routes[0], "save/state.go:") {
				t.Fatalf("seeded codec write escaped: %v", routes)
			}
		})
	}
	sources["save/state.go"] = codec
}

func cloutUnparen(expression ast.Expr) ast.Expr {
	for {
		parenthesized, ok := expression.(*ast.ParenExpr)
		if !ok {
			return expression
		}
		expression = parenthesized.X
	}
}

// cloutWriters is a conservative source-ownership guard, not a whole-program
// alias/reflect proof. Besides explicit writes, it reports taking the field's
// address: that hands mutation authority to an alias or another function.
// Ordinary value reads remain allowed. The codec is scanned, with exactly one
// same-field round-trip copy in each of its two named entry points.
func cloutWriters(t *testing.T, sources map[string][]byte) []string {
	t.Helper()
	writers := []string{}
	codecCopies := map[string]int{}
	codecOwners := map[string]string{"EncodeStateVersion": "normalized", "RestoreState": "source"}
	for name, source := range sources {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, name, source, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, declaration := range file.Decls {
			function := ""
			if value, ok := declaration.(*ast.FuncDecl); ok && value.Recv == nil {
				function = value.Name.Name
			}
			ast.Inspect(declaration, func(node ast.Node) bool {
				record := func(position token.Pos) {
					writers = append(writers, name+":"+strings.TrimPrefix(fileSet.Position(position).String(), name+":"))
				}
				switch value := node.(type) {
				case *ast.AssignStmt:
					for _, left := range value.Lhs {
						if selector, ok := cloutUnparen(left).(*ast.SelectorExpr); ok && selector.Sel.Name == "CloutLifetime" {
							record(selector.Pos())
						}
					}
				case *ast.IncDecStmt:
					if selector, ok := cloutUnparen(value.X).(*ast.SelectorExpr); ok && selector.Sel.Name == "CloutLifetime" {
						record(selector.Pos())
					}
				case *ast.UnaryExpr:
					if selector, ok := cloutUnparen(value.X).(*ast.SelectorExpr); value.Op == token.AND && ok && selector.Sel.Name == "CloutLifetime" {
						record(value.Pos())
					}
				case *ast.KeyValueExpr:
					if key, ok := value.Key.(*ast.Ident); ok && key.Name == "CloutLifetime" {
						if name == "save/state.go" && codecOwners[function] != "" {
							if field, ok := cloutUnparen(value.Value).(*ast.SelectorExpr); ok && field.Sel.Name == "CloutLifetime" {
								if owner, ok := cloutUnparen(field.X).(*ast.Ident); ok && owner.Name == codecOwners[function] {
									codecCopies[function]++
									return true
								}
							}
						}
						record(key.Pos())
					}
				}
				return true
			})
		}
	}
	if _, present := sources["save/state.go"]; present {
		for function := range codecOwners {
			if codecCopies[function] != 1 {
				writers = append(writers, fmt.Sprintf("save/state.go:%s requires exactly one round-trip copy, got %d", function, codecCopies[function]))
			}
		}
	}
	sort.Strings(writers)
	return writers
}

func cloutRepositorySources(t *testing.T) map[string][]byte {
	t.Helper()
	sources := map[string][]byte{}
	root := filepath.Join("..")
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "testdata" || entry.Name() == ".gocache" || strings.HasPrefix(entry.Name(), ".")) && path != root {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, relativeErr := filepath.Rel(root, path)
		if relativeErr != nil {
			return relativeErr
		}
		data, readErr := os.ReadFile(path)
		sources[filepath.ToSlash(relative)] = data
		return readErr
	}); err != nil {
		t.Fatal(err)
	}
	if len(sources) < 100 {
		t.Fatalf("scanned only %d server files; the walk is broken", len(sources))
	}
	if len(sources["save/state.go"]) == 0 {
		t.Fatal("save codec is absent from the source inventory")
	}
	return sources
}

// AC4 source guard under Option A. All repository-local server Go sources are
// inspected, including the codec; the two named persistence copies are not mints.
// Catalog checks below concern only the pinned and fixture catalogs, not every
// future reward/effect union or an independent whole-feature acceptance proof.
func TestGaiaLawNoCloutWriterUnderOptionA(t *testing.T) {
	sources := cloutRepositorySources(t)
	if writers := cloutWriters(t, sources); len(writers) != 0 {
		t.Fatalf("Clout mutation routes under Option A: %v", writers)
	}
	t.Logf("scanned %d non-test server source files, including both codec copies", len(sources))
	// Seeded failing case: a second writer in the achievement hook is found.
	seeded := map[string][]byte{"production/seeded_hook.go": []byte("package production\nfunc seeded(founder *struct{ CloutLifetime int64 }) {\n\tfounder.CloutLifetime += 2\n\tfounder.CloutLifetime++\n\t_ = struct{ CloutLifetime int64 }{CloutLifetime: 1}\n}\n")}
	if got := cloutWriters(t, seeded); !reflect.DeepEqual(got, []string{"production/seeded_hook.go:3:2", "production/seeded_hook.go:4:2", "production/seeded_hook.go:5:36"}) {
		t.Fatalf("seeded writers = %v", got)
	}
	// Neither of these two current catalogs supplies a Clout resource.
	for _, path := range []string{"../../balance/catalogs/phase0.json", "../../balance/testdata/axis-stack/economy-v5-fixture.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		catalog, err := economy.LoadCatalog(data)
		if err != nil {
			t.Fatal(err)
		}
		for _, resource := range catalog.Resources() {
			if strings.Contains(resource.ID, "clout") {
				t.Fatalf("%s declares a Clout resource %q", path, resource.ID)
			}
		}
	}
}
