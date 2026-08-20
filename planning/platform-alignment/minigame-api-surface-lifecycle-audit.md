# Minigame API & Surface lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `e3ba58b`; 2026-08-20.

This pass re-derived the active RFC from its complete specification, fifteen owner rulings,
plan/log, generated API authority, composed account/gameserver handlers, Minigame/Pitch and Soul
Recovery producers, client tree, canonical docs, implementation history, and cross-party verdicts.
Temporary discrimination mutations were restored. No product code, owner-authored RFC/design text,
canonical product documentation, player copy, or implementation checkbox was persistently changed.

## Bottom line

The authenticated backend playability seam is real. A cold real-Postgres/HTTP composition test
creates and reconnects to a Pitch session, drives every tenant transition, auto-resolves, retries
durable bytes, proves terminal `current=none`, then starts/reconnects/progresses/resolves Soul
Recovery and exercises its watchdog terminal path. Severing the production tenant-content resolver
makes that test fail at create with the exact `500 internal_invariant/minigame_api` seam.

The active RFC is nevertheless not closeout-ready:

- no `minigame_session` or `soul_recovery` component exists;
- no Pitch table component mounts through a generated tenant-surface registry;
- the client has generated DTOs/operation metadata but no generated/thin HTTP client;
- MA-C9 still does not enumerate exact props, callbacks, state transitions, error/reconnect behavior,
  keyboard behavior, or copy keys that an implementer can build without inventing UX;
- AC2's “byte-match” test accepts extra response bytes and the generated `APIError` permits a
  category/detail cross-product instead of literal ruled pairs;
- AC3 proves command-flood nonmutation, but recovery flooding is only a 429 against a stateless
  stub and never compares authoritative progress before/after rejection; and
- AC4's named enumeration covers four minigame operations only. Adding a client `founder_id` to
  the recovery finish schema leaves the complete Account unit population green.

The canonical backend docs are unusually honest: they describe the composed API and explicitly
leave toys/disclosure rendering to the UI successor. The main lifecycle drift is in the active RFC
body, status/dependency header, plan wording, and unassembled final review union.

## Current cold evidence

All valid commands ran from repository root and reached their declared populations:

- `make test-go GO_PACKAGES='./account ./minigame ./minigameapi ./gameserver'
  GO_TEST_FLAGS='-count=1'` — all four packages green;
- `make test-save-integration SAVE_TEST_PACKAGES=./account` — cold real-Postgres Account population
  green;
- `make test-save-integration SAVE_TEST_PACKAGES=./gameserver` — cold real-Postgres composed
  gameserver population green;
- `make api-check` — generated OpenAPI/TypeScript/compatibility outputs unchanged; and
- `make typecheck` — TypeScript/Svelte zero diagnostics.

One attempted final batch launched the Account and gameserver Compose suites concurrently. They
share the declared Postgres service and collided on truncation/catalog seeding, producing deadlocks
and duplicate keys. That attempt did not measure either target and is excluded. The immediate
sequential reruns above both passed; the invalid parallel run is retained here rather than hidden.

## Restored discrimination probes

1. Replaced `runtimeCatalogs.ResolveTenantContent` with an unconditional miss. The cold gameserver
   integration population failed `TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration`
   at Pitch create with `500 internal_invariant/minigame_api`. Restoring the resolver returned the
   full package to green. AC1's producer/composition seam discriminates.
2. Bypassed the per-session Soul Recovery progress limiter. The Account integration population
   failed because the seventh heartbeat returned 200 instead of 429. Restoring the limiter returned
   it to green. This proves rejection, but the fixture's recovery adapter has no call/progress
   counter and therefore cannot prove the rejected beat left state unchanged.
3. Added optional `founder_id` authority to `SoulRecoveryFinishRequest`. All cold Account unit tests
   stayed green. The named privacy enumeration checks only the four minigame operations; the
   recovery wire test rejects Founder input only on `start`, not progress/cancel/resolve.
4. Appended an extra byte to the `minigame_revision` deterministic rejection. All cold Account unit
   tests stayed green because the claimed byte-match table uses substring containment. The runtime
   pair is mechanically correct at HEAD, but the literal byte oracle does not protect it.

