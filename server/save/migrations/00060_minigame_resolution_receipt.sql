-- +goose Up
ALTER TABLE minigame_sessions
    ADD COLUMN resolution_receipt jsonb,
    ADD COLUMN resolution_company_revision bigint,
    ADD COLUMN resolution_founder_revision bigint,
    ADD CONSTRAINT minigame_resolution_receipt_shape CHECK (
        (resolution_receipt IS NULL AND resolution_company_revision IS NULL AND resolution_founder_revision IS NULL) OR
        (jsonb_typeof(resolution_receipt) = 'object' AND
         resolution_company_revision > 0 AND resolution_company_revision <= 9007199254740991 AND
         resolution_founder_revision > 0 AND resolution_founder_revision <= 9007199254740991)
    );

-- Legacy resolved rows predate the composer and remain explicitly
-- distinguishable by an all-null receipt tuple. Every post-migration resolve
-- must write the complete tuple in its claimed→resolved transition.
-- +goose StatementBegin
CREATE FUNCTION enforce_minigame_resolution_receipt() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.resolution_receipt IS NOT NULL OR OLD.resolution_company_revision IS NOT NULL OR
       OLD.resolution_founder_revision IS NOT NULL THEN
        IF NEW.resolution_receipt IS DISTINCT FROM OLD.resolution_receipt OR
           NEW.resolution_company_revision IS DISTINCT FROM OLD.resolution_company_revision OR
           NEW.resolution_founder_revision IS DISTINCT FROM OLD.resolution_founder_revision THEN
            RAISE EXCEPTION 'minigame resolution receipt is immutable';
        END IF;
    ELSIF NEW.status = 'resolved' AND OLD.status <> 'resolved' THEN
        IF NEW.resolution_receipt IS NULL OR NEW.resolution_company_revision IS NULL OR
           NEW.resolution_founder_revision IS NULL THEN
            RAISE EXCEPTION 'resolved minigame session requires its complete receipt tuple';
        END IF;
    ELSIF NEW.resolution_receipt IS NOT NULL OR NEW.resolution_company_revision IS NOT NULL OR
          NEW.resolution_founder_revision IS NOT NULL THEN
        RAISE EXCEPTION 'unresolved minigame session cannot carry a resolution receipt';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER minigame_resolution_receipt_guard
BEFORE UPDATE ON minigame_sessions
FOR EACH ROW EXECUTE FUNCTION enforce_minigame_resolution_receipt();

-- +goose Down
DROP TRIGGER minigame_resolution_receipt_guard ON minigame_sessions;
DROP FUNCTION enforce_minigame_resolution_receipt();
ALTER TABLE minigame_sessions
    DROP CONSTRAINT minigame_resolution_receipt_shape,
    DROP COLUMN resolution_founder_revision,
    DROP COLUMN resolution_company_revision,
    DROP COLUMN resolution_receipt;
