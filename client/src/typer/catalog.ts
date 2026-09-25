// Terminal Typer content (rfc/minigame-terminal-typer.md TT3), the TS mirror
// of server/typer/catalog.go. One shared fixture proves both loaders agree.
export const TYPER_SCHEMA_VERSION = 1 as const;
export const TYPER_ERA_TIER_MIN = 1;
export const TYPER_ERA_TIER_MAX = 9;

export interface TyperPolicy {
  readonly run_length: number; readonly timed_budget_ms: number; readonly max_line_bytes: number;
  readonly elapsed_hardcap_ms: number; readonly misses_hardcap: number;
  readonly run_length_reason_key: string; readonly line_bytes_reason_key: string;
  readonly elapsed_reason_key: string; readonly misses_reason_key: string;
}
export interface TyperEra { readonly era_id: string; readonly min_tier: number; readonly copy_key: string }
export interface TyperPrompt { readonly prompt_id: string; readonly era_id: string; readonly text: string; readonly scene_copy_key: string }
export interface TyperCatalog {
  readonly schema_version: 1;
  readonly policy: TyperPolicy;
  readonly eras: readonly TyperEra[];
  readonly prompts: readonly TyperPrompt[];
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const MAX_SAFE = Number.MAX_SAFE_INTEGER;

export function parseTyperCatalog(source: unknown, declaredCopyKeys: ReadonlySet<string>): TyperCatalog {
  if (declaredCopyKeys.size === 0) throw new SyntaxError("Typer copy registry is empty");
  const root = exactObject(source, ["schema_version", "policy", "eras", "prompts"], "Typer catalog");
  if (root.schema_version !== TYPER_SCHEMA_VERSION) throw new SyntaxError("invalid Typer schema version");
  const policyRow = exactObject(root.policy, ["run_length", "timed_budget_ms", "max_line_bytes", "elapsed_hardcap_ms", "misses_hardcap",
    "run_length_reason_key", "line_bytes_reason_key", "elapsed_reason_key", "misses_reason_key"], "Typer policy");
  const policy: TyperPolicy = Object.freeze({
    run_length: safe(policyRow.run_length, 1, MAX_SAFE), timed_budget_ms: safe(policyRow.timed_budget_ms, 1, MAX_SAFE),
    max_line_bytes: safe(policyRow.max_line_bytes, 1, 1024), elapsed_hardcap_ms: safe(policyRow.elapsed_hardcap_ms, 1, MAX_SAFE),
    misses_hardcap: safe(policyRow.misses_hardcap, 1, MAX_SAFE),
    run_length_reason_key: key(policyRow.run_length_reason_key, declaredCopyKeys), line_bytes_reason_key: key(policyRow.line_bytes_reason_key, declaredCopyKeys),
    elapsed_reason_key: key(policyRow.elapsed_reason_key, declaredCopyKeys), misses_reason_key: key(policyRow.misses_reason_key, declaredCopyKeys),
  });
  if (!Array.isArray(root.eras) || root.eras.length === 0 || !Array.isArray(root.prompts) || root.prompts.length === 0) throw new SyntaxError("Typer rows must be non-empty arrays");
  const eraTier = new Map<string, number>();
  let priorTier = -1;
  const eras = Object.freeze(root.eras.map((item) => {
    const row = exactObject(item, ["era_id", "min_tier", "copy_key"], "Typer era");
    const id = identifier(row.era_id), tier = safe(row.min_tier, 0, TYPER_ERA_TIER_MAX);
    if (tier <= priorTier || eraTier.has(id)) throw new SyntaxError("Typer eras must be tier-ascending and unique");
    priorTier = tier; eraTier.set(id, tier);
    return Object.freeze({ era_id: id, min_tier: tier, copy_key: key(row.copy_key, declaredCopyKeys) });
  }));
  let prior = "";
  const prompts = Object.freeze(root.prompts.map((item) => {
    const row = exactObject(item, ["prompt_id", "era_id", "text", "scene_copy_key"], "Typer prompt");
    const id = identifier(row.prompt_id);
    if (prior !== "" && byteCompare(prior, id) >= 0) throw new SyntaxError("Typer prompts must be byte-sorted and unique");
    prior = id;
    if (typeof row.era_id !== "string" || !eraTier.has(row.era_id)) throw new SyntaxError("Typer prompt references an unknown era");
    if (typeof row.text !== "string" || !validPromptText(row.text, policy.max_line_bytes)) throw new SyntaxError("invalid Typer prompt text");
    return Object.freeze({ prompt_id: id, era_id: row.era_id, text: row.text, scene_copy_key: key(row.scene_copy_key, declaredCopyKeys) });
  }));
  const catalog: TyperCatalog = Object.freeze({ schema_version: TYPER_SCHEMA_VERSION, policy, eras, prompts });
  for (let tier = TYPER_ERA_TIER_MIN; tier <= TYPER_ERA_TIER_MAX; tier++) {
    if (typerPool(catalog, tier).length < policy.run_length) throw new SyntaxError("a reachable Typer era_tier cannot deal a full run");
  }
  return catalog;
}

export function typerPool(catalog: TyperCatalog, eraTier: number): TyperPrompt[] {
  const tiers = new Map(catalog.eras.map((era) => [era.era_id, era.min_tier]));
  return catalog.prompts.filter((prompt) => (tiers.get(prompt.era_id) ?? Infinity) <= eraTier);
}

export async function typerContentHash(bytes: string | Uint8Array): Promise<string> {
  const data: Uint8Array<ArrayBuffer> = typeof bytes === "string" ? new TextEncoder().encode(bytes) : new Uint8Array(bytes);
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", data));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function validPromptText(text: string, maxBytes: number): boolean {
  if (text.length === 0 || text.length > maxBytes || text.startsWith(" ") || text.endsWith(" ") || text.includes("  ")) return false;
  for (let index = 0; index < text.length; index++) { const code = text.charCodeAt(index); if (code < 0x20 || code > 0x7e) return false; }
  return true;
}

function key(value: unknown, declared: ReadonlySet<string>): string {
  const result = identifier(value);
  if (!declared.has(result)) throw new SyntaxError(`unknown Typer copy key ${result}`);
  return result;
}
function identifier(value: unknown): string {
  if (typeof value !== "string" || !mechanical.test(value)) throw new SyntaxError("invalid Typer identifier");
  return value;
}
function safe(value: unknown, minimum: number, maximum: number): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError("Typer integer outside exact domain");
  return value;
}
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const row = source as Record<string, unknown>, actual = Object.keys(row).sort(byteCompare), expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((value, index) => value !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return row;
}
export function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
