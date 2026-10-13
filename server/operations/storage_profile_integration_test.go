package operations

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func runStoragePressureProfile(t *testing.T, ctx context.Context, client *http.Client, received <-chan []byte) {
	t.Helper()
	paths := []string{"/pressure-persistent", "/pressure-backup"}
	var states [2]syscall.Statfs_t
	for index, path := range paths {
		if err := syscall.Statfs(path, &states[index]); err != nil ||
			states[index].Blocks*uint64(states[index].Bsize) != 20<<20 || states[index].Bavail != states[index].Blocks {
			t.Fatal("storage proof requires its empty, declared 20MiB disposable filesystems")
		}
	}
	if states[0].Fsid == states[1].Fsid {
		t.Fatal("persistent and backup subjects must be separate filesystems")
	}
	observe := func(subject *testing.T) [2]float64 {
		subject.Helper()
		var usage [2]float64
		for index, path := range paths {
			var err error
			usage[index], err = FilesystemUsage(path)
			if err != nil {
				subject.Fatal(err)
			}
		}
		// Only the filesystem producer is under test. Restart/journal values
		// are controls, not actual Docker or journal observations.
		if err := RecordHost("/textfile", usage[0], usage[1], 0, 0, 1000); err != nil {
			subject.Fatal(err)
		}
		return usage
	}
	if usage := observe(t); usage != [2]float64{} {
		t.Fatal("empty subjects did not produce zero usage")
	}
	for _, storage := range []string{"persistent", "backup"} {
		query := fmt.Sprintf(`cloud_clicker_filesystem_usage_ratio{storage=%q}`, storage)
		if !waitForQueryValue(ctx, client, query, "0") {
			t.Fatal("actual healthy host textfile never reached Prometheus")
		}
	}
	if !waitForAlertState(ctx, client, "CloudClickerStoragePressure", "inactive") {
		t.Fatal("storage alert had no healthy baseline")
	}
	for index, storage := range []string{"persistent", "backup"} {
		t.Run(storage, func(t *testing.T) {
			path := filepath.Join(paths[index], "operations-pressure-fixture")
			file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = file.Close()
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					t.Error(err)
				}
			})
			block := make([]byte, 1<<20)
			writeMiB := func(count int) {
				for range count {
					if written, err := file.Write(block); err != nil || written != len(block) {
						t.Fatal("bounded pressure allocation incomplete")
					}
				}
				if err := file.Sync(); err != nil {
					t.Fatal(err)
				}
			}
			query := fmt.Sprintf(`cloud_clicker_filesystem_usage_ratio{storage=%q}`, storage)
			writeMiB(16)
			if usage := observe(t); usage[index] != 0.8 || usage[1-index] != 0 {
				t.Fatal("actual allocation did not reach the exact 80% boundary")
			}
			if !waitForQueryValue(ctx, client, query, "0.8") {
				t.Fatal("exact-boundary observation never reached Prometheus")
			}
			requireNoOperationsAlert(t, ctx, client, `ALERTS{alertname="CloudClickerStoragePressure"}`, 30*time.Second)
			failedAt := time.Now().UTC()
			writeMiB(1)
			if usage := observe(t); usage[index] != 0.85 || usage[1-index] != 0 {
				t.Fatal("actual bounded allocation did not cross 80% on only its subject")
			}
			if !waitForQueryValue(ctx, client, query, "0.85") ||
				!waitForAlertState(ctx, client, "CloudClickerStoragePressure", "firing") {
				t.Fatal("actual filesystem pressure did not fire through node-exporter")
			}
			firingStart := waitForOperationsNotification(t, ctx, received, "CloudClickerStoragePressure", "node-exporter", "firing", failedAt, time.Time{}, map[string]string{"storage": storage})
			t.Log("actual 85% filesystem allocation reached the configured receiver")
			clearedAt := time.Now().UTC()
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if usage := observe(t); usage != [2]float64{} {
				t.Fatal("removing only the fixture allocation did not recover filesystem usage")
			}
			if !waitForQueryValue(ctx, client, query, "0") ||
				!waitForAlertState(ctx, client, "CloudClickerStoragePressure", "inactive") {
				t.Fatal("actual filesystem recovery did not resolve its rule")
			}
			waitForOperationsNotification(t, ctx, received, "CloudClickerStoragePressure", "node-exporter", "resolved", clearedAt, firingStart, map[string]string{"storage": storage})
			t.Log("that exact storage alert resolved after freeing the fixture's file")
		})
		if t.Failed() {
			return
		}
	}
}
