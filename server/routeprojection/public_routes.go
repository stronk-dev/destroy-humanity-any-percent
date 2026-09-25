package routeprojection

import (
	"context"
	"errors"
	"time"
)

// PublicRoutePageMaximum is the C12 list bound.
const PublicRoutePageMaximum = 100

var ErrInvalidPublicRoutePage = errors.New("invalid public route page")

// PublicRoute is the C12 RoutePage row: permanent public registry facts only.
// The public name is the approved player name only once published; reserved,
// pending (unmoderated) and expired names show the house name. The first
// executor is withheld (nil) once that Founder's account is anonymized, and
// also when no ownership row exists, so no path can leak an orphaned identity.
type PublicRoute struct {
	RouteID        string
	PublicName     string
	FirstExecutor  *string
	CreditedAt     time.Time
	NamingStatus   string
	NamingDeadline time.Time
	AdoptionCount  int64
}

// PublicRoutePage reads route_id-ordered rows after the exclusive keyset bound.
// route_id is the ordering because it is immutable; credited_at is not (an
// earlier execution can re-credit a route).
func (p *Projector) PublicRoutePage(ctx context.Context, after string, limit int) ([]PublicRoute, bool, error) {
	if p == nil || limit < 1 || limit > PublicRoutePageMaximum {
		return nil, false, ErrInvalidPublicRoutePage
	}
	rows, err := p.db.QueryContext(ctx, `
		SELECT routes.route_id,
		       CASE WHEN routes.name_state='published' THEN routes.name ELSE routes.house_name END,
		       CASE WHEN founders.account_id IS NULL THEN NULL ELSE routes.first_founder_id::text END,
		       routes.occurred_at, routes.name_state, routes.naming_reserved_until, routes.execution_count
		FROM registry_routes routes
		LEFT JOIN account_founders founders ON founders.founder_id=routes.first_founder_id
		WHERE $1='' OR routes.route_id > $1
		ORDER BY routes.route_id
		LIMIT $2`, after, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	result := []PublicRoute{}
	for rows.Next() {
		var route PublicRoute
		if err := rows.Scan(&route.RouteID, &route.PublicName, &route.FirstExecutor, &route.CreditedAt, &route.NamingStatus, &route.NamingDeadline, &route.AdoptionCount); err != nil {
			return nil, false, err
		}
		result = append(result, route)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	if len(result) > limit {
		return result[:limit], true, nil
	}
	return result, false, nil
}
