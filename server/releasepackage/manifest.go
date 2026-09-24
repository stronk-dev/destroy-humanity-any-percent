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

	yaml "go.yaml.in/yaml/v2"
)

const ReleaseManifestPath = "release-manifest.json"

var (
	commitPattern    = regexp.MustCompile(`^[0-9a-f]{40}$`)
	migrationPattern = regexp.MustCompile(`^([0-9]{5})_[a-z0-9_]+\.sql$`)
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
	RehearsalImages      []Image `json:"rehearsal_images,omitempty"`
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
	RehearsalImages      []Image
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
		Images: append([]Image(nil), input.Images...), RehearsalImages: append([]Image(nil), input.RehearsalImages...), Artifacts: artifacts}
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
	if manifest.SchemaVersion != 1 || !validReleaseVersion(manifest.ReleaseVersion) || !commitPattern.MatchString(manifest.SourceCommit) ||
		manifest.Platform != "linux/amd64" || manifest.DockerEngineVersion == "" || manifest.DockerComposeVersion == "" || manifest.DatabaseMigration < 1 ||
		manifest.CompanySaveVersion < 1 || manifest.CompanySaveVersion > save.LatestCompanyVersion ||
		manifest.FounderSaveVersion < 1 || manifest.FounderSaveVersion > save.LatestFounderVersion || manifest.EpochID < 1 ||
		!hashPattern.MatchString(manifest.ConstantsHash) || !hashPattern.MatchString(manifest.CopyHash) || len(manifest.Images) != len(releaseImageNames) ||
		len(manifest.RehearsalImages) > 1 || len(manifest.Artifacts) == 0 {
		return ErrInvalidContent
	}
	wantImages := releaseImageNames
	for index, image := range manifest.Images {
		if image.Name != wantImages[index] || !imageReferencePattern.MatchString(image.Reference) || !hashPattern.MatchString(image.RuntimeConfigSHA256) || !validRelativePath(image.SBOMPath) || !hashPattern.MatchString(image.SBOMSHA256) {
			return fmt.Errorf("%w: image %d", ErrInvalidContent, index)
		}
	}
	for _, image := range manifest.Images {
		if (image.Name == "gameserver") != strings.HasPrefix(image.Reference, "sha256:") {
			return fmt.Errorf("%w: image reference forms", ErrInvalidContent)
		}
	}
	if len(manifest.RehearsalImages) == 1 {
		image := manifest.RehearsalImages[0]
		if image.Name != "playwright" || !imageReferencePattern.MatchString(image.Reference) || strings.HasPrefix(image.Reference, "sha256:") ||
			!hashPattern.MatchString(image.RuntimeConfigSHA256) || image.SBOMPath != "sbom/playwright.spdx.json" || !hashPattern.MatchString(image.SBOMSHA256) {
			return fmt.Errorf("%w: rehearsal image", ErrInvalidContent)
		}
	}
	gameserverImage := manifest.Images[2]
	if gameserverImage.Reference != gameserverImage.RuntimeConfigSHA256 {
		return fmt.Errorf("%w: gameserver config identity", ErrInvalidContent)
	}
	prior := ""
	for _, artifact := range manifest.Artifacts {
		if !validRelativePath(artifact.Path) || artifact.Path == ReleaseManifestPath || !hashPattern.MatchString(artifact.SHA256) || artifact.Path <= prior {
			return ErrInvalidContent
		}
		prior = artifact.Path
	}
	for _, required := range []string{".env.example", "Caddyfile", "Dockerfile.gameserver", "LICENSE", "compose.yml", "compose.rotation.yml", "config.schema.json", "release-manifest.schema.json", "rehearsal-evidence.schema.json", "images/gameserver.docker.tar", "sbom/application.spdx.json", "third-party-licenses.txt", "site/index.html", "site/third-party-licenses.txt", "gameserver", "deployment-backup", "deployment-release", "deployment-operations", "deployment-rehearsal", "operations/prometheus.yml", "operations/cloud-clicker-alerts.yml", "operations/cloud-clicker-alerts.test.yml", "operations/alertmanager.example.yml", "operations/journald.template.conf", "operations/cloud-clicker-observe.service", "operations/cloud-clicker-observe.timer", "operations/operations.env.example", "content/balance/epochs/phase0.json"} {
		if !hasArtifact(manifest.Artifacts, required) {
			return fmt.Errorf("%w: missing release artifact %q", ErrInvalidContent, required)
		}
	}
	hasBrowser := hasArtifact(manifest.Artifacts, "deployment-browser")
	hasBrowserSBOM := hasArtifact(manifest.Artifacts, "sbom/playwright.spdx.json")
	if hasBrowser != hasBrowserSBOM || (len(manifest.RehearsalImages) == 1) != hasBrowser {
		return fmt.Errorf("%w: incomplete rehearsal browser closure", ErrInvalidContent)
	}
	for _, image := range manifest.Images {
		if artifactHash(manifest.Artifacts, image.SBOMPath) != image.SBOMSHA256 {
			return fmt.Errorf("%w: image SBOM mismatch %s", ErrInvalidContent, image.Name)
		}
	}
	for _, image := range manifest.RehearsalImages {
		if artifactHash(manifest.Artifacts, image.SBOMPath) != image.SBOMSHA256 {
			return fmt.Errorf("%w: rehearsal image SBOM mismatch %s", ErrInvalidContent, image.Name)
		}
	}
	return nil
}

