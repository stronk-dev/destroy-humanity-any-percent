package deploymentbackup

import (
	"errors"
	"io/fs"
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
	directoryIdentity os.FileInfo
	staleIdentities   map[string]os.FileInfo
	Backups           []string `json:"backups"`
	StaleTemporaries  []string `json:"stale_temporaries"`
	ActiveTemporaries []string `json:"active_temporaries"`
	Foreign           []string `json:"foreign"`
}

func InspectTarget(directory string, now time.Time) (TargetState, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return TargetState{}, err
	}
	defer root.Close()
	identity, err := root.Stat(".")
	if err != nil {
		return TargetState{}, err
	}
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return TargetState{}, err
	}
	state := TargetState{directory: filepath.Clean(directory), directoryIdentity: identity, staleIdentities: map[string]os.FileInfo{},
		Backups: []string{}, StaleTemporaries: []string{}, ActiveTemporaries: []string{}, Foreign: []string{}}
	for _, entry := range entries {
		path := filepath.Join(directory, entry.Name())
		info, err := root.Lstat(entry.Name())
		if err != nil {
			return TargetState{}, err
		}
		regular := info.Mode().IsRegular()
		switch {
		case regular && strings.HasSuffix(entry.Name(), ".ccbackup") && !strings.HasPrefix(entry.Name(), "."):
			state.Backups = append(state.Backups, path)
		case regular && ownTemporary.MatchString(entry.Name()) && now.Sub(info.ModTime()) > StaleTemporaryAge:
			state.StaleTemporaries = append(state.StaleTemporaries, path)
			state.staleIdentities[path] = info
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
	if state.directory == "" || state.directoryIdentity == nil {
		return ErrInvalid
	}
	for _, path := range state.StaleTemporaries {
		if filepath.Dir(path) != state.directory || state.staleIdentities[path] == nil {
			return ErrInvalid
		}
	}
	ownership, err := lockTarget(state.directory, true)
	if err != nil {
		return err
	}
	defer ownership.Close()
	lockedIdentity, err := ownership.Stat()
	if err != nil || !os.SameFile(state.directoryIdentity, lockedIdentity) {
		return errors.Join(ErrInvalid, err)
	}
	root, err := os.OpenRoot(state.directory)
	if err != nil {
		return err
	}
	defer root.Close()
	rootIdentity, err := root.Stat(".")
	if err != nil || !os.SameFile(lockedIdentity, rootIdentity) {
		return errors.Join(ErrInvalid, err)
	}
	return removeStaleTemporariesInRoot(root, state, now)
}

// Keep all destructive lookups on the same directory handle as the inspected,
// locked target; a later path/alias retarget cannot redirect these removals.
func removeStaleTemporariesInRoot(root *os.Root, state TargetState, now time.Time) error {
	for _, path := range state.StaleTemporaries {
		name := filepath.Base(path)
		info, err := root.Lstat(name)
		if err != nil || state.staleIdentities[path] == nil || !os.SameFile(state.staleIdentities[path], info) ||
			!info.Mode().IsRegular() || !ownTemporary.MatchString(name) || now.Sub(info.ModTime()) <= StaleTemporaryAge {
			return errors.Join(ErrInvalid, err)
		}
		if err := root.Remove(name); err != nil {
			return err
		}
	}
	return nil
}
