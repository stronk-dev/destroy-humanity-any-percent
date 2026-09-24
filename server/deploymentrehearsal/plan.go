package deploymentrehearsal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const maximumCommandOutput = 1 << 20

// ProbeRejectedExit is the only exit code that proves a negative row: the
// named fixture was fully prepared and its production gate rejected it. The
// rehearsal command never uses it for usage, validation or setup failures.
const ProbeRejectedExit = 3

// populationProducers names the rehearsal subcommand that alone may produce a
// population's result. Every other population is produced by `probe`, which
// fails loudly (exit 2) for a population it cannot yet exercise.
var populationProducers = map[string]string{
	"clean_linux_amd64_bundle_only_install": "install-candidate",
	"phase0_browser_flow_through_caddy":     "run-browser",
	"empty_database_backup_restore":         "recover-empty",
	"populated_database_identity_restore":   "recover-populated",
	"rpo_within_six_hours":                  "recover-populated",
	"rto_within_four_hours":                 "recover-populated",
	"bounded_drain_and_restart":             "lifecycle-release",
	"exact_previous_release_rollback":       "lifecycle-rollback",
}

// boundToProducer requires the row to invoke the rehearsal tool itself with
// the population's producer, so a constant command cannot stand in for proof.
func boundToProducer(check PlannedCheck) bool {
	if len(check.Command) < 2 || !filepath.IsAbs(check.Command[0]) || filepath.Base(check.Command[0]) != "deployment-rehearsal" {
		return false
	}
	producer := populationProducers[check.Name]
	if producer == "" {
		producer = "probe"
	}
	if check.Command[1] != producer {
		return false
	}
	populations := 0
	for _, argument := range check.Command[2:] {
		if strings.HasPrefix(argument, "--population=") {
			populations++
			if argument != "--population="+check.Name {
				return false
			}
		}
	}
	return producer == "probe" && populations == 1 || producer != "probe" && populations == 0
}

type ExecutionPlan struct {
	SchemaVersion          int            `json:"schema_version"`
	RunID                  string         `json:"run_id"`
	ManifestSHA256         string         `json:"manifest_sha256"`
	PreviousManifestSHA256 string         `json:"previous_manifest_sha256"`
	Checks                 []PlannedCheck `json:"checks"`
}

type PlannedCheck struct {
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	Step           string   `json:"step"`
	Command        []string `json:"command"`
	ExpectedExit   int      `json:"expected_exit"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

type requiredPlanCheck struct {
	name string
	step string
}

var requiredPlanChecks = []requiredPlanCheck{
	{name: "source_checkout_present", step: "host_preflight"},
	{name: "missing_or_malformed_secret", step: "host_preflight"},
	{name: "duplicate_key_id_or_value", step: "host_preflight"},
	{name: "invalid_origin_or_proxy_depth", step: "host_preflight"},
	{name: "clean_linux_amd64_bundle_only_install", step: "candidate_install"},
	{name: "phase0_browser_flow_through_caddy", step: "browser_phase0"},
	{name: "non_clean_restore_target", step: "empty_backup_restore"},
	{name: "truncated_or_corrupt_backup", step: "empty_backup_restore"},
	{name: "wrong_age_identity", step: "empty_backup_restore"},
	{name: "wrong_release_manifest", step: "empty_backup_restore"},
	{name: "empty_database_backup_restore", step: "empty_backup_restore"},
	{name: "interrupted_backup_writer", step: "populated_backup_restore"},
	{name: "populated_database_identity_restore", step: "populated_backup_restore"},
	{name: "rpo_or_rto_above_bound", step: "incident_recovery"},
	{name: "rpo_within_six_hours", step: "incident_recovery"},
	{name: "rto_within_four_hours", step: "incident_recovery"},
	{name: "gameserver_restart_during_admitted_work", step: "candidate_release"},
	{name: "wrong_epoch_or_artifact_set", step: "candidate_release"},
	{name: "bounded_drain_and_restart", step: "candidate_release"},
	{name: "missing_previous_image_or_backup", step: "previous_release_rollback"},
	{name: "irreversible_or_down_migration", step: "previous_release_rollback"},
	{name: "exact_previous_release_rollback", step: "previous_release_rollback"},
	{name: "current_previous_key_overlap", step: "key_rotation"},
	{name: "public_metrics_route", step: "operations_alerts_retention"},
	{name: "health_only_alert_receiver", step: "operations_alerts_retention"},
	{name: "severed_alert_rule_or_counter", step: "operations_alerts_retention"},
	{name: "early_journal_eviction", step: "operations_alerts_retention"},
	{name: "incomplete_or_guarded_observation", step: "operations_alerts_retention"},
	{name: "private_metrics_and_alert_delivery", step: "operations_alerts_retention"},
	{name: "fourteen_day_journal_budget", step: "operations_alerts_retention"},
	{name: "removed_catalog", step: "provider_off_supply_chain"},
	{name: "removed_client", step: "provider_off_supply_chain"},
	{name: "removed_license", step: "provider_off_supply_chain"},
	{name: "removed_config", step: "provider_off_supply_chain"},
	{name: "removed_helper", step: "provider_off_supply_chain"},
	{name: "changed_image_digest", step: "provider_off_supply_chain"},
	{name: "changed_runtime_config_digest", step: "provider_off_supply_chain"},
	{name: "changed_sbom", step: "provider_off_supply_chain"},
	{name: "seeded_source_secret", step: "provider_off_supply_chain"},
	{name: "seeded_image_secret", step: "provider_off_supply_chain"},
	{name: "provider_off_operation", step: "provider_off_supply_chain"},
	{name: "six_image_sbom_license_provenance", step: "provider_off_supply_chain"},
}

type CheckResult struct {
	SchemaVersion  int       `json:"schema_version"`
	Name           string    `json:"name"`
	Kind           string    `json:"kind"`
	Step           string    `json:"step"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	CommandSHA256  string    `json:"command_sha256"`
	OutputSHA256   string    `json:"output_sha256"`
	ExitCode       int       `json:"exit_code"`
	Result         string    `json:"result"`
	GuardExhausted bool      `json:"guard_exhausted"`
}

