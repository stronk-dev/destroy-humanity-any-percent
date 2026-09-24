package deploymentrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"filippo.io/age"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/epochseed"
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

func TestDockerRuntimeInitialInstallRequiresCleanStateAndCleansFailedStart(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	if err := os.WriteFile(filepath.Join(bundle.Root, "candidate-byte"), []byte("candidate"), 0o600); err != nil {
		t.Fatal(err)
	}
	dockerRoot := t.TempDir()
	runner := &commandFixture{output: func(call []string) ([]byte, error) {
		if len(call) > 1 && call[1] == "info" {
			return []byte(dockerRoot + "\n"), nil
		}
		return nil, nil
	}}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	runtime.Client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if err := runtime.PreflightInstall(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	psChecks, volumeChecks := 0, 0
	for _, call := range runner.calls {
		joined := strings.Join(call, " ")
		if strings.Contains(joined, " ps --all --quiet") {
			psChecks++
		}
		if strings.Contains(joined, "volume ls --quiet --filter=name=^cloud-clicker_postgres_data$") {
			volumeChecks++
		}
		if strings.Contains(joined, " backup inspect ") {
			t.Fatalf("initial install inspected nonexistent database: %s", joined)
		}
	}
	if psChecks != 2 || volumeChecks != 2 {
		t.Fatalf("clean state not checked around preflight: ps=%d volume=%d calls=%v", psChecks, volumeChecks, runner.calls)
	}

	for name, occupied := range map[string]func([]string) []byte{
		"container": func(call []string) []byte {
			if strings.Contains(strings.Join(call, " "), " ps --all --quiet") {
				return []byte("existing-container\n")
			}
			return nil
		},
		"volume": func(call []string) []byte {
			if strings.Contains(strings.Join(call, " "), "volume ls --quiet") {
				return []byte("cloud-clicker_postgres_data\n")
			}
			return nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			severed := &commandFixture{output: func(call []string) ([]byte, error) { return occupied(call), nil }}
			if err := dockerFixtureRuntime(severed, t.TempDir(), "").PreflightInstall(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
				t.Fatalf("occupied host accepted: %v", err)
			}
			for _, call := range severed.calls {
				if slices.Contains(call, "run") {
					t.Fatalf("occupied host reached candidate helper: %v", call)
				}
			}
		})
	}

	runner.calls = nil
	if err := runtime.StartInstall(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 2 || !strings.Contains(strings.Join(runner.calls[0], " "), "up --detach --wait postgres") ||
		!strings.Contains(strings.Join(runner.calls[1], " "), "--force-recreate gameserver caddy backup prometheus alertmanager node-exporter") {
		t.Fatalf("initial start sequence=%v", runner.calls)
	}
	runner.calls = nil
	if err := runtime.AbortInstall(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(runner.calls[0], " ")
	for _, required := range []string{"down", "--volumes", "--remove-orphans"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("install cleanup missing %q: %s", required, joined)
		}
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
	recipient, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	// writeEnvelope replaces the host backup with a real authenticated
	// envelope, because release now re-reads the host bytes it will roll back to.
	writeEnvelope := func(preUpgrade bool) deploymentbackup.Header {
		t.Helper()
		_ = os.Remove(filepath.Join(target, "20260822T180000Z-acde00000001.ccbackup"))
		written, _, err := deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target, BackupID: "20260822T180000Z-acde00000001",
			ServerID: "server-1", ReleaseManifestSHA256: bundle.ManifestSHA256, EpochID: bundle.Manifest.EpochID,
			StartedAt: time.Date(2026, 8, 22, 18, 0, 0, 0, time.UTC), Now: func() time.Time { return time.Date(2026, 8, 22, 18, 1, 0, 0, time.UTC) },
			Recipient: recipient.Recipient().String(), Dump: bytes.NewReader([]byte("PGDMP\x01release fixture")), PreUpgrade: preUpgrade})
		if err != nil {
			t.Fatal(err)
		}
		return written
	}
	header := writeEnvelope(true)
	hostBackup := filepath.Join(target, header.BackupID+".ccbackup")
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
	recoveryHeader := writeEnvelope(false)
	runner.output = func(call []string) ([]byte, error) {
		if slices.Contains(call, "create") {
			return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(hostBackup), "header": recoveryHeader})
		}
		if slices.Contains(call, "restore") {
			return json.Marshal(map[string]any{"status": "restored", "header": recoveryHeader})
		}
		return nil, errors.New("unexpected command")
	}
	recovery, err := runtime.CreateRecoveryBackup(context.Background(), bundle)
	if err != nil || recovery != backup {
		t.Fatalf("recovery backup=%+v err=%v", recovery, err)
	}
	if err := runtime.RestoreRecoveryBackup(context.Background(), bundle, recovery); err != nil {
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

	severed.output = func(call []string) ([]byte, error) {
		if slices.Contains(call, "restore") {
			return json.Marshal(map[string]any{"status": "restored", "header": header})
		}
		return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(hostBackup), "header": header})
	}
	badRuntime = dockerFixtureRuntime(&severed, target, identity)
	if _, err := badRuntime.CreateRecoveryBackup(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("pre-upgrade backup accepted as scheduled recovery backup: %v", err)
	}
	if err := badRuntime.RestoreRecoveryBackup(context.Background(), bundle, backup); !errors.Is(err, ErrInvalid) {
		t.Fatalf("pre-upgrade backup accepted by recovery restore: %v", err)
	}

	// Rollback authority is the host envelope, not the container's report.
	header = writeEnvelope(true)
	reportThenCreate := func(reported deploymentbackup.Header) error {
		runner := &commandFixture{output: func([]string) ([]byte, error) {
			return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(hostBackup), "header": reported})
		}}
		_, err := dockerFixtureRuntime(runner, target, identity).CreatePreUpgradeBackup(context.Background(), bundle)
		return err
	}
	if err := reportThenCreate(header); err != nil {
		t.Fatalf("exact host envelope rejected: %v", err)
	}
	differing := header
	differing.CompletedAt = differing.CompletedAt.Add(time.Second)
	if err := reportThenCreate(differing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("reported header differing from the host envelope accepted: %v", err)
	}
	data, err := os.ReadFile(hostBackup)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(hostBackup, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reportThenCreate(header); !errors.Is(err, ErrInvalid) {
		t.Fatalf("corrupt host pre-upgrade payload accepted as rollback authority: %v", err)
	}
}

