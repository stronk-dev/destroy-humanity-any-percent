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
	"sort"
	"strings"
	"time"
)

const maximumCommandOutput = 1 << 20

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
	if plan.SchemaVersion != 1 || !identifierPattern.MatchString(plan.RunID) || !hashPattern.MatchString(plan.ManifestSHA256) ||
		!hashPattern.MatchString(plan.PreviousManifestSHA256) || plan.ManifestSHA256 == plan.PreviousManifestSHA256 ||
		len(plan.Checks) != len(RequiredPopulations) {
		return ErrInvalid
	}
	required := make([]string, 0, len(RequiredPopulations))
	for name := range RequiredPopulations {
		required = append(required, name)
	}
	sort.Strings(required)
	for index, check := range plan.Checks {
		kind := RequiredPopulations[check.Name]
		if check.Name != required[index] || check.Kind != kind || !slicesContains(RequiredSteps, check.Step) ||
			len(check.Command) == 0 || len(check.Command) > 32 || check.TimeoutSeconds < 1 || check.TimeoutSeconds > 4*60*60 ||
			kind == "positive" && check.ExpectedExit != 0 || kind == "negative" && check.ExpectedExit != 1 ||
			invalidCommand(check.Command) {
			return ErrInvalid
		}
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
