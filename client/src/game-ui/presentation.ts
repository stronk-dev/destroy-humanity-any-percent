import source from "./presentation.generated.json";

import type { CopyKey } from "../copy";

type NamedBinding = Readonly<{ id: string; title_key: CopyKey }>;
type DescribedBinding = NamedBinding & Readonly<{ description_key: CopyKey }>;
type GeneratorBinding = DescribedBinding & Readonly<{ cap_reason_key: CopyKey | null }>;
type CosmeticBinding = DescribedBinding & Readonly<{ disclosure_key: CopyKey; purchasable: false; stateful: false }>;
type ConstantBinding = Readonly<{ id: string; value: string }>;
type TextBinding = Readonly<{ id: string; text_key: CopyKey }>;
const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

function exact(value: object, keys: readonly string[], label: string): void {
  const actual = Object.keys(value).sort(), expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
}

export interface GameUIPresentation {
  readonly schemaVersion: 3;
  readonly cloutReachNotes: ReadonlyMap<string, TextBinding>;
  readonly constants: ReadonlyMap<string, string>;
  readonly generators: ReadonlyMap<string, GeneratorBinding>;
  readonly upgrades: ReadonlyMap<string, DescribedBinding>;
  readonly manualActions: ReadonlyMap<string, DescribedBinding>;
  readonly cosmeticStubs: ReadonlyMap<string, CosmeticBinding>;
  readonly gates: ReadonlyMap<string, NamedBinding>;
  readonly exitTypes: ReadonlyMap<string, NamedBinding>;
  readonly networkSlots: ReadonlyMap<string, NamedBinding>;
}

function rows<T extends { readonly id: string }>(values: readonly T[], keys: readonly string[], label: string): ReadonlyMap<string, T> {
  const result = new Map<string, T>();
  for (const [index, value] of values.entries()) {
    exact(value, keys, label);
    if (!mechanicalID.test(value.id) || index > 0 && values[index - 1].id >= value.id) throw new SyntaxError(`${label} IDs must be byte-sorted and unique`);
    result.set(value.id, Object.freeze({ ...value }));
  }
  return result;
}

function constants(values: readonly ConstantBinding[]): ReadonlyMap<string, string> {
  const result = new Map<string, string>();
  for (const [index, value] of values.entries()) {
    exact(value, ["id", "value"], "presentation constant");
    if (!mechanicalID.test(value.id) || index > 0 && values[index - 1].id >= value.id || typeof value.value !== "string" || value.value.length === 0) throw new SyntaxError("presentation constants must be byte-sorted unique non-empty strings");
    result.set(value.id, value.value);
  }
  return result;
}

exact(source, ["clout_reach_notes", "constants", "cosmetic_stubs", "exit_types", "gates", "generators", "manual_actions", "network_slots", "schema_version", "upgrades"], "game UI presentation");
if (source.schema_version !== 3) throw new SyntaxError("game UI presentation must be schema v3");

export const GAME_UI_PRESENTATION: GameUIPresentation = Object.freeze({
  schemaVersion: 3,
  cloutReachNotes: rows(source.clout_reach_notes as TextBinding[], ["id", "text_key"], "Clout-reach presentation"),
  constants: constants(source.constants as ConstantBinding[]),
  generators: rows(source.generators as GeneratorBinding[], ["cap_reason_key", "description_key", "id", "title_key"], "generator presentation"),
  upgrades: rows(source.upgrades as DescribedBinding[], ["description_key", "id", "title_key"], "upgrade presentation"),
  manualActions: rows(source.manual_actions as DescribedBinding[], ["description_key", "id", "title_key"], "manual presentation"),
  cosmeticStubs: rows(source.cosmetic_stubs as CosmeticBinding[], ["description_key", "disclosure_key", "id", "purchasable", "stateful", "title_key"], "cosmetic presentation"),
  gates: rows(source.gates as NamedBinding[], ["id", "title_key"], "gate presentation"),
  exitTypes: rows(source.exit_types as NamedBinding[], ["id", "title_key"], "exit presentation"),
  networkSlots: rows(source.network_slots as NamedBinding[], ["id", "title_key"], "network-slot presentation"),
});

export function requirePresentation<T>(values: ReadonlyMap<string, T>, id: string): T {
  const value = values.get(id);
  if (!value) throw new RangeError(`missing presentation binding for ${id}`);
  return value;
}

export function requirePresentationConstant(id: string): string {
  return requirePresentation(GAME_UI_PRESENTATION.constants, id);
}
