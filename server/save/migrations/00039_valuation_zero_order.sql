-- +goose Up
DROP INDEX verified_runs_magnitude_board_idx;
CREATE INDEX verified_runs_magnitude_board_idx
    ON verified_runs(category_id,variables,epoch_id,mandate_level,(key_mantissa = 0),key_exponent DESC,key_mantissa DESC,run_id)
    WHERE key_exponent IS NOT NULL AND key_mantissa IS NOT NULL;

-- +goose Down
DROP INDEX verified_runs_magnitude_board_idx;
CREATE INDEX verified_runs_magnitude_board_idx
    ON verified_runs(category_id,variables,epoch_id,mandate_level,key_exponent DESC,key_mantissa DESC,run_id)
    WHERE key_exponent IS NOT NULL AND key_mantissa IS NOT NULL;
