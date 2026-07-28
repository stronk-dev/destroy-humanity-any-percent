-- +goose Up
CREATE TABLE intent_records (
    stream_id uuid NOT NULL REFERENCES save_streams(id),
    intent_id uuid NOT NULL,
    request_hash text NOT NULL CHECK (request_hash ~ '^sha256:[0-9a-f]{64}$'),
    outcome text NOT NULL CHECK (outcome IN ('applied', 'rejected')),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt) = 'object'),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (stream_id, intent_id)
);

CREATE INDEX intent_records_created_at_idx ON intent_records (created_at);

CREATE TABLE events (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id uuid NOT NULL REFERENCES save_streams(id),
    revision bigint NOT NULL CHECK (revision > 0),
    schema_version integer NOT NULL CHECK (schema_version = 1),
    kind text NOT NULL CHECK (kind IN ('generator_purchased', 'invariant_reported', 'compensation')),
    intent_id uuid,
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX events_stream_revision_idx ON events (stream_id, revision, event_id);

-- +goose Down
DROP TABLE events;
DROP TABLE intent_records;
