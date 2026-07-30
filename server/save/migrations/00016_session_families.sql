-- +goose Up
CREATE TABLE session_families (
    family_id uuid PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz
);

INSERT INTO session_families(family_id,account_id,created_at,revoked_at)
SELECT family_id,account_id,min(created_at),max(revoked_at)
FROM (
    SELECT family_id,account_id,created_at,revoked_at FROM sessions
    UNION ALL
    SELECT family_id,account_id,created_at,revoked_at FROM access_tokens
) existing
GROUP BY family_id,account_id;

ALTER TABLE sessions
    ADD CONSTRAINT sessions_family_fk
    FOREIGN KEY (family_id) REFERENCES session_families(family_id) ON DELETE CASCADE;

ALTER TABLE access_tokens
    ADD CONSTRAINT access_tokens_family_fk
    FOREIGN KEY (family_id) REFERENCES session_families(family_id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE access_tokens DROP CONSTRAINT access_tokens_family_fk;
ALTER TABLE sessions DROP CONSTRAINT sessions_family_fk;
DROP TABLE session_families;
