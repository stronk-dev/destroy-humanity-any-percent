// Package deploymentbackup owns the encrypted, atomic Phase-0 backup envelope.
package deploymentbackup

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
)

const magic = "CLOUD-CLICKER-BACKUP-V1\n"

var (
	ErrInvalid = errors.New("invalid deployment backup")
	backupID   = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12}$`)
	hash       = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type Header struct {
	SchemaVersion         int       `json:"schema_version"`
	BackupID              string    `json:"backup_id"`
	ServerID              string    `json:"server_id"`
	ReleaseManifestSHA256 string    `json:"release_manifest_sha256"`
	EpochID               int64     `json:"epoch_id"`
	StartedAt             time.Time `json:"started_at"`
	CompletedAt           time.Time `json:"completed_at"`
	PayloadSHA256         string    `json:"payload_sha256"`
	PayloadBytes          int64     `json:"payload_bytes"`
	PreUpgrade            bool      `json:"pre_upgrade"`
	UpgradeResolved       bool      `json:"upgrade_resolved"`
}

type CreateInput struct {
	Directory             string
	BackupID              string
	ServerID              string
	ReleaseManifestSHA256 string
	EpochID               int64
	StartedAt             time.Time
	Now                   func() time.Time
	Recipient             string
	Dump                  io.Reader
	PreUpgrade            bool
}

type authenticatedHeader struct {
	BackupID              string    `json:"backup_id"`
	ServerID              string    `json:"server_id"`
	ReleaseManifestSHA256 string    `json:"release_manifest_sha256"`
	EpochID               int64     `json:"epoch_id"`
	StartedAt             time.Time `json:"started_at"`
	CompletedAt           time.Time `json:"completed_at"`
	PreUpgrade            bool      `json:"pre_upgrade"`
	UpgradeResolved       bool      `json:"upgrade_resolved"`
}

func Create(input CreateInput) (Header, string, error) {
	if input.Now == nil || input.Dump == nil || !backupID.MatchString(input.BackupID) || input.ServerID == "" ||
		!hash.MatchString(input.ReleaseManifestSHA256) || input.EpochID < 1 || input.StartedAt.IsZero() {
		return Header{}, "", ErrInvalid
	}
	recipient, err := age.ParseX25519Recipient(input.Recipient)
	if err != nil {
		return Header{}, "", ErrInvalid
	}
	info, err := os.Stat(input.Directory)
	if err != nil || !info.IsDir() || info.Mode()&0o022 != 0 {
		return Header{}, "", ErrInvalid
	}
	header := Header{SchemaVersion: 1, BackupID: input.BackupID, ServerID: input.ServerID,
		ReleaseManifestSHA256: input.ReleaseManifestSHA256, EpochID: input.EpochID,
		StartedAt: input.StartedAt.UTC(), CompletedAt: input.Now().UTC(), PreUpgrade: input.PreUpgrade}
	if header.CompletedAt.Before(header.StartedAt) {
		return Header{}, "", ErrInvalid
	}
	payload, err := os.CreateTemp(input.Directory, ".backup-payload-*.tmp")
	if err != nil {
		return Header{}, "", err
	}
	payloadPath := payload.Name()
	defer os.Remove(payloadPath)
	encrypted, err := age.Encrypt(payload, recipient)
	if err == nil {
		authenticated, _ := json.Marshal(authenticatedFrom(header))
		_, err = encrypted.Write(append(authenticated, '\n'))
	}
	if err == nil {
		_, err = io.Copy(encrypted, input.Dump)
	}
	if encrypted != nil {
		if closeErr := encrypted.Close(); err == nil {
			err = closeErr
		}
	}
	if syncErr := payload.Sync(); err == nil {
		err = syncErr
	}
	if closeErr := payload.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Header{}, "", err
	}
	payload, err = os.Open(payloadPath)
	if err != nil {
		return Header{}, "", err
	}
	defer payload.Close()
	hasher := sha256.New()
	payloadSize, err := io.Copy(hasher, payload)
	if err != nil {
		return Header{}, "", err
	}
	if _, err := payload.Seek(0, io.SeekStart); err != nil {
		return Header{}, "", err
	}
	header.PayloadSHA256, header.PayloadBytes = "sha256:"+hex.EncodeToString(hasher.Sum(nil)), payloadSize
	if err := validateHeader(header); err != nil {
		return Header{}, "", ErrInvalid
	}
	headerBytes, _ := json.Marshal(header)
	finalPath := filepath.Join(input.Directory, input.BackupID+".ccbackup")
	if _, err := os.Lstat(finalPath); !errors.Is(err, os.ErrNotExist) {
		return Header{}, "", ErrInvalid
	}
	temporary, err := os.CreateTemp(input.Directory, ".backup-envelope-*.tmp")
	if err != nil {
		return Header{}, "", err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err == nil {
		_, err = temporary.Write(append(append([]byte(magic), headerBytes...), '\n'))
	}
	if err == nil {
		_, err = io.Copy(temporary, payload)
	}
	if err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Header{}, "", err
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		return Header{}, "", err
	}
	return header, finalPath, nil
}

func ReadHeader(path string) (Header, error) {
	file, err := os.Open(path)
	if err != nil {
		return Header{}, err
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64<<10)
	first, err := reader.ReadString('\n')
	if err != nil || first != magic {
		return Header{}, ErrInvalid
	}
	line, err := reader.ReadBytes('\n')
	if err != nil || len(line) > 64<<10 {
		return Header{}, ErrInvalid
	}
	var header Header
	decoder := json.NewDecoder(strings.NewReader(string(line)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&header) != nil || validateHeader(header) != nil {
		return Header{}, ErrInvalid
	}
	hasher := sha256.New()
	written, err := io.Copy(hasher, reader)
	if err != nil || written != header.PayloadBytes || "sha256:"+hex.EncodeToString(hasher.Sum(nil)) != header.PayloadSHA256 {
		return Header{}, ErrInvalid
	}
	return header, nil
}

func Restore(path, expectedManifest string, identity age.Identity, output io.Writer) (Header, error) {
	if identity == nil || output == nil || !hash.MatchString(expectedManifest) {
		return Header{}, fmt.Errorf("%w: restore inputs", ErrInvalid)
	}
	file, err := os.Open(path)
	if err != nil {
		return Header{}, err
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64<<10)
	first, _ := reader.ReadString('\n')
	line, lineErr := reader.ReadBytes('\n')
	var header Header
	decoder := json.NewDecoder(strings.NewReader(string(line)))
	decoder.DisallowUnknownFields()
	if first != magic || lineErr != nil || decoder.Decode(&header) != nil || validateHeader(header) != nil || header.ReleaseManifestSHA256 != expectedManifest {
		return Header{}, fmt.Errorf("%w: envelope header", ErrInvalid)
	}
	payload, err := os.CreateTemp("", ".cloud-clicker-restore-*.age")
	if err != nil {
		return Header{}, err
	}
	payloadPath := payload.Name()
	defer os.Remove(payloadPath)
	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(payload, hasher), reader)
	if closeErr := payload.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil || written != header.PayloadBytes || "sha256:"+hex.EncodeToString(hasher.Sum(nil)) != header.PayloadSHA256 {
		return Header{}, fmt.Errorf("%w: encrypted payload checksum", ErrInvalid)
	}
	payload, err = os.Open(payloadPath)
	if err != nil {
		return Header{}, err
	}
	defer payload.Close()
	decrypted, err := age.Decrypt(payload, identity)
	if err != nil {
		return Header{}, fmt.Errorf("%w: age decryption", ErrInvalid)
	}
	plain := bufio.NewReaderSize(decrypted, 64<<10)
	authenticatedLine, err := plain.ReadBytes('\n')
	var authenticated authenticatedHeader
	decoder = json.NewDecoder(strings.NewReader(string(authenticatedLine)))
	decoder.DisallowUnknownFields()
	if err != nil {
		return Header{}, fmt.Errorf("%w: authenticated metadata framing: %v", ErrInvalid, err)
	}
	if decodeErr := decoder.Decode(&authenticated); decodeErr != nil {
		return Header{}, fmt.Errorf("%w: authenticated metadata encoding: %v", ErrInvalid, decodeErr)
	}
	if differences := authenticatedDifferences(authenticated, authenticatedFrom(header)); len(differences) != 0 {
		return Header{}, fmt.Errorf("%w: authenticated metadata fields %s", ErrInvalid, strings.Join(differences, ","))
	}
	if _, err := io.Copy(output, plain); err != nil {
		return Header{}, err
	}
	return header, nil
}

type Retention struct {
	Keep    []string
	Delete  []string
	Invalid []string
}

type ScheduleStatus struct {
	LatestCompletedAt time.Time `json:"latest_completed_at"`
	Missing           bool      `json:"missing"`
	Late              bool      `json:"late"`
	Invalid           []string  `json:"invalid"`
}

func PlanRetention(paths []string, now time.Time) Retention {
	type item struct {
		path   string
		header Header
	}
	valid := []item{}
	result := Retention{}
	for _, path := range paths {
		header, err := ReadHeader(path)
		if err != nil || header.CompletedAt.After(now) {
			result.Invalid = append(result.Invalid, path)
			continue
		}
		valid = append(valid, item{path, header})
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].header.CompletedAt.After(valid[j].header.CompletedAt) })
	daily := map[string]bool{}
	for index, entry := range valid {
		age := now.Sub(entry.header.CompletedAt)
		day := entry.header.CompletedAt.UTC().Format("2006-01-02")
		keep := index == 0 || age <= 7*24*time.Hour || entry.header.PreUpgrade && !entry.header.UpgradeResolved
		if !keep && age <= 30*24*time.Hour && !daily[day] {
			keep, daily[day] = true, true
		}
		if keep {
			result.Keep = append(result.Keep, entry.path)
		} else {
			result.Delete = append(result.Delete, entry.path)
		}
	}
	sort.Strings(result.Keep)
	sort.Strings(result.Delete)
	sort.Strings(result.Invalid)
	return result
}

func EvaluateSchedule(paths []string, now time.Time) ScheduleStatus {
	status := ScheduleStatus{}
	for _, path := range paths {
		header, err := ReadHeader(path)
		if err != nil || header.CompletedAt.After(now) {
			status.Invalid = append(status.Invalid, path)
			continue
		}
		if status.LatestCompletedAt.IsZero() || header.CompletedAt.After(status.LatestCompletedAt) {
			status.LatestCompletedAt = header.CompletedAt
		}
	}
	sort.Strings(status.Invalid)
	status.Missing = status.LatestCompletedAt.IsZero()
	status.Late = !status.Missing && now.Sub(status.LatestCompletedAt) > 6*time.Hour
	return status
}

// ApplyRetention derives and executes the governed policy from the complete
// population. Callers cannot submit a hand-written deletion plan that omits a
// newest or unresolved pre-upgrade backup.
func ApplyRetention(paths []string, now time.Time) (Retention, error) {
	plan := PlanRetention(paths, now)
	if len(plan.Invalid) != 0 {
		return plan, ErrInvalid
	}
	protected := make(map[string]bool, len(plan.Keep))
	for _, path := range plan.Keep {
		protected[path] = true
		if _, err := ReadHeader(path); err != nil {
			return plan, ErrInvalid
		}
	}
	for _, path := range plan.Delete {
		if protected[path] {
			return plan, ErrInvalid
		}
		if _, err := ReadHeader(path); err != nil {
			return plan, ErrInvalid
		}
	}
	for _, path := range plan.Delete {
		if err := os.Remove(path); err != nil {
			return plan, err
		}
	}
	return plan, nil
}

func validateHeader(header Header) error {
	if header.SchemaVersion != 1 || !backupID.MatchString(header.BackupID) || header.ServerID == "" || !hash.MatchString(header.ReleaseManifestSHA256) ||
		header.EpochID < 1 || header.StartedAt.IsZero() || header.CompletedAt.IsZero() || !hash.MatchString(header.PayloadSHA256) || header.PayloadBytes < 1 || header.UpgradeResolved && !header.PreUpgrade {
		return ErrInvalid
	}
	return nil
}

func authenticatedFrom(header Header) authenticatedHeader {
	return authenticatedHeader{BackupID: header.BackupID, ServerID: header.ServerID,
		ReleaseManifestSHA256: header.ReleaseManifestSHA256, EpochID: header.EpochID,
		StartedAt: header.StartedAt, CompletedAt: header.CompletedAt, PreUpgrade: header.PreUpgrade,
		UpgradeResolved: header.UpgradeResolved}
}

func authenticatedEqual(left, right authenticatedHeader) bool {
	return left.BackupID == right.BackupID && left.ServerID == right.ServerID &&
		left.ReleaseManifestSHA256 == right.ReleaseManifestSHA256 && left.EpochID == right.EpochID &&
		left.StartedAt.Equal(right.StartedAt) && left.CompletedAt.Equal(right.CompletedAt) &&
		left.PreUpgrade == right.PreUpgrade && left.UpgradeResolved == right.UpgradeResolved
}

func authenticatedDifferences(left, right authenticatedHeader) []string {
	differences := []string{}
	if left.BackupID != right.BackupID {
		differences = append(differences, "backup_id")
	}
	if left.ServerID != right.ServerID {
		differences = append(differences, "server_id")
	}
	if left.ReleaseManifestSHA256 != right.ReleaseManifestSHA256 {
		differences = append(differences, "release_manifest_sha256")
	}
	if left.EpochID != right.EpochID {
		differences = append(differences, "epoch_id")
	}
	if !left.StartedAt.Equal(right.StartedAt) {
		differences = append(differences, "started_at")
	}
	if !left.CompletedAt.Equal(right.CompletedAt) {
		differences = append(differences, "completed_at")
	}
	if left.PreUpgrade != right.PreUpgrade {
		differences = append(differences, "pre_upgrade")
	}
	if left.UpgradeResolved != right.UpgradeResolved {
		differences = append(differences, "upgrade_resolved")
	}
	return differences
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (header Header) String() string {
	return fmt.Sprintf("%s %s", header.BackupID, header.CompletedAt.Format(time.RFC3339))
}
