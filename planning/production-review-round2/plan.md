# Production implementation — round-2 review remediation

Findings from the two-lens review of `4479cd7`..`b1fe65a` (spec-compliance + adversarial).
This job exists to fix them; the log below is the authoritative findings record, in priority
order. R1/R2 are demonstrated permanent-brick bugs and go first; R3 includes a primitive-level
Go/TS divergence with blast radius beyond its trigger. R4–R8 are coverage and integrity debt.
