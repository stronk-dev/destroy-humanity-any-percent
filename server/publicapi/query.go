package publicapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var queryNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var canonicalIntegerPattern = regexp.MustCompile(`^(?:0|-?[1-9][0-9]*)$`)

// InvalidQueryError names the declared query parameter whose value failed its
// descriptor; handlers map it to the operation's typed 400.
type InvalidQueryError struct{ Parameter string }

func (err *InvalidQueryError) Error() string { return "invalid query parameter " + err.Parameter }

func validQueryParameters(operation Operation, definitions map[string]*Schema) bool {
	pathNames := map[string]bool{}
	for _, parameter := range operation.Parameters {
		pathNames[parameter.Name] = true
	}
	hasCursor := false
	last := ""
	for _, parameter := range operation.Query {
		if !queryNamePattern.MatchString(parameter.Name) || parameter.Name <= last || pathNames[parameter.Name] || parameter.Schema == nil ||
			validateSchema(parameter.Schema, definitions, 0) != nil || !scalarQuerySchema(parameter.Schema) {
			return false
		}
		if parameter.Name == "cursor" {
			if parameter.Required || parameter.Schema.Kind != SchemaString {
				return false
			}
			hasCursor = true
		}
		last = parameter.Name
	}
	return hasCursor == (operation.CursorKey != "")
}

func scalarQuerySchema(schema *Schema) bool {
	return schema.Kind == SchemaString || schema.Kind == SchemaInteger || schema.Kind == SchemaBoolean
}

func cloneQueryParameters(source []QueryParameter) []QueryParameter {
	if source == nil {
		return nil
	}
	result := make([]QueryParameter, len(source))
	for index, parameter := range source {
		result[index] = QueryParameter{Name: parameter.Name, Schema: cloneSchema(parameter.Schema), Required: parameter.Required}
	}
	return result
}

// ParseQuery decodes the operation's declared query parameters: each appears at
// most once, integers are strict base-10 (no sign, padding, or leading zero),
// and every value satisfies its descriptor. Undeclared parameters are ignored.
func (registry *Registry) ParseQuery(operationID string, values url.Values) (map[string]any, error) {
	operation, ok := registry.Operation(operationID)
	if !ok {
		return nil, ErrInvalidOperation
	}
	result := map[string]any{}
	for _, parameter := range operation.Query {
		raw, present := values[parameter.Name]
		if !present {
			if parameter.Required {
				return nil, &InvalidQueryError{Parameter: parameter.Name}
			}
			continue
		}
		if len(raw) != 1 {
			return nil, &InvalidQueryError{Parameter: parameter.Name}
		}
		value, err := decodeQueryValue(parameter.Schema, raw[0])
		if err != nil {
			return nil, &InvalidQueryError{Parameter: parameter.Name}
		}
		result[parameter.Name] = value
	}
	return result, nil
}

func decodeQueryValue(schema *Schema, raw string) (any, error) {
	var value any
	var encoded string
	switch schema.Kind {
	case SchemaString:
		if raw == "" {
			return nil, ErrInvalidSchema
		}
		value, encoded = raw, strconv.Quote(raw)
	case SchemaInteger:
		if !canonicalIntegerPattern.MatchString(raw) || strings.HasPrefix(raw, "-") && schema.Minimum != nil && *schema.Minimum >= 0 {
			return nil, ErrInvalidSchema
		}
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, ErrInvalidSchema
		}
		value, encoded = parsed, raw
	case SchemaBoolean:
		if raw != "true" && raw != "false" {
			return nil, ErrInvalidSchema
		}
		value, encoded = raw == "true", raw
	default:
		return nil, ErrInvalidSchema
	}
	if err := validateValue(schema, jsonValue(encoded), nil, 0); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}
	return value, nil
}

// jsonValue decodes one scalar exactly as ValidateJSON would, so query values
// share the descriptor validator with response bodies.
func jsonValue(encoded string) any {
	decoder := json.NewDecoder(bytes.NewReader([]byte(encoded)))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return nil
	}
	return value
}
