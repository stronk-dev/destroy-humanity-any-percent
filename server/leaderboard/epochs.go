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

	"cloud-clicker/server/save"
)

var (
	ErrInvalidEpoch  = errors.New("invalid balance epoch")
	artifactPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	changelogPattern = regexp.MustCompile(`^changelog/epoch-([0-9]+)\.md$`)
)

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
	var currentID int64
	var currentStarted time.Time
	err = tx.QueryRowContext(ctx, `SELECT epoch_id,started_at FROM epochs WHERE ended_at IS NULL FOR UPDATE`).Scan(&currentID, &currentStarted)
	if err == nil {
		if startedAt.Before(currentStarted) {
			return Epoch{}, ErrInvalidEpoch
		}
		if _, err := tx.ExecContext(ctx, `UPDATE epochs SET ended_at=$2 WHERE epoch_id=$1`, currentID, startedAt); err != nil {
			return Epoch{}, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Epoch{}, err
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES($1,$2,$3) RETURNING epoch_id`, name, startedAt, changelogRef).Scan(&epochID); err != nil {
		return Epoch{}, err
	}
	if changelogRef != fmt.Sprintf("changelog/epoch-%d.md", epochID) {
		return Epoch{}, ErrInvalidEpoch
	}
	if err := insertCatalogSet(ctx, tx, constantsHash, normalized); err != nil {
		return Epoch{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, constantsHash); err != nil {
		return Epoch{}, err
	}
	if err := tx.Commit(); err != nil {
		return Epoch{}, err
	}
	return Epoch{ID: epochID, Name: name, StartedAt: startedAt, ChangelogRef: changelogRef, Hashes: []string{constantsHash}}, nil
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
	var epochID int64
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id FROM epochs WHERE ended_at IS NULL FOR SHARE`).Scan(&epochID); err != nil {
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
