# Run Genesis archival remediation — append-only log

## 2026-08-01 — opened from designated full-span verdict

The implemented replay system remains archived as shipped; its acceptance record is reopened
fix-forward. Scope is limited to the four failed gate claims and Transport F2/F3. The account
integration `GuildSettlements` value is an empty constructor fixture, not production composition;
the real `guild.Service.PendingSettlements` seam stays visibly unchecked under the active
`cmd/gameserver` plan until the Commons participation-weight and world-snapshot owner contracts
exist. No push or history rewrite is authorized.

## 2026-08-01 — CI, KV-1, and provenance repair

The client CI checkout now fetches full history. A repository-owned checker verifies that the job
which invokes `make verify-client` has `fetch-depth: 0`; adversarial fixtures reject shallow,
depth-one, and missing-gate variants. KV-1 now owns `server/replaycatalog/` and
`server/leaderboard/categories.go`, and the full history fixture remains green.

Review provenance correction: the three disputed entries in the archived Run Genesis log were
**Review by: Darwin (`/root/l7b_independent_review`)**, **Recorded by: root Codex**. They were not
reviews by the project's designated Claude reviewer and must not be cited as that gate. The later
2026-08-01 designated full-span verdict is the authoritative archival review. `AGENTS.md` now
requires every future verdict to label reviewer and recorder explicitly.

## 2026-08-01 — sequential corpus debt closed in sequence

The 51-intent run now carries its own immutable seven-artifact bundle. Its economy hardcap reaches
the numeric-domain boundary without mutating catalogs between commands; the same evolving run
performs a max purchase that emits `invariant_reported`, keeps the spawned offer alive, and expires
it during the next ordinary intent. Go asserts all three at generation and TypeScript asserts all
three are present before replaying the entire sequence to byte-identical receipts/events/state.

## 2026-08-01 — Transport F2/F3 implementation

Outbound wire v2 adds the closed `cursor_effect` field. `compensation` is exactly `historical`;
all other event kinds are `advance`. Go and TypeScript reject mismatched pairs through one shared
wire corpus. The client now owns a concrete Company/Founder cursor implementation: historical
events deliver without cursor mutation, same-revision forward events dedupe by event ID, and real
forward gaps request full sync.

Migration 00042 upgrades queued events and the event trigger without editing applied migration
00040. It retains the strict receipt cap while allowing authoritative event history to commit at
any payload size. Oversized events then take the relay's bounded deterministic dead-letter path.
The real-Postgres fixture commits a >60 KiB compensation event, observes its historical marker in
the outbox, and independently proves oversized receipts still fail. The migration Down refuses to
destroy legal oversized event rows.

Seam review found that a dead-lettered event is not guaranteed to create a later numeric gap: a
same-revision receipt can follow it. The relay now publishes a bounded `resync_required` system
frame when an event reaches its fifth deterministic failure; a regression test proves the poison
row is dead-lettered and the recovery signal is emitted exactly then.

## 2026-08-01 — implementation diff review

**Review by: root Codex. Recorded by: root Codex. Range: `6141a0f..6332c6c`.**

Reviewed the CI workflow/guard fixture, KV-1 registry growth, sequential bundle and both-runtime
assertions, v2 kind/effect matrix, migration Up/Down behavior, live queue reservations, relay
failure path, client scope cursor, and canonical docs. The live private queue counts reservations
by revision but does not require monotonic revisions, so historical compensation reaches the
client as designed. Review found one adjacent recovery defect: a dead-lettered event could be
followed by a same-revision receipt, hiding the presumed forward gap. Commit `6332c6c` adds the
explicit recovery signal and regression proof. No other open defect was found in this scope.

This is an implementer self-review and does **not** satisfy plan item 7's independent gate. The
real Guild settlement composition box also remains open: the false fixture claim is corrected,
but the production owner contracts and composed test have not landed, so the checkbox convention
forbids marking it complete.

## 2026-08-02 — designated independent review (6141a0f..8578fb1) — APPROVED

Review by: the project's designated Claude reviewer. Recorded by: same.

All four reopen items verified closed in source: both CI jobs fetch full history with a
repository-owned checker + adversarial workflow fixtures; KV-1 registers `server/replaycatalog/`
and `server/leaderboard/categories.go`; the sequential 51-intent run now carries max-mode (with
`invariant_reported` asserted at generation AND before TS replay), a live offer expiring during a
later ordinary intent, and its own immutable seven-artifact bundle; the false composition claim
is corrected in the ledger with the seam visibly open. **The provenance question is resolved by
the record itself: the disputed entries were Review by: Darwin, Recorded by: root Codex — a
delegated review mislabeled as the designated gate; the AGENTS.md authorship rule now prevents
recurrence, and the 2026-08-01 designated full-span verdict stands as the authoritative archival
review.**

**Transport F2/F3: RATIFIED as owner contract.** Wire v2's `cursor_effect` is exactly the
declared-backward-compensation rule my finding demanded: closed to `advance|historical`,
biconditionally locked to the `compensation` kind in BOTH validators, historical events delivered
as audit output without cursor mutation, same-revision forward events deduped by event ID, real
forward gaps → full sync — with nine shared wire vectors enforcing the pairing and a concrete
client cursor implementation. Migration 00042 upgrades without editing applied 00040; oversized
events commit and take the bounded dead-letter path while receipts keep the strict cap; the
dead-letter `resync_required` frame closes the silent-gap seam my F12 noted. The RFC edit is
ratified as written; the changelog gains this ratification line.

Honestly open, agreed: real `guild.Service.PendingSettlements` composition — blocked on the two
owner contracts drafted today (Commons participation weight; world snapshot schema).
