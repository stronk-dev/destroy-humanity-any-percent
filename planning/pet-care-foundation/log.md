# Pet Care Foundation implementation log

## 2026-08-05 — owner rulings and implementation start

- Owner accepted C1-C8. The central deliverable is the reusable Founder-only transition and replay
  boundary; care must never mutate through Company `ApplyLogged` or spend Company state.
- Reconciled the normative body at the decision sites: fixed mechanical stat IDs, care through
  `ApplyFounderLogged`, and no Company resource/time/token cost.
- Implementation starts with persistence and replay plumbing. Production pet content and numeric
  policy rows remain owner/harness data and will not be invented.

## 2026-08-05 — Founder persistence/replay envelope

- Added `save.Store.ApplyFounderLogged`, a Founder-scope-only mutation path with a typed
  `FounderReplayCommand`. The Store assigns sequence and database server time under the locked
  stream; replay inputs must echo both exactly, so feature code cannot invent attended time or
  import live Company context.
- Appended migration 00054 for immutable `founder_log` rows. Applied and rejected commands are
  both recorded, only applied rows name a save revision, every row pins constants identity, and a
  database trigger rejects any non-Founder stream even if application validation is bypassed.
- Real-Postgres coverage proves applied state/revision, rejected no-op, byte-hash idempotent replay,
  Company-scope refusal before callback, exact row ordering, null-vs-applied revision semantics,
  and SQL immutability. Envelope tests reject mismatched server stamps/commands, unknown keys, and
  non-object resolved inputs.
- `make test-go GO_PACKAGES='./save'`, `make vet GO_PACKAGES='./save'`, and root
  `make test-save-integration` pass. Pet's pure care union and Founder replay verifier remain the
  next independently reviewable slice.

## 2026-08-05 — independent rejection of Founder boundary (`7620311^..7620311`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** not approved; one HIGH and two MEDIUM findings require remediation.
- HIGH: exported legacy `Store.ApplyIntent` still accepted Founder streams, allowing an unlogged
  Founder mutation that bypassed `ApplyFounderLogged` and `founder_log` entirely.
- MEDIUM: receipt enqueueing hardcoded `scope=company`, so Founder receipts entered durable
  transport with false scope. MEDIUM: migration 00054's direct-insert archive check did not lock
  the parent stream and could race archival.
- The authoritative database timestamp, envelope closure, intended-path serialization,
  applied/rejected log semantics, transaction composition, immutability, retained archive history,
  migration direction, and kernel bump held. Coverage debt was named for concurrency, rollback,
  Founder event/outbox, and archive-race proofs.

## 2026-08-05 — Founder boundary remediation

- Founder scope now rejects the legacy unlogged mutation callback before invocation. Receipt
  enqueueing accepts and validates the actual stream scope, and transport documentation names
  Founder receipts.
- Forward-only migration 00055 replaces the already-applied 00054 trigger with a parent-row
  locking recheck, so archive and direct log insertion have a serial order.
- The real-Postgres fixture now covers the legacy bypass, Founder event and applied/rejected
  receipt scope/revisions, concurrent same-revision commands, a failure after the log insert that
  rolls back every write, and an archive-first/direct-insert race. Kernel semantics advance to
  0.3.26. Independent remediation review remains required.

## 2026-08-05 — remaining Pet implementation contracts

- Applying the boundary to a real pet transition exposed four gaps not resolved by C1-C8: replay
  needs segment genesis around externally-owned Exit mutations; wall-clock stamps cannot supply
  attended time; the Founder-only save/table/activation wire is unnamed; and the allegedly closed
  mood/FSM/status unions have no literal members.
- Filed C9-C12 with proposed executable contracts. No pet numeric balance row is requested; these
  are replay, persistence, time-authority, and wire-shape decisions that code must not invent.

## 2026-08-05 — independent remediation review (`cd60175^..cd60175`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** not approved; one new MEDIUM in the migrated production consumer.
- Every original HIGH/MEDIUM finding closed. The reviewer verified legacy-bypass refusal,
  Founder-scoped event/receipt rows, forward-only archive serialization, same-revision command
  serialization, post-log rollback, route-hint migration inputs, and kernel 0.3.26.
- New MEDIUM: `buy_route_hint` requests carrying unknown/invalid fields entered the new Founder
  handler but skipped Company `ApplyLogged`'s deterministic invalid arm; a known route could still
  be purchased. The fix must log a closed invalid resolved variant before projector/catalog reads.

