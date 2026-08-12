package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/harness"
)

func TestWriteRelevanceReportRejectsFailuresBeforeWriting(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report.json")
	report := harness.RelevanceReport{Failures: []string{"relevance_floor:upgrade.dead"}}

	err := writeRelevanceReport(output, report)
	if err == nil || !strings.Contains(err.Error(), "relevance_floor:upgrade.dead") {
		t.Fatalf("writeRelevanceReport error = %v", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("failing report was written: %v", statErr)
	}
	diagnosticBytes, readErr := os.ReadFile(relevanceDiagnosticPath(output))
	if readErr != nil {
		t.Fatal(readErr)
	}
	diagnostic := string(diagnosticBytes)
	if !strings.Contains(diagnostic, `"kind": "non_authoritative_relevance_diagnostic"`) ||
		!strings.Contains(diagnostic, `"authoritative": false`) || !strings.Contains(diagnostic, `"relevance_floor:upgrade.dead"`) {
		t.Fatalf("non-authoritative diagnostic=%s", diagnosticBytes)
	}
}

func TestWriteRelevanceReportRemovesStaleDiagnosticOnSuccess(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report.json")
	diagnostic := relevanceDiagnosticPath(output)
	if err := os.WriteFile(diagnostic, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := harness.RelevanceReport{}
	if err := writeRelevanceReport(output, report); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(diagnostic); !os.IsNotExist(statErr) {
		t.Fatalf("stale diagnostic survived success: %v", statErr)
	}
}
