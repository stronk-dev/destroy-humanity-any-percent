package deploymentbackup

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"cloud-clicker/server/epochseed"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresInspectionInput struct {
	DatabaseURLFile string
	ReleaseManifest string
	ContentRoot     string
	RequireIdentity bool
}

type PostgresInspection struct {
	SchemaVersion     int    `json:"schema_version"`
	DatabaseBytes     int64  `json:"database_bytes"`
	DatabaseMigration int    `json:"database_migration"`
	EpochID           int64  `json:"epoch_id"`
	ConstantsHash     string `json:"constants_hash"`
	ArtifactsVerified int    `json:"artifacts_verified"`
}

// InspectPostgres runs from the private backup service. It is the production
// boundary for database sizing and post-start manifest/epoch/artifact checks;
// the host-side release helper never needs a routable Postgres endpoint.
func InspectPostgres(ctx context.Context, input PostgresInspectionInput) (PostgresInspection, error) {
	if input.DatabaseURLFile == "" || input.ReleaseManifest == "" || input.ContentRoot == "" {
		return PostgresInspection{}, ErrInvalid
	}
	manifestBytes, err := os.ReadFile(input.ReleaseManifest)
	if err != nil {
		return PostgresInspection{}, err
	}
	epochBytes, err := os.ReadFile(filepath.Join(input.ContentRoot, filepath.FromSlash(epochseed.Path)))
	if err != nil {
		return PostgresInspection{}, err
	}
	manifest, _, _, _, err := validateReleaseInputs(manifestBytes, epochBytes)
	if err != nil {
		return PostgresInspection{}, err
	}
	bundle, err := epochseed.Load(input.ContentRoot)
	if err != nil || bundle.Seed.CurrentEpochID != manifest.EpochID || bundle.Hash != manifest.ConstantsHash {
		return PostgresInspection{}, errors.Join(ErrInvalid, err)
	}
	databaseURL, err := readSecretFile(input.DatabaseURLFile)
	if err != nil {
		return PostgresInspection{}, err
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return PostgresInspection{}, err
	}
	defer database.Close()
	if err := database.PingContext(ctx); err != nil {
		return PostgresInspection{}, err
	}
	result := PostgresInspection{SchemaVersion: 1, EpochID: manifest.EpochID, ConstantsHash: manifest.ConstantsHash}
	if err := database.QueryRowContext(ctx, `SELECT pg_database_size(current_database()), COALESCE(max(version_id) FILTER (WHERE is_applied),0) FROM goose_db_version`).Scan(&result.DatabaseBytes, &result.DatabaseMigration); err != nil || result.DatabaseBytes < 1 {
		return PostgresInspection{}, errors.Join(ErrInvalid, err)
	}
	if !input.RequireIdentity {
		return result, nil
	}
	if result.DatabaseMigration != manifest.DatabaseMigration {
		return PostgresInspection{}, ErrInvalid
	}
	var current bool
	if err := database.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM epochs e JOIN epoch_hashes h USING(epoch_id) WHERE e.ended_at IS NULL AND e.epoch_id=$1 AND h.constants_hash=$2)`, manifest.EpochID, manifest.ConstantsHash).Scan(&current); err != nil || !current {
		return PostgresInspection{}, errors.Join(ErrInvalid, err)
	}
	for name, expected := range bundle.Artifacts {
		var actual []byte
		if err := database.QueryRowContext(ctx, `SELECT bytes FROM catalog_artifacts WHERE constants_hash=$1 AND artifact_name=$2`, bundle.Hash, name).Scan(&actual); err != nil || string(actual) != string(expected) {
			return PostgresInspection{}, errors.Join(ErrInvalid, err)
		}
		result.ArtifactsVerified++
	}
	var count int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, bundle.Hash).Scan(&count); err != nil || count != len(bundle.Artifacts) || result.ArtifactsVerified != count {
		return PostgresInspection{}, errors.Join(ErrInvalid, err)
	}
	return result, nil
}
