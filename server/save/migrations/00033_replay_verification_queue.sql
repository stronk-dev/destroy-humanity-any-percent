-- +goose Up
CREATE TABLE verification_queue (
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','claimed','verified','dead')),
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    claimed_at timestamptz,
    completed_at timestamptz,
    verdict text CHECK (verdict IN ('verified','log_gap','state_divergence','constants_mismatch','clock_violation','engine_mismatch')),
    last_error text,
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES run_epochs(company_stream_id, run_seq)
);

CREATE INDEX verification_queue_work_idx ON verification_queue(status,available_at,company_stream_id,run_seq);

CREATE TABLE verification_dead_letters (
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL,
    verdict text NOT NULL CHECK (verdict IN ('log_gap','state_divergence','constants_mismatch','clock_violation','engine_mismatch')),
    detail text NOT NULL CHECK (detail <> ''),
    failed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES verification_queue(company_stream_id, run_seq)
);

CREATE TRIGGER verification_dead_letters_immutable
BEFORE UPDATE OR DELETE ON verification_dead_letters
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER verification_dead_letters_immutable ON verification_dead_letters;
DROP TABLE verification_dead_letters;
DROP INDEX verification_queue_work_idx;
DROP TABLE verification_queue;