func TestDockerRuntimeObservesStrictPrivateRecoveryIdentity(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	digest := "sha256:" + strings.Repeat("c", 64)
	domain := deploymentbackup.RecoveryDomainIdentity{Rows: 1, SHA256: digest}
	identity := deploymentbackup.RecoveryIdentity{SchemaVersion: 2, DatabaseMigration: bundle.Manifest.DatabaseMigration, VerifiedRunRows: 1,
		Database: domain, Player: domain, Founder: domain, Company: domain, Events: domain, Board: domain, Epoch: domain}
	runner := &commandFixture{output: func(call []string) ([]byte, error) {
		if !slices.Contains(call, "recovery-identity") || !slices.Contains(call, "--database-url-file=/run/secrets/database-url") {
			return nil, errors.New("wrong private identity command")
		}
		return json.Marshal(map[string]any{"status": "observed", "identity": identity})
	}}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	observed, err := runtime.InspectRecoveryIdentity(context.Background(), bundle)
	if err != nil || observed != identity {
		t.Fatalf("identity=%+v err=%v", observed, err)
	}
	identity.DatabaseMigration--
	if _, err := runtime.InspectRecoveryIdentity(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong-migration recovery identity accepted: %v", err)
	}
	runner.output = func([]string) ([]byte, error) {
		data, _ := json.Marshal(map[string]any{"status": "observed", "identity": observed})
		return append(data, []byte("\n{}")...), nil
	}
	if _, err := runtime.InspectRecoveryIdentity(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("trailing recovery identity output accepted: %v", err)
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

func TestDockerRuntimeRecoveryCoreExcludesUnobservedBackupWriter(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	runner := &commandFixture{}
	runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
	runtime.Client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if err := runtime.StartRecoveryCore(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(runner.calls[0], " ")
	for _, required := range []string{"--no-deps", "--force-recreate", "gameserver", "caddy"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("recovery core missing %q: %s", required, joined)
		}
	}
	for _, excluded := range []string{"backup", "prometheus", "alertmanager", "node-exporter"} {
		if strings.Contains(joined, " "+excluded) {
			t.Fatalf("recovery core started %q: %s", excluded, joined)
		}
	}
}

func dockerFixtureRuntime(runner CommandRunner, target, identity string) DockerRuntime {
	return DockerRuntime{Runner: runner, PublicOrigin: "https://game.example", ReceiverHealthURL: "http://alertmanager:9093/-/healthy",
		BackupTarget: target, MetricsDirectory: target, AgeRecipient: "age1fixture",
		AgeIdentityFile: identity, ServerID: "server-1", DrainTimeout: 20 * time.Second,
		RotationLedgerPath: filepath.Join(target, "rotation-ledger.jsonl")}
}

func dockerFixtureBundle(t *testing.T) Bundle {
	t.Helper()
	root := t.TempDir()
	return Bundle{Root: root, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), Manifest: releasepackage.ReleaseManifest{
		ReleaseVersion: "1.0.0", DatabaseMigration: 74, EpochID: 8, ConstantsHash: "sha256:" + strings.Repeat("b", 64), Images: []releasepackage.Image{
			{Name: "alertmanager", Reference: "alertmanager:v1@sha256:" + strings.Repeat("1", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("7", 64)},
			{Name: "caddy", Reference: "caddy:v1@sha256:" + strings.Repeat("1", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("4", 64)},
			{Name: "gameserver", Reference: "sha256:" + strings.Repeat("2", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("2", 64)},
			{Name: "node-exporter", Reference: "node-exporter:v1@sha256:" + strings.Repeat("6", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("8", 64)},
			{Name: "postgres", Reference: "postgres:v1@sha256:" + strings.Repeat("3", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("5", 64)},
			{Name: "prometheus", Reference: "prometheus:v1@sha256:" + strings.Repeat("9", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("a", 64)},
		}}}
}

func TestDockerRuntimeVerifiesRestoreInputsWithoutRuntimeCommands(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	target := t.TempDir()
	identityPath := filepath.Join(t.TempDir(), "identity")
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 22, 18, 1, 0, 0, time.UTC)
	create := func(id string, manifest string, preUpgrade bool) BackupReference {
		t.Helper()
		header, path, err := deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target, BackupID: id, ServerID: "server-1",
			ReleaseManifestSHA256: manifest, EpochID: bundle.Manifest.EpochID, StartedAt: now.Add(-time.Minute),
			Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
			Dump: bytes.NewReader([]byte("PGDMP\x01restore-input fixture")), PreUpgrade: preUpgrade})
		if err != nil {
			t.Fatal(err)
		}
		return BackupReference{ID: header.BackupID, Path: path}
	}
	valid := create("20260822T180000Z-acde00000001", bundle.ManifestSHA256, true)
	runner := &commandFixture{}
	runtime := dockerFixtureRuntime(runner, target, identityPath)
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, valid); err != nil {
		t.Fatalf("valid restore inputs rejected: %v", err)
	}

	missing := BackupReference{ID: "20260822T180000Z-acde00000009", Path: filepath.Join(target, "20260822T180000Z-acde00000009.ccbackup")}
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, missing); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing backup accepted: %v", err)
	}
	wrongManifest := create("20260822T180000Z-acde00000002", "sha256:"+strings.Repeat("e", 64), true)
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, wrongManifest); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong-manifest backup accepted: %v", err)
	}
	recoveryClass := create("20260822T180000Z-acde00000003", bundle.ManifestSHA256, false)
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, recoveryClass); !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-pre-upgrade backup accepted as rollback input: %v", err)
	}
	corrupt := create("20260822T180000Z-acde00000004", bundle.ManifestSHA256, true)
	data, err := os.ReadFile(corrupt.Path)
	if err != nil {
		t.Fatal(err)
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(corrupt.Path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, corrupt); !errors.Is(err, ErrInvalid) {
		t.Fatalf("corrupt backup payload accepted: %v", err)
	}
	if err := os.Chmod(identityPath, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, valid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("group-readable identity accepted: %v", err)
	}
	if err := os.Remove(identityPath); err != nil {
		t.Fatal(err)
	}
	if err := runtime.VerifyRestoreInputs(context.Background(), bundle, valid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing identity accepted: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("restore-input verification issued runtime commands: %v", runner.calls)
	}
}

