-- +goose Up
ALTER TABLE company_compact_memberships
    ADD COLUMN projected_revision bigint NOT NULL DEFAULT 0 CHECK (projected_revision >= 0);

-- +goose Down
ALTER TABLE company_compact_memberships DROP COLUMN projected_revision;
