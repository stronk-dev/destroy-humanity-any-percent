package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/releasepackage"
)

const nestedBSDText = "Copyright (c) scoped contributor\nRedistribution and use in source and binary forms are permitted under these conditions.\nNeither the name of the contributor nor its contributors may endorse derived products.\n"

func scopedGoFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	pkg := filepath.Join(root, "nested", "pkg")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(root, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(root, "NOTICE"), "Root attribution\n")
	writeFixture(t, filepath.Join(root, "nested", "COPYING"), nestedBSDText)
	writeFixture(t, filepath.Join(root, "nested", "NOTICE"), "Nested attribution\n")
	return root, pkg
}

func TestGoAttributionFollowsImportedAncestors(t *testing.T) {
	root, pkg := scopedGoFixture(t)
	// An unimported subtree cannot poison or widen this dependency's inventory.
	writeFixture(t, filepath.Join(root, "unrelated", "LICENSE"), "unsupported unrelated grant")
	text, license, notice, err := goAttributionAt(root, []string{pkg, pkg})
	if err != nil || license != "BSD-3-Clause AND MIT" || strings.Count(text, "Copyright (c) scoped contributor") != 1 ||
		!strings.Contains(text, "nested/COPYING") || !strings.Contains(notice, "Root attribution") || !strings.Contains(notice, "Nested attribution") {
		t.Fatalf("linked scope attribution incomplete: %v %q %q %q", err, license, text, notice)
	}
}

func TestGoAttributionRefusesIncompleteScopes(t *testing.T) {
	for _, variant := range []string{"unknown alongside known", "outside root", "missing root license", "symlink license", "empty notice"} {
		t.Run(variant, func(t *testing.T) {
			root, pkg := scopedGoFixture(t)
			switch variant {
			case "unknown alongside known":
				writeFixture(t, filepath.Join(root, "nested", "LICENSE"), "Mozilla Public License, version 2.0\n")
			case "outside root":
				pkg = t.TempDir()
			case "missing root license":
				if err := os.Remove(filepath.Join(root, "LICENSE")); err != nil {
					t.Fatal(err)
				}
			case "symlink license":
				if err := os.Symlink(filepath.Join(root, "LICENSE"), filepath.Join(root, "nested", "LICENSE")); err != nil {
					t.Fatal(err)
				}
			case "empty notice":
				writeFixture(t, filepath.Join(root, "nested", "NOTICE"), " \n")
			}
			if _, _, _, err := goAttributionAt(root, []string{pkg}); !errors.Is(err, releasepackage.ErrInvalidContent) {
				t.Fatalf("incomplete linked scope accepted: %v", err)
			}
		})
	}
}

func TestGoInventoryPreservesNestedContribution(t *testing.T) {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "../../.."))
	args := append([]string{"list", "-deps", "-json"}, shippedGoCommands...)
	command := exec.Command("go", args...)
	command.Dir = filepath.Join(root, "server")
	data, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	const imported = "github.com/prometheus/client_golang/internal/github.com/golang/gddo/httputil"
	var moduleDir, version string
	for {
		var pkg struct {
			ImportPath string
			Module     *struct{ Path, Version, Dir string }
		}
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if pkg.ImportPath == imported && pkg.Module != nil {
			moduleDir, version = pkg.Module.Dir, pkg.Module.Version
		}
	}
	if moduleDir == "" || version == "" {
		t.Fatal("no actual shipped-command package witness")
	}
	license, err := os.ReadFile(filepath.Join(moduleDir, "internal/github.com/golang/gddo/LICENSE"))
	if err != nil || len(license) == 0 {
		t.Fatal("actual nested license missing", err)
	}
	dependencies, err := discoverGoDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	dependencies, err = releasepackage.SortDependencies(dependencies)
	if err != nil {
		t.Fatal(err)
	}
	var selected []releasepackage.Dependency
	for _, dependency := range dependencies {
		if dependency.Name == "github.com/prometheus/client_golang" {
			selected = append(selected, dependency)
		}
	}
	if len(selected) != 1 {
		t.Fatal("actual module inventory binding missing")
	}
	output, err := releasepackage.ThirdPartyNotices(selected)
	if err != nil {
		t.Fatal(err)
	}
	normalize := func(data []byte) string { return strings.Join(strings.Fields(string(data)), " ") }
	if !strings.Contains(normalize(output), normalize(license)) {
		t.Fatalf("actual imported %s in module%s omits its nested BSD attribution from delivered output", imported, version)
	}
	t.Logf("actual imported%s nested license delivered", imported)
}
