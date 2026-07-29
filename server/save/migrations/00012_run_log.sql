-- +goose Up
CREATE TABLE run_log (
    company_stream_id uuid NOT NULL REFERENCES save_streams(id),
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    run_id text GENERATED ALWAYS AS (company_stream_id::text || ':' || run_seq::text) STORED,
    seq bigint NOT NULL CHECK (seq > 0 AND seq <= 9007199254740991),
    intent_id uuid NOT NULL,
    canonical_payload bytea NOT NULL CHECK (octet_length(canonical_payload) > 1),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt) = 'object'),
    applied_revision bigint CHECK (applied_revision > 0),
    server_ts_ms bigint NOT NULL CHECK (server_ts_ms > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (run_id, seq),
    UNIQUE (company_stream_id, intent_id)
);

CREATE INDEX run_log_stream_run_idx ON run_log(company_stream_id, run_seq, seq);

-- +goose Down
DROP TABLE run_log;
