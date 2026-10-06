import { describe, expect, it } from "vitest";

import population from "../../testdata/reputation/purchase-input-shape-v1.json";
import corpus from "../../testdata/replay/reputation-tree-v1.json";
import corpusBytes from "../../testdata/replay/reputation-tree-v1.json?raw";
import {
  applyFounderLogged, canonicalJSONString, encodeFounderReplayState,
  loadReplayCatalogBundle, restoreFounderReplayState,
  type ReplayArtifacts, type ReplayCatalogBundle,
} from "../src/replay";

const fields = ["kind", "node_id", "resolved_cost", "reputation_level", "reputation_spent_before", "owned_before"];
const mutations = ["missing", "null", "alias", "duplicate", "alias_extra"];
type Control = (typeof population.controls)[number];
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();

function sourceCase(control: Control) {
  const row = corpus.cases.find((candidate) => candidate.name === control.source_case);
  if (!row) throw new Error(`unknown input-shape source ${control.source_case}`);
  const preState = structuredClone(row.pre_state);
  const inputs = { ...structuredClone(row.replay_inputs), resolved: structuredClone(row.replay_inputs.resolved) as unknown as Record<string, unknown> };
  let postStateJSON = row.post_state_json;
  if (control.zero_earned) {
    preState.reputation_level = 0;
    inputs.resolved.reputation_level = 0;
    const post = JSON.parse(postStateJSON) as Record<string, unknown>;
    post.reputation_level = 0;
    postStateJSON = canonicalJSONString(post);
  }
  return { row, preState, inputs, postStateJSON };
}

function expectedRaw(control: Control, field: string, mutation: string): string {
  const resolved = sourceCase(control).inputs.resolved;
  const parts: string[] = [];
  for (const key of fields) {
    const value = JSON.stringify(resolved[key]);
    let entry = `${JSON.stringify(key)}:${value}`;
    if (key === field) {
      switch (mutation) {
        case "missing": continue;
        case "null": entry = `${JSON.stringify(key)}:null`; break;
        case "alias": entry = `${JSON.stringify(key.toUpperCase())}:${value}`; break;
        case "duplicate": parts.push(entry); break;
        case "alias_extra": parts.push(`${JSON.stringify(key.toUpperCase())}:${value}`); break;
        default: throw new Error(`unknown shape mutation ${mutation}`);
      }
    }
    parts.push(entry);
  }
  return `{${parts.join(",")}}`;
}

async function restore(control: Control) {
  const fixture = sourceCase(control);
  const name = fixture.row.bundle as keyof typeof corpus.bundles;
  const source = corpus.bundles[name];
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  const catalogs = await bundles.get(name)!;
  const state = restoreFounderReplayState(fixture.preState, fixture.row.state_version, catalogs);
  return { ...fixture, catalogs, state };
}

async function assertControl(control: Control, rawResolved?: string): Promise<void> {
  const fixture = await restore(control);
  if (rawResolved !== undefined) fixture.inputs.resolved = JSON.parse(rawResolved) as Record<string, unknown>;
  const result = await applyFounderLogged(fixture.state, canonicalJSONString(fixture.row.canonical_payload), fixture.catalogs, fixture.inputs);
  expect(result.outcome).toBe(fixture.row.outcome);
  expect(result.resultConstantsHash).toBe(corpus.bundles[fixture.row.bundle as keyof typeof corpus.bundles].constants_hash);
  expect(canonicalJSONString(result.receipt)).toBe(fixture.row.receipt_json);
  expect(canonicalJSONString(result.events)).toBe(fixture.row.events_json);
  expect(canonicalJSONString(encodeFounderReplayState(result.state))).toBe(fixture.postStateJSON);
}

describe("R5 frozen Reputation purchase-input shape", () => {
  it("binds source bytes, all nineteen controls and the complete six-field mutation product", async () => {
    expect(Object.keys(population).sort()).toEqual(["cases", "controls", "schema_version", "source_sha256"]);
    expect(population.schema_version).toBe(1);
    const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(corpusBytes));
    expect(population.source_sha256).toBe(Array.from(new Uint8Array(digest), (value) => value.toString(16).padStart(2, "0")).join(""));
    const original = corpus.cases.filter((row) => row.replay_inputs.resolved.kind === "purchase_reputation_node");
    expect(original).toHaveLength(18);
    expect(population.controls).toEqual([
      ...original.map((row) => ({ name: row.name, source_case: row.name, zero_earned: false })),
      { name: "zero-earned-unaffordable", source_case: "rejects-cost-one-over-available", zero_earned: true },
    ]);
    expect(population.controls).toHaveLength(19);
    expect(population.cases).toHaveLength(570);
    expect(population.cases.filter((row) => row.mutation === "duplicate")).toHaveLength(114);
    let index = 0;
    for (const control of population.controls) {
      for (const field of fields) {
        for (const mutation of mutations) {
          expect(population.cases[index++]).toEqual({ name: `${control.name}/${field}/${mutation}`, control: control.name, field, mutation,
            raw_resolved: expectedRaw(control, field, mutation) });
        }
      }
    }
    expect(index).toBe(population.cases.length);
  });

  it.each(population.controls)("preserves the complete unmodified $name result", async (control) => {
    await assertControl(control);
  });

  it.each(population.cases.filter((row) => row.mutation !== "duplicate"))("refuses malformed $name without state mutation", async (row) => {
    const control = population.controls.find((candidate) => candidate.name === row.control)!;
    const fixture = await restore(control);
    const before = canonicalJSONString(encodeFounderReplayState(fixture.state));
    fixture.inputs.resolved = JSON.parse(row.raw_resolved) as Record<string, unknown>;
    await expect(applyFounderLogged(fixture.state, canonicalJSONString(fixture.row.canonical_payload), fixture.catalogs, fixture.inputs)).rejects.toThrow();
    expect(canonicalJSONString(encodeFounderReplayState(fixture.state))).toBe(before);
  });

  // This public API receives an object, not JSON bytes. Parsing has already
  // erased duplicate keys: these are honest normalization controls, not refusals.
  it.each(population.cases.filter((row) => row.mutation === "duplicate"))("preserves the normalized duplicate $name result", async (row) => {
    const control = population.controls.find((candidate) => candidate.name === row.control)!;
    expect(JSON.parse(row.raw_resolved)).toEqual(sourceCase(control).inputs.resolved);
    await assertControl(control, row.raw_resolved);
  });
});
