import fixtureJSON from "../../testdata/replay/apply-logged-v1.json";
import { describe, expect, it } from "vitest";

import {
  applyLogged,
  applyLoggedExit,
  canonicalJSONString,
  encodeReplayStateV12,
  loadReplayCatalogBundle,
  restoreReplayStateV12,
  type ReplayArtifacts,
} from "../src/replay";

interface FixtureCase {
  readonly name: string;
  readonly pre_state: unknown;
  readonly canonical_payload: Record<string, unknown>;
  readonly replay_inputs: unknown;
  readonly outcome: "applied" | "rejected";
  readonly receipt: unknown;
  readonly events: readonly unknown[];
  readonly post_state: unknown;
}

interface TerminalFixtureCase {
  readonly name: string;
  readonly pre_state: unknown;
  readonly canonical_payload: Record<string, unknown>;
  readonly replay_inputs: unknown;
  readonly outcome: "applied" | "rejected";
  readonly receipt: unknown;
  readonly founder_output: unknown;
  readonly final_company: unknown;
  readonly new_company: unknown;
  readonly founder_events: readonly unknown[];
  readonly company_ended_events: readonly unknown[];
  readonly company_started_events: readonly unknown[];
}

const fixture = fixtureJSON as {
  readonly version: number;
  readonly constants_hash: string;
  readonly artifacts: ReplayArtifacts;
  readonly cases: readonly FixtureCase[];
  readonly terminal_cases: readonly TerminalFixtureCase[];
};

describe("TypeScript ApplyLogged cross-runtime fixture", () => {
  it("loads the exact six artifact bytes under their Go constants hash", async () => {
    expect(fixture.version).toBe(1);
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    expect(bundle.constantsHash).toBe(fixture.constants_hash);

    const relabeled = `sha256:${"a".repeat(64)}`;
    await expect(loadReplayCatalogBundle(relabeled, fixture.artifacts)).rejects.toThrow(/label mismatch/);
  });

  it.each(fixture.cases)("replays $name to the Go receipt, events, and state", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const state = restoreReplayStateV12(testCase.pre_state, bundle.economy);
    const transition = await applyLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(canonicalJSONString(testCase.receipt));
    expect(canonicalJSONString(transition.events)).toBe(canonicalJSONString(testCase.events));
    expect(canonicalJSONString(encodeReplayStateV12(transition.state))).toBe(canonicalJSONString(testCase.post_state));
  });

  it.each(fixture.terminal_cases)("replays terminal '$name' to the Go receipt, events, and next run", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const state = restoreReplayStateV12(testCase.pre_state, bundle.economy);
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(canonicalJSONString(testCase.receipt));
    expect(canonicalJSONString(transition.founder)).toBe(canonicalJSONString(testCase.founder_output));
    expect(canonicalJSONString(encodeReplayStateV12(transition.finalCompany))).toBe(canonicalJSONString(testCase.final_company));
    expect(canonicalJSONString(encodeReplayStateV12(transition.newCompany!))).toBe(canonicalJSONString(testCase.new_company));
    expect(canonicalJSONString(transition.founderEvents)).toBe(canonicalJSONString(testCase.founder_events));
    expect(canonicalJSONString(transition.companyEndedEvents)).toBe(canonicalJSONString(testCase.company_ended_events));
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(canonicalJSONString(testCase.company_started_events));
  });

  it("fails closed on a hidden catalog input and clock regression", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const testCase = fixture.cases[0]!;
    const state = restoreReplayStateV12(testCase.pre_state, bundle.economy);
    const replay = structuredClone(testCase.replay_inputs) as Record<string, any>;
    replay.evaluated_at_ms = state.evaluatedThroughMs - 1;
    await expect(applyLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, replay)).rejects.toThrow(/clock regression/);

    const founderCase = fixture.cases.find((value) => value.name === "cross-gate")!;
    const founderState = restoreReplayStateV12(founderCase.pre_state, bundle.economy);
    const founderReplay = structuredClone(founderCase.replay_inputs) as Record<string, any>;
    founderReplay.resolved.founder_carry.founder_constants_hash = `sha256:${"b".repeat(64)}`;
    await expect(applyLogged(founderState, canonicalJSONString(founderCase.canonical_payload), bundle, founderReplay)).rejects.toThrow(/founder catalog mismatch/);
  });
});
