package achievements

import (
	"fmt"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const AchievementCatalogSchemaName = "AchievementCatalogV1"

// AchievementSchemas exports C18's closed structural v1 descriptor. Semantic
// registry/proof bindings, row ordering, cardinality and total-node limits
// remain mandatory LoadCatalog checks. No recursive/open JSON arm is introduced.
func AchievementSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	enum := func(values ...string) *publicapi.Schema {
		sort.Strings(values)
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
	}
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	array := func(items *publicapi.Schema) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: items}
	}
	ref := func(name string) *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name} }
	conditionName := func(depth int) string { return fmt.Sprintf("AchievementConditionV1Depth%d", depth) }
	definitions := []publicapi.NamedSchema{}
	for depth := maximumConditionDepth; depth >= 1; depth-- {
		arms := []*publicapi.Schema{
			object(field("kind", enum(string(ConditionFactPresent))), field("fact_kind", id())),
			object(field("kind", enum(string(ConditionCounterAtLeast))), field("counter", id()), field("minimum", integer(0, decimal.MaxExactInteger))),
			object(field("kind", enum(string(ConditionExitCountAtLeast))), field("count", integer(1, decimal.MaxExactInteger))),
			object(field("kind", enum(string(ConditionOwnsGeneratorAtLeast))), field("generator_id", id()), field("count", integer(1, decimal.MaxExactInteger))),
		}
		if depth < maximumConditionDepth {
			arms = append(arms, object(field("kind", enum(string(ConditionAllOf))), field("conditions", array(ref(conditionName(depth+1))))))
		}
		definitions = append(definitions, publicapi.NamedSchema{Name: conditionName(depth), Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: arms}})
	}
	definitions = append(definitions,
		publicapi.NamedSchema{Name: "AchievementProofV1", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			object(field("kind", enum(string(ProofProvenance))), field("event_kinds", array(id()))),
			object(field("kind", enum(string(ProofBurn))), field("event_kind", id()), field("resource_id", id()), field("minimum", &publicapi.Schema{Kind: publicapi.SchemaString})),
			object(field("kind", enum(string(ProofPossession))), field("justification_copy_key", id())),
		}}},
		publicapi.NamedSchema{Name: "AchievementDefinitionV1", Schema: object(
			field("id", id()), field("condition_scope", enum(string(ScopeCareer), string(ScopeRun))),
			field("condition", ref(conditionName(1))), field("proof", ref("AchievementProofV1")),
			field("score_grant", integer(1, decimal.MaxExactInteger)), field("copy_key", id()),
		)},
		publicapi.NamedSchema{Name: AchievementCatalogSchemaName, Schema: object(
			field("schema_version", integer(CatalogSchemaVersion, CatalogSchemaVersion)),
			field("achievements", array(ref("AchievementDefinitionV1"))),
		)},
	)
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
}
