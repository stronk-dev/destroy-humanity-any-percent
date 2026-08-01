package replayverify

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"cloud-clicker/server/kernel"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
)

var (
	uuidPattern                 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	ErrVerificationDeferred     = errors.New("pinned replay engine is unavailable")
	ErrVerificationClaimLost    = errors.New("verification claim lost")
	errDeterministicCatalogData = errors.New("deterministic catalog evidence is invalid")
	errDeterministicReplayData  = errors.New("deterministic replay evidence is invalid")
)

const (
	verificationFailureLimit = 5
	verificationLease        = 5 * time.Minute
	verificationBackoff      = 5 * time.Second
	versionBackoff           = 5 * time.Minute
)

type VerificationInvariant struct {
	Kind     string
	StreamID string
	RunSeq   int64
	Attempts int
	Detail   string
}

type InvariantSink interface {
	ReportVerificationInvariant(VerificationInvariant)
}

type Repository struct {
	db     *sql.DB
	sink   InvariantSink
	verify func(context.Context, string, int64) (production.ReplayVerdict, error)
	// fault is an integration-test seam for failures that database/sql cannot
	// deterministically inject between a successful query and rows.Err.
	fault func(string) error
}

// Projector implementations must be idempotent by (company_stream_id,
// run_seq). The queue also holds and token-checks its claim row in the same
// transaction, but idempotency remains the projector's recovery contract.
type Projector interface {
	ProjectVerifiedRun(context.Context, *sql.Tx, string, int64) error
}

type claim struct {
	StreamID string
	RunSeq   int64
	Token    string
	Attempts int
}

func NewRepository(db *sql.DB, sink InvariantSink) (*Repository, error) {
	if db == nil || sink == nil {
		return nil, errors.New("replay verifier requires database and invariant sink")
	}
	return &Repository{db: db, sink: sink}, nil
}

// ProcessNext claims one per-company head run with a crash lease. A verified
// run is projected and marked in one token-owned transaction. Deterministic
// replay evidence is dead-lettered immediately; transient operational errors
// retry and eventually enter the separate poison lane without inventing a
// gameplay verdict.
func (repository *Repository) ProcessNext(ctx context.Context, projector Projector) (bool, error) {
	if repository == nil || projector == nil {
		return false, errors.New("verification projector required")
	}
	poisoned, err := repository.poisonExpiredClaim(ctx)
	if err != nil || poisoned {
		return poisoned, err
	}
	claimed, err := repository.claimNext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	verify := repository.VerifyStoredRun
	if repository.verify != nil {
		verify = repository.verify
	}
	verdict, verifyErr := verify(ctx, claimed.StreamID, claimed.RunSeq)
	if errors.Is(verifyErr, ErrVerificationDeferred) {
		return true, repository.deferVersion(ctx, claimed)
	}
	if verifyErr != nil {
		return true, repository.failTransient(ctx, claimed, verifyErr)
	}

	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return true, repository.failTransient(ctx, claimed, err)
	}
	defer tx.Rollback()
	if err := ownClaim(ctx, tx, claimed); err != nil {
		return true, err
	}
	if verdict == production.ReplayVerified {
		if err := projector.ProjectVerifiedRun(ctx, tx, claimed.StreamID, claimed.RunSeq); err != nil {
			_ = tx.Rollback()
			return true, repository.failTransient(ctx, claimed, err)
		}
		if err := archiveVerifiedRun(ctx, tx, claimed.StreamID, claimed.RunSeq); err != nil {
			_ = tx.Rollback()
			return true, repository.failTransient(ctx, claimed, err)
		}
		if err := markVerified(ctx, tx, claimed); err != nil {
			return true, err
		}
	} else if err := markDeterministicDead(ctx, tx, claimed, verdict); err != nil {
		return true, err
	}
	return true, tx.Commit()
}

type runArchive struct {
	SchemaVersion   int               `json:"schema_version"`
	CompanyStreamID string            `json:"company_stream_id"`
	RunSeq          int64             `json:"run_seq"`
	Pin             runArchivePin     `json:"pin"`
	Genesis         runArchiveGenesis `json:"genesis"`
	Entries         []runArchiveEntry `json:"entries"`
}

