package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud-clicker/server/releasepackage"
)

func main() {
	result, err := run(os.Args[1:])
	if err != nil {
		fail(err)
	}
	fmt.Printf("release secret scan: %d tracked files, no findings\n", result.TrackedFiles)
}

var commandOutput = func(root string, args ...string) ([]byte, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	return command.Output()
}

var now = time.Now

func run(args []string) (releasepackage.SecretScanResult, error) {
	set := flag.NewFlagSet("release-secret-scan", flag.ContinueOnError)
	root := set.String("root", "..", "repository root")
	archive := set.String("gameserver-archive", "", "optional gameserver docker-save archive")
	manifestSHA256 := set.String("manifest-sha256", "", "candidate release manifest SHA-256")
	resultOutput := set.String("output", "", "exclusive structured result path")
	if set.Parse(args) != nil || set.NArg() != 0 {
		return releasepackage.SecretScanResult{}, releasepackage.ErrInvalidContent
	}
	started := now().UTC()
	output, err := commandOutput(*root, "ls-files", "-z")
	if err != nil {
		return releasepackage.SecretScanResult{}, err
	}
	parts := bytes.Split(output, []byte{0})
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) != 0 {
			paths = append(paths, string(part))
		}
	}
	findings, err := releasepackage.ScanTrackedFiles(*root, paths)
	if err != nil {
		return releasepackage.SecretScanResult{}, err
	}
	if strings.TrimSpace(*archive) != "" {
		imageFindings, err := releasepackage.ScanDockerArchive(*archive)
		if err != nil {
			return releasepackage.SecretScanResult{}, err
		}
		findings = append(findings, imageFindings...)
	}
	if err := releasepackage.RequireNoSecrets(findings); err != nil {
		return releasepackage.SecretScanResult{}, err
	}
	commitBytes, err := commandOutput(*root, "rev-parse", "HEAD")
	if err != nil {
		return releasepackage.SecretScanResult{}, err
	}
	result := releasepackage.SecretScanResult{SchemaVersion: 1, ManifestSHA256: *manifestSHA256, SourceCommit: strings.TrimSpace(string(commitBytes)),
		StartedAt: started, CompletedAt: now().UTC(), TrackedFiles: len(paths), ImageArchiveScanned: strings.TrimSpace(*archive) != "", ObjectiveCompleted: true}
	structured := *manifestSHA256 != "" || *resultOutput != ""
	if structured {
		if *manifestSHA256 == "" || *resultOutput == "" || strings.TrimSpace(*archive) == "" || releasepackage.ValidateSecretScanResult(result) != nil {
			return releasepackage.SecretScanResult{}, releasepackage.ErrInvalidContent
		}
		if err := releasepackage.WriteSecretScanResult(*resultOutput, result); err != nil {
			return releasepackage.SecretScanResult{}, err
		}
	}
	return result, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
