# Production implementation — round-2 review remediation

Findings from the two-lens review of `4479cd7`..`b1fe65a` (spec-compliance + adversarial).
This job exists to fix them; the log below is the authoritative findings record, in priority
order. R1/R2 are demonstrated permanent-brick bugs and go first; R3 is a catalog/progress
operator divergence (the initial primitive diagnosis was corrected in the log). R4–R8 are
coverage and integrity debt.

## Remediation sequence

1. Review/accept `rfc/production-hardcap-saturation.md`; implement R1 and its near-cap property
   corpus.
2. Review/accept `rfc/millisecond-cursor-canonicalization.md`; implement R2 as save version 4.
3. Review/accept `rfc/resource-log-domain-parity.md`; implement the corrected R3 diagnosis without
   changing Decimal division semantics.
4. Specify and land R4–R8 as a separate contract-assertion/integrity follow-up after the three
   correctness repairs are green.
5. Resume the accepted Balance Harness Foundation only after R1–R3 are archived into canonical
   docs, so its first baseline cannot encode a known brick or parity divergence.