type runArchivePin struct {
	EpochID       int64  `json:"epoch_id"`
	ConstantsHash string `json:"constants_hash"`
	EngineVersion string `json:"engine_version"`
	BuildVCSHash  string `json:"build_vcs_hash"`
	Seed          string `json:"seed"`
}

type runArchiveGenesis struct {
	Version int             `json:"version"`
	State   json.RawMessage `json:"state"`
}

type runArchiveEntry struct {
	Sequence         int64             `json:"seq"`
	IntentID         string            `json:"intent_id"`
	CanonicalPayload json.RawMessage   `json:"canonical_payload"`
	ReplayInputs     json.RawMessage   `json:"replay_inputs"`
	Receipt          json.RawMessage   `json:"receipt"`
	AppliedRevision  *int64            `json:"applied_revision"`
	ServerTSMS       int64             `json:"server_ts_ms"`
	Events           []runArchiveEvent `json:"events"`
}

type runArchiveEvent struct {
	EventID       string          `json:"event_id"`
	StreamID      string          `json:"stream_id"`
	EventSequence int64           `json:"event_seq"`
	Kind          string          `json:"kind"`
	SchemaVersion int             `json:"schema_version"`
	IntentID      string          `json:"intent_id"`
	Payload       json.RawMessage `json:"payload"`
}

