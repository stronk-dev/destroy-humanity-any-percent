package harness

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Preserve exact byte equality: excerpts diagnose drift, never normalize or
// accept it. In particular an identity mismatch does not excuse other changes.
func comparePinnedReputationReport(path string, observed []byte) error {
	pinned, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read pinned Reputation report %q: %w", path, err)
	}
	if bytes.Equal(pinned, observed) {
		return nil
	}
	offset := 0
	for offset < min(len(pinned), len(observed)) && pinned[offset] == observed[offset] {
		offset++
	}
	excerpt := func(data []byte) string {
		return fmt.Sprintf("%+q", data[max(0, offset-32):min(len(data), offset+32)])
	}
	return fmt.Errorf("Reputation report %q drifted at byte %d (zero-based); pinned bytes=%d sha256=%s context=%s; observed bytes=%d sha256=%s context=%s; inspect the first difference and remaining report before any separately authorized regeneration",
		path, offset, len(pinned), reputationMeasurementSHA(pinned), excerpt(pinned),
		len(observed), reputationMeasurementSHA(observed), excerpt(observed))
}

func TestPinnedReputationReportComparison(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	pinned := []byte("{\n \"count\": 1\n}\n")
	if err := os.WriteFile(path, pinned, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		name     string
		observed []byte
		offset   int
	}{
		{"identical", bytes.Clone(pinned), -1},
		{"value", []byte("{\n \"count\": 2\n}\n"), 12},
		{"formatting_is_still_drift", append([]byte(" "), pinned...), 0},
		{"shorter", pinned[:len(pinned)-1], len(pinned) - 1},
		{"longer", append(bytes.Clone(pinned), '\n'), len(pinned)},
		{"empty", nil, 0},
		{"binary_context", []byte{0xff, '\n', 0}, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			err := comparePinnedReputationReport(path, row.observed)
			if row.offset < 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("changed report was accepted")
			}
			for _, want := range []string{path, fmt.Sprintf("at byte %d (zero-based)", row.offset),
				fmt.Sprintf("pinned bytes=%d sha256=%x", len(pinned), sha256.Sum256(pinned)),
				fmt.Sprintf("observed bytes=%d sha256=%x", len(row.observed), sha256.Sum256(row.observed))} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("diagnostic omitted %q: %v", want, err)
				}
			}
			if strings.ContainsAny(err.Error(), "\n\r\x00") {
				t.Fatal("report bytes escaped the quoted, single-line context")
			}
		})
	}
	if err := comparePinnedReputationReport(path+".missing", pinned); !errors.Is(err, fs.ErrNotExist) || strings.Contains(err.Error(), "drifted") {
		t.Fatalf("missing report was misreported as measured drift: %v", err)
	}
	unchanged, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(unchanged, pinned) {
		t.Fatal("comparison changed the pinned report")
	}
}

func TestPinnedReputationReportDiagnosticUsesRealReports(t *testing.T) {
	for _, report := range []string{reputationCareerReportPath, reputationRelevancePath} {
		t.Run(filepath.Base(report), func(t *testing.T) {
			path := filepath.Join(repositoryRootForReputation, report)
			pinned, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := comparePinnedReputationReport(path, pinned); err != nil {
				t.Fatal(err)
			}
			// A late-byte mutation in each actual pin must not dump megabytes,
			// excuse formatting, or produce the old uninformative <nil> message.
			observed := bytes.Clone(pinned)
			observed[len(observed)-1] ^= 1
			err = comparePinnedReputationReport(path, observed)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("at byte %d (zero-based)", len(pinned)-1)) || len(err.Error()) > len(path)+1100 {
				t.Fatalf("real report diagnostic is absent, wrong or unbounded: %v", err)
			}
			if strings.Contains(err.Error(), "<nil>") {
				t.Fatal("drift diagnostic still reports a nil read error")
			}
		})
	}
}
