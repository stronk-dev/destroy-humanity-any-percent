-- +goose Up
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

CREATE TABLE guild_projection_events (
    event_id uuid PRIMARY KEY,
    event_kind text NOT NULL CHECK (event_kind IN ('guild_tithe_accrued')),
    projected_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE guild_activity_windows (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    window_start timestamptz NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    xp bigint NOT NULL CHECK (xp >= 0 AND xp <= 9007199254740991),
    PRIMARY KEY (guild_id,window_start,account_id)
);

CREATE TABLE guild_clearing_results (
    guild_id uuid NOT NULL REFERENCES guilds(guild_id),
    boundary_seq bigint NOT NULL CHECK (boundary_seq > 0 AND boundary_seq <= 9007199254740991),
    account_id uuid NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    debit_units bigint NOT NULL CHECK (debit_units >= 0),
    credit_units bigint NOT NULL CHECK (credit_units >= 0),
    allocations jsonb NOT NULL CHECK (jsonb_typeof(allocations)='array'),
    committed_at timestamptz NOT NULL,
    PRIMARY KEY (guild_id,boundary_seq,account_id)
);

-- +goose Down
DROP TABLE guild_clearing_results;
DROP TABLE guild_activity_windows;
DROP TABLE guild_projection_events;
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced',
    'incorporated', 'faction_stock_saturated'
));
