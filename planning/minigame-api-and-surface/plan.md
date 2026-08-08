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
- [ ] Prove create → command → terminal auto-resolve → retry/current over the composed real socket.
- [x] Register/generate the recovery and minigame operation schemas through API Foundation.
- [ ] Implement surface components after UI Foundation, per MA-C9.
- [ ] Hand the complete range off for designated cross-party review; do not self-archive.
