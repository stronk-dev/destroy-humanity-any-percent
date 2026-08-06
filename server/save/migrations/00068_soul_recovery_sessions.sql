-- +goose Up
CREATE TABLE soul_recovery_sessions (
    session_id uuid PRIMARY KEY,
    founder_id uuid NOT NULL REFERENCES account_founders(founder_id) ON DELETE CASCADE,
    founder_stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id) ON DELETE CASCADE,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    constants_hash text NOT NULL CHECK (constants_hash ~ '^sha256:[0-9a-f]{64}$'),
    activity_id text NOT NULL CHECK (activity_id ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    founder_attended_start_ms bigint NOT NULL CHECK (founder_attended_start_ms >= 0 AND founder_attended_start_ms <= 9007199254740991),
    required_duration_ms bigint NOT NULL CHECK (required_duration_ms > 0 AND required_duration_ms <= 9007199254740991),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','claimed','resolved','cancelled')),
    start_request_hash text NOT NULL CHECK (start_request_hash ~ '^sha256:[0-9a-f]{64}$'),
    terminal_request_hash text CHECK (terminal_request_hash IS NULL OR terminal_request_hash ~ '^sha256:[0-9a-f]{64}$'),
    claim_token uuid,
    claimed_at timestamptz,
    terminal_receipt jsonb CHECK (terminal_receipt IS NULL OR jsonb_typeof(terminal_receipt)='object'),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    terminal_at timestamptz,
    FOREIGN KEY (company_stream_id,run_seq) REFERENCES run_epochs(company_stream_id,run_seq),
    CHECK ((claim_token IS NULL) = (claimed_at IS NULL)),
    CHECK ((status='claimed') = (claim_token IS NOT NULL)),
    CHECK ((status IN ('resolved','cancelled')) = (terminal_receipt IS NOT NULL)),
    CHECK ((status IN ('claimed','resolved','cancelled')) = (terminal_request_hash IS NOT NULL)),
    CHECK ((status IN ('resolved','cancelled')) = (terminal_at IS NOT NULL))
);

CREATE UNIQUE INDEX soul_recovery_one_active_founder_idx
    ON soul_recovery_sessions(founder_id) WHERE status IN ('active','claimed');

-- +goose StatementBegin
CREATE FUNCTION enforce_soul_recovery_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status IN ('resolved','cancelled') THEN
        RAISE EXCEPTION 'terminal soul recovery session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.founder_id <> OLD.founder_id OR
       NEW.founder_stream_id <> OLD.founder_stream_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.activity_id <> OLD.activity_id OR NEW.founder_attended_start_ms <> OLD.founder_attended_start_ms OR
       NEW.required_duration_ms <> OLD.required_duration_ms OR NEW.start_request_hash <> OLD.start_request_hash OR
       NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'soul recovery session identity is immutable';
    END IF;
    IF NOT ((OLD.status='active' AND NEW.status='claimed') OR
            (OLD.status='claimed' AND NEW.status IN ('claimed','active','resolved','cancelled'))) THEN
        RAISE EXCEPTION 'invalid soul recovery session transition';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER soul_recovery_transition_guard
BEFORE UPDATE ON soul_recovery_sessions
FOR EACH ROW EXECUTE FUNCTION enforce_soul_recovery_transition();

ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'upgrade_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered', 'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced', 'incorporated', 'faction_stock_saturated',
    'guild_tithe_accrued', 'guild_activity_evaluated', 'meter_band_changed.v1',
    'achievement_earned.v1', 'pet_care_applied.v1', 'pet_status_changed.v1',
    'minigame_resolved.v1', 'minigame_rating_changed.v1', 'doctrine_picked', 'compute_credit_spent',
    'fiscal_period_harvested.v1', 'fiscal_credit_spent.v1', 'opportunity_spawned.v1',
    'opportunity_expired.v1', 'opportunity_claimed.v1', 'buff_started.v1', 'buff_expired.v1',
    'soul_price_paid.v1', 'soul_band_changed.v1', 'soul_depleted.v1',
    'soul_recovery_started.v1', 'soul_recovery_cancelled.v1', 'soul_recovered.v1'
));

-- +goose Down
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'upgrade_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered', 'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced', 'incorporated', 'faction_stock_saturated',
    'guild_tithe_accrued', 'guild_activity_evaluated', 'meter_band_changed.v1',
    'achievement_earned.v1', 'pet_care_applied.v1', 'pet_status_changed.v1',
    'minigame_resolved.v1', 'minigame_rating_changed.v1', 'doctrine_picked', 'compute_credit_spent',
    'fiscal_period_harvested.v1', 'fiscal_credit_spent.v1', 'opportunity_spawned.v1',
    'opportunity_expired.v1', 'opportunity_claimed.v1', 'buff_started.v1', 'buff_expired.v1'
));
DROP TRIGGER soul_recovery_transition_guard ON soul_recovery_sessions;
DROP FUNCTION enforce_soul_recovery_transition();
DROP INDEX soul_recovery_one_active_founder_idx;
DROP TABLE soul_recovery_sessions;
