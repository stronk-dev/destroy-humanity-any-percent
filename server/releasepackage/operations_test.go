package releasepackage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOperationsProfileRequiresEveryAlertAndPrivacyBoundary(t *testing.T) {
	root := filepath.Join("..", "..", "deployment")
	if err := ValidateOperationsProfile(root); err != nil {
		t.Fatal(err)
	}
	copyRoot := filepath.Join(t.TempDir(), "deployment")
	if err := copyTree(filepath.Join(root, "operations"), filepath.Join(copyRoot, "operations")); err != nil {
		t.Fatal(err)
	}
	rulesPath := filepath.Join(copyRoot, "operations", "cloud-clicker-alerts.yml")
	rules, _ := os.ReadFile(rulesPath)
	rules = append(rules, []byte("\n# founder_id\n")...)
	if err := os.WriteFile(rulesPath, rules, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateOperationsProfile(copyRoot); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("private label accepted: %v", err)
	}
}

func TestOperationsProfileRejectsMissingAlertAndResolvedEvidence(t *testing.T) {
	root := filepath.Join("..", "..", "deployment")
	copyProfile := func(t *testing.T) string {
		t.Helper()
		copyRoot := filepath.Join(t.TempDir(), "deployment")
		if err := copyTree(filepath.Join(root, "operations"), filepath.Join(copyRoot, "operations")); err != nil {
			t.Fatal(err)
		}
		return copyRoot
	}
	t.Run("missing rule", func(t *testing.T) {
		copyRoot := copyProfile(t)
		path := filepath.Join(copyRoot, "operations", "cloud-clicker-alerts.yml")
		data, _ := os.ReadFile(path)
		data = bytes.Replace(data, []byte("alert: CloudClickerCleanupJobFailed"), []byte("alert: SeveredCleanupJobFailed"), 1)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := ValidateOperationsProfile(copyRoot); !errors.Is(err, ErrInvalidContent) {
			t.Fatalf("missing alert accepted: %v", err)
		}
	})
	t.Run("missing resolved state", func(t *testing.T) {
		copyRoot := copyProfile(t)
		path := filepath.Join(copyRoot, "operations", "cloud-clicker-alerts.test.yml")
		data, _ := os.ReadFile(path)
		needle := []byte("        alertname: CloudClickerCleanupJobFailed\n        exp_alerts: []")
		replacement := []byte("        alertname: SeveredCleanupJobFailed\n        exp_alerts: []")
		if !bytes.Contains(data, needle) {
			t.Fatal("resolved fixture shape changed")
		}
		data = bytes.Replace(data, needle, replacement, 1)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := ValidateOperationsProfile(copyRoot); !errors.Is(err, ErrInvalidContent) {
			t.Fatalf("missing resolved evidence accepted: %v", err)
		}
	})
}
