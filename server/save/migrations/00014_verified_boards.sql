-- +goose Up
CREATE TABLE verification_projection_events (
    event_id uuid PRIMARY KEY,
    claimed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE verified_runs (
    run_id text PRIMARY KEY,
    event_id uuid NOT NULL UNIQUE REFERENCES verification_projection_events(event_id),
    founder_id uuid NOT NULL,
    category_id text NOT NULL CHECK (category_id ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    variables jsonb NOT NULL CHECK (
        jsonb_typeof(variables) = 'object' AND
        variables ?& ARRAY['commons','advisor','glitched'] AND
        (variables - 'commons' - 'advisor' - 'glitched') = '{}'::jsonb AND
        jsonb_typeof(variables->'commons') = 'boolean' AND
        jsonb_typeof(variables->'advisor') = 'boolean' AND
        jsonb_typeof(variables->'glitched') = 'boolean'
    ),
    epoch_id bigint NOT NULL REFERENCES epochs(epoch_id),
    mandate_level integer NOT NULL CHECK (mandate_level >= 0 AND mandate_level <= 20),
    key_ms bigint,
    key_int bigint,
    verified_at timestamptz NOT NULL,
    world_first boolean NOT NULL DEFAULT false,
    CHECK ((key_ms IS NOT NULL)::int + (key_int IS NOT NULL)::int = 1)
);

CREATE UNIQUE INDEX verified_runs_world_first_idx
    ON verified_runs(category_id,variables,epoch_id) WHERE world_first;
CREATE INDEX verified_runs_time_board_idx
    ON verified_runs(category_id,variables,epoch_id,mandate_level,key_ms,run_id) WHERE key_ms IS NOT NULL;
CREATE INDEX verified_runs_count_board_idx
    ON verified_runs(category_id,variables,epoch_id,mandate_level,key_int DESC,run_id) WHERE key_int IS NOT NULL;
CREATE TRIGGER verified_runs_immutable BEFORE UPDATE OR DELETE ON verified_runs FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER verified_runs_immutable ON verified_runs;
DROP TABLE verified_runs;
DROP TABLE verification_projection_events;
