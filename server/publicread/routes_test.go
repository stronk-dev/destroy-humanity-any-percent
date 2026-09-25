package publicread

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"cloud-clicker/server/routeprojection"
)

type fakeRoutes struct {
	rows  []routeprojection.PublicRoute // route_id-ordered
	err   error
	calls []string
}

func (fake *fakeRoutes) PublicRoutePage(_ context.Context, after string, limit int) ([]routeprojection.PublicRoute, bool, error) {
	fake.calls = append(fake.calls, after)
	if fake.err != nil {
		return nil, false, fake.err
	}
	page := []routeprojection.PublicRoute{}
	for _, row := range fake.rows {
		if after == "" || row.RouteID > after {
			page = append(page, row)
		}
	}
	if len(page) > limit {
		return page[:limit], true, nil
	}
	return page, false, nil
}

func routeRows() []routeprojection.PublicRoute {
	credited := time.Date(2026, 9, 3, 8, 0, 0, 7_000_000, time.UTC)
	founder := "01986666-0000-4000-8000-0000000000aa"
	return []routeprojection.PublicRoute{
		{RouteID: "route.acquihire_out_of_bounds", PublicName: "House A", FirstExecutor: &founder, CreditedAt: credited, NamingStatus: "reserved", NamingDeadline: credited.Add(72 * time.Hour), AdoptionCount: 3},
		{RouteID: "route.ipo_sequence_break", PublicName: "Player Name", FirstExecutor: nil, CreditedAt: credited, NamingStatus: "published", NamingDeadline: credited.Add(72 * time.Hour), AdoptionCount: 1},
	}
}

func TestRoutesReaderServesTheC12RoutePageWithKeysetPaging(t *testing.T) {
	routes := &fakeRoutes{rows: routeRows()}
	router := boardRouterWithRoutes(t, routes)
	first := getBoard(router, "/api/public/v1/registry/routes", url.Values{"limit": {"1"}})
	want := `{"items":[{"adoption_count":3,"credited_at":"2026-09-03T08:00:00.007Z","first_executor_founder_id":"01986666-0000-4000-8000-0000000000aa",` +
		`"naming_deadline":"2026-09-06T08:00:00.007Z","naming_status":"reserved","public_name":"House A","route_id":"route.acquihire_out_of_bounds"}],"next_cursor":`
	if first.Code != 200 || first.Header().Get("Cache-Control") != "public,max-age=300" || len(first.Body.String()) < len(want) || first.Body.String()[:len(want)] != want {
		t.Fatalf("first routes page %d %v\n%s", first.Code, first.Header(), first.Body.String())
	}
	var page publicRoutePage
	if err := json.Unmarshal(first.Body.Bytes(), &page); err != nil || page.NextCursor == nil {
		t.Fatal(first.Body.String())
	}
	second := getBoard(router, "/api/public/v1/registry/routes", url.Values{"limit": {"1"}, "cursor": {*page.NextCursor}})
	var secondPage publicRoutePage
	if second.Code != 200 || json.Unmarshal(second.Body.Bytes(), &secondPage) != nil || len(secondPage.Items) != 1 || secondPage.NextCursor != nil ||
		secondPage.Items[0].FirstExecutorFounderID != nil || routes.calls[1] != "route.acquihire_out_of_bounds" {
		t.Fatalf("second routes page %d %s calls=%v", second.Code, second.Body.String(), routes.calls)
	}
	if bound := getBoard(router, "/api/public/v1/registry/routes", url.Values{"limit": {"2"}, "cursor": {*page.NextCursor}}); bound.Code != 400 || bound.Body.String() != string(invalidCursor) {
		t.Fatalf("the route cursor must bind the limit: %d %s", bound.Code, bound.Body.String())
	}
	if limit := getBoard(router, "/api/public/v1/registry/routes", url.Values{"limit": {"0"}}); limit.Code != 400 || limit.Body.String() != string(invalidLimit) {
		t.Fatalf("limit 0: %d", limit.Code)
	}
	failing := boardRouterWithRoutes(t, &fakeRoutes{err: errors.New("database down")})
	if failure := getBoard(failing, "/api/public/v1/registry/routes", url.Values{}); failure.Code != 500 || failure.Body.String() != string(internalPublic) {
		t.Fatalf("reader failure: %d", failure.Code)
	}
}
