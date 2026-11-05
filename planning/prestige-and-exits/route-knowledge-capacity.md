# RP-771 — eligible Wind Down fails at Route Knowledge capacity

**Status:** verified current defect; no remedy adopted. Product/test baseline
`6827120c4b411f64a8a7a8c714fbe5ccd55e79a9`,2026-11-05. Found during independent
review by Codex `/root/prestige_lifecycle_remediation_review`; recorded by Codex root.

## Executed counterexample

The existing real-Postgres `TestPrestigeWindDownEligibleStateMatrixIntegration`
creates valid persisted Founder/Company streams with Tier1 and one prior Exit.
A private test-only extension sets `RouteKnowledgeBalance=9007199254740991`,
the admitted exact-integer maximum, then submits the otherwise-eligible
`wind_down` with current revisions. A companion also owns all seven declared
hints; Routes D5's only spending action cannot free balance there because an
already-owned hint refuses before debit.

Both return `invalid production engine state` rather than applying Wind Down.
`finishExitResolved` refuses the positive Route Knowledge grant above remaining
headroom. Original six matrix controls still pass. Both failed cases retain
identical nonempty encoded Founder/Company state, unchanged revisions and zero
matching event, intent-record and run-log writes. This is a direct AC5 failure,
not corruption or a new assertion that every other Exit path is impossible.
The fixture does not establish natural reachability at current balance values.

Independent first probe82182 reproduces the failure. Full matrix63330 confirms
both cap cases and six controls, but ignores encoding errors: do not use it as
complete byte-rollback proof. Corrected59962 requires all four encodings to
succeed and be nonempty, then repeats the same eight cases: **two failures/six
passes**, with full rollback checks actually executed. No main-tree mutation.
Actual Go/Postgres images are Linux-amd64 under local Desktop emulation; this
is not native supported-host or browser evidence. All handles terminal; owned
Postgres removed, pre-existing network/caches preserved.

## Retained reproduction

The [small test-only patch](route-knowledge-cap-repro.patch) retains the executed
counterexample rather than adding an unexplained failing test to ordinary CI.
It asserts successful Wind Down and is intentionally **red** on current behavior;
it does not turn the defect into an expected-success observation.

SHA-256 bindings:

- Original `server/production/prestige_integration_test.go`:
  `2cca513b295ce9e35d9b8db7bf886c33d0ae74d22cd68312d9694d1f0fce2c2e`.
- Retained patch: `323ce066820846046cbf018edf3d740cc8b9c34f13a7639b8087ca31ad362548`.
- Executed private source: `03cfafc19e761beede570c3c82f989970c46569f97088061b3cd70f845c47dd6`.
- Corrected result log: `bcfc6a4d6a132135472fb46a8bc4e37c9305818926927c6bb7452918fe717a1a`.

Raw source/overlay/log and manifest remain in ignored
`.cache/prestige-remediation-review/`. To reproduce without editing tracked Go:
check the original source binding, copy that file to
`.cache/rp771-repro/server/production/prestige_integration_test.go`, then apply
the patch there with `git apply --directory=.cache/rp771-repro
planning/prestige-and-exits/route-knowledge-cap-repro.patch`. Create an overlay:

```json
{"Replace":{"/workspace/server/production/prestige_integration_test.go":"/workspace/.cache/rp771-repro/server/production/prestige_integration_test.go"}}
```

Save it as `.cache/rp771-repro/overlay.json` and use the existing declared service:

```sh
make test-save-integration SAVE_TEST_PACKAGES=./production SAVE_TEST_FLAGS='-overlay /workspace/.cache/rp771-repro/overlay.json -run TestPrestigeWindDownEligibleStateMatrixIntegration -v' SAVE_TEST_COUNT=1
```

These are private test fixtures, not compiler/module caches. If source has changed,
rebase the reproduction explicitly; do not claim the old executed binding applies.

## Required contract resolution

Prestige D1/AC5 requires eligible Wind Down; Routes D5/C6 specifies earning and
permanent hint spending but no overflow disposition. P2c explicitly saturates
**Reputation**, not Route Knowledge. The ledger's positive-production saturation
exception does not establish a scalar reward policy; ordinary grants remain
strict. P3/AC2 also promises at least the offered payout after further progress.

A repair must therefore settle Route Knowledge representation/capacity and
preview/grant semantics, including curriculum bonuses, replay, projection and
other grant producers. Preserving all grants through wider exact representation
and adopting a visible bounded balance with explicit grant/preview rules are
different contracts. No choice is inferred here. Do not silently clamp, delete
the valid-state case, narrow save validity or call hint spending a universal escape.
After the contract is resolved, integrate the retained regression with the
implementation, exercise real persistence/replay and both runtimes, and obtain
independent review. RP-038/full acceptance and RP-771 remain open.
