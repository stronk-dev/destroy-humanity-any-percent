package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloud-clicker/server/operations"
	"cloud-clicker/server/releasepackage"
)

func main() {
	if len(os.Args) < 2 {
		fail("unknown", operations.ErrInvalid)
	}
	var err error
	switch os.Args[1] {
	case "alert-test":
		err = runAlertTest(os.Args[2:])
	case "alert-observe":
		err = runAlertObserve(os.Args[2:])
	case "host-observe":
		err = runHostObserve(os.Args[2:])
	case "journal-observe":
		err = runJournalObserve(os.Args[2:])
	case "journal-render":
		err = runJournalRender(os.Args[2:])
	case "record":
		err = runRecord(os.Args[2:])
	default:
		err = operations.ErrInvalid
	}
	if err != nil {
		fail(os.Args[1], err)
	}
}

func runAlertObserve(args []string) error {
	set := flag.NewFlagSet("alert-observe", flag.ContinueOnError)
	bundle := set.String("bundle", "", "exact candidate release bundle")
	alertmanager := set.String("alertmanager-url", "", "private Alertmanager URL")
	receiverHealth := set.String("receiver-health-url", "", "configured receiver health URL")
	output := set.String("output", "", "exclusive alert-delivery observation")
	if set.Parse(args) != nil || set.NArg() != 0 || !filepath.IsAbs(*bundle) || !filepath.IsAbs(*output) {
		return operations.ErrInvalid
	}
	if err := releasepackage.ValidateBundle(*bundle); err != nil {
		return err
	}
	manifest, manifestBytes, err := releasepackage.LoadReleaseManifest(*bundle)
	if err != nil {
		return err
	}
	prometheusImage := ""
	for _, image := range manifest.Images {
		if image.Name == "prometheus" {
			prometheusImage = image.Reference
			break
		}
	}
	if prometheusImage == "" {
		return operations.ErrInvalid
	}
	rules := filepath.Join(*bundle, "operations")
	if _, err := commandOutput("docker", "run", "--rm", "--platform", "linux/amd64", "-v", rules+":/rules:ro", prometheusImage,
		"promtool", "test", "rules", "/rules/cloud-clicker-alerts.test.yml"); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client := &http.Client{Timeout: 5 * time.Second}
	manifestHash := fmt.Sprintf("sha256:%x", sha256.Sum256(manifestBytes))
	observation, err := operations.ObserveReleaseFloorAlertDelivery(ctx, client, *alertmanager, *receiverHealth, manifestHash, time.Now)
	if err != nil {
		return err
	}
	return operations.WriteAlertDeliveryObservation(*output, observation)
}

var commandOutput = func(name string, args ...string) ([]byte, error) { return exec.Command(name, args...).Output() }

func runHostObserve(args []string) error {
	set := flag.NewFlagSet("host-observe", flag.ContinueOnError)
	bundle := set.String("bundle", "", "exact current release bundle")
	metricsDirectory := set.String("metrics-dir", "", "node-exporter textfile directory")
	persistentPath := set.String("persistent-path", "", "Postgres volume host path")
	backupPath := set.String("backup-path", "", "off-host backup mount path")
	journalPath := set.String("journal-path", "", "persistent journal directory")
	journalBudgetPath := set.String("journal-budget-bytes-file", "", "rendered measured journal budget")
	if err := set.Parse(args); err != nil || set.NArg() != 0 {
		return operations.ErrInvalid
	}
	for _, path := range []string{*bundle, *metricsDirectory, *persistentPath, *backupPath, *journalPath, *journalBudgetPath} {
		if !filepath.IsAbs(path) {
			return operations.ErrInvalid
		}
	}
	container, err := commandOutput("docker", "compose", "--project-name", "cloud-clicker", "-f", filepath.Join(*bundle, "compose.yml"), "ps", "--quiet", "gameserver")
	containerID := strings.TrimSpace(string(container))
	if err != nil || containerID == "" || strings.ContainsAny(containerID, " \t\r\n") {
		return operations.ErrInvalid
	}
	restarts, err := commandOutput("docker", "inspect", "--format={{.RestartCount}}", containerID)
	restartCount, parseErr := strconv.ParseUint(strings.TrimSpace(string(restarts)), 10, 64)
	if err != nil || parseErr != nil {
		return errors.Join(operations.ErrInvalid, err, parseErr)
	}
	persistentUsage, err := filesystemUsage(*persistentPath)
	if err != nil {
		return err
	}
	backupUsage, err := filesystemUsage(*backupPath)
	if err != nil {
		return err
	}
	journalBytes, err := directoryUsage(*journalPath)
	if err != nil {
		return err
	}
	budgetBytes, err := os.ReadFile(*journalBudgetPath)
	if err != nil {
		return err
	}
	journalBudget, err := strconv.ParseUint(strings.TrimSpace(string(budgetBytes)), 10, 64)
	if err != nil {
		return operations.ErrInvalid
	}
	return operations.RecordHost(*metricsDirectory, persistentUsage, backupUsage, restartCount, journalBytes, journalBudget)
}

