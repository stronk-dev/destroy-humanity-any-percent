import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/care-transition-v1.json";
import { parsePetCatalog } from "../src/pet/catalog";
import { PET_BEHAVIOR_STATES } from "../src/pet/grammar";
import { parsePetCareStates } from "../src/pet/state";
import { applyPetCareTransition } from "../src/pet/transition";

const petID = "01986666-0000-7000-8000-000000000001";

describe("pet care transition", () => {
  it("matches the shared Go/TypeScript state, receipt facts, and FSM vectors", () => {
    const catalog = parsePetCatalog(fixture.catalog);
    for (const testCase of fixture.cases) {
      const state = parsePetCareStates({ [petID]: testCase.state }, {
        action_ids: catalog.actions.map((row) => row.action_id),
        behavior_ids: PET_BEHAVIOR_STATES,
      })[petID]!;
      expect(applyPetCareTransition(state, catalog, testCase.input), testCase.name).toEqual(testCase.expected);
    }
  });
});
