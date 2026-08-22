package main

import (
	"errors"
	"testing"

	"cloud-clicker/server/deploymentrehearsal"
)

func TestRunValidateRejectsMissingAndInvalidEvidence(t *testing.T) {
	if _, err := runValidate(nil); !errors.Is(err, deploymentrehearsal.ErrInvalid) {
		t.Fatalf("empty evidence path accepted: %v", err)
	}
	path := t.TempDir() + "/missing.json"
	if _, err := runValidate([]string{"--evidence", path}); err == nil {
		t.Fatal("missing evidence accepted")
	}
}
