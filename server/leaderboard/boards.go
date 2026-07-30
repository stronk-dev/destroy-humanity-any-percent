package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	uuidPattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	mechanicalPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type Variables struct {
	Commons  bool `json:"commons"`
	Advisor  bool `json:"advisor"`
	Glitched bool `json:"glitched"`
}

type VerifiedRun struct {
	EventID      string
	RunID        string
	FounderID    string
	CategoryID   string
	Variables    Variables
	EpochID      int64
	MandateLevel int
	KeyMS        *int64
	KeyInt       *int64
	VerifiedAt   time.Time
}

type BoardEntry struct {
	RunID      string    `json:"run_id"`
	FounderID  string    `json:"founder_id"`
	Rank       int64     `json:"rank"`
	Key        int64     `json:"key"`
	VerifiedAt time.Time `json:"verified_at"`
	WorldFirst bool      `json:"world_first"`
}

type Cursor struct {
	Key   int64
	RunID string
}

func (repository *Repository) ProjectVerifiedRun(ctx context.Context, run VerifiedRun) (bool, error) {
	if !uuidPattern.MatchString(run.EventID) || !uuidPattern.MatchString(run.FounderID) || run.RunID == "" || !mechanicalPattern.MatchString(run.CategoryID) ||
		run.EpochID < 1 || run.MandateLevel < 0 || run.MandateLevel > 20 || run.VerifiedAt.IsZero() || (run.KeyMS == nil) == (run.KeyInt == nil) {
		return false, ErrInvalidEpoch
	}
	variables, _ := json.Marshal(run.Variables)
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO verification_projection_events(event_id) VALUES($1) ON CONFLICT DO NOTHING`, run.EventID)
	if err != nil {
		return false, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}
	var imported bool
	if err := tx.QueryRowContext(ctx, `SELECT imported FROM account_founders WHERE founder_id=$1`, run.FounderID).Scan(&imported); errors.Is(err, sql.ErrNoRows) {
		return false, ErrInvalidEpoch
	} else if err != nil {
		return false, err
	}
	if imported {
		return false, ErrInvalidEpoch
	}
	var insertedRunID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,key_int,verified_at,world_first)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true)
		ON CONFLICT (category_id,variables,epoch_id) WHERE world_first DO NOTHING
		RETURNING run_id`, run.RunID, run.EventID, run.FounderID, run.CategoryID, variables, run.EpochID, run.MandateLevel, run.KeyMS, run.KeyInt, run.VerifiedAt.UTC()).Scan(&insertedRunID)
	worldFirst := err == nil
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,key_int,verified_at,world_first)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,false)`, run.RunID, run.EventID, run.FounderID, run.CategoryID, variables, run.EpochID, run.MandateLevel, run.KeyMS, run.KeyInt, run.VerifiedAt.UTC())
	}
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return worldFirst, nil
}

func (repository *Repository) TimeBoard(ctx context.Context, categoryID string, variables Variables, epochID int64, mandateLevel, limit int, after *Cursor) ([]BoardEntry, error) {
	if !mechanicalPattern.MatchString(categoryID) || epochID < 1 || mandateLevel < 0 || mandateLevel > 20 || limit < 1 || limit > 100 || after != nil && after.RunID == "" {
		return nil, ErrInvalidEpoch
	}
	encoded, _ := json.Marshal(variables)
	afterKey, afterRun, hasAfter := int64(0), "", false
	if after != nil {
		afterKey, afterRun, hasAfter = after.Key, after.RunID, true
	}
	rows, err := repository.db.QueryContext(ctx, `
		WITH ranked AS (
			SELECT run_id,founder_id,key_ms,verified_at,world_first,
			       rank() OVER (ORDER BY key_ms ASC) AS competition_rank
			FROM verified_runs
			WHERE category_id=$1 AND variables=$2 AND epoch_id=$3 AND mandate_level=$4 AND key_ms IS NOT NULL
		)
		SELECT run_id,founder_id,competition_rank,key_ms,verified_at,world_first
		FROM ranked
		WHERE NOT $5 OR (key_ms,run_id) > ($6,$7)
		ORDER BY key_ms,run_id LIMIT $8`, categoryID, encoded, epochID, mandateLevel, hasAfter, afterKey, afterRun, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []BoardEntry
	for rows.Next() {
		var entry BoardEntry
		if err := rows.Scan(&entry.RunID, &entry.FounderID, &entry.Rank, &entry.Key, &entry.VerifiedAt, &entry.WorldFirst); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("time board: %w", err)
	}
	return entries, nil
}

func (repository *Repository) CountBoard(ctx context.Context, categoryID string, variables Variables, epochID int64, mandateLevel, limit int, after *Cursor) ([]BoardEntry, error) {
	if !mechanicalPattern.MatchString(categoryID) || epochID < 1 || mandateLevel < 0 || mandateLevel > 20 || limit < 1 || limit > 100 || after != nil && after.RunID == "" {
		return nil, ErrInvalidEpoch
	}
	encoded, _ := json.Marshal(variables)
	afterKey, afterRun, hasAfter := int64(0), "", false
	if after != nil {
		afterKey, afterRun, hasAfter = after.Key, after.RunID, true
	}
	rows, err := repository.db.QueryContext(ctx, `
		WITH ranked AS (
			SELECT run_id,founder_id,key_int,verified_at,world_first,
			       rank() OVER (ORDER BY key_int DESC) AS competition_rank
			FROM verified_runs
			WHERE category_id=$1 AND variables=$2 AND epoch_id=$3 AND mandate_level=$4 AND key_int IS NOT NULL
		)
		SELECT run_id,founder_id,competition_rank,key_int,verified_at,world_first
		FROM ranked
		WHERE NOT $5 OR key_int < $6 OR (key_int=$6 AND run_id>$7)
		ORDER BY key_int DESC,run_id LIMIT $8`, categoryID, encoded, epochID, mandateLevel, hasAfter, afterKey, afterRun, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []BoardEntry
	for rows.Next() {
		var entry BoardEntry
		if err := rows.Scan(&entry.RunID, &entry.FounderID, &entry.Rank, &entry.Key, &entry.VerifiedAt, &entry.WorldFirst); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
