package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"filippo.io/age"

	"cloud-clicker/server/deploymentconfig"
)

const maximumScenarioConfigBytes = 64 << 10

var scenarioServerID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type ScenarioConfig struct {
	SchemaVersion      int    `json:"schema_version"`
	RunID              string `json:"run_id"`
	CandidateBundle    string `json:"candidate_bundle"`
	PreviousBundle     string `json:"previous_bundle"`
	WorkDirectory      string `json:"work_directory"`
	ArtifactsDirectory string `json:"artifacts_directory"`
	OperatorState      string `json:"operator_state"`
	BackupTarget       string `json:"backup_target"`
	MetricsDirectory   string `json:"metrics_directory"`
	AgeIdentityFile    string `json:"age_identity_file"`
	PublicOrigin       string `json:"public_origin"`
	ReceiverHealthURL  string `json:"receiver_health_url"`
	AgeRecipient       string `json:"age_recipient"`
	ServerID           string `json:"server_id"`
}

func LoadScenarioConfig(path string) (ScenarioConfig, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return ScenarioConfig{}, ErrInvalid
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 || info.Size() < 1 || info.Size() > maximumScenarioConfigBytes {
		return ScenarioConfig{}, ErrInvalid
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ScenarioConfig{}, err
	}
	config, err := DecodeScenarioConfig(data)
	if err != nil || validateScenarioFilesystem(config) != nil {
		return ScenarioConfig{}, ErrInvalid
	}
	return config, nil
}

func DecodeScenarioConfig(data []byte) (ScenarioConfig, error) {
	var config ScenarioConfig
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&config) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateScenarioConfig(config) != nil {
		return ScenarioConfig{}, ErrInvalid
	}
	return config, nil
}

func ValidateScenarioConfig(config ScenarioConfig) error {
	paths := []string{config.CandidateBundle, config.PreviousBundle, config.WorkDirectory, config.ArtifactsDirectory,
		config.OperatorState, config.BackupTarget, config.MetricsDirectory, config.AgeIdentityFile}
	for _, path := range paths {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return ErrInvalid
		}
	}
	if config.SchemaVersion != 1 || !identifierPattern.MatchString(config.RunID) || !scenarioServerID.MatchString(config.ServerID) ||
		!deploymentconfig.ValidProductionOrigin(config.PublicOrigin) {
		return ErrInvalid
	}
	receiver, err := url.Parse(config.ReceiverHealthURL)
	if err != nil || receiver.Host == "" || receiver.User != nil || receiver.Fragment != "" || receiver.RawQuery != "" ||
		receiver.Scheme != "http" && receiver.Scheme != "https" {
		return ErrInvalid
	}
	if _, err := age.ParseX25519Recipient(config.AgeRecipient); err != nil {
		return ErrInvalid
	}
	if config.CandidateBundle == config.PreviousBundle || pathsOverlap(config.CandidateBundle, config.PreviousBundle) {
		return ErrInvalid
	}
	mutable := []string{config.WorkDirectory, config.ArtifactsDirectory, config.OperatorState, config.BackupTarget, config.MetricsDirectory}
	for index, left := range mutable {
		for _, right := range mutable[index+1:] {
			if pathsOverlap(left, right) {
				return ErrInvalid
			}
		}
		for _, bundle := range []string{config.CandidateBundle, config.PreviousBundle} {
			if pathsOverlap(left, bundle) {
				return ErrInvalid
			}
		}
	}
	for _, directory := range mutable {
		if pathsOverlap(config.AgeIdentityFile, directory) {
			return ErrInvalid
		}
	}
	return nil
}

func validateScenarioFilesystem(config ScenarioConfig) error {
	for _, directory := range []string{config.CandidateBundle, config.PreviousBundle, config.WorkDirectory, config.ArtifactsDirectory,
		config.OperatorState, config.BackupTarget, config.MetricsDirectory} {
		info, err := os.Lstat(directory)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
	}
	identity, err := os.Lstat(config.AgeIdentityFile)
	if err != nil || !identity.Mode().IsRegular() || identity.Mode()&os.ModeSymlink != 0 || identity.Mode().Perm()&0o077 != 0 || identity.Size() < 1 {
		return ErrInvalid
	}
	return nil
}

func pathsOverlap(left, right string) bool {
	relative, err := filepath.Rel(left, right)
	if err == nil && relativeWithin(relative) {
		return true
	}
	relative, err = filepath.Rel(right, left)
	return err == nil && relativeWithin(relative)
}

func relativeWithin(relative string) bool {
	return relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