## 2026-08-05 — invalid-Founder-command remediation

- The Founder route-hint handler now emits a deterministic rejected decision with resolved
  `{kind:"invalid",detail}` before any catalog or projection read. Its Postgres fixture proves an
  extra-field request remains at revision 1, unlocks nothing, records a rejected Founder-log row,
  then allows the canonical command to apply at the same expected revision.
- Kernel semantics advance to 0.3.27. Independent follow-up review remains required.

## 2026-08-05 — independent approval of invalid-command remediation (`17faed3^..17faed3`)

- **Review by:** independent Codex subagent `/root/l7b_independent_review`
- **Recorded by:** Codex
- **Decision:** approved; no findings.
- The reviewer verified the invalid arm executes before any catalog/projector access, persists the
  closed resolved variant, creates no revision/event or state mutation, preserves terminal
  idempotency, and leaves the same expected revision available to the canonical command.
- Independent production/save tests, vet, kernel-history verification, root Postgres integration,
  and exact-range diff check passed. Full root `make verify` also passes at kernel 0.3.27.

## 2026-08-05 — designated reviewer verdict: ApplyFounderLogged (7620311..1ba07fd) — APPROVE

Review by: the project's designated Claude reviewer. Recorded by: same.

**Approved.** The Founder mutation boundary is a faithful mirror of Company `ApplyLogged`, verified
at source: shared `applyIntent` core so it inherits FOR-UPDATE serialization (double-apply proven
impossible by the real-Postgres two-goroutine test), same-transaction idempotency with
byte-identical recorded receipts, applied AND rejected immutable history (reject_immutable_change
trigger), revision CAS, archive/genesis guards, and — critically — it commits the Founder stream
ONLY and rejects ambient company state in its closed replay envelope (an `ambient_company_state`
key is refused). All four prior findings fixed to the letter with regressions: the buy_route_hint
legacy bypass (the old RunLogSeq==0 unlogged founder path) closed at BOTH store guards and dispatch;
the founder-outbox scope corrected to `scope='founder'`; the archive race fixed forward-only with
FOR SHARE recheck; and malformed-command determinism fixed (invalid arm runs first — with the
honest nuance that for buy_route_hint no literal state-apply was actually reachable, so the real
defect was receipt-category/log-record determinism + defense-in-depth for future founder intents).

