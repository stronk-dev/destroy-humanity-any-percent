-- +goose Up
ALTER TABLE account_founders DROP CONSTRAINT account_founders_account_id_fkey;
ALTER TABLE account_founders ALTER COLUMN account_id DROP NOT NULL;
ALTER TABLE account_founders
    ADD CONSTRAINT account_founders_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts(account_id) ON DELETE SET NULL;

-- +goose Down
-- Orphaned Founder history cannot be reattached to a deleted account. A
-- downgrade to the old cascade schema therefore drops only those anonymized
-- ownership rows; save streams remain independently retained.
DELETE FROM account_founders WHERE account_id IS NULL;
ALTER TABLE account_founders DROP CONSTRAINT account_founders_account_id_fkey;
ALTER TABLE account_founders ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE account_founders
    ADD CONSTRAINT account_founders_account_id_fkey
    FOREIGN KEY (account_id) REFERENCES accounts(account_id) ON DELETE CASCADE;
