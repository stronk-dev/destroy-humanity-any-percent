package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RankingKind is the closed public board key union (API Foundation C12).
type RankingKind string

const (
	RankingTimeMS    RankingKind = "time_ms"
	RankingCount     RankingKind = "count"
	RankingMagnitude RankingKind = "magnitude"
)

// PublicBoardPageMaximum is the C12 list bound.
const PublicBoardPageMaximum = 100

var (
	ErrUnknownPublicEpoch    = errors.New("unknown public board epoch")
	ErrUnknownPublicCategory = errors.New("unknown public board category")
)

// rankingKindForTimer mirrors the projector's key selection: timed categories
// rank by milliseconds, the untimed (valuation) category by magnitude.
func rankingKindForTimer(timer CategoryTimer) (RankingKind, bool) {
	switch timer {
	case TimerRTA, TimerAttended:
		return RankingTimeMS, true
	case TimerNone:
		return RankingMagnitude, true
	}
	return "", false
}

// PublicBoardRankingKind resolves a category's ranking kind from the pinned
// categories artifacts of every stored constants hash accepted into the epoch (C13:
// "ranking kind resolves from the category's pinned catalog after load"). A
// category absent from every accepted catalog is unknown; accepted catalogs
// that disagree on its timer are an invariant failure, never a guess.
func (repository *Repository) PublicBoardRankingKind(ctx context.Context, categoryID string, epochID int64) (RankingKind, error) {
	if repository == nil || epochID < 1 {
		return "", ErrUnknownPublicEpoch
	}
	if !mechanicalPattern.MatchString(categoryID) {
		return "", ErrUnknownPublicCategory
	}
	var exists bool
	if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM epochs WHERE epoch_id=$1)`, epochID).Scan(&exists); err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownPublicEpoch
	}
	// Only accepted hashes whose catalog bytes are stored can own board rows:
	// the projector loads the run's pinned categories artifact and cannot
	// project a run whose catalog is absent, so skipping catalog-less hashes
	// hides no row. A stored hash missing its routes pair still fails loudly.
	rows, err := repository.db.QueryContext(ctx, `
		SELECT hashes.constants_hash FROM epoch_hashes hashes
		WHERE hashes.epoch_id=$1 AND EXISTS (
			SELECT 1 FROM catalog_artifacts artifacts
			WHERE artifacts.constants_hash=hashes.constants_hash AND artifacts.artifact_name='categories')
		ORDER BY hashes.constants_hash`, epochID)
	if err != nil {
		return "", err
	}
	hashes := []string{}
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			rows.Close()
			return "", err
		}
		hashes = append(hashes, hash)
	}
	if err := rows.Close(); err != nil {
		return "", err
	}
	catalogs := make([]*CategoryCatalog, 0, len(hashes))
	for _, hash := range hashes {
		catalog, err := loadPinnedCategoryCatalog(ctx, repository.db, hash)
		if err != nil {
			return "", err
		}
		catalogs = append(catalogs, catalog)
	}
	return resolveRankingKind(categoryID, epochID, catalogs)
}

// resolveRankingKind requires every catalog that declares the category to
// agree on its ranking kind; a category no catalog declares is unknown.
func resolveRankingKind(categoryID string, epochID int64, catalogs []*CategoryCatalog) (RankingKind, error) {
	kind := RankingKind("")
	for _, catalog := range catalogs {
		for _, category := range catalog.Categories {
			if category.ID != categoryID {
				continue
			}
			resolved, ok := rankingKindForTimer(category.Timer)
			if !ok || kind != "" && kind != resolved {
				return "", fmt.Errorf("%w: category %s ranking kind disagrees across epoch %d", ErrInvalidEpoch, categoryID, epochID)
			}
			kind = resolved
		}
	}
	if kind == "" {
		return "", ErrUnknownPublicCategory
	}
	return kind, nil
}

// PublicBoardQuery is one normalized board page request.
type PublicBoardQuery struct {
	CategoryID   string
	Variables    Variables
	EpochID      int64
	MandateLevel int
	Limit        int
	AfterKey     *int64
	AfterMag     *MagnitudeKey
	AfterRunID   string
}

// PublicBoardRow is a ranking-kind-neutral board row.
type PublicBoardRow struct {
	RunID      string
	FounderID  string
	Rank       int64
	Key        int64
	Magnitude  MagnitudeKey
	VerifiedAt time.Time
	WorldFirst bool
}

// PublicBoardPage reads limit+1 rows so the caller learns whether a next page
// exists without an empty trailing page.
func (repository *Repository) PublicBoardPage(ctx context.Context, kind RankingKind, query PublicBoardQuery) ([]PublicBoardRow, bool, error) {
	if repository == nil || !mechanicalPattern.MatchString(query.CategoryID) || !validVariables(query.Variables) || query.EpochID < 1 ||
		query.MandateLevel < 0 || query.MandateLevel > 20 || query.Limit < 1 || query.Limit > PublicBoardPageMaximum {
		return nil, false, ErrInvalidEpoch
	}
	result := []PublicBoardRow{}
	switch kind {
	case RankingTimeMS, RankingCount:
		var after *Cursor
		if query.AfterKey != nil {
			if query.AfterRunID == "" || query.AfterMag != nil {
				return nil, false, ErrInvalidEpoch
			}
			after = &Cursor{Key: *query.AfterKey, RunID: query.AfterRunID}
		}
		read := repository.timeBoard
		if kind == RankingCount {
			read = repository.countBoard
		}
		entries, err := read(ctx, query.CategoryID, query.Variables, query.EpochID, query.MandateLevel, query.Limit+1, after)
		if err != nil {
			return nil, false, err
		}
		for _, entry := range entries {
			result = append(result, PublicBoardRow{RunID: entry.RunID, FounderID: entry.FounderID, Rank: entry.Rank, Key: entry.Key,
				VerifiedAt: entry.VerifiedAt, WorldFirst: entry.WorldFirst})
		}
	case RankingMagnitude:
		var after *MagnitudeCursor
		if query.AfterMag != nil {
			if query.AfterRunID == "" || query.AfterKey != nil || !validMagnitudeKey(*query.AfterMag) {
				return nil, false, ErrInvalidEpoch
			}
			after = &MagnitudeCursor{Key: *query.AfterMag, RunID: query.AfterRunID}
		}
		entries, err := repository.magnitudeBoard(ctx, query.CategoryID, query.Variables, query.EpochID, query.MandateLevel, query.Limit+1, after)
		if err != nil {
			return nil, false, err
		}
		for _, entry := range entries {
			result = append(result, PublicBoardRow{RunID: entry.RunID, FounderID: entry.FounderID, Rank: entry.Rank, Magnitude: entry.Key,
				VerifiedAt: entry.VerifiedAt, WorldFirst: entry.WorldFirst})
		}
	default:
		return nil, false, ErrInvalidEpoch
	}
	if len(result) > query.Limit {
		return result[:query.Limit], true, nil
	}
	return result, false, nil
}
