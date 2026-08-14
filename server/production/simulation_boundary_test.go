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
	simulationEntrypoints := map[string]bool{
		"SimulateTransition":                true,
		"SimulateAdvance":                   true,
		"SimulateContentDynamicsActivePlay": true,
		"SimulateContentDynamicsFiscal":     true,
		"SimulateContentDynamicsPitch":      true,
		"SimulateContentDynamicsPermits":    true,
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.SelectorExpr:
			found = found || simulationEntrypoints[value.Sel.Name]
		case *ast.Ident:
			found = found || simulationEntrypoints[value.Name]
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
	advance := []byte("package seeded\nfunc invalid() { production.SimulateAdvance() }\n")
	if found, err := hasSimulationCall(advance, "advance.go"); err != nil || !found {
		t.Fatalf("source guard misses advance call: found=%v err=%v", found, err)
	}
	alias := []byte("package seeded\nvar advance = production.SimulateAdvance\nfunc invalid() { advance() }\n")
	if found, err := hasSimulationCall(alias, "alias.go"); err != nil || !found {
		t.Fatalf("source guard misses function-value alias: found=%v err=%v", found, err)
	}
	contentDynamics := []byte("package seeded\nfunc invalid() { production.SimulateContentDynamicsPitch() }\n")
	if found, err := hasSimulationCall(contentDynamics, "content_dynamics.go"); err != nil || !found {
		t.Fatalf("source guard misses content-dynamics call: found=%v err=%v", found, err)
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
		if strings.HasPrefix(relative, "harness"+string(filepath.Separator)) || relative == filepath.Join("production", "simulation.go") ||
			relative == filepath.Join("production", "content_dynamics.go") {
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
