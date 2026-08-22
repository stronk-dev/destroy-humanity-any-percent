package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"cloud-clicker/server/deploymentbackup"
)

const scheduleInterval = 6 * time.Hour

type createFlags struct {
	target, databaseURLFile, manifest, epoch, recipient, serverID string
	preUpgrade                                                    bool
}

func main() {
	if len(os.Args) < 2 {
		fail("usage: deployment-backup <create|schedule|restore|retention>")
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
	default:
		err = errors.New("unknown deployment backup command")
	}
	if err != nil {
		fail(err.Error())
	}
}

func runCreate(ctx context.Context, args []string) error {
	flags, err := parseCreateFlags("create", args)
	if err != nil {
		return err
	}
	header, path, err := createOnce(ctx, flags, time.Now().UTC())
	if err != nil {
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
			return err
		}
		paths, err := backupPaths(flags.target)
		if err != nil {
			return err
		}
		plan, err := deploymentbackup.ApplyRetention(paths, time.Now().UTC())
		if err != nil {
			return err
		}
		if err := emit(map[string]any{"status": "completed", "backup": path, "header": header, "retention": plan}); err != nil {
			return err
		}
		next := started.Add(scheduleInterval)
		delay := time.Until(next)
		if delay <= 0 {
			return errors.New("backup exceeded its six-hour schedule interval")
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
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *backup == "" || *manifest == "" || *identity == "" || *targetURL == "" {
		return errors.New("restore requires backup, release-manifest, identity-file and target-database-url-file")
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
	set.BoolVar(&values.preUpgrade, "pre-upgrade", false, "protect as an unresolved pre-upgrade backup")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || values.target == "" || values.databaseURLFile == "" || values.manifest == "" || values.epoch == "" || values.recipient == "" || values.serverID == "" {
		return createFlags{}, errors.New("backup requires target, database-url-file, release-manifest, epoch, age-recipient and server-id")
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

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
