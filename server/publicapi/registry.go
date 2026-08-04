package publicapi

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
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

type Operation struct {
	ID        string
	Method    string
	Path      string
	Surface   Surface
	Auth      AuthMode
	Public    bool
	Request   string
	CursorKey string
	Responses map[int]string
}

type Registry struct {
	schemas    map[string]*Schema
	operations []Operation
	byID       map[string]Operation
}

var operationIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

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
		if operation.CursorKey != "" && (definitions[operation.CursorKey] == nil || definitions[operation.CursorKey].Kind != SchemaObject) {
			return nil, fmt.Errorf("%w: cursor key schema", ErrInvalidOperation)
		}
		for status, schema := range operation.Responses {
			if status < 100 || status > 599 || definitions[schema] == nil {
				return nil, fmt.Errorf("%w: response schema", ErrInvalidOperation)
			}
		}
		key := operation.Method + "\x00" + operation.Path
		if seenRoutes[key] {
			return nil, fmt.Errorf("%w: duplicate route", ErrInvalidOperation)
		}
		seenRoutes[key] = true
		operation.Responses = cloneResponses(operation.Responses)
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
	return operation, ok
}

func (registry *Registry) Operations() []Operation {
	if registry == nil {
		return nil
	}
	result := make([]Operation, len(registry.operations))
	for index, operation := range registry.operations {
		operation.Responses = cloneResponses(operation.Responses)
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
	schema := operation.Responses[status]
	if schema == "" {
		return ErrInvalidOperation
	}
	return ValidateJSON(schema, data, registry.schemas)
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

func cloneResponses(source map[int]string) map[int]string {
	result := make(map[int]string, len(source))
	for status, schema := range source {
		result[status] = schema
	}
	return result
}

func SortedResponseStatuses(operation Operation) []int {
	result := make([]int, 0, len(operation.Responses))
	for status := range operation.Responses {
		result = append(result, status)
	}
	sort.Ints(result)
	return result
}
