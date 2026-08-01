-- +goose Up
-- Events are authoritative history, not optional transport payloads. Receipts
-- remain bounded before commit; an oversized event is allowed into the outbox
-- and is handled by the relay's bounded deterministic dead-letter lane.
ALTER TABLE transport_player_outbox
    DROP CONSTRAINT transport_player_outbox_payload_check;
ALTER TABLE transport_player_outbox
    ADD CONSTRAINT transport_player_outbox_payload_check
    CHECK (
        jsonb_typeof(payload) = 'object'
        AND (message_kind = 'event' OR octet_length(payload::text) <= 61440)
    );

-- Existing queued event payloads are upgraded before the v2 encoder sees them.
UPDATE transport_player_outbox
SET payload = payload || jsonb_build_object(
    'cursor_effect',
    CASE WHEN payload->>'kind' = 'compensation' THEN 'historical' ELSE 'advance' END
)
WHERE message_kind = 'event' AND NOT payload ? 'cursor_effect';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enqueue_transport_player_event() RETURNS trigger LANGUAGE plpgsql AS $$
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
            'cursor_effect',CASE WHEN NEW.kind='compensation' THEN 'historical' ELSE 'advance' END,
            'payload',NEW.payload
        ),
        NEW.occurred_at
    );
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- A rollback is data-preserving: refuse to reinstate the old cap if rows that
-- were legal under this migration would be destroyed or made invalid.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM transport_player_outbox
        WHERE message_kind='event' AND octet_length(payload::text) > 61440
    ) THEN
        RAISE EXCEPTION 'cannot roll back transport event sizing while oversized event rows exist';
    END IF;
END;
$$;
-- +goose StatementEnd

UPDATE transport_player_outbox
SET payload = payload - 'cursor_effect'
WHERE message_kind = 'event';

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enqueue_transport_player_event() RETURNS trigger LANGUAGE plpgsql AS $$
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

ALTER TABLE transport_player_outbox
    DROP CONSTRAINT transport_player_outbox_payload_check;
ALTER TABLE transport_player_outbox
    ADD CONSTRAINT transport_player_outbox_payload_check
    CHECK (jsonb_typeof(payload) = 'object' AND octet_length(payload::text) <= 61440);
