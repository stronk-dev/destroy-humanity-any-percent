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
	"strings"
	"time"
	"unicode/utf8"
)

const claimLease = 5 * time.Minute

var (
	ErrInvalidSession    = errors.New("invalid minigame session")
	ErrSessionBusy       = errors.New("minigame session is busy")
	ErrSessionGone       = errors.New("minigame session is unavailable")
	ErrClaimLost         = errors.New("minigame session claim lost")
	ErrExclusiveActivity = errors.New("another exclusive activity is active")
	ErrAPIIdempotency    = errors.New("minigame API idempotency conflict")

	uuidPattern       = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	uuidV7Pattern     = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	mechanicalPattern = regexp.MustCompile("^[a-z][a-z0-9_]*(?:\\.[a-z][a-z0-9_]*)*$")
	versionPattern    = regexp.MustCompile("^[0-9]+\\.[0-9]+\\.[0-9]+$")
	hashPattern       = regexp.MustCompile("^sha256:[0-9a-f]{64}$")
	opaqueIDPattern   = regexp.MustCompile("^[A-Za-z0-9-]{1,64}$")
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
	SessionID                 string
	MinigameID                string
	FounderID                 string
	CompanyStreamID           string
	RunSeq                    int64
	EngineRef                 string
	EngineVersion             string
	ConstantsHash             string
	ScalingInputs             json.RawMessage
	Seed                      string
	Mode                      Mode
	Status                    Status
	Revision                  int64
	Genesis                   json.RawMessage
	State                     json.RawMessage
	Result                    json.RawMessage
	ClaimToken                string
	ClaimedAt                 *time.Time
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	ResolvedAt                *time.Time
	ResolutionReceipt         json.RawMessage
	ResolutionCompanyRevision *int64
	ResolutionFounderRevision *int64
}

type sessionCommand struct {
	Sequence        int64
	Command         json.RawMessage
	AppliedRevision int64
	ServerTSMS      int64
}

type Repository struct {
	db *sql.DB
}

type APIReceipt struct {
	SessionID string
	Response  json.RawMessage
}

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("minigame repository requires database")
	}
	return &Repository{db: db}, nil
}

func (repository *Repository) create(ctx context.Context, input CreateSession) (Session, error) {
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
	created, err := createTx(ctx, tx, input)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return Session{}, err
	}
	return created, nil
}

// CreateTx inserts a prepared deterministic session inside a coordinator-
// owned Founder→Company transaction. The caller must already hold the Founder
// serialization lock; the database still rechecks recovery exclusivity and
// pinned-run ownership before the insert.
func CreateTx(ctx context.Context, tx *sql.Tx, input CreateSession) (Session, error) {
	if tx == nil || !validCreate(input) {
		return Session{}, ErrInvalidSession
	}
	return createTx(ctx, tx, input)
}

func createTx(ctx context.Context, tx *sql.Tx, input CreateSession) (Session, error) {
	var exclusive bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM soul_recovery_sessions WHERE founder_id=$1 AND status IN ('active','claimed'))`,
		input.FounderID).Scan(&exclusive); err != nil {
		return Session{}, err
	}
	if exclusive {
		return Session{}, ErrExclusiveActivity
	}
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM minigame_sessions WHERE founder_id=$1 AND status IN ('active','claimed'))`,
		input.FounderID).Scan(&exclusive); err != nil {
		return Session{}, err
	}
	if exclusive {
		return Session{}, ErrExclusiveActivity
	}
	var ownsRun bool
	if err := tx.QueryRowContext(ctx, ownsRunSQL, input.CompanyStreamID, input.FounderID, input.RunSeq, input.ConstantsHash).Scan(&ownsRun); errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrInvalidSession
	} else if err != nil {
		return Session{}, err
	}
	return scanSession(tx.QueryRowContext(ctx, createSessionSQL, input.SessionID, input.MinigameID, input.FounderID,
		input.CompanyStreamID, input.RunSeq, input.EngineRef, input.EngineVersion, input.ConstantsHash,
		string(input.ScalingInputs), input.Seed, input.Mode, string(input.Genesis)))
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

// Current returns the sole active or claimed session for a Founder. Migration
// 00071 makes cardinality a database invariant rather than an API assumption.
func (repository *Repository) Current(ctx context.Context, founderID string) (Session, bool, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) {
		return Session{}, false, ErrInvalidSession
	}
	result, err := scanSession(repository.db.QueryRowContext(ctx, currentSessionSQL, founderID))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, false, nil
	}
	return result, err == nil, err
}

