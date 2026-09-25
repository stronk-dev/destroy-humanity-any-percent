-- Reputation Tree v1 R3/R7: a run pinned to a bundle with a reputation_tree
-- artifact also freezes exactly one reputation.founder_bonus row.
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION require_fiscal_frozen_contributions() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE expected_count integer;
DECLARE actual_count integer;
BEGIN
    SELECT (CASE WHEN fiscal.bytes IS NULL THEN 0
                 ELSE 1 + jsonb_array_length(convert_from(fiscal.bytes,'UTF8')::jsonb->'generator_level_rows') END)
         + (CASE WHEN tree.bytes IS NULL THEN 0 ELSE 1 END)
      INTO expected_count
      FROM (SELECT NEW.constants_hash AS constants_hash) pin
      LEFT JOIN catalog_artifacts fiscal
        ON fiscal.constants_hash=pin.constants_hash AND fiscal.artifact_name='fiscal'
      LEFT JOIN catalog_artifacts tree
        ON tree.constants_hash=pin.constants_hash AND tree.artifact_name='reputation_tree';
    SELECT count(*) INTO actual_count FROM run_frozen_contributions contribution
     WHERE contribution.company_stream_id=NEW.company_stream_id AND contribution.run_seq=NEW.run_seq;
    IF actual_count <> expected_count THEN
        RAISE EXCEPTION 'run pin requires complete frozen Founder contributions: expected %, got %', expected_count, actual_count;
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION require_fiscal_frozen_contributions() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE expected_count integer;
DECLARE actual_count integer;
BEGIN
    SELECT CASE WHEN artifact.bytes IS NULL THEN 0
                ELSE 1 + jsonb_array_length(convert_from(artifact.bytes,'UTF8')::jsonb->'generator_level_rows') END
      INTO expected_count
      FROM (SELECT NEW.constants_hash AS constants_hash) pin
      LEFT JOIN catalog_artifacts artifact
        ON artifact.constants_hash=pin.constants_hash AND artifact.artifact_name='fiscal';
    SELECT count(*) INTO actual_count FROM run_frozen_contributions contribution
     WHERE contribution.company_stream_id=NEW.company_stream_id AND contribution.run_seq=NEW.run_seq;
    IF actual_count <> expected_count THEN
        RAISE EXCEPTION 'run pin requires complete frozen Fiscal contributions: expected %, got %', expected_count, actual_count;
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd
