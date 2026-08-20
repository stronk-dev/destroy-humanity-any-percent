# CURRENT STATE — read this after `AGENTS.md`

**Last updated: 2026-08-20 at repository commit `cb162a3`, product coordinate `190a4fa`.** This
file is a mutable resume brief. The evidence-backed repository program lives in
`planning/platform-alignment/`; RFC lifecycle truth still lives in `rfc/README.md` after
reconciliation.

## 1. What genuinely exists

- `HEAD` and `origin/main` were identical at `cb162a3` (`0` ahead, `0` behind) when this wave
  began. The GitHub repository is public; `cb162a3` is planning-only and the last product change
  remains `190a4fa`.
- The numeric core, economy/save/production engines, offline accrual, route/gate foundation,
  Commons/guild server foundations, account/session backend, realtime transport, replay/epoch
  machinery, and T0–T1 content have substantial executable evidence.
- Epoch 7 identity is
  `sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789`.
  Epoch 8 identity is
  `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
- The live content is a bounded vertical slice: nine generators, ten upgrades, tiers 0–1, one real
  minigame (The Pitch), a scripted first-company ending, and run-2 starter consequences.
- A real Chromium path proves account bootstrap, Postgres, snapshot v2, and the world WebSocket
  subscription through the production Game UI runtime. It reaches the Desk; it does not yet prove
  the complete first-hour browser workflow.

## 2. Current release truth

The repository is **not a coherent 1.0 release platform** at this commit.

- Current-product hosted CI has no successful workflow conclusion. Runs `32009994004`,
  `32096019304`, `32212696707`, `32328790752`, and the planning-only checkpoint run
  `32404232364` were cancelled because the harness remained inside
  `balance-harness -mode=check` when its 30-minute budget expired. Server, client, schema, browser,
  and composed-browser jobs passed individually. The same harness check completes locally at 12
  workers in roughly 18 minutes, but exposes no phase/row timings; R-001's instrumentation wave
  remains open.
- No production Dockerfile, Caddyfile, full-stack Compose contract, deploy/rollback flow,
  backup/restore rehearsal, or clean-host self-host proof exists.
- Account import/deletion/security primitives exist on the backend, but the UI has no recovery,
  export, import, deletion, or email-attach workflow. The one-time recovery code is silently stored
  in localStorage.
- The ruled local-only outage fallback is absent from the production client. Silent anonymous
  server bootstrap is the adopted default; older Account/tech text still needs body reconciliation.
- Accessibility proof covers axe, reduced motion, focus preservation, and Begin-via-Enter; it does
  not cover the full keyboard/screen-reader/zoom/touch release workflow.
- Tiers 2–8, most minigames, combat engines, player-facing social/world/events systems, and the
  designed terminal endings remain unimplemented.

See `planning/platform-alignment/release-platform-audit.md` and `capability-map.md` for the trace.

## 3. Record and lifecycle truth

- Game UI AC1 is blocked because its normative body still contradicts accepted GU-C25–GU-C28
  rulings. The ruling author must reconcile it before implementation resumes.
- Permits and First Content Epoch behavior appears landed and reviewed, but their live plans still
  show mint-time work open. They require a transactional closeout/range-union audit, not blind box
  flipping.
- Minigame Platform, Account, Transport, Prestige, and CI plans contain blockers or composition
  claims superseded by later work. Each requires acceptance reconciliation against the tree.
- `planning/run-genesis-archival-remediation/` still needs its recorded resolved state moved through
  the normal reviewed archival lane.
- The 2026-08-05 coverage map is historical and stale. It must not be used as a current count until
  its six slices are revalidated.
- `design/research/README.md` still contains a ruled-looking private-repository disposition that
  contradicts the public remote. This is owner decision D-002; an implementation agent must not
  silently rewrite the policy.
- The mandated design backlog, research dossiers/matrix, and coverage map are gitignored and absent
  from a fresh clone. The tracked interim ledger is `planning/platform-alignment/backlog.md` until
  D-002 establishes a durable public/private shared-memory model.

## 4. Critical path

```
R-001 current-head harness diagnosis
        +
active-RFC / planning closeout reconciliation
        +
owner release decisions and Game UI body reconciliation
        -> accepted, dependency-ordered implementation
        -> integrated release proof
```

Only the first two lanes are presently `READY`. The exact queue and blocked work are in
`planning/platform-alignment/execution-queue.md`.

## 5. Owner decisions now required

- Next public milestone label and exact release floor.
- Public/private repository disposition.
- Sunset/self-host covenant and deliverable.
- Recovery credential UX, data-export scope, and deletion disclosure.
- Supported deployment/backup topology.
- T0–T1 vertical slice versus nine-tier/ending content scope for the chosen milestone.

The full decision records and evidence prerequisites are in
`planning/platform-alignment/decision-queue.md`.

## 6. Working rules

- Do not start from old handoff prose. Start from the platform-alignment execution queue and the
  exact active RFC.
- Do not turn backend primitives into shipped-feature claims without a consumer, real workflow,
  and executable witness.
- Do not weaken or extend a measurement budget from a run that did not reach its objective.
- Claude/Codex cross-party designated review and full range-union remain mandatory before archival.
- Push, publish, deploy, and PR creation remain owner-authorized external actions only.