func LoadExecutionPlan(path string) (ExecutionPlan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ExecutionPlan{}, err
	}
	if containsForbiddenEvidenceKey(data) {
		return ExecutionPlan{}, ErrInvalid
	}
	var plan ExecutionPlan
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&plan) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateExecutionPlan(plan) != nil {
		return ExecutionPlan{}, ErrInvalid
	}
	return plan, nil
}

func ValidateExecutionPlan(plan ExecutionPlan) error {
	populations := planPopulations()
	if plan.SchemaVersion != 1 || !identifierPattern.MatchString(plan.RunID) || !hashPattern.MatchString(plan.ManifestSHA256) ||
		!hashPattern.MatchString(plan.PreviousManifestSHA256) || plan.ManifestSHA256 == plan.PreviousManifestSHA256 ||
		len(plan.Checks) != len(populations) {
		return ErrInvalid
	}
	if len(requiredPlanChecks) != len(populations) {
		return ErrInvalid
	}
	seen := make(map[string]bool, len(requiredPlanChecks))
	for index, check := range plan.Checks {
		required := requiredPlanChecks[index]
		kind := populations[check.Name]
		if required.name != check.Name || required.step != check.Step || seen[check.Name] || kind == "" || check.Kind != kind ||
			len(check.Command) == 0 || len(check.Command) > 32 || check.TimeoutSeconds < 1 || check.TimeoutSeconds > 4*60*60 ||
			kind == "positive" && check.ExpectedExit != 0 || kind == "negative" && check.ExpectedExit != ProbeRejectedExit ||
			invalidCommand(check.Command) || !boundToProducer(check) {
			return ErrInvalid
		}
		seen[check.Name] = true
	}
	return nil
}

func ExecutePlan(ctx context.Context, plan ExecutionPlan, outputDirectory string, now func() time.Time) error {
	if ValidateExecutionPlan(plan) != nil || outputDirectory == "" || now == nil {
		return ErrInvalid
	}
	if entries, err := os.ReadDir(outputDirectory); err != nil || len(entries) != 0 {
		return errors.Join(ErrInvalid, err)
	}
	for _, check := range plan.Checks {
		started := now().UTC()
		if started.IsZero() {
			return ErrInvalid
		}
		checkCtx, cancel := context.WithTimeout(ctx, time.Duration(check.TimeoutSeconds)*time.Second)
		command := exec.CommandContext(checkCtx, check.Command[0], check.Command[1:]...)
		var output boundedBuffer
		command.Stdout, command.Stderr = &output, &output
		err := command.Run()
		guarded := errors.Is(checkCtx.Err(), context.DeadlineExceeded) || output.overflow
		cancel()
		exitCode := 0
		if err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				exitCode = exit.ExitCode()
			} else {
				exitCode = -1
			}
		}
		completed := now().UTC()
		result := "passed"
		if guarded || exitCode != check.ExpectedExit || !completed.After(started) {
			result = "failed"
		}
		resultRecord := CheckResult{SchemaVersion: 1, Name: check.Name, Kind: check.Kind, Step: check.Step,
			StartedAt: started, CompletedAt: completed, CommandSHA256: hashJSON(check.Command), OutputSHA256: hashBytes(output.data),
			ExitCode: exitCode, Result: result, GuardExhausted: guarded}
		encoded, _ := json.Marshal(resultRecord)
		path := filepath.Join(outputDirectory, check.Name+".json")
		if writeNewEvidence(path, append(encoded, '\n')) != nil || result != "passed" {
			return ErrInvalid
		}
	}
	return nil
}

type boundedBuffer struct {
	data     []byte
	overflow bool
}

func (buffer *boundedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	remaining := maximumCommandOutput - len(buffer.data)
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
		}
		buffer.data = append(buffer.data, data...)
	}
	if written > remaining {
		buffer.overflow = true
	}
	return written, nil
}

func invalidCommand(command []string) bool {
	base := strings.ToLower(filepath.Base(command[0]))
	if base == "sh" || base == "bash" || base == "zsh" || base == "fish" || base == "sudo" || strings.TrimSpace(command[0]) != command[0] {
		return true
	}
	for _, argument := range command {
		lower := strings.ToLower(argument)
		for _, forbidden := range []string{"password=", "database_url=", "database-url=", "recovery_code=", "recovery-code=", "secret_value=", "token="} {
			if strings.Contains(lower, forbidden) {
				return true
			}
		}
		if strings.ContainsRune(argument, 0) || strings.ContainsAny(argument, "\r\n") {
			return true
		}
	}
	return false
}

func hashJSON(value any) string {
	data, _ := json.Marshal(value)
	return hashBytes(data)
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func writeNewEvidence(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	return errors.Join(err, file.Close())
}

func slicesContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
