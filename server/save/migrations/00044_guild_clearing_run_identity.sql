-- +goose Up
ALTER TABLE guild_clearing_results
    ADD COLUMN founder_id uuid,
    ADD COLUMN company_stream_id uuid,
    ADD COLUMN run_seq bigint;

ALTER TABLE guild_clearing_results ADD CONSTRAINT guild_clearing_results_run_identity_check CHECK (
    (founder_id IS NULL AND company_stream_id IS NULL AND run_seq IS NULL) OR
    (founder_id IS NOT NULL AND company_stream_id IS NOT NULL AND run_seq > 0 AND run_seq <= 9007199254740991)
);

CREATE INDEX guild_clearing_results_run_pending
    ON guild_clearing_results(founder_id,company_stream_id,run_seq,guild_id,boundary_seq)
    WHERE founder_id IS NOT NULL;

-- Existing pre-identity rows remain explicitly legacy/unclaimable. Attribution
-- cannot be reconstructed safely after a Founder or run transition.

-- +goose Down
DROP INDEX guild_clearing_results_run_pending;
ALTER TABLE guild_clearing_results DROP CONSTRAINT guild_clearing_results_run_identity_check;
ALTER TABLE guild_clearing_results
    DROP COLUMN run_seq,
    DROP COLUMN company_stream_id,
    DROP COLUMN founder_id;
