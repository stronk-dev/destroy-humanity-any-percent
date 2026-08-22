package deploymentbackup

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/releasepackage"
	"filippo.io/age"
)

const (
	metadataName = "metadata.json"
	manifestName = "release-manifest.json"
	epochName    = "epoch.json"
	dumpName     = "database.dump"
	metadataMax  = 1 << 20
	manifestMax  = 16 << 20
	epochMax     = 4 << 20
)

type PackageMetadata struct {
	SchemaVersion         int       `json:"schema_version"`
	BackupID              string    `json:"backup_id"`
	ServerID              string    `json:"server_id"`
	ReleaseManifestSHA256 string    `json:"release_manifest_sha256"`
	EpochSHA256           string    `json:"epoch_sha256"`
	EpochID               int64     `json:"epoch_id"`
	DatabaseDumpSHA256    string    `json:"database_dump_sha256"`
	DatabaseDumpBytes     int64     `json:"database_dump_bytes"`
	StartedAt             time.Time `json:"started_at"`
	CompletedAt           time.Time `json:"completed_at"`
	PreUpgrade            bool      `json:"pre_upgrade"`
	UpgradeResolved       bool      `json:"upgrade_resolved"`
}

type PackageInput struct {
	Directory        string
	BackupID         string
	ServerID         string
	StartedAt        time.Time
	Now              func() time.Time
	Recipient        string
	Dump             io.ReadSeeker
	ReleaseManifest  []byte
	EpochDeclaration []byte
	PreUpgrade       bool
}

type ExtractedPackage struct {
	Header           Header
	Metadata         PackageMetadata
	DumpPath         string
	Manifest         releasepackage.ReleaseManifest
	ReleaseManifest  []byte
	EpochDeclaration []byte
}

type epochDeclaration struct {
	SchemaVersion  int   `json:"schema_version"`
	CurrentEpochID int64 `json:"current_epoch_id"`
}

// CreatePackage binds a custom-format database dump to the exact release and
// epoch bytes needed to interpret it, then commits the encrypted envelope with
// the atomic Create boundary.
func CreatePackage(input PackageInput) (Header, string, error) {
	if input.Now == nil || input.Dump == nil {
		return Header{}, "", ErrInvalid
	}
	manifest, manifestDigest, epochDigest, epochID, err := validateReleaseInputs(input.ReleaseManifest, input.EpochDeclaration)
	if err != nil {
		return Header{}, "", err
	}
	_ = manifest
	dumpHash := sha256.New()
	dumpBytes, err := io.Copy(dumpHash, input.Dump)
	if err != nil || dumpBytes < 1 {
		return Header{}, "", ErrInvalid
	}
	if _, err := input.Dump.Seek(0, io.SeekStart); err != nil {
		return Header{}, "", err
	}
	completedAt := input.Now().UTC()
	metadata := PackageMetadata{
		SchemaVersion: 1, BackupID: input.BackupID, ServerID: input.ServerID,
		ReleaseManifestSHA256: manifestDigest, EpochSHA256: epochDigest, EpochID: epochID,
		DatabaseDumpSHA256: "sha256:" + hex.EncodeToString(dumpHash.Sum(nil)), DatabaseDumpBytes: dumpBytes,
		StartedAt: input.StartedAt.UTC(), CompletedAt: completedAt, PreUpgrade: input.PreUpgrade,
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return Header{}, "", err
	}
	reader, writer := io.Pipe()
	writeResult := make(chan error, 1)
	go func() {
		writeResult <- writePackageTar(writer, completedAt, metadataBytes, input.ReleaseManifest, input.EpochDeclaration, input.Dump, dumpBytes)
	}()
	header, path, createErr := Create(CreateInput{
		Directory: input.Directory, BackupID: input.BackupID, ServerID: input.ServerID,
		ReleaseManifestSHA256: manifestDigest, EpochID: epochID, StartedAt: input.StartedAt,
		Now: func() time.Time { return completedAt }, Recipient: input.Recipient, Dump: reader,
		PreUpgrade: input.PreUpgrade,
	})
	if createErr != nil {
		_ = reader.CloseWithError(createErr)
	}
	writeErr := <-writeResult
	if createErr != nil {
		return Header{}, "", createErr
	}
	if writeErr != nil {
		return Header{}, "", writeErr
	}
	return header, path, nil
}

