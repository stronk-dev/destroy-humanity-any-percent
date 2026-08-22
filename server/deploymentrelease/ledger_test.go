package deploymentrelease

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotationLedgerEnforcesAllGovernedOverlaps(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	for _, fixture := range []struct {
		family  KeyFamily
		overlap time.Duration
	}{{FamilyJWT, JWTOverlap}, {FamilyBootstrap, BootstrapOverlap}, {FamilyCursor, CursorOverlap}} {
		t.Run(string(fixture.family), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "rotation-ledger.jsonl")
			if err := ActivateRotation(path, fixture.family, "new-id", "old-id", "operator-1", base); err != nil {
				t.Fatal(err)
			}
			if err := RemovePrevious(path, fixture.family, "new-id", "old-id", "operator-1", base.Add(fixture.overlap-time.Nanosecond)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("premature removal accepted: %v", err)
			}
			if err := RemovePrevious(path, fixture.family, "new-id", "old-id", "operator-1", base.Add(fixture.overlap)); err != nil {
				t.Fatal(err)
			}
			records, err := ReadRotationLedger(path)
			if err != nil || len(records) != 2 || records[1].Action != "removed" {
				t.Fatalf("records=%+v err=%v", records, err)
			}
			data, _ := os.ReadFile(path)
			if containsSecretField(data) {
				t.Fatalf("ledger admits secret-shaped fields: %s", data)
			}
		})
	}
}

func TestRotationLedgerRejectsRewriteAndUnmatchedRemoval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rotation-ledger.jsonl")
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	if err := ActivateRotation(path, FamilyJWT, "new-id", "old-id", "operator-1", base); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.WriteString(`{"schema_version":1,"family":"jwt","action":"removed","current_id":"new-id","previous_id":"old-id","occurred_at":"2026-08-22T12:31:00Z","operator":"operator-1"}` + "\n")
	_ = file.Close()
	if _, err := ReadRotationLedger(path); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unmatched rewritten removal accepted: %v", err)
	}
}

func TestRotationLedgerRequiresContinuousNeverReusedKeyIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rotation-ledger.jsonl")
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	if err := ActivateRotation(path, FamilyJWT, "key-2", "key-1", "operator-1", base); err != nil {
		t.Fatal(err)
	}
	if err := RemovePrevious(path, FamilyJWT, "key-2", "key-1", "operator-1", base.Add(JWTOverlap)); err != nil {
		t.Fatal(err)
	}
	if err := ActivateRotation(path, FamilyJWT, "key-3", "unrelated", "operator-1", base.Add(time.Hour)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("discontinuous rotation accepted: %v", err)
	}
	if err := ActivateRotation(path, FamilyJWT, "key-1", "key-2", "operator-1", base.Add(time.Hour)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("retired key ID reuse accepted: %v", err)
	}
	if err := ActivateRotation(path, FamilyJWT, "key-3", "key-2", "operator-1", base.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func TestReleaseLedgerRequiresExplicitFailureStageAndImmutableOrdering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	record := fixtureReleaseRecord(base)
	if err := AppendReleaseRecord(path, record); err != nil {
		t.Fatal(err)
	}
	failed := fixtureReleaseRecord(base.Add(2 * time.Minute))
	failed.Result, failed.FailureStage, failed.RollbackUntil = "failed", "authenticated_smoke", time.Time{}
	if err := AppendReleaseRecord(path, failed); err != nil {
		t.Fatal(err)
	}
	invalid := fixtureReleaseRecord(base.Add(4 * time.Minute))
	invalid.Result, invalid.RollbackUntil = "failed", time.Time{}
	if err := AppendReleaseRecord(path, invalid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("stage-less failure accepted: %v", err)
	}
	invalid.FailureStage = "password_from_operator"
	if err := AppendReleaseRecord(path, invalid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown secret-shaped failure stage accepted: %v", err)
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(strings.ToLower(string(data)), "password") {
		t.Fatalf("release ledger leaked secret-shaped field: %s", data)
	}
}

func TestOperatorLockRejectsConcurrentMutation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release-ledger.jsonl.lock")
	first, err := acquireOperatorLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if second, err := acquireOperatorLock(path); !errors.Is(err, ErrInvalid) {
		if second != nil {
			_ = second.Close()
		}
		t.Fatalf("concurrent operator lock accepted: %v", err)
	}
}

func fixtureReleaseRecord(start time.Time) ReleaseRecord {
	return ReleaseRecord{SchemaVersion: 1, Action: "release", ReleaseVersion: "1.0.0",
		ManifestSHA256: "sha256:" + strings.Repeat("a", 64), PreviousVersion: "0.9.0",
		PreviousManifestSHA256: "sha256:" + strings.Repeat("e", 64), ImageDigests: []string{
			"alertmanager:v1@sha256:" + strings.Repeat("a", 64), "caddy:v1@sha256:" + strings.Repeat("b", 64), "sha256:" + strings.Repeat("c", 64),
			"node-exporter:v1@sha256:" + strings.Repeat("d", 64), "postgres:v1@sha256:" + strings.Repeat("e", 64), "prometheus:v1@sha256:" + strings.Repeat("f", 64),
		}, BackupID: "20260822T180000Z-acde00000001", StartedAt: start, CompletedAt: start.Add(time.Minute), RollbackUntil: start.Add(time.Minute).Add(RollbackWindow), Result: "succeeded", Operator: "operator-1"}
}
