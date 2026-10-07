package publicread

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/publicapi"
)

const (
	GetRunGenesisOperation   = "get_public_run_genesis"
	GetRunReplayLogOperation = "get_public_run_replay_log"
	GetRunVerdictOperation   = "get_public_run_verdict"
	EvidenceHashHeader       = "X-Content-SHA256"
)

var unknownRun = exactErrors([2]string{"unknown_id", "run"})[0]

type EvidenceReader interface {
	PublicRunEvidence(context.Context, string, int64) (leaderboard.PublicRunEvidence, error)
}

// C14's field set translated directly to mechanical snake_case wire names.
// Genesis version is available in the replay archive, not an additional v1 field.
type publicRunVerdict struct {
	CatalogURL      string `json:"catalog_url"`
	ConstantsHash   string `json:"constants_hash"`
	EngineVersion   string `json:"engine_version"`
	GenesisSHA256   string `json:"genesis_sha256"`
	GenesisURL      string `json:"genesis_url"`
	ReplayLogSHA256 string `json:"replay_log_sha256"`
	ReplayLogURL    string `json:"replay_log_url"`
	RunID           string `json:"run_id"`
	Verdict         string `json:"verdict"`
}

func evidenceSchemas() []publicapi.NamedSchema {
	return []publicapi.NamedSchema{{Name: "PublicRunVerdict", Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
		field("catalog_url", stringSchema("")),
		field("constants_hash", stringSchema("sha256-prefixed")),
		field("engine_version", stringSchema("semver")),
		field("genesis_sha256", stringSchema("sha256-prefixed")),
		field("genesis_url", stringSchema("")),
		field("replay_log_sha256", stringSchema("sha256-prefixed")),
		field("replay_log_url", stringSchema("")),
		field("run_id", stringSchema("")),
		field("verdict", &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{"verified"}}),
	}}}}
}

func evidenceOperations(errorResponse func(int, ...[]byte) publicapi.Response) []publicapi.Operation {
	operations := []publicapi.Operation{}
	for _, resource := range []struct{ id, suffix, contentType string }{
		{GetRunGenesisOperation, "genesis", publicapi.ContentJSON},
		{GetRunReplayLogOperation, "replay-log", publicapi.ContentGzip},
		{GetRunVerdictOperation, "verdict", publicapi.ContentJSON},
	} {
		success := publicapi.Response{Kind: publicapi.ResponseRaw, Status: http.StatusOK, ContentType: resource.contentType, ContentHashHeader: EvidenceHashHeader}
		if resource.id == GetRunVerdictOperation {
			success = publicapi.Response{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: "PublicRunVerdict"}
		}
		operations = append(operations, publicapi.Operation{
			ID: resource.id, Method: http.MethodGet, Path: "/api/public/v1/runs/{stream}/{seq}/" + resource.suffix,
			Surface: publicapi.SurfacePublicV1, Auth: publicapi.AuthNone, Public: true,
			Parameters: []publicapi.Parameter{{Name: "stream", Schema: stringSchema("uuid")}, {Name: "seq", Schema: integer(1, 9_007_199_254_740_991)}},
			Responses: []publicapi.Response{success, errorResponse(http.StatusNotFound, unknownRun),
				errorResponse(http.StatusTooManyRequests, exactErrors([2]string{"rate_limited", "ip"})...),
				errorResponse(http.StatusInternalServerError, internalPublic)},
		})
	}
	return operations
}

type EvidenceHandler struct {
	OperationID string
	Registry    *publicapi.Registry
	Runtime     *publicapi.Runtime
	Evidence    EvidenceReader
}

func (handler EvidenceHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if handler.Registry == nil || handler.Runtime == nil || handler.Evidence == nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	stream, seqText := chi.URLParam(request, "stream"), chi.URLParam(request, "seq")
	seq, err := strconv.ParseInt(seqText, 10, 64)
	if err != nil || seq < 1 || seq > 9_007_199_254_740_991 || strconv.FormatInt(seq, 10) != seqText {
		writeExact(response, http.StatusNotFound, unknownRun)
		return
	}
	evidence, err := handler.Evidence.PublicRunEvidence(request.Context(), stream, seq)
	if errors.Is(err, leaderboard.ErrUnknownPublicRun) {
		writeExact(response, http.StatusNotFound, unknownRun)
		return
	}
	if err != nil || evidence.RunID != stream+":"+seqText {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	base := "/api/public/v1/runs/" + stream + "/" + seqText
	var body []byte
	var contentType, hash string
	switch handler.OperationID {
	case GetRunGenesisOperation:
		body, contentType, hash = evidence.Genesis, publicapi.ContentJSON, evidence.GenesisSHA256
	case GetRunReplayLogOperation:
		body, contentType, hash = evidence.ReplayLog, publicapi.ContentGzip, evidence.ReplayLogSHA256
	case GetRunVerdictOperation:
		body, err = json.Marshal(publicRunVerdict{
			CatalogURL: "/api/public/v1/catalogs/" + evidence.ConstantsHash, ConstantsHash: evidence.ConstantsHash,
			EngineVersion: evidence.EngineVersion, GenesisSHA256: evidence.GenesisSHA256, GenesisURL: base + "/genesis",
			ReplayLogSHA256: evidence.ReplayLogSHA256, ReplayLogURL: base + "/replay-log", RunID: evidence.RunID, Verdict: "verified",
		})
		body = append(body, '\n')
		if err != nil || handler.Registry.ValidateResponse(handler.OperationID, http.StatusOK, body) != nil {
			writeExact(response, http.StatusInternalServerError, internalPublic)
			return
		}
		if handler.Runtime.WriteCached(response, request, publicapi.CacheVerification, publicapi.ContentJSON, body) != nil {
			writeExact(response, http.StatusInternalServerError, internalPublic)
		}
		return
	default:
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	hash = strings.TrimPrefix(hash, "sha256:")
	if handler.Registry.ValidateRawResponse(handler.OperationID, http.StatusOK, contentType, body, hash) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
		return
	}
	operation, _ := handler.Registry.Operation(handler.OperationID)
	if handler.Runtime.WriteCached(response, request, publicapi.CacheVerification, contentType, body, operation.Responses[0].ContentHashHeader) != nil {
		writeExact(response, http.StatusInternalServerError, internalPublic)
	}
}

func sortOperations(operations []publicapi.Operation) []publicapi.Operation {
	sort.Slice(operations, func(i, j int) bool { return operations[i].ID < operations[j].ID })
	return operations
}
