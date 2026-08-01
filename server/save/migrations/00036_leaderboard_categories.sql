-- +goose Up
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_pkey;
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_event_id_key;
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_variables_check;
ALTER TABLE verified_runs ADD PRIMARY KEY (run_id,category_id);
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_variables_check CHECK (
    jsonb_typeof(variables) = 'object' AND
    variables ?& ARRAY['commons','advisor','glitched','faction'] AND
    (variables - 'commons' - 'advisor' - 'glitched' - 'faction') = '{}'::jsonb AND
    jsonb_typeof(variables->'commons') = 'boolean' AND
    jsonb_typeof(variables->'advisor') = 'boolean' AND
    jsonb_typeof(variables->'glitched') = 'boolean' AND
    (jsonb_typeof(variables->'faction') = 'null' OR (
        jsonb_typeof(variables->'faction') = 'string' AND
        (variables->>'faction') ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'
    ))
);

CREATE TRIGGER verification_projection_events_immutable
BEFORE UPDATE OR DELETE ON verification_projection_events
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER verification_projection_events_immutable ON verification_projection_events;
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_variables_check;
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_pkey;
ALTER TABLE verified_runs ADD PRIMARY KEY (run_id);
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_event_id_key UNIQUE (event_id);
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_variables_check CHECK (
    jsonb_typeof(variables) = 'object' AND
    variables ?& ARRAY['commons','advisor','glitched'] AND
    (variables - 'commons' - 'advisor' - 'glitched') = '{}'::jsonb AND
    jsonb_typeof(variables->'commons') = 'boolean' AND
    jsonb_typeof(variables->'advisor') = 'boolean' AND
    jsonb_typeof(variables->'glitched') = 'boolean'
);
