package deploymentrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/releasepackage"
)

const RollbackWindow = 7 * 24 * time.Hour

type Bundle struct {
	Root           string
	Manifest       releasepackage.ReleaseManifest
	ManifestSHA256 string
}

type BackupReference struct {
	ID   string
	Path string
}

type DrainEvidence struct {
	ReadinessDown    bool
	CourtesyFrame    bool
	IntentsRefused   bool
	AdmittedComplete bool
	JobsFlushed      bool
	SocketsClosed    bool
	WithinBound      bool
}

func (evidence DrainEvidence) Valid() bool {
	return evidence.ReadinessDown && evidence.CourtesyFrame && evidence.IntentsRefused && evidence.AdmittedComplete &&
		evidence.JobsFlushed && evidence.SocketsClosed && evidence.WithinBound
}

type Runtime interface {
	Preflight(context.Context, Bundle) error
	CreatePreUpgradeBackup(context.Context, Bundle) (BackupReference, error)
	Prepare(context.Context, Bundle) error
	DrainCurrent(context.Context, Bundle) (DrainEvidence, error)
	Start(context.Context, Bundle) error
	VerifyIdentity(context.Context, Bundle) error
	AuthenticatedSmoke(context.Context, Bundle) error
	StopFailed(context.Context, Bundle) error
	ResetDatabase(context.Context, Bundle) error
	Restore(context.Context, Bundle, BackupReference) error
	// VerifyRestoreInputs proves, without mutating runtime state, that the
	// backup and restore credential can drive Restore for this bundle.
	VerifyRestoreInputs(context.Context, Bundle, BackupReference) error
}

type InstallRuntime interface {
	Runtime
	PreflightInstall(context.Context, Bundle) error
	StartInstall(context.Context, Bundle) error
	AbortInstall(context.Context, Bundle) error
}

type Controller struct {
	Runtime    Runtime
	LedgerPath string
	Operator   string
	Now        func() time.Time
	Load       func(string) (Bundle, error)
}

type ReleaseRequest struct {
	CurrentBundle   string
	CandidateBundle string
}

type InstallRequest struct {
	Bundle string
}

type RollbackRequest struct {
	FailedBundle   string
	PreviousBundle string
	Backup         BackupReference
}

func (controller Controller) Install(ctx context.Context, request InstallRequest) error {
	started, err := controller.validate()
	if err != nil {
		return err
	}
	runtime, ok := controller.Runtime.(InstallRuntime)
	if !ok {
		return ErrInvalid
	}
	lock, err := acquireOperatorLock(controller.LedgerPath + ".lock")
	if err != nil {
		return err
	}
	defer lock.Close()
	bundle, err := controller.loadBundle(request.Bundle)
	if err != nil {
		return controller.fail(rejectedInputRecord("install", request.Bundle, controller.Operator, started), "bundle", err)
	}
	base := controller.baseRecord("install", bundle, Bundle{}, started)
	records, ledgerErr := ReadReleaseLedger(controller.LedgerPath)
	if ledgerErr == nil {
		for _, record := range records {
			if record.Action != "install" || record.Result != "failed" {
				return controller.fail(base, "install_authority", ErrInvalid)
			}
		}
	} else if !errors.Is(ledgerErr, os.ErrNotExist) {
		return errors.Join(ErrInvalid, ledgerErr)
	}
	if err := runtime.Prepare(ctx, bundle); err != nil {
		return controller.fail(base, "supply_chain", err)
	}
	if err := runtime.PreflightInstall(ctx, bundle); err != nil {
		return controller.fail(base, "preflight", err)
	}
	if err := runtime.StartInstall(ctx, bundle); err != nil {
		return controller.fail(base, "startup_migration", errors.Join(err, runtime.AbortInstall(ctx, bundle)))
	}
	if err := runtime.VerifyIdentity(ctx, bundle); err != nil {
		return controller.fail(base, "epoch_artifact_identity", errors.Join(err, runtime.AbortInstall(ctx, bundle)))
	}
	if err := runtime.AuthenticatedSmoke(ctx, bundle); err != nil {
		return controller.fail(base, "authenticated_smoke", errors.Join(err, runtime.AbortInstall(ctx, bundle)))
	}
	base.CompletedAt = controller.Now().UTC()
	base.Result = "succeeded"
	if err := AppendReleaseRecord(controller.LedgerPath, base); err != nil {
		return errors.Join(err, runtime.AbortInstall(ctx, bundle))
	}
	return nil
}

