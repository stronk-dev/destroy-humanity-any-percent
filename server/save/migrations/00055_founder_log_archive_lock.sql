-- +goose Up
-- The original insert guard in 00054 checked archived_at without locking the
-- parent stream. Replace it forward-only: a direct log insert now conflicts
-- with archival, and the predicate is rechecked after either waiter proceeds.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_founder_log_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM 1
    FROM save_streams
    WHERE id=NEW.founder_stream_id
      AND owner_kind='founder'
      AND scope='founder'
      AND archived_at IS NULL
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'founder log requires an active Founder-scope stream';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- Restore the historical 00054 definition exactly for a reversible rollback.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_founder_log_insert() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM save_streams
        WHERE id=NEW.founder_stream_id AND owner_kind='founder' AND scope='founder' AND archived_at IS NULL
    ) THEN
        RAISE EXCEPTION 'founder log requires an active Founder-scope stream';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
