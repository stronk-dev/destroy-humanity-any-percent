package minigame

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

const claimLease = 5 * time.Minute

var (
	ErrInvalidSession = errors.New("invalid minigame session")
	ErrSessionBusy    = errors.New("minigame session is busy")
	ErrSessionGone    = errors.New("minigame session is unavailable")
	ErrClaimLost      = errors.New("minigame session claim lost")

	uuidPattern       = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	uuidV7Pattern     = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	mechanicalPattern = regexp.MustCompile("^[a-z][a-z0-9_]*(?:\\.[a-z][a-z0-9_]*)*$")
	versionPattern    = regexp.MustCompile("^[0-9]+\\.[0-9]+\\.[0-9]+$")
	hashPattern       = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
)

type Mode string

const (
	ModeSolo          Mode = "solo"
	ModeAsyncSnapshot Mode = "async_snapshot"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusClaimed  Status = "claimed"
	StatusResolved Status = "resolved"
)

type CreateSession struct {
	SessionID       string
	MinigameID      string
	FounderID       string
	CompanyStreamID string
	RunSeq          int64
	EngineRef       string
	EngineVersion   string
	ConstantsHash   string
	ScalingInputs   json.RawMessage
	Seed            string
	Mode            Mode
	Genesis         json.RawMessage
}

type Session struct {
	SessionID       string
	MinigameID      string
	FounderID       string
	CompanyStreamID string
	RunSeq          int64
	EngineRef       string
	EngineVersion   string
	ConstantsHash   string
	ScalingInputs   json.RawMessage
	Seed            string
	Mode            Mode
	Status          Status
	Revision        int64
	Genesis         json.RawMessage
	State           json.RawMessage
	Result          json.RawMessage
	ClaimToken      string
	ClaimedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ResolvedAt      *time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("minigame repository requires database")
	}
	return &Repository{db: db}, nil
}

func (repository *Repository) Create(ctx context.Context, input CreateSession) (Session, error) {
	if repository == nil || !validCreate(input) {
		return Session{}, ErrInvalidSession
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()
	if err := lockFounder(ctx, tx, input.FounderID); err != nil {
		return Session{}, err
	}
	var ownsRun bool
	if err := tx.QueryRowContext(ctx, ownsRunSQL, input.CompanyStreamID, input.FounderID, input.RunSeq, input.ConstantsHash).Scan(&ownsRun); errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrInvalidSession
	} else if err != nil {
		return Session{}, err
	}
	row := tx.QueryRowContext(ctx, createSessionSQL, input.SessionID, input.MinigameID, input.FounderID,
		input.CompanyStreamID, input.RunSeq, input.EngineRef, input.EngineVersion, input.ConstantsHash,
		string(input.ScalingInputs), input.Seed, input.Mode, string(input.Genesis))
	created, err := scanSession(row)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return created, nil
}

func (repository *Repository) Load(ctx context.Context, founderID, sessionID string) (Session, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) {
		return Session{}, ErrInvalidSession
	}
	result, err := scanSession(repository.db.QueryRowContext(ctx, loadSessionSQL, sessionID, founderID))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrSessionGone
	}
	return result, err
}

// Claim serializes a server-side play or resolution command. The founder row
// is locked before the session row; this is the C13 account/session lock order.
func (repository *Repository) Claim(ctx context.Context, founderID, sessionID string) (Session, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) {
		return Session{}, ErrInvalidSession
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()
	if err := lockFounder(ctx, tx, founderID); err != nil {
		return Session{}, err
	}
	row := tx.QueryRowContext(ctx, claimSessionSQL, sessionID, founderID, intervalLiteral(claimLease))
	claimed, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		var status Status
		lookupErr := tx.QueryRowContext(ctx, statusSQL, sessionID, founderID).Scan(&status)
		if errors.Is(lookupErr, sql.ErrNoRows) || status == StatusResolved {
			return Session{}, ErrSessionGone
		}
		if lookupErr != nil {
			return Session{}, lookupErr
		}
		return Session{}, ErrSessionBusy
	}
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return claimed, nil
}

func (repository *Repository) CompletePlay(ctx context.Context, founderID, sessionID, claimToken string, state json.RawMessage) (Session, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) ||
		!uuidPattern.MatchString(claimToken) || !validJSONObject(state) {
		return Session{}, ErrInvalidSession
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback()
	if err := lockFounder(ctx, tx, founderID); err != nil {
		return Session{}, err
	}
	row := tx.QueryRowContext(ctx, completePlaySQL, sessionID, founderID, claimToken, string(state))
	updated, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrClaimLost
	}
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return updated, nil
}

// ResolveTx is the session half of C17's company-then-session transaction. The
// caller owns and has already locked the Company stream.
func ResolveTx(ctx context.Context, tx *sql.Tx, sessionID, claimToken string, result json.RawMessage) (Session, error) {
	if tx == nil || !uuidV7Pattern.MatchString(sessionID) || !uuidPattern.MatchString(claimToken) || !validJSONObject(result) {
		return Session{}, ErrInvalidSession
	}
	resolved, err := scanSession(tx.QueryRowContext(ctx, resolveSessionSQL, sessionID, claimToken, string(result)))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrClaimLost
	}
	return resolved, err
}

