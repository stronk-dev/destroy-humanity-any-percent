# Meters Foundation implementation log

## 2026-08-03 — acceptance reconciliation

The owner declared all six drafted foundations accepted, including Meters C1–C12. The worktree
contained only the earlier C1–C7 summary while retaining stale contradictory draft prose and open
C8–C12 proposals. Reconciliation made the accepted contract singular:

- exactly ten independent Trust standing/grievance axes plus Company p(doom);
- Externality remains addressed ledger facts and Founder Soul remains the existing read-only carry;
- integer `[0,100]` values, exact attended-time arithmetic, causal input arms, and derived bands;
- band events never become moral acts; save v15 owns complete value/remainder maps;
- meters joins pinned epoch/replay identity after Purchasable Content; not-spendable is structural.

`DESIGN-GAP:` no literal production band IDs/floors, initial values, decay rates, or seed input
bindings have been owner-supplied. Per house law, implementation will use discriminating test
catalogs and will not improvise the epoch artifact. The mint and archive remain blocked on those
balance-data rows; foundational engine work proceeds.

## 2026-08-03 — catalog/schema foundation

- Added strict JSON Schema, Go, and TypeScript authorities for exactly eleven Company meters: ten
  Trust axes and p(doom). Externality/Soul/public-axis rows, wrong scopes, decorative spendable
  flags, duplicate sources, zero inputs, invalid bands, and resource-ID collisions reject.
- The closed input union contains only ledger-fact and contribution-slot arms. Slots validate
  against the implemented multiplier vocabulary; every numeric field remains integer `[0,100]`.
- Added a fail-closed package import boundary preventing meters from importing economy or
  production state owners. The boundary's negative fixtures run in the normal root verification.
- `make test-go GO_PACKAGES='./meters'`, `make typecheck`, `make test-client` (6,509 assertions),
  `make verify-schema`, and `make verify-meters-boundary` pass from the repository root.
- No production meter artifact was invented. Schema verification labels its fixture `pre-mint`.

## 2026-08-03 — catalog-foundation full gate

`make verify` passed at `8f5263a`: Go vet and every server package; formulas and balance harness;
strict TypeScript/Svelte checks; production client build; 6,509 client assertions; kernel/copy/
schema/package-boundary gates; and 19,536 browser assertions.

## 2026-08-03 — deterministic transition kernel

- Added matching Go and TypeScript state authorities with exact complete value, decay-remainder,
  and input-remainder key sets.
- The pure transition advances decay first, then causal facts and active contribution sources,
  clamps once, clears decay phase at its target, and emits only a prior-to-final band change.
- Added one shared JSON corpus consumed by both runtimes. It discriminates split attended time,
  negative rate input, ledger-fact band crossing, linear decay, stale target phase, and hard-bound
  saturation.
- Focused Go, client, and TypeScript type-check gates pass using the normal repository toolchain.
- Save v15 and `ApplyLogged` binding remain intentionally uncommitted until the meter catalog can
  be part of the pinned artifact bundle; accepting an optional/unhashed runtime catalog would
  violate replay identity rather than make progress.