func writePackageTar(pipe *io.PipeWriter, timestamp time.Time, metadata, manifest, epoch []byte, dump io.Reader, dumpBytes int64) error {
	tarWriter := tar.NewWriter(pipe)
	closeWith := func(err error) error {
		_ = tarWriter.Close()
		_ = pipe.CloseWithError(err)
		return err
	}
	entries := []struct {
		name string
		data []byte
	}{{metadataName, metadata}, {manifestName, manifest}, {epochName, epoch}}
	for _, entry := range entries {
		if err := tarWriter.WriteHeader(&tar.Header{Name: entry.name, Mode: 0o600, Size: int64(len(entry.data)), ModTime: timestamp}); err != nil {
			return closeWith(err)
		}
		if _, err := tarWriter.Write(entry.data); err != nil {
			return closeWith(err)
		}
	}
	if err := tarWriter.WriteHeader(&tar.Header{Name: dumpName, Mode: 0o600, Size: dumpBytes, ModTime: timestamp}); err != nil {
		return closeWith(err)
	}
	if written, err := io.CopyN(tarWriter, dump, dumpBytes); err != nil || written != dumpBytes {
		return closeWith(errors.Join(ErrInvalid, err))
	}
	if err := tarWriter.Close(); err != nil {
		return closeWith(err)
	}
	return pipe.Close()
}

// ExtractPackage authenticates the envelope and all internal identities before
// exposing the custom-format dump to pg_restore. Destination must be empty.
func ExtractPackage(path, expectedManifest string, identity age.Identity, destination string) (ExtractedPackage, error) {
	if err := requireEmptyDirectory(destination); err != nil {
		return ExtractedPackage{}, err
	}
	reader, writer := io.Pipe()
	type restoreResult struct {
		header Header
		err    error
	}
	restored := make(chan restoreResult, 1)
	go func() {
		header, err := Restore(path, expectedManifest, identity, writer)
		_ = writer.CloseWithError(err)
		restored <- restoreResult{header: header, err: err}
	}()
	result, extractErr := extractPackageTar(tar.NewReader(reader), destination)
	if extractErr != nil {
		_ = reader.CloseWithError(extractErr)
	}
	restore := <-restored
	if restore.err != nil {
		return ExtractedPackage{}, restore.err
	}
	if extractErr != nil {
		return ExtractedPackage{}, extractErr
	}
	result.Header = restore.header
	if result.Metadata.BackupID != restore.header.BackupID || result.Metadata.ServerID != restore.header.ServerID ||
		result.Metadata.ReleaseManifestSHA256 != restore.header.ReleaseManifestSHA256 || result.Metadata.EpochID != restore.header.EpochID ||
		!result.Metadata.StartedAt.Equal(restore.header.StartedAt) || !result.Metadata.CompletedAt.Equal(restore.header.CompletedAt) ||
		result.Metadata.PreUpgrade != restore.header.PreUpgrade || result.Metadata.UpgradeResolved != restore.header.UpgradeResolved {
		return ExtractedPackage{}, ErrInvalid
	}
	return result, nil
}

