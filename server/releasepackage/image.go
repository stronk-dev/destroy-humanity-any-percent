package releasepackage

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type dockerArchiveManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type dockerImageConfig struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	RootFS       struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
	Config struct {
		User       string            `json:"User"`
		Entrypoint []string          `json:"Entrypoint"`
		Labels     map[string]string `json:"Labels"`
	} `json:"config"`
}

// ValidateDockerArchive binds an offline docker-save archive to the bare
// content-addressed image ID used by the release Compose file.
func ValidateDockerArchive(path, expectedReference, releaseVersion, sourceCommit string) error {
	if !strings.HasPrefix(expectedReference, "sha256:") || !hashPattern.MatchString(expectedReference) {
		return fmt.Errorf("%w: gameserver image must use its bare image ID", ErrInvalidContent)
	}
	file, err := os.Open(path)
	if err != nil {
		return errors.Join(ErrInvalidContent, err)
	}
	defer file.Close()
	reader := tar.NewReader(file)
	var manifestBytes, configBytes []byte
	hash := strings.TrimPrefix(expectedReference, "sha256:")
	wantConfigs := map[string]bool{hash + ".json": true, "blobs/sha256/" + hash: true}
	seen := map[string]bool{}
	layerDiffIDs := map[string]string{}
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return errors.Join(ErrInvalidContent, err)
		}
		cleanName := strings.TrimSuffix(header.Name, "/")
		if cleanName == "" || !validRelativePath(cleanName) || seen[header.Name] || header.Size < 0 || header.Size > 512<<20 {
			return ErrInvalidContent
		}
		seen[header.Name] = true
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA && header.Typeflag != tar.TypeDir {
			return ErrInvalidContent
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if header.Name == "manifest.json" || wantConfigs[header.Name] {
			data, err := io.ReadAll(io.LimitReader(reader, header.Size+1))
			if err != nil || int64(len(data)) != header.Size {
				return ErrInvalidContent
			}
			if header.Name == "manifest.json" {
				manifestBytes = data
			} else {
				configBytes = data
			}
		} else {
			diffID, err := archiveMemberDiffID(reader, header.Size)
			if err != nil {
				return err
			}
			layerDiffIDs[header.Name] = diffID
		}
	}
	if len(manifestBytes) == 0 || len(configBytes) == 0 || digest(configBytes) != expectedReference {
		return ErrInvalidContent
	}
	var manifest []dockerArchiveManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	if decoder.Decode(&manifest) != nil || decoder.Decode(&struct{}{}) != io.EOF || len(manifest) != 1 || !wantConfigs[manifest[0].Config] || len(manifest[0].Layers) == 0 {
		return ErrInvalidContent
	}
	var config dockerImageConfig
	decoder = json.NewDecoder(bytes.NewReader(configBytes))
	if decoder.Decode(&config) != nil || decoder.Decode(&struct{}{}) != io.EOF || config.Architecture != "amd64" || config.OS != "linux" ||
		config.Config.User != "65532:65532" || len(config.Config.Entrypoint) != 1 || config.Config.Entrypoint[0] != "/usr/local/bin/gameserver" ||
		config.Config.Labels["org.opencontainers.image.version"] != releaseVersion || config.Config.Labels["org.opencontainers.image.revision"] != sourceCommit ||
		config.Config.Labels["org.opencontainers.image.source"] != "https://github.com/stronk-dev/destroy-humanity-any-percent" ||
		config.Config.Labels["org.opencontainers.image.licenses"] != "MIT" {
		return ErrInvalidContent
	}
	if config.RootFS.Type != "layers" || len(config.RootFS.DiffIDs) != len(manifest[0].Layers) {
		return ErrInvalidContent
	}
	for index, layer := range manifest[0].Layers {
		if !hashPattern.MatchString(config.RootFS.DiffIDs[index]) || layerDiffIDs[layer] != config.RootFS.DiffIDs[index] {
			return fmt.Errorf("%w: gameserver layer content identity mismatch", ErrInvalidContent)
		}
	}
	return nil
}

// OCI DiffIDs hash uncompressed layer bytes, not a gzip member's compressed
// blob identity. Reuse the existing archive scanner's decompression bound and
// supported encodings; an unreadable or oversized member is never skipped.
func archiveMemberDiffID(source io.Reader, size int64) (string, error) {
	buffered := bufio.NewReader(io.LimitReader(source, size))
	prefix, err := buffered.Peek(4)
	if err != nil && !errors.Is(err, io.EOF) || bytes.HasPrefix(prefix, zstdMagic) {
		return "", errors.Join(ErrInvalidContent, err)
	}
	var input io.Reader = buffered
	if bytes.HasPrefix(prefix, gzipMagic) {
		decompressor, err := gzip.NewReader(buffered)
		if err != nil {
			return "", errors.Join(ErrInvalidContent, err)
		}
		defer decompressor.Close()
		input = decompressor
	}
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(input, maximumSecretScanDecompressed+1))
	if err != nil || written > maximumSecretScanDecompressed {
		return "", errors.Join(ErrInvalidContent, err)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
