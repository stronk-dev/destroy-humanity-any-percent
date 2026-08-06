package faction

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhase0StockWriterClosure(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	serverRoot := filepath.Dir(workingDirectory)
	allowed := map[string]map[string]bool{
		"StockUnits":         {"faction/hook.go": true, "guild/clearing_store.go": true, "production/replay.go": true, "production/soul_suppression.go": true},
		"StockProgressMS":    {"faction/hook.go": true, "production/soul_suppression.go": true},
		"ConsumedStockUnits": {"guild/clearing_store.go": true, "production/replay.go": true, "production/soul_suppression.go": true},
	}
	found := map[string]bool{}
	err = filepath.WalkDir(serverRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel(serverRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		check := func(expression ast.Expr) {
			selector, ok := expression.(*ast.SelectorExpr)
			if !ok {
				return
			}
			files, tracked := allowed[selector.Sel.Name]
			if !tracked {
				return
			}
			if !files[relative] {
				t.Errorf("unregistered %s writer in %s", selector.Sel.Name, relative)
			}
			found[selector.Sel.Name] = true
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.AssignStmt:
				for _, expression := range typed.Lhs {
					check(expression)
				}
			case *ast.IncDecStmt:
				check(typed.X)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"StockUnits", "StockProgressMS"} {
		if !found[field] {
			t.Fatalf("expected registered %s writer was not found", field)
		}
	}
}
