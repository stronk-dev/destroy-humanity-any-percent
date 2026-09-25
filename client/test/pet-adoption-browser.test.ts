import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import AdoptionCard from "../src/game-ui/pet/AdoptionCard.svelte";
import PetAdoptionHarness from "./PetAdoptionHarness.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const availability = { cap: 1, count: 0, name_keys: ["pet.name.server_room_cat.n00", "pet.name.server_room_cat.n01", "pet.name.server_room_cat.n02"], starter_species_id: "pet_species.server_room_cat" };
const adopted = { pet_id: "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa", name_key: "pet.name.server_room_cat.n01", palette_id: "pet_palette.fur_02",
  species_id: "pet_species.server_room_cat", status_band: "high" as const, temperament: "curious" };

async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function host(width?: number): HTMLElement {
  const target = document.createElement("main");
  if (width) target.style.inlineSize = `${width}px`;
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  return target;
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.includes(text));
  if (!found) throw new Error(`missing ${text}: ${target.textContent}`);
  return found;
}

it.skipIf(!browser)("adopts by keyboard, focuses the welcome, announces once, and shows no price text", async () => {
  const adoptions: [string, string][] = [];
  const initial = { availability, adopted: undefined, era: "era_1995" as const, pending: false, controlsEnabled: true,
    rejection: null, reducedMotion: false, onAdopt: (species: string, name: string) => { adoptions.push([species, name]); } };
  const target = host();
  const app = mount(PetAdoptionHarness, { target, props: { initial } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    await assertAxe(target, "card");
    const radios = [...target.querySelectorAll<HTMLInputElement>("input[type=radio]")];
    expect(radios.map((radio) => radio.checked)).toEqual([true, false, false]);
    radios[1]!.focus(); radios[1]!.click();
    await settle();
    button(target, "action.adopt").click();
    expect(adoptions).toEqual([["pet_species.server_room_cat", "pet.name.server_room_cat.n01"]]);
    for (const token of ["$", "0.00", "Free", "free", "limited"]) expect(target.textContent, token).not.toContain(token);
    app.update({ adopted, availability: { ...availability, count: 1 } });
    await settle();
    expect(document.activeElement?.id).toBe("pet-adoption-heading");
    const live = target.querySelector("[aria-live=polite]")!;
    const first = live.textContent;
    expect(first).toContain("announce adopted");
    const sprite = target.querySelector<HTMLElement>("[data-family=cat]")!;
    expect(sprite.getAttribute("aria-hidden")).toBe("true");
    expect(sprite.dataset.animate).toBe("true");
    await assertAxe(target, "welcome");
    // A resync re-delivering the same pet must not announce again or steal focus.
    const elsewhere = document.createElement("button");
    document.body.append(elsewhere); elsewhere.focus();
    app.update({ adopted: { ...adopted } });
    await settle();
    expect(document.activeElement).toBe(elsewhere);
    elsewhere.remove();
    await settle();
    expect(live.textContent).toBe(first);
    expect(target.querySelectorAll("[aria-live]")).toHaveLength(1);
  } finally { unmount(app as never); target.remove(); }
});

it.skipIf(!browser)("renders a static pose under reduced motion", async () => {
  const target = host();
  const app = mount(AdoptionCard, { target, props: { availability: { ...availability, count: 1 }, adopted, era: "era_1995", pending: false, controlsEnabled: true, rejection: null, reducedMotion: true, onAdopt: () => {} } });
  try {
    await settle();
    const sprite = target.querySelector<HTMLElement>("[data-family=cat]")!;
    expect(sprite.dataset.animate).toBe("false");
    for (const part of sprite.querySelectorAll("div")) expect(getComputedStyle(part).animationName).toBe("none");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("collapses on Not now to a persistent entry point and shows rejections inline", async () => {
  const target = host(320);
  const app = mount(AdoptionCard, { target, props: { availability, adopted: undefined, era: "era_1995", pending: false, controlsEnabled: true,
    rejection: "pet.adoption.reject.adoption_cap_reached", reducedMotion: false, onAdopt: () => {} } });
  try {
    await settle();
    expect(target.scrollWidth).toBeLessThanOrEqual(target.clientWidth);
    const adopt = button(target, "action.adopt");
    expect(adopt.getAttribute("aria-describedby")).toBe("pet-adoption-rejection");
    expect(target.querySelector("#pet-adoption-rejection")?.textContent).toContain("adoption cap reached");
    await assertAxe(target, "rejection");
    button(target, "action.later").click();
    await settle();
    const entry = button(target, "entry point");
    expect(target.querySelector("fieldset")).toBeNull();
    entry.click();
    await settle();
    expect(target.querySelector("fieldset")).not.toBeNull();
  } finally { unmount(app); target.remove(); }
});
