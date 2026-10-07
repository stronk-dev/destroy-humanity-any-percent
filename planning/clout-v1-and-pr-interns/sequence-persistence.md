# Persisted opportunity / compute-burst sequences

Test-only scope predeclared `fe0e4196` at clean `f038a467`, with instrument
amendments `132d2925` and `6c3c4173` before their corrected experiments.
Authority: accepted Clout CV3/CV4 and AC8. This supplements ordinary persistence
evidence; it does not complete CV4, AC6, the natural player journey or Clout.

## Actual population [V]

`server/production/axis_sequence_integration_test.go` consumes the committed
`testdata/axis-stack/sequence-research-v1.json` as a command/mode/time plan,
not an expected-output substitute. It reconstructs raw requests using the
recorded intent IDs and each actual loaded head revision. Canonical replay
payloads deliberately omit intent IDs; the replay command owns those IDs.

All sixteen paths remain: four actual opportunity effects × two gaps
(3114/90000000ms) × online/offline/online or offline/online/offline. The actual
Service resolver, Store, frozen Founder contributions and Postgres execute them;
no test callback replaces the production transition or persistence logic.
The unchanged UNMINTED bundle and admitted synthetic inventory/history are
described in [the sequence observation](sequence-research.md). This is not
natural acquisition, minted content, all seeds or all combo-cap combinations.

Database clock minus26h keeps the long path's commands in the past. Diagnostic
setup pins real Company19/run2 genesis and frozen contributions at revision1,
then inserts SEVEN identical initial snapshots, revisions1..7. This explicit
padding exercises retention DELETE on real old rows; it is not prior gameplay
or invented run-log history. Founder21 and its genuine catalog grants remain
unchanged. Every later gameplay revision is produced by the real Store.

The original136 commands are retained. Two new-ID, same-time refusals per path
exercise `opportunity_not_pending` after claim and `burst_active` after spend.
Exact census:168 logged attempts,120 applied/48 refused,168 identical retries
after later state advancement,32 body-only same-ID conflicts,48 injected write
faults and16 corrupt-event-outbox negative controls. Parent fails on any deficit.

## Complete persistence oracles [V]

Each applied command advances exactly one revision; each refusal preserves the
complete encoded head and emits no events. Latest Company retention is exactly
five actual rows with the correct min/max revisions. Real claim clears pending,
the three buff arms install their actual effect and Lucky installs no buff;
PR2 ownership, attainment12/earned0, compute bank2000/burst3000, original buff
expiry, final beige101, positive permits and provision0(short)/1440(long) bind.

Every new attempt has exactly one run-log row, intent record and RECEIPT outbox
row. Complete receipt payload, owner/stream/scope/revision/hash agree. Event
delivery is separate: every event has exactly one outbox counterpart with its
source ID, owner/stream/scope/revision/hash/time and complete published JSON
envelope, with equal nonzero populations. One transactional payload corruption
per path demonstrably invalidates this bijection; rollback restores all rows.

Every later identical retry returns the original exact receipt with Replay=true;
each conflict preserves the original expected revision while changing only
opportunity_id or amount_ms and returns typed idempotency_conflict. Both leave
all twelve recorded tables unchanged, not merely counters or the latest head.

All168 stored transitions, including refusals, replay from immutable genesis
through actual ApplyLogged. Full stored receipts, ordered events and final
encoded Company head agree. Genesis, pin, frozen rows, Founder genesis/log,
stream ownership and verification queue remain unchanged, as does the complete
Founder head. Runs remain OPEN: completed-run verification returns log_gap;
this does not fabricate a completed-history proof.

Fault population is four short online/offline/online paths × actual claim and
burst × six writes: save_revisions INSERT, events INSERT, run_log INSERT,
intent_records INSERT, receipt-only transport_player_outbox INSERT and
save_revisions DELETE. Narrow owned PG triggers require the exact P0001/message
sentinel and compare all twelve tables after rollback. The outbox fault reaches
the late receipt write, not early event delivery; DELETE reaches a real row.
Owned triggers/functions are removed on exit without overwriting unknown objects.
Nontransactional sequence gaps are not claimed rollback failures.

