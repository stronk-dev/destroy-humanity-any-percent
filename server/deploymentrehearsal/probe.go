package deploymentrehearsal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/deploymentconfig"
	"cloud-clicker/server/releasepackage"
)

type ProbeOutcome int

const (
	ProbeAccepted ProbeOutcome = iota
	ProbeRejected
)

type ProbeRequest struct {
	Population      string
	CandidateBundle string
	PreviousBundle  string
	WorkDirectory   string
}

type bundleMutation struct {
	path   string
	mutate func([]byte) ([]byte, error)
}

var bundleMutations = map[string]bundleMutation{
	"removed_catalog":               {path: "content/balance/catalogs/phase0.json"},
	"removed_client":                {path: "site/index.html"},
	"removed_license":               {path: "LICENSE"},
	"removed_config":                {path: "config.schema.json"},
	"removed_helper":                {path: "deployment-release"},
	"changed_image_digest":          {path: "release-manifest.json", mutate: replaceManifestHash("reference")},
	"changed_runtime_config_digest": {path: "release-manifest.json", mutate: replaceManifestHash("runtime_config_sha256")},
	"changed_sbom":                  {path: "sbom/caddy.spdx.json", mutate: appendByte},
}

// RunProbe returns ProbeRejected only when a fully prepared, named negative
// fixture reaches the release-package gate and that gate rejects it. Setup,
// input and unsupported-population errors are returned separately so they can
// never satisfy a rehearsal row merely by exiting nonzero.
func RunProbe(request ProbeRequest) (ProbeOutcome, error) {
	if err := validateProbeRequest(request); err != nil {
		return ProbeAccepted, err
	}
	if request.Population == "six_image_sbom_license_provenance" {
		if err := releasepackage.ValidateBundle(request.CandidateBundle); err != nil {
			return ProbeAccepted, err
		}
		if err := releasepackage.ValidateBundle(request.PreviousBundle); err != nil {
			return ProbeAccepted, err
		}
		return ProbeAccepted, nil
	}
	if request.Population == "missing_or_malformed_secret" || request.Population == "duplicate_key_id_or_value" ||
		request.Population == "invalid_origin_or_proxy_depth" {
		return runConfigNegativeProbe(request.Population, deploymentconfig.Load)
	}
	if request.Population == "public_metrics_route" {
		return runPreparedBundleProbe(request, preparePublicMetricsRoute, releasepackage.ValidateBundle)
	}
	mutation, ok := bundleMutations[request.Population]
	if !ok {
		return ProbeAccepted, ErrInvalid
	}
	return runBundleMutationProbe(request, mutation, releasepackage.ValidateBundle)
}

func runPreparedBundleProbe(request ProbeRequest, prepare func(string) error, validate func(string) error) (outcome ProbeOutcome, resultErr error) {
	root, err := os.MkdirTemp(request.WorkDirectory, "bundle-probe-")
	if err != nil {
		return ProbeAccepted, err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(root); cleanupErr != nil {
			outcome = ProbeAccepted
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()
	if err := hardlinkBundle(request.CandidateBundle, root); err != nil {
		return ProbeAccepted, err
	}
	if err := prepare(root); err != nil {
		return ProbeAccepted, err
	}
	if err := validate(root); err != nil {
		return ProbeRejected, nil
	}
	return ProbeAccepted, nil
}

func preparePublicMetricsRoute(root string) error {
	caddyPath := filepath.Join(root, "Caddyfile")
	if err := rewriteHardlink(caddyPath, func(data []byte) ([]byte, error) {
		return append(append([]byte(nil), data...), []byte("\nhandle /metrics {\n\treverse_proxy gameserver:8080\n}\n")...), nil
	}); err != nil {
		return err
	}
	return refreshManifestArtifact(root, "Caddyfile")
}

func refreshManifestArtifact(root, artifactPath string) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifactPath)))
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(root, releasepackage.ReleaseManifestPath)
	return rewriteHardlink(manifestPath, func(manifestData []byte) ([]byte, error) {
		var manifest releasepackage.ReleaseManifest
		if json.Unmarshal(manifestData, &manifest) != nil {
			return nil, ErrInvalid
		}
		found := false
		for index := range manifest.Artifacts {
			if manifest.Artifacts[index].Path == artifactPath {
				manifest.Artifacts[index].SHA256 = hashBytes(data)
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalid
		}
		encoded, err := json.MarshalIndent(manifest, "", "  ")
		return append(encoded, '\n'), err
	})
}

type configLoader func([]string, deploymentconfig.ReadFile) (deploymentconfig.Config, error)

type configFixture struct {
	environment []string
	secrets     map[string][]byte
}

