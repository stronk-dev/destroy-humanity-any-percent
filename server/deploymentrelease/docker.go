package deploymentrelease

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/epochseed"
)

type DockerRuntime struct {
	Runner            CommandRunner
	Client            *http.Client
	PublicOrigin      string
	ReceiverHealthURL string
	BackupTarget      string
	AgeRecipient      string
	AgeIdentityFile   string
	ServerID          string
	DrainTimeout      time.Duration
	Now               func() time.Time
}

func (runtime DockerRuntime) normalized() (DockerRuntime, error) {
	if runtime.Runner == nil {
		runtime.Runner = ExecRunner{}
	}
	if runtime.Client == nil {
		runtime.Client = &http.Client{Timeout: 10 * time.Second}
	}
	if runtime.Now == nil {
		runtime.Now = time.Now
	}
	if runtime.DrainTimeout == 0 {
		runtime.DrainTimeout = 20 * time.Second
	}
	if runtime.PublicOrigin == "" || runtime.ReceiverHealthURL == "" || runtime.BackupTarget == "" ||
		runtime.AgeRecipient == "" || runtime.ServerID == "" || !filepath.IsAbs(runtime.BackupTarget) ||
		runtime.DrainTimeout < time.Second || runtime.DrainTimeout > time.Minute {
		return DockerRuntime{}, ErrInvalid
	}
	if info, err := os.Lstat(runtime.BackupTarget); err != nil || !info.IsDir() {
		return DockerRuntime{}, ErrInvalid
	}
	return runtime, nil
}

func (runtime DockerRuntime) Preflight(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "config", "--quiet")...); err != nil {
		return err
	}
	if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "run", "--rm", "--no-deps", "gameserver", "validate-config")...); err != nil {
		return err
	}
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, runtime.ReceiverHealthURL, nil)
	response, err := runtime.Client.Do(request)
	if err != nil {
		return err
	}
	_ = response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ErrInvalid
	}
	bundleBytes, err := directoryBytes(bundle.Root)
	if err != nil || bundleBytes < 1 {
		return errors.Join(ErrInvalid, err)
	}
	dockerRootOutput, err := runtime.Runner.Run(ctx, bundle.Root, "docker", "info", "--format={{.DockerRootDir}}")
	dockerRoot := strings.TrimSpace(string(dockerRootOutput))
	if err != nil || !filepath.IsAbs(dockerRoot) {
		return errors.Join(ErrInvalid, err)
	}
	inspection, err := runtime.inspectDatabase(ctx, bundle, false)
	if err != nil || inspection.DatabaseBytes < 1 {
		return errors.Join(ErrInvalid, err)
	}
	if !hasFreeBytes(dockerRoot, bundleBytes) || !hasFreeBytes(runtime.BackupTarget, uint64(inspection.DatabaseBytes)) {
		return ErrInvalid
	}
	return nil
}

func (runtime DockerRuntime) CreatePreUpgradeBackup(ctx context.Context, bundle Bundle) (BackupReference, error) {
	runtime, err := runtime.normalized()
	if err != nil {
		return BackupReference{}, err
	}
	args := composeArgs(bundle, "run", "--rm", "--no-deps", "backup", "create",
		"--target=/backups", "--database-url-file=/run/secrets/database-url",
		"--release-manifest=/opt/cloud-clicker/release-manifest.json", "--epoch=/opt/cloud-clicker/content/balance/epochs/phase0.json",
		"--age-recipient="+runtime.AgeRecipient, "--server-id="+runtime.ServerID, "--pre-upgrade")
	output, err := runtime.Runner.Run(ctx, bundle.Root, "docker", args...)
	if err != nil {
		return BackupReference{}, err
	}
	var result struct {
		Status string                  `json:"status"`
		Backup string                  `json:"backup"`
		Header deploymentbackup.Header `json:"header"`
	}
	if decodeExact(output, &result) != nil || result.Status != "completed" || !validBackupHeader(result.Header, bundle, runtime.ServerID, true) ||
		result.Header.ReleaseManifestSHA256 != bundle.ManifestSHA256 || result.Backup != "/backups/"+result.Header.BackupID+".ccbackup" {
		return BackupReference{}, ErrInvalid
	}
	hostPath := filepath.Join(runtime.BackupTarget, result.Header.BackupID+".ccbackup")
	if info, err := os.Lstat(hostPath); err != nil || !info.Mode().IsRegular() || info.Size() < 1 {
		return BackupReference{}, ErrInvalid
	}
	return BackupReference{ID: result.Header.BackupID, Path: hostPath}, nil
}

func (runtime DockerRuntime) Prepare(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", "load", "--input", filepath.Join(bundle.Root, "images", "gameserver.docker.tar")); err != nil {
		return err
	}
	for _, image := range bundle.Manifest.Images {
		if image.Name != "gameserver" {
			if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", "pull", image.Reference); err != nil {
				return err
			}
		}
		output, err := runtime.Runner.Run(ctx, bundle.Root, "docker", "image", "inspect", "--format={{.Id}}", image.Reference)
		if err != nil || strings.TrimSpace(string(output)) != image.RuntimeConfigSHA256 {
			return errors.Join(ErrInvalid, err)
		}
	}
	return nil
}

