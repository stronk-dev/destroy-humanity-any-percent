package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type recoveryRuntimeFixture struct {
	identities []deploymentbackup.RecoveryIdentity
	backups    []deploymentrelease.BackupReference
	restored   []deploymentrelease.BackupReference
	stops      int
	resets     int
	starts     int
	verifies   int
	smokes     int
	failAt     string
}

func (runtime *recoveryRuntimeFixture) InspectRecoveryIdentity(context.Context, deploymentrelease.Bundle) (deploymentbackup.RecoveryIdentity, error) {
	if runtime.failAt == "inspect" {
		return deploymentbackup.RecoveryIdentity{}, errors.New("injected inspect failure")
	}
	if len(runtime.identities) == 0 {
		return deploymentbackup.RecoveryIdentity{}, errors.New("unexpected identity inspection")
	}
	identity := runtime.identities[0]
	runtime.identities = runtime.identities[1:]
	return identity, nil
}
func (runtime *recoveryRuntimeFixture) CreateRecoveryBackup(context.Context, deploymentrelease.Bundle) (deploymentrelease.BackupReference, error) {
	if runtime.failAt == "backup" {
		return deploymentrelease.BackupReference{}, errors.New("injected backup failure")
	}
	if len(runtime.backups) == 0 {
		return deploymentrelease.BackupReference{}, errors.New("unexpected backup")
	}
	backup := runtime.backups[0]
	runtime.backups = runtime.backups[1:]
	return backup, nil
}
func (runtime *recoveryRuntimeFixture) StopFailed(context.Context, deploymentrelease.Bundle) error {
	runtime.stops++
	if runtime.failAt == "stop" {
		return errors.New("injected stop failure")
	}
	return nil
}
func (runtime *recoveryRuntimeFixture) ResetDatabase(context.Context, deploymentrelease.Bundle) error {
	runtime.resets++
	if runtime.failAt == "reset" {
		return errors.New("injected reset failure")
	}
	return nil
}
func (runtime *recoveryRuntimeFixture) StartRecoveryCore(context.Context, deploymentrelease.Bundle) error {
	runtime.starts++
	if runtime.failAt == "start" {
		return errors.New("injected start failure")
	}
	return nil
}
func (runtime *recoveryRuntimeFixture) VerifyIdentity(context.Context, deploymentrelease.Bundle) error {
	runtime.verifies++
	if runtime.failAt == "verify" {
		return errors.New("injected verify failure")
	}
	return nil
}
func (runtime *recoveryRuntimeFixture) RestoreRecoveryBackup(_ context.Context, _ deploymentrelease.Bundle, backup deploymentrelease.BackupReference) error {
	runtime.restored = append(runtime.restored, backup)
	if runtime.failAt == "restore" {
		return errors.New("injected restore failure")
	}
	return nil
}
func (runtime *recoveryRuntimeFixture) AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error {
	runtime.smokes++
	if runtime.failAt == "smoke" {
		return errors.New("injected smoke failure")
	}
	return nil
}

func TestRecoveryProducerRestoresExactEmptyAndPopulatedIdentities(t *testing.T) {
	config, bundle, dependencies, runtime := recoveryProducerFixture(t)
	checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.ManifestSHA256 != bundle.ManifestSHA256 || len(runtime.restored) != 1 || runtime.restored[0].ID != checkpoint.EmptyBackup.ID {
		t.Fatalf("checkpoint=%+v restored=%+v", checkpoint, runtime.restored)
	}

	runtime.identities = []deploymentbackup.RecoveryIdentity{checkpoint.PopulatedIdentity}
	dependencies.now = sequenceClock(time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC), time.Date(2026, 8, 23, 12, 6, 0, 0, time.UTC))
	observation, err := runPopulatedRecovery(context.Background(), config, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	if observation.Objectives.RPOSeconds != 60 || observation.Objectives.RTOSeconds != 60 || !observation.Objectives.RestoredIdentityMatch ||
		len(runtime.restored) != 2 || runtime.restored[1].ID != checkpoint.PopulatedBackup.ID || runtime.smokes != 1 {
		t.Fatalf("observation=%+v runtime=%+v", observation, runtime)
	}
	if _, err := DecodeObjectiveObservation(mustRead(t, filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["objective_observation"]))); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryProducerRejectsWrongPopulationAndIdentityMismatch(t *testing.T) {
	t.Run("empty initial population", func(t *testing.T) {
		config, _, dependencies, runtime := recoveryProducerFixture(t)
		runtime.identities[0] = recoveryIdentity(false, "a")
		if _, err := runEmptyRecovery(context.Background(), config, dependencies); !errors.Is(err, ErrInvalid) {
			t.Fatalf("empty initial population accepted: %v", err)
		}
		if runtime.stops != 0 || runtime.resets != 0 {
			t.Fatalf("invalid population reached destructive phase: %+v", runtime)
		}
	})
	t.Run("populated mismatch", func(t *testing.T) {
		config, _, dependencies, runtime := recoveryProducerFixture(t)
		checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
		if err != nil {
			t.Fatal(err)
		}
		mismatch := checkpoint.PopulatedIdentity
		mismatch.Company.SHA256 = hashForBuild("f")
		runtime.identities = []deploymentbackup.RecoveryIdentity{mismatch}
		dependencies.now = sequenceClock(time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC))
		if _, err := runPopulatedRecovery(context.Background(), config, dependencies); !errors.Is(err, ErrInvalid) {
			t.Fatalf("mismatched populated identity accepted: %v", err)
		}
		if runtime.smokes != 0 {
			t.Fatal("identity mismatch reached authenticated smoke")
		}
	})
}

