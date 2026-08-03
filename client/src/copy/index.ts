import generatedCatalog from "./generated/catalog.json";

import { parseCanonical } from "../numeric";
import { COPY_HASH, COPY_MAX_TEXT_LINES, COPY_MAX_TEXT_UTF8_BYTES, type CopyKey, type CopyParamsByKey } from "./generated/types";

export { COPY_HASH, COPY_KEYS, COPY_MAX_TEXT_LINES, COPY_MAX_TEXT_UTF8_BYTES, type CopyKey, type CopyParamsByKey } from "./generated/types";

export type CopyEra = "era_1995" | "era_2000";
export type CopyTone = "corporate" | "diegetic" | "lore_card" | "achievement";
export type CopyParamType = "string" | "integer" | "canonical_decimal";

export interface CopyParamDefinition {
  readonly name: string;
  readonly type: CopyParamType;
}

export interface CopyEntry {
  readonly key: string;
  readonly text: string;
  readonly params: readonly CopyParamDefinition[];
  readonly eraVariants: Readonly<Partial<Record<CopyEra, string>>> | null;
  readonly provenance: readonly string[];
  readonly tone: CopyTone;
}

export interface CopyCatalog {
  readonly schemaVersion: 1;
  readonly entries: readonly CopyEntry[];
  readonly byKey: ReadonlyMap<string, CopyEntry>;
}

export interface CopyInvariant {
  readonly kind: "copy_resolution_failed";
  readonly key: string;
  readonly detail: string;
}

export type ResolveCopyOptions =
  | { readonly mode?: "development"; readonly reportInvariant?: (invariant: CopyInvariant) => void }
  | { readonly mode: "production"; readonly reportInvariant: (invariant: CopyInvariant) => void };

const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const paramName = /^[a-z][a-z0-9_]*$/;
const tones = new Set<CopyTone>(["achievement", "corporate", "diegetic", "lore_card"]);
const paramTypes = new Set<CopyParamType>(["canonical_decimal", "integer", "string"]);
const eras = new Set<CopyEra>(["era_1995", "era_2000"]);

function syntax(message: string): never {
  throw new SyntaxError(message);
}

function exactObject(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) syntax(`${label} must be an object`);
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) {
    syntax(`${label} expected exact keys ${expected.join(", ")}; got ${actual.join(", ")}`);
  }
  return value as Record<string, unknown>;
}

