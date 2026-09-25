import { substream } from "../combat/rng";
import { COPY_KEYS } from "../copy";
import { parseTyperCatalog, typerContentHash, typerPool, TYPER_ERA_TIER_MAX, TYPER_SCHEMA_VERSION, type TyperCatalog, type TyperPrompt } from "./catalog";

// Terminal Typer pure engine (TT4), the TS mirror of server/typer/engine.go.
export const TYPER_ENGINE_VERSION = "1.0.0" as const;
export const TYPER_SCALING_DESTINATION = "typer.era_tier" as const;
const RUN_SUBSTREAM = "typer.run.v1";
const PROMPT_SUBSTREAM = "typer.prompts.v1";

export type TyperPhase = "ready" | "typing" | "terminal";
export type TyperAssist = "timed" | "untimed";
export type TyperCommand = Readonly<{ kind: "begin"; assist_level: TyperAssist }> | Readonly<{ kind: "submit_line"; text: string }> | Readonly<{ kind: "end_run" }>;
export interface TyperSubmission { readonly outcome: "cleared" | "miss"; readonly first_mismatch_index: number | null }
export interface TyperSnapshot {
  readonly assist_level: TyperAssist | null; readonly clean_lines: number; readonly current_prompt_id: string | null;
  readonly current_prompt_misses: number; readonly current_prompt_text: string | null; readonly deadline_server_ms: number | null;
  readonly era_tier: number; readonly last_server_ms: number | null; readonly last_submission: TyperSubmission | null;
  readonly lines_cleared: number; readonly misses: number; readonly phase: TyperPhase; readonly prompt_index: number;
  readonly prompts_total: number; readonly revision: number; readonly started_server_ms: number | null;
  readonly typer_content_hash: string; readonly typer_schema_version: number;
}
export interface TyperResult {
  readonly outcome: "completed" | "ended_early" | "timed_out"; readonly rating_delta: null;
  readonly score_facts: readonly { readonly kind: string; readonly value: number }[];
}
export type TyperRejectionCode = "illegal_phase" | "invalid_assist_level" | "invalid_text" | "line_too_long";
export class TyperRejection extends Error { constructor(readonly code: TyperRejectionCode, readonly detail: string) { super(`${code}: ${detail}`); } }

export interface TyperContentInput { readonly content: string; readonly content_hash: string; readonly content_schema_version: number }
export interface TyperCreateInput extends TyperContentInput { readonly seed: bigint; readonly mode: "solo"; readonly scaling_inputs: Readonly<Record<string, number>> }
export interface TyperApplyInput extends TyperCreateInput { readonly revision: number; readonly snapshot: string; readonly command: string; readonly server_time_ms: number }

type Mutable = { -readonly [K in keyof TyperSnapshot]: TyperSnapshot[K] };

export async function createTyper(input: TyperCreateInput): Promise<string> {
  const catalog = await resolveCatalog(input);
  if (input.mode !== "solo") throw new SyntaxError("Typer is solo only");
  const eraTier = validScaling(input.scaling_inputs, catalog);
  return encodeSnapshot({ assist_level: null, clean_lines: 0, current_prompt_id: null, current_prompt_misses: 0, current_prompt_text: null,
    deadline_server_ms: null, era_tier: eraTier, last_server_ms: null, last_submission: null, lines_cleared: 0, misses: 0, phase: "ready",
    prompt_index: 0, prompts_total: catalog.policy.run_length, revision: 1, started_server_ms: null,
    typer_content_hash: input.content_hash, typer_schema_version: input.content_schema_version });
}

export async function applyTyper(input: TyperApplyInput): Promise<{ readonly snapshot: string; readonly result: TyperResult | null }> {
  const catalog = await resolveCatalog(input);
  const eraTier = validScaling(input.scaling_inputs, catalog);
  const snapshot = decodeSnapshot(input.snapshot);
  if (snapshot.revision !== input.revision || snapshot.typer_content_hash !== input.content_hash || snapshot.typer_schema_version !== input.content_schema_version ||
    snapshot.era_tier !== eraTier) throw new SyntaxError("Typer snapshot diverges from its identity");
  validateAgainstCatalog(snapshot, catalog, input.seed);
  const command = decodeCommand(input.command);
  const result = transition(snapshot, command, catalog, input.seed, input.server_time_ms);
  snapshot.revision = input.revision + 1;
  return { snapshot: encodeSnapshot(snapshot), result };
}

