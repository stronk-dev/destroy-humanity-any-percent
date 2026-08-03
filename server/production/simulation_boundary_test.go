package production

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hasSimulationCall(data []byte, filename string) (bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), filename, data, 0)
	if err != nil {
		return false, err
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch function := call.Fun.(type) {
		case *ast.SelectorExpr:
			found = found || function.Sel.Name == "SimulateTransition"
		case *ast.Ident:
			found = found || function.Name == "SimulateTransition"
		}
		return !found
	})
	return found, nil
}

func TestSimulationEntrypointCallersAreHarnessOrTests(t *testing.T) {
	seed := []byte("package seeded\nfunc invalid() { production.SimulateTransition() }\n")
	if found, err := hasSimulationCall(seed, "seed.go"); err != nil || !found {
		t.Fatalf("source guard misses seeded call: found=%v err=%v", found, err)
	}
	decoy := []byte("package seeded\nconst text = `SimulateTransition()`\n")
	if found, err := hasSimulationCall(decoy, "decoy.go"); err != nil || found {
		t.Fatalf("source guard rejects decoy: found=%v err=%v", found, err)
	}

	serverRoot := ".."
	err := filepath.WalkDir(serverRoot, func(path string, entry fs.DirEntry, walkErr error) error {
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
		if strings.HasPrefix(relative, "harness"+string(filepath.Separator)) || relative == filepath.Join("production", "simulation.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found, err := hasSimulationCall(data, relative)
		if err != nil {
			return err
		}
		if found {
			t.Errorf("%s calls simulation-only production entrypoint", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
