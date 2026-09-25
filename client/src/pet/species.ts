import { MAX_EXACT_INTEGER } from "../numeric";

// Pet Adoption v1 (rfc/pet-adoption-v1.md PA2/PA4): the pet_species artifact
// and the grinding-proof, nonce-derived adoption draws. Byte parity with
// server/pet/species.go is proven by the shared fixtures and vectors.

export const PET_TEMPERAMENT_ORDER = ["lazy", "playful", "curious", "sassy", "shy", "chaotic"] as const;
export type PetTemperament = (typeof PET_TEMPERAMENT_ORDER)[number];

export interface PetSpeciesRow {
  readonly species_id: string;
  readonly availability: "starter";
  readonly visual_family: "cat";
  readonly allowed_temperaments: readonly PetTemperament[];
  readonly palette_ids: readonly string[];
  readonly name_keys: readonly string[];
  readonly name_copy_key: string;
  readonly description_copy_key: string;
}

export interface PetSpeciesCatalog {
  readonly schema_version: 1;
  readonly max_pets_per_founder: number;
  readonly species: readonly PetSpeciesRow[];
}

export interface PetSpeciesDeclarations { readonly copyKeys: ReadonlySet<string>; readonly companionKeys: ReadonlySet<string> }

const speciesID = /^pet_species\.[a-z][a-z0-9_]*$/u;
const paletteID = /^pet_palette\.[a-z0-9_]+$/u;
const nameKey = /^pet\.name\.[a-z0-9_]+\.[a-z0-9_]+$/u;
const copyKey = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;

export function parsePetSpeciesCatalog(source: unknown, declarations: PetSpeciesDeclarations): PetSpeciesCatalog {
  const root = exactObject(source, ["schema_version", "max_pets_per_founder", "species"], "pet_species");
  if (root.schema_version !== 1 || !Number.isSafeInteger(root.max_pets_per_founder) || (root.max_pets_per_founder as number) < 1 ||
    (root.max_pets_per_founder as number) > MAX_EXACT_INTEGER || !Array.isArray(root.species) || root.species.length === 0 || root.species.length > 64) {
    throw new SyntaxError("invalid pet_species schema, cap, or species list");
  }
  let starters = 0;
  const species: PetSpeciesRow[] = [];
  for (const item of root.species) {
    const row = exactObject(item, ["species_id", "availability", "visual_family", "allowed_temperaments", "palette_ids", "name_keys", "name_copy_key", "description_copy_key"], "pet_species row");
    if (typeof row.species_id !== "string" || !speciesID.test(row.species_id) || species.length > 0 && byteCompare(species.at(-1)!.species_id, row.species_id) >= 0) {
      throw new SyntaxError("pet species ids must be namespaced, sorted, and unique");
    }
    if (row.availability !== "starter" || row.visual_family !== "cat") throw new SyntaxError("unknown pet availability or visual family");
    starters += 1;
    const temperaments = parseTemperaments(row.allowed_temperaments);
    const palettes = sortedUniqueMatching(row.palette_ids, paletteID, "palette_ids");
    const names = sortedUniqueMatching(row.name_keys, nameKey, "name_keys");
    for (const key of [...names, row.name_copy_key, row.description_copy_key]) {
      if (typeof key !== "string" || !copyKey.test(key)) throw new SyntaxError(`pet copy key ${String(key)} is not mechanical`);
      if (!declarations.copyKeys.has(key)) throw new SyntaxError(`pet copy key ${key} is not registered`);
      if (!declarations.companionKeys.has(key)) throw new SyntaxError(`pet copy key ${key} is not companion tone`);
    }
    species.push(Object.freeze({ species_id: row.species_id, availability: "starter", visual_family: "cat", allowed_temperaments: temperaments,
      palette_ids: palettes, name_keys: names, name_copy_key: row.name_copy_key as string, description_copy_key: row.description_copy_key as string }));
  }
  if (starters !== 1) throw new SyntaxError("exactly one starter pet species is required");
  return Object.freeze({ schema_version: 1, max_pets_per_founder: root.max_pets_per_founder as number, species: Object.freeze(species) });
}

export function petSpeciesRow(catalog: PetSpeciesCatalog, id: string): PetSpeciesRow | undefined {
  return catalog.species.find((row) => row.species_id === id);
}

export interface PetAdoptionDraws { readonly pet_id: string; readonly temperament: PetTemperament; readonly palette_id: string }

