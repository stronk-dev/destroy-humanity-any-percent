// Package routeprojection maintains idempotent founder and public Registry
// projections from authoritative company-stream route events.
package routeprojection

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

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

var ErrProjection = errors.New("route projection failed")

type CatalogResolver interface {
	ResolveRoutes(constantsHash string) (*routes.Catalog, bool)
}

type Projector struct {
	db       *sql.DB
	catalogs CatalogResolver
}

func New(db *sql.DB, catalogs CatalogResolver) (*Projector, error) {
	if db == nil || catalogs == nil {
		return nil, ErrProjection
	}
	return &Projector{db: db, catalogs: catalogs}, nil
}

type executionPayload struct {
	RouteID string `json:"route_id"`
	GateID  string `json:"gate_id"`
	RunID   struct {
		CompanyStreamID string `json:"company_stream_id"`
		RunSeq          int64  `json:"run_seq"`
	} `json:"run_id"`
	FounderID string `json:"founder_id"`
}
type hintPayload struct {
	RouteID string `json:"route_id"`
	Cost    int64  `json:"cost"`
}

func (p *Projector) Project(ctx context.Context, source []save.EventRecord) error {
	events := append([]save.EventRecord(nil), source...)
	sort.Slice(events, func(i, j int) bool {
		if !events[i].OccurredAt.Equal(events[j].OccurredAt) {
			return events[i].OccurredAt.Before(events[j].OccurredAt)
		}
		return strings.Compare(events[i].EventID, events[j].EventID) < 0
	})
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, event := range events {
		switch event.Kind {
		case save.EventRouteExecuted:
			if err := p.projectExecution(ctx, tx, event); err != nil {
				return err
			}
		case save.EventRouteHintPurchased:
			if err := p.projectHint(ctx, tx, event); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (p *Projector) projectExecution(ctx context.Context, tx *sql.Tx, event save.EventRecord) error {
	var payload executionPayload
	if err := decodeStrict(event.Payload, &payload); err != nil {
		return fmt.Errorf("%w: route event: %v", ErrProjection, err)
	}
	catalog, ok := p.catalogs.ResolveRoutes(event.ConstantsHash)
	if !ok {
		return fmt.Errorf("%w: unknown catalog %s", ErrProjection, event.ConstantsHash)
	}
	route, ok := catalog.Route(payload.RouteID)
	if !ok || payload.FounderID != event.OwnerID || payload.RunID.CompanyStreamID != event.StreamID || payload.RunID.RunSeq < 1 {
		return fmt.Errorf("%w: route event identity mismatch", ErrProjection)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO route_projection_events(event_id,route_id,founder_id,company_stream_id,run_seq,occurred_at) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, event.EventID, payload.RouteID, payload.FounderID, payload.RunID.CompanyStreamID, payload.RunID.RunSeq, event.OccurredAt)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil || inserted == 0 {
		return err
	}

	result, err = tx.ExecContext(ctx, `INSERT INTO registry_routes(route_id,first_event_id,first_founder_id,occurred_at,house_name,name,name_state,naming_reserved_until,execution_count) VALUES ($1,$2,$3,$4,$5,$5,'reserved',$6,1) ON CONFLICT DO NOTHING`, payload.RouteID, event.EventID, payload.FounderID, event.OccurredAt, route.HouseName, event.OccurredAt.Add(72*time.Hour))
	if err != nil {
		return err
	}
	registryFirst, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if registryFirst == 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE registry_routes SET execution_count=execution_count+1 WHERE route_id=$1`, payload.RouteID); err != nil {
			return err
		}
	}

	result, err = tx.ExecContext(ctx, `INSERT INTO founder_route_executions(founder_id,route_id,first_event_id,first_occurred_at,last_occurred_at,execution_count) VALUES ($1,$2,$3,$4,$4,1) ON CONFLICT DO NOTHING`, payload.FounderID, payload.RouteID, event.EventID, event.OccurredAt)
	if err != nil {
		return err
	}
	founderFirst, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if founderFirst == 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE founder_route_executions SET last_occurred_at=$3,execution_count=execution_count+1 WHERE founder_id=$1 AND route_id=$2`, payload.FounderID, payload.RouteID, event.OccurredAt); err != nil {
			return err
		}
	}

	policy := catalog.KnowledgePolicy()
	if registryFirst == 1 {
		if err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "registry_first", policy.RegistryFirstBonus); err != nil {
			return err
		}
	}
	if founderFirst == 1 {
		if err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "founder_first", policy.FounderFirstGrant); err != nil {
			return err
		}
	} else if err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "repeat", policy.RepeatGrant); err != nil {
		return err
	}
	return nil
}

