import { describe, expect, it } from "vitest";
import era1995 from "../../../ui/themes/era_1995.json";
import era2000 from "../../../ui/themes/era_2000.json";
import { THEME_KEYS, UI_THEMES, installTheme, parseTheme } from "../../src/ui/themes";

describe("UI theme contract", () => {
  it("loads both exact 41-token era artifacts", () => {
    expect(parseTheme(era1995)).toEqual(UI_THEMES.era_1995);
    expect(parseTheme(era2000)).toEqual(UI_THEMES.era_2000);
    expect(Object.values(THEME_KEYS).flat()).toHaveLength(41);
    expect(UI_THEMES.era_1995.motion).toEqual({
      duration_fast: "0ms",
      duration_base: "0ms",
      duration_slow: "0ms",
      easing: "linear",
      budget: "none",
    });
  });

  it("rejects missing, extra, and out-of-domain tokens", () => {
    const missing = structuredClone(era1995) as Record<string, any>;
    delete missing.color.accent;
    expect(() => parseTheme(missing)).toThrow(/exact keys/);

    const extra = structuredClone(era1995) as Record<string, any>;
    extra.motion.sparkle = "mandatory";
    expect(() => parseTheme(extra)).toThrow(/exact keys/);

    const literal = structuredClone(era1995) as Record<string, any>;
    literal.motion.duration_fast = "0.1s";
    expect(() => parseTheme(literal)).toThrow(/token domain/);
  });

  it.skipIf(typeof document === "undefined")("installs only the declared data era and token values", () => {
    const root = document.createElement("main");
    installTheme(root, UI_THEMES.era_1995);
    expect(root.dataset.era).toBe("era_1995");
    expect(root.style.getPropertyValue("--cc-color-accent")).toBe("#000080");
    expect([...root.style].filter((name) => name.startsWith("--cc-"))).toHaveLength(41);

    installTheme(root, UI_THEMES.era_2000);
    expect(root.dataset.era).toBe("era_2000");
    expect(root.style.getPropertyValue("--cc-color-accent")).toBe("#3366cc");
    expect([...root.style].filter((name) => name.startsWith("--cc-"))).toHaveLength(41);
  });
});
