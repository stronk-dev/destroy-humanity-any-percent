-- +goose Up
CREATE TABLE founder_genesis (
    founder_stream_id uuid PRIMARY KEY REFERENCES save_streams(id),
    revision bigint NOT NULL CHECK (revision > 0 AND revision <= 9007199254740991),
    state bytea NOT NULL CHECK (octet_length(state) > 0),
    version integer NOT NULL CHECK (version > 0),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    FOREIGN KEY (constants_hash) REFERENCES catalog_sets(constants_hash)
);

-- Backfill any pre-migration Founder log from the exact pre-command revision
-- named by its first immutable replay envelope. Fail below rather than invent
-- a genesis when retention has already removed that revision.
WITH first_log AS (
    SELECT DISTINCT ON (founder_stream_id)
        founder_stream_id,
        (replay_inputs->'command'->>'revision')::bigint AS revision
    FROM founder_log
    ORDER BY founder_stream_id,seq
)
INSERT INTO founder_genesis(founder_stream_id,revision,state,version,constants_hash)
SELECT first_log.founder_stream_id,first_log.revision,
       convert_to(revision.state::text,'UTF8'),revision.version,revision.constants_hash
FROM first_log
JOIN save_revisions revision
  ON revision.stream_id=first_log.founder_stream_id
 AND revision.revision=first_log.revision;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM founder_log log
        LEFT JOIN founder_genesis genesis USING (founder_stream_id)
        WHERE genesis.founder_stream_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot backfill immutable Founder genesis from retained revisions';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER founder_genesis_immutable
BEFORE UPDATE OR DELETE ON founder_genesis
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose StatementBegin
CREATE FUNCTION require_founder_genesis() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM founder_genesis genesis
        WHERE genesis.founder_stream_id=NEW.founder_stream_id
          AND (NEW.seq <> 1 OR genesis.revision=(NEW.replay_inputs->'command'->>'revision')::bigint)
    ) THEN
        RAISE EXCEPTION 'Founder log requires immutable genesis at its first command revision';
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER founder_log_requires_genesis
AFTER INSERT ON founder_log
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION require_founder_genesis();

-- +goose Down
DROP TRIGGER founder_log_requires_genesis ON founder_log;
DROP FUNCTION require_founder_genesis();
DROP TRIGGER founder_genesis_immutable ON founder_genesis;
DROP TABLE founder_genesis;
