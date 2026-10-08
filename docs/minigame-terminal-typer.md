# Terminal Typer

Terminal Typer (`rfc/minigame-terminal-typer.md`) is the Tier-1, solo, session-skill typing tenant
on the Minigame Platform. **Status: fixture-first.** The engine, content loader and shared
content-gate corpus exist. No production epoch pins the `typer` artifact yet (OD-11), and the
TT-PA1 command clock, TT-PA2 server unlock and TT-PA3 content chain are implemented in fixture
paths. TT-PA4's public command/snapshot wire remains blocked by the API C2 conflict
(`planning/minigame-terminal-typer/`); no default player workflow is claimed.

## Content (`balance/testdata/typer-v1.json`, schema v1)

The exact-key catalog has `policy`, `eras` and `prompts`. Both loaders, `server/typer` and
`client/src/typer`, reject:

- unknown or missing keys;
- non-printable-ASCII prompt text, and leading, trailing or doubled spaces;
- unsorted prompt IDs;
- unknown eras or copy keys;
- null or nonnumeric era `min_tier` values (a genuine numeric zero is valid);
- `max_line_bytes` outside 1..1024;
- any reachable `era_tier` in 1..9 whose cumulative pool (OD-6) cannot deal `run_length`
  prompts.

The fixture's fifteen prompts are **provisional placeholders** (OD-9) and must be ratified or
replaced before any production mint.

## Engine (`typer` `1.0.0`)

- **Snapshot.** `typer.snapshot.v1` has exactly eighteen keys. Only the current prompt is ever
  exposed. Both replay engines reject negative counters, unknown assist levels, malformed
  submission feedback and miss counters above the pinned catalog hardcap before applying a command.
  Nonnullable numeric fields must be numbers, never JSON `null`. Nonnull submission feedback
  must contain both `outcome` and `first_mismatch_index`, including an explicit null index
  for a cleared line. Go refuses these malformed inputs rather than coercing null to zero.
  A `typing` snapshot must have `prompt_index < prompts_total`; Go refuses an
  exhausted-but-still-typing snapshot before prompt lookup, matching TS's refusal.
  A genuine final prompt still completes; terminal snapshots may retain the full count.
- **Scaling.** The `typer.era_tier` input and stored snapshot identity must be in the pinned
  TT1 clamp range 1..9, even if content declares a valid tier-zero era with a full prompt pool.
- **Prompt order.** A downward Fisher–Yates over the byte-sorted eligible pool, using the
  `typer.prompts.v1` substream of the `typer.run.v1` run seed.
- **Commands.** `begin {assist_level: timed|untimed}`, `submit_line {text}` and `end_run`. Exact
  keys are required, so any client-authored score, time or prompt ID is rejected.
  Both engines admit only `solo` for creation and command application, in every phase;
  TypeScript validates the runtime mode rather than relying on its compile-time type.
- **Time.** Only `ApplyInput.ServerTimeMs`, the platform's per-command server sample (TT-PA1),
  supplies time. `t = max(sample, last_server_ms)`, so the run clock never rewinds.
- **Deadline.** A timed run has deadline `t_begin + timed_budget_ms`. A submit stamped exactly at
  the deadline is scored; one stamped later ends the run `timed_out`, unscored.
- **Comparison.**
  - `invalid_text` is decided first: C0, U+007F, malformed raw UTF-8 and unpaired JSON surrogate
    escapes are refused. A valid U+FFFD is a scored miss against the ASCII prompt. Go validates
    raw JSON before decoding so malformed input cannot be substituted into that valid character.
  - Then `line_too_long`, measured on raw bytes.
  - Then the closed OD-7 map (‘ ’ → `'`, “ ” → `"`, NBSP → space, — → `--`), with no other
    normalization.
  - Then an exact byte compare. A miss reports `first_mismatch_index`.
- **Result.** `completed | ended_early | timed_out`. The facts are `typer.assisted`,
  `typer.clean_lines` (used for payout and quality, identical for timed and untimed runs),
  `typer.elapsed_ms`, `typer.lines_cleared` and `typer.misses`. `rating_delta` is always null.
