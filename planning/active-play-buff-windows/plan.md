# Active-Play Buff Windows implementation plan

RFC: `rfc/active-play-buff-windows.md`

- [ ] Resolve A1-A8 and reconcile the normative body in the same ruling edit (rulings exist;
  implementation proof remains pending with A9-A16).
- [ ] Resolve implementation blockers A9-A16 and reconcile their exact catalog, transition, and wire
  contracts.
- [ ] Implement the exact opportunities artifact and Company-v18 activation/save state.
- [ ] Implement the ruled attended-time scheduler and deterministic spawn/effect selection.
- [ ] Implement claim, payout saturation, buff contributions, expiry, and manual-action binding.
- [ ] Add exact Go/TS wire/event/replay vectors and migration corpus.
- [ ] Run normal root gates and obtain both mandatory full-range reviews before docs/archive.

Current blocker: A9-A16. A1-A8 settle the gameplay direction, but the shipped Company transition
cannot yet represent the promised scheduler/rejection behavior, and the artifact, save, replay,
saturation, contribution, and activation shapes are not byte-exact. No Active-Play mechanic code has
started.
