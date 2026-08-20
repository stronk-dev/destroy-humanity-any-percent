# Deployment Foundation lifecycle audit

Coordinate: product tree `190a4fa`; audit checkpoint after `e7f8a91`; 2026-08-20.

This pass re-derived the draft Deployment Foundation from its complete RFC, active index, current
gameserver entry point and composition, transport/API policies, Make/Actions topology, every
Compose and packaging file, dependency-license audit, canonical operations documentation, and Git
remote state. It built and invoked the real binary but made no product, deployment, RFC, design,
canonical-product-doc, balance, content, or secret change.

## Bottom line

There is a real, buildable 29 MB arm64 gameserver binary with fail-closed mandatory key checks,
startup migration/epoch synchronization, health/readiness, and a bounded process drain. There is
also a real public remote and hosted Actions. Those are useful primitives.

They do not compose into the Deployment RFC's claimed product:

- all five Compose files are test-only; no production Dockerfile, Caddyfile, full-stack Compose,
  image, static-client server, clean-host contract, or deploy workflow exists;
- the binary reads epoch declarations, moderation data, and transport policy from a repository
  directory at runtime. It does not embed catalogs or client assets and serves no static client;
- WebSocket origin is catalog-pinned to `http://localhost:5173`, while the account API's trusted
  proxy depth is hardcoded to zero. Neither is deployment-injected as DP5 claims;
- runtime wiring accepts only one JWT key and one bootstrap-receipt key. The underlying libraries
  support previous keys, but `cmd/gameserver` cannot configure them and hardcodes the JWT key ID;
- no backup, restore, rollback, disaster, release-orchestration, or third-party-license delivery
  artifact exists; and
- hosted CI does not reach a current-head verdict because the harness job is killed at 30 minutes.
  There is no CD stage.

The draft's central “local-only / THE PUSH” premise is also obsolete: `origin/main` exists, the
repository is public, Actions runs exist, and the original product coordinate is present remotely.
AC5 is therefore the only acceptance outcome already true, despite the RFC remaining draft.

## Direct evidence

- `rg --files -g 'compose*.yml' -g 'compose*.yaml' -g 'Dockerfile*' -g 'Caddyfile*'` found only
  `compose.browser-test.yml`, `compose.ci-test.yml`, `compose.game-ui-test.yml`,
  `compose.numeric-test.yml`, and `compose.save-test.yml`.
- `make build-gameserver` succeeded. `file` identified a Mach-O arm64 executable and `du` measured
  29 MB. Invoking it with no runtime credentials exited nonzero with `invalid gameserver
  composition`, demonstrating the mandatory-secret refusal rather than only reading it.
- `server/cmd/gameserver/main.go` defaults `CLOUD_CLICKER_REPOSITORY_ROOT` to `.`, passes that root
  into composition, supplies `SigningKeys{CurrentID:"runtime", Current:key}`, and supplies only
  one current bootstrap key.
- `server/gameserver/composition.go` reads `moderation/guild-names.txt` and
  `balance/transport/phase0.json` from that root. No `go:embed`, `embed.FS`, `http.FileServer`,
  `ServeFile`, `client/dist`, or static route exists in the production server path.
- `balance/transport/phase0.json` contains only `http://localhost:5173`; the WebSocket handler
  performs an exact lookup against that list. `account.Phase0APIConfig` leaves
  `TrustedProxyHops` at zero even though `balance/api/phase0.json` declares one; the public API
  policy that reads the latter is not composed into the gameserver.
- A tracked-file scan found no private-key block or GitHub token pattern. The only checked-in
  database credentials are explicitly disposable test values in CI/test Compose. This is a
  bounded source scan, not proof about an eventual image, host, or secret store.
- The dependency audit classifies linked/runtime dependencies as permissive, but its required
  `third-party-licenses.txt` client deliverable does not exist.
- No tracked filename or implementation contains a backup/restore tool, `pg_dump`, `pg_restore`,
  production rollback runbook, or disaster rehearsal. Save-schema restore tests are application
  migration evidence, not database recovery.

## Acceptance classification

