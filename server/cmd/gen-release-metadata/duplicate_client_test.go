package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/releasepackage"
)

func TestClientInventoryValidatesEveryShippedPackageCopyBeforeDeduplicating(t *testing.T) {
	isc := "Copyright (c) other fixture\nPermission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.\nTHE SOFTWARE IS PROVIDED \"AS IS\" AND THE AUTHOR DISCLAIMS ALL WARRANTIES.\n"
	for _, route := range []string{"JavaScript", "CSS"} {
		for _, variant := range []string{"identical", "missing license", "empty license", "unknown license", "metadata mismatch", "different notice", "different license", "different version"} {
			t.Run(route+"/"+variant, func(t *testing.T) {
				root := clientFixture(t, "MIT")
				// The original directory sorts first; this second physical package
				// has the same name/version despite its distinct peer-resolution path.
				module := "node_modules/.pnpm/pad-end@1.0.2_peer@2.0.0/node_modules/pad-end/index.js"
				if route == "CSS" {
					module = strings.TrimSuffix(module, "index.js") + "theme.css"
				}
				directory := filepath.Dir(filepath.Join(root, "client", filepath.FromSlash(module)))
				version, declared, text := "1.0.2", "MIT", mitText
				switch variant {
				case "empty license":
					text = " \n"
				case "unknown license":
					text = "unrecognized fixture license"
				case "metadata mismatch":
					declared = "Apache-2.0"
				case "different notice":
					text = strings.Replace(mitText, "Copyright (c) fixture", "Copyright (c) second fixture", 1)
				case "different license":
					declared, text = "ISC", isc
				case "different version":
					version = "1.0.3"
					text = strings.Replace(mitText, "Copyright (c) fixture", "Copyright (c) second fixture", 1)
				}
				manifest, err := json.Marshal(packageManifest{Name: "pad-end", Version: version, License: declared})
				if err != nil {
					t.Fatal(err)
				}
				writeFixture(t, filepath.Join(directory, "package.json"), string(manifest))
				if variant != "missing license" {
					writeFixture(t, filepath.Join(directory, "LICENSE"), text)
				}
				writeFixture(t, filepath.Join(root, "client", filepath.FromSlash(module)), "fixture shipped module\n")
				if route == "JavaScript" {
					path := filepath.Join(root, "client", "dist", "assets", "index.js.map")
					data, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					var graph map[string]json.RawMessage
					if err := json.Unmarshal(data, &graph); err != nil {
						t.Fatal(err)
					}
					var sources []string
					if err := json.Unmarshal(graph["sources"], &sources); err != nil {
						t.Fatal(err)
					}
					graph["sources"], err = json.Marshal(append(sources, "../../"+module))
					if err != nil {
						t.Fatal(err)
					}
					data, err = json.Marshal(graph)
					if err != nil {
						t.Fatal(err)
					}
					writeFixture(t, path, string(data))
				} else {
					writeCSSFixtureGraph(t, root, map[string]string{"assets/style.css": "body{color:red}\n"}, []string{module})
				}
				dependencies, err := discoverClientDependencies(root)
				if variant != "identical" && variant != "different version" {
					if !errors.Is(err, releasepackage.ErrInvalidContent) || dependencies != nil {
						t.Fatalf("later shipped copy omitted from license validation: error=%v partial_dependencies=%d", err, len(dependencies))
					}
					return
				}
				want := 2
				if variant == "different version" {
					want = 3
				}
				if err != nil || len(dependencies) != want {
					t.Fatalf("valid shipped copies refused: error=%v dependencies=%d want=%d", err, len(dependencies), want)
				}
				notices, err := releasepackage.ThirdPartyNotices(dependencies)
				if err != nil || strings.Count(string(notices), "pad-end 1.0.2 (npm)") != 1 || !strings.Contains(string(notices), "Copyright (c) fixture") {
					t.Fatal("deduplicated copy absent from delivered notice")
				}
				if variant == "different version" && (!strings.Contains(string(notices), "pad-end 1.0.3 (npm)") || !strings.Contains(string(notices), "Copyright (c) second fixture")) {
					t.Fatal("distinct version's notice dropped")
				}
			})
		}
	}
}

// R8 replaced this existing Go-graph test while changing only client inventory.
// Retain its real producer check: helper-only dependencies must remain in the
// application metadata, not just the gameserver's direct imports.
func TestGoDependencyInventoryCoversEveryShippedBinary(t *testing.T) {
	dependencies, err := discoverGoDependencies("../../..")
	if err != nil {
		t.Fatal(err)
	}
	foundAge, foundChromedp := false, false
	for _, dependency := range dependencies {
		foundAge = foundAge || dependency.Name == "filippo.io/age"
		foundChromedp = foundChromedp || dependency.Name == "github.com/chromedp/chromedp"
	}
	if !foundAge || !foundChromedp {
		t.Fatalf("shipped helper dependencies omitted: age=%t browser=%t", foundAge, foundChromedp)
	}
}
