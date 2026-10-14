package doctrine

import "cloud-clicker/server/publicapi"

const DoctrineCatalogSchemaName = "DoctrineCatalogV1"

// DoctrineSchemas exports D1's closed identity-only catalog grammar for C18.
// Adjacent boundary agreement, ordering, branching and pinned Route bindings
// remain mandatory owner checks; no doctrine effects are added here.
func DoctrineSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	row := &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
		field("doctrine_ids", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: id()}),
		field("gate_id", id()), field("source_tier", integer(0, 8)), field("transition_id", id()),
	}}
	return []publicapi.NamedSchema{{Name: DoctrineCatalogSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: []publicapi.Field{
		field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
		field("transitions", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: row}),
	}}}}
}
