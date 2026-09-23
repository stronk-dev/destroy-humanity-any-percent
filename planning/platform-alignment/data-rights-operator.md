# D-008/D-009/D-015 — operator-persisted surfaces

Research checkpoint: 2026-09-23. Source after `dfdd6f4`. Method and controls:
[`data-rights-operator-plan.md`](data-rights-operator-plan.md). This is a
source/config and synthetic-test classification, not an inspection of private
host data, a clean-host R-006/R-007 verdict, legal advice or a new retention
rule. No live secret, backup, journal or operator service was accessed.

| Surface and sink | Exact current fields/content and configured lifecycle | Rights boundary / evidence limit |
|---|---|---|
| Encrypted off-host `.ccbackup` | The backup package contains `database.dump` (a complete Postgres custom-format dump), `release-manifest.json`, `epoch.json` and `metadata.json` inside an age X25519-encrypted tar. The plaintext authenticated envelope header contains `schema_version`, `backup_id`, `server_id`, manifest SHA-256, `epoch_id`, start/complete times, encrypted payload SHA-256/bytes and pre-upgrade/upgrade-resolved flags. Encrypted metadata additionally contains epoch and dump digests/size. Scheduled cadence is six hours plus pre-upgrade. Retention keeps all completed six-hourly objects for seven days and one per UTC day through day 30, **but always keeps the newest valid and every unresolved pre-upgrade object**. | A dump can contain all account/Founder, Guild, receipts, event and nested data present when taken. Account deletion does not edit encrypted prior objects. Thirty days is **not a hard maximum** while newest/unresolved protection applies. `RestorePostgresBackup` authenticates and restores the entire dump into a clean DB; it has no post-restore deletion replay. A pre-deletion backup can therefore restore a deleted account (RP-125, source-derived, not executed as a deletion scenario). Header server ID is plaintext operator identity; account rows are inside encrypted payload. |
| Application and service journal | All seven Compose services use `journald` with distinct `cloud-clicker-<service>` tags. Gameserver uses JSON `slog`; inspected ordinary calls use bounded error classes, causes and invariant kinds rather than raw errors/identifiers. `journal-observe` measures bytes/day and `journal-render` writes a host policy only for a complete observation with 14-day retention and enough budget for fourteen peak days; alert threshold is 80% of the journal reserve. Current stack has no raw-IP producer or enabled security sink. | Journal entries are host-persisted, not Postgres rows. A source allowlist for gameserver logging does not prove every dependency/container message is free of player data or that a supported host retained/purged exactly 14 days. No actual host journal or incident extract was inspected. R-006/R-007 must prove configured behavior; D-015 must classify incident extracts/access separately. |
| Private Prometheus and Alertmanager | Compose gives Prometheus `prometheus_data` with explicit `--storage.tsdb.retention.time=30d`; Alertmanager has `alertmanager_data` with **no explicit retention flag in the template**. Both expose private Compose ports, not host ports. Gameserver metrics use bounded route/method/status, job/result and invariant-kind labels; host textfiles hold backup/release/restore success/failure timestamps, filesystem use ratios, restart count and journal bytes/budget. Caddy, node-exporter, Prometheus and Alertmanager are also scraped. Alertmanager's configured receiver can be local or remote. | These are operational time series and alert records, not a player gameplay telemetry license. The inspected gameserver/host metric labels have no account/Founder ID or payload; that does not prove every external collector/receiver field or alert history is player-data-free. A private network is an access boundary, not deletion. No clean-host persistence/purge or receiver-delivery observation was performed here. |
| Release and rotation ledgers | Mode-0700 operator-state directory holds mode-0600 append-only JSONL. `ReleaseRecord` stores action, release/previous version and manifest digests, image digests, backup ID, times, rollback deadline, result/failure stage and `operator`. `RotationRecord` stores key family/action, current/previous **IDs** (not secret bytes), time and `operator`. The accepted overlap rules govern key removal, not ledger-row expiry. | The `operator` identifier is caller-supplied, validated as a bounded identifier but may identify a person. No record expiry, access-review or operator-data disclosure rule is present (RP-126). The ledgers do not contain key/password/recovery-code values by schema, but this is not proof no operator artifact elsewhere does. |
| Rehearsal and recovery-identity artifacts | The R-006 evidence schema contains run/manifest/version IDs, host OS/architecture/tool versions, step time/result/hash, population result/severing flag, RPO/RTO times, restored-identity match and artifact hashes. Recovery identity emits table/ownership/state/event/board/epoch counts and SHA-256 identities instead of row content. Bundle/release records also preserve content and image digests. | These can correlate exact releases, servers, runs and operator actions. Their lack of full rows is not equivalent to an adopted access or retention policy. The exact supported-host rehearsal is still unexecuted; local schema validators and synthetic fixtures cannot attest a deployed operator's actual files. |

## Executed evidence and refusal boundary

Cold `make test-go GO_PACKAGES='./deploymentbackup ./operations
./deploymentrelease' GO_TEST_FLAGS='-count=1'` passed all three package suites.
The repository's isolated `make test-deployment-backup` passed four real
Postgres 16 integration populations, including empty/populated encrypted
backup→restore identity, non-clean-target refusal and same-count mutation
detection. Its temporary `cloud-clicker-backup-test` Compose container,
network and ephemeral project volumes were removed by the Make target.
Those populations **do not** delete an account after backup and then restore
that prior backup; RP-125 is a source-derived consequence of full-dump restore,
not a separately executed resurrection witness. Neither local suite is
R-006/R-007 on the exact release bundle and supported clean host.

## Decision handoff

- **D-008:** a player export is not an encrypted operator backup. Decide whether
  any operator-held historical/incident material is disclosed or separately
  accessible; do not hand out a full shared database dump as one person's data.
- **D-009:** state backup lag and restoration behavior precisely. A deleted
  account may remain in protected backup objects beyond day 30 and may return
  after restoring a pre-deletion snapshot unless an accepted post-restore
  reconciliation exists. No such replay is currently implemented.
- **D-015:** rule purpose, access, trigger, duration and cleanup owner for backup
  exceptions, journal/incident extracts, Prometheus and Alertmanager stores,
  textfile observations, ledgers (including operator identity) and rehearsal
  artifacts. DP6/DP7 operational retention does not answer complete product
  or operator-data rights. R-006/R-007 must validate the built host path.

No operator runtime/secret was touched, and no policy or deployment change is
authorized by this dossier.
