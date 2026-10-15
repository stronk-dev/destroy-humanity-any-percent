package operations

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// CleanupFailureFromJournal consumes the complete filtered journal population.
// A cursor hash identifies an observation across gameserver/observer restarts,
// without retaining the journal message or container/player identifiers.
func CleanupFailureFromJournal(input io.Reader) (string, error) {
	if input == nil {
		return "", ErrInvalid
	}
	decoder := json.NewDecoder(input)
	latest := ""
	for {
		var row struct {
			Tag       string          `json:"CONTAINER_TAG"`
			Cursor    string          `json:"__CURSOR"`
			Timestamp string          `json:"__REALTIME_TIMESTAMP"`
			Partial   json.RawMessage `json:"CONTAINER_PARTIAL_MESSAGE"`
			Message   string          `json:"MESSAGE"`
		}
		if err := decoder.Decode(&row); errors.Is(err, io.EOF) {
			return latest, nil
		} else if err != nil {
			return "", errors.Join(ErrInvalid, err)
		}
		stamp, err := strconv.ParseUint(row.Timestamp, 10, 64)
		if row.Tag != "cloud-clicker-gameserver" || row.Cursor == "" || stamp == 0 || err != nil || len(row.Partial) != 0 {
			return "", ErrInvalid
		}
		var message struct {
			Class string `json:"event_class"`
			Job   string `json:"job_name"`
		}
		if json.Unmarshal([]byte(row.Message), &message) != nil {
			return "", ErrInvalid
		}
		if message.Class != "composed_job_failure" {
			continue
		}
		if _, known := jobNames[message.Job]; !known {
			return "", ErrInvalid
		}
		if message.Job == "credential_cleanup" || message.Job == "route_name_expiry" {
			latest = fmt.Sprintf("%x", sha256.Sum256([]byte(row.Cursor)))
		}
	}
}

type cleanupJournalState struct {
	SchemaVersion int    `json:"schema_version"`
	CursorHash    string `json:"cursor_hash"`
	DetectedAt    int64  `json:"detected_at"`
}

var cleanupJournalWrite sync.Mutex

// RecordCleanupJournal holds a newly discovered failure for the existing five-
// minute alert window, even when the original process is gone or the observer
// resumed late. Re-reading the same journal event never extends that window.
func RecordCleanupJournal(directory, cursorHash string, at time.Time) error {
	// The declared systemd oneshot has one collector instance. Serialize library
	// callers within that instance too, including observation fixtures.
	cleanupJournalWrite.Lock()
	defer cleanupJournalWrite.Unlock()
	if !filepath.IsAbs(directory) || at.IsZero() || at.Unix() <= 0 || (cursorHash != "" && !validCursorHash(cursorHash)) {
		return ErrInvalid
	}
	state := cleanupJournalState{SchemaVersion: 1}
	statePath := filepath.Join(directory, "cleanup-state.json")
	if info, err := os.Lstat(statePath); err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		data, err := os.ReadFile(statePath)
		if err != nil {
			return err
		}
		// Do not inherit the new-file default: a present JSON object without
		// a version is corrupt state, not a valid version1 empty observation.
		state = cleanupJournalState{}
		if json.Unmarshal(data, &state) != nil || state.SchemaVersion != 1 ||
			(state.CursorHash != "" && (!validCursorHash(state.CursorHash) || state.DetectedAt <= 0)) ||
			(state.CursorHash == "" && state.DetectedAt != 0) {
			return ErrInvalid
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	changed := cursorHash != "" && cursorHash != state.CursorHash
	if changed {
		state.CursorHash, state.DetectedAt = cursorHash, at.UTC().Unix()
	}
	data := []byte(fmt.Sprintf(`# HELP cloud_clicker_cleanup_failure_detected_timestamp_seconds Host detection time of the latest journal cleanup failure.
# TYPE cloud_clicker_cleanup_failure_detected_timestamp_seconds gauge
cloud_clicker_cleanup_failure_detected_timestamp_seconds %d
# HELP cloud_clicker_cleanup_journal_observed_timestamp_seconds Last complete host cleanup-journal collection.
# TYPE cloud_clicker_cleanup_journal_observed_timestamp_seconds gauge
cloud_clicker_cleanup_journal_observed_timestamp_seconds %d
`, state.DetectedAt, at.UTC().Unix()))
	// Publish before committing the seen cursor. A failed metric write cannot
	// consume a failure; a failed state write causes safe re-observation.
	if err := writeMetricFile(directory, "cleanup.prom", data); err != nil {
		return err
	}
	if changed {
		data, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return writeMetricFile(directory, "cleanup-state.json", append(data, '\n'))
	}
	return nil
}

func validCursorHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}
