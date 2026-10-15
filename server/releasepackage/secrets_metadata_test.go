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
	} else if field == "multistream-split-layer-name" {
		mutate(inner, "name")
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
		prefix := 0
		if strings.HasPrefix(field, "multistream-") {
			count := 1
			if strings.HasPrefix(field, "multistream-third-") {
				count = 2
			}
			for range count {
				first := gzip.NewWriter(&buffer)
				if strings.HasPrefix(field, "multistream-split-") {
					prefix = len(layerBytes) / 2
					if _, err := first.Write(layerBytes[:prefix]); err != nil {
						t.Fatal(err)
					}
				}
				if err := first.Close(); err != nil {
					t.Fatal(err)
				}
			}
		}
		compressor := gzip.NewWriter(&buffer)
		compressor.Name, compressor.Comment, compressor.Extra = "ordinary-layer", "ordinary", []byte("ordinary")
		switch field {
		case "gzip-name", "multistream-name", "multistream-third-name":
			compressor.Name = value
		case "gzip-comment", "multistream-comment", "multistream-third-comment":
			compressor.Comment = value
		case "gzip-extra", "multistream-extra", "multistream-third-extra":
			compressor.Extra = []byte(value)
		}
		if _, err := compressor.Write(layerBytes[prefix:]); err != nil {
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

func TestSecretScanConcatenatedGzipMetadata(t *testing.T) {
	seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	for _, member := range []string{"multistream-", "multistream-third-"} {
		for _, field := range []string{"name", "comment", "extra"} {
			t.Run(member+field, func(t *testing.T) {
				path, _ := secretMetadataArchive(t, true, member+field, seed)
				findings, err := ScanDockerArchive(path)
				if err != nil || len(findings) != 1 || findings[0].Rule != "seeded-fixture" ||
					strings.Contains(findings[0].Path, seed) {
					t.Fatalf("later-member metadata not safely detected: findings%d err%v", len(findings), err)
				}
				if err := RequireNoSecrets(findings); !errors.Is(err, ErrInvalidContent) || strings.Contains(err.Error(), seed) {
					t.Fatal("later-member metadata gate admitted or exposed its value")
				}
			})
		}
	}
	for _, field := range []string{"multistream-name", "multistream-third-name", "multistream-split-clean", "multistream-split-layer-name"} {
		t.Run("payload-control/"+field, func(t *testing.T) {
			value, want := "ordinary", 0
			if field == "multistream-split-layer-name" {
				value, want = seed, 1
			}
			path, _ := secretMetadataArchive(t, true, field, value)
			findings, err := ScanDockerArchive(path)
			if err != nil || len(findings) != want {
				t.Fatal("concatenated tar payload semantics changed", len(findings), err)
			}
		})
	}
}

func TestSecretScanConcatenatedGzipBoundsAndIntegrity(t *testing.T) {
	encode := func(parts ...string) []byte {
		t.Helper()
		var compressed bytes.Buffer
		for _, part := range parts {
			writer := gzip.NewWriter(&compressed)
			if _, err := writer.Write([]byte(part)); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
		}
		return compressed.Bytes()
	}
	for _, row := range []struct {
		name  string
		parts []string
		limit int
		want  string
		valid bool
	}{
		{"exact combined bound", []string{"abcd", "efgh"}, 8, "abcdefgh", true},
		{"later member crosses bound", []string{"abcd", "efgh"}, 7, "", false},
		{"third member crosses bound", []string{"abcd", "efgh", "i"}, 8, "", false},
		{"empty members retain payload", []string{"", "abcd", "", "efgh", ""}, 8, "abcdefgh", true},
		{"empty exact zero bound", []string{"", ""}, 0, "", true},
		{"nonempty zero bound", []string{"", "x"}, 0, "", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			plain, findings, err := readSecretScanGzip("ordinary", encode(row.parts...), row.limit)
			if row.valid {
				if err != nil || string(plain) != row.want || len(findings) != 0 {
					t.Fatal("valid combined payload/bound refused", err)
				}
			} else if !errors.Is(err, ErrInvalidContent) || plain != nil || findings != nil {
				t.Fatal("combined bound violation returned usable content", err)
			}
		})
	}
	for _, kind := range []string{"checksum", "truncated", "trailing"} {
		t.Run(kind, func(t *testing.T) {
			data := encode("abcd", "efgh")
			switch kind {
			case "checksum":
				data[len(data)-8] ^= 1
			case "truncated":
				data = data[:len(data)-2]
			case "trailing":
				data = append(data, 'x', 'y')
			}
			plain, findings, err := readSecretScanGzip("ordinary", data, 8)
			if !errors.Is(err, ErrInvalidContent) || plain != nil || findings != nil {
				t.Fatal("invalid later stream returned usable content", err)
			}
		})
	}
	t.Run("recognized payload crosses member boundary", func(t *testing.T) {
		seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
		findings, err := scanArchiveMember("ordinary", encode(seed[:10], seed[10:]), 0)
		if err != nil || len(findings) != 1 || findings[0].Rule != "seeded-fixture" {
			t.Fatal("member boundary hid a recognized payload", len(findings), err)
		}
	})
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
