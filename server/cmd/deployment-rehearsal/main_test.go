package main

import (
	"context"
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

func TestRunValidateBuildRejectsMissingRecord(t *testing.T) {
	if _, err := runValidateBuild(nil); !errors.Is(err, deploymentrehearsal.ErrInvalid) {
		t.Fatalf("empty build record accepted: %v", err)
	}
	if _, err := runValidateBuild([]string{"--record", t.TempDir() + "/missing.json"}); err == nil {
		t.Fatal("missing build record accepted")
	}
}

func TestRunPlanRequiresReviewedPlanAndEmptyOutput(t *testing.T) {
	if err := runPlan(context.Background(), nil); !errors.Is(err, deploymentrehearsal.ErrInvalid) {
		t.Fatalf("empty plan accepted: %v", err)
	}
	if err := runPlan(context.Background(), []string{"--plan", t.TempDir() + "/missing.json", "--output", t.TempDir()}); err == nil {
		t.Fatal("missing plan accepted")
	}
}

func TestRunProbeDistinguishesInvalidSetupFromExpectedRejection(t *testing.T) {
	if outcome, err := runProbe(nil); !errors.Is(err, deploymentrehearsal.ErrInvalid) || outcome == deploymentrehearsal.ProbeRejected {
		t.Fatalf("missing probe inputs counted as rejection: outcome=%d err=%v", outcome, err)
	}
}
