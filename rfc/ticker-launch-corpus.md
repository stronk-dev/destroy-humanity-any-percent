# RFC: Ticker v1 & the Launch Corpus Slice

- **Status:** draft — not implementation authority
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-09-25
- **Design refs:** `design/07` Phase 1 ("ticker/dispatch system v1 with the launch corpus slice")
  and §Sequencing principles 2–4; `design/08` §1 (voice rules 1–6), §2 (era table, T0–T2), §3
  ("Ticker corpus"); `design/09` §6 (dispatches are the ticker's surface) and §7 (watchdogs);
  `design/05` §2 (feed, NPC fallback) and the header's AI/solo-fallback law; `design/11` §1b
  (run-1 presence budget, T0 voice), §5 (surface inventory), §6 (strings as data, weights,
  no-repeat windows, copy linter), §7 (reduced motion); `design/12` §2 (ticker lines are drop-in
  data) and §5 step 3; `design/13` §2 (tier-up ceremony line, while-you-were-gone replay)
- **Depends on:** `rfc/feed-and-dispatch-foundation.md` (draft; this RFC needs its FD1 dispatch
  producer, **narrowed** per release-manifest VQ-3); the Copy Pipeline (archived,
  `docs/copy-pipeline.md`); WebSocket Transport (implementing, `docs/transport.md`); API
  Foundation (accepted, `docs/api-foundation.md`); Game UI (archived, `docs/game-ui.md`);
  `rfc/accessibility-player-workflows.md` (draft; A1/A3/A4 floor); `garage-player-surfaces`
  (undrafted; placement in era chrome and the `era_2010` copy era); `tier2-content` (undrafted;
  the T1→T2 gate)
- **Parent / amends:** child of `rfc/feed-and-dispatch-foundation.md`. It adds one dispatch kind
  (`ticker_ambient`) to FD1's closed list. It also amends the Copy Pipeline reference registry
  (a non-epoch content artifact class) and the API Foundation operation set (one read
  operation). Each amendment is listed under Deviations.
- **Supersedes / superseded by:** —
- **Planning:** `planning/ticker-launch-corpus/` (once implementing)
- **Evidence coordinate:** repository HEAD `c2d9bbc`. This is a static trace of design, docs,
  catalogs and schemas. No product code was run for this draft.

## Summary

This RFC specifies the v0.1 news ticker: a mechanism plus a sized, categorized slot plan for its
first corpus. The text of every line is owner-authored and **absent** from this RFC. Every line
slot below is a labelled `PLACEHOLDER` key.

The ticker has two server-authored streams:

1. A **personal stream.** Deterministic, stateless selections are drawn from a declarative
   corpus. Each selection is conditioned on the player's authoritative state and era, and is
   delivered privately through a read-only API operation.
2. A **world stream.** A rate-shaped `ticker_ambient` dispatch is published on `feed`, at most
   one per world interval. It draws on public aggregate world gauges, and labelled NPC-company
   bulletins take over when population is low (design law 6).

When the server is unreachable, a client-bundled fallback pool plays, labelled OFFLINE.

The slice is compatible with two owner rulings:

- **D-016.** Nothing about what a player saw is recorded, counted or logged.
- **D-017.** No player-authored text, no player names, no string parameters and no path from a
  client to `feed`.

## Motivation

`rfc/v0.1-garage-release-manifest.md` row G17 records the ticker as Absent, with the content
slice UNOWNED. The feed draft owns dispatch *mechanics*. It carries no content, no NPC fallback
(manifest E-2) and no player surface. This RFC fills three gaps without widening any of them:

- **Corpus format and selection.** A line is data (category, era, conditions, weight, parameter
  bindings). Its copy key resolves through the existing pipeline, so voice, provenance, the
  denylist and the statistic detector already apply.
- **Delivery.** The personal stream needs no per-player server tick (law 2). The world stream is
  a single world-level emission, never a per-player fan-out.
- **Surface contract.** A moving ticker is exactly the case WCAG 2.2.2 (Pause, Stop, Hide) and
  4.1.3 (Status Messages) govern. The contract is fixed here so that `garage-player-surfaces`
  places a finished component rather than inventing its accessibility.

**Out of scope:**

- the prose of any line (owner-authored, evidence discipline 6);
- the FD2 user-submission/curation pipeline and any user-authored broadcast (D-017);
- the real-player activity feed (`design/05` §2 "<name> sold their company"), which carries
  player names (D-017);
- feed prominence by Clout;
- GM dispatch composer and war log (`design/09` §7);
- HN/Slack/screenshot visual dispatch formats (text lines only);
- community-milestone and epoch-beat dispatches;
- ending datasets (`design/01` Endings A/B);
- eras beyond `era_2010`;
- hot-reload orchestration (see Deviations);
- the while-you-were-gone replay unless OD-6 rules it in.

## Specification

### TL1 — Artifacts and where they live

| Artifact | Path | Role |
|---|---|---|
| Corpus schema | `balance/ticker.schema.json` | Strict JSON Schema. It rejects unknown keys. |
| Shipped corpus | `balance/ticker/launch.v1.json` | The mechanics of every *adopted* line. It never contains a placeholder or fixture key. |
| Slot plan | `balance/ticker/launch.plan.v1.json` | The TL8 slot table as data: planned `line_id`, category, eras, conditions, weight, parameter bindings and claim requirement. It has **no text** and is never loaded by the server or the client. `ticker-check` uses it to report adoption progress. |
| Copy | `copy/catalog/ticker-launch.json` | Owner-authored copy entries, created only when the owner authors or adopts them. Same schema-version-1 entry grammar as every catalog. |
| Fixtures | `balance/ticker-testdata/*.json`, `copy/…` fixture catalogs under test directories | Mechanics proofs use only `fixture.ticker.*` keys. They are never loaded in production. |
| Golden vectors | `balance/ticker-testdata/selection-vectors.v1.json` | Shared by the Go and TS suites (TL4). |
| Generated client index | `client/src/ticker/generated/index.ts` | Built by `make copy-generate`: `line_id → {copy_key, category, npc_id}`, the fallback pool, `TICKER_CORPUS_HASH` and the presentation constants. |

The corpus is **not** an epoch artifact (OD-1). It is not registered in
`balance/epochs/phase0.json` and does not enter `constants_hash`. A ticker line never changes
state, a receipt, a replay or a leaderboard, so editing one must not mint an epoch or segment
boards.

Its identity is a new `ticker_corpus_hash`: SHA-256 over the repository's existing canonical-JSON
bytes of the shipped corpus. It is pinned as an added field in
`deployment/content-manifest.v1.json` beside `constants_hash` and `copy_hash`, and `copy-check`
recomputes it.

Copy-key references from the corpus are validated through a new registry,
`balance/content-artifacts.v1.json` (`{name, path, schema}` rows, byte-sorted). The Copy Pipeline's
reference walker reads it in addition to the epoch registry. `copy/references.v1.json` gains rows
for `ticker` at `/lines/*/copy_key` and `/npc_companies/*/name_copy_key`, both
`field_kind: copy_key`. The walker's "unknown epoch artifact" error becomes "unknown content
artifact", and it covers both registries.

### TL2 — Corpus schema (version 1)

```jsonc
{
  "schema_version": 1,
  "corpus_id": "launch",                      // mechanical ID
  "slot_ms": 30000,                           // personal slot length; integer 10000..120000
  "window_slots": 64,                         // entries per personal window; 8..256
  "no_repeat_window": 12,                     // W; 1..window_slots-1
  "refetch_min_interval_ms": 120000,          // TL5 client policy; 30000..3600000
  "world": {
    "interval_ms": 300000,                    // one world line per interval; 60000..3600000
    "ttl_ms": 900000,                         // display lifetime; interval_ms..86400000
    "npc_population_floor": 10                // TL6 published formula; 0..100000
  },
  "presentation": {
    "scroll_px_per_second": 60,               // motion mode only; 20..200
    "max_template_chars": 140                 // TL7 length gate
  },
  "eras": ["era_1995", "era_2000", "era_2010"],   // byte-sorted, ⊆ CopyEra
  "tier_era": [                               // exact tier → era map, ascending tier
    {"tier": 0, "era": "era_1995"},
    {"tier": 1, "era": "era_2000"},
    {"tier": 2, "era": "era_2010"}
  ],
  "npc_companies": [                          // byte-sorted by npc_id
    {"npc_id": "npc_a", "name_copy_key": "ticker.placeholder.npc_a_name"}
  ],
  "lines": [                                  // byte-sorted by line_id; line_id == copy_key
    {
      "line_id": "ticker.placeholder.t0_ambient_01",
      "copy_key": "ticker.placeholder.t0_ambient_01",
      "category": "era_ambient",
      "eras": ["era_1995"],
      "weight": 10,                           // integer 1..1000
      "conditions": [],                       // conjunction; TL3
      "params": {},                           // copy param name → TL3 param source
      "npc_id": null,                         // required non-null iff category = npc_company
      "claim_ids": []                         // required non-empty iff category = dated_fact
    }
  ]
}
```

**Categories** (a closed set; it grows only by RFC):

| Category | Stream | Selected by | Conditions may use |
|---|---|---|---|
| `era_ambient` | personal | weighted window pick | personal facts |
| `dated_fact` | personal | weighted window pick | personal facts |
| `state_reactive` | personal | weighted window pick | personal facts (at least one condition required) |
| `tier_up` | personal | TL4 ceremony pick, one per (run, tier) | `tier` only (exactly one `eq` condition) |
| `while_away` | personal | reserved. Zero lines are allowed and loading fails if any exist until OD-6 is ruled in. | — |
| `world_ambient` | world | TL6 world pick | world facts |
| `npc_company` | world | TL6 world pick, NPC branch | world facts |
| `offline_fallback` | client-local | TL4 fallback pick | none (era only) |

**Load-time validation.** The Go loader and `make ticker-check` (part of `make copy-check`)
enforce the same rules:

1. Exact keys at every level. `line_id` values are unique and byte-sorted, and
   `line_id == copy_key`. All IDs match the copy mechanical-ID grammar.
2. Every `copy_key` and `name_copy_key` exists in the compiled copy catalog. A line's `params`
   names equal its copy entry's `params` names exactly, and each source's type (TL3) equals the
   copy param type.
3. `eras` ⊆ the corpus `eras`, which ⊆ the copy runtime's `CopyEra` set. A line may target
   `era_2010` only once that era exists in `CopyEra`; until then, the loader fails on such a
   line. A line whose copy entry has `era_variants` must cover exactly the line's `eras`, or be
   `null`.
4. `dated_fact` lines: `claim_ids` is non-empty, and every claim is also listed on the copy entry's
   `provenance`. Claims resolve in `copy/provenance.v1.json` with `status: "verified"` (the
   pipeline already refuses to ship non-verified claims).
5. `world_ambient`/`npc_company` lines use only world facts and world param sources.
   `npc_company` lines name an existing `npc_id`. Personal lines use only personal ones.
6. **Pool floor.** For each shipped era, the count of `era_ambient` lines with **no** conditions
   is ≥ `no_repeat_window + 1`. For each shipped era, `offline_fallback` has ≥ 3 lines. The world
   stream has ≥ 1 unconditional `npc_company` line. These floors guarantee that every pick has a
   non-repeating candidate and that the NPC fallback always has a line.
7. **Shipping gate.** The shipped corpus and the production copy catalog contain no key with a
   `placeholder` segment and no key under `fixture.`. `launch.plan.v1.json` is the only file
   where `ticker.placeholder.*` keys may appear.
8. **Length.** Each copy template (every era variant), with each `{param}` counted as 12
   characters, is ≤ `presentation.max_template_chars` UTF-16 code units. `text_kind` must be
   plain (longform is rejected).
9. **Tier map.** `tier_era` is contiguous from tier 0 and each era appears once. The server
   treats a tier absent from the map as "no ticker corpus" (TL5).

### TL3 — Facts, operators and parameter sources

The closed v1 vocabulary is read only from fields the live `game_ui_snapshot.v3` projection and
the transport world snapshot already carry. Adding a fact requires an RFC.

**Personal facts** (evaluated at the Founder state the operation reads; TL5):

| Fact | Type | Derivation |
|---|---|---|
| `tier` | int | `run.tier` |
| `exit_count` | int | `run.exit_count` |
| `run_seq` | int | `run.run_seq` |
| `run_elapsed_ms` | int | `server_now_ms − run.run_started_at_ms` |
| `generators_owned_total` | int | Σ `generators[].owned` |
| `generator_owned:<generator_id>` | int | that generator's `owned` |
| `upgrade_owned:<upgrade_id>` | bool | that upgrade's `owned` |
| `resource_at_cap:<resource_id>` | bool | `cap` non-null and `amount ≥ cap` (numeric-core comparison) |
| `resource_log10_floor:<resource_id>` | int | `floor(log10(amount))` via the Go `Decimal` exponent; `−1` when `amount` is 0 |
| `gate_eligible` | bool | `transitions.cross_gate.eligible` (false when `cross_gate` is null) |

**World facts** (the transport world-snapshot state current at emission):
`online_founders`, `commons_active_founders`, `commons_health_ppm` (all int).

**Condition row:** `{"fact": "<fact>", "op": "eq"|"ne"|"gte"|"lte"|"in", "value": <int|bool|int[]>}`.

- `gte`/`lte`/`in` apply only to int facts.
- `in` takes 1–32 byte-sorted, unique ints.
- Bool facts accept only `eq`/`ne`.
- Every `<generator_id>`, `<upgrade_id>` or `<resource_id>` must exist in the current epoch's
  economy catalog at load and in `ticker-check`. A later epoch that removes an ID fails CI until
  the corpus is updated.

**Parameter sources** are numeric only, which enforces D-017 structurally:

| Source | Copy param type |
|---|---|
| `state.exit_count`, `state.run_seq`, `state.generators_owned_total` | `integer` |
| `state.resource_amount:<resource_id>` | `canonical_decimal` (the canonical string, verbatim) |
| `world.online_founders`, `world.commons_active_founders` | `integer` |

There is no `string` parameter source in v1. A line's own text and an NPC name key are the only
words the ticker ever shows. The copy pipeline's statistic detector ignores runtime
placeholders, and numbers that come from parameters are game state, not real-world claims.

### TL4 — Selection determinism

All selection is a pure function of (corpus, inputs). Nothing is persisted, and nothing about a
past selection is stored.

```text
H(parts…)  = SHA-256( "cloud-clicker/ticker/v1" ‖ 0x00 ‖ join(parts, 0x00) )
             where parts are ASCII: hashes as their "sha256:…" text, ints as base-10.
U(parts…)  = first 8 bytes of H(parts…) as big-endian uint64.

pick(C, u):  sort C by line_id bytewise ascending; T = Σ weight;
             x = u mod T; return the first line whose running weight sum > x.
             (Modulo bias ≤ T/2^64, which is documented and accepted.)

eligible(E, state, era, categories) =
             lines with category ∈ categories, era ∈ line.eras, and every condition true.

personal_window(corpus, state, founder_id, run_seq, s0, n):
  era = tier_era[state.tier];  if absent → return []
  E   = eligible(corpus, state, era, {era_ambient, dated_fact, state_reactive})
  recent = empty sequence (max length W = no_repeat_window);  out = []
  for s in [s0 − W, s0 + n − 1]:
      C = E \ recent
      if C empty: C = E \ {last element of recent}
      if C empty: C = E
      p = pick(C, U(corpus_hash, "personal", founder_id, run_seq, s))
      append p to recent (drop oldest beyond W)
      if s ≥ s0: append {slot: s, line: p} to out
  return out

ceremony(corpus, founder_id, run_seq, tier):
  era = tier_era[tier];  C = tier_up lines with era ∈ eras and condition tier == tier
  return C empty ? null : pick(C, U(corpus_hash, "ceremony", founder_id, run_seq, tier))

world_line(corpus, world, k):         # k = world slot index
  if world.online_founders < npc_population_floor:
       C = eligible(npc_company)
  else C = eligible(world_ambient);  if C empty: C = eligible(npc_company)
  return pick(C, U(corpus_hash, "world", k))

fallback_line(corpus, era, seed_id, s):   # client-local, TL6
  C = offline_fallback lines with era ∈ eras
  pick with the same no-repeat walk as personal_window, keyed
  U(corpus_hash, "offline", seed_id, s)
```

The slot is `s = floor(server_ms / slot_ms)`, using the server clock for personal windows and the
local clock for the offline fallback.

**The warm-up walk.** Walking from `s0 − W` makes the no-repeat property hold across window
boundaries without any stored history. With unchanged state, two windows computed from different
`s0` agree on every overlapping slot. When state changes, the eligible set changes and a repeat
across that one boundary is allowed. This is the accepted cost of statelessness.

**Parity.** `pick`, `personal_window`, `ceremony`, `world_line` and `fallback_line` have a Go
implementation (server) and a TS implementation (the client fallback, plus the shared vector
suite). Both pass `selection-vectors.v1.json`. Only the server's result is ever displayed for
the personal and world streams. The TS copy of the server functions exists so that the vector
suite can fail on divergence (the same discipline as the numeric golden vectors).

### TL5 — Personal stream delivery (read-only operation)

A new API Foundation operation, `get_ticker_window`:

- **Request.** `GET /api/v1/founder/ticker?from_slot=<int>`, authenticated exactly like
  `get_game_ui_snapshot`.
- **Parameters.** `from_slot` is optional and defaults to the current slot. When present, it must
  lie in `[current_slot − 1, current_slot + window_slots]`. Otherwise the operation returns 400
  `ticker_slot_out_of_range`.
- **Response** `ticker_window.v1`, exact keys:
  ```jsonc
  {
    "schema_version": 1,
    "ticker_corpus_hash": "sha256:…",
    "founder_revision": 123,          // the Founder revision the facts were read at
    "server_now_ms": 0,
    "slot_ms": 30000,
    "window_start_slot": 0,
    "entries": [ {"slot": 0, "line_id": "…", "params": {"n": 3}} ],   // length 0 or window_slots
    "ceremony": {"tier": 1, "line_id": "…", "params": {}} | null      // for the current tier ≥ 1
  }
  ```
- **Parameter values.** An `integer` source is a safe JSON integer. A `canonical_decimal` source
  is the canonical string.
- **Empty windows.** `entries` is empty only when the tier has no corpus era (TL2 rule 9). The
  client then hides the personal stream rather than showing lines from the wrong era.
- **The operation is read-only.** It reads committed Founder state through the same projection
  read path as `get_game_ui_snapshot`. It performs no persistence, no outbox insert, no event and
  no audit row, and it emits nothing on any WebSocket channel.
- **Params reflect state at `founder_revision`.** They may therefore trail live state by up to
  one refetch interval. This is disclosed in `docs/`; a ticker line is never an authoritative
  number.
- **Client fetch policy** (in `runtime.ts`, the only module permitted network access):
  - Fetch after each snapshot bind.
  - Fetch immediately when the bound `run.tier` changes.
  - Fetch when fewer than `window_slots / 4` future entries remain.
  - Fetch when the bound `founder_revision` has advanced and at least `refetch_min_interval_ms`
    has passed since the last fetch.
  - Never more than once per `refetch_min_interval_ms`, except on a tier change or bind.
- **Response handling.** The newest `founder_revision` wins. A response whose
  `ticker_corpus_hash` differs from the built `TICKER_CORPUS_HASH` is discarded. The client then
  shows the fallback and emits exactly one `ticker_corpus_mismatch` invariant through the
  existing production invariant reporter (the same sink contract as `copy_resolution_failed`).
- **Ceremony.** When the client observes the decoded `gate_crossed` event for tier *t*, it shows
  `ceremony` for *t* once, from the post-gate refresh the Game UI already performs. The ceremony
  host itself belongs to `garage-player-surfaces`; this RFC supplies the line.

### TL6 — World stream, NPC fallback and offline fallback

**World emission.**

- A single emitter in the world goroutine (never per player, never per connection) publishes at
  most one line per world slot. The world slot is `k = floor(server_ms / world.interval_ms)`.
- The emitter publishes at the first world-aggregator sample at or after the slot start, using
  the world snapshot current at that sample.
- **Payload.** The line travels as an FD1 dispatch of kind `ticker_ambient` on `feed`, using the
  transport's curated-`event` envelope rules for `feed` as FD1 binds them. Exact keys:
  ```jsonc
  {
    "dispatch_id": "ticker_ambient.<k>",   // deterministic: a restart re-publishes the same ID
    "kind": "ticker_ambient",
    "line_id": "…",
    "params": {"n": 4},                    // numeric only (TL3 world sources)
    "npc_id": "npc_a" | null,              // non-null iff an npc_company line
    "published_at_ms": 0,
    "expires_at_ms": 0,                    // published_at_ms + world.ttl_ms (the FD4 watchdog)
    "ticker_corpus_hash": "sha256:…"
  }
  ```
- **Decoding.** The Go encoder and TS decoder share a wire corpus and reject extra keys. A
  payload carrying any string other than these mechanical IDs and hashes is invalid.
- **Deduplication.** Clients dedupe by `dispatch_id` and drop lines past `expires_at_ms`.
  Transport's feed history (N = 50) serves late joiners.
- **The anti-firehose property.** The emission count is a function of wall time only. It is
  independent of population, clicks, intents and connections.

**The NPC fallback (law 6).**

- **Published formula** (rendered in `docs/` and in the NPC label's disclosure tooltip):
  *if players online < `npc_population_floor`, the world line is an NPC-company bulletin;
  otherwise it is a world bulletin, falling back to an NPC bulletin if no world line is
  eligible.*
- It runs with zero players online, so the stream is never dead.
- NPC lines pass through the same encoder, validation and wire schema as every other line (bots
  never get a different path).
- NPC lines are always labelled: a text chip that resolves the NPC company's `name_copy_key`,
  plus an owner-authored "NPC" label key. The label is never hue alone (A3).

**The run-1 presence budget** (`design/11` §1b).

- The world stream is a multiplayer signal, so the client neither subscribes to `feed` nor
  renders a world or NPC line while the bound `run.exit_count` is 0 (OD-4).
- The personal stream (single-player satire) runs from T0. The visitor counter remains run 1's
  only multiplayer signal.

**The offline fallback.**

- The client plays `offline_fallback` lines through `fallback_line` whenever no valid personal
  window covers the current slot. That covers four cases: server unreachable, the labelled
  local-only mode, a failed fetch, and a corpus-hash mismatch.
- `seed_id` is the persisted Founder ID when one exists, and the literal `local` otherwise.
- The surface shows an owner-authored OFFLINE label while fallback is active.
- The fallback pool, like every client-bundled line, is era-tagged. It is not conditioned on
  state, and it is never presented as server-authored.

### TL7 — UI surface contract (`TickerSurface`)

This RFC ships the component, and `garage-player-surfaces` places it in each era's persistent
chrome (OD-9). The component imports no transport code. It receives decoded windows, world
lines and the fallback state from `runtime.ts`, and resolves every string through `t()` with the
bound era.

1. **Structure.** A `<section>` labelled by an owner-authored heading key, and a pause/resume
   control. Behind a disclosure sits a "recent lines" list of the last 10 displayed lines, in
   DOM order as a plain list, so every line stays readable at the reader's own pace.
2. **Live regions.**
   - The ticker line container is **not** a live region (`aria-live="off"`; no `role="marquee"`,
     `role="status"` or `role="alert"`). Ambient, dated, state, world, NPC and fallback lines
     make **zero** AT announcements.
   - The one exception is the `ceremony` line. It is announced **once per tier-up**, with
     `polite` politeness only, through the Game UI's existing status region (not a new region).
     It is deduplicated by `(run_seq, tier)` and never announced as an alert. When the ceremony
     host moves focus to a new heading (A1), the ceremony line is folded into that context rather
     than announced separately, so there is no duplicate announcement.
   - The OFFLINE label's appearance uses the existing status channel once per transition.
3. **Motion.**
   - Without reduced motion, a line scrolls horizontally once at `scroll_px_per_second` (the
     era-appropriate marquee). The next line starts at the first slot boundary after the scroll
     completes. Slots that are passed are skipped, never queued. There is no blink and no flash
     (WCAG 2.3.1).
   - Under `prefers-reduced-motion: reduce`, lines are static and swap in place at slot
     boundaries, with no transform, fade or transition. Long text wraps rather than clipping.
   - Following A4, one `MediaQueryList` listener drives the mode live. Entering reduced motion
     mid-scroll ends the scroll immediately with the full line shown statically. Leaving it
     affects only the next line.
   - A future independent in-game motion setting is D-018's choice (not assumed here). If
     adopted, it drives the same switch.
4. **Pause (WCAG 2.2.2).**
   - A native `<button>` with `aria-pressed` and owner-authored labels. It is reachable by Tab and
     operable by Enter/Space, with a hit area of at least 24×24 CSS px.
   - Paused means no motion and no advancement: the current line stays until resumed, and new
     world lines queue behind at most one pending line, dropping any that expire.
   - Hovering the ticker, or focus within it, also suspends motion; it resumes on leave or blur
     unless the button is pressed.
   - The pause preference persists per device in `localStorage`, wrapped in try/catch and
     defaulting to unpaused. It is never sent to the server (OD-12).
   - While `document.hidden`, the ticker suspends and then resumes at the current slot with no
     catch-up burst.
5. **Reflow (A3).** At a 320 CSS-px viewport and at 400% zoom, the ticker causes no horizontal
   page scroll. The reduced-motion and paused modes wrap lines. The motion mode clips inside its
   own container only (the page never scrolls).
6. **Labels.** World, NPC and OFFLINE provenance is shown as text chips, not colour alone. An NPC
   chip's tooltip or disclosure states the TL6 formula with `npc_population_floor` as a
   parameter (owner copy).
7. **Failure.** A line whose copy resolution fails follows the Copy Pipeline production contract
   (loud mechanical key plus one invariant). The ticker never substitutes other text.

### TL8 — The launch slot plan (all PLACEHOLDER; copy pending owner)

Every row below is a **slot**, not a line. The keys are labelled placeholders. TL2 rule 7 makes
them unshippable, and a shipped line receives a final key when the owner authors it. Weights are
starting values and are config (the balance-numbers convention).

| Category | Era | Slots | Placeholder key pattern | Weight | Conditions / notes |
|---|---|---:|---|---:|---|
| `era_ambient` | era_1995 | 14 | `ticker.placeholder.t0_ambient_01..14` | 10 | None. The T0 "news blurbs of 1995–96" frame; earnest-webmaster voice, warmth not mockery (`design/11` §1b); era banned-word list applies. |
| `era_ambient` | era_2000 | 14 | `ticker.placeholder.t1_ambient_01..14` | 10 | None. The web-1.0 frame. |
| `era_ambient` | era_2010 | 14 | `ticker.placeholder.t2_ambient_01..14` | 10 | None. The flat-startup frame. **Blocked** until `era_2010` exists (TL2 rule 3). |
| `dated_fact` | era_1995 | 5 | `ticker.placeholder.t0_dated_01..05` | 6 | Candidate seeds and flags are listed below. Each needs a verified claim anchored in `design/research/provenance-extracts.md`. |
| `dated_fact` | era_2000 | 4 | `ticker.placeholder.t1_dated_01..04` | 6 | As above. |
| `dated_fact` | era_2010 | 0 | — | — | No tracked, verified T2 source exists (OD-8). |
| `state_reactive` | era_1995 | 5 | `ticker.placeholder.t0_state_01..05` | 30 | Suggested fact hooks: first generator (`generators_owned_total gte 1`), a resource at cap, `resource_log10_floor gte 3/6`, `gate_eligible eq true`. |
| `state_reactive` | era_2000 | 5 | `ticker.placeholder.t1_state_01..05` | 30 | Hooks as above, plus `exit_count gte 1` and `run_seq gte 2` (the run-2 "strictly faster" beat). |
| `state_reactive` | era_2010 | 2 | `ticker.placeholder.t2_state_01..02` | 30 | Blocked with `era_2010`. |
| `tier_up` | era_2000 | 2 | `ticker.placeholder.tier_up_1_01..02` | 1 | `tier eq 1` (T0→T1 ceremony). |
| `tier_up` | era_2010 | 2 | `ticker.placeholder.tier_up_2_01..02` | 1 | `tier eq 2`. Blocked with `era_2010` and `tier2-content`. |
| `world_ambient` | all | 6 | `ticker.placeholder.world_01..06` | 10 | World facts only; params only from `world.*`. Era variants, or era-neutral. |
| `npc_company` | all | 8 | `ticker.placeholder.npc_<a..d>_01..02` | 10 | Two per NPC company, at least one unconditional. Fictional companies only. |
| NPC names | — | 4 | `ticker.placeholder.npc_a_name..npc_d_name` | — | Fictional; must pass the denylist; never resembling a real firm. |
| `offline_fallback` | each era | 3 × 3 | `ticker.placeholder.t<0..2>_offline_01..03` | 10 | Unconditional. The T2 rows are blocked with `era_2010`. |
| Chrome keys | — | 6 | `ticker.placeholder.ui_heading`, `ui_pause`, `ui_resume`, `ui_recent`, `label_npc`, `label_offline` (plus `label_world` if OD-9 keeps a world chip) | — | Pause and resume labels must name the action plainly. |

**Totals.** 90 line slots (69 are T0/T1-shippable before `era_2010` exists; 21 wait on it), plus 4 NPC names
and 6–7 chrome keys. That is roughly 15% of `design/11` §5's ~600-line 1.0 target (OD-2).

**Dated-fact candidate seeds** (pointers only; the prose is owner-authored). They come from the
untracked dossier `design/research/era-1995-satire.md` §6b "Tickers". Each shipped line needs a
new, tracked `provenance-extracts.md` heading with public HTTPS sources, because the dossier
itself is unpublished (Copy Pipeline A1).

| Seed | Era | Dossier status | Flag |
|---|---|---|---|
| Browser company's 1995 IPO-day valuation | 1995 | `[V]` | 🟢 fact register |
| 1994 first banner ad click-through (44%) | 1995 | `[V]` | **Verify #2:** any modern-CTR comparison stays out until a citable figure is verified |
| Astronomer's 1995 newspaper prediction | 1995 | `[V]` | The subject is living; name the event, keep the frame respectful |
| Networking pioneer's 1996 collapse prediction and reversal | 1995 | `[V]` | The subject is living; respectful frame |
| Mass free-trial disc mailing (1996) | 1995 | `[V]` claim-as-claim | **Verify #1:** ship only as "an executive claims…"; never as a measurement; never name the executive (🟡→🔴) |
| Screensaver news company declines $450M (1997) | 1995 | `[V]` | 🟢 |
| Screensaver news company accepts $7M (1999) | 2000 | `[V]` | 🟢 |
| Homepage neighbourhood acquired for $3.57B (1999) | 2000 | `[V]` | 🟢 |
| Sock-puppet Super Bowl ad, $1.9M (2000) | 2000 | `[V]` | **Verify #9:** never "seventeen"; only "12 of 61" or "over a dozen"; no sock-puppet likeness (🔴) |
| 2001 crash, ~five trillion | 2000 | `[V]` per `culture-genx.md` | Figure re-sourced into the tracked extracts |

**Excluded until verified** (the dossier's §6c verify-before-shipping ledger): the grocery-site
"prospectus-adjacent" clause (🟡 overclaim), the PointCast magazine-cover quote (#6), the Barlow
Declaration text (#7), the Krugman fax-machine quote (#10), Kagi (#5), Kozmo specifics (#3), and
turbo-button history (#8). A slot may not be filled from this list until its claim is `verified`
in `copy/provenance.v1.json`.

The `societal-satire.md` §1 ticker seeds are T2-shaped. They contain unsourced specific numbers
and have no verify ledger, so they are **not** eligible for `dated_fact`. Under OD-3 they may
inform `era_ambient` lines only after their numbers are rewritten as in-fiction parameters or
removed.

**Real-world names** appear only in the `dated_fact` factual register, naming the event rather
than the person (the dossier's legal matrix and `internet-culture.md` §6). All lines pass
`moderation/copy-denylist.txt`. Voice rules 5 and 6 apply without exception.

### TL9 — D-016 and D-017 conformance (structural, not promised)

- **D-016 (no gameplay telemetry).**
  - The ticker has no impression, click, dwell, skip or pause event.
  - Nothing it does is stored, whether server rows, outbox rows or events.
  - No log line carries `line_id` together with an account or Founder identity.
  - No metric has a `line_id`, `category` or Founder label.
  - No A/B assignment exists. The seed inputs are the Founder ID, run and slot only.
  - The pause preference is device-local.
  - The world stream reads only gauges the world snapshot already publishes. It creates no new
    aggregate.
- **D-017 (no public UGC).**
  - No player-authored input reaches any ticker field.
  - No line names or counts an individual player.
  - The parameter vocabulary is numeric-only.
  - The feed payload is closed and carries mechanical IDs only.
  - Clients cannot publish (the transport law, re-asserted).
  - FD2's submission pipeline is not implemented by this RFC.

### TL10 — Docs

The implementing change adds `docs/ticker.md` (the corpus format, the selection function, the
published NPC formula, the staleness disclosure and the surface contract). It updates the
following pages in the same change:

- `docs/copy-pipeline.md` (the content-artifact registry and `ticker_corpus_hash`);
- `docs/transport.md` (the `ticker_ambient` feed payload);
- `docs/api-foundation.md` (`get_ticker_window`);
- `docs/game-ui.md` (`TickerSurface`).

## Deviations from design

1. **Hot reload (law 4).** The corpus is declarative data, loaded at boot. Hot-reload
   orchestration does not exist for any artifact (`docs/economy-kernel.md`), so a content change
   is a deploy. The artifact format is reload-ready.
2. **Content packs (`design/12` §1).** No pack loader exists. The corpus lives at
   `balance/ticker/` as a non-epoch content artifact (OD-1).
3. **Selection model (`design/11` §6).** "Priority + cooldown" is realized as integer weights, a
   corpus-global no-repeat window, and category pre-emption (`tier_up` is a separate one-shot
   pick). There are no per-line cooldowns.
4. **Parameterized company names (`design/08` §3).** There are no string parameters in v1: no
   naming mechanic exists (Presentation v3 owns the pre-naming `Founder` constant), and string
   sources would weaken the D-017 guarantee. Numbers only.
5. **The live feed of player actions, and feed prominence by Clout (`design/05` §2).** Both are
   excluded (D-017; Clout v1 unshipped). The world stream is aggregate-only.
6. **While-you-were-gone ticker replay (`design/13` §2, tagged `[v0.1]`).** Reserved and
   deferred (OD-6). The snapshot exposes no offline span to condition on (DESIGN-GAP TG-1).
7. **Visual dispatch formats (`design/09` §6).** HN pages and Slack screenshots are out. v1 is
   one-line text.
8. **Bank sizing (`design/08` §3, "~30% of each bank").** The slice is sized per era and
   category, with pool floors derived from the no-repeat window (OD-2).
9. **Amendments to parents.**
   - FD1 gains the `ticker_ambient` kind.
   - The Copy Pipeline gains a non-epoch content-artifact registry and a third hash in the
     content manifest.
   - API Foundation gains `get_ticker_window`.

   Each owning doc changes in the implementing change.
10. **Voice rule 1 versus rule 3 (`design/08` §1).** Rule 1 prescribes "one precise fake
    statistic"; rule 3 requires every ticker statistic to be sourced. This RFC applies OD-3's
    recommended reading and flags the text for the ruling author to reconcile (evidence
    discipline 5). It does not edit `design/08`.

## Acceptance criteria

Each criterion names the demonstrated failing case it must ship with (evidence discipline 1).
Gate claims use `-count=1`.

1. **TL-AC1 Corpus load.** The Go loader and `ticker-check` accept the fixture corpus and reject
   each of the following with a named error: an unknown key; a duplicate or unsorted `line_id`;
   `line_id ≠ copy_key`; weight 0 or 1001; an unknown category, fact, op or param source; a
   param-type mismatch against the copy entry; an unknown generator/resource/upgrade ID; a
   `dated_fact` without claims or with a non-verified claim; a pool-floor violation; an
   over-length template; a `longform` entry; an `era_2010` line while `CopyEra` lacks it; any
   `while_away` line.
   *Failing case:* each rejection fixture passes under a loader with that rule deleted, and the
   suite demonstrates this with a mutation.
2. **TL-AC2 Shipping gate.** `copy-check` fails when the shipped corpus or production catalog
   contains a `placeholder` segment or a `fixture.` key.
   *Failing case:* a fixture production corpus containing `ticker.placeholder.t0_ambient_01` is
   red.
3. **TL-AC3 Determinism parity.** The Go and TS suites both pass `selection-vectors.v1.json`.
   The vectors include discriminating cases: changing only the Founder ID, the slot, the run,
   the corpus hash or the line sort order changes the expected pick.
   *Failing case:* replacing bytewise sort with insertion order, or swapping `mod` for a
   truncating division, turns at least one vector red in each suite.
4. **TL-AC4 No-repeat and boundary.** For 10,000 seeded (Founder, `s0`) pairs over the fixture
   corpus, no `line_id` recurs within `no_repeat_window` consecutive slots, and windows with
   different `s0` agree on overlapping slots.
   *Failing case:* removing the warm-up walk fails the cross-boundary assertion.
5. **TL-AC5 Conditions.** Every fact and op is evaluated against fixture snapshots. A line gated
   `tier gte 1` is never selected at tier 0, and a `resource_at_cap` line appears only at cap.
   *Failing case:* an evaluator that ignores conditions fails. A `resource_log10_floor` that uses
   float64 instead of the Decimal exponent fails a vector at `1e308 < amount`.
6. **TL-AC6 Read-only operation.** `get_ticker_window` is in the generated API client and
   `api-check` pin. It returns 401 unauthenticated and 400 `ticker_slot_out_of_range` outside
   the bounds. In a real-Postgres integration test, a call leaves every table's row count and the
   player outbox unchanged and publishes nothing on any channel.
   *Failing case:* a test-only injected write in the handler turns the row-count assertion red.
7. **TL-AC7 D-016.** A test scans the metrics registry, a captured structured log of 100 window
   calls, and the migration set, and finds no ticker-labelled series, no log record pairing
   `line_id` with an account or Founder identity, and no ticker table.
   *Failing case:* adding a `line_id` metric label or a log field turns each scan red.
8. **TL-AC8 D-017 wire closure.** The shared feed-payload corpus is rejected by both the Go
   encoder and the TS decoder when it has an extra key (e.g. `player_name`), a string-valued
   param, a missing `npc_id`, or a non-null `npc_id` on a `world_ambient` line.
   *Failing case:* each vector passes if exact-key validation is disabled.
9. **TL-AC9 Rate shaping.** A simulated 24 h with 0, 1, 500 and 5,000 connected Founders emits
   exactly `24 h / interval_ms` `ticker_ambient` dispatches in every case. A restart mid-slot
   re-publishes the same `dispatch_id`, and the client shows it once.
   *Failing case:* a per-connection emitter makes the 5,000 case count wrong, and a random
   `dispatch_id` makes the restart case show two lines.
10. **TL-AC10 NPC fallback.** Below the floor, every world line has a non-null `npc_id` and
    renders the NPC chip. At or above the floor, world lines are chosen unless none is eligible.
    With zero players, lines are still emitted.
    *Failing case:* disabling the NPC branch leaves the zero-population run with no line.
11. **TL-AC11 Run-1 budget.** With `exit_count = 0`, the browser opens no `feed` subscription and
    renders no world or NPC line. After the first Exit, it subscribes.
    *Failing case:* a build that subscribes at bootstrap fails the subscription assertion.
12. **TL-AC12 Offline fallback.** With the gameserver unreachable, or on a corpus-hash mismatch,
    the surface shows era-matched fallback lines with the OFFLINE label, and a mismatch emits
    exactly one `ticker_corpus_mismatch` invariant.
    *Failing case:* removing the label or the invariant turns the browser assertion red.
13. **TL-AC13 Accessibility.** Across Chromium, Firefox and WebKit:
    - axe WCAG 2.2 AA is clean.
    - A MutationObserver on every live region records **0** announcements over 120 s of ambient,
      world and fallback lines, and **exactly 1** polite announcement for a tier-up.
    - With reduced motion emulated from mount, and toggled mid-scroll, no line has a running
      animation or transform, and the full line is present.
    - The pause button toggles `aria-pressed` by keyboard, and freezes both slot advancement and
      motion (the line text is unchanged across 3 slots).
    - Focus within the ticker suspends motion.
    - A 320 px viewport and 400% zoom produce no horizontal page scroll.

    *Failing cases:* setting `aria-live="polite"` on the line container makes the announcement
    count non-zero; removing the `MediaQueryList` listener fails the mid-scroll toggle; a
    CSS-only pause fails the slot-freeze assertion.
14. **TL-AC14 Copy and provenance.** `make copy-check` passes with the adopted catalog. Every
    `dated_fact` entry cites a verified claim whose anchor exists in the tracked
    `provenance-extracts.md`.
    *Failing case:* a fixture `dated_fact` citing a claim with status other than `verified`, or
    an anchor into an untracked dossier, is red.
15. **TL-AC15 Composed witness.** A real gameserver, Postgres and Chromium (the
    `test-game-ui-composed` lane) bootstrap an account. The run fetches a window, renders a
    resolved line, and then renders the tier-1 ceremony line once after Gate.
    *Failing case:* severing the operation (returning 503) turns the witness into the OFFLINE
    fallback state and fails the "server line rendered" assertion, which proves the witness
    depends on the producer.
16. **TL-AC16 Canon.** The `docs/` pages in TL10 are updated in the implementing change. The RFC
    index row, status and archival follow RFC-0000, gated on a cross-party designated review
    whose cited ranges union to the full span.

**Release-content gate (not an implementation gate).** The v0.1 ship claim needs every
T0/T1-shippable slot either filled by owner-adopted copy or explicitly dropped by the owner.
T2 slots follow `era_2010`. Mechanics may be implemented and archived fixture-first before any
line is adopted, like the First Content Epoch foundations.

## Owner decisions

Recommended defaults are listed first. None is inferred by implementers.

1. **OD-1 Corpus artifact class.**
   - *Recommend:* a non-epoch content artifact with its own `ticker_corpus_hash` in the content
     manifest, and a content-artifact registry for copy-reference validation.
   - *Alternative:* register `ticker` as an epoch artifact. Every line edit then mints an epoch
     and segments boards.
2. **OD-2 Slice size and allocation.**
   - *Recommend:* the TL8 table: 90 line slots, 14 unconditional ambient per era, no-repeat 12,
     slot 30 s, window 64.
   - *Alternative:* the T0/T1-only subset (69 slots) for the first mint, with T2 following.
3. **OD-3 Voice rule 1 versus rule 3.**
   - *Recommend:* "fake" precise numbers are allowed only about in-fiction entities (the
     player's company via parameters, NPC companies). Every number about the real world carries
     a verified claim.
   - The ruling author reconciles `design/08` §1's text.
4. **OD-4 Run-1 gate for the world stream.**
   - *Recommend:* `exit_count ≥ 1` as the machine reading of "the activity feed unlocks at the
     scripted first failure".
   - *Alternative:* a specific scripted Run-End marker, if the owner means something narrower.
5. **OD-5 Personal delivery path.**
   - *Recommend:* the dedicated read operation `get_ticker_window`, which leaves snapshot v3
     untouched.
   - *Alternative:* a `ticker` block in a new snapshot v4.
6. **OD-6 While-you-were-gone replay (`design/13` §2 `[v0.1]`).**
   - *Recommend:* defer. The return-sequence owner (`garage-player-surfaces`) first exposes an
     authoritative offline-span fact. The `while_away` category is then enabled by an amendment.
   - *Alternative:* rule it in now, and this RFC adds `last_offline_span_ms` as a DESIGN-GAP
     producer requirement.
7. **OD-7 World cadence and NPC formula.**
   - *Recommend:* `interval_ms` 300,000, `ttl_ms` 900,000, `npc_population_floor` 10 online
     Founders, published in-game on the NPC chip and in `docs/`.
8. **OD-8 T2 dated facts.**
   - *Recommend:* zero until a tracked, verified T2 extract exists. T2 ships `era_ambient`,
     `state_reactive` and `offline_fallback` lines only.
9. **OD-9 Surface ownership.**
   - *Recommend:* this RFC ships `TickerSurface` and its contract. `garage-player-surfaces` owns
     placement in each era's chrome and the ceremony host.
   - Also decide whether a separate "WIRE"/world chip exists (recommend yes, text-labelled).
10. **OD-10 Default motion.**
    - *Recommend:* the scrolling marquee in motion mode (era-authentic), with a pause control and
      static swap under reduced motion.
    - *Alternative:* static swap for everyone.
    - Any independent in-game motion setting remains D-018's choice.
11. **OD-11 Copy tone.**
    - *Recommend:* reuse the existing tones: `diegetic` for personal lines, `corporate` for world
      and NPC lines, `lore_card` never. This avoids a copy-schema change.
    - *Alternative:* a new `ticker` tone.
12. **OD-12 Pause persistence.**
    - *Recommend:* device-local only. This creates no new player-data class under
      D-008/D-009/D-015.
13. **OD-13 FD1 narrowing.**
    - *Recommend:* accept the `ticker_ambient` kind into FD1 in the same edit that narrows the
      feed draft per VQ-3 (server-authored dispatches plus NPC fallback, no FD2 submissions).
14. **OD-14 Authoring workflow.**
    - *Recommend:* Marco authors, or explicitly adopts, each line into
      `copy/catalog/ticker-launch.json`.
    - Agent-drafted candidates (`design/11` §6 "Claude drafts at volume") live outside the copy
      catalog until adopted.
    - Implementers never fill a slot.

## Open questions

- **DESIGN-GAP TG-1.** No authoritative offline-span fact exists for `while_away` (OD-6).
- **DESIGN-GAP TG-2.** `era_2010` does not exist in `CopyEra`. All T2 rows wait on the
  `garage-player-surfaces`/`tier2-content` producer.
- **Blocker.** FD1 is a draft whose body still contains FD2's submission pipeline (release
  manifest E-2). This RFC cannot be accepted before FD is narrowed (OD-13).

## Changelog

- 2026-09-25: created (draft — not implementation authority) as the v0.1 child of Feed & Dispatch
  Foundation, from a static trace at `c2d9bbc`. All corpus text is pending owner authorship.
