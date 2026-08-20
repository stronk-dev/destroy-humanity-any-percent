# Release platform audit at `190a4fa`

> Repository/remote reality audit, 2026-08-20. This is not a 1.0 scope decision. `[V]` means
> directly verified from the named repository file, command, or hosted run. Product, schema,
> balance, RFC mechanics, and ruled copy were held unchanged.

## Method

The population was the RFC index, live planning records, canonical docs, production client/server
code, content artifacts, Make/Actions gates, packaging files, and current GitHub metadata. Each
capability was traced as intent -> producer -> consumer -> real workflow -> executable witness.
The audit was fixed to `190a4fa04958cc2a3b4e689804cd55682f6c6420`.

## Verified strengths

- **[V]** Go/TypeScript numeric parity, versioned Postgres saves, replay, authoritative intents,
  lazy production, and offline catch-up have cold executable evidence.
- **[V]** A real Chromium path reaches the Desk through Vite, the composed gameserver, Postgres,
  snapshot v2, and a real Centrifuge world subscription.
- **[V]** Epochs 7/8 contain substantive T0–T1 content and a composed server-side first-hour proof.
- **[V]** Account recovery authentication, refresh-family replay revocation, import normalization,
  and deletion/anonymization have real-Postgres tests.
- **[V]** Axe WCAG 2.2 AA checks cover all five Phase-A surfaces; reduced motion and persistent
  focus have executable checks.

## Verified defects

1. **RP-001:** current-head Actions has no successful workflow. Push run `32009994004` and nightly
   runs `32096019304`, `32212696707`, and `32328790752` are cancelled; the harness job is killed at
   30 minutes while `balance-harness -mode=check` is running.
2. **RP-002/RP-003:** only test Compose files exist. There is no production Dockerfile, Caddyfile,
   full-stack Compose contract, clean-host boot, backup/restore, or rollback proof.
3. **RP-004/RP-005:** account deletion exists only on the backend and data export is absent.
   Settings exposes neither capability.
4. **RP-006:** the one-time recovery code is silently stored with tokens in localStorage. There is
   no display/download, second-device recovery, or recovery UI. Optional email is unconfigured.
5. **RP-007:** `design/11 §1b` rules silent server-anonymous startup plus a labeled local-only
   outage fallback and later import. Startup implements the default; no local gameplay-save owner
   or fallback exists, and older Account/tech wording remains unreconciled.
6. **RP-008:** the composed browser witness reaches Desk. The full first-hour/Exit/run-end browser
   path remains Game UI AC1 and is blocked on body reconciliation.
7. **RP-009:** accessibility evidence is strong at component level but does not cover the complete
   keyboard, screen-reader, zoom/reflow, coarse-pointer, or account-rights workflows.
8. **RP-010:** only tiers 0–1 of nine and one real minigame exist. World/feed/events, most social
   surfaces, combat engines, tiers 2–8, and the terminal endings are absent.
9. **RP-011/RP-012:** third-party license delivery, a self-host bundle, export, sunset covenant,
   and preservation rehearsal do not exist as implemented release capabilities.
10. **RP-013/RP-015:** GitHub says the repo is public, while internal research says private and the
    ignore policy removes the shared backlog/research/coverage records from fresh clones.
11. **RP-014:** multiple tracked current-status and planning records contain superseded claims.
12. **RP-075:** the current runtime cannot accept its declared deployment origin/proxy policy.
    WebSocket origin is pinned to `http://localhost:5173`; Account trusted-proxy depth is hardcoded
    to zero while the uncomposed public API policy declares one.
13. **RP-076:** key-rotation primitives exist in Account, but production composition supplies only
    one JWT key (with hardcoded ID `runtime`) and one bootstrap key. Previous/cursor keys cannot be
    deployment-configured.
14. **RP-077:** the deployable-binary claim overstates the artifact. The built 29 MB gameserver
    reads repository files at runtime, embeds neither catalogs nor client assets, and serves no
    static client.
15. **RP-078:** Deployment AC5's public push is already true, while the draft/index still route
    “THE PUSH” as future work; AC1–AC4 remain unmet, primitive-only, contradicted, or partial.
16. **RP-079/RP-081:** invalid or missing browser credentials have no recover-existing-account
    branch; the production Settings surface has no rights/recovery controls even though a test-only
    destructive-confirmation navigation primitive exists.
17. **RP-080:** Settings claims offline progress is parked locally, but the runtime persists no
    gameplay save or intent queue and has no later flush/import consumer.
18. **RP-082/RP-083:** restored desired-behavior probes fail in Chromium, Firefox, and WebKit:
    lifecycle preemption drops focus to `<body>`, and Desk width is 647 px in a 320 px container.
19. **RP-084:** the production reduced-motion preference reaches CSS tokens but not the numeric
    shell's interpolation/pulse mode, and mid-session preference changes have no listener.
20. **RP-085:** health/readiness, job failure propagation, and drain are real, but production has
    no metrics export, request/access correlation, SLO/alert/dashboard/runbook, and composes the
    optional Production invariant counter as nil.
21. **RP-086/RP-087:** the documented 30-day intent-record pruner is never scheduled, and no
    complete retention schedule governs inactive anonymous accounts, archived identities,
    append-only histories, dead letters, projections, or application logs.
22. **RP-088:** designed-sunset research falsely treats a one-command self-host bundle as an
    existing near-zero-cost deliverable; current packaging has no production Compose/client,
    export, bots-default workflow, final artifact, mirror, or sunset rehearsal.

## Provider-off posture

**[V]** No mandatory external identity, mail, analytics, ad, payment, or AI provider exists in the
implemented runtime. This is a genuine strength, but not a self-hosting proof without production
packaging, secrets validation, clean-host boot, backup, and restore.

The deeper artifact/configuration trace and criterion verdicts are in
`deployment-foundation-lifecycle-audit.md`.
The complete player recovery/rights trace is in `account-rights-release-audit.md`.
The task-level accessibility trace is in `accessibility-release-audit.md`.
The supervision/observability/retention/preservation trace is in
`operations-retention-preservation-audit.md`.

## Conclusion

The honest claim is **a tested T0–T1 server-authoritative vertical slice with a real browser
bootstrap path**. The repository does not yet substantiate “1.0,” “self-hostable,” “recoverable,”
or “complete game.” D-001 must define the next milestone before RFCs implement a release floor.

This audit cannot choose that floor, substitute axe for assistive user testing, infer durability
without a restore rehearsal, or make any active RFC archival-eligible.
