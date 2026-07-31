-- +goose Up
-- Closed membership rows remain immutable except for FK anonymization during
-- account deletion. The prior trigger accidentally rejected that cascade.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION protect_guild_membership_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'guild membership history is append-only';
    END IF;
    IF NEW.membership_id <> OLD.membership_id OR NEW.guild_id <> OLD.guild_id OR
       NEW.joined_at <> OLD.joined_at THEN
        RAISE EXCEPTION 'guild membership identity/history is immutable';
    END IF;
    IF OLD.left_at IS NOT NULL THEN
        IF NOT (OLD.account_id IS NOT NULL AND NEW.account_id IS NULL AND
                NEW.left_at IS NOT DISTINCT FROM OLD.left_at AND
                NEW.role IS NOT DISTINCT FROM OLD.role) THEN
            RAISE EXCEPTION 'closed guild membership history is immutable';
        END IF;
        RETURN NEW;
    END IF;
    IF NEW.left_at IS NOT NULL AND NEW.left_at < OLD.joined_at OR
       NEW.account_id IS DISTINCT FROM OLD.account_id AND
       NOT (OLD.account_id IS NOT NULL AND NEW.account_id IS NULL AND NEW.left_at IS NOT NULL) THEN
        RAISE EXCEPTION 'guild membership identity/history is immutable';
    END IF;
    RETURN NEW;
END $$;
-- +goose StatementEnd

ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced',
    'incorporated', 'faction_stock_saturated', 'guild_tithe_accrued',
    'guild_activity_evaluated'
));

ALTER TABLE guild_projection_events DROP CONSTRAINT guild_projection_events_event_kind_check;
ALTER TABLE guild_projection_events ADD CONSTRAINT guild_projection_events_event_kind_check
    CHECK (event_kind IN ('guild_tithe_accrued','guild_activity_evaluated'));

-- Preserve the account identity of a deletion-presence message until it has
-- been published, even though the FK-owned account_id is anonymized.
ALTER TABLE guild_presence_outbox ADD COLUMN account_ref uuid;
UPDATE guild_presence_outbox SET account_ref=account_id WHERE account_id IS NOT NULL;

-- A committed sequence is idempotent only for the exact same authoritative
-- member snapshot.
ALTER TABLE guild_clearing_results ADD COLUMN snapshot_hash text;
UPDATE guild_clearing_results SET snapshot_hash='sha256:' || repeat('0',64);
ALTER TABLE guild_clearing_results ALTER COLUMN snapshot_hash SET NOT NULL;
ALTER TABLE guild_clearing_results ADD CONSTRAINT guild_clearing_results_snapshot_hash_check
    CHECK (snapshot_hash ~ '^sha256:[0-9a-f]{64}$');

-- +goose Down
ALTER TABLE guild_clearing_results DROP CONSTRAINT guild_clearing_results_snapshot_hash_check;
ALTER TABLE guild_clearing_results DROP COLUMN snapshot_hash;
ALTER TABLE guild_presence_outbox DROP COLUMN account_ref;
ALTER TABLE guild_projection_events DROP CONSTRAINT guild_projection_events_event_kind_check;
ALTER TABLE guild_projection_events ADD CONSTRAINT guild_projection_events_event_kind_check
    CHECK (event_kind IN ('guild_tithe_accrued'));
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced',
    'incorporated', 'faction_stock_saturated', 'guild_tithe_accrued'
));
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