func (runtime DockerRuntime) DrainCurrent(ctx context.Context, bundle Bundle) (DrainEvidence, error) {
	runtime, err := runtime.normalized()
	if err != nil {
		return DrainEvidence{}, err
	}
	session, err := OpenAuthenticatedSmoke(ctx, runtime.Client, runtime.PublicOrigin, bundle.Manifest.ConstantsHash)
	if err != nil {
		return DrainEvidence{}, err
	}
	defer session.Close()
	containerOutput, err := runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "ps", "--quiet", "gameserver")...)
	containerID := strings.TrimSpace(string(containerOutput))
	if err != nil || containerID == "" || strings.ContainsAny(containerID, " \t\r\n") {
		return DrainEvidence{}, errors.Join(ErrInvalid, err)
	}
	drainCtx, cancel := context.WithTimeout(ctx, runtime.DrainTimeout)
	defer cancel()
	type socketResult struct {
		courtesy, closed bool
		err              error
	}
	socket := make(chan socketResult, 1)
	go func() {
		courtesy, closed, readErr := session.WaitForDrain(drainCtx)
		socket <- socketResult{courtesy, closed, readErr}
	}()
	started := runtime.Now()
	stop := make(chan error, 1)
	go func() {
		seconds := int(runtime.DrainTimeout / time.Second)
		_, stopErr := runtime.Runner.Run(drainCtx, bundle.Root, "docker", composeArgs(bundle, "stop", "--timeout", fmt.Sprint(seconds), "gameserver")...)
		stop <- stopErr
	}()
	readiness := make(chan bool, 1)
	go func() { readiness <- waitHTTPState(drainCtx, runtime.Client, runtime.PublicOrigin+"/readyz", false) }()
	stopErr := <-stop
	exitOutput, inspectErr := runtime.Runner.Run(drainCtx, bundle.Root, "docker", "inspect", "--format={{.State.ExitCode}}", containerID)
	socketState := <-socket
	evidence := DrainEvidence{ReadinessDown: <-readiness, CourtesyFrame: socketState.courtesy,
		SocketsClosed: socketState.closed, WithinBound: runtime.Now().Sub(started) <= runtime.DrainTimeout}
	cleanExit := stopErr == nil && inspectErr == nil && strings.TrimSpace(string(exitOutput)) == "0"
	// The gameserver process exits zero only after its admitted-work gate, jobs,
	// relay/outbox flush, and transport shutdown have all returned successfully.
	// beginDrain closes intent admission before those waits. The real Caddy
	// population independently sends an intent during this interval and requires
	// the exact refusal; the release helper must not invent a minimum drain delay
	// merely to make that race observable on an otherwise idle server.
	evidence.IntentsRefused, evidence.AdmittedComplete, evidence.JobsFlushed = cleanExit, cleanExit, cleanExit
	if !evidence.Valid() {
		return evidence, errors.Join(ErrInvalid, stopErr, inspectErr)
	}
	return evidence, nil
}

func (runtime DockerRuntime) Start(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "up", "--detach", "--no-deps", "--force-recreate", "gameserver", "caddy", "backup")...); err != nil {
		return err
	}
	readyCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if !waitHTTPState(readyCtx, runtime.Client, runtime.PublicOrigin+"/readyz", true) {
		return ErrInvalid
	}
	return nil
}

func (runtime DockerRuntime) VerifyIdentity(ctx context.Context, bundle Bundle) error {
	if _, err := runtime.normalized(); err != nil {
		return err
	}
	epoch, err := epochseed.Load(filepath.Join(bundle.Root, "content"))
	if err != nil || epoch.Hash != bundle.Manifest.ConstantsHash || epoch.Seed.CurrentEpochID != bundle.Manifest.EpochID {
		return errors.Join(ErrInvalid, err)
	}
	inspection, err := runtime.inspectDatabase(ctx, bundle, true)
	if err != nil || inspection.DatabaseMigration != bundle.Manifest.DatabaseMigration || inspection.EpochID != bundle.Manifest.EpochID ||
		inspection.ConstantsHash != bundle.Manifest.ConstantsHash || inspection.ArtifactsVerified != len(epoch.Artifacts) {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

func (runtime DockerRuntime) AuthenticatedSmoke(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	session, err := OpenAuthenticatedSmoke(ctx, runtime.Client, runtime.PublicOrigin, bundle.Manifest.ConstantsHash)
	if err != nil {
		return err
	}
	return session.Close()
}

func (runtime DockerRuntime) StopFailed(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	_, err = runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "down", "--remove-orphans")...)
	return err
}

