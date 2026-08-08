# RFC: Permits & the T3→T4 Gate (pre-mint)

- **Status:** candidate implemented — byte-exact fixtures assembled and gated; awaiting owner
  ratification of the manifest hashes and designated implementation review before acceptance.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/02 §1` (Permits = a Company-run constraint resource alongside
  Energy/Water), `design/02 §role-law` (capacity role: "Energy/Water/Permits caps — the Kittens
  model"), `design/01 §T4` (permits as data-center-era constraint), `design/02 §Externality`
  (constraint-dodging — "unpermitted turbines" — feeds Externality; "Legal% = every permit first").
- **Depends on:** Economy Kernel + Production Engine (archived — this is two rows in the existing
  grammar), Route Registry (archived — one gate row), First Content Epoch (draft — consumes these
  bytes; FCE-C1).
- **Planning:** `planning/permits-and-t3-gate/` (once implementing)

## Summary

The epoch-6 bundle cannot load without `gate.t3_to_t4` (FCE-B1), and the owner ruled the gate is
permits-backed, not cash-only. This RFC introduces the game's SECOND economy resource —
`company.permits` — using ONLY existing grammar shapes (one resource row, one generator class, one
gate row), and defines the T3→T4 gate over it. No new engine mechanics, no new schema fields, no
save-schema bump expected.

## Specification

All amounts are PROVISIONAL BYTES (constants are config; the composed harness and epoch-7 retunes
own tuning). Shapes are exact.

### P1 — The resource row (economy artifact, additive)

```json
{"id": "company.permits", "scope": "company", "numeric_kind": "decimal",
 "initial": "0", "minimum": "0",
 "hardcap": {"amount": "24", "reason_key": "resource.company_permits.cap.phase0"}}
```

A small, VISIBLE hardcap (the hardcaps-never-softcaps law; the Kittens constraint model). Raising
the cap via capacity-role generators is the design's stated destination but is NOT in this
contract — the capacity role is not yet loader-implemented; a later epoch owns it.

### P2 — The faucet: one T3-era generator class (economy artifact, additive)

```json
{"id": "generator.legal_dept",
 "price": {"resource_id": "company.cash", "base": "1e8",
           "curve": {"kind": "geometric", "ratio": "1.15e0"}},
 "production": {"resource_id": "company.permits", "base_rate": "1e-3"}}
```

The idle-honest faucet: permits accrue lazily from a purchased legal department (closed-form
compatible; offline-progress law applies unchanged). `base_rate 1e-3`/s ≈ 1 permit per ~17 min per
department, saturating at the cap. **DESIGN-GAP (flagged, owner-confirmed direction):** design
names permits as a constraint resource and the Kittens cap model but never names the faucet
mechanism; the generator faucet is the chosen minimal shape (over a manual filing action, which
would violate the idle-default law as sole faucet). Confirmed by the FCE-C1 owner ruling round.

### P3 — The gate row (routes artifact, additive — closes FCE-B1)

```json
{"gate_id": "gate.t3_to_t4",
 "requirement": [{"resource_id": "company.cash",    "amount": "1e12"},
                 {"resource_id": "company.permits", "amount": "12"}],
 "routes": []}