func (p *Projector) grant(ctx context.Context, tx *sql.Tx, event save.EventRecord, routeID, founderID, source string, amount int64) error {
	if amount <= 0 || amount > decimal.MaxExactInteger {
		return ErrProjection
	}
	payload, _ := json.Marshal(map[string]any{"route_id": routeID, "amount": amount, "source": source})
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,'route_knowledge_granted',$3,$4,$5,$6)`, event.StreamID, event.Revision, nullIntent(event.IntentID), event.ConstantsHash, event.OccurredAt, payload); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance) VALUES($1,$2) ON CONFLICT(founder_id) DO UPDATE SET route_knowledge_balance=founder_route_state.route_knowledge_balance+EXCLUDED.route_knowledge_balance WHERE founder_route_state.route_knowledge_balance <= $3`, founderID, amount, decimal.MaxExactInteger-amount)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return fmt.Errorf("%w: Route Knowledge overflow", ErrProjection)
	}
	return nil
}

func (p *Projector) projectHint(ctx context.Context, tx *sql.Tx, event save.EventRecord) error {
	var payload hintPayload
	if err := decodeStrict(event.Payload, &payload); err != nil || payload.Cost <= 0 {
		return fmt.Errorf("%w: hint event", ErrProjection)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO route_hint_projection_events(event_id,founder_id,route_id,cost) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, event.EventID, event.OwnerID, payload.RouteID, payload.Cost)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil || inserted == 0 {
		return err
	}
	result, err = tx.ExecContext(ctx, `UPDATE founder_route_state SET route_knowledge_balance=route_knowledge_balance-$2 WHERE founder_id=$1 AND route_knowledge_balance >= $2`, event.OwnerID, payload.Cost)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return fmt.Errorf("%w: projected hint has insufficient balance", ErrProjection)
	}
	return nil
}

func (p *Projector) RepairFounder(ctx context.Context, founderID string, state *save.State) error {
	if state == nil {
		return ErrProjection
	}
	var balance int64
	err := p.db.QueryRowContext(ctx, `SELECT route_knowledge_balance FROM founder_route_state WHERE founder_id=$1`, founderID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		var grants, costs int64
		if err := p.db.QueryRowContext(ctx, `SELECT COALESCE(sum(CASE WHEN e.kind='route_knowledge_granted' THEN (e.payload->>'amount')::bigint ELSE 0 END),0), COALESCE(sum(CASE WHEN e.kind='route_hint_purchased' THEN (e.payload->>'cost')::bigint ELSE 0 END),0) FROM events e JOIN save_streams s ON s.id=e.stream_id WHERE s.owner_id=$1 AND e.kind IN ('route_knowledge_granted','route_hint_purchased')`, founderID).Scan(&grants, &costs); err != nil {
			return err
		}
		balance = grants - costs
		if balance < 0 || balance > decimal.MaxExactInteger {
			return ErrProjection
		}
		if _, err := p.db.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance) VALUES($1,$2) ON CONFLICT(founder_id) DO UPDATE SET route_knowledge_balance=EXCLUDED.route_knowledge_balance`, founderID, balance); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	state.RouteKnowledgeBalance = balance
	return nil
}

func (p *Projector) FounderDistinctRoutes(ctx context.Context, founderID string) (int64, error) {
	var count int64
	if err := p.db.QueryRowContext(ctx, `SELECT count(*) FROM founder_route_executions WHERE founder_id=$1`, founderID).Scan(&count); err != nil {
		return 0, err
	}
	if count > 0 {
		return count, nil
	}
	if err := p.db.QueryRowContext(ctx, `SELECT count(DISTINCT payload->>'route_id') FROM events WHERE kind='route_executed' AND payload->>'founder_id'=$1`, founderID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (p *Projector) SubmitName(ctx context.Context, routeID, founderID, name string, now time.Time) error {
	if strings.TrimSpace(name) == "" || len([]rune(name)) > 80 {
		return ErrProjection
	}
	result, err := p.db.ExecContext(ctx, `UPDATE registry_routes SET name=$3,name_state='pending' WHERE route_id=$1 AND first_founder_id=$2 AND name_state='reserved' AND naming_reserved_until>$4`, routeID, founderID, name, now.UTC())
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ErrProjection
	}
	return nil
}

func (p *Projector) ResolveName(ctx context.Context, routeID string, approved bool) error {
	state, reset := "published", false
	if !approved {
		state, reset = "house", true
	}
	result, err := p.db.ExecContext(ctx, `UPDATE registry_routes SET name_state=$2,name=CASE WHEN $3 THEN house_name ELSE name END WHERE route_id=$1 AND name_state='pending'`, routeID, state, reset)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ErrProjection
	}
	return nil
}

func (p *Projector) ExpireNames(ctx context.Context, now time.Time) (int64, error) {
	result, err := p.db.ExecContext(ctx, `UPDATE registry_routes SET name_state='house' WHERE name_state='reserved' AND naming_reserved_until <= $1`, now.UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func decodeStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}
func nullIntent(value string) any {
	if value == "" {
		return nil
	}
	return value
}
