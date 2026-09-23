package deploymentrehearsal

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbrowser"
	"cloud-clicker/server/releasepackage"
)

type browserRunnerFixture struct {
	t              *testing.T
	config         ScenarioConfig
	manifestDigest string
	resultManifest string
	fail           bool
	skipOutput     bool
}

func (runner browserRunnerFixture) Run(_ context.Context, _ string, name string, arguments ...string) ([]byte, error) {
	runner.t.Helper()
	if runner.fail {
		return nil, errors.New("injected docker failure")
	}
	for _, required := range []string{"--platform=linux/amd64", "--network=host", "--read-only", "--security-opt=no-new-privileges",
		"--cap-drop=ALL", "playwright:v1@sha256:" + strings.Repeat("a", 64), "--origin=" + runner.config.PublicOrigin,
		"--manifest-sha256=" + runner.manifestDigest} {
		if name != "docker" || !containsArgument(arguments, required) {
			runner.t.Fatalf("browser command missing %q: %s %s", required, name, strings.Join(arguments, " "))
		}
	}
	if containsArgument(arguments, "--allow-loopback-http") {
		runner.t.Fatal("production browser command enabled loopback HTTP")
	}
	if runner.skipOutput {
		return []byte("driver claimed success\n"), nil
	}
	manifest := runner.resultManifest
	if manifest == "" {
		manifest = runner.manifestDigest
	}
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	result := deploymentbrowser.Result{SchemaVersion: 2, ManifestSHA256: manifest, StartedAt: start, CompletedAt: start.Add(time.Second),
		Surface: "run_2_desk", BootstrapCommitted: true, CredentialsPresent: true, WebSocketObserved: true,
		ManualIntentObserved: true, ManualIntentStatus: 200, GateCrossed: true, RunEndObserved: true,
		NextRunObserved: true, InitialRunSeq: 1, FinalRunSeq: 2, ActionCount: 3, IntentResponseCount: 3,
		ObjectiveCompleted: true}
	data, _ := json.Marshal(result)
	path := filepath.Join(runner.config.ArtifactsDirectory, requiredRunArtifactFiles["browser_result"])
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		runner.t.Fatal(err)
	}
	return []byte("browser passed\n"), nil
}

func TestProductBrowserUsesCandidateDriverImageAndTypedResult(t *testing.T) {
	config := validScenarioConfig(t)
	manifestBytes := []byte("candidate manifest\n")
	digest := hashBytes(manifestBytes)
	runner := browserRunnerFixture{t: t, config: config, manifestDigest: digest}
	dependencies := browserDependencies(runner, manifestBytes)
	result, err := runBrowser(context.Background(), config, dependencies)
	if err != nil || result.ManifestSHA256 != digest || !result.ObjectiveCompleted {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestProductBrowserRejectsCommandAndResultFailures(t *testing.T) {
	for name, mutate := range map[string]func(*browserRunnerFixture, *browserRunDependencies){
		"docker failure": func(runner *browserRunnerFixture, _ *browserRunDependencies) { runner.fail = true },
		"missing output": func(runner *browserRunnerFixture, _ *browserRunDependencies) { runner.skipOutput = true },
		"wrong manifest": func(runner *browserRunnerFixture, _ *browserRunDependencies) {
			runner.resultManifest = hashForBuild("f")
		},
		"invalid bundle": func(_ *browserRunnerFixture, dependencies *browserRunDependencies) {
			dependencies.validateBundle = func(string) error { return ErrInvalid }
		},
	} {
		t.Run(name, func(t *testing.T) {
			config := validScenarioConfig(t)
			manifestBytes := []byte("candidate manifest\n")
			runner := browserRunnerFixture{t: t, config: config, manifestDigest: hashBytes(manifestBytes)}
			dependencies := browserDependencies(runner, manifestBytes)
			mutate(&runner, &dependencies)
			dependencies.runner = runner
			if _, err := runBrowser(context.Background(), config, dependencies); err == nil {
				t.Fatal("invalid product browser result accepted")
			}
		})
	}

	t.Run("output overwrite", func(t *testing.T) {
		config := validScenarioConfig(t)
		path := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["browser_result"])
		if err := os.WriteFile(path, []byte("prior\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		manifestBytes := []byte("candidate manifest\n")
		runner := browserRunnerFixture{t: t, config: config, manifestDigest: hashBytes(manifestBytes)}
		if _, err := runBrowser(context.Background(), config, browserDependencies(runner, manifestBytes)); err == nil {
			t.Fatal("browser output overwrite accepted")
		}
	})
}

func browserDependencies(runner browserRunnerFixture, manifestBytes []byte) browserRunDependencies {
	manifest := releasepackage.ReleaseManifest{RehearsalImages: []releasepackage.Image{{Name: "playwright",
		Reference: "playwright:v1@sha256:" + strings.Repeat("a", 64)}}}
	return browserRunDependencies{runner: runner, validateBundle: func(string) error { return nil },
		loadManifest: func(string) (releasepackage.ReleaseManifest, []byte, error) { return manifest, manifestBytes, nil }}
}
