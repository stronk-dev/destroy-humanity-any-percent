package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "index.js"), "console.log('fixture')\n//# sourceMappingURL=index.js.map\n")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "index.js.map"),
		`{"version":3,"sources":["../../src/main.ts","../../node_modules/.pnpm/pad-end@1.0.2/node_modules/pad-end/index.js","../../node_modules/.pnpm/@scope+pkg@2.0.0/node_modules/@scope/pkg/dist/index.js"],"mappings":""}`)
	writeFixture(t, filepath.Join(root, "client", "dist", "css-dependency-graph.json"),
		`{"schema_version":2,"assets":[],"package_css_modules":[],"package_css_assets":[]}`)
	return root
}

func writeCSSFixtureGraph(t *testing.T, root string, styles map[string]string, modules []string, assets ...string) {
	t.Helper()
	graph := struct {
		SchemaVersion int `json:"schema_version"`
		Assets        []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"assets"`
		PackageCSSModules []string `json:"package_css_modules"`
		PackageCSSAssets  []string `json:"package_css_assets"`
	}{SchemaVersion: 2, PackageCSSModules: append([]string{}, modules...), PackageCSSAssets: append([]string{}, assets...)}
	graph.Assets = make([]struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	}, 0, len(styles))
	for path, content := range styles {
		writeFixture(t, filepath.Join(root, "client", "dist", filepath.FromSlash(path)), content)
		graph.Assets = append(graph.Assets, struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		}{Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(content)))})
	}
	encoded, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(root, "client", "dist", "css-dependency-graph.json"), string(encoded))
}

func TestClientInventoryRejectsPartialAssetMapPopulation(t *testing.T) {
	root := clientFixture(t, "MIT")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "worker.js"),
		"import 'missing-package'\n//# sourceMappingURL=worker.js.map\n")
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("emitted JavaScript asset without its map accepted: %v", err)
	}

	root = clientFixture(t, "MIT")
	if err := os.Remove(filepath.Join(root, "client", "dist", "assets", "index.js")); err != nil {
		t.Fatal(err)
	}
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("orphan sourcemap without its JavaScript asset accepted: %v", err)
	}

	root = clientFixture(t, "MIT")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "index.js"),
		"console.log('fixture')\n//# sourceMappingURL=some-other.js.map\n")
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("asset linked to a different sourcemap accepted: %v", err)
	}
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

func TestClientInventoryIncludesCSSOnlyPackageAndBindsStyles(t *testing.T) {
	root := clientFixture(t, "MIT")
	module := "node_modules/.pnpm/style-only@1.0.0/node_modules/style-only/theme.css"
	packageRoot := filepath.Join(root, "client", "node_modules", ".pnpm", "style-only@1.0.0", "node_modules", "style-only")
	writeFixture(t, filepath.Join(packageRoot, "package.json"), `{"name":"style-only","version":"1.0.0","license":"MIT"}`)
	writeFixture(t, filepath.Join(packageRoot, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(packageRoot, "theme.css"), "body { color: red; }\n")
	writeCSSFixtureGraph(t, root, map[string]string{"assets/style.css": "body{color:red}\n"}, []string{module})
	dependencies, err := discoverClientDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, dependency := range dependencies {
		if dependency.Name == "style-only" && dependency.Version == "1.0.0" && dependency.License == "MIT" {
			found = true
		}
	}
	if !found {
		t.Fatalf("CSS-only package omitted from shipped inventory: %+v", dependencies)
	}
	notices, err := releasepackage.ThirdPartyNotices(dependencies)
	if err != nil || !strings.Contains(string(notices), "style-only 1.0.0 (npm)") {
		t.Fatalf("CSS-only package missing from delivered notices: %v", err)
	}

	graphPath := filepath.Join(root, "client", "dist", "css-dependency-graph.json")
	if err := os.Remove(graphPath); err != nil {
		t.Fatal(err)
	}
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("missing CSS build graph accepted: %v", err)
	}
	writeCSSFixtureGraph(t, root, map[string]string{"assets/style.css": "body{color:red}\n"}, []string{module})
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "unrecorded.css"), "body{color:blue}\n")
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("unrecorded CSS asset accepted: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "client", "dist", "assets", "unrecorded.css")); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "style.css"), "body{color:blue}\n")
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("CSS bytes changed after provenance accepted: %v", err)
	}
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "style.css"), "body{color:red}\n")
	if err := os.Remove(filepath.Join(packageRoot, "theme.css")); err != nil {
		t.Fatal(err)
	}
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("missing package stylesheet accepted: %v", err)
	}
}

func TestClientInventoryIncludesCSSURLAssetOnlyPackage(t *testing.T) {
	root := clientFixture(t, "MIT")
	module := "node_modules/.pnpm/asset-only@1.0.0/node_modules/asset-only/logo.svg"
	packageRoot := filepath.Join(root, "client", "node_modules", ".pnpm", "asset-only@1.0.0", "node_modules", "asset-only")
	writeFixture(t, filepath.Join(packageRoot, "package.json"), `{"name":"asset-only","version":"1.0.0","license":"MIT"}`)
	writeFixture(t, filepath.Join(packageRoot, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(packageRoot, "logo.svg"), `<svg><text>fixture-logo</text></svg>`)
	writeCSSFixtureGraph(t, root, map[string]string{"assets/style.css": "body{background:url(data:image/svg+xml;base64,fixture)}\n"}, nil, module)
	dependencies, err := discoverClientDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	notices, err := releasepackage.ThirdPartyNotices(dependencies)
	if err != nil || !strings.Contains(string(notices), "asset-only 1.0.0 (npm)") {
		t.Fatalf("CSS URL asset package missing from delivered notices: %v", err)
	}
	if err := os.Remove(filepath.Join(packageRoot, "logo.svg")); err != nil {
		t.Fatal(err)
	}
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("missing CSS URL package asset accepted: %v", err)
	}
}

func TestClientInventoryIncludesWorkerCSSOmittedByJavaScriptMap(t *testing.T) {
	root := clientFixture(t, "MIT")
	module := "node_modules/.pnpm/style-only@1.0.0/node_modules/style-only/theme.css"
	packageRoot := filepath.Join(root, "client", "node_modules", ".pnpm", "style-only@1.0.0", "node_modules", "style-only")
	writeFixture(t, filepath.Join(packageRoot, "package.json"), `{"name":"style-only","version":"1.0.0","license":"MIT"}`)
	writeFixture(t, filepath.Join(packageRoot, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(packageRoot, "theme.css"), "body { color: red; }\n")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "worker.js"),
		"console.log('worker')\n//# sourceMappingURL=worker.js.map\n")
	writeFixture(t, filepath.Join(root, "client", "dist", "assets", "worker.js.map"),
		`{"version":3,"sources":["../../src/worker.ts"],"mappings":""}`)
	writeCSSFixtureGraph(t, root, map[string]string{}, []string{module})
	dependencies, err := discoverClientDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	notices, err := releasepackage.ThirdPartyNotices(dependencies)
	if err != nil || !strings.Contains(string(notices), "style-only 1.0.0 (npm)") {
		t.Fatalf("worker CSS package omitted by JavaScript map missing from graph notice: %v", err)
	}
}

func TestClientInventoryRejectsStaleCSSGraphSchema(t *testing.T) {
	root := clientFixture(t, "MIT")
	writeFixture(t, filepath.Join(root, "client", "dist", "css-dependency-graph.json"),
		`{"schema_version":1,"assets":[],"package_css_modules":[]}`)
	if _, err := discoverClientDependencies(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("stale CSS graph without URL-asset provenance accepted: %v", err)
	}
}
