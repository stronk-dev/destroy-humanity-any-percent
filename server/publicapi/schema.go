// Package publicapi owns the additive public HTTP contract and its generated artifacts.
package publicapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"time"

	"cloud-clicker/server/decimal"
)

var ErrInvalidSchema = errors.New("invalid API schema")

const maximumRuntimeSchemaDepth = 64

type SchemaKind string

const (
	SchemaObject  SchemaKind = "object"
	SchemaArray   SchemaKind = "array"
	SchemaString  SchemaKind = "string"
	SchemaInteger SchemaKind = "integer"
	SchemaBoolean SchemaKind = "boolean"
	SchemaNull    SchemaKind = "null"
	SchemaRef     SchemaKind = "ref"
	SchemaOneOf   SchemaKind = "oneOf"
)

type Field struct {
	Name     string
	Schema   *Schema
	Required bool
}

type Schema struct {
	Kind       SchemaKind
	Fields     []Field
	Items      *Schema
	Enum       []string
	Minimum    *int64
	Maximum    *int64
	Format     string
	Ref        string
	Alternates []*Schema
}

type NamedSchema struct {
	Name   string
	Schema *Schema
}

var (
	schemaNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	fieldNamePattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	sha256Pattern     = regexp.MustCompile(`^[0-9a-f]{64}$`)
	uuidPattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func ValidateSchemaDefinitions(definitions []NamedSchema) (map[string]*Schema, error) {
	sources := make(map[string]*Schema, len(definitions))
	last := ""
	for _, definition := range definitions {
		if !schemaNamePattern.MatchString(definition.Name) || definition.Name <= last || definition.Schema == nil {
			return nil, ErrInvalidSchema
		}
		sources[definition.Name], last = definition.Schema, definition.Name
	}
	for _, definition := range definitions {
		if err := validateSchema(definition.Schema, sources, 0); err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidSchema, definition.Name, err)
		}
	}
	if err := validateReferenceGraph(sources); err != nil {
		return nil, err
	}
	for name, schema := range sources {
		if err := validateExpandedSchemaDepth(schema, sources, 0); err != nil {
			return nil, fmt.Errorf("%w: %s: expanded depth", ErrInvalidSchema, name)
		}
	}
	result := make(map[string]*Schema, len(sources))
	for name, schema := range sources {
		result[name] = cloneSchema(schema)
	}
	return result, nil
}

