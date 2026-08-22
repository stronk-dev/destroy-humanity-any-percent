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
	raw, err := os.CreateTemp(input.Directory, ".backup-dump-*.tmp")
	if err != nil {
		return Header{}, "", err
	}
	rawPath := raw.Name()
	defer os.Remove(rawPath)
	if _, err = io.Copy(raw, input.Dump); err == nil {
		err = raw.Sync()
	}
	if closeErr := raw.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Header{}, "", err
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
		raw, err = os.Open(rawPath)
	}
	if err == nil {
		_, err = io.Copy(encrypted, raw)
		_ = raw.Close()
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
		return Header{}, ErrInvalid
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
		return Header{}, ErrInvalid
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
		return Header{}, ErrInvalid
	}
	payload, err = os.Open(payloadPath)
	if err != nil {
		return Header{}, err
	}
	defer payload.Close()
	decrypted, err := age.Decrypt(payload, identity)
	if err != nil {
		return Header{}, ErrInvalid
	}
	plain := bufio.NewReaderSize(decrypted, 64<<10)
	authenticatedLine, err := plain.ReadBytes('\n')
	var authenticated authenticatedHeader
	decoder = json.NewDecoder(strings.NewReader(string(authenticatedLine)))
	decoder.DisallowUnknownFields()
	if err != nil || decoder.Decode(&authenticated) != nil || authenticated != authenticatedFrom(header) {
		return Header{}, ErrInvalid
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

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (header Header) String() string {
	return fmt.Sprintf("%s %s", header.BackupID, header.CompletedAt.Format(time.RFC3339))
}
