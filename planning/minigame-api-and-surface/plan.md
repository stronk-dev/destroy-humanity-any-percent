# Minigame API & Surface implementation plan

- [x] Compose the minigame repository, closed Pitch tenant registry, pinned content resolver, and
  platform service in `gameserver.Compose`; expose it on `Composition`.
- [ ] Attach the authenticated minigame coordinator API after MA-C10–MA-C14 are ruled.
- [ ] Implement replay-owned session sequencing and request-receipt idempotency.
- [ ] Prove create → command → terminal auto-resolve → retry/current over the composed real socket.
- [ ] Register/generate the recovery and minigame operation schemas through API Foundation.
- [ ] Implement surface components after UI Foundation, per MA-C9.
- [ ] Hand the complete range off for designated cross-party review; do not self-archive.
