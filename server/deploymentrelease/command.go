package deploymentrelease

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner interface {
	Run(context.Context, string, string, ...string) ([]byte, error)
}

type ExecRunner struct{}

// CommandFailure preserves the failed command's exit status and stderr as
// separate fields. Callers that need a specific tool refusal must classify
// those fields, not search arbitrary wrapped error text.
type CommandFailure struct {
	Command  string
	ExitCode int
	Stderr   string
	Cause    error
}

func (failure *CommandFailure) Error() string {
	return fmt.Sprintf("%s failed: %v: %s", failure.Command, failure.Cause, failure.Stderr)
}

func (failure *CommandFailure) Unwrap() error { return failure.Cause }

func (ExecRunner) Run(ctx context.Context, directory, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		exitCode := -1
		var exited *exec.ExitError
		if errors.As(err, &exited) {
			exitCode = exited.ExitCode()
		}
		return nil, &CommandFailure{Command: name, ExitCode: exitCode, Stderr: strings.TrimSpace(stderr.String()), Cause: err}
	}
	return stdout.Bytes(), nil
}

func (runtime DockerRuntime) composeArgs(bundle Bundle, values ...string) []string {
	args := []string{"compose", "--project-name", "cloud-clicker", "--file", bundle.Root + "/compose.yml"}
	if runtime.rotationOverlay {
		args = append(args, "--file", bundle.Root+"/compose.rotation.yml")
	}
	return append(args, values...)
}