func TestEmptyRecoveryFailsClosedAtEveryRuntimeBoundary(t *testing.T) {
	for _, stage := range []string{"inspect", "backup", "stop", "reset", "start", "verify", "restore"} {
		t.Run(stage, func(t *testing.T) {
			config, _, dependencies, runtime := recoveryProducerFixture(t)
			runtime.failAt = stage
			if _, err := runEmptyRecovery(context.Background(), config, dependencies); err == nil {
				t.Fatalf("%s failure accepted", stage)
			}
			if _, err := os.Lstat(filepath.Join(config.WorkDirectory, recoveryCheckpointName)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("%s failure committed checkpoint: %v", stage, err)
			}
		})
	}
}

func TestRecoveryProducerRejectsCheckpointAndObjectiveForgery(t *testing.T) {
	config, _, dependencies, runtime := recoveryProducerFixture(t)
	checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(config.WorkDirectory, recoveryCheckpointName)
	data := mustRead(t, path)
	if err := os.WriteFile(path, append(data, []byte("{}")...), 0o600); err != nil {
		t.Fatal(err)
	}
	runtime.identities = []deploymentbackup.RecoveryIdentity{checkpoint.PopulatedIdentity}
	if _, err := runPopulatedRecovery(context.Background(), config, dependencies); !errors.Is(err, ErrInvalid) {
		t.Fatalf("trailing checkpoint accepted: %v", err)
	}
}

func TestRecoveryProducerRejectsObjectiveOverrunAndOverwriteBeforeMutation(t *testing.T) {
	t.Run("rto overrun", func(t *testing.T) {
		config, _, dependencies, runtime := recoveryProducerFixture(t)
		checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
		if err != nil {
			t.Fatal(err)
		}
		runtime.identities = []deploymentbackup.RecoveryIdentity{checkpoint.PopulatedIdentity}
		dependencies.now = sequenceClock(time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC), time.Date(2026, 8, 23, 16, 5, 1, 0, time.UTC))
		if _, err := runPopulatedRecovery(context.Background(), config, dependencies); err == nil {
			t.Fatal("RTO above four hours accepted")
		}
	})
	t.Run("output overwrite", func(t *testing.T) {
		config, _, dependencies, runtime := recoveryProducerFixture(t)
		if _, err := runEmptyRecovery(context.Background(), config, dependencies); err != nil {
			t.Fatal(err)
		}
		output := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["objective_observation"])
		if err := os.WriteFile(output, []byte("existing"), 0o600); err != nil {
			t.Fatal(err)
		}
		stops := runtime.stops
		if _, err := runPopulatedRecovery(context.Background(), config, dependencies); !errors.Is(err, ErrInvalid) {
			t.Fatalf("objective overwrite accepted: %v", err)
		}
		if runtime.stops != stops {
			t.Fatal("objective overwrite reached destructive recovery")
		}
	})
	t.Run("smoke failure", func(t *testing.T) {
		config, _, dependencies, runtime := recoveryProducerFixture(t)
		checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
		if err != nil {
			t.Fatal(err)
		}
		runtime.identities = []deploymentbackup.RecoveryIdentity{checkpoint.PopulatedIdentity}
		runtime.failAt = "smoke"
		dependencies.now = sequenceClock(time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC))
		if _, err := runPopulatedRecovery(context.Background(), config, dependencies); err == nil {
			t.Fatal("failed authenticated smoke accepted")
		}
		if _, err := os.Lstat(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["objective_observation"])); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("failed smoke committed objective: %v", err)
		}
	})
}

