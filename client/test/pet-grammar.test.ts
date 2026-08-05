import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/grammar-v1.json";
import {
  PET_BEHAVIOR_EVENTS,
  PET_BEHAVIOR_PRNG_LABEL,
  PET_BEHAVIOR_QUEUE_HARDCAP,
  PET_BEHAVIOR_STATES,
  PET_CARE_REJECTION_DETAILS,
  PET_MOODS,
  PET_STAT_IDS,
  PET_STATUS_BANDS,
  validPetBehaviorQueueLength,
} from "../src/pet/grammar";

describe("pet wire grammar", () => {
  it("matches the shared Go/TypeScript fixture", () => {
    expect(fixture.version).toBe(1);
    expect(PET_STAT_IDS).toEqual(fixture.stat_ids);
    expect(PET_STATUS_BANDS).toEqual(fixture.status_bands);
    expect(PET_MOODS).toEqual(fixture.moods);
    expect(PET_BEHAVIOR_STATES).toEqual(fixture.behavior_states);
    expect(PET_BEHAVIOR_EVENTS).toEqual(fixture.behavior_events);
    expect(PET_CARE_REJECTION_DETAILS).toEqual(fixture.care_rejection_details);
    expect(PET_BEHAVIOR_QUEUE_HARDCAP).toBe(fixture.behavior_queue_hardcap);
    expect(PET_BEHAVIOR_PRNG_LABEL).toBe(fixture.behavior_prng_label);
  });

  it("enforces the hard queue bound", () => {
    for (const length of fixture.valid_queue_lengths) expect(validPetBehaviorQueueLength(length), `${length}`).toBe(true);
    for (const length of fixture.invalid_queue_lengths) expect(validPetBehaviorQueueLength(length), `${length}`).toBe(false);
    expect(validPetBehaviorQueueLength(1.5)).toBe(false);
  });

  it("does not expose mutable protocol authority", () => {
    for (const values of [PET_STAT_IDS, PET_STATUS_BANDS, PET_MOODS, PET_BEHAVIOR_STATES, PET_BEHAVIOR_EVENTS, PET_CARE_REJECTION_DETAILS]) {
      expect(Object.isFrozen(values)).toBe(true);
      expect(() => (values as unknown as string[]).push("invented")).toThrow();
    }
  });
});
