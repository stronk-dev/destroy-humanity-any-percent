package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/operations"
)

const scheduleInterval = 6 * time.Hour

type createFlags struct {
	target, databaseURLFile, manifest, epoch, recipient, serverID, metricsDirectory string
	preUpgrade                                                                      bool
}

func main() {
	if len(os.Args) < 2 {
		fail("unknown", deploymentbackup.ErrInvalid)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	var err error
	switch os.Args[1] {
	case "create":
		err = runCreate(ctx, os.Args[2:])
	case "schedule":
		err = runSchedule(ctx, os.Args[2:])
	case "restore":
		err = runRestore(ctx, os.Args[2:])
	case "retention":
		err = runRetention(os.Args[2:])
	case "inspect":
		err = runInspect(ctx, os.Args[2:])
	case "recovery-identity":
		err = runRecoveryIdentity(ctx, os.Args[2:])
	default:
		err = errors.New("unknown deployment backup command")
	}
	if err != nil {
		fail(os.Args[1], err)
	}
}

func runRecoveryIdentity(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("recovery-identity", flag.ContinueOnError)
	databaseURL := set.String("database-url-file", "", "Postgres URL secret file")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *databaseURL == "" {
		return errors.New("recovery-identity requires database-url-file")
	}
	identity, err := deploymentbackup.InspectRecoveryIdentity(ctx, deploymentbackup.RecoveryIdentityInput{DatabaseURLFile: *databaseURL})
	if err != nil {
		return err
	}
	return emit(map[string]any{"status": "observed", "identity": identity})
}

func runInspect(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("inspect", flag.ContinueOnError)
	databaseURL := set.String("database-url-file", "", "Postgres URL secret file")
	manifest := set.String("release-manifest", "", "release-manifest.json")
	content := set.String("content-root", "", "release content root")
	requireIdentity := set.Bool("require-identity", false, "require exact migration, epoch and artifact identity")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *databaseURL == "" || *manifest == "" || *content == "" {
		return errors.New("inspect requires database-url-file, release-manifest and content-root")
	}
	inspection, err := deploymentbackup.InspectPostgres(ctx, deploymentbackup.PostgresInspectionInput{
		DatabaseURLFile: *databaseURL, ReleaseManifest: *manifest, ContentRoot: *content, RequireIdentity: *requireIdentity,
	})
	if err != nil {
		return err
	}
	return emit(map[string]any{"status": "inspected", "inspection": inspection})
}

func runCreate(ctx context.Context, args []string) error {
	flags, err := parseCreateFlags("create", args)
	if err != nil {
		return err
	}
	header, path, err := createOnce(ctx, flags, time.Now().UTC())
	if err != nil {
		return errors.Join(err, operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
	}
	if err := operations.RecordOperation(flags.metricsDirectory, "backup", "success", header.CompletedAt); err != nil {
		return err
	}
	return emit(map[string]any{"status": "completed", "backup": path, "header": header})
}

func runSchedule(ctx context.Context, args []string) error {
	flags, err := parseCreateFlags("schedule", args)
	if err != nil {
		return err
	}
	if flags.preUpgrade {
		return errors.New("scheduled backups cannot be marked pre-upgrade")
	}
	for {
		started := time.Now().UTC()
		header, path, err := createOnce(ctx, flags, started)
		if err != nil {
			return errors.Join(err, operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
		}
		paths, err := backupPaths(flags.target)
		if err != nil {
			return errors.Join(err, operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
		}
		if err := operations.RecordOperation(flags.metricsDirectory, "backup", "success", header.CompletedAt); err != nil {
			return err
		}
		plan, err := deploymentbackup.ApplyRetention(paths, time.Now().UTC())
		if err != nil {
			return errors.Join(err, operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
		}
		if err := emit(map[string]any{"status": "completed", "backup": path, "header": header, "retention": plan}); err != nil {
			return errors.Join(err, operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
		}
		next := started.Add(scheduleInterval)
		delay := time.Until(next)
		if delay <= 0 {
			return errors.Join(errors.New("backup exceeded its six-hour schedule interval"), operations.RecordOperation(flags.metricsDirectory, "backup", "failure", time.Now().UTC()))
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func runRestore(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("restore", flag.ContinueOnError)
	backup := set.String("backup", "", "encrypted .ccbackup path")
	manifest := set.String("release-manifest", "", "expected release-manifest.json")
	identity := set.String("identity-file", "", "restore-only age X25519 identity file")
	targetURL := set.String("target-database-url-file", "", "clean target Postgres URL secret file")
	metricsDirectory := set.String("metrics-dir", "", "node-exporter textfile directory")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *backup == "" || *manifest == "" || *identity == "" || *targetURL == "" || *metricsDirectory == "" {
		return errors.New("restore requires backup, release-manifest, identity-file, target-database-url-file and metrics-dir")
	}
	manifestBytes, err := os.ReadFile(*manifest)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(manifestBytes)
	header, err := deploymentbackup.RestorePostgresBackup(ctx, deploymentbackup.PostgresRestoreInput{
		BackupPath: *backup, ExpectedManifestSHA256: "sha256:" + hex.EncodeToString(sum[:]),
		IdentityFile: *identity, TargetDatabaseURLFile: *targetURL,
	})
	if err != nil {
		return errors.Join(err, operations.RecordOperation(*metricsDirectory, "restore", "failure", time.Now().UTC()))
	}
	if err := operations.RecordOperation(*metricsDirectory, "restore", "success", time.Now().UTC()); err != nil {
		return err
	}
	return emit(map[string]any{"status": "restored", "header": header})
}

func runRetention(args []string) error {
	set := flag.NewFlagSet("retention", flag.ContinueOnError)
	target := set.String("target", "", "dedicated backup target")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *target == "" {
		return errors.New("retention requires target")
	}
	paths, err := backupPaths(*target)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	status := deploymentbackup.EvaluateSchedule(paths, now)
	if status.Missing || status.Late || len(status.Invalid) != 0 {
		_ = emit(map[string]any{"status": "blocked", "schedule": status})
		return errors.New("backup population is missing, late or invalid")
	}
	plan, err := deploymentbackup.ApplyRetention(paths, now)
	if err != nil {
		return err
	}
	return emit(map[string]any{"status": "completed", "schedule": status, "retention": plan})
}

func parseCreateFlags(name string, args []string) (createFlags, error) {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	var values createFlags
	set.StringVar(&values.target, "target", "", "dedicated off-host backup target")
	set.StringVar(&values.databaseURLFile, "database-url-file", "", "Postgres URL secret file")
	set.StringVar(&values.manifest, "release-manifest", "", "release-manifest.json")
	set.StringVar(&values.epoch, "epoch", "", "epoch declaration")
	set.StringVar(&values.recipient, "age-recipient", "", "public age X25519 recipient")
	set.StringVar(&values.serverID, "server-id", "", "source server UUID")
	set.StringVar(&values.metricsDirectory, "metrics-dir", "", "node-exporter textfile directory")
	set.BoolVar(&values.preUpgrade, "pre-upgrade", false, "protect as an unresolved pre-upgrade backup")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || values.target == "" || values.databaseURLFile == "" || values.manifest == "" || values.epoch == "" || values.recipient == "" || values.serverID == "" || values.metricsDirectory == "" {
		return createFlags{}, errors.New("backup requires target, database-url-file, release-manifest, epoch, age-recipient, server-id and metrics-dir")
	}
	return values, nil
}

func createOnce(ctx context.Context, flags createFlags, started time.Time) (deploymentbackup.Header, string, error) {
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return deploymentbackup.Header{}, "", err
	}
	return deploymentbackup.CreatePostgresBackup(ctx, deploymentbackup.PostgresBackupInput{
		Directory: flags.target, BackupID: started.Format("20060102T150405Z") + "-" + hex.EncodeToString(suffix),
		ServerID: flags.serverID, StartedAt: started, Now: time.Now, Recipient: flags.recipient,
		DatabaseURLFile: flags.databaseURLFile, ReleaseManifest: flags.manifest,
		EpochDeclaration: flags.epoch, PreUpgrade: flags.preUpgrade,
	})
}

func backupPaths(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(paths)
	return paths, nil
}

func emit(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

func fail(command string, err error) {
	command, class := boundedFailure(command, err)
	slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("deployment backup command failed", "command", command, "error_class", class)
	os.Exit(1)
}

func boundedFailure(command string, err error) (string, string) {
	if command != "create" && command != "schedule" && command != "restore" && command != "retention" && command != "inspect" && command != "recovery-identity" {
		command = "unknown"
	}
	class := "operation_failed"
	if errors.Is(err, deploymentbackup.ErrInvalid) || errors.Is(err, operations.ErrInvalid) {
		class = "invalid_input"
	}
	return command, class
}
