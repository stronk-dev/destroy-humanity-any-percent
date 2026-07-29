-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation',
    'gate_crossed', 'route_executed', 'route_hint_purchased', 'route_knowledge_granted'
));

CREATE TABLE route_projection_events (
    event_id uuid PRIMARY KEY REFERENCES events(event_id),
    route_id text NOT NULL,
    founder_id uuid NOT NULL,
    company_stream_id uuid NOT NULL REFERENCES save_streams(id),
    run_seq bigint NOT NULL CHECK (run_seq > 0),
    occurred_at timestamptz NOT NULL,
    UNIQUE (founder_id, company_stream_id, run_seq, route_id)
);

CREATE TABLE founder_route_executions (
    founder_id uuid NOT NULL,
    route_id text NOT NULL,
    first_event_id uuid NOT NULL REFERENCES events(event_id),
    first_occurred_at timestamptz NOT NULL,
    last_occurred_at timestamptz NOT NULL,
    execution_count bigint NOT NULL CHECK (execution_count > 0),
    PRIMARY KEY (founder_id, route_id)
);

CREATE TABLE founder_route_state (
    founder_id uuid PRIMARY KEY,
    route_knowledge_balance bigint NOT NULL DEFAULT 0 CHECK (route_knowledge_balance >= 0)
);

CREATE TABLE route_hint_projection_events (
    event_id uuid PRIMARY KEY REFERENCES events(event_id),
    founder_id uuid NOT NULL,
    route_id text NOT NULL,
    cost bigint NOT NULL CHECK (cost > 0)
);

CREATE TABLE registry_routes (
    route_id text PRIMARY KEY,
    first_event_id uuid NOT NULL REFERENCES events(event_id),
    first_founder_id uuid NOT NULL,
    occurred_at timestamptz NOT NULL,
    house_name text NOT NULL,
    name text NOT NULL,
    name_state text NOT NULL CHECK (name_state IN ('reserved', 'pending', 'published', 'house')),
    naming_reserved_until timestamptz NOT NULL,
    execution_count bigint NOT NULL CHECK (execution_count > 0)
);

-- +goose Down
DROP TABLE registry_routes;
DROP TABLE route_hint_projection_events;
DROP TABLE founder_route_state;
DROP TABLE founder_route_executions;
DROP TABLE route_projection_events;
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events ADD CONSTRAINT events_kind_check CHECK (kind IN (
    'generator_purchased', 'invariant_reported', 'compensation'
));
