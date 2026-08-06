# Provenance extracts (tracked, publishable)

This file is the **tracked, publishable** provenance source for shipped-copy claims. The full
internal research dossiers are unpublished (see the repo's gitignore); each entry here restates the
claim-relevant fact with its public primary sources, so the copy pipeline's shipping gate validates
against version-controlled, publishable material only. Entries are append-only; anchors are stable.

## Speedrun category taxonomy

Speedrunning communities define per-game competitive categories — e.g. **Any%** (finish by any
means), **100%**, **Glitchless**, **Low%** — plus community-maintained rules and variables;
leaderboards are segmented by category and variable values, and category taxonomies proliferate
per community rather than by any central authority.

Sources: <https://en.wikipedia.org/wiki/Speedrun>

## Legal denylist third-party marks

`blizzard`, `club penguin`, `habbo`, and `neopets` are third-party marks. They are listed in
`moderation/copy-denylist.txt` **defensively**: the copy pipeline fails CI if any shipped player-facing
string contains them. Their presence here is the provenance for those denylist rows.

Sources: <https://tmsearch.uspto.gov/>
