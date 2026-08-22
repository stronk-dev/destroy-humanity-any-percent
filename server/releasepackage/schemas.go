package releasepackage

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type schemaEnvelope struct {
	Schema               string                     `json:"$schema"`
	ID                   string                     `json:"$id"`
	Type                 string                     `json:"type"`
	AdditionalProperties *bool                      `json:"additionalProperties"`
	Required             []string                   `json:"required"`
	Properties           map[string]json.RawMessage `json:"properties"`
}

func ValidateDeploymentSchemas(root string) error {
	wants := map[string][]string{
		"config.schema.json":           {"CLOUD_CLICKER_PUBLIC_ORIGIN", "CLOUD_CLICKER_SERVER_ID", "CLOUD_CLICKER_JWT_CURRENT_ID", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_ID", "CLOUD_CLICKER_DATABASE_URL_SECRET_FILE", "CLOUD_CLICKER_POSTGRES_PASSWORD_SECRET_FILE", "CLOUD_CLICKER_JWT_CURRENT_SECRET_FILE", "CLOUD_CLICKER_BOOTSTRAP_CURRENT_SECRET_FILE"},
		"release-manifest.schema.json": {"schema_version", "release_version", "source_commit", "platform", "docker_engine_version", "docker_compose_version", "database_migration", "company_save_version", "founder_save_version", "epoch_id", "constants_hash", "copy_hash", "images", "artifacts"},
	}
	for name, required := range wants {
		data, err := os.ReadFile(filepath.Join(root, "deployment", name))
		if err != nil || validateSchema(data, name, required) != nil {
			return ErrInvalidContent
		}
	}
	return nil
}

func validateSchema(data []byte, name string, required []string) error {
	var schema schemaEnvelope
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&schema) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		schema.Schema != "https://json-schema.org/draft/2020-12/schema" || !strings.HasSuffix(schema.ID, "/deployment/"+name) ||
		schema.Type != "object" || schema.AdditionalProperties == nil || *schema.AdditionalProperties || len(schema.Properties) == 0 {
		return ErrInvalidContent
	}
	actual := append([]string(nil), schema.Required...)
	want := append([]string(nil), required...)
	sort.Strings(actual)
	sort.Strings(want)
	if !sameStrings(actual, want) {
		return ErrInvalidContent
	}
	for _, field := range required {
		if len(schema.Properties[field]) == 0 {
			return ErrInvalidContent
		}
	}
	return nil
}
