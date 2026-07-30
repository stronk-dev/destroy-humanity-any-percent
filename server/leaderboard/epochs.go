package leaderboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

var (
	ErrInvalidEpoch  = errors.New("invalid balance epoch")
	artifactPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	changelogPattern = regexp.MustCompile(`^changelog/epoch-([0-9]+)\.md$`)
)

const epochMutationLock int64 = 0x434c4f554445504f

type Artifact struct {
	Name  string
	Bytes []byte
}

type Epoch struct {
	ID           int64
	Name         string
	StartedAt    time.Time
	EndedAt      *time.Time
	ChangelogRef string
	Hashes       []string
}

type Repository struct {
	db             *sql.DB
	repositoryRoot string
}

func NewRepository(db *sql.DB, repositoryRoot string) (*Repository, error) {
	if db == nil || repositoryRoot == "" {
		return nil, ErrInvalidEpoch
	}
	return &Repository{db: db, repositoryRoot: repositoryRoot}, nil
}

func (repository *Repository) MintEpoch(ctx context.Context, name string, startedAt time.Time, changelogRef string, artifacts []Artifact) (Epoch, error) {
	if strings.TrimSpace(name) == "" || startedAt.IsZero() || !changelogPattern.MatchString(changelogRef) {
		return Epoch{}, ErrInvalidEpoch
	}
	if err := ValidateChangelog(repository.repositoryRoot, changelogRef); err != nil {
		return Epoch{}, err
	}
	constantsHash, normalized, err := validateArtifacts(artifacts)
	if err != nil {
		return Epoch{}, err
	}
	startedAt = save.CanonicalServerTime(startedAt)
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Epoch{}, err
	}
	defer tx.Rollback()
	if err := lockEpochMutation(ctx, tx); err != nil {
		return Epoch{}, err
	}
	var currentID int64
	var currentStarted time.Time
	err = tx.QueryRowContext(ctx, `SELECT epoch_id,started_at FROM epochs WHERE ended_at IS NULL FOR UPDATE`).Scan(&currentID, &currentStarted)
	nextID := int64(1)
	if err == nil {
		nextID = currentID + 1
		if startedAt.Before(currentStarted) {
			return Epoch{}, ErrInvalidEpoch
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Epoch{}, err
	}
	if changelogRef != fmt.Sprintf("changelog/epoch-%d.md", nextID) {
		return Epoch{}, ErrInvalidEpoch
	}
	if currentID != 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE epochs SET ended_at=$2 WHERE epoch_id=$1`, currentID, startedAt); err != nil {
			return Epoch{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES($1,$2,$3,$4)`, nextID, name, startedAt, changelogRef); err != nil {
		return Epoch{}, err
	}
	if err := advanceEpochSequence(ctx, tx, nextID); err != nil {
		return Epoch{}, err
	}
	if err := insertCatalogSet(ctx, tx, constantsHash, normalized); err != nil {
		return Epoch{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, nextID, constantsHash); err != nil {
		return Epoch{}, err
	}
	if err := tx.Commit(); err != nil {
		return Epoch{}, err
	}
	return Epoch{ID: nextID, Name: name, StartedAt: startedAt, ChangelogRef: changelogRef, Hashes: []string{constantsHash}}, nil
}

