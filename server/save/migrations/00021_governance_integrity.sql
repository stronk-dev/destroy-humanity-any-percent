-- +goose Up
-- An epoch may change exactly once: the current row can acquire its closing
-- timestamp. Everything else is historical governance evidence.
-- +goose StatementBegin
CREATE FUNCTION protect_epoch_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'UPDATE'
       AND OLD.ended_at IS NULL
       AND NEW.ended_at IS NOT NULL
       AND NEW.epoch_id = OLD.epoch_id
       AND NEW.name = OLD.name
       AND NEW.started_at = OLD.started_at
       AND NEW.changelog_ref = OLD.changelog_ref THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'immutable epoch history';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER epochs_history_guard
    BEFORE UPDATE OR DELETE ON epochs
    FOR EACH ROW EXECUTE FUNCTION protect_epoch_history();

ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_run_id_format CHECK (
    run_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}:[1-9][0-9]*$'
);

-- +goose Down
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_run_id_format;
DROP TRIGGER epochs_history_guard ON epochs;
DROP FUNCTION protect_epoch_history();
