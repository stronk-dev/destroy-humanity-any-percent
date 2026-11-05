-- +goose Up
CREATE TABLE exit_knowledge_projection_events (
    event_id uuid PRIMARY KEY REFERENCES events(event_id),
    founder_id uuid NOT NULL,
    amount bigint NOT NULL CHECK (amount BETWEEN 0 AND 9007199254740991)
);
CREATE INDEX exit_knowledge_projection_founder_idx
    ON exit_knowledge_projection_events(founder_id);

-- +goose Down
DROP TABLE exit_knowledge_projection_events;
