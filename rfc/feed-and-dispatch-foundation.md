# RFC: Feed & Dispatch Foundation (curated presence, amplification-gated)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/05 §2` (the live feed, presence, dispatches — curated events, rate-shaped at source), `design/05 §Neopets adoptions` (amplification-gated curation: free on owned surfaces, editorial gate on anything broadcast), the no-free-text law (`design/00`)
- **Research:** `events-playstyles.md §Ops lessons` (dispatches = highest-ROI narrative device; GW2 stuck-event watchdogs), `neopets-social-history.md §3` (the Neopian Times editorial-UGC model; vote-based → social-capital-market hazard), `social-spaces.md` (moderation model)
- **Depends on:** Transport (implemented — `feed` channel exists, shallow history N=50); Events + World Layer (drafted — dispatches narrate their outputs)
- **Owner ruling honored:** breadth-first — the feed/dispatch MECHANICS + the curation gate, not the feed content.
- **Planning:** `planning/feed-and-dispatch-foundation/` (once implementing)

## Summary

Transport's `feed` channel exists and carries nothing. This RFC gives it a source: server-curated
dispatches (in-fiction HN front pages, leaked Slack screenshots — the highest-ROI narrative
device), the amplification-gated curation model (the one broadcast surface in a no-free-text
world), and the rate-shaping that keeps an idle MMO's feed from ever scaling per-click.

## Specification

### FD1 — Dispatches (server-authored, rate-shaped)

A dispatch is a server-emitted feed item: `{dispatch_id, kind, era_skin, copy_key, refs,
published_at}`. Kinds (closed, grows by RFC): `world_milestone` (a community milestone tier fell),
`world_first` (a category's first verified completion — the Ethical% broadcast moment),
`epoch_beat` (a mint/patch note in corporate voice), `situation` (a Layer-3 server event beat).
**Dispatches are the ONLY thing that publishes to `feed`** beyond curated events — no per-click,
per-purchase, or per-player firehose ever reaches a public channel (the transport D2 law, now
sourced). Rate-shaped at emission (the aggregator batches; a milestone fires one dispatch, not one
per contributor).

### FD2 — The amplification-gated curation model (the design law, structural)

The design's UGC template: **free expression on OWNED surfaces, an editorial gate on anything
BROADCAST.** This RFC ships the BROADCAST side's gate — nothing user-authored reaches `feed`
without passing a gate, and passing the gate is itself a coveted reward (the Neopian Times model).
Phase-0 broadcast surfaces are SERVER-authored only (dispatches); user-authored broadcast (a
future "your run made the front page" surface) lands behind the gate this establishes. The gate is
a closed pipeline: submission → automated checks (the copy legal lint, structured-only content) →
curation queue → published-as-reward. **Vote-based amplification is FORBIDDEN** (the Neopets
Beauty Contest lesson: vote-based UGC becomes a social-capital market — every amplification
decision is server/editorial, never player-vote).

### FD3 — Presence (aggregate, never per-player firehose)

Presence on `feed`-adjacent surfaces = the transport aggregate (join/leave + periodic count,
roster only on subscribe where the surface needs it — already the transport discipline). The feed
shows *curated* activity (a world-first, a milestone), never a raw activity stream. The Break
Room's ambient-population model (design/05 §3) is presence, not feed — this RFC owns the feed;
presence stays transport's.

### FD4 — The watchdog (the GW2 ops lesson)

Every dispatch and any future feed-event has a max-lifetime + forced-resolution watchdog (GW2's
stuck-event bug lesson: every event needs a watchdog). Dispatches expire from history (the shallow
N=50 already); situations that stall alarm the ops surface. No feed item can wedge.

## Acceptance criteria

1. Dispatch emission: a world-milestone/world-first fires exactly one rate-shaped dispatch to
   `feed` (not one per contributor — the anti-firehose proof); the wire vector conforms.
2. Amplification gate: user-authored broadcast is impossible without the gate (grep-proven: no
   path from a client message to `feed`); vote-based amplification has no code path (forbidden).
3. Curation pipeline: submission → legal-lint → queue → publish-as-reward round-trip (the
   editorial-UGC model, server-side).
4. Presence stays aggregate (no per-player feed firehose — the transport law re-asserted at the
   feed surface).
5. Watchdog: a stalled feed item forces resolution; history expiry proven.

## Changelog

- 2026-08-03: created (draft) — the feed's source + the amplification-gated curation law
  structural; dispatches as the narrative device.
