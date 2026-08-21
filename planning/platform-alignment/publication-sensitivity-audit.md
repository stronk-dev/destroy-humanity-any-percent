# D-002 Publication Sensitivity Audit

**Audit date:** 2026-08-21  
**Authority:** owner-ratified D-002 posture in `owner-ruling-packet.md`  
**Scope:** ignored files under `design/` and `planning/` that were presented as durable shared
memory, not ordinary build/cache output  
**Effect:** classification and routing only. Nothing in this audit publishes, force-adds, deletes,
or rewrites an ignored artifact.

## Verdict

The current ignore boundary is not an acceptable durable-memory model for a public repository, but
blindly tracking the ignored population would also be unsafe.

- There are **96 ignored artifacts** in scope.
- **89 are durable prose/data**, not generated output: 70 research files, `design/BACKLOG.md`, 17
  coverage-map files, and one consolidated historical fix queue.
- **Seven are generated diagnostic JSON files**. They are not canonical memory and should remain
  ignored.
- The high-confidence credential/path scan found **no private key, GitHub token, OpenAI-style key,
  or email-address match**. This is evidence against obvious secrets, not a guarantee that prose is
  publication-safe.
- One absolute workstation path is present in `design/research/cattery-reusables.md:3`.
- The sibling-repository research exposes more than that path: it names a live hostname, ports,
  service topology, filenames, endpoints, authentication gaps, and deployment mechanics. Treat it
  as private-source-derived material until the owner confirms the sibling repository itself is
  public or adopts a sanitized synthesis.
- The research corpus is rich in attributed quotations, allegations, court/regulatory events,
  trademark/copyright analysis, named people and companies, and sometimes explicitly
  model-derived or verification-pending claims. A secret scan cannot authorize its publication.

The industry-standard boundary for this repository is:

1. track durable project control-plane records;
2. keep reproducible generated diagnostics ignored;
3. publish research as source-linked synthesis with short necessary quotations and explicit
   verification status;
4. exclude private sibling implementation details, credentials, personal paths, and unneeded
   operational identifiers;
5. label historical snapshots as historical so they cannot become a second current authority.

## Reproducible population

The audited population is the sorted output of:

```sh
git ls-files --others --ignored --exclude-standard design planning
```

| Class | Count | Disposition |
|---|---:|---|
| A — durable control-plane candidates | 13 | Track after the stale disposition text is author-reconciled and each snapshot receives a current/historical authority label. |
| B — requires targeted sanitization or adoption | 9 | Do not track in current form. Apply the named correction or obtain the named owner ruling first. |
| C — research publication-rights pass | 67 | File-by-file source, quotation, allegation, privacy, and verification pass before tracking. |
| D — generated diagnostics | 7 | Keep ignored; retain only canonical summary evidence in tracked logs/reports. |
| **Total** | **96** | |

### Class A — durable control-plane candidates (13)

- `design/BACKLOG.md`
- `planning/coverage-map/README.md`
- `planning/coverage-map/deferred-and-dropped.md`
- `planning/coverage-map/fce-mint-punchlist.md`
- `planning/coverage-map/gap-backlog.md`
- `planning/coverage-map/map.md`
- `planning/coverage-map/research-integration.md`
- all six files under `planning/coverage-map/validated/`

No credential or machine-path hit was found in this class. These are useful repository memory, but
the coverage map is a 2026-08-05 reconstruction and contains old lifecycle/status claims. Before it
is tracked, its README and master outputs must say whether they are historical evidence or a
current authority; merely exposing stale rows would worsen repository truth.

### Class B — targeted sanitization or adoption (9)

