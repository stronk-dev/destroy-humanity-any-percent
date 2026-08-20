# Dependency graph

This graph separates the release spine from gameplay breadth. A consumer RFC cannot make its
producer, content, migration, refusal, or acceptance witness appear by assertion.

## Release spine

```
R-001 first-wave diagnosis -> accepted CI/harness observability owner
    -> authority-preserving hosted measurement
    -> CI Baseline repair/current-head green

R-002 release audit -> D-001 milestone floor -> D-007 content scope
                                      |              |
                                      |              -> tier/content RFC sequence
                                      v
                         reconciled Deployment RFC
                          /          |          \
                   packaging     backup/restore   safe deploy
                       |                |              |
                       +------ R-006 clean-host proof-+

D-002 repo disposition ------------------------------^
D-003 sunset deliverable -> export/self-host successor^
account default/fallback body reconciliation + D-005/D-008/D-009
    -> account rights + recovery RFC -> R-003
Transport D4/T4 client recovery + Account token rotation
    -> durable Game UI sessions and drain recovery
API AC4/C18 body reconciliation
    -> all-v1 registry/query/header/raw-client authority
    -> generated browser client + public formula/readers/evidence/privacy proof
Game UI body reconciliation -> AC1 implementation -> R-004
accessibility release contract -----------------------> R-005
```

The release is not the Deployment RFC alone. It is the intersection of a current CI verdict,
supported packaging, recoverable data, player-facing account rights, accessible workflows, honest
content scope, and an integrated run.

## Existing gameplay producers and consumers

```
Combat Shared Data
    -> Duel Engine ----\
    -> Lane Engine -----+-> Bots & Integration -> pet/minigame player surfaces

Minigame Platform + Account/API + Game UI
    -> Minigame & Recovery API + Surface
        -> The Pitch / Soul Recovery visible workflows
        -> later individual minigame RFCs

Leaderboards/Epochs + Replay + Public API
    -> board readers/run evidence
        -> speedrun chrome, spectator pages, world-first consumers

Commons/Guilds + World Layer + Feed/Dispatch + Events L1
    -> Layer-2 disasters
    -> Layer-3 world arcs
    -> conspiracy/media/narrative consumers

Purchasable Content + T0-T1 proof
    -> tier 2..8 content epochs
        -> designed endings
```

## Shared-resource ownership rules

- Account/recovery/export schemas belong to one account/data-portability contract; individual UI
  RFCs consume them.
- Release packaging, backup/restore, and sunset artifacts share one versioned release manifest.
- Accessibility acceptance follows user tasks and is consumed by every surface; component-local
  axe checks are a lower-level producer.
- Public board/evidence DTOs are API Foundation resources; spectator/speedrun consumers may not
  invent alternate wire types.
- Each content epoch binds real data to its producer engine and at least one default workflow.
