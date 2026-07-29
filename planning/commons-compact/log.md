# Commons Compact — append-only log

## 2026-07-29 — acceptance and start

- Re-read RFC-0000, AGENTS.md, the complete Commons RFC, and the normative research formula/starting-value sections.
- Owner direction to implement the RFC queue is explicit acceptance. C1-C8 close the draft's executable gaps without changing its design: bounded ratios are integer ppm, large values remain RFC-0001 Decimals, and every provisional value is balance data.
- Marked the RFC implementing. No push will be performed.

## 2026-07-29 — pure catalog and numeric boundary

- Added the strict Commons catalog in Go/TypeScript and shipped the research starting values as data. Bounded ratios are ppm; the multiplier remains canonical Decimal.
- Implemented source-derived Enclosure and the Health/Solidarity modifier in both runtimes with shared edge vectors. Non-members return an absent contribution; members return exactly `commons.member` in the fixed Commons slot.
- Declared the one Commons multiplier source in the economy catalog and added schema/cross-catalog checks plus a compile-enforced two-way production boundary.
- Focused Go, TypeScript, 6,396 Node tests, schema verification, and the import gate pass. Review correction: the economy schema's `oneOf` made target `all` match both alternatives; changed it to `anyOf`, matching the already-canonical Go/TypeScript loader behavior.

## 2026-07-29 — save v6, membership intents, and cohort assignment

- Added save v6 company membership/tithe/Solidarity state with strict hourly sample validation and a corpus-backed v1-v5 migration to empty non-member state.
- Added strict `sign_compact` and `leave_compact` intent parsing, accrue-before-change semantics, typed repeat/non-member rejections, revision-tied membership events, and leave-time Solidarity erasure. Direct transition tests cover sign, duplicate sign, leave, and clean re-sign.
- Generalized post-commit event projection to support multiple independent projectors without coupling production to Commons.
- Added migrations and an idempotent Commons projector. Founder cohort assignment is server-context-derived, protected by a per-server/bracket advisory transaction lock, fills the oldest cohort below its catalog target, and persists across runs/leaves.
- Real Postgres concurrency tests prove two simultaneous first signs share one eligible cohort, replay does not inflate membership, leaving updates only run membership, and founder cohort identity remains stable.
- Batch review found no unresolved correctness issue. All Go packages and the full Postgres integration target pass.

## 2026-07-29 — lazy Health/Capacity pipeline and population gate

- Added a neutral post-accrual hook contract. Production calls the hook but imports no Commons code; the composition adapter derives Enclosure from accepted multiplier contributions, advances hourly 30-day Solidarity buckets, computes tithed Capacity from committed accrual receipts, and emits `compact_sampled`.
- Added idempotent sample projection, member/cohort/server read models, labeled NPC denominator weight below 40 members, guildless effective-Health fallback, asymmetric closed-form Health smoothing, and the `commons.member` contribution provider. The first sample uses the declared neutral NPC Health; subsequent intents consume the projected snapshot.
- Added explicit collapse merge. Founder/member/sample rows move under one advisory lock, source cohorts close, and target standing is recomputed from member numerator/denominator inputs rather than averaged from rounded Health.
- Extended the production integration test through sign → member accrual → sample projection → Health snapshot; the real-Postgres suite and replay checks pass.
- Added the SplitMix64 population harness gate over 128 seeds. Mean modifiers at 200 vs 20,000 members remain within the shipped 100-ppm bound and their 95% intervals overlap.
- Formula generation now fingerprints the executable Enclosure, modifier, aggregate-Health, and smoothing authorities and publishes the Commons formula family plus the exact source-weight table from balance data.

## 2026-07-29 — ambient events and canonical documentation

- Added catalog-owned collapsed/strained/healthy thresholds. Server Health band crossings now append immutable band events; entering/leaving collapse adds cascade/recovery events.
- Added the idempotent mid-T3 recruitment operator. It verifies the authoritative company stream and refuses Founders who have ever signed or already received the offer; Postgres uniqueness proves once-per-career delivery.
- Expanded the closed Go/DB event registry and strict payload tests for the entire Commons family.
- Published `docs/commons.md`, updated production/save canonical docs for the six-intent surface and save v6, and updated repository/onboarding status. Formula generation publishes an explicit empty Phase-0 source-weight array rather than JSON null.
- Focused Go/TypeScript/schema tests and the full Postgres integration target pass after the dispatch migration.

## 2026-07-29 — complete-diff review correction

- Found a real ordering defect before archive: PostgreSQL `now()` is transaction-stable, so the final `compact_sampled` and `compact_left` events from one leave share a timestamp. UUID ordering could project leave first and reject the legitimate final sample.
- Bound Commons projection order for equal timestamps as sign → sample → leave, then raw event ID. Added an end-to-end sign → sample → leave integration regression and verified the projection ends non-member while retaining the final sample/Health update.

## 2026-07-29 — Capacity accumulation correction

- The archive review found that `compact_sampled.capacity` is the tithe from one accrual receipt, but the projection was replacing the prior value. That violated D2's absolute-sum Capacity rule even though the latest-state Health inputs were correct.
- Capacity now adds once per idempotently projected sample under a company-scoped advisory transaction lock; Health/Compliance remain latest-state inputs. Replaying either sample adds nothing. A real-Postgres regression proves sequential `1e0` and `2e0` tithes expose `3e0` at cohort and server scope.
- Replaced the Commons boundary's source-text scan with an executable dependency-graph check, so test-only imports do not create false failures while a real production dependency still fails closed.
- The full existing balance report was regenerated out of tree. Pacing, milestones, outcomes, balances, and invariants are unchanged; only the expected catalog/constants and save-state hashes differ.

## 2026-07-29 — dependency-bound surface split

- Final acceptance review found two design obligations that cannot be truthfully implemented in this server foundation: the every-run incorporation/Open Source front door, and the cohort-panel/guild/monthly-vote surface. The repository has no implemented incorporation, faction, guild, transport, or client shell contract, and inventing those models here would violate RFC-0000.
- Split those obligations into draft `commons-onboarding-and-governance.md`, linked in both directions. Its unresolved ballot and UI schemas are explicit `DESIGN-GAP`s with named dependencies. The current RFC now states the implemented server boundary exactly; no shipped behavior was weakened or silently deferred.
- The same claim audit corrected two stale normative sentences: Phase 0 computes cohort/server Health with the guildless 80/20 fallback (not a nonexistent guild scope), and a maximally loyal collapsed member retains 40% of the *bonus* / ×3 of the maximum ×6, not the draft's mathematically impossible “≤47%” claim. The accepted immediate leave/re-sign behavior is now also explicit as a deviation from the research cooldown proposal.
- The formula artifact previously published the formula shapes and source table but omitted NPC, tithe, smoothing, cohort, and population controls needed to render them honestly. Artifact schema v4 now publishes every Commons catalog control, with a structural regression test guarding the complete field set.
- The final data-boundary audit found the 1.5 collective exponent hard-coded in both runtimes despite C1 assigning it to the Commons catalog. Added `collective_exponent_ppm=1500000`, strict 1.0–10.0 validation, runtime use in Go/TypeScript, generated-artifact publication, and a quadratic mutation test proving the formula actually reads the catalog.
