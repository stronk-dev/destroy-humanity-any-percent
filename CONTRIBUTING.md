# Contributing

Cloud Clicker is built almost entirely by AI coding agents working under human direction — the
code, the design documents, and the game copy are AI-produced; art and audio come from public
sources. That is not a caveat, it is how this project works.

## AI-generated contributions are welcome

Contributions made with or by generative AI are explicitly allowed — the maintainers' own workflow
is agentic. What matters is the same thing that matters for any contribution:

- **You are responsible for what you submit.** Review it before we do; "the model wrote it" does
  not lower the bar.
- **It must pass the gates.** `make verify` (formatting, tests, typecheck, the kernel and copy
  guards) is the floor for any change; non-trivial changes come with tests.
- **Follow the process.** Implementation follows accepted RFCs (`rfc/README.md`); design intent
  lives in `design/`; missing specification is a `DESIGN-GAP` to raise, never something to
  improvise. Read `rfc/0000-rfc-process.md` first.
- **No real-world assets with unclear provenance.** Art/audio must come from sources whose license
  permits our use; player-facing text follows the project's voice rules and claim-verification
  discipline.

## Licensing

The project aims to be as public-domain as possible: the Unlicense wherever our dependencies
permit, complying with the more restrictive copyleft license where a dependency requires it. By
contributing you agree to your contribution being distributed under the project's license terms.

## Ground rules

No real money anywhere in the project, in any direction. Server-authoritative design is law.
Balance lives in data files, never constants in code. See `CLAUDE.md` for the full set of
non-negotiable design laws — they are settled and not open for drive-by "improvements."
