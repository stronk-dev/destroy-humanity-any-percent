-- +goose Up
-- Existing rows deliberately remain NULL: no pre-migration verifier can
-- reconstruct inputs that were never recorded. Every new row is fail-closed.
ALTER TABLE run_log ADD COLUMN replay_inputs jsonb;

-- +goose StatementBegin
CREATE FUNCTION require_run_log_replay_inputs() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.replay_inputs IS NULL OR jsonb_typeof(NEW.replay_inputs) <> 'object' THEN
        RAISE EXCEPTION 'new run-log rows require replay inputs';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER run_log_replay_inputs_required
BEFORE INSERT ON run_log
FOR EACH ROW EXECUTE FUNCTION require_run_log_replay_inputs();

-- +goose Down
DROP TRIGGER run_log_replay_inputs_required ON run_log;
DROP FUNCTION require_run_log_replay_inputs();
ALTER TABLE run_log DROP COLUMN replay_inputs;
