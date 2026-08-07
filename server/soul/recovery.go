package soul

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

const recoveryClaimLease = 5 * time.Minute

var (
	ErrInvalidRecovery     = errors.New("invalid soul recovery session")
	ErrRecoveryActive      = errors.New("soul recovery session already active")
	ErrRecoveryBusy        = errors.New("soul recovery session is busy")
	ErrRecoveryGone        = errors.New("soul recovery session is unavailable")
	ErrRecoveryClaimLost   = errors.New("soul recovery claim lost")
	ErrRecoveryIdempotency = errors.New("soul recovery idempotency conflict")
	ErrRecoveryToken       = errors.New("invalid soul recovery progress token")
	recoveryUUIDPattern    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	recoveryUUIDV7Pattern  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	recoveryHashPattern    = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	recoveryMechanicalID   = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type RecoveryStatus string

const (
	RecoveryActive    RecoveryStatus = "active"
	RecoveryClaimed   RecoveryStatus = "claimed"
	RecoveryResolved  RecoveryStatus = "resolved"
	RecoveryCancelled RecoveryStatus = "cancelled"
)

type RecoverySession struct {
	SessionID              string
	FounderID              string
	FounderStreamID        string
	CompanyStreamID        string
	RunSeq                 int64
	ConstantsHash          string
	ActivityID             string
	FounderAttendedStartMS int64
	RequiredDurationMS     int64
	Status                 RecoveryStatus
	StartRequestHash       string
	TerminalRequestHash    string
	ClaimToken             string
	ProgressToken          string
	AttendedProgressMS     int64
	LastProgressServerMS   int64
	ClaimedAt              *time.Time
	TerminalReceipt        json.RawMessage
	CreatedAt              time.Time
	UpdatedAt              time.Time
	TerminalAt             *time.Time
}

type StartRecovery struct {
	SessionID              string
	FounderID              string
	FounderStreamID        string
	CompanyStreamID        string
	RunSeq                 int64
	ConstantsHash          string
	ActivityID             string
	FounderAttendedStartMS int64
	RequiredDurationMS     int64
	FounderRevision        int64
	CompanyRevision        int64
	RequestHash            string
	ServerNowMS            int64
}

type ProgressRecovery struct {
	SessionID             string
	FounderID             string
	ProgressToken         string
	ServerNowMS           int64
	RecoveryBeatCeilingMS int64
}

type RecoveryRepository struct{ db *sql.DB }

func NewRecoveryRepository(db *sql.DB) (*RecoveryRepository, error) {
	if db == nil {
		return nil, ErrInvalidRecovery
	}
	return &RecoveryRepository{db: db}, nil
}

// Start serializes on the Founder identity, proves both supplied stream
// coordinates are still current, and commits only the authoritative session
// row plus its Founder event. It intentionally creates no replay-log row:
// suppression does not begin until a terminal coordinator command.
func (repository *RecoveryRepository) Start(ctx context.Context, input StartRecovery) (RecoverySession, error) {
	if repository == nil || !validStartRecovery(input) {
		return RecoverySession{}, ErrInvalidRecovery
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return RecoverySession{}, err
	}
	defer tx.Rollback()
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT true FROM account_founders WHERE founder_id=$1 AND archived_at IS NULL FOR UPDATE`, input.FounderID).Scan(&active); errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, ErrRecoveryGone
	} else if err != nil {
		return RecoverySession{}, err
	}
	if existing, loadErr := loadRecoveryRow(tx.QueryRowContext(ctx, recoveryActiveByFounderSQL+` FOR UPDATE`, input.FounderID)); loadErr == nil {
		rotated, rotateErr := loadRecoveryRow(tx.QueryRowContext(ctx, rotateRecoveryProgressTokenSQL, existing.SessionID))
		if rotateErr != nil {
			return RecoverySession{}, rotateErr
		}
		if err := tx.Commit(); err != nil {
			return RecoverySession{}, err
		}
		return rotated, nil
	} else if !errors.Is(loadErr, sql.ErrNoRows) {
		return RecoverySession{}, loadErr
	}
	var otherExclusive bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM minigame_sessions WHERE founder_id=$1 AND status IN ('active','claimed'))`,
		input.FounderID).Scan(&otherExclusive); err != nil {
		return RecoverySession{}, err
	}
	if otherExclusive {
		return RecoverySession{}, ErrRecoveryActive
	}
	if err := verifyRecoveryCoordinate(ctx, tx, input); err != nil {
		return RecoverySession{}, err
	}
	created, err := loadRecoveryRow(tx.QueryRowContext(ctx, createRecoverySQL,
		input.SessionID, input.FounderID, input.FounderStreamID, input.CompanyStreamID, input.RunSeq,
		input.ConstantsHash, input.ActivityID, input.FounderAttendedStartMS, input.RequiredDurationMS, input.RequestHash,
		input.ServerNowMS))
	if err != nil {
		return RecoverySession{}, err
	}
	payload, _ := json.Marshal(map[string]any{"session_id": input.SessionID, "activity_id": input.ActivityID,
		"company_stream_id": input.CompanyStreamID, "run_seq": input.RunSeq,
		"founder_attended_start_ms": input.FounderAttendedStartMS, "required_duration_ms": input.RequiredDurationMS})
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
		VALUES($1,$2,1,'soul_recovery_started.v1',$3,$4,$5)`, input.FounderStreamID, input.FounderRevision,
		input.SessionID, input.ConstantsHash, payload); err != nil {
		return RecoverySession{}, err
	}
	if err := tx.Commit(); err != nil {
		return RecoverySession{}, err
	}
	return created, nil
}

// RequireInactiveTx is the race-safe guard used after a Company or Founder
// stream lock has serialized against session start. A literal command retry is
// resolved by the owning Store before this guard runs.
func (repository *RecoveryRepository) RequireInactiveTx(ctx context.Context, tx *sql.Tx, founderID, _ string) error {
	if repository == nil || tx == nil || !recoveryUUIDPattern.MatchString(founderID) {
		return ErrInvalidRecovery
	}
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM soul_recovery_sessions WHERE founder_id=$1 AND status IN ('active','claimed'))`,
		founderID).Scan(&active); err != nil {
		return err
	}
	if active {
		return ErrRecoveryActive
	}
	return nil
}

