package releasepackage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/save"
)

const ReleaseManifestPath = "release-manifest.json"

var (
	releaseVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
	commitPattern         = regexp.MustCompile(`^[0-9a-f]{40}$`)
	migrationPattern      = regexp.MustCompile(`^([0-9]{5})_[a-z0-9_]+\.sql$`)
)

type Image struct {
	Name                string `json:"name"`
	Reference           string `json:"reference"`
	RuntimeConfigSHA256 string `json:"runtime_config_sha256"`
	SBOMPath            string `json:"sbom_path"`
	SBOMSHA256          string `json:"sbom_sha256"`
}

type ReleaseManifest struct {
	SchemaVersion        int     `json:"schema_version"`
	ReleaseVersion       string  `json:"release_version"`
	SourceCommit         string  `json:"source_commit"`
	Platform             string  `json:"platform"`
	DockerEngineVersion  string  `json:"docker_engine_version"`
	DockerComposeVersion string  `json:"docker_compose_version"`
	DatabaseMigration    int     `json:"database_migration"`
	CompanySaveVersion   int     `json:"company_save_version"`
	FounderSaveVersion   int     `json:"founder_save_version"`
	EpochID              int64   `json:"epoch_id"`
	ConstantsHash        string  `json:"constants_hash"`
	CopyHash             string  `json:"copy_hash"`
	Images               []Image `json:"images"`
	Artifacts            []File  `json:"artifacts"`
}

type ManifestInput struct {
	ReleaseVersion       string
	SourceCommit         string
	DockerEngineVersion  string
	DockerComposeVersion string
	DatabaseMigration    int
	Closure              Closure
	Images               []Image
}

func BuildReleaseManifest(bundleRoot string, input ManifestInput) (ReleaseManifest, []byte, error) {
	artifacts, err := hashBundleFiles(bundleRoot)
	if err != nil {
		return ReleaseManifest{}, nil, err
	}
	manifest := ReleaseManifest{SchemaVersion: 1, ReleaseVersion: input.ReleaseVersion, SourceCommit: input.SourceCommit,
		Platform: "linux/amd64", DockerEngineVersion: input.DockerEngineVersion, DockerComposeVersion: input.DockerComposeVersion,
		DatabaseMigration: input.DatabaseMigration, CompanySaveVersion: save.LatestCompanyVersion, FounderSaveVersion: save.LatestFounderVersion,
		EpochID: input.Closure.EpochID, ConstantsHash: input.Closure.ConstantsHash, CopyHash: input.Closure.CopyHash,
		Images: append([]Image(nil), input.Images...), Artifacts: artifacts}
	if err := ValidateReleaseManifest(manifest); err != nil {
		return ReleaseManifest{}, nil, err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ReleaseManifest{}, nil, err
	}
	return manifest, append(encoded, '\n'), nil
}

func ValidateReleaseManifest(manifest ReleaseManifest) error {
	if manifest.SchemaVersion != 1 || !releaseVersionPattern.MatchString(manifest.ReleaseVersion) || !commitPattern.MatchString(manifest.SourceCommit) ||
		manifest.Platform != "linux/amd64" || manifest.DockerEngineVersion == "" || manifest.DockerComposeVersion == "" || manifest.DatabaseMigration < 1 ||
		manifest.CompanySaveVersion != save.LatestCompanyVersion || manifest.FounderSaveVersion != save.LatestFounderVersion || manifest.EpochID < 1 ||
		!hashPattern.MatchString(manifest.ConstantsHash) || !hashPattern.MatchString(manifest.CopyHash) || len(manifest.Images) != 3 || len(manifest.Artifacts) == 0 {
		return ErrInvalidContent
	}
	wantImages := []string{"caddy", "gameserver", "postgres"}
	for index, image := range manifest.Images {
		if image.Name != wantImages[index] || !imageReferencePattern.MatchString(image.Reference) || !hashPattern.MatchString(image.RuntimeConfigSHA256) || !validRelativePath(image.SBOMPath) || !hashPattern.MatchString(image.SBOMSHA256) {
			return fmt.Errorf("%w: image %d", ErrInvalidContent, index)
		}
	}
	if strings.HasPrefix(manifest.Images[0].Reference, "sha256:") || !strings.HasPrefix(manifest.Images[1].Reference, "sha256:") || strings.HasPrefix(manifest.Images[2].Reference, "sha256:") {
		return fmt.Errorf("%w: image reference forms", ErrInvalidContent)
	}
	if manifest.Images[1].Reference != manifest.Images[1].RuntimeConfigSHA256 {
		return fmt.Errorf("%w: gameserver config identity", ErrInvalidContent)
	}
	prior := ""
	for _, artifact := range manifest.Artifacts {
		if !validRelativePath(artifact.Path) || artifact.Path == ReleaseManifestPath || !hashPattern.MatchString(artifact.SHA256) || artifact.Path <= prior {
			return ErrInvalidContent
		}
		prior = artifact.Path
	}
	for _, required := range []string{".env.example", "Caddyfile", "Dockerfile.gameserver", "LICENSE", "compose.yml", "compose.rotation.yml", "config.schema.json", "release-manifest.schema.json", "images/gameserver.docker.tar", "sbom/application.spdx.json", "third-party-licenses.txt", "site/index.html", "site/third-party-licenses.txt", "gameserver", "deployment-backup", "content/balance/epochs/phase0.json"} {
		if !hasArtifact(manifest.Artifacts, required) {
			return fmt.Errorf("%w: missing release artifact %q", ErrInvalidContent, required)
		}
	}
	for _, image := range manifest.Images {
		if artifactHash(manifest.Artifacts, image.SBOMPath) != image.SBOMSHA256 {
			return fmt.Errorf("%w: image SBOM mismatch %s", ErrInvalidContent, image.Name)
		}
	}
	return nil
}

