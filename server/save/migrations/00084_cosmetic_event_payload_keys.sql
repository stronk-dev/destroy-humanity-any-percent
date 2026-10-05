-- Cosmetic Shop v1 §4.5/AC9: reject extra or missing cosmetic event payload keys
-- at the database boundary. Other event kinds keep their existing payload rules.
-- The existing events.payload CHECK already requires a JSON object.
-- +goose Up
ALTER TABLE events ADD CONSTRAINT events_cosmetic_payload_keys_check CHECK (
    CASE kind
        WHEN 'cosmetic_acquired.v1' THEN
            payload ? 'cosmetic_id' AND payload ? 'order_number'
            AND payload - 'cosmetic_id' - 'order_number' = '{}'::jsonb
        WHEN 'cosmetic_equipped.v1' THEN
            payload ? 'cosmetic_id' AND payload ? 'pet_id' AND payload ? 'replaced_cosmetic_id'
            AND payload - 'cosmetic_id' - 'pet_id' - 'replaced_cosmetic_id' = '{}'::jsonb
        WHEN 'cosmetic_unequipped.v1' THEN
            payload ? 'cosmetic_id' AND payload ? 'pet_id'
            AND payload - 'cosmetic_id' - 'pet_id' = '{}'::jsonb
        ELSE TRUE
    END
);

-- +goose Down
ALTER TABLE events DROP CONSTRAINT events_cosmetic_payload_keys_check;
