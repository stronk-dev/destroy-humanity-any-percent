package releasepackage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type BundleInput struct {
	RepositoryRoot         string
	Output                 string
	ServerBinary           string
	ClientDist             string
	MetadataDirectory      string
	GameserverImageArchive string
	ReleaseVersion         string
	SourceCommit           string
	DockerEngineVersion    string
	DockerComposeVersion   string
	Images                 map[string]string
	ImageSBOMs             map[string]string
}

// AssembleBundle creates the release unit from declared build outputs. The
// destination must not already contain bytes: a failed build can therefore
// never be mistaken for a previously successful release.
func AssembleBundle(input BundleInput) (ReleaseManifest, error) {
	if input.RepositoryRoot == "" || input.ServerBinary == "" || input.ClientDist == "" || input.MetadataDirectory == "" || input.GameserverImageArchive == "" || len(input.Images) != 3 || len(input.ImageSBOMs) != 3 {
		return ReleaseManifest{}, ErrInvalidContent
	}
	if err := requireEmptyDestination(input.Output); err != nil {
		return ReleaseManifest{}, err
	}
	if err := ValidateDeploymentSchemas(input.RepositoryRoot); err != nil {
		return ReleaseManifest{}, err
	}

	closure, err := StageRuntimeContent(input.RepositoryRoot, filepath.Join(input.Output, "content"))
	if err != nil {
		return ReleaseManifest{}, err
	}
	for source, destination := range map[string]string{
		filepath.Join(input.RepositoryRoot, "deployment", ".env.example"):                  ".env.example",
		filepath.Join(input.RepositoryRoot, "deployment", "Caddyfile"):                     "Caddyfile",
		filepath.Join(input.RepositoryRoot, "deployment", "Dockerfile.gameserver"):         "Dockerfile.gameserver",
		filepath.Join(input.RepositoryRoot, "deployment", "compose.rotation.template.yml"): "compose.rotation.yml",
		filepath.Join(input.RepositoryRoot, "deployment", "config.schema.json"):            "config.schema.json",
		filepath.Join(input.RepositoryRoot, "deployment", "release-manifest.schema.json"):  "release-manifest.schema.json",
		filepath.Join(input.RepositoryRoot, "LICENSE"):                                     "LICENSE",
		filepath.Join(input.MetadataDirectory, "third-party-licenses.txt"):                 "third-party-licenses.txt",
		filepath.Join(input.MetadataDirectory, "sbom.spdx.json"):                           "sbom/application.spdx.json",
	} {
		if err := copyRegularFile(source, filepath.Join(input.Output, filepath.FromSlash(destination)), 0o644); err != nil {
			return ReleaseManifest{}, err
		}
	}
	if err := copyLinuxAMD64Binary(input.ServerBinary, filepath.Join(input.Output, "gameserver")); err != nil {
		return ReleaseManifest{}, err
	}
	if err := copyRegularFile(input.GameserverImageArchive, filepath.Join(input.Output, "images", "gameserver.docker.tar"), 0o644); err != nil {
		return ReleaseManifest{}, err
	}
	if err := ValidateDockerArchive(filepath.Join(input.Output, "images", "gameserver.docker.tar"), input.Images["gameserver"], input.ReleaseVersion, input.SourceCommit); err != nil {
		return ReleaseManifest{}, err
	}
	findings, err := ScanDockerArchive(filepath.Join(input.Output, "images", "gameserver.docker.tar"))
	if err != nil || RequireNoSecrets(findings) != nil {
		return ReleaseManifest{}, ErrInvalidContent
	}
	if err := copyTree(input.ClientDist, filepath.Join(input.Output, "site")); err != nil {
		return ReleaseManifest{}, err
	}
	if err := copyRegularFile(filepath.Join(input.MetadataDirectory, "third-party-licenses.txt"), filepath.Join(input.Output, "site", "third-party-licenses.txt"), 0o644); err != nil {
		return ReleaseManifest{}, err
	}
	if info, err := os.Stat(filepath.Join(input.Output, "site", "index.html")); err != nil || !info.Mode().IsRegular() {
		return ReleaseManifest{}, ErrInvalidContent
	}

	template, err := os.ReadFile(filepath.Join(input.RepositoryRoot, "deployment", "compose.template.yml"))
	if err != nil {
		return ReleaseManifest{}, err
	}
	compose, err := RenderCompose(template, input.Images)
	if err != nil {
		return ReleaseManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(input.Output, "compose.yml"), compose, 0o644); err != nil {
		return ReleaseManifest{}, err
	}

	images := make([]Image, 0, 3)
	for _, name := range []string{"caddy", "gameserver", "postgres"} {
		reference, source := input.Images[name], input.ImageSBOMs[name]
		if !imageReferencePattern.MatchString(reference) || source == "" {
			return ReleaseManifest{}, fmt.Errorf("%w: image %s", ErrInvalidContent, name)
		}
		destination := filepath.Join("sbom", name+".spdx.json")
		if err := copyRegularFile(source, filepath.Join(input.Output, destination), 0o644); err != nil {
			return ReleaseManifest{}, err
		}
		data, err := os.ReadFile(filepath.Join(input.Output, destination))
		if err != nil || len(data) == 0 {
			return ReleaseManifest{}, ErrInvalidContent
		}
		images = append(images, Image{Name: name, Reference: reference, SBOMPath: filepath.ToSlash(destination), SBOMSHA256: digest(data)})
	}
	migration, err := CurrentMigration(input.RepositoryRoot)
	if err != nil {
		return ReleaseManifest{}, err
	}
	manifest, encoded, err := BuildReleaseManifest(input.Output, ManifestInput{ReleaseVersion: input.ReleaseVersion, SourceCommit: input.SourceCommit,
		DockerEngineVersion: input.DockerEngineVersion, DockerComposeVersion: input.DockerComposeVersion,
		DatabaseMigration: migration, Closure: closure, Images: images})
	if err != nil {
		return ReleaseManifest{}, err
	}
	if err := os.WriteFile(filepath.Join(input.Output, ReleaseManifestPath), encoded, 0o644); err != nil {
		return ReleaseManifest{}, err
	}
	findings, err = ScanTree(input.Output)
	if err != nil || RequireNoSecrets(findings) != nil {
		return ReleaseManifest{}, ErrInvalidContent
	}
	if err := ValidateBundle(input.Output); err != nil {
		return ReleaseManifest{}, err
	}
	return manifest, nil
}

