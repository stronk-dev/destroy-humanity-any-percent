package formulas

import (
	"encoding/json"
	"fmt"
	"strconv"

	"cloud-clicker/server/publicapi"
)

// ClientSchemaModule derives the client validator metadata from the same C18
// descriptor. int64 bounds are decimal strings, so TypeScript does not round
// the schema itself through an IEEE-754 JSON number.
func ClientSchemaModule() ([]byte, error) {
	schemas := Schemas()
	_, err := publicapi.ValidateSchemaDefinitions(schemas)
	if err != nil {
		return nil, err
	}
	rows := make([]clientNamedSchema, 0, len(schemas))
	for _, schema := range schemas {
		rows = append(rows, clientNamedSchema{Name: schema.Name, Schema: clientDescriptor(schema.Schema)})
	}
	encoded, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("// Code generated from formulas.Schemas(); DO NOT EDIT.\n\nexport const FORMULA_SCHEMA_NAME = %q as const;\nexport const FORMULA_SCHEMAS = %s as const;\n", SchemaName, encoded)), nil
}

type clientNamedSchema struct {
	Name   string       `json:"name"`
	Schema clientSchema `json:"schema"`
}
type clientField struct {
	Name     string       `json:"name"`
	Required bool         `json:"required"`
	Schema   clientSchema `json:"schema"`
}
type clientSchema struct {
	Kind       publicapi.SchemaKind `json:"kind"`
	Fields     []clientField        `json:"fields,omitempty"`
	Items      *clientSchema        `json:"items,omitempty"`
	Enum       []string             `json:"enum,omitempty"`
	Minimum    *string              `json:"minimum,omitempty"`
	Maximum    *string              `json:"maximum,omitempty"`
	Format     string               `json:"format,omitempty"`
	Ref        string               `json:"ref,omitempty"`
	Alternates []clientSchema       `json:"alternates,omitempty"`
}

func clientDescriptor(schema *publicapi.Schema) clientSchema {
	result := clientSchema{Kind: schema.Kind, Enum: schema.Enum, Format: schema.Format, Ref: schema.Ref}
	for _, field := range schema.Fields {
		result.Fields = append(result.Fields, clientField{Name: field.Name, Required: field.Required, Schema: clientDescriptor(field.Schema)})
	}
	if schema.Items != nil {
		items := clientDescriptor(schema.Items)
		result.Items = &items
	}
	if schema.Minimum != nil {
		value := strconv.FormatInt(*schema.Minimum, 10)
		result.Minimum = &value
	}
	if schema.Maximum != nil {
		value := strconv.FormatInt(*schema.Maximum, 10)
		result.Maximum = &value
	}
	for _, alternate := range schema.Alternates {
		result.Alternates = append(result.Alternates, clientDescriptor(alternate))
	}
	return result
}
