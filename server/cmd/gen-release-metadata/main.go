package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
	arguments := append([]string{"list", "-deps", "-json=Dir,Standard,Module"}, shippedGoCommands...)
	command := exec.Command("go", arguments...)
	command.Dir = filepath.Join(root, "server")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go dependency inventory: %w", err)
	}
	type moduleSource struct {
		Path, Version, Dir string
		Main               bool
	}
	type moduleScopes struct {
		source moduleSource
		dirs   []string
	}
	modules := make(map[string]*moduleScopes)
	var standard []string
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg struct {
			Dir      string
			Standard bool
			Module   *moduleSource
		}
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil || pkg.Dir == "" {
			return nil, fmt.Errorf("%w: invalid Go package inventory", releasepackage.ErrInvalidContent)
		}
		if pkg.Module == nil {
			if !pkg.Standard {
				return nil, fmt.Errorf("%w: unowned Go package", releasepackage.ErrInvalidContent)
			}
			standard = append(standard, pkg.Dir)
			continue
		}
		if pkg.Module.Main {
			continue
		}
		module := *pkg.Module
		if module.Path == "" || module.Version == "" || module.Dir == "" {
			return nil, fmt.Errorf("%w: invalid Go module inventory", releasepackage.ErrInvalidContent)
		}
		identity := module.Path + "@" + module.Version
		scopes, exists := modules[identity]
		if !exists {
			scopes = &moduleScopes{source: module}
			modules[identity] = scopes
		} else if scopes.source.Dir != module.Dir {
			return nil, fmt.Errorf("%w: conflicting Go module directories", releasepackage.ErrInvalidContent)
		}
		scopes.dirs = append(scopes.dirs, pkg.Dir)
	}
	result := make([]releasepackage.Dependency, 0, len(modules)+1)
	for identity, scopes := range modules {
		text, license, notice, err := goAttributionAt(scopes.source.Dir, scopes.dirs)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", identity, err)
		}
		module := scopes.source
		result = append(result, releasepackage.Dependency{Name: module.Path, Version: module.Version, Kind: "go", License: license, LicenseText: text, NoticeText: notice,
			Download: "https://proxy.golang.org/" + module.Path + "/@v/" + module.Version + ".zip", PackageURL: "pkg:golang/" + module.Path + "@" + module.Version})
	}
	goLicenseRoot := runtime.GOROOT()
	if _, _, err := licenseAt(goLicenseRoot); err != nil {
		// Homebrew stores the upstream license above its libexec GOROOT.
		goLicenseRoot = filepath.Dir(goLicenseRoot)
	}
	text, license, notice, err := goAttributionAt(goLicenseRoot, standard)
	if err != nil {
		return nil, fmt.Errorf("Go standard library attribution: %w", err)
	}
	version := strings.TrimPrefix(runtime.Version(), "go")
	result = append(result, releasepackage.Dependency{Name: "go.dev/stdlib", Version: version, Kind: "go", License: license, LicenseText: text, NoticeText: notice,
		Download: "https://go.dev/dl/go" + version + ".src.tar.gz", PackageURL: "pkg:golang/go.dev/stdlib@" + version})
	return result, nil
}

// goAttributionAt includes only the module root and ancestors of actually
// imported packages. An unrelated subtree cannot add or hide a license grant.
func goAttributionAt(root string, packageDirs []string) (string, string, string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", "", err
	}
	scopes := map[string]bool{root: true}
	for _, directory := range packageDirs {
		directory, err = filepath.EvalSymlinks(directory)
		if err != nil {
			return "", "", "", err
		}
		relative, err := filepath.Rel(root, directory)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", "", "", fmt.Errorf("%w: Go package outside source root", releasepackage.ErrInvalidContent)
		}
		for {
			scopes[directory] = true
			if directory == root {
				break
			}
			directory = filepath.Dir(directory)
		}
	}
	directories := make([]string, 0, len(scopes))
	for directory := range scopes {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	var licenses, notices bytes.Buffer
	labels := make(map[string]bool)
	rootLicense := false
	for _, directory := range directories {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return "", "", "", err
		}
		for _, entry := range entries {
			isLicense := isLicenseName(entry.Name())
			isNotice := isNoticeName(entry.Name())
			if !isLicense && !isNotice {
				continue
			}
			if !entry.Type().IsRegular() {
				return "", "", "", releasepackage.ErrInvalidContent
			}
			data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil || len(bytes.TrimSpace(data)) == 0 {
				return "", "", "", releasepackage.ErrInvalidContent
			}
			display, err := filepath.Rel(root, filepath.Join(directory, entry.Name()))
			if err != nil {
				return "", "", "", err
			}
			destination := &notices
			if isLicense {
				label, err := releasepackage.DetectPermissiveLicense(string(data))
				if err != nil {
					return "", "", "", fmt.Errorf("%w: unrecognized scoped license %s", err, filepath.ToSlash(display))
				}
				for _, item := range strings.Split(label, " AND ") {
					labels[item] = true
				}
				rootLicense = rootLicense || directory == root
				destination = &licenses
			}
			fmt.Fprintf(destination, "----- %s -----\n%s\n", filepath.ToSlash(display), strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n")))
		}
	}
	if !rootLicense {
		return "", "", "", releasepackage.ErrInvalidContent
	}
	ordered := make([]string, 0, len(labels))
	for label := range labels {
		ordered = append(ordered, label)
	}
	sort.Strings(ordered)
	return licenses.String(), strings.Join(ordered, " AND "), notices.String(), nil
}

