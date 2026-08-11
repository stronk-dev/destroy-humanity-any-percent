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
}
