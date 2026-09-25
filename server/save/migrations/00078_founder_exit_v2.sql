-- Reputation Tree v1 R6/R7: the append-only exit.v2 Founder-log arm (exit.v1
-- plus the Exit plan's reputation_purchases) is Company-linked like exit.v1.
-- +goose Up
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'exit.v2', 'resolve_minigame_session', 'soul_recovery')) =
    (source_company_stream_id IS NOT NULL)
);

-- +goose Down
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'resolve_minigame_session', 'soul_recovery')) =
    (source_company_stream_id IS NOT NULL)
);
