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

The [2026-11-05 real SQL/HTTP supplement](#2026-11-05--actual-garden-producer-to-verified-public-archive)
advances the original controlled-reader finding below: actual verification, compaction and
board authorization now execute. The deployed/default-release boundary remains unproved.

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

## 2026-11-05 — Actual Garden producer to verified public archive

Source `ce932586` (runtime source `c5130acf`, kernel0.3.226), migration86.
Question/method/limits were declared in the
[payload plan](data-rights-payload-core-plan.md#2026-11-05--rp-606-actual-producer--verified-board-archive-extension).
This closes the older synthetic-reader/compaction evidence gap, not RP-606's
privacy conflict. PASS means the observation completed; **the initialized arm
fails the adopted hidden-salt boundary**.

Two fresh disposable compositions use the unchanged complete `grown` Garden
corpus catalog: actual Company18/Founder25. Initial Company tier/lifetime value
and Fiscal credit are explicit controlled genesis, not earned progression.
Real Account HTTP supplies credentials; the progressed fixture Founder is linked
before those credentials are issued. In the initialized arm, real authenticated
Fiscal-unlock and Garden-plant requests generate the salt; it is never seeded by
the observer. The other arm omits those actions and remains null-salted. This is
a same-format control, not a single-field causal comparison.

Both execute actual Wind Down, the normal verification worker and SQL compactor.
Each queue reaches verified, archive terminal sequence1 and zero unreferenced live
rows; each gets four real `verified_runs` rows. No board row is inserted by the
observer, no archive is constructed and no reader is replaced. Before completion,
the same public archive path returns404. Afterward unauthenticated HTTP returns200,
exact stored gzip bytes, matching digest/ETag and
`public,max-age=31536000,immutable`. The salt is still the current Founder salt
following Exit into the next Company.

| Actual source / reader | Dormant arm | Server-initialized arm |
|---|---|---|
| Private Founder head | No salt | Same generated salt retained |
| Private Founder replay | No matching value | Sequence2 contains the generated salt |
| FK-retained Company run-log row1 | No matching value | Same salt in replay inputs |
| Verified gzip archive → public HTTP | Null field, no salt value | Same salt in original bytes |
| Ordinary Garden view/command receipts | Salt-free | Salt-free |
| Public verdict / genesis | Salt-free | Salt-free |

The exact exposed JSON location is
`entries[0].replay_inputs.resolved.founder_carry.founder_extensions.server_garden.salt_hex`.
Both raw-byte search and a separate decoded-string/path walk agree. Access,
refresh and recovery credentials are absent from the downloaded evidence. Actual
Founder histories verify. The existing public evidence consumer also downloads
all three resources, checks storage/hash/cache bindings and Go-reverifies the
Company archive; corrupting its first receipt refuses. That consumer's ordinary
credential checks can pass while the separately measured Garden salt leaks.
The optional TypeScript public seam is explicitly not requested and gets no credit;
DB-pinned catalogs are supplied, so this is not the complete public catalog loop.

Initial direct chain68860 passes1.51s. Added lineage/consumer checks30181 pass1.67s
(two subcases, no skips); no product/assertion/guard/corpus change. In that second
execution, initialized salt commitment is
`sha256:eafa6dded2d85fb1170f1a1b0d3835cbd7b93d2c5fe5eebbf8b95024c1fbd200`;
public archive digest is
`sha256:1902bceb5a5f027887243542262d124d4f0e5b941e4cdf28e9b4990d6d26aa99`.
Those bind this observation, not future random salts/IDs/times or a golden file.
Raw local logs/results are `.cache/garden-public-archive-observation/`;
`evidence.json` SHA256
`b660bb9ba7a213c578dd4c8573e41520e7bda834fe2760598a18a90cfe1b9b53`.

The [retained manual observer](probes/garden-verified-public-archive.go.txt),
SHA256 `399b00ab25800961ae39095d309b36ce000918ca9679a7d0a5f6fe13fccbd49b`,
is byte-identical to the executed final temporary test. To reproduce at compatible
source, copy it to the previously absent
`server/gameserver/rights_garden_archive_observation_test.go`, then run from root:

```sh
make test-save-integration SAVE_TEST_PACKAGES=./gameserver SAVE_TEST_FLAGS='-run TestRightsGardenVerifiedArchiveObservationIntegration -v' SAVE_TEST_COUNT=1
```

Remove only that temporary copy after all handles terminate. It is deliberately
outside ordinary CI: a future privacy fix may yield exposed=false rather than
being forced to preserve the leak. Root removed its copy after execution and only
the newly owned tmpfs PG `fea6446f...`; shared network/caches remain. Both actual
Go/PG images are Linux-amd64 under ARM64 Desktop emulation. No supported native
host, Caddy/release-binary, default epoch8, natural pacing, deployed-player,
delete/restore or complete rights proof is claimed.

Next: specify and prove a public Company-verification representation that excludes
protected persistent state while retaining private original history and honest
verification/source binding. Existing sampled null admission does not authorize
redaction. Authored SG8/AC10/OD-12 and API C14 need an accepted contract before
publication; no salt rotation, stored-byte rewrite, weaker proof or blanket
ranking suppression follows. Research author/first filter: Codex root.

**Independent evidence review:** APPROVED complete eight-path `ce932586..e153eaf6`.
Review by: fresh independent Codex `/root/garden_public_archive_evidence_review`.
Recorded by: Codex root. Its exact retained observer passes cold1.47s, both arms/no
skips, with a freshly generated salt, all actual private/public source bindings,
real board authorization and Go replay/corrupt-receipt refusal. Affected vet passes.
No TypeScript credit. It independently confirms the privacy FAILURE, not repair,
default/deployed exposure or feature/release approval. All handles terminal,
temporary copy/owned tmpfs PG removed, shared resources preserved, SQL released.
Ignored `.cache/garden-public-archive-review/review-results.json`, digest checked
by root: `afeabf2c50b4582a4ec0330b6e945d0624f18301a3c39e2d3cc7a9f242265ba3`.

## 2026-11-05 — Public Company replay projection feasibility

Predeclared in the [payload plan](data-rights-payload-core-plan.md#2026-11-05--rp-606-public-company-replay-feasibility),
source `3a1b8737`, unchanged runtime/kernel0.3.226/migration86. This is an **unapplied
in-memory experiment**, not a privacy fix or adopted public wire format.

A fresh actual HTTP unlock/plant → ordinary manual command → Wind Down run passes
normal verification/SQL compaction (terminal sequence2), with four actual board
rows and original-byte unauthenticated download. Both archived Company inputs carry
the same server-generated salt. Private Founder head/history retain it; the live
Company source row is correctly pruned in this run. Progressed genesis and initial
Fiscal budget use the prior fixture-only grown catalog, not natural progression
or default epoch8. No board insertion, synthetic archive or changed worker guard.
The exported original genesis/archive/pinned catalog bytes feed both actual replay
libraries offline. The optional TypeScript HTTP/catalog-download seam is not run.

| Arm | Go / TypeScript verdict | Company result |
|---|---|---|
| Original | verified / verified | Baseline |
| Replace only salt value with null | verified / verified | Identical complete state, receipt and ordered events at both entries, including final and next Company |
| Remove salt member | state_divergence / state_divergence | Existing private codec rejects it |
| Remove Garden extension | state_divergence / state_divergence | Existing private codec rejects it |

The null arm changes two carry inputs, at
`entries[0..1].replay_inputs.resolved.founder_carry.founder_extensions.server_garden.salt_hex`.
Its full Founder output changes in each runtime: **it is not a valid private-state
substitution**. Founder digests establish within-runtime changes only; differing
Go/TS Founder encodings are not claimed byte-equal. All complete public view objects
are independently compared across runtimes and between original/null arms.
Receipt corruption, ordered-event corruption, sequence removal and wrong constants
identity still refuse in both verifying arms and both runtimes. These are real
consumer refusals, not merely changed metadata assertions.

The four existing Company harvest cases match complete committed state/receipt/
event goldens in both runtimes. They already omit salt: zero changed inputs, not
four additional elision successes. Each deliberately nonterminal run returns
`log_gap`; independently wrong payload and resolved harvest hashes each return
`state_divergence`. The prior private-Founder growth-dependency study remains the
private-verification evidence; it was not repeated.

Original gzip SHA256 is
`f9e4cafe345d725e3cfda49e2a845c8c7b6c1d52d09f50dad94c812722959973`.
Read-only SQL before/after and a final direct query match it; source-file bytes
also remain unchanged. Under the same canonical JSON serialization the original
object hashes `82ae720c9c40ef8bf08b013810d3ebd16c8b3adb7281dddd522501cfc3d10ce6`,
and the derived object `ffc89e832d437fb0fe1158578ec7a33c88706f4c97f8cb431c195d8f1b3f3636`.
Thus the derived representation has a different identity; it cannot silently claim
to satisfy C14's current original-byte promise. Raw exported fixture evidence stays
ignored in `.cache/garden-public-projection-feasibility/`; no raw salt is committed.

Executed producer53839 PASS1.10s/no skips; final Go70709 PASS0.04s (package0.343s).
Final Node exits0, and the cross-runtime comparison exits0 after explicitly
requiring nonempty event lists at both entries. Affected `make vet GO_PACKAGES=./production`
exits0. The initial TypeScript instrument omitted Exit events, so its first output
is excluded as complete equality evidence. The first extraction correction also
changed the harvest loop incorrectly and failed with ReferenceError; retained as
invalid, then corrected narrowly. Final results include both Exit and harvest
events. The earlier vet handle's lost exit status is not credited; the directly
observed repeat passed. No product/library/assertion/corpus change was made.

Reproduction uses small retained manual observers, outside ordinary CI. Start with
no other SQL owner and an absent `.cache/garden-public-projection-feasibility/source.json`
(the exporter refuses replacement), then from repository root:

```sh
mkdir -p .cache/garden-public-projection-feasibility
cp planning/platform-alignment/probes/garden-verified-public-archive.go.txt server/gameserver/rights_garden_archive_observation_test.go
git apply --unidiff-zero planning/platform-alignment/probes/garden-public-projection-source.patch
make test-save-integration SAVE_TEST_PACKAGES=./gameserver SAVE_TEST_FLAGS='-run TestGardenProjectionSourceObservationIntegration -v' SAVE_TEST_COUNT=1
cp planning/platform-alignment/probes/garden-public-projection.go.txt server/production/rights_garden_projection_observation_test.go
make test-go GO_PACKAGES=./production GO_TEST_FLAGS='-count=1 -run TestGardenProjectionFeasibilityObservation -v'
cp planning/platform-alignment/probes/garden-public-projection.mjs.txt .cache/garden-public-projection-feasibility/observe.mjs
node .cache/garden-public-projection-feasibility/observe.mjs
```

Bind the fresh source rather than reuse this run's random IDs/times/salt/hash:

```sh
python3 - <<'PY'
import base64, hashlib, json
from pathlib import Path
p = Path('.cache/garden-public-projection-feasibility')
b = (p / 'source.json').read_bytes()
s = json.loads(b)
(p / 'source-binding.json').write_text(json.dumps({
    'source_file_sha256': hashlib.sha256(b).hexdigest(),
    'original_archive_sha256': 'sha256:' + hashlib.sha256(base64.b64decode(s['archive'])).hexdigest()
}) + '\n')
PY
cp planning/platform-alignment/probes/garden-public-projection-compare.py.txt .cache/garden-public-projection-feasibility/compare.py
python3 .cache/garden-public-projection-feasibility/compare.py
make vet GO_PACKAGES=./production
```

Remove only the two temporary test copies after all handles terminate. Root retained
the executed source and verified the zero-context producer patch reconstructs its
exact bytes, removed its temporary tests/owned tmpfs PG and preserved shared
network/caches. The SQL image is Linux-amd64 under ARM64 Desktop emulation.
No all-historical-version/all-command, mature Garden-production, natural pacing,
release-binary, native supported-host, R-006, deletion/restore or full rights proof.

The result supports specifying a separately versioned **Company-only** public
representation with its own identity and explicit secret exclusion. A future
public codec may omit the secret rather than fabricate dormant Garden state; the
null surrogate here is only a feasibility instrument. Private original history,
source binding and refusal strength remain required. SG8/AC10/OD-12 and API C14
still require an accepted reconciliation before implementation/publication.
RP-606 stays OPEN. Ignored evidence manifest `evidence.json` SHA256
`04fd43829cbe80d0736f49ebdc1f13f3d92ab15b51711b917e1526fe3be3345b`.
Author/first filter: Codex root.

**Independent technical evidence review:** APPROVED complete nine-path
`3a1b8737..ddda9f4c`. Review by: fresh independent Codex
`/root/garden_public_projection_review`. Recorded by: Codex root. Fresh actual
producer PASS1.14s/no skips, Go PASS0.03s/package0.339s, Node/comparison/vet exit0.
Independent deletion of the null-arm TypeScript Exit events makes the comparator
exit1; honest results are restored. All original/source hashes unchanged. It
confirms feasibility only, not privacy repair/protocol adoption/full acceptance.
All handles terminal; temporary tests/owned PG removed, shared resources preserved,
root evidence restored, SQL released. Ignored review manifest
`.cache/garden-public-projection-review/review-results.json`, root-checked SHA256
`945b29fd2b588a578f3f32ca19c399a006c4dabcbe5a0713f88242a3a69b9b15`.
