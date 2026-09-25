package production

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
)

// cloutCodecFiles may carry CloutLifetime through encode/restore; they
// write no new value (the Gaia law, design/02 §6b).
var cloutCodecFiles = map[string]bool{"save/state.go": true}

// cloutWriters returns "file:line" for every assignment, increment, or
// composite-literal key that writes CloutLifetime in the given Go sources.
func cloutWriters(t *testing.T, sources map[string][]byte) []string {
	t.Helper()
	writers := []string{}
	for name, source := range sources {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, name, source, 0)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			record := func(position token.Pos) {
				writers = append(writers, name+":"+strings.TrimPrefix(fileSet.Position(position).String(), name+":"))
			}
			switch value := node.(type) {
			case *ast.AssignStmt:
				for _, left := range value.Lhs {
					if selector, ok := left.(*ast.SelectorExpr); ok && selector.Sel.Name == "CloutLifetime" {
						record(selector.Pos())
					}
				}
			case *ast.IncDecStmt:
				if selector, ok := value.X.(*ast.SelectorExpr); ok && selector.Sel.Name == "CloutLifetime" {
					record(selector.Pos())
				}
			case *ast.KeyValueExpr:
				if key, ok := value.Key.(*ast.Ident); ok && key.Name == "CloutLifetime" {
					record(key.Pos())
				}
			}
			return true
		})
	}
	sort.Strings(writers)
	return writers
}

// TestGaiaLawNoCloutWriterUnderOptionA is AC4 under the owner's Option A
// ruling: outside the save codec nothing writes Clout (the allowed writer set
// is empty), and no economy resource offers a clout arm to a reward union.
func TestGaiaLawNoCloutWriterUnderOptionA(t *testing.T) {
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
		relative, _ := filepath.Rel(root, path)
		if cloutCodecFiles[filepath.ToSlash(relative)] {
			return nil
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
	if writers := cloutWriters(t, sources); len(writers) != 0 {
		t.Fatalf("Clout writers outside the codec under Option A: %v", writers)
	}
	// Seeded failing case: a second writer in the achievement hook is found.
	seeded := map[string][]byte{"production/seeded_hook.go": []byte("package production\nfunc seeded(founder *struct{ CloutLifetime int64 }) {\n\tfounder.CloutLifetime += 2\n\tfounder.CloutLifetime++\n\t_ = struct{ CloutLifetime int64 }{CloutLifetime: 1}\n}\n")}
	if got := cloutWriters(t, seeded); !reflect.DeepEqual(got, []string{"production/seeded_hook.go:3:2", "production/seeded_hook.go:4:2", "production/seeded_hook.go:5:36"}) {
		t.Fatalf("seeded writers = %v", got)
	}
	// No reward union can name a Clout arm: no pinned or fixture economy
	// resource is Clout.
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
