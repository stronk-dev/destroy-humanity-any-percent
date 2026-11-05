package routeprojection

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

// synchronizeKnowledge includes committed Exit credits and hint debits even if
// their post-commit delivery has not arrived yet. All callers hold the same
// Founder projection row before claiming either event population. The claims and
// balance update share one transaction and one source snapshot.
//
// routeDelta belongs to a route grant/compensation fact just written by this
// transaction. A missing row rebuilds from all facts, including that delta;
// an existing row applies it alongside the newly claimed Exit/hint changes.
func (p *Projector) synchronizeKnowledge(ctx context.Context, tx *sql.Tx, founderID string, routeDelta int64) (int64, int64, error) {
	insert, err := tx.ExecContext(ctx, `INSERT INTO founder_route_state(founder_id,route_knowledge_balance,route_knowledge_debt)
		VALUES($1,0,0) ON CONFLICT DO NOTHING`, founderID)
	if err != nil {
		return 0, 0, err
	}
	inserted, err := insert.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	var balance, debt int64
	if err := tx.QueryRowContext(ctx, `SELECT route_knowledge_balance,route_knowledge_debt
		FROM founder_route_state WHERE founder_id=$1 FOR UPDATE`, founderID).Scan(&balance, &debt); err != nil {
		return 0, 0, err
	}
	var rebuiltText, changeText string
	var validSource bool
	err = tx.QueryRowContext(ctx, `WITH source AS MATERIALIZED (
		SELECT e.event_id,e.kind,e.payload,s.scope FROM events e
		JOIN save_streams s ON s.id=e.stream_id
		WHERE s.owner_kind='founder' AND s.owner_id=$1 AND (
			(e.kind='founder_advanced' AND
				($2 OR NOT EXISTS(SELECT 1 FROM exit_knowledge_projection_events c WHERE c.event_id=e.event_id))) OR
			(e.kind='route_hint_purchased' AND s.scope='founder' AND
				($2 OR NOT EXISTS(SELECT 1 FROM route_hint_projection_events c WHERE c.event_id=e.event_id))) OR
			($2 AND e.kind='route_knowledge_granted')
		)
	), exit_claims AS (
		INSERT INTO exit_knowledge_projection_events(event_id,founder_id,amount)
		SELECT event_id,$1,(payload->>'route_knowledge')::bigint FROM source WHERE kind='founder_advanced'
		ON CONFLICT DO NOTHING RETURNING amount
	), hint_claims AS (
		INSERT INTO route_hint_projection_events(event_id,founder_id,route_id,cost)
		SELECT event_id,$1,payload->>'route_id',(payload->>'cost')::bigint FROM source WHERE kind='route_hint_purchased'
		ON CONFLICT DO NOTHING RETURNING cost
	)
	SELECT (
		COALESCE(sum((payload->>'route_knowledge')::numeric) FILTER (WHERE kind='founder_advanced'),0) +
		COALESCE(sum((payload->>'amount')::numeric) FILTER (WHERE kind='route_knowledge_granted' AND NOT EXISTS (
			SELECT 1 FROM events compensation WHERE compensation.kind='compensation'
			AND compensation.payload->>'compensates_event_id'=source.event_id::text)),0) -
		COALESCE(sum((payload->>'cost')::numeric) FILTER (WHERE kind='route_hint_purchased'),0)
	)::text, (
		COALESCE((SELECT sum(amount) FROM exit_claims),0) - COALESCE((SELECT sum(cost) FROM hint_claims),0)
	)::text, COALESCE(bool_and(kind!='founder_advanced' OR
		(scope='founder' AND COALESCE(payload->>'founder_id'=$1::text,false))),true)
	FROM source`, founderID, inserted == 1).Scan(&rebuiltText, &changeText, &validSource)
	if err != nil {
		return 0, 0, err
	}
	if !validSource {
		return 0, 0, ErrProjection
	}
	netText := changeText
	if inserted == 1 {
		netText = rebuiltText
	}
	net, ok := new(big.Int).SetString(netText, 10)
	if !ok {
		return 0, 0, ErrProjection
	}
	if inserted == 0 {
		net.Add(net, big.NewInt(balance))
		net.Sub(net, big.NewInt(debt))
		net.Add(net, big.NewInt(routeDelta))
	}
	if new(big.Int).Abs(new(big.Int).Set(net)).Cmp(big.NewInt(decimal.MaxExactInteger)) > 0 {
		return 0, 0, fmt.Errorf("%w: Route Knowledge overflow", ErrProjection)
	}
	balance, debt = 0, 0
	if net.Sign() >= 0 {
		balance = net.Int64()
	} else {
		debt = new(big.Int).Neg(net).Int64()
	}
	if _, err := tx.ExecContext(ctx, `UPDATE founder_route_state SET route_knowledge_balance=$2,route_knowledge_debt=$3
		WHERE founder_id=$1`, founderID, balance, debt); err != nil {
		return 0, 0, err
	}
	return balance, debt, nil
}

func (p *Projector) projectExitKnowledge(ctx context.Context, tx *sql.Tx, event save.EventRecord) error {
	var matches bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM events e JOIN save_streams s ON s.id=e.stream_id
		WHERE e.event_id=$1 AND e.kind='founder_advanced' AND e.stream_id=$2 AND s.owner_kind='founder'
		AND s.owner_id=$3 AND s.scope='founder' AND e.payload=$4::jsonb)`,
		event.EventID, event.StreamID, event.OwnerID, event.Payload).Scan(&matches); err != nil {
		return err
	}
	if !matches {
		return ErrProjection
	}
	_, _, err := p.synchronizeKnowledge(ctx, tx, event.OwnerID, 0)
	return err
}
