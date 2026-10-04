# Public board after account deletion — retained current-HEAD witness

**Population:** the composed first-hour Postgres test in `server/gameserver/first_hour_composed_integration_test.go`, extended under RP-118. **Product source:** `f9ab1037` (no production change in this witness). **Date:** 2026-10-04. This is an executed behavior observation, not a deletion or retention policy.

## What ran

The unmodified composed first-hour test passed cold first. Its real bootstrap account and server-side script produced two terminal runs. The composed verifier and leaderboard projector naturally verified and archived them; this witness inserted no board, queue, archive or event row. The new assertion waited for a projected `any_percent` row, then read its exact Founder/run identity through the mounted unauthenticated `/api/public/v1/boards/{category}` HTTP route. It refreshed the bootstrap session through the real endpoint, called authenticated `DELETE /api/v1/account`, checked the resulting account/Founder/stream state, and read the same public route again.

The native Postgres 16 service ran in a uniquely named Docker Compose project with an out-of-tree override to the cached ARM64 image. The repository's `postgres:16-alpine` tag is amd64 on this host. Neither the base Compose file nor the shared test containers changed. The focused cold run passed the joined gameserver test, the Account deletion integration control, and the public-board integration control with `-count=1`; non-Postgres gameserver/leaderboard/account tests and `go vet` passed.

| Founder-scoped family | Before delete | After delete |
|---|---:|---:|
| `verified_runs` | 6 | 6 |
| Projection markers joined to those runs | 6 | 6 |
| Verification queue rows | 2 | 2 |
| `run_log_archive` rows | 2 | 2 |
| Events joined through retained streams | 1,136 | 1,136 |

The event count is illustrative, not a fixed oracle: another passing run had 1,145 before and after. The test requires a nonempty event family and exact before/after equality within its own run. The delete returned 204. The account row vanished; the Founder row remained archived and unlinked; all Founder-owned save streams (at least two) were archived. The public HTTP board still returned the previously identified `founder_id`/`run_id` pair. This is stronger than RP-130's directly seeded relational row: the row here came from the real projector and was read through the public route. It is also different from Claude's historical `ae98287f` diagnostic: that run included an artificially premature queue row to elicit a real-worker dead letter, but at that source coordinate no public board route was mounted, and its temporary test was removed.

## Discrimination and limits

The isolated board page contained no unmatched Founder. Flipping the `glitched` board-variable partition returned no matching run before or after deletion. A temporary production-reader mutation that hid archived Founders left the before-delete path intact but made the witness fail specifically at the post-delete public-board assertion; the mutation was restored byte-exact. The test's first draft used `accounts.id` instead of the real `accounts.account_id` column and failed before this result; the schema error was corrected rather than weakening the post-delete check.

This population has no dead-letter or poison row; it does not reproduce the old branch's complete producer→dead-letter chain. It does not exercise export, backup restoration, device-local data, deleted-account re-association through other payloads, player-facing deletion disclosure, or legal sufficiency. The row's survival is a **current behavior**, not an instruction to retain it. D-009/D-015 must choose whether public ranking is retained, anonymized or withdrawn after account deletion, how archived streams and derived rows are treated, and what players are told. Any later behavior change needs an accepted contract and a replacement witness for the chosen policy.
