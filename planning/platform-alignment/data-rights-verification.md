# D-008/D-009/D-015 — verification and board payloads

Research checkpoint: 2026-09-23. Product source `7e8aa70` (planning-only
commits since). Predeclared method:
[`data-rights-verification-plan.md`](data-rights-verification-plan.md). This
completes the selected content-bearing fields in the five verification/board
tables; it is not an export schema, deletion/retention policy or public-board
proof. The SQL-column inventory is in
[`data-rights-fields-history-boards.md`](data-rights-fields-history-boards.md).

## Producer → persisted value → reader

| Field | Current content and proof boundary |
|---|---|
| `verification_queue.last_error` | Pending/claimed retry stores `boundedDetail(cause)` from `VerifyStoredRun`, SQL/catalog access or `ProjectVerifiedRun` failure. The helper trims whitespace, substitutes a generic phrase for empty/non-UTF-8 input, and truncates to 512 Unicode runes. It does **not** classify or redact secret/personal substrings. Version skew gets a fixed deferred-error phrase. Verified completion clears `last_error`; deterministic-dead and poison-dead completion retain their details in terminal queue rows. Terminal rows reject update/delete by trigger. |
| `verification_dead_letters.detail` | A deterministic nonverified replay verdict writes fixed text `replay verification returned <closed verdict>` into an immutable row and the terminal queue's `last_error`. It does not interpolate an arbitrary underlying exception. The row and queue key remain linked to Company stream/run. |
| `verification_poison_dead_letters.detail` | A transient error at the fifth attempt writes the bounded `err.Error()` text verbatim; an expired final claim copies prior nonempty `last_error` through the same bound, else uses a fixed phrase. The immutable poison row and terminal queue duplicate that text, and the invariant sink receives it. The code does not prove the possible upstream error strings are secret-free (RP-129). A length cap is not sanitization. |
| `verified_runs.variables` | Current Up constraint allows exactly `commons`, `advisor`, `glitched` booleans and `faction` null or mechanical-ID string. The queue projector derives Commons from event membership, Advisor from the terminal run, Glitched from executed routes and Faction from incorporated event history, checking the terminal claim against history. It is structured gameplay classification, not free-text telemetry. The retained row also has Founder/run identity and ranking keys; internal Time/Count/Magnitude board readers return Founder/run IDs. No mounted public board reader is inferred. |
| `verification_projection_events.event_id` | This immutable marker deduplicates one source event, while one event can produce multiple category rows or none (imported/drifted cases). It is a join key, not a proof of no personal linkage. The cold projector fixture left four markers and five `verified_runs` rows; their join yielded five matches, not a one-to-one map. |

All five families retain stream/run/event/Founder linkage after the current
`DeleteAccount` path archives Founder/streams and unlinks the account. That
conclusion comes from the SQL FKs and deletion source, **not** a joined test
that first populates the board and dead letters, then deletes the account.
RP-118 continues to own that missing rights witness. The immutable terminal
and board guards mean D-015 cannot assume an unmodified generic delete job.

## Executed checks and limitations

Cold real-Postgres tests each passed with `-count=1 -v`:

- `TestVerificationTransientCatalogFailureRetriesThenPoisonsIntegration`:
  five failed attempts created a poison row. Read-only SQL observed a dead
  queue row, `last_error = poison.detail`, and the exact injected synthetic
  error text `temporary catalog connection failure`. The test asserts exact
  equality of the poison detail to its injected error; it does not seed an
  actual secret or classify every production error source.
- `TestVerificationDeterministicCatalogEvidenceDeadLettersIntegration`:
  read-only SQL observed `constants_mismatch` and the fixed detail phrase
  `replay verification returned constants_mismatch`.
- `TestVerificationProjectionFailureRollsBackThenPoisonsIntegration`:
  read-only SQL observed a dead queue row retaining its synthetic projection
  failure text. This exercises another arbitrary-error producer.
- `TestQueueProjectorCategoriesVariablesPreTimerAndRetryIntegration`:
  read-only SQL found five board rows, four projection markers and exactly the
  four permitted `variables` keys. The marker-to-board join gave five matches
  because one event can feed more than one category. This is a fixture
  population, not a public board or post-delete proof.

The temporary Postgres service/network was removed after use; the named Go
cache volume remains. No private production rows or secrets were read.

## Decision handoff

- **D-008:** decide whether verified rows and operational diagnostics belong
  in player export, and how per-run records, shared world-first status and
  mutable/immutable failure history are represented.
- **D-009/D-015:** rule purpose, access/disclosure, duration and forward-only
  cleanup mechanism for terminal queue/immutable dead letters and board rows.
  RP-129 requires a bounded diagnostic-data policy or accepted sanitizer for
  arbitrary upstream error text; it does not assert an observed secret leak.
- The joined populated-board/dead-letter account deletion witness, host
  journal/alert observation and player-facing public board remain open. No
  product or release-state claim follows from this tranche.