func filesystemUsage(path string) (float64, error) {
	var state syscall.Statfs_t
	if syscall.Statfs(path, &state) != nil || state.Blocks == 0 || state.Bavail > state.Blocks {
		return 0, operations.ErrInvalid
	}
	return float64(state.Blocks-state.Bavail) / float64(state.Blocks), nil
}

func directoryUsage(root string) (uint64, error) {
	var total uint64
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return operations.ErrInvalid
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() < 0 {
			return operations.ErrInvalid
		}
		total += uint64(info.Size())
		return nil
	})
	return total, err
}

func runAlertTest(args []string) error {
	set := flag.NewFlagSet("alert-test", flag.ContinueOnError)
	alertmanager := set.String("alertmanager-url", "", "private Alertmanager URL")
	receiverHealth := set.String("receiver-health-url", "", "configured receiver health URL")
	if err := set.Parse(args); err != nil || set.NArg() != 0 {
		return operations.ErrInvalid
	}
	client := &http.Client{Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := operations.VerifyAlertDelivery(ctx, client, *alertmanager, *receiverHealth)
	return err
}

func runJournalObserve(args []string) error {
	set := flag.NewFlagSet("journal-observe", flag.ContinueOnError)
	population := set.String("population", "", "predeclared workload population")
	startedText := set.String("started-at", "", "RFC3339 measurement start")
	journalMaxUse := set.Uint64("journal-max-use-bytes", 0, "configured SystemMaxUse bytes")
	journalPath := set.String("journal-path", "/var/log/journal", "persistent journal filesystem path")
	rawIP := set.Bool("raw-ip-enabled", false, "explicitly enabled separate security sink")
	securityRetention := set.Int64("security-retention-seconds", 0, "separate security sink retention")
	if err := set.Parse(args); err != nil || set.NArg() != 0 {
		return operations.ErrInvalid
	}
	started, err := time.Parse(time.RFC3339, *startedText)
	if err != nil || *population == "" || *journalMaxUse == 0 || !filepath.IsAbs(*journalPath) {
		return operations.ErrInvalid
	}
	completed := time.Now().UTC()
	if !completed.After(started) {
		return operations.ErrInvalid
	}
	command := exec.Command("journalctl", "--since=@"+strconv.FormatInt(started.Unix(), 10), "--until=@"+strconv.FormatInt(completed.Unix(), 10), "--output=json", "--no-pager",
		"SYSLOG_IDENTIFIER=cloud-clicker-caddy", "+", "SYSLOG_IDENTIFIER=cloud-clicker-gameserver", "+", "SYSLOG_IDENTIFIER=cloud-clicker-postgres", "+", "SYSLOG_IDENTIFIER=cloud-clicker-backup",
		"+", "SYSLOG_IDENTIFIER=cloud-clicker-prometheus", "+", "SYSLOG_IDENTIFIER=cloud-clicker-alertmanager", "+", "SYSLOG_IDENTIFIER=cloud-clicker-node-exporter")
	output, err := command.Output()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 64<<10), 2<<20)
	samples := 0
	for scanner.Scan() {
		var row map[string]any
		if json.Unmarshal(scanner.Bytes(), &row) != nil || len(row) == 0 {
			return operations.ErrInvalid
		}
		samples++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	duration := completed.Sub(started)
	peakPerDay := uint64((float64(len(output)) * float64(24*time.Hour)) / float64(duration))
	if peakPerDay < uint64(len(output)) {
		peakPerDay = uint64(len(output))
	}
	var filesystem syscall.Statfs_t
	if syscall.Statfs(*journalPath, &filesystem) != nil {
		return operations.ErrInvalid
	}
	observation := operations.JournalObservation{SchemaVersion: 1, Population: *population, StartedAt: started.UTC(), CompletedAt: completed,
		ObjectiveCompleted: true, Samples: samples, ObservedBytes: uint64(len(output)), PeakBytesPerDay: peakPerDay,
		FilesystemBytes: uint64(filesystem.Blocks) * uint64(filesystem.Bsize), JournalMaxUseBytes: *journalMaxUse,
		JournalRetentionSeconds: int64(operations.JournalRetention / time.Second), StorageAlertFraction: 0.8,
		RawIPEnabled: *rawIP, SecuritySinkSeparate: *rawIP, SecurityRetentionSeconds: *securityRetention}
	if err := operations.ValidateJournalObservation(observation); err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(observation)
}

