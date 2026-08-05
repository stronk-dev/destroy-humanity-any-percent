-- +goose Up
CREATE TABLE minigame_faucet_window (
    founder_id uuid NOT NULL REFERENCES account_founders(founder_id) ON DELETE CASCADE,
    minigame_id text NOT NULL CHECK (minigame_id ~ '^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$'),
    attended_day bigint NOT NULL CHECK (attended_day >= 0 AND attended_day <= 9007199254740991),
    quota_used bigint NOT NULL CHECK (quota_used >= 0 AND quota_used <= 9007199254740991),
    conversion_remainder_ppm bigint NOT NULL CHECK (
        conversion_remainder_ppm >= 0 AND conversion_remainder_ppm < 1000000
    ),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (founder_id,minigame_id,attended_day)
);

-- +goose Down
DROP TABLE minigame_faucet_window;
