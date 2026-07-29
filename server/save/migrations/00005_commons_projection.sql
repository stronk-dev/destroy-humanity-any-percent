-- +goose Up
CREATE TABLE commons_cohorts (
    cohort_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id uuid NOT NULL,
    activity_bracket text NOT NULL CHECK (activity_bracket ~ '^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$'),
    cohort_seq bigint NOT NULL CHECK (cohort_seq > 0),
    member_count integer NOT NULL CHECK (member_count >= 0),
    standing_numerator bigint NOT NULL DEFAULT 0 CHECK (standing_numerator >= 0),
    standing_denominator bigint NOT NULL DEFAULT 0 CHECK (standing_denominator >= 0),
    closed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (server_id, activity_bracket, cohort_seq)
);

CREATE TABLE founder_commons_assignments (
    founder_id uuid PRIMARY KEY,
    server_id uuid NOT NULL,
    activity_bracket text NOT NULL,
    cohort_id uuid NOT NULL REFERENCES commons_cohorts(cohort_id),
    first_signed_at timestamptz NOT NULL,
    last_signed_at timestamptz NOT NULL
);

CREATE TABLE company_compact_memberships (
    company_stream_id uuid PRIMARY KEY REFERENCES save_streams(id),
    founder_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0),
    cohort_id uuid NOT NULL REFERENCES commons_cohorts(cohort_id),
    member boolean NOT NULL,
    tithe_ppm bigint NOT NULL CHECK (tithe_ppm BETWEEN 0 AND 1000000),
    updated_at timestamptz NOT NULL,
    UNIQUE (founder_id, company_stream_id, run_seq)
);

CREATE TABLE commons_projection_events (
    event_id uuid PRIMARY KEY REFERENCES events(event_id),
    kind text NOT NULL CHECK (kind IN ('compact_signed', 'compact_left')),
    founder_id uuid NOT NULL,
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0),
    occurred_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE commons_projection_events;
DROP TABLE company_compact_memberships;
DROP TABLE founder_commons_assignments;
DROP TABLE commons_cohorts;