func runJournalRender(args []string) error {
	set := flag.NewFlagSet("journal-render", flag.ContinueOnError)
	observationPath := set.String("observation", "", "validated journal observation JSON")
	templatePath := set.String("template", "", "journald policy template")
	output := set.String("output", "", "new absolute policy path")
	budgetOutput := set.String("budget-output", "", "new absolute measured budget path")
	if err := set.Parse(args); err != nil || set.NArg() != 0 || *observationPath == "" || *templatePath == "" || *output == "" || *budgetOutput == "" {
		return operations.ErrInvalid
	}
	observationBytes, err := os.ReadFile(*observationPath)
	if err != nil {
		return err
	}
	observation, err := operations.DecodeJournalObservation(observationBytes)
	if err != nil {
		return err
	}
	template, err := os.ReadFile(*templatePath)
	if err != nil {
		return err
	}
	policy, err := operations.RenderJournalPolicy(template, observation)
	if err != nil {
		return err
	}
	if err := operations.WriteJournalPolicy(*output, policy); err != nil {
		return err
	}
	return operations.WriteJournalBudget(*budgetOutput, observation.JournalMaxUseBytes)
}

func runRecord(args []string) error {
	set := flag.NewFlagSet("record", flag.ContinueOnError)
	directory := set.String("metrics-dir", "", "node-exporter textfile directory")
	operation := set.String("operation", "", "backup, release, or restore")
	result := set.String("result", "", "success or failure")
	atText := set.String("at", "", "RFC3339 result time")
	if err := set.Parse(args); err != nil || set.NArg() != 0 {
		return operations.ErrInvalid
	}
	at, err := time.Parse(time.RFC3339, *atText)
	if err != nil {
		return errors.Join(operations.ErrInvalid, err)
	}
	return operations.RecordOperation(*directory, *operation, *result, at)
}

func fail(command string, err error) {
	command, class := boundedFailure(command, err)
	slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("deployment operations command failed", "command", command, "error_class", class)
	os.Exit(1)
}

func boundedFailure(command string, err error) (string, string) {
	if command != "alert-test" && command != "alert-observe" && command != "host-observe" && command != "journal-observe" && command != "journal-render" && command != "record" {
		command = "unknown"
	}
	class := "operation_failed"
	if errors.Is(err, operations.ErrInvalid) {
		class = "invalid_input"
	}
	return command, class
}