- **Encoding.** Snapshots are Go-canonical: sorted keys at every level, with Go's `<`, `>` and `&`
  escaping mirrored in TS.

## Parity and gates

`make typer-corpus` regenerates `testdata/typer/content-gate-v1.json` from the Go engine, and
`make typer-corpus-check` fails when the file is stale. `client/test/typer-content-gate.test.ts`
replays every step in TS: rejections, the terminal snapshot, the result and the transition budget
(all command attempts, including rejections).
The corpus covers every prompt, all three outcomes, both modes, every rejection code, the
exact-deadline and one-late vectors, a backwards clock, the smart-punctuation map, a case miss and
a fullwidth look-alike miss.

## Content pinning and composition (TT-PA3)

`typer` is a replay-catalog artifact name in both runtimes.

- **What its presence requires.** A pinned `typer` artifact requires `minigame_api`. It is legal
  only together with a `typer` definition row in `minigames` and a
  `{"engine_ref":"typer","engine_version":"1.0.0","minigame_id":"typer"}` tenant in
  `minigame_api`. Any one of the three without the others hard-fails load in Go and TS.
- **Content lookup.** `CatalogBundle.TenantContent(engine_ref, engine_version)` resolves the pinned
  bytes for Pitch or Typer. A Pitch-only bundle still resolves Pitch and does not resolve Typer.
- **Chain identity.** Both loaders require the Typer definition row itself to name
  `engine_ref: typer` and `engine_version: 1.0.0`; matching only its `minigame_id` is insufficient.
- **Starting a session.** The start coordinator requires the requested tenant's own content rather
  than Pitch's.
- **Registration.** `gameserver.Compose` registers `typer.NewTenant()` beside Pitch.

The TT1 row fixture is `testdata/minigame/pitch-typer-v3.json`, with the two-tenant API fixture
`balance/testdata/minigame-api-typer-candidate-v1.json`. No live epoch pins either one.

## Child surface status

`TyperTable` is an unregistered presentation-only child. It offers both timed and untimed starts,
accepts native text-field input (including paste and composition), and retains input focus
when that input still owns focus as a prompt advances. Responses never override a newer player
focus choice. Mode, Submit and End controls remain native tab stops while pending, expose
`aria-disabled`, and guard their callbacks against duplicate commands; the section exposes
`aria-busy`. Terminal transitions hand off focus from a removed child control to Leave, without
stealing newer focus or running after unmount.

A separate polite live region announces the new prompt text once per prompt ID change;
miss/clear feedback uses its own status. Chromium/WebKit checks execute forward/backward Tab
traversal, pending Enter/Space, prompt/lifecycle focus, reflow and axe. A complete native-keyboard
untimed component run now uses the actual TS engine and pinned placeholder content, including a
miss, correction, all prompts, exact result facts and Leave. The harness supplies synthetic
command times: this is not a public server/receipt/payout journey. Firefox/manual assistive-
technology and full public registration/acceptance remain open.

The countdown samples local monotonic time once when the authoritative response time or
snapshot revision changes. Both display-clock fields are assigned from the same non-reactive
local sample; the effect never reads a state field it writes. A browser case injects a clock
that advances on every read and checks ready/timed states and subsequent server revisions,
so coarse browser-clock equality cannot hide a reactive update loop. Engine timestamps and
payout remain server-owned and unchanged.

## API status

The v1 error taxonomy carries the Typer rejection details (`invalid_assist_level`, `invalid_text`,
`line_too_long`) and the TT-PA2 start details (`tier_required`, `curriculum_exit_required`). These
are exact 409 bytes on the minigame operations. **No v1 route carries Typer commands or snapshots
yet.** The accepted API Foundation C2 law forbids growing request unions inside v1. The transport
shape is an open DESIGN-GAP; see `planning/minigame-terminal-typer/log.md`.
