-- +goose Up
ALTER TABLE commons_member_samples ADD COLUMN run_seq bigint;
UPDATE commons_member_samples s
SET run_seq=m.run_seq
FROM company_compact_memberships m
WHERE m.company_stream_id=s.company_stream_id;
ALTER TABLE commons_member_samples ALTER COLUMN run_seq SET NOT NULL;
ALTER TABLE commons_member_samples ADD CONSTRAINT commons_member_samples_run_seq_check CHECK (run_seq > 0);

-- +goose Down
ALTER TABLE commons_member_samples DROP CONSTRAINT commons_member_samples_run_seq_check;
ALTER TABLE commons_member_samples DROP COLUMN run_seq;
