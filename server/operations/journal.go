package operations

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	JournalRetention  = 14 * 24 * time.Hour
	SecurityRetention = 7 * 24 * time.Hour
)

type JournalObservation struct {
	SchemaVersion            int       `json:"schema_version"`
	Population               string    `json:"population"`
	StartedAt                time.Time `json:"started_at"`
	CompletedAt              time.Time `json:"completed_at"`
	ObjectiveCompleted       bool      `json:"objective_completed"`
	GuardExhausted           bool      `json:"guard_exhausted"`
	Samples                  int       `json:"samples"`
	ObservedBytes            uint64    `json:"observed_bytes"`
	PeakBytesPerDay          uint64    `json:"peak_bytes_per_day"`
	FilesystemBytes          uint64    `json:"filesystem_bytes"`
	JournalMaxUseBytes       uint64    `json:"journal_max_use_bytes"`
	JournalRetentionSeconds  int64     `json:"journal_retention_seconds"`
	StorageAlertFraction     float64   `json:"storage_alert_fraction"`
	RawIPEnabled             bool      `json:"raw_ip_enabled"`
	SecuritySinkSeparate     bool      `json:"security_sink_separate"`
	SecurityRetentionSeconds int64     `json:"security_retention_seconds"`
}

var populationPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{2,63}$`)

func DecodeJournalObservation(data []byte) (JournalObservation, error) {
	var observation JournalObservation
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observation) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateJournalObservation(observation) != nil {
		return JournalObservation{}, ErrInvalid
	}
	return observation, nil
}

func ValidateJournalObservation(observation JournalObservation) error {
	required := observation.PeakBytesPerDay
	if required > math.MaxUint64/14 {
		return ErrInvalid
	}
	required *= 14
	if observation.SchemaVersion != 1 || !populationPattern.MatchString(observation.Population) || observation.StartedAt.IsZero() ||
		!observation.CompletedAt.After(observation.StartedAt) || !observation.ObjectiveCompleted || observation.GuardExhausted || observation.Samples < 2 ||
		observation.ObservedBytes < 1 || observation.PeakBytesPerDay < observation.ObservedBytes || observation.FilesystemBytes < 1 ||
		observation.JournalMaxUseBytes < required || observation.JournalMaxUseBytes >= observation.FilesystemBytes ||
		observation.JournalRetentionSeconds != int64(JournalRetention/time.Second) || observation.StorageAlertFraction != 0.8 {
		return ErrInvalid
	}
	// The shipped profile has no raw-IP producer or enable switch. Refuse an
	// operator assertion in place of a real separate sink and purge witness.
	if observation.RawIPEnabled || observation.SecuritySinkSeparate || observation.SecurityRetentionSeconds != 0 {
		return ErrInvalid
	}
	// The journal-budget alert evaluates usage at 80% of the measured reserve,
	// so it necessarily fires before journald reaches its eviction boundary.
	if uint64(float64(observation.JournalMaxUseBytes)*observation.StorageAlertFraction) >= observation.JournalMaxUseBytes {
		return ErrInvalid
	}
	return nil
}

func RenderJournalPolicy(template []byte, observation JournalObservation) ([]byte, error) {
	if ValidateJournalObservation(observation) != nil || len(template) == 0 {
		return nil, ErrInvalid
	}
	result := string(template)
	values := map[string]string{
		"@@MAX_RETENTION_SEC@@": fmt.Sprint(observation.JournalRetentionSeconds),
		"@@SYSTEM_MAX_USE@@":    fmt.Sprint(observation.JournalMaxUseBytes),
	}
	for token, value := range values {
		if strings.Count(result, token) != 1 {
			return nil, ErrInvalid
		}
		result = strings.Replace(result, token, value, 1)
	}
	if strings.Contains(result, "@@") {
		return nil, ErrInvalid
	}
	return []byte(result), nil
}

func WriteJournalPolicy(path string, data []byte) error {
	if !filepath.IsAbs(path) || len(data) == 0 {
		return ErrInvalid
	}
	if info, err := os.Lstat(path); err == nil || !errors.Is(err, os.ErrNotExist) || info != nil {
		return ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
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
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	written = true
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func WriteJournalBudget(path string, budget uint64) error {
	if budget == 0 {
		return ErrInvalid
	}
	return WriteJournalPolicy(path, []byte(fmt.Sprintf("%d\n", budget)))
}
