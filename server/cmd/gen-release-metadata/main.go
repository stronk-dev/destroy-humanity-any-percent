package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"cloud-clicker/server/releasepackage"
)

type packageManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	License string `json:"license"`
}

var shippedGoCommands = []string{
	"./cmd/gameserver",
	"./cmd/deployment-backup",
	"./cmd/deployment-browser",
	"./cmd/deployment-release",
	"./cmd/deployment-operations",
	"./cmd/deployment-rehearsal",
}

func main() {
	root := flag.String("root", "..", "repository root")
	output := flag.String("output", "", "empty metadata output directory")
	version := flag.String("version", "", "release version")
	commit := flag.String("commit", "", "full source commit")
	created := flag.String("created", "", "RFC3339 release creation time")
	flag.Parse()
	createdAt, err := time.Parse(time.RFC3339, *created)
	if err != nil || *output == "" || *version == "" || *commit == "" {
		fail(releasepackage.ErrInvalidContent)
	}
	if entries, err := os.ReadDir(*output); err == nil && len(entries) != 0 || err != nil && !errors.Is(err, os.ErrNotExist) {
		fail(releasepackage.ErrInvalidContent)
	}
	if err := os.MkdirAll(*output, 0o755); err != nil {
		fail(err)
	}
	dependencies, err := discoverDependencies(*root)
	if err != nil {
		fail(err)
	}
	notices, err := releasepackage.ThirdPartyNotices(dependencies)
	if err != nil {
		fail(err)
	}
	sbom, err := releasepackage.BuildSPDX("Cloud Clicker", *version, *commit, createdAt, dependencies)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(*output, "third-party-licenses.txt"), notices, 0o644); err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(*output, "sbom.spdx.json"), sbom, 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("release metadata: %d dependencies\n", len(dependencies))
}

func discoverDependencies(root string) ([]releasepackage.Dependency, error) {
	goDependencies, err := discoverGoDependencies(root)
	if err != nil {
		return nil, err
	}
	clientDependencies, err := discoverClientDependencies(root)
	if err != nil {
		return nil, err
	}
	combined := append(goDependencies, clientDependencies...)
	for _, dependency := range combined {
		if err := releasepackage.ValidateDependencies([]releasepackage.Dependency{dependency}); err != nil {
			return nil, fmt.Errorf("dependency %s@%s: %w", dependency.Name, dependency.Version, err)
		}
	}
	return releasepackage.SortDependencies(combined)
}

func discoverGoDependencies(root string) ([]releasepackage.Dependency, error) {
	format := `{{with .Module}}{{if not .Main}}{{.Path}}{{"\t"}}{{.Version}}{{"\t"}}{{.Dir}}{{"\n"}}{{end}}{{end}}`
	arguments := []string{"list", "-deps", "-f", format}
	arguments = append(arguments, shippedGoCommands...)
	command := exec.Command("go", arguments...)
	command.Dir = filepath.Join(root, "server")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go dependency inventory: %w", err)
	}
	unique := map[string]releasepackage.Dependency{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 || fields[0] == "" || fields[1] == "" || fields[2] == "" {
			return nil, fmt.Errorf("%w: invalid Go inventory row %q", releasepackage.ErrInvalidContent, line)
		}
		identity := fields[0] + "@" + fields[1]
		if _, exists := unique[identity]; exists {
			continue
		}
		text, license, err := licenseAt(fields[2])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", identity, err)
		}
		unique[identity] = releasepackage.Dependency{Name: fields[0], Version: fields[1], Kind: "go", License: license, LicenseText: text,
			Download: "https://proxy.golang.org/" + fields[0] + "/@v/" + fields[1] + ".zip", PackageURL: "pkg:golang/" + fields[0] + "@" + fields[1]}
	}
	goLicense, license, err := licenseAt(runtime.GOROOT())
	if err != nil {
		// Homebrew keeps the upstream Go LICENSE one directory above its
		// libexec GOROOT; official toolchains keep it at GOROOT.
		goLicense, license, err = licenseAt(filepath.Dir(runtime.GOROOT()))
		if err != nil {
			return nil, fmt.Errorf("Go standard library license: %w", err)
		}
	}
	version := strings.TrimPrefix(runtime.Version(), "go")
	unique["go.dev/stdlib@"+version] = releasepackage.Dependency{Name: "go.dev/stdlib", Version: version, Kind: "go", License: license, LicenseText: goLicense,
		Download: "https://go.dev/dl/go" + version + ".src.tar.gz", PackageURL: "pkg:golang/go.dev/stdlib@" + version}
	result := make([]releasepackage.Dependency, 0, len(unique))
	for _, dependency := range unique {
		result = append(result, dependency)
	}
	return result, nil
}

