package deploymentrehearsal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	mutation, ok := bundleMutations[request.Population]
	if !ok {
		return ProbeAccepted, ErrInvalid
	}
	return runBundleMutationProbe(request, mutation, releasepackage.ValidateBundle)
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
