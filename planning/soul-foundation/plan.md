# Soul Foundation implementation plan

RFC: `rfc/soul-foundation.md`

- [x] Resolve SB1-SB9 and reconcile the normative body in the same ruling edit.
- [x] Resolve implementation blockers SB10-SB16, including once-only state and the Company
  suppression authority.
- [x] Wait for the exact Fiscal v19 artifact/activation chain before enabling Soul v20.
- [x] Resolve implementation blockers SB17-SB23: literal artifact enums, debit errors/events,
  recovery persistence/commands, the suppressed transition, v20 activation bytes, consumer schema
  versions, and cross-stream event order.
- [x] Implement the strict Soul artifact, dormant-to-active save transition, and pure band/gate API.
- [ ] Implement the ruled atomic debit component and real recovery activity coordinator. The debit
  component and recovery persistence/replay/atomicity are complete. SB24 selects a progress
  heartbeat, but its token acquisition, watchdog transaction entrypoint, and exact catalog/wire
  shapes remain unresolved by the accepted text (SB25-SB27 in the implementation log).
- [x] Add pet/minigame consumers, exact wire/event/replay vectors, and migration corpus.
- [ ] Route the Trust/Soul correlation and ending-copy obligations to named owners.
- [ ] Run normal root gates and obtain both mandatory full-range reviews before docs/archive.

Current blocker: SB24 chooses the missing heartbeat, but the command cannot yet be implemented
without inventing three public/persistence contracts. The request requires a `claim_token`, while
start/reconnect never issues or renews one and the existing claim token is transaction-internal.
Lazy watchdog cancellation from an ordinary command cannot run inside the existing Company-locked
guard without reversing the Founder-then-Company order. Finally, the exact Soul artifact policy and
progress/start/watchdog receipt keys were not reconciled into the SB10/SB19 schemas. SB25-SB27 must
be ruled before a production recovery row can mint.