func TestDockerRuntimeComposesRotationOverlayOnlyForAnOpenPair(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	target := t.TempDir()
	ledger := filepath.Join(target, "rotation-ledger.jsonl")
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	stopArgs := func() ([]string, error) {
		t.Helper()
		runner := &commandFixture{}
		err := dockerFixtureRuntime(runner, target, "").StopFailed(context.Background(), bundle)
		if len(runner.calls) == 0 {
			return nil, err
		}
		return runner.calls[0], err
	}
	overlay := bundle.Root + "/compose.rotation.yml"
	if call, err := stopArgs(); err != nil || slices.Contains(call, overlay) {
		t.Fatalf("no-ledger composition call=%v err=%v", call, err)
	}
	if err := ActivateRotation(ledger, FamilyJWT, "jwt-2", "jwt-1", "operator-1", now); err != nil {
		t.Fatal(err)
	}
	if call, err := stopArgs(); !errors.Is(err, ErrInvalid) || call != nil {
		t.Fatalf("single open overlap composed without its previous key: call=%v err=%v", call, err)
	}
	if err := ActivateRotation(ledger, FamilyBootstrap, "boot-2", "boot-1", "operator-1", now); err != nil {
		t.Fatal(err)
	}
	call, err := stopArgs()
	if err != nil || !slices.Contains(call, overlay) || slices.Index(call, overlay) > slices.Index(call, "down") {
		t.Fatalf("open JWT+bootstrap overlap omitted rotation overlay: call=%v err=%v", call, err)
	}
	if err := RemovePrevious(ledger, FamilyJWT, "jwt-2", "jwt-1", "operator-1", now.Add(JWTOverlap)); err != nil {
		t.Fatal(err)
	}
	if call, err := stopArgs(); !errors.Is(err, ErrInvalid) || call != nil {
		t.Fatalf("remaining bootstrap overlap composed without its previous key: call=%v err=%v", call, err)
	}
	if err := RemovePrevious(ledger, FamilyBootstrap, "boot-2", "boot-1", "operator-1", now.Add(BootstrapOverlap)); err != nil {
		t.Fatal(err)
	}
	if call, err := stopArgs(); err != nil || slices.Contains(call, overlay) {
		t.Fatalf("closed overlaps still composed overlay: call=%v err=%v", call, err)
	}
	relative := dockerFixtureRuntime(&commandFixture{}, target, "")
	relative.RotationLedgerPath = "rotation-ledger.jsonl"
	if err := relative.StopFailed(context.Background(), bundle); !errors.Is(err, ErrInvalid) {
		t.Fatalf("relative rotation ledger accepted: %v", err)
	}
}

