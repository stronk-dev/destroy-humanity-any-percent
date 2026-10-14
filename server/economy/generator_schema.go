package economy

import (
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/publicapi"
)

const GeneratorSchemaName = "EconomyGenerator"

// GeneratorSchemas exports the K2/Purchasable C13 component. Artifact version
// selects V1 (no production), V2 (production), or V4 (typed content); the complete
// artifact descriptor is still separate. Relationships remain owner checks.
func GeneratorSchemas() []publicapi.NamedSchema {
	field := func(name string, schema *publicapi.Schema) publicapi.Field {
		return publicapi.Field{Name: name, Schema: schema, Required: true}
	}
	object := func(fields ...publicapi.Field) *publicapi.Schema {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
		return &publicapi.Schema{Kind: publicapi.SchemaObject, Fields: fields}
	}
	id := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "mechanical-id"}
	}
	value := func() *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Format: "canonical-decimal"}
	}
	positive := func() *publicapi.Schema {
		minimum, maximum := int64(1), int64(decimal.MaxExactInteger)
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimum, Maximum: &maximum}
	}
	enum := func(value string) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaString, Enum: []string{value}}
	}
	ref := func(name string) *publicapi.Schema { return &publicapi.Schema{Kind: publicapi.SchemaRef, Ref: name} }
	price := object(field("resource_id", id()), field("base", value()), field("curve", ref(CostCurveSchemaName)))
	production := object(field("resource_id", id()), field("base_rate", value()))
	roles := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
		object(field("kind", enum(string(RoleProvision))), field("generator_id", id())),
		object(field("kind", enum(string(RoleSynergyFeed))), field("pool_id", id())),
		object(field("kind", enum(string(RoleManualOutput))), field("action_id", id()), field("per_purchased_ppm", positive())),
		object(field("kind", enum(string(RoleStockRate))), field("per_purchased_ppm", positive())),
	}}
	fields := func() []publicapi.Field {
		return []publicapi.Field{field("id", id()), field("price", ref("EconomyPriceV1"))}
	}
	legacy := object(fields()...)
	producingFields := append(fields(), field("production", ref("EconomyProductionV1")))
	producing := object(producingFields...)
	minimumTier, maximumTier := int64(0), int64(3)
	contentFields := append(fields(), field("production", ref("EconomyProductionV1")),
		field("tier", &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: &minimumTier, Maximum: &maximumTier}), field("category", id()),
		field("ladder", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: object(field("purchased_at", positive()), field("multiplier_ppm", positive()))}),
		field("roles", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: ref("EconomyGeneratorRoleV1")}))
	for _, optional := range []publicapi.Field{
		field("provisions", object(field("generator_id", id()), field("rate_ppm", positive()))),
		field("provisioned_hardcap", object(field("count", positive()), field("reason_key", id()))),
	} {
		optional.Required = false
		contentFields = append(contentFields, optional)
	}
	definitions := []publicapi.NamedSchema{
		{Name: GeneratorSchemaName, Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{ref("EconomyGeneratorV1"), ref("EconomyGeneratorV2"), ref("EconomyGeneratorV4")}}},
		{Name: "EconomyGeneratorV1", Schema: legacy}, {Name: "EconomyGeneratorV2", Schema: producing}, {Name: "EconomyGeneratorV4", Schema: object(contentFields...)},
		{Name: "EconomyPriceV1", Schema: price}, {Name: "EconomyProductionV1", Schema: production}, {Name: "EconomyGeneratorRoleV1", Schema: roles},
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	return definitions
}