Every mutation was removed. Restored product paths have no diff.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Proven composed backend integration** | One real-Postgres/HTTP client uses mounted authenticated operations for complete Pitch and Recovery lifecycles, reconnect/token rotation, durable retries, terminal current state, and watchdog behavior. The production resolver-sever mutation fails exactly. | Preserve the composed mutation and include later shared API/content changes in the final current-history union. This does not prove a human UI. |
| AC2 | **Partial** | All eight RFC operations are registry-mounted and generated; drift/pins and many exact wire negatives are real. The shared error schema admits invalid category/detail pairings, and a rejection with extra bytes passes the alleged byte-match test. | Generate or validate literal operation/status error pairs and compare exact bytes for every deterministic mapping; retain additive drift negatives. |
| AC3 | **Partial** | Authenticated command flooding returns 429 before execution and byte-compares the same current session after refill. Recovery's seventh heartbeat returns 429 and the limiter bypass fails the test, but the adapter is stateless and no production state/progress snapshot is compared. | Drive heartbeat flooding through the composed coordinator or a stateful adapter and prove session/token/progress bytes unchanged before ordinary progress resumes. |
| AC4 | **Failed enumeration oracle** | Foreign/missing minigame sessions are byte-indistinguishable and minigame request identity fields are rejected. The enumeration contains four minigame operations, while this RFC registers four Recovery operations too; a recovery finish identity-field mutation survives all Account unit tests. | Enumerate all eight operations and every server-owned identity/clock field; include response hidden-info checks and one mutation per request family. |
| AC5 | **Unmet and specification-blocked** | No surface components, tenant-surface registry, Pitch table mount, recovery scheduler wiring, browser/a11y workflow, API client, or governed surface copy exists. MA-C9 names desired shapes but never enumerates the exact executable component contract. | Ruling author reconciles and completes MA3/C9; API supplies the ruled generated client; then implement both surfaces, tenant mount, lifecycle/error/reconnect/keyboard behavior, and discriminating browser proof. |

## Producer-to-consumer trace

### Pitch

1. The minted Pitch tenant, frozen policy/content, session repository, command coordinator, atomic
   resolution, and current-session query exist.
2. Eight relevant authenticated operations mount through the private registry; four are minigame
   operations. Generated TypeScript describes DTOs and operation metadata.
3. No production client imports those minigame DTOs or calls those paths. Repository-wide client
   search finds engine/catalog/replay utilities only, not an interactive table or host surface.
4. The required `engine_ref,engine_version -> child` registry does not exist. The only surface
   registry is UI Foundation's generic mount catalog and contains the five Game UI screens.
5. The composed HTTP lifecycle proves the backend API, not a player workflow.

### Soul Recovery

1. Three minted activities and the authenticated coordinator exist. The framework-neutral
   `RecoveryScheduler` exists and correctly pauses/reconnects without replaying queued beats.
2. Nothing constructs that scheduler from a live activity's ruled ceiling/3 cadence. Canonical
   `docs/soul-recovery.md` states this wiring is owed by the UI.
3. No recovery surface renders disclosure, progress, cancel, reconnect-required, or terminal
   return behavior, and no toy receives the non-authoritative local seed.
4. The composed backend session proof therefore stops before the named human capability.

## Normative, plan, and review drift

1. The Summary still says “no new persistence” although C3/C5 ruled and shipped Founder-v21
   sequence state plus receipt rows. Appended “amended” text did not reconcile the normative body.
2. MA2 still says it adds the already-archived Recovery routes and includes `session_expired`, which
   C1 explicitly removed. MA1 still mandates a catalog per-session budget, which C8 struck.
3. The dependency header still calls UI Foundation accepted and Game UI draft. UI Foundation is
   archived and Game UI is accepted/implementing. The status mentions only the old composition
   verdict although AC1–AC4 later received backend verdicts.
4. MA3/C9 remains directional after its own finding demanded exact props/events/states, tenant
   registry, error/reconnect rendering, keyboard/cancel behavior, and copy keys. It also assumes a
   generated client that API Foundation has not implemented.
5. The plan calls four authenticated minigame endpoints “public,” omits the Recovery enumeration
   and AC2/AC3 literal proof debt, and has no body-reconciliation step.
6. Designated verdicts validly cover the original backend slices through `dba3bf7`; later shared
   API/Game-UI/content edits have successor reviews in other logs. The active MA log does not
   assemble those ranges into one current criterion union, and the future surface range is absent
   by definition. The final review checkbox correctly remains open.

## Smallest honest closeout order

1. Ruling author reconciles Summary, MA1, MA2, MA3, dependency/status header, and AC2–AC5 body with
   C1/C3/C5/C8/C9 and current repository truth; enumerate the buildable component contract.
2. Under API Foundation's accepted C9 rule, provide the generated/thin authenticated HTTP client
   and an all-source boundary negative before a surface invents raw fetch calls.
3. Repair AC2's literal error-pair/byte oracles, AC3's recovery nonmutation proof, and AC4's full
   eight-operation enumeration. These are backend closeout witnesses, not new mechanics.
4. Implement the generated tenant-surface registry and Pitch child, then the Recovery surface and
   scheduler wiring, using governed copy and fail-closed loading/error/reconnect states.
5. Prove keyboard/axe behavior and full API→component→tenant/toy→terminal-return workflows with
   mutations that sever dispatch, child registration, cadence, visibility pause, cancel, and
   reconnect.
6. Reconcile plan/log/docs and assemble the original plus successor plus surface ranges into one
   exact current-history union; obtain the designated cross-party verdict before archival.
