import paletteSource from "./palette.json";

// Pet Adoption v1 PA8.5: the single pet visual contract. The CSS-only sprite
// renders only from this spec; cosmetic-shop-v1 extends the pose union here.
export type PetVisualFamily = "cat";
export type PetPose = "content" | "low" | "withdrawn";
export type PetStatusBand = "floor" | "high" | "low" | "normal";

export interface PetVisualSpec {
  readonly family: PetVisualFamily;
  readonly palette_id: string;
  readonly pose: PetPose;
  readonly animate: boolean;
}

const poses: Readonly<Record<PetStatusBand, PetPose>> = { high: "content", normal: "content", low: "low", floor: "withdrawn" };

export function petVisualSpec(identity: Readonly<{ palette_id: string }>, statusBand: PetStatusBand, reducedMotion: boolean): PetVisualSpec {
  const pose = poses[statusBand];
  if (!pose) throw new RangeError(`unknown pet status band ${String(statusBand)}`);
  if (!PET_SWATCHES.has(identity.palette_id)) throw new RangeError(`unknown pet palette ${identity.palette_id}`);
  return Object.freeze({ family: "cat", palette_id: identity.palette_id, pose, animate: !reducedMotion });
}

export interface PetSwatch { readonly fur: string; readonly outline: string }

const hex = /^#[0-9a-f]{6}$/u;

export function parsePetPalette(source: unknown): ReadonlyMap<string, PetSwatch> {
  const root = source as { schema_version?: unknown; swatches?: unknown };
  if (root === null || typeof root !== "object" || root.schema_version !== 1 || !Array.isArray(root.swatches) || Object.keys(root).length !== 2) throw new SyntaxError("invalid pet palette");
  const result = new Map<string, PetSwatch>();
  for (const row of root.swatches as Record<string, unknown>[]) {
    if (Object.keys(row).sort().join(",") !== "fur,outline,palette_id" || typeof row.palette_id !== "string" || typeof row.fur !== "string" ||
      typeof row.outline !== "string" || !hex.test(row.fur) || !hex.test(row.outline) || result.has(row.palette_id)) throw new SyntaxError("invalid pet swatch");
    result.set(row.palette_id, Object.freeze({ fur: row.fur, outline: row.outline }));
  }
  return result;
}

export const PET_SWATCHES = parsePetPalette(paletteSource);

// WCAG 2.x relative luminance and contrast ratio (non-text contrast, 1.4.11).
export function contrastRatio(left: string, right: string): number {
  const luminance = (value: string): number => {
    const channels = [1, 3, 5].map((offset) => Number.parseInt(value.slice(offset, offset + 2), 16) / 255)
      .map((channel) => channel <= 0.03928 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4);
    return 0.2126 * channels[0]! + 0.7152 * channels[1]! + 0.0722 * channels[2]!;
  };
  const [a, b] = [luminance(left), luminance(right)].sort((x, y) => y - x);
  return (a! + 0.05) / (b! + 0.05);
}
