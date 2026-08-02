-- +goose Up
ALTER TABLE commons_member_samples ALTER COLUMN run_seq DROP NOT NULL;

-- Migration 00043 could only infer a sample's run from the current membership
-- row. A later re-sign advances membership.updated_at; samples older than that
-- boundary are deliberately unlabeled so the live resolver uses entry weight.
UPDATE commons_member_samples sample
SET run_seq=NULL
FROM company_compact_memberships membership
WHERE membership.company_stream_id=sample.company_stream_id
  AND sample.updated_at<membership.updated_at;

-- +goose Down
UPDATE commons_member_samples sample
SET run_seq=membership.run_seq
FROM company_compact_memberships membership
WHERE membership.company_stream_id=sample.company_stream_id
  AND sample.run_seq IS NULL;

ALTER TABLE commons_member_samples ALTER COLUMN run_seq SET NOT NULL;
