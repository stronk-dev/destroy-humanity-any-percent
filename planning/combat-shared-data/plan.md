# Combat Shared Data & Arithmetic — implementation plan

- **Assignee:** Codex
- **RFC:** `rfc/combat-data-model.md`
- **Started:** 2026-07-29

2026-10-14 C3 correction (RP-604): TypeScript integer stores reject non-BigInt inputs before
coercion; damage rejects non-boolean crit flags before the zero-power return. All 31 old-source
refusals fail in Node and Chromium/WebKit; final 315 Node/630 native checks, cold Go/vet,
both division gates and typecheck pass. No valid arithmetic, data, kernel activation or battle
consumer changes. Fresh independent corrective review APPROVES the complete
six-path d1fb8d57..efeae834 range; the owning log records its executed evidence
and limits. Adjacent construction and full parent acceptance remain separately open.

1. [x] Extract one shared SplitMix64/rejection-sampling authority and add labeled substreams.
   - C4/AC3's sequence supplement now pins 72 draws across three fixture labels and
     three seeds, retains an independent closed-form fixture generator and exercises
     overlapping stream lifetimes with interleaved/reversed/chunked/new-consumer draws.
     A shared/resettable factory passes the old first-draw pattern but fails the new
     sequence comparison. This is test evidence, not new RNG behavior or battle logs;
     fresh independent corrective review APPROVES complete18fd6c06..be66902b, test-only.
     The owning log records independent fixture regeneration and Go/TS execution; this
     does not approve adjacent Combat construction or the full parent feature.
2. [x] Implement exact int32/int64 combat arithmetic and Temperament chart in Go and TypeScript.
   - C4's lane ordering projection now compares `(tick, player_slot, placement_seq)`
     identically in Go/TS without signed-int32 subtraction overflow. Shared vectors,
     independent 324-pair ranks and actual reversed-rotation sorting pass. This does not
     choose placement legality, duplicate policy or construct a battle log/engine;
     full C4/parent acceptance remains open. Owning log records the range after `efeae834`.
   - C1's six runtime values and36 temperament/runtime identities now have pure Go/TS
     constructors and revalidated chart lookup. All1,296 directed pairs match the existing
     literal temperament chart; twelve shared invalid pairs refuse in both slots.
     This does not populate tags/species/sprites, supply a battle consumer or close C2/C5.
3. [ ] Add the strict combat catalog/schema and the complete Phase-0 fixture.
   - C2's species row is now structurally decoded in Go/TS: exact root/stat keys,
     signed-int32 fields, existing temperament/runtime axes, four ordered declared
     moves and ordered tags. `TestSpeciesRecordVectors` and
     `client/test/combat-species.test.ts` cover the shared accepted/refusal population,
     all36 identity pairs and independent storage. This does not choose stat minima,
     effects, tags or balance, construct the blocked full catalog, or close item3/AC1.
   - C2's independent team/deck reference envelopes are now implemented in Go/TS:
     exact three/eight slots, ordered declared references, strict wire refusal and
     isolated output storage. `TestLoadoutReferenceVectors` and
     `client/test/combat-references.test.ts` pass, including current native
     Chromium/WebKit. This does not provide the blocked catalog/effect schemas,
     engine legality or a live player consumer; full item 3 remains open.
   - C2's common lane unit/building/spell row now decodes in Go/TS: exact required
     fields, signed-int32 numeric representation, existing temperament axis, three
     target values, ordered tags and optional bypass marker. `TestLaneRecordVectors`
     and `client/test/combat-lane-record.test.ts` consume100 shared accepted/refusal
     cases; all54 axis/marker combinations and storage controls pass. This does not
     supply effects, battle-legal minima, cheap-cycle/bypass coverage, full catalog
     or a live engine. Item3/AC1 remain open; the new range needs designated review.
4. [ ] Add fixture-only Trust/Soul tables and cross-runtime golden vectors.
5. [ ] Add source lint and full parent acceptance/property coverage. Both Go and TypeScript
   division boundaries now execute locally; `TestRepositoryCombatDivisionBoundary` and its
   AST/recursive negative controls cover Go, and `make verify-combat-boundary` covers TypeScript.
   Full parent acceptance and designated review remain open; this is not a catalog/table closeout.
6. [ ] Publish canonical docs and record independent review before archival.

2026-10-08 C1/C3 / AC2/AC4 supplement: the existing shared vector suite now covers exact
cycle edges, HP/stamina clamp domains, signed store boundaries and late-only saturation.
Both runtimes, affected Go/vet, source-division gates and type checks pass locally; the
owning log records the exact review range. No production arithmetic changes or parent
acceptance claim; catalog/effect and Trust/Soul specification gaps remain unchanged.
