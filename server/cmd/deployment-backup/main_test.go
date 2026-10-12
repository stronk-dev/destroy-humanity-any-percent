package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"

	"cloud-clicker/server/deploymentbackup"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

func TestRecoveryIdentityRequiresDatabaseSecretPath(t *testing.T) {
	if err := runRecoveryIdentity(context.Background(), nil); err == nil {
		t.Fatal("recovery identity accepted without database URL secret")
	}
	command, class := boundedFailure("recovery-identity", errors.New("database unavailable"))
	if command != "recovery-identity" || class != "operation_failed" {
		t.Fatalf("command=%q class=%q", command, class)
	}
}

func TestBackupFailureOutputIsBounded(t *testing.T) {
	command, class := boundedFailure("recovery_code=private", errors.New("database_url=private"))
	if command != "unknown" || class != "operation_failed" {
		t.Fatalf("unbounded failure escaped: command=%q class=%q", command, class)
	}
}

func TestNonCleanRestoreRefusalHasItsOwnClass(t *testing.T) {
	if _, class := boundedFailure("restore", fmt.Errorf("restore: %w", deploymentbackup.ErrNonCleanTarget)); class != "non_clean_target" {
		t.Fatalf("non-clean refusal class=%q", class)
	}
	if _, class := boundedFailure("restore", deploymentbackup.ErrInvalid); class != "invalid_input" {
		t.Fatalf("generic invalid restore class=%q", class)
	}
}

func TestBackupPathsAdmitsEveryTargetEntryForValidation(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "partial.tmp"), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(target, "unexpected-directory"), 0o700); err != nil {
		t.Fatal(err)
	}
	paths, err := backupPaths(target)
	if err != nil || len(paths) != 2 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if err := runRetention([]string{"--target", target}); err == nil {
		t.Fatal("invalid target entries were silently excluded")
	}
}

func TestCreateFlagsRequireEveryReleaseIdentityInput(t *testing.T) {
	if _, err := parseCreateFlags("create", []string{"--target=/backups"}); err == nil {
		t.Fatal("incomplete backup command accepted")
	}
	flags, err := parseCreateFlags("create", []string{
		"--target=/backups", "--database-url-file=/run/secrets/database-url",
		"--release-manifest=/release-manifest.json", "--epoch=/epoch.json",
		"--age-recipient=age1fixture", "--server-id=server", "--metrics-dir=/operations",
	})
	if err != nil || flags.target != "/backups" {
		t.Fatalf("flags=%+v err=%v", flags, err)
	}
}

func TestScheduleRoundNeverCreatesIntoABlockedTarget(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	type result struct {
		outcome        scheduleOutcome
		creates        int
		failureWritten bool
		successWritten bool
	}
	runRound := func(t *testing.T, prepare func(target string)) result {
		t.Helper()
		target, metrics := t.TempDir(), t.TempDir()
		if err := os.Chmod(target, 0o700); err != nil {
			t.Fatal(err)
		}
		prepare(target)
		creates := 0
		round := scheduleRound{target: target, metricsDirectory: metrics, now: func() time.Time { return now },
			emit: func(any) error { return nil },
			create: func(started time.Time) (deploymentbackup.Header, string, error) {
				creates++
				return deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target, BackupID: "20260924T120000Z-abcdef123456",
					ServerID: "server-1", ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8, StartedAt: started,
					Now: func() time.Time { return now }, Recipient: identity.Recipient().String(), Dump: bytes.NewReader([]byte("PGDMP\x01schedule"))})
			}}
		outcome, _ := round.run()
		_, failureErr := os.Stat(filepath.Join(metrics, "backup-failure.prom"))
		_, successErr := os.Stat(filepath.Join(metrics, "backup-success.prom"))
		return result{outcome: outcome, creates: creates, failureWritten: failureErr == nil, successWritten: successErr == nil}
	}
	write := func(t *testing.T, path string, modified time.Time) {
		t.Helper()
		if err := os.WriteFile(path, []byte("partial"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, modified, modified); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("foreign lost+found blocks without creating", func(t *testing.T) {
		got := runRound(t, func(target string) {
			if err := os.Mkdir(filepath.Join(target, "lost+found"), 0o700); err != nil {
				t.Fatal(err)
			}
		})
		if got.outcome != scheduleBlocked || got.creates != 0 || !got.failureWritten || got.successWritten {
			t.Fatalf("foreign entry round=%+v", got)
		}
	})
	t.Run("corrupt backup blocks without creating", func(t *testing.T) {
		got := runRound(t, func(target string) {
			write(t, filepath.Join(target, "20260923T120000Z-abcdef123456.ccbackup"), now.Add(-time.Hour))
		})
		if got.outcome != scheduleBlocked || got.creates != 0 || !got.failureWritten {
			t.Fatalf("corrupt backup round=%+v", got)
		}
	})
	t.Run("stale crash temporary is removed then the round creates", func(t *testing.T) {
		var stale string
		got := runRound(t, func(target string) {
			stale = filepath.Join(target, ".backup-envelope-4242.tmp")
			write(t, stale, now.Add(-2*time.Hour))
		})
		if got.outcome != scheduleCreated || got.creates != 1 || got.failureWritten || !got.successWritten {
			t.Fatalf("stale temporary round=%+v", got)
		}
		if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("stale crash temporary retained: %v", err)
		}
	})
	t.Run("active temporary defers without failure", func(t *testing.T) {
		got := runRound(t, func(target string) {
			write(t, filepath.Join(target, ".backup-payload-4242.tmp"), now.Add(-time.Minute))
		})
		if got.outcome != scheduleDeferred || got.creates != 0 || got.failureWritten || got.successWritten {
			t.Fatalf("active temporary round=%+v", got)
		}
	})
}

