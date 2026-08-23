package deploymentrehearsal

import (
	"encoding/json"
	"slices"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
)

type manifestAuthority struct {
	SourceCommit   string `json:"source_commit"`
	ReleaseVersion string `json:"release_version"`
	EpochID        int64  `json:"epoch_id"`
	Images         []struct {
		Reference string `json:"reference"`
	} `json:"images"`
}

func decodeManifestAuthority(data []byte) (manifestAuthority, error) {
	var manifest manifestAuthority
	if json.Unmarshal(data, &manifest) != nil || !commitPattern.MatchString(manifest.SourceCommit) ||
		!versionPattern.MatchString(manifest.ReleaseVersion) || manifest.EpochID < 1 || len(manifest.Images) != 6 {
		return manifestAuthority{}, ErrInvalid
	}
	for _, image := range manifest.Images {
		if image.Reference == "" {
			return manifestAuthority{}, ErrInvalid
		}
	}
	return manifest, nil
}

func validateOperatorRecords(evidence Evidence, artifacts map[string][]byte) error {
	candidate, err := decodeManifestAuthority(artifacts["candidate_manifest"])
	if err != nil {
		return err
	}
	previous, err := decodeManifestAuthority(artifacts["previous_manifest"])
	if err != nil || candidate.ReleaseVersion != evidence.ReleaseVersion || previous.ReleaseVersion != evidence.PreviousReleaseVersion {
		return ErrInvalid
	}
	install, err := deploymentrelease.DecodeReleaseLedger(artifacts["install_ledger"])
	if err != nil || len(install) != 1 || !matchesInstall(install[0], candidate, evidence.ManifestSHA256) ||
		install[0].StartedAt.Before(evidence.StartedAt) || install[0].CompletedAt.After(evidence.CompletedAt) {
		return ErrInvalid
	}
	lifecycle, err := deploymentrelease.DecodeReleaseLedger(artifacts["release_ledger"])
	if err != nil || len(lifecycle) != 3 || !matchesInstall(lifecycle[0], previous, evidence.PreviousManifestSHA256) ||
		!matchesTransition(lifecycle[1], "release", candidate, evidence.ManifestSHA256, previous, evidence.PreviousManifestSHA256) ||
		!matchesTransition(lifecycle[2], "rollback", previous, evidence.PreviousManifestSHA256, candidate, evidence.ManifestSHA256) ||
		lifecycle[0].Operator != lifecycle[1].Operator || lifecycle[1].Operator != lifecycle[2].Operator ||
		lifecycle[1].BackupID != lifecycle[2].BackupID || lifecycle[1].RollbackUntil.Sub(lifecycle[1].CompletedAt) != deploymentrelease.RollbackWindow ||
		lifecycle[2].StartedAt.After(lifecycle[1].RollbackUntil) {
		return ErrInvalid
	}
	for _, record := range lifecycle {
		if record.StartedAt.Before(evidence.StartedAt) || record.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	}
	header, err := deploymentbackup.DecodeHeader(artifacts["backup_header"])
	if err != nil || !header.PreUpgrade || header.UpgradeResolved || header.BackupID != lifecycle[1].BackupID ||
		header.ReleaseManifestSHA256 != evidence.PreviousManifestSHA256 || header.EpochID != previous.EpochID ||
		header.StartedAt.Before(evidence.StartedAt) || header.CompletedAt.After(evidence.CompletedAt) {
		return ErrInvalid
	}
	rotation, err := deploymentrelease.DecodeRotationLedger(artifacts["rotation_ledger"])
	if err != nil || validateRotationAuthority(rotation) != nil {
		return ErrInvalid
	}
	return nil
}

func matchesInstall(record deploymentrelease.ReleaseRecord, manifest manifestAuthority, digest string) bool {
	return record.Action == "install" && record.Result == "succeeded" && record.ReleaseVersion == manifest.ReleaseVersion &&
		record.ManifestSHA256 == digest && slices.Equal(record.ImageDigests, manifestReferences(manifest))
}

func matchesTransition(record deploymentrelease.ReleaseRecord, action string, manifest manifestAuthority, digest string,
	previous manifestAuthority, previousDigest string) bool {
	return record.Action == action && record.Result == "succeeded" && record.ReleaseVersion == manifest.ReleaseVersion &&
		record.ManifestSHA256 == digest && record.PreviousVersion == previous.ReleaseVersion &&
		record.PreviousManifestSHA256 == previousDigest && slices.Equal(record.ImageDigests, manifestReferences(manifest))
}

func manifestReferences(manifest manifestAuthority) []string {
	result := make([]string, len(manifest.Images))
	for index, image := range manifest.Images {
		result[index] = image.Reference
	}
	return result
}

func validateRotationAuthority(records []deploymentrelease.RotationRecord) error {
	families := []deploymentrelease.KeyFamily{deploymentrelease.FamilyJWT, deploymentrelease.FamilyBootstrap, deploymentrelease.FamilyCursor}
	if len(records) != len(families)*2 {
		return ErrInvalid
	}
	for index, family := range families {
		activated, removed := records[index*2], records[index*2+1]
		overlap, err := deploymentrelease.MinimumOverlap(family)
		if err != nil || activated.Family != family || removed.Family != family || activated.Action != "activated" || removed.Action != "removed" ||
			activated.CurrentID != removed.CurrentID || activated.PreviousID != removed.PreviousID || activated.Operator != removed.Operator ||
			removed.OccurredAt.Sub(activated.OccurredAt) < overlap {
			return ErrInvalid
		}
	}
	return nil
}
