import fixtureJSON from "../../testdata/replay/apply-logged-v1.json";
import { describe, expect, it } from "vitest";

import {
  applyLogged,
  applyLoggedExit,
  canonicalJSONString,
  encodeReplayStateV12,
  loadReplayCatalogBundle,
  restoreReplayStateV12,
  verifyReplayRun,
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
  readonly receipt_json: string;
  readonly events_json: string;
  readonly post_state_json: string;
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
  readonly receipt_json: string;
  readonly founder_output_json: string;
  readonly final_company_json: string;
  readonly new_company_json: string;
  readonly founder_events_json: string;
  readonly company_ended_events_json: string;
  readonly company_started_events_json: string;
}

const fixture = fixtureJSON as {
  readonly version: number;
  readonly constants_hash: string;
  readonly artifacts: ReplayArtifacts;
  readonly cases: readonly FixtureCase[];
  readonly terminal_cases: readonly TerminalFixtureCase[];
  readonly additional_bundles: readonly {
    readonly constants_hash: string;
    readonly artifacts: ReplayArtifacts;
    readonly case: FixtureCase;
  }[];
  readonly full_run: {
    readonly genesis: unknown;
    readonly entries: readonly {
      readonly seq: number;
      readonly canonical_payload: Record<string, unknown>;
      readonly replay_inputs: unknown;
      readonly receipt_json: string;
      readonly events_json: string;
      readonly terminal: boolean;
    }[];
    readonly final_state_json: string;
  };
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
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeReplayStateV12(transition.state))).toBe(testCase.post_state_json);
  });

  it.each(fixture.additional_bundles)("replays an additional Go-authored catalog bundle", async (special) => {
    const bundle = await loadReplayCatalogBundle(special.constants_hash, special.artifacts);
    const state = restoreReplayStateV12(special.case.pre_state, bundle.economy);
    const transition = await applyLogged(state, canonicalJSONString(special.case.canonical_payload), bundle, special.case.replay_inputs);

    expect(transition.outcome).toBe(special.case.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(special.case.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(special.case.events_json);
    expect(canonicalJSONString(encodeReplayStateV12(transition.state))).toBe(special.case.post_state_json);
    expect(transition.invariants).toEqual([{ kind: "afford_fallback", intent_id: "01986666-0201-7000-8000-000000000201", detail: "generator.beige_tower" }]);
  });

  it.each(fixture.terminal_cases)("replays terminal '$name' to the Go receipt, events, and next run", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const state = restoreReplayStateV12(testCase.pre_state, bundle.economy);
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayStateV12(transition.finalCompany))).toBe(testCase.final_company_json);
    expect(canonicalJSONString(encodeReplayStateV12(transition.newCompany!))).toBe(testCase.new_company_json);
    expect(canonicalJSONString(transition.founderEvents)).toBe(testCase.founder_events_json);
    expect(canonicalJSONString(transition.companyEndedEvents)).toBe(testCase.company_ended_events_json);
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(testCase.company_started_events_json);
  });

  it("verifies the 51-entry mixed full-run corpus and all closed failure verdicts", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const entries = fixture.full_run.entries.map((entry) => ({ seq: entry.seq, canonicalPayload: canonicalJSONString(entry.canonical_payload), replayInputs: entry.replay_inputs, receiptJSON: entry.receipt_json, eventsJSON: entry.events_json, terminal: entry.terminal }));
    const identity = { constantsHash: fixture.constants_hash };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, identity)).resolves.toBe("verified");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries.filter((entry) => entry.seq !== 20), identity)).resolves.toBe("log_gap");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, { constantsHash: `sha256:${"f".repeat(64)}` })).resolves.toBe("constants_mismatch");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, { ...identity, engineMismatch: true })).resolves.toBe("engine_mismatch");

    const driftedClock = structuredClone(entries);
    (driftedClock[10]!.replayInputs as Record<string, unknown>).evaluated_at_ms = 1;
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, driftedClock, identity)).resolves.toBe("clock_violation");
    const tampered = structuredClone(entries);
    tampered[10] = { ...tampered[10]!, receiptJSON: "{}" };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, tampered, identity)).resolves.toBe("state_divergence");
  });

  it("fails closed on a hidden catalog input and clock regression", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const testCase = fixture.cases[0]!;
    const state = restoreReplayStateV12(testCase.pre_state, bundle.economy);
    const replay = structuredClone(testCase.replay_inputs) as Record<string, any>;
    replay.evaluated_at_ms = state.evaluatedThroughMs - 1;
    await expect(applyLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, replay)).rejects.toThrow(/clock regression/);

    const rejectedCase = fixture.cases.find((value) => value.name === "unknown-buy-rejects-before-accrual")!;
    const rejectedState = restoreReplayStateV12(rejectedCase.pre_state, bundle.economy);
    const rejectedReplay = structuredClone(rejectedCase.replay_inputs) as Record<string, any>;
    rejectedReplay.evaluated_at_ms = rejectedState.evaluatedThroughMs - 1;
    await expect(applyLogged(rejectedState, canonicalJSONString(rejectedCase.canonical_payload), bundle, rejectedReplay)).rejects.toThrow(/clock regression/);

    const founderCase = fixture.cases.find((value) => value.name === "cross-gate")!;
    const founderState = restoreReplayStateV12(founderCase.pre_state, bundle.economy);
    const founderReplay = structuredClone(founderCase.replay_inputs) as Record<string, any>;
    founderReplay.resolved.founder_carry.founder_constants_hash = `sha256:${"b".repeat(64)}`;
    await expect(applyLogged(founderState, canonicalJSONString(founderCase.canonical_payload), bundle, founderReplay)).rejects.toThrow(/founder catalog mismatch/);
  });
});
