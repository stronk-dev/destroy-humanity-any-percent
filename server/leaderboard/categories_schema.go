package leaderboard

import (
	"math"
	"sort"

	"cloud-clicker/server/publicapi"
)

const CategoryCatalogSchemaName = "CategoryCatalogV1"

// CategorySchemas exports the C18 wire descriptor for the existing v1 catalog.
// The five canonical row arms are closed; this does not introduce arbitrary
// predicates or future category definitions. LoadCategoryCatalog still owns
// array cardinality/order, gate membership and canonical Phase-0 policy values.
func CategorySchemas() []publicapi.NamedSchema {
	any := categoryObject(categoryField("kind", categoryEnum(string(PredicateAny))))
	allGates := categoryObject(categoryField("kind", categoryEnum(string(PredicateAllGates))))
	completion := categoryObject(
		categoryField("kind", categoryEnum(string(PredicateFactsSuperset))),
		categoryField("set_ref", categoryEnum("completion_set")),
	)
	rows := []*publicapi.Schema{
		categoryRow("any_percent", TimerRTA, any),
		categoryRow("ethical_percent", TimerAttended, categoryObject(
			categoryField("kind", categoryEnum(string(PredicateFactsDisjoint))),
			categoryField("set_ref", categoryEnum("forbidden_set")),
		)),
		categoryRow("hundred_percent", TimerRTA, categoryObject(
			categoryField("all", categoryArray(&publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{allGates, completion}})),
			categoryField("kind", categoryEnum(string(PredicateAllOf))),
		)),
		categoryRow("low_percent", TimerRTA, categoryObject(
			categoryField("field", categoryEnum("generators_purchased_total")),
			categoryField("kind", categoryEnum(string(PredicateCountAtMost))),
			categoryField("literal", categoryInteger(0, math.MaxInt64)),
		)),
		categoryRow("valuation", TimerNone, any),
	}
	return []publicapi.NamedSchema{{Name: CategoryCatalogSchemaName, Schema: categoryObject(
		categoryField("categories", categoryArray(&publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: rows})),
		categoryField("fact_sets", categoryObject(
			categoryField("completion_set", categoryArray(&publicapi.Schema{Kind: publicapi.SchemaString})),
			categoryField("forbidden_set", categoryArray(&publicapi.Schema{Kind: publicapi.SchemaString})),
		)),
		categoryField("full_gate_set", categoryArray(&publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"})),
		categoryField("schema_version", categoryInteger(1, 1)),
	)}}
}

func categoryRow(id string, timer CategoryTimer, predicate *publicapi.Schema) *publicapi.Schema {
	return categoryObject(
		categoryField("id", categoryEnum(id)),
		categoryField("name_key", categoryEnum("category."+id)),
		categoryField("predicate", predicate),
		categoryField("timer", categoryEnum(string(timer))),
	)
}

func categoryField(name string, schema *publicapi.Schema) publicapi.Field {
	return publicapi.Field{Name: name, Schema: schema, Required: true}
}

func categoryObject(fields ...publicapi.Field) *publicapi.Schema {
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
}

func categoryEnum(values ...string) *publicapi.Schema {
	sort.Strings(values)
	return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: values}
}

func categoryArray(items *publicapi.Schema) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaArray, Items: items}
}

func categoryInteger(minimum, maximum int64) *publicapi.Schema {
	return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
}