func copyLinuxAMD64Binary(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil || !isLinuxAMD64ELF(data) {
		return fmt.Errorf("%w: gameserver is not linux/amd64 ELF", ErrInvalidContent)
	}
	return writeNewRegularFile(destination, data, 0o755)
}

func isLinuxAMD64ELF(data []byte) bool {
	return len(data) >= 64 && string(data[:4]) == "\x7fELF" && data[4] == 2 && data[5] == 1 &&
		(binary.LittleEndian.Uint16(data[16:18]) == 2 || binary.LittleEndian.Uint16(data[16:18]) == 3) &&
		binary.LittleEndian.Uint16(data[18:20]) == 62
}

func copyRegularFile(source, destination string, mode fs.FileMode) error {
	info, err := os.Lstat(source)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("%w: release input %q", ErrInvalidContent, source)
	}
	data, err := os.ReadFile(source)
	if err != nil || len(data) == 0 {
		return fmt.Errorf("%w: empty release input %q", ErrInvalidContent, source)
	}
	return writeNewRegularFile(destination, data, mode)
}

func writeNewRegularFile(destination string, data []byte, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func copyTree(source, destination string) error {
	if source == "" {
		return ErrInvalidContent
	}
	paths := []string{}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil || !validRelativePath(filepath.ToSlash(relative)) || entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalidContent
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return ErrInvalidContent
		}
		paths = append(paths, filepath.ToSlash(relative))
		return nil
	})
	if err != nil || len(paths) == 0 {
		return errors.Join(ErrInvalidContent, err)
	}
	sort.Strings(paths)
	for _, relative := range paths {
		if err := copyRegularFile(filepath.Join(source, filepath.FromSlash(relative)), filepath.Join(destination, filepath.FromSlash(relative)), 0o644); err != nil {
			return err
		}
	}
	return nil
}
