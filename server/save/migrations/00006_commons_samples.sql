-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left', 'compact_sampled'
));

ALTER TABLE commons_projection_events DROP CONSTRAINT commons_projection_events_kind_check;
ALTER TABLE commons_projection_events ADD CONSTRAINT commons_projection_events_kind_check CHECK (kind IN ('compact_signed', 'compact_left', 'compact_sampled'));

CREATE TABLE commons_member_samples (
    company_stream_id uuid PRIMARY KEY REFERENCES save_streams(id),
    founder_id uuid NOT NULL,
    cohort_id uuid NOT NULL REFERENCES commons_cohorts(cohort_id),
    weight_ppm bigint NOT NULL CHECK (weight_ppm BETWEEN 0 AND 1000000),
    compliance_ppm bigint NOT NULL CHECK (compliance_ppm BETWEEN 0 AND 1000000),
    solidarity_ppm bigint NOT NULL CHECK (solidarity_ppm BETWEEN 0 AND 1000000),
    enclosure text NOT NULL,
    capacity text NOT NULL,
    sampled_ms bigint NOT NULL CHECK (sampled_ms > 0),
    updated_at timestamptz NOT NULL
);

CREATE TABLE commons_health_scopes (
    scope_kind text NOT NULL CHECK (scope_kind IN ('cohort', 'server')),
    scope_id uuid NOT NULL,
    raw_health_ppm bigint NOT NULL CHECK (raw_health_ppm BETWEEN 0 AND 1000000),
    health_ppm bigint NOT NULL CHECK (health_ppm BETWEEN 0 AND 1000000),
    capacity text NOT NULL,
    real_members integer NOT NULL CHECK (real_members >= 0),
    npc_weight_ppm bigint NOT NULL CHECK (npc_weight_ppm >= 0),
    evaluated_at timestamptz NOT NULL,
    PRIMARY KEY (scope_kind, scope_id)
);

-- +goose Down
DROP TABLE commons_health_scopes;
DROP TABLE commons_member_samples;
ALTER TABLE commons_projection_events DROP CONSTRAINT commons_projection_events_kind_check;
ALTER TABLE commons_projection_events ADD CONSTRAINT commons_projection_events_kind_check CHECK (kind IN ('compact_signed', 'compact_left'));
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted',
    'compact_signed', 'compact_left'
));
