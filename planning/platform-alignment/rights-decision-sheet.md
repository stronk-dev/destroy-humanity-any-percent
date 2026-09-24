# D-008 / D-009 / D-015 — owner decision sheet

**Status:** prepared options for Marco, 2026-09-24; **no option is adopted.** Coordinate: product
source `7e8aa70` (no `server/` diff through HEAD at preparation). Evidence is the 60-table census
in [`data-rights-inventory.md`](data-rights-inventory.md), its five SQL-column tranches, the
browser/payload/event/receipt/verification/operator maps, and the executed RP-119/RP-122/RP-127–
RP-130 probes. This sheet turns those records into choices. It is not a legal opinion, privacy
notice or retention policy, and it does not select anything on Marco's behalf. Legal review is
still required before any adopted policy becomes public copy.

## Facts every option must respect

| Fact at this source | Evidence | Consequence for the choice |
|---|---|---|
| `DELETE /api/v1/account` deletes account, email, session, token and family rows; it **unlinks and archives** Founder mappings and save streams rather than deleting them. | `TestAccountSessionIntegration`; `data-rights-fields-account-save.md` | Any Founder- or stream-keyed row whose only deletion authority is an FK cascade survives deletion (RP-122). Choosing "delete" for such a family needs new code, not a policy line. |
| The bootstrap receipt row keeps `account_id`, digest and times permanently; migration `00073` forbids row deletion and `account_id` changes. | Inventory "Executed deletion diagnostic" | "No account-linked row remains" is currently false. Deleting it needs a new append-only migration plus replay/idempotency analysis. |
| Immutable history (`run_log_archive`, events, genesis, Founder logs, verified runs, Minigame API receipts, terminal Soul rows, poison dead letters) is enforced by triggers. | `data-rights-payload-core.md`, `-receipts.md`, `-verification.md` | Erasure inside those tables conflicts with replay/verification integrity. A delete option must name the replacement integrity mechanism; an anonymize option must name which join keys are severed. |
| Verified boards keep a row linked to an archived Founder after deletion (RP-130); Commons World-count and health sampling still count a deleted account's active membership (RP-119). | `data-rights-board-delete.md`, `data-rights-commons-delete.md` | Public or aggregate surfaces continue to reflect deleted players unless a transition is specified. |
| Guild shared event JSON retains a deleted account UUID even where the FK is nulled (RP-120). | `data-rights-fields-guild.md` | Shared-history disclosure needs its own answer; per-account Guild receipts already cascade. |
| Minigame API receipt triggers block parent-session deletion even through FK cascade (RP-128); terminal Soul rows keep their progress token (RP-127). | `data-rights-receipts.md` | Any expiry of Minigame sessions needs a migration/contract change; the Soul token is historical capability material, not a proved live bypass. |
| Full encrypted backups keep deleted accounts up to 30 days, longer for protected newest/pre-upgrade copies; a full restore re-creates deleted accounts with no deletion replay (RP-125). | `data-rights-operator.md`; RFC DP6 | "Deleted within N days" must either include backup lag honestly or add a post-restore deletion replay. |
| Browser `localStorage` holds credentials, the unrevealed recovery code, transport positions and personal-best history; nothing clears it on deletion (RP-123). | `data-rights-browser.md` | Device-local residue is part of the deletion disclosure unless the client clears it. |
| `POST /api/v1/founder/import` accepts only an offline-anonymous save onto a **pristine** Founder, resets to run 1 and marks it `imported` (excluded from boards). | `docs/accounts-and-sessions.md` | A portable export that re-imports a played account is not supported by the current import contract. |
| `PruneIntentRecords` has no production caller; no anonymous-account inactivity cleanup exists. | Inventory | Any adopted duration for intent receipts or inactive accounts needs a composed job and a fired-cleanup witness (R-007). |

## D-008 — what a player can export