// CreateReceipt and CommandReceipt are read-before-execute idempotency gates.
// A hash mismatch is deterministic evidence; a missing row is not an error.
func (repository *Repository) CreateReceipt(ctx context.Context, founderID, idempotencyKey, requestHash string) (APIReceipt, bool, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !opaqueIDPattern.MatchString(idempotencyKey) || !hashPattern.MatchString(requestHash) {
		return APIReceipt{}, false, ErrInvalidSession
	}
	var result APIReceipt
	var storedHash string
	var response []byte
	err := repository.db.QueryRowContext(ctx, loadCreateReceiptSQL, founderID, idempotencyKey).Scan(&storedHash, &result.SessionID, &response)
	if errors.Is(err, sql.ErrNoRows) {
		return APIReceipt{}, false, nil
	}
	if err != nil {
		return APIReceipt{}, false, err
	}
	if storedHash != requestHash {
		return APIReceipt{}, false, ErrAPIIdempotency
	}
	result.Response, _ = canonicalJSONObject(response)
	if !validJSONObject(result.Response) {
		return APIReceipt{}, false, ErrInvalidSession
	}
	return result, true, nil
}

func (repository *Repository) CommandReceipt(ctx context.Context, founderID, sessionID, commandID, requestHash string) (APIReceipt, bool, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) ||
		!opaqueIDPattern.MatchString(commandID) || !hashPattern.MatchString(requestHash) {
		return APIReceipt{}, false, ErrInvalidSession
	}
	var storedHash string
	var response []byte
	err := repository.db.QueryRowContext(ctx, loadCommandReceiptSQL, sessionID, founderID, commandID).Scan(&storedHash, &response)
	if errors.Is(err, sql.ErrNoRows) {
		return APIReceipt{}, false, nil
	}
	if err != nil {
		return APIReceipt{}, false, err
	}
	if storedHash != requestHash {
		return APIReceipt{}, false, ErrAPIIdempotency
	}
	canonical, ok := canonicalJSONObject(response)
	if !ok {
		return APIReceipt{}, false, ErrInvalidSession
	}
	return APIReceipt{SessionID: sessionID, Response: canonical}, true, nil
}

// Claim serializes a server-side play or resolution command. The founder row
// is locked before the session row; this is the C13 account/session lock order.
func (repository *Repository) claim(ctx context.Context, founderID, sessionID string) (Session, error) {
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

func (repository *Repository) completePlay(ctx context.Context, founderID, sessionID, claimToken string, command, state json.RawMessage) (Session, error) {
	return repository.completePlayWithReceipt(ctx, founderID, sessionID, claimToken, command, state, "", "", nil)
}

// CompletePlayWithReceipt commits the nonterminal command, authoritative
// snapshot, and exact API response under one claim token.
func (repository *Repository) CompletePlayWithReceipt(ctx context.Context, founderID, sessionID, claimToken string,
	command, state json.RawMessage, commandID, requestHash string, response json.RawMessage,
) (Session, error) {
	return repository.completePlayWithReceipt(ctx, founderID, sessionID, claimToken, command, state, commandID, requestHash, response)
}

func (repository *Repository) completePlayWithReceipt(ctx context.Context, founderID, sessionID, claimToken string,
	command, state json.RawMessage, commandID, requestHash string, response json.RawMessage,
) (Session, error) {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) ||
		!uuidPattern.MatchString(claimToken) || !validJSONObject(command) || !validJSONObject(state) {
		return Session{}, ErrInvalidSession
	}
	withReceipt := commandID != "" || requestHash != "" || len(response) != 0
	if withReceipt && (!opaqueIDPattern.MatchString(commandID) || !hashPattern.MatchString(requestHash) || !validJSONObject(response)) {
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
	if err := appendSessionCommand(ctx, tx, sessionID, founderID, claimToken, command, resolutionIdentity{}); err != nil {
		return Session{}, err
	}
	if withReceipt {
		if err := InsertCommandReceiptTx(ctx, tx, founderID, sessionID, claimToken, commandID, requestHash, response); err != nil {
			return Session{}, err
		}
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

// InsertCreateReceiptTx binds a successful, transaction-scoped session insert
// to the client's Founder-scoped create key.
func InsertCreateReceiptTx(ctx context.Context, tx *sql.Tx, founderID, idempotencyKey, requestHash, sessionID string, response json.RawMessage) error {
	if tx == nil || !uuidPattern.MatchString(founderID) || !opaqueIDPattern.MatchString(idempotencyKey) ||
		!hashPattern.MatchString(requestHash) || !uuidV7Pattern.MatchString(sessionID) || !validJSONObject(response) {
		return ErrInvalidSession
	}
	result, err := tx.ExecContext(ctx, insertCreateReceiptSQL, founderID, idempotencyKey, requestHash, sessionID, string(response))
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrInvalidSession
	}
	return nil
}

// InsertCommandReceiptTx is called before the claimed→active/resolved update;
// its INSERT SELECT makes a stale claim token incapable of publishing bytes.
func InsertCommandReceiptTx(ctx context.Context, tx *sql.Tx, founderID, sessionID, claimToken, commandID, requestHash string, response json.RawMessage) error {
	if tx == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) ||
		!uuidPattern.MatchString(claimToken) || !opaqueIDPattern.MatchString(commandID) ||
		!hashPattern.MatchString(requestHash) || !validJSONObject(response) {
		return ErrInvalidSession
	}
	result, err := tx.ExecContext(ctx, insertCommandReceiptSQL, sessionID, founderID, claimToken, commandID, requestHash, string(response))
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrClaimLost
	}
	return nil
}

