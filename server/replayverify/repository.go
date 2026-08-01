package replayverify

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"cloud-clicker/server/kernel"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type Repository struct{ db *sql.DB }

type Projector interface {
	ProjectVerifiedRun(context.Context, *sql.Tx, string, int64) error
}

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("replay verifier requires database")
	}
	return &Repository{db: db}, nil
}

// ProcessNext claims one ended run with a five-minute crash lease. A verified
// run is projected and marked in one transaction; every other verdict is
// dead-lettered immutably. Projection is mandatory so no unprojected run can
// be marked verified.
func (repository *Repository) ProcessNext(ctx context.Context, projector Projector) (bool, error) {
	if repository == nil || projector == nil {
		return false, errors.New("verification projector required")
	}
	var streamID string
	var runSeq int64
	err := repository.db.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT company_stream_id,run_seq FROM verification_queue
			WHERE (status='pending' AND available_at<=now()) OR (status='claimed' AND claimed_at<now()-interval '5 minutes')
			ORDER BY available_at,company_stream_id,run_seq FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE verification_queue q SET status='claimed',claimed_at=now(),attempts=attempts+1
		FROM candidate c WHERE q.company_stream_id=c.company_stream_id AND q.run_seq=c.run_seq
		RETURNING q.company_stream_id,q.run_seq`).Scan(&streamID, &runSeq)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	verdict, verifyErr := repository.VerifyStoredRun(ctx, streamID, runSeq)
	if verifyErr != nil {
		_, _ = repository.db.ExecContext(ctx, `UPDATE verification_queue SET status='pending',claimed_at=NULL,available_at=now()+interval '5 seconds',last_error=$3 WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed'`, streamID, runSeq, verifyErr.Error())
		return true, verifyErr
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return true, err
	}
	defer tx.Rollback()
	if verdict == production.ReplayVerified {
		if err := projector.ProjectVerifiedRun(ctx, tx, streamID, runSeq); err != nil {
			return true, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE verification_queue SET status='verified',verdict='verified',completed_at=now(),claimed_at=NULL,last_error=NULL WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed'`, streamID, runSeq); err != nil {
			return true, err
		}
	} else {
		detail := "replay verification returned " + string(verdict)
		if _, err := tx.ExecContext(ctx, `INSERT INTO verification_dead_letters(company_stream_id,run_seq,verdict,detail) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, streamID, runSeq, verdict, detail); err != nil {
			return true, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE verification_queue SET status='dead',verdict=$3,completed_at=now(),claimed_at=NULL,last_error=$4 WHERE company_stream_id=$1 AND run_seq=$2 AND status='claimed'`, streamID, runSeq, verdict, detail); err != nil {
			return true, err
		}
	}
	return true, tx.Commit()
}

// VerifyStoredRun maps legacy NULL inputs to log_gap and derives engine drift
// from immutable database evidence rather than caller input.
func (repository *Repository) VerifyStoredRun(ctx context.Context, streamID string, runSeq int64) (production.ReplayVerdict, error) {
	if repository == nil || !uuidPattern.MatchString(streamID) || runSeq < 1 {
		return production.ReplayStateDivergence, errors.New("invalid run")
	}
	var hash, engine string
	var genesis []byte
	var version int
	if err := repository.db.QueryRowContext(ctx, `SELECT p.constants_hash,p.engine_version,g.state,g.version FROM run_epochs p JOIN run_genesis g USING(company_stream_id,run_seq) WHERE p.company_stream_id=$1 AND p.run_seq=$2`, streamID, runSeq).Scan(&hash, &engine, &genesis, &version); errors.Is(err, sql.ErrNoRows) {
		return production.ReplayLogGap, nil
	} else if err != nil {
		return production.ReplayStateDivergence, err
	}
	var drift bool
	if err := repository.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM run_version_drift WHERE company_stream_id=$1 AND run_seq=$2)`, streamID, runSeq).Scan(&drift); err != nil {
		return production.ReplayStateDivergence, err
	}
	artifacts, err := repository.artifacts(ctx, hash)
	if err != nil {
		return production.ReplayConstantsMismatch, nil
	}
	bundle, err := replaycatalog.Load(hash, artifacts)
	if err != nil {
		return production.ReplayConstantsMismatch, nil
	}
	rows, err := repository.db.QueryContext(ctx, `SELECT seq,canonical_payload,replay_inputs::text,receipt::text FROM run_log WHERE company_stream_id=$1 AND run_seq=$2 ORDER BY seq`, streamID, runSeq)
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
		if json.Unmarshal(entry.ReplayInputs, &envelope) != nil {
			return production.ReplayStateDivergence, nil
		}
		entry.Terminal = envelope.Resolved.Kind == "exit"
		entry.EventsJSON, err = repository.events(ctx, envelope.Command.IntentID)
		if err != nil {
			return production.ReplayStateDivergence, err
		}
		if entry.Terminal && envelope.Resolved.NextConstantsHash != "" && envelope.Resolved.NextConstantsHash != hash {
			nextArtifacts, loadErr := repository.artifacts(ctx, envelope.Resolved.NextConstantsHash)
			if loadErr != nil {
				return production.ReplayConstantsMismatch, nil
			}
			next, loadErr := replaycatalog.Load(envelope.Resolved.NextConstantsHash, nextArtifacts)
			if loadErr != nil {
				return production.ReplayConstantsMismatch, nil
			}
			bundle.Next = &next
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return production.ReplayStateDivergence, err
	}
	return production.VerifyReplayRun(genesis, version, bundle, entries, hash, drift || engine != kernel.Version), nil
}

func (repository *Repository) artifacts(ctx context.Context, hash string) (map[string][]byte, error) {
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
	if len(result) != 6 {
		return nil, fmt.Errorf("incomplete artifact bundle")
	}
	return result, rows.Err()
}

func (repository *Repository) events(ctx context.Context, intentID string) ([]byte, error) {
	rows, err := repository.db.QueryContext(ctx, `SELECT kind,schema_version,intent_id,payload::text FROM events WHERE intent_id=$1 ORDER BY event_seq`, intentID)
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
			return nil, fmt.Errorf("invalid event payload")
		}
		values = append(values, map[string]any{"kind": kind, "schema_version": schema, "intent_id": id, "payload": payload})
	}
	return json.Marshal(values)
}
