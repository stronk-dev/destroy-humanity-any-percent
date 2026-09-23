# D-008/D-009 — populated-board account-delete probe

**Status:** predeclared diagnostic, 2026-09-23. No product or retention policy change.

## Question

When the real `DELETE /api/v1/account` path runs with a populated
`verified_runs` row, does the transaction succeed, and what identity links
remain in the board row, Founder row and archived save stream afterward?
The current Account integration test deletes an account but has no verified
board row; the Leaderboards projector test populates boards but never deletes
the account. Joining those two facts by inference would be too weak.

## Population and method

Temporarily augment `TestAccountSessionIntegration` on the declared Postgres
test service. Before its existing delete request, insert one schema-valid
`verification_projection_events` and `verified_runs` pair tied to an actual
Founder of that account and to the seeded epoch. Record the pre-delete join.
Run that test cold through the root Make target with verbose output. After the
API delete, observe account absence, archived/unlinked Founder, retained board
row and whether its `founder_id` still resolves to the Founder/streams.

The seeded board row is a **relational diagnostic**, not a claim that the
production projector produced this particular row. The existing cold
projector fixture separately demonstrates board creation. The temporary
probe must be removed by exact patch after execution; no permanent test
assertion should lock an owner-unruled board retention choice.

## Discrimination and limits

The pre-delete join must have exactly one row; a missing seed invalidates the
probe. An HTTP error, a vanished board row, an unarchived or still-linked
Founder, or a severed board→Founder join must produce a failing observation,
not a silently skipped test. This does not establish public board behavior,
player export semantics, deletion sufficiency, legal compliance, or backup
restoration. Those remain D-008/D-009/R-003 and clean-host work.
