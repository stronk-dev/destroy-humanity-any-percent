package account

import (
	"net/http"

	"cloud-clicker/server/publicapi"
)

// Account lifecycle descriptors document the existing handlers. In particular,
// Founder display is currently an exact empty object, not a profile-data map.
func accountAPISchemas() []publicapi.NamedSchema {
	return []publicapi.NamedSchema{
		{Name: "AccountCreated", Schema: apiObject(
			apiField("account_id", apiString("uuid-v7")),
			apiField("created_at", apiString("date-time-utc-ms")),
			apiField("recovery_code", apiString("")),
		)},
		{Name: "AccountEmptyRequest", Schema: apiObject()},
		{Name: "FounderCreated", Schema: apiObject(
			apiField("created_at", apiString("date-time-utc-ms")),
			apiField("id", apiString("uuid-v7")),
			apiField("imported", &publicapi.Schema{Kind: publicapi.SchemaBoolean}),
		)},
		{Name: "FounderProfile", Schema: apiObject(
			apiField("created_at", apiString("date-time-utc-ms")),
			apiField("display", apiObject()),
			apiField("id", apiString("uuid-v7")),
		)},
	}
}

func accountAPIOperations() []publicapi.Operation {
	response := func(status int, schema string, pairs ...apiErrorPair) publicapi.Response {
		return publicapi.Response{Kind: publicapi.ResponseSchema, Status: status, ContentType: publicapi.ContentJSON,
			SchemaRef: schema, ExactJSON: exactAPIErrorJSON(pairs...)}
	}
	return []publicapi.Operation{
		{ID: "create_account", Method: http.MethodPost, Path: "/api/v1/account", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthNone, Request: "AccountEmptyRequest", Responses: []publicapi.Response{
				response(201, "AccountCreated"),
				response(400, "APIError", apiErrorPair{"invalid", "body"}),
				response(429, "APIError", apiErrorPair{"rate_limited", "ip"}),
				response(500, "APIError", apiErrorPair{"internal_invariant", "account_create"}),
			}},
		{ID: "create_founder", Method: http.MethodPost, Path: "/api/v1/founder", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Request: "AccountEmptyRequest", Responses: []publicapi.Response{
				response(201, "FounderCreated"),
				response(400, "APIError", apiErrorPair{"invalid", "body"}),
				response(401, "APIError", apiErrorPair{"unauthorized", "access_token"}),
				response(404, "APIError", apiErrorPair{"unknown_id", "account"}),
				response(429, "APIError", apiErrorPair{"rate_limited", "account"}, apiErrorPair{"rate_limited", "ip"}),
				response(500, "APIError", apiErrorPair{"internal_invariant", "founder_create"}),
			}},
		{ID: "get_founder", Method: http.MethodGet, Path: "/api/v1/founder", Surface: publicapi.SurfacePrivateV1,
			Auth: publicapi.AuthAccessToken, Responses: []publicapi.Response{
				response(200, "FounderProfile"),
				response(401, "APIError", apiErrorPair{"unauthorized", "access_token"}),
				response(404, "APIError", apiErrorPair{"unknown_id", "founder"}),
				response(429, "APIError", apiErrorPair{"rate_limited", "account"}, apiErrorPair{"rate_limited", "ip"}),
			}},
	}
}