function transition(snapshot: Mutable, command: TyperCommand, catalog: TyperCatalog, seed: bigint, serverTimeMS: number): TyperResult | null {
  if (command.kind === "begin") {
    if (snapshot.phase !== "ready") throw new TyperRejection("illegal_phase", "begin requires ready phase");
    const t = advanceClock(snapshot, serverTimeMS);
    snapshot.assist_level = command.assist_level;
    snapshot.started_server_ms = t;
    if (command.assist_level === "timed") snapshot.deadline_server_ms = t + catalog.policy.timed_budget_ms;
    snapshot.phase = "typing";
    setCurrent(snapshot, typerPromptOrder(catalog, seed, snapshot.era_tier)[0]!);
    return null;
  }
  if (command.kind === "submit_line") {
    if (snapshot.phase !== "typing") throw new TyperRejection("illegal_phase", "submit_line requires typing phase");
    if (new TextEncoder().encode(command.text).length > catalog.policy.max_line_bytes) throw new TyperRejection("line_too_long", "submitted line exceeds max_line_bytes");
    const t = advanceClock(snapshot, serverTimeMS);
    if (snapshot.deadline_server_ms !== null && t > snapshot.deadline_server_ms) return terminate(snapshot, catalog, "timed_out");
    const normalized = normalizeTyperLine(command.text), target = snapshot.current_prompt_text!;
    if (normalized === target) {
      snapshot.lines_cleared++;
      if (snapshot.current_prompt_misses === 0) snapshot.clean_lines++;
      snapshot.prompt_index++;
      snapshot.current_prompt_misses = 0;
      snapshot.last_submission = { outcome: "cleared", first_mismatch_index: null };
      if (snapshot.prompt_index === catalog.policy.run_length) return terminate(snapshot, catalog, "completed");
      setCurrent(snapshot, typerPromptOrder(catalog, seed, snapshot.era_tier)[snapshot.prompt_index]!);
      return null;
    }
    const index = firstMismatchIndex(normalized, target);
    snapshot.misses = Math.min(snapshot.misses + 1, catalog.policy.misses_hardcap);
    snapshot.current_prompt_misses = Math.min(snapshot.current_prompt_misses + 1, catalog.policy.misses_hardcap);
    snapshot.last_submission = { outcome: "miss", first_mismatch_index: index };
    return null;
  }
  if (snapshot.phase !== "ready" && snapshot.phase !== "typing") throw new TyperRejection("illegal_phase", "end_run requires a non-terminal phase");
  advanceClock(snapshot, serverTimeMS);
  return terminate(snapshot, catalog, "ended_early");
}

function advanceClock(snapshot: Mutable, sample: number): number {
  const t = snapshot.last_server_ms !== null && snapshot.last_server_ms > sample ? snapshot.last_server_ms : sample;
  snapshot.last_server_ms = t;
  return t;
}

function setCurrent(snapshot: Mutable, prompt: TyperPrompt): void { snapshot.current_prompt_id = prompt.prompt_id; snapshot.current_prompt_text = prompt.text; }

function terminate(snapshot: Mutable, catalog: TyperCatalog, outcome: TyperResult["outcome"]): TyperResult {
  snapshot.phase = "terminal"; snapshot.current_prompt_id = null; snapshot.current_prompt_text = null;
  let elapsed = 0;
  if (snapshot.started_server_ms !== null && snapshot.last_server_ms !== null) {
    const end = snapshot.deadline_server_ms !== null && snapshot.deadline_server_ms < snapshot.last_server_ms ? snapshot.deadline_server_ms : snapshot.last_server_ms;
    elapsed = Math.min(end - snapshot.started_server_ms, catalog.policy.elapsed_hardcap_ms);
  }
  return { outcome, rating_delta: null, score_facts: [
    { kind: "typer.assisted", value: snapshot.assist_level === "untimed" ? 1 : 0 },
    { kind: "typer.clean_lines", value: snapshot.clean_lines },
    { kind: "typer.elapsed_ms", value: elapsed },
    { kind: "typer.lines_cleared", value: snapshot.lines_cleared },
    { kind: "typer.misses", value: snapshot.misses },
  ] };
}

const NORMALIZATION: readonly [string, string][] = [["‘", "'"], ["’", "'"], ["“", "\""], ["”", "\""], [" ", " "], ["—", "--"]];

/** OD-7's closed smart-punctuation map and nothing else. */
export function normalizeTyperLine(text: string): string {
  let result = "";
  for (const character of text) result += NORMALIZATION.find(([from]) => from === character)?.[1] ?? character;
  return result;
}

/** First differing UTF-8 byte offset, or the shorter length on a prefix. */
export function firstMismatchIndex(submitted: string, target: string): number {
  const left = new TextEncoder().encode(submitted), right = new TextEncoder().encode(target);
  const limit = Math.min(left.length, right.length);
  for (let index = 0; index < limit; index++) if (left[index] !== right[index]) return index;
  return limit;
}

/** TT4.2: downward Fisher–Yates with the typer.prompts.v1 substream of the run seed. */
export function typerPromptOrder(catalog: TyperCatalog, seed: bigint, eraTier: number): TyperPrompt[] {
  const pool = typerPool(catalog, eraTier);
  const runSeed = substream(seed, RUN_SUBSTREAM).next();
  const random = substream(runSeed, PROMPT_SUBSTREAM);
  for (let index = pool.length - 1; index > 0; index--) {
    const swap = Number(random.bound(BigInt(index + 1)));
    [pool[index], pool[swap]] = [pool[swap]!, pool[index]!];
  }
  return pool.slice(0, catalog.policy.run_length);
}