// discoverClientDependencies inventories exactly the npm packages whose
// modules the built client ships, from the build's own sourcemaps. Direct
// package.json entries are not the shipped graph: bundlers inline transitive
// packages and aliases can select a different installed version.
func discoverClientDependencies(root string) ([]releasepackage.Dependency, error) {
	dist := filepath.Join(root, "client", "dist")
	var maps []string
	var scripts []string
	var styles []string
	err := filepath.WalkDir(dist, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: symlink in client build %s", releasepackage.ErrInvalidContent, path)
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".map") {
			maps = append(maps, path)
		}
		if !entry.IsDir() && isClientScript(entry.Name()) {
			scripts = append(scripts, path)
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
			styles = append(styles, path)
		}
		return nil
	})
	if err != nil || len(maps) == 0 || len(scripts) == 0 {
		return nil, fmt.Errorf("%w: build the client with sourcemaps before inventorying shipped dependencies: %v", releasepackage.ErrInvalidContent, err)
	}
	sort.Strings(maps)
	mapSet := make(map[string]bool, len(maps))
	for _, mapPath := range maps {
		mapSet[mapPath] = true
		asset := strings.TrimSuffix(mapPath, ".map")
		if err := requireLinkedSourceMap(asset, mapPath); err != nil {
			return nil, err
		}
	}
	for _, script := range scripts {
		if !mapSet[script+".map"] {
			return nil, fmt.Errorf("%w: shipped JavaScript lacks a sourcemap: %s", releasepackage.ErrInvalidContent, script)
		}
	}
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
	cssPackages, err := discoverCSSPackages(root, dist, styles)
	if err != nil {
		return nil, err
	}
	for _, directory := range cssPackages {
		packages[directory] = "CSS build graph"
	}
	directories := make([]string, 0, len(packages))
	for directory := range packages {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	result := make([]releasepackage.Dependency, 0, len(directories))
	seen := map[string]releasepackage.Dependency{}
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
		text, detected, err := licenseAt(directory)
		if err != nil || installed.License != detected {
			return nil, fmt.Errorf("%w: %s: package license mismatch declared=%q detected=%q: %v", releasepackage.ErrInvalidContent, identity, installed.License, detected, err)
		}
		notice, err := noticeAt(directory)
		if err != nil {
			return nil, fmt.Errorf("%w: %s NOTICE: %v", releasepackage.ErrInvalidContent, identity, err)
		}
		// A peer-resolution path can ship another physical copy of the same
		// package. Its identity is not evidence that its notice was inspected.
		// One output row is safe only when every copy has matching attribution.
		if previous, duplicate := seen[identity]; duplicate {
			if previous.License != detected || previous.LicenseText != text || previous.NoticeText != notice {
				return nil, fmt.Errorf("%w: conflicting shipped attribution for %s", releasepackage.ErrInvalidContent, identity)
			}
			continue
		}
		escaped := strings.ReplaceAll(installed.Name, "/", "%2f")
		archive := strings.TrimPrefix(installed.Name[strings.LastIndex(installed.Name, "/")+1:], "@") + "-" + installed.Version + ".tgz"
		dependency := releasepackage.Dependency{Name: installed.Name, Version: installed.Version, Kind: "npm", License: detected, LicenseText: text, NoticeText: notice,
			Download: "https://registry.npmjs.org/" + escaped + "/-/" + archive, PackageURL: "pkg:npm/" + installed.Name + "@" + installed.Version}
		seen[identity] = dependency
		result = append(result, dependency)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: built client ships no npm package", releasepackage.ErrInvalidContent)
	}
	return result, nil
}

