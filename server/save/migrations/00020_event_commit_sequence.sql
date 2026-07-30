-- +goose Up
ALTER TABLE events ADD COLUMN event_seq bigserial NOT NULL;

-- Existing intent replays used the closed Prestige-kind order before this
-- sequence existed. Preserve that relative order during upgrade; all new rows
-- receive their exact insertion order directly from the sequence.
WITH ranked AS (
    SELECT event_id,
           row_number() OVER (
               ORDER BY occurred_at,
                        intent_id NULLS FIRST,
                        CASE kind
                            WHEN 'founder_advanced' THEN 0
                            WHEN 'run_ended' THEN 1
                            WHEN 'run_started' THEN 2
                            ELSE 3
                        END,
                        revision,
                        stream_id,
                        event_id
           ) AS seq
    FROM events
)
UPDATE events
SET event_seq = ranked.seq
FROM ranked
WHERE events.event_id = ranked.event_id;

SELECT setval(
    'events_event_seq_seq',
    COALESCE((SELECT max(event_seq) FROM events), 1),
    EXISTS (SELECT 1 FROM events)
);

CREATE UNIQUE INDEX events_commit_sequence_idx ON events(event_seq);

-- +goose Down
DROP INDEX events_commit_sequence_idx;
ALTER TABLE events DROP COLUMN event_seq;
