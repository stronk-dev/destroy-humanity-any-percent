# History rewrite: the unpublication filter (2026-08-06)

Owner-approved `git filter-branch` over all 570 unpushed commits of main, removing the
internal-unpublished paths (the research corpus, the internal backlog and trackers) from every
commit; the tracked publishable provenance-extracts file was preserved. Nothing had ever been
pushed; no external ref cited the old hashes. The map below is old -> new, position-paired
(no commits were pruned). Planning-log verdicts citing pre-filter hashes resolve through it.

commits: 570
