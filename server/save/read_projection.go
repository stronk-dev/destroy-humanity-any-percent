package save

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cloud-clicker/server/economy"
)

// LoadSiblingLatestAtDatabaseTime reads an active same-Founder sibling head and
// the database clock in one statement. It is advisory read context only: mutations
// must re-read under their transaction's locks, not reuse this snapshot or clock.
func (s *Store) LoadSiblingLatestAtDatabaseTime(ctx context.Context, streamID string, scope economy.Scope) (Loaded, int64, error) {
	if s == nil || s.db == nil || !uuidPattern.MatchString(streamID) || scope != economy.ScopeCompany && scope != economy.ScopeFounder {
		return Loaded{}, 0, ErrInvalidStream
	}
	var loaded Loaded
	var state []byte
	var serverMS int64
	err := s.db.QueryRowContext(ctx, `
		SELECT sibling.id,sibling.owner_kind,sibling.owner_id,sibling.scope,
		       r.revision,r.version,r.constants_hash,r.created_at,r.state,
		       (extract(epoch FROM clock_timestamp())*1000)::bigint
		FROM save_streams source
		JOIN save_streams sibling ON sibling.owner_kind=source.owner_kind
		     AND sibling.owner_id=source.owner_id AND sibling.scope=$2 AND sibling.archived_at IS NULL
		JOIN LATERAL (SELECT * FROM save_revisions WHERE stream_id=sibling.id ORDER BY revision DESC LIMIT 1) r ON true
		WHERE source.id=$1 AND source.owner_kind='founder' AND source.archived_at IS NULL`, streamID, scope).Scan(
		&loaded.Revision.StreamID, &loaded.Key.OwnerKind, &loaded.Key.OwnerID, &loaded.Key.Scope,
		&loaded.Revision.Number, &loaded.Revision.Version, &loaded.Revision.ConstantsHash,
		&loaded.Revision.CreatedAt, &state, &serverMS)
	if errors.Is(err, sql.ErrNoRows) {
		return Loaded{}, 0, ErrNotFound
	}
	if err != nil {
		return Loaded{}, 0, err
	}
	catalog, ok := s.catalogs.Resolve(loaded.Revision.ConstantsHash)
	if !ok {
		return Loaded{}, 0, fmt.Errorf("%w: unknown catalog %s", ErrInvalidState, loaded.Revision.ConstantsHash)
	}
	loaded.State, err = RestoreState(state, loaded.Revision.Version, catalog, loaded.Key.Scope, loaded.Revision.CreatedAt)
	if err != nil {
		return Loaded{}, 0, err
	}
	return loaded, serverMS, nil
}
