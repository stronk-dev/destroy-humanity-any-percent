# Minigame, Soul and Transport rows after account deletion — bounded result

**Coordinate:** product `7e8aa70`, predeclared at `8a02a3c` in
[`data-rights-minigame-soul-delete-plan.md`](data-rights-minigame-soul-delete-plan.md); run
2026-09-24 by Claude.
**Classification:** [V] for the executed deletion outcome on seeded rows. This is not a composed
Minigame/Soul API workflow.

## Population and execution

A temporary extension of `TestAccountSessionIntegration` ran after the test's real applied
`perform_manual_batch` intent and before its real authenticated `DELETE /api/v1/account`. It
seeded rows for the account's **active** imported Founder, its Company stream and its pinned run
epoch:

- one `active` `minigame_sessions` row;
- one `minigame_create_receipts` row and one `minigame_command_receipts` row;
- one `minigame_faucet_window` row;
- one `active` `soul_recovery_sessions` row with a non-null `progress_token`.

The `intent_records` and `transport_player_outbox` rows came from the real intent route; the probe
counted them but did not seed them.

The declared service ran cold with no skip:

```text
make test-save-integration SAVE_TEST_PACKAGES='./account' SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -v' SAVE_TEST_COUNT=1
RP-122 before delete: minigame_sessions=1 create_receipts=1 command_receipts=1 faucet=1 soul_sessions=1 soul_tokens=1 intent_records=1 player_outbox=1
RP-122 delete status=204
RP-122 after delete:  minigame_sessions=1 create_receipts=1 command_receipts=1 faucet=1 soul_sessions=1 soul_tokens=1 intent_records=1 player_outbox=1
RP-122 founder linked=0 unarchived=0
RP-122 control in-tx after soul/faucet delete: ... faucet=0 soul_sessions=0 soul_tokens=0 ...
RP-122 control parent session delete err=ERROR: minigame API receipts are immutable (SQLSTATE P0001)
RP-122 after rollbacks: (identical to after delete)
--- PASS: TestAccountSessionIntegration (0.62s)
```

The pre-delete validity control held: every family counted one for an active, account-linked
Founder. Account deletion succeeded (204) with an active Minigame session carrying immutable API
receipts, so RP-128 does not block account deletion. That is because deletion archives the
Founder instead of deleting it. The Founder was unlinked and archived, but **every counted row
survived**, including the active Soul session's progress token.

In a rollback-only transaction, deleting the Soul and faucet rows made those counts zero. This
shows the queries observe removal. A separate rollback-only parent-session delete still failed
with the RP-128 immutability error after account deletion. Both transactions rolled back, and the
final counts matched the post-delete counts. The temporary test edit was restored with
`git checkout` and `git diff --exit-code` returned 0. The Postgres container and network were
removed.

## Boundary

RP-122 is now an executed outcome for these eight row families rather than a source-only trace.
It does not cover:

- the Minigame or Soul API coordinators, since the rows were seeded directly;
- terminal-session variants or `minigame_session_commands` bytes;
- whether a seeded *active* session stays actionable, which is unreachable without credentials
  after deletion;
- export, backup restore or any retention duration.

Whether these rows should be deleted, severed or retained, and for how long, remains owner
D-009/D-015 work. See [`rights-decision-sheet.md`](rights-decision-sheet.md). Deleting any of them
on account deletion needs new code, and for Minigame sessions an RP-128 migration.
