-- +goose Up
ALTER TABLE guild_presence_outbox
    ADD COLUMN active_count bigint,
    ADD COLUMN claim_token uuid,
    ADD COLUMN claimed_until timestamptz;

UPDATE guild_presence_outbox AS outbox
SET active_count=(SELECT count(*) FROM guild_members AS member WHERE member.guild_id=outbox.guild_id AND member.left_at IS NULL);

ALTER TABLE guild_presence_outbox
    ALTER COLUMN active_count SET NOT NULL,
    ADD CONSTRAINT guild_presence_active_count_check CHECK (active_count >= 0),
    ADD CONSTRAINT guild_presence_claim_pair_check CHECK ((claim_token IS NULL)=(claimed_until IS NULL));

-- +goose Down
ALTER TABLE guild_presence_outbox
    DROP CONSTRAINT guild_presence_claim_pair_check,
    DROP CONSTRAINT guild_presence_active_count_check,
    DROP COLUMN claimed_until,
    DROP COLUMN claim_token,
    DROP COLUMN active_count;
