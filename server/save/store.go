package save

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"cloud-clicker/server/economy"
)

var (
	ErrInvalidStream = errors.New("invalid save stream")
	ErrConflict      = errors.New("save revision conflict")
	ErrArchived      = errors.New("save stream is archived")
	ErrNotFound      = errors.New("save stream not found")
)

var (
	uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type OwnerKind string

const (
	OwnerFounder OwnerKind = "founder"
	OwnerGuild   OwnerKind = "guild"
	OwnerWorld   OwnerKind = "world"
)

type StreamKey struct {
	OwnerKind OwnerKind
	OwnerID   string
	Scope     economy.Scope
}

type WriteContext struct {
	Cause    string
	IntentID string
}

type Revision struct {
	StreamID      string
	OwnerID       string
	Number        int64
	Version       int
	ConstantsHash string
	CreatedAt     time.Time
}

type Loaded struct {
	Key        StreamKey
	Revision   Revision
	ArchivedAt *time.Time
	State      *State
}

type CatalogResolver interface {
	Resolve(constantsHash string) (*economy.Catalog, bool)
}

type Store struct {
	db       *sql.DB
	catalogs CatalogResolver
	logger   *slog.Logger
}

func NewStore(db *sql.DB, catalogs CatalogResolver, logger *slog.Logger) (*Store, error) {
	if db == nil || catalogs == nil {
		return nil, ErrInvalidStream
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Store{db: db, catalogs: catalogs, logger: logger}, nil
}

func (s *Store) CreateStream(ctx context.Context, key StreamKey, constantsHash string, state *State, writeContext WriteContext) (Revision, error) {
	if err := validateWrite(key, constantsHash, state, writeContext); err != nil {
		s.logRejection(err, writeContext)
		return Revision{}, err
	}
	if err := validateInitialCursors(key.Scope, state); err != nil {
		s.logRejection(err, writeContext)
		return Revision{}, err
	}
	encodedState, err := s.validatedState(constantsHash, key.Scope, state)
	if err != nil {
		s.logRejection(err, writeContext)
		return Revision{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Revision{}, err
	}
	defer tx.Rollback()
	var revision Revision
	err = tx.QueryRowContext(ctx, `
		INSERT INTO save_streams (owner_kind, owner_id, scope) VALUES ($1, $2, $3)
		RETURNING id`, key.OwnerKind, key.OwnerID, key.Scope).Scan(&revision.StreamID)
	if err != nil {
		return Revision{}, fmt.Errorf("create save stream: %w", err)
	}
	revision.Number, revision.Version, revision.ConstantsHash = 1, CurrentVersion, constantsHash
	err = tx.QueryRowContext(ctx, `
		INSERT INTO save_revisions (stream_id, revision, version, state, constants_hash)
		VALUES ($1, 1, $2, $3, $4) RETURNING created_at`,
		revision.StreamID, CurrentVersion, encodedState, constantsHash).Scan(&revision.CreatedAt)
	if err != nil {
		return Revision{}, fmt.Errorf("create save revision: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Revision{}, err
	}
	return revision, nil
}

func (s *Store) Write(ctx context.Context, streamID string, expectedRevision int64, constantsHash string, state *State, writeContext WriteContext) (Revision, error) {
	if !uuidPattern.MatchString(streamID) || expectedRevision < 1 || !hashPattern.MatchString(constantsHash) || state == nil || state.Ledger == nil || writeContext.Cause == "" {
		return Revision{}, ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Revision{}, err
	}
	defer tx.Rollback()
	var scope economy.Scope
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT scope, archived_at FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&scope, &archivedAt); errors.Is(err, sql.ErrNoRows) {
		return Revision{}, ErrNotFound
	} else if err != nil {
		return Revision{}, err
	}
	if archivedAt.Valid {
		return Revision{}, ErrArchived
	}
	if state.Ledger.Scope() != scope {
		return Revision{}, ErrInvalidStream
	}
	encodedState, err := s.validatedState(constantsHash, scope, state)
	if err != nil {
		s.logRejection(err, writeContext)
		return Revision{}, err
	}
	var latest int64
	if err := tx.QueryRowContext(ctx, `SELECT max(revision) FROM save_revisions WHERE stream_id=$1`, streamID).Scan(&latest); err != nil {
		return Revision{}, err
	}
	if latest != expectedRevision {
		return Revision{}, fmt.Errorf("%w: got %d, current %d", ErrConflict, expectedRevision, latest)
	}
	revision := Revision{StreamID: streamID, Number: latest + 1, Version: CurrentVersion, ConstantsHash: constantsHash}
	err = tx.QueryRowContext(ctx, `INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash) VALUES ($1,$2,$3,$4,$5) RETURNING created_at`, streamID, revision.Number, CurrentVersion, encodedState, constantsHash).Scan(&revision.CreatedAt)
	if err != nil {
		return Revision{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, streamID, revision.Number-5); err != nil {
		return Revision{}, err
	}
	if err := tx.Commit(); err != nil {
		return Revision{}, err
	}
	return revision, nil
}

func (s *Store) LoadLatest(ctx context.Context, streamID string) (Loaded, error) {
	var loaded Loaded
	var archived sql.NullTime
	var state []byte
	err := s.db.QueryRowContext(ctx, `SELECT s.owner_kind,s.owner_id,s.scope,s.archived_at,r.revision,r.version,r.constants_hash,r.created_at,r.state FROM save_streams s JOIN LATERAL (SELECT * FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true WHERE s.id=$1`, streamID).Scan(&loaded.Key.OwnerKind, &loaded.Key.OwnerID, &loaded.Key.Scope, &archived, &loaded.Revision.Number, &loaded.Revision.Version, &loaded.Revision.ConstantsHash, &loaded.Revision.CreatedAt, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return Loaded{}, ErrNotFound
	}
	if err != nil {
		return Loaded{}, err
	}
	loaded.Revision.StreamID = streamID
	if archived.Valid {
		loaded.ArchivedAt = &archived.Time
	}
	catalog, ok := s.catalogs.Resolve(loaded.Revision.ConstantsHash)
	if !ok {
		return Loaded{}, fmt.Errorf("%w: unknown catalog %s", ErrInvalidState, loaded.Revision.ConstantsHash)
	}
	loaded.State, err = RestoreState(state, loaded.Revision.Version, catalog, loaded.Key.Scope, loaded.Revision.CreatedAt)
	if err != nil {
		return Loaded{}, err
	}
	return loaded, nil
}

func (s *Store) Archive(ctx context.Context, streamID string, expectedRevision int64) error {
	if !uuidPattern.MatchString(streamID) || expectedRevision < 1 {
		return ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx,
		`SELECT archived_at FROM save_streams WHERE id=$1 FOR UPDATE`, streamID,
	).Scan(&archivedAt); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if archivedAt.Valid {
		return ErrArchived
	}
	var latest int64
	if err := tx.QueryRowContext(ctx,
		`SELECT max(revision) FROM save_revisions WHERE stream_id=$1`, streamID,
	).Scan(&latest); err != nil {
		return err
	}
	if latest != expectedRevision {
		return fmt.Errorf("%w: got %d, current %d", ErrConflict, expectedRevision, latest)
	}
	result, err := tx.ExecContext(ctx,
		`UPDATE save_streams SET archived_at=now() WHERE id=$1`, streamID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrNotFound
	}
	return tx.Commit()
}

func (s *Store) validatedState(hash string, scope economy.Scope, state *State) ([]byte, error) {
	catalog, ok := s.catalogs.Resolve(hash)
	if !ok {
		return nil, fmt.Errorf("%w: unknown catalog %s", ErrInvalidState, hash)
	}
	encoded, err := EncodeState(state)
	if err != nil {
		return nil, err
	}
	if _, err := RestoreState(encoded, CurrentVersion, catalog, scope, time.Time{}); err != nil {
		return nil, err
	}
	return encoded, nil
}

func validateWrite(key StreamKey, hash string, state *State, ctx WriteContext) error {
	validPair := key.OwnerKind == OwnerFounder && (key.Scope == economy.ScopeCompany || key.Scope == economy.ScopeFounder) || key.OwnerKind == OwnerGuild && key.Scope == economy.ScopeGuild || key.OwnerKind == OwnerWorld && key.Scope == economy.ScopeWorld
	if !validPair || !uuidPattern.MatchString(key.OwnerID) || !hashPattern.MatchString(hash) || state == nil || state.Ledger == nil || state.Ledger.Scope() != key.Scope || ctx.Cause == "" {
		return ErrInvalidStream
	}
	return nil
}

func validateInitialCursors(scope economy.Scope, state *State) error {
	if scope == economy.ScopeCompany && !state.EvaluatedThrough.Equal(state.ManualTokenRefilledAt) {
		return fmt.Errorf("%w: new company cursors must share one canonical server instant", ErrInvalidState)
	}
	return nil
}

func (s *Store) logRejection(err error, ctx WriteContext) {
	s.logger.Error("save rejected", "cause", ctx.Cause, "intent_id", ctx.IntentID, "error", err)
}
