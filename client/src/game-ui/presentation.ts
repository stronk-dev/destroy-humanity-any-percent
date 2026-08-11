import source from "./presentation.generated.json";

import type { CopyKey } from "../copy";

type NamedBinding = Readonly<{ id: string; title_key: CopyKey }>;
type DescribedBinding = NamedBinding & Readonly<{ description_key: CopyKey }>;
type GeneratorBinding = DescribedBinding & Readonly<{ cap_reason_key: CopyKey | null }>;
type CosmeticBinding = DescribedBinding & Readonly<{ disclosure_key: CopyKey; purchasable: false; stateful: false }>;

export interface GameUIPresentation {
  readonly schemaVersion: 2;
  readonly generators: ReadonlyMap<string, GeneratorBinding>;
  readonly upgrades: ReadonlyMap<string, DescribedBinding>;
  readonly manualActions: ReadonlyMap<string, DescribedBinding>;
  readonly cosmeticStubs: ReadonlyMap<string, CosmeticBinding>;
  readonly gates: ReadonlyMap<string, NamedBinding>;
  readonly exitTypes: ReadonlyMap<string, NamedBinding>;
}

function rows<T extends { readonly id: string }>(values: readonly T[], label: string): ReadonlyMap<string, T> {
  const result = new Map<string, T>();
  for (const value of values) {
    if (result.has(value.id)) throw new SyntaxError(`${label} IDs must be unique`);
    result.set(value.id, Object.freeze({ ...value }));
  }
  return result;
}

export const GAME_UI_PRESENTATION: GameUIPresentation = Object.freeze({
  schemaVersion: 2,
  generators: rows(source.generators as GeneratorBinding[], "generator presentation"),
  upgrades: rows(source.upgrades as DescribedBinding[], "upgrade presentation"),
  manualActions: rows(source.manual_actions as DescribedBinding[], "manual presentation"),
  cosmeticStubs: rows(source.cosmetic_stubs as CosmeticBinding[], "cosmetic presentation"),
  gates: rows(source.gates as NamedBinding[], "gate presentation"),
  exitTypes: rows(source.exit_types as NamedBinding[], "exit presentation"),
});

export function requirePresentation<T>(values: ReadonlyMap<string, T>, id: string): T {
  const value = values.get(id);
  if (!value) throw new RangeError(`missing presentation binding for ${id}`);
  return value;
}
