package account

import (
	"net/http"

	"cloud-clicker/server/publicapi"
)

// These descriptors register the existing handlers, not browser renewal policy.
// Both operations reuse the account owner's existing two-token DTO.
func sessionAPISchemas() []publicapi.NamedSchema {
	return []publicapi.NamedSchema{
		{Name: "SessionCreateRequest", Schema: apiObject(
			apiField("account_id", apiString("uuid-v7")),
			apiField("recovery_code", apiString("")),
		)},
		{Name: "SessionRefreshRequest", Schema: apiObject(
			apiField("refresh_token", apiString("")),
		)},
	}
}

func sessionAPIOperations() []publicapi.Operation {
	responses := func(unauthorized ...apiErrorPair) []publicapi.Response {
		return []publicapi.Response{
			{Kind: publicapi.ResponseSchema, Status: http.StatusOK, ContentType: publicapi.ContentJSON, SchemaRef: "BootstrapSession"},
			{Kind: publicapi.ResponseSchema, Status: http.StatusBadRequest, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: exactAPIErrorJSON(apiErrorPair{"invalid", "body"})},
			{Kind: publicapi.ResponseSchema, Status: http.StatusUnauthorized, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: exactAPIErrorJSON(unauthorized...)},
			{Kind: publicapi.ResponseSchema, Status: http.StatusTooManyRequests, ContentType: publicapi.ContentJSON, SchemaRef: "APIError", ExactJSON: exactAPIErrorJSON(apiErrorPair{"rate_limited", "ip"})},
		}
	}
	return []publicapi.Operation{
		{ID: "create_session", Method: http.MethodPost, Path: "/api/v1/session", Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthNone,
			Request: "SessionCreateRequest", Responses: responses(apiErrorPair{"unauthorized", "credential"})},
		{ID: "refresh_session", Method: http.MethodPost, Path: "/api/v1/session/refresh", Surface: publicapi.SurfacePrivateV1, Auth: publicapi.AuthNone,
			Request: "SessionRefreshRequest", Responses: responses(apiErrorPair{"unauthorized", "refresh_token"}, apiErrorPair{"refresh_reused", "session_family_revoked"})},
	}
}