| Option | Contents | Cost and consequence |
|---|---|---|
| **A — current state** | Account identifiers the player already holds, the active Founder and Company canonical state (versioned save JSON + constants hash), personal-best history from this device, and a human-readable summary. | Smallest schema; mirrors data the client already renders. Does **not** satisfy a request for history. Re-import would still hit the pristine-only import rule. |
| **B — state plus own history** | A plus the player's own revisions (latest five retained), run/Founder logs and archives, intent receipts, verified-board rows, Minigame/Soul sessions and receipts, and Commons/Guild rows **filtered to the player's own contributions**. | Needs per-family filters so other members' Guild/Commons data never leaks; archived Founders must be reachable from the account before deletion. Largest review surface (nested payload classification is still incomplete). |
| **C — portable self-host bundle** | B, signed and versioned, importable into a self-hosted instance with provenance retained and boards excluding imported history. | Requires a new import contract (played accounts, signature trust, replay under the source epoch/catalogs). Interacts with D-003's deferred self-host covenant; do not promise before the supported bundle exists. |

Also state: whether export is available after deletion is requested (a grace window) or only
before; the machine format (JSON with published schema is the default expectation under the
transparency rule); and whether other players' identifiers ever appear.

## D-009 — what deletion does and what the player is told

For each family, choose **Delete** (row removed; needs code where the FK does not already
cascade), **Sever** (row kept, every join to the account/Founder cut, player-visible names removed),
or **Retain** (kept as-is for a stated purpose and duration). The "current" column is behavior at
this source, not a recommendation.

| Family | Current | Choice needed |
|---|---|---|
| Account, email, sessions, tokens, families | Delete | Confirm. |
| Bootstrap receipt tombstone (`account_id`, digest, times) | Retain forever | Keep, sever `account_id` via new migration, or bounded expiry. |
| Founder mappings, save streams/revisions, intent receipts | Sever (unlink + archive) | Keep sever, or delete with an integrity replacement. |
| Run/Founder history, genesis, archives, events | Retain (immutable) | Retain for replay/verification with a duration, or sever further. |
| Verified boards / public rankings (RP-130) | Retain, linked to archived Founder | Remain on public boards anonymously, or be withdrawn from public readers. |
| Active Commons membership (RP-119) | Still counted | Deletion leaves the Commons (live counts drop, historic contributions kept), or counts as historic only. |
| Guild shared history (RP-120) | UUID retained in JSON | Scrub/replace the UUID, or retain with disclosure. |
| Minigame sessions/receipts, Soul sessions, player outbox (RP-122/RP-127/RP-128) | Retain | Delete (needs RP-128 migration), sever, or retain with expiry. |
| Verification dead letters/poison text (RP-129) | Retain (immutable) | Bounded retention or redaction rule. |
| Browser storage (RP-123) | Untouched | Client clears on successful deletion, and/or disclose device residue. |
| Backups (RP-125) | Up to 30 days, longer if protected | Disclose the maximum lag, add a post-restore deletion replay, or both. |
| Operator ledgers `operator` field (RP-126) | Retain | Operator-identity policy (this is operator data, not player data). |

The disclosure copy ("what remains and for how long") is owner-authored; agents may not write it.

## D-015 — retention by family

For each of the 11 table groups plus browser, backups, journals, metrics and operator artifacts,
supply **purpose**, **trigger** (time, run end, account deletion, inactivity) and **duration** or
"until service end", plus the cleanup owner. Already ruled and not reopened here: logs 14 days,
raw IP maximum 7 days, backups 7-day six-hourly + 30-day daily (D-006/D-011). Specific open items:

1. Intent receipts: adopt a duration (the documented 30 days has no running job) or retire the doc claim.
2. Anonymous accounts with no activity: inactivity trigger and duration, or none.
3. Minigame session expiry: if chosen, it needs the RP-128 migration first.
4. Terminal Soul progress token: clear on terminal state (a migration) or retain with the session.
5. Dead letters and projection events: a bounded window once resolved, or retain.
6. Alertmanager volume: explicit retention (currently no flag; Prometheus is 30 days).

## Ruling template

```text
D-008 export: [A/B/C + before/after deletion + format + other-player filtering]
D-009 deletion: [per-family Delete/Sever/Retain from the table above, with durations]
D-009 disclosure copy owner: [Marco; legal review by ...]
D-015 retention: [per-family purpose/trigger/duration/cleanup owner; items 1–6]
Legal review: [who, when; no public claim before it]
```

After a ruling, an Account/retention RFC turns it into routes, migrations and jobs with
failing-control witnesses, and R-003/R-007 test the built workflow. Until then no deletion or
export claim beyond current behavior may be made.
