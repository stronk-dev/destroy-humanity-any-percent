package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryCombatDivisionBoundary(t *testing.T) {
	if err := checkDirectory(filepath.Join("..", "..", "combat")); err != nil {
		t.Fatal(err)
	}
}

func TestCombatBoundaryUsesASTAndRecurses(t *testing.T) {
	for _, control := range []struct {
		name, source string
		refused      bool
	}{
		{"binary", "package combat\nfunc f(a,b int) int { return a / b }", true},
		{"assignment", "package combat\nfunc f(a,b int) int { a /= b; return a }", true},
		{"constant", "package combat\nconst ratio = 7 / 2", true},
		{"nested closure", "package combat\nvar f = func() int { return 7 / 2 }", true},
		{"URL cannot hide operator", "package combat\nconst url = `https://example.test`; const ratio = 7 / 2", true},
		{"text and comments", "package combat\nconst label = `13/10`; // a / b\n/* a /= b */", false},
		{"helper call", "package combat\nfunc f(a,b int) int { return sharedFloor(a,b) }", false},
		{"malformed", "package combat\nfunc broken(", true},
	} {
		t.Run(control.name, func(t *testing.T) {
			root := t.TempDir()
			nested := filepath.Join(root, "future", "engine")
			if err := os.MkdirAll(nested, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(nested, "arithmetic.go")
			if err := os.WriteFile(path, []byte(control.source), 0o644); err != nil {
				t.Fatal(err)
			}
			err := checkDirectory(root)
			if (err != nil) != control.refused {
				t.Fatalf("refused=%v err=%v", control.refused, err)
			}
			if err != nil && !strings.Contains(err.Error(), path) {
				t.Fatalf("refusal omitted actual source: %v", err)
			}
		})
	}
}

func TestCombatBoundaryIncludesTestsAndFailsMissingEmptyOrLinkedTrees(t *testing.T) {
	root := t.TempDir()
	if err := checkDirectory(filepath.Join(root, "missing")); err == nil {
		t.Fatal("missing tree accepted")
	}
	if err := checkDirectory(root); err == nil {
		t.Fatal("empty population accepted")
	}
	if err := os.WriteFile(filepath.Join(root, "division_test.go"), []byte("package combat\nconst ratio = 7 / 2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkDirectory(root); err == nil {
		t.Fatal("test-file division accepted")
	}
	linked := t.TempDir()
	if err := os.Symlink(root, filepath.Join(linked, "engine")); err != nil {
		t.Fatal(err)
	}
	if err := checkDirectory(linked); err == nil {
		t.Fatal("linked subtree bypassed boundary")
	}
}
