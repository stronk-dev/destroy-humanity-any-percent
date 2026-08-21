# RFC: Leaderboard Readers & Player Surface

- **Status:** draft — not implementation authority
- **Author:** Codex (drafted from the Leaderboards lifecycle audit; owner acceptance required)
- **Design refs:** `design/05 §6`, `design/08 §6`
- **Depends on:** Leaderboards & Balance Epochs, API Foundation, Game UI, Route Registry; D-017 for
  any player-authored/Exhibition scope

## Purpose

Turn the existing verified-run/epoch backend into a real player capability without inventing a
second API client or calling storage primitives a shipped surface.

## Required scope before acceptance

1. Registry-owned generated public readers for categories, current/historical epochs, exact board
   pages/cursors, verified run archives, and their pinned catalog bytes.
2. A Svelte player surface for board/category/variable/epoch browsing, historical navigation,
   exact ties, loading/empty/error/refusal states, and the Route Registry alongside records.
3. A browser workflow that retrieves an archive plus pinned catalog, invokes the bundled
   TypeScript verifier, displays the six-verdict result, and refuses missing/mismatched evidence.
4. World-first dispatch/feed production and consumption, plus a separately keyed `machine` class,
   only if retained in the accepted milestone.
5. Accessibility, privacy/cache policy, pagination stability, provider-off behavior, current real
   data, and producer-severing acceptance proofs.

## Explicit exclusions

- Player-authored predicates, promotion and Exhibition remain blocked on D-017 and their rights/
  moderation contract.
- Mandate mechanics/content require a separate accepted gameplay/balance RFC.
- Abandoned-run/archive retention belongs to D-015 and the accepted operations contract.
- This draft authorizes no implementation until its exact operation schemas, UI states, copy keys,
  dependencies, and acceptance matrix are reviewed and the status is explicitly changed.

## Open acceptance work

- Choose the Phase-0 preview subset of readers/surfaces.
- Enumerate exact API operation IDs, filters, headers, cursor and archive byte contracts.
- Enumerate mounted routes, components, copy keys, keyboard/focus behavior and failure states.
- Define composed API→browser→archive→validator and Route Registry/record witnesses.
