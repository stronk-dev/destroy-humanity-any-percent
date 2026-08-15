# Dependency license audit — 2026-08-14 (pre-PUSH follow-up)

**Audited by:** Claude. **Method:** license classification of every module actually linked into
the server binary (`go list -deps ./...` → owning modules, license file content matched, not
metadata trusted) and every client runtime dependency's shipped LICENSE file. Full raw
classification retained in the session scratchpad; module list reproducible from `server/go.mod`.

## Verdict: CLEAN for distribution. Two owner items, no blockers among dependencies.

## Server (Go) — 37 modules linked into the binary

| License | Count |
|---|---|
| MIT | 19 |
| BSD-3-Clause | 9 |
| Apache-2.0 | 8 |
| BSD-2-Clause | 1 |

- **No copyleft (GPL/LGPL/AGPL), no unlicensed, no unrecognized licenses in linked code.**
- The wider module *graph* (131 entries) contains many undownloaded, unlinked entries (test-only
  and tool dependencies of dependencies — ClickHouse, Docker, OTel, mysql drivers via goose etc.).
  They are not compiled in and impose no distribution obligations. Re-check if direct
  dependencies change.
- The server binary is operated, not distributed to players; permissive-license obligations are
  minimal either way.

## Client (bundled and served to browsers) — 3 runtime dependencies

| Package | License | LICENSE file present |
|---|---|---|
| `break_infinity.js` 2.2.0 | MIT | yes |
| `svelte` 5.56.8 | MIT | yes |
| `@antimatter-dimensions/notations` 1.6.0 | MIT | yes |

devDependencies (vite, vitest, playwright, typescript, …) are build-time only and not shipped.

## Owner items

1. **RESOLVED 2026-08-15 — MIT.** `LICENSE` added at the repository root, copyright Marco van
   Dijk. The owner's first preference was the Unlicense; assessed as viable but the weakest
   footing of the options — public-domain dedication is not fully recognised under Dutch/EU law
   (moral rights are inalienable), so it degrades to its embedded permissive grant, and some
   organisations refuse Unlicense dependencies for that ambiguity. Under the owner's own stated
   condition ("is that viable, else MIT is fine") this resolved to MIT, which is practically
   identical, cleanly enforceable in the EU, and matches every dependency already shipped.
   Note recorded for completeness: neither MIT nor the Unlicense prevents a monetised closed
   fork; the owner considered and declined AGPL-3.0 with that trade-off stated.

   *Original finding:* **The repository itself has NO LICENSE file.** Published at
   `github.com/stronk-dev/destroy-humanity-any-percent` — legally "all rights reserved," which
   may be intended, but it is currently an accident of omission rather than a decision. Design
   law 1 (free, no money ever) suggests the *game* being freely playable; that does not force
   the *source* license either way. Owner decision needed before THE PUSH: pick a license (or
   an explicit "source-available, all rights reserved" README statement).
2. **MIT attribution in the client bundle.** MIT requires the license text to accompany
   "substantial portions" of distributed copies. The minified bundle strips comments. Standard,
   low-effort fix at Deployment time: emit a `third-party-licenses.txt` (three short MIT texts)
   served with the client, linked from the site footer or About screen. Fold into the Deployment
   Foundation RFC as an acceptance item.

## Follow-up trigger

Re-run this audit when: a new direct dependency lands in `server/go.mod` or
`client/package.json` `dependencies`; or before THE PUSH if more than 30 days pass.
