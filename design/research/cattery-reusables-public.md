# Reusable Small-Service Patterns

**Public synthesis date:** 2026-08-21
**Status:** research synthesis, not design or implementation authority
**Source boundary:** derived from a separately reviewed sibling implementation; project-specific
hostnames, paths, ports, routes, credentials, topology and filenames are intentionally excluded.

## Findings worth carrying forward

Small browser services benefit from a deliberately narrow operational shape:

- Bind application services to localhost or an internal network and give one reverse proxy
  ownership of public HTTP/TLS traffic.
- Give every long-running service a cheap health endpoint and make dependency readiness explicit.
- Authenticate administrative webhooks. Prefer a signed request over a reusable bearer value, and
  never expose the hook listener directly to the public network.
- Coalesce repeated rebuild/deploy requests with a single-flight lock, a short debounce and a
  pending-work marker so an update arriving during a run is not silently lost.
- Derive idle progress from persisted timestamps. A periodic loop may trigger materialisation, but
  elapsed time—not the number of timer callbacks—must determine the result.
- Treat WebSocket delivery as an optimisation over a recoverable state path. Initial state and
  reconnect recovery need an authoritative snapshot or history source.
- Keep shared-world actions server-validated and make actor type explicit. Human, system and bot
  actions should use the same state transition boundary when their permissions are equivalent.
- Separate long-lived relationship state from short-lived session expression. The former belongs
  in authoritative persistence; the latter can remain local and disposable.
- Model visible behavior as small composable states and queued transitions. Personality should
  bias deterministic choices rather than create a second rules engine.
- Respect reduced-motion preferences and make DOM-first animation degrade cleanly when the visual
  layer is absent.

## Limits

These are reusable patterns, not evidence that the sibling architecture fits Cloud Clicker. The
sibling system was small, lightly authenticated and not designed for MMO scale. Cloud Clicker's
binding server-authoritative, Postgres, transport, save-versioning and bot-parity rules remain in
`design/06-tech.md`, accepted RFCs and canonical docs.

Any implementation must be specified through the appropriate RFC. This synthesis does not approve
a deployment topology, endpoint, authentication mechanism, data model, pet mechanic or balance
constant.
