package operations

import (
	"errors"
	"math"
	"path/filepath"
	"testing"
)

func TestFilesystemUsageReadsSubjectAndRejectsMissingPath(t *testing.T) {
	directory := t.TempDir()
	usage, err := FilesystemUsage(directory)
	if err != nil || math.IsNaN(usage) || math.IsInf(usage, 0) || usage < 0 || usage > 1 {
		t.Fatalf("actual filesystem usage invalid: %v %v", usage, err)
	}
	if _, err := FilesystemUsage(filepath.Join(directory, "missing")); !errors.Is(err, ErrInvalid) {
		t.Fatal("unavailable filesystem observation reported success")
	}
}