func validateExpandedSchemaDepth(schema *Schema, definitions map[string]*Schema, depth int) error {
	if schema == nil || depth > maximumRuntimeSchemaDepth {
		return ErrInvalidSchema
	}
	if schema.Kind == SchemaRef {
		return validateExpandedSchemaDepth(definitions[schema.Ref], definitions, depth+1)
	}
	for _, field := range schema.Fields {
		if err := validateExpandedSchemaDepth(field.Schema, definitions, depth+1); err != nil {
			return err
		}
	}
	if schema.Items != nil {
		if err := validateExpandedSchemaDepth(schema.Items, definitions, depth+1); err != nil {
			return err
		}
	}
	for _, alternate := range schema.Alternates {
		if err := validateExpandedSchemaDepth(alternate, definitions, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func validateReferenceGraph(definitions map[string]*Schema) error {
	edges := make(map[string][]string, len(definitions))
	for name, schema := range definitions {
		seen := map[string]bool{}
		collectSchemaRefs(schema, seen)
		for target := range seen {
			edges[name] = append(edges[name], target)
		}
		sort.Strings(edges[name])
	}
	state := map[string]uint8{}
	var visit func(string) error
	visit = func(name string) error {
		if state[name] == 1 {
			return fmt.Errorf("%w: reference cycle at %s", ErrInvalidSchema, name)
		}
		if state[name] == 2 {
			return nil
		}
		state[name] = 1
		for _, target := range edges[name] {
			if err := visit(target); err != nil {
				return err
			}
		}
		state[name] = 2
		return nil
	}
	for name := range definitions {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}

func collectSchemaRefs(schema *Schema, result map[string]bool) {
	if schema.Kind == SchemaRef {
		result[schema.Ref] = true
		return
	}
	for _, field := range schema.Fields {
		collectSchemaRefs(field.Schema, result)
	}
	if schema.Items != nil {
		collectSchemaRefs(schema.Items, result)
	}
	for _, alternate := range schema.Alternates {
		collectSchemaRefs(alternate, result)
	}
}

func cloneSchema(source *Schema) *Schema {
	if source == nil {
		return nil
	}
	result := &Schema{Kind: source.Kind, Enum: append([]string(nil), source.Enum...), Format: source.Format, Ref: source.Ref}
	if source.Minimum != nil {
		value := *source.Minimum
		result.Minimum = &value
	}
	if source.Maximum != nil {
		value := *source.Maximum
		result.Maximum = &value
	}
	result.Fields = make([]Field, len(source.Fields))
	for index, field := range source.Fields {
		result.Fields[index] = Field{Name: field.Name, Required: field.Required, Schema: cloneSchema(field.Schema)}
	}
	result.Items = cloneSchema(source.Items)
	result.Alternates = make([]*Schema, len(source.Alternates))
	for index, alternate := range source.Alternates {
		result.Alternates[index] = cloneSchema(alternate)
	}
	return result
}

func validateSchema(schema *Schema, definitions map[string]*Schema, depth int) error {
	if schema == nil || depth > 32 {
		return ErrInvalidSchema
	}
	switch schema.Kind {
	case SchemaObject:
		if schema.Items != nil || len(schema.Enum) != 0 || schema.Minimum != nil || schema.Maximum != nil || schema.Format != "" || schema.Ref != "" || len(schema.Alternates) != 0 {
			return ErrInvalidSchema
		}
		last := ""
		for _, field := range schema.Fields {
			if !fieldNamePattern.MatchString(field.Name) || field.Name <= last || validateSchema(field.Schema, definitions, depth+1) != nil {
				return ErrInvalidSchema
			}
			last = field.Name
		}
	case SchemaArray:
		if schema.Items == nil || len(schema.Fields) != 0 || len(schema.Enum) != 0 || schema.Minimum != nil || schema.Maximum != nil || schema.Format != "" || schema.Ref != "" || len(schema.Alternates) != 0 {
			return ErrInvalidSchema
		}
		return validateSchema(schema.Items, definitions, depth+1)
	case SchemaString:
		if len(schema.Fields) != 0 || schema.Items != nil || schema.Minimum != nil || schema.Maximum != nil || schema.Ref != "" || len(schema.Alternates) != 0 || !validStringFormat(schema.Format) {
			return ErrInvalidSchema
		}
		last := ""
		for _, value := range schema.Enum {
			if value <= last {
				return ErrInvalidSchema
			}
			last = value
		}
	case SchemaInteger:
		if len(schema.Fields) != 0 || schema.Items != nil || len(schema.Enum) != 0 || schema.Format != "" || schema.Ref != "" || len(schema.Alternates) != 0 || schema.Minimum != nil && schema.Maximum != nil && *schema.Minimum > *schema.Maximum {
			return ErrInvalidSchema
		}
	case SchemaBoolean, SchemaNull:
		if len(schema.Fields) != 0 || schema.Items != nil || len(schema.Enum) != 0 || schema.Minimum != nil || schema.Maximum != nil || schema.Format != "" || schema.Ref != "" || len(schema.Alternates) != 0 {
			return ErrInvalidSchema
		}
	case SchemaRef:
		if !schemaNamePattern.MatchString(schema.Ref) || definitions[schema.Ref] == nil || len(schema.Fields) != 0 || schema.Items != nil || len(schema.Enum) != 0 || schema.Minimum != nil || schema.Maximum != nil || schema.Format != "" || len(schema.Alternates) != 0 {
			return ErrInvalidSchema
		}
	case SchemaOneOf:
		if len(schema.Alternates) < 2 || len(schema.Fields) != 0 || schema.Items != nil || len(schema.Enum) != 0 || schema.Minimum != nil || schema.Maximum != nil || schema.Format != "" || schema.Ref != "" {
			return ErrInvalidSchema
		}
		for _, alternate := range schema.Alternates {
			if err := validateSchema(alternate, definitions, depth+1); err != nil {
				return err
			}
		}
	default:
		return ErrInvalidSchema
	}
	return nil
}

func ValidateJSON(schemaName string, data []byte, definitions map[string]*Schema) error {
	schema := definitions[schemaName]
	if schema == nil {
		return ErrInvalidSchema
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("%w: JSON: %v", ErrInvalidSchema, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("%w: trailing JSON", ErrInvalidSchema)
	}
	if err := validateValue(schema, value, definitions, 0); err != nil {
		return fmt.Errorf("%w: value: %v", ErrInvalidSchema, err)
	}
	return nil
}

func validateValue(schema *Schema, value any, definitions map[string]*Schema, depth int) error {
	if schema == nil || depth > maximumRuntimeSchemaDepth {
		return ErrInvalidSchema
	}
	switch schema.Kind {
	case SchemaObject:
		object, ok := value.(map[string]any)
		if !ok || len(object) > len(schema.Fields) {
			return ErrInvalidSchema
		}
		for _, field := range schema.Fields {
			item, present := object[field.Name]
			if !present {
				if field.Required {
					return ErrInvalidSchema
				}
				continue
			}
			if err := validateValue(field.Schema, item, definitions, depth+1); err != nil {
				return err
			}
		}
		for key := range object {
			index := sort.Search(len(schema.Fields), func(index int) bool { return schema.Fields[index].Name >= key })
			if index == len(schema.Fields) || schema.Fields[index].Name != key {
				return ErrInvalidSchema
			}
		}
	case SchemaArray:
		array, ok := value.([]any)
		if !ok {
			return ErrInvalidSchema
		}
		for _, item := range array {
			if err := validateValue(schema.Items, item, definitions, depth+1); err != nil {
				return err
			}
		}
	case SchemaString:
		text, ok := value.(string)
		if !ok || len(schema.Enum) != 0 && !sortedContains(schema.Enum, text) || !matchesFormat(schema.Format, text) {
			return ErrInvalidSchema
		}
	case SchemaInteger:
		number, ok := value.(json.Number)
		if !ok {
			return ErrInvalidSchema
		}
		integer, err := strconv.ParseInt(string(number), 10, 64)
		if err != nil || schema.Minimum != nil && integer < *schema.Minimum || schema.Maximum != nil && integer > *schema.Maximum {
			return ErrInvalidSchema
		}
	case SchemaBoolean:
		if _, ok := value.(bool); !ok {
			return ErrInvalidSchema
		}
	case SchemaNull:
		if value != nil {
			return ErrInvalidSchema
		}
	case SchemaRef:
		return validateValue(definitions[schema.Ref], value, definitions, depth+1)
	case SchemaOneOf:
		matches := 0
		for _, alternate := range schema.Alternates {
			if validateValue(alternate, value, definitions, depth+1) == nil {
				matches++
			}
		}
		if matches != 1 {
			return ErrInvalidSchema
		}
	default:
		return ErrInvalidSchema
	}
	return nil
}

func validStringFormat(format string) bool {
	switch format {
	case "", "sha256", "uuid", "date-time-ms", "canonical-decimal":
		return true
	default:
		return false
	}
}

func matchesFormat(format, value string) bool {
	switch format {
	case "":
		return true
	case "sha256":
		return sha256Pattern.MatchString(value)
	case "uuid":
		return uuidPattern.MatchString(value)
	case "date-time-ms":
		parsed, err := time.Parse("2006-01-02T15:04:05.000Z", value)
		return err == nil && parsed.UTC().Format("2006-01-02T15:04:05.000Z") == value
	case "canonical-decimal":
		_, err := decimal.ParseCanonical(value)
		return err == nil
	default:
		return false
	}
}

func sortedContains(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}