// The build hook records main/worker CSS modules and their URL resources because
// Vite's production sourcemap does not cover extracted CSS or all worker inputs.
// Its asset hashes bind the graph to the exact CSS bytes being inventoried.
func discoverCSSPackages(root, dist string, styles []string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(dist, "css-dependency-graph.json"))
	if err != nil {
		return nil, fmt.Errorf("%w: missing CSS build graph: %v", releasepackage.ErrInvalidContent, err)
	}
	var graph struct {
		SchemaVersion int `json:"schema_version"`
		Assets        []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"assets"`
		PackageCSSModules []string `json:"package_css_modules"`
		PackageCSSAssets  []string `json:"package_css_assets"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&graph) != nil || decoder.Decode(&struct{}{}) != io.EOF || graph.SchemaVersion != 2 || graph.Assets == nil || graph.PackageCSSModules == nil || graph.PackageCSSAssets == nil {
		return nil, fmt.Errorf("%w: invalid CSS build graph", releasepackage.ErrInvalidContent)
	}
	actual := make(map[string]string, len(styles))
	for _, style := range styles {
		relative, err := filepath.Rel(dist, style)
		if err != nil {
			return nil, err
		}
		content, err := os.ReadFile(style)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(content)
		actual[filepath.ToSlash(relative)] = hex.EncodeToString(sum[:])
	}
	previous := ""
	for _, asset := range graph.Assets {
		if asset.Path <= previous || !strings.HasSuffix(asset.Path, ".css") || actual[asset.Path] != asset.SHA256 {
			return nil, fmt.Errorf("%w: CSS asset not bound to build graph: %s", releasepackage.ErrInvalidContent, asset.Path)
		}
		previous = asset.Path
	}
	if len(graph.Assets) != len(actual) {
		return nil, fmt.Errorf("%w: CSS build graph omits an emitted asset", releasepackage.ErrInvalidContent)
	}
	packages := map[string]bool{}
	for _, resourceSet := range []struct {
		paths   []string
		cssOnly bool
	}{
		{graph.PackageCSSModules, true},
		{graph.PackageCSSAssets, false},
	} {
		previous = ""
		for _, resource := range resourceSet.paths {
			path := filepath.FromSlash(resource)
			if resource <= previous || !filepath.IsLocal(path) || filepath.ToSlash(filepath.Clean(path)) != resource ||
				(resourceSet.cssOnly && !strings.HasSuffix(resource, ".css")) || !strings.Contains("/"+resource, "/node_modules/") {
				return nil, fmt.Errorf("%w: invalid CSS package resource %q", releasepackage.ErrInvalidContent, resource)
			}
			previous = resource
			resolved := filepath.Join(root, "client", path)
			if info, err := os.Stat(resolved); err != nil || info.IsDir() {
				return nil, fmt.Errorf("%w: missing CSS package resource %q", releasepackage.ErrInvalidContent, resource)
			}
			directory, err := packageDirectory(resolved)
			if err != nil {
				return nil, fmt.Errorf("%w: CSS package resource has no package manifest %q", releasepackage.ErrInvalidContent, resource)
			}
			packages[directory] = true
		}
	}
	directories := make([]string, 0, len(packages))
	for directory := range packages {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	return directories, nil
}

func isClientScript(name string) bool {
	return strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".mjs") || strings.HasSuffix(name, ".cjs")
}

// Every shipped script must have its own linked map. An orphan or stale map
// cannot authorize a license inventory for unrelated shipped bytes.
func requireLinkedSourceMap(asset, mapPath string) error {
	data, err := os.ReadFile(asset)
	if err != nil {
		return fmt.Errorf("%w: sourcemap without its asset %s: %v", releasepackage.ErrInvalidContent, mapPath, err)
	}
	var marker string
	switch {
	case isClientScript(asset):
		marker = "//# sourceMappingURL=" + filepath.Base(mapPath)
	case strings.HasSuffix(asset, ".css"):
		marker = "/*# sourceMappingURL=" + filepath.Base(mapPath) + " */"
	default:
		return fmt.Errorf("%w: unsupported client sourcemap asset %s", releasepackage.ErrInvalidContent, asset)
	}
	if !strings.HasSuffix(strings.TrimSpace(string(data)), marker) || strings.Count(string(data), "sourceMappingURL=") != 1 {
		return fmt.Errorf("%w: asset is not linked to its sourcemap %s", releasepackage.ErrInvalidContent, asset)
	}
	return nil
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

func isLicenseName(name string) bool {
	name = strings.ToLower(name)
	return name == "license" || strings.HasPrefix(name, "license.") || name == "copying" || strings.HasPrefix(name, "copying.")
}

func isNoticeName(name string) bool {
	name = strings.ToLower(name)
	return name == "notice" || strings.HasPrefix(name, "notice.")
}

func licenseAt(directory string) (string, string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", "", err
	}
	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() && isLicenseName(entry.Name()) {
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

// Attribution notices are delivered separately from the texts that establish
// the package's license. A NOTICE cannot widen the license allowlist or supply
// a missing LICENSE, and every present notice must be readable and nonempty.
func noticeAt(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	var candidates []string
	for _, entry := range entries {
		if !isNoticeName(entry.Name()) {
			continue
		}
		if !entry.Type().IsRegular() {
			return "", releasepackage.ErrInvalidContent
		}
		candidates = append(candidates, entry.Name())
	}
	sort.Strings(candidates)
	var combined bytes.Buffer
	for _, candidate := range candidates {
		data, err := os.ReadFile(filepath.Join(directory, candidate))
		if err != nil || len(bytes.TrimSpace(data)) == 0 {
			return "", releasepackage.ErrInvalidContent
		}
		fmt.Fprintf(&combined, "----- %s -----\n%s\n", candidate, strings.TrimSpace(strings.ReplaceAll(string(data), "\r\n", "\n")))
	}
	return combined.String(), nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
