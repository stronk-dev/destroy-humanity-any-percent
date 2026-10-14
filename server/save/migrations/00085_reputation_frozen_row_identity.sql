-- Reputation R3/AC6: the expected count alone does not establish that the
-- Reputation row exists. Bind its identity to the pinned tree's declaration.
-- Previously applied migrations and historical pins remain untouched.
-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION require_fiscal_frozen_contributions() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE expected_count integer;
DECLARE actual_count integer;
DECLARE tree_present boolean;
DECLARE tree_bonus jsonb;
BEGIN
    SELECT (CASE WHEN fiscal.bytes IS NULL THEN 0
                 ELSE 1 + jsonb_array_length(convert_from(fiscal.bytes,'UTF8')::jsonb->'generator_level_rows') END)
         + (CASE WHEN tree.bytes IS NULL THEN 0 ELSE 1 END),
           tree.bytes IS NOT NULL,
           CASE WHEN tree.bytes IS NULL THEN NULL
                ELSE convert_from(tree.bytes,'UTF8')::jsonb->'bonus' END
      INTO expected_count, tree_present, tree_bonus
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
    IF tree_present AND NOT EXISTS (
        SELECT 1 FROM run_frozen_contributions contribution
         WHERE contribution.company_stream_id=NEW.company_stream_id
           AND contribution.run_seq=NEW.run_seq
           AND contribution.source_id=tree_bonus->>'source_id'
           AND contribution.slot=tree_bonus->>'slot'
           AND contribution.target=tree_bonus->>'target'
    ) THEN
        RAISE EXCEPTION 'run pin requires complete frozen Founder contributions: missing declared Reputation row';
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