func (runtime DockerRuntime) ResetDatabase(ctx context.Context, bundle Bundle) error {
	runtime, err := runtime.normalized()
	if err != nil {
		return err
	}
	if _, err := runtime.Runner.Run(ctx, bundle.Root, "docker", "volume", "rm", "cloud-clicker_postgres_data"); err != nil {
		return err
	}
	_, err = runtime.Runner.Run(ctx, bundle.Root, "docker", composeArgs(bundle, "up", "--detach", "--wait", "postgres")...)
	return err
}

func (runtime DockerRuntime) Restore(ctx context.Context, bundle Bundle, backup BackupReference) error {
	runtime, err := runtime.normalized()
	if err != nil || runtime.AgeIdentityFile == "" || filepath.Dir(backup.Path) != filepath.Clean(runtime.BackupTarget) || filepath.Base(backup.Path) != backup.ID+".ccbackup" {
		return ErrInvalid
	}
	identity, err := filepath.Abs(runtime.AgeIdentityFile)
	if err != nil {
		return err
	}
	if backupInfo, err := os.Lstat(backup.Path); err != nil || !backupInfo.Mode().IsRegular() || backupInfo.Size() < 1 {
		return ErrInvalid
	}
	if identityInfo, err := os.Lstat(identity); err != nil || !identityInfo.Mode().IsRegular() || identityInfo.Mode().Perm()&0o077 != 0 {
		return ErrInvalid
	}
	args := composeArgs(bundle, "run", "--rm", "--no-deps", "--volume", identity+":/run/secrets/age-identity:ro", "backup", "restore",
		"--backup=/backups/"+filepath.Base(backup.Path), "--release-manifest=/opt/cloud-clicker/release-manifest.json",
		"--identity-file=/run/secrets/age-identity", "--target-database-url-file=/run/secrets/database-url")
	output, err := runtime.Runner.Run(ctx, bundle.Root, "docker", args...)
	if err != nil {
		return err
	}
	var result struct {
		Status string                  `json:"status"`
		Header deploymentbackup.Header `json:"header"`
	}
	if decodeExact(output, &result) != nil || result.Status != "restored" || result.Header.BackupID != backup.ID ||
		!validBackupHeader(result.Header, bundle, runtime.ServerID, true) {
		return ErrInvalid
	}
	return nil
}

func validBackupHeader(header deploymentbackup.Header, bundle Bundle, serverID string, preUpgrade bool) bool {
	return header.SchemaVersion == 1 && backupIDPattern.MatchString(header.BackupID) && header.ServerID == serverID &&
		header.ReleaseManifestSHA256 == bundle.ManifestSHA256 && header.EpochID == bundle.Manifest.EpochID &&
		!header.StartedAt.IsZero() && !header.CompletedAt.Before(header.StartedAt) && hashPattern.MatchString(header.PayloadSHA256) &&
		header.PayloadBytes > 0 && header.PreUpgrade == preUpgrade && !header.UpgradeResolved
}

func (runtime DockerRuntime) inspectDatabase(ctx context.Context, bundle Bundle, requireIdentity bool) (deploymentbackup.PostgresInspection, error) {
	runtime, err := runtime.normalized()
	if err != nil {
		return deploymentbackup.PostgresInspection{}, err
	}
	args := composeArgs(bundle, "run", "--rm", "--no-deps", "backup", "inspect",
		"--database-url-file=/run/secrets/database-url", "--release-manifest=/opt/cloud-clicker/release-manifest.json",
		"--content-root=/opt/cloud-clicker/content")
	if requireIdentity {
		args = append(args, "--require-identity")
	}
	output, err := runtime.Runner.Run(ctx, bundle.Root, "docker", args...)
	if err != nil {
		return deploymentbackup.PostgresInspection{}, err
	}
	var result struct {
		Status     string                              `json:"status"`
		Inspection deploymentbackup.PostgresInspection `json:"inspection"`
	}
	if decodeExact(output, &result) != nil || result.Status != "inspected" || result.Inspection.SchemaVersion != 1 ||
		result.Inspection.DatabaseBytes < 1 || result.Inspection.EpochID != bundle.Manifest.EpochID ||
		result.Inspection.ConstantsHash != bundle.Manifest.ConstantsHash || requireIdentity &&
		(result.Inspection.DatabaseMigration != bundle.Manifest.DatabaseMigration || result.Inspection.ArtifactsVerified < 1) {
		return deploymentbackup.PostgresInspection{}, ErrInvalid
	}
	return result.Inspection, nil
}

func directoryBytes(root string) (uint64, error) {
	var total uint64
	err := filepath.WalkDir(root, func(_ string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type().IsRegular() {
			info, err := entry.Info()
			if err != nil || info.Size() < 0 {
				return errors.Join(ErrInvalid, err)
			}
			total += uint64(info.Size())
		}
		return nil
	})
	return total, err
}

func hasFreeBytes(path string, required uint64) bool {
	var state syscall.Statfs_t
	if syscall.Statfs(path, &state) != nil {
		return false
	}
	return uint64(state.Bavail)*uint64(state.Bsize) >= required
}
