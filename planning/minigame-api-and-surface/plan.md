# Minigame API & Surface implementation plan

Original Pitch review finished (2026-10-09): the complete 23-path Claude range
`50a3a514^..50a3a514` has a designated Codex **CHANGES REQUIRED** disposition in the log.
RP-131, RP-415 and RP-478 remain distinct from their later corrections. The latter two are
locally repaired, not independently approved by this original-range review. Complete MA
range coverage, corrective review, authored contract/copy reconciliation and acceptance remain.

Current bounded correction (2026-10-08, RP-478): Pitch declares zero Tab indices for native
controls, matching the existing Typer pattern. The revised existing component flow uses actual
Enter/Space for Start/cards/play/buy/end-shop/Return and Leave, with launcher/terminal Tab and
reverse card traversal. Old-source native WebKit fails before Start; corrected affected files,
types/build/boundaries pass. The log owns evidence and pending review. No whole accessibility,
full server keyboard journey, new mechanics or GS7 proposal adoption is claimed.

Current checkpoint (2026-10-08): RP-446 adds native Soul Recovery start/attended progress/
leave/reconnect/cancel through the production-built client, real service and SQL to the normal
composed target. Native manual actions/purchases first establish real positive idle production,
so unchanged balances during recovery are not a zero-producer control. The explicit full-duration
target also executes unchanged five-minute Repot
through native Finish, exact 5-Soul grant, unchanged Company balances and reload, after disclosed
Soul 40 setup. This is MA2/MA3 player-integration evidence, not a natural debit/history-replay,
all-duration/AT/whole-feature or release claim. Latest results and pending review range belong
in the owning log. The preceding Pitch stored-outcome checkpoint is complete locally;
designated review and the remaining contracts are still independent obligations. The HTTP
Pitch lifecycle now reconstructs the whole application over the same Postgres after a
nonterminal command, preserving the complete snapshot, exact receipts and once-only payout.
The browser journey now separately kills its actual gameserver after a native nonterminal
command commits but before its reply reaches the player. Existing automatic same-ID retry,
persisted snapshot/history, socket recovery and continued payout are required without reload
or replacement bootstrap. This strengthens solo Pitch, not death inside a transaction,
host/DB crash, clean-host recovery or a generic async API. RP-459 repairs the HTTP test's
reused-database bootstrap-receipt isolation. The owning log
records execution and review status; no whole-feature acceptance checkbox changes.
The bounded Recovery journey now also loses a real committed heartbeat reply while mounted:
visible pause, frozen persisted progress, native same-session token rotation and no missed-gap
credit before fresh attended progress. This is one transport failure case, not general reliability.
The same player path now separately survives actual gameserver SIGKILL after a heartbeat commits:
new process, unchanged stored session, WebSocket recovery and native reconnect without account
replacement/reload. The latest log owns execution, the incomplete fault control and pending review.

- [x] Compose the minigame repository, closed Pitch tenant registry, pinned content resolver, and
  platform service in `gameserver.Compose`; expose it on `Composition`.
- [x] Attach the authenticated minigame coordinator adapter after MA-C10–MA-C14 are ruled.
- [x] Implement Founder-v21 replay-owned create sequencing, atomic session creation, and create
  request-receipt idempotency.
- [x] Implement transaction-scoped command-receipt persistence primitives.
- [x] Implement the typed command coordinator with pre-execution receipt replay, atomic nonterminal
  completion, and terminal auto-resolution receipt composition.
- [x] Reject Exit from a frozen active-minigame predicate in both replay runtimes.
- [x] Wire authenticated create/command/resolve/current handlers through the generated API registry.
- [x] Prove create → command → terminal auto-resolve → retry/current over the composed real socket.
- [x] Register/generate the recovery and minigame operation schemas through API Foundation.
- [x] Extend the composed real-socket lifecycle through the Soul Recovery half required by AC1.
- [x] Prove authenticated minigame-command flooding is bounded by the ruled account/session policy.
- [x] Enumerate and verify the privacy contract for all four public minigame endpoints.
- [x] Implement surface components after UI Foundation, per MA-C9. (Pitch `50a3a51`, composed witness `06edddba`, recovery `b92a05de`; tests in those commits.)
- [ ] Hand the complete range off for designated cross-party review; do not self-archive.

## Q-002 — platform-alignment backend witness closeout

- [x] Replace the AC2 substring oracle with exact status/body-byte witnesses for every
  deterministic minigame error mapping, and prove that an invalid category/detail pairing and an
  otherwise-valid body with one appended byte both fail the exact oracle. The live registry
  narrowing is owned and reviewed separately by the API Foundation corrective lane.
- [x] Extend AC3 with a stateful Soul Recovery heartbeat-flood witness: the rejected seventh
  heartbeat must leave authoritative progress/session/token state unchanged, and ordinary clock
  refill must admit exactly the next state transition.
- [x] Extend AC4 privacy enumeration across all eight Minigame and Soul Recovery operations,
  rejecting client-authored Founder, Company, and server-clock authority in every request and
  hidden authority in every response; demonstrate a `founder_id` leak mutation in each Recovery
  request family.
- [x] Run cold root Go tests, sequential Account and Gameserver Postgres integration targets,
  `api-check`, and client typecheck; restore every mutation before the clean run.
- [x] Record a Codex exact-range first-filter and hand the bounded backend-test range to Claude for
  the mandatory designated cross-party review. Do not claim surface completion or archive MA.

Corrected Q-002 and its separate API Foundation registry lane received designated approval at
`bfd9b65`. This closes AC2/AC3/AC4's bounded backend witnesses only; the surface and archival boxes
above remain open.

2026-10-09: original Recovery surface `b92a05de` now has a complete designated Codex
CHANGES REQUIRED disposition, separate from later repairs. RP-480's current terminal/error
race is locally corrected with both reply orders through the generated browser port;
60 affected native checks,84 units and types/build/boundaries pass. The new range after
`5c9d0293` awaits cross-party review. No full MA acceptance or archival follows.