## Failed instruments and discriminating probes [V]

- `62521/be1591` fails all sixteen adapters before Handle: the initial adapter
  omitted the replay envelope's intent ID. Exact raw reconstruction was
  predeclared in132d2925; no parser/runtime/output/population change.
- `11110/cd71d6` applies the first command in each path, then fails receipt
  census because it counted event rows too. Actual migrations40/42 establish
  the two message kinds. Amendment6c3c4173 filters receipts AND adds full event
  bijection,16 corruption controls and the late receipt-write fault predicate.
  It does not drop legitimate events or change migration/production bytes.
- Corrected baseline `1595/ce7c1e` passes8.116s with the entire exact census.
- Live claim resolution set to nil: compiling `64184/0f44a2` fails1.868s.
  Missing applied claim evidence is rejected before persistence. Four selected
  paths therefore cannot reach their exact DB sentinels; this is evidence
  admission discrimination, NOT successful execution of those rollback stages.
- Stored-intent lookup bypassed with AND false: compiling `90276/b30b4c`
  fails2.681s on all sixteen later-retry oracles. Actual receipts become
  revision_conflict/Replay=false after later state changes; all48 DB faults
  still execute. No byte/source-pin check substitutes for this behavior.
- Applied retention selector changed to through0: compiling `72787/2a0349`
  fails0.771s in every path at its first apply:8 actual rows[1,8], not5[4,8].

All mutants restored exactly after their handles terminated. No runtime,
save/schema/kernel/balance/RFC/owner-copy/CI change or residual probe remains.
An unquoted pipe in a later Make selector (`8f326d`) caused shell command errors
and no selected population; the correctly quoted command below was then run.
This launch error is not a product failure or a green test run.

## Final verification / limits [V]

- Broad actual production Integration population `33766/1b9130` passes23.538s.
- Focused `67854/18f02a`: all12 research/integration populations execute without
  skips6.866s; new SQL population4.70s has the exact census above. Existing
  anchor1215/16 and activation14fault populations still execute and pass.
- Full client `55740/9cf38a`:9640 pass/340 explicitly visible skips.
- `5047/8facb9`: zero TS/Svelte errors/warnings, full vet, client build213modules
  and CI topology with13 rejected negative controls pass.
- Cold `72556/d7fd17`: production39.331s fails ONLY original27 AC6 partitions;
  economy7.070s/decimal.244s/kernel.160s pass. Concurrent check times are not
  performance measurements and authorize no budget or acceptance-bound change.

Native declared Compose runs Postgres16.15/linux-arm64, not AMD64/hosted Actions.
Existing Actions server core provides real PG and discovers this normal Go test;
client discovers unchanged replay companions. Source routing/topology is not
hosted execution. Original AC6, historical RP-131 and deployment/release holds
remain. Existing Compose orphan warning is not cleanup authority.

Root reproduction (cold native DB population, no writer/artifact refresh):

```sh
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-run TestAxisSequencePersistenceIntegration -v'
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-run "TestAxis.*Research|TestAxisActivationWriteFaultsIntegration|TestAxisSequencePersistenceIntegration" -v'
make test-save-integration SAVE_TEST_PACKAGES='./production' SAVE_TEST_FLAGS='-run Integration -v'
make test-client
make typecheck vet build-client verify-ci-topology
make test-go GO_PACKAGES='./production ./economy ./decimal ./kernel' GO_TEST_FLAGS='-count=1'
```

The last command must still expose27 AC6 failures. Existing reports/source pins
are unchanged. Next reconcile remaining accepted CV4 migration-corpus and CV5/
AC3 receipt obligations before selecting a distinct implementation lane;
author-owned contradictions are findings, not permission to rewrite them.
R-012 representation and RP-308 episode meaning remain separate gates.
Entire new span afterf038a467 INCLUDING predeclarations/corrections/records
needs designated Claude review, independently of all older spans. No checkbox,
acceptance/status/archive/mint/push/deploy/release call or shortened1.0 scope.
