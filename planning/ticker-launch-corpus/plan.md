# Ticker launch corpus — current implementation boundary

Authority: accepted `rfc/ticker-launch-corpus.md`. Its owner-acceptance exception explicitly
blocks implementation until Feed & Dispatch is narrowed under VQ-3. No ticker runtime,
production copy, content mint, API operation or UI is authorized by this plan.

Before construction, the RFC author must also reconcile TL4/TL-AC4 (RP-479): its literal
fixed-length warm-up does not establish the required overlap consistency. The retained
diagnostic tests unchanged inputs at the recommended 12-slot/64-entry settings, with 14
unconditional ambient lines. It distinguishes internal no-repeat from overlapping-window
agreement. This is a specification counterexample, not a replacement selection algorithm.

Next sequence:

1. Author reconciles Feed's server-only/numeric-only dispatch contract with ruled VQ-3/OD-13.
2. Author corrects TL4 to satisfy the existing determinism, no-repeat, bounded delivery and
   stateless/privacy requirements, with explicit cross-window vectors. Do not weaken TL-AC4
   or add stored selection/impression history to make the old algorithm pass.
3. After the named blockers are resolved and reviewed, implement TL1–TL4 grammar and both
   runtime selectors with shared positive/negative vectors; then generated API, single world
   emitter, client fallback/accessibility and composed player proof under TL5–TL10.
4. Owner-authored/adopted copy and production activation remain separate release-content gates.

Probe commands and results belong in `log.md`. Do not add this expected-red diagnostic to CI
or call its self-test a green ticker acceptance result.
