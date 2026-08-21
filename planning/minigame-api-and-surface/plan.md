# Minigame API & Surface implementation plan

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
- [ ] Implement surface components after UI Foundation, per MA-C9.
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
