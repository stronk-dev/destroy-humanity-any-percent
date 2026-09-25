package leaderboard

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// PublicEpochPageMaximum is the ruled public list bound (API Foundation C12).
const PublicEpochPageMaximum = 100

// PublicEpoch is one public epoch-registry row: the epoch, its byte-sorted
// accepted constants hashes, and the exact changelog bytes staged at its
// reference.
type PublicEpoch struct {
	ID           int64
	Name         string
	StartedAt    time.Time
	EndedAt      *time.Time
	ChangelogRef string
	Changelog    []byte
	Hashes       []string
}

// PublicEpochPage returns up to limit epochs newest-first whose ID is below
// before (0 starts at the newest), and whether more rows follow. It reads one
// statement snapshot. A missing, empty, or non-UTF-8 changelog fails closed.
func (repository *Repository) PublicEpochPage(ctx context.Context, before int64, limit int) ([]PublicEpoch, bool, error) {
	if repository == nil || before < 0 || limit < 1 || limit > PublicEpochPageMaximum {
		return nil, false, ErrInvalidEpoch
	}
	var bound any
	if before > 0 {
		bound = before
	}
	rows, err := repository.db.QueryContext(ctx, `
SELECT e.epoch_id, e.name, e.started_at, e.ended_at, e.changelog_ref,
       COALESCE(string_agg(h.constants_hash, ',' ORDER BY h.constants_hash COLLATE "C"), '')
FROM epochs e
LEFT JOIN epoch_hashes h ON h.epoch_id = e.epoch_id
WHERE $1::bigint IS NULL OR e.epoch_id < $1::bigint
GROUP BY e.epoch_id
ORDER BY e.epoch_id DESC
LIMIT $2`, bound, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	page := make([]PublicEpoch, 0, limit)
	more := false
	for rows.Next() {
		if len(page) == limit {
			more = true
			break
		}
		var epoch PublicEpoch
		var ended *time.Time
		var hashes string
		if err := rows.Scan(&epoch.ID, &epoch.Name, &epoch.StartedAt, &ended, &epoch.ChangelogRef, &hashes); err != nil {
			return nil, false, err
		}
		epoch.StartedAt = epoch.StartedAt.UTC()
		if ended != nil {
			value := ended.UTC()
			epoch.EndedAt = &value
		}
		epoch.Hashes = publicEpochHashes(hashes)
		page = append(page, epoch)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	for index := range page {
		changelog, err := repository.publicChangelog(page[index].ChangelogRef)
		if err != nil {
			return nil, false, err
		}
		page[index].Changelog = changelog
	}
	return page, more, nil
}

func (repository *Repository) publicChangelog(reference string) ([]byte, error) {
	if err := ValidateChangelog(repository.repositoryRoot, reference); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(repository.repositoryRoot, filepath.FromSlash(reference)))
	if err != nil || !utf8.Valid(data) {
		return nil, ErrInvalidEpoch
	}
	return data, nil
}

// publicEpochHashes splits the aggregated hash list and byte-sorts it in Go,
// so the public order never depends on the database's aggregation order or
// collation.
func publicEpochHashes(aggregated string) []string {
	if aggregated == "" {
		return []string{}
	}
	hashes := strings.Split(aggregated, ",")
	sort.Strings(hashes)
	return hashes
}
