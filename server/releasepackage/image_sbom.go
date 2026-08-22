package releasepackage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// NormalizeImageSPDX binds Syft's discovered package graph to the immutable
// runtime config identity and the release's declared creation time. Syft may
// otherwise emit a random document namespace and wall-clock timestamp, which
// would make an identical release input produce different bundle bytes.
func NormalizeImageSPDX(input []byte, runtimeConfigSHA256 string, created time.Time) ([]byte, error) {
	if !hashPattern.MatchString(runtimeConfigSHA256) || created.IsZero() {
		return nil, ErrInvalidContent
	}
	var document map[string]any
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	if decoder.Decode(&document) != nil || decoder.Decode(&struct{}{}) != io.EOF || len(document) == 0 {
		return nil, ErrInvalidContent
	}
	if document["spdxVersion"] != "SPDX-2.3" || document["dataLicense"] != "CC0-1.0" || document["SPDXID"] != "SPDXRef-DOCUMENT" {
		return nil, ErrInvalidContent
	}
	creation, ok := document["creationInfo"].(map[string]any)
	if !ok {
		return nil, ErrInvalidContent
	}
	if packages, ok := document["packages"].([]any); !ok || len(packages) == 0 {
		return nil, ErrInvalidContent
	}
	name := strings.ReplaceAll(runtimeConfigSHA256, ":", "-")
	document["name"] = name
	document["documentNamespace"] = "https://github.com/stronk-dev/destroy-humanity-any-percent/sbom/image/" + strings.TrimPrefix(runtimeConfigSHA256, "sha256:")
	creation["created"] = created.UTC().Format(time.RFC3339)
	creation["creators"] = []any{"Tool: syft-v1.51.0", "Tool: cloud-clicker/normalize-image-sbom"}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, errors.Join(ErrInvalidContent, err)
	}
	encoded = append(encoded, '\n')
	if ValidateImageSPDX(encoded, runtimeConfigSHA256) != nil {
		return nil, ErrInvalidContent
	}
	return encoded, nil
}

func WriteNormalizedImageSPDX(inputPath, outputPath, runtimeConfigSHA256 string, created time.Time) error {
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	output, err := NormalizeImageSPDX(input, runtimeConfigSHA256, created)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(output); err == nil {
		err = file.Sync()
	}
	return errors.Join(err, file.Close())
}
