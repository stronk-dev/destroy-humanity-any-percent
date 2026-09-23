# D-008/D-009/D-015 — current browser-persisted data

Research checkpoint: 2026-09-23, source `a6956de` plus this planning-only
checkpoint. Method and controls: [`data-rights-browser-plan.md`](data-rights-browser-plan.md).
This is a source and unit-test map, not a browser-profile forensic claim, a
privacy notice, or an adopted export/deletion/retention rule.

## Storage API and exact keys

Production client source uses `localStorage` through `GameUIRuntime` and
`GameUIApp`/`timing.ts`. A search of current production source found no other
client-managed `sessionStorage`, IndexedDB, Cache API, cookie or service-worker
storage call. That source negative does not rule out browser HTTP cache or
operator-side persistence. Runtime storage is injectable in unit tests; in
the browser it defaults to the origin's `localStorage`.

| Key | Exact stored shape; writer → reader | Clear/expiry and account boundary |
|---|---|---|
| `cloud-clicker.bootstrap-key.v1` | Random 32-byte hex idempotency key, written before `POST /api/v1/bootstrap`; the next bootstrap attempt reuses it. | Removed only after a successful bootstrap response and credential write. Failed/interrupted bootstrap can leave it indefinitely in the same browser profile. It is a retry key, not a recovery credential. |
| `cloud-clicker.credentials.v1` | JSON `{accessToken,refreshToken,accountID,recoveryCode}` from the bootstrap response. `hasCredentials`, authenticated snapshots/intents, and socket connect read it. | Overwritten by another successful bootstrap; no runtime logout, account-deletion cleanup or expiry removal path. Parsing invalid bytes returns absent without deleting them. `recoveryCode` is kept in the origin's localStorage even though the ruled player posture is one-time display/copy/download at bootstrap. No current Game UI code reads the value to display/copy/download it (RP-123). A token expiring server-side does not remove these local bytes. |
| `cloud-clicker.transport.v1.<founderID>` | JSON channel-to-`{epoch,offset}` map, limited on read to `player:<founderID>` and `world`; written on socket subscription/publication. | Removed on authoritative resync for that Founder. Normal unsubscribe, credential expiry, account deletion and profile reuse do not remove it. The key and player channel contain the Founder UUID; the world position is stored inside that same Founder-keyed document. |
| `cloud-clicker.timing.v1.<founderID>` | JSON `{schema_version:1,records:[{category,founder_id,pb_rta_ms,run_seq,splits:[{gate_id,rta_ms}]}]}`. Run-end UI writes it; later-run personal-best display reads it. | No `removeItem` capability in the timing storage interface and no cleanup/expiry path. It remains after server account deletion or changing accounts in the same browser profile. Corrupt/unreadable bytes are treated as empty, but are not removed. The Founder UUID, run category and split timings are player-linked local history, not anonymous statistics. |

`latestSnapshot`, cursor, timer, active socket and UI state are in memory and
not additional persistent browser stores. Unsubscribing stops the socket but
does not clear stored positions. A server-side `DELETE /api/v1/account` has no
mechanism to cascade into an origin's localStorage; the current Game UI has no
mounted player delete/export/recovery action that would perform local cleanup.

## Evidence and limits

`make test-client` passed cold: 39 test files passed, 2 skipped; 6,662 tests
passed, 22 skipped. The Game UI runtime test observes the retry-key write,
credential write and successful retry-key removal. Transport tests check
position persistence/resync. Timing tests check schema parsing, split ordering
and personal-best lookup. None executes account deletion with all four keys
populated, proves a one-time recovery-code player workflow, or verifies
logout/profile-switch cleanup. Green unit tests therefore do not promote
these rights outcomes.

The recovery-code issue is concrete: the server and owner record say the code
is returned once for the user to copy/download; current Game UI stores it
silently as a required `GameUICredentials` field and has no display/transfer
consumer. RP-123 tracks the mismatch. This research does not choose how a
future accepted Account/UI contract should handle loss of device storage or
whether an optional explicit device-only retention feature is permissible.

## Decision handoff

- **D-008:** choose whether and how browser-only timing/splits are exportable,
  and distinguish them from server-authoritative save/history export. Transport
  offsets and credentials are not automatically portable player data.
- **D-009:** author deletion behavior and disclosure for these four origin keys,
  including a same-device account switch and a remote deletion where no browser
  is available to clean storage. A server deletion response alone cannot prove
  client cleanup.
- **D-015:** choose purpose and lifecycle for bootstrap retry state, active
  tokens, one-time recovery code, Founder transport positions and personal-best
  timing. The exact client change belongs to an accepted Account/Game UI RFC,
  not this inventory.

Backups, journals, metrics and operator artifacts remain separate non-DB
surfaces. Nested game-state/receipt/event payloads remain separately open.
