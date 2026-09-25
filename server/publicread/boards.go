package publicread

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

const ListBoardOperation = "list_public_board"

var (
	invalidEpoch     = exactErrors([2]string{"invalid", "epoch"})[0]
	invalidMandate   = exactErrors([2]string{"invalid", "mandate"})[0]
	invalidVariables = exactErrors([2]string{"invalid", "variables"})[0]
	unknownCategory  = exactErrors([2]string{"unknown_id", "category"})[0]
	unknownEpoch     = exactErrors([2]string{"unknown_id", "epoch"})[0]
)

// BoardReader is the leaderboard-owned board source (C10 narrow interface).
type BoardReader interface {
	PublicBoardRankingKind(ctx context.Context, categoryID string, epochID int64) (leaderboard.RankingKind, error)
	PublicBoardPage(ctx context.Context, kind leaderboard.RankingKind, query leaderboard.PublicBoardQuery) ([]leaderboard.PublicBoardRow, bool, error)
}

func boardSchemas() []publicapi.NamedSchema {
	const maxExactInteger = int64(9_007_199_254_740_991)
	const maxMagnitudeExponent = int64(8_999_999_999_999_999)
	flag := integer(0, 1)
	return []publicapi.NamedSchema{
		{Name: "PublicBoardCursorKey", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("exponent", nullable(integer(-maxMagnitudeExponent, maxMagnitudeExponent))),
			field("key", nullable(integer(-maxExactInteger, maxExactInteger))),
			field("quantized_mantissa", nullable(integer(0, 999_999_999_999))),
			field("run_id", stringSchema("")),
		}}},
		{Name: "PublicBoardItem", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("founder_id", stringSchema("uuid")),
			field("key", &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "PublicBoardKey"}),
			field("rank", integer(1, maxExactInteger)),
			field("run_id", stringSchema("")),
			field("verified_at", stringSchema("date-time-ms")),
			field("world_first", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
		}}},
		{Name: "PublicBoardKey", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
				field("kind", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"count", "time_ms"}}),
				field("value", integer(-maxExactInteger, maxExactInteger)),
			}},
			{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
				field("exponent", integer(-maxMagnitudeExponent, maxMagnitudeExponent)),
				field("kind", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"magnitude"}}),
				field("quantized_mantissa", integer(0, 999_999_999_999)),
			}},
		}}},
		{Name: "PublicBoardPage", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("category_id", stringSchema("mechanical-id")),
			field("epoch_id", integer(1, maxExactInteger)),
			field("items", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "PublicBoardItem"}}),
			field("mandate_level", integer(0, 20)),
			field("next_cursor", nullable(stringSchema(""))),
			field("ranking_kind", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"count", "magnitude", "time_ms"}}),
			field("variables", &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "PublicBoardVariables"}),
		}}},
		{Name: "PublicBoardVariables", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("advisor", flag),
			field("commons", flag),
			field("faction", nullable(stringSchema("mechanical-id"))),
			field("glitched", flag),
		}}},
	}
}

func boardOperation(errorResponse func(int, ...[]byte) publicapi.Response) publicapi.Operation {
	const maxExactInteger = int64(9_007_199_254_740_991)
	return publicapi.Operation{
		ID: ListBoardOperation, Method: http.MethodGet, Path: "/api/public/v1/boards/{category}",
		Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
		Parameters: []publicapi.Parameter{{Name: "category", Schema: stringSchema("mechanical-id")}},
		Query: []publicapi.QueryParameter{
			{Name: "cursor", Schema: stringSchema("")},
			{Name: "epoch", Schema: integer(1, maxExactInteger), Required: true},
			{Name: "limit", Schema: integer(1, leaderboard.PublicBoardPageMaximum)},
			{Name: "mandate", Schema: integer(0, 20), Required: true},
			{Name: "variables", Schema: stringSchema(""), Required: true},
		},
		CursorKey: "PublicBoardCursorKey",
		Responses: []publicapi.Response{
			{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: "PublicBoardPage"},
			errorResponse(http.StatusBadRequest, invalidCursor, invalidEpoch, invalidLimit, invalidMandate, invalidVariables),
			errorResponse(http.StatusNotFound, unknownCategory, unknownEpoch),
			errorResponse(http.StatusTooManyRequests, exactErrors([2]string{"rate_limited", "ip"})...),
			errorResponse(http.StatusInternalServerError, internalPublic),
		},
	}
}

type BoardHandler struct {
	Registry *publicapi.Registry
	Cursors  *publicapi.CursorCodec
	Runtime  *publicapi.Runtime
	Boards   BoardReader
}

type boardCursorKey struct {
	Exponent          *int64 `json:"exponent"`
	Key               *int64 `json:"key"`
	QuantizedMantissa *int64 `json:"quantized_mantissa"`
	RunID             string `json:"run_id"`
}

type publicBoardItem struct {
	FounderID  string         `json:"founder_id"`
	Key        map[string]any `json:"key"`
	Rank       int64          `json:"rank"`
	RunID      string         `json:"run_id"`
	VerifiedAt string         `json:"verified_at"`
	WorldFirst bool           `json:"world_first"`
}

