-- +goose Up
-- Immutable command rows may disappear only as the database-owned cascade of
-- their parent session retention/deletion. A direct child-row delete remains
-- forbidden, while account deletion cannot be bricked by historical sessions.
-- +goose StatementBegin
CREATE FUNCTION enforce_minigame_command_immutability() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' AND NOT EXISTS (
        SELECT 1 FROM minigame_sessions WHERE session_id=OLD.session_id
    ) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'minigame session command is immutable';
END;
$$;
-- +goose StatementEnd

DROP TRIGGER minigame_session_commands_immutable ON minigame_session_commands;
CREATE TRIGGER minigame_session_commands_immutable
BEFORE UPDATE OR DELETE ON minigame_session_commands
FOR EACH ROW EXECUTE FUNCTION enforce_minigame_command_immutability();

-- +goose Down
DROP TRIGGER minigame_session_commands_immutable ON minigame_session_commands;
DROP FUNCTION enforce_minigame_command_immutability();
CREATE TRIGGER minigame_session_commands_immutable
BEFORE UPDATE OR DELETE ON minigame_session_commands
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();
