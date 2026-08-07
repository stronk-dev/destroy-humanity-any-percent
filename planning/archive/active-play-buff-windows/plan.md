# Active-Play Buff Windows implementation plan

RFC: `rfc/active-play-buff-windows.md`

- [x] Resolve A1-A8 and reconcile the normative body in the same ruling edit (proved across the
  designated-review range `32e5a63^..d3b18ef`, which includes the two pre-restart Active-Play
  commits omitted from the original handoff).
- [x] Resolve implementation blockers A9-A16 and reconcile their exact catalog, transition, and wire
  contracts.
- [x] Implement the exact opportunities artifact and Company-v18 activation/save state.
- [x] Implement the ruled attended-time scheduler and deterministic spawn/effect selection.
- [x] Implement claim, payout saturation, buff contributions, expiry, and manual-action binding.
- [x] Add exact Go/TS wire/event/replay vectors and migration corpus.
- [x] Remediate the designated verdict over `32e5a63^..d3b18ef` plus `45944ca`, then obtain the
  mandatory cross-party re-review over that full span and the remediation range before writing
  canonical docs and archive only if that verdict approves the full range.

Current state: implemented and archival-eligible under the designated round-3 verdict `e9ebb0c`.
The archival commit cites the complete reviewed union `32e5a63^..d3b18ef` + `45944ca` +
`1f5a2f8^..16eb935` + `45f082d` and adds `docs/active-play.md` as the canonical description.
