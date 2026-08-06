-- +goose Up
CREATE TABLE run_frozen_contributions (
    company_stream_id uuid NOT NULL,
    run_seq bigint NOT NULL CHECK (run_seq > 0 AND run_seq <= 9007199254740991),
    source_id text NOT NULL CHECK (source_id ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    slot text NOT NULL CHECK (slot IN ('upgrades','milestones','faction','doctrine','commons','trust','event_buffs','prestige')),
    target text NOT NULL CHECK (target ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$' OR target='all'),
    factor text NOT NULL CHECK (factor <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (company_stream_id,run_seq,source_id),
    FOREIGN KEY (company_stream_id,run_seq) REFERENCES run_epochs(company_stream_id,run_seq)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TRIGGER run_frozen_contributions_immutable
BEFORE UPDATE OR DELETE ON run_frozen_contributions
FOR EACH ROW EXECUTE FUNCTION reject_immutable_change();

-- +goose StatementBegin
CREATE FUNCTION require_fiscal_frozen_contributions() RETURNS trigger LANGUAGE plpgsql AS $$
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

CREATE CONSTRAINT TRIGGER run_epoch_requires_fiscal_frozen_contributions
AFTER INSERT ON run_epochs
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION require_fiscal_frozen_contributions();

-- +goose Down
DROP TRIGGER run_epoch_requires_fiscal_frozen_contributions ON run_epochs;
DROP FUNCTION require_fiscal_frozen_contributions();
DROP TRIGGER run_frozen_contributions_immutable ON run_frozen_contributions;
DROP TABLE run_frozen_contributions;
