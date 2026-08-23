package releasepackage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

type SecretScanResult struct {
	SchemaVersion       int       `json:"schema_version"`
	ManifestSHA256      string    `json:"manifest_sha256"`
	SourceCommit        string    `json:"source_commit"`
	StartedAt           time.Time `json:"started_at"`
	CompletedAt         time.Time `json:"completed_at"`
	TrackedFiles        int       `json:"tracked_files"`
	ImageArchiveScanned bool      `json:"image_archive_scanned"`
	Findings            int       `json:"findings"`
	ObjectiveCompleted  bool      `json:"objective_completed"`
	GuardExhausted      bool      `json:"guard_exhausted"`
}

func ValidateSecretScanResult(result SecretScanResult) error {
	if result.SchemaVersion != 1 || !hashPattern.MatchString(result.ManifestSHA256) || !commitPattern.MatchString(result.SourceCommit) ||
		result.StartedAt.IsZero() || !result.CompletedAt.After(result.StartedAt) || result.TrackedFiles < 1 || !result.ImageArchiveScanned ||
		result.Findings != 0 || !result.ObjectiveCompleted || result.GuardExhausted {
		return ErrInvalidContent
	}
	return nil
}

func DecodeSecretScanResult(data []byte) (SecretScanResult, error) {
	var result SecretScanResult
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateSecretScanResult(result) != nil {
		return SecretScanResult{}, ErrInvalidContent
	}
	return result, nil
}

func WriteSecretScanResult(path string, result SecretScanResult) error {
	if !filepath.IsAbs(path) || ValidateSecretScanResult(result) != nil {
		return ErrInvalidContent
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written := false
	defer func() {
		_ = file.Close()
		if !written {
			_ = os.Remove(path)
		}
	}()
	if _, err = file.Write(append(data, '\n')); err == nil {
		err = file.Sync()
	}
	if err == nil {
		err = file.Close()
	}
	if err != nil {
		return errors.Join(ErrInvalidContent, err)
	}
	written = true
	return nil
}