**Two residuals actioned/recorded:**
1. **KV-1 registry gap — FIXED directly by the reviewer:** `server/save/founderlog.go` (the exact
   Founder analog of the registered `runlog.go`, owning ValidateFounderReplayInputs / insertFounderLog
   / determinism logic) was NOT in `kernel/affecting-paths.json` — a future founder-ONLY edit would
   slip the kernel-drift gate (the span passed only because it co-touched registered intent.go).
   Added as a surgical append-only registry line (guard-integrity fix, no replay-byte change, no
   version bump — exactly the growth the guard's append-only rule accepts). Codex to confirm it
   lands green.
2. **Range-union gap — RECORDED, blocks any future archival of this span:** the cited Darwin ranges
   union to {7620311, cd60175, 17faed3}; **a2696b5 is UNREVIEWED** — and it is NOT docs-trivia, it
   added acceptance blockers C9–C12 to pet-care AND +63 normative lines to minigame-platform. Per
   the range-union rule, this span is NOT archival-eligible until a2696b5's RFC changes are
   reviewed. (Those changes are the blockers I ruled this turn — so they ARE now owner-reviewed via
   the rulings; recording that the review-of-record for a2696b5 is this turn's ruling pass.) Also
   noted: no single verdict approves the cumulative HEAD; the boundary rests on
   rejection-then-remediation per-commit, acceptable for a live slice but the archival gate must
   cite the remediated end-state.

Dormant, tied to future work: `scopeMatchesChannel` needs a `founder` arm when founder snapshots
land (Pet C7/C8); the 7620311 transient bisect-break is intermediate-only.

**Verdict: the boundary is sound and the buy_route_hint migration is clean. Founder-replay
verification is correctly deferred (Pet C9). Proceed.**

## 2026-08-05 — owner ruling: APPROVE rewrite of 80456c1 (my error)

Approved: rewrite unpushed `80456c1` to place `server/save/founderlog.go` in strict lexical
position (before `genesis.go`), replay `4005f00` on top. Preconditions verified in-repo:
unpushed (no remote contains it), referenced by NO verdict hash (my prior verdict referenced the
CHANGE — "add founderlog.go" — never the commit `80456c1`, which Codex created after), and
protocol-violating with NO forward remedy — `verify-kernel-version.mjs`'s `validateGuard` enforces
strict ascending sort (`paths[i-1] >= entry → throw`) at EVERY historical commit, and the
`history-corrections.json` mechanism forgives only missing-version-bumps (`assertBump`), never a
malformed/unsorted registry. So a forward correction cannot help; the rewrite is the only fix.

**This is my mistake.** Last turn I fixed a real KV-1 coverage gap (founderlog.go absent from the
registry) but did the insert POSITIONALLY (after runlog.go, next to its analog) instead of in the
registry's required strict lexical order (founderlog < genesis < runlog). I didn't check
`validateGuard`'s own sort invariant before editing its data. The working tree is already
corrected (founderlog.go now precedes genesis.go); the rebase carries that correct state into the
commit. **Lesson recorded: editing kernel-guarded config (`affecting-paths.json`) means respecting
the guard's structural invariants — strict lexical sort, append-only-per-commit — not just adding
the right path.** Direct edits to guard data get a `validateGuard`-style sanity check before commit.

## 2026-08-05 — designated reviewer verdict: Founder genesis batch (70d2d4a..7af623c) — APPROVE

Review by: the project's designated Claude reviewer. Recorded by: same. **This is the review of
record for the batch** — the prior ApplyFounderLogged approval covered 7620311..1ba07fd, and
1ba07fd is the PARENT of this batch's base; F2 (range-union) correctly flagged the batch was
otherwise unreviewed. Now covered.

**Approved — sound and complete for what it claims (the immutable-Founder-genesis HALF of A4).**
No correctness defects. Verified independently at source:
- **The poisoned-history repair (70d2d4a) is clean AND proven over the WHOLE rewritten span** —
  the reviewer RAN the guard (exit 0, non-shallow so the full guardIntroduction..HEAD walk),
  which means every commit in the rewrite has a sorted registry and a satisfied bump rule.
  founderlog.go is in strict lexical order at every descendant; the subsequent commits replayed
  intact. Repairing a fail-closed guard correctly (no NEW unsorted/unbumped commit anywhere in
  history) was the thing that most needed independent proof — confirmed.
- **Genesis mirrors run-genesis faithfully:** byte-identical pre-command Founder state via the raw
  jsonb bytes (the `state::text` equivalent, NOT a Go re-encode); genesis + first founder_log in
  one transaction; the deferrable AFTER-INSERT constraint binds seq=1 to the genesis revision so a
  founder command CANNOT run without genesis; ON CONFLICT + FOR SHARE recheck + the seq-1 PK force
  exactly one concurrent-first winner; immutable (UPDATE rejected, tested); fail-closed backfill
  RAISEs if a pruned revision made genesis unrecoverable (never fabricates history); Founder-only
  (no Company read). Kernel 0.3.28 justified, both new paths registered.
- **B1-B3 honesty confirmed:** grep proves Exit writes no founder_log/genesis — Exit-as-founder-log
  is honestly deferred behind the B1-B3 rulings, genesis claims only its half.

Findings (none blocking): **F1 LOW — no real-Postgres fault-injection test for the founder
both-or-none property** (run-genesis has runExitFault; applyIntent has no fault hook). The property
holds structurally (single-tx rollback + the genesis-less deferred-constraint rejection IS tested —
the stronger guarantee), but add the fault-injection parity test when the Exit-founder-log slice
lands (it'll need the hook anyway). F3 INFO — an available `(founder_stream_id, revision) →
save_revisions` FK is unused hardening (parity with run-genesis, not a regression); DELETE-
immutability untested though the same trigger covers it — add the one-line negative.

**Verdict: proceed with Exit-as-founder-log against the B1-B3 rulings; the genesis foundation is
solid.**

## 2026-08-05 — C12a cross-runtime Pet grammar

- Implemented the owner-ruled Phase-A stat IDs, status bands, moods, behavior states/events, care
  rejection details, queue hardcap, and PRNG label in isolated Go and TypeScript modules.
- One shared fixture asserts literal member order and queue boundaries in both suites. Go returns
  defensive vocabulary copies so a caller cannot mutate protocol authority at runtime.
- Added canonical docs scoped honestly to the grammar; no pet identity, state, action, decay,
  threshold, species, or temperament row is claimed or synthesized.
- `DESIGN-GAP:` C3/C11 still do not enumerate the exact pet catalog and Founder-state wire keys,
  including action/decay rows, remainder and cooldown maps, behavior queue entries, and the bond
  graph. Those mechanics remain blocked rather than inferred from prose.

## 2026-08-05 — remaining Pet wire bounce C13-C14

- Filed an exact fixture-catalog proposal for stat decay, actions, trust, mood thresholds, and
  behavior candidates; all numeric values remain balance data.
- Filed one replay-owned Founder pet-state proposal that resolves the mutable-table/snapshot and
  stored/derived-mood contradictions. Immutable identity stays relational; mutable mechanics stay
  inside the Founder transition boundary; writable bonds remain deferred.

## 2026-08-05 — protocol authority hardening

- Self-review confirmed Go vocabulary accessors return defensive copies, then found the
  TypeScript `as const` arrays were compile-time readonly but still mutable at runtime through a
  cast. Froze every exported array and added a mutation regression so protocol vocabulary cannot
  drift inside a running client.

## 2026-08-05 — designated reviewer verdict: Minigame+Pet foundation round (95a3e35..75b0a87) — APPROVE (code)

Review by: the project's designated Claude reviewer. Recorded by: same. **This IS the review of
record** — the review found (F1 HIGH governance) that NO prior verdict covered this 8-commit
round; the delegated approvals stopped at the Founder-genesis batch, entirely before 95a3e35.
That is the FIFTH consecutive batch where the designated pass is the actual coverage. Filing this
verdict clears the archival-eligibility block (F1) for the code that passes.

**Code APPROVED — no correctness defects, verified at source:**
- Scaling grammar (C28): exact-key row loader (DisallowUnknownFields + trailing-EOF + recursive
  uniqueJSONKeys), closed op union, duplicate-destination load error, correct negative floor
  (-5 floordiv 2 = -3 tested), big.Int math with MaxExactInteger bounds. **The Fairness Law is
  genuinely loader-PROVABLE and tested** (seeded ranked+power fixture fails) — caveat F2: `ranked`
  is a caller parameter until the registry composition binds it, honestly documented.
- Payout (C30): big.Int overflow-safe by construction (converted ≤ reduced ≤ score ≤ Max), proven
  at the boundary; credited resource must be a declared catalog ID; remainder carry partition-
  invariant; server-authoritative so no TS conversion needed.
- Fallback (C23/C27): three arms exact-key, bot_ref/npc_profile = mechanical id + semver,
  mixed/unknown/extra all rejected+tested.
- **Recursive duplicate-JSON-key rejection is real** (recurses objects AND arrays, nested-dup
  tested) and applied across all three loaders; pre-existing wire uses canonical-bytes equality
  which also catches nested dups — no loader accepts ambiguous keys.
- Pet vocabulary (C12/C12a): Go/TS enum members byte-identical against a shared fixture, queue
  hardcap 8, PRNG label pet.behavior.v1; 'runtime-frozen' = Object.freeze (TS) + unexported arrays
  with defensive-copy accessors (Go) guarding the shared protocol-authority slice.
- Kernel 0.3.32→0.3.38 one-bump-per-semantic-commit (F4: 0.3.38 is an over-bump for an
  Object.freeze no-wire change — harmless, forced by the watched path); KV-1 registry adds the new
  minigame/pet paths in strict lexical order; formula artifact schema 7→10 exports the grammar.
- Honesty confirmed: C31/C32/C33 + Pet C13/C14 are genuinely DESIGN-GAP-deferred, none half-claimed.

Non-blocking, queue with the composition: F2 (bind `ranked` to the registry when it composes),
F3 (sends_per_day/per_send_cap validated-but-unused + cap-forfeit reason key — lands with C32),
F4 (over-bump, informational).

**Verdict: the round is sound. C31-C33 + Pet C13-C14 are now ruled (this turn); the Fairness Law,
recursive strict decode, overflow-safe payout, and cross-runtime pet grammar all hold. Proceed.**

## PROCESS ESCALATION — the delegated-review coverage gap is now systemic (5 consecutive batches)

For five batches running the delegated (Darwin) approvals have stopped short of the substantive
commit, leaving the actual implementation unreviewed until the designated pass. This is no longer
an incident — it's a structural gap in how the delegated review scopes its range. **Recommend to
Marco:** either (a) the delegated review must cite a range whose union covers the FULL
implementation span it claims to approve (the existing range-union rule, actually enforced on the
delegated reviewer, not just the archival gate), or (b) the designated adversarial pass is made a
MANDATORY gate before any foundation batch is archival-eligible, not an after-the-fact catch. The
pattern is benign so far only because the designated pass has caught every gap — but it means the
delegated approvals are not currently load-bearing.

## 2026-08-05 — C13-C14 implementation recheck

- C14 correctly resolves mutable authority into the replay-owned Founder snapshot, but the current
  Exit validator still requires Founder and Company revision versions to match. The revision rows
  are already independently versioned; the coupling is an application guard, not a database
  limitation.
- C13 still omits literal keys for mood-threshold and behavior-candidate rows. Filed C15 with a
  complete exact-key/order proposal rather than choosing a fixture grammar in code.
- Filed C16 for the shared activation seam: Founder-only v17, pinned pet artifact identity, mixed
  Founder/Company Exit semantics, and Go/TypeScript replay parity. No pet content or deploy-current
  catalog fallback was introduced.

## 2026-08-05 — C14 pure replay-owned state boundary

- Implemented the exact C14 mutable-state object in isolated Go and TypeScript modules without
  activating a Founder save version. Both validators require complete four-stat value/remainder
  maps, declared action cooldowns, Trust/remainder bounds, the closed behavior-state union,
  declared bounded queue entries, and exact attended/PRNG cursors. Mood is rejected as stored data.
- One shared fixture fixes every JSON key and representative value across runtimes. Adversarial
  tests reject missing stats, unknown action/behavior IDs, oversized queues, invalid remainders,
  stored mood, duplicate declarations, unknown fields, and nested duplicate JSON keys on the Go
  raw-byte boundary.
- Kernel semantics advance to 0.3.41. C15 still owns the two missing fixture-catalog row shapes;
  C16 still owns pinned artifact identity and Founder-only save activation. No production pet,
  balance literal, or second mutable authority was introduced.

## 2026-08-05 — designated reviewer verdict: payout/faucet/pet-state round (75b0a87..58118e9) — APPROVE (code)

Review by: the designated Claude reviewer. Recorded by: same. Review of record — no prior verdict
covered this range (6th consecutive batch; see the standing process escalation).

**Code APPROVED — no correctness defects, verified at source:**
- **Faucet window (C33), the concurrency-critical one:** minigame_faucet_window carries BOTH
  quota_used AND conversion_remainder_ppm keyed (founder_id, minigame_id, attended_day); DOUBLE
  serialization (FOR UPDATE on account_founders THEN on the window row); both orders traced -
  quota 0->1->2 never double-spends, and carry is order-independent by the modular recurrence
  (rem = (s1+s2)*c mod 1e6 either way), matching one combined ConvertPayout. The 9eaf71e
  two-goroutine real-Postgres race proves it. Carry survives cross-session (on the window, not
  session state) and resets cross-day. HONEST deferral: claim-token exactly-once resolve is NOT
  yet proven (applyFaucetWindowTx alone isn't idempotent) - the token-guarded composer is [ ].
- **Payout facts (C32):** payout_score_fact_id + cap_reason_key validated fail-closed against
  declared sets; forfeit reason distinct from numeric saturation.
- **Pet state replay-ownership (C14), the architectural one:** grep confirms NO separate mutable
  pet_care_state table; state is Founder-save jsonb; MOOD IS DERIVED (a stored mood field is
  rejected both runtimes); complete stat/remainder maps, Trust bounds, queue hardcap 8, exact
  cursors, strict keys, nested-dup rejection, deep-clone decode (no leaked mutable authority).
- Kernel 0.3.38->0.3.41 one-bump-per-behavior-commit, test/docs commits correctly don't bump;
  KV-1 covers the new paths; formula authorities added.
- C15/C16/C34/C35 genuinely deferred (plan boxes [ ]).

Non-blocking: F1 (TS pet validator can't catch nested-dup/trailing-byte since it takes parsed
JSON - Go is the server-authoritative ingest gate, honestly scoped; add a TS raw-string entry or
a doc note if the client ever becomes a validation authority); F2 (payout_score_fact_id declared
set is test-injected until the tenant result_schema binding lands with the resolve composer);
F3 (over-quota churns remainder harmlessly - resets at day roll).

**Verdict: sound. C15/C16/C34/C35 ruled this turn (incl. the independent Founder-version-axis
reconciliation). Proceed.**

## 2026-08-05 — C16 independent save-version axes, infrastructure slice

- Replaced Exit's historical `Founder version == Company version` rule with four explicit floors:
  current Founder/Company and next Founder/Company, each derived from its pinned bundle. The
  terminal Company state cannot change version; resulting Founder and next Company versions must
  meet their own floors and cannot regress.
- Existing paired v16 Meters/Achievements activation supplies `16/16` floors as a special case.
  Unit coverage proves legacy migration, paired activation, and a mixed Founder-v16/Company-v14
  tuple; it also rejects a missing floor, decode-only v15, terminal changes, and either axis
  missing its declared next floor.
- `DESIGN-GAP:` C16 names a complete `pets` epoch artifact, but the implemented C15 artifact slice
  contains only mood/behavior grammar and C13 still supplies no complete production species,
  action, decay, or Trust catalog bytes. No artifact set, constants identity, or Founder version
  number is minted from partial fixture data. That owner-gated content/artifact step remains before
  the C14 pet map can enter the Founder save.

## 2026-08-05 — C17 ordering/artifact bounce and verification handoff

- Filed C17 after proving the scalar Founder version cannot encode two optional mechanics in
  arbitrary activation order. Proposed the existing queue order (`minigames=v17`, `pets=v18`) or
  an explicit feature-vector successor; no version or partial artifact was silently assigned.
- Normal root gates are green for commits `dc635f7..4c56865`: focused Go/TypeScript suites,
  typecheck, kernel full-history guard, formula drift, harness drift, client build/boundaries,
  schema/copy checks, browser suite, and the repository Postgres integration target. No push or
  balance mint occurred.

## 2026-08-05 — C15 exact pet catalog row families

- Implemented strict Go and TypeScript loaders for the two owner-ruled row families against one
  shared fixture. Mood thresholds contain every closed mood exactly once with strictly ascending
  ppm floors; behavior candidates use only the closed state/event unions, positive exact tick
  durations, and unique `(from_state,event,to_state)` tuples.
- Both loaders reject unknown keys, missing/duplicate moods, nonascending thresholds, invalid
  enum members, zero durations, and duplicate transition tuples. The Go raw-byte boundary also
  rejects nested duplicate JSON keys before typed decoding.
- This closes only C15's wire grammar. No species, temperament, action, decay, Trust, or
  production balance row was synthesized, and the C14 state remains outside the Founder save
  until C16's pinned-artifact/independent-axis transition lands.

## 2026-08-05 — C36/C17 ruled: Founder version stays a SCALAR chain (feature-vector deferred)

The scalar-vs-feature-vector fork (the one owner decision blocking the Founder save object) is
decided: SCALAR monotonic chain, fixed total order minigames=v17 then pets=v18; v18 requires v17's
artifact pinned. Same discipline as the Company axis (save at N carries the union of all fields <=N;
unpinned mechanics hold empty/default state; activation is always artifact-gated). Keeps Exit/replay
LINEAR instead of 2^K feature-subset combinatorics. Feature-vector envelope is a named-successor-RFC
escape hatch, reached only if an epoch ever needs a higher Founder mechanic without a lower one.

- v17 (minigames) Founder-save maps: minigame_ratings{elo,season_member,games_counted} +
  minigame_offline_quality{grade_ppm,last_founder_attended_ms,decay_remainder_ppm}; ratings live IN
  the save (not a side table — resolves C29's fork toward the save, replay-owned discipline);
  resolution facts stay in the minigame_resolution.v1 log arm; minigames artifact <=> floor>=17.
- v18 (pets): pet-state map added; complete `pets` artifact = C13 top-level union
  {schema_version,stat_policy,actions,trust_policy,mood_policy,behavior_policy}, with C15's mood +
  DETERMINISTIC behavior rows SUPERSEDING C13's mood/temperament sketch. pets artifact <=> floor>=18.
- Structure ruled, numbers deferred. Closes the "C15 fixture = partial catalog" hazard C17 named.

## Review governance: Codex independently recommends (c). Still MARCO'S call — not adopted.
Both agents now recommend (c) [delegated range-union coverage + retain designated archival gate].
Strong concurrence, but the rule change is the owner's; the designated pass continues meanwhile
(which is what (b) would formalize). Escalation remains open on Marco's desk.

## 2026-08-05 — designated reviewer verdict: version-floor/grammar round (dc635f7..0c3707f) — APPROVE (code)

Review by: the designated Claude reviewer. Recorded by: same. Review of record — no prior verdict
covered dc635f7..0c3707f (7th consecutive batch; standing governance escalation still open).

**Code APPROVED — no correctness defects. The critical item verified sound at source:**
- **Independent version floors (3a3cae5) — the one place a soundness hole could hide:** the old
  equality gate (founderVersion==companyVersion + atomic coupling) is REPLACED by per-axis
  lower-bound + no-regress checks in exit.go validateExitVersionTransition. It does NOT admit an
  invalid tuple: downgrades blocked by explicit regress checks; an artifact-not-pinned version is
  caught by the EXACT upper bound downstream in the SAME Exit transaction (validatedState ->
  ValidateFoundationState: v16 iff foundations pinned, >=15 unpinned rejected); floors are
  pinned-epoch-derived (exitVersionFloors reads each bundle's foundationsActive, server-computed in
  prestige.go, never client-supplied); zero/missing floors fail closed (writableStateVersion(0)
  => decode-only reject). Lower bound in the validator + exact upper bound in the pinning = sound.
- Offline-quality grammar (C34): greatest-met-threshold grade, neutral floor = lowest, ascending
  thresholds, watermark-only clock (no wall time); grammar/wire-parity only, decay compute not yet
  built (consistent stage). Pet mood/behavior catalog (C15): exact-key rows over closed enums,
  ascending full mood set, deterministic behavior candidates, queue hardcap 8, shared byte-identical
  fixture. Grade-boundary parity: shared fixture at/just-below/floor/max, Go+TS green.
- Kernel 0.3.41->0.3.45: 4 semantic bumps one-per-code-commit in lockstep; docs commit 0c3707f
  correctly no-bump; KV-1 registry covers all touched semantic paths (adds client/src/minigame/).
- Deferral honesty genuine: versionFloors() emits only 14/16 — no v17/v18 secretly shipped; C15
  fixture correctly NOT pinned as the complete pets artifact.

Non-blocking (2 LOW): F1 — the "mixed-axis Exit regression" test feeds a {Founder:16,Company:14}
floor tuple that production CANNOT currently emit (versionFloors returns founder==company today);
the validator is correctly built axis-independent AHEAD of the v17/v18 feature, so the test proves
the arithmetic, not a reachable mixed run. ACTION for the v17/v18 implementation (now unblocked by
the C36/C17 scalar ruling): add the END-TO-END reachable mixed-run test (minigames=v17 while
Company=14/16) so the "mixed-axis" box is backed by a reachable scenario, not just a unit-domain
input. F2 — TS parsers receive already-JSON.parse'd objects so cannot see physical duplicate keys;
raw-dup-key rejection is server-side only (matches every repo parser; server is the wire gatekeeper).

**Verdict: sound. Round approved. C36/C17 ruled this turn unblock the v17/v18 scalar schema.**

## 2026-08-05 — complete pets artifact and Founder v18 landing

Implemented by: Codex.

- Added the complete C17 pet artifact loader in Go/TypeScript: stat grid, actions, Trust, full mood
  thresholds, and deterministic `(from_state,event)` behavior transitions. The earlier C15 fixture
  cannot be mistaken for this complete artifact.
- Added replay-owned Founder v18 `pets`, with exact structural state validation plus catalog-owned
  action/behavior declarations. v18 requires the pinned v17 minigames artifact; Company remains
  v14/v16.
- Proved the reachable scalar chain v16 -> v17 -> v18 through the production Exit settlement path
  while every new Company run remains v16. Numeric/content rows remain unminted balance data.
- Kernel bumped 0.3.45 -> 0.3.46 with the shared save/replay change. Normal root gates are green;
  no content mint and no push occurred.

Self-review remediation `3085d4d`: pre-v17 codecs now reject rather than silently discard
nonempty minigame/pet maps. The regression and kernel 0.3.47 are in the independent-review range.

Cross-stream remediation `00420ae`: Founder Exit replay now activates empty v17/v18 schema maps
from the hash-matched next bundle before validating the result state. The Founder verifier resolves
that bundle without deploy-current fallback, and v18 replay is proven after a v17 head. Kernel
0.3.48; included in the independent-review range.
