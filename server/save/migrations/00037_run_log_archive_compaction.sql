-- +goose Up
DROP TRIGGER run_log_immutable ON run_log;
DROP FUNCTION IF EXISTS protect_active_run_log();

-- +goose StatementBegin
CREATE FUNCTION protect_active_run_log() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' AND EXISTS (
        SELECT 1 FROM run_log_archive archive WHERE archive.run_id=OLD.run_id
    ) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'immutable active run log';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER run_log_immutable
BEFORE UPDATE OR DELETE ON run_log
FOR EACH ROW EXECUTE FUNCTION protect_active_run_log();

-- +goose Down
DROP TRIGGER run_log_immutable ON run_log;
DROP FUNCTION protect_active_run_log();
CREATE TRIGGER run_log_immutable
BEFORE UPDATE OR DELETE ON run_log
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
