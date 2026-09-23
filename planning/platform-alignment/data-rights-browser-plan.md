# D-008/D-009/D-015 — browser-persisted rights inventory plan

Status: predeclared research only; no player-data policy or product edit.
Coordinate: source at current HEAD, starting from game-table SQL-column map
`a6956de`.

## Question and population

Enumerate every production browser-persisted value (localStorage,
sessionStorage, IndexedDB, Cache API, cookies if client-managed) used by the
current client/site code. For each: exact key/schema, writer, reader, expiry or
clear path, account/Founder binding, and whether account deletion, logout,
credential replacement or browser profile reuse actually removes it. Include
transport resume positions and speedrun timing, not just credentials.

## Method and controls

1. Search all production client/site paths for storage APIs and key strings;
   inspect complete read/write/clear call chains and relevant tests. Classify
   ephemeral memory separately from persistent browser state.
2. Execute the nearest unit/browser tests cold where they cover persistence.
   If no test performs account deletion with populated local state, report
   that as a gap rather than promoting a source trace to integrated proof.
3. Positive control: retained credentials/recovery code or Founder timing must
   not be labeled anonymous. Negative control: removing one key is not proof
   all same-account browser state is gone; server-side deletion is not a
   browser storage cascade. Check for any storage API outside localStorage.

## Exit and refusal

Publish a bounded browser-surface map, update the parent inventory, backlog for
new contradictions, D-008/D-009/D-015 evidence, 1.0 board and append-only logs.
This may inform owner decisions but cannot choose export scope, retention,
player copy or a legal conclusion. No product/test behavior, accepted RFC or
authored content changes in this wave.
