import { describe, expect, it } from "vitest";
import migrations from "../../testdata/save-migrations.json";
import baseline from "../../testdata/save-migrations-baseline.json";
import source from "../../testdata/axis-stack/activation-research-v1.json";
import sourceRaw from "../../testdata/axis-stack/activation-research-v1.json?raw";
import { applyLogged, applyLoggedExit, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Action { pre_state: unknown; post_state: unknown; canonical_payload: unknown; replay_inputs: unknown;
  receipt_json: string; events_json: string; post_state_json: string }
interface Exit { pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown; receipt_json: string;
  founder_output_json: string; final_company_json: string; new_company_json: string;
  founder_events_json: string; company_ended_events_json: string; company_started_events_json: string }
interface Entry { name: string; source_case: string; source_state: string; operation: string;
  from_version: number; expected_version: number; input_patch: Record<string, unknown>; error_stage: string }
const corpus = source as unknown as { current: { constants_hash: string; artifacts: ReplayArtifacts };
  next: { constants_hash: string; artifacts: ReplayArtifacts }; rows: { id: string; old: Action; exit: Exit; next: Action[] }[] };
const entries = migrations.company_cases as Entry[];
const currentPromise = loadReplayCatalogBundle(corpus.current.constants_hash, corpus.current.artifacts);
const nextPromise = loadReplayCatalogBundle(corpus.next.constants_hash, corpus.next.artifacts);
const names = ["company-v18-to-v19-new-run", "company-v19-derived-mismatch-rejected", "company-v19-field-before-version-rejected",
  "company-v19-attained-not-superset-of-earned-rejected", "pre-activation-run-replays-through-exit"];
const states = ["exit_pre", "post_purchase", "old_post", "post_purchase", "old_pre"];
const operations = ["new_run_activation", "decode", "decode", "decode", "replay_exit"];
const stages = ["none", "pinned_derived", "structural", "pinned_superset", "none"];
const fromVersions = [18, 19, 18, 19, 18], toVersions = [19, 19, 18, 19, 19];
function restore(raw: unknown, version: number, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("migration foundation missing");
  return restoreReplayState(raw, version, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
function admitted(raw: unknown, version: number, bundle: ReplayCatalogBundle) {
  const state = restore(raw, version, bundle);
  expect(state.wireVersion).toBe(version);
  expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(raw));
  return state;
}
async function step(state: ReturnType<typeof restore>, action: Action, bundle: ReplayCatalogBundle, version: number) {
  expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(action.pre_state));
  const result = await applyLogged(state, canonicalJSONString(action.canonical_payload), bundle, action.replay_inputs);
  expect(result.outcome).toBe("applied");
  expect(canonicalJSONString(result.receipt)).toBe(action.receipt_json);
  expect(canonicalJSONString(result.events)).toBe(action.events_json);
  expect(canonicalJSONString(encodeReplayState(result.state))).toBe(action.post_state_json);
  return admitted(encodeReplayState(result.state), version, bundle);
}

describe("CV4 exact Company migration corpus at real load/Exit boundaries", () => {
  it("requires all20 distinct names, five exact boundaries and byte-pinned source", async () => {
    expect(Object.keys(migrations).sort()).toEqual(["cases", "company_cases", "company_source", "corpus_version", "founder_cases", "founder_source"]);
    expect(migrations.corpus_version).toBe(10);
    expect(migrations.cases).toHaveLength(11); expect(migrations.founder_cases).toHaveLength(4); expect(entries).toHaveLength(5);
    expect(baseline.schema_version).toBe(1); expect(baseline.minimum_case_count).toBe(20);
    const allNames = [...migrations.cases, ...migrations.founder_cases, ...entries].map(row => row.name);
    expect(new Set(allNames).size).toBe(20); expect(allNames.sort()).toEqual([...baseline.required_case_names].sort());
    expect(entries.map(row => row.name)).toEqual(names);
    for (const [index, row] of entries.entries()) {
      expect(Object.keys(row).sort()).toEqual(["error_stage", "expected_version", "from_version", "input_patch", "name", "operation", "source_case", "source_state"]);
      expect([row.source_state, row.operation, row.error_stage, row.from_version, row.expected_version])
        .toEqual([states[index], operations[index], stages[index], fromVersions[index], toVersions[index]]);
      expect(row.source_case).toBe("veteran/online/3114");
    }
    expect(migrations.company_source.path).toBe("axis-stack/activation-research-v1.json");
    const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(sourceRaw)));
    expect([...digest].map(byte => byte.toString(16).padStart(2, "0")).join("")).toBe(migrations.company_source.sha256);
  });

  it.each(entries)("executes $name on pinned source, not a name-only migration", async entry => {
    const matches = corpus.rows.filter(row => row.id === entry.source_case);
    expect(matches).toHaveLength(1);
    const row = matches[0]!; expect(row.next).toHaveLength(2);
    const current = await currentPromise, next = await nextPromise;
    const raw = entry.source_state === "exit_pre" ? row.exit.pre_state : entry.source_state === "old_pre" ? row.old.pre_state
      : entry.source_state === "old_post" ? row.old.post_state : entry.source_state === "post_purchase" ? row.next[1]!.post_state : undefined;
    expect(raw).toBeDefined();
    const bundle = entry.source_state === "post_purchase" ? next : current;
    let state = admitted(raw, entry.from_version, bundle);
    if (entry.operation === "decode") {
      const patches: Record<string, Record<string, unknown>> = {
        structural: { achievements_attained_run: [], attainment_score_run: 0 },
        pinned_derived: { attainment_score_run: 3 },
        pinned_superset: { achievements_attained_run: [], attainment_score_run: 0,
          achievements_earned_run: ["achievement.generators_purchased_1"], achievement_score_run: 2 },
      };
      expect(entry.input_patch).toEqual(patches[entry.error_stage]);
      const patched = { ...structuredClone(raw as Record<string, unknown>), ...entry.input_patch };
      const before = canonicalJSONString(patched);
      const errors: Record<string, RegExp> = { structural: /^save v18 fields are not exact$/u,
        pinned_derived: /attainment score does not derive from attained IDs/u,
        pinned_superset: /run-scoped earned achievement achievement.generators_purchased_1 is not attained/u };
      expect(() => restore(patched, entry.from_version, bundle)).toThrow(errors[entry.error_stage]);
      expect(canonicalJSONString(patched)).toBe(before);
      const fixed = structuredClone(patched);
      if (entry.error_stage === "structural") { delete fixed.achievements_attained_run; delete fixed.attainment_score_run; }
      else if (entry.error_stage === "pinned_derived") fixed.attainment_score_run = 2;
      else if (entry.error_stage === "pinned_superset") { fixed.achievements_earned_run = []; fixed.achievement_score_run = 0; }
      else throw new Error("unknown negative stage");
      admitted(fixed, entry.from_version, bundle);
      return;
    }
    expect(entry.input_patch).toEqual({}); expect(entry.error_stage).toBe("none");
    expect(state.wireVersion).toBe(18); expect(state.achievementsAttainedRun).toBeNull(); expect(state.attainmentScoreRun).toBe(0);
    if (entry.operation === "replay_exit") state = await step(state, row.old, current, 18);
    else expect(entry.operation).toBe("new_run_activation");
    expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.exit.pre_state));
    const exit = await applyLoggedExit(state, canonicalJSONString(row.exit.canonical_payload), withNextReplayCatalogBundle(current, next), row.exit.replay_inputs);
    expect(exit.outcome).toBe("applied"); expect(canonicalJSONString(exit.receipt)).toBe(row.exit.receipt_json);
    expect(canonicalJSONString(exit.founder)).toBe(row.exit.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(exit.finalCompany))).toBe(row.exit.final_company_json);
    expect(canonicalJSONString(encodeReplayState(exit.newCompany!))).toBe(row.exit.new_company_json);
    expect(canonicalJSONString(exit.founderEvents)).toBe(row.exit.founder_events_json);
    expect(canonicalJSONString(exit.companyEndedEvents)).toBe(row.exit.company_ended_events_json);
    expect(canonicalJSONString(exit.companyStartedEvents)).toBe(row.exit.company_started_events_json);
    const terminal = admitted(encodeReplayState(exit.finalCompany), 18, current);
    expect(terminal.achievementsAttainedRun).toBeNull(); expect(terminal.attainmentScoreRun).toBe(0);
    state = admitted(encodeReplayState(exit.newCompany!), 19, next);
    expect(state.runSeq).toBe(3); expect(state.achievementsAttainedRun?.size).toBe(0); expect(state.attainmentScoreRun).toBe(0);
    expect(state.achievementsEarnedRun.size).toBe(0); expect(state.achievementScoreRun).toBe(0);
    if (entry.operation === "new_run_activation") {
      for (const action of row.next) state = await step(state, action, next, 19);
      expect([...state.achievementsAttainedRun!]).toEqual(["achievement.generators_purchased_1"]); expect(state.attainmentScoreRun).toBe(2);
    }
  });
});
