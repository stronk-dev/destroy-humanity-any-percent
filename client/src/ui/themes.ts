import era1995Source from "../../../ui/themes/era_1995.json";
import era2000Source from "../../../ui/themes/era_2000.json";

export const UI_ERAS = ["era_1995", "era_2000"] as const;
export type UIEra = (typeof UI_ERAS)[number];

export const THEME_KEYS = Object.freeze({
  color: ["bg", "surface", "text", "text_muted", "accent", "accent_text", "border", "link", "danger", "success"],
  type: ["font_ui", "font_display", "font_mono", "size_base", "size_small", "size_large", "size_display", "weight_normal", "weight_bold", "line_height"],
  space: ["unit", "xs", "sm", "md", "lg", "xl"],
  border: ["width", "radius", "style", "bevel"],
  chrome: ["window_bg", "window_border", "titlebar_bg", "titlebar_text", "button_face", "button_shadow"],
  motion: ["duration_fast", "duration_base", "duration_slow", "easing", "budget"],
} as const);

export type ThemeGroup = keyof typeof THEME_KEYS;
type ThemeValues<G extends ThemeGroup> = Readonly<Record<(typeof THEME_KEYS)[G][number], string>>;

export interface UITheme {
  readonly schemaVersion: 1;
  readonly era: UIEra;
  readonly color: ThemeValues<"color">;
  readonly type: ThemeValues<"type">;
  readonly space: ThemeValues<"space">;
  readonly border: ThemeValues<"border">;
  readonly chrome: ThemeValues<"chrome">;
  readonly motion: ThemeValues<"motion">;
}

export const UI_LAYOUT_CUSTOM_PROPERTIES = Object.freeze([
  "--cc-layout-columns",
  "--cc-layout-max-inline-size",
  "--cc-layout-progress",
] as const);

const rootKeys = ["border", "chrome", "color", "era", "motion", "schema_version", "space", "type"];
const hexColor = /^#[0-9a-f]{3}(?:[0-9a-f]{3})?$/i;
const px = /^(?:0|[1-9]\d*)px$/;
const duration = /^(?:0|[1-9]\d*)ms$/;
const lineHeight = /^(?:1(?:\.\d+)?)$/;
const font = /^(?:Tahoma, "MS Sans Serif", Geneva, sans-serif|Verdana, Arial, Helvetica, sans-serif|"Courier New", monospace)$/;

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

function validToken(group: ThemeGroup, key: string, value: unknown): value is string {
  if (typeof value !== "string" || value.length === 0 || value.length > 160) return false;
  if (group === "color") return hexColor.test(value);
  if (group === "type") {
    if (key.startsWith("font_")) return font.test(value);
    if (key.startsWith("size_")) return px.test(value);
    if (key.startsWith("weight_")) return value === "400" || value === "700";
    return key === "line_height" && lineHeight.test(value);
  }
  if (group === "space") return px.test(value);
  if (group === "border") {
    if (key === "width" || key === "radius") return px.test(value);
    if (key === "style") return value === "solid";
    return key === "bevel" && (value === "none" || value === "outset");
  }
  if (group === "chrome") {
    return key === "window_bg"
      ? hexColor.test(value) || /^linear-gradient\(180deg, #[0-9a-f]{6}, #[0-9a-f]{6}\)$/i.test(value)
      : hexColor.test(value);
  }
  if (key.startsWith("duration_")) return duration.test(value);
  if (key === "easing") return value === "linear" || value === "ease-in-out";
  return key === "budget" && (value === "none" || value === "respect");
}

export function parseTheme(source: unknown): UITheme {
  const root = exactObject(source, rootKeys, "theme");
  if (root.schema_version !== 1) syntax("theme.schema_version must be 1");
  if (typeof root.era !== "string" || !UI_ERAS.includes(root.era as UIEra)) syntax("theme.era is unknown");
  const parsed: Partial<Record<ThemeGroup, Readonly<Record<string, string>>>> = {};
  for (const group of Object.keys(THEME_KEYS) as ThemeGroup[]) {
    const row = exactObject(root[group], THEME_KEYS[group], `theme.${group}`);
    for (const key of THEME_KEYS[group]) {
      if (!validToken(group, key, row[key])) syntax(`theme.${group}.${key} is outside its token domain`);
    }
    parsed[group] = Object.freeze({ ...(row as Record<string, string>) });
  }
  return Object.freeze({
    schemaVersion: 1,
    era: root.era as UIEra,
    color: parsed.color as ThemeValues<"color">,
    type: parsed.type as ThemeValues<"type">,
    space: parsed.space as ThemeValues<"space">,
    border: parsed.border as ThemeValues<"border">,
    chrome: parsed.chrome as ThemeValues<"chrome">,
    motion: parsed.motion as ThemeValues<"motion">,
  });
}

export const UI_THEMES: Readonly<Record<UIEra, UITheme>> = Object.freeze({
  era_1995: parseTheme(era1995Source),
  era_2000: parseTheme(era2000Source),
});

export function installTheme(root: HTMLElement, theme: UITheme, reducedMotion = false): void {
  root.dataset.era = theme.era;
  root.dataset.reducedMotion = String(reducedMotion);
  for (const group of Object.keys(THEME_KEYS) as ThemeGroup[]) {
    const values = theme[group] as Readonly<Record<string, string>>;
    for (const key of THEME_KEYS[group]) {
      const value = reducedMotion && group === "motion" && key.startsWith("duration_") ? "0ms" : values[key];
      root.style.setProperty(`--cc-${group}-${key}`, value);
    }
  }
}
