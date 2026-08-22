# Designed sunset and preservation — public issue brief

**Evidence refreshed:** 2026-08-22

**Repository coordinate:** current planning state after the Game UI archival verdict `f199f9a`

**Authority:** research only; owner decision D-003 remains required

## Question

What preservation posture can Cloud Clicker honestly choose now, and which promises require a
supported deliverable first?

This is the source-neutral derivative required by
`planning/platform-alignment/publication-rights-batch-12.md`. The ignored raw dossier remains
private, noncanonical source. This brief does not adopt covenant wording, a notice period, export
scope, retention semantics, a binary promise, an archive destination or sunset choreography.

## Verified external evidence

- The European Citizens' Initiative was submitted on 26 January 2026 after verification of
  1,294,188 statements of support. The organisers presented it to the Commission on 23 February;
  Parliament held a hearing on 16 April and a plenary debate on 21 May. On 16 June the Commission
  declined, at that stage, to propose an EU obligation to keep discontinued games playable and
  committed to pursue an industry end-of-life code of conduct by the end of 2026. Source:
  [official ECI response](https://citizens-initiative.europa.eu/stop-destroying-videogames-commissions-reply-european-citizens-initiative_en).
- Video Games Europe argues that mandatory private-server compatibility can require prohibitive
  engineering work and raises player-safety, security, intellectual-property and reputation risks.
  This is an industry position, not an adopted Cloud Clicker conclusion. Source:
  [Video Games Europe position paper](https://www.videogameseurope.eu/wp-content/uploads/2025/11/VGE-Position-Discontinuation-of-Support-to-Online-Games-04072025.pdf).
- Ubisoft released The Crew 2's offline mode in October 2025, with local progression and an explicit
  list of unavailable online features. Source:
  [Ubisoft's offline-mode notes](https://www.ubisoft.com/en-us/game/the-crew/the-crew-2/news-updates/4kuseh7DXQ6fuf4tFoHix0/the-crew-2-offline-mode-patchnote).
- Knockout City's official Private Server Edition page still offers a free Windows package with
  hosted/LAN/solo play and bot support. Source:
  [official Private Server Edition page](https://www.knockoutcity.com/private-server-edition).

These sources establish that offline or private-server preservation is feasible for some games and
can require substantial product-specific work. They do not establish a universal legal duty, a
universal implementation recipe or a Cloud Clicker release promise.

## Cloud Clicker reality at HEAD

The tracked `deployment-foundation-lifecycle-audit.md` and
`operations-retention-preservation-audit.md` establish the current boundary:

| Concern | Proven now | Missing outcome |
|---|---|---|
| Source availability | The repository is public and MIT-licensed. | Public source alone is not a supported package or covenant. |
| Server primitive | A real Go gameserver binary builds and fails closed without mandatory current credentials. | All tracked Compose files are test-only; there is no production Dockerfile, Caddyfile, full-stack Compose contract or clean-host witness. |
| Runtime content/client | Declarative catalogs, a versioned save schema and a client build exist. | The server reads repository files at runtime, embeds neither required catalogs nor the client, and serves no static client. |
| Empty-world play | Bot fallback is a binding design law and selected bot primitives exist. | No supported self-host workflow proves the Phase-0 preview playable end to end with bots by default. |
| Save continuity | Save migrations and NaN refusal are implemented. | No player export, portable import bundle, provenance contract or clean-host import rehearsal exists. |
| Lifecycle and operations | Health/readiness, bounded drain and credential cleanup are composed. | Backup, restore, rollback, retention, metrics/alerts, incident ownership, preservation mirror and final-world artifact are absent. |

The architecture is preservation-friendly. Preservation is still an engineering and operations
project. No self-host, export or sunset claim is authorized until the exact deliverable is built and
rehearsed.

## D-003 choices now ready for the owner

Choose the public commitment boundary separately from the engineering target:

1. **No covenant yet.** State only the current facts: public MIT source, no supported self-host
   package. Revisit after Deployment and R-006.
2. **Source-continuity covenant only.** Promise a defined source-availability/mirror posture without
   claiming that a runnable bundle, imports or an empty-world experience exist.
3. **Supported-bundle covenant after proof.** Target a documented binary/image plus Postgres/Caddy
   package, import path and bot-default workflow, but make the promise effective only after the
   clean-host, corrupt-backup, rollback and missing/wrong-config controls pass R-006.
4. **Source plus supported bundle after proof.** Combine 2 and 3, again without making the future
   package a current claim.

For any covenant, rule these independently: export/import scope and provenance; notice period;
final-world artifact and privacy boundary; mirror/deposit target; and whether commitments may only
ratchet player-ward. Account retention/deletion and operator obligations remain separately blocked
under O-003/O-004; this research cannot choose them.

## Routing

- D-003 owner ruling chooses among the four postures and the five independent obligations.
- Deployment owns the supported package, origin/proxy/key configuration, client/license delivery,
  release sequence, backup/restore/rollback and clean-host proof.
- R-006 validates only the deliverable actually built; it cannot validate a future promise.
- A later accepted sunset-covenant RFC owns player-facing wording and any final-world/mirror work.
