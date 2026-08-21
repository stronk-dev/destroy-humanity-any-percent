package publicapi

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
)

var ErrInvalidOperation = errors.New("invalid API operation")

type Surface string

const (
	SurfacePrivateV1 Surface = "private_v1"
	SurfacePublicV1  Surface = "public_v1"
)

type AuthMode string

const (
	AuthNone        AuthMode = "none"
	AuthAccessToken AuthMode = "access_token"
)

type ResponseKind string

const (
	ResponseSchema ResponseKind = "schema"
	ResponseRaw    ResponseKind = "raw"
)

const (
	ContentJSON = "application/json"
	ContentGzip = "application/gzip"
)

// Response is the closed C19 operation-response union. Schema responses are
// exact JSON DTOs. Raw responses are repository-owned immutable evidence bytes;
// their declared hash header is mandatory and their media type is deliberately
// limited to the two evidence formats the API contract ships.
type Response struct {
	Kind              ResponseKind
	Status            int
	ContentType       string
	SchemaRef         string
	ContentHashHeader string
}

type Operation struct {
	ID         string
	Method     string
	Path       string
	Surface    Surface
	Auth       AuthMode
	Public     bool
	Parameters []Parameter
	Request    string
	CursorKey  string
	Responses  []Response
}

type Parameter struct {
	Name   string
	Schema *Schema
}

type Registry struct {
	schemas    map[string]*Schema
	operations []Operation
	byID       map[string]Operation
}

type Binding struct {
	OperationID string
	Handler     http.Handler
}

type Middleware func(http.Handler) http.Handler

var operationIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var responseHeaderPattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9-]*$`)

func NewRegistry(schemas []NamedSchema, operations []Operation) (*Registry, error) {
	definitions, err := ValidateSchemaDefinitions(schemas)
	if err != nil {
		return nil, err
	}
	result := &Registry{schemas: definitions, operations: append([]Operation(nil), operations...), byID: map[string]Operation{}}
	lastID := ""
	seenRoutes := map[string]bool{}
	for index, operation := range result.operations {
		if !operationIDPattern.MatchString(operation.ID) || operation.ID <= lastID || !validMethod(operation.Method) ||
			!validOperationPath(operation.Path, operation.Surface) || (operation.Auth != AuthNone && operation.Auth != AuthAccessToken) ||
			operation.Public != (operation.Surface == SurfacePublicV1) || operation.Public && operation.Auth != AuthNone || len(operation.Responses) == 0 {
			return nil, fmt.Errorf("%w: operations[%d]", ErrInvalidOperation, index)
		}
		if operation.Request != "" && definitions[operation.Request] == nil {
			return nil, fmt.Errorf("%w: request schema", ErrInvalidOperation)
		}
		if !validPathParameters(operation.Path, operation.Parameters, definitions) {
			return nil, fmt.Errorf("%w: path parameters", ErrInvalidOperation)
		}
		if operation.CursorKey != "" && (definitions[operation.CursorKey] == nil || definitions[operation.CursorKey].Kind != SchemaObject) {
			return nil, fmt.Errorf("%w: cursor key schema", ErrInvalidOperation)
		}
		lastResponse := ""
		for _, response := range operation.Responses {
			key := fmt.Sprintf("%03d\x00%s", response.Status, response.ContentType)
			if key <= lastResponse || !validResponse(response, definitions) {
				return nil, fmt.Errorf("%w: response descriptor", ErrInvalidOperation)
			}
			lastResponse = key
		}
		key := operation.Method + "\x00" + operation.Path
		if seenRoutes[key] {
			return nil, fmt.Errorf("%w: duplicate route", ErrInvalidOperation)
		}
		seenRoutes[key] = true
		operation.Responses = cloneResponses(operation.Responses)
		operation.Parameters = cloneParameters(operation.Parameters)
		result.operations[index], result.byID[operation.ID], lastID = operation, operation, operation.ID
	}
	return result, nil
}

func (registry *Registry) Operation(id string) (Operation, bool) {
	if registry == nil {
		return Operation{}, false
	}
	operation, ok := registry.byID[id]
	operation.Responses = cloneResponses(operation.Responses)
	operation.Parameters = cloneParameters(operation.Parameters)
	return operation, ok
}

func (registry *Registry) Operations() []Operation {
	if registry == nil {
		return nil
	}
	result := make([]Operation, len(registry.operations))
	for index, operation := range registry.operations {
		operation.Responses = cloneResponses(operation.Responses)
		operation.Parameters = cloneParameters(operation.Parameters)
		result[index] = operation
	}
	return result
}

func (registry *Registry) Schemas() []NamedSchema {
	if registry == nil {
		return nil
	}
	names := make([]string, 0, len(registry.schemas))
	for name := range registry.schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]NamedSchema, len(names))
	for index, name := range names {
		result[index] = NamedSchema{Name: name, Schema: cloneSchema(registry.schemas[name])}
	}
	return result
}

// Mount binds every operation exactly once through the registry's declared
// auth mode. Missing, extra, unsorted, or nil bindings fail before chi sees a
// route, so runtime mounting cannot silently drift from generation.
func (registry *Registry) Mount(router chi.Router, bindings []Binding, middleware map[AuthMode]Middleware) error {
	if registry == nil || router == nil || len(bindings) != len(registry.operations) {
		return ErrInvalidOperation
	}
	for index, operation := range registry.operations {
		binding := bindings[index]
		if binding.OperationID != operation.ID || binding.Handler == nil {
			return ErrInvalidOperation
		}
		handler := binding.Handler
		if operation.Auth != AuthNone {
			wrap := middleware[operation.Auth]
			if wrap == nil {
				return ErrInvalidOperation
			}
			handler = wrap(handler)
		}
		router.Method(operation.Method, operation.Path, handler)
	}
	return nil
}

func (registry *Registry) cursorSchema(operationID string) (string, bool) {
	if registry == nil {
		return "", false
	}
	operation, ok := registry.byID[operationID]
	return operation.CursorKey, ok && operation.CursorKey != ""
}

func (registry *Registry) ValidateResponse(operationID string, status int, data []byte) error {
	operation, ok := registry.byID[operationID]
	if !ok {
		return ErrInvalidOperation
	}
	for _, response := range operation.Responses {
		if response.Status == status && response.Kind == ResponseSchema {
			return ValidateJSON(response.SchemaRef, data, registry.schemas)
		}
	}
	return ErrInvalidOperation
}

func (registry *Registry) ValidateRequest(operationID string, data []byte) error {
	operation, ok := registry.byID[operationID]
	if !ok {
		return ErrInvalidOperation
	}
	if operation.Request == "" {
		if len(data) != 0 {
			return ErrInvalidOperation
		}
		return nil
	}
	return ValidateJSON(operation.Request, data, registry.schemas)
}

func (registry *Registry) ValidateRawResponse(operationID string, status int, contentType string, data []byte, contentHash string) error {
	operation, ok := registry.byID[operationID]
	if !ok || !sha256Pattern.MatchString(contentHash) {
		return ErrInvalidOperation
	}
	for _, response := range operation.Responses {
		if response.Status == status && response.Kind == ResponseRaw && response.ContentType == contentType {
			digest := fmt.Sprintf("%x", sha256.Sum256(data))
			if digest != contentHash {
				return ErrInvalidOperation
			}
			return nil
		}
	}
	return ErrInvalidOperation
}

func validResponse(response Response, definitions map[string]*Schema) bool {
	if response.Status < 100 || response.Status > 599 {
		return false
	}
	switch response.Kind {
	case ResponseSchema:
		return response.ContentType == ContentJSON && definitions[response.SchemaRef] != nil && response.ContentHashHeader == ""
	case ResponseRaw:
		return (response.ContentType == ContentJSON || response.ContentType == ContentGzip) && response.SchemaRef == "" &&
			responseHeaderPattern.MatchString(response.ContentHashHeader)
	default:
		return false
	}
}

func validMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func validOperationPath(path string, surface Surface) bool {
	prefix := "/api/v1/"
	if surface == SurfacePublicV1 {
		prefix = "/api/public/v1/"
	} else if surface != SurfacePrivateV1 {
		return false
	}
	if !strings.HasPrefix(path, prefix) || strings.Contains(path, "//") || strings.ContainsAny(path, "?#") {
		return false
	}
	for _, segment := range strings.Split(strings.TrimPrefix(path, prefix), "/") {
		if segment == "" {
			return false
		}
		if strings.HasPrefix(segment, "{") {
			if !strings.HasSuffix(segment, "}") || !fieldNamePattern.MatchString(segment[1:len(segment)-1]) {
				return false
			}
		} else if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(segment) {
			return false
		}
	}
	return true
}

func cloneResponses(source []Response) []Response {
	return append([]Response(nil), source...)
}

func cloneParameters(source []Parameter) []Parameter {
	result := make([]Parameter, len(source))
	for index, parameter := range source {
		result[index] = Parameter{Name: parameter.Name, Schema: cloneSchema(parameter.Schema)}
	}
	return result
}

func validPathParameters(path string, parameters []Parameter, definitions map[string]*Schema) bool {
	want := []string{}
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			want = append(want, segment[1:len(segment)-1])
		}
	}
	if len(want) != len(parameters) {
		return false
	}
	for index, parameter := range parameters {
		if parameter.Name != want[index] || parameter.Schema == nil || validateSchema(parameter.Schema, definitions, 0) != nil ||
			parameter.Schema.Kind == SchemaObject || parameter.Schema.Kind == SchemaArray || parameter.Schema.Kind == SchemaOneOf || parameter.Schema.Kind == SchemaNull {
			return false
		}
	}
	return true
}

func SortedResponseStatuses(operation Operation) []int {
	result := make([]int, 0, len(operation.Responses))
	for _, response := range operation.Responses {
		if len(result) == 0 || result[len(result)-1] != response.Status {
			result = append(result, response.Status)
		}
	}
	return result
}