// discoverClientDependencies inventories exactly the npm packages whose
// modules the built client ships, from the build's own sourcemaps. Direct
// package.json entries are not the shipped graph: bundlers inline transitive
// packages and aliases can select a different installed version.
func discoverClientDependencies(root string) ([]releasepackage.Dependency, error) {
	dist := filepath.Join(root, "client", "dist")
	var maps []string
	err := filepath.WalkDir(dist, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".map") {
			maps = append(maps, path)
		}
		return nil
	})
	if err != nil || len(maps) == 0 {
		return nil, fmt.Errorf("%w: build the client with sourcemaps before inventorying shipped dependencies: %v", releasepackage.ErrInvalidContent, err)
	}
	sort.Strings(maps)
	packages := map[string]string{}
	for _, mapPath := range maps {
		data, err := os.ReadFile(mapPath)
		if err != nil {
			return nil, err
		}
		var sourceMap struct {
			SourceRoot string   `json:"sourceRoot"`
			Sources    []string `json:"sources"`
		}
		if json.Unmarshal(data, &sourceMap) != nil || len(sourceMap.Sources) == 0 {
			return nil, fmt.Errorf("%w: unreadable sourcemap %s", releasepackage.ErrInvalidContent, filepath.Base(mapPath))
		}
		for _, source := range sourceMap.Sources {
			if !strings.Contains(source, "node_modules/") {
				continue
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(mapPath), filepath.FromSlash(sourceMap.SourceRoot), filepath.FromSlash(source)))
			directory, err := packageDirectory(resolved)
			if err != nil {
				return nil, fmt.Errorf("%w: shipped module %s has no package manifest", releasepackage.ErrInvalidContent, source)
			}
			packages[directory] = source
		}
	}
	directories := make([]string, 0, len(packages))
	for directory := range packages {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	result := make([]releasepackage.Dependency, 0, len(directories))
	seen := map[string]bool{}
	for _, directory := range directories {
		packageBytes, err := os.ReadFile(filepath.Join(directory, "package.json"))
		if err != nil {
			return nil, err
		}
		var installed packageManifest
		if json.Unmarshal(packageBytes, &installed) != nil || installed.Name == "" || installed.Version == "" {
			return nil, fmt.Errorf("%w: shipped package manifest %s", releasepackage.ErrInvalidContent, directory)
		}
		identity := installed.Name + "@" + installed.Version
		if seen[identity] {
			continue
		}
		seen[identity] = true
		text, detected, err := licenseAt(directory)
		if err != nil || installed.License != detected {
			return nil, fmt.Errorf("%s: package license mismatch declared=%q detected=%q: %w", identity, installed.License, detected, err)
		}
		escaped := strings.ReplaceAll(installed.Name, "/", "%2f")
		archive := strings.TrimPrefix(installed.Name[strings.LastIndex(installed.Name, "/")+1:], "@") + "-" + installed.Version + ".tgz"
		result = append(result, releasepackage.Dependency{Name: installed.Name, Version: installed.Version, Kind: "npm", License: detected, LicenseText: text,
			Download: "https://registry.npmjs.org/" + escaped + "/-/" + archive, PackageURL: "pkg:npm/" + installed.Name + "@" + installed.Version})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: built client ships no npm package", releasepackage.ErrInvalidContent)
	}
	return result, nil
}

// packageDirectory walks up from a shipped module to the nearest enclosing
// package root directly beneath a node_modules directory.
func packageDirectory(module string) (string, error) {
	for directory := filepath.Dir(module); ; directory = filepath.Dir(directory) {
		parent := filepath.Dir(directory)
		scopedParent := filepath.Dir(parent)
		underModules := filepath.Base(parent) == "node_modules" ||
			strings.HasPrefix(filepath.Base(parent), "@") && filepath.Base(scopedParent) == "node_modules"
		if underModules {
			if _, err := os.Stat(filepath.Join(directory, "package.json")); err == nil {
				return directory, nil
			}
		}
		if parent == directory {
			return "", os.ErrNotExist
		}
	}
}

func licenseAt(directory string) (string, string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", "", err
	}
	var candidates []string
	for _, entry := range entries {
		name := strings.ToLower(entry.Name())
		if !entry.IsDir() && (name == "license" || strings.HasPrefix(name, "license.") || name == "copying" || strings.HasPrefix(name, "copying.")) {
			candidates = append(candidates, entry.Name())
		}
	}
	sort.Strings(candidates)
	if len(candidates) == 0 {
		return "", "", releasepackage.ErrInvalidContent
	}
	var combined bytes.Buffer
	for _, candidate := range candidates {
		data, err := os.ReadFile(filepath.Join(directory, candidate))
		if err != nil || len(bytes.TrimSpace(data)) == 0 {
			return "", "", releasepackage.ErrInvalidContent
		}
		fmt.Fprintf(&combined, "----- %s -----\n%s\n", candidate, strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n")))
	}
	text := combined.String()
	license, err := releasepackage.DetectPermissiveLicense(text)
	return text, license, err
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
