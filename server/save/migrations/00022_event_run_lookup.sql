-- +goose Up
CREATE INDEX events_stream_kind_run_seq_idx
    ON events(stream_id, kind, (payload->>'run_seq'));

-- +goose Down
DROP INDEX events_stream_kind_run_seq_idx;