func (repository *RecoveryRepository) Load(ctx context.Context, founderID, sessionID string) (RecoverySession, error) {
	if repository == nil || !recoveryUUIDPattern.MatchString(founderID) || !recoveryUUIDV7Pattern.MatchString(sessionID) {
		return RecoverySession{}, ErrInvalidRecovery
	}
	result, err := loadRecoveryRow(repository.db.QueryRowContext(ctx, recoveryByIDSQL, sessionID, founderID))
	if errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, ErrRecoveryGone
	}
	return result, err
}

func (repository *RecoveryRepository) HasActive(ctx context.Context, founderID string) (bool, error) {
	if repository == nil || !recoveryUUIDPattern.MatchString(founderID) {
		return false, ErrInvalidRecovery
	}
	var found bool
	err := repository.db.QueryRowContext(ctx, `SELECT true FROM soul_recovery_sessions WHERE founder_id=$1 AND status IN ('active','claimed')`, founderID).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return found, err
}

func (repository *RecoveryRepository) Active(ctx context.Context, founderID string) (RecoverySession, error) {
	if repository == nil || !recoveryUUIDPattern.MatchString(founderID) {
		return RecoverySession{}, ErrInvalidRecovery
	}
	result, err := loadRecoveryRow(repository.db.QueryRowContext(ctx, recoveryActiveByFounderSQL, founderID))
	if errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, ErrRecoveryGone
	}
	return result, err
}