func LoadBundle(root string) (Bundle, error) {
	if err := releasepackage.ValidateBundle(root); err != nil {
		return Bundle{}, errors.Join(ErrInvalid, err)
	}
	data, err := os.ReadFile(filepath.Join(root, releasepackage.ReleaseManifestPath))
	if err != nil {
		return Bundle{}, err
	}
	var manifest releasepackage.ReleaseManifest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&manifest) != nil || decoder.Decode(&struct{}{}) != io.EOF || releasepackage.ValidateReleaseManifest(manifest) != nil {
		return Bundle{}, ErrInvalid
	}
	sum := sha256.Sum256(data)
	return Bundle{Root: root, Manifest: manifest, ManifestSHA256: "sha256:" + hex.EncodeToString(sum[:])}, nil
}

func (controller Controller) Release(ctx context.Context, request ReleaseRequest) error {
	started, err := controller.validate()
	if err != nil {
		return err
	}
	lock, err := acquireOperatorLock(controller.LedgerPath + ".lock")
	if err != nil {
		return err
	}
	defer lock.Close()
	candidate, err := controller.loadBundle(request.CandidateBundle)
	if err != nil {
		return controller.fail(rejectedInputRecord("release", request.CandidateBundle, controller.Operator, started), "candidate_bundle", err)
	}
	current, err := controller.loadBundle(request.CurrentBundle)
	if err != nil {
		record := controller.baseRecord("release", candidate, Bundle{}, started)
		record.PreviousVersion, record.PreviousManifestSHA256 = "", ""
		return controller.fail(record, "current_bundle", err)
	}
	base := controller.baseRecord("release", candidate, current, started)
	versionOrder, versionErr := releasepackage.CompareReleaseVersions(candidate.Manifest.ReleaseVersion, current.Manifest.ReleaseVersion)
	if versionErr != nil || versionOrder <= 0 || candidate.Manifest.DatabaseMigration < current.Manifest.DatabaseMigration {
		return controller.fail(base, "compatibility", ErrInvalid)
	}
	// The candidate config preflight runs inside the candidate image, which a
	// bare config-ID reference can only resolve after the bundle archive loads.
	if err := controller.Runtime.Prepare(ctx, candidate); err != nil {
		return controller.fail(base, "supply_chain", err)
	}
	if err := controller.Runtime.Preflight(ctx, candidate); err != nil {
		return controller.fail(base, "preflight", err)
	}
	backup, err := controller.Runtime.CreatePreUpgradeBackup(ctx, current)
	if err != nil || !backupIDPattern.MatchString(backup.ID) || backup.Path == "" {
		return controller.fail(base, "preupgrade_backup", errors.Join(ErrInvalid, err))
	}
	base.BackupID = backup.ID
	evidence, err := controller.Runtime.DrainCurrent(ctx, current)
	if err != nil || !evidence.Valid() {
		return controller.fail(base, "bounded_drain", errors.Join(ErrInvalid, err))
	}
	if err := controller.Runtime.Start(ctx, candidate); err != nil {
		return controller.fail(base, "startup_migration", err)
	}
	if err := controller.Runtime.VerifyIdentity(ctx, candidate); err != nil {
		return controller.fail(base, "epoch_artifact_identity", err)
	}
	if err := controller.Runtime.AuthenticatedSmoke(ctx, candidate); err != nil {
		return controller.fail(base, "authenticated_smoke", err)
	}
	base.CompletedAt = controller.Now().UTC()
	base.RollbackUntil = base.CompletedAt.Add(RollbackWindow)
	base.Result = "succeeded"
	return AppendReleaseRecord(controller.LedgerPath, base)
}

