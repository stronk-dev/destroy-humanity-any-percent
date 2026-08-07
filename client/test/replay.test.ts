import fixtureJSON from "../../testdata/replay/apply-logged-v1.json";
import attendanceFixtureJSON from "../../testdata/founder-attendance-v1.json";
import { describe, expect, it } from "vitest";

import {
  applyFounderLogged,
  applyLogged,
  applyLoggedExit,
  canonicalJSONString,
  completedFounderAttendedMS,
  encodeFounderReplayState,
  encodeReplayState,
  encodeReplayStateV14,
  loadReplayCatalogBundle,
  parseFounderAttendanceSample,
  restoreFounderReplayState,
  restoreReplayState,
  verifyReplayRun,
  verifyReplayRunDetailed,
  verifyFounderReplayHistory,
  validateFounderAttendanceSample,
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
	readonly state_version: 14 | 15 | 16 | 17 | 18 | 19 | 20;
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
  readonly doctrine_run: {
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
  readonly active_play_run: {
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
  readonly active_play_exit: {
    readonly constants_hash: string; readonly artifacts: ReplayArtifacts; readonly next_constants_hash: string; readonly next_artifacts: ReplayArtifacts; readonly case: TerminalFixtureCase;
  };
  readonly founder_constants_hash: string;
  readonly founder_artifacts: ReplayArtifacts;
  readonly founder_cases: readonly FounderFixtureCase[];
	readonly pet_founder_constants_hash: string;
	readonly pet_founder_artifacts: ReplayArtifacts;
	readonly pet_founder_cases: readonly FounderFixtureCase[];
	readonly minigame_constants_hash: string;
	readonly minigame_artifacts: ReplayArtifacts;
	readonly minigame_company_case: FixtureCase;
	readonly minigame_founder_case: FounderFixtureCase;
	readonly soul_constants_hash: string;
	readonly soul_artifacts: ReplayArtifacts;
	readonly soul_company_case: FixtureCase;
	readonly soul_founder_case: FounderFixtureCase;
	readonly founder_run: {
		readonly founder_stream_id: string; readonly founder_id: string; readonly genesis_revision: number; readonly genesis_version: 14 | 15 | 16 | 17 | 18;
    readonly genesis_constants_hash: string; readonly genesis: unknown; readonly head_revision: number; readonly head_version: 14 | 15 | 16;
		readonly head_constants_hash: string; readonly head_state: unknown;
    readonly entries: readonly {
      readonly seq: number; readonly intent_id: string; readonly constants_hash: string; readonly canonical_payload: Record<string, unknown>;
      readonly replay_inputs: unknown; readonly receipt_json: string; readonly events_json: string; readonly applied_revision: number | null;
      readonly server_ts_ms: number; readonly source: null | { readonly company_stream_id: string; readonly run_seq: number; readonly run_log_seq: number };
    }[];
  };
};

const attendanceFixture = attendanceFixtureJSON as {
  readonly version: number;
  readonly cases: readonly {
    readonly name: string; readonly founder_age_ms: number; readonly actual_founder_revision: number; readonly expected_founder_revision: number;
    readonly sample: unknown; readonly valid: boolean;
  }[];
};

describe("TypeScript ApplyLogged cross-runtime fixture", () => {
  it.each(attendanceFixture.cases)("validates shared Founder attendance vector $name", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.founder_constants_hash, fixture.founder_artifacts);
    const state = restoreFounderReplayState({ ...(fixture.founder_run.genesis as Record<string, unknown>), age_ms: testCase.founder_age_ms }, fixture.founder_run.genesis_version, bundle);
    expect(completedFounderAttendedMS(state)).toBe(testCase.founder_age_ms);
    const validate = () => validateFounderAttendanceSample(state, testCase.actual_founder_revision, testCase.expected_founder_revision, parseFounderAttendanceSample(testCase.sample));
    if (testCase.valid) expect(validate()).toBe((testCase.sample as { effective_founder_attended_ms: number }).effective_founder_attended_ms);
    else expect(validate).toThrow();
  });

  it.each(fixture.founder_cases)("replays Founder $name to the Go receipt, events, and state", async (testCase) => {
    const bundle = await loadReplayCatalogBundle(fixture.founder_constants_hash, fixture.founder_artifacts);
    const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, bundle);
    const transition = await applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

    expect(transition.outcome).toBe(testCase.outcome);
    expect(transition.resultConstantsHash).toBe(testCase.result_constants_hash);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(testCase.post_state_json);
  });

	it.each(fixture.pet_founder_cases)("replays pet Founder $name to the Go receipt, events, and state", async (testCase) => {
		const bundle = await loadReplayCatalogBundle(fixture.pet_founder_constants_hash, fixture.pet_founder_artifacts);
		const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, bundle);
		const transition = await applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);

		expect(transition.outcome).toBe(testCase.outcome);
		expect(transition.resultConstantsHash).toBe(testCase.result_constants_hash);
		expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
		expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
		expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(testCase.post_state_json);
	});

	it("replays one certified minigame resolution across both Go/TS log arms", async () => {
		const bundle = await loadReplayCatalogBundle(fixture.minigame_constants_hash, fixture.minigame_artifacts);
		if (!bundle.meters || !bundle.achievements) throw new Error("minigame replay fixture lacks foundation catalogs");
		const foundationCatalogs = { meters: bundle.meters, achievements: bundle.achievements };
		const companyCase = fixture.minigame_company_case;
		const company = restoreReplayState(companyCase.pre_state, 16, bundle.economy, foundationCatalogs);
		const companyTransition = await applyLogged(company, canonicalJSONString(companyCase.canonical_payload), bundle, companyCase.replay_inputs);
		expect(canonicalJSONString(companyTransition.receipt)).toBe(companyCase.receipt_json);
		expect(canonicalJSONString(companyTransition.events)).toBe(companyCase.events_json);
		expect(canonicalJSONString(encodeReplayState(companyTransition.state))).toBe(companyCase.post_state_json);
		const founderCase = fixture.minigame_founder_case;
		const founder = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, bundle);
		const founderTransition = await applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), bundle, founderCase.replay_inputs);
		expect(canonicalJSONString(founderTransition.receipt)).toBe(founderCase.receipt_json);
		expect(canonicalJSONString(founderTransition.events)).toBe(founderCase.events_json);
		expect(canonicalJSONString(encodeFounderReplayState(founderTransition.state))).toBe(founderCase.post_state_json);
		const founderWire = founderCase.replay_inputs as { command: { intent_id: string; founder_stream_id: string; founder_id: string; server_ts_ms: number } };
		const companyWire = companyCase.replay_inputs as { command: { company_stream_id: string; run_seq: number; run_log_seq: number } };
		await expect(verifyFounderReplayHistory(founderCase.pre_state, 1, 17, fixture.minigame_constants_hash,
			founderWire.command.founder_stream_id, founderWire.command.founder_id, [{ seq: 1, intentId: founderWire.command.intent_id,
			constantsHash: fixture.minigame_constants_hash, canonicalPayload: canonicalJSONString(founderCase.canonical_payload),
			replayInputs: founderCase.replay_inputs, receiptJSON: founderCase.receipt_json, eventsJSON: founderCase.events_json,
			appliedRevision: 2, serverTSMS: founderWire.command.server_ts_ms, source: { companyStreamId: companyWire.command.company_stream_id,
				runSeq: companyWire.command.run_seq, runLogSeq: companyWire.command.run_log_seq } }],
			{ revision: 2, version: 17, constantsHash: fixture.minigame_constants_hash, state: founderCase.post_state }, [bundle])).resolves.toBe("verified");
	});

	it("replays one Soul recovery across the Company suppression and Founder audit arms", async () => {
		const bundle = await loadReplayCatalogBundle(fixture.soul_constants_hash, fixture.soul_artifacts);
		if (!bundle.meters || !bundle.achievements || !bundle.soul) throw new Error("Soul replay fixture lacks pinned catalogs");
		const companyCase = fixture.soul_company_case;
		const company = restoreReplayState(companyCase.pre_state, 16, bundle.economy,
			{ meters: bundle.meters, achievements: bundle.achievements });
		const companyTransition = await applyLogged(company, canonicalJSONString(companyCase.canonical_payload), bundle, companyCase.replay_inputs);
		expect(canonicalJSONString(companyTransition.receipt)).toBe(companyCase.receipt_json);
		expect(canonicalJSONString(companyTransition.events)).toBe(companyCase.events_json);
		expect(canonicalJSONString(encodeReplayState(companyTransition.state))).toBe(companyCase.post_state_json);
		const founderCase = fixture.soul_founder_case;
		const founder = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, bundle);
		const founderTransition = await applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), bundle, founderCase.replay_inputs);
		expect(canonicalJSONString(founderTransition.receipt)).toBe(founderCase.receipt_json);
		expect(canonicalJSONString(founderTransition.events)).toBe(founderCase.events_json);
		expect(canonicalJSONString(encodeFounderReplayState(founderTransition.state))).toBe(founderCase.post_state_json);
		const founderWire = founderCase.replay_inputs as { command: { intent_id: string; founder_stream_id: string; founder_id: string; server_ts_ms: number } };
		const companyWire = companyCase.replay_inputs as { command: { company_stream_id: string; run_seq: number; run_log_seq: number } };
		await expect(verifyFounderReplayHistory(founderCase.pre_state, 1, 20, fixture.soul_constants_hash,
			founderWire.command.founder_stream_id, founderWire.command.founder_id, [{ seq: 1, intentId: founderWire.command.intent_id,
			constantsHash: fixture.soul_constants_hash, canonicalPayload: canonicalJSONString(founderCase.canonical_payload),
			replayInputs: founderCase.replay_inputs, receiptJSON: founderCase.receipt_json, eventsJSON: founderCase.events_json,
			appliedRevision: 2, serverTSMS: founderWire.command.server_ts_ms, source: { companyStreamId: companyWire.command.company_stream_id,
				runSeq: companyWire.command.run_seq, runLogSeq: companyWire.command.run_log_seq } }],
			{ revision: 2, version: 20, constantsHash: fixture.soul_constants_hash, state: founderCase.post_state }, [bundle])).resolves.toBe("verified");
	});

  it("replays the sequential doctrine and Compute Credit corpus", async () => {
    const run = fixture.doctrine_run;
    const bundle = await loadReplayCatalogBundle(run.constants_hash, run.artifacts);
    if (!bundle.meters || !bundle.achievements || !bundle.doctrines) throw new Error("doctrine fixture lacks its pinned catalogs");
    const state = restoreReplayState(run.genesis, 17, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements, doctrines: bundle.doctrines });
    for (const entry of run.entries) {
      const transition = await applyLogged(state, canonicalJSONString(entry.canonical_payload), bundle, entry.replay_inputs);
      expect(canonicalJSONString(transition.receipt), `receipt seq ${entry.seq}`).toBe(entry.receipt_json);
      expect(canonicalJSONString(transition.events), `events seq ${entry.seq}`).toBe(entry.events_json);
    }
    expect(canonicalJSONString(encodeReplayState(state))).toBe(run.final_state_json);
  });

  it("replays the sequential Active-Play scheduler, buffs, Lucky payout, and rollback corpus", async () => {
    const run = fixture.active_play_run;
    const bundle = await loadReplayCatalogBundle(run.constants_hash, run.artifacts);
    if (!bundle.meters || !bundle.achievements || !bundle.doctrines || !bundle.opportunities) throw new Error("Active-Play fixture lacks its pinned catalogs");
    const state = restoreReplayState(run.genesis, 18, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
      doctrines: bundle.doctrines, opportunities: bundle.opportunities });
    for (const entry of run.entries) {
      const transition = await applyLogged(state, canonicalJSONString(entry.canonical_payload), bundle, entry.replay_inputs);
      expect(canonicalJSONString(transition.receipt), `receipt seq ${entry.seq}`).toBe(entry.receipt_json);
      expect(canonicalJSONString(transition.events), `events seq ${entry.seq}`).toBe(entry.events_json);
    }
    expect(canonicalJSONString(encodeReplayState(state))).toBe(run.final_state_json);
  });

  it("replays Active-Play Exit reset and next-run scheduler initialization", async () => {
    const fixtureExit = fixture.active_play_exit;
    const bundle = await loadReplayCatalogBundle(fixtureExit.constants_hash, fixtureExit.artifacts);
    if (!bundle.meters || !bundle.achievements || !bundle.doctrines || !bundle.opportunities) throw new Error("Active-Play Exit fixture lacks its pinned catalogs");
    const testCase = fixtureExit.case;
    const state = restoreReplayState(testCase.pre_state, 18, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
      doctrines: bundle.doctrines, opportunities: bundle.opportunities });
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(encodeReplayState(transition.finalCompany))).toBe(testCase.final_company_json);
    expect(canonicalJSONString(encodeReplayState(transition.newCompany!))).toBe(testCase.new_company_json);
    expect(transition.newCompany!.activeBuffs).toEqual([]);
    expect(transition.newCompany!.pendingOpportunity).toBeNull();
    expect(transition.newCompany!.opportunitySpawnSeq).toBe(0);
    expect(transition.newCompany!.nextOpportunityAttendedMs).toBeGreaterThan(0);
  });

  it("verifies the Go-authored Founder career from genesis without Company state", async () => {
    const bundle = await loadReplayCatalogBundle(fixture.founder_constants_hash, fixture.founder_artifacts);
    const run = fixture.founder_run;
    const entries = run.entries.map((entry) => ({ seq: entry.seq, intentId: entry.intent_id, constantsHash: entry.constants_hash,
      canonicalPayload: canonicalJSONString(entry.canonical_payload), replayInputs: entry.replay_inputs, receiptJSON: entry.receipt_json, eventsJSON: entry.events_json,
      appliedRevision: entry.applied_revision, serverTSMS: entry.server_ts_ms, source: entry.source === null ? null : { companyStreamId: entry.source.company_stream_id, runSeq: entry.source.run_seq, runLogSeq: entry.source.run_log_seq } }));
    const head = { revision: run.head_revision, version: run.head_version, constantsHash: run.head_constants_hash, state: run.head_state };
    await expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, entries, head, [bundle])).resolves.toBe("verified");

    const poisoned = structuredClone(entries); poisoned[2]!.eventsJSON = "[]";
    await expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, poisoned, head, [bundle])).resolves.toBe("state_divergence");
    const gap = structuredClone(entries); gap[1]!.seq = 7;
    await expect(verifyFounderReplayHistory(run.genesis, run.genesis_revision, run.genesis_version, run.genesis_constants_hash, run.founder_stream_id, run.founder_id, gap, head, [bundle])).resolves.toBe("log_gap");
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

  it("keeps Founder and Company v17 envelopes scope-specific", async () => {
    const activeCase = fixture.additional_bundles.find((value) => value.case.name === "active-foundation-offline-5001ms")!;
    const artifacts: ReplayArtifacts = { ...activeCase.artifacts, minigames: JSON.stringify({ schema_version: 2, rating_seasons: [], minigames: [] }) };
    const bundle = await loadReplayCatalogBundle(await artifactHash(artifacts), artifacts);
    const source = { ...(fixture.founder_run.head_state as Record<string, unknown>), minigame_ratings: {}, minigame_offline_quality: {} };
    const state = restoreFounderReplayState(source, 17, bundle);
    expect(state.wireVersion).toBe(17);
    expect(state.minigameRatings).toEqual({});
    const catalogs = { meters: bundle.meters!, achievements: bundle.achievements!, doctrines: {} as never };
    expect(() => restoreReplayState(source, 17, bundle.economy, catalogs)).toThrow();

    const companySource: Record<string, unknown> = { ...(activeCase.case.pre_state as Record<string, unknown>), compute_burst_remaining_ms: 12_345 };
    const company = restoreReplayState(companySource, 17, bundle.economy, catalogs);
    expect(company.wireVersion).toBe(17);
    expect(company.computeBurstRemainingMs).toBe(12_345);
    expect(encodeReplayState(company)).toMatchObject({ compute_burst_remaining_ms: 12_345 });
    delete companySource.compute_burst_remaining_ms;
    expect(() => restoreReplayState(companySource, 17, bundle.economy, catalogs)).toThrow();
  });

  it("activates the exact Founder v19 Fiscal envelope only under a pinned Fiscal artifact", async () => {
		const artifacts = structuredClone(fixture.pet_founder_artifacts) as unknown as Record<string, string>;
		const economy = JSON.parse(artifacts.economy!) as any;
		economy.multiplier_sources.push(
			{ id: "fiscal.generator.beige_tower", slot: "prestige", target: "generator.beige_tower", provider: "fiscal" },
			{ id: "fiscal.hoard", slot: "prestige", target: "all", provider: "fiscal" },
		);
		artifacts.economy = JSON.stringify(economy);
		artifacts.fiscal = JSON.stringify((await import("../../balance/testdata/fiscal-foundation-v1.json")).default.baseline);
		const bundle = await loadReplayCatalogBundle(await artifactHash(artifacts as unknown as ReplayArtifacts), artifacts as unknown as ReplayArtifacts);
		const source = { ...(fixture.pet_founder_cases[0]!.pre_state as Record<string, unknown>), fiscal_credit: 17,
			fiscal_period_opened_wall_ms: 1_786_000_000_000, fiscal_period_seq: 9,
			fiscal_generator_levels: { "generator.beige_tower": 3 }, fiscal_unlocks: ["unlock.arcade"] };
		const state = restoreFounderReplayState(source, 19, bundle);
		expect(state.fiscalCredit).toBe(17); expect(state.fiscalGeneratorLevels["generator.beige_tower"]).toBe(3);
		expect(encodeFounderReplayState(state)).toMatchObject({ fiscal_credit: 17, fiscal_period_seq: 9, fiscal_unlocks: ["unlock.arcade"] });
		const withoutArtifact = { ...bundle, fiscal: undefined };
		expect(() => restoreFounderReplayState(source, 19, withoutArtifact)).toThrow();
		const missing: Record<string, unknown> = { ...source }; delete missing.fiscal_generator_levels;
		expect(() => restoreFounderReplayState(missing, 19, bundle)).toThrow();
  });

	it("activates Founder v20 only under exact Soul and bumped consumer artifacts", async () => {
		const artifacts = structuredClone(fixture.pet_founder_artifacts) as unknown as Record<string, string>;
		const economy = JSON.parse(artifacts.economy!) as any;
		economy.multiplier_sources.push(
			{ id: "fiscal.generator.beige_tower", slot: "prestige", target: "generator.beige_tower", provider: "fiscal" },
			{ id: "fiscal.hoard", slot: "prestige", target: "all", provider: "fiscal" },
		);
		artifacts.economy = JSON.stringify(economy);
		artifacts.fiscal = JSON.stringify((await import("../../balance/testdata/fiscal-foundation-v1.json")).default.baseline);
		const minigames = JSON.parse(artifacts.minigames!) as any; minigames.schema_version = 3;
		for (const row of minigames.minigames) row.soul_gate = "human_hobby";
		artifacts.minigames = JSON.stringify(minigames);
		const pets = JSON.parse(artifacts.pets!) as any; pets.schema_version = 2;
		for (const row of pets.actions) row.soul_gate = "ordinary";
		artifacts.pets = JSON.stringify(pets);
		artifacts.soul = JSON.stringify({ schema_version: 1, policy: { soul_floor: 0, soul_initial: 100, soul_max: 100,
			recovery_beat_ceiling_ms: 5000, max_session_wall_ms: 86400000 }, bands: [
			{ band_member: "near_zero", min_inclusive: 0, max_inclusive: 9, human_content_locked: true, reason_key: "category.low_percent" },
			{ band_member: "hollow", min_inclusive: 10, max_inclusive: 39, human_content_locked: false, reason_key: "category.ethical_percent" },
			{ band_member: "dimming", min_inclusive: 40, max_inclusive: 74, human_content_locked: false, reason_key: "category.hundred_percent" },
			{ band_member: "whole", min_inclusive: 75, max_inclusive: 100, human_content_locked: false, reason_key: "category.any_percent" },
		], debit_sources: [], recovery_activities: [], ending_policy: { whole_variant: "earnest_ascension", depleted_variant: "training_data" } });
		const bundle = await loadReplayCatalogBundle(await artifactHash(artifacts as unknown as ReplayArtifacts), artifacts as unknown as ReplayArtifacts);
		const source = { ...(fixture.pet_founder_cases[0]!.pre_state as Record<string, unknown>), fiscal_credit: 0,
			fiscal_period_opened_wall_ms: 1_786_000_000_000, fiscal_period_seq: 0,
			fiscal_generator_levels: { "generator.beige_tower": 0 }, fiscal_unlocks: [], soul: 73,
			soul_exhausted_source_ids: [] };
		const state = restoreFounderReplayState(source, 20, bundle);
		expect(state.soul).toBe(73); expect(state.soulExhaustedSourceIds.size).toBe(0);
		expect(encodeFounderReplayState(state)).toMatchObject({ soul: 73, soul_exhausted_source_ids: [] });
		const missing: Record<string, unknown> = { ...source }; delete missing.soul_exhausted_source_ids;
		expect(() => restoreFounderReplayState(missing, 20, bundle)).toThrow();
		expect(() => restoreFounderReplayState(source, 19, { ...bundle, soul: undefined })).toThrow(/fields are not exact|inactive Soul/);
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
    const version = Object.hasOwn(special.case.pre_state as object, "compute_burst_remaining_ms") ? 17 : active ? 16 : 14;
    const state = restoreReplayState(special.case.pre_state, version, bundle.economy, active ? { meters: bundle.meters!, achievements: bundle.achievements! } : undefined);
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
