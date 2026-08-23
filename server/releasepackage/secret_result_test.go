package releasepackage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSecretScanResultRequiresCompletedSourceAndImagePopulation(t *testing.T) {
	valid := validSecretScanResult()
	if err := ValidateSecretScanResult(valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*SecretScanResult){
		"no files":        func(value *SecretScanResult) { value.TrackedFiles = 0 },
		"no image":        func(value *SecretScanResult) { value.ImageArchiveScanned = false },
		"finding":         func(value *SecretScanResult) { value.Findings = 1 },
		"incomplete":      func(value *SecretScanResult) { value.ObjectiveCompleted = false },
		"guard exhausted": func(value *SecretScanResult) { value.GuardExhausted = true },
	} {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if err := ValidateSecretScanResult(value); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("invalid secret result accepted: %v", err)
			}
		})
	}
}

func TestSecretScanResultDecodeAndWriteFailClosed(t *testing.T) {
	valid := validSecretScanResult()
	data, _ := json.Marshal(valid)
	if _, err := DecodeSecretScanResult(data); err != nil {
		t.Fatal(err)
	}
	unknown := append(append([]byte(nil), data[:len(data)-1]...), []byte(`,"summary":"clean"}`)...)
	if _, err := DecodeSecretScanResult(unknown); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("unknown summary accepted: %v", err)
	}
	path := filepath.Join(t.TempDir(), "secret-scan.json")
	if err := WriteSecretScanResult(path, valid); err != nil {
		t.Fatal(err)
	}
	if err := WriteSecretScanResult(path, valid); err == nil {
		t.Fatal("secret result overwritten")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode=%v err=%v", info.Mode(), err)
	}
}

func validSecretScanResult() SecretScanResult {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	return SecretScanResult{SchemaVersion: 1, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), SourceCommit: strings.Repeat("b", 40),
		StartedAt: start, CompletedAt: start.Add(time.Second), TrackedFiles: 100, ImageArchiveScanned: true, ObjectiveCompleted: true}
}
