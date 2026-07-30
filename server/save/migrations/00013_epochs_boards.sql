-- +goose Up
CREATE TABLE catalog_sets (
    constants_hash text PRIMARY KEY CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE catalog_artifacts (
    constants_hash text NOT NULL REFERENCES catalog_sets(constants_hash),
    artifact_name text NOT NULL CHECK (artifact_name ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    bytes bytea NOT NULL CHECK (octet_length(bytes) > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (constants_hash, artifact_name)
);

CREATE TABLE epochs (
    epoch_id bigserial PRIMARY KEY,
    name text NOT NULL CHECK (name <> ''),
    started_at timestamptz NOT NULL,
    ended_at timestamptz,
    changelog_ref text NOT NULL CHECK (changelog_ref ~ '^changelog/epoch-[0-9]+\.md$'),
    CHECK (ended_at IS NULL OR ended_at >= started_at)
);

CREATE UNIQUE INDEX epochs_one_current_idx ON epochs((true)) WHERE ended_at IS NULL;

CREATE TABLE epoch_hashes (
    epoch_id bigint NOT NULL REFERENCES epochs(epoch_id),
    constants_hash text NOT NULL REFERENCES catalog_sets(constants_hash),
    accepted_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (epoch_id, constants_hash)
);

CREATE TABLE run_epochs (
    company_stream_id uuid NOT NULL REFERENCES save_streams(id),
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    epoch_id bigint NOT NULL REFERENCES epochs(epoch_id),
    constants_hash text NOT NULL,
    engine_version text NOT NULL CHECK (engine_version ~ '^[0-9]+\.[0-9]+\.[0-9]+$'),
    build_vcs_hash text NOT NULL CHECK (build_vcs_hash <> ''),
    seed text NOT NULL CHECK (seed ~ '^[0-9]+$'),
    pinned_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (epoch_id, constants_hash) REFERENCES epoch_hashes(epoch_id, constants_hash)
);

CREATE TABLE run_log_archive (
    run_id text PRIMARY KEY,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id),
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    terminal_seq bigint NOT NULL CHECK (terminal_seq > 0 AND terminal_seq <= 9007199254740991),
    encoding text NOT NULL CHECK (encoding = 'gzip+json.v1'),
    bytes bytea NOT NULL CHECK (octet_length(bytes) > 0),
    sha256 text NOT NULL CHECK (sha256 ~ '^sha256:[0-9a-f]{64}$'),
    archived_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (company_stream_id, run_seq),
    CHECK (run_id = company_stream_id::text || ':' || run_seq::text)
);

-- Immutable replay/governance evidence: correction means append a new epoch,
-- accepted hash, or verified record, never rewrite historical bytes.
-- +goose StatementBegin
CREATE FUNCTION reject_immutable_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'immutable leaderboard history';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER catalog_sets_immutable BEFORE UPDATE OR DELETE ON catalog_sets FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
CREATE TRIGGER catalog_artifacts_immutable BEFORE UPDATE OR DELETE ON catalog_artifacts FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
CREATE TRIGGER epoch_hashes_immutable BEFORE UPDATE OR DELETE ON epoch_hashes FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
CREATE TRIGGER run_epochs_immutable BEFORE UPDATE OR DELETE ON run_epochs FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
CREATE TRIGGER run_log_archive_immutable BEFORE UPDATE OR DELETE ON run_log_archive FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER run_log_archive_immutable ON run_log_archive;
DROP TRIGGER run_epochs_immutable ON run_epochs;
DROP TRIGGER epoch_hashes_immutable ON epoch_hashes;
DROP TRIGGER catalog_artifacts_immutable ON catalog_artifacts;
DROP TRIGGER catalog_sets_immutable ON catalog_sets;
DROP FUNCTION reject_immutable_change();
DROP TABLE run_log_archive;
DROP TABLE run_epochs;
DROP TABLE epoch_hashes;
DROP INDEX epochs_one_current_idx;
DROP TABLE epochs;
DROP TABLE catalog_artifacts;
DROP TABLE catalog_sets;
