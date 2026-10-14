# D-008/D-009/D-015 — newer saved content and receipt fields

Source checkpoint: 2026-10-14 at `be66902b`. Method:
[focused payload extension](data-rights-payload-core-plan.md#2026-10-14-focused-extension-at-be66902b).
This is a bounded source/codec/transport map, not a policy, legal classification,
complete historical export schema or proof of populated account deletion.

## Version and persistence boundary

`save/state.go` supports scope-specific maxima **Company 19 / Founder 25**;
`save/runlog.go` currently writes Company replay-inputs 12. These are supported
ceilings, not evidence that every live revision has those versions. `WireVersion`,
the revision's SQL version and pinned activation rules choose the actual encoding;
zero runtime WireVersion defaults to legacy CurrentVersion 14, not automatically 25.
The September core map's Company 18/Founder 21/replay-inputs 8 are dated facts.

No table or column was added by migrations 00075–00084, as the preceding event
refresh established. New nested fields nevertheless appear in versioned revisions,
genesis, command/replay logs, receipts and outbox copies. All these stores must be
considered separately from a current rendered snapshot.

## Selected saved-state fields

| Family | Exact selected persisted fields / producer → consumer | Linkage and coverage limit |
|---|---|---|
| Company19 attainment | `achievements_attained_run[]`, `attainment_score_run`; `save.companyStateV19`, Production achievement/Clout transitions → restore/replay and authoritative projection. | Achievement IDs are catalog references attached to the Company/run, not a new account ID. `axis_stack` is economy configuration, not an additional save field here. Other Company fields are outside this tranche. |
| Founder22 Reputation | New `reputation_spent`, `reputation_nodes_owned[]`; existing level/unlock fields also change through `production/reputation_intent.go`. `save.stateV22` → restore, next-run contribution/starter handling and replay. | Per-Founder purchase history and node choices; the owned ID is shared content, while enclosing ownership identifies the player's choice. Company replay-inputs9 adds spent/unlock/owned to `founder_carry.founder_extensions`. |
| Founder23 pet identity | `pet_identities[pet UUIDv7]`: `species_id`, `temperament`, `palette_id`, `name_key`, `adopted_at_ms`, `adopted_at_attended_ms`; `pet_adoption_intent.go` → `save.wirePetIdentity`, restore and replay. | Pet UUID joins adoption, care and equipment; wall/attended adoption times remain. `name_key` must resolve in the pinned species catalog, not arbitrary free text. Replay-inputs10 carries this identity map. |
| Existing pet care alongside adoption | `pets[pet ID]`: `stats_ppm`, `stat_decay_remainders_ppm`, `cooldown_until_attended_ms`, `trust_ppm`, `trust_decay_remainder_ppm`, `evaluated_through_attended_ms`, `behavior_state`, `behavior_entered_at_attended_ms`, `behavior_queue[{behavior_id,due_attended_ms}]`, `behavior_prng_cursor`; `pet.CareState` → save/replay/projection. | Care and identity have matching key sets in adopted state. This is mutable per-pet state, not erased by a new identity record or receipt acknowledgment. It predates Founder23; adoption links it to the new immutable identity. |
| Founder24 cosmetics | `cosmetics: {owned[], equipped:{pet ID:cosmetic ID}}`; `production/cosmetic_intent.go` → `save.wireCosmetics` and `cosmetic.State`, restore/replay/overlay. | Owned/equipped choices join to the same pet IDs. Empty collections are persisted, not null defaults. Replay-inputs11 carries the full cosmetic state. |
| Founder25 Garden | `server_garden: {salt_hex,tick_anchor_wall_ms,tick_seq,substrate_id,substrate_set_wall_ms,plots[],seed_collection[]}`; plots have `row,col,species_id,age_ticks,matured_effect_ppm`. `garden.State` / `save.stateV25` → restore, Garden advance/harvest, contribution and replay. | Salt is nullable or a server-hidden16-hex value; clocks, collection and plots are per-Founder state. Ordinary Exit preserves Garden, including its salt. Replay-inputs12 carries the complete encoded Garden, not merely visible plots. This is the protected-field issue below. |

Source ownership: `server/save/state.go`, `server/pet/identity.go`,
`server/pet/state.go`, `server/cosmetic/state.go`, `server/garden/state.go`,
`server/production/replay.go`. Encoded data remains scoped by its enclosing
Founder/stream/run even when a nested object has no explicit account UUID.

## Selected applied receipt and private resolved-input variants

| Producer | Base transition receipt fields / related retained inputs |
|---|---|
| Direct Reputation purchase | `intent_id,outcome,founder_revision,fiscal_sweep,node_id,resolved_cost,reputation_available_before,reputation_available_after,reputation_unlock_ppm_after,effective_from`. Exit-plan purchases separately retain `{node_id,resolved_cost}` in Founder `exit.v2.reputation_purchases`; run_started v2 carries the starter/bonus summary described in the event dossier. |
| Pet adoption | `intent_id,outcome,founder_revision,pet_id,species_id,temperament,palette_id,name_key,adopted_at_attended_ms,status_band,eligible_action_ids[]`. Private Founder resolved inputs retain `kind,attendance,adoption_nonce`; the nonce reproduces draws, not a session credential. This does not classify every attendance field or rejected receipt. |
| Cosmetics | `intent_id,outcome,founder_revision,kind,cosmetics,event:{kind,payload}`. The receipt contains the full owned/equipped state, not only the changed item. Acquire's private resolved inputs also retain active Company stream/revision/run/tier coordinates; equip/unequip omit that context. |
| Garden commands | `intent_id,outcome,founder_revision,kind,event:{kind,payload}`. Private `garden_command.v1` resolved inputs retain `kind,server_ms,garden_salt_hex?,advance`; `advance` contains tick/catchup facts, matured plot references and spawned species/recipe references. The private salt field is not an ordinary event/receipt field. |
| Garden harvest | Distinct base Founder `{intent_id,outcome,founder_revision,kind,harvest}`, Company-log `{intent_id,outcome,harvest_hash,credited_resource_id,credited,forfeited_units,cap_reason_key,faucet_applied,company_revision,founder_revision}` and API receipts. The API has `intent_id,outcome,kind,founder_revision,company_revision,harvest,credited_resource_id,credited,forfeited_units,cap_reason_key,faucet_applied,garden_advance`. Harvest includes plot coordinates/species/units, total units, discovered seeds and hash. Company resolved inputs additionally retain policy/faucet evidence and a Founder-log coordinate; the Founder arm retains attendance/server time and optional initialization salt. |

The base producer objects are not always the final stored receipt. The shared
`ApplyFounderLogged` transition in `production/founder_replay.go` decorates an
applied result with `fiscal_sweep` when Fiscal is pinned; Garden triggers also
receive `garden_advance` through `decorateGardenAdvance`. Their existing nested
sweep/advance results must travel with that final receipt. This tranche does not
flatten every historic Fiscal sweep variant.

These are current source maps, not executed equality observations for every receipt.
Save's ordinary applied/rejected paths copy normalized decision receipts into
`intent_records`, the corresponding run/Founder log and `transport_player_outbox`.
Multi-stream Garden harvest has deliberately different API, Founder and Company
receipts, so they must not be presumed byte-equal. Verified run compaction copies
Company replay inputs and matching event bodies into immutable gzip history.
Account deletion archives Founder-owned streams; it does not itself rewrite these
feature payloads. Fresh joined feature-data deletion/restore remains unproved.

## RP-606 — protected Garden salt meets original-byte public verification

Garden OD-12(b), adopted with the recommended defaults, keeps its per-Founder
salt from clients/exports; SG9/AC10 repeat the hidden-information requirement.
The actual `founderCarry` / `buildReplayInputs` producer nevertheless includes
the full Garden under `resolved.founder_carry.founder_extensions.server_garden`.
That is appropriate **private** replay material, not inherently safe public data.

Source chain: `production/intents.go` uses this carry for logged Company inputs;
`replayverify/archiveVerifiedRun` copies `run_log.replay_inputs::text` into gzip
archive entries without redaction. `leaderboard/PublicRunEvidence` authorizes
by an actual `verified_runs` board row and checks pin/genesis/archive metadata
and hash; it does not inspect nested protected fields. The public C14 handler
serves those exact bytes under their digest, with a one-year immutable cache.

Executed **controlled** observer, not SQL:

- Existing complete Garden fixture → valid Founder 25 save encoder and actual
  Founder/Company replay-input builders.
- Constructed archive wrapper and controlled evidence reader → real registry-
  mounted `get_public_run_replay_log`, request-ID runtime and gzip response.
- Salted case: codec contains the known fixture salt, replay inputs contain it,
  HTTP 200, exact compressed-byte equality, decoded response contains it,
  `public,max-age=31536000,immutable`.
- Dormant/unsalted format control: same version/HTTP/cache/byte preservation,
  but none of those three payloads contains the salt. Garden is reset to its
  activation state, not changed only in the salt field.

The archive wrapper is **not** produced by SQL compaction and the reader bypasses
database authorization. No natural verified Garden run, deployed disclosure,
public catalog loop or complete re-verification was demonstrated. `PASS` means
the observation ran, not that privacy passed. This confirms a source/transport
contract conflict that must be resolved before Garden publication; it does not
prove that a live player is currently exposed.

The [retained probe](probes/garden-public-replay.go.txt) is the exact Go observer.
To repeat at compatible source, temporarily install it as
`server/production/rights_payload_observation_test.go` with normal patch tools,
run `make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestRightsGardenPayloadObservation -v'`,
then remove that temporary test. It uses only committed synthetic fixtures.
The executed temporary file was removed; no product or existing test assertion was changed.

Do not strip or replace stored fields, rotate the carried salt, weaken public
verification or silently suppress a required feature. Garden OD-12/SG8/AC10 and
API C14 need an authored, reviewed public-evidence contract that satisfies both
privacy and independent verification. D-008's broader export choice cannot by
itself repair an unauthenticated raw download or reopen the existing exclusion.

## Other checks and limits

### RP-606 replay-dependency follow-up at d6ea5884

The [predeclared comparison](data-rights-payload-core-plan.md#2026-10-14-rp-606-replay-dependency-follow-up)
now runs the actual Go and TypeScript transitions, rather than treating an
HTTP byte-preservation test as proof of replay necessity. Three populations,
six arms each, execute in each runtime. This is a selected pure-replay population,
not natural play, SQL, public authorization/compaction or the full public verifier.

| Population | Original control | Three other valid salts | Missing member / null |
|---|---|---|---|
| Populated retained-Garden Company Exit | Applies; normal unmodified activation separately matches committed receipt/new-Company state. Retained arm reuses the existing literal permanent-Garden witness. | All apply with identical final/new Company state, receipt and all three event sets. Full Founder carry changes and preserves each replacement salt. | Missing `salt_hex` refuses at carry decoding. Explicit null **applies**, with identical Company outputs and a changed, null-salted Founder carry. This is codec admission, not lawful redaction/carry behavior. |
| Founder initialization (`plants-and-anchors`) | Applies and matches committed receipt, events and full post-state. | All apply with identical receipts/events, but persist their respective different salts. No growth tick has yet drawn. | Missing resolved `garden_salt_hex` and explicit null both refuse. TS rejects explicit null at its string decoder; Go rejects absent initialization material during advance. No error-text parity claimed. |
| Founder growth (`switches-substrate-after-growth`) | Applies and matches committed receipt, events and full post-state. | All three refuse because the recomputed advance differs from the recorded advance. | Missing saved salt member refuses restore; explicit null restores but refuses advance because initialization material was not supplied. |

The negative controls establish **real Garden recomputation dependence**, not
merely that a parser requires a key. Conversely, a strict-key refusal in Company
carry is not evidence that Company arithmetic consumed the salt. The sampled Exit
does not exercise every ordinary Company command, contribution or replay verdict.
Its surviving null/substituted carry must not be called a valid public format:
it changes Founder evidence, violates original-byte identity and does not satisfy
SG2's unchanged carry or the later private Garden computation.

Engineering consequence: the next contract work must distinguish public Company
verification inputs from protected Founder/Garden replay, with explicit proof
scope, schema/version/hash binding and privacy negatives. This is a direction for
author reconciliation, **not adoption of a particular split or permission to weaken
independent verification**. The evidence rules out a blanket "remove the seed and
re-run existing replay" remedy; it does not establish a complete alternative.
The existing OD-12 exclusion remains settled, and RP-606 remains OPEN.

Exact manual observers are retained as
[Go](probes/garden-salt-dependency.go.txt) and
[TypeScript](probes/garden-salt-dependency.ts.txt). Temporarily install them with
normal patch tools at `server/production/garden_salt_dependency_observation_test.go`
and `client/test/garden-salt-dependency-observation.test.ts`. Commands:

```sh
make test-go GO_PACKAGES='./production' GO_TEST_FLAGS='-count=1 -run TestGardenSaltDependencyObservation -v'
client/node_modules/.bin/vitest run --root client test/garden-salt-dependency-observation.test.ts --disableConsoleIntercept
```

Final Go observer passes cold (0.672s); Node runs three declarations and prints all
18 comparisons (1.19s total). PASS means the controls and observations completed,
not that the unsafe public disclosure was repaired. The native probe's first run
failed an unchanged control because it supplied an indented fixture command;
normal canonical-command encoding fixed the instrument. Its second run stopped
on the observer's dereference of an admitted null salt, not a product panic; the
null-safe observer records that admission explicitly. Both failed observations
are excluded as complete populations. No expectation/corpus was regenerated.
The temporary test-tree files were removed; these research snippets are outside
ordinary CI and do not institutionalize unsafe disclosure as a passing gate.

Cold existing save tests pass: `TestFounderV2` executes six version 20–25 codec
cases; `TestFounderReputationMigrationCorpus` and `TestCompanyAxisMigrationCorpus`
execute their populated codec/refusal corpora. These are not SQL migrations,
all historical variants or deletion/export witnesses. The declared Docker socket
again returns exit 28/HTTP 000/zero bytes after 5.001443s; it is unresponsive, not
proved stopped. No restart, cleanup or database substitute was used.

Remaining: fuller historical receipt/resolved-input maps, nested shared-party
data, actual authorization→compaction→public reader and populated deletion/
restore proofs, explicit rights policy and reviewed implementation. No release,
feature approval or legal conclusion follows from this tranche.
