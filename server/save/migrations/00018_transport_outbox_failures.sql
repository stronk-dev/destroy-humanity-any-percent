-- +goose Up
ALTER TABLE transport_receipt_outbox
    ADD COLUMN attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0 AND attempt_count <= 1000),
    ADD COLUMN last_error text CHECK (last_error IS NULL OR char_length(last_error) BETWEEN 1 AND 512),
    ADD COLUMN dead_lettered_at timestamptz,
    ADD CONSTRAINT transport_receipt_size CHECK (octet_length(receipt::text) <= 61440),
    ADD CONSTRAINT transport_receipt_terminal CHECK (published_at IS NULL OR dead_lettered_at IS NULL);

DROP INDEX transport_receipt_outbox_pending_idx;
CREATE INDEX transport_receipt_outbox_pending_idx
    ON transport_receipt_outbox (outbox_id)
    WHERE published_at IS NULL AND dead_lettered_at IS NULL;

-- +goose Down
DROP INDEX transport_receipt_outbox_pending_idx;
ALTER TABLE transport_receipt_outbox
    DROP CONSTRAINT transport_receipt_terminal,
    DROP CONSTRAINT transport_receipt_size,
    DROP COLUMN dead_lettered_at,
    DROP COLUMN last_error,
    DROP COLUMN attempt_count;
CREATE INDEX transport_receipt_outbox_pending_idx
    ON transport_receipt_outbox (outbox_id)
    WHERE published_at IS NULL;
