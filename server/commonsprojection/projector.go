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
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
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
	ResolveCommons(constantsHash string) (*commons.Catalog, bool)
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
		leftPriority, rightPriority := projectionPriority(events[i].Kind), projectionPriority(events[j].Kind)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return strings.Compare(events[i].EventID, events[j].EventID) < 0
	})
	for _, event := range events {
		switch event.Kind {
		case save.EventCompactSigned, save.EventCompactLeft:
			if err := p.project(ctx, event); err != nil {
				return err
			}
		case save.EventCompactSampled:
			if err := p.projectSample(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func projectionPriority(kind save.EventKind) int {
	switch kind {
	case save.EventCompactSigned:
		return 0
	case save.EventCompactSampled:
		return 1
	case save.EventCompactLeft:
		return 2
	default:
		return 3
	}
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
	catalog, ok := p.policies.ResolveCommons(event.ConstantsHash)
	if !ok || catalog.CohortTargetSize <= 0 {
		return fmt.Errorf("%w: catalog", ErrProjection)
	}
	target := catalog.CohortTargetSize
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

type samplePayload struct {
	FounderID string `json:"founder_id"`
	RunID     struct {
		CompanyStreamID string `json:"company_stream_id"`
		RunSeq          int64  `json:"run_seq"`
	} `json:"run_id"`
	WeightPPM     int64  `json:"weight_ppm"`
	CompliancePPM int64  `json:"compliance_ppm"`
	Enclosure     string `json:"enclosure"`
	Capacity      string `json:"capacity"`
	SolidarityPPM int64  `json:"solidarity_ppm"`
	SampledMS     int64  `json:"sampled_ms"`
}

func (p *Projector) projectSample(ctx context.Context, event save.EventRecord) error {
	var payload samplePayload
	if err := decodeStrict(event.Payload, &payload); err != nil || payload.FounderID != event.OwnerID || payload.RunID.CompanyStreamID != event.StreamID || payload.RunID.RunSeq < 1 || payload.WeightPPM < 0 || payload.WeightPPM > commons.PPM || payload.CompliancePPM < 0 || payload.CompliancePPM > commons.PPM || payload.SolidarityPPM < 0 || payload.SolidarityPPM > commons.PPM || payload.SampledMS <= 0 {
		return fmt.Errorf("%w: sample identity", ErrProjection)
	}
	if value, err := decimal.ParseCanonical(payload.Enclosure); err != nil || value.Lt(decimal.Zero) || value.Gt(decimal.One) {
		return ErrProjection
	}
	if value, err := decimal.ParseCanonical(payload.Capacity); err != nil || value.Lt(decimal.Zero) {
		return ErrProjection
	}
	catalog, ok := p.policies.ResolveCommons(event.ConstantsHash)
	if !ok {
		return ErrProjection
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var cohortID, serverID string
	var member bool
	if err := tx.QueryRowContext(ctx, `SELECT m.cohort_id,a.server_id,m.member FROM company_compact_memberships m JOIN founder_commons_assignments a ON a.founder_id=m.founder_id WHERE m.company_stream_id=$1 AND m.founder_id=$2 AND m.run_seq=$3`, payload.RunID.CompanyStreamID, payload.FounderID, payload.RunID.RunSeq).Scan(&cohortID, &serverID, &member); err != nil || !member {
		return fmt.Errorf("%w: sample without active membership", ErrProjection)
	}
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
	// Capacity is the absolute sum of every tithe, while the other fields are
	// latest-state samples. Serialize per company so concurrent post-commit
	// projectors cannot lose one of those additive contributions.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, payload.RunID.CompanyStreamID); err != nil {
		return err
	}
	capacity := decimal.Zero
	var priorCapacity string
	err = tx.QueryRowContext(ctx, `SELECT capacity FROM commons_member_samples WHERE company_stream_id=$1 FOR UPDATE`, payload.RunID.CompanyStreamID).Scan(&priorCapacity)
	if err == nil {
		capacity, err = decimal.ParseCanonical(priorCapacity)
		if err != nil {
			return ErrProjection
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	increment, err := decimal.ParseCanonical(payload.Capacity)
	if err != nil {
		return ErrProjection
	}
	capacity = capacity.Add(increment).Quantize(decimal.CanonicalSignificantDigits)
	if !capacity.IsStateValue() || capacity.Lt(decimal.Zero) {
		return ErrProjection
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO commons_member_samples(company_stream_id,founder_id,cohort_id,weight_ppm,compliance_ppm,solidarity_ppm,enclosure,capacity,sampled_ms,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(company_stream_id) DO UPDATE SET weight_ppm=EXCLUDED.weight_ppm,compliance_ppm=EXCLUDED.compliance_ppm,solidarity_ppm=EXCLUDED.solidarity_ppm,enclosure=EXCLUDED.enclosure,capacity=EXCLUDED.capacity,sampled_ms=EXCLUDED.sampled_ms,updated_at=EXCLUDED.updated_at`, payload.RunID.CompanyStreamID, payload.FounderID, cohortID, payload.WeightPPM, payload.CompliancePPM, payload.SolidarityPPM, payload.Enclosure, capacity.String(), payload.SampledMS, event.OccurredAt); err != nil {
		return err
	}
	var previousServerHealth int64
	hadPreviousServer := tx.QueryRowContext(ctx, `SELECT health_ppm FROM commons_health_scopes WHERE scope_kind='server' AND scope_id=$1`, serverID).Scan(&previousServerHealth) == nil
	if err := p.refreshScope(ctx, tx, catalog, "cohort", cohortID, event.OccurredAt); err != nil {
		return err
	}
	if err := p.refreshScope(ctx, tx, catalog, "server", serverID, event.OccurredAt); err != nil {
		return err
	}
	if hadPreviousServer {
		var current int64
		if err := tx.QueryRowContext(ctx, `SELECT health_ppm FROM commons_health_scopes WHERE scope_kind='server' AND scope_id=$1`, serverID).Scan(&current); err != nil {
			return err
		}
		if err := emitHealthTransitions(ctx, tx, catalog, event, serverID, previousServerHealth, current); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func healthBand(catalog *commons.Catalog, healthPPM int64) string {
	if healthPPM < catalog.CollapseHealthPPM {
		return "collapsed"
	}
	if healthPPM >= catalog.HealthyHealthPPM {
		return "healthy"
	}
	return "strained"
}

func emitHealthTransitions(ctx context.Context, tx *sql.Tx, catalog *commons.Catalog, event save.EventRecord, serverID string, previous, current int64) error {
	from, to := healthBand(catalog, previous), healthBand(catalog, current)
	if from == to {
		return nil
	}
	bandPayload, _ := json.Marshal(map[string]any{"scope_kind": "server", "scope_id": serverID, "from_band": from, "to_band": to, "health_ppm": current})
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,'compact_health_band_changed',$3,$4,$5,$6)`, event.StreamID, event.Revision, nullIntent(event.IntentID), event.ConstantsHash, event.OccurredAt, bandPayload); err != nil {
		return err
	}
	var kind save.EventKind
	if to == "collapsed" {
		kind = save.EventCompactCascadeStarted
	} else if from == "collapsed" {
		kind = save.EventCompactRecovered
	}
	if kind == "" {
		return nil
	}
	payload, _ := json.Marshal(map[string]any{"scope_kind": "server", "scope_id": serverID, "health_ppm": current})
	_, err := tx.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,$3,$4,$5,$6,$7)`, event.StreamID, event.Revision, kind, nullIntent(event.IntentID), event.ConstantsHash, event.OccurredAt, payload)
	return err
}

func (p *Projector) refreshScope(ctx context.Context, tx *sql.Tx, catalog *commons.Catalog, scopeKind, scopeID string, now time.Time) error {
	query := `SELECT s.weight_ppm,s.compliance_ppm,s.capacity FROM commons_member_samples s JOIN company_compact_memberships m ON m.company_stream_id=s.company_stream_id JOIN founder_commons_assignments a ON a.founder_id=s.founder_id WHERE m.member=true AND a.cohort_id=$1`
	if scopeKind == "server" {
		query = `SELECT s.weight_ppm,s.compliance_ppm,s.capacity FROM commons_member_samples s JOIN company_compact_memberships m ON m.company_stream_id=s.company_stream_id JOIN founder_commons_assignments a ON a.founder_id=s.founder_id WHERE m.member=true AND a.server_id=$1`
	}
	rows, err := tx.QueryContext(ctx, query, scopeID)
	if err != nil {
		return err
	}
	var samples []commons.MemberSample
	for rows.Next() {
		var sample commons.MemberSample
		var raw string
		if err := rows.Scan(&sample.WeightPPM, &sample.CompliancePPM, &raw); err != nil {
			rows.Close()
			return err
		}
		sample.Capacity, err = decimal.ParseCanonical(raw)
		if err != nil {
			rows.Close()
			return err
		}
		samples = append(samples, sample)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	aggregate, err := commons.AggregateHealth(catalog, samples)
	if err != nil {
		return err
	}
	health := aggregate.HealthPPM
	var previous int64
	var evaluated time.Time
	err = tx.QueryRowContext(ctx, `SELECT health_ppm,evaluated_at FROM commons_health_scopes WHERE scope_kind=$1 AND scope_id=$2 FOR UPDATE`, scopeKind, scopeID).Scan(&previous, &evaluated)
	if err == nil {
		elapsed := now.Sub(evaluated).Milliseconds()
		if elapsed < 0 {
			elapsed = 0
		}
		health, err = commons.SmoothHealthPPM(catalog, previous, aggregate.HealthPPM, elapsed)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO commons_health_scopes(scope_kind,scope_id,raw_health_ppm,health_ppm,capacity,real_members,npc_weight_ppm,evaluated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(scope_kind,scope_id) DO UPDATE SET raw_health_ppm=EXCLUDED.raw_health_ppm,health_ppm=EXCLUDED.health_ppm,capacity=EXCLUDED.capacity,real_members=EXCLUDED.real_members,npc_weight_ppm=EXCLUDED.npc_weight_ppm,evaluated_at=EXCLUDED.evaluated_at`, scopeKind, scopeID, aggregate.HealthPPM, health, aggregate.Capacity.String(), aggregate.RealMembers, aggregate.NPCWeightPPM, now)
	return err
}

type HealthSnapshot struct {
	HealthPPM       int64
	CohortHealthPPM int64
	ServerHealthPPM int64
	CohortCapacity  string
	ServerCapacity  string
	NPCWeightPPM    int64
}

func (p *Projector) Snapshot(ctx context.Context, founderID string) (HealthSnapshot, error) {
	var cohortID, serverID string
	if err := p.db.QueryRowContext(ctx, `SELECT cohort_id,server_id FROM founder_commons_assignments WHERE founder_id=$1`, founderID).Scan(&cohortID, &serverID); err != nil {
		return HealthSnapshot{}, err
	}
	var result HealthSnapshot
	if err := p.db.QueryRowContext(ctx, `SELECT health_ppm,capacity,npc_weight_ppm FROM commons_health_scopes WHERE scope_kind='cohort' AND scope_id=$1`, cohortID).Scan(&result.CohortHealthPPM, &result.CohortCapacity, &result.NPCWeightPPM); err != nil {
		return HealthSnapshot{}, err
	}
	if err := p.db.QueryRowContext(ctx, `SELECT health_ppm,capacity FROM commons_health_scopes WHERE scope_kind='server' AND scope_id=$1`, serverID).Scan(&result.ServerHealthPPM, &result.ServerCapacity); err != nil {
		return HealthSnapshot{}, err
	}
	result.HealthPPM = (result.CohortHealthPPM*800_000 + result.ServerHealthPPM*200_000) / commons.PPM
	return result, nil
}

func (p *Projector) CompactSnapshot(ctx context.Context, founderID string) (commons.ContributionSnapshot, error) {
	snapshot, err := p.Snapshot(ctx, founderID)
	return commons.ContributionSnapshot{HealthPPM: snapshot.HealthPPM}, err
}

// MergeCollapsed merges every additional under-floor cohort into the oldest
// compatible cohort. It never averages rounded Health values: refreshScope
// recomputes numerator/denominator inputs from member samples after the move.
func (p *Projector) MergeCollapsed(ctx context.Context, constantsHash, serverID, activityBracket string, now time.Time) (int64, error) {
	catalog, ok := p.policies.ResolveCommons(constantsHash)
	if !ok {
		return 0, ErrProjection
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, serverID+":"+activityBracket); err != nil {
		return 0, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT cohort_id,member_count FROM commons_cohorts WHERE server_id=$1 AND activity_bracket=$2 AND closed_at IS NULL ORDER BY created_at,cohort_id FOR UPDATE`, serverID, activityBracket)
	if err != nil {
		return 0, err
	}
	type cohort struct {
		id      string
		members int
	}
	var cohorts []cohort
	for rows.Next() {
		var item cohort
		if err := rows.Scan(&item.id, &item.members); err != nil {
			rows.Close()
			return 0, err
		}
		cohorts = append(cohorts, item)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	if len(cohorts) < 2 {
		if err := tx.Commit(); err != nil {
			return 0, err
		}
		return 0, nil
	}
	target := cohorts[0]
	var merged int64
	for _, source := range cohorts[1:] {
		if source.members >= catalog.CohortMergeFloor {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE founder_commons_assignments SET cohort_id=$1 WHERE cohort_id=$2`, target.id, source.id); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE company_compact_memberships SET cohort_id=$1 WHERE cohort_id=$2`, target.id, source.id); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE commons_member_samples SET cohort_id=$1 WHERE cohort_id=$2`, target.id, source.id); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=member_count+$2 WHERE cohort_id=$1`, target.id, source.members); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=0,closed_at=$2 WHERE cohort_id=$1`, source.id, now); err != nil {
			return 0, err
		}
		target.members += source.members
		merged++
	}
	if merged > 0 {
		if err := p.refreshScope(ctx, tx, catalog, "cohort", target.id, now); err != nil {
			return 0, err
		}
		if err := p.refreshScope(ctx, tx, catalog, "server", serverID, now); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return merged, nil
}

// OfferRecruitment is called by the authoritative progress boundary at mid-T3.
// The unique founder row makes the ambient offer once-per-career and replay-safe.
func (p *Projector) OfferRecruitment(ctx context.Context, streamID, founderID string, revision int64, constantsHash string, occurredAt time.Time) (bool, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var ownerID string
	var scope string
	if err := tx.QueryRowContext(ctx, `SELECT owner_id,scope FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&ownerID, &scope); err != nil || ownerID != founderID || scope != "company" {
		return false, ErrProjection
	}
	var already bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM founder_commons_assignments WHERE founder_id=$1) OR EXISTS(SELECT 1 FROM commons_recruitment_offers WHERE founder_id=$1)`, founderID).Scan(&already); err != nil {
		return false, err
	}
	if already {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}
	payload, _ := json.Marshal(map[string]any{"founder_id": founderID, "reason_key": "compact.recruitment.mid_t3"})
	var eventID string
	if err := tx.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,'compact_recruitment_offered',NULL,$3,$4,$5) RETURNING event_id`, streamID, revision, constantsHash, occurredAt, payload).Scan(&eventID); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO commons_recruitment_offers(founder_id,event_id,offered_at) VALUES($1,$2,$3)`, founderID, eventID, occurredAt); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func nullIntent(value string) any {
	if value == "" {
		return nil
	}
	return value
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
