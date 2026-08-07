-- +goose Up
CREATE UNIQUE INDEX minigame_one_active_founder_idx
    ON minigame_sessions(founder_id) WHERE status IN ('active','claimed');

CREATE TABLE minigame_create_receipts (
    founder_id uuid NOT NULL REFERENCES account_founders(founder_id) ON DELETE CASCADE,
    idempotency_key text NOT NULL CHECK (idempotency_key ~ '^[A-Za-z0-9-]{1,64}$'),
    request_hash text NOT NULL CHECK (request_hash ~ '^sha256:[0-9a-f]{64}$'),
    session_id uuid NOT NULL UNIQUE REFERENCES minigame_sessions(session_id) ON DELETE CASCADE,
    response jsonb NOT NULL CHECK (jsonb_typeof(response) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (founder_id,idempotency_key)
);

CREATE TABLE minigame_command_receipts (
    session_id uuid NOT NULL REFERENCES minigame_sessions(session_id) ON DELETE CASCADE,
    command_id text NOT NULL CHECK (command_id ~ '^[A-Za-z0-9-]{1,64}$'),
    request_hash text NOT NULL CHECK (request_hash ~ '^sha256:[0-9a-f]{64}$'),
    response jsonb NOT NULL CHECK (jsonb_typeof(response) = 'object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (session_id,command_id)
);

-- +goose StatementBegin
CREATE FUNCTION enforce_minigame_api_receipt_immutability() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'minigame API receipts are immutable';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER minigame_create_receipt_immutability
BEFORE UPDATE OR DELETE ON minigame_create_receipts
FOR EACH ROW EXECUTE FUNCTION enforce_minigame_api_receipt_immutability();

CREATE TRIGGER minigame_command_receipt_immutability
BEFORE UPDATE OR DELETE ON minigame_command_receipts
FOR EACH ROW EXECUTE FUNCTION enforce_minigame_api_receipt_immutability();

-- +goose Down
DROP TRIGGER minigame_command_receipt_immutability ON minigame_command_receipts;
DROP TRIGGER minigame_create_receipt_immutability ON minigame_create_receipts;
DROP FUNCTION enforce_minigame_api_receipt_immutability();
DROP TABLE minigame_command_receipts;
DROP TABLE minigame_create_receipts;
DROP INDEX minigame_one_active_founder_idx;
