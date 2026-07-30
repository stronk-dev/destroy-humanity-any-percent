-- +goose Up
CREATE TABLE transport_receipt_outbox (
    outbox_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    founder_id uuid NOT NULL,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    intent_id uuid NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0 AND revision <= 9007199254740991),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    claim_token uuid,
    claimed_until timestamptz,
    published_at timestamptz,
    UNIQUE (company_stream_id, intent_id),
    CHECK ((claim_token IS NULL) = (claimed_until IS NULL))
);

CREATE INDEX transport_receipt_outbox_pending_idx
    ON transport_receipt_outbox (outbox_id)
    WHERE published_at IS NULL;

-- +goose Down
DROP TABLE transport_receipt_outbox;