func validCreate(input CreateSession) bool {
	if !uuidV7Pattern.MatchString(input.SessionID) || !mechanicalPattern.MatchString(input.MinigameID) ||
		!uuidPattern.MatchString(input.FounderID) || !uuidPattern.MatchString(input.CompanyStreamID) ||
		input.RunSeq < 1 || input.RunSeq > 9_007_199_254_740_991 || !mechanicalPattern.MatchString(input.EngineRef) ||
		!versionPattern.MatchString(input.EngineVersion) || !hashPattern.MatchString(input.ConstantsHash) ||
		!validJSONObject(input.ScalingInputs) || !validJSONObject(input.Genesis) ||
		(input.Mode != ModeSolo && input.Mode != ModeAsyncSnapshot) {
		return false
	}
	_, err := strconv.ParseUint(input.Seed, 10, 64)
	return err == nil && (input.Seed == "0" || input.Seed[0] != '0')
}

func validJSONObject(value []byte) bool {
	canonical, ok := canonicalJSONObject(value)
	return ok && bytes.Equal(canonical, value)
}

func canonicalJSONObject(value []byte) (json.RawMessage, bool) {
	if len(value) == 0 || !utf8.Valid(value) || !json.Valid(value) {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, false
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return nil, false
	}
	encoded, err := json.Marshal(object)
	return encoded, err == nil
}

func lockFounder(ctx context.Context, tx *sql.Tx, founderID string) error {
	var active bool
	err := tx.QueryRowContext(ctx, lockFounderSQL, founderID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSessionGone
	}
	return err
}

const sessionColumns = "session_id,minigame_id,founder_id,company_stream_id,run_seq,engine_ref," +
	"engine_version,constants_hash,scaling_inputs::text,seed,mode,status,revision,genesis::text,state::text," +
	"COALESCE(result::text,''),COALESCE(claim_token::text,''),claimed_at,created_at,updated_at,resolved_at"

const ownsRunSQL = "SELECT true FROM save_streams s JOIN run_epochs r ON r.company_stream_id=s.id AND r.run_seq=$3 " +
	"WHERE s.id=$1 AND s.owner_kind='founder' AND s.owner_id=$2 AND s.scope='company' AND s.archived_at IS NULL " +
	"AND r.constants_hash=$4"
const createSessionSQL = "INSERT INTO minigame_sessions(session_id,minigame_id,founder_id,company_stream_id,run_seq," +
	"engine_ref,engine_version,constants_hash,scaling_inputs,seed,mode,genesis,state) " +
	"VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12) RETURNING " + sessionColumns
const loadSessionSQL = "SELECT " + sessionColumns + " FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2"
const claimSessionSQL = "UPDATE minigame_sessions SET status='claimed',claim_token=gen_random_uuid()," +
	"claimed_at=clock_timestamp(),updated_at=clock_timestamp() WHERE session_id=$1 AND founder_id=$2 AND " +
	"(status='active' OR (status='claimed' AND claimed_at<clock_timestamp()-$3::interval)) RETURNING " + sessionColumns
const statusSQL = "SELECT status FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2"
const completePlaySQL = "UPDATE minigame_sessions SET state=$4,status='active',revision=revision+1,claim_token=NULL," +
	"claimed_at=NULL,updated_at=clock_timestamp() WHERE session_id=$1 AND founder_id=$2 AND status='claimed' " +
	"AND claim_token=$3 RETURNING " + sessionColumns
const resolveSessionSQL = "UPDATE minigame_sessions SET result=$3,status='resolved',revision=revision+1," +
	"claim_token=NULL,claimed_at=NULL,updated_at=clock_timestamp(),resolved_at=clock_timestamp() " +
	"WHERE session_id=$1 AND status='claimed' AND claim_token=$2 RETURNING " + sessionColumns
const lockFounderSQL = "SELECT true FROM account_founders WHERE founder_id=$1 AND archived_at IS NULL FOR UPDATE"

type rowScanner interface {
	Scan(...any) error
}

func scanSession(row rowScanner) (Session, error) {
	var result Session
	var scaling, genesis, state, resolved string
	var claimedAt, resolvedAt sql.NullTime
	err := row.Scan(&result.SessionID, &result.MinigameID, &result.FounderID, &result.CompanyStreamID,
		&result.RunSeq, &result.EngineRef, &result.EngineVersion, &result.ConstantsHash, &scaling,
		&result.Seed, &result.Mode, &result.Status, &result.Revision, &genesis, &state, &resolved,
		&result.ClaimToken, &claimedAt, &result.CreatedAt, &result.UpdatedAt, &resolvedAt)
	if err != nil {
		return Session{}, err
	}
	result.ScalingInputs, _ = canonicalJSONObject([]byte(scaling))
	result.Genesis, _ = canonicalJSONObject([]byte(genesis))
	result.State, _ = canonicalJSONObject([]byte(state))
	if resolved != "" {
		result.Result, _ = canonicalJSONObject([]byte(resolved))
	}
	if claimedAt.Valid {
		value := claimedAt.Time
		result.ClaimedAt = &value
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time
		result.ResolvedAt = &value
	}
	return result, nil
}

func intervalLiteral(value time.Duration) string {
	return strconv.FormatInt(value.Milliseconds(), 10) + " milliseconds"
}
