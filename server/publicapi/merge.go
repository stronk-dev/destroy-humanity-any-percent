package publicapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// MergeRegistries unions independently mounted registries into one generation
// authority (one OpenAPI document, one TypeScript module, one compatibility
// pin). A schema name shared between registries must be the byte-identical
// definition; operation IDs and routes must not collide. Runtime mounting keeps
// using each source registry.
func MergeRegistries(registries ...*Registry) (*Registry, error) {
	definitions := map[string]*Schema{}
	encoded := map[string][]byte{}
	var operations []Operation
	for _, registry := range registries {
		if registry == nil {
			return nil, ErrInvalidOperation
		}
		for _, definition := range registry.Schemas() {
			bytesValue, err := json.Marshal(openAPISchema(definition.Schema))
			if err != nil {
				return nil, err
			}
			if prior, ok := encoded[definition.Name]; ok {
				if !bytes.Equal(prior, bytesValue) {
					return nil, fmt.Errorf("%w: conflicting schema %s", ErrInvalidSchema, definition.Name)
				}
				continue
			}
			encoded[definition.Name], definitions[definition.Name] = bytesValue, definition.Schema
		}
		operations = append(operations, registry.Operations()...)
	}
	schemas := make([]NamedSchema, 0, len(definitions))
	for name, schema := range definitions {
		schemas = append(schemas, NamedSchema{Name: name, Schema: schema})
	}
	sort.Slice(schemas, func(left, right int) bool { return schemas[left].Name < schemas[right].Name })
	sort.Slice(operations, func(left, right int) bool { return operations[left].ID < operations[right].ID })
	return NewRegistry(schemas, operations)
}
