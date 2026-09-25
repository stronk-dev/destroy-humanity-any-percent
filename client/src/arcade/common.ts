import { COPY_KEYS } from "../copy";
import { arcadeContentHash, parseArcadeCatalog, ARCADE_SCHEMA_VERSION, type ArcadeCatalog } from "./catalog";

export interface ArcadeContentInput { readonly content: string; readonly content_hash: string; readonly content_schema_version: number }
export interface ArcadeCreateInput extends ArcadeContentInput { readonly seed: bigint; readonly mode: "solo"; readonly scaling_inputs: Readonly<Record<string, number>> }
export interface ArcadeApplyInput extends ArcadeCreateInput { readonly revision: number; readonly snapshot: string; readonly command: string }
export interface ArcadeResult { readonly outcome: string; readonly rating_delta: null; readonly score_facts: readonly { readonly kind: string; readonly value: number }[] }

export class ArcadeRejection extends Error { constructor(readonly code: string, readonly detail: string) { super(`${code}: ${detail}`); } }

export async function resolveArcadeCatalog(input: ArcadeContentInput): Promise<ArcadeCatalog> {
  if (input.content_schema_version !== ARCADE_SCHEMA_VERSION || input.content_hash !== await arcadeContentHash(input.content)) throw new SyntaxError("arcade content identity mismatch");
  return parseArcadeCatalog(JSON.parse(input.content), new Set(COPY_KEYS));
}

/** AR2's only scaling input: the literal breadth 1 at the toy's destination. */
export function requireLiteralScaling(values: Readonly<Record<string, number>>, destination: string): void {
  const keys = Object.keys(values);
  if (keys.length !== 1 || values[destination] !== 1) throw new SyntaxError("invalid arcade scaling inputs");
}

/** Go-canonical JSON: recursively sorted keys and Go's <, >, & escaping. */
export function encodeCanonical(value: unknown): string {
  const sorted = (item: unknown): unknown => item === null || typeof item !== "object" ? item :
    Array.isArray(item) ? item.map(sorted) :
    Object.fromEntries(Object.keys(item as object).sort().map((key) => [key, sorted((item as Record<string, unknown>)[key])]));
  return JSON.stringify(sorted(value)).replace(new RegExp("[<>&\\u2028\\u2029]", "gu"), (character) => `\\u${character.charCodeAt(0).toString(16).padStart(4, "0")}`);
}

export function parseCommandObject(source: string, onInvalid: () => never): Record<string, unknown> {
  let value: unknown;
  try { value = JSON.parse(source); } catch { onInvalid(); }
  if (value === null || typeof value !== "object" || Array.isArray(value)) onInvalid();
  return value as Record<string, unknown>;
}

export function keysOf(row: Record<string, unknown>): string { return Object.keys(row).sort().join("\0"); }

export function safeInteger(value: unknown): value is number { return typeof value === "number" && Number.isSafeInteger(value); }
