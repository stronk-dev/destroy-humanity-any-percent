# RP-122 — Minigame, Soul and Transport rows after account deletion

## Producer-backed supplement — declared 2026-10-15 at 6a6a32b9

Extend the existing composed Pitch/Soul lifecycle after all its original outcomes,
in both literal and purchased-generator fixture populations. No direct session or
receipt seed: the real APIs already produce one resolved Pitch, immutable command/
create/command receipts and a faucet window, plus resolved and watchdog-cancelled
Soul sessions. Observe all six table families before/after the authenticated
Account DELETE. Every before-family must be populated or the observation is invalid.
Use captured durable session/Founder identifiers, not the disappearing account join.

Compare stable ordered SQL-row JSON representations (including nested JSON and
hex-encoded command bytea), without publishing their contents/tokens. Report each
family as removed, retained-identical, partially-removed, expanded or modified;
do not assert that retaining or erasing it is the desired policy. A missing-session
query must observe zero, and the classification controls must distinguish outcomes.
Check live access before delete, its204 response, archived/unlinked Founder and
post-delete access/refresh refusals. Retention conclusions require an actual run;
neither denied access nor row counts alone prove content erasure.

Run the existing root Postgres lifecycle selector cold. This is producer-backed
terminal Minigame/Soul evidence, not active-session deletion, Guild, every nested
field/history, browser/participant rights, expiry or backup/restore proof. The
five-minute recovery progression and24-hour watchdog use the existing injected
clock; do not call them natural wall-clock player observations. No product/privacy
policy, schema, migration, copy, CI or release authority changes. Preserve the dated
seeded experiment below; reconcile its dossier and RP-122 after execution.

## Historical seeded experiment

**Status:** predeclared diagnostic, 2026-09-24; product source `7e8aa70` (no later `server/`
diff at HEAD `a64ed0e`). No Account, Minigame, Soul, Transport or retention behavior change is
authorized by this probe.

## Question and population

RP-122 is source-derived: account deletion archives Founder mappings and save streams instead of
deleting them, so rows whose only deletion authority is an `ON DELETE CASCADE` FK to
`account_founders` or `save_streams` should survive. Does the real authenticated
`DELETE /api/v1/account` route leave these rows in place, and does it succeed at all when a
Minigame session with immutable API receipts (RP-128) belongs to the deleted account?

Temporarily extend the existing real-Postgres `TestAccountSessionIntegration`. After its final
import and applied `perform_manual_batch` intent, and before deletion, seed for the account's
**active** imported Founder and its active Company stream/run epoch:

- one `active` `minigame_sessions` row plus one `minigame_create_receipts` and one
  `minigame_command_receipts` row (the RP-128 trigger population);
- one `minigame_faucet_window` row; and
- one `active` `soul_recovery_sessions` row with a non-null `progress_token`.

The applied intent already produced by the real route supplies the `intent_records` receipt and
`transport_player_outbox` population; the probe counts them rather than seeding them.

## Controls, limits and exit

- **Pre-delete validity:** every seeded family and the real intent receipt/outbox rows count at
  least one for the active Founder, or the run is invalid.
- **Outcome recorded either way:** the delete status is observed and printed before the test's
  existing 204 assertion. A non-204 is recorded as a finding, not repaired.
- **Post-delete:** count each family again by the seeded Founder/stream identifiers; count the
  soul row's non-null `progress_token`; confirm the Founder mapping is unlinked/archived.
- **Negative control (sensitivity of the oracle):** in a rollback-only transaction after deletion,
  delete the `soul_recovery_sessions` and `minigame_faucet_window` rows for that Founder and
  demand those counts fall to zero inside the transaction (the queries can observe removal);
  and demand that a parent `minigame_sessions` delete in the same kind of transaction still fails
  with `minigame API receipts are immutable` (RP-128 persists after account deletion).
- Run cold with `make test-save-integration SAVE_TEST_PACKAGES='./account'
  SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -v' SAVE_TEST_COUNT=1` on the declared
  Postgres service, confirm no skip, then remove the temporary test edit exactly and prove
  `git diff --exit-code -- server/account/account_integration_test.go`.

Record the observed counts in a bounded result file, RP-122 and the decision/1.0 trackers. Seeded
Minigame/Soul rows are not produced by their real API coordinators; do not call this a composed
Minigame/Soul workflow, an export schema or a retention policy. Whether these rows should be
deleted, anonymized or retained, and for how long, is D-009/D-015 owner work.