function plainText(value: unknown, label: string): string {
  if (typeof value !== "string" || value !== value.normalize("NFC") || new TextEncoder().encode(value).length > COPY_MAX_TEXT_UTF8_BYTES || value.split("\n").length > COPY_MAX_TEXT_LINES) {
    syntax(`${label} is not bounded NFC text`);
  }
  if (/[^\P{Cc}\n]/u.test(value) || /[<`*_~\[\]|\\]|-->| {2,}$|^ {4}|^\s{0,3}(?:#{1,6}(?:\s|$)|>|[-+]\s|\d+[.)](?:\s|$))|^\s{0,3}(?:-\s*){3,}$|^\s{0,3}===+\s*$/mu.test(value)) syntax(`${label} must be plain text`);
  return value;
}

function placeholderNames(text: string, label: string): readonly string[] {
  const names = new Set<string>();
  for (let index = 0; index < text.length; ) {
    if (text.startsWith("{{", index) || text.startsWith("}}", index)) { index += 2; continue; }
    if (text[index] === "{") {
      const end = text.indexOf("}", index + 1);
      if (end === -1) syntax(`${label} has an unclosed placeholder`);
      const name = text.slice(index + 1, end);
      if (!paramName.test(name)) syntax(`${label} has invalid placeholder {${name}}`);
      names.add(name);
      index = end + 1;
      continue;
    }
    if (text[index] === "}") syntax(`${label} has an unmatched closing brace`);
    index += 1;
  }
  return [...names].sort();
}

function parseEntry(value: unknown, label: string): CopyEntry {
  const row = exactObject(value, ["era_variants", "key", "params", "provenance", "text", "tone"], label);
  if (typeof row.key !== "string" || !mechanicalID.test(row.key)) syntax(`${label}.key is invalid`);
  if (typeof row.tone !== "string" || !tones.has(row.tone as CopyTone)) syntax(`${label}.tone is invalid`);
  if (!Array.isArray(row.params)) syntax(`${label}.params must be an array`);
  const params = row.params.map((raw, index) => {
    const param = exactObject(raw, ["name", "type"], `${label}.params[${index}]`);
    if (typeof param.name !== "string" || !paramName.test(param.name) || typeof param.type !== "string" || !paramTypes.has(param.type as CopyParamType)) syntax(`${label}.params[${index}] is invalid`);
    return Object.freeze({ name: param.name, type: param.type as CopyParamType });
  });
  if (params.some((param, index) => index > 0 && params[index - 1].name >= param.name)) syntax(`${label}.params must be sorted and unique`);
  const text = plainText(row.text, `${label}.text`);
  const expected = params.map((param) => param.name);
  if (placeholderNames(text, `${label}.text`).join("\0") !== expected.join("\0")) syntax(`${label}.text placeholders differ from params`);
  let eraVariants: CopyEntry["eraVariants"] = null;
  if (row.era_variants !== null) {
    if (row.era_variants === null || typeof row.era_variants !== "object" || Array.isArray(row.era_variants)) syntax(`${label}.era_variants is invalid`);
    const source = row.era_variants as Record<string, unknown>;
    const keys = Object.keys(source);
    if (keys.length === 0 || keys.some((key) => !eras.has(key as CopyEra)) || keys.some((key, index) => index > 0 && keys[index - 1] >= key)) syntax(`${label}.era_variants is invalid`);
    const parsed: Partial<Record<CopyEra, string>> = {};
    for (const key of keys as CopyEra[]) {
      const variant = plainText(source[key], `${label}.era_variants.${key}`);
      if (placeholderNames(variant, `${label}.era_variants.${key}`).join("\0") !== expected.join("\0")) syntax(`${label}.era_variants.${key} placeholders differ from params`);
      parsed[key] = variant;
    }
    eraVariants = Object.freeze(parsed);
  }
  if (!Array.isArray(row.provenance) || row.provenance.some((claim) => typeof claim !== "string") || row.provenance.some((claim, index) => index > 0 && (row.provenance as string[])[index - 1] >= claim)) syntax(`${label}.provenance must be sorted unique strings`);
  return Object.freeze({ key: row.key, text, params: Object.freeze(params), eraVariants, provenance: Object.freeze([...(row.provenance as string[])]), tone: row.tone as CopyTone });
}

export function loadCopyCatalog(bytes: string | Uint8Array): CopyCatalog {
  const source = typeof bytes === "string" ? bytes : new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  let value: unknown;
  try { value = JSON.parse(source); } catch (error) { syntax(`copy catalog is invalid JSON: ${String(error)}`); }
  const root = exactObject(value, ["entries", "schema_version"], "copy catalog");
  if (root.schema_version !== 1 || !Array.isArray(root.entries) || root.entries.length === 0) syntax("copy catalog must be non-empty schema version 1");
  const entries = root.entries.map((entry, index) => parseEntry(entry, `copy catalog.entries[${index}]`));
  if (entries.some((entry, index) => index > 0 && entries[index - 1].key >= entry.key)) syntax("copy catalog keys must be byte-sorted and unique");
  return Object.freeze({ schemaVersion: 1, entries: Object.freeze(entries), byKey: new Map(entries.map((entry) => [entry.key, entry])) });
}

export async function hashCopyCatalog(bytes: string | Uint8Array): Promise<string> {
  const encoded: Uint8Array<ArrayBuffer> = typeof bytes === "string" ? new TextEncoder().encode(bytes) : new Uint8Array(bytes);
  const digest = await crypto.subtle.digest("SHA-256", encoded);
  return `sha256:${[...new Uint8Array(digest)].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

export async function verifyCopyCatalogHash(bytes: string | Uint8Array, expectedHash: string): Promise<boolean> {
  return await hashCopyCatalog(bytes) === expectedHash;
}

function render(template: string, entry: CopyEntry, params: Readonly<Record<string, unknown>>): string {
  let result = "";
  for (let index = 0; index < template.length; ) {
    if (template.startsWith("{{", index)) { result += "{"; index += 2; continue; }
    if (template.startsWith("}}", index)) { result += "}"; index += 2; continue; }
    if (template[index] === "{") {
      const end = template.indexOf("}", index + 1);
      const name = template.slice(index + 1, end);
      const definition = entry.params.find((param) => param.name === name)!;
      const value = params[name];
      if (definition.type === "integer" && (!Number.isSafeInteger(value))) throw new TypeError(`${entry.key}.${name} must be a safe integer`);
      if (definition.type === "string" && typeof value !== "string") throw new TypeError(`${entry.key}.${name} must be a string`);
      if (definition.type === "canonical_decimal") {
        if (typeof value !== "string") throw new TypeError(`${entry.key}.${name} must be a canonical Decimal string`);
        parseCanonical(value);
      }
      result += String(value);
      index = end + 1;
      continue;
    }
    result += template[index];
    index += 1;
  }
  return result;
}

export function resolveCopy<K extends CopyKey>(catalog: CopyCatalog, key: K, params: CopyParamsByKey[K], era?: CopyEra, options?: ResolveCopyOptions): string;
export function resolveCopy(catalog: CopyCatalog, key: string, params: Readonly<Record<string, unknown>>, era?: CopyEra, options?: ResolveCopyOptions): string;
export function resolveCopy(catalog: CopyCatalog, key: string, params: Readonly<Record<string, unknown>>, era?: CopyEra, options: ResolveCopyOptions = {}): string {
  try {
    const entry = catalog.byKey.get(key);
    if (!entry) throw new RangeError(`unknown copy key ${key}`);
    const actual = Object.keys(params).sort();
    const expected = entry.params.map((param) => param.name);
    if (actual.join("\0") !== expected.join("\0")) throw new TypeError(`${key} requires exact params ${expected.join(", ")}; got ${actual.join(", ")}`);
    return render(era ? entry.eraVariants?.[era] ?? entry.text : entry.text, entry, params);
  } catch (error) {
    if ((options.mode ?? "development") !== "production") throw error;
    if (typeof options.reportInvariant !== "function") throw new Error("production copy resolution requires an invariant reporter", { cause: error });
    options.reportInvariant({ kind: "copy_resolution_failed", key, detail: error instanceof Error ? error.message : String(error) });
    return key;
  }
}

const generatedBytes = `${JSON.stringify(generatedCatalog, null, 2)}\n`;
export const applicationCopyCatalog = loadCopyCatalog(generatedBytes);

export function t<K extends CopyKey>(key: K, params: CopyParamsByKey[K], era?: CopyEra, options?: ResolveCopyOptions): string {
  return resolveCopy(applicationCopyCatalog, key, params, era, options);
}
