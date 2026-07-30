# Guild Model — implementation log

## 2026-07-30 — accepted and started

- The independent Faction remediation review approved FB-1/F2a and the LOW batch, so Guild's GA
  parent contract is implemented and archived. The owner explicitly placed Guild next.
- GB supplies the complete strict Phase-0 catalog; GC supplies a unique integer-only clearing
  answer, including the non-redistribution rule and labeled NPC counterparty. No balance mechanics
  need to be inferred.
- Implementation order is catalog/epoch → storage/lifecycle → tithe/Health/exchange → transport
  resolver → canonical docs and full verification. Per-change commits remain behind the mandatory
  independent review gate before archival.