function decodeCommand(source: string): TyperCommand {
  let value: unknown;
  try { value = JSON.parse(source); } catch { throw new TyperRejection("illegal_phase", "command is not JSON"); }
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new TyperRejection("illegal_phase", "command is not an object");
  const row = value as Record<string, unknown>, keys = Object.keys(row).sort().join("\0");
  if (row.kind === "begin") {
    if (keys !== "assist_level\0kind" || row.assist_level !== "timed" && row.assist_level !== "untimed") throw new TyperRejection("invalid_assist_level", "begin requires assist_level timed or untimed");
    return { kind: "begin", assist_level: row.assist_level };
  }
  if (row.kind === "submit_line") {
    if (keys !== "kind\0text" || typeof row.text !== "string") throw new TyperRejection("invalid_text", "submit_line schema mismatch");
    if (!validSubmittedText(row.text)) throw new TyperRejection("invalid_text", "submitted line contains invalid characters");
    return { kind: "submit_line", text: row.text };
  }
  if (row.kind === "end_run") {
    if (keys !== "kind") throw new TyperRejection("illegal_phase", "end_run schema mismatch");
    return { kind: "end_run" };
  }
  throw new TyperRejection("illegal_phase", "unknown command kind");
}

// TT4.4 step 1. U+FFFD is refused in both runtimes because Go's JSON decoder
// substitutes it for invalid UTF-8; lone surrogates are invalid UTF-16 here.
function validSubmittedText(text: string): boolean {
  for (let index = 0; index < text.length; index++) {
    const code = text.charCodeAt(index);
    if (code < 0x20 || code === 0x7f || code === 0xfffd) return false;
    if (code >= 0xd800 && code <= 0xdbff) { const next = text.charCodeAt(index + 1); if (!(next >= 0xdc00 && next <= 0xdfff)) return false; index++; continue; }
    if (code >= 0xdc00 && code <= 0xdfff) return false;
  }
  return true;
}

function decodeSnapshot(source: string): Mutable {
  const value = JSON.parse(source) as Mutable;
  const keys = ["assist_level", "clean_lines", "current_prompt_id", "current_prompt_misses", "current_prompt_text", "deadline_server_ms", "era_tier",
    "last_server_ms", "last_submission", "lines_cleared", "misses", "phase", "prompt_index", "prompts_total", "revision", "started_server_ms",
    "typer_content_hash", "typer_schema_version"];
  if (Object.keys(value).sort().join("\0") !== keys.join("\0") || value.typer_schema_version !== TYPER_SCHEMA_VERSION ||
    !/^sha256:[0-9a-f]{64}$/.test(value.typer_content_hash) || !["ready", "typing", "terminal"].includes(value.phase) ||
    value.prompt_index < 0 || value.prompt_index > value.prompts_total || value.lines_cleared !== value.prompt_index ||
    value.clean_lines > value.lines_cleared || (value.phase === "typing") !== (value.current_prompt_id !== null) ||
    (value.assist_level === null) !== (value.started_server_ms === null) ||
    (value.deadline_server_ms !== null) !== (value.assist_level === "timed")) throw new SyntaxError("invalid Typer snapshot");
  return value;
}

function validateAgainstCatalog(snapshot: Mutable, catalog: TyperCatalog, seed: bigint): void {
  if (snapshot.prompts_total !== catalog.policy.run_length) throw new SyntaxError("Typer snapshot/catalog divergence");
  if (snapshot.phase === "typing") {
    const expected = typerPromptOrder(catalog, seed, snapshot.era_tier)[snapshot.prompt_index];
    if (!expected || expected.prompt_id !== snapshot.current_prompt_id || expected.text !== snapshot.current_prompt_text) throw new SyntaxError("Typer snapshot prompt divergence");
  }
}

/** Go-canonical JSON: sorted keys and Go's <, >, & escaping. */
export function encodeSnapshot(snapshot: TyperSnapshot): string {
  const sorted = (value: unknown): unknown => value === null || typeof value !== "object" ? value :
    Object.fromEntries(Object.keys(value as object).sort().map((key) => [key, sorted((value as Record<string, unknown>)[key])]));
  return JSON.stringify(sorted(snapshot)).replace(new RegExp("[<>&\\u2028\\u2029]", "gu"), (character) => `\\u${character.charCodeAt(0).toString(16).padStart(4, "0")}`);
}

async function resolveCatalog(input: TyperContentInput): Promise<TyperCatalog> {
  if (input.content_schema_version !== TYPER_SCHEMA_VERSION || input.content_hash !== await typerContentHash(input.content)) throw new SyntaxError("Typer content identity mismatch");
  return parseTyperCatalog(JSON.parse(input.content), new Set(COPY_KEYS));
}

function validScaling(values: Readonly<Record<string, number>>, catalog: TyperCatalog): number {
  const keys = Object.keys(values), tier = values[TYPER_SCALING_DESTINATION];
  if (keys.length !== 1 || tier === undefined || !Number.isSafeInteger(tier) || tier < 0 || tier > TYPER_ERA_TIER_MAX || typerPool(catalog, tier).length < catalog.policy.run_length) {
    throw new SyntaxError("invalid Typer scaling inputs");
  }
  return tier;
}