| Artifact | Required treatment |
|---|---|
| `design/research/README.md` | Its “one private repo, never publish research” body contradicts the 2026-08-21 owner ruling. The ruling author must reconcile the normative body before the ignore policy changes. |
| `planning/coverage-map/decisions-log.md` | Preserve the historical 2026-08-06/07 sequence, but append the 2026-08-21 superseding public-shared-memory ruling and clearly mark the former final disposition superseded. This is ruling-author work. |
| `planning/codex-fixes-2026-07-30.md` | Labels the repo local-only and acts as a large historical second queue. Either archive it with a prominent frozen-through date and links to canonical successor records, or prove it redundant and remove it through an explicit owner-approved cleanup. |
| `design/research/cattery-reusables.md` | Remove the absolute local path and do not publish sibling topology, live hostname, endpoint/authentication details, or file-level extraction without an explicit sibling-publication ruling. Prefer a source-neutral reusable-pattern synthesis. |
| `design/research/cicd-deploy.md` | Remove sibling-repository file/topology recipe details and machine-specific benchmark identity; retain only Cloud Clicker conclusions and public-source evidence. Revalidate its now-stale CI cost claims against R-001. |
| three `planning/coverage-map/draft-artifacts/*.json` files | These are unadopted first-content drafts, not neutral planning metadata. Track only if the owning content authority adopts them or retain them as explicitly noncanonical examples without player-facing authority. |
| `planning/coverage-map/mint-content-rows-proposal.md` | Contains proposed/owner-authored content rows and mint history. Track only with an explicit historical/noncanonical label and confirmation that all adopted copy already has its canonical home. |

### Class C — research publication-rights pass (67)

This is every `design/research/*.md` file except the three Class-B research files above. The class is
not a claim that every file is legally risky. It records that the corpus deliberately mixes:

- direct and paraphrased source material;
- quotations of varying length;
- company and named-person conduct;
- litigation, criminal, safety and abuse allegations;
- trademark, patent, copyright and platform-policy analysis;
- `[P]` partial and `[M]` model-synthesis claims;
- “verify before shipping” queues that were designed for player copy, not public-source
  publication.

The pass must be file-by-file and record: source URL/primary-source preference, quote necessity and
length, allegation attribution, current verification state, personal/private operational detail,
and the chosen correction. It may approve a file unchanged, shorten/attribute it, replace it with a
synthesis, or keep only a public index plus a named private source store. “No secret regex hit” is
not an approval criterion.

**Progress:** Batch 01 reviewed `numeric-core.md`, `economy-kernel.md`, `browser-rendering.md`, and
`balance-enforcement.md` completely and found them eligible as dated technical research without
file edits. **Four of 67 are reviewed; 63 remain.** See `publication-rights-batch-01.md`.

### Class D — generated diagnostics (7)

- six `planning/archive/t0-t1-content/*.diagnostic.json` files;
- `planning/platform-alignment/r001-registered.diagnostic.json`.

These are failed/intermediate measurement products. Their durable findings already belong in
tracked logs, verdicts and canonical evidence artifacts. Tracking every diagnostic would create
noise and accidental second authorities; the existing `planning/**/*.diagnostic.json` ignore rule
is correct.

## Exact blockers and next queue

1. **Ruling-author reconciliation:** update the stale disposition bodies in
   `design/research/README.md` and `planning/coverage-map/decisions-log.md`. This audit does not
   perform that author-owned edit.
2. **Targeted sanitization:** produce public-safe replacements or explicit exclusions for the two
   sibling-derived research files; resolve the historical fix queue and four draft-content files.
3. **Control-plane tracking batch:** after steps 1–2, remove only the applicable durable paths from
   `.gitignore`, add the approved Class-A/B artifacts normally, and prove a fresh clone contains
   every artifact named by `AGENTS.md` as shared memory.
4. **Research rights batches:** review Class C in small thematic groups with a per-file disposition
   manifest. Do not make one 67-file blanket approval.
5. **Fresh-clone gate:** fail if a tracked authority links to an absent ignored artifact without a
   named durable private-store contract, or if generated diagnostics are treated as canonical
   evidence.

No step above authorizes a push, publication, deployment, content adoption, or product change.
