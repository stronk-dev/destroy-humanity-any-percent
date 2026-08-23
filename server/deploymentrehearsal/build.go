package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"time"
)

type BuildRecord struct {
	SchemaVersion           int          `json:"schema_version"`
	Role                    string       `json:"role"`
	ReleaseVersion          string       `json:"release_version"`
	SourceCommit            string       `json:"source_commit"`
	SourceDateEpoch         int64        `json:"source_date_epoch"`
	CreatedAt               time.Time    `json:"created_at"`
	DockerEngine            string       `json:"docker_engine"`
	DockerCompose           string       `json:"docker_compose"`
	BuildkitImage           string       `json:"buildkit_image"`
	SyftImage               string       `json:"syft_image"`
	ManifestSHA256          string       `json:"manifest_sha256"`
	GameserverArchiveSHA256 string       `json:"gameserver_archive_sha256"`
	Images                  []BuildImage `json:"images"`
	RehearsalImages         []BuildImage `json:"rehearsal_images,omitempty"`
	IndependentRebuild      bool         `json:"independent_rebuild"`
	RebuildManifestSHA256   string       `json:"rebuild_manifest_sha256"`
	RebuildArchiveSHA256    string       `json:"rebuild_archive_sha256"`
	NormalizedSBOMsEqual    bool         `json:"normalized_sboms_equal"`
	ObjectiveCompleted      bool         `json:"objective_completed"`
	GuardExhausted          bool         `json:"guard_exhausted"`
}

type BuildImage struct {
	Name                string `json:"name"`
	Reference           string `json:"reference"`
	RuntimeConfigSHA256 string `json:"runtime_config_sha256"`
	SBOMSHA256          string `json:"sbom_sha256"`
}

var (
	commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	imagePattern  = regexp.MustCompile(`^(?:[a-z0-9][a-z0-9./_-]*(?::[A-Za-z0-9._-]+)?@)?sha256:[0-9a-f]{64}$`)
)

func LoadBuildRecord(path string) (BuildRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BuildRecord{}, err
	}
	return DecodeBuildRecord(data)
}

func DecodeBuildRecord(data []byte) (BuildRecord, error) {
	if containsForbiddenEvidenceKey(data) {
		return BuildRecord{}, ErrInvalid
	}
	var record BuildRecord
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&record) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateBuildRecord(record) != nil {
		return BuildRecord{}, ErrInvalid
	}
	return record, nil
}

func ValidateBuildRecord(record BuildRecord) error {
	if record.SchemaVersion != 1 && record.SchemaVersion != 2 || record.Role != "previous" && record.Role != "candidate" ||
		!versionPattern.MatchString(record.ReleaseVersion) || !commitPattern.MatchString(record.SourceCommit) ||
		record.SourceDateEpoch < 1 || record.CreatedAt.IsZero() || record.CreatedAt.Unix() != record.SourceDateEpoch ||
		record.DockerEngine == "" || record.DockerCompose == "" || !imagePattern.MatchString(record.BuildkitImage) ||
		!imagePattern.MatchString(record.SyftImage) || !hashPattern.MatchString(record.ManifestSHA256) ||
		!hashPattern.MatchString(record.GameserverArchiveSHA256) || !record.IndependentRebuild ||
		record.RebuildManifestSHA256 != record.ManifestSHA256 || record.RebuildArchiveSHA256 != record.GameserverArchiveSHA256 ||
		!record.NormalizedSBOMsEqual || !record.ObjectiveCompleted || record.GuardExhausted || len(record.Images) != 6 ||
		record.SchemaVersion == 1 && len(record.RehearsalImages) != 0 ||
		record.SchemaVersion == 2 && record.Role == "candidate" && len(record.RehearsalImages) != 1 ||
		record.SchemaVersion == 2 && record.Role == "previous" && len(record.RehearsalImages) > 1 {
		return ErrInvalid
	}
	names := []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"}
	for index, image := range record.Images {
		if image.Name != names[index] || !imagePattern.MatchString(image.Reference) || !hashPattern.MatchString(image.RuntimeConfigSHA256) ||
			!hashPattern.MatchString(image.SBOMSHA256) || image.Name == "gameserver" && image.Reference != image.RuntimeConfigSHA256 ||
			image.Name != "gameserver" && image.Reference == image.RuntimeConfigSHA256 {
			return ErrInvalid
		}
	}
	if len(record.RehearsalImages) == 1 {
		image := record.RehearsalImages[0]
		if image.Name != "playwright" || !imagePattern.MatchString(image.Reference) || !hashPattern.MatchString(image.RuntimeConfigSHA256) ||
			!hashPattern.MatchString(image.SBOMSHA256) || image.Reference == image.RuntimeConfigSHA256 {
			return ErrInvalid
		}
	}
	return nil
}
