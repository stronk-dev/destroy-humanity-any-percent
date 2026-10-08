import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/pet-adoption-v1.json";
import { applyFounderLogged, canonicalJSONString, encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, verifyFounderReplayHistory, type FounderReplayHead, type FounderReplayLogEntry, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// AC4/AC7: the TS Founder replay byte-matches the Go-authored adoption corpus
// (testdata/replay/pet-adoption-v1.json, regenerated only from Go).
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();
function bundle(name: string): Promise<ReplayCatalogBundle> {
  const source = (corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>)[name]!;
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  return bundles.get(name)!;
}

async function recordedHistory(row: typeof corpus.cases[number]) {
  const catalogs = await bundle(row.bundle);
  const inputs = structuredClone(row.replay_inputs);
  inputs.command.founder_log_seq = 1;
  const entry = { seq: 1, intentId: inputs.command.intent_id, constantsHash: catalogs.constantsHash,
    canonicalPayload: canonicalJSONString(row.canonical_payload), replayInputs: inputs, receiptJSON: row.receipt_json,
    eventsJSON: row.events_json, appliedRevision: row.outcome === "applied" ? inputs.command.revision + 1 : null,
    serverTSMS: inputs.command.server_ts_ms, source: null };
  const head = { revision: inputs.command.revision + (row.outcome === "applied" ? 1 : 0),
    version: row.state_version as FounderReplayHead["version"], constantsHash: catalogs.constantsHash,
    state: JSON.parse(row.post_state_json) as Record<string, unknown> };
  const verify = (entries: readonly FounderReplayLogEntry[] = [entry], target: FounderReplayHead = head,
    available: readonly ReplayCatalogBundle[] = [catalogs]) => verifyFounderReplayHistory(row.pre_state,
    inputs.command.revision, head.version, catalogs.constantsHash, inputs.command.founder_stream_id,
    inputs.command.founder_id, entries, target, available);
  return { entry, head, verify, catalogs };
}

describe("pet adoption cross-runtime corpus", () => {
  it("pins every reachable PA4.2 row, the cap ordering, and care on the adopted pet", () => {
    expect(corpus.cases.map((row) => row.name)).toEqual(expect.arrayContaining(["rejects-inactive-adoption", "rejects-unknown-species",
      "rejects-unknown-name", "applies-starter-adoption", "rejects-second-adoption-at-cap", "rejects-unknown-species-at-cap", "care-applies-on-adopted-pet"]));
  });

  it.each(corpus.cases)("replays $name to the Go receipt, events, and state", async (testCase) => {
    const catalogs = await bundle(testCase.bundle);
    const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, catalogs);
    const transition = await applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), catalogs, testCase.replay_inputs);
    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(testCase.post_state_json);
  });

  it("re-derives draws from the nonce: a tampered nonce cannot reproduce the receipt", async () => {
    const applied = corpus.cases.find((row) => row.name === "applies-starter-adoption")!;
    const catalogs = await bundle(applied.bundle);
    const inputs = structuredClone(applied.replay_inputs) as { resolved: Record<string, unknown> };
    inputs.resolved.adoption_nonce = "deadbeefcafebabe0011223344556677";
    const state = restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs);
    const transition = await applyFounderLogged(state, canonicalJSONString(applied.canonical_payload), catalogs, inputs);
    expect(canonicalJSONString(transition.receipt)).not.toBe(applied.receipt_json);
    const stripped = structuredClone(applied.replay_inputs) as { resolved: Record<string, unknown> };
    delete stripped.resolved.adoption_nonce;
    await expect(applyFounderLogged(restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs), canonicalJSONString(applied.canonical_payload), catalogs, stripped)).rejects.toThrow();
  });

  it.each(corpus.cases)("verifies the pinned complete history for $name", async (row) => {
    // Preserve the source revision and expected outputs; only the local
    // sequence is rebased. This is not a persisted multi-command chronology.
    const history = await recordedHistory(row);
    expect(await history.verify()).toBe("verified");
  });

  it.each(["nonce", "missing-nonce", "receipt-pet", "missing-event", "event-name", "identity-name",
    "care-watermark", "applied-revision", "head-revision", "missing-artifacts"])("refuses corrupted history: %s", async (change) => {
    const history = await recordedHistory(corpus.cases.find((row) => row.name === "applies-starter-adoption")!);
    expect(await history.verify()).toBe("verified");
    const entry = structuredClone(history.entry), head = structuredClone(history.head);
    const receipt = JSON.parse(entry.receiptJSON) as { pet_id: string; name_key: string };
    const differentName = receipt.name_key === "pet.name.server_room_cat.n01" ? "pet.name.server_room_cat.n02" : "pet.name.server_room_cat.n01";
    const resolved = entry.replayInputs.resolved as Record<string, unknown>;
    let expected = "state_divergence";
    switch (change) {
      case "nonce": resolved.adoption_nonce = "deadbeefcafebabe0011223344556677"; break;
      case "missing-nonce": delete resolved.adoption_nonce; break;
      case "receipt-pet": receipt.pet_id = "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb"; entry.receiptJSON = canonicalJSONString(receipt); break;
      case "missing-event": entry.eventsJSON = "[]"; break;
      case "event-name": {
        const events = JSON.parse(entry.eventsJSON) as { kind: string; payload: Record<string, unknown> }[];
        const adopted = events.find((event) => event.kind === "pet_adopted.v1");
        expect(adopted).toBeDefined();
        adopted!.payload.name_key = differentName;
        entry.eventsJSON = canonicalJSONString(events);
        break;
      }
      case "identity-name": (head.state.pet_identities as Record<string, Record<string, unknown>>)[receipt.pet_id]!.name_key = differentName; break;
      case "care-watermark": {
        const care = (head.state.pets as Record<string, Record<string, unknown>>)[receipt.pet_id]!;
        care.evaluated_through_attended_ms = (care.evaluated_through_attended_ms as number) + 1;
        break;
      }
      case "applied-revision": entry.appliedRevision! += 1; expected = "log_gap"; break;
      case "head-revision": head.revision += 1; break;
      case "missing-artifacts": expected = "constants_mismatch"; break;
    }
    expect(await history.verify([entry], head, change === "missing-artifacts" ? [] : undefined)).toBe(expected);
    expect(await history.verify(), "negative control must not mutate the honest history").toBe("verified");
  });

  it("refuses a pin missing only pet_species without falling back to another bundle", async () => {
    const history = await recordedHistory(corpus.cases.find((row) => row.name === "applies-starter-adoption")!);
    const stripped = { ...history.catalogs.artifacts };
    expect(stripped.pet_species).toBeTypeOf("string");
    delete stripped.pet_species;
    expect(Object.keys(stripped)).toHaveLength(Object.keys(history.catalogs.artifacts).length - 1);
    await expect(loadReplayCatalogBundle(history.catalogs.constantsHash, stripped)).rejects.toThrow("replay artifact label mismatch");
    expect(await history.verify([history.entry], history.head, [await bundle("tree")])).toBe("constants_mismatch");
    expect(await history.verify()).toBe("verified");
  });
});
