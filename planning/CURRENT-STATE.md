# CURRENT STATE — read this after CLAUDE.md, before anything else

**Last updated: 2026-08-16.** Written so a fresh agent (either side) can resume with no
conversational context. `CLAUDE.md` has the process laws; `rfc/README.md` has the RFC index;
this file has **where the project actually is right now and what is blocking what.**

---

## 1. Where the project is

- **The repository is published.** `origin/main` exists at the owner-controlled GitHub remote.
  Published commits are immutable; unpublished rewrites remain limited to the explicit protocol
  carve-out and require owner authorization plus byte-identity proof.
- Epoch 6 (`First Content`), epoch 7 (`T0-T1 Playable Content`), and epoch 8
  (`First-Hour Payoff`) are minted. Epoch 8 pins nineteen artifacts at
  `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
- **The T0–T1 first hour is implemented, designated-approved, and ARCHIVED** (archival verdict
  `b5d161d` over `755f428`): 97 deterministic pacing runs, zero dead purchasables,
  branch-specific first-company endings, run-2 starter packages, and a pinned-seed proof through
  the real gameserver and Postgres. All six acceptance criteria were re-walked against the tree
  and executed, with AC0 and AC4 each backed by a mutation probe. Canonical behavior now lives in
  `docs/t0-t1-playable-content.md`; the RFC and its full history are in `rfc/archive/` and
  `planning/archive/t0-t1-content/`.
- The Phase-A screens exist and work. A real Chromium drives a real gameserver against Postgres
  over a live WebSocket: bootstrap → Vision Slide → Desk → intents → visitor counter.
- **The working branch is 39 commits ahead of `origin/main`.** Everything since `ebcfc15` — the
  instrument repairs, both content epochs, AC0's offline-catchup fix, the whole first-hour payoff,
  the MIT licence — is local only, so hosted CI has not seen any of it (including a browser-gate
  repair that stays red on Actions purely because the fix has not reached the remote). Pushing and
  deployment remain owner-only actions.

## 2. Critical path

```
[HERE] Game-UI acceptance sweep + archival → Deployment Foundation → THE PUSH
```

The detailed T0–T1 instrument, balance, mint, rewrite, and review history is frozen in
`planning/archive/t0-t1-content/log.md`. Canonical shipped behavior lives in
`docs/t0-t1-playable-content.md`.

## 3. Ratified live identities

- Epoch 7: `sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789`.
- Epoch 8: `sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6`.
- The exact scenario, policy, curriculum, content, copy, and report hashes are recorded in the
  archived T0–T1 RFC and planning log. Changing any live balance byte requires the normal epoch
  and owner-ratification protocol.

## 4. Open debts

- Game UI still needs its post-content acceptance sweep and archival verdict. **Sweep it against
  the tree, not the log** — the T0–T1 close found AC0 unmet after the record implied otherwise, and
  found AC1/AC4 absent entirely after a mint had already been signed.
- Deployment carries two recorded items from the licence audit: the repo now has an MIT `LICENSE`,
  and the client bundle still needs a `third-party-licenses.txt` (three dependency MIT texts)
  served alongside it, since minification strips the notices MIT requires.
- Minigame surfaces (MA-C9), Soul Recovery's carried UI debt, Minigame Platform AC6, and Pet Care
  AC3 remain with their named successors.
- The Offer Sheet's “nothing on this sheet is hidden” copy versus withholding an unbound future
  network-slot row must be ruled before such content ships.
- Deployment and any push remain owner-gated.

## 5. Conventions established in-session that are NOT in CLAUDE.md's original text

CLAUDE.md's "Evidence discipline" section now carries rules 1–6; they are not repeated here. Two
additions that live only in this brief:

1. **Decision rounds go to Marco as `AskUserQuestion` dialogs WITH the context embedded in the
   question/option text** — prose written before a tool call is invisible to him. Plain language;
   no project jargon ("ceiling", "budget", "oracle") without a one-line translation.
2. **The instrument is a tool, not the project.** 2026-08-13's owner ruling exists because 48
   harness commits accumulated against 2 content commits. When verification cost starts compounding
   (tiers, attestations, identity hashes), stop and ask the owner what the check is FOR.

## 6. Working rhythm

Claude drafts RFCs and rules blockers; Codex implements and files blockers; **every batch gets a
designated cross-party review by the other side before archival** (Claude reviews Codex's
implementations; Codex reviews Claude's RFC/ruling commits). Marco rules design questions and
ratifies content hashes. Reviews are delegated to subagents with explicit probe instructions —
note the per-session subagent cap (`CLAUDE_CODE_MAX_SUBAGENTS_PER_SESSION`); it is cumulative,
not concurrent, and resuming an agent is free.

**THE PUSH is Marco's alone**, always.