func archiveVerifiedRun(ctx context.Context, tx *sql.Tx, streamID string, runSeq int64) error {
	runID := streamID + ":" + fmt.Sprintf("%d", runSeq)
	var existingHash string
	if err := tx.QueryRowContext(ctx, `SELECT sha256 FROM run_log_archive WHERE run_id=$1`, runID).Scan(&existingHash); err == nil {
		var liveRows int64
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND run_seq=$2`, streamID, runSeq).Scan(&liveRows); err != nil {
			return err
		}
		if liveRows != 0 {
			return errors.New("archive exists while active run log remains")
		}
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	archive := runArchive{SchemaVersion: 1, CompanyStreamID: streamID, RunSeq: runSeq, Entries: []runArchiveEntry{}}
	var genesis []byte
	if err := tx.QueryRowContext(ctx, `SELECT p.epoch_id,p.constants_hash,p.engine_version,p.build_vcs_hash,p.seed,g.version,g.state
		FROM run_epochs p JOIN run_genesis g USING(company_stream_id,run_seq)
		WHERE p.company_stream_id=$1 AND p.run_seq=$2`, streamID, runSeq).
		Scan(&archive.Pin.EpochID, &archive.Pin.ConstantsHash, &archive.Pin.EngineVersion, &archive.Pin.BuildVCSHash, &archive.Pin.Seed,
			&archive.Genesis.Version, &genesis); err != nil {
		return err
	}
	if !json.Valid(genesis) {
		return errors.New("invalid run genesis JSON")
	}
	archive.Genesis.State = bytes.Clone(genesis)
	rows, err := tx.QueryContext(ctx, `SELECT seq,intent_id,canonical_payload,replay_inputs::text,receipt::text,applied_revision,server_ts_ms
		FROM run_log WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY seq`, streamID, runSeq)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var entry runArchiveEntry
		var replayInputs, receipt string
		var applied sql.NullInt64
		if err := rows.Scan(&entry.Sequence, &entry.IntentID, &entry.CanonicalPayload, &replayInputs, &receipt, &applied, &entry.ServerTSMS); err != nil {
			return err
		}
		entry.ReplayInputs, entry.Receipt = json.RawMessage(replayInputs), json.RawMessage(receipt)
		if !json.Valid(entry.CanonicalPayload) || !json.Valid(entry.ReplayInputs) || !json.Valid(entry.Receipt) {
			return errors.New("invalid run-log JSON")
		}
		if applied.Valid {
			value := applied.Int64
			entry.AppliedRevision = &value
		}
		archive.Entries = append(archive.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if len(archive.Entries) == 0 {
		return errors.New("cannot archive empty run log")
	}
	for index := range archive.Entries {
		archive.Entries[index].Events, err = archiveEvents(ctx, tx, streamID, archive.Entries[index].IntentID)
		if err != nil {
			return err
		}
	}
	var encoded bytes.Buffer
	zipper, err := gzip.NewWriterLevel(&encoded, gzip.BestCompression)
	if err != nil {
		return err
	}
	zipper.Header.ModTime = time.Unix(0, 0).UTC()
	zipper.Header.OS = 255
	encoder := json.NewEncoder(zipper)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(archive); err != nil {
		_ = zipper.Close()
		return err
	}
	if err := zipper.Close(); err != nil {
		return err
	}
	digest := sha256.Sum256(encoded.Bytes())
	sha := "sha256:" + hex.EncodeToString(digest[:])
	terminalSeq := archive.Entries[len(archive.Entries)-1].Sequence
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_log_archive(run_id,company_stream_id,run_seq,terminal_seq,encoding,bytes,sha256)
		VALUES($1,$2,$3,$4,'gzip+json.v1',$5,$6)`, runID, streamID, runSeq, terminalSeq, encoded.Bytes(), sha); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM run_log WHERE company_stream_id=$1 AND run_seq=$2`, streamID, runSeq)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil || deleted != int64(len(archive.Entries)) {
		return errors.Join(err, errors.New("run-log compaction count mismatch"))
	}
	return nil
}

func archiveEvents(ctx context.Context, tx *sql.Tx, companyStreamID, intentID string) ([]runArchiveEvent, error) {
	rows, err := tx.QueryContext(ctx, `SELECT e.event_id,e.stream_id,e.event_seq,e.kind,e.schema_version,e.intent_id,e.payload::text
		FROM events e
		JOIN save_streams company ON company.id=$1 AND company.owner_kind='founder' AND company.scope='company'
		WHERE e.intent_id=$2 AND (e.stream_id=company.id OR e.stream_id IN (
			SELECT founder.id FROM save_streams founder
			WHERE founder.owner_kind='founder' AND founder.scope='founder' AND founder.owner_id=company.owner_id
		)) ORDER BY CASE WHEN e.stream_id=company.id THEN 1 ELSE 0 END,e.event_seq,e.event_id`, companyStreamID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []runArchiveEvent{}
	for rows.Next() {
		var event runArchiveEvent
		var payload string
		if err := rows.Scan(&event.EventID, &event.StreamID, &event.EventSequence, &event.Kind, &event.SchemaVersion, &event.IntentID, &payload); err != nil {
			return nil, err
		}
		event.Payload = json.RawMessage(payload)
		if !json.Valid(event.Payload) {
			return nil, errors.New("invalid archive event payload")
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (repository *Repository) claimNext(ctx context.Context) (claim, error) {
	var result claim
	err := repository.db.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT q.company_stream_id,q.run_seq
			FROM verification_queue q
			WHERE (((q.status='pending' AND q.available_at<=clock_timestamp()) OR
			         (q.status='claimed' AND q.claimed_at<clock_timestamp()-$1::interval)))
			  AND q.attempts < $2
			  AND NOT EXISTS (
				SELECT 1 FROM verification_queue earlier
				WHERE earlier.company_stream_id=q.company_stream_id
				  AND earlier.run_seq<q.run_seq
				  AND earlier.status IN ('pending','claimed')
			  )
			ORDER BY q.available_at,q.company_stream_id,q.run_seq
			FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE verification_queue q
		SET status='claimed',claimed_at=clock_timestamp(),claim_token=gen_random_uuid(),attempts=attempts+1
		FROM candidate c
		WHERE q.company_stream_id=c.company_stream_id AND q.run_seq=c.run_seq
		RETURNING q.company_stream_id,q.run_seq,q.claim_token,q.attempts`,
		intervalLiteral(verificationLease), verificationFailureLimit).
		Scan(&result.StreamID, &result.RunSeq, &result.Token, &result.Attempts)
	return result, err
}

