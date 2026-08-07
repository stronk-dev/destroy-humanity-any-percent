# Minigame API & Surface implementation plan

- [x] Compose the minigame repository, closed Pitch tenant registry, pinned content resolver, and
  platform service in `gameserver.Compose`; expose it on `Composition`.
- [ ] Attach the authenticated minigame coordinator API after MA-C10–MA-C14 are ruled.
- [x] Implement Founder-v21 replay-owned create sequencing, atomic session creation, and create
  request-receipt idempotency.
- [x] Implement transaction-scoped command-receipt persistence primitives.
- [ ] Wire the typed command handler, terminal auto-resolution receipt, and current-session read.
- [ ] Prove create → command → terminal auto-resolve → retry/current over the composed real socket.
- [ ] Register/generate the recovery and minigame operation schemas through API Foundation.
- [ ] Implement surface components after UI Foundation, per MA-C9.
- [ ] Hand the complete range off for designated cross-party review; do not self-archive.
