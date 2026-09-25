import source from "./presentation.json";

import { COPY_KEYS, type CopyKey } from "../../copy";

// Cosmetic Shop v1 §7.2/§8: presentation rows and the curtain-binding
// contract. Every parodied pattern binds exactly one disclosure key; an
// unbound, duplicated, or unknown pattern rejects the catalog at load time
// (and therefore at build/test time: this module loads eagerly).
export type CurtainPattern = "paid_cosmetic_dlc" | "reference_price_anchor" | "checkout_flow" | "re_release";
export type WearerReaction = "annoyed" | "none";
export interface Curtain { readonly pattern: CurtainPattern; readonly key: CopyKey }
export interface CosmeticPresentation {
  readonly cosmetic_id: string; readonly title_key: CopyKey; readonly description_key: CopyKey; readonly anchor_key: CopyKey | null;
  readonly render_key: string; readonly wearer_reaction: WearerReaction; readonly curtains: readonly Curtain[];
}
export interface CosmeticShopPresentation {
  readonly cosmetics: ReadonlyMap<string, CosmeticPresentation>;
  readonly shop: { readonly heading_key: CopyKey; readonly curtains: readonly Curtain[] };
  /** The exported curtain list design/08 §7's honesty appendix enumerates. */
  readonly curtainList: readonly { readonly element: string; readonly pattern: CurtainPattern; readonly key: CopyKey }[];
}

const PATTERNS = new Set<string>(["paid_cosmetic_dlc", "reference_price_anchor", "checkout_flow", "re_release"]);
const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;

function object(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  return value as Record<string, unknown>;
}
function exact(value: Record<string, unknown>, keys: readonly string[], label: string): void {
  const actual = Object.keys(value).sort(), expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} keys are not exact`);
}
function copyKey(value: unknown, keys: ReadonlySet<string>, label: string): CopyKey {
  if (typeof value !== "string" || !keys.has(value)) throw new SyntaxError(`${label} is not a declared copy key`);
  return value as CopyKey;
}
function curtains(value: unknown, keys: ReadonlySet<string>, label: string): Curtain[] {
  if (!Array.isArray(value)) throw new SyntaxError(`${label} curtains must be an array`);
  const seen = new Set<string>();
  return value.map((raw) => {
    const row = object(raw, `${label} curtain`);
    exact(row, ["key", "pattern"], `${label} curtain`);
    if (typeof row.pattern !== "string" || !PATTERNS.has(row.pattern)) throw new SyntaxError(`${label} binds an unknown pattern`);
    if (seen.has(row.pattern)) throw new SyntaxError(`${label} binds ${row.pattern} twice`);
    seen.add(row.pattern);
    return { pattern: row.pattern as CurtainPattern, key: copyKey(row.key, keys, `${label} curtain key`) };
  });
}

export function parseCosmeticShopPresentation(input: unknown, keys: ReadonlySet<string> = new Set(COPY_KEYS)): CosmeticShopPresentation {
  const root = object(input, "cosmetic presentation");
  exact(root, ["cosmetic_shop", "cosmetics", "schema_version"], "cosmetic presentation");
  if (root.schema_version !== 1 || !Array.isArray(root.cosmetics)) throw new SyntaxError("invalid cosmetic presentation");
  const cosmetics = new Map<string, CosmeticPresentation>();
  const curtainList: { element: string; pattern: CurtainPattern; key: CopyKey }[] = [];
  for (const raw of root.cosmetics) {
    const row = object(raw, "cosmetic presentation row");
    exact(row, ["anchor_key", "cosmetic_id", "curtains", "description_key", "render_key", "title_key", "wearer_reaction"], "cosmetic presentation row");
    if (typeof row.cosmetic_id !== "string" || !mechanical.test(row.cosmetic_id) || cosmetics.has(row.cosmetic_id)) throw new SyntaxError("invalid cosmetic presentation id");
    if (typeof row.render_key !== "string" || !mechanical.test(row.render_key)) throw new SyntaxError("invalid render_key");
    if (row.wearer_reaction !== "annoyed" && row.wearer_reaction !== "none") throw new SyntaxError("invalid wearer_reaction");
    const anchor = row.anchor_key === null ? null : copyKey(row.anchor_key, keys, "anchor_key");
    const bound = curtains(row.curtains, keys, row.cosmetic_id);
    const patterns = new Set(bound.map((curtain) => curtain.pattern));
    if (!patterns.has("paid_cosmetic_dlc")) throw new SyntaxError(`${row.cosmetic_id} is missing its paid_cosmetic_dlc curtain`);
    if ((anchor !== null) !== patterns.has("reference_price_anchor")) throw new SyntaxError(`${row.cosmetic_id} anchor and reference_price_anchor curtain must bind together`);
    if (patterns.has("checkout_flow")) throw new SyntaxError("checkout_flow binds only to the shop");
    // re_release is reserved for a row that declares a re-release (OD-1); v1 has none.
    if (patterns.has("re_release")) throw new SyntaxError("re_release is reserved and unbound in v1");
    cosmetics.set(row.cosmetic_id, Object.freeze({ cosmetic_id: row.cosmetic_id, title_key: copyKey(row.title_key, keys, "title_key"),
      description_key: copyKey(row.description_key, keys, "description_key"), anchor_key: anchor, render_key: row.render_key,
      wearer_reaction: row.wearer_reaction, curtains: Object.freeze(bound) }));
    for (const curtain of bound) curtainList.push({ element: `cosmetic:${row.cosmetic_id}`, ...curtain });
  }
  const shopRow = object(root.cosmetic_shop, "cosmetic shop presentation");
  exact(shopRow, ["curtains", "heading_key"], "cosmetic shop presentation");
  const shopCurtains = curtains(shopRow.curtains, keys, "shop");
  if (shopCurtains.length !== 1 || shopCurtains[0]!.pattern !== "checkout_flow") throw new SyntaxError("the shop must bind exactly the checkout_flow curtain");
  for (const curtain of shopCurtains) curtainList.push({ element: "cosmetic_shop", ...curtain });
  return Object.freeze({ cosmetics, shop: Object.freeze({ heading_key: copyKey(shopRow.heading_key, keys, "heading_key"), curtains: Object.freeze(shopCurtains) }),
    curtainList: Object.freeze(curtainList) });
}

export const COSMETIC_SHOP_PRESENTATION = parseCosmeticShopPresentation(source);
