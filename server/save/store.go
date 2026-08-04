package save

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
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
	StreamID       string
	OwnerID        string
	Number         int64
	Version        int
	ConstantsHash  string
	CreatedAt      time.Time
	RunLogSequence int64
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

// StatePolicyValidator lets catalog bundles enforce data-owned bounds without
// making the persistence package import feature packages.
type StatePolicyValidator interface {
	ValidateState(constantsHash string, state *State) error
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
	version := VersionForState(state)
	revision.Number, revision.Version, revision.ConstantsHash = 1, version, constantsHash
	err = tx.QueryRowContext(ctx, `
		INSERT INTO save_revisions (stream_id, revision, version, state, constants_hash)
		VALUES ($1, 1, $2, $3, $4) RETURNING created_at`,
		revision.StreamID, version, encodedState, constantsHash).Scan(&revision.CreatedAt)
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
	revision, err := s.WriteInTransaction(ctx, tx, streamID, expectedRevision, constantsHash, state, writeContext)
	if err != nil {
		return Revision{}, err
	}
	if err := tx.Commit(); err != nil {
		return Revision{}, err
	}
	return revision, nil
}

// WriteInTransaction applies the same validation, locking, conflict check, and
// retention policy as Write inside a caller-owned transaction. It exists for
// cross-owner operations, such as account import, whose relational marker must
// commit atomically with the save revision.
func (s *Store) WriteInTransaction(ctx context.Context, tx *sql.Tx, streamID string, expectedRevision int64, constantsHash string, state *State, writeContext WriteContext) (Revision, error) {
	if tx == nil || !uuidPattern.MatchString(streamID) || expectedRevision < 1 || !hashPattern.MatchString(constantsHash) || state == nil || state.Ledger == nil || writeContext.Cause == "" {
		return Revision{}, ErrInvalidStream
	}
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
	var latestVersion int
	if err := tx.QueryRowContext(ctx, `SELECT revision,version FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, streamID).Scan(&latest, &latestVersion); err != nil {
		return Revision{}, err
	}
	if latest != expectedRevision {
		return Revision{}, fmt.Errorf("%w: got %d, current %d", ErrConflict, expectedRevision, latest)
	}
	version := VersionForState(state)
	if version != latestVersion {
		return Revision{}, fmt.Errorf("%w: ordinary write cannot change save version %d to %d", ErrInvalidState, latestVersion, version)
	}
	revision := Revision{StreamID: streamID, Number: latest + 1, Version: version, ConstantsHash: constantsHash}
	err = tx.QueryRowContext(ctx, `INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash) VALUES ($1,$2,$3,$4,$5) RETURNING created_at`, streamID, revision.Number, version, encodedState, constantsHash).Scan(&revision.CreatedAt)
	if err != nil {
		return Revision{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, streamID, revision.Number-5); err != nil {
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

// LoadSiblingLatest resolves another active scope owned by the same Founder as
// the supplied stream. Callers use it for cross-scope read context only; any
// mutation that depends on the result must re-read under ApplyExitTransaction's
// Founder-then-Company locks.
func (s *Store) LoadSiblingLatest(ctx context.Context, streamID string, scope economy.Scope) (Loaded, error) {
	if !uuidPattern.MatchString(streamID) || scope != economy.ScopeCompany && scope != economy.ScopeFounder {
		return Loaded{}, ErrInvalidStream
	}
	var siblingID string
	err := s.db.QueryRowContext(ctx, `
		SELECT sibling.id
		FROM save_streams source
		JOIN save_streams sibling ON sibling.owner_kind=source.owner_kind AND sibling.owner_id=source.owner_id AND sibling.scope=$2 AND sibling.archived_at IS NULL
		WHERE source.id=$1 AND source.owner_kind='founder' AND source.archived_at IS NULL`, streamID, scope).Scan(&siblingID)
	if errors.Is(err, sql.ErrNoRows) {
		return Loaded{}, ErrNotFound
	}
	if err != nil {
		return Loaded{}, err
	}
	return s.LoadLatest(ctx, siblingID)
}

func (s *Store) CountEvents(ctx context.Context, streamID string, kind EventKind) (int64, error) {
	if !uuidPattern.MatchString(streamID) {
		return 0, ErrInvalidStream
	}
	var count int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind=$2`, streamID, kind).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CountRunEvents(ctx context.Context, streamID string, kind EventKind, runSeq int64) (int64, error) {
	if !uuidPattern.MatchString(streamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return 0, ErrInvalidStream
	}
	var count int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind=$2 AND payload->>'run_seq'=$3`,
		streamID, kind, fmt.Sprintf("%d", runSeq)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// ExecutedRouteIDs returns the distinct route facts committed for one run at
// or before the supplied Company revision. The revision boundary lets callers
// assemble a terminal run record before ApplyExitTransaction takes its locks:
// if another intent advances the stream meanwhile, the Exit compare-and-swap
// fails and the retry observes the newer fact set.
func (s *Store) ExecutedRouteIDs(ctx context.Context, streamID string, runSeq, throughRevision int64) ([]string, error) {
	if !uuidPattern.MatchString(streamID) || runSeq < 1 || runSeq > decimal.MaxExactInteger || throughRevision < 1 {
		return nil, ErrInvalidStream
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT payload
		FROM events
		WHERE stream_id=$1 AND revision <= $2 AND kind='route_executed'
		ORDER BY revision,event_id`, streamID, throughRevision)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seen := make(map[string]bool)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var payload struct {
			RouteID string `json:"route_id"`
			RunID   struct {
				CompanyStreamID string `json:"company_stream_id"`
				RunSeq          int64  `json:"run_seq"`
			} `json:"run_id"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil || payload.RunID.CompanyStreamID != streamID || !mechanicalIDPattern.MatchString(payload.RouteID) {
			return nil, ErrInvalidStream
		}
		if payload.RunID.RunSeq == runSeq {
			seen[payload.RouteID] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(seen))
	for routeID := range seen {
		result = append(result, routeID)
	}
	sort.Strings(result)
	return result, nil
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
	if validator, ok := s.catalogs.(StatePolicyValidator); ok {
		if err := validator.ValidateState(hash, state); err != nil {
			return nil, fmt.Errorf("%w: catalog state policy: %v", ErrInvalidState, err)
		}
	}
	version := VersionForState(state)
	encoded, err := EncodeStateVersion(state, version)
	if err != nil {
		return nil, err
	}
	if _, err := RestoreState(encoded, version, catalog, scope, time.Time{}); err != nil {
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