func DecodeReleaseManifest(data []byte) (ReleaseManifest, error) {
	var manifest ReleaseManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&manifest) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateReleaseManifest(manifest) != nil {
		return ReleaseManifest{}, ErrInvalidContent
	}
	return manifest, nil
}

func LoadReleaseManifest(root string) (ReleaseManifest, []byte, error) {
	data, err := os.ReadFile(filepath.Join(root, ReleaseManifestPath))
	if err != nil {
		return ReleaseManifest{}, nil, errors.Join(ErrInvalidContent, err)
	}
	manifest, err := DecodeReleaseManifest(data)
	return manifest, data, err
}

func ValidateBundle(root string) error {
	manifest, _, err := LoadReleaseManifest(root)
	if err != nil {
		return err
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
	compose, err := os.ReadFile(filepath.Join(root, "compose.yml"))
	if err != nil || ValidateCompose(compose) != nil {
		return fmt.Errorf("%w: invalid release Compose", ErrInvalidContent)
	}
	if err := bindComposeImages(compose, manifest.Images); err != nil {
		return err
	}
	if err := bindContentIdentity(filepath.Join(root, "content"), manifest); err != nil {
		return err
	}
	caddyfile, err := os.ReadFile(filepath.Join(root, "Caddyfile"))
	if err != nil || ValidateCaddyfile(caddyfile) != nil {
		return fmt.Errorf("%w: invalid release Caddyfile", ErrInvalidContent)
	}
	dockerfile, err := os.ReadFile(filepath.Join(root, "Dockerfile.gameserver"))
	if err != nil || ValidateGameserverDockerfile(dockerfile) != nil {
		return fmt.Errorf("%w: invalid gameserver Dockerfile", ErrInvalidContent)
	}
	if data, err := os.ReadFile(filepath.Join(root, "sbom", "application.spdx.json")); err != nil || ValidateSPDX(data) != nil {
		return fmt.Errorf("%w: invalid application SPDX document", ErrInvalidContent)
	}
	if err := ValidateOperationsProfile(root); err != nil {
		return err
	}
	for _, image := range append(append([]Image(nil), manifest.Images...), manifest.RehearsalImages...) {
		path := image.SBOMPath
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || ValidateImageSPDX(data, image.RuntimeConfigSHA256) != nil {
			return fmt.Errorf("%w: invalid SPDX document %q", ErrInvalidContent, path)
		}
	}
	if err := ValidateDockerArchive(filepath.Join(root, "images", "gameserver.docker.tar"), manifest.Images[2].Reference, manifest.ReleaseVersion, manifest.SourceCommit); err != nil {
		return err
	}
	for _, name := range []string{"config.schema.json", "release-manifest.schema.json", "rehearsal-evidence.schema.json"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return ErrInvalidContent
		}
		required := map[string][]string{
			"config.schema.json":             {"CLOUD_CLICKER_PUBLIC_ORIGIN", "CLOUD_CLICKER_SERVER_ID", "CLOUD_CLICKER_JWT_CURRENT_ID", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID", "CLOUD_CLICKER_BACKUP_TARGET", "CLOUD_CLICKER_AGE_RECIPIENT", "CLOUD_CLICKER_OPERATIONS_METRICS", "CLOUD_CLICKER_RECEIVER_HEALTH_URL", "CLOUD_CLICKER_ALERTMANAGER_CONFIG", "CLOUD_CLICKER_DATABASE_URL_SECRET_FILE", "CLOUD_CLICKER_POSTGRES_PASSWORD_SECRET_FILE", "CLOUD_CLICKER_JWT_CURRENT_SECRET_FILE", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_SECRET_FILE"},
			"release-manifest.schema.json":   {"schema_version", "release_version", "source_commit", "platform", "docker_engine_version", "docker_compose_version", "database_migration", "company_save_version", "founder_save_version", "epoch_id", "constants_hash", "copy_hash", "images", "artifacts"},
			"rehearsal-evidence.schema.json": {"schema_version", "run_id", "manifest_sha256", "previous_manifest_sha256", "release_version", "previous_release_version", "started_at", "completed_at", "host", "tools", "steps", "populations", "objectives", "artifacts", "exclusions", "objective_completed", "guard_exhausted"},
		}[name]
		if validateSchemaRehearsalClosure(data, name, required, len(manifest.RehearsalImages) == 1) != nil {
			return ErrInvalidContent
		}
	}
	return nil
}

