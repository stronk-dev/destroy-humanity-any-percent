# Epoch 6 — First Content

The fixture-first foundations are now one live, dependency-complete content bundle. This epoch
activates Meters, Achievements, Doctrines and Compute Credit, the Minigame and Pet policies,
Fiscal Quarters, Soul and recovery activities, The Pitch, and the public minigame API catalog.
It also adds `company.permits`, the Legal Department generator, and the Tier-3-to-Tier-4 gate.

The accepted constants identity is
`sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a`.
All source-to-production copies below are byte-identical to the owner-ratified promotion manifest.

| artifact | production path | ratified source | consumed verdict |
|---|---|---|---|
| achievements | `balance/achievements/first-content.json` | `balance/testdata/first-content/achievements-v1.json` | Candidate review `b0277a1` (`{3530b08, d0dccb4, a0ca14e}`) |
| categories | `balance/categories/phase0.json` | `balance/testdata/first-content/categories-v1.json` | FCE-C10 + Candidate review `b0277a1` |
| commons | `balance/commons/phase0.json` | same production artifact | Commons Compact archived verdict |
| doctrines | `balance/doctrines/first-content.json` | `balance/testdata/first-content/doctrines-v1.json` | Doctrine & Compute Credit archived verdict |
| economy | `balance/catalogs/phase0.json` | `balance/testdata/first-content/economy-v3.json` | Candidate review `b0277a1` |
| factions | `balance/factions/phase0.json` | same production artifact | Faction & Incorporation archived verdict |
| fiscal | `balance/fiscal/first-content.json` | `balance/testdata/first-content/fiscal-v1.json` | Candidate review `b0277a1` |
| guilds | `balance/guilds/phase0.json` | same production artifact | Guild Model archived verdict |
| meters | `balance/meters/first-content.json` | `balance/testdata/first-content/meters-v1.json` | Candidate review `b0277a1` |
| minigame_api | `balance/minigame-api/first-content.json` | `balance/testdata/minigame-api-candidate-v1.json` | Minigame API verdict `ce69a4b` |
| minigames | `balance/minigames/first-content.json` | `testdata/minigame/pitch-v3.json` | Minigame Platform verdict |
| pets | `balance/pets/first-content.json` | `balance/testdata/first-content/pets-v2.json` | Candidate review `b0277a1` |
| pitch | `balance/pitch.json` | `balance/testdata/pitch-v1.json` | The Pitch verdict `c76101a` |
| prestige | `balance/prestige/phase0.json` | same production artifact | Prestige & Exits verdict |
| routes | `balance/routes/phase0.json` | `balance/testdata/permits-t3-gate-candidate-v1.json` | Permits verdict `88e2054` |
| soul | `balance/soul/first-content.json` | `balance/testdata/first-content/soul-v1.json` | Candidate review `b0277a1` |

The candidate review independently recomputed every source SHA and loaded the complete bundle in
both runtimes. The composed pacing report then ran 300 deterministic cases. It found no invariant
failure and no Casual-policy movement; seven Chaos-policy observations moved by 33.3–60%. Marco
reviewed those findings and authorized this mint in FCE5.6. Those provisional values stand for
epoch 6; new-content scenarios and retuning belong to the epoch-7 lane.

Activation is new-run-bound. Existing runs keep their pinned catalog and semantics. A founder who
adopts this epoch initializes the shipped Company and Founder feature schemas without retroactive
achievement grants or a fabricated starter pet. Historical accepted hashes remain registered.
