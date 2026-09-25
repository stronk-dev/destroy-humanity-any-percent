// Package publicread owns the unauthenticated /api/public/v1/ read operations
// (API Foundation A3/A6). Every operation is a public registry row; handlers
// serve exact DTO bytes validated against that row before writing.
package publicread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

const (
	ListEpochsOperation = "list_public_epochs"
	// DefaultPageLimit and the 1..100 bound are ruled by C12.
	DefaultPageLimit = 50
)

func integer(minimum, maximum int64) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
}

func field(name string, schema *publicapi.Schema) publicapi.Field {
	return publicapi.Field{Name: name, Schema: schema, Required: true}
}

func stringSchema(format string) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaString, Format: format}
}

func nullable(schema *publicapi.Schema) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{{Kind: publicapi.SchemaNull}, schema}}
}

func exactErrors(pairs ...[2]string) [][]byte {
	result := make([][]byte, len(pairs))
	for index, pair := range pairs {
		result[index] = []byte(`{"category":` + strconv.Quote(pair[0]) + `,"detail":` + strconv.Quote(pair[1]) + "}\n")
	}
	return result
}

var (
	invalidCursor  = exactErrors([2]string{"invalid", "cursor"})[0]
	invalidLimit   = exactErrors([2]string{"invalid", "limit"})[0]
	internalPublic = exactErrors([2]string{"internal_invariant", "public_api"})[0]
)

// Schemas are the public DTO descriptors (C12 EpochPage literal). APIError is
// the single account-owned definition shared with the private registry.
func Schemas() []publicapi.NamedSchema {
	const maxExactInteger = int64(9_007_199_254_740_991)
	schemas := []publicapi.NamedSchema{
		account.APIErrorSchema(),
		{Name: "PublicEpoch", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("accepted_hashes", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: stringSchema("sha256-prefixed")}),
			field("changelog_markdown", stringSchema("")),
			field("changelog_ref", stringSchema("")),
			field("ended_at", nullable(stringSchema("date-time-ms"))),
			field("epoch_id", integer(1, maxExactInteger)),
			field("name", stringSchema("")),
			field("started_at", stringSchema("date-time-ms")),
		}}},
		{Name: "PublicEpochCursorKey", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("epoch_id", integer(1, maxExactInteger)),
		}}},
		{Name: "PublicEpochPage", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
			field("items", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: "PublicEpoch"}}),
			field("next_cursor", nullable(stringSchema(""))),
		}}},
	}
	schemas = append(schemas, boardSchemas()...)
	sort.Slice(schemas, func(left, right int) bool { return schemas[left].Name < schemas[right].Name })
	return schemas
}

// Operations are the public registry rows. Error responses are narrowed to the
// exact pairs each handler can emit.
func Operations() []publicapi.Operation {
	errorResponse := func(status int, body ...[]byte) publicapi.Response {
		return publicapi.Response{Kind: publicapi.ResponseSchema, Status: status, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: body}
	}
	return []publicapi.Operation{boardOperation(errorResponse), {
		ID: ListEpochsOperation, Method: http.MethodGet, Path: "/api/public/v1/epochs",
		Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
		Query: []publicapi.QueryParameter{
			{Name: "cursor", Schema: stringSchema("")},
			{Name: "limit", Schema: integer(1, leaderboard.PublicEpochPageMaximum)},
		},
		CursorKey: "PublicEpochCursorKey",
		Responses: []publicapi.Response{
			{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: "PublicEpochPage"},
			errorResponse(http.StatusBadRequest, invalidCursor, invalidLimit),
			errorResponse(http.StatusTooManyRequests, exactErrors([2]string{"rate_limited", "ip"})...),
			errorResponse(http.StatusInternalServerError, internalPublic),
		},
	}}
}

// Registry is the immutable public-read authority used for mounting, cursor
// codecs, and (merged with the private registry) generated artifacts.
func Registry() (*publicapi.Registry, error) {
	return publicapi.NewRegistry(Schemas(), Operations())
}

