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
	"math/big"
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

type registryDecision struct {
	registryFirst    bool
	displacedEvent   string
	displacedFounder string
}

type grantRecord struct {
	EventID       string
	StreamID      string
	Revision      int64
	ConstantsHash string
	OccurredAt    time.Time
	Amount        int64
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
	if err := lockRegistryRoutes(ctx, tx, events); err != nil {
		return err
	}
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

	registry, err := p.decideRegistry(ctx, tx, event, payload, route.HouseName)
	if err != nil {
		return err
	}
	if registry.displacedEvent != "" {
		if err := p.compensateRegistryGrant(ctx, tx, registry.displacedEvent, registry.displacedFounder); err != nil {
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
	if registry.registryFirst {
		if _, err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "registry_first", policy.RegistryFirstBonus); err != nil {
			return err
		}
	}
	if founderFirst == 1 {
		if _, err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "founder_first", policy.FounderFirstGrant); err != nil {
			return err
		}
	} else if _, err := p.grant(ctx, tx, event, payload.RouteID, payload.FounderID, "repeat", policy.RepeatGrant); err != nil {
		return err
	}
	return nil
}

func lockRegistryRoutes(ctx context.Context, tx *sql.Tx, events []save.EventRecord) error {
	routeIDs := make(map[string]struct{})
	for _, event := range events {
		if event.Kind != save.EventRouteExecuted {
			continue
		}
		var payload executionPayload
		if err := decodeStrict(event.Payload, &payload); err != nil {
			return fmt.Errorf("%w: route event: %v", ErrProjection, err)
		}
		routeIDs[payload.RouteID] = struct{}{}
	}
	ordered := make([]string, 0, len(routeIDs))
	for routeID := range routeIDs {
		ordered = append(ordered, routeID)
	}
	sort.Strings(ordered)
	for _, routeID := range ordered {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 1381191751))`, routeID); err != nil {
			return err
		}
	}
	return nil
}

func (p *Projector) decideRegistry(ctx context.Context, tx *sql.Tx, event save.EventRecord, payload executionPayload, houseName string) (registryDecision, error) {
	var currentEvent, currentFounder string
	var currentOccurred time.Time
	err := tx.QueryRowContext(ctx, `SELECT first_event_id,first_founder_id,occurred_at FROM registry_routes WHERE route_id=$1 FOR UPDATE`, payload.RouteID).Scan(&currentEvent, &currentFounder, &currentOccurred)
	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `INSERT INTO registry_routes(route_id,first_event_id,first_founder_id,occurred_at,house_name,name,name_state,naming_reserved_until,execution_count) VALUES ($1,$2,$3,$4,$5,$5,'reserved',$6,1)`, payload.RouteID, event.EventID, payload.FounderID, event.OccurredAt, houseName, event.OccurredAt.Add(72*time.Hour))
		return registryDecision{registryFirst: true}, err
	}
	if err != nil {
		return registryDecision{}, err
	}
	if !eventBefore(event.OccurredAt, event.EventID, currentOccurred, currentEvent) {
		_, err = tx.ExecContext(ctx, `UPDATE registry_routes SET execution_count=execution_count+1 WHERE route_id=$1`, payload.RouteID)
		return registryDecision{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE registry_routes SET first_event_id=$2,first_founder_id=$3,occurred_at=$4,house_name=$5,name=$5,name_state='reserved',naming_reserved_until=$6,execution_count=execution_count+1 WHERE route_id=$1`, payload.RouteID, event.EventID, payload.FounderID, event.OccurredAt, houseName, event.OccurredAt.Add(72*time.Hour))
	if err != nil {
		return registryDecision{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return registryDecision{}, fmt.Errorf("%w: Registry displacement", ErrProjection)
	}
	return registryDecision{registryFirst: true, displacedEvent: currentEvent, displacedFounder: currentFounder}, nil
}

func eventBefore(leftTime time.Time, leftID string, rightTime time.Time, rightID string) bool {
	if !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}
	return strings.Compare(leftID, rightID) < 0
}

