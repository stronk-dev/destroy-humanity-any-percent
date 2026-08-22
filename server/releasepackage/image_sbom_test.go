package releasepackage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeImageSPDXIsDeterministicAndIdentityBound(t *testing.T) {
	raw := []byte(`{"spdxVersion":"SPDX-2.3","dataLicense":"CC0-1.0","SPDXID":"SPDXRef-DOCUMENT","name":"mutable","documentNamespace":"https://example.invalid/random","creationInfo":{"created":"2026-08-23T12:00:01Z","creators":["Tool: syft"]},"packages":[{"name":"fixture"}]}`)
	identity := "sha256:" + strings.Repeat("a", 64)
	created := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	first, err := NormalizeImageSPDX(raw, identity, created)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NormalizeImageSPDX(raw, identity, created)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("normalization changed bytes: err=%v", err)
	}
	if ValidateImageSPDX(first, identity) != nil || !bytes.Contains(first, []byte(`"created": "2026-08-23T12:00:00Z"`)) || bytes.Contains(first, []byte("random")) {
		t.Fatalf("normalized SPDX was not identity/time bound: %s", first)
	}
	path := filepath.Join(t.TempDir(), "image.spdx.json")
	if err := os.WriteFile(path, first, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteNormalizedImageSPDX(path, path, identity, created); err == nil {
		t.Fatal("normalizer overwrote an existing artifact")
	}
}

func TestNormalizeImageSPDXRejectsEmptyPackageGraphAndWrongIdentity(t *testing.T) {
	raw := []byte(`{"spdxVersion":"SPDX-2.3","dataLicense":"CC0-1.0","SPDXID":"SPDXRef-DOCUMENT","creationInfo":{},"packages":[]}`)
	if _, err := NormalizeImageSPDX(raw, "sha256:"+strings.Repeat("a", 64), time.Now()); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("empty package graph accepted: %v", err)
	}
	if _, err := NormalizeImageSPDX(raw, "mutable", time.Now()); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("mutable identity accepted: %v", err)
	}
}
