-- +goose Up
CREATE TABLE founder_log (
    founder_stream_id uuid NOT NULL REFERENCES save_streams(id),
    seq bigint NOT NULL CHECK (seq > 0 AND seq <= 9007199254740991),
    intent_id uuid NOT NULL,
    canonical_payload bytea NOT NULL CHECK (octet_length(canonical_payload) > 1),
    replay_inputs jsonb NOT NULL CHECK (jsonb_typeof(replay_inputs) = 'object'),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt) = 'object'),
    applied_revision bigint CHECK (applied_revision > 0 AND applied_revision <= 9007199254740991),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    server_ts_ms bigint NOT NULL CHECK (server_ts_ms > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (founder_stream_id,seq),
    UNIQUE (founder_stream_id,intent_id)
);

-- +goose StatementBegin
CREATE FUNCTION enforce_founder_log_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM save_streams
        WHERE id=NEW.founder_stream_id AND owner_kind='founder' AND scope='founder' AND archived_at IS NULL
    ) THEN
        RAISE EXCEPTION 'founder log requires an active Founder-scope stream';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER founder_log_scope_guard
BEFORE INSERT ON founder_log
FOR EACH ROW EXECUTE FUNCTION enforce_founder_log_insert();

CREATE TRIGGER founder_log_immutable
BEFORE UPDATE OR DELETE ON founder_log
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER founder_log_immutable ON founder_log;
DROP TRIGGER founder_log_scope_guard ON founder_log;
DROP FUNCTION enforce_founder_log_insert();
DROP TABLE founder_log;
