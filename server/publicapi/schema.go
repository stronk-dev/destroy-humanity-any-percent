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
)

var ErrInvalidSchema = errors.New("invalid API schema")

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
	result := make(map[string]*Schema, len(definitions))
	last := ""
	for _, definition := range definitions {
		if !schemaNamePattern.MatchString(definition.Name) || definition.Name <= last || definition.Schema == nil {
			return nil, ErrInvalidSchema
		}
		result[definition.Name], last = definition.Schema, definition.Name
	}
	for _, definition := range definitions {
		if err := validateSchema(definition.Schema, result, 0); err != nil {
			return nil, fmt.Errorf("%w: %s: %v", ErrInvalidSchema, definition.Name, err)
		}
	}
	return result, nil
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
	if err := validateValue(schema, value, definitions); err != nil {
		return fmt.Errorf("%w: value: %v", ErrInvalidSchema, err)
	}
	return nil
}

func validateValue(schema *Schema, value any, definitions map[string]*Schema) error {
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
			if err := validateValue(field.Schema, item, definitions); err != nil {
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
			if err := validateValue(schema.Items, item, definitions); err != nil {
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
		return validateValue(definitions[schema.Ref], value, definitions)
	case SchemaOneOf:
		matches := 0
		for _, alternate := range schema.Alternates {
			if validateValue(alternate, value, definitions) == nil {
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
		return canonicalDecimalPattern.MatchString(value)
	default:
		return false
	}
}

var canonicalDecimalPattern = regexp.MustCompile(`^(?:0|[1-9][0-9]*(?:\.[0-9]+)?)(?:e[+-]?[0-9]+)?$`)

func sortedContains(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}
