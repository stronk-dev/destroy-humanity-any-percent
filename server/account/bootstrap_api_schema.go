package account

import (
	"net/http"

	"cloud-clicker/server/publicapi"
)

func bootstrapAPISchemas() []publicapi.NamedSchema {
	return []publicapi.NamedSchema{
		{Name: "BootstrapAccount", Schema: apiObject(
			apiField("account_id", apiString("uuid-v7")),
			apiField("created_at", apiString("date-time-ms")),
			apiField("recovery_code", apiString("")),
		)},
		{Name: "BootstrapRequest", Schema: apiObject(
			apiField("idempotency_key", apiString("opaque-id")),
		)},
		{Name: "BootstrapResponse", Schema: apiObject(
			apiField("account", apiRef("BootstrapAccount")),
			apiField("game_ui_snapshot", &publicapi.Schema{Kind: publicapi.SchemaOneOf, Alternates: []*publicapi.Schema{
				apiRef("GameUISnapshot"), apiRef("GameUISnapshotV1"), apiRef("GameUISnapshotV2"),
			}}),
			apiField("session", apiRef("BootstrapSession")),
		)},
		{Name: "BootstrapSession", Schema: apiObject(
			apiField("access_token", apiString("")),
			apiField("refresh_token", apiString("")),
		)},
	}
}

func bootstrapAPIOperations() []publicapi.Operation {
	return []publicapi.Operation{{ID: "create_bootstrap", Method: http.MethodPost, Path: "/api/v1/bootstrap",
		Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthNone, Request: "BootstrapRequest", Responses: []publicapi.Response{
			{Kind: publicapi.ResponseSchema, Status: http.StatusCreated, ContentType: publicapi.ContentJSON, SchemaRef: "BootstrapResponse"},
			{Kind: publicapi.ResponseSchema, Status: http.StatusBadRequest, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
			{Kind: publicapi.ResponseSchema, Status: http.StatusConflict, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
			{Kind: publicapi.ResponseSchema, Status: http.StatusTooManyRequests, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
			{Kind: publicapi.ResponseSchema, Status: http.StatusInternalServerError, ContentType: publicapi.ContentJSON, SchemaRef: "APIError"},
		}}}
}
