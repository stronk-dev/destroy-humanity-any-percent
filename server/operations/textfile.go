package operations

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var operationNames = map[string]struct{}{
	"backup": {}, "release": {}, "restore": {},
}

var operationResults = map[string]struct{}{
	"success": {}, "failure": {},
}

func RecordOperation(directory, operation, result string, at time.Time) error {
	if !filepath.IsAbs(directory) || strings.ContainsAny(directory, "\r\n") || at.IsZero() {
		return ErrInvalid
	}
	if _, ok := operationNames[operation]; !ok {
		return ErrInvalid
	}
	if _, ok := operationResults[result]; !ok {
		return ErrInvalid
	}
	name := operation + "_last_" + result + "_timestamp_seconds"
	data := []byte(fmt.Sprintf("# HELP cloud_clicker_%s Unix timestamp of the latest %s %s.\n# TYPE cloud_clicker_%s gauge\ncloud_clicker_%s %d\n",
		name, operation, result, name, name, at.UTC().Unix()))
	return writeMetricFile(directory, operation+"-"+result+".prom", data)
}

func RecordHost(directory string, persistentUsage, backupUsage float64, restartCount, journalBytes, journalBudget uint64) error {
	if !filepath.IsAbs(directory) || strings.ContainsAny(directory, "\r\n") || persistentUsage < 0 || persistentUsage > 1 || backupUsage < 0 || backupUsage > 1 || journalBudget < 1 {
		return ErrInvalid
	}
	data := []byte(fmt.Sprintf(`# HELP cloud_clicker_filesystem_usage_ratio Used fraction of the bounded storage subject.
# TYPE cloud_clicker_filesystem_usage_ratio gauge
cloud_clicker_filesystem_usage_ratio{storage="backup"} %.9g
cloud_clicker_filesystem_usage_ratio{storage="persistent"} %.9g
# HELP cloud_clicker_gameserver_restart_count Docker restart count for the current gameserver container.
# TYPE cloud_clicker_gameserver_restart_count counter
cloud_clicker_gameserver_restart_count %d
# HELP cloud_clicker_journal_usage_bytes Current persistent journal bytes for the Cloud Clicker workload.
# TYPE cloud_clicker_journal_usage_bytes gauge
cloud_clicker_journal_usage_bytes %d
# HELP cloud_clicker_journal_budget_bytes Measured reserved journal budget.
# TYPE cloud_clicker_journal_budget_bytes gauge
cloud_clicker_journal_budget_bytes %d
`, backupUsage, persistentUsage, restartCount, journalBytes, journalBudget))
	return writeMetricFile(directory, "host.prom", data)
}

func writeMetricFile(directory, filename string, data []byte) error {
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || filepath.Base(filename) != filename || len(data) == 0 {
		return errors.Join(ErrInvalid, err)
	}
	final := filepath.Join(directory, filename)
	if existing, statErr := os.Lstat(final); statErr == nil && (!existing.Mode().IsRegular() || existing.Mode()&os.ModeSymlink != 0) {
		return ErrInvalid
	} else if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return errors.Join(ErrInvalid, statErr)
	}
	temporary, err := os.CreateTemp(directory, "."+strings.TrimSuffix(filename, filepath.Ext(filename))+"-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(0o644); err != nil {
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, final); err != nil {
		return err
	}
	committed = true
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer directoryHandle.Close()
	return directoryHandle.Sync()
}