func (repository *Repository) poisonExpiredClaim(ctx context.Context) (bool, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var report VerificationInvariant
	var detail sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT company_stream_id,run_seq,attempts,last_error
		FROM verification_queue
		WHERE attempts >= $1 AND (
			(status='pending' AND available_at<=clock_timestamp()) OR
			(status='claimed' AND claimed_at<clock_timestamp()-$2::interval)
		)
		ORDER BY available_at,company_stream_id,run_seq
		FOR UPDATE SKIP LOCKED LIMIT 1`, verificationFailureLimit, intervalLiteral(verificationLease)).
		Scan(&report.StreamID, &report.RunSeq, &report.Attempts, &detail)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	report.Kind = "verification_poison_dead_letter"
	report.Detail = "verification claim expired at failure limit"
	if detail.Valid && strings.TrimSpace(detail.String) != "" {
		report.Detail = boundedDetail(errors.New(detail.String))
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO verification_poison_dead_letters(company_stream_id,run_seq,attempts,detail)
		VALUES($1,$2,$3,$4)`, report.StreamID, report.RunSeq, report.Attempts, report.Detail); err != nil {
		return false, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE verification_queue
		SET status='dead',verdict=NULL,completed_at=clock_timestamp(),claimed_at=NULL,claim_token=NULL,last_error=$3
		WHERE company_stream_id=$1 AND run_seq=$2 AND status IN ('pending','claimed')`, report.StreamID, report.RunSeq, report.Detail)
	if err = requireOne(result, err); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	repository.sink.ReportVerificationInvariant(report)
	return true, nil
}

func ownClaim(ctx context.Context, tx *sql.Tx, claimed claim) error {
	var owned bool
	err := tx.QueryRowContext(ctx, `SELECT true FROM verification_queue
		WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3
		FOR UPDATE`, claimed.StreamID, claimed.RunSeq, claimed.Token).Scan(&owned)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVerificationClaimLost
	}
	return err
}

func markVerified(ctx context.Context, tx *sql.Tx, claimed claim) error {
	result, err := tx.ExecContext(ctx, `UPDATE verification_queue
		SET status='verified',verdict='verified',completed_at=clock_timestamp(),claimed_at=NULL,claim_token=NULL,last_error=NULL
		WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3`,
		claimed.StreamID, claimed.RunSeq, claimed.Token)
	return requireOne(result, err)
}

func markDeterministicDead(ctx context.Context, tx *sql.Tx, claimed claim, verdict production.ReplayVerdict) error {
	if verdict == production.ReplayVerified || !knownNonVerifiedVerdict(verdict) {
		return errors.New("invalid deterministic replay verdict")
	}
	detail := "replay verification returned " + string(verdict)
	if _, err := tx.ExecContext(ctx, `INSERT INTO verification_dead_letters(company_stream_id,run_seq,verdict,detail)
		VALUES($1,$2,$3,$4)`, claimed.StreamID, claimed.RunSeq, verdict, detail); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE verification_queue
		SET status='dead',verdict=$4,completed_at=clock_timestamp(),claimed_at=NULL,claim_token=NULL,last_error=$5
		WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3`,
		claimed.StreamID, claimed.RunSeq, claimed.Token, verdict, detail)
	return requireOne(result, err)
}

func (repository *Repository) deferVersion(ctx context.Context, claimed claim) error {
	result, err := repository.db.ExecContext(ctx, `UPDATE verification_queue
		SET status='pending',attempts=GREATEST(attempts-1,0),available_at=clock_timestamp()+$4::interval,
		    claimed_at=NULL,claim_token=NULL,last_error=$5
		WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3`,
		claimed.StreamID, claimed.RunSeq, claimed.Token, intervalLiteral(versionBackoff), ErrVerificationDeferred.Error())
	return requireOne(result, err)
}

func (repository *Repository) failTransient(ctx context.Context, claimed claim, cause error) error {
	detail := boundedDetail(cause)
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Join(cause, err)
	}
	defer tx.Rollback()
	if err := ownClaim(ctx, tx, claimed); err != nil {
		return errors.Join(cause, err)
	}
	if claimed.Attempts < verificationFailureLimit {
		result, updateErr := tx.ExecContext(ctx, `UPDATE verification_queue
			SET status='pending',available_at=clock_timestamp()+$4::interval,
			    claimed_at=NULL,claim_token=NULL,last_error=$5
			WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3`,
			claimed.StreamID, claimed.RunSeq, claimed.Token, intervalLiteral(verificationBackoff), detail)
		if updateErr = requireOne(result, updateErr); updateErr != nil {
			return errors.Join(cause, updateErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.Join(cause, commitErr)
		}
		return cause
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO verification_poison_dead_letters(company_stream_id,run_seq,attempts,detail)
		VALUES($1,$2,$3,$4)`, claimed.StreamID, claimed.RunSeq, claimed.Attempts, detail); err != nil {
		return errors.Join(cause, err)
	}
	result, err := tx.ExecContext(ctx, `UPDATE verification_queue
		SET status='dead',verdict=NULL,completed_at=clock_timestamp(),claimed_at=NULL,claim_token=NULL,last_error=$4
		WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed' AND claim_token=$3`,
		claimed.StreamID, claimed.RunSeq, claimed.Token, detail)
	if err = requireOne(result, err); err != nil {
		return errors.Join(cause, err)
	}
	if err := tx.Commit(); err != nil {
		return errors.Join(cause, err)
	}
	repository.sink.ReportVerificationInvariant(VerificationInvariant{Kind: "verification_poison_dead_letter",
		StreamID: claimed.StreamID, RunSeq: claimed.RunSeq, Attempts: claimed.Attempts, Detail: detail})
	return cause
}