func extractPackageTar(reader *tar.Reader, destination string) (ExtractedPackage, error) {
	var result ExtractedPackage
	want := []string{metadataName, manifestName, epochName, dumpName}
	for _, name := range want {
		header, err := reader.Next()
		if err != nil || header.Name != name || header.Typeflag != tar.TypeReg || header.Size < 1 {
			return ExtractedPackage{}, ErrInvalid
		}
		switch name {
		case metadataName:
			data, err := readBounded(reader, header.Size, metadataMax)
			if err != nil || decodeStrict(data, &result.Metadata) != nil {
				return ExtractedPackage{}, ErrInvalid
			}
		case manifestName:
			data, err := readBounded(reader, header.Size, manifestMax)
			if err != nil {
				return ExtractedPackage{}, ErrInvalid
			}
			result.ReleaseManifest = data
		case epochName:
			data, err := readBounded(reader, header.Size, epochMax)
			if err != nil {
				return ExtractedPackage{}, ErrInvalid
			}
			result.EpochDeclaration = data
		case dumpName:
			if header.Size != result.Metadata.DatabaseDumpBytes || !hash.MatchString(result.Metadata.DatabaseDumpSHA256) {
				return ExtractedPackage{}, ErrInvalid
			}
			dumpPath := filepath.Join(destination, dumpName)
			file, err := os.OpenFile(dumpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
			if err != nil {
				return ExtractedPackage{}, err
			}
			hasher := sha256.New()
			written, copyErr := io.CopyN(io.MultiWriter(file, hasher), reader, header.Size)
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil || written != header.Size || "sha256:"+hex.EncodeToString(hasher.Sum(nil)) != result.Metadata.DatabaseDumpSHA256 {
				return ExtractedPackage{}, ErrInvalid
			}
			result.DumpPath = dumpPath
		}
	}
	if _, err := reader.Next(); err != io.EOF {
		return ExtractedPackage{}, ErrInvalid
	}
	manifest, manifestDigest, epochDigest, epochID, err := validateReleaseInputs(result.ReleaseManifest, result.EpochDeclaration)
	if err != nil || manifestDigest != result.Metadata.ReleaseManifestSHA256 || epochDigest != result.Metadata.EpochSHA256 || epochID != result.Metadata.EpochID {
		return ExtractedPackage{}, ErrInvalid
	}
	result.Manifest = manifest
	if result.Metadata.SchemaVersion != 1 || !backupID.MatchString(result.Metadata.BackupID) || result.Metadata.ServerID == "" ||
		result.Metadata.DatabaseDumpBytes < 1 || result.Metadata.StartedAt.IsZero() || result.Metadata.CompletedAt.Before(result.Metadata.StartedAt) ||
		result.Metadata.UpgradeResolved && !result.Metadata.PreUpgrade {
		return ExtractedPackage{}, ErrInvalid
	}
	return result, nil
}

func validateReleaseInputs(manifestBytes, epochBytes []byte) (releasepackage.ReleaseManifest, string, string, int64, error) {
	var manifest releasepackage.ReleaseManifest
	if decodeStrict(manifestBytes, &manifest) != nil || releasepackage.ValidateReleaseManifest(manifest) != nil {
		return manifest, "", "", 0, ErrInvalid
	}
	var epoch epochDeclaration
	if decodeStrict(epochBytes, &epoch) != nil || epoch.SchemaVersion != 1 || epoch.CurrentEpochID != manifest.EpochID {
		return manifest, "", "", 0, ErrInvalid
	}
	epochDigest := digest(epochBytes)
	found := false
	for _, artifact := range manifest.Artifacts {
		if artifact.Path == "content/balance/epochs/phase0.json" {
			found = artifact.SHA256 == epochDigest
			break
		}
	}
	if !found {
		return manifest, "", "", 0, ErrInvalid
	}
	return manifest, digest(manifestBytes), epochDigest, manifest.EpochID, nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrInvalid
	}
	return nil
}

func readBounded(reader io.Reader, size, maximum int64) ([]byte, error) {
	if size < 1 || size > maximum {
		return nil, ErrInvalid
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}
	return data, nil
}

func requireEmptyDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() || info.Mode()&0o022 != 0 {
		return ErrInvalid
	}
	entries, err := os.ReadDir(path)
	if err != nil || len(entries) != 0 {
		return ErrInvalid
	}
	return nil
}
