package account

import (
	"net/http"

	"cloud-clicker/server/publicapi"
)

// Server Garden SG9: the one garden read, a NEW operation (allowed widening
// under API Foundation C2; no existing v1 union grows).
func gardenAPISchemas() []publicapi.NamedSchema {
	integer := func(minimum, maximum int64) *publicapi.Schema {
		return &publicapi.Schema{Kind: publicapi.SchemaInteger, Minimum: apiInteger(minimum), Maximum: apiInteger(maximum)}
	}
	nullableInteger := &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{integer(0, apiMaxExactInteger), {Kind: publicapi.SchemaNull}}}
	return []publicapi.NamedSchema{
		{Name: "GardenActive", Schema: apiObject(
			apiField("founder_revision", integer(1, apiMaxExactInteger)),
			apiField("garden", apiRef("GardenView")),
			apiField("kind", apiString("", "active")),
			apiField("server_ms", integer(0, apiMaxExactInteger)),
		)},
		{Name: "GardenCurrentResponse", Schema: &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
			apiRef("GardenActive"), apiRef("GardenInactive"), apiRef("GardenLocked"),
		}}},
		{Name: "GardenInactive", Schema: apiObject(apiField("kind", apiString("", "inactive")))},
		{Name: "GardenLocked", Schema: apiObject(
			apiField("kind", apiString("", "locked")),
			apiField("unlock_id", apiString("mechanical-id")),
		)},
		{Name: "GardenPlot", Schema: apiObject(
			apiField("age_ticks", integer(0, 1_000_000)),
			apiField("col", integer(0, 5)),
			apiField("dormant", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
			apiField("maturation_ticks", integer(1, 1_000_000)),
			apiField("row", integer(0, 5)),
			apiField("species_id", apiString("mechanical-id")),
			apiField("stage", apiString("", "growing", "mature")),
		)},
		{Name: "GardenView", Schema: apiObject(
			apiField("height", integer(1, 6)),
			apiField("max_height", integer(1, 6)),
			apiField("max_width", integer(1, 6)),
			apiField("next_tick_wall_ms", nullableInteger),
			apiField("pending_catchup_forfeited_ms", integer(0, apiMaxExactInteger)),
			apiField("plots", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiRef("GardenPlot")}),
			apiField("seed_collection", &publicapi.Schema{Kind: publicapi.SchemaArray, Items: apiString("mechanical-id")}),
			apiField("species_total", integer(1, 256)),
			apiField("substrate_id", apiString("mechanical-id")),
			apiField("substrate_lockout_until_ms", nullableInteger),
			apiField("tick_seq", integer(0, apiMaxExactInteger)),
			apiField("width", integer(1, 6)),
		)},
	}
}

func gardenAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{{ID: "get_current_garden", Method: http.MethodGet, Path: "/api/v1/garden/current",
		Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthAccessToken, Responses: minigameAPIResponses("GardenCurrentResponse", "")}}
}
