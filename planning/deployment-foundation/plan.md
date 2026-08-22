# Deployment Foundation implementation plan

- **RFC:** `rfc/deployment-foundation.md`
- **Owner:** Marco
- **Implementer:** Codex
- **Started:** 2026-08-22
- **State:** implementing
- **Implementation baseline:** `cd102d7` (accepted after designated cross-party review)

## Delivery sequence

The release package is the integration boundary. Work is split so each batch has one authority,
its own executable failure cases and a bounded review range. No batch may claim a later criterion.

| Batch | Scope | Acceptance witness | State |
|---|---|---|---|
| DP-A | Canonical production config, file-backed secrets, origin/proxy boundary, current/previous signing-key composition and shared startup/preflight decoder | Cold Go tests cover every required field plus missing file, malformed secret, insecure origin, wrong proxy depth, half-pair and duplicate ID/value negatives | ready for designated review |
| DP-B | Repository-independent gameserver/client/Caddy image inputs, production Compose topology, machine-readable schema, license/SBOM/release manifest generation and validation | Bundle builds from tracked inputs; clean extracted bundle reaches startup without a checkout; removed artifact and changed digest fixtures fail | ready for designated review (DP-A review remains independent) |
| DP-C | Encrypted off-host backup, restore, retention and manifest binding | Empty/populated Postgres restore witnesses plus corrupt/truncated/wrong-identity/wrong-manifest/partial-file negatives | ready for designated review |
| DP-D | Stop-drain-start release helper, release record, seven-day previous-version rollback without Down migrations | Real Caddy HTTP/WebSocket release and rollback population; severed courtesy frame, drain, migration, epoch and smoke paths fail | pending |
| DP-E | Private operations profile, metrics, journald policy and seven blocking alerts | Private reachability and retention fixtures; all seven alerts fire and severed metric/rule/receiver paths fail | pending |
| DP-F | Exact-manifest R-006 clean-host, provider-off, supply-chain and recovery rehearsal; canonical docs and lifecycle closeout | Clean Linux/amd64 release bundle proves AC1–AC8, RPO/RTO and rollback, then receives both required review gates | pending |

## Batch protocol

For every batch:

1. predeclare the exact paths, positive population, negative/severing fixtures and claims in the
   append-only log before changing product behavior;
2. update canonical `docs/` in the same commit as behavior;
3. run cold (`-count=1`) focused tests and the smallest relevant root verification target;
4. record the implementation range and Codex first-filter without calling it the designated pass;
5. hand the exact range to Claude for the cross-party designated review before treating the batch
   as approved; and
6. keep clean-host/destructive/release lanes outside push CI unless measured D-014 headroom and a
   separately authorized CI change permit otherwise.

## Acceptance map

| RFC criterion | Owning batch | Required proof |
|---|---|---|
| AC1 reproducible clean-host package | DP-B, final DP-F | Extracted exact bundle, no checkout/writable source tree, real Caddy browser path, removed-artifact failures |
| AC2 network/config/secrets | DP-A, DP-B | One public Caddy hop; production negatives; tracked/image seeded-secret scan |
| AC3 release/drain | DP-D | Real courtesy frame and bounded drain through release helper; five severing failures |
| AC4 backup/restore/RPO/RTO | DP-C, final DP-F | Empty and populated restore identity; six named corruption/interruption failures; measured objectives |
| AC5 rollback | DP-D, final DP-F | Previous manifest plus pre-upgrade backup on clean volume; no Down migration; wrong-input refusals |
| AC6 rotation | DP-A, DP-D | Current/previous runtime verification and governed overlap/removal ledger mutations |
| AC7 operations | DP-E | Private metrics, exact retention, storage-pressure evidence and seven fired/resolved alerts |
| AC8 provider-off/supply chain | DP-B, final DP-F | No optional provider credentials; digest/SBOM/license/provenance checks with tampering failures |
| AC9 existing public/CI fact | DP-F | Retain evidence without promoting it to deployment proof |
| AC10 claim gate/closeout | DP-F | Exact-manifest R-006, current docs/registers/logs and full-range dual review before archival |

## Final closeout gates

- [ ] AC1 reproducible clean-host package passes with a demonstrated removed-artifact failure.
- [ ] AC2 network/config/secrets passes with every named negative and a seeded-secret failure.
- [ ] AC3 release/drain passes with all five severing cases.
- [ ] AC4 backup/restore passes with identity equality, corrupt inputs rejected, RPO ≤ 6 h and RTO ≤ 4 h.
- [ ] AC5 rollback passes against the previous exact manifest without a Down migration.
- [ ] AC6 rotation passes for JWT/bootstrap/cursor and refuses premature removal.
- [ ] AC7 operations passes retention, privacy and all seven fired/resolved alert fixtures.
- [ ] AC8 provider-off and supply-chain checks pass, including digest and attribution failures.
- [ ] AC9 remains accurately described as an existing external-state fact only.
- [ ] AC10 exact-manifest R-006 and transactional closeout receive Codex first-filter and Claude designated approval.
