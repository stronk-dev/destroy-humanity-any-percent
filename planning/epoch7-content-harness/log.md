# Epoch-7 content-dynamics harness implementation log

Append-only. This is the successor evidence lane registered by the First Content Epoch sign-off.

## 2026-08-14 — candidate-runway acceptance bounce: EH-C11–EH-C14

- Resumed at the owner-ratified T0–T1 bytes. The existing lane owns only the empty registry and
  immutable snapshot generator/validator; no content-dynamics scenario parser, four-arm runner,
  candidate report mode, or report schema exists.
- The mint cannot yet be represented honestly: T01-C14 names a singular `relevance` artifact but
  the repo has two ratified policies and no cycle-free replay-bundle owner; the only promotion
  loader hardcodes the epoch-6 cardinality of sixteen; and the scenario/report bytes remain
  directional rather than exact.
- The single Active-Play pair is additionally underdetermined because the pinned catalog can draw
  a non-window Lucky effect and a fresh Company has no owned building target. A favorable seed or
  synthetic pending row would violate EH-C3.
- Filed EH-C11–EH-C14 in the RFC with concrete proposed contracts. No candidate artifact, epoch
  coordinate, production balance byte, or mint authority was changed. The relevance H-A/H-B LOW
  closures are independent and ready for their designated review.

## 2026-08-10 — implementation bounce: current harness cannot own the four mechanics

- Audited the active harness rather than treating “full bundle loaded” as “full bundle simulated.”
  `harness.Suite` passes only economy, Routes, and Commons into Company `Transition`; its two
  policies cannot execute a Founder Fiscal command, Active-Play schedule, Pitch tenant session, or
  platform payout.
- Adding four milestone rows would therefore be false coverage. Reimplementing the missing math in
  the harness would be worse: it would create the second authority the ApplyLogged and tenant
  boundaries were built to prevent.
- Filed EH-C1–EH-C7 in `rfc/first-content-epoch.md` with concrete proposed contracts: one
  full-bundle production-owned simulation lane, a closed four-arm grammar, literal policies and
  cardinality, an observation/invariant split, the existing separate-commit baseline discipline,
  and pinned-epoch identity.
- No balance bytes, baselines, or runtime semantics changed. Work is blocked only on owner rulings;
  the rest of the forward batch remains independently reviewable.

## 2026-08-10 — implementation audit: active content and historical bytes are absent

- Confirmed the epoch-6 seed has sixteen artifacts but no `opportunities` row, and the minted
  economy catalog has no `active_play` multiplier declarations. The exact replay-catalog load
  therefore has no Active-Play policy. Using the reviewed fixture would violate the ruled
  full-bundle identity instead of measuring live content.
- Confirmed an epoch-seed path alone cannot satisfy EH-C6 after a future mint: its artifact paths
  resolve mutable production files, while historical accepted hashes retain identity but not the
  bytes needed to rerun the old golden.
- Filed EH-C8 with a runner-first/fixture-only implementation lane and a hard requirement that the
  first four-arm golden wait for an epoch that actually pins opportunities. Filed EH-C9 with a
  generated, full-set, content-addressed bundle snapshot contract so every registered golden stays
  executable under its original bytes.
- No runner, golden, registry, production artifact, or balance baseline was authored past these
  unresolved identity contracts.

## 2026-08-10 — implementation handoff: EH-C9 snapshots + EH-C8 empty registry

- **Review by:** Codex self-review only. **Recorded by:** Codex. This is ready for designated
  review as a partial implementation slice; it does not claim the four-arm runner or a golden.
- Added the strict five-coordinate `content_dynamics.v1` registry with zero production entries,
  exactly as EH-C8 requires while epoch 6 has no Opportunities owner. `make content-harness` is
  therefore an honest no-op rather than a skip or synthetic-policy lane.
- Added generated, immutable full-bundle snapshots. Generation accepts only the complete active
  `epochseed.Bundle` at its accepted current epoch, writes content-addressed bytes once, and
  refuses a later rewrite. Loading verifies registry/manifest coordinates, accepted epoch hash,
  raw-byte-sorted declarations, per-file SHA-256, exact directory set, and recomputed bundle hash.
- Adversarial tests cover missing, extra, tampered, manifest-rehashed, hand-subsetted, and rewrite
  attempts. The baseline guard now discovers future content-dynamics golden paths from registry
  history, and its accepted input surface includes the governed content-dynamics tree.
- Normal repository commands passed: focused cold `make test-go
  GO_PACKAGES='./harness ./cmd/balance-harness' GO_TEST_FLAGS='-count=1'`, `make content-harness`,
  and read-only `make harness-check`.
- Still open in this RFC: the production-owned four-arm simulation seams, strict scenario/report
  grammar, literal-cardinality runner, and the separate first golden after a real epoch pins
  Opportunities. No balance artifact or golden was created.

## 2026-08-14 — OWNER RULING: EH-C11–EH-C14 all ACCEPTED as proposed

Ruled by Marco (recorded by Claude from the owner's explicit per-contract selections, each made
with its alternatives presented):

1. **EH-C11 ACCEPTED.** One minted `relevance` artifact: the cumulative T1 schema-v2 policy,
   byte-for-byte, at `balance/relevance/t0-t1.json`. The T0 policy and both scenarios remain
   harness inputs. The strict decoder/validator moves to a cycle-free `server/relevancepolicy`
   package consumed by both harness and replaycatalog; the parsed policy joins `CatalogBundle`.
   Epoch 7 = exactly eighteen sorted artifacts. (Mint-both REJECTED as redundant weight;
   mint-neither REJECTED — it would reopen T01-C14.)
2. **EH-C12 ACCEPTED.** `planning/t0-t1-content/promotion-manifest.candidate.v1.json` in the
   shipped manifest-v1 grammar, status `ratified`, the exact eighteen rows; carryover rows keep
   production hashes and consumed verdicts; loader cardinality generalized; epoch-6 test still
   pins sixteen, a new epoch-7 test pins the exact set, hashes, and recomputed bundle hash.
   (Grammar v2 REJECTED — churn without need.)
3. **EH-C13 ACCEPTED.** Content-dynamics scenario/report schema v1 exactly as proposed: strict
   roots, byte-sorted run IDs and observation rows, canonical Decimal-string values, candidate at
   `testdata/harness/content-dynamics/scenarios/epoch-7-candidate.v1.json`; the derived budget and
   the scenario bytes + hash return for owner pinning before any golden is accepted. (Reusing the
   relevance schema REJECTED — couples unrelated contracts.)
4. **EH-C14 ACCEPTED.** One pinned deterministic founder/run coordinate whose first natural draw
   is duration-bearing, with the named minimum owned target; the engine performs the natural
   spawn/claim and fails loud on mismatch; the harness never manufactures an opportunity. The
   exact coordinate and target return with the scenario hash for owner acceptance. (Runtime seed
   search REJECTED — adapts to content changes instead of failing on them; vacuous-run tolerance
   REJECTED — fail-loud law.)

**The epoch-7 candidate runner is unblocked.** Deliverables back through designated review:
the runner + candidate mode, the eighteen-row manifest, the scenario bytes/hash and C14
coordinate for owner pinning, then the EH-C10 candidate report toward the mint.