func (p *Projector) compensateRegistryGrant(ctx context.Context, tx *sql.Tx, executionEventID, founderID string) error {
	var grant grantRecord
	var matches int64
	err := tx.QueryRowContext(ctx, `
		SELECT awarded.event_id,awarded.stream_id,awarded.revision,awarded.constants_hash,awarded.occurred_at,
		       (awarded.payload->>'amount')::bigint,count(*) OVER ()
		FROM events execution
		JOIN events awarded ON awarded.stream_id=execution.stream_id AND awarded.revision=execution.revision
		JOIN save_streams stream ON stream.id=awarded.stream_id
		WHERE execution.event_id=$1 AND stream.owner_id=$2 AND awarded.kind='route_knowledge_granted'
		  AND awarded.payload->>'route_id'=(execution.payload->>'route_id')
		  AND awarded.payload->>'source'='registry_first'
		  AND NOT EXISTS (
		      SELECT 1 FROM events compensation
		      WHERE compensation.kind='compensation'
		        AND compensation.payload->>'compensates_event_id'=awarded.event_id::text
		  )`, executionEventID, founderID).Scan(&grant.EventID, &grant.StreamID, &grant.Revision, &grant.ConstantsHash, &grant.OccurredAt, &grant.Amount, &matches)
	if err != nil || matches != 1 || grant.Amount <= 0 || grant.Amount > decimal.MaxExactInteger {
		if err == nil {
			err = ErrProjection
		}
		return fmt.Errorf("%w: active Registry grant: %v", ErrProjection, err)
	}
	payload, err := json.Marshal(map[string]any{
		"compensates_event_id": grant.EventID,
		"reason_key":           "route.registry_first_reordered",
	})
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,'compensation',NULL,$3,$4,$5)`, grant.StreamID, grant.Revision, grant.ConstantsHash, grant.OccurredAt, payload); err != nil {
		return err
	}
	return p.applyKnowledgeDelta(ctx, tx, founderID, -grant.Amount)
}

func (p *Projector) grant(ctx context.Context, tx *sql.Tx, event save.EventRecord, routeID, founderID, source string, amount int64) (string, error) {
	if amount <= 0 || amount > decimal.MaxExactInteger {
		return "", ErrProjection
	}
	payload, _ := json.Marshal(map[string]any{"route_id": routeID, "amount": amount, "source": source})
	var eventID string
	if err := tx.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,1,'route_knowledge_granted',$3,$4,$5,$6) RETURNING event_id`, event.StreamID, event.Revision, nullIntent(event.IntentID), event.ConstantsHash, event.OccurredAt, payload).Scan(&eventID); err != nil {
		return "", err
	}
	if err := p.applyKnowledgeDelta(ctx, tx, founderID, amount); err != nil {
		return "", err
	}
	return eventID, nil
}

func (p *Projector) applyKnowledgeDelta(ctx context.Context, tx *sql.Tx, founderID string, delta int64) error {
	if delta == 0 || delta < -decimal.MaxExactInteger || delta > decimal.MaxExactInteger {
		return ErrProjection
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance,route_knowledge_debt) VALUES($1,0,0) ON CONFLICT DO NOTHING`, founderID); err != nil {
		return err
	}
	var balance, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT route_knowledge_balance,route_knowledge_debt FROM founder_route_state WHERE founder_id=$1 FOR UPDATE`, founderID).Scan(&balance, &debt); err != nil {
		return err
	}
	if delta > 0 {
		repaid := min(debt, delta)
		debt -= repaid
		spendable := delta - repaid
		if balance > decimal.MaxExactInteger-spendable {
			return fmt.Errorf("%w: Route Knowledge overflow", ErrProjection)
		}
		balance += spendable
	} else {
		correction := -delta
		consumed := min(balance, correction)
		balance -= consumed
		correction -= consumed
		if debt > decimal.MaxExactInteger-correction {
			return fmt.Errorf("%w: Route Knowledge debt overflow", ErrProjection)
		}
		debt += correction
	}
	result, err := tx.ExecContext(ctx, `UPDATE founder_route_state SET route_knowledge_balance=$2,route_knowledge_debt=$3 WHERE founder_id=$1`, founderID, balance, debt)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return ErrProjection
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
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance,route_knowledge_debt) VALUES($1,0,0) ON CONFLICT DO NOTHING`, founderID)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	var balance, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT route_knowledge_balance,route_knowledge_debt FROM founder_route_state WHERE founder_id=$1 FOR UPDATE`, founderID).Scan(&balance, &debt); err != nil {
		return err
	}
	if inserted == 1 {
		var netText string
		if err := tx.QueryRowContext(ctx, `
			SELECT (
				COALESCE(sum((e.payload->>'amount')::numeric) FILTER (
					WHERE e.kind='route_knowledge_granted' AND NOT EXISTS (
						SELECT 1 FROM events compensation
						WHERE compensation.kind='compensation'
						  AND compensation.payload->>'compensates_event_id'=e.event_id::text
					)
				),0)
				- COALESCE(sum((e.payload->>'cost')::numeric) FILTER (WHERE e.kind='route_hint_purchased'),0)
			)::text
			FROM events e JOIN save_streams s ON s.id=e.stream_id
			WHERE s.owner_id=$1 AND e.kind IN ('route_knowledge_granted','route_hint_purchased')`, founderID).Scan(&netText); err != nil {
			return err
		}
		net, ok := new(big.Int).SetString(netText, 10)
		limit := big.NewInt(decimal.MaxExactInteger)
		if !ok || new(big.Int).Abs(new(big.Int).Set(net)).Cmp(limit) > 0 {
			return ErrProjection
		}
		if net.Sign() >= 0 {
			balance = net.Int64()
		} else {
			debt = new(big.Int).Neg(net).Int64()
		}
		if _, err := tx.ExecContext(ctx, `UPDATE founder_route_state SET route_knowledge_balance=$2,route_knowledge_debt=$3 WHERE founder_id=$1`, founderID, balance, debt); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
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
