-- +goose Up
ALTER TABLE save_streams
    ADD CONSTRAINT save_streams_id_owner_unique UNIQUE (id,owner_id);
ALTER TABLE run_epochs
    ADD CONSTRAINT run_epochs_stream_run_hash_unique UNIQUE (company_stream_id,run_seq,constants_hash);
ALTER TABLE minigame_sessions
    ADD CONSTRAINT minigame_session_founder_owns_stream
        FOREIGN KEY (company_stream_id,founder_id) REFERENCES save_streams(id,owner_id),
    ADD CONSTRAINT minigame_session_uses_pinned_hash
        FOREIGN KEY (company_stream_id,run_seq,constants_hash)
        REFERENCES run_epochs(company_stream_id,run_seq,constants_hash);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_minigame_session_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status = 'resolved' THEN
        RAISE EXCEPTION 'resolved minigame session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.minigame_id <> OLD.minigame_id OR
       NEW.founder_id <> OLD.founder_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.engine_ref <> OLD.engine_ref OR
       NEW.engine_version <> OLD.engine_version OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.scaling_inputs <> OLD.scaling_inputs OR NEW.seed <> OLD.seed OR NEW.mode <> OLD.mode OR
       NEW.genesis <> OLD.genesis OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'minigame session genesis is immutable';
    END IF;
    IF NOT (
        (OLD.status = 'active' AND NEW.status = 'claimed') OR
        (OLD.status = 'claimed' AND NEW.status IN ('claimed','active','resolved'))
    ) THEN
        RAISE EXCEPTION 'invalid minigame session transition';
    END IF;
    IF NEW.status = 'active' AND NOT (
        (NEW.state = OLD.state AND NEW.revision = OLD.revision) OR
        NEW.revision = OLD.revision + 1
    ) THEN
        RAISE EXCEPTION 'minigame play revision does not match claim outcome';
    END IF;
    IF NEW.status = 'resolved' AND NEW.revision <> OLD.revision + 1 THEN
        RAISE EXCEPTION 'resolved minigame session must advance exactly once';
    END IF;
    IF NEW.status = 'claimed' AND NEW.revision <> OLD.revision THEN
        RAISE EXCEPTION 'claiming cannot advance minigame session revision';
    END IF;
    IF NEW.status = 'claimed' AND (
        NEW.state <> OLD.state OR NEW.result IS DISTINCT FROM OLD.result OR
        NEW.resolved_at IS DISTINCT FROM OLD.resolved_at
    ) THEN
        RAISE EXCEPTION 'claim transition changed authoritative session data';
    END IF;
    IF NEW.status = 'active' AND (
        NEW.result IS DISTINCT FROM OLD.result OR NEW.resolved_at IS DISTINCT FROM OLD.resolved_at
    ) THEN
        RAISE EXCEPTION 'active transition changed terminal session data';
    END IF;
    IF NEW.updated_at < OLD.updated_at THEN
        RAISE EXCEPTION 'minigame session time moved backwards';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_minigame_session_transition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.status = 'resolved' THEN
        RAISE EXCEPTION 'resolved minigame session is immutable';
    END IF;
    IF NEW.session_id <> OLD.session_id OR NEW.minigame_id <> OLD.minigame_id OR
       NEW.founder_id <> OLD.founder_id OR NEW.company_stream_id <> OLD.company_stream_id OR
       NEW.run_seq <> OLD.run_seq OR NEW.engine_ref <> OLD.engine_ref OR
       NEW.engine_version <> OLD.engine_version OR NEW.constants_hash <> OLD.constants_hash OR
       NEW.scaling_inputs <> OLD.scaling_inputs OR NEW.seed <> OLD.seed OR NEW.mode <> OLD.mode OR
       NEW.genesis <> OLD.genesis OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'minigame session genesis is immutable';
    END IF;
    IF NOT (
        (OLD.status = 'active' AND NEW.status = 'claimed') OR
        (OLD.status = 'claimed' AND NEW.status IN ('claimed','active','resolved'))
    ) THEN
        RAISE EXCEPTION 'invalid minigame session transition';
    END IF;
    IF NEW.status = 'active' AND NOT (
        (NEW.state = OLD.state AND NEW.revision = OLD.revision) OR
        NEW.revision = OLD.revision + 1
    ) THEN
        RAISE EXCEPTION 'minigame play revision does not match claim outcome';
    END IF;
    IF NEW.status = 'resolved' AND NEW.revision <> OLD.revision + 1 THEN
        RAISE EXCEPTION 'resolved minigame session must advance exactly once';
    END IF;
    IF NEW.status = 'claimed' AND NEW.revision <> OLD.revision THEN
        RAISE EXCEPTION 'claiming cannot advance minigame session revision';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

ALTER TABLE minigame_sessions
    DROP CONSTRAINT minigame_session_uses_pinned_hash,
    DROP CONSTRAINT minigame_session_founder_owns_stream;
ALTER TABLE run_epochs DROP CONSTRAINT run_epochs_stream_run_hash_unique;
ALTER TABLE save_streams DROP CONSTRAINT save_streams_id_owner_unique;
