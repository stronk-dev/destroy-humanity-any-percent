package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/deploymentbrowser"
	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type browserRunDependencies struct {
	runner         hostCommandRunner
	validateBundle func(string) error
	loadManifest   func(string) (releasepackage.ReleaseManifest, []byte, error)
}

func RunBrowser(ctx context.Context, config ScenarioConfig) (deploymentbrowser.Result, error) {
	return runBrowser(ctx, config, browserRunDependencies{runner: deploymentrelease.ExecRunner{}, validateBundle: releasepackage.ValidateBundle,
		loadManifest: releasepackage.LoadReleaseManifest})
}

func runBrowser(ctx context.Context, config ScenarioConfig, dependencies browserRunDependencies) (deploymentbrowser.Result, error) {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.runner == nil ||
		dependencies.validateBundle == nil || dependencies.loadManifest == nil || dependencies.validateBundle(config.CandidateBundle) != nil {
		return deploymentbrowser.Result{}, ErrInvalid
	}
	manifest, manifestBytes, err := dependencies.loadManifest(config.CandidateBundle)
	if err != nil || len(manifest.RehearsalImages) != 1 || manifest.RehearsalImages[0].Name != "playwright" {
		return deploymentbrowser.Result{}, errors.Join(ErrInvalid, err)
	}
	manifestDigest := hashBytes(manifestBytes)
	output := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["browser_result"])
	if _, err := os.Lstat(output); !errors.Is(err, os.ErrNotExist) {
		return deploymentbrowser.Result{}, ErrInvalid
	}
	driver := filepath.Join(config.CandidateBundle, "deployment-browser")
	arguments := []string{"run", "--rm", "--platform=linux/amd64", "--network=host", "--read-only", "--tmpfs=/tmp:size=256m,mode=1777",
		"--shm-size=256m", "--security-opt=no-new-privileges", "--cap-drop=ALL",
		"--volume=" + driver + ":/usr/local/bin/deployment-browser:ro",
		"--volume=" + config.ArtifactsDirectory + ":/evidence",
		manifest.RehearsalImages[0].Reference, "/usr/local/bin/deployment-browser", "--origin=" + config.PublicOrigin,
		"--output=/evidence/" + requiredRunArtifactFiles["browser_result"], "--manifest-sha256=" + manifestDigest}
	if _, err := dependencies.runner.Run(ctx, config.WorkDirectory, "docker", arguments...); err != nil {
		return deploymentbrowser.Result{}, err
	}
	data, err := os.ReadFile(output)
	if err != nil {
		return deploymentbrowser.Result{}, err
	}
	result, err := deploymentbrowser.DecodeResult(data)
	if err != nil || result.ManifestSHA256 != manifestDigest {
		return deploymentbrowser.Result{}, errors.Join(ErrInvalid, err)
	}
	return result, nil
}

func containsArgument(arguments []string, target string) bool {
	for _, argument := range arguments {
		if argument == target || strings.HasPrefix(argument, target) {
			return true
		}
	}
	return false
}