func (repository *Repository) AddHotfix(ctx context.Context, constantsHash string, artifacts []Artifact) error {
	computed, normalized, err := validateArtifacts(artifacts)
	if err != nil || computed != constantsHash {
		return ErrInvalidEpoch
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockEpochMutation(ctx, tx); err != nil {
		return err
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id FROM epochs WHERE ended_at IS NULL FOR UPDATE`).Scan(&epochID); err != nil {
		return err
	}
	if err := insertCatalogSet(ctx, tx, constantsHash, normalized); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, constantsHash); err != nil {
		return err
	}
	return tx.Commit()
}

// ReconcileSeed makes the repository manifest and database agree before a
// gameserver can become ready. An empty database is reconstructed from the
// manifest's complete epoch/hash history; an existing database must remain a
// prefix of that history and may advance by only one deployed epoch.
func (repository *Repository) ReconcileSeed(ctx context.Context, bundle epochseed.Bundle, startedAt time.Time) error {
	if startedAt.IsZero() || bundle.Hash == "" || epochseed.Validate(bundle.Seed) != nil ||
		!epochseed.Accepts(epochseed.Current(bundle.Seed), bundle.Hash) {
		return ErrInvalidEpoch
	}
	artifacts := make([]Artifact, 0, len(bundle.Seed.Artifacts))
	seen := make(map[string]bool, len(bundle.Seed.Artifacts))
	for _, declaration := range bundle.Seed.Artifacts {
		data := bundle.Artifacts[declaration.Name]
		if len(data) == 0 || seen[declaration.Name] {
			return ErrInvalidEpoch
		}
		seen[declaration.Name] = true
		artifacts = append(artifacts, Artifact{Name: declaration.Name, Bytes: data})
	}
	if len(seen) != len(bundle.Artifacts) {
		return ErrInvalidEpoch
	}
	computed, normalized, err := validateArtifacts(artifacts)
	if err != nil || computed != bundle.Hash {
		return ErrInvalidEpoch
	}
	currentSeed := epochseed.Current(bundle.Seed)
	for _, declared := range bundle.Seed.Epochs {
		if err := ValidateChangelog(repository.repositoryRoot, declared.ChangelogRef); err != nil {
			return err
		}
	}
	startedAt = save.CanonicalServerTime(startedAt)
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockEpochMutation(ctx, tx); err != nil {
		return err
	}
	databaseEpochs, err := loadEpochRows(ctx, tx)
	if err != nil {
		return err
	}
	if len(databaseEpochs) > len(bundle.Seed.Epochs) {
		return ErrInvalidEpoch
	}
	for index, databaseEpoch := range databaseEpochs {
		declared := bundle.Seed.Epochs[index]
		hashesMatch := equalStrings(databaseEpoch.Hashes, declared.AcceptedHashes)
		if databaseEpoch.ID == currentSeed.ID {
			hashesMatch = stringSubset(databaseEpoch.Hashes, declared.AcceptedHashes)
		}
		if databaseEpoch.ID != declared.ID || databaseEpoch.Name != declared.Name || databaseEpoch.ChangelogRef != declared.ChangelogRef || !hashesMatch {
			return ErrInvalidEpoch
		}
	}
	currentDatabaseID := int64(0)
	if len(databaseEpochs) > 0 {
		currentDatabaseID = databaseEpochs[len(databaseEpochs)-1].ID
		if databaseEpochs[len(databaseEpochs)-1].EndedAt != nil {
			return ErrInvalidEpoch
		}
	}
	if currentDatabaseID == 0 {
		if err := bootstrapEpochHistory(ctx, tx, bundle.Seed, startedAt); err != nil {
			return err
		}
		if err := insertCatalogSet(ctx, tx, bundle.Hash, normalized); err != nil {
			return err
		}
		return tx.Commit()
	}
	switch {
	case currentDatabaseID == currentSeed.ID:
		// Current row already exists; the exact current bytes may still need an
		// idempotent catalog insert after a process/database restore.
	case currentDatabaseID+1 == currentSeed.ID:
		last := databaseEpochs[len(databaseEpochs)-1]
		if startedAt.Before(last.StartedAt) {
			return ErrInvalidEpoch
		}
		if _, err := tx.ExecContext(ctx, `UPDATE epochs SET ended_at=$2 WHERE epoch_id=$1`, currentDatabaseID, startedAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES($1,$2,$3,$4)`, currentSeed.ID, currentSeed.Name, startedAt, currentSeed.ChangelogRef); err != nil {
			return err
		}
		if err := advanceEpochSequence(ctx, tx, currentSeed.ID); err != nil {
			return err
		}
	default:
		return ErrInvalidEpoch
	}
	if err := insertDeclaredEpochHashes(ctx, tx, currentSeed); err != nil {
		return err
	}
	if err := insertCatalogSet(ctx, tx, bundle.Hash, normalized); err != nil {
		return err
	}
	hashes, err := loadEpochHashes(ctx, tx, currentSeed.ID)
	if err != nil || !equalStrings(hashes, currentSeed.AcceptedHashes) {
		return ErrInvalidEpoch
	}
	return tx.Commit()
}

func insertDeclaredEpochHashes(ctx context.Context, tx *sql.Tx, declared epochseed.Epoch) error {
	for _, hash := range declared.AcceptedHashes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, hash); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2) ON CONFLICT DO NOTHING`, declared.ID, hash); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapEpochHistory(ctx context.Context, tx *sql.Tx, seed epochseed.Seed, currentStartedAt time.Time) error {
	for _, declared := range seed.Epochs {
		for _, hash := range declared.AcceptedHashes {
			if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, hash); err != nil {
				return err
			}
		}
	}
	step := time.Millisecond
	firstStartedAt := currentStartedAt.Add(-time.Duration(len(seed.Epochs)-1) * step)
	for index, declared := range seed.Epochs {
		startedAt := firstStartedAt.Add(time.Duration(index) * step)
		var endedAt any
		if index+1 < len(seed.Epochs) {
			endedAt = startedAt.Add(step)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO epochs(epoch_id,name,started_at,ended_at,changelog_ref) VALUES($1,$2,$3,$4,$5)`,
			declared.ID, declared.Name, startedAt, endedAt, declared.ChangelogRef); err != nil {
			return err
		}
		if err := insertDeclaredEpochHashes(ctx, tx, declared); err != nil {
			return err
		}
	}
	return advanceEpochSequence(ctx, tx, seed.CurrentEpochID)
}

