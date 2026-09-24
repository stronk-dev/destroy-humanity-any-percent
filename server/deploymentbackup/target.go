package deploymentbackup

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// StaleTemporaryAge is how long Create's own temporary file may go unmodified
// before it is treated as crash debris rather than an in-progress backup.
const StaleTemporaryAge = time.Hour

var ownTemporary = regexp.MustCompile(`^\.backup-(?:payload|envelope)-[0-9]+\.tmp$`)

// TargetState classifies every entry of a backup target. Nothing is excluded:
// an entry is a backup, one of Create's own temporaries, or foreign.
type TargetState struct {
	Backups           []string `json:"backups"`
	StaleTemporaries  []string `json:"stale_temporaries"`
	ActiveTemporaries []string `json:"active_temporaries"`
	Foreign           []string `json:"foreign"`
}

func InspectTarget(directory string, now time.Time) (TargetState, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return TargetState{}, err
	}
	state := TargetState{Backups: []string{}, StaleTemporaries: []string{}, ActiveTemporaries: []string{}, Foreign: []string{}}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return TargetState{}, err
		}
		regular := info.Mode().IsRegular()
		switch {
		case regular && strings.HasSuffix(entry.Name(), ".ccbackup") && !strings.HasPrefix(entry.Name(), "."):
			state.Backups = append(state.Backups, path)
		case regular && ownTemporary.MatchString(entry.Name()) && now.Sub(info.ModTime()) > StaleTemporaryAge:
			state.StaleTemporaries = append(state.StaleTemporaries, path)
		case regular && ownTemporary.MatchString(entry.Name()):
			state.ActiveTemporaries = append(state.ActiveTemporaries, path)
		default:
			state.Foreign = append(state.Foreign, path)
		}
	}
	for _, list := range [][]string{state.Backups, state.StaleTemporaries, state.ActiveTemporaries, state.Foreign} {
		sort.Strings(list)
	}
	return state, nil
}

// RemoveStaleTemporaries deletes only Create's own temporaries that were
// already classified stale, rechecking name, type and age at removal time.
func RemoveStaleTemporaries(state TargetState, now time.Time) error {
	for _, path := range state.StaleTemporaries {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || !ownTemporary.MatchString(filepath.Base(path)) || now.Sub(info.ModTime()) <= StaleTemporaryAge {
			return errors.Join(ErrInvalid, err)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}
