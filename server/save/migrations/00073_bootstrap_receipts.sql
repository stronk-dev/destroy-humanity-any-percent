-- +goose Up
CREATE TABLE bootstrap_receipts (
    request_digest bytea PRIMARY KEY CHECK (octet_length(request_digest) = 32),
    account_id uuid NOT NULL,
    key_id text,
    nonce bytea,
    ciphertext bytea,
    created_at timestamptz NOT NULL,
    refresh_expires_at timestamptz NOT NULL,
    tombstoned_at timestamptz,
    CHECK (refresh_expires_at > created_at),
    CHECK (
        (tombstoned_at IS NULL AND key_id IS NOT NULL AND octet_length(nonce) = 12 AND octet_length(ciphertext) > 16)
        OR
        (tombstoned_at IS NOT NULL AND key_id IS NULL AND nonce IS NULL AND ciphertext IS NULL)
    )
);

CREATE INDEX bootstrap_receipts_expiry_idx
    ON bootstrap_receipts(refresh_expires_at,request_digest) WHERE tombstoned_at IS NULL;

-- +goose StatementBegin
CREATE FUNCTION enforce_bootstrap_receipt_lifecycle() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'bootstrap receipt tombstones are permanent';
    END IF;
    IF OLD.request_digest <> NEW.request_digest
       OR OLD.account_id <> NEW.account_id
       OR OLD.created_at <> NEW.created_at
       OR OLD.refresh_expires_at <> NEW.refresh_expires_at
       OR OLD.tombstoned_at IS NOT NULL
       OR NEW.tombstoned_at IS NULL
       OR NEW.key_id IS NOT NULL
       OR NEW.nonce IS NOT NULL
       OR NEW.ciphertext IS NOT NULL THEN
        RAISE EXCEPTION 'bootstrap receipts permit only live-to-tombstone transition';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER bootstrap_receipt_lifecycle
BEFORE UPDATE OR DELETE ON bootstrap_receipts
FOR EACH ROW EXECUTE FUNCTION enforce_bootstrap_receipt_lifecycle();

-- +goose Down
DROP TRIGGER bootstrap_receipt_lifecycle ON bootstrap_receipts;
DROP FUNCTION enforce_bootstrap_receipt_lifecycle();
DROP TABLE bootstrap_receipts;