func runConfigNegativeProbe(population string, load configLoader) (ProbeOutcome, error) {
	base := rehearsalProductionConfig()
	if _, err := load(base.environment, configReader(base.secrets)); err != nil {
		return ProbeAccepted, err
	}
	mutations, ok := configProbeMutations(population)
	if !ok || len(mutations) == 0 {
		return ProbeAccepted, ErrInvalid
	}
	for _, mutate := range mutations {
		fixture := cloneConfigFixture(base)
		mutate(&fixture)
		if _, err := load(fixture.environment, configReader(fixture.secrets)); err == nil {
			return ProbeAccepted, nil
		}
	}
	return ProbeRejected, nil
}

func configProbeMutations(population string) ([]func(*configFixture), bool) {
	switch population {
	case "missing_or_malformed_secret":
		return []func(*configFixture){
			func(value *configFixture) { delete(value.secrets, "/run/secrets/database-url") },
			func(value *configFixture) { value.secrets["/run/secrets/jwt-current"] = []byte("not-base64\n") },
			func(value *configFixture) { value.secrets["/run/secrets/bootstrap-current"] = encodedProbeKey(31, 3) },
			func(value *configFixture) { removeConfigEnvironment(value, "CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE") },
			func(value *configFixture) { removeConfigEnvironment(value, "CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE") },
		}, true
	case "duplicate_key_id_or_value":
		return []func(*configFixture){
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_JWT_PREVIOUS_ID", "jwt-current")
			},
			func(value *configFixture) {
				value.secrets["/run/secrets/jwt-previous"] = append([]byte(nil), value.secrets["/run/secrets/jwt-current"]...)
			},
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID", "bootstrap-current")
			},
			func(value *configFixture) {
				value.secrets["/run/secrets/bootstrap-previous"] = append([]byte(nil), value.secrets["/run/secrets/bootstrap-current"]...)
			},
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_CURSOR_PREVIOUS_ID", "cursor-current")
			},
			func(value *configFixture) {
				value.secrets["/run/secrets/cursor-previous"] = append([]byte(nil), value.secrets["/run/secrets/cursor-current"]...)
			},
		}, true
	case "invalid_origin_or_proxy_depth":
		return []func(*configFixture){
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_PUBLIC_ORIGIN", "http://play.example.test")
			},
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_PUBLIC_ORIGIN", "https://play.example.test/game")
			},
			func(value *configFixture) {
				setConfigEnvironment(value, "CLOUD_CLICKER_PUBLIC_ORIGIN", "https://PLAY.example.test")
			},
			func(value *configFixture) { setConfigEnvironment(value, "CLOUD_CLICKER_TRUSTED_PROXY_HOPS", "0") },
			func(value *configFixture) { setConfigEnvironment(value, "CLOUD_CLICKER_TRUSTED_PROXY_HOPS", "2") },
		}, true
	default:
		return nil, false
	}
}

func rehearsalProductionConfig() configFixture {
	environment := []string{
		"CLOUD_CLICKER_DEPLOYMENT_MODE=production",
		"CLOUD_CLICKER_PUBLIC_ORIGIN=https://play.example.test",
		"CLOUD_CLICKER_TRUSTED_PROXY_HOPS=1",
		"CLOUD_CLICKER_CONTENT_ROOT=/opt/cloud-clicker/content",
		"CLOUD_CLICKER_SERVER_ID=01986666-b001-4000-8000-000000000001",
		"DATABASE_URL_FILE=/run/secrets/database-url",
		"CLOUD_CLICKER_JWT_CURRENT_ID=jwt-current",
		"CLOUD_CLICKER_JWT_CURRENT_KEY_FILE=/run/secrets/jwt-current",
		"CLOUD_CLICKER_JWT_PREVIOUS_ID=jwt-previous",
		"CLOUD_CLICKER_JWT_PREVIOUS_KEY_FILE=/run/secrets/jwt-previous",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID=bootstrap-current",
		"CLOUD_CLICKER_BOOTSTRAP_CURRENT_KEY_FILE=/run/secrets/bootstrap-current",
		"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_ID=bootstrap-previous",
		"CLOUD_CLICKER_BOOTSTRAP_PREVIOUS_KEY_FILE=/run/secrets/bootstrap-previous",
		"CLOUD_CLICKER_CURSOR_CURRENT_ID=cursor-current",
		"CLOUD_CLICKER_CURSOR_CURRENT_KEY_FILE=/run/secrets/cursor-current",
		"CLOUD_CLICKER_CURSOR_PREVIOUS_ID=cursor-previous",
		"CLOUD_CLICKER_CURSOR_PREVIOUS_KEY_FILE=/run/secrets/cursor-previous",
	}
	return configFixture{environment: environment, secrets: map[string][]byte{
		"/run/secrets/database-url":       []byte("postgres://cloud:fixture@clicker-db/cloud?sslmode=disable\n"),
		"/run/secrets/jwt-current":        encodedProbeKey(32, 1),
		"/run/secrets/jwt-previous":       encodedProbeKey(32, 2),
		"/run/secrets/bootstrap-current":  encodedProbeKey(32, 3),
		"/run/secrets/bootstrap-previous": encodedProbeKey(32, 4),
		"/run/secrets/cursor-current":     encodedProbeKey(32, 5),
		"/run/secrets/cursor-previous":    encodedProbeKey(32, 6),
	}}
}

