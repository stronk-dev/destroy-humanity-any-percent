import axe from "axe-core";
import { expect } from "vitest";
import type { GameUISnapshot } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import { NOW, meterRow, settle, v4 } from "./garage-surfaces-fixture";

// Shared assertions and public fixture builders; no test registration.


export const browser = typeof document !== "undefined";


export async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

export function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) =>
    (candidate.querySelector(":scope > [data-nav-label]") ?? candidate).textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}`);
  return found;
}

// GS0.5/GS2: shared-state public fixtures, not a real network outage or earn.
export function sharedStateVisibleText(parent: Element, selector: string, text: string): HTMLElement {
  const nodes = [...parent.querySelectorAll<HTMLElement>(selector)].filter((node) => node.textContent === text);
  expect(nodes, `shared-state visible exact ${text}`).toHaveLength(1);
  const node = nodes[0]!, rectangle = node.getBoundingClientRect(), style = getComputedStyle(node);
  expect(rectangle.width).toBeGreaterThan(0); expect(rectangle.height).toBeGreaterThan(0);
  expect(style.display).not.toBe("none"); expect(style.visibility).toBe("visible"); expect(Number(style.opacity)).toBeGreaterThan(0);
  return node;
}

// RP-330: real semantic layout over decoder-admitted public fixtures, not
// production/persisted values or assistive-technology evidence.
export const semanticMeterRows = v4.features.meters!.meters.map((row, index) => meterRow(row.meter_id, index * 7));
export function semanticMeterSnapshot(rows = semanticMeterRows): ParsedGameUISnapshot {
  return parseGameUISnapshot({ ...v4, features: { ...v4.features, meters: { meters: rows } } });
}

export async function settleMeterLayout(): Promise<void> {
  // Native media-query change events run at rendering time, not merely after
  // the driver's viewport RPC. Observe one real frame, then flush Svelte.
  await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
  await settle();
}

export async function assertSemanticMeters(target: HTMLElement, rows: typeof semanticMeterRows, narrow: boolean): Promise<void> {
  const { page } = await import("vitest/browser");
  const surface = target.querySelector(".meters")!;
  const name = (row: (typeof rows)[number]) => {
    const presentation = FEATURES_PRESENTATION.trustMeters.get(row.meter_id);
    return presentation
      ? `${t(presentation.constituency_key, {}, "era_1995")} ${t(presentation.axis_key, {}, "era_1995")}`
      : t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995");
  };
  // Collect every name before asserting so one missing label cannot hide the others.
  expect(rows.map((row) => ({ name: name(row), count: page.getByRole("meter", { name: name(row), exact: true }).elements().length })))
    .toEqual(rows.map((row) => ({ name: name(row), count: 1 })));
  const value = (element: Element, row: (typeof rows)[number]) => {
    const meter = element.querySelector("meter")!;
    expect(meter).not.toBeNull();
    // Table-header context does not supply a native meter's accessible name.
    // Find this exact value by role/name in both layouts, including after resize.
    expect(page.getByRole("meter", { name: name(row), exact: true }).element()).toBe(meter);
    expect([meter.min, meter.max, meter.value]).toEqual([row.min, row.max, row.value]);
    expect([...element.querySelectorAll("span")].map((node) => node.textContent)).toEqual([
      t("meters.value_frame", { value: row.value, max: row.max }, "era_1995"),
      t(FEATURES_PRESENTATION.meterBands.get(row.band_id)!, {}, "era_1995"),
    ]);
  };
  expect(surface.querySelectorAll("meter")).toHaveLength(rows.length);
  if (narrow) {
    expect(surface.querySelectorAll("table")).toHaveLength(0);
    const pairs = [...surface.querySelectorAll<HTMLElement>("dl > div")];
    expect(pairs).toHaveLength(rows.length);
    expect(pairs.map((pair) => pair.dataset.meterId).sort()).toEqual(rows.map((row) => row.meter_id));
    for (const row of rows) {
      const pair = pairs.find((candidate) => candidate.dataset.meterId === row.meter_id)!;
      expect([...pair.children].map((child) => child.tagName)).toEqual(["DT", "DD"]);
      const term = pair.querySelector("dt")!, description = pair.querySelector("dd")!;
      const presentation = FEATURES_PRESENTATION.trustMeters.get(row.meter_id);
      const label = presentation
        ? `${t(presentation.constituency_key, {}, "era_1995")} ${t(presentation.axis_key, {}, "era_1995")}`
        : t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995");
      expect(term.textContent?.replace(/\s+/gu, " ").trim()).toBe(label);
      expect(term.id).not.toBe("");
      expect([...target.querySelectorAll("[id]")].filter((node) => node.id === term.id)).toHaveLength(1);
      const meter = description.querySelector("meter")!;
      expect(meter).not.toBeNull();
      expect(document.getElementById(meter.getAttribute("aria-labelledby")!)).toBe(term);
      value(description, row);
    }
  } else {
    expect(surface.querySelectorAll("dl")).toHaveLength(0);
    expect(surface.querySelectorAll("table")).toHaveLength(1);
    expect(surface.querySelectorAll("tbody tr")).toHaveLength(5);
    expect(surface.querySelectorAll("td")).toHaveLength(10);
    expect(surface.querySelectorAll('thead th[scope="col"]')).toHaveLength(3);
    expect(surface.querySelectorAll('tbody th[scope="row"]')).toHaveLength(5);
    expect([...surface.querySelectorAll("thead th")].map((node) => node.textContent)).toEqual([
      t("meters.constituency_label", {}, "era_1995"), t("meters.axis.standing", {}, "era_1995"), t("meters.axis.grievance", {}, "era_1995"),
    ]);
    for (const row of rows) {
      const presentation = FEATURES_PRESENTATION.trustMeters.get(row.meter_id);
      if (presentation) {
        const tableRow = [...surface.querySelectorAll("tbody tr")].find((node) => node.querySelector("th")!.textContent === t(presentation.constituency_key, {}, "era_1995"))!;
        value(tableRow.querySelectorAll("td")[presentation.axis_key === "meters.axis.standing" ? 0 : 1]!, row);
      } else {
        expect(surface.querySelector(".doom h2")!.textContent).toBe(t(FEATURES_PRESENTATION.doomMeter.title_key, {}, "era_1995"));
        value(surface.querySelector(".doom")!, row);
      }
    }
  }
  const root = document.documentElement;
  expect(root.scrollWidth).toBeLessThanOrEqual(root.clientWidth + 1);
  expect([...target.querySelectorAll("*")].filter((node) => node.getBoundingClientRect().right > root.clientWidth + 1)).toEqual([]);
  await assertAxe(target, narrow ? "semantic narrow meters" : "wide meter table");
}
export const withRipeFiscal = (): GameUISnapshot => ({ ...v4,
  features: { ...v4.features, fiscal: { ...v4.features.fiscal!, period: { ...v4.features.fiscal!.period, opened_wall_ms: NOW - 250_000 } } } });

export const petRow = { eligible_action_ids: ["care.feed", "care.pet"], name_key: "pet.name.server_room_cat.n04", palette_id: "pet_palette.fur_02",
  pet_id: "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa", species_id: "pet_species.server_room_cat", status_band: "normal" as const, temperament: "sassy" as const };
export const withPet = (): GameUISnapshot => ({ ...v4,
  facts: [...v4.facts.map((fact) => fact.fact_id === "feature.pets" ? { ...fact, value: true } : fact)],
  features: { ...v4.features,
    pet_adoption: { pet_adoption: { cap: 1, count: 1, name_keys: ["pet.name.server_room_cat.n04"], starter_species_id: "pet_species.server_room_cat" }, pets: [petRow] },
    cosmetics: { active: true, items: [{ acquirable: false, cosmetic_id: "horse_armor", lock: null, owned: true, worn_by: [petRow.pet_id] }], wearers: [{ pet_id: petRow.pet_id, worn: "horse_armor" }] } } } as GameUISnapshot);

// Care supplement: native host/runtime-double evidence, not a server care or
// raw-stat/cooldown projection. All values below are public PA7 fields.
export const careText = (key: CopyKey): string => t(key, {}, "era_1995");