type publicBoardPage struct {
	CategoryID   string                   `json:"category_id"`
	EpochID      int64                    `json:"epoch_id"`
	Items        []publicBoardItem        `json:"items"`
	MandateLevel int                      `json:"mandate_level"`
	NextCursor   *string                  `json:"next_cursor"`
	RankingKind  string                   `json:"ranking_kind"`
	Variables    publicapi.BoardVariables `json:"variables"`
}

func (handler BoardHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if handler.Registry == nil || handler.Cursors == nil || handler.Runtime == nil || handler.Boards == nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	query, err := handler.Registry.ParseQuery(ListBoardOperation, request.URL.Query())
	if err != nil {
		var invalid *publicapi.InvalidQueryError
		if !errors.As(err, &invalid) {
			writeExact(response, http.StatusInternalServerError, internalPublic)
			return
		}
		body := map[string][]byte{"cursor": invalidCursor, "epoch": invalidEpoch, "limit": invalidLimit, "mandate": invalidMandate, "variables": invalidVariables}[invalid.Parameter]
		if body == nil {
			writeExact(response, http.StatusInternalServerError, internalPublic)
			return
		}
		writeExact(response, http.StatusBadRequest, body)
		return
	}
	variables, err := publicapi.DecodeBoardVariables(query["variables"].(string))
	if err != nil {
		writeExact(response, http.StatusBadRequest, invalidVariables)
		return
	}
	category := chi.URLParam(request, "category")
	filter := publicapi.BoardFilter{Category: category, Variables: variables, EpochID: query["epoch"].(int64),
		MandateLevel: int(query["mandate"].(int64)), Limit: DefaultPageLimit}
	if value, ok := query["limit"].(int64); ok {
		filter.Limit = int(value)
	}
	kind, err := handler.Boards.PublicBoardRankingKind(request.Context(), category, filter.EpochID)
	switch {
	case errors.Is(err, leaderboard.ErrUnknownPublicCategory):
		writeExact(response, http.StatusNotFound, unknownCategory)
		return
	case errors.Is(err, leaderboard.ErrUnknownPublicEpoch):
		writeExact(response, http.StatusNotFound, unknownEpoch)
		return
	case err != nil:
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	filterSHA256, err := publicapi.BoardFilterSHA256(filter)
	if err != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	boardQuery := leaderboard.PublicBoardQuery{CategoryID: category, EpochID: filter.EpochID, MandateLevel: filter.MandateLevel, Limit: filter.Limit,
		Variables: leaderboard.Variables{Advisor: variables.Advisor == 1, Commons: variables.Commons == 1, Glitched: variables.Glitched == 1, Faction: variables.Faction}}
	if token, ok := query["cursor"].(string); ok {
		var key boardCursorKey
		if handler.Cursors.Decode(token, ListBoardOperation, filterSHA256, &key) != nil || !applyBoardCursor(kind, key, &boardQuery) {
			writeExact(response, http.StatusBadRequest, invalidCursor)
			return
		}
	}
	rows, more, err := handler.Boards.PublicBoardPage(request.Context(), kind, boardQuery)
	if err != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	page := publicBoardPage{CategoryID: category, EpochID: filter.EpochID, Items: make([]publicBoardItem, len(rows)), MandateLevel: filter.MandateLevel,
		RankingKind: string(kind), Variables: variables}
	for index, row := range rows {
		item := publicBoardItem{FounderID: row.FounderID, Rank: row.Rank, RunID: row.RunID, VerifiedAt: formatTime(row.VerifiedAt), WorldFirst: row.WorldFirst}
		if kind == leaderboard.RankingMagnitude {
			item.Key = map[string]any{"exponent": row.Magnitude.Exponent, "kind": string(kind), "quantized_mantissa": row.Magnitude.Mantissa}
		} else {
			item.Key = map[string]any{"kind": string(kind), "value": row.Key}
		}
		page.Items[index] = item
	}
	if more && len(rows) != 0 {
		last := rows[len(rows)-1]
		key := boardCursorKey{RunID: last.RunID}
		if kind == leaderboard.RankingMagnitude {
			exponent, mantissa := last.Magnitude.Exponent, last.Magnitude.Mantissa
			key.Exponent, key.QuantizedMantissa = &exponent, &mantissa
		} else {
			value := last.Key
			key.Key = &value
		}
		token, err := handler.Cursors.Encode(ListBoardOperation, filterSHA256, key)
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
	if handler.Registry.ValidateResponse(ListBoardOperation, http.StatusOK, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	if handler.Runtime.WriteCached(response, request, publicapi.CacheBoards, publicapi.ContentJSON, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
	}
}

// applyBoardCursor accepts only the cursor arm matching the resolved ranking
// kind; the MAC already binds the complete normalized filter.
func applyBoardCursor(kind leaderboard.RankingKind, key boardCursorKey, query *leaderboard.PublicBoardQuery) bool {
	if key.RunID == "" {
		return false
	}
	query.AfterRunID = key.RunID
	if kind == leaderboard.RankingMagnitude {
		if key.Key != nil || key.Exponent == nil || key.QuantizedMantissa == nil {
			return false
		}
		query.AfterMag = &leaderboard.MagnitudeKey{Exponent: *key.Exponent, Mantissa: *key.QuantizedMantissa}
		return true
	}
	if key.Key == nil || key.Exponent != nil || key.QuantizedMantissa != nil {
		return false
	}
	query.AfterKey = key.Key
	return true
}