func encodedProbeKey(size int, value byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{value}, size)) + "\n")
}

func cloneConfigFixture(source configFixture) configFixture {
	result := configFixture{environment: append([]string(nil), source.environment...), secrets: make(map[string][]byte, len(source.secrets))}
	for path, value := range source.secrets {
		result.secrets[path] = append([]byte(nil), value...)
	}
	return result
}

func configReader(secrets map[string][]byte) deploymentconfig.ReadFile {
	return func(path string) ([]byte, error) {
		value, ok := secrets[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return append([]byte(nil), value...), nil
	}
}

func setConfigEnvironment(value *configFixture, name, replacement string) {
	prefix := name + "="
	for index, entry := range value.environment {
		if strings.HasPrefix(entry, prefix) {
			value.environment[index] = prefix + replacement
			return
		}
	}
}

func removeConfigEnvironment(value *configFixture, name string) {
	prefix := name + "="
	result := value.environment[:0]
	for _, entry := range value.environment {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}
	value.environment = result
}

func validateProbeRequest(request ProbeRequest) error {
	if request.Population == "" || !filepath.IsAbs(request.CandidateBundle) || !filepath.IsAbs(request.PreviousBundle) ||
		!filepath.IsAbs(request.WorkDirectory) || filepath.Clean(request.CandidateBundle) != request.CandidateBundle ||
		filepath.Clean(request.PreviousBundle) != request.PreviousBundle || filepath.Clean(request.WorkDirectory) != request.WorkDirectory ||
		request.CandidateBundle == request.PreviousBundle || pathContains(request.CandidateBundle, request.WorkDirectory) ||
		pathContains(request.PreviousBundle, request.WorkDirectory) {
		return ErrInvalid
	}
	for _, path := range []string{request.CandidateBundle, request.PreviousBundle, request.WorkDirectory} {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return errors.Join(ErrInvalid, err)
		}
	}
	return nil
}

func runBundleMutationProbe(request ProbeRequest, mutation bundleMutation, validate func(string) error) (outcome ProbeOutcome, resultErr error) {
	root, err := os.MkdirTemp(request.WorkDirectory, "bundle-probe-")
	if err != nil {
		return ProbeAccepted, err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(root); cleanupErr != nil {
			outcome = ProbeAccepted
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()
	if err := hardlinkBundle(request.CandidateBundle, root); err != nil {
		return ProbeAccepted, err
	}
	target := filepath.Join(root, filepath.FromSlash(mutation.path))
	if mutation.mutate == nil {
		if err := os.Remove(target); err != nil {
			return ProbeAccepted, err
		}
	} else if err := rewriteHardlink(target, mutation.mutate); err != nil {
		return ProbeAccepted, err
	}
	if err := validate(root); err != nil {
		return ProbeRejected, nil
	}
	return ProbeAccepted, nil
}

func hardlinkBundle(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		if entry.IsDir() {
			if relative == "." {
				return nil
			}
			return os.Mkdir(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return ErrInvalid
		}
		return os.Link(path, target)
	})
}

func rewriteHardlink(path string, mutate func([]byte) ([]byte, error)) error {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return errors.Join(ErrInvalid, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	changed, err := mutate(data)
	if err != nil || len(changed) == 0 || string(changed) == string(data) {
		return errors.Join(ErrInvalid, err)
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return os.WriteFile(path, changed, info.Mode().Perm())
}

func replaceManifestHash(field string) func([]byte) ([]byte, error) {
	return func(data []byte) ([]byte, error) {
		needle := []byte(`"` + field + `": "sha256:`)
		index := strings.Index(string(data), string(needle))
		if index < 0 {
			return nil, ErrInvalid
		}
		changed := append([]byte(nil), data...)
		hashIndex := index + len(needle)
		if changed[hashIndex] == '0' {
			changed[hashIndex] = '1'
		} else {
			changed[hashIndex] = '0'
		}
		return changed, nil
	}
}

func appendByte(data []byte) ([]byte, error) {
	return append(append([]byte(nil), data...), ' '), nil
}

func pathContains(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
