-- +goose Up
CREATE TABLE run_version_drift (
    company_stream_id uuid NOT NULL REFERENCES save_streams(id),
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    observed_version text NOT NULL CHECK (observed_version ~ '^[0-9]+\.[0-9]+\.[0-9]+$'),
    first_seen timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id, run_seq),
    FOREIGN KEY (company_stream_id, run_seq) REFERENCES run_epochs(company_stream_id, run_seq)
);

CREATE TRIGGER run_version_drift_immutable
    BEFORE UPDATE OR DELETE ON run_version_drift
    FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER run_version_drift_immutable ON run_version_drift;
DROP TABLE run_version_drift;
