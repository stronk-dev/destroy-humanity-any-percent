-- +goose Up
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_check;
ALTER TABLE verified_runs ADD COLUMN key_exponent bigint;
ALTER TABLE verified_runs ADD COLUMN key_mantissa bigint;
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_key_check CHECK (
    (key_ms IS NOT NULL)::int +
    (key_int IS NOT NULL)::int +
    ((key_exponent IS NOT NULL AND key_mantissa IS NOT NULL))::int = 1 AND
    ((key_exponent IS NULL AND key_mantissa IS NULL) OR
     (key_exponent = 0 AND key_mantissa = 0) OR
     (key_exponent BETWEEN -8999999999999999 AND 8999999999999999 AND
      key_mantissa BETWEEN 100000000000 AND 999999999999))
);
CREATE INDEX verified_runs_magnitude_board_idx
    ON verified_runs(category_id,variables,epoch_id,mandate_level,key_exponent DESC,key_mantissa DESC,run_id)
    WHERE key_exponent IS NOT NULL AND key_mantissa IS NOT NULL;

-- +goose Down
DROP INDEX verified_runs_magnitude_board_idx;
ALTER TABLE verified_runs DROP CONSTRAINT verified_runs_key_check;
ALTER TABLE verified_runs DROP COLUMN key_mantissa;
ALTER TABLE verified_runs DROP COLUMN key_exponent;
ALTER TABLE verified_runs ADD CONSTRAINT verified_runs_check
    CHECK ((key_ms IS NOT NULL)::int + (key_int IS NOT NULL)::int = 1);
