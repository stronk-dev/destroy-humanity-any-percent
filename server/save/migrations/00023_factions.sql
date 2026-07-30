-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced',
    'incorporated', 'faction_stock_saturated'
));

ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_variables_check;
ALTER TABLE verified_runs DISABLE TRIGGER verified_runs_immutable;
UPDATE verified_runs SET variables = variables || '{"faction":null}'::jsonb;
ALTER TABLE verified_runs ENABLE TRIGGER verified_runs_immutable;
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_variables_check CHECK (
    jsonb_typeof(variables) = 'object' AND
    variables ?& ARRAY['commons','advisor','glitched','faction'] AND
    (variables - 'commons' - 'advisor' - 'glitched' - 'faction') = '{}'::jsonb AND
    jsonb_typeof(variables->'commons') = 'boolean' AND
    jsonb_typeof(variables->'advisor') = 'boolean' AND
    jsonb_typeof(variables->'glitched') = 'boolean' AND
    (variables->'faction' = 'null'::jsonb OR jsonb_typeof(variables->'faction') = 'string')
);

-- +goose Down
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_variables_check;
ALTER TABLE verified_runs DISABLE TRIGGER verified_runs_immutable;
UPDATE verified_runs SET variables = variables - 'faction';
ALTER TABLE verified_runs ENABLE TRIGGER verified_runs_immutable;
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_variables_check CHECK (
    jsonb_typeof(variables) = 'object' AND
    variables ?& ARRAY['commons','advisor','glitched'] AND
    (variables - 'commons' - 'advisor' - 'glitched') = '{}'::jsonb AND
    jsonb_typeof(variables->'commons') = 'boolean' AND
    jsonb_typeof(variables->'advisor') = 'boolean' AND
    jsonb_typeof(variables->'glitched') = 'boolean'
);

ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced'
));