func (controller Controller) Rollback(ctx context.Context, request RollbackRequest) error {
	started, err := controller.validate()
	if err != nil {
		return err
	}
	lock, err := acquireOperatorLock(controller.LedgerPath + ".lock")
	if err != nil {
		return err
	}
	defer lock.Close()
	failed, err := controller.loadBundle(request.FailedBundle)
	if err != nil {
		return controller.fail(rejectedInputRecord("rollback", request.FailedBundle, controller.Operator, started), "failed_bundle", err)
	}
	previous, err := controller.loadBundle(request.PreviousBundle)
	if err != nil {
		record := controller.baseRecord("rollback", failed, Bundle{}, started)
		record.PreviousVersion, record.PreviousManifestSHA256 = "", ""
		return controller.fail(record, "previous_bundle", err)
	}
	base := controller.baseRecord("rollback", previous, failed, started)
	base.BackupID = request.Backup.ID
	records, err := ReadReleaseLedger(controller.LedgerPath)
	if err != nil {
		return err
	}
	approved, authorized := rollbackAuthority(records, failed, previous, request.Backup)
	if !authorized || approved.Result != "succeeded" && approved.Result != "failed" || approved.RollbackUntil.IsZero() || approved.ReleaseVersion != failed.Manifest.ReleaseVersion ||
		approved.ManifestSHA256 != failed.ManifestSHA256 || approved.PreviousVersion != previous.Manifest.ReleaseVersion ||
		approved.PreviousManifestSHA256 != previous.ManifestSHA256 || approved.BackupID != request.Backup.ID || request.Backup.Path == "" ||
		started.After(approved.RollbackUntil) {
		return controller.fail(base, "rollback_authority", ErrInvalid)
	}
	// Every rollback input is proved before the failed release stops or the
	// live database volume is removed.
	if err := controller.Runtime.Prepare(ctx, previous); err != nil {
		return controller.fail(base, "supply_chain", err)
	}
	if err := controller.Runtime.Preflight(ctx, previous); err != nil {
		return controller.fail(base, "preflight", err)
	}
	if err := controller.Runtime.VerifyRestoreInputs(ctx, previous, request.Backup); err != nil {
		return controller.fail(base, "restore_inputs", err)
	}
	if err := controller.Runtime.StopFailed(ctx, failed); err != nil {
		return controller.fail(base, "stop_failed_release", err)
	}
	if err := controller.Runtime.ResetDatabase(ctx, previous); err != nil {
		return controller.fail(base, "clean_database", err)
	}
	if err := controller.Runtime.Restore(ctx, previous, request.Backup); err != nil {
		return controller.fail(base, "exact_restore", err)
	}
	if err := controller.Runtime.Start(ctx, previous); err != nil {
		return controller.fail(base, "previous_startup", err)
	}
	if err := controller.Runtime.VerifyIdentity(ctx, previous); err != nil {
		return controller.fail(base, "epoch_artifact_identity", err)
	}
	if err := controller.Runtime.AuthenticatedSmoke(ctx, previous); err != nil {
		return controller.fail(base, "authenticated_smoke", err)
	}
	base.CompletedAt = controller.Now().UTC()
	base.Result = "succeeded"
	base.PreviousVersion, base.PreviousManifestSHA256 = "", ""
	base.RollbackUntil = time.Time{}
	return AppendReleaseRecord(controller.LedgerPath, base)
}

