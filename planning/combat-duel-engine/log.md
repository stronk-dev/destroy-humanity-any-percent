# Combat Duel Engine implementation log

## 2026-08-10 — Codex acceptance review

D4 usefully hardens commitment, exhaustion, stamina saturation, and snapshot keys, but the active
parent still ships arithmetic/RNG only: there is no combat catalog, move-effect union, fixture, or
Obedience/Soul policy table. The duel draft also lacks exact input, action/hash, resolution-order,
event, error, and fuzz bytes. Implementing it now would require inventing core combat rules.

DU-C1–DU-C8 record the dependency and proposed executable contracts. Status remains draft; no duel
engine code was started.
