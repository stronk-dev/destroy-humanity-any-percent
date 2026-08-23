package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
)

func TestScenarioConfigStrictlyBindsPrivateRuntimeInputs(t *testing.T) {
	config := validScenarioConfig(t)
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "scenario.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadScenarioConfig(path)
	if err != nil || loaded.RunID != config.RunID {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}

	unknown := append(encoded[:len(encoded)-1], []byte(`,"invented":true}`)...)
	if _, err := DecodeScenarioConfig(unknown); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown field accepted: %v", err)
	}
	if _, err := DecodeScenarioConfig(append(encoded, []byte(`{}`)...)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("trailing value accepted: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarioConfig(path); !errors.Is(err, ErrInvalid) {
		t.Fatalf("public scenario input accepted: %v", err)
	}
}

func TestScenarioConfigRejectsEveryUnsafeFieldClass(t *testing.T) {
	for name, mutate := range map[string]func(*ScenarioConfig){
		"relative path": func(value *ScenarioConfig) { value.WorkDirectory = "relative" },
		"nested mutable path": func(value *ScenarioConfig) {
			value.ArtifactsDirectory = filepath.Join(value.WorkDirectory, "artifacts")
		},
		"bundle collision":      func(value *ScenarioConfig) { value.PreviousBundle = value.CandidateBundle },
		"invalid public origin": func(value *ScenarioConfig) { value.PublicOrigin = "http://game.example" },
		"credentialed receiver": func(value *ScenarioConfig) { value.ReceiverHealthURL = "https://user:pass@receiver.example/health" },
		"queried receiver":      func(value *ScenarioConfig) { value.ReceiverHealthURL = "https://receiver.example/health?secret=no" },
		"invalid recipient":     func(value *ScenarioConfig) { value.AgeRecipient = "age1invalid" },
		"invalid server":        func(value *ScenarioConfig) { value.ServerID = "server-1" },
	} {
		t.Run(name, func(t *testing.T) {
			config := validScenarioConfig(t)
			mutate(&config)
			if err := ValidateScenarioConfig(config); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unsafe scenario config accepted: %v", err)
			}
		})
	}
}

func TestScenarioConfigRejectsMissingSymlinkedAndPublicRuntimeInputs(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		config := validScenarioConfig(t)
		if err := os.Remove(config.MetricsDirectory); err != nil {
			t.Fatal(err)
		}
		if err := validateScenarioFilesystem(config); !errors.Is(err, ErrInvalid) {
			t.Fatalf("missing metrics directory accepted: %v", err)
		}
	})
	t.Run("symlinked directory", func(t *testing.T) {
		config := validScenarioConfig(t)
		target := config.WorkDirectory + "-real"
		if err := os.Rename(config.WorkDirectory, target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, config.WorkDirectory); err != nil {
			t.Fatal(err)
		}
		if err := validateScenarioFilesystem(config); !errors.Is(err, ErrInvalid) {
			t.Fatalf("symlinked work directory accepted: %v", err)
		}
	})
	t.Run("public identity", func(t *testing.T) {
		config := validScenarioConfig(t)
		if err := os.Chmod(config.AgeIdentityFile, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := validateScenarioFilesystem(config); !errors.Is(err, ErrInvalid) {
			t.Fatalf("public age identity accepted: %v", err)
		}
	})
}

func validScenarioConfig(t *testing.T) ScenarioConfig {
	t.Helper()
	root := t.TempDir()
	directory := func(name string) string {
		path := filepath.Join(root, name)
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
		return path
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	identityPath := filepath.Join(root, "age-identity.txt")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return ScenarioConfig{SchemaVersion: 1, RunID: "r006-runtime-001", CandidateBundle: directory("candidate"),
		PreviousBundle: directory("previous"), WorkDirectory: directory("work"), ArtifactsDirectory: directory("artifacts"),
		OperatorState: directory("operator"), BackupTarget: directory("backup"), MetricsDirectory: directory("metrics"),
		AgeIdentityFile: identityPath, PublicOrigin: "https://game.example", ReceiverHealthURL: "https://receiver.example/health",
		AgeRecipient: identity.Recipient().String(), ServerID: "01986666-b001-4000-8000-000000000001"}
}
