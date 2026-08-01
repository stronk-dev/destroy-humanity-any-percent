-- +goose Up
ALTER TABLE events DROP CONSTRAINT events_schema_version_check;
ALTER TABLE events ADD CONSTRAINT events_schema_version_check CHECK (
    (kind='run_ended' AND schema_version IN (1,2)) OR
    (kind<>'run_ended' AND schema_version=1)
);

-- +goose Down
ALTER TABLE events DROP CONSTRAINT events_schema_version_check;
ALTER TABLE events ADD CONSTRAINT events_schema_version_check CHECK (schema_version=1);
