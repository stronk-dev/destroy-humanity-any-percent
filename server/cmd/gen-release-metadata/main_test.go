package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/releasepackage"
)

const mitText = "Copyright (c) fixture\n\nPermission is hereby granted, free of charge, to any person obtaining a copy.\n\nTHE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND.\n"

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func clientFixture(t *testing.T, license string) string {
	t.Helper()
	root := t.TempDir()
	// A transitive package reached only through the build graph; it is not a
	// direct package.json dependency.
	store := filepath.Join(root, "client", "node_modules", ".pnpm", "pad-end@1.0.2", "node_modules", "pad-end")
	writeFixture(t, filepath.Join(store, "package.json"), `{"name":"pad-end","version":"1.0.2","license":"`+license+`"}`)
	writeFixture(t, filepath.Join(store, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(store, "index.js"), "module.exports = 1\n")
	scoped := filepath.Join(root, "client", "node_modules", ".pnpm", "@scope+pkg@2.0.0", "node_modules", "@scope", "pkg")
	writeFixture(t, filepath.Join(scoped, "package.json"), `{"name":"@scope/pkg","version":"2.0.0","license":"MIT"}`)
	writeFixture(t, filepath.Join(scoped, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(scoped, "dist", "index.js"), "export {}\n")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "index.js.map"),
		`{"version":3,"sources":["../../src/main.ts","../../node_modules/.pnpm/pad-end@1.0.2/node_modules/pad-end/index.js","../../node_modules/.pnpm/@scope+pkg@2.0.0/node_modules/@scope/pkg/dist/index.js"],"mappings":""}`)
	return root
}

func TestClientInventoryFollowsShippedModulesNotDirectDependencies(t *testing.T) {
	dependencies, err := discoverClientDependencies(clientFixture(t, "MIT"))
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, dependency := range dependencies {
		names = append(names, dependency.Name+"@"+dependency.Version+"="+dependency.License)
	}
	if strings.Join(names, ",") != "@scope/pkg@2.0.0=MIT,pad-end@1.0.2=MIT" {
		t.Fatalf("shipped client inventory=%v", names)
	}
	if _, err := discoverClientDependencies(clientFixture(t, "Apache-2.0")); err == nil {
		t.Fatal("license metadata mismatch accepted")
	}
	if _, err := discoverClientDependencies(t.TempDir()); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("missing build output accepted: %v", err)
	}
}
