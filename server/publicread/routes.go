package publicread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/routeprojection"
)

const ListRoutesOperation = "list_public_routes"

// RouteReader is the route-projection-owned registry source (C10).
type RouteReader interface {
	PublicRoutePage(ctx context.Context, after string, limit int) ([]routeprojection.PublicRoute, bool, error)
}

func routeSchemas() []publicapi.NamedSchema {
	const maxExactInteger = int64(9_007_199_254_740_991)
	return []publicapi.NamedSchema{
		{Name: "PublicRoute", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("adoption_count", integer(1, maxExactInteger)),
			field("credited_at", stringSchema("date-time-ms")),
			field("first_executor_founder_id", nullable(stringSchema("uuid"))),
			field("naming_deadline", stringSchema("date-time-ms")),
			field("naming_status", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"house", "pending", "published", "reserved"}}),
			field("public_name", stringSchema("")),
			field("route_id", stringSchema("mechanical-id")),
		}}},
		{Name: "PublicRouteCursorKey", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("route_id", stringSchema("mechanical-id")),
		}}},
		{Name: "PublicRoutePage", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("items", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "PublicRoute"}}),
			field("next_cursor", nullable(stringSchema(""))),
		}}},
	}
}

func routesOperation(errorResponse func(int, ...[]byte) publicapi.Response) publicapi.Operation {
	return publicapi.Operation{
		ID: ListRoutesOperation, Method: http.MethodGet, Path: "/api/public/v1/registry/routes",
		Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
		Query: []publicapi.QueryParameter{
			{Name: "cursor", Schema: stringSchema("")},
			{Name: "limit", Schema: integer(1, routeprojection.PublicRoutePageMaximum)},
		},
		CursorKey: "PublicRouteCursorKey",
		Responses: []publicapi.Response{
			{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: "PublicRoutePage"},
			errorResponse(http.StatusBadRequest, invalidCursor, invalidLimit),
			errorResponse(http.StatusTooManyRequests, exactErrors([2]string{"rate_limited", "ip"})...),
			errorResponse(http.StatusInternalServerError, internalPublic),
		},
	}
}

// RoutesFilterSHA256 binds the route cursor to its normalized filter {"limit":N}.
func RoutesFilterSHA256(limit int) string {
	digest := sha256.Sum256([]byte(`{"limit":` + strconv.Itoa(limit) + `}`))
	return hex.EncodeToString(digest[:])
}

type RoutesHandler struct {
	Registry *publicapi.Registry
	Cursors  *publicapi.CursorCodec
	Runtime  *publicapi.Runtime
	Routes   RouteReader
}

type publicRoute struct {
	AdoptionCount          int64   `json:"adoption_count"`
	CreditedAt             string  `json:"credited_at"`
	FirstExecutorFounderID *string `json:"first_executor_founder_id"`
	NamingDeadline         string  `json:"naming_deadline"`
	NamingStatus           string  `json:"naming_status"`
	PublicName             string  `json:"public_name"`
	RouteID                string  `json:"route_id"`
}

type publicRoutePage struct {
	Items      []publicRoute `json:"items"`
	NextCursor *string       `json:"next_cursor"`
}

type routeCursorKey struct {
	RouteID string `json:"route_id"`
}

func (handler RoutesHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if handler.Registry == nil || handler.Cursors == nil || handler.Runtime == nil || handler.Routes == nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	query, err := handler.Registry.ParseQuery(ListRoutesOperation, request.URL.Query())
	if err != nil {
		var invalid *publicapi.InvalidQueryError
		switch {
		case errors.As(err, &invalid) && invalid.Parameter == "cursor":
			writeExact(response, http.StatusBadRequest, invalidCursor)
		case errors.As(err, &invalid) && invalid.Parameter == "limit":
			writeExact(response, http.StatusBadRequest, invalidLimit)
		default:
			writeExact(response, http.StatusInternalServerError, internalPublic)
		}
		return
	}
	limit := DefaultPageLimit
	if value, ok := query["limit"].(int64); ok {
		limit = int(value)
	}
	filter := RoutesFilterSHA256(limit)
	after := ""
	if token, ok := query["cursor"].(string); ok {
		var key routeCursorKey
		if handler.Cursors.Decode(token, ListRoutesOperation, filter, &key) != nil || key.RouteID == "" {
			writeExact(response, http.StatusBadRequest, invalidCursor)
			return
		}
		after = key.RouteID
	}
	routes, more, err := handler.Routes.PublicRoutePage(request.Context(), after, limit)
	if err != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	page := publicRoutePage{Items: make([]publicRoute, len(routes))}
	for index, route := range routes {
		page.Items[index] = publicRoute{AdoptionCount: route.AdoptionCount, CreditedAt: formatTime(route.CreditedAt), FirstExecutorFounderID: route.FirstExecutor,
			NamingDeadline: formatTime(route.NamingDeadline), NamingStatus: route.NamingStatus, PublicName: route.PublicName, RouteID: route.RouteID}
	}
	if more && len(routes) != 0 {
		token, err := handler.Cursors.Encode(ListRoutesOperation, filter, routeCursorKey{RouteID: routes[len(routes)-1].RouteID})
		if err != nil {
			writeExact(response, http.StatusInternalServerError, internalPublic)
			return
		}
		page.NextCursor = &token
	}
	body, err := json.Marshal(page)
	if err != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	body = append(body, '\n')
	if handler.Registry.ValidateResponse(ListRoutesOperation, http.StatusOK, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	if handler.Runtime.WriteCached(response, request, publicapi.CacheRegistry, publicapi.ContentJSON, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
	}
}
