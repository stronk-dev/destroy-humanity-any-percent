package deploymentrelease

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/releasepackage"
)

type commandFixture struct {
	calls  [][]string
	output func([]string) ([]byte, error)
}

func TestDockerRuntimePreflightRunsCandidateConfigAndPrivateDatabaseInspection(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	if err := os.WriteFile(filepath.Join(bundle.Root, "candidate-byte"), []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	inspection := deploymentbackup.PostgresInspection{SchemaVersion: 1, DatabaseBytes: 4096,
		DatabaseMigration: bundle.Manifest.DatabaseMigration, EpochID: bundle.Manifest.EpochID,
		ConstantsHash: bundle.Manifest.ConstantsHash, ArtifactsVerified: 3}
	runner := &commandFixture{output: func(call []string) ([]byte, error) {
		if slices.Contains(call, "inspect") {
			return json.Marshal(map[string]any{"status": "inspected", "inspection": inspection})
		}
		if len(call) > 1 && call[1] == "info" {
			return []byte(bundle.Root + "\n"), nil
		}
		return nil, nil
	}}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	runtime.ReceiverHealthURL = "http://alert-receiver.invalid/-/healthy"
	runtime.Client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if err := runtime.Preflight(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	foundCandidateConfig := false
	for _, call := range runner.calls {
		if slices.Contains(call, "exec") {
			t.Fatalf("preflight exec-ed the old running container: %v", call)
		}
		if slices.Contains(call, "gameserver") && slices.Contains(call, "validate-config") && slices.Contains(call, "run") {
			foundCandidateConfig = true
		}
	}
	if !foundCandidateConfig {
		t.Fatalf("candidate config preflight absent: %v", runner.calls)
	}
	foundAlertmanagerConfig := false
	for _, call := range runner.calls {
		if strings.Contains(strings.Join(call, " "), "entrypoint=amtool") && slices.Contains(call, "check-config") {
			foundAlertmanagerConfig = true
		}
	}
	if !foundAlertmanagerConfig {
		t.Fatalf("candidate Alertmanager config preflight absent: %v", runner.calls)
	}

	inspection.DatabaseMigration--
	if _, err := runtime.inspectDatabase(context.Background(), bundle, true); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong private database identity accepted: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func (runner *commandFixture) Run(_ context.Context, _ string, name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	runner.calls = append(runner.calls, call)
	if runner.output == nil {
		return nil, nil
	}
	return runner.output(call)
}

func TestDockerRuntimeBindsBackupAndRestoreOutputToExactManifest(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	target := t.TempDir()
	identity := filepath.Join(t.TempDir(), "identity")
	if err := os.WriteFile(identity, []byte("AGE-SECRET-KEY-fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	header := deploymentbackup.Header{SchemaVersion: 1, BackupID: "20260822T180000Z-acde00000001", ServerID: "server-1",
		ReleaseManifestSHA256: bundle.ManifestSHA256, EpochID: bundle.Manifest.EpochID,
		StartedAt: time.Date(2026, 8, 22, 18, 0, 0, 0, time.UTC), CompletedAt: time.Date(2026, 8, 22, 18, 1, 0, 0, time.UTC),
		PayloadSHA256: "sha256:" + strings.Repeat("d", 64), PayloadBytes: 4096, PreUpgrade: true}
	hostBackup := filepath.Join(target, header.BackupID+".ccbackup")
	if err := os.WriteFile(hostBackup, []byte("encrypted fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner := &commandFixture{}
	runner.output = func(call []string) ([]byte, error) {
		if slices.Contains(call, "create") {
			return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(hostBackup), "header": header})
		}
		if slices.Contains(call, "restore") {
			return json.Marshal(map[string]any{"status": "restored", "header": header})
		}
		return nil, errors.New("unexpected command")
	}
	runtime := dockerFixtureRuntime(runner, target, identity)
	backup, err := runtime.CreatePreUpgradeBackup(context.Background(), bundle)
	if err != nil || backup.ID != header.BackupID || backup.Path != hostBackup {
		t.Fatalf("backup=%+v err=%v", backup, err)
	}
	if err := runtime.Restore(context.Background(), bundle, backup); err != nil {
		t.Fatal(err)
	}
	for _, call := range runner.calls {
		joined := strings.Join(call, " ")
		if strings.Contains(strings.ToLower(joined), " down ") || strings.Contains(strings.ToLower(joined), "migrate down") {
			t.Fatalf("rollback command admits a Down migration: %s", joined)
		}
	}

	severed := *runner
	severed.calls = nil
	badHeader := header
	badHeader.ReleaseManifestSHA256 = "sha256:" + strings.Repeat("e", 64)
	severed.output = func(call []string) ([]byte, error) {
		return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(hostBackup), "header": badHeader})
	}
	badRuntime := dockerFixtureRuntime(&severed, target, identity)
	if _, err := badRuntime.CreatePreUpgradeBackup(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong-manifest backup output accepted: %v", err)
	}
	symlink := filepath.Join(target, "20260822T180000Z-acde00000002.ccbackup")
	if err := os.Symlink(hostBackup, symlink); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Restore(context.Background(), bundle, BackupReference{ID: "20260822T180000Z-acde00000002", Path: symlink}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("symlinked backup accepted: %v", err)
	}
}

func TestDockerRuntimePrepareInspectsEveryRuntimeConfigDigest(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	runner := &commandFixture{}
	runner.output = func(call []string) ([]byte, error) {
		if len(call) >= 5 && call[1] == "image" && call[2] == "inspect" {
			reference := call[len(call)-1]
			for _, image := range bundle.Manifest.Images {
				if image.Reference == reference {
					return []byte(image.RuntimeConfigSHA256 + "\n"), nil
				}
			}
		}
		return []byte("ok\n"), nil
	}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	if err := runtime.Prepare(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	inspectCount, pullCount := 0, 0
	for _, call := range runner.calls {
		if len(call) > 2 && call[1] == "image" && call[2] == "inspect" {
			inspectCount++
		}
		if len(call) > 1 && call[1] == "pull" {
			pullCount++
		}
	}
	if inspectCount != 6 || pullCount != 5 {
		t.Fatalf("prepare calls=%v", runner.calls)
	}

	wrong := &commandFixture{output: func(call []string) ([]byte, error) {
		if len(call) > 2 && call[1] == "image" && call[2] == "inspect" {
			return []byte("sha256:" + strings.Repeat("f", 64) + "\n"), nil
		}
		return nil, nil
	}}
	if err := dockerFixtureRuntime(wrong, t.TempDir(), "").Prepare(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong runtime config digest accepted: %v", err)
	}
}

func TestDockerRuntimeRequiresDeliveredAlertNotHealthOnly(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	runner := &commandFixture{}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	if err := runtime.verifyAlertDelivery(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("delivery calls=%v", runner.calls)
	}
	joined := strings.Join(runner.calls[0], " ")
	for _, required := range []string{"deployment-operations", "alert-test", "http://alertmanager:9093", runtime.ReceiverHealthURL} {
		if !strings.Contains(joined, required) {
			t.Fatalf("delivery proof missing %q: %s", required, joined)
		}
	}
	if !strings.Contains(joined, " alertmanager alert-test ") {
		t.Fatalf("delivery verifier does not inherit Alertmanager receiver egress: %s", joined)
	}
}

func TestDockerRuntimeStartRecreatesCandidateServerProxyAndBackupWorker(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	runner := &commandFixture{}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	runtime.Client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if err := runtime.Start(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("start calls=%v", runner.calls)
	}
	joined := strings.Join(runner.calls[0], " ")
	for _, required := range []string{"--force-recreate", "gameserver", "caddy", "backup"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("candidate service %q not recreated: %s", required, joined)
		}
	}
}

func dockerFixtureRuntime(runner CommandRunner, target, identity string) DockerRuntime {
	return DockerRuntime{Runner: runner, PublicOrigin: "https://game.example", ReceiverHealthURL: "http://alertmanager:9093/-/healthy",
		BackupTarget: target, MetricsDirectory: target, AgeRecipient: "age1fixture",
		AgeIdentityFile: identity, ServerID: "server-1", DrainTimeout: 20 * time.Second}
}

func dockerFixtureBundle(t *testing.T) Bundle {
	t.Helper()
	root := t.TempDir()
	return Bundle{Root: root, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), Manifest: releasepackage.ReleaseManifest{
		ReleaseVersion: "1.0.0", EpochID: 8, ConstantsHash: "sha256:" + strings.Repeat("b", 64), Images: []releasepackage.Image{
			{Name: "alertmanager", Reference: "alertmanager:v1@sha256:" + strings.Repeat("1", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("7", 64)},
			{Name: "caddy", Reference: "caddy:v1@sha256:" + strings.Repeat("1", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("4", 64)},
			{Name: "gameserver", Reference: "sha256:" + strings.Repeat("2", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("2", 64)},
			{Name: "node-exporter", Reference: "node-exporter:v1@sha256:" + strings.Repeat("6", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("8", 64)},
			{Name: "postgres", Reference: "postgres:v1@sha256:" + strings.Repeat("3", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("5", 64)},
			{Name: "prometheus", Reference: "prometheus:v1@sha256:" + strings.Repeat("9", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("a", 64)},
		}}}
}