func (repository *Repository) releaseClaim(ctx context.Context, founderID, sessionID, claimToken string) error {
	if repository == nil || !uuidPattern.MatchString(founderID) || !uuidV7Pattern.MatchString(sessionID) ||
		!uuidPattern.MatchString(claimToken) {
		return ErrInvalidSession
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockFounder(ctx, tx, founderID); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, releaseClaimSQL, sessionID, founderID, claimToken)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrClaimLost
	}
	return tx.Commit()
}

// ResolveTx is the session half of C17's company-then-session transaction. The
// caller owns and has already locked the Company stream.
func resolveTx(ctx context.Context, tx *sql.Tx, identity resolutionIdentity, command, state, result,
	receipt json.RawMessage, companyRevision, founderRevision int64,
) (Session, error) {
	if tx == nil || !validResolutionIdentity(identity) ||
		!validJSONObject(command) || !validJSONObject(state) || !validJSONObject(result) || !validJSONObject(receipt) ||
		companyRevision < 1 || companyRevision > 9_007_199_254_740_991 || founderRevision < 1 || founderRevision > 9_007_199_254_740_991 {
		return Session{}, ErrInvalidSession
	}
	if err := appendSessionCommand(ctx, tx, identity.sessionID, identity.founderID, identity.claimToken, command, identity); err != nil {
		return Session{}, err
	}
	resolved, err := scanSession(tx.QueryRowContext(ctx, resolveSessionSQL, identity.sessionID, identity.claimToken,
		string(state), string(result), identity.founderID, identity.companyStreamID, identity.runSeq,
		identity.engineRef, identity.engineVersion, identity.constantsHash, string(receipt), companyRevision, founderRevision))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrClaimLost
	}
	return resolved, err
}

func appendSessionCommand(ctx context.Context, tx *sql.Tx, sessionID, founderID, claimToken string, command json.RawMessage, identity resolutionIdentity) error {
	query := appendPlayCommandSQL
	args := []any{sessionID, founderID, claimToken, []byte(command)}
	if identity.sessionID != "" {
		query = appendResolvedCommandSQL
		args = append(args, identity.companyStreamID, identity.runSeq, identity.engineRef, identity.engineVersion, identity.constantsHash)
	}
	var sequence int64
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&sequence); errors.Is(err, sql.ErrNoRows) {
		return ErrClaimLost
	} else if err != nil {
		return err
	}
	return nil
}

func loadClaimedReplay(ctx context.Context, tx *sql.Tx, identity resolutionIdentity) (Session, []sessionCommand, error) {
	if tx == nil || !validResolutionIdentity(identity) {
		return Session{}, nil, ErrInvalidSession
	}
	session, err := scanSession(tx.QueryRowContext(ctx, loadClaimedReplaySQL, identity.sessionID, identity.founderID,
		identity.companyStreamID, identity.runSeq, identity.engineRef, identity.engineVersion, identity.constantsHash,
		identity.claimToken))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, nil, ErrClaimLost
	}
	if err != nil {
		return Session{}, nil, err
	}
	rows, err := tx.QueryContext(ctx, loadSessionCommandsSQL, identity.sessionID)
	if err != nil {
		return Session{}, nil, err
	}
	defer rows.Close()
	commands := make([]sessionCommand, 0, session.Revision-1)
	for rows.Next() {
		var command sessionCommand
		var raw []byte
		if err := rows.Scan(&command.Sequence, &raw, &command.AppliedRevision, &command.ServerTSMS); err != nil {
			return Session{}, nil, err
		}
		canonical, ok := canonicalJSONObject(raw)
		if !ok || !bytes.Equal(canonical, raw) || command.Sequence != int64(len(commands)+1) ||
			command.AppliedRevision != command.Sequence+1 || command.ServerTSMS < 1 {
			return Session{}, nil, ErrTenantDivergence
		}
		command.Command = canonical
		commands = append(commands, command)
	}
	if err := rows.Err(); err != nil {
		return Session{}, nil, err
	}
	if int64(len(commands)) != session.Revision-1 {
		return Session{}, nil, ErrTenantDivergence
	}
	return session, commands, nil
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
	normalized, ok := normalizeJSONValue(object)
	if !ok {
		return nil, false
	}
	encoded, err := json.Marshal(normalized)
	return encoded, err == nil
}

