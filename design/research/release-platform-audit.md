# Release platform audit at `190a4fa`

> Repository/remote reality audit, 2026-08-20. This is not a 1.0 scope decision. Evidence labels:
> `[V]` directly verified from the named repository file, command, or hosted run; `[P]` partial;
> `[M]` unsupported synthesis. No product, design intent, or ruled copy was changed by this audit.

## Question and method

**Question:** does the current repository support a coherent public release claim, and which
capabilities are proven integrations versus mechanical fragments?

The population was the active/archived RFC index, live planning directories, canonical docs,
production client/server code, balance/content artifacts, Make/Actions gates, packaging files, and
the current GitHub repository/runs. Each capability was traced:

```
intent -> producer -> consumer -> real content/workflow -> executable witness
```

The audit was fixed to `190a4fa04958cc2a3b4e689804cd55682f6c6420`. A backend route without a
player workflow, a component without real data, and a green sub-job inside a cancelled workflow
were deliberately treated as incomplete.

## Confirmed strengths

1. **[V] Numeric/save/production integrity is real.** Go/TypeScript golden vectors, versioned save
   migrations, authoritative intents, replay, closed-form accrual, and offline catch-up are
   exercised by cold package tests and composed Postgres paths.
2. **[V] The browser is connected to the real server.** `make test-game-ui-composed` drives
   Chromium through Vite, the composed gameserver, Postgres bootstrap, snapshot v2, and a real
   Centrifuge world subscription. The current-head hosted `game-ui-composed` job passed.
3. **[V] T0–T1 server content is substantive.** Epochs 7/8 bind nine generators, ten upgrades,
   branch-specific first-company endings, and run-2 starter consequences to composed proof.
4. **[V] Account security primitives are unusually complete.** Real-Postgres tests cover Argon2id
   recovery, refresh-family replay revocation, import normalization, and deletion/anonymization.
5. **[V] Accessibility primitives exist.** Axe WCAG 2.2 AA checks cover all five Phase-A surfaces;
   reduced motion and focus preservation have tests.

## Confirmed release defects

| ID | Finding | Evidence | Route |
|---|---|---|---|
| **RP-001** | **[V] Current-head CI has no successful workflow verdict.** | Push run `32009994004` and nightly runs `32096019304`, `32212696707`, `32328790752` are cancelled. In the push run, five main jobs pass; harness is killed at 30 minutes while `balance-harness -mode=check` is still running. | R-001; CI Baseline. |
| **RP-002** | **[V] Production packaging is absent.** | Only test Compose files exist. No production Dockerfile, Caddyfile, or full-stack Compose contract exists; Deployment RFC remains draft. | D-001/D-006; Deployment rewrite. |
| **RP-003** | **[V] Backup/restore and rollback are unproved.** | No operator runbook or clean-host data-loss rehearsal exists. Save migrations are not a deployment backup strategy. | R-006; Deployment/ops contract. |
| **RP-004** | **[V] Player data export is absent.** | Account API exposes bootstrap/session/founder/import/state/intents/deletion, not export. Settings has no controls. | D-008; account successor. |
| **RP-005** | **[V] Account deletion is backend-only.** | `DELETE /api/v1/account` and Postgres proof exist; the Settings surface renders only save status and an account note. | D-009; R-003. |
| **RP-006** | **[V] Account recoverability is not a user workflow.** | The one-time recovery code is silently persisted with tokens in localStorage. No UI displays/downloads it or starts a recovery session. The copy says optional email may be added, but no UI exists and server docs say email is `not_configured`. | D-005; R-003; owner copy adoption required. |
| **RP-007** | **[V] Fully offline-anonymous play is a false current capability.** | RFC/design promise a local save and later import. Production Game UI bootstraps a server account; no local gameplay save owner exists. | D-004. |
| **RP-008** | **[V] Game UI acceptance stops before the full first hour.** | Current composed browser proof reaches Desk. Game UI AC1 remains open and its body is blocked on GU-C25–GU-C28 reconciliation. | R-004 after author action. |
| **RP-009** | **[V] Accessibility proof is component-heavy, workflow-light.** | Axe covers surfaces and Enter starts the attempt. No full keyboard first hour, screen-reader, zoom/reflow, coarse-pointer, or account-rights workflow is proven. | R-005. |
| **RP-010** | **[V] The designed product is mostly absent.** | T0–T1 of nine tiers and one real minigame exist. World/feed/events, most social/player surfaces, combat engines, tiers 2–8, and the designed terminal endings do not. | D-001/D-007; content DAG. |
| **RP-011** | **[V] Third-party license delivery is incomplete.** | MIT `LICENSE` exists; `planning/CURRENT-STATE.md` records the missing client `third-party-licenses.txt`; no production bundle supplies it. | Deployment/package acceptance. |
| **RP-012** | **[V] Sunset/self-host claims are intent only.** | `designed-sunset.md` proposes export and a downloadable bundle, but no accepted covenant, export, production bundle, or rehearsal exists. | D-003; R-006. |
| **RP-013** | **[V] Repository disposition contradicts reality.** | GitHub reports the repository public; CI log records the public-repo owner choice. `design/research/README.md` still calls it private. | D-002; owner reconciliation. |
| **RP-014** | **[V] Lifecycle records are stale enough to misroute work.** | Current state had wrong ahead/behind count; README denies the live UI adapter; Deployment assumes no push; active plans leave already-landed composition/mint steps open; July batch is still “current handoff.” | Platform-alignment reconciliation. |

## Provider-off and failure posture

**[V] There is no mandatory external identity, mail, analytics, ad, payment, or AI provider in the
implemented runtime.** That is a genuine strength. It is not yet a self-hosting proof because the
production package, secrets contract, clean-host boot, backup, and restore paths do not exist.
Optional email is explicitly unconfigured and has no player surface, so its failure behavior is
currently “feature absent,” not graceful provider degradation.

## Release conclusion

The repository supports an honest claim of a **tested T0–T1 server-authoritative vertical slice
with a real browser bootstrap path**. It does not support “1.0,” “self-hostable,” “recoverable,” or
“complete game” without owner decisions and subsequent RFC work. The exact next milestone label is
therefore D-001, not an implementation-agent assumption.

## What this audit cannot prove

- It does not choose the 1.0 content/release floor.
- It does not substitute automated axe output for assistive-technology user testing.
- It does not infer durability from unit/integration tests without a deployment rehearsal.
- It does not validate legal obligations beyond recording the repository's own license/compliance
  commitments.
- It does not make any active RFC archival-eligible.

