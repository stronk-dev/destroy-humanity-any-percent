package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/operations"
)

type runtimeFlags struct {
	state, operator, origin, receiver, backupTarget, recipient, identity, metricsDirectory string
}

func main() {
	if len(os.Args) < 2 {
		fail("unknown", deploymentrelease.ErrInvalid)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	var err error
	switch os.Args[1] {
	case "install":
		err = runInstall(ctx, os.Args[2:])
	case "release":
		err = runRelease(ctx, os.Args[2:])
	case "rollback":
		err = runRollback(ctx, os.Args[2:])
	case "rotation-activate":
		err = runRotation(os.Args[2:], false)
	case "rotation-remove":
		err = runRotation(os.Args[2:], true)
	default:
		err = errors.New("unknown deployment release command")
	}
	if err != nil {
		fail(os.Args[1], err)
	}
}

func runInstall(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("install", flag.ContinueOnError)
	bundle := set.String("bundle", "", "exact release bundle to install on an empty host")
	values := addRuntimeFlags(set, false)
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *bundle == "" {
		return errors.New("install requires bundle and operator runtime flags")
	}
	controller, err := controllerFromFlags(*values)
	if err != nil {
		return err
	}
	err = controller.Install(ctx, deploymentrelease.InstallRequest{Bundle: *bundle})
	result := "success"
	if err != nil {
		result = "failure"
	}
	return errors.Join(err, operations.RecordOperation(values.metricsDirectory, "release", result, time.Now().UTC()))
}

func runRelease(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("release", flag.ContinueOnError)
	current := set.String("current-bundle", "", "currently running exact release bundle")
	candidate := set.String("candidate-bundle", "", "candidate exact release bundle")
	values := addRuntimeFlags(set, false)
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *current == "" || *candidate == "" {
		return errors.New("release requires current-bundle, candidate-bundle, and operator runtime flags")
	}
	controller, err := controllerFromFlags(*values)
	if err != nil {
		return err
	}
	err = controller.Release(ctx, deploymentrelease.ReleaseRequest{CurrentBundle: *current, CandidateBundle: *candidate})
	result := "success"
	if err != nil {
		result = "failure"
	}
	return errors.Join(err, operations.RecordOperation(values.metricsDirectory, "release", result, time.Now().UTC()))
}

func runRollback(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("rollback", flag.ContinueOnError)
	failed := set.String("failed-bundle", "", "failed exact release bundle")
	previous := set.String("previous-bundle", "", "immediately previous exact release bundle")
	backupID := set.String("backup-id", "", "pre-upgrade backup identifier from the release ledger")
	backupPath := set.String("backup", "", "exact encrypted pre-upgrade backup path")
	values := addRuntimeFlags(set, true)
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *failed == "" || *previous == "" || *backupID == "" || *backupPath == "" {
		return errors.New("rollback requires failed-bundle, previous-bundle, backup-id, backup, and operator runtime flags")
	}
	if values.identity == "" {
		return errors.New("rollback requires an age-identity-file")
	}
	controller, err := controllerFromFlags(*values)
	if err != nil {
		return err
	}
	return controller.Rollback(ctx, deploymentrelease.RollbackRequest{FailedBundle: *failed, PreviousBundle: *previous,
		Backup: deploymentrelease.BackupReference{ID: *backupID, Path: *backupPath}})
}

func addRuntimeFlags(set *flag.FlagSet, restore bool) *runtimeFlags {
	values := &runtimeFlags{}
	set.StringVar(&values.state, "operator-state", "", "durable operator-state directory")
	set.StringVar(&values.operator, "operator", "", "operator identity recorded in ledgers")
	set.StringVar(&values.origin, "public-origin", "", "canonical Caddy origin")
	set.StringVar(&values.receiver, "receiver-health-url", "", "private alert-receiver health URL")
	set.StringVar(&values.backupTarget, "backup-target", "", "separately mounted backup target")
	set.StringVar(&values.recipient, "age-recipient", "", "public age X25519 recipient")
	set.StringVar(&values.metricsDirectory, "metrics-dir", "", "node-exporter textfile directory")
	if restore {
		set.StringVar(&values.identity, "age-identity-file", "", "restore-only age identity file")
	}
	return values
}

func controllerFromFlags(values runtimeFlags) (deploymentrelease.Controller, error) {
	if values.state == "" || values.operator == "" || values.origin == "" || values.receiver == "" || values.backupTarget == "" || values.recipient == "" || values.metricsDirectory == "" {
		return deploymentrelease.Controller{}, errors.New("operator-state, operator, public-origin, receiver-health-url, backup-target, age-recipient and metrics-dir are required")
	}
	state, err := filepath.Abs(values.state)
	if err != nil {
		return deploymentrelease.Controller{}, err
	}
	serverID := os.Getenv("CLOUD_CLICKER_SERVER_ID")
	if serverID == "" {
		return deploymentrelease.Controller{}, errors.New("CLOUD_CLICKER_SERVER_ID is required")
	}
	runtime := deploymentrelease.DockerRuntime{PublicOrigin: values.origin, ReceiverHealthURL: values.receiver,
		BackupTarget: values.backupTarget, MetricsDirectory: values.metricsDirectory, AgeRecipient: values.recipient,
		AgeIdentityFile: values.identity, ServerID: serverID, DrainTimeout: 20 * time.Second}
	return deploymentrelease.Controller{Runtime: runtime, LedgerPath: filepath.Join(state, "release-ledger.jsonl"), Operator: values.operator, Now: time.Now}, nil
}

func runRotation(args []string, remove bool) error {
	name := "rotation-activate"
	if remove {
		name = "rotation-remove"
	}
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	state := set.String("operator-state", "", "durable operator-state directory")
	family := set.String("family", "", "jwt, bootstrap, or cursor")
	current := set.String("current-id", "", "current key identifier")
	previous := set.String("previous-id", "", "previous key identifier")
	operator := set.String("operator", "", "operator identity")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *state == "" {
		return errors.New("rotation requires operator-state, family, current-id, previous-id and operator")
	}
	path := filepath.Join(*state, "rotation-ledger.jsonl")
	if remove {
		return deploymentrelease.RemovePrevious(path, deploymentrelease.KeyFamily(*family), *current, *previous, *operator, time.Now().UTC())
	}
	return deploymentrelease.ActivateRotation(path, deploymentrelease.KeyFamily(*family), *current, *previous, *operator, time.Now().UTC())
}

func fail(command string, err error) {
	command, class := boundedFailure(command, err)
	slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("deployment release command failed", "command", command, "error_class", class)
	os.Exit(1)
}

func boundedFailure(command string, err error) (string, string) {
	if command != "install" && command != "release" && command != "rollback" && command != "rotation-activate" && command != "rotation-remove" {
		command = "unknown"
	}
	class := "operation_failed"
	if errors.Is(err, deploymentrelease.ErrInvalid) || errors.Is(err, operations.ErrInvalid) {
		class = "invalid_input"
	}
	return command, class
}
