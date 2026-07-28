-- +goose Up
CREATE TABLE save_streams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_kind text NOT NULL CHECK (owner_kind IN ('founder', 'guild', 'world')),
    owner_id uuid NOT NULL,
    scope text NOT NULL CHECK (scope IN ('company', 'founder', 'guild', 'world')),
    archived_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (owner_kind, owner_id, scope),
    CHECK (
        (owner_kind = 'founder' AND scope IN ('company', 'founder')) OR
        (owner_kind = 'guild' AND scope = 'guild') OR
        (owner_kind = 'world' AND scope = 'world')
    )
);

CREATE TABLE save_revisions (
    stream_id uuid NOT NULL REFERENCES save_streams(id),
    revision bigint NOT NULL CHECK (revision > 0),
    version integer NOT NULL CHECK (version > 0),
    state jsonb NOT NULL CHECK (jsonb_typeof(state) = 'object'),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (stream_id, revision)
);

-- +goose Down
DROP TABLE save_revisions;
DROP TABLE save_streams;
