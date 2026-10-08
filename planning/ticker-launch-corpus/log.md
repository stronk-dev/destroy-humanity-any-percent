# Ticker launch corpus — append-only work log

## 2026-10-09 — TL4 overlap claim: bounded diagnostic

Before implementation, inspect the complete accepted RFC and its explicit Feed-narrowing
exception. No product changes are allowed yet. TL4 additionally claims a 12-slot warm-up makes
arbitrary overlapping personal windows agree. Test that claim directly, without altering its
SHA-256 domain, uint64 interpretation, weighted byte-sorted pick, exclusion or warm-up walk.

Initial read-only selector probe uses 14 equal-weight unconditional lines, identical Founder,
run, eligibility and hash, and adjacent 64-entry windows starting at 64/65. It exits 1 with 50 of 63
overlapping slots different, while both individual windows obey the 12-slot no-repeat bound.
That initial hash covered the line-set only; it is a selector diagnostic, not a full corpus hash.
The retained diagnostic instead hashes a complete synthetic TL2-shaped fixture, including
world/offline pool rows, using recursive canonical object-key ordering. No prose is authored
or loaded into the shipping copy pipeline. It does not prove production-loader acceptance.

Final commands:

- `node planning/ticker-launch-corpus/warmup-diagnostic.mjs --self-test`: exit 0. Exact-repeat
  and reversed-input-order controls agree; the comparator accepts equal records and detects
  one deliberately changed record; both windows contain 64 entries with zero internal repeats.
  The retained complete-fixture hash is
  `sha256:a89d5bb632ec23d6bd426f6b67be7c86136d4622943d9c9c068ec8c9e5d950b0`.
- `node planning/ticker-launch-corpus/warmup-diagnostic.mjs`: exit 1 at the actual TL-AC4
  overlap assertion: 55 mismatches of 63 shared slots. At slot 65, the first request picks
  `fixture.ticker.ambient_01` and the second picks `fixture.ticker.ambient_07`.
  A first diagnostic version printed the full array twice in its assertion error; the final
  version keeps the same zero-mismatch requirement with compact count diagnostics.

Why: each request starts with a different empty recent-history state. Excluding the preceding
12 choices makes those differences influence subsequent choices; walking 12 steps does not
guarantee the two histories converge. The privacy law forbids silently curing that by storing
what the player saw. RP-479 routes this to the RFC author alongside Feed's existing hold.
This is not an empirical pacing study, production bug, passed TL-AC4 or newly playable feature.
Only this diagnostic and its owning records changed; no unrelated software suites are needed.

Review by: Codex (diagnostic author/first filter). Recorded by: Codex. This does not amend
the author-owned RFC, adopt an alternative algorithm, narrow acceptance or unblock Feed.
Exact new diagnostic/record range starts after `f553640d` through this containing commit;
designated cross-party review pending. No acceptance, archival, content mint, CI or push change.
