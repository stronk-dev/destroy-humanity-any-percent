package deploymentbackup

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

// StaleTemporaryAge makes an unmodified own temporary eligible for crash-debris
// cleanup. Cleanup must also exclude live creators before removing it.
const StaleTemporaryAge = time.Hour

var ownTemporary = regexp.MustCompile(`^\.backup-(?:payload|envelope)-[0-9]+\.tmp$`)

// ErrTargetBusy means a supported helper still owns the target. Callers should
// defer instead of recording a failed backup or removing a live temporary.
var ErrTargetBusy = errors.New("deployment backup target in use")

// Creators hold shared directory locks through their final cleanup. An exclusive
// cleanup lock excludes all of them without adding a foreign lockfile or blocking
// concurrent creators (whose completed destinations already cannot be replaced).
func lockTarget(directory string, exclusive bool) (*os.File, error) {
	handle, err := os.Open(directory)
	if err != nil {
		return nil, err
	}
	info, err := handle.Stat()
	if err != nil || !info.IsDir() {
		handle.Close()
		return nil, errors.Join(ErrInvalid, err)
	}
	mode := syscall.LOCK_SH
	if exclusive {
		mode = syscall.LOCK_EX
	}
	if err := syscall.Flock(int(handle.Fd()), mode|syscall.LOCK_NB); err != nil {
		handle.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrTargetBusy
		}
		return nil, err
	}
	return handle, nil
}

// TargetState classifies every entry of a backup target. Nothing is excluded:
// an entry is a backup, one of Create's own temporaries, or foreign.
type TargetState struct {
	directory         string
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
	state := TargetState{directory: filepath.Clean(directory), Backups: []string{}, StaleTemporaries: []string{}, ActiveTemporaries: []string{}, Foreign: []string{}}
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
// already classified stale, excluding live creators and rechecking target,
// name, type and age at removal time. The inspection is not proof of writer death.
func RemoveStaleTemporaries(state TargetState, now time.Time) error {
	if len(state.StaleTemporaries) == 0 {
		return nil
	}
	if state.directory == "" {
		return ErrInvalid
	}
	for _, path := range state.StaleTemporaries {
		if filepath.Dir(path) != state.directory {
			return ErrInvalid
		}
	}
	ownership, err := lockTarget(state.directory, true)
	if err != nil {
		return err
	}
	defer ownership.Close()
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
