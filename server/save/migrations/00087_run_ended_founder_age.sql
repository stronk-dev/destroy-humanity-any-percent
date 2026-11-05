-- Prestige D3/AC7: immutable post-Exit Founder age in run_ended v4.
-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_schema_version_check;
ALTER TABLE events ADD CONSTRAINT events_schema_version_check CHECK (
    (kind='run_ended' AND schema_version IN (1,2,3,4)) OR
    (kind IN ('opportunity_claimed.v1','buff_started.v1','run_started') AND schema_version IN (1,2)) OR
    (kind NOT IN ('run_ended','opportunity_claimed.v1','buff_started.v1','run_started') AND schema_version=1)
);


-- +goose Down
ALTER TABLE events DROP CONSTRAINT events_schema_version_check;
ALTER TABLE events ADD CONSTRAINT events_schema_version_check CHECK (
    (kind='run_ended' AND schema_version IN (1,2,3)) OR
    (kind IN ('opportunity_claimed.v1','buff_started.v1','run_started') AND schema_version IN (1,2)) OR
    (kind NOT IN ('run_ended','opportunity_claimed.v1','buff_started.v1','run_started') AND schema_version=1)
);

