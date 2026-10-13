package operations

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

func TestRecordHostRejectsNonfiniteUsageWithoutRefreshingObservation(t *testing.T) {
	for _, usage := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, subject := range []string{"persistent", "backup"} {
			t.Run(subject+"/"+fmtFloat(usage), func(t *testing.T) {
				directory := t.TempDir()
				if err := RecordHost(directory, 0.25, 0.5, 0, 0, 1000); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(directory, "host.prom")
				before, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				original, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				persistent, backup := 0.25, 0.5
				if subject == "persistent" {
					persistent = usage
				} else {
					backup = usage
				}
				if err := RecordHost(directory, persistent, backup, 1, 0, 1000); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nonfinite %s observation reported success: %v", subject, err)
				}
				after, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(path)
				if err != nil || string(data) != string(original) || !after.ModTime().Equal(before.ModTime()) {
					t.Fatal("invalid observation replaced or refreshed last valid metrics")
				}
			})
		}
	}
}

func TestRecordOperationPreservesSuccessAcrossLaterFailure(t *testing.T) {
	directory := t.TempDir()
	succeeded := time.Unix(1_700_000_000, 0)
	failed := succeeded.Add(time.Hour)
	if err := RecordOperation(directory, "backup", "success", succeeded); err != nil {
		t.Fatal(err)
	}
	if err := RecordOperation(directory, "backup", "failure", failed); err != nil {
		t.Fatal(err)
	}
	success, err := os.ReadFile(filepath.Join(directory, "backup-success.prom"))
	if err != nil || !strings.Contains(string(success), "cloud_clicker_backup_last_success_timestamp_seconds 1700000000") {
		t.Fatalf("success observation lost: %q err=%v", success, err)
	}
	failure, err := os.ReadFile(filepath.Join(directory, "backup-failure.prom"))
	if err != nil || !strings.Contains(string(failure), "cloud_clicker_backup_last_failure_timestamp_seconds 1700003600") {
		t.Fatalf("failure observation missing: %q err=%v", failure, err)
	}
	entries, _ := os.ReadDir(directory)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("partial metric survived: %s", entry.Name())
		}
	}
}

func TestRecordOperationOutcomeDoesNotDependOnTimestampOrder(t *testing.T) {
	for _, operation := range []string{"backup", "release", "restore"} {
		for _, delay := range []time.Duration{0, time.Millisecond, -time.Second} {
			t.Run(operation+"/"+delay.String(), func(t *testing.T) {
				directory := t.TempDir()
				at := time.Unix(1_700_000_000, 100_000_000)
				for index, result := range []string{"success", "failure", "success"} {
					observedAt := at
					if index != 0 {
						observedAt = at.Add(delay)
					}
					if err := RecordOperation(directory, operation, result, observedAt); err != nil {
						t.Fatal(err)
					}
					data, err := os.ReadFile(filepath.Join(directory, operation+"-result.prom"))
					if err != nil {
						t.Fatal(err)
					}
					parser := expfmt.NewTextParser(model.UTF8Validation)
					metrics, err := parser.TextToMetricFamilies(strings.NewReader(string(data)))
					family := metrics["cloud_clicker_"+operation+"_last_result"]
					if err != nil || len(metrics) != 1 || family == nil || len(family.Metric) != 1 ||
						len(family.Metric[0].Label) != 0 || family.Metric[0].Gauge == nil {
						t.Fatalf("invalid bounded last outcome metric: %s error=%v", data, err)
					}
					want := float64(0)
					if result == "success" {
						want = 1
					}
					if got := family.Metric[0].Gauge.GetValue(); got != want {
						t.Fatalf("latest %s outcome=%v want=%v", result, got, want)
					}
				}
			})
		}
	}
}

func TestRecordOperationRejectsUnboundedNamesAndSymlinkTarget(t *testing.T) {
	directory := t.TempDir()
	if err := RecordOperation(directory, "player-controlled", "success", time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unbounded operation accepted: %v", err)
	}
	target := filepath.Join(directory, "outside")
	if err := os.WriteFile(target, []byte("protected"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(directory, "release-success.prom")); err != nil {
		t.Fatal(err)
	}
	if err := RecordOperation(directory, "release", "success", time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("symlink metric target accepted: %v", err)
	}
	data, _ := os.ReadFile(target)
	if string(data) != "protected" {
		t.Fatalf("symlink target changed: %q", data)
	}
}

func TestRecordOperationCannotReportSuccessWithoutItsOutcomeFile(t *testing.T) {
	directory := t.TempDir()
	protected := filepath.Join(directory, "protected")
	if err := os.WriteFile(protected, []byte("untouched"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(protected, filepath.Join(directory, "backup-result.prom")); err != nil {
		t.Fatal(err)
	}
	if err := RecordOperation(directory, "backup", "success", time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("failed outcome publication reported success: %v", err)
	}
	data, err := os.ReadFile(protected)
	if err != nil || string(data) != "untouched" {
		t.Fatalf("protected target changed: %q error=%v", data, err)
	}
}

func TestRecordHostUsesOnlyBoundedStorageLabels(t *testing.T) {
	directory := t.TempDir()
	if err := RecordHost(directory, 0.75, 0.5, 3, 80, 100); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "host.prom"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`storage="backup"} 0.5`, `storage="persistent"} 0.75`, "cloud_clicker_gameserver_restart_count 3", "cloud_clicker_journal_usage_bytes 80"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing host metric %q in %s", want, data)
		}
	}
	if err := RecordHost(directory, 0, 0, 0, 101, 100); err != nil {
		t.Fatalf("over-budget journal pressure was hidden: %v", err)
	}
	data, _ = os.ReadFile(filepath.Join(directory, "host.prom"))
	if !strings.Contains(string(data), "cloud_clicker_journal_usage_bytes 101") {
		t.Fatalf("over-budget journal pressure absent: %s", data)
	}
}

func TestContainerHostBindsBirthAndRejectsUnknownBirth(t *testing.T) {
	directory := t.TempDir()
	created := time.Unix(1_700_000_000, 0)
	if err := RecordContainerHost(directory, 0.25, 0.5, 3, 0, 1000, created); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "host.prom")
	original, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(original), "cloud_clicker_gameserver_restart_count 3\n") ||
		!strings.Contains(string(original), "cloud_clicker_gameserver_container_created_timestamp_seconds 1700000000\n") {
		t.Fatal("restart count lost its bound container birth")
	}
	for _, invalid := range []time.Time{{}, time.Unix(0, 0), time.Unix(-1, 0)} {
		if err := RecordContainerHost(directory, 0.25, 0.5, 3, 0, 1000, invalid); !errors.Is(err, ErrInvalid) {
			t.Fatal("unknown/invalid container birth accepted")
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != string(original) {
			t.Fatal("invalid birth replaced the last valid observation")
		}
	}
}
