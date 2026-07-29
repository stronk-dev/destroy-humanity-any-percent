-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left', 'compact_sampled',
    'compact_health_band_changed', 'compact_cascade_started', 'compact_recovered',
    'compact_recruitment_offered'
));

CREATE TABLE commons_recruitment_offers (
    founder_id uuid PRIMARY KEY,
    event_id uuid NOT NULL UNIQUE REFERENCES events(event_id),
    offered_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE commons_recruitment_offers;
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left', 'compact_sampled'
));