func TestDeriveDrainEvidenceRequiresEveryObservation(t *testing.T) {
	good := func() DrainEvidence {
		return deriveDrainEvidence(true, true, true, nil, nil, "0\n", 5*time.Second, 20*time.Second)
	}
	if !good().Valid() {
		t.Fatalf("complete drain rejected: %+v", good())
	}
	for name, evidence := range map[string]DrainEvidence{
		"readiness never withdrawn": deriveDrainEvidence(false, true, true, nil, nil, "0", time.Second, 20*time.Second),
		"no courtesy frame":         deriveDrainEvidence(true, false, true, nil, nil, "0", time.Second, 20*time.Second),
		"socket left open":          deriveDrainEvidence(true, true, false, nil, nil, "0", time.Second, 20*time.Second),
		"stop failed":               deriveDrainEvidence(true, true, true, errors.New("stop timed out"), nil, "0", time.Second, 20*time.Second),
		"exit not inspected":        deriveDrainEvidence(true, true, true, nil, errors.New("inspect failed"), "0", time.Second, 20*time.Second),
		"nonzero exit":              deriveDrainEvidence(true, true, true, nil, nil, "137", time.Second, 20*time.Second),
		"empty exit code":           deriveDrainEvidence(true, true, true, nil, nil, "", time.Second, 20*time.Second),
		"over the drain bound":      deriveDrainEvidence(true, true, true, nil, nil, "0", 21*time.Second, 20*time.Second),
	} {
		if evidence.Valid() {
			t.Fatalf("%s accepted as a complete drain: %+v", name, evidence)
		}
	}
}