// Progress is the only non-terminal writer for attended recovery progress.
// It serializes on the session row, compares the capability in constant time,
// and deliberately emits no event or replay row.
func (repository *RecoveryRepository) Progress(ctx context.Context, input ProgressRecovery) (RecoverySession, error) {
	if repository == nil || !recoveryUUIDV7Pattern.MatchString(input.SessionID) ||
		!recoveryUUIDPattern.MatchString(input.FounderID) || !recoveryUUIDPattern.MatchString(input.ProgressToken) ||
		input.ServerNowMS < 0 || input.ServerNowMS > 9_007_199_254_740_991 || input.RecoveryBeatCeilingMS < 1 {
		return RecoverySession{}, ErrInvalidRecovery
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return RecoverySession{}, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT true FROM account_founders WHERE founder_id=$1 AND archived_at IS NULL FOR UPDATE`, input.FounderID).Scan(new(bool)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RecoverySession{}, ErrRecoveryGone
		}
		return RecoverySession{}, err
	}
	session, err := loadRecoveryRow(tx.QueryRowContext(ctx, recoveryByIDSQL+` FOR UPDATE`, input.SessionID, input.FounderID))
	if errors.Is(err, sql.ErrNoRows) || err == nil && session.Status != RecoveryActive {
		return RecoverySession{}, ErrRecoveryGone
	}
	if err != nil {
		return RecoverySession{}, err
	}
	if subtle.ConstantTimeCompare([]byte(session.ProgressToken), []byte(input.ProgressToken)) != 1 {
		return RecoverySession{}, ErrRecoveryToken
	}
	gap := input.ServerNowMS - session.LastProgressServerMS
	if gap < 0 {
		return RecoverySession{}, ErrInvalidRecovery
	}
	increment := int64(0)
	if gap <= input.RecoveryBeatCeilingMS {
		increment = gap
	}
	if session.AttendedProgressMS > 9_007_199_254_740_991-increment {
		return RecoverySession{}, ErrInvalidRecovery
	}
	updated, err := loadRecoveryRow(tx.QueryRowContext(ctx, progressRecoverySQL, session.SessionID,
		session.AttendedProgressMS+increment, input.ServerNowMS))
	if err != nil {
		return RecoverySession{}, err
	}
	if err := tx.Commit(); err != nil {
		return RecoverySession{}, err
	}
	return updated, nil
}

// ClaimTx runs after the save coordinator has acquired Founder then Company
// locks. Expired leases recover; terminal rows return their durable receipt.
func (repository *RecoveryRepository) ClaimTx(ctx context.Context, tx *sql.Tx, founderID, sessionID, requestHash string) (RecoverySession, error) {
	if repository == nil || tx == nil || !recoveryUUIDPattern.MatchString(founderID) ||
		!recoveryUUIDV7Pattern.MatchString(sessionID) || !recoveryHashPattern.MatchString(requestHash) {
		return RecoverySession{}, ErrInvalidRecovery
	}
	claimed, err := loadRecoveryRow(tx.QueryRowContext(ctx, claimRecoverySQL, sessionID, founderID,
		recoveryClaimLease.Milliseconds(), requestHash))
	if err == nil {
		return claimed, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, err
	}
	existing, err := loadRecoveryRow(tx.QueryRowContext(ctx, recoveryByIDSQL+` FOR UPDATE`, sessionID, founderID))
	if errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, ErrRecoveryGone
	}
	if err != nil {
		return RecoverySession{}, err
	}
	if existing.Status == RecoveryResolved || existing.Status == RecoveryCancelled {
		if existing.TerminalRequestHash != requestHash {
			return RecoverySession{}, ErrRecoveryIdempotency
		}
		return existing, nil
	}
	return RecoverySession{}, ErrRecoveryBusy
}

func (repository *RecoveryRepository) FinishTx(ctx context.Context, tx *sql.Tx, sessionID, claimToken string,
	status RecoveryStatus, requestHash string, receipt json.RawMessage,
) (RecoverySession, error) {
	if repository == nil || tx == nil || !recoveryUUIDV7Pattern.MatchString(sessionID) || !recoveryUUIDPattern.MatchString(claimToken) ||
		(status != RecoveryResolved && status != RecoveryCancelled) || !recoveryHashPattern.MatchString(requestHash) || !validRecoveryReceipt(receipt) {
		return RecoverySession{}, ErrInvalidRecovery
	}
	result, err := loadRecoveryRow(tx.QueryRowContext(ctx, finishRecoverySQL, sessionID, claimToken, status, requestHash, receipt))
	if errors.Is(err, sql.ErrNoRows) {
		return RecoverySession{}, ErrRecoveryClaimLost
	}
	return result, err
}

func verifyRecoveryCoordinate(ctx context.Context, tx *sql.Tx, input StartRecovery) error {
	var founderRevision int64
	var founderHash string
	if err := tx.QueryRowContext(ctx, `SELECT r.revision,r.constants_hash FROM save_streams s JOIN LATERAL
		(SELECT revision,constants_hash FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
		WHERE s.id=$1 AND s.owner_kind='founder' AND s.owner_id=$2 AND s.scope='founder' AND s.archived_at IS NULL FOR UPDATE OF s`,
		input.FounderStreamID, input.FounderID).Scan(&founderRevision, &founderHash); err != nil {
		return err
	}
	var companyRevision, runSeq int64
	var companyHash string
	if err := tx.QueryRowContext(ctx, `SELECT r.revision,r.constants_hash,(r.state->>'run_seq')::bigint FROM save_streams s JOIN LATERAL
		(SELECT revision,constants_hash,state FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
		WHERE s.id=$1 AND s.owner_kind='founder' AND s.owner_id=$2 AND s.scope='company' AND s.archived_at IS NULL FOR UPDATE OF s`,
		input.CompanyStreamID, input.FounderID).Scan(&companyRevision, &companyHash, &runSeq); err != nil {
		return err
	}
	if founderRevision != input.FounderRevision || companyRevision != input.CompanyRevision || runSeq != input.RunSeq ||
		founderHash != input.ConstantsHash || companyHash != input.ConstantsHash {
		return ErrInvalidRecovery
	}
	return nil
}

func validStartRecovery(input StartRecovery) bool {
	return recoveryUUIDV7Pattern.MatchString(input.SessionID) && recoveryUUIDPattern.MatchString(input.FounderID) &&
		recoveryUUIDPattern.MatchString(input.FounderStreamID) && recoveryUUIDPattern.MatchString(input.CompanyStreamID) &&
		input.RunSeq > 0 && input.RunSeq <= 9_007_199_254_740_991 && recoveryHashPattern.MatchString(input.ConstantsHash) &&
		recoveryMechanicalID.MatchString(input.ActivityID) && input.FounderAttendedStartMS >= 0 &&
		input.FounderAttendedStartMS <= 9_007_199_254_740_991 && input.RequiredDurationMS > 0 &&
		input.RequiredDurationMS <= 9_007_199_254_740_991 && input.FounderRevision > 0 && input.CompanyRevision > 0 &&
		recoveryHashPattern.MatchString(input.RequestHash) && input.ServerNowMS >= 0 && input.ServerNowMS <= 9_007_199_254_740_991
}

func validRecoveryReceipt(receipt []byte) bool {
	var object map[string]json.RawMessage
	return len(receipt) > 0 && json.Valid(receipt) && json.Unmarshal(receipt, &object) == nil && object != nil
}

type recoveryScanner interface{ Scan(...any) error }

func loadRecoveryRow(row recoveryScanner) (RecoverySession, error) {
	var result RecoverySession
	var claimToken, progressToken, requestHash string
	var receipt []byte
	var claimedAt, terminalAt sql.NullTime
	err := row.Scan(&result.SessionID, &result.FounderID, &result.FounderStreamID, &result.CompanyStreamID,
		&result.RunSeq, &result.ConstantsHash, &result.ActivityID, &result.FounderAttendedStartMS,
		&result.RequiredDurationMS, &result.Status, &result.StartRequestHash, &requestHash, &claimToken,
		&progressToken, &result.AttendedProgressMS, &result.LastProgressServerMS,
		&claimedAt, &receipt, &result.CreatedAt, &result.UpdatedAt, &terminalAt)
	if err != nil {
		return RecoverySession{}, err
	}
	result.TerminalRequestHash, result.ClaimToken, result.ProgressToken = requestHash, claimToken, progressToken
	if claimedAt.Valid {
		value := claimedAt.Time
		result.ClaimedAt = &value
	}
	if terminalAt.Valid {
		value := terminalAt.Time
		result.TerminalAt = &value
	}
	if len(receipt) != 0 {
		result.TerminalReceipt = append(json.RawMessage(nil), receipt...)
	}
	return result, nil
}

const recoveryColumns = `session_id,founder_id,founder_stream_id,company_stream_id,run_seq,constants_hash,activity_id,
	founder_attended_start_ms,required_duration_ms,status,start_request_hash,COALESCE(terminal_request_hash,''),
	COALESCE(claim_token::text,''),progress_token::text,attended_progress_ms,last_progress_server_ms,
	claimed_at,COALESCE(terminal_receipt::text,'')::bytea,created_at,updated_at,terminal_at`
const recoveryByIDSQL = `SELECT ` + recoveryColumns + ` FROM soul_recovery_sessions WHERE session_id=$1 AND founder_id=$2`
const recoveryActiveByFounderSQL = `SELECT ` + recoveryColumns + ` FROM soul_recovery_sessions WHERE founder_id=$1 AND status IN ('active','claimed')`
const createRecoverySQL = `INSERT INTO soul_recovery_sessions(session_id,founder_id,founder_stream_id,company_stream_id,run_seq,
	constants_hash,activity_id,founder_attended_start_ms,required_duration_ms,start_request_hash,progress_token,
	attended_progress_ms,last_progress_server_ms,created_at,updated_at)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,gen_random_uuid(),0,$11::bigint,
		to_timestamp(($11::bigint)::double precision/1000),to_timestamp(($11::bigint)::double precision/1000)) RETURNING ` + recoveryColumns
const rotateRecoveryProgressTokenSQL = `UPDATE soul_recovery_sessions SET progress_token=gen_random_uuid(),updated_at=clock_timestamp()
	WHERE session_id=$1 AND status='active' RETURNING ` + recoveryColumns
const progressRecoverySQL = `UPDATE soul_recovery_sessions SET attended_progress_ms=$2,last_progress_server_ms=$3,updated_at=clock_timestamp()
	WHERE session_id=$1 AND status='active' RETURNING ` + recoveryColumns
const claimRecoverySQL = `UPDATE soul_recovery_sessions SET status='claimed',claim_token=gen_random_uuid(),claimed_at=clock_timestamp(),
	updated_at=clock_timestamp(),terminal_request_hash=$4 WHERE session_id=$1 AND founder_id=$2 AND
	(status='active' OR (status='claimed' AND claimed_at<clock_timestamp()-($3::bigint*interval '1 millisecond')))
	AND (terminal_request_hash IS NULL OR terminal_request_hash=$4) RETURNING ` + recoveryColumns
const finishRecoverySQL = `UPDATE soul_recovery_sessions SET status=$3,terminal_request_hash=$4,terminal_receipt=$5,
	claim_token=NULL,claimed_at=NULL,terminal_at=clock_timestamp(),updated_at=clock_timestamp()
	WHERE session_id=$1 AND status='claimed' AND claim_token=$2 AND terminal_request_hash=$4 RETURNING ` + recoveryColumns
