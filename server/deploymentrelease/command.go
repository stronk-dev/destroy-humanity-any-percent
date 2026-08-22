package deploymentrelease

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner interface {
	Run(context.Context, string, string, ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, directory, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%s failed: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func composeArgs(bundle Bundle, values ...string) []string {
	return append([]string{"compose", "--project-name", "cloud-clicker", "--file", bundle.Root + "/compose.yml"}, values...)
}
