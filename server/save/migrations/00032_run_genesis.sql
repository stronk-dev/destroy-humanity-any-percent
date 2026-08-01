-- +goose Up
CREATE TABLE run_genesis (
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    state bytea NOT NULL CHECK (octet_length(state) > 0),
    version integer NOT NULL CHECK (version > 0),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES run_epochs(company_stream_id, run_seq),
    FOREIGN KEY (constants_hash) REFERENCES catalog_sets(constants_hash)
);

CREATE TRIGGER run_genesis_immutable
BEFORE UPDATE OR DELETE ON run_genesis
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose StatementBegin
CREATE FUNCTION require_run_genesis() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM run_genesis g
        WHERE g.company_stream_id=NEW.company_stream_id AND g.run_seq=NEW.run_seq
    ) THEN
        RAISE EXCEPTION 'run pin requires immutable genesis';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER run_epoch_requires_genesis
AFTER INSERT ON run_epochs
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION require_run_genesis();

-- +goose Down
DROP TRIGGER run_epoch_requires_genesis ON run_epochs;
DROP FUNCTION require_run_genesis();
DROP TRIGGER run_genesis_immutable ON run_genesis;
DROP TABLE run_genesis;
