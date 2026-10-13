package deploymentrelease

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"filippo.io/age"
)

func TestCreatedBackupTimestampsMustBelongToTheInvokedCommand(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	for _, preUpgrade := range []bool{false, true} {
		kind := "recovery"
		if preUpgrade {
			kind = "pre-upgrade"
		}
		for _, vector := range []struct {
			name                string
			started, completed  time.Time
			before, after       time.Time
			accepted            bool
			commandMustNotStart bool
		}{
			{"fresh interior", start.Add(time.Second), end.Add(-time.Second), start, end, true, false},
			{"exact invocation edges", start, end, start, end, true, false},
			{"same instant invocation", start, start, start, start, true, false},
			{"old intact backup", start.Add(-24 * time.Hour), start.Add(-23 * time.Hour), start, end, false, false},
			{"started before invocation", start.Add(-time.Nanosecond), end, start, end, false, false},
			{"completed after reply", start, end.Add(time.Nanosecond), start, end, false, false},
			{"future intact backup", end.Add(time.Second), end.Add(2 * time.Second), start, end, false, false},
			{"zero invocation clock", start, end, time.Time{}, end, false, true},
			{"zero completion clock", start, end, start, time.Time{}, false, false},
			{"backwards invocation clock", start, end, end, start, false, false},
		} {
			t.Run(kind+"/"+vector.name, func(t *testing.T) {
				bundle := dockerFixtureBundle(t)
				target := t.TempDir()
				header, path, err := deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target,
					BackupID: vector.started.Format("20060102T150405Z") + "-abcdef123456", ServerID: "server-1",
					ReleaseManifestSHA256: bundle.ManifestSHA256, EpochID: bundle.Manifest.EpochID,
					StartedAt: vector.started, Now: func() time.Time { return vector.completed },
					Recipient: identity.Recipient().String(), Dump: bytes.NewReader([]byte("PGDMP\x01invocation fixture")), PreUpgrade: preUpgrade})
				if err != nil {
					t.Fatal(err)
				}
				if actual, err := deploymentbackup.ReadHeader(path); err != nil || actual != header {
					t.Fatal("invalid envelope cannot establish the timestamp population")
				}
				var restored bytes.Buffer
				if _, err := deploymentbackup.Restore(path, bundle.ManifestSHA256, identity, &restored); err != nil ||
					!bytes.Equal(restored.Bytes(), []byte("PGDMP\x01invocation fixture")) {
					t.Fatal("timestamp fixture is not an actual intact authenticated backup")
				}
				runner := &commandFixture{output: func(call []string) ([]byte, error) {
					if !strings.Contains(strings.Join(call, " "), " backup create ") {
						return nil, errors.New("unexpected backup command")
					}
					return json.Marshal(map[string]any{"status": "completed", "backup": "/backups/" + filepath.Base(path), "header": header})
				}}
				runtime := dockerFixtureRuntime(runner, target, "")
				reads := 0
				runtime.Now = func() time.Time {
					reads++
					if reads == 1 {
						return vector.before
					}
					return vector.after
				}
				var reference BackupReference
				if preUpgrade {
					reference, err = runtime.CreatePreUpgradeBackup(context.Background(), bundle)
				} else {
					reference, err = runtime.CreateRecoveryBackup(context.Background(), bundle)
				}
				if vector.accepted {
					if err != nil || reference != (BackupReference{ID: header.BackupID, Path: path}) {
						t.Fatalf("fresh intact backup refused: reference=%+v error=%v", reference, err)
					}
				} else if !errors.Is(err, ErrInvalid) || reference != (BackupReference{}) {
					t.Fatalf("backup outside command's timestamp window admitted: reference=%+v error=%v", reference, err)
				}
				if vector.commandMustNotStart && len(runner.calls) != 0 {
					t.Fatal("invalid invocation clock ran the backup command")
				}
			})
		}
	}
}