// bindComposeImages requires every image the hash-bound Compose file will run
// to be the manifest's declared reference for that service; the backup worker
// runs the Postgres image for its pg_dump/pg_restore major.
func bindComposeImages(compose []byte, images []Image) error {
	var model composeModel
	if yaml.Unmarshal(compose, &model) != nil {
		return ErrInvalidContent
	}
	declared := map[string]string{}
	for _, image := range images {
		declared[image.Name] = image.Reference
	}
	declared["backup"] = declared["postgres"]
	if len(model.Services) != len(declared) {
		return fmt.Errorf("%w: Compose service set differs from manifest images", ErrInvalidContent)
	}
	for name, service := range model.Services {
		if want, ok := declared[name]; !ok || service.Image != want {
			return fmt.Errorf("%w: Compose image for %s differs from the manifest", ErrInvalidContent, name)
		}
	}
	return nil
}

// bindContentIdentity re-derives the runtime closure from the bundle's own
// content tree with the same epoch authority used at build time. The staged
// files must be exactly that closure and its epoch, constants and copy
// identity must equal the manifest's claims. Database migration is compiled
// into the gameserver binary and is bound at runtime instead.
func bindContentIdentity(content string, manifest ReleaseManifest) error {
	closure, err := DeriveRuntimeClosure(content)
	if err != nil {
		return fmt.Errorf("%w: bundle content does not derive a runtime closure", ErrInvalidContent)
	}
	if err := ValidateStagedContent(content, closure); err != nil {
		return err
	}
	if closure.EpochID != manifest.EpochID || closure.ConstantsHash != manifest.ConstantsHash || closure.CopyHash != manifest.CopyHash {
		return fmt.Errorf("%w: manifest content identity differs from bundle content", ErrInvalidContent)
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
