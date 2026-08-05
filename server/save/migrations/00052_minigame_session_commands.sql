-- +goose Up
CREATE TABLE minigame_session_commands (
    session_id uuid NOT NULL REFERENCES minigame_sessions(session_id) ON DELETE CASCADE,
    seq bigint NOT NULL CHECK (seq > 0 AND seq <= 9007199254740991),
    command bytea NOT NULL CHECK (octet_length(command) > 1),
    applied_revision bigint NOT NULL CHECK (applied_revision > 1 AND applied_revision <= 9007199254740991),
    server_ts_ms bigint NOT NULL CHECK (server_ts_ms > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (session_id,seq),
    UNIQUE (session_id,applied_revision),
    CHECK (applied_revision = seq + 1)
);

CREATE TRIGGER minigame_session_commands_immutable
BEFORE UPDATE OR DELETE ON minigame_session_commands
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose Down
DROP TRIGGER minigame_session_commands_immutable ON minigame_session_commands;
DROP TABLE minigame_session_commands;