func TestScheduleFailureAfterSameSecondSuccessIsObservable(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	target, metrics := t.TempDir(), t.TempDir()
	now := time.Date(2026, 10, 13, 12, 0, 0, 100_000_000, time.UTC)
	creates := 0
	round := scheduleRound{target: target, metricsDirectory: metrics, now: func() time.Time { return now },
		emit: func(any) error { return nil },
		create: func(started time.Time) (deploymentbackup.Header, string, error) {
			creates++
			return deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target,
				BackupID: fmt.Sprintf("20261013T120000Z-%012x", creates), ServerID: "server-1",
				ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8, StartedAt: started,
				Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
				Dump: bytes.NewReader([]byte("PGDMP\x01schedule"))})
		}}
	readGauge := func(file, name string) float64 {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(metrics, file))
		if err != nil {
			t.Fatal(err)
		}
		parser := expfmt.NewTextParser(model.UTF8Validation)
		families, err := parser.TextToMetricFamilies(strings.NewReader(string(data)))
		metric := families[name]
		if err != nil || metric == nil || len(metric.Metric) != 1 || metric.Metric[0].Gauge == nil {
			t.Fatalf("unreadable production metric: %s error=%v", data, err)
		}
		return metric.Metric[0].Gauge.GetValue()
	}
	if outcome, err := round.run(); err != nil || outcome != scheduleCreated || creates != 1 {
		t.Fatalf("healthy encrypted baseline: outcome=%s creates=%d error=%v", outcome, creates, err)
	}
	successAt := readGauge("backup-success.prom", "cloud_clicker_backup_last_success_timestamp_seconds")
	foreign := filepath.Join(target, "foreign-entry")
	if err := os.WriteFile(foreign, []byte("must not be deleted by the worker"), 0o600); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Millisecond)
	if outcome, err := round.run(); err != nil || outcome != scheduleBlocked || creates != 1 {
		t.Fatalf("blocked target: outcome=%s creates=%d error=%v", outcome, creates, err)
	}
	failureAt := readGauge("backup-failure.prom", "cloud_clicker_backup_last_failure_timestamp_seconds")
	if successAt != failureAt {
		t.Fatal("invalid same-second regression population")
	}
	if got := readGauge("backup-result.prom", "cloud_clicker_backup_last_result"); got != 0 {
		t.Fatalf("actual blocked scheduled backup was hidden: latest result=%v", got)
	}
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	if outcome, err := round.run(); err != nil || outcome != scheduleCreated || creates != 2 {
		t.Fatalf("recovered target: outcome=%s creates=%d error=%v", outcome, creates, err)
	}
	if got := readGauge("backup-result.prom", "cloud_clicker_backup_last_result"); got != 1 {
		t.Fatalf("successful retry did not clear the failed outcome: %v", got)
	}
	// No real Postgres dump is used here; this exercises the actual scheduler,
	// encrypted target admission and its production textfile producer.
}
