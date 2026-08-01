-- +goose Up
CREATE TABLE transport_player_outbox (
    outbox_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    founder_id uuid NOT NULL,
    stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    message_kind text NOT NULL CHECK (message_kind IN ('receipt','event')),
    source_id uuid NOT NULL,
    scope text NOT NULL CHECK (scope IN ('company','founder')),
    revision bigint NOT NULL CHECK (revision > 0 AND revision <= 9007199254740991),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object' AND octet_length(payload::text) <= 61440),
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    claim_token uuid,
    claimed_until timestamptz,
    published_at timestamptz,
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0 AND attempt_count <= 1000),
    last_error text CHECK (last_error IS NULL OR char_length(last_error) BETWEEN 1 AND 512),
    dead_lettered_at timestamptz,
    UNIQUE (message_kind, source_id),
    CHECK ((claim_token IS NULL) = (claimed_until IS NULL)),
    CHECK (published_at IS NULL OR dead_lettered_at IS NULL)
);

INSERT INTO transport_player_outbox(
    founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload,occurred_at,
    claim_token,claimed_until,published_at,attempt_count,last_error,dead_lettered_at
)
SELECT founder_id,company_stream_id,'receipt',intent_id,'company',revision,constants_hash,receipt,occurred_at,
       claim_token,claimed_until,published_at,attempt_count,last_error,dead_lettered_at
FROM transport_receipt_outbox
ORDER BY outbox_id;

CREATE INDEX transport_player_outbox_pending_idx
    ON transport_player_outbox (outbox_id)
    WHERE published_at IS NULL AND dead_lettered_at IS NULL;

-- +goose StatementBegin
CREATE FUNCTION enqueue_transport_player_event() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    owner uuid;
    stream_scope text;
BEGIN
    SELECT owner_id,scope INTO owner,stream_scope
    FROM save_streams
    WHERE id=NEW.stream_id AND owner_kind='founder' AND scope IN ('company','founder');
    IF NOT FOUND THEN
        RETURN NEW;
    END IF;
    INSERT INTO transport_player_outbox(
        founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload,occurred_at
    ) VALUES (
        owner,NEW.stream_id,'event',NEW.event_id,stream_scope,NEW.revision,NEW.constants_hash,
        jsonb_build_object(
            'event_id',NEW.event_id::text,
            'kind',NEW.kind::text,
            'scope',stream_scope,
            'rev',NEW.revision,
            'payload',NEW.payload
        ),
        NEW.occurred_at
    );
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER events_transport_player_outbox
AFTER INSERT ON events
FOR EACH ROW EXECUTE FUNCTION enqueue_transport_player_event();

DROP TABLE transport_receipt_outbox;

-- +goose Down
CREATE TABLE transport_receipt_outbox (
    outbox_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    founder_id uuid NOT NULL,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    intent_id uuid NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0 AND revision <= 9007199254740991),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt) = 'object' AND octet_length(receipt::text) <= 61440),
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    claim_token uuid,
    claimed_until timestamptz,
    published_at timestamptz,
    attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0 AND attempt_count <= 1000),
    last_error text CHECK (last_error IS NULL OR char_length(last_error) BETWEEN 1 AND 512),
    dead_lettered_at timestamptz,
    UNIQUE (company_stream_id,intent_id),
    CHECK ((claim_token IS NULL) = (claimed_until IS NULL)),
    CHECK (published_at IS NULL OR dead_lettered_at IS NULL)
);
INSERT INTO transport_receipt_outbox(
    founder_id,company_stream_id,intent_id,revision,constants_hash,receipt,occurred_at,
    claim_token,claimed_until,published_at,attempt_count,last_error,dead_lettered_at
)
SELECT founder_id,stream_id,source_id,revision,constants_hash,payload,occurred_at,
       claim_token,claimed_until,published_at,attempt_count,last_error,dead_lettered_at
FROM transport_player_outbox WHERE message_kind='receipt' ORDER BY outbox_id;
CREATE INDEX transport_receipt_outbox_pending_idx
    ON transport_receipt_outbox (outbox_id)
    WHERE published_at IS NULL AND dead_lettered_at IS NULL;
DROP TRIGGER events_transport_player_outbox ON events;
DROP FUNCTION enqueue_transport_player_event();
DROP TABLE transport_player_outbox;
