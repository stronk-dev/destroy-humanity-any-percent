-- +goose Up
ALTER TABLE verification_queue
    ADD COLUMN claim_token uuid,
    ADD CONSTRAINT verification_queue_claim_pair CHECK ((claim_token IS NULL) = (claimed_at IS NULL));

CREATE TABLE verification_poison_dead_letters (
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    attempts integer NOT NULL CHECK (attempts > 0),
    detail text NOT NULL CHECK (detail <> ''),
    failed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES verification_queue(company_stream_id, run_seq)
);

CREATE TRIGGER verification_poison_dead_letters_immutable
BEFORE UPDATE OR DELETE ON verification_poison_dead_letters
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose StatementBegin
CREATE FUNCTION reject_terminal_verification_queue_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status IN ('verified','dead') THEN
        RAISE EXCEPTION 'terminal verification record is immutable';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER verification_queue_terminal_immutable
BEFORE UPDATE OR DELETE ON verification_queue
FOR EACH ROW EXECUTE FUNCTION reject_terminal_verification_queue_change();

-- +goose Down
DROP TRIGGER verification_queue_terminal_immutable ON verification_queue;
DROP FUNCTION reject_terminal_verification_queue_change();
DROP TRIGGER verification_poison_dead_letters_immutable ON verification_poison_dead_letters;
DROP TABLE verification_poison_dead_letters;
ALTER TABLE verification_queue DROP CONSTRAINT verification_queue_claim_pair;
ALTER TABLE verification_queue DROP COLUMN claim_token;