// EpochReader is the leaderboard-owned page source.
type EpochReader interface {
	PublicEpochPage(ctx context.Context, before int64, limit int) ([]leaderboard.PublicEpoch, bool, error)
}

type EpochsHandler struct {
	Registry *publicapi.Registry
	Cursors  *publicapi.CursorCodec
	Runtime  *publicapi.Runtime
	Epochs   EpochReader
}

type publicEpoch struct {
	AcceptedHashes    []string `json:"accepted_hashes"`
	ChangelogMarkdown string   `json:"changelog_markdown"`
	ChangelogRef      string   `json:"changelog_ref"`
	EndedAt           *string  `json:"ended_at"`
	EpochID           int64    `json:"epoch_id"`
	Name              string   `json:"name"`
	StartedAt         string   `json:"started_at"`
}

type publicEpochPage struct {
	Items      []publicEpoch `json:"items"`
	NextCursor *string       `json:"next_cursor"`
}

type epochCursorKey struct {
	EpochID int64 `json:"epoch_id"`
}

// EpochsFilterSHA256 is the normalized filter the epoch cursor binds to: the
// canonical exact object {"limit":N}.
func EpochsFilterSHA256(limit int) string {
	digest := sha256.Sum256([]byte(`{"limit":` + strconv.Itoa(limit) + `}`))
	return hex.EncodeToString(digest[:])
}

func formatTime(value time.Time) string { return value.UTC().Format("2006-01-02T15:04:05.000Z") }

func (handler EpochsHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if handler.Registry == nil || handler.Cursors == nil || handler.Runtime == nil || handler.Epochs == nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	query, err := handler.Registry.ParseQuery(ListEpochsOperation, request.URL.Query())
	if err != nil {
		var invalid *publicapi.InvalidQueryError
		if errors.As(err, &invalid) && invalid.Parameter == "cursor" {
			writeExact(response, http.StatusBadRequest, invalidCursor)
		} else if errors.As(err, &invalid) && invalid.Parameter == "limit" {
			writeExact(response, http.StatusBadRequest, invalidLimit)
		} else {
			writeExact(response, http.StatusInternalServerError, internalPublic)
		}
		return
	}
	limit := DefaultPageLimit
	if value, ok := query["limit"].(int64); ok {
		limit = int(value)
	}
	filter := EpochsFilterSHA256(limit)
	before := int64(0)
	if token, ok := query["cursor"].(string); ok {
		var key epochCursorKey
		if handler.Cursors.Decode(token, ListEpochsOperation, filter, &key) != nil {
			writeExact(response, http.StatusBadRequest, invalidCursor)
			return
		}
		before = key.EpochID
	}
	epochs, more, err := handler.Epochs.PublicEpochPage(request.Context(), before, limit)
	if err != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	page := publicEpochPage{Items: make([]publicEpoch, len(epochs))}
	for index, epoch := range epochs {
		item := publicEpoch{AcceptedHashes: epoch.Hashes, ChangelogMarkdown: string(epoch.Changelog), ChangelogRef: epoch.ChangelogRef,
			EpochID: epoch.ID, Name: epoch.Name, StartedAt: formatTime(epoch.StartedAt)}
		if epoch.EndedAt != nil {
			ended := formatTime(*epoch.EndedAt)
			item.EndedAt = &ended
		}
		page.Items[index] = item
	}
	if more && len(epochs) != 0 {
		token, err := handler.Cursors.Encode(ListEpochsOperation, filter, epochCursorKey{EpochID: epochs[len(epochs)-1].ID})
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
	// The served bytes are validated against the registry row before writing.
	if handler.Registry.ValidateResponse(ListEpochsOperation, http.StatusOK, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	if handler.Runtime.WriteCached(response, request, publicapi.CacheCatalogEpoch, publicapi.ContentJSON, body) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
	}
}

func writeExact(response http.ResponseWriter, status int, body []byte) {
	response.Header().Set("Content-Type", publicapi.ContentJSON)
	response.WriteHeader(status)
	_, _ = response.Write(body)
}