// VerifyStoredRun maps legacy NULL inputs to log_gap and derives every verdict
// from immutable database evidence. Operational failures are returned as
// errors; they never become permanent verdicts.
func (repository *Repository) VerifyStoredRun(ctx context.Context, streamID string, runSeq int64) (production.ReplayVerdict, error) {
	if repository == nil || !uuidPattern.MatchString(streamID) || runSeq < 1 {
		return production.ReplayStateDivergence, errors.New("invalid run")
	}
	var hash, engine string
	var genesis []byte
	var version int
	if err := repository.db.QueryRowContext(ctx, `SELECT p.constants_hash,p.engine_version,g.state,g.version
		FROM run_epochs p JOIN run_genesis g USING(company_stream_id,run_seq)
		WHERE p.company_stream_id=$1 AND p.run_seq=$2`, streamID, runSeq).Scan(&hash, &engine, &genesis, &version); errors.Is(err, sql.ErrNoRows) {
		return production.ReplayLogGap, nil
	} else if err != nil {
		return production.ReplayStateDivergence, err
	}
	var drift bool
	if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM run_version_drift WHERE company_stream_id=$1 AND run_seq=$2)`, streamID, runSeq).Scan(&drift); err != nil {
		return production.ReplayStateDivergence, err
	}
	if drift {
		return production.ReplayEngineMismatch, nil
	}
	if engine != kernel.Version {
		return production.ReplayStateDivergence, ErrVerificationDeferred
	}
	var legacyGap bool
	if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM run_log WHERE company_stream_id=$1 AND run_seq=$2 AND replay_inputs IS NULL
	)`, streamID, runSeq).Scan(&legacyGap); err != nil {
		return production.ReplayStateDivergence, err
	}
	if legacyGap {
		return production.ReplayLogGap, nil
	}

	bundles := map[string]production.CatalogBundle{}
	loadBundle := func(constantsHash string) (production.CatalogBundle, error) {
		if bundle, ok := bundles[constantsHash]; ok {
			return bundle, nil
		}
		artifacts, err := repository.artifacts(ctx, constantsHash)
		if err != nil {
			return production.CatalogBundle{}, err
		}
		bundle, err := replaycatalog.Load(constantsHash, artifacts)
		if err != nil {
			return production.CatalogBundle{}, fmt.Errorf("%w: %v", errDeterministicCatalogData, err)
		}
		bundles[constantsHash] = bundle
		return bundle, nil
	}
	bundle, err := loadBundle(hash)
	if errors.Is(err, errDeterministicCatalogData) {
		return production.ReplayConstantsMismatch, nil
	}
	if err != nil {
		return production.ReplayStateDivergence, err
	}

	rows, err := repository.db.QueryContext(ctx, `SELECT seq,canonical_payload,replay_inputs::text,receipt::text
		FROM run_log WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY seq`, streamID, runSeq)
	if err != nil {
		return production.ReplayStateDivergence, err
	}
	defer rows.Close()
	entries := []production.ReplayLogEntry{}
	for rows.Next() {
		var entry production.ReplayLogEntry
		var inputs sql.NullString
		var receipt string
		if err := rows.Scan(&entry.Sequence, &entry.CanonicalPayload, &inputs, &receipt); err != nil {
			return production.ReplayStateDivergence, err
		}
		if !inputs.Valid {
			return production.ReplayLogGap, nil
		}
		entry.ReplayInputs, entry.ReceiptJSON = []byte(inputs.String), []byte(receipt)
		var envelope struct {
			Command struct {
				IntentID string `json:"intent_id"`
			} `json:"command"`
			Resolved struct {
				Kind              string `json:"kind"`
				NextConstantsHash string `json:"next_constants_hash"`
			} `json:"resolved"`
		}
		if json.Unmarshal(entry.ReplayInputs, &envelope) != nil || !uuidPattern.MatchString(envelope.Command.IntentID) {
			return production.ReplayStateDivergence, nil
		}
		entry.Terminal = envelope.Resolved.Kind == "exit"
		entry.EventsJSON, err = repository.events(ctx, streamID, envelope.Command.IntentID)
		if errors.Is(err, errDeterministicReplayData) {
			return production.ReplayStateDivergence, nil
		}
		if err != nil {
			return production.ReplayStateDivergence, err
		}
		if entry.Terminal && envelope.Resolved.NextConstantsHash != "" && envelope.Resolved.NextConstantsHash != hash {
			next, loadErr := loadBundle(envelope.Resolved.NextConstantsHash)
			if errors.Is(loadErr, errDeterministicCatalogData) {
				return production.ReplayConstantsMismatch, nil
			}
			if loadErr != nil {
				return production.ReplayStateDivergence, loadErr
			}
			entry.NextCatalog = &next
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return production.ReplayStateDivergence, err
	}
	if err := repository.runFault("run_log_rows"); err != nil {
		return production.ReplayStateDivergence, err
	}
	return production.VerifyReplayRun(genesis, version, bundle, entries, hash, false), nil
}

