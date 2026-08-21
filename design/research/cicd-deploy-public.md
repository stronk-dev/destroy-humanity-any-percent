# CI and Deployment Research — Public Synthesis

**Public synthesis date:** 2026-08-21
**Status:** dated research synthesis, not deployment or CI authority
**Current measurement:** R-001 local observation, recorded 2026-08-21

## Current evidence

The complete local observation took **974.510 seconds**. Standard pacing accounted for
**739.441 seconds** and active relevance for **212.927 seconds**. The original dossier's estimate
that a broad simulation workload was necessarily the dominant CI cost is superseded by this
instrumented result. A hosted-Linux observation is still required before choosing a permanent CI
topology or optimisation target.

The durable conclusions are narrower:

- Measure named phases and objectives before optimising or assigning CI budgets. A gate must expose
  exclusions, guard exhaustion and whether it reached its actual objective.
- Keep the routine blocking path short enough to run habitually. Put long statistical searches or
  far-horizon observations on explicit slower lanes without weakening the blocking assertions.
- CI and deployment are different jobs. A container builder or deployment orchestrator is not a
  substitute for a test runner with pass/fail dependencies and a matrix.
- Prefer repository-native verification entry points so local and hosted runs execute the same
  commands. Cache only performance, never truth-producing artifacts or required baselines.
- Build immutable, content-addressed release inputs. Production balance/content changes should pass
  the same schema, parity, policy and acceptance gates as code changes.
- Use expand/contract database evolution: additive schema first, compatible binary next, resumable
  backfill separately, destructive contraction only after old readers are gone. Production rollback
  normally rolls back the binary, not an already-applied migration.
- A single-node service should optimise for honest reconnect and recovery before adopting complex
  zero-downtime topology. WebSocket drain and recovery require explicit tests; ordinary HTTP
  shutdown behavior is not proof for upgraded connections.
- Bind internal services narrowly, expose health/readiness separately, authenticate deployment
  triggers and coalesce duplicate deploy requests without losing a request that arrives mid-run.
- Use synthetic, versioned save fixtures for tests and previews. Do not copy production player data
  into ephemeral environments.
- Policy gates prove bounded facts: manifest coverage, schema conformance, allowed origins and
  attestation integrity. They do not prove that a human assertion is true or that simulated pacing
  matches real players.

## Public sources to re-check when drafting an RFC

- [GitHub Actions documentation](https://docs.github.com/en/actions)
- [Docker Compose documentation](https://docs.docker.com/compose/)
- [Caddy reverse proxy documentation](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)
- [Go `http.Server.Shutdown`](https://pkg.go.dev/net/http#Server.Shutdown)
- [Evolutionary Database Design](https://martinfowler.com/articles/evodb.html)
- [Playwright CI guidance](https://playwright.dev/docs/ci)
- [Content Security Policy `connect-src`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/connect-src)
- [Sigstore documentation](https://docs.sigstore.dev/)

Vendor limits, prices, versions and platform policies are time-sensitive. Reverify them from primary
sources at RFC time. The R-001 hosted measurement, D-006 deployment decisions, D-011 release
integrity and D-014 CI topology remain the authority route. This synthesis approves no workflow,
runner, vendor, secret layout, host topology, deployment script or player-facing copy.
