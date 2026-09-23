# Populated-board account deletion — bounded result

**Coordinate:** product source `7e8aa70`, predeclaration
`data-rights-board-delete-plan.md` at `d753360`; run 2026-09-23.
**Classification:** [V] for the executed relational outcome; no public-reader,
player-export or policy conclusion.

## Executed population

The existing `TestAccountSessionIntegration` creates an account, Founder,
sessions, imported Founder and ten streams, then calls the real authenticated
`DELETE /api/v1/account` endpoint and checks account/session removal plus
Founder/stream archival. A temporary, uncommitted diagnostic addition inserted
one schema-valid `verification_projection_events` marker and one `verified_runs`
`any_percent` row before that request. The row used a real Founder ID from the
account, an existing epoch, a valid `stream_uuid:1` run ID and the current
four-key board variables. A pre-delete join back to the account had to equal
one; a missing or invalid seed would fail the test.

The exact root command was:

```text
make test-save-integration SAVE_TEST_PACKAGES='./account' SAVE_TEST_FLAGS='-run TestAccountSessionIntegration -v' SAVE_TEST_COUNT=1
```

It ran on the declared Postgres 16 service and exited 0, with no skip:

```text
=== RUN   TestAccountSessionIntegration
board deletion probe: pre_delete_join=1 retained_board=1 archived_unlinked_founder=1 archived_streams=2
--- PASS: TestAccountSessionIntegration (0.73s)
```

The temporary addition also asserted that the HTTP delete returned its normal
204, the original account/session/receipt checks passed, the board row still
existed, its `founder_id` joined to an `account_id IS NULL` and archived Founder,
and two streams of that Founder were archived. Removing the board row or
breaking that join would have failed the diagnostic assertion. The temporary
test edit was removed by exact patch; `git diff --exit-code --
server/account/account_integration_test.go` returned 0 afterward.

## What this establishes and what it does not

The account deletion transaction currently **retains** the verified board
identity path `verified_runs.founder_id → account_founders.founder_id → archived
save_streams.owner_id` while removing the account link. The archived stream
contents are a separate classification and retention question. The result is
consistent with the Account RFC's archival model, but it does not establish
that the remaining linkage is acceptable under D-009 or disclosed to players.

The board row was inserted directly to isolate deletion/foreign-key behavior;
the production projector and its own cold Postgres fixture are separate
evidence. No joined test here populated an archived run log and verification
dead letters, exercised a public board reader, exported player data, restored
a pre-delete backup or proved that a deleted Founder cannot be associated by
some other payload. RP-118 therefore narrows but stays open; RP-130 records the
observed retained linkage for the owner/legal deletion and retention rulings.