func ValidateBundle(root string) error {
	data, err := os.ReadFile(filepath.Join(root, ReleaseManifestPath))
	if err != nil {
		return errors.Join(ErrInvalidContent, err)
	}
	var manifest ReleaseManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&manifest) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateReleaseManifest(manifest) != nil {
		return ErrInvalidContent
	}
	actual, err := hashBundleFiles(root)
	if err != nil || len(actual) != len(manifest.Artifacts) {
		return errors.Join(ErrInvalidContent, err)
	}
	for index := range actual {
		if actual[index] != manifest.Artifacts[index] {
			return fmt.Errorf("%w: artifact mismatch %q", ErrInvalidContent, actual[index].Path)
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "sbom", "application.spdx.json")); err != nil || ValidateSPDX(data) != nil {
		return fmt.Errorf("%w: invalid application SPDX document", ErrInvalidContent)
	}
	for _, image := range manifest.Images {
		path := image.SBOMPath
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || ValidateImageSPDX(data, image.RuntimeConfigSHA256) != nil {
			return fmt.Errorf("%w: invalid SPDX document %q", ErrInvalidContent, path)
		}
	}
	if err := ValidateDockerArchive(filepath.Join(root, "images", "gameserver.docker.tar"), manifest.Images[1].Reference, manifest.ReleaseVersion, manifest.SourceCommit); err != nil {
		return err
	}
	for _, name := range []string{"config.schema.json", "release-manifest.schema.json"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return ErrInvalidContent
		}
		required := map[string][]string{
			"config.schema.json":           {"CLOUD_CLICKER_PUBLIC_ORIGIN", "CLOUD_CLICKER_SERVER_ID", "CLOUD_CLICKER_JWT_CURRENT_ID", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID", "CLOUD_CLICKER_BACKUP_TARGET", "CLOUD_CLICKER_AGE_RECIPIENT", "CLOUD_CLICKER_DATABASE_URL_SECRET_FILE", "CLOUD_CLICKER_POSTGRES_PASSWORD_SECRET_FILE", "CLOUD_CLICKER_JWT_CURRENT_SECRET_FILE", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_SECRET_FILE"},
			"release-manifest.schema.json": {"schema_version", "release_version", "source_commit", "platform", "docker_engine_version", "docker_compose_version", "database_migration", "company_save_version", "founder_save_version", "epoch_id", "constants_hash", "copy_hash", "images", "artifacts"},
		}[name]
		if validateSchema(data, name, required) != nil {
			return ErrInvalidContent
		}
	}
	return nil
}

func CurrentMigration(root string) (int, error) {
	entries, err := os.ReadDir(filepath.Join(root, "server", "save", "migrations"))
	if err != nil || len(entries) == 0 {
		return 0, ErrInvalidContent
	}
	current := 0
	for _, entry := range entries {
		match := migrationPattern.FindStringSubmatch(entry.Name())
		if entry.IsDir() || len(match) != 2 {
			return 0, fmt.Errorf("%w: migration path %q", ErrInvalidContent, entry.Name())
		}
		value := 0
		for _, digit := range match[1] {
			value = value*10 + int(digit-'0')
		}
		if value != current+1 {
			return 0, fmt.Errorf("%w: noncontiguous migration %q", ErrInvalidContent, entry.Name())
		}
		current = value
	}
	return current, nil
}

func hashBundleFiles(root string) ([]File, error) {
	files := []File{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == ReleaseManifestPath {
			return nil
		}
		if !validRelativePath(relative) || entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalidContent
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, File{Path: relative, SHA256: digest(data)})
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrInvalidContent, err)
	}
	sort.Slice(files, func(left, right int) bool { return files[left].Path < files[right].Path })
	return files, nil
}

func hasArtifact(artifacts []File, path string) bool { return artifactHash(artifacts, path) != "" }

func artifactHash(artifacts []File, path string) string {
	index := sort.Search(len(artifacts), func(index int) bool { return artifacts[index].Path >= path })
	if index < len(artifacts) && artifacts[index].Path == path {
		return artifacts[index].SHA256
	}
	return ""
}
