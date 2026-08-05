import fixtureJSON from "../../testdata/replay/apply-logged-v1.json";
import { describe, expect, it } from "vitest";

import {
  applyFounderLogged,
  applyLogged,
  applyLoggedExit,
  canonicalJSONString,
  encodeFounderReplayState,
  encodeReplayState,
  encodeReplayStateV14,
  loadReplayCatalogBundle,
  restoreFounderReplayState,
  restoreReplayState,
  verifyReplayRun,
  verifyReplayRunDetailed,
  verifyFounderReplayHistory,
  withNextReplayCatalogBundle,
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

interface FounderFixtureCase extends FixtureCase {
  readonly state_version: 14 | 15 | 16;
  readonly result_constants_hash: string;
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
    readonly constants_hash: string;
    readonly artifacts: ReplayArtifacts;
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
  readonly rejected_exit_run: {
    readonly genesis: unknown;
    readonly entries: readonly {
      readonly seq: number; readonly canonical_payload: Record<string, unknown>; readonly replay_inputs: unknown;
      readonly receipt_json: string; readonly events_json: string; readonly terminal: boolean;
    }[];
  };
  readonly active_foundation_exit: {
    readonly constants_hash: string; readonly artifacts: ReplayArtifacts; readonly next_constants_hash: string; readonly next_artifacts: ReplayArtifacts; readonly case: TerminalFixtureCase;
  };
  readonly founder_constants_hash: string;
  readonly founder_artifacts: ReplayArtifacts;
  readonly founder_cases: readonly FounderFixtureCase[];
  readonly founder_run: {
    readonly founder_stream_id: string; readonly founder_id: string; readonly genesis_revision: number; readonly genesis_version: 14 | 15 | 16;
    readonly genesis_constants_hash: string; readonly genesis: unknown; readonly head_revision: number; readonly head_version: 14 | 15 | 16;
    readonly head_constants_hash: string; readonly head_state: unknown;
    readonly entries: readonly {
      readonly seq: number; readonly intent_id: string; readonly constants_hash: string; readonly canonical_payload: Record<string, unknown>;
      readonly replay_inputs: unknown; readonly receipt_json: string; readonly events_json: string; readonly applied_revision: number | null;
      readonly server_ts_ms: number; readonly source: null | { readonly company_stream_id: string; readonly run_seq: number; readonly run_log_seq: number };
    }[];
  };
};

describe("TypeScript ApplyLogged cross-runtime fixture", () => {
  it.each(fixture.founder_cases)("replays Founder $name to the Go receipt, events, and state", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.founder_constants_hash, fixture.founder_artifacts);
    const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, bundle);
    const transition = applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(transition.resultConstantsHash).toBe(testCase.result_constants_hash);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(testCase.post_state_json);
  });

  it("verifies the Go-authored Founder career from genesis without Company state", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.founder_constants_hash, fixture.founder_artifacts);
    const run = fixture.founder_run;
    const entries = run.entries.map((entry) => ({ seq: entry.seq, intentId: entry.intent_id, constantsHash: entry.constants_hash,
      canonicalPayload: canonicalJSONString(entry.canonical_payload), replayInputs: entry.replay_inputs, receiptJSON: entry.receipt_json, eventsJSON: entry.events_json,
      appliedRevision: entry.applied_revision, serverTSMS: entry.server_ts_ms, source: entry.source === null ? null : { companyStreamId: entry.source.company_stream_id, runSeq: entry.source.run_seq, runLogSeq: entry.source.run_log_seq } }));
    const head = { revision: run.head_revision, version: run.head_version, constantsHash: run.head_constants_hash, state: run.head_state };
    expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, entries, head, [bundle])).toBe("verified");

    const poisoned = structuredClone(entries); poisoned[2]!.eventsJSON = "[]";
    expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, poisoned, head, [bundle])).toBe("state_divergence");
    const gap = structuredClone(entries); gap[1]!.seq = 7;
    expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, gap, head, [bundle])).toBe("log_gap");
  });

  it("migrates save v13 to the closed v14 purchasable-content shape", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const current = structuredClone(fixture.cases[0]!.pre_state) as Record<string, unknown>;
    const legacy = structuredClone(current);
    for (const field of ["upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"]) delete legacy[field];
    const migrated = restoreReplayState(legacy, 13, bundle.economy);
    expect(canonicalJSONString(encodeReplayStateV14(migrated))).toBe(canonicalJSONString(current));
    for (const field of ["upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm"]) {
      const malformed = structuredClone(current);
      delete malformed[field];
      expect(() => restoreReplayState(malformed, 14, bundle.economy)).toThrow(SyntaxError);
    }
  });

  it("loads the exact seven artifact bytes under their Go constants hash", async () => {
    expect(fixture.version).toBe(1);
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    expect(bundle.constantsHash).toBe(fixture.constants_hash);

    const relabeled = `sha256:${"a".repeat(64)}`;
    await expect(loadReplayCatalogBundle(relabeled, fixture.artifacts)).rejects.toThrow(/label mismatch/);

		const malformedCategories = JSON.parse(fixture.artifacts.categories) as Record<string, unknown>;
		malformedCategories.full_gate_set = [];
		const malformed = { ...fixture.artifacts, categories: JSON.stringify(malformedCategories) };
		await expect(loadReplayCatalogBundle(await artifactHash(malformed), malformed)).rejects.toThrow(/gate set differs from routes/);
  });

  it.each(fixture.cases)("replays $name to the Go receipt, events, and state", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const state = restoreReplayState(testCase.pre_state, 14, bundle.economy);
    const transition = await applyLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeReplayStateV14(transition.state))).toBe(testCase.post_state_json);
  });

  it.each(fixture.additional_bundles)("replays an additional Go-authored catalog bundle", async (special) => {
    const bundle = await loadReplayCatalogBundle(special.constants_hash, special.artifacts);
    const active = Object.hasOwn(special.case.pre_state as object, "meter_values");
    const state = restoreReplayState(special.case.pre_state, active ? 16 : 14, bundle.economy, active ? { meters: bundle.meters!, achievements: bundle.achievements! } : undefined);
    const transition = await applyLogged(state, canonicalJSONString(special.case.canonical_payload), bundle, special.case.replay_inputs);

    expect(transition.outcome).toBe(special.case.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(special.case.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(special.case.events_json);
    expect(canonicalJSONString(encodeReplayState(transition.state))).toBe(special.case.post_state_json);
    if (special.case.name === "buy-generator-max-fallback-invariant") expect(transition.invariants).toEqual([{ kind: "afford_fallback", intent_id: "01986666-0201-7000-8000-000000000201", detail: "generator.beige_tower" }]);
    else if (special.case.name.startsWith("active-foundation-offline-")) expect(transition.events.map((value) => value.kind)).toEqual(["achievement_earned.v1"]);
    else if (special.case.name === "active-foundation-band-crossing") expect(transition.events.map((value) => value.kind)).toEqual(["meter_band_changed.v1", "achievement_earned.v1"]);
    else expect(transition.invariants).toEqual([]);
  });

  it("keeps historical v2 replayable and rejects v2 globally once foundations are active", async () => {
    const legacy = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const legacyCase = fixture.cases.find((value) => value.name === "manual-online")!;
    const legacyInputs = structuredClone(legacyCase.replay_inputs) as Record<string, unknown>;
    legacyInputs.v = 2;
    const legacyState = restoreReplayState(legacyCase.pre_state, 14, legacy.economy);
    await expect(applyLogged(legacyState, canonicalJSONString(legacyCase.canonical_payload), legacy, legacyInputs)).resolves.toMatchObject({ outcome: legacyCase.outcome });

    const activeCase = fixture.additional_bundles.find((value) => value.case.name === "active-foundation-offline-5001ms")!;
    const active = await loadReplayCatalogBundle(activeCase.constants_hash, activeCase.artifacts);
    const activeInputs = structuredClone(activeCase.case.replay_inputs) as Record<string, unknown>;
    activeInputs.v = 2;
    const activeState = restoreReplayState(activeCase.case.pre_state, 16, active.economy, { meters: active.meters!, achievements: active.achievements! });
    await expect(applyLogged(activeState, canonicalJSONString(activeCase.case.canonical_payload), active, activeInputs)).rejects.toThrow(/invalid replay envelope/);

    const missingCarry = structuredClone(activeCase.case.replay_inputs) as Record<string, any>;
    delete missingCarry.resolved.founder_carry;
    await expect(applyLogged(restoreReplayState(activeCase.case.pre_state, 16, active.economy, { meters: active.meters!, achievements: active.achievements! }), canonicalJSONString(activeCase.case.canonical_payload), active, missingCarry)).rejects.toThrow(/missing active Founder carry/);
    missingCarry.v = 3;
    await expect(applyLogged(restoreReplayState(activeCase.case.pre_state, 16, active.economy, { meters: active.meters!, achievements: active.achievements! }), canonicalJSONString(activeCase.case.canonical_payload), active, missingCarry)).resolves.toMatchObject({ outcome: "applied" });

    const activation = fixture.active_foundation_exit;
    const next = await loadReplayCatalogBundle(activation.constants_hash, activation.artifacts);
    const activatingBundle = withNextReplayCatalogBundle(legacy, next);
    const terminal = fixture.terminal_cases.find((value) => value.name === "wind-down-scripted-first")!;
    const activatingInputs = structuredClone(terminal.replay_inputs) as Record<string, any>;
    activatingInputs.v = 2;
    activatingInputs.resolved.next_constants_hash = next.constantsHash;
    delete activatingInputs.resolved.founder_carry.achievements_earned_lifetime;
    delete activatingInputs.resolved.founder_carry.achievement_score_lifetime;
    const activatingState = restoreReplayState(terminal.pre_state, 14, legacy.economy);
    await expect(applyLoggedExit(activatingState, canonicalJSONString(terminal.canonical_payload), activatingBundle, activatingInputs)).rejects.toThrow(/activation requires replay inputs v3/);
  });

  it("derives active Founder carry identity and score from the pinned achievement artifact", async () => {
    const activeCase = fixture.additional_bundles.find((value) => value.case.name === "active-foundation-offline-5001ms")!;
    const bundle = await loadReplayCatalogBundle(activeCase.constants_hash, activeCase.artifacts);
    const state = () => restoreReplayState(activeCase.case.pre_state, 16, bundle.economy, { meters: bundle.meters!, achievements: bundle.achievements! });
    for (const mutate of [
      (carry: Record<string, any>) => { carry.achievements_earned_lifetime = ["achievement.unknown"]; },
      (carry: Record<string, any>) => { carry.achievement_score_lifetime += 1; },
    ]) {
      const replay = structuredClone(activeCase.case.replay_inputs) as Record<string, any>;
      mutate(replay.resolved.founder_carry);
      await expect(applyLogged(state(), canonicalJSONString(activeCase.case.canonical_payload), bundle, replay)).rejects.toThrow();
    }
  });

  it("replays active foundation Exit and next-catalog score retune byte-for-byte", async () => {
    const fixtureExit = fixture.active_foundation_exit;
    const current = await loadReplayCatalogBundle(fixtureExit.constants_hash, fixtureExit.artifacts);
    const next = await loadReplayCatalogBundle(fixtureExit.next_constants_hash, fixtureExit.next_artifacts);
    const bundle = withNextReplayCatalogBundle(current, next);
    const testCase = fixtureExit.case;
    const state = restoreReplayState(testCase.pre_state, 16, bundle.economy, { meters: bundle.meters!, achievements: bundle.achievements! });
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(transition.finalCompany))).toBe(testCase.final_company_json);
    expect(canonicalJSONString(encodeReplayState(transition.newCompany!))).toBe(testCase.new_company_json);
    expect(canonicalJSONString(transition.founderEvents)).toBe(testCase.founder_events_json);
    expect(canonicalJSONString(transition.companyEndedEvents)).toBe(testCase.company_ended_events_json);
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(testCase.company_started_events_json);
    expect(transition.founder.achievement_score_lifetime).toBe(11);
    expect(transition.companyEndedEvents.some((value) => value.kind === "achievement_earned.v1")).toBe(true);
  });

  it("rejects an active Exit whose run set overlaps Founder lifetime ownership", async () => {
    const fixtureExit = fixture.active_foundation_exit;
    const current = await loadReplayCatalogBundle(fixtureExit.constants_hash, fixtureExit.artifacts);
    const next = await loadReplayCatalogBundle(fixtureExit.next_constants_hash, fixtureExit.next_artifacts);
    const bundle = withNextReplayCatalogBundle(current, next);
    const preState = structuredClone(fixtureExit.case.pre_state) as Record<string, any>;
    preState.achievements_earned_run = ["achievement.first_gate"];
    preState.achievement_score_run = 4;
    const state = restoreReplayState(preState, 16, bundle.economy, { meters: bundle.meters!, achievements: bundle.achievements! });
    await expect(applyLoggedExit(state, canonicalJSONString(fixtureExit.case.canonical_payload), bundle, fixtureExit.case.replay_inputs)).rejects.toThrow(/already owned for life/);
  });

  it.each(fixture.terminal_cases)("replays terminal '$name' to the Go receipt, events, and next run", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const state = restoreReplayState(testCase.pre_state, 14, bundle.economy);
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayStateV14(transition.finalCompany))).toBe(testCase.final_company_json);
    expect(canonicalJSONString(encodeReplayStateV14(transition.newCompany!))).toBe(testCase.new_company_json);
    expect(canonicalJSONString(transition.founderEvents)).toBe(testCase.founder_events_json);
    expect(canonicalJSONString(transition.companyEndedEvents)).toBe(testCase.company_ended_events_json);
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(testCase.company_started_events_json);
  });

  it("verifies the 51-entry mixed full-run corpus and all closed failure verdicts", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.full_run.constants_hash, fixture.full_run.artifacts);
    const entries = fixture.full_run.entries.map((entry) => ({ seq: entry.seq, canonicalPayload: canonicalJSONString(entry.canonical_payload), replayInputs: entry.replay_inputs, receiptJSON: entry.receipt_json, eventsJSON: entry.events_json, terminal: entry.terminal }));
    expect(entries.some((entry) => entry.canonicalPayload.includes('"mode":"max"'))).toBe(true);
    expect(entries.some((entry) => entry.eventsJSON.includes('"kind":"invariant_reported"'))).toBe(true);
    expect(entries.some((entry) => entry.eventsJSON.includes('"kind":"exit_offer_expired"'))).toBe(true);
    const identity = { constantsHash: fixture.full_run.constants_hash };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, identity)).resolves.toBe("verified");
    const detailed = await verifyReplayRunDetailed(fixture.full_run.genesis, bundle, entries, identity);
    expect(detailed.verdict).toBe("verified");
    expect(canonicalJSONString(encodeReplayStateV14(detailed.finalState!))).toBe(fixture.full_run.final_state_json);
    const v12Genesis = structuredClone(fixture.full_run.genesis) as Record<string, unknown>;
    delete v12Genesis.generators_purchased_total;
    await expect(verifyReplayRun(v12Genesis, bundle, entries, { ...identity, genesisVersion: 12 })).resolves.toBe("verified");
    const nonzeroV12 = structuredClone(v12Genesis) as Record<string, any>;
    nonzeroV12.generators["generator.beige_tower"] = 37;
    expect(restoreReplayState(nonzeroV12, 12, bundle.economy).generatorPurchasedTotal).toBe(37);
    const saturatedV12 = structuredClone(v12Genesis) as Record<string, any>;
    saturatedV12.generators["generator.beige_tower"] = Number.MAX_SAFE_INTEGER;
    expect(restoreReplayState(saturatedV12, 12, bundle.economy).generatorPurchasedTotal).toBe(Number.MAX_SAFE_INTEGER);
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, { ...identity, genesisVersion: 12 })).resolves.toBe("constants_mismatch");
    await expect(verifyReplayRun(v12Genesis, bundle, entries, { ...identity, genesisVersion: 11 })).resolves.toBe("constants_mismatch");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries.filter((entry) => entry.seq !== 20), identity)).resolves.toBe("log_gap");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, { constantsHash: `sha256:${"f".repeat(64)}` })).resolves.toBe("constants_mismatch");
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, entries, { ...identity, engineMismatch: true })).resolves.toBe("engine_mismatch");

    const driftedClock = structuredClone(entries);
    (driftedClock[10]!.replayInputs as Record<string, unknown>).evaluated_at_ms = 1;
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, driftedClock, identity)).resolves.toBe("clock_violation");
    const tampered = structuredClone(entries);
    tampered[10] = { ...tampered[10]!, receiptJSON: "{}" };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, tampered, identity)).resolves.toBe("state_divergence");
    const corrupt = structuredClone(entries); corrupt[10] = { ...corrupt[10]!, receiptJSON: "not-json" };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, corrupt, identity)).resolves.toBe("state_divergence");
    const eventTamper = structuredClone(entries); eventTamper[1] = { ...eventTamper[1]!, eventsJSON: "[]" };
    await expect(verifyReplayRun(fixture.full_run.genesis, bundle, eventTamper, identity)).resolves.toBe("state_divergence");
  });

  it("continues after a rejected exit and never treats a final rejection as terminal", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const entries = fixture.rejected_exit_run.entries.map((entry) => ({ seq: entry.seq, canonicalPayload: canonicalJSONString(entry.canonical_payload), replayInputs: entry.replay_inputs, receiptJSON: entry.receipt_json, eventsJSON: entry.events_json, terminal: entry.terminal }));
    const identity = { constantsHash: fixture.constants_hash };
    await expect(verifyReplayRun(fixture.rejected_exit_run.genesis, bundle, entries, identity)).resolves.toBe("log_gap");
    const tampered = structuredClone(entries); tampered[1] = { ...tampered[1]!, receiptJSON: "{}" };
    await expect(verifyReplayRun(fixture.rejected_exit_run.genesis, bundle, tampered, identity)).resolves.toBe("state_divergence");
    await expect(verifyReplayRun(fixture.rejected_exit_run.genesis, bundle, entries.slice(0, 1), identity)).resolves.toBe("log_gap");
    const clock = structuredClone(entries); (clock[0]!.replayInputs as Record<string, unknown>).evaluated_at_ms = 1;
    await expect(verifyReplayRun(fixture.rejected_exit_run.genesis, bundle, clock, identity)).resolves.toBe("clock_violation");
  });

  it("fails closed on a hidden catalog input and clock regression", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts);
    const testCase = fixture.cases[0]!;
    const state = restoreReplayState(testCase.pre_state, 14, bundle.economy);
    const replay = structuredClone(testCase.replay_inputs) as Record<string, any>;
    replay.evaluated_at_ms = state.evaluatedThroughMs - 1;
    await expect(applyLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, replay)).rejects.toThrow(/replay clock violation/);

    const rejectedCase = fixture.cases.find((value) => value.name === "unknown-buy-rejects-before-accrual")!;
    const rejectedState = restoreReplayState(rejectedCase.pre_state, 14, bundle.economy);
    const rejectedReplay = structuredClone(rejectedCase.replay_inputs) as Record<string, any>;
    rejectedReplay.evaluated_at_ms = rejectedState.evaluatedThroughMs - 1;
    await expect(applyLogged(rejectedState, canonicalJSONString(rejectedCase.canonical_payload), bundle, rejectedReplay)).rejects.toThrow(/replay clock violation/);

    const founderCase = fixture.cases.find((value) => value.name === "cross-gate")!;
    const founderState = restoreReplayState(founderCase.pre_state, 14, bundle.economy);
    const founderReplay = structuredClone(founderCase.replay_inputs) as Record<string, any>;
    founderReplay.resolved.founder_carry.founder_constants_hash = `sha256:${"b".repeat(64)}`;
    await expect(applyLogged(founderState, canonicalJSONString(founderCase.canonical_payload), bundle, founderReplay)).rejects.toThrow(/founder catalog mismatch/);
  });
});

async function artifactHash(artifacts: ReplayArtifacts): Promise<string> {
	const encoder = new TextEncoder(); const chunks: Uint8Array[] = [];
	for (const name of Object.keys(artifacts).sort() as (keyof ReplayArtifacts)[]) {
		const nameBytes = encoder.encode(name); const data = encoder.encode(artifacts[name]);
		chunks.push(frame(nameBytes.length), nameBytes, frame(data.length), data);
	}
	const input = new Uint8Array(chunks.reduce((sum, value) => sum + value.length, 0)); let offset = 0;
	for (const chunk of chunks) { input.set(chunk, offset); offset += chunk.length; }
	const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", input));
	return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function frame(value: number): Uint8Array { const result = new Uint8Array(8); new DataView(result.buffer).setBigUint64(0, BigInt(value)); return result; }
