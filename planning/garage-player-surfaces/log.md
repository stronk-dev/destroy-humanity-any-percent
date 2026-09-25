# Garage Player Surfaces implementation log

## 2026-09-25 — Predeclaration, Batch A (Claude)

**Implemented by:** Claude, on the owner's 2026-09-24 direction that Claude implements accepted
RFCs. Every batch awaits Codex's designated cross-party review; nothing here is self-approved.

Order: GS0.2 runtime outcome fix first (client-only), then the v4 projection with the live arms,
then the surfaces. The snapshot version is the next free Game UI snapshot version at landing
(v4 is free: `schema_version` 3 is the current maximum). Copy for new keys is implementer-drafted
candidate text in `copy/catalog/garage-surfaces-candidate.json`, marked for owner adoption; it is
not ruled copy. Each gate ships with a demonstrated severing probe.

## 2026-09-25 — GS0.2 intent outcomes (Claude)

`runtime.intent()` now resolves the typed outcome, and a non-2xx intent throws
`GameUIRequestError`. `act()` takes a `scope`: a Founder-scoped intent sends the Founder revision
and is skipped while that revision is unknown. Rejections render in one host `role="status"` line
through `noticeForOutcome`, which checks surface rows first, then the category fallbacks. Only
transport failures and 401/404/5xx set `offline`.

Server fact recorded: a stale revision is HTTP 200 `rejected revision_conflict/expected_revision`
(`server/production/intents.go`), not 409. It maps to one authoritative refresh plus the
`intent.conflict` line, with no auto-retry.

**Evidence (cold):**
- `game-ui-intent-outcome.test.ts`: 3/3.
- New browser case: a rejected buy shows "Not enough funds for that."; a revision conflict
  refreshes exactly once.
- Full lanes pass: `make typecheck test-client verify-client-boundary copy-check test-browser`
  (20136 browser tests).

**Severing probes:**
- **S1:** drop the notice assignment, which restores F2. The chromium test fails (1 failed).
- **S2:** make the rejection parser accept extra keys. The unit test fails (1 failed | 2 passed).

**Process note:** the generated copy outputs in this commit were built in a scratch worktree from
this lane's catalog only. The concurrent Typer lane's uncommitted `typer-candidate.json` keys are
therefore not committed here.
