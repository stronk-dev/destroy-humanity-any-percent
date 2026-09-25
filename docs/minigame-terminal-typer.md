# Terminal Typer

Terminal Typer (`rfc/minigame-terminal-typer.md`) is the Tier-1, solo, session-skill typing tenant
on the Minigame Platform. **Status: fixture-first.** The engine, content loader and shared
content-gate corpus exist. No production epoch pins the `typer` artifact yet (OD-11), and the
platform amendments TT-PA1–PA4 land in later batches (`planning/minigame-terminal-typer/`).

## Content (`balance/testdata/typer-v1.json`, schema v1)

The exact-key catalog has `policy`, `eras` and `prompts`. Both loaders, `server/typer` and
`client/src/typer`, reject:

- unknown or missing keys;
- non-printable-ASCII prompt text, and leading, trailing or doubled spaces;
- unsorted prompt IDs;
- unknown eras or copy keys;
- `max_line_bytes` outside 1..1024;
- any reachable `era_tier` in 1..9 whose cumulative pool (OD-6) cannot deal `run_length`
  prompts.

The fixture's fifteen prompts are **provisional placeholders** (OD-9) and must be ratified or
replaced before any production mint.

## Engine (`typer` `1.0.0`)

- **Snapshot.** `typer.snapshot.v1` has exactly eighteen keys. Only the current prompt is ever
  exposed.
- **Prompt order.** A downward Fisher–Yates over the byte-sorted eligible pool, using the
  `typer.prompts.v1` substream of the `typer.run.v1` run seed.
- **Commands.** `begin {assist_level: timed|untimed}`, `submit_line {text}` and `end_run`. Exact
  keys are required, so any client-authored score, time or prompt ID is rejected.
- **Time.** Only `ApplyInput.ServerTimeMs`, the platform's per-command server sample (TT-PA1),
  supplies time. `t = max(sample, last_server_ms)`, so the run clock never rewinds.
- **Deadline.** A timed run has deadline `t_begin + timed_budget_ms`. A submit stamped exactly at
  the deadline is scored; one stamped later ends the run `timed_out`, unscored.
- **Comparison.**
  - `invalid_text` is decided first: C0, U+007F and U+FFFD are refused. Go substitutes U+FFFD for
    invalid UTF-8, so both runtimes refuse it.
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
replays every step in TS: rejections, the terminal snapshot, the result and the transition budget.
The corpus covers every prompt, all three outcomes, both modes, every rejection code, the
exact-deadline and one-late vectors, a backwards clock, the smart-punctuation map, a case miss and
a fullwidth look-alike miss.
