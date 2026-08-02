-- +goose Up
ALTER TABLE guild_clearing_results
    ADD COLUMN membership_id uuid REFERENCES guild_members(membership_id);

CREATE INDEX guild_clearing_results_membership_pending
    ON guild_clearing_results(membership_id,founder_id,company_stream_id,run_seq,guild_id,boundary_seq)
    WHERE membership_id IS NOT NULL;

-- Existing results predate membership-period attribution and remain
-- deliberately unclaimable. Timestamp backfills cannot distinguish a
-- leave/rejoin from a boundary committed in the same millisecond.

-- +goose Down
DROP INDEX guild_clearing_results_membership_pending;
ALTER TABLE guild_clearing_results DROP COLUMN membership_id;