func normalizeJSONValue(value any) (any, bool) {
	switch typed := value.(type) {
	case nil, bool, string:
		return typed, true
	case json.Number:
		if strings.ContainsAny(typed.String(), ".eE") {
			return nil, false
		}
		integer, err := strconv.ParseInt(typed.String(), 10, 64)
		if err != nil || integer < -9_007_199_254_740_991 || integer > 9_007_199_254_740_991 {
			return nil, false
		}
		return integer, true
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			normalized, ok := normalizeJSONValue(child)
			if !ok {
				return nil, false
			}
			result[index] = normalized
		}
		return result, true
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			normalized, ok := normalizeJSONValue(child)
			if !ok {
				return nil, false
			}
			result[key] = normalized
		}
		return result, true
	default:
		return nil, false
	}
}

type resolutionIdentity struct {
	sessionID       string
	minigameID      string
	founderID       string
	companyStreamID string
	runSeq          int64
	engineRef       string
	engineVersion   string
	constantsHash   string
	claimToken      string
}

func validResolutionIdentity(value resolutionIdentity) bool {
	return uuidV7Pattern.MatchString(value.sessionID) && mechanicalPattern.MatchString(value.minigameID) && uuidPattern.MatchString(value.founderID) &&
		uuidPattern.MatchString(value.companyStreamID) && value.runSeq > 0 && value.runSeq <= 9_007_199_254_740_991 &&
		mechanicalPattern.MatchString(value.engineRef) && versionPattern.MatchString(value.engineVersion) &&
		hashPattern.MatchString(value.constantsHash) && uuidPattern.MatchString(value.claimToken)
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
	"COALESCE(result::text,''),COALESCE(claim_token::text,''),claimed_at,created_at,updated_at,resolved_at," +
	"COALESCE(resolution_receipt::text,''),resolution_company_revision,resolution_founder_revision"

const ownsRunSQL = "SELECT true FROM save_streams s JOIN run_epochs r ON r.company_stream_id=s.id AND r.run_seq=$3 " +
	"WHERE s.id=$1 AND s.owner_kind='founder' AND s.owner_id=$2 AND s.scope='company' AND s.archived_at IS NULL " +
	"AND r.constants_hash=$4"
const createSessionSQL = "INSERT INTO minigame_sessions(session_id,minigame_id,founder_id,company_stream_id,run_seq," +
	"engine_ref,engine_version,constants_hash,scaling_inputs,seed,mode,genesis,state) " +
	"VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12) RETURNING " + sessionColumns
const loadSessionSQL = "SELECT " + sessionColumns + " FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2"
const currentSessionSQL = "SELECT " + sessionColumns + " FROM minigame_sessions WHERE founder_id=$1 AND status IN ('active','claimed')"
const claimSessionSQL = "UPDATE minigame_sessions SET status='claimed',claim_token=gen_random_uuid()," +
	"claimed_at=clock_timestamp(),updated_at=clock_timestamp() WHERE session_id=$1 AND founder_id=$2 AND " +
	"(status='active' OR (status='claimed' AND claimed_at<clock_timestamp()-$3::interval)) RETURNING " + sessionColumns
const statusSQL = "SELECT status FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2"
const completePlaySQL = "UPDATE minigame_sessions SET state=$4,status='active',revision=revision+1,claim_token=NULL," +
	"claimed_at=NULL,updated_at=clock_timestamp() WHERE session_id=$1 AND founder_id=$2 AND status='claimed' " +
	"AND claim_token=$3 RETURNING " + sessionColumns
const releaseClaimSQL = "UPDATE minigame_sessions SET status='active',claim_token=NULL,claimed_at=NULL," +
	"updated_at=clock_timestamp() WHERE session_id=$1 AND founder_id=$2 AND status='claimed' AND claim_token=$3"
const resolveSessionSQL = "UPDATE minigame_sessions SET state=$3,result=$4,status='resolved',revision=revision+1," +
	"claim_token=NULL,claimed_at=NULL,updated_at=clock_timestamp(),resolved_at=clock_timestamp()," +
	"resolution_receipt=$11,resolution_company_revision=$12,resolution_founder_revision=$13 " +
	"WHERE session_id=$1 AND status='claimed' AND claim_token=$2 AND founder_id=$5 AND company_stream_id=$6 " +
	"AND run_seq=$7 AND engine_ref=$8 AND engine_version=$9 AND constants_hash=$10 RETURNING " + sessionColumns
const appendPlayCommandSQL = "INSERT INTO minigame_session_commands(session_id,seq,command,applied_revision,server_ts_ms) " +
	"SELECT session_id,revision,$4,revision+1,floor(extract(epoch FROM clock_timestamp())*1000)::bigint " +
	"FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2 AND status='claimed' AND claim_token=$3 " +
	"RETURNING seq"
const appendResolvedCommandSQL = "INSERT INTO minigame_session_commands(session_id,seq,command,applied_revision,server_ts_ms) " +
	"SELECT session_id,revision,$4,revision+1,floor(extract(epoch FROM clock_timestamp())*1000)::bigint " +
	"FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2 AND status='claimed' AND claim_token=$3 " +
	"AND company_stream_id=$5 AND run_seq=$6 AND engine_ref=$7 AND engine_version=$8 AND constants_hash=$9 RETURNING seq"
const loadClaimedReplaySQL = "SELECT " + sessionColumns + " FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2 " +
	"AND company_stream_id=$3 AND run_seq=$4 AND engine_ref=$5 AND engine_version=$6 AND constants_hash=$7 " +
	"AND status='claimed' AND claim_token=$8 FOR UPDATE"
const loadSessionCommandsSQL = "SELECT seq,command,applied_revision,server_ts_ms FROM minigame_session_commands " +
	"WHERE session_id=$1 ORDER BY seq"
const loadCreateReceiptSQL = "SELECT request_hash,session_id,response::text FROM minigame_create_receipts WHERE founder_id=$1 AND idempotency_key=$2"
const loadCommandReceiptSQL = "SELECT r.request_hash,r.response::text FROM minigame_command_receipts r JOIN minigame_sessions s USING(session_id) " +
	"WHERE r.session_id=$1 AND s.founder_id=$2 AND r.command_id=$3"
const insertCreateReceiptSQL = "INSERT INTO minigame_create_receipts(founder_id,idempotency_key,request_hash,session_id,response) " +
	"SELECT $1,$2,$3,$4,$5 FROM minigame_sessions WHERE session_id=$4 AND founder_id=$1"
const insertCommandReceiptSQL = "INSERT INTO minigame_command_receipts(session_id,command_id,request_hash,response) " +
	"SELECT session_id,$4,$5,$6 FROM minigame_sessions WHERE session_id=$1 AND founder_id=$2 AND status='claimed' AND claim_token=$3"
const lockFounderSQL = "SELECT true FROM account_founders WHERE founder_id=$1 AND archived_at IS NULL FOR UPDATE"

type rowScanner interface {
	Scan(...any) error
}

func scanSession(row rowScanner) (Session, error) {
	var result Session
	var scaling, genesis, state, resolved, resolutionReceipt string
	var claimedAt, resolvedAt sql.NullTime
	var companyRevision, founderRevision sql.NullInt64
	err := row.Scan(&result.SessionID, &result.MinigameID, &result.FounderID, &result.CompanyStreamID,
		&result.RunSeq, &result.EngineRef, &result.EngineVersion, &result.ConstantsHash, &scaling,
		&result.Seed, &result.Mode, &result.Status, &result.Revision, &genesis, &state, &resolved,
		&result.ClaimToken, &claimedAt, &result.CreatedAt, &result.UpdatedAt, &resolvedAt,
		&resolutionReceipt, &companyRevision, &founderRevision)
	if err != nil {
		return Session{}, err
	}
	result.ScalingInputs, _ = canonicalJSONObject([]byte(scaling))
	result.Genesis, _ = canonicalJSONObject([]byte(genesis))
	result.State, _ = canonicalJSONObject([]byte(state))
	if resolved != "" {
		result.Result, _ = canonicalJSONObject([]byte(resolved))
	}
	if resolutionReceipt != "" {
		result.ResolutionReceipt, _ = canonicalJSONObject([]byte(resolutionReceipt))
	}
	if companyRevision.Valid {
		value := companyRevision.Int64
		result.ResolutionCompanyRevision = &value
	}
	if founderRevision.Valid {
		value := founderRevision.Int64
		result.ResolutionFounderRevision = &value
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
