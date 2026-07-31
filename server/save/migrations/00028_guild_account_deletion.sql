-- +goose Up
ALTER TABLE guilds ALTER COLUMN founder_account DROP NOT NULL;
ALTER TABLE guilds DROP CONSTRAINT guilds_founder_account_fkey;
ALTER TABLE guilds ADD CONSTRAINT guilds_founder_account_fkey FOREIGN KEY(founder_account) REFERENCES accounts(account_id) ON DELETE SET NULL;

ALTER TABLE guild_members ALTER COLUMN account_id DROP NOT NULL;
ALTER TABLE guild_members DROP CONSTRAINT guild_members_account_id_fkey;
ALTER TABLE guild_members ADD CONSTRAINT guild_members_account_id_fkey FOREIGN KEY(account_id) REFERENCES accounts(account_id) ON DELETE SET NULL;

ALTER TABLE guild_events DROP CONSTRAINT guild_events_actor_account_fkey;
ALTER TABLE guild_events ADD CONSTRAINT guild_events_actor_account_fkey FOREIGN KEY(actor_account) REFERENCES accounts(account_id) ON DELETE SET NULL;
ALTER TABLE guild_events DROP CONSTRAINT guild_events_subject_account_fkey;
ALTER TABLE guild_events ADD CONSTRAINT guild_events_subject_account_fkey FOREIGN KEY(subject_account) REFERENCES accounts(account_id) ON DELETE SET NULL;

ALTER TABLE guild_presence_outbox ALTER COLUMN account_id DROP NOT NULL;
ALTER TABLE guild_presence_outbox DROP CONSTRAINT guild_presence_outbox_account_id_fkey;
ALTER TABLE guild_presence_outbox ADD CONSTRAINT guild_presence_outbox_account_id_fkey FOREIGN KEY(account_id) REFERENCES accounts(account_id) ON DELETE SET NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION protect_guild_membership_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN RAISE EXCEPTION 'guild membership history is append-only'; END IF;
    IF NEW.membership_id <> OLD.membership_id OR NEW.guild_id <> OLD.guild_id OR
       NEW.joined_at <> OLD.joined_at OR OLD.left_at IS NOT NULL OR
       (NEW.left_at IS NOT NULL AND NEW.left_at < OLD.joined_at) OR
       NEW.account_id IS DISTINCT FROM OLD.account_id AND NOT (OLD.account_id IS NOT NULL AND NEW.account_id IS NULL AND NEW.left_at IS NOT NULL) THEN
        RAISE EXCEPTION 'guild membership identity/history is immutable';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION protect_guild_membership_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN RAISE EXCEPTION 'guild membership history is append-only'; END IF;
    IF NEW.membership_id <> OLD.membership_id OR NEW.guild_id <> OLD.guild_id OR
       NEW.account_id <> OLD.account_id OR NEW.joined_at <> OLD.joined_at OR
       OLD.left_at IS NOT NULL OR (NEW.left_at IS NOT NULL AND NEW.left_at < OLD.joined_at) THEN
        RAISE EXCEPTION 'guild membership identity/history is immutable';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd
ALTER TABLE guild_presence_outbox DROP CONSTRAINT guild_presence_outbox_account_id_fkey;
ALTER TABLE guild_presence_outbox ADD CONSTRAINT guild_presence_outbox_account_id_fkey FOREIGN KEY(account_id) REFERENCES accounts(account_id);
ALTER TABLE guild_presence_outbox ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE guild_events DROP CONSTRAINT guild_events_subject_account_fkey;
ALTER TABLE guild_events ADD CONSTRAINT guild_events_subject_account_fkey FOREIGN KEY(subject_account) REFERENCES accounts(account_id);
ALTER TABLE guild_events DROP CONSTRAINT guild_events_actor_account_fkey;
ALTER TABLE guild_events ADD CONSTRAINT guild_events_actor_account_fkey FOREIGN KEY(actor_account) REFERENCES accounts(account_id);
ALTER TABLE guild_members DROP CONSTRAINT guild_members_account_id_fkey;
ALTER TABLE guild_members ADD CONSTRAINT guild_members_account_id_fkey FOREIGN KEY(account_id) REFERENCES accounts(account_id);
ALTER TABLE guild_members ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE guilds DROP CONSTRAINT guilds_founder_account_fkey;
ALTER TABLE guilds ADD CONSTRAINT guilds_founder_account_fkey FOREIGN KEY(founder_account) REFERENCES accounts(account_id);
ALTER TABLE guilds ALTER COLUMN founder_account SET NOT NULL;
