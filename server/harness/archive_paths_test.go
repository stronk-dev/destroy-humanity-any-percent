package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestActiveSourcesDoNotReferencePreArchiveT0T1Evidence(t *testing.T) {
	root := filepath.Join("..", "..")
	oldSlashPath := strings.Join([]string{"planning", "t0-t1-content"}, "/")
	oldJoinPath := `"` + strings.Join([]string{"planning", "t0-t1-content"}, `", "`) + `"`

	paths := []string{filepath.Join(root, "Makefile"), filepath.Join(root, "server"), filepath.Join(root, "client", "tools")}
	for _, path := range paths {
		err := filepath.WalkDir(path, func(candidate string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if filepath.Base(candidate) == "archive_paths_test.go" {
				return nil
			}
			extension := filepath.Ext(candidate)
			if filepath.Base(candidate) != "Makefile" && extension != ".go" && extension != ".mjs" {
				return nil
			}
			data, err := os.ReadFile(candidate)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), oldSlashPath) || strings.Contains(string(data), oldJoinPath) {
				t.Errorf("active source still references pre-archive T0-T1 evidence: %s", candidate)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