func TestReadinessDownRequiresTheGameserverDrainingAnswer(t *testing.T) {
	status := http.StatusBadGateway
	var lock sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		response.WriteHeader(status)
	}))
	wait := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		return waitHTTPState(ctx, server.Client(), server.URL+"/readyz", false)
	}
	for _, proxyStatus := range []int{http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusNoContent} {
		lock.Lock()
		status = proxyStatus
		lock.Unlock()
		if wait() {
			t.Fatalf("status %d counted as the gameserver withdrawing readiness", proxyStatus)
		}
	}
	lock.Lock()
	status = http.StatusServiceUnavailable
	lock.Unlock()
	if !wait() {
		t.Fatal("gameserver draining 503 not observed")
	}
	url := server.URL
	server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if waitHTTPState(ctx, http.DefaultClient, url+"/readyz", false) {
		t.Fatal("a vanished upstream counted as readiness withdrawal")
	}
}

func TestDockerRuntimeVerifyIdentityBindsEveryInspectedIdentity(t *testing.T) {
	bundle := dockerFixtureBundle(t)
	content := filepath.Join(bundle.Root, "content")
	if _, err := releasepackage.StageRuntimeContent(filepath.Join("..", ".."), content); err != nil {
		t.Fatal(err)
	}
	epoch, err := epochseed.Load(content)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Manifest.ConstantsHash, bundle.Manifest.EpochID = epoch.Hash, epoch.Seed.CurrentEpochID
	exact := deploymentbackup.PostgresInspection{SchemaVersion: 1, DatabaseBytes: 4096, DatabaseMigration: bundle.Manifest.DatabaseMigration,
		EpochID: bundle.Manifest.EpochID, ConstantsHash: bundle.Manifest.ConstantsHash, ArtifactsVerified: len(epoch.Artifacts)}
	verify := func(inspection deploymentbackup.PostgresInspection) error {
		runner := &commandFixture{output: func(call []string) ([]byte, error) {
			if slices.Contains(call, "inspect") {
				return json.Marshal(map[string]any{"status": "inspected", "inspection": inspection})
			}
			return nil, errors.New("unexpected command")
		}}
		return dockerFixtureRuntime(runner, t.TempDir(), "").VerifyIdentity(context.Background(), bundle)
	}
	if err := verify(exact); err != nil {
		t.Fatalf("exact identity rejected: %v", err)
	}
	for name, mutate := range map[string]func(*deploymentbackup.PostgresInspection){
		"migration": func(value *deploymentbackup.PostgresInspection) { value.DatabaseMigration-- },
		"epoch":     func(value *deploymentbackup.PostgresInspection) { value.EpochID++ },
		"constants hash": func(value *deploymentbackup.PostgresInspection) {
			value.ConstantsHash = "sha256:" + strings.Repeat("e", 64)
		},
		"artifact count": func(value *deploymentbackup.PostgresInspection) { value.ArtifactsVerified-- },
	} {
		inspection := exact
		mutate(&inspection)
		if err := verify(inspection); !errors.Is(err, ErrInvalid) {
			t.Fatalf("wrong %s accepted: %v", name, err)
		}
	}
}