func (repository *Repository) Current(ctx context.Context) (Epoch, error) {
	var epoch Epoch
	if err := repository.db.QueryRowContext(ctx, `SELECT epoch_id,name,started_at,changelog_ref FROM epochs WHERE ended_at IS NULL`).Scan(&epoch.ID, &epoch.Name, &epoch.StartedAt, &epoch.ChangelogRef); err != nil {
		return Epoch{}, err
	}
	rows, err := repository.db.QueryContext(ctx, `SELECT constants_hash FROM epoch_hashes WHERE epoch_id=$1 ORDER BY constants_hash`, epoch.ID)
	if err != nil {
		return Epoch{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return Epoch{}, err
		}
		epoch.Hashes = append(epoch.Hashes, hash)
	}
	return epoch, rows.Err()
}

func ValidateChangelog(repositoryRoot, reference string) error {
	if !changelogPattern.MatchString(reference) || repositoryRoot == "" {
		return ErrInvalidEpoch
	}
	path := filepath.Join(repositoryRoot, filepath.FromSlash(reference))
	data, err := os.ReadFile(path)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return ErrInvalidEpoch
	}
	return nil
}

func validateArtifacts(artifacts []Artifact) (string, map[string][]byte, error) {
	if len(artifacts) == 0 {
		return "", nil, ErrInvalidEpoch
	}
	normalized := make(map[string][]byte, len(artifacts))
	for _, artifact := range artifacts {
		if !artifactPattern.MatchString(artifact.Name) || len(artifact.Bytes) == 0 || normalized[artifact.Name] != nil {
			return "", nil, ErrInvalidEpoch
		}
		normalized[artifact.Name] = append([]byte(nil), artifact.Bytes...)
	}
	hash, err := save.ConstantsHashArtifacts(normalized)
	if err != nil {
		return "", nil, ErrInvalidEpoch
	}
	return hash, normalized, nil
}

func insertCatalogSet(ctx context.Context, tx *sql.Tx, constantsHash string, artifacts map[string][]byte) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, constantsHash); err != nil {
		return err
	}
	names := make([]string, 0, len(artifacts))
	for name := range artifacts {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		result, err := tx.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, constantsHash, name, artifacts[name])
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 0 {
			var existing []byte
			if err := tx.QueryRowContext(ctx, `SELECT bytes FROM catalog_artifacts WHERE constants_hash=$1 AND artifact_name=$2`, constantsHash, name).Scan(&existing); err != nil || string(existing) != string(artifacts[name]) {
				return ErrInvalidEpoch
			}
		}
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, constantsHash).Scan(&count); err != nil || count != len(artifacts) {
		return ErrInvalidEpoch
	}
	return nil
}

func lockEpochMutation(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, epochMutationLock)
	return err
}

func advanceEpochSequence(ctx context.Context, tx *sql.Tx, epochID int64) error {
	_, err := tx.ExecContext(ctx, `SELECT setval(pg_get_serial_sequence('epochs','epoch_id'),$1,true)`, epochID)
	return err
}

func loadEpochRows(ctx context.Context, tx *sql.Tx) ([]Epoch, error) {
	rows, err := tx.QueryContext(ctx, `SELECT epoch_id,name,started_at,ended_at,changelog_ref FROM epochs ORDER BY epoch_id`)
	if err != nil {
		return nil, err
	}
	var epochs []Epoch
	for rows.Next() {
		var epoch Epoch
		if err := rows.Scan(&epoch.ID, &epoch.Name, &epoch.StartedAt, &epoch.EndedAt, &epoch.ChangelogRef); err != nil {
			rows.Close()
			return nil, err
		}
		epochs = append(epochs, epoch)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for index := range epochs {
		hashes, err := loadEpochHashes(ctx, tx, epochs[index].ID)
		if err != nil {
			return nil, err
		}
		epochs[index].Hashes = hashes
	}
	return epochs, nil
}

func loadEpochHashes(ctx context.Context, tx *sql.Tx, epochID int64) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT constants_hash FROM epoch_hashes WHERE epoch_id=$1 ORDER BY constants_hash`, epochID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, rows.Err()
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func stringSubset(subset, set []string) bool {
	for _, value := range subset {
		index := sort.SearchStrings(set, value)
		if index >= len(set) || set[index] != value {
			return false
		}
	}
	return true
}
