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
