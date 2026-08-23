package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type hostRunnerFixture struct {
	values map[string][]byte
	errAt  string
}

func (runner hostRunnerFixture) Run(_ context.Context, _ string, name string, arguments ...string) ([]byte, error) {
	key := strings.Join(append([]string{name}, arguments...), " ")
	if key == runner.errAt {
		return nil, errors.New("injected host command failure")
	}
	value, ok := runner.values[key]
	if !ok && strings.HasPrefix(key, "docker compose --project-name cloud-clicker --file ") && strings.HasSuffix(key, "compose.yml ps --all --quiet") {
		return nil, nil
	}
	if !ok {
		return nil, errors.New("unexpected host command: " + key)
	}
	return append([]byte(nil), value...), nil
}

func TestHostObservationUsesExactCleanProviderOffMachineState(t *testing.T) {
	config := validScenarioConfig(t)
	dependencies := validHostDependencies()
	observation, err := observeHost(context.Background(), config, dependencies)
	if err != nil || observation.Host.Distribution != "debian-13" || observation.Host.DockerEngine != "28.4.0" ||
		!observation.Host.CleanStart || !observation.Host.ProviderCredentialsAbsent {
		t.Fatalf("observation=%+v err=%v", observation, err)
	}
	path := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["host_observation"])
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if decoded, err := DecodeHostObservation(data); err != nil || decoded.Host != observation.Host {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	if _, err := observeHost(context.Background(), config, dependencies); err == nil {
		t.Fatal("host observation overwrite accepted")
	}
}

func TestHostObservationRejectsEveryUnsupportedOrDirtyBoundary(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, *ScenarioConfig, *hostObservationDependencies){
		"provider credential": func(_ *testing.T, _ *ScenarioConfig, dependencies *hostObservationDependencies) {
			dependencies.environment = append(dependencies.environment, "OPENAI_API_KEY=present")
		},
		"checkout metadata": func(t *testing.T, config *ScenarioConfig, _ *hostObservationDependencies) {
			if err := os.Mkdir(filepath.Join(config.WorkDirectory, ".git"), 0o700); err != nil {
				t.Fatal(err)
			}
		},
		"running project": func(_ *testing.T, _ *ScenarioConfig, _ *hostObservationDependencies) {},
		"wrong os": func(_ *testing.T, _ *ScenarioConfig, dependencies *hostObservationDependencies) {
			dependencies.runner.(hostRunnerFixture).values["uname -s"] = []byte("Darwin\n")
		},
		"wrong architecture": func(_ *testing.T, _ *ScenarioConfig, dependencies *hostObservationDependencies) {
			dependencies.runner.(hostRunnerFixture).values["uname -m"] = []byte("aarch64\n")
		},
		"bad distribution": func(_ *testing.T, _ *ScenarioConfig, dependencies *hostObservationDependencies) {
			dependencies.readFile = func(string) ([]byte, error) { return []byte("ID=debian\nVERSION_ID=rolling\n"), nil }
		},
		"invalid bundle": func(_ *testing.T, _ *ScenarioConfig, dependencies *hostObservationDependencies) {
			dependencies.validateBundle = func(string) error { return ErrInvalid }
		},
	} {
		t.Run(name, func(t *testing.T) {
			config := validScenarioConfig(t)
			dependencies := validHostDependencies()
			if name == "running project" {
				runner := dependencies.runner.(hostRunnerFixture)
				runner.values["docker compose --project-name cloud-clicker --file "+filepath.Join(config.CandidateBundle, "compose.yml")+" ps --all --quiet"] = []byte("container\n")
				dependencies.runner = runner
			} else {
				mutate(t, &config, &dependencies)
			}
			if _, err := observeHost(context.Background(), config, dependencies); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unsupported host accepted: %v", err)
			}
		})
	}
}

func TestProviderCredentialDetectorRejectsEveryReleaseProviderFamily(t *testing.T) {
	for _, name := range []string{"AWS_ACCESS_KEY_ID", "AZURE_CLIENT_SECRET", "GOOGLE_APPLICATION_CREDENTIALS", "OPENAI_API_KEY",
		"ANTHROPIC_API_KEY", "STRIPE_SECRET_KEY", "SENDGRID_API_KEY", "TWILIO_AUTH_TOKEN", "SMTP_PASSWORD", "SENTRY_DSN",
		"DATADOG_API_KEY", "NEW_RELIC_LICENSE_KEY"} {
		if providerCredentialsAbsent([]string{"PATH=/usr/bin", name + "=present"}) {
			t.Fatalf("provider credential %s accepted", name)
		}
	}
	if !providerCredentialsAbsent([]string{"PATH=/usr/bin", "CLOUD_CLICKER_SERVER_ID=01986666-b001-4000-8000-000000000001"}) {
		t.Fatal("non-secret local server identity treated as provider credential")
	}
}

func validHostDependencies() hostObservationDependencies {
	clock := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	return hostObservationDependencies{runner: hostRunnerFixture{values: map[string][]byte{
		"uname -s": []byte("Linux\n"),
		"uname -m": []byte("x86_64\n"),
		"uname -r": []byte("6.12.0\n"),
		"docker version --format={{.Server.Version}}":                          []byte("28.4.0\n"),
		"docker compose version --short":                                       []byte("2.39.4\n"),
		"docker volume ls --quiet --filter=name=^cloud-clicker_postgres_data$": nil,
	}}, readFile: func(string) ([]byte, error) { return []byte("ID=debian\nVERSION_ID=\"13\"\n"), nil },
		environment: []string{"PATH=/usr/bin"}, validateBundle: func(string) error { return nil }, now: func() time.Time {
			clock = clock.Add(time.Second)
			return clock
		}}
}
