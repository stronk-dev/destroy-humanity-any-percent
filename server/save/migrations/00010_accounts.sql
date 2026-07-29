-- +goose Up
CREATE TABLE accounts (
    account_id uuid PRIMARY KEY,
    recovery_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE account_emails (
    account_id uuid PRIMARY KEY REFERENCES accounts(account_id) ON DELETE CASCADE,
    email text NOT NULL UNIQUE,
    verified_at timestamptz
);

CREATE TABLE account_founders (
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    founder_id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    archived_at timestamptz,
    imported boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX account_founders_one_active_idx
    ON account_founders(account_id) WHERE archived_at IS NULL;

CREATE TABLE sessions (
    token_hash bytea PRIMARY KEY,
    family_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    revoked_at timestamptz,
    CHECK (expires_at > created_at)
);

CREATE INDEX sessions_account_idx ON sessions(account_id, family_id);

CREATE TABLE access_tokens (
    jti uuid PRIMARY KEY,
    family_id uuid NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    founder_id uuid NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);

CREATE INDEX access_tokens_family_idx ON access_tokens(family_id);

-- +goose Down
DROP TABLE access_tokens;
DROP TABLE sessions;
DROP TABLE account_founders;
DROP TABLE account_emails;
DROP TABLE accounts;
