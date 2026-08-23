package deploymentrehearsal

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBuildRecordRequiresIndependentExactRebuild(t *testing.T) {
	valid := validBuildRecord()
	if err := ValidateBuildRecord(valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*BuildRecord){
		"no rebuild":       func(value *BuildRecord) { value.IndependentRebuild = false },
		"manifest differs": func(value *BuildRecord) { value.RebuildManifestSHA256 = hashForBuild("e") },
		"archive differs":  func(value *BuildRecord) { value.RebuildArchiveSHA256 = hashForBuild("e") },
		"SBOM differs":     func(value *BuildRecord) { value.NormalizedSBOMsEqual = false },
		"guarded":          func(value *BuildRecord) { value.GuardExhausted = true },
		"mutable image":    func(value *BuildRecord) { value.Images[0].Reference = "alertmanager:latest" },
		"wrong order": func(value *BuildRecord) {
			value.Images[0], value.Images[1] = value.Images[1], value.Images[0]
		},
		"v2 candidate missing rehearsal image": func(value *BuildRecord) {
			value.SchemaVersion = 2
			value.Role = "candidate"
		},
		"v1 rehearsal image": func(value *BuildRecord) {
			value.RehearsalImages = []BuildImage{validRehearsalImage()}
		},
	} {
		t.Run(name, func(t *testing.T) {
			value := validBuildRecord()
			mutate(&value)
			if err := ValidateBuildRecord(value); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid build record accepted: %v", err)
			}
		})
	}
}

func TestCandidateBuildRecordBindsRehearsalBrowserImage(t *testing.T) {
	value := validBuildRecord()
	value.SchemaVersion = 2
	value.Role = "candidate"
	value.RehearsalImages = []BuildImage{validRehearsalImage()}
	if err := ValidateBuildRecord(value); err != nil {
		t.Fatal(err)
	}
	value.RehearsalImages[0].Reference = "playwright:latest"
	if err := ValidateBuildRecord(value); !errors.Is(err, ErrInvalid) {
		t.Fatalf("mutable rehearsal image accepted: %v", err)
	}
}

func validRehearsalImage() BuildImage {
	return BuildImage{Name: "playwright", Reference: "mcr.microsoft.com/playwright:v1@" + hashForBuild("9"),
		RuntimeConfigSHA256: hashForBuild("8"), SBOMSHA256: hashForBuild("7")}
}

func validBuildRecord() BuildRecord {
	created := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	names := []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"}
	images := make([]BuildImage, len(names))
	for index, name := range names {
		reference := name + ":v1@" + hashForBuild(string(rune('1'+index)))
		config := hashForBuild(string(rune('a' + index)))
		if name == "gameserver" {
			reference = config
		}
		images[index] = BuildImage{Name: name, Reference: reference, RuntimeConfigSHA256: config, SBOMSHA256: hashForBuild("d")}
	}
	manifest, archive := hashForBuild("a"), hashForBuild("b")
	return BuildRecord{SchemaVersion: 1, Role: "previous", ReleaseVersion: "0.1.0-preview.0", SourceCommit: strings.Repeat("c", 40),
		SourceDateEpoch: created.Unix(), CreatedAt: created, DockerEngine: "28.4.0", DockerCompose: "2.39.4",
		BuildkitImage: "moby/buildkit:v1@" + hashForBuild("1"), SyftImage: "anchore/syft:v1@" + hashForBuild("2"),
		ManifestSHA256: manifest, GameserverArchiveSHA256: archive, Images: images, IndependentRebuild: true,
		RebuildManifestSHA256: manifest, RebuildArchiveSHA256: archive, NormalizedSBOMsEqual: true, ObjectiveCompleted: true}
}

func hashForBuild(fill string) string { return "sha256:" + strings.Repeat(fill, 64) }
