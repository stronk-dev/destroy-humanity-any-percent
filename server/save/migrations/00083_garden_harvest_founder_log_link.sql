-- Server Garden SG6: the coordinator's garden_harvest_credited Founder-log arm
-- is Company-linked (it carries the credit_garden_harvest run-log coordinate);
-- the Founder-only garden_harvest rejection arm is not.
-- +goose Up
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'exit.v2', 'resolve_minigame_session', 'soul_recovery', 'garden_harvest_credited')) =
    (source_company_stream_id IS NOT NULL)
);

-- +goose Down
ALTER TABLE founder_log DROP CONSTRAINT founder_log_multistream_source_shape;
ALTER TABLE founder_log ADD CONSTRAINT founder_log_multistream_source_shape CHECK (
    ((replay_inputs->'resolved'->>'kind') IN ('exit.v1', 'exit.v2', 'resolve_minigame_session', 'soul_recovery')) =
    (source_company_stream_id IS NOT NULL)
);
