-- +goose Up
ALTER TABLE founder_route_state
    ADD COLUMN route_knowledge_debt bigint NOT NULL DEFAULT 0
    CHECK (route_knowledge_debt BETWEEN 0 AND 9007199254740991);

-- +goose Down
ALTER TABLE founder_route_state DROP COLUMN route_knowledge_debt;
