package deploymentbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	stdhash "hash"
	"regexp"
	"strings"
)

var postgresIdentifier = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

type RecoveryIdentityInput struct {
	DatabaseURLFile string
}

type RecoveryDomainIdentity struct {
	Rows   int64  `json:"rows"`
	SHA256 string `json:"sha256"`
}

// RecoveryIdentity is a content identity, not a data export. Only row counts
// and hashes leave the private backup service.
type RecoveryIdentity struct {
	SchemaVersion     int                    `json:"schema_version"`
	DatabaseMigration int                    `json:"database_migration"`
	VerifiedRunRows   int64                  `json:"verified_run_rows"`
	Database          RecoveryDomainIdentity `json:"database"`
	Player            RecoveryDomainIdentity `json:"player"`
	Founder           RecoveryDomainIdentity `json:"founder"`
	Company           RecoveryDomainIdentity `json:"company"`
	Events            RecoveryDomainIdentity `json:"events"`
	Board             RecoveryDomainIdentity `json:"board"`
	Epoch             RecoveryDomainIdentity `json:"epoch"`
}

type identityQuery struct {
	name string
	sql  string
}

func InspectRecoveryIdentity(ctx context.Context, input RecoveryIdentityInput) (RecoveryIdentity, error) {
	if input.DatabaseURLFile == "" {
		return RecoveryIdentity{}, ErrInvalid
	}
	databaseURL, err := readSecretFile(input.DatabaseURLFile)
	if err != nil {
		return RecoveryIdentity{}, err
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return RecoveryIdentity{}, err
	}
	defer database.Close()
	if err := database.PingContext(ctx); err != nil {
		return RecoveryIdentity{}, err
	}

	var migration int
	if err := database.QueryRowContext(ctx, `SELECT COALESCE(max(version_id) FILTER (WHERE is_applied),0) FROM goose_db_version`).Scan(&migration); err != nil {
		return RecoveryIdentity{}, err
	}
	allTables, err := discoverIdentityTables(ctx, database, `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		return RecoveryIdentity{}, err
	}
	eventTables, err := discoverIdentityTables(ctx, database, `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public' AND (tablename='events' OR tablename='intent_records' OR tablename LIKE '%\_events' ESCAPE '\') ORDER BY tablename`)
	if err != nil {
		return RecoveryIdentity{}, err
	}

	identity := RecoveryIdentity{SchemaVersion: 2, DatabaseMigration: migration}
	if identity.Database, err = hashIdentityQueries(ctx, database, tableQueries(allTables)); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Player, err = hashIdentityQueries(ctx, database, tableQueries([]string{"accounts", "account_emails", "account_founders"})); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Founder, err = hashIdentityQueries(ctx, database, streamIdentityQueries("founder")); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Company, err = hashIdentityQueries(ctx, database, streamIdentityQueries("company")); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Events, err = hashIdentityQueries(ctx, database, tableQueries(eventTables)); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Board, err = hashIdentityQueries(ctx, database, tableQueries([]string{"verification_projection_events", "verified_runs"})); err != nil {
		return RecoveryIdentity{}, err
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM public.verified_runs`).Scan(&identity.VerifiedRunRows); err != nil {
		return RecoveryIdentity{}, err
	}
	if identity.Epoch, err = hashIdentityQueries(ctx, database, tableQueries([]string{"catalog_sets", "catalog_artifacts", "epochs", "epoch_hashes", "run_epochs"})); err != nil {
		return RecoveryIdentity{}, err
	}
	if ValidateRecoveryIdentity(identity) != nil {
		return RecoveryIdentity{}, ErrInvalid
	}
	return identity, nil
}

func ValidateRecoveryIdentity(identity RecoveryIdentity) error {
	if identity.SchemaVersion != 2 || identity.DatabaseMigration < 1 || identity.Database.Rows < 1 ||
		identity.VerifiedRunRows < 0 || identity.VerifiedRunRows > identity.Board.Rows {
		return ErrInvalid
	}
	for _, domain := range []RecoveryDomainIdentity{identity.Database, identity.Player, identity.Founder, identity.Company, identity.Events, identity.Board, identity.Epoch} {
		if domain.Rows < 0 || !hash.MatchString(domain.SHA256) {
			return ErrInvalid
		}
	}
	return nil
}

func ValidateEmptyRecoveryIdentity(identity RecoveryIdentity) error {
	if ValidateRecoveryIdentity(identity) != nil || identity.Epoch.Rows < 1 || identity.Player.Rows != 0 || identity.Founder.Rows != 0 ||
		identity.Company.Rows != 0 || identity.Events.Rows != 0 || identity.Board.Rows != 0 || identity.VerifiedRunRows != 0 {
		return ErrInvalid
	}
	return nil
}

func ValidatePopulatedRecoveryIdentity(identity RecoveryIdentity) error {
	if ValidateRecoveryIdentity(identity) != nil || identity.VerifiedRunRows < 1 {
		return ErrInvalid
	}
	for _, domain := range []RecoveryDomainIdentity{identity.Player, identity.Founder, identity.Company, identity.Events, identity.Board, identity.Epoch} {
		if domain.Rows < 1 {
			return ErrInvalid
		}
	}
	return nil
}

func discoverIdentityTables(ctx context.Context, database *sql.DB, query string) ([]string, error) {
	rows, err := database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if rows.Scan(&name) != nil || !postgresIdentifier.MatchString(name) {
			return nil, ErrInvalid
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil || len(names) == 0 {
		return nil, errors.Join(ErrInvalid, err)
	}
	return names, nil
}

func tableQueries(names []string) []identityQuery {
	queries := make([]identityQuery, 0, len(names))
	for _, name := range names {
		quoted := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
		queries = append(queries, identityQuery{name: name,
			sql: `SELECT to_jsonb(t)::text FROM public.` + quoted + ` AS t ORDER BY to_jsonb(t)::text`})
	}
	return queries
}

func streamIdentityQueries(scope string) []identityQuery {
	quotedScope := `'` + strings.ReplaceAll(scope, `'`, `''`) + `'`
	return []identityQuery{
		{name: "save_streams:" + scope, sql: `SELECT to_jsonb(t)::text FROM public.save_streams AS t WHERE scope=` + quotedScope + ` ORDER BY to_jsonb(t)::text`},
		{name: "save_revisions:" + scope, sql: `SELECT to_jsonb(r)::text FROM public.save_revisions AS r JOIN public.save_streams AS s ON s.id=r.stream_id WHERE s.scope=` + quotedScope + ` ORDER BY to_jsonb(r)::text`},
	}
}

func hashIdentityQueries(ctx context.Context, database *sql.DB, queries []identityQuery) (RecoveryDomainIdentity, error) {
	if database == nil || len(queries) == 0 {
		return RecoveryDomainIdentity{}, ErrInvalid
	}
	digest := sha256.New()
	var count int64
	for _, query := range queries {
		if query.name == "" || query.sql == "" {
			return RecoveryDomainIdentity{}, ErrInvalid
		}
		writeIdentityValue(digest, []byte(query.name))
		rows, err := database.QueryContext(ctx, query.sql)
		if err != nil {
			return RecoveryDomainIdentity{}, err
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				rows.Close()
				return RecoveryDomainIdentity{}, err
			}
			writeIdentityValue(digest, []byte(value))
			count++
		}
		err = rows.Err()
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return RecoveryDomainIdentity{}, err
		}
	}
	return RecoveryDomainIdentity{Rows: count, SHA256: "sha256:" + hex.EncodeToString(digest.Sum(nil))}, nil
}

func writeIdentityValue(digest stdhash.Hash, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write(value)
}

func CompareRecoveryIdentity(before, after RecoveryIdentity) error {
	if ValidateRecoveryIdentity(before) != nil || ValidateRecoveryIdentity(after) != nil || before != after {
		return fmt.Errorf("%w: recovery identity mismatch", ErrInvalid)
	}
	return nil
}
