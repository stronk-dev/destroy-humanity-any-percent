-- +goose Up
ALTER TABLE founder_genesis
ADD CONSTRAINT founder_genesis_revision_fkey
FOREIGN KEY (founder_stream_id,revision) REFERENCES save_revisions(stream_id,revision);

ALTER TABLE run_log
ADD CONSTRAINT run_log_source_coordinates_key UNIQUE (company_stream_id,run_seq,seq);

ALTER TABLE founder_log
    ADD COLUMN source_company_stream_id uuid,
    ADD COLUMN source_run_seq bigint,
    ADD COLUMN source_run_log_seq bigint,
    ADD CONSTRAINT founder_log_source_all_or_none CHECK (
        (source_company_stream_id IS NULL AND source_run_seq IS NULL AND source_run_log_seq IS NULL)
        OR
        (source_company_stream_id IS NOT NULL AND source_run_seq IS NOT NULL AND source_run_log_seq IS NOT NULL)
    ),
    ADD CONSTRAINT founder_log_exit_source_shape CHECK (
        ((replay_inputs->'resolved'->>'kind') = 'exit.v1') = (source_company_stream_id IS NOT NULL)
    ),
    ADD CONSTRAINT founder_log_source_run_fkey
        FOREIGN KEY (source_company_stream_id,source_run_seq,source_run_log_seq)
        REFERENCES run_log(company_stream_id,run_seq,seq)
        DEFERRABLE INITIALLY DEFERRED;

-- +goose Down
ALTER TABLE founder_log DROP CONSTRAINT founder_log_source_run_fkey;
ALTER TABLE founder_log DROP CONSTRAINT founder_log_exit_source_shape;
ALTER TABLE founder_log DROP CONSTRAINT founder_log_source_all_or_none;
ALTER TABLE founder_log DROP COLUMN source_run_log_seq;
ALTER TABLE founder_log DROP COLUMN source_run_seq;
ALTER TABLE founder_log DROP COLUMN source_company_stream_id;
ALTER TABLE run_log DROP CONSTRAINT run_log_source_coordinates_key;
ALTER TABLE founder_genesis DROP CONSTRAINT founder_genesis_revision_fkey;
