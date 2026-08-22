package releasepackage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploymentSchemasAreClosedAndCoverCanonicalRequiredFields(t *testing.T) {
	root := filepath.Join("..", "..")
	if err := ValidateDeploymentSchemas(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "deployment", "config.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSchema(data, "config.schema.json", []string{"CLOUD_CLICKER_PUBLIC_ORIGIN", "invented_required_field"}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("forged required field set accepted: %v", err)
	}
}
