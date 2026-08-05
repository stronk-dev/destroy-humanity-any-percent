import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/state-v1.json";
import { parsePetCareStates } from "../src/pet/state";

describe("pet replay-owned state", () => {
  it("matches the shared Go/TypeScript fixture", () => {
    expect(fixture.version).toBe(1);
    const states = parsePetCareStates(fixture.pets, fixture.declarations);
    const state = states["01986666-7000-7000-8000-000000000001"]!;
    expect(state.stats_ppm.hunger).toBe(800000);
    expect(state.trust_ppm).toBe(850000);
    expect(state.behavior_state).toBe("active");
    expect(state.behavior_queue.map((entry) => entry.behavior_id)).toEqual(["behavior.nap", "behavior.pounce"]);
    expect(state).not.toHaveProperty("mood");
  });

  it("rejects unknown declarations and incomplete or invented state", () => {
    const base = structuredClone(fixture.pets) as Record<string, Record<string, unknown>>;
    const id = Object.keys(base)[0]!;
    const cases: unknown[] = [];
    const missingStat = structuredClone(base);
    delete (missingStat[id]!.stats_ppm as Record<string, unknown>).hunger;
    cases.push(missingStat);
    const unknownCooldown = structuredClone(base);
    (unknownCooldown[id]!.cooldown_until_attended_ms as Record<string, unknown>)["care.unknown"] = 1;
    cases.push(unknownCooldown);
    const unknownBehavior = structuredClone(base);
    ((unknownBehavior[id]!.behavior_queue as Record<string, unknown>[])[0]!).behavior_id = "behavior.unknown";
    cases.push(unknownBehavior);
    const badRemainder = structuredClone(base);
    badRemainder[id]!.trust_decay_remainder_ppm = 9_007_199_254_740_992;
    cases.push(badRemainder);
    const badCursor = structuredClone(base);
    badCursor[id]!.behavior_prng_cursor = 1;
    cases.push(badCursor);
    const badQueueOrder = structuredClone(base);
    badQueueOrder[id]!.behavior_queue = [{ behavior_id: "behavior.pounce", due_attended_ms: 2 }, { behavior_id: "behavior.nap", due_attended_ms: 1 }];
    cases.push(badQueueOrder);
    const longQueue = structuredClone(base);
    longQueue[id]!.behavior_queue = Array.from({ length: 9 }, () => ({ behavior_id: "behavior.nap", due_attended_ms: 0 }));
    cases.push(longQueue);
    const storedMood = structuredClone(base);
    storedMood[id]!.mood = "neutral";
    cases.push(storedMood);
    for (const value of cases) expect(() => parsePetCareStates(value, fixture.declarations)).toThrow();
    expect(() => parsePetCareStates(base, { ...fixture.declarations, action_ids: ["care.feed", "care.feed"] })).toThrow();
  });
});
