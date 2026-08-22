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

type clientManifest struct {
	Dependencies map[string]string `json:"dependencies"`
}

type packageManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	License string `json:"license"`
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
	command := exec.Command("go", "list", "-deps", "-f", format, "./cmd/gameserver")
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

func discoverClientDependencies(root string) ([]releasepackage.Dependency, error) {
	data, err := os.ReadFile(filepath.Join(root, "client", "package.json"))
	if err != nil {
		return nil, err
	}
	var manifest clientManifest
	if json.Unmarshal(data, &manifest) != nil || len(manifest.Dependencies) == 0 {
		return nil, fmt.Errorf("%w: client dependency manifest", releasepackage.ErrInvalidContent)
	}
	names := make([]string, 0, len(manifest.Dependencies))
	for name := range manifest.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]releasepackage.Dependency, 0, len(names))
	for _, name := range names {
		directory := filepath.Join(root, "client", "node_modules", filepath.FromSlash(name))
		packageBytes, err := os.ReadFile(filepath.Join(directory, "package.json"))
		if err != nil {
			return nil, err
		}
		var installed packageManifest
		if json.Unmarshal(packageBytes, &installed) != nil || installed.Name != name || installed.Version != manifest.Dependencies[name] {
			return nil, fmt.Errorf("%w: installed client dependency %s expected=%s got=%s@%s", releasepackage.ErrInvalidContent, name, manifest.Dependencies[name], installed.Name, installed.Version)
		}
		text, detected, err := licenseAt(directory)
		if err != nil || installed.License != detected {
			return nil, fmt.Errorf("%s: package license mismatch: %w", name, err)
		}
		escaped := strings.ReplaceAll(name, "/", "%2f")
		archive := strings.TrimPrefix(name[strings.LastIndex(name, "/")+1:], "@") + "-" + installed.Version + ".tgz"
		result = append(result, releasepackage.Dependency{Name: name, Version: installed.Version, Kind: "npm", License: detected, LicenseText: text,
			Download: "https://registry.npmjs.org/" + escaped + "/-/" + archive, PackageURL: "pkg:npm/" + name + "@" + installed.Version})
	}
	return result, nil
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
