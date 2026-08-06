-- +goose Up
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'resolve_minigame_session', 'soul_recovery')) =
    (source_company_stream_id IS NOT NULL)
);

-- +goose Down
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'resolve_minigame_session')) =
    (source_company_stream_id IS NOT NULL)
);
