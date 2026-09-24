package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type hostCommandRunner interface {
	Run(context.Context, string, string, ...string) ([]byte, error)
}

type hostObservationDependencies struct {
	runner         hostCommandRunner
	readFile       func(string) ([]byte, error)
	environment    []string
	now            func() time.Time
	validateBundle func(string) error
}

var (
	distributionID      = regexp.MustCompile(`^[a-z0-9]+$`)
	distributionVersion = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)*$`)
)

func ObserveHost(ctx context.Context, config ScenarioConfig) (HostObservation, error) {
	return observeHost(ctx, config, hostObservationDependencies{runner: deploymentrelease.ExecRunner{}, readFile: os.ReadFile,
		environment: os.Environ(), now: time.Now, validateBundle: releasepackage.ValidateBundle})
}

func observeHost(ctx context.Context, config ScenarioConfig, dependencies hostObservationDependencies) (HostObservation, error) {
	if ValidateScenarioConfig(config) != nil || dependencies.runner == nil || dependencies.readFile == nil ||
		dependencies.now == nil || dependencies.validateBundle == nil {
		return HostObservation{}, ErrInvalid
	}
	started := dependencies.now().UTC()
	if started.IsZero() || dependencies.validateBundle(config.CandidateBundle) != nil || dependencies.validateBundle(config.PreviousBundle) != nil ||
		!sourceMetadataAbsent(config) || !providerCredentialsAbsent(dependencies.environment) {
		return HostObservation{}, ErrInvalid
	}
	osName, err := runHostValue(ctx, dependencies.runner, config.WorkDirectory, "uname", "-s")
	if err != nil || strings.ToLower(osName) != "linux" {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	architecture, err := runHostValue(ctx, dependencies.runner, config.WorkDirectory, "uname", "-m")
	if err != nil || architecture != "x86_64" && architecture != "amd64" {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	kernel, err := runHostValue(ctx, dependencies.runner, config.WorkDirectory, "uname", "-r")
	if err != nil || kernel == "" || len(kernel) > 128 {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	distributionBytes, err := dependencies.readFile("/etc/os-release")
	distribution, distributionErr := parseDistribution(distributionBytes)
	if err != nil || distributionErr != nil {
		return HostObservation{}, errors.Join(ErrInvalid, err, distributionErr)
	}
	dockerEngine, err := runHostValue(ctx, dependencies.runner, config.WorkDirectory, "docker", "version", "--format={{.Server.Version}}")
	if err != nil || dockerEngine == "" || len(dockerEngine) > 128 {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	dockerCompose, err := runHostValue(ctx, dependencies.runner, config.WorkDirectory, "docker", "compose", "version", "--short")
	if err != nil || dockerCompose == "" || len(dockerCompose) > 128 {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	containers, err := dependencies.runner.Run(ctx, config.CandidateBundle, "docker", "compose", "--project-name", "cloud-clicker",
		"--file", filepath.Join(config.CandidateBundle, "compose.yml"), "ps", "--all", "--quiet")
	if err != nil || strings.TrimSpace(string(containers)) != "" {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	// A clean start owns no project volume at all: retained certificates,
	// metrics or alert state are prior operator state, not a fresh host.
	volumes, err := dependencies.runner.Run(ctx, config.WorkDirectory, "docker", "volume", "ls", "--quiet", "--filter=name=^cloud-clicker_")
	if err != nil {
		return HostObservation{}, errors.Join(ErrInvalid, err)
	}
	for _, name := range strings.Fields(string(volumes)) {
		if strings.HasPrefix(name, "cloud-clicker_") {
			return HostObservation{}, ErrInvalid
		}
	}
	completed := dependencies.now().UTC()
	observation := HostObservation{SchemaVersion: 1, StartedAt: started, CompletedAt: completed,
		Host: Host{OS: "linux", Architecture: "amd64", Distribution: distribution, Kernel: kernel,
			DockerEngine: dockerEngine, DockerCompose: dockerCompose, CleanStart: true, SourceCheckoutAbsent: true,
			ProviderCredentialsAbsent: true}, ObjectiveCompleted: true}
	if err := WriteHostObservation(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["host_observation"]), observation); err != nil {
		return HostObservation{}, err
	}
	return observation, nil
}

func runHostValue(ctx context.Context, runner hostCommandRunner, directory, name string, arguments ...string) (string, error) {
	data, err := runner.Run(ctx, directory, name, arguments...)
	return strings.TrimSpace(string(data)), err
}

func parseDistribution(data []byte) (string, error) {
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		name, value, found := strings.Cut(line, "=")
		if !found || name != "ID" && name != "VERSION_ID" {
			continue
		}
		value = strings.Trim(value, `"`)
		if _, exists := values[name]; exists {
			return "", ErrInvalid
		}
		values[name] = value
	}
	if !distributionID.MatchString(values["ID"]) || !distributionVersion.MatchString(values["VERSION_ID"]) {
		return "", ErrInvalid
	}
	result := values["ID"] + "-" + strings.ReplaceAll(values["VERSION_ID"], ".", "-")
	if !identifierPattern.MatchString(result) {
		return "", ErrInvalid
	}
	return result, nil
}

func sourceMetadataAbsent(config ScenarioConfig) bool {
	for _, directory := range []string{config.CandidateBundle, config.PreviousBundle, config.WorkDirectory,
		config.ArtifactsDirectory, config.InstallOperatorState, config.LifecycleOperatorState, config.BackupTarget, config.MetricsDirectory} {
		if _, err := os.Lstat(filepath.Join(directory, ".git")); !errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}

func providerCredentialsAbsent(environment []string) bool {
	exact := map[string]bool{
		"GOOGLE_APPLICATION_CREDENTIALS": true, "OPENAI_API_KEY": true, "ANTHROPIC_API_KEY": true,
		"STRIPE_SECRET_KEY": true, "SENDGRID_API_KEY": true, "MAILGUN_API_KEY": true, "SENTRY_DSN": true,
	}
	prefixes := []string{"AWS_", "AZURE_", "GCP_", "GOOGLE_CLOUD_", "TWILIO_", "SMTP_", "DATADOG_", "NEW_RELIC_"}
	for _, entry := range environment {
		name, value, found := strings.Cut(entry, "=")
		if !found || value == "" {
			continue
		}
		if exact[name] {
			return false
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				return false
			}
		}
	}
	return true
}
