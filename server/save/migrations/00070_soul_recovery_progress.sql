-- +goose Up
-- Soul had no production recovery rows before heartbeat capabilities shipped,
-- so this migration deliberately has no fabricated backfill/default values.
ALTER TABLE soul_recovery_sessions
    ADD COLUMN progress_token uuid NOT NULL,
    ADD COLUMN attended_progress_ms bigint NOT NULL
        CHECK (attended_progress_ms >= 0 AND attended_progress_ms <= 9007199254740991),
    ADD COLUMN last_progress_server_ms bigint NOT NULL
        CHECK (last_progress_server_ms >= 0 AND last_progress_server_ms <= 9007199254740991),
    ADD CONSTRAINT soul_recovery_progress_active_check CHECK (
        status NOT IN ('active','claimed') OR
        (progress_token IS NOT NULL AND attended_progress_ms >= 0 AND last_progress_server_ms >= 0)
    );

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_soul_recovery_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status IN ('resolved','cancelled') THEN
        RAISE EXCEPTION 'terminal soul recovery session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.founder_id <> OLD.founder_id OR
       NEW.founder_stream_id <> OLD.founder_stream_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.activity_id <> OLD.activity_id OR NEW.founder_attended_start_ms <> OLD.founder_attended_start_ms OR
       NEW.required_duration_ms <> OLD.required_duration_ms OR NEW.start_request_hash <> OLD.start_request_hash OR
       NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'soul recovery session identity is immutable';
    END IF;
    IF NEW.attended_progress_ms < OLD.attended_progress_ms OR
       NEW.last_progress_server_ms < OLD.last_progress_server_ms THEN
        RAISE EXCEPTION 'soul recovery progress cannot regress';
    END IF;
    IF NOT ((OLD.status='active' AND NEW.status IN ('active','claimed')) OR
            (OLD.status='claimed' AND NEW.status IN ('claimed','active','resolved','cancelled'))) THEN
        RAISE EXCEPTION 'invalid soul recovery session transition';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_soul_recovery_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status IN ('resolved','cancelled') THEN
        RAISE EXCEPTION 'terminal soul recovery session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.founder_id <> OLD.founder_id OR
       NEW.founder_stream_id <> OLD.founder_stream_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.activity_id <> OLD.activity_id OR NEW.founder_attended_start_ms <> OLD.founder_attended_start_ms OR
       NEW.required_duration_ms <> OLD.required_duration_ms OR NEW.start_request_hash <> OLD.start_request_hash OR
       NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'soul recovery session identity is immutable';
    END IF;
    IF NOT ((OLD.status='active' AND NEW.status='claimed') OR
            (OLD.status='claimed' AND NEW.status IN ('claimed','active','resolved','cancelled'))) THEN
        RAISE EXCEPTION 'invalid soul recovery session transition';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
ALTER TABLE soul_recovery_sessions
    DROP CONSTRAINT soul_recovery_progress_active_check,
    DROP COLUMN last_progress_server_ms,
    DROP COLUMN attended_progress_ms,
    DROP COLUMN progress_token;
