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
- Retained zero-context patch: `2b30fd93c981d843f72646f914a3eaaafa6d5100bf4b4ed1f741a50d520d3f24`.
- Executed private source: `03cfafc19e761beede570c3c82f989970c46569f97088061b3cd70f845c47dd6`.
- Corrected result log: `bcfc6a4d6a132135472fb46a8bc4e37c9305818926927c6bb7452918fe717a1a`.

Raw source/overlay/log and manifest remain in ignored
`.cache/prestige-remediation-review/`. The reviewer's original context-bearing
patch has SHA`323ce066820846046cbf018edf3d740cc8b9c34f13a7639b8087ca31ad362548`;
its diff context markers trigger Git's space-before-tab check. The retained
zero-context representation applies to the bound source and reconstructs the
identical executed private-source SHA above; no assertion changes or guard
exceptions. To reproduce without editing tracked Go:
check the original source binding, copy that file to
`.cache/rp771-repro/server/production/prestige_integration_test.go`, then apply
the patch there with `git apply --unidiff-zero --directory=.cache/rp771-repro
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

## 2026-11-11 — Positive stored offer and projection boundaries

At source `d66e4ca1`, the existing Wind Down/scripted Exit and projection capacity
populations passed cold on real Postgres, including the projector's current
overflow refusal. Production passed in 1.043s and Route Projection in 0.274s.
The [draft remedy](../../rfc/route-knowledge-capacity.md) preserves the distinction
between the Exit's saved balance and the projector's independently summed facts.

An overlay of `TestPrestigeWindDownAndScriptedExitIntegration` used the existing
offer-age population. After decoding its stored offer and before writing the
progressed Company, set `stored.PayoutPreview.RouteKnowledge = 10` and marshal
`stored` back into `offered.State.OfferState.TermsJSON`. In the `fresh` case,
load and write the Founder with balance `9007199254740989` using the new revision
for `accept_exit_offer`. Leave the `one_millisecond` Founder's balance unchanged.
Run just those subtests using the Compose Go overlay mechanism above. The fresh
offer failed with `invalid production engine state`. The control applied and
the original committed-payout assertion verified the positive floor. This is
admitted stored input. Newly generated acquisition/acquihire offers carry zero.
The raw overlay and result remain in ignored `.cache/rp771-offer-capacity/`.

- Unmodified suite log SHA-256: `60d1f112b43ef488400780cfcfafe0ea8e9d9986f26deddee3682fa679565f3f`.
- Overlay source SHA-256: `97a7d0ac86992a3adf51b4ae83f12e01d333bf428e56aec1ba3a186112c36546`.
- Overlay result SHA-256: `0ebc2140501e12e37684df7ebf973f61bc8d732f466703720d4b4f36863e8847`.

With maximum `M = 9007199254740991`, a Node `BigInt` model started at `M−2`
and applied a credit of 10 and debit of 50 in both orders. Exact addition yields
`9007199254740949` in both. Clipping each arriving credit at `M` yields
`9007199254740941` credit-first and `9007199254740949` debit-first. This rules
out clipping callbacks as an implementation of a deterministic cap. In the
same model `M+M` is `18014398509481982`, and JavaScript `Number` aliases `M+1`
with `M+2`. These are arithmetic checks. The acceptance population still needs
an actual scripted Exit whose combined grant exceeds `M` and actual replayers
and SQL consumers that preserve the amount.
