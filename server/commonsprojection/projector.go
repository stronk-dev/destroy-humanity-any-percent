// Package commonsprojection maintains idempotent membership and cohort read
// models from authoritative company-stream Compact events.
package commonsprojection

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"cloud-clicker/server/save"
)

var ErrProjection = errors.New("commons projection failed")

type AssignmentContext struct {
	ServerID        string
	ActivityBracket string
}
type AssignmentResolver interface {
	ResolveAssignment(founderID string) (AssignmentContext, bool)
}
type PolicyResolver interface {
	CompactCohortTarget(constantsHash string) (int, bool)
}

type Projector struct {
	db          *sql.DB
	assignments AssignmentResolver
	policies    PolicyResolver
}

func New(db *sql.DB, assignments AssignmentResolver, policies PolicyResolver) (*Projector, error) {
	if db == nil || assignments == nil || policies == nil {
		return nil, ErrProjection
	}
	return &Projector{db: db, assignments: assignments, policies: policies}, nil
}

type membershipPayload struct {
	FounderID string `json:"founder_id"`
	RunID     struct {
		CompanyStreamID string `json:"company_stream_id"`
		RunSeq          int64  `json:"run_seq"`
	} `json:"run_id"`
	TithePPM    int64 `json:"tithe_ppm"`
	PriorMember bool  `json:"prior_member"`
	NewMember   bool  `json:"new_member"`
}

func (p *Projector) Project(ctx context.Context, source []save.EventRecord) error {
	events := append([]save.EventRecord(nil), source...)
	sort.Slice(events, func(i, j int) bool {
		if !events[i].OccurredAt.Equal(events[j].OccurredAt) {
			return events[i].OccurredAt.Before(events[j].OccurredAt)
		}
		return strings.Compare(events[i].EventID, events[j].EventID) < 0
	})
	for _, event := range events {
		if event.Kind != save.EventCompactSigned && event.Kind != save.EventCompactLeft {
			continue
		}
		if err := p.project(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (p *Projector) project(ctx context.Context, event save.EventRecord) error {
	var payload membershipPayload
	if err := decodeStrict(event.Payload, &payload); err != nil || payload.FounderID != event.OwnerID || payload.RunID.CompanyStreamID != event.StreamID || payload.RunID.RunSeq < 1 || payload.TithePPM < 0 || payload.TithePPM > 1_000_000 {
		return fmt.Errorf("%w: membership event identity", ErrProjection)
	}
	assignment, ok := p.assignments.ResolveAssignment(payload.FounderID)
	if !ok || assignment.ServerID == "" || assignment.ActivityBracket == "" {
		return fmt.Errorf("%w: assignment context", ErrProjection)
	}
	target, ok := p.policies.CompactCohortTarget(event.ConstantsHash)
	if !ok || target <= 0 {
		return fmt.Errorf("%w: catalog", ErrProjection)
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO commons_projection_events(event_id,kind,founder_id,company_stream_id,run_seq,occurred_at) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, event.EventID, event.Kind, payload.FounderID, payload.RunID.CompanyStreamID, payload.RunID.RunSeq, event.OccurredAt)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil || inserted == 0 {
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, assignment.ServerID+":"+assignment.ActivityBracket); err != nil {
		return err
	}
	var cohortID string
	err = tx.QueryRowContext(ctx, `SELECT cohort_id FROM founder_commons_assignments WHERE founder_id=$1`, payload.FounderID).Scan(&cohortID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT cohort_id FROM commons_cohorts WHERE server_id=$1 AND activity_bracket=$2 AND closed_at IS NULL AND member_count<$3 ORDER BY created_at,cohort_id LIMIT 1 FOR UPDATE`, assignment.ServerID, assignment.ActivityBracket, target).Scan(&cohortID)
		if errors.Is(err, sql.ErrNoRows) {
			if err := tx.QueryRowContext(ctx, `INSERT INTO commons_cohorts(server_id,activity_bracket,cohort_seq,member_count) VALUES($1,$2,COALESCE((SELECT max(cohort_seq)+1 FROM commons_cohorts WHERE server_id=$1 AND activity_bracket=$2),1),0) RETURNING cohort_id`, assignment.ServerID, assignment.ActivityBracket).Scan(&cohortID); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO founder_commons_assignments(founder_id,server_id,activity_bracket,cohort_id,first_signed_at,last_signed_at) VALUES($1,$2,$3,$4,$5,$5)`, payload.FounderID, assignment.ServerID, assignment.ActivityBracket, cohortID, event.OccurredAt); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=member_count+1 WHERE cohort_id=$1`, cohortID); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if _, err := tx.ExecContext(ctx, `UPDATE founder_commons_assignments SET last_signed_at=$2 WHERE founder_id=$1`, payload.FounderID, event.OccurredAt); err != nil {
		return err
	}
	if event.Kind == save.EventCompactSigned {
		if payload.PriorMember || !payload.NewMember {
			return ErrProjection
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO company_compact_memberships(company_stream_id,founder_id,run_seq,cohort_id,member,tithe_ppm,updated_at) VALUES($1,$2,$3,$4,true,$5,$6) ON CONFLICT(company_stream_id) DO UPDATE SET founder_id=EXCLUDED.founder_id,run_seq=EXCLUDED.run_seq,cohort_id=EXCLUDED.cohort_id,member=true,tithe_ppm=EXCLUDED.tithe_ppm,updated_at=EXCLUDED.updated_at`, payload.RunID.CompanyStreamID, payload.FounderID, payload.RunID.RunSeq, cohortID, payload.TithePPM, event.OccurredAt)
	} else {
		if !payload.PriorMember || payload.NewMember {
			return ErrProjection
		}
		result, err = tx.ExecContext(ctx, `UPDATE company_compact_memberships SET member=false,tithe_ppm=0,updated_at=$2 WHERE company_stream_id=$1 AND founder_id=$3`, payload.RunID.CompanyStreamID, event.OccurredAt, payload.FounderID)
		if err == nil {
			var affected int64
			affected, err = result.RowsAffected()
			if err == nil && affected != 1 {
				err = ErrProjection
			}
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (p *Projector) FounderCohort(ctx context.Context, founderID string) (string, error) {
	var id string
	err := p.db.QueryRowContext(ctx, `SELECT cohort_id FROM founder_commons_assignments WHERE founder_id=$1`, founderID).Scan(&id)
	return id, err
}

func decodeStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON")
	}
	return nil
}
