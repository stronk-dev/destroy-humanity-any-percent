import source from "./features-presentation.json";

import { applicationCopyCatalog, type CopyKey } from "../copy";

// Mechanical ID -> copy key rows for the GS1/GS3 surfaces. Rows live in data,
// never derived from IDs; an unknown ID is withheld by the surface.
export interface TrustMeterRow { readonly meter_id: string; readonly constituency_key: CopyKey; readonly axis_key: CopyKey }
export interface FeaturesPresentation {
  readonly doomMeter: Readonly<{ meter_id: string; title_key: CopyKey; tooltip_key: CopyKey }>;
  readonly trustMeters: ReadonlyMap<string, TrustMeterRow>;
  readonly meterBands: ReadonlyMap<string, CopyKey>;
  readonly fiscalUnlocks: ReadonlyMap<string, Readonly<{ title_key: CopyKey; description_key: CopyKey }>>;
  // GS5: active-play effect rows; an unknown effect row withholds its opportunity.
  readonly opportunityEffects: ReadonlyMap<string, Readonly<{ title_key: CopyKey; description_key: CopyKey }>>;
  // GS4: care actions in catalog order and the PA7 status bands.
  readonly petActions: ReadonlyMap<string, CopyKey>;
  readonly petBands: ReadonlyMap<string, CopyKey>;
}

function key(value: unknown): CopyKey {
  if (typeof value !== "string" || !applicationCopyCatalog.byKey.has(value)) throw new SyntaxError(`features presentation references unknown copy ${String(value)}`);
  return value as CopyKey;
}

function id(value: unknown): string {
  if (typeof value !== "string" || !/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value)) throw new SyntaxError("features presentation ID is invalid");
  return value;
}

function exact(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(value).sort();
  if (actual.join("\0") !== [...keys].sort().join("\0")) throw new SyntaxError(`${label} fields are not exact`);
  return value as Record<string, unknown>;
}

function sortedRows<T>(values: unknown, idField: string, label: string, parse: (row: Record<string, unknown>) => T): ReadonlyMap<string, T> {
  if (!Array.isArray(values)) throw new SyntaxError(`${label} must be an array`);
  const result = new Map<string, T>();
  let prior = "";
  for (const raw of values) {
    const row = raw as Record<string, unknown>;
    const rowID = id(row?.[idField]);
    if (rowID <= prior) throw new SyntaxError(`${label} must be byte-sorted and unique`);
    prior = rowID;
    result.set(rowID, parse(row));
  }
  return result;
}

export function parseFeaturesPresentation(value: unknown): FeaturesPresentation {
  const root = exact(value, ["doom_meter", "fiscal_unlocks", "meter_bands", "opportunity_effects", "pet_actions", "pet_bands", "schema_version", "trust_meters"], "features presentation");
  if (root.schema_version !== 2) throw new SyntaxError("features presentation must be schema v2");
  const doom = exact(root.doom_meter, ["meter_id", "title_key", "tooltip_key"], "doom meter");
  return Object.freeze({
    doomMeter: Object.freeze({ meter_id: id(doom.meter_id), title_key: key(doom.title_key), tooltip_key: key(doom.tooltip_key) }),
    trustMeters: sortedRows(root.trust_meters, "meter_id", "trust meters", (row) => {
      exact(row, ["axis_key", "constituency_key", "meter_id"], "trust meter");
      return Object.freeze({ meter_id: row.meter_id as string, constituency_key: key(row.constituency_key), axis_key: key(row.axis_key) });
    }),
    meterBands: sortedRows(root.meter_bands, "band_id", "meter bands", (row) => { exact(row, ["band_id", "title_key"], "meter band"); return key(row.title_key); }),
    fiscalUnlocks: sortedRows(root.fiscal_unlocks, "unlock_id", "fiscal unlocks", (row) => {
      exact(row, ["description_key", "title_key", "unlock_id"], "fiscal unlock");
      return Object.freeze({ title_key: key(row.title_key), description_key: key(row.description_key) });
    }),
    opportunityEffects: sortedRows(root.opportunity_effects, "effect_row_id", "opportunity effects", (row) => {
      exact(row, ["description_key", "effect_row_id", "title_key"], "opportunity effect");
      return Object.freeze({ title_key: key(row.title_key), description_key: key(row.description_key) });
    }),
    petActions: sortedRows(root.pet_actions, "action_id", "pet actions", (row) => { exact(row, ["action_id", "title_key"], "pet action"); return key(row.title_key); }),
    petBands: sortedRows(root.pet_bands, "band_id", "pet bands", (row) => { exact(row, ["band_id", "title_key"], "pet band"); return key(row.title_key); }),
  });
}

export const FEATURES_PRESENTATION = parseFeaturesPresentation(source);
