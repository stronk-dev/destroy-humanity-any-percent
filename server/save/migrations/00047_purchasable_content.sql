-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'upgrade_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_tithe_raised', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered',
    'exit_offer_spawned', 'exit_offer_expired', 'exit_offer_declined',
    'run_ended', 'run_started', 'founder_advanced',
    'incorporated', 'faction_stock_saturated', 'guild_tithe_accrued',
    'guild_activity_evaluated'
));

-- +goose Down
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