```

Cash keeps the shipped ladder (1e9 → **1e12** → 1e15, and the Tier-3 1e12 pacing coordinate);
permits are the constraint the era is ABOUT. Both debit at crossing (the established gate
semantics). `routes: []` in v1 — permit-dodging skips (the Externality-priced
`Non-Road Engine Clip` class) are later route content on this gate, exactly as design/02
§Externality frames them.

### P4 — Copy & integration

- Copy rows for the cap reason key and the generator's player-facing family, through the copy
  pipeline in the same change (names per the existing resource/generator copy conventions).
- The doctrine fixture's `gate.t3_to_t4` reference becomes loadable UNCHANGED (FCE-B1's goal).
- No save-schema bump expected: resource balances are id-keyed. If any code path hardcodes
  `company.cash` as "the" resource, that is an implementation finding to surface, not a contract
  change here.

## Acceptance criteria

1. Both loaders (Go + TS) accept the extended economy and routes artifacts; doctrine
   `ValidateRoutes` passes against the extended routes bytes; composed bundle parity fixture green.
2. Production accrual: a company with N legal departments accrues permits at the closed-form lazy
   rate, saturating at the visible cap with the reason key; offline accrual honors the standard
   90%/24h policy.
3. Gate crossing debits BOTH requirements exactly; insufficient permits rejects with the standard
   typed rejection; the crossing replays byte-identically.
4. Chronology/depletion and route-registry gates green; the change lands as `BALANCE-CHANGE:`
   with its own designated review BEFORE the First Content Epoch consumes the bytes (FCE-C1).

## Changelog

- 2026-08-07: created (draft) — commissioned by the FCE-C1 owner ruling (permits now); minimal
  shapes over existing grammar; faucet DESIGN-GAP flagged and direction owner-confirmed.

## Codex acceptance-review blockers (2026-08-07 — PT-C1–PT-C6)

The three literal rows fit the schema-3 economy/routes grammars, and the existing production,
ledger, gate-debit, replay, and new-run initialization paths are catalog-driven. Implementation is
still blocked on the following contracts. They are content/governance decisions; code must not
silently choose them.

### PT-C1 — The required pre-mint `BALANCE-CHANGE:` is impossible under the epoch guard

P4/AC4 require this change to land as `BALANCE-CHANGE:` **before** the First Content Epoch consumes
it. The live artifact authority points `economy` and `routes` at the two `phase0.json` files. The
fail-closed epoch guard permits changing either active file only when the same commit changes the
seed; a `BALANCE-CHANGE:` seed change must append exactly one epoch. Therefore a separate Permits
balance commit would itself mint epoch 6, making the First Content Epoch epoch 7. A hotfix would
prematurely extend epoch 5 and cannot honestly carry the mandated `BALANCE-CHANGE:` subject. The
artifact authority also forbids replacing an existing artifact name/path, so staging under a new
production path does not solve the active-byte transition.

**Proposed contract:** implement and review Permits against byte-exact candidate fixtures first,
without touching the active epoch artifacts. The single owner-gated First Content Epoch mint then
copies those reviewed rows into the active `economy` and `routes` documents and appends epoch 6 in
that same `BALANCE-CHANGE:` commit. Its changelog consumes the Permits verdict and its tests prove
the production bytes equal the reviewed candidates. Reconcile AC4 from “before the First Content
Epoch” to “reviewed before, activated atomically by the First Content Epoch.” Alternative: mint a
dedicated Permits epoch 6 and renumber First Content to epoch 7; do not let implementation choose.

### PT-C2 — “T3-era generator” is not enforceable by the existing generator grammar

`generator.legal_dept` is a schema-3 generator row. `buy_generator` checks identity, count,
affordability, and exact arithmetic; it does not inspect Company tier. Even schema-4's `tier` field
classifies relevance groups and does not gate purchase eligibility. As written, the Legal
Department is purchasable in any tier whenever the Company has `1e8` Cash, contradicting the
normative “T3-era” description.

**Proposed contract:** preserve the no-new-mechanics scope and rule the generator globally
purchasable, with its `1e8` price as pacing rather than an authorization boundary; rewrite
“T3-era” as “priced for the approach to T3.” If mechanical T3 availability is required, this RFC
must own a generator-availability grammar/runtime/replay extension (or reuse a fully specified
existing window mechanism); that is no longer a three-row content RFC.

### PT-C3 — The faucet's exact multiplier semantics are unstated

The existing production path applies every eligible `target:"all"` contribution to every
Company generator. Consequently Legal Department output is not always `N × 1e-3` permits/second:
Commons, faction/guild, event, prestige, and future all-target factors can scale it, and offline
efficiency applies through the standard accrual path. Both Go and TypeScript will implement that
automatically, but AC2 currently reads like an unmultiplied fixed faucet.

**Proposed contract:** adopt the existing production law explicitly. Online permit rate is
`N × 1e-3 × contributionFactor(generator.legal_dept)` and offline accrual applies the pinned
offline policy, with one ledger quantization and accrual-only saturation at 24. The shared parity
fixture covers neutral, non-neutral all-target, 24-hour offline, and near-cap saturation cases. If
Permits must ignore production multipliers, that requires a new production kind and contradicts
the existing-grammar ruling.

### PT-C4 — The promised generator copy has no binding surface

The resource hardcap reason is a real registered copy-bearing field, so
`resource.company_permits.cap.phase0` can be resolved and completeness-gated. Economy generator
rows have no `copy_key`/`name_key`, and the Copy Pipeline deliberately forbids deriving display
text from mechanical IDs. Adding an orphan `generator.legal_dept.*` copy family would not make the
generator player-facing or prove any binding; there is also no existing generator-copy convention
for P4 to invoke.

**Proposed contract:** narrow this RFC's copy work to the bound cap-reason row, with exact owner
text supplied in the ruling, and carry Legal Department title/description to the Game UI content
surface that owns generator presentation. Alternative: extend the economy generator grammar with
an explicit copy field, register it in `copy/references.v1.json`, and bump/port the catalog schema;
that is a real cross-runtime mechanic and must be specified here rather than implied. In either
case, delete “names per the existing ... conventions,” because no such convention exists.

### PT-C5 — Activation and save-shape behavior must be new-run-bound

Adding a resource and generator changes the exact key sets of Company balances and generator maps.
Pinned in-flight runs correctly retain the old economy bytes; new state creation is catalog-driven
and can initialize both new keys. The RFC currently says only “no save-schema bump expected,”
without binding which side of the epoch boundary owns the new maps. A hotfix-style interpretation
would risk validating old state against new exact key sets.

**Proposed contract:** activation is new-run-bound with the epoch-6 economy bytes. An epoch-5 run
finishes and replays under its pinned one-resource/one-generator catalog; Exit into epoch 6 creates
the next run with `company.permits:"0"` and `generator.legal_dept:0`; a fresh epoch-6 founder gets
the same keys at genesis. No migration mutates an in-flight run and no save version changes. Add
the cross-epoch Exit, fresh-genesis, and old-run replay fixtures to AC1/AC3.

### PT-C6 — The candidate-byte manifest and copy literal are incomplete

The mechanical JSON snippets are exact, but the review cannot bind final artifact bytes from
additive snippets alone: insertion order, unchanged surrounding bytes, candidate paths/hashes,
and the cap-reason English row are unspecified. FCE-C6 requires a literal promotion manifest, so
leaving these choices to implementation would recreate the same gap one layer earlier.

**Proposed contract:** before status becomes accepted, append a literal Permits manifest naming
the candidate economy path, routes path, copy source path, SHA-256 of each complete document,
schema versions, exact copy row, and commands/content gates. State that the resource is inserted
after `company.cash`, the generator after `generator.beige_tower`, and `gate.t3_to_t4` between the
T2→T3 and T4→T5 rows; gate requirements remain byte-ordered `company.cash`, then
`company.permits`. FCE's promotion table consumes these hashes rather than reconstructing rows.

## Owner rulings on PT-C1–PT-C6 (2026-08-07)

- **PT-C1 — accepted as proposed (primary arm).** Permits is implemented and reviewed against
  BYTE-EXACT CANDIDATE FIXTURES, never touching the active epoch artifacts; the single owner-gated
  First Content Epoch mint copies the reviewed rows into the active `economy`/`routes` documents
  and appends epoch 6 in the SAME `BALANCE-CHANGE:` commit, its changelog consuming the Permits
  verdict and its tests proving production bytes = reviewed candidates. AC4 is RECONCILED by this
  ruling: "reviewed before, activated atomically by the First Content Epoch." The dedicated
  Permits-epoch alternative is REJECTED (one owner-gated mint, not two).
- **PT-C2 — accepted, recommended arm.** `generator.legal_dept` is globally purchasable; the
  `1e8` price is PACING, not an authorization boundary. P2's "T3-era" wording is reconciled to
  "priced for the approach to T3" by this ruling. A generator-availability grammar is NOT
  introduced here.
- **PT-C3 — accepted as proposed.** The existing production law applies EXPLICITLY: online permit
  rate = `N × 1e-3 × contributionFactor(generator.legal_dept)` (all-target multipliers included),
  offline via the pinned policy, one ledger quantization, accrual-only saturation at 24. AC2 is
  reconciled to this reading. The shared parity fixture covers neutral, non-neutral all-target,
  24-hour offline, and near-cap saturation. A multiplier-immune production kind is REJECTED.
- **PT-C4 — accepted, narrow arm.** This RFC's copy work is exactly ONE bound row — the cap
  reason — with the owner text supplied here verbatim:
  **`resource.company_permits.cap.phase0` = "The county issues at most 24 concurrent permits. The
  county does not care about your roadmap."**
  Legal Department title/description is CARRIED to the Game-UI content surface that owns generator
  presentation (recorded debt there, not here). P4's "names per the existing … conventions" phrase
  is STRUCK by this ruling (no such convention exists). The generator-copy grammar extension is
  REJECTED for this RFC.
- **PT-C5 — accepted as proposed.** Activation is NEW-RUN-BOUND with the epoch-6 economy bytes:
  an epoch-5 run finishes and replays under its pinned one-resource/one-generator catalog; Exit
  into epoch 6 creates the next run with `company.permits: "0"` and `generator.legal_dept: 0`; a
  fresh epoch-6 founder gets the same keys at genesis; no migration mutates an in-flight run; no
  save version changes. The cross-epoch Exit, fresh-genesis, and old-run replay fixtures are added
  to AC1/AC3 by this ruling.
- **PT-C6 — accepted as proposed, with the assembly split.** Insertion order is RULED now:
  resource after `company.cash`; generator after `generator.beige_tower`; `gate.t3_to_t4` between
  the T2→T3 and T4→T5 rows; gate requirements byte-ordered `company.cash` then `company.permits`;
  the copy row is the PT-C4 literal. Codex assembles the two complete candidate documents to these
  rulings and files their SHA-256 + paths + schema versions + gate commands as the Permits
  manifest; owner ratification of those hashes (the FCE-C3 pattern) completes acceptance. FCE's
  promotion table consumes the ratified hashes.

## Changelog (rulings round)

- 2026-08-07: PT-C1–C6 ALL RULED — candidate-fixture landing with atomic mint activation; global
  purchasability (price = pacing); explicit multiplier semantics; single bound copy row with owner
  text; new-run-bound activation fixtures; insertion order ruled, manifest hashes owed by
  assembly + ratification.

## Candidate manifest — **RATIFIED by Marco, 2026-08-08**

**Owner ratification (post-designated-APPROVE of {7d9cb37, 90633a6, d30ab9e}):** the three
candidate documents are RATIFIED as byte-pins — economy `31af760c…990`, routes `6c7c4350…df2`
(at its post-F1 path `balance/testdata/permits-t3-gate-candidate-v1.json`), copy source
`e87b0224…512`. The generated-copy hash `8462d6d5…8a2` is recorded as the point-in-time
composition attestation at 7d9cb37 per the re-review's N1 (the composed catalog is a shared
living artifact; the cell is refreshed at mint time). The epoch-6 mint MUST copy the ratified
documents byte-identically (PT-C1: activated atomically by the First Content Epoch). Any edit to
a ratified document records a replacement hash here.

### Original manifest (2026-08-07 — as filed for ratification)

The RFC's human-form `24` and `12` amounts are encoded below as RFC-0001 canonical Decimals
`2.4e1` and `1.2e1`. This is a representation normalization only; the ruled values are unchanged.

| Role | Path | Schema | SHA-256 |
|---|---|---:|---|
| Economy candidate | `balance/testdata/valid/permits-economy-candidate-v1.json` | economy 3 | `31af760c45394401af5847b0e1be5790281d88555b9380a73bd5ca87a0105990` |
| Routes candidate | `balance/testdata/permits-t3-gate-candidate-v1.json` | routes 1 | `6c7c4350bcd43840a141fb5c0525d9779f11ed0ed836a8783f21f22f6c880df2` |
| Bound copy source | `copy/catalog/permits-candidate.json` | copy 1 | `e87b0224f6d17a25ffb018ab09cf0898e720e8c783bee470ae1ac7bf29e1a512` |
| Composed generated copy | `client/src/copy/generated/catalog.json` | copy 1 | `8462d6d5a8ccde29508d51e46c256eeac52a569988f8561e8bed87fcdfcb08a2` |

Gates: `make copy-check`, `make test`, and the final repository `make verify`. The production mint
must copy the candidate economy/routes documents byte-for-byte apart from their already-active
surrounding path and must consume the designated verdict for this candidate range.

## Changelog (candidate assembly)

- 2026-08-07: byte-exact economy/routes/copy candidates assembled; shared Go/TypeScript loader,
  doctrine-route, multiplier, offline, and hardcap proofs added; manifest hashes filed for owner
  ratification. No active balance artifact or epoch seed changed.
