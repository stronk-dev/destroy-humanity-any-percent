package releasepackage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func secretMetadataArchive(t *testing.T, compressed bool, field, value string) (string, string) {
	t.Helper()
	inner := &tar.Header{Name: "content.txt", Mode: 0o644, Typeflag: tar.TypeReg,
		Uname: "ordinary-owner", Gname: "ordinary-group", PAXRecords: map[string]string{"fixture.note": "ordinary"},
		Xattrs: map[string]string{"user.fixture": "ordinary"}}
	outer := &tar.Header{Name: "layer/layer.tar", Mode: 0o644, Typeflag: tar.TypeReg,
		Uname: "ordinary-owner", Gname: "ordinary-group", PAXRecords: map[string]string{"fixture.note": "ordinary"},
		Xattrs: map[string]string{"user.fixture": "ordinary"}}
	mutate := func(header *tar.Header, property string) {
		switch property {
		case "name":
			header.Name = "metadata-" + value
		case "link":
			header.Typeflag = tar.TypeSymlink
			header.Linkname = value
		case "uname":
			header.Uname = value
		case "gname":
			header.Gname = value
		case "pax-value", "directory-pax":
			header.PAXRecords["fixture.note"] = value
			if property == "directory-pax" {
				header.Typeflag = tar.TypeDir
			}
		case "pax-key":
			header.PAXRecords["fixture."+value] = "ordinary"
		case "xattr-value":
			header.Xattrs["user.fixture"] = value
		case "xattr-key":
			header.Xattrs["user."+value] = "ordinary"
		default:
			t.Fatalf("unknown metadata fixture field %q", property)
		}
	}
	if strings.HasPrefix(field, "outer-") && field != "outer-directory-pax" {
		mutate(outer, strings.TrimPrefix(field, "outer-"))
	} else if strings.HasPrefix(field, "layer-") {
		mutate(inner, strings.TrimPrefix(field, "layer-"))
	}
	var layer bytes.Buffer
	writer := tar.NewWriter(&layer)
	payload := []byte("ordinary payload\n")
	if inner.Typeflag != tar.TypeReg {
		payload = nil
	}
	inner.Size = int64(len(payload))
	if err := writer.WriteHeader(inner); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	layerBytes := layer.Bytes()
	layerID := digest(layerBytes)
	if compressed {
		var buffer bytes.Buffer
		compressor := gzip.NewWriter(&buffer)
		compressor.Name, compressor.Comment, compressor.Extra = "ordinary-layer", "ordinary", []byte("ordinary")
		switch field {
		case "gzip-name":
			compressor.Name = value
		case "gzip-comment":
			compressor.Comment = value
		case "gzip-extra":
			compressor.Extra = []byte(value)
		}
		if _, err := compressor.Write(layerBytes); err != nil {
			t.Fatal(err)
		}
		if err := compressor.Close(); err != nil {
			t.Fatal(err)
		}
		layerBytes = buffer.Bytes()
	}
	// Reuse the existing image identity template, then bind the actual metadata-
	// bearing layer's uncompressed hash. The image validator must accept it first.
	base, baseID := fixtureDockerArchive(t, t.TempDir(), "amd64")
	baseBytes, err := os.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(bytes.NewReader(baseBytes))
	var config map[string]any
	for {
		header, err := reader.Next()
		if err != nil {
			t.Fatalf("image identity fixture missing: %v", err)
		}
		if header.Name == strings.TrimPrefix(baseID, "sha256:")+".json" {
			if err := json.NewDecoder(reader).Decode(&config); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	config["rootfs"] = map[string]any{"type": "layers", "diff_ids": []string{layerID}}
	configBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	imageID := digest(configBytes)
	configName := strings.TrimPrefix(imageID, "sha256:") + ".json"
	manifest, err := json.Marshal([]dockerArchiveManifest{{Config: configName, Layers: []string{outer.Name}}})
	if err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	writer = tar.NewWriter(&archive)
	entries := []struct {
		header *tar.Header
		data   []byte
	}{
		{&tar.Header{Name: configName, Mode: 0o644, Typeflag: tar.TypeReg}, configBytes},
		{&tar.Header{Name: "manifest.json", Mode: 0o644, Typeflag: tar.TypeReg}, manifest},
		{outer, layerBytes},
	}
	if field == "outer-directory-pax" {
		entries = append(entries, struct {
			header *tar.Header
			data   []byte
		}{&tar.Header{Name: "extra/", Mode: 0o755, Typeflag: tar.TypeDir,
			PAXRecords: map[string]string{"fixture.note": value}}, nil})
	}
	for _, entry := range entries {
		entry.header.Size = int64(len(entry.data))
		if err := writer.WriteHeader(entry.header); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(entry.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "metadata.docker.tar")
	if err := os.WriteFile(path, archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDockerArchive(path, imageID, "0.1.0-preview.1", strings.Repeat("d", 40)); err != nil {
		t.Fatalf("metadata image fails identity/layer validation before scan: %v", err)
	}
	return path, imageID
}

func TestSecretScanCoversImageMetadataWithoutLeakingPaths(t *testing.T) {
	seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	fields := []string{"outer-name", "outer-uname", "outer-gname", "outer-pax-value", "outer-pax-key", "outer-xattr-value", "outer-xattr-key", "outer-directory-pax",
		"layer-name", "layer-link", "layer-uname", "layer-gname", "layer-pax-value", "layer-pax-key", "layer-xattr-value", "layer-xattr-key", "layer-directory-pax"}
	for _, compressed := range []bool{false, true} {
		population := append([]string(nil), fields...)
		if compressed {
			population = append(population, "gzip-name", "gzip-comment", "gzip-extra")
		}
		for _, field := range population {
			t.Run(fmt.Sprintf("gzip=%t/%s", compressed, field), func(t *testing.T) {
				path, _ := secretMetadataArchive(t, compressed, field, seed)
				findings, err := ScanDockerArchive(path)
				if err != nil || len(findings) == 0 {
					t.Fatalf("metadata-only sentinel missed: findings=%d err=%v", len(findings), err)
				}
				found := false
				for _, finding := range findings {
					found = found || finding.Rule == "seeded-fixture"
					if strings.Contains(finding.Path, seed) {
						t.Fatal("secret-bearing metadata entered a diagnostic path")
					}
				}
				if !found {
					t.Fatal("refusal lacks the sentinel rule")
				}
				gate := RequireNoSecrets(findings)
				if !errors.Is(gate, ErrInvalidContent) || strings.Contains(gate.Error(), seed) {
					t.Fatal("secret gate admitted metadata or exposed its value")
				}
			})
		}
		t.Run(fmt.Sprintf("gzip=%t/clean-metadata", compressed), func(t *testing.T) {
			path, _ := secretMetadataArchive(t, compressed, "clean", "ordinary")
			findings, err := ScanDockerArchive(path)
			if err != nil || len(findings) != 0 || RequireNoSecrets(findings) != nil {
				t.Fatal("clean metadata image refused")
			}
		})
	}
}

func TestSecretScanMetadataErrorPathsDoNotExposeSecretNames(t *testing.T) {
	seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	for name, data := range map[string][]byte{"gzip": {0x1f, 0x8b, 0x08}, "zstd": {0x28, 0xb5, 0x2f, 0xfd}} {
		t.Run(name, func(t *testing.T) {
			_, err := scanArchiveMember("image!"+seed, data, 0)
			if !errors.Is(err, ErrInvalidContent) || strings.Contains(err.Error(), seed) {
				t.Fatal("unreadable compression diagnostic exposes secret-bearing name")
			}
		})
	}
	t.Run("direct-gate-path", func(t *testing.T) {
		err := RequireNoSecrets([]SecretFinding{{Path: "image!" + seed, Rule: "seeded-fixture"}})
		if !errors.Is(err, ErrInvalidContent) || strings.Contains(err.Error(), seed) {
			t.Fatal("secret gate exposes an externally supplied finding path")
		}
	})
}
