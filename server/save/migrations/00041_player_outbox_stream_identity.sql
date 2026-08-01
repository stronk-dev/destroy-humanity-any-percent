-- +goose Up
ALTER TABLE transport_player_outbox
    DROP CONSTRAINT IF EXISTS transport_player_outbox_message_kind_source_id_key;
ALTER TABLE transport_player_outbox
    DROP CONSTRAINT IF EXISTS transport_player_outbox_message_stream_source_key;
ALTER TABLE transport_player_outbox
    ADD CONSTRAINT transport_player_outbox_message_stream_source_key
    UNIQUE (message_kind,stream_id,source_id);

-- +goose Down
ALTER TABLE transport_player_outbox
    DROP CONSTRAINT IF EXISTS transport_player_outbox_message_stream_source_key;
ALTER TABLE transport_player_outbox
    ADD CONSTRAINT transport_player_outbox_message_stream_source_key
    UNIQUE (message_kind,stream_id,source_id);