func (repository *Repository) artifacts(ctx context.Context, hash string) (map[string][]byte, error) {
	if err := repository.runFault("catalog_query"); err != nil {
		return nil, err
	}
	rows, err := repository.db.QueryContext(ctx, `SELECT artifact_name,bytes FROM catalog_artifacts WHERE constants_hash=$1`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string][]byte{}
	for rows.Next() {
		var name string
		var data []byte
		if err := rows.Scan(&name, &data); err != nil {
			return nil, err
		}
		result[name] = bytes.Clone(data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := repository.runFault("catalog_rows"); err != nil {
		return nil, err
	}
	if len(result) != 7 {
		return nil, fmt.Errorf("%w: incomplete artifact bundle", errDeterministicCatalogData)
	}
	return result, nil
}

func (repository *Repository) events(ctx context.Context, companyStreamID, intentID string) ([]byte, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT e.kind,e.schema_version,e.intent_id,e.payload::text
		FROM events e
		JOIN save_streams company ON company.id=$1 AND company.owner_kind='founder' AND company.scope='company'
		WHERE e.intent_id=$2 AND (
			e.stream_id=company.id OR e.stream_id IN (
				SELECT founder.id FROM save_streams founder
				WHERE founder.owner_kind='founder' AND founder.scope='founder' AND founder.owner_id=company.owner_id
			)
		)
		ORDER BY CASE WHEN e.stream_id=company.id THEN 1 ELSE 0 END,e.event_seq,e.event_id`, companyStreamID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []map[string]any{}
	for rows.Next() {
		var kind, id, payloadText string
		var schema int
		if err := rows.Scan(&kind, &schema, &id, &payloadText); err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewBufferString(payloadText))
		decoder.UseNumber()
		var payload any
		if decoder.Decode(&payload) != nil {
			return nil, fmt.Errorf("%w: invalid event payload", errDeterministicReplayData)
		}
		values = append(values, map[string]any{"kind": kind, "schema_version": schema, "intent_id": id, "payload": payload})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := repository.runFault("event_rows"); err != nil {
		return nil, err
	}
	return json.Marshal(values)
}

func (repository *Repository) runFault(step string) error {
	if repository.fault == nil {
		return nil
	}
	return repository.fault(step)
}

func knownNonVerifiedVerdict(verdict production.ReplayVerdict) bool {
	switch verdict {
	case production.ReplayLogGap, production.ReplayStateDivergence, production.ReplayConstantsMismatch,
		production.ReplayClockViolation, production.ReplayEngineMismatch:
		return true
	default:
		return false
	}
}

func requireOne(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrVerificationClaimLost
	}
	return nil
}

func boundedDetail(err error) string {
	detail := strings.TrimSpace(err.Error())
	if detail == "" || !utf8.ValidString(detail) {
		detail = "verification operation failed"
	}
	runes := []rune(detail)
	if len(runes) > 512 {
		detail = string(runes[:512])
	}
	return detail
}

func intervalLiteral(duration time.Duration) string {
	return fmt.Sprintf("%d milliseconds", duration.Milliseconds())
}
