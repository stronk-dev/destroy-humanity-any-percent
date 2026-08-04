-- +goose Up
CREATE TABLE minigame_sessions (
    session_id uuid PRIMARY KEY,
    minigame_id text NOT NULL CHECK (minigame_id ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    founder_id uuid NOT NULL REFERENCES account_founders(founder_id) ON DELETE CASCADE,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    engine_ref text NOT NULL CHECK (engine_ref ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    engine_version text NOT NULL CHECK (engine_version ~ '^[0-9]+\.[0-9]+\.[0-9]+$'),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    scaling_inputs jsonb NOT NULL CHECK (jsonb_typeof(scaling_inputs) = 'object'),
    seed text NOT NULL CHECK (seed ~ '^(0|[1-9][0-9]{0,19})$' AND seed::numeric <= 18446744073709551615),
    mode text NOT NULL CHECK (mode IN ('solo','async_snapshot')),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','claimed','resolved')),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0 AND revision <= 9007199254740991),
    genesis jsonb NOT NULL CHECK (jsonb_typeof(genesis) = 'object'),
    state jsonb NOT NULL CHECK (jsonb_typeof(state) = 'object'),
    result jsonb CHECK (result IS NULL OR jsonb_typeof(result) = 'object'),
    claim_token uuid,
    claimed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    resolved_at timestamptz,
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES run_epochs(company_stream_id, run_seq),
    CHECK ((claim_token IS NULL) = (claimed_at IS NULL)),
    CHECK ((status = 'claimed') = (claim_token IS NOT NULL)),
    CHECK ((status = 'resolved') = (result IS NOT NULL)),
    CHECK ((status = 'resolved') = (resolved_at IS NOT NULL))
);

CREATE INDEX minigame_sessions_founder_status_idx
    ON minigame_sessions(founder_id,status,created_at,session_id);

-- +goose StatementBegin
CREATE FUNCTION enforce_minigame_session_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status = 'resolved' THEN
        RAISE EXCEPTION 'resolved minigame session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.minigame_id <> OLD.minigame_id OR
       NEW.founder_id <> OLD.founder_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.engine_ref <> OLD.engine_ref OR
       NEW.engine_version <> OLD.engine_version OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.scaling_inputs <> OLD.scaling_inputs OR NEW.seed <> OLD.seed OR NEW.mode <> OLD.mode OR
       NEW.genesis <> OLD.genesis OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'minigame session genesis is immutable';
    END IF;
    IF NOT (
        (OLD.status = 'active' AND NEW.status = 'claimed') OR
        (OLD.status = 'claimed' AND NEW.status IN ('claimed','active','resolved'))
    ) THEN
        RAISE EXCEPTION 'invalid minigame session transition';
    END IF;
    IF NEW.status IN ('active','resolved') AND NEW.revision <> OLD.revision + 1 THEN
        RAISE EXCEPTION 'minigame session revision must advance exactly once';
    END IF;
    IF NEW.status = 'claimed' AND NEW.revision <> OLD.revision THEN
        RAISE EXCEPTION 'claiming cannot advance minigame session revision';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER minigame_session_transition_guard
BEFORE UPDATE ON minigame_sessions
FOR EACH ROW EXECUTE FUNCTION enforce_minigame_session_transition();

-- +goose Down
DROP TRIGGER minigame_session_transition_guard ON minigame_sessions;
DROP FUNCTION enforce_minigame_session_transition();
DROP INDEX minigame_sessions_founder_status_idx;
DROP TABLE minigame_sessions;
