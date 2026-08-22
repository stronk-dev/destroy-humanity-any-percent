// Package deploymentrelease owns the fail-closed operator state used by
// releases, rollback, and secret rotation. Ledgers are append-only JSONL and
// intentionally contain identifiers and outcomes, never secret values.
package deploymentrelease

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	ErrInvalid = errors.New("invalid deployment release state")
	identifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
)

type KeyFamily string

const (
	FamilyJWT       KeyFamily = "jwt"
	FamilyBootstrap KeyFamily = "bootstrap"
	FamilyCursor    KeyFamily = "cursor"

	JWTOverlap       = 30 * time.Minute
	BootstrapOverlap = 31 * 24 * time.Hour
	CursorOverlap    = 366 * 24 * time.Hour
)

type RotationRecord struct {
	SchemaVersion int       `json:"schema_version"`
	Family        KeyFamily `json:"family"`
	Action        string    `json:"action"`
	CurrentID     string    `json:"current_id"`
	PreviousID    string    `json:"previous_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	Operator      string    `json:"operator"`
}

type ReleaseRecord struct {
	SchemaVersion          int       `json:"schema_version"`
	Action                 string    `json:"action"`
	ReleaseVersion         string    `json:"release_version"`
	ManifestSHA256         string    `json:"manifest_sha256"`
	PreviousVersion        string    `json:"previous_version,omitempty"`
	PreviousManifestSHA256 string    `json:"previous_manifest_sha256,omitempty"`
	ImageDigests           []string  `json:"image_digests"`
	BackupID               string    `json:"backup_id"`
	StartedAt              time.Time `json:"started_at"`
	CompletedAt            time.Time `json:"completed_at"`
	RollbackUntil          time.Time `json:"rollback_until,omitempty"`
	Result                 string    `json:"result"`
	FailureStage           string    `json:"failure_stage,omitempty"`
	Operator               string    `json:"operator"`
}

func MinimumOverlap(family KeyFamily) (time.Duration, error) {
	switch family {
	case FamilyJWT:
		return JWTOverlap, nil
	case FamilyBootstrap:
		return BootstrapOverlap, nil
	case FamilyCursor:
		return CursorOverlap, nil
	default:
		return 0, ErrInvalid
	}
}

func ActivateRotation(path string, family KeyFamily, currentID, previousID, operator string, now time.Time) error {
	if !validRotationFields(family, currentID, previousID, operator, now) || currentID == previousID {
		return ErrInvalid
	}
	lock, err := acquireOperatorLock(path + ".lock")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := ReadRotationLedger(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	latest := -1
	for index := range records {
		if records[index].Family != family {
			continue
		}
		if records[index].CurrentID == currentID || records[index].PreviousID == currentID {
			return ErrInvalid
		}
		latest = index
	}
	if latest >= 0 && (records[latest].Action != "removed" || records[latest].CurrentID != previousID) {
		return ErrInvalid
	}
	return appendJSONLine(path, RotationRecord{SchemaVersion: 1, Family: family, Action: "activated",
		CurrentID: currentID, PreviousID: previousID, OccurredAt: now.UTC(), Operator: operator})
}

func RemovePrevious(path string, family KeyFamily, currentID, previousID, operator string, now time.Time) error {
	if !validRotationFields(family, currentID, previousID, operator, now) || currentID == previousID {
		return ErrInvalid
	}
	lock, err := acquireOperatorLock(path + ".lock")
	if err != nil {
		return err
	}
	defer lock.Close()
	records, err := ReadRotationLedger(path)
	if err != nil {
		return err
	}
	var activation *RotationRecord
	for index := len(records) - 1; index >= 0; index-- {
		record := records[index]
		if record.Family != family {
			continue
		}
		if record.Action != "activated" || record.CurrentID != currentID || record.PreviousID != previousID {
			return ErrInvalid
		}
		activation = &record
		break
	}
	if activation == nil {
		return ErrInvalid
	}
	overlap, _ := MinimumOverlap(family)
	if now.UTC().Before(activation.OccurredAt.Add(overlap)) {
		return ErrInvalid
	}
	return appendJSONLine(path, RotationRecord{SchemaVersion: 1, Family: family, Action: "removed",
		CurrentID: currentID, PreviousID: previousID, OccurredAt: now.UTC(), Operator: operator})
}

func ReadRotationLedger(path string) ([]RotationRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []RotationRecord{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 64<<10)
	for scanner.Scan() {
		var record RotationRecord
		if decodeExact(scanner.Bytes(), &record) != nil || !validRotationFields(record.Family, record.CurrentID, record.PreviousID, record.Operator, record.OccurredAt) ||
			record.SchemaVersion != 1 || record.CurrentID == record.PreviousID || record.Action != "activated" && record.Action != "removed" {
			return nil, ErrInvalid
		}
		if len(result) > 0 && record.OccurredAt.Before(result[len(result)-1].OccurredAt) {
			return nil, ErrInvalid
		}
		if record.Action == "activated" {
			latest := -1
			for index := range result {
				if result[index].Family != record.Family {
					continue
				}
				if result[index].CurrentID == record.CurrentID || result[index].PreviousID == record.CurrentID {
					return nil, ErrInvalid
				}
				latest = index
			}
			if latest >= 0 && (result[latest].Action != "removed" || result[latest].CurrentID != record.PreviousID) {
				return nil, ErrInvalid
			}
		} else {
			matched := false
			for index := len(result) - 1; index >= 0; index-- {
				prior := result[index]
				if prior.Family != record.Family {
					continue
				}
				overlap, _ := MinimumOverlap(record.Family)
				matched = prior.Action == "activated" && prior.CurrentID == record.CurrentID && prior.PreviousID == record.PreviousID &&
					!record.OccurredAt.Before(prior.OccurredAt.Add(overlap))
				break
			}
			if !matched {
				return nil, ErrInvalid
			}
		}
		result = append(result, record)
	}
	if err := scanner.Err(); err != nil || len(result) == 0 {
		return nil, errors.Join(ErrInvalid, err)
	}
	return result, nil
}

func AppendReleaseRecord(path string, record ReleaseRecord) error {
	if !validReleaseRecord(record) {
		return ErrInvalid
	}
	if records, err := ReadReleaseLedger(path); err == nil {
		if !record.StartedAt.After(records[len(records)-1].CompletedAt) {
			return ErrInvalid
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return appendJSONLine(path, record)
}

func ReadReleaseLedger(path string) ([]ReleaseRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := []ReleaseRecord{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 256<<10)
	for scanner.Scan() {
		var record ReleaseRecord
		if decodeExact(scanner.Bytes(), &record) != nil || !validReleaseRecord(record) {
			return nil, ErrInvalid
		}
		if len(result) > 0 && !record.StartedAt.After(result[len(result)-1].CompletedAt) {
			return nil, ErrInvalid
		}
		result = append(result, record)
	}
	if err := scanner.Err(); err != nil || len(result) == 0 {
		return nil, errors.Join(ErrInvalid, err)
	}
	return result, nil
}

func validRotationFields(family KeyFamily, currentID, previousID, operator string, at time.Time) bool {
	_, err := MinimumOverlap(family)
	return err == nil && identifier.MatchString(currentID) && identifier.MatchString(previousID) && identifier.MatchString(operator) && !at.IsZero()
}

func validReleaseRecord(record ReleaseRecord) bool {
	if record.SchemaVersion != 1 || record.Action != "install" && record.Action != "release" && record.Action != "rollback" ||
		!identifier.MatchString(record.ReleaseVersion) || !hashPattern.MatchString(record.ManifestSHA256) ||
		!identifier.MatchString(record.Operator) || record.StartedAt.IsZero() ||
		record.CompletedAt.Before(record.StartedAt) || record.Result != "succeeded" && record.Result != "failed" ||
		(record.Result == "succeeded") != (record.FailureStage == "") || len(record.ImageDigests) != 0 && len(record.ImageDigests) != 6 {
		return false
	}
	if record.Result == "failed" && (!validFailureStage(record.Action, record.FailureStage) || record.BackupID != "none" && !backupIDPattern.MatchString(record.BackupID)) {
		return false
	}
	if record.Result == "succeeded" && (len(record.ImageDigests) != 6 || record.Action == "install" && record.BackupID != "none" || record.Action != "install" && !backupIDPattern.MatchString(record.BackupID)) {
		return false
	}
	if (record.PreviousVersion == "") != (record.PreviousManifestSHA256 == "") {
		return false
	}
	if record.PreviousVersion != "" {
		if !identifier.MatchString(record.PreviousVersion) || !hashPattern.MatchString(record.PreviousManifestSHA256) {
			return false
		}
	}
	if record.Action == "install" {
		if record.PreviousVersion != "" || record.PreviousManifestSHA256 != "" || !record.RollbackUntil.IsZero() || record.BackupID != "none" {
			return false
		}
	} else if record.Result == "succeeded" && record.Action == "release" {
		if record.PreviousVersion == "" || record.RollbackUntil.Before(record.CompletedAt) || record.RollbackUntil.After(record.CompletedAt.Add(RollbackWindow)) {
			return false
		}
	} else if record.Result == "failed" && record.Action == "release" && !record.RollbackUntil.IsZero() {
		if record.PreviousVersion == "" || !backupIDPattern.MatchString(record.BackupID) || record.RollbackUntil.Before(record.CompletedAt) || record.RollbackUntil.After(record.CompletedAt.Add(RollbackWindow)) {
			return false
		}
	} else if !record.RollbackUntil.IsZero() {
		return false
	}
	for _, digest := range record.ImageDigests {
		if !imageDigestPattern.MatchString(digest) {
			return false
		}
	}
	return true
}

func validFailureStage(action, stage string) bool {
	install := map[string]bool{"bundle": true, "install_authority": true, "preflight": true, "supply_chain": true,
		"startup_migration": true, "epoch_artifact_identity": true, "authenticated_smoke": true}
	release := map[string]bool{"candidate_bundle": true, "current_bundle": true, "compatibility": true, "preflight": true,
		"preupgrade_backup": true, "supply_chain": true, "bounded_drain": true, "startup_migration": true,
		"epoch_artifact_identity": true, "authenticated_smoke": true}
	rollback := map[string]bool{"failed_bundle": true, "previous_bundle": true, "rollback_authority": true, "preflight": true,
		"stop_failed_release": true, "clean_database": true, "exact_restore": true, "previous_startup": true,
		"epoch_artifact_identity": true, "authenticated_smoke": true}
	if action == "install" {
		return install[stage]
	}
	if action == "release" {
		return release[stage]
	}
	return rollback[stage]
}

func appendJSONLine(path string, value any) error {
	if path == "" {
		return ErrInvalid
	}
	encoded, err := json.Marshal(value)
	if err != nil || bytes.Contains(encoded, []byte{'\n'}) {
		return ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	directory, err := os.Lstat(filepath.Dir(path))
	if err != nil || !directory.IsDir() || directory.Mode().Perm()&0o077 != 0 {
		return ErrInvalid
	}
	if existing, err := os.Lstat(path); err == nil {
		if !existing.Mode().IsRegular() || existing.Mode().Perm() != 0o600 {
			return ErrInvalid
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if info, statErr := file.Stat(); statErr != nil || info.Mode().Perm() != 0o600 {
		_ = file.Close()
		return ErrInvalid
	}
	if _, err = fmt.Fprintf(file, "%s\n", encoded); err == nil {
		err = file.Sync()
	}
	return errors.Join(err, file.Close())
}

func acquireOperatorLock(path string) (*os.File, error) {
	if path == "" {
		return nil, ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	directory, err := os.Lstat(filepath.Dir(path))
	if err != nil || !directory.IsDir() || directory.Mode().Perm()&0o077 != 0 {
		return nil, ErrInvalid
	}
	if existing, err := os.Lstat(path); err == nil && (!existing.Mode().IsRegular() || existing.Mode().Perm() != 0o600) {
		return nil, ErrInvalid
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, errors.Join(ErrInvalid, err)
	}
	return file, nil
}

func decodeExact(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrInvalid
	}
	return nil
}

var (
	hashPattern        = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	imageDigestPattern = regexp.MustCompile(`^(?:[a-z0-9][a-z0-9./_-]*(?::[A-Za-z0-9._-]+)?@)?sha256:[0-9a-f]{64}$`)
	backupIDPattern    = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12}$`)
)

func containsSecretField(data []byte) bool {
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "key_value")
}