func rollbackAuthority(records []ReleaseRecord, failed, previous Bundle, backup BackupReference) (ReleaseRecord, bool) {
	for index := len(records) - 1; index >= 0; index-- {
		record := records[index]
		if record.Action == "release" {
			return record, true
		}
		// A failed attempt against this exact authority may be retried. Any
		// successful or unrelated intervening rollback closes the chain.
		if record.Result != "failed" || record.ReleaseVersion != previous.Manifest.ReleaseVersion ||
			record.ManifestSHA256 != previous.ManifestSHA256 || record.PreviousVersion != failed.Manifest.ReleaseVersion ||
			record.PreviousManifestSHA256 != failed.ManifestSHA256 || record.BackupID != backup.ID {
			return ReleaseRecord{}, false
		}
	}
	return ReleaseRecord{}, false
}

func (controller Controller) validate() (time.Time, error) {
	if controller.Runtime == nil || controller.LedgerPath == "" || !identifier.MatchString(controller.Operator) || controller.Now == nil {
		return time.Time{}, ErrInvalid
	}
	now := controller.Now().UTC()
	if now.IsZero() {
		return time.Time{}, ErrInvalid
	}
	return now, nil
}

func (controller Controller) loadBundle(path string) (Bundle, error) {
	if controller.Load != nil {
		return controller.Load(path)
	}
	return LoadBundle(path)
}

func (controller Controller) baseRecord(action string, target, previous Bundle, started time.Time) ReleaseRecord {
	digests := make([]string, len(target.Manifest.Images))
	for index := range target.Manifest.Images {
		digests[index] = target.Manifest.Images[index].Reference
	}
	return ReleaseRecord{SchemaVersion: 1, Action: action, ReleaseVersion: target.Manifest.ReleaseVersion,
		ManifestSHA256: target.ManifestSHA256, PreviousVersion: previous.Manifest.ReleaseVersion,
		PreviousManifestSHA256: previous.ManifestSHA256, ImageDigests: digests, BackupID: "none",
		StartedAt: started, Operator: controller.Operator}
}

func (controller Controller) fail(record ReleaseRecord, stage string, cause error) error {
	record.CompletedAt = controller.Now().UTC()
	if record.CompletedAt.Before(record.StartedAt) {
		record.CompletedAt = record.StartedAt
	}
	record.Result, record.FailureStage = "failed", stage
	if record.Action == "release" && backupIDPattern.MatchString(record.BackupID) && record.PreviousVersion != "" {
		record.RollbackUntil = record.CompletedAt.Add(RollbackWindow)
	}
	return errors.Join(cause, AppendReleaseRecord(controller.LedgerPath, record))
}

func rejectedInputRecord(action, path, operator string, started time.Time) ReleaseRecord {
	data, err := os.ReadFile(filepath.Join(path, releasepackage.ReleaseManifestPath))
	if err != nil {
		data = nil
	}
	sum := sha256.Sum256(data)
	version := "invalid"
	var partial struct {
		ReleaseVersion string                 `json:"release_version"`
		Images         []releasepackage.Image `json:"images"`
	}
	_ = json.Unmarshal(data, &partial)
	if identifier.MatchString(partial.ReleaseVersion) {
		version = partial.ReleaseVersion
	}
	digests := []string(nil)
	if len(partial.Images) == 6 {
		candidate := make([]string, 6)
		valid := true
		for index := range partial.Images {
			candidate[index] = partial.Images[index].Reference
			valid = valid && imageDigestPattern.MatchString(candidate[index])
		}
		if valid {
			digests = candidate
		}
	}
	return ReleaseRecord{SchemaVersion: 1, Action: action, ReleaseVersion: version,
		ManifestSHA256: "sha256:" + hex.EncodeToString(sum[:]), ImageDigests: digests,
		BackupID: "none", StartedAt: started, Operator: operator}
}

func (bundle Bundle) String() string {
	return fmt.Sprintf("%s (%s)", bundle.Manifest.ReleaseVersion, bundle.ManifestSHA256)
}
