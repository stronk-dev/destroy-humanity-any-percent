import source from "./axis-presentation.json";

import { applicationCopyCatalog, type CopyKey } from "../copy";
import { GAME_UI_PRESENTATION, requirePresentation } from "./presentation";

// Clout v1 CV9: presentation rows for the axis-scaled (PR Intern) upgrades.
// They live beside the pinned presentation catalog, as the Garage lane's
// features presentation does, until a content mint folds them into it; an ID
// in both is a configuration error. Unknown IDs still fail loud.
type Binding = Readonly<{ id: string; title_key: CopyKey; description_key: CopyKey }>;

export function parseAxisPresentation(value: unknown): ReadonlyMap<string, Binding> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError("axis presentation must be an object");
  const root = value as Record<string, unknown>;
  if (Object.keys(root).sort().join("\0") !== "schema_version\0upgrades" || root.schema_version !== 1 || !Array.isArray(root.upgrades)) throw new SyntaxError("axis presentation must be exact schema v1");
  const result = new Map<string, Binding>();
  let prior = "";
  for (const raw of root.upgrades) {
    const row = raw as Record<string, unknown>;
    if (row === null || typeof row !== "object" || Object.keys(row).sort().join("\0") !== "description_key\0id\0title_key") throw new SyntaxError("axis presentation row fields are not exact");
    const id = row.id;
    if (typeof id !== "string" || id <= prior || !/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(id)) throw new SyntaxError("axis presentation IDs must be byte-sorted mechanical IDs");
    if (GAME_UI_PRESENTATION.upgrades.has(id)) throw new SyntaxError(`axis presentation duplicates pinned upgrade ${id}`);
    for (const key of [row.title_key, row.description_key]) if (typeof key !== "string" || !applicationCopyCatalog.byKey.has(key)) throw new SyntaxError(`axis presentation references unknown copy ${String(key)}`);
    result.set(id, Object.freeze({ id, title_key: row.title_key as CopyKey, description_key: row.description_key as CopyKey }));
    prior = id;
  }
  return result;
}

export const AXIS_PRESENTATION = parseAxisPresentation(source);

export function upgradePresentation(id: string): Binding {
  return AXIS_PRESENTATION.get(id) ?? requirePresentation(GAME_UI_PRESENTATION.upgrades, id);
}