func recoveryProducerFixture(t *testing.T) (ScenarioConfig, deploymentrelease.Bundle, recoveryDependencies, *recoveryRuntimeFixture) {
	t.Helper()
	config := validScenarioConfig(t)
	bundle := deploymentrelease.Bundle{Root: config.CandidateBundle, ManifestSHA256: hashForBuild("a"), Manifest: releasepackage.ReleaseManifest{
		ReleaseVersion: "1.0.0", DatabaseMigration: 74, EpochID: 8, ConstantsHash: hashForBuild("b")}}
	populated := recoveryIdentity(true, "c")
	empty := recoveryIdentity(false, "d")
	populatedHeader := recoveryHeader(config, bundle, "20260823T120000Z-abcdef123456",
		time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC), time.Date(2026, 8, 23, 12, 1, 0, 0, time.UTC))
	emptyHeader := recoveryHeader(config, bundle, "20260823T120300Z-abcdef123457",
		time.Date(2026, 8, 23, 12, 3, 0, 0, time.UTC), time.Date(2026, 8, 23, 12, 4, 0, 0, time.UTC))
	headers := map[string]deploymentbackup.Header{}
	backups := make([]deploymentrelease.BackupReference, 0, 2)
	for _, header := range []deploymentbackup.Header{populatedHeader, emptyHeader} {
		path := filepath.Join(config.BackupTarget, header.BackupID+".ccbackup")
		if err := os.WriteFile(path, []byte("encrypted fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
		headers[path] = header
		backups = append(backups, deploymentrelease.BackupReference{ID: header.BackupID, Path: path})
	}
	runtime := &recoveryRuntimeFixture{identities: []deploymentbackup.RecoveryIdentity{populated, empty, empty}, backups: backups}
	dependencies := recoveryDependencies{loadBundle: func(string) (deploymentrelease.Bundle, error) { return bundle, nil },
		readHeader: func(path string) (deploymentbackup.Header, error) { return headers[path], nil }, runtime: runtime,
		now: sequenceClock(time.Date(2026, 8, 23, 12, 2, 0, 0, time.UTC))}
	return config, bundle, dependencies, runtime
}

func recoveryIdentity(populated bool, value string) deploymentbackup.RecoveryIdentity {
	domain := deploymentbackup.RecoveryDomainIdentity{Rows: 1, SHA256: hashForBuild(value)}
	empty := deploymentbackup.RecoveryDomainIdentity{SHA256: hashForBuild(value)}
	identity := deploymentbackup.RecoveryIdentity{SchemaVersion: 1, DatabaseMigration: 74, Database: domain, Epoch: domain,
		Player: empty, Founder: empty, Company: empty, Events: empty, Board: empty}
	if populated {
		identity.Player, identity.Founder, identity.Company, identity.Events, identity.Board = domain, domain, domain, domain, domain
	}
	return identity
}

func recoveryHeader(config ScenarioConfig, bundle deploymentrelease.Bundle, id string, started, completed time.Time) deploymentbackup.Header {
	return deploymentbackup.Header{SchemaVersion: 1, BackupID: id, ServerID: config.ServerID,
		ReleaseManifestSHA256: bundle.ManifestSHA256, EpochID: bundle.Manifest.EpochID, StartedAt: started,
		CompletedAt: completed, PayloadSHA256: hashForBuild("e"), PayloadBytes: 1024}
}

func sequenceClock(values ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}

func TestRecoveryCheckpointRejectsWrongBackupClass(t *testing.T) {
	config, bundle, dependencies, _ := recoveryProducerFixture(t)
	checkpoint, err := runEmptyRecovery(context.Background(), config, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint.PopulatedBackup.Header.PreUpgrade = true
	if validateRecoveryCheckpoint(checkpoint, config, bundle) == nil {
		t.Fatal("pre-upgrade backup accepted as recovery checkpoint")
	}
	checkpoint.PopulatedBackup.Header.PreUpgrade = false
	checkpoint.PopulatedBackup.Path = filepath.Join(config.BackupTarget, "wrong.ccbackup")
	if validateRecoveryCheckpoint(checkpoint, config, bundle) == nil {
		t.Fatal("wrong recovery backup path accepted")
	}
}

var _ recoveryRuntime = (*recoveryRuntimeFixture)(nil)
