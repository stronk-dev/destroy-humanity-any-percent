package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHostAndObjectiveObservationsAreStrictExclusiveEvidence(t *testing.T) {
	evidence := validEvidence()
	host := HostObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt, CompletedAt: evidence.StartedAt.Add(time.Second),
		Host: evidence.Host, ObjectiveCompleted: true}
	objectives := ObjectiveObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
		CompletedAt: evidence.CompletedAt, Objectives: evidence.Objectives, ObjectiveCompleted: true}

	for name, value := range map[string]any{"host": host, "objectives": objectives} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if name == "host" {
			if _, err := DecodeHostObservation(encoded); err != nil {
				t.Fatal(err)
			}
		} else if _, err := DecodeObjectiveObservation(encoded); err != nil {
			t.Fatal(err)
		}
	}

	root := t.TempDir()
	hostPath := filepath.Join(root, "host.json")
	if err := WriteHostObservation(hostPath, host); err != nil {
		t.Fatal(err)
	}
	if err := WriteHostObservation(hostPath, host); err == nil {
		t.Fatal("host observation overwrite accepted")
	}
	info, err := os.Stat(hostPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("host evidence mode=%v err=%v", info.Mode(), err)
	}
}

func TestHostAndObjectiveObservationsRejectForgedState(t *testing.T) {
	evidence := validEvidence()
	host := HostObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt, CompletedAt: evidence.StartedAt.Add(time.Second),
		Host: evidence.Host, ObjectiveCompleted: true}
	objectives := ObjectiveObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
		CompletedAt: evidence.CompletedAt, Objectives: evidence.Objectives, ObjectiveCompleted: true}

	for name, mutate := range map[string]func(*HostObservation){
		"dirty source":        func(value *HostObservation) { value.Host.SourceCheckoutAbsent = false },
		"provider credential": func(value *HostObservation) { value.Host.ProviderCredentialsAbsent = false },
		"guarded":             func(value *HostObservation) { value.GuardExhausted = true },
		"zero interval":       func(value *HostObservation) { value.CompletedAt = value.StartedAt },
	} {
		t.Run("host "+name, func(t *testing.T) {
			value := host
			mutate(&value)
			if err := ValidateHostObservation(value); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid host observation accepted: %v", err)
			}
		})
	}
	for name, mutate := range map[string]func(*ObjectiveObservation){
		"rpo": func(value *ObjectiveObservation) {
			value.Objectives.NewestValidBackupAt = value.Objectives.IncidentAt.Add(-MaximumRPO - time.Second)
			value.Objectives.RPOSeconds = int64((MaximumRPO + time.Second) / time.Second)
		},
		"identity": func(value *ObjectiveObservation) { value.Objectives.RestoredIdentityMatch = false },
		"guarded":  func(value *ObjectiveObservation) { value.GuardExhausted = true },
	} {
		t.Run("objective "+name, func(t *testing.T) {
			value := objectives
			mutate(&value)
			if err := ValidateObjectiveObservation(value); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid objective observation accepted: %v", err)
			}
		})
	}
}
