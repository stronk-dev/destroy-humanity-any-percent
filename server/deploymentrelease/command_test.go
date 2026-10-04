package deploymentrelease

import (
	"context"
	"errors"
	"testing"
)

func TestNonCleanRestoreRefusalRejectsUnattributedErrorText(t *testing.T) {
	err := errors.New(`unrelated Docker failure echoed {"error_class":"non_clean_target"} from an earlier log`)
	if IsNonCleanRestoreRefusal(err) {
		t.Fatal("unattributed substring accepted as a dirty-database restore refusal")
	}
}

func TestNonCleanRestoreRefusalRequiresExactCommandFailureRecord(t *testing.T) {
	const genuine = `{"time":"2026-09-24T12:00:00Z","level":"ERROR","msg":"deployment backup command failed","command":"restore","error_class":"non_clean_target"}`
	makeFailure := func(stderr string) *CommandFailure {
		return &CommandFailure{Command: "docker", ExitCode: 1, Stderr: stderr, Cause: errors.New("exit status 1")}
	}
	if !IsNonCleanRestoreRefusal(makeFailure(genuine)) {
		t.Fatal("exact structured restore refusal rejected")
	}
	for name, failure := range map[string]*CommandFailure{
		"other backup command": makeFailure(`{"time":"2026-09-24T12:00:00Z","level":"ERROR","msg":"deployment backup command failed","command":"inspect","error_class":"non_clean_target"}`),
		"other outer command":  {Command: "sh", ExitCode: 1, Stderr: genuine, Cause: errors.New("exit status 1")},
		"wrong exit":           {Command: "docker", ExitCode: 2, Stderr: genuine, Cause: errors.New("exit status 2")},
		"extra log line":       makeFailure("unrelated failure\n" + genuine),
		"embedded record":      makeFailure("operation failed: " + genuine),
		"other class":          makeFailure(`{"time":"2026-09-24T12:00:00Z","level":"ERROR","msg":"deployment backup command failed","command":"restore","error_class":"operation_failed"}`),
	} {
		t.Run(name, func(t *testing.T) {
			if IsNonCleanRestoreRefusal(failure) {
				t.Fatal("non-clean refusal accepted without its exact command record")
			}
		})
	}
}

func TestExecRunnerRetainsFailureProvenance(t *testing.T) {
	_, err := (ExecRunner{}).Run(context.Background(), t.TempDir(), "sh", "-c", "printf 'failed\\n' >&2; exit 1")
	var failure *CommandFailure
	if !errors.As(err, &failure) || failure.Command != "sh" || failure.ExitCode != 1 || failure.Stderr != "failed" || failure.Cause == nil {
		t.Fatalf("command failure lost executable, exit or stderr provenance: %+v", failure)
	}
}