// PA4.3–PA4.5. The client-chosen intent_id never enters any hash input.
export async function drawPetAdoption(founderId: string, nonceHex: string, serverTimeMs: number,
  row: Pick<PetSpeciesRow, "allowed_temperaments" | "palette_ids">): Promise<PetAdoptionDraws> {
  const founder = uuidBytes(founderId);
  if (!/^[0-9a-f]{32}$/u.test(nonceHex)) throw new SyntaxError("pet adoption nonce must be 16 lowercase hex bytes");
  if (!Number.isSafeInteger(serverTimeMs) || serverTimeMs < 0 || serverTimeMs > 0xffffffffffff || row.allowed_temperaments.length === 0 || row.palette_ids.length === 0) {
    throw new SyntaxError("invalid pet adoption draw input");
  }
  const nonce = hexBytes(nonceHex);
  const idDigest = await sha256(concat(ascii("pet.id.v1"), founder, nonce));
  const id = new Uint8Array(16);
  let time = BigInt(serverTimeMs);
  for (let index = 5; index >= 0; index -= 1) { id[index] = Number(time & 0xffn); time >>= 8n; }
  id[6] = 0x70 | (idDigest[0]! & 0x0f);
  id[7] = idDigest[1]!;
  id[8] = 0x80 | (idDigest[2]! & 0x3f);
  id.set(idDigest.subarray(3, 10), 9);
  const temperament = row.allowed_temperaments[await labelledIndex("pet.temperament.v1", nonce, row.allowed_temperaments.length)]!;
  const palette = row.palette_ids[await labelledIndex("pet.palette.v1", nonce, row.palette_ids.length)]!;
  return Object.freeze({ pet_id: formatUUID(id), temperament, palette_id: palette });
}

async function labelledIndex(label: string, nonce: Uint8Array, size: number): Promise<number> {
  const digest = await sha256(concat(ascii(label), nonce));
  return Number(new DataView(digest.buffer, digest.byteOffset, 8).getBigUint64(0) % BigInt(size));
}

async function sha256(bytes: Uint8Array): Promise<Uint8Array> {
  return new Uint8Array(await crypto.subtle.digest("SHA-256", bytes as Uint8Array<ArrayBuffer>));
}

function parseTemperaments(source: unknown): readonly PetTemperament[] {
  if (!Array.isArray(source) || source.length === 0 || source.length > PET_TEMPERAMENT_ORDER.length) throw new SyntaxError("allowed_temperaments must be non-empty");
  let cursor = 0;
  for (const value of source) {
    let found = false;
    while (cursor < PET_TEMPERAMENT_ORDER.length) {
      const candidate = PET_TEMPERAMENT_ORDER[cursor]!;
      cursor += 1;
      if (candidate === value) { found = true; break; }
    }
    if (!found) throw new SyntaxError(`temperament ${String(value)} is unknown, duplicated, or out of canonical order`);
  }
  return Object.freeze([...source] as PetTemperament[]);
}

function sortedUniqueMatching(source: unknown, pattern: RegExp, label: string): readonly string[] {
  if (!Array.isArray(source) || source.length === 0 || source.length > 256) throw new SyntaxError(`${label} must be non-empty`);
  source.forEach((value, index) => {
    if (typeof value !== "string" || !pattern.test(value) || index > 0 && byteCompare(source[index - 1] as string, value) >= 0) throw new SyntaxError(`${label} must be sorted, unique, and well-formed`);
  });
  return Object.freeze([...source] as string[]);
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(source).sort(byteCompare), expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} keys are not exact`);
  return source as Record<string, unknown>;
}

function uuidBytes(value: string): Uint8Array {
  if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(value)) throw new SyntaxError("founder id is not a lowercase canonical UUID");
  return hexBytes(value.replaceAll("-", ""));
}

function hexBytes(value: string): Uint8Array {
  const bytes = new Uint8Array(value.length / 2);
  for (let index = 0; index < bytes.length; index += 1) bytes[index] = Number.parseInt(value.slice(index * 2, index * 2 + 2), 16);
  return bytes;
}

function formatUUID(bytes: Uint8Array): string {
  const text = [...bytes].map((value) => value.toString(16).padStart(2, "0")).join("");
  return `${text.slice(0, 8)}-${text.slice(8, 12)}-${text.slice(12, 16)}-${text.slice(16, 20)}-${text.slice(20)}`;
}

function ascii(value: string): Uint8Array { return new TextEncoder().encode(value); }

function concat(...parts: Uint8Array[]): Uint8Array {
  const result = new Uint8Array(parts.reduce((sum, part) => sum + part.length, 0));
  let offset = 0;
  for (const part of parts) { result.set(part, offset); offset += part.length; }
  return result;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index += 1) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
