import { describe, expect, it } from "vitest";
import content from "../../balance/testdata/typer-v1.json?raw";
import corpusSource from "../../testdata/typer/content-gate-v1.json";
import { COPY_KEYS } from "../src/copy";
import { parseTyperCatalog, typerContentHash } from "../src/typer/catalog";
import { applyTyper, createTyper, encodeSnapshot, firstMismatchIndex, normalizeTyperLine, TyperRejection, typerPromptOrder, type TyperResult } from "../src/typer/engine";

interface Corpus {
  readonly version: number;
  readonly typer_content_hash: string;
  readonly transition_budget: number;
  readonly scenarios: readonly { readonly name: string; readonly seed: string; readonly era_tier: number;
    readonly steps: readonly { readonly command: unknown; readonly server_time_ms: number; readonly expect: string }[];
    readonly expected_terminal: unknown; readonly expected_result: TyperResult; readonly covers_prompts: readonly string[] }[];
}

const corpus = corpusSource as unknown as Corpus;

describe("Typer shared content gate", () => {
  it("byte-replays every Go-generated scenario, rejections included", async () => {
    expect(await typerContentHash(content)).toBe(corpus.typer_content_hash);
    let transitions = 0;
    for (const scenario of corpus.scenarios) {
      const identity = { content, content_hash: corpus.typer_content_hash, content_schema_version: 1, seed: BigInt(scenario.seed), mode: "solo" as const,
        scaling_inputs: { "typer.era_tier": scenario.era_tier } };
      let snapshot = await createTyper(identity);
      let result: TyperResult | null = null;
      let revision = 1;
      for (const step of scenario.steps) {
        transitions++;
        try {
          const output = await applyTyper({ ...identity, revision, snapshot, command: JSON.stringify(step.command), server_time_ms: step.server_time_ms });
          expect("applied", `${scenario.name} ${JSON.stringify(step.command)}`).toBe(step.expect);
          snapshot = output.snapshot; result = output.result; revision++;
        } catch (error) {
          if (!(error instanceof TyperRejection)) throw error;
          expect(error.code, `${scenario.name} ${JSON.stringify(step.command)}`).toBe(step.expect);
        }
      }
      expect(snapshot, scenario.name).toBe(JSON.stringify(scenario.expected_terminal).replace(/[<>&]/gu, (c) => `\\u${c.charCodeAt(0).toString(16).padStart(4, "0")}`));
      expect(JSON.stringify(result), scenario.name).toBe(JSON.stringify(scenario.expected_result));
    }
    expect(transitions).toBe(corpus.transition_budget);
  });

  it("covers every prompt, all three outcomes, and both modes", () => {
    const catalog = parseTyperCatalog(JSON.parse(content), new Set(COPY_KEYS));
    expect([...new Set(corpus.scenarios.flatMap((row) => row.covers_prompts))].sort()).toEqual(catalog.prompts.map((row) => row.prompt_id));
    expect(new Set(corpus.scenarios.map((row) => row.expected_result.outcome))).toEqual(new Set(["completed", "ended_early", "timed_out"]));
    expect(new Set(corpus.scenarios.flatMap((row) => row.steps.map((step) => step.expect)))).toEqual(
      new Set(["applied", "illegal_phase", "invalid_assist_level", "invalid_text", "line_too_long"]));
  });

  it("matches Go on normalization and mismatch offsets", () => {
    expect(normalizeTyperLine("echo ‘hi’ “x” a—b")).toBe("echo 'hi' \"x\" a--b");
    expect(normalizeTyperLine("–")).toBe("–");
    expect(normalizeTyperLine("\uff4c\uff53 -la")).toBe("\uff4c\uff53 -la");
    expect(firstMismatchIndex("LS -la", "ls -la")).toBe(0);
    expect(firstMismatchIndex("ls -l", "ls -la")).toBe(5);
    expect(firstMismatchIndex("ls -lah", "ls -la")).toBe(6);
  });

  it("scores valid U+FFFD as a miss but rejects unpaired surrogate JSON escapes", async () => {
    const identity = { content, content_hash: corpus.typer_content_hash, content_schema_version: 1, seed: 42n, mode: "solo" as const,
      scaling_inputs: { "typer.era_tier": 1 } };
    const ready = await createTyper(identity);
    const begun = await applyTyper({ ...identity, revision: 1, snapshot: ready, command: '{"kind":"begin","assist_level":"untimed"}', server_time_ms: 1 });
    const miss = await applyTyper({ ...identity, revision: 2, snapshot: begun.snapshot,
      command: JSON.stringify({ kind: "submit_line", text: "�" }), server_time_ms: 2 });
    expect(JSON.parse(miss.snapshot).last_submission).toEqual({ first_mismatch_index: 0, outcome: "miss" });
    for (const command of [String.raw`{"kind":"submit_line","text":"\ud800"}`, String.raw`{"kind":"submit_line","text":"\udc00"}`]) {
      await expect(applyTyper({ ...identity, revision: 2, snapshot: begun.snapshot, command, server_time_ms: 2 }))
        .rejects.toMatchObject({ code: "invalid_text" });
    }
  });

  it("rejects malformed snapshot state that the Go tenant validator rejects", async () => {
    const identity = { content, content_hash: corpus.typer_content_hash, content_schema_version: 1, seed: 42n, mode: "solo" as const,
      scaling_inputs: { "typer.era_tier": 1 } };
    const ready = await createTyper(identity);
    const begun = await applyTyper({ ...identity, revision: 1, snapshot: ready,
      command: '{"assist_level":"untimed","kind":"begin"}', server_time_ms: 1 });
    const missesHardcap = parseTyperCatalog(JSON.parse(content), new Set(COPY_KEYS)).policy.misses_hardcap;
    for (const [name, mutate] of [
      ["negative misses", (row: Record<string, any>) => { row.misses = -1; }],
      ["negative clean lines", (row: Record<string, any>) => { row.clean_lines = -1; }],
      ["negative prompt misses", (row: Record<string, any>) => { row.current_prompt_misses = -1; }],
      ["unknown assist level", (row: Record<string, any>) => { row.assist_level = "custom"; }],
      ["malformed submission", (row: Record<string, any>) => { row.last_submission = { outcome: "miss", first_mismatch_index: null }; }],
      ["misses above pinned cap", (row: Record<string, any>) => { row.misses = missesHardcap + 1; }],
    ] as const) {
      const row = JSON.parse(begun.snapshot) as Record<string, any>;
      mutate(row);
      await expect(applyTyper({ ...identity, revision: 2, snapshot: JSON.stringify(row),
        command: '{"kind":"end_run"}', server_time_ms: 2 }), name).rejects.toThrow(SyntaxError);
    }
  });

  it("refuses typing at prompts_total but completes the genuine final prompt", async () => {
    const identity = { content, content_hash: corpus.typer_content_hash, content_schema_version: 1, seed: 42n, mode: "solo" as const,
      scaling_inputs: { "typer.era_tier": 1 } };
    let snapshot = await createTyper(identity);
    let revision = 1;
    const begin = await applyTyper({ ...identity, revision, snapshot, command: '{"assist_level":"untimed","kind":"begin"}', server_time_ms: revision });
    snapshot = begin.snapshot; revision++;
    let current = JSON.parse(snapshot);
    while (current.prompt_index < current.prompts_total - 1) {
      const output = await applyTyper({ ...identity, revision, snapshot,
        command: JSON.stringify({ kind: "submit_line", text: current.current_prompt_text }), server_time_ms: revision });
      snapshot = output.snapshot; revision++;
      current = JSON.parse(snapshot);
    }
    const invalid = { ...current, prompt_index: current.prompts_total, lines_cleared: current.prompts_total };
    await expect(applyTyper({ ...identity, revision, snapshot: JSON.stringify(invalid), command: '{"kind":"end_run"}', server_time_ms: revision }))
      .rejects.toThrow(SyntaxError);
    const completed = await applyTyper({ ...identity, revision, snapshot,
      command: JSON.stringify({ kind: "submit_line", text: current.current_prompt_text }), server_time_ms: revision });
    expect(JSON.parse(completed.snapshot)).toMatchObject({ phase: "terminal", prompt_index: current.prompts_total, lines_cleared: current.prompts_total });
    expect(completed.result).toMatchObject({ outcome: "completed", score_facts: [
      { kind: "typer.assisted", value: 1 }, { kind: "typer.clean_lines", value: current.prompts_total },
      { kind: "typer.elapsed_ms", value: current.prompts_total }, { kind: "typer.lines_cleared", value: current.prompts_total }, { kind: "typer.misses", value: 0 },
    ] });
  });

  it("rejects catalog defects the Go loader rejects", () => {
    const keys = new Set(COPY_KEYS);
    const mutate = (change: (value: Record<string, any>) => void) => { const value = JSON.parse(content); change(value); return value; };
    for (const change of [
      (v: Record<string, any>) => { v.extra = 1; },
      (v: Record<string, any>) => { v.prompts[0].text = "ls é"; },
      (v: Record<string, any>) => { v.prompts[0].text = "ls  -la"; },
      (v: Record<string, any>) => { v.policy.run_length = 99; },
      (v: Record<string, any>) => { v.prompts[0].scene_copy_key = "typer.prompt.missing.scene"; },
    ]) expect(() => parseTyperCatalog(mutate(change), keys)).toThrow();
  });

  it("rejects scaling below the pinned tier-one clamp even with a valid tier-zero era", async () => {
    const value = JSON.parse(content);
    value.eras[0].min_tier = 0;
    const tierZeroContent = JSON.stringify(value);
    expect(() => parseTyperCatalog(value, new Set(COPY_KEYS))).not.toThrow();
    await expect(createTyper({ content: tierZeroContent, content_hash: await typerContentHash(tierZeroContent), content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: { "typer.era_tier": 0 } })).rejects.toThrow("invalid Typer scaling inputs");
  });

  it("orders prompts through the typer.prompts.v1 substream of the run seed", () => {
    const catalog = parseTyperCatalog(JSON.parse(content), new Set(COPY_KEYS));
    const first = corpus.scenarios[0]!;
    const begin = typerPromptOrder(catalog, BigInt(first.seed), 1);
    expect(first.covers_prompts).toEqual(begin.map((row) => row.prompt_id).sort());
  });

  it("escapes snapshot strings exactly like Go's encoder", () => {
    const encoded = encodeSnapshot({ assist_level: null, clean_lines: 0, current_prompt_id: "x", current_prompt_misses: 0, current_prompt_text: "a > b & <c>",
      deadline_server_ms: null, era_tier: 1, last_server_ms: null, last_submission: null, lines_cleared: 0, misses: 0, phase: "typing", prompt_index: 0,
      prompts_total: 8, revision: 1, started_server_ms: null, typer_content_hash: "sha256:0", typer_schema_version: 1 });
    expect(encoded).toContain(String.raw`"current_prompt_text":"a \u003e b \u0026 \u003cc\u003e"`);
  });
});