| AC | Verdict | Evidence and limitation | Required closeout |
|---|---|---|---|
| AC1 | **Unmet** | No production container, Caddy/static client, production Compose, supported host definition, or clean-host readiness run exists. The binary is disk-dependent and is not the claimed single artifact. | Owner rules D-001/D-006 and topology. Author reconciles DP1; accepted implementation supplies package, client/license delivery, clean-host positive and missing-file/config negatives. |
| AC2 | **Mechanical primitive only** | Gameserver composition and process shutdown prove startup migration/sync/readiness and bounded in-process drain. No deployment owner invokes stop→start across versions, checks the restart broadcast through the deployed proxy, rehearses migration failure, or proves rollback. | Reconciled release sequence plus an integration rehearsal over the actual package and database, including mid-write, failed migration, stale epoch, and bounded-drain controls. |
| AC3 | **Contradicted/incomplete** | Hosted Actions exists and most jobs pass, but the harness repeatedly exhausts 30 minutes, so current push CI has no verdict. The workflow has no delivery/deploy stage and no package/clean-host gate. | Resolve R-001 under accepted authority; define CI versus CD honestly; add package/rehearsal gates only after the deployment RFC is accepted. Never raise the timeout from the killed run. |
| AC4 | **Partial** | Missing current JWT/bootstrap credentials fail closed and no production secret value was found in tracked source. Runtime cannot configure previous rotation keys, cursor keys, proxy depth, or deploy origin as DP5 promises; no image/host secret scan or rotation rehearsal exists. | Exact deployment config schema, current/previous key inputs, origin/proxy wiring, secret-store contract, missing/invalid/rotation negatives, and tracked/image scan. |
| AC5 | **Proven external-state fact; body stale** | A public `origin` exists and hosted Actions runs on pushed commits. The product coordinate is remotely published. This does not prove the other four criteria or make the draft accepted. | Author removes the obsolete local-only/push narrative and records the already-completed phase transition without relabeling deployment complete. |

## Provider-off and operator trace

No mandatory identity, mail, analytics, ad, payment, AI, Redis, or other cloud provider is composed.
Postgres is the only external runtime dependency. Provider-off operation is therefore a genuine
architectural strength. A source tree plus a local binary is still not a supported self-host
bundle: the operator lacks the reproducible artifact, TLS/static tier, configurable public origin,
proxy contract, backup/restore, upgrade/rollback, secret rotation, and clean-host witness.

## Normative and lifecycle drift

1. Summary and DP4 say the repository is local-only and the push is pending; that external action
   already happened.
2. DP1 says the single binary has embedded catalogs; current canonical Economy docs explicitly say
   `go:embed` and hot reload are deferred, and production composition reads the repository tree.
3. DP3 says the full `make verify` gate already runs hosted. Current CI splits related targets,
   includes a blocking harness omitted by the draft's old dependency wording, and never reaches a
   full verdict at current product HEAD.
4. DP5 says origin, trusted proxy, cursor secret, Postgres, and rotating account keys are injected
   deployment data. Only Postgres plus one JWT/bootstrap key reach current composition.
5. The RFC depends on “CI Baseline (implementing)” even though that RFC remains active with a more
   complex current topology and lifecycle defects. No `planning/deployment-foundation/` exists,
   correctly reflecting that the draft has never been accepted for implementation.
6. `rfc/README.md` still labels the draft “THE PUSH,” routing a completed external fact as future
   work while the actual packaging/recovery work remains unaccepted.

## Smallest honest next order

1. Owner rules the release label/content floor (D-001/D-007), supported topology/RPO/RTO/rotation
   ownership (D-006), repository/research disposition (D-002), and self-host/sunset deliverable
   (D-003).
2. The Deployment author rewrites the obsolete Summary, dependencies, DP1, DP3–DP5, acceptance,
   and index label. The rewrite must distinguish the already-proven push from unimplemented deploy.
3. Accept a buildable Deployment RFC that owns artifact contents, origin/proxy/current+previous
   key configuration, license delivery, clean-host boot, release sequencing, backup/restore,
   rollback, and discriminating failure fixtures.
4. Implement only that accepted scope, run R-006 on the real deliverable, then hand the exact range
   to designated cross-party review. A green source-tree test suite cannot substitute for it.
