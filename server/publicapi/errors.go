package publicapi

// APIErrorSchema is the single API-wide typed-rejection descriptor (A4/A5).
// Public and private registries share it without making public readers import
// the private account repository. Exact operation/status pairs remain owned by
// each operation; the two enums alone intentionally are not a pair oracle.
func APIErrorSchema() NamedSchema {
	return NamedSchema{Name: "APIError", Schema: &Schema{Kind: SchemaObject, Fields: []Field{
		{Name: "category", Required: true, Schema: &Schema{Kind: SchemaString, Enum: []string{
			"conflict", "idempotency_conflict", "internal_invariant", "invalid", "not_configured", "not_eligible", "rate_limited", "refresh_reused", "unauthorized", "unknown_id",
		}}},
		{Name: "detail", Required: true, Schema: &Schema{Kind: SchemaString, Enum: []string{
			"access_token", "account", "account_create", "body", "bootstrap", "bootstrap_expired", "category", "company_stream", "credential", "curriculum_exit_required", "cursor", "duplicate_card", "epoch", "exclusive_activity", "fiscal_unlock_required", "founder", "founder_create", "founder_state", "game_ui_snapshot", "garden", "hack_slots_full", "hand_too_large", "human_content_locked", "illegal_phase", "insufficient_currency", "invalid_assist_level", "invalid_text", "ip", "limit", "line_too_long", "mandate", "minigame_api", "minigame_command", "minigame_create", "minigame_revision", "minigame_session", "minigame_tenant", "public_api", "recovery_progress", "recovery_session", "recovery_token", "refresh_token", "run", "session_family_revoked", "session_id", "soul_recovery", "soul_recovery_cancel", "soul_recovery_not_ready", "soul_recovery_progress", "soul_recovery_resolve", "soul_recovery_start", "tier_required", "unknown_card", "unknown_offer", "variables",
		}}},
	}}}
}
