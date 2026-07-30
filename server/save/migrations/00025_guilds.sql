-- +goose Up
CREATE TABLE guilds (
    guild_id uuid PRIMARY KEY CHECK (guild_id::text ~ '^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'),
    name text NOT NULL CHECK (char_length(name) BETWEEN 3 AND 24),
    created_at timestamptz NOT NULL,
    founder_account uuid NOT NULL REFERENCES accounts(account_id),
    join_policy text NOT NULL CHECK (join_policy IN ('open','invite','apply')),
    revision bigint NOT NULL CHECK (revision > 0 AND revision <= 9007199254740991),
    guild_xp bigint NOT NULL DEFAULT 0 CHECK (guild_xp >= 0),
    below_floor_since timestamptz,
    disbanded_at timestamptz
);
CREATE UNIQUE INDEX guilds_active_name_idx ON guilds (lower(name)) WHERE disbanded_at IS NULL;

CREATE TABLE guild_members (
    membership_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    account_id uuid NOT NULL REFERENCES accounts(account_id),
    joined_at timestamptz NOT NULL,
    left_at timestamptz,
    role text NOT NULL CHECK (role IN ('leader','officer','member')),
    CHECK (left_at IS NULL OR left_at >= joined_at)
);
CREATE UNIQUE INDEX guild_members_one_active_account_idx ON guild_members (account_id) WHERE left_at IS NULL;
CREATE UNIQUE INDEX guild_members_one_active_leader_idx ON guild_members (guild_id) WHERE left_at IS NULL AND role='leader';
CREATE INDEX guild_members_active_guild_idx ON guild_members (guild_id, account_id) WHERE left_at IS NULL;

-- +goose StatementBegin
CREATE FUNCTION protect_guild_membership_history() RETURNS trigger LANGUAGE plpgsql AS $$
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
CREATE TRIGGER guild_membership_history_guard BEFORE UPDATE OR DELETE ON guild_members
    FOR EACH ROW EXECUTE FUNCTION protect_guild_membership_history();

CREATE TABLE guild_applications (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    account_id uuid NOT NULL REFERENCES accounts(account_id),
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    admitted boolean,
    PRIMARY KEY (guild_id, account_id, created_at),
    CHECK ((resolved_at IS NULL) = (admitted IS NULL))
);
CREATE UNIQUE INDEX guild_applications_one_pending_idx ON guild_applications(guild_id,account_id) WHERE resolved_at IS NULL;

CREATE TABLE guild_invitations (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    account_id uuid NOT NULL REFERENCES accounts(account_id),
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    accepted boolean,
    PRIMARY KEY (guild_id, account_id, created_at),
    CHECK ((resolved_at IS NULL) = (accepted IS NULL))
);
CREATE UNIQUE INDEX guild_invitations_one_pending_idx ON guild_invitations(guild_id,account_id) WHERE resolved_at IS NULL;

CREATE TABLE guild_account_revisions (
    account_id uuid PRIMARY KEY REFERENCES accounts(account_id) ON DELETE CASCADE,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0 AND revision <= 9007199254740991)
);

CREATE TABLE guild_intent_records (
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    intent_id uuid NOT NULL,
    request_hash text NOT NULL CHECK (request_hash ~ '^sha256:[0-9a-f]{64}$'),
    outcome text NOT NULL CHECK (outcome IN ('applied','rejected')),
    receipt jsonb NOT NULL CHECK (jsonb_typeof(receipt)='object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (account_id,intent_id)
);

CREATE TABLE guild_events (
    event_seq bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    revision bigint NOT NULL CHECK (revision > 0),
    kind text NOT NULL CHECK (kind IN ('guild_created','member_joined','member_left','role_changed','guild_disbanded','exchange_cleared','guild_xp_accrued')),
    actor_account uuid REFERENCES accounts(account_id),
    subject_account uuid REFERENCES accounts(account_id),
    intent_id uuid,
    occurred_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload)='object'),
    UNIQUE (guild_id,revision,kind,subject_account)
);

CREATE TABLE guild_health_inputs (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    window_start timestamptz NOT NULL,
    active_founders bigint NOT NULL CHECK (active_founders >= 0),
    tithed_xp bigint NOT NULL CHECK (tithed_xp >= 0),
    PRIMARY KEY (guild_id,window_start)
);

CREATE TABLE guild_exchange_boundaries (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    boundary_id text NOT NULL,
    committed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (guild_id,boundary_id)
);

-- Membership changes are durable before any socket publication. A successor
-- relay claims these rows and publishes presence on guild:{id}.
CREATE TABLE guild_presence_outbox (
    outbox_id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    account_id uuid NOT NULL REFERENCES accounts(account_id),
    kind text NOT NULL CHECK (kind IN ('joined','left')),
    guild_revision bigint NOT NULL CHECK (guild_revision > 0),
    occurred_at timestamptz NOT NULL,
    published_at timestamptz
);
CREATE INDEX guild_presence_outbox_pending_idx ON guild_presence_outbox(outbox_id) WHERE published_at IS NULL;

-- +goose Down
DROP TABLE guild_presence_outbox;
DROP TABLE guild_exchange_boundaries;
DROP TABLE guild_health_inputs;
DROP TABLE guild_events;
DROP TABLE guild_intent_records;
DROP TABLE guild_account_revisions;
DROP TABLE guild_invitations;
DROP TABLE guild_applications;
DROP TRIGGER guild_membership_history_guard ON guild_members;
DROP FUNCTION protect_guild_membership_history;
DROP TABLE guild_members;
DROP TABLE guilds;
