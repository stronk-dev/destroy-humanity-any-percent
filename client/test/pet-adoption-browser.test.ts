import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import { t, type CopyKey } from "../src/copy";
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

for (const preexisting of [false, true]) {
  it.skipIf(!browser)(`does not celebrate a pet read without a local applied adoption (preexisting ${preexisting})`, async () => {
    const target = host();
    const outside = document.createElement("button"); target.before(outside); outside.focus();
    let calls = 0;
    const app = mount(PetAdoptionHarness, { target, props: { initial: {
      availability: { ...availability, count: preexisting ? 1 : 0 }, adopted: preexisting ? adopted : undefined,
      era: "era_1995", pending: false, controlsEnabled: true, rejection: null, reducedMotion: true,
      onAdopt: async () => { calls += 1; return undefined; },
    } } }) as unknown as { update(patch: Record<string, unknown>): void };
    try {
      await settle();
      if (!preexisting) { app.update({ adopted, availability: { ...availability, count: 1 } }); await settle(); }
      expect(target.querySelector("[aria-live=polite]")?.textContent, "saved or remote state is not an applied local receipt").toBe("");
      expect(document.activeElement, "a read must not take keyboard focus").toBe(outside);
      expect(target.querySelector("h2")?.textContent).toBe(t("pet.adoption.welcome.title", { pet_name: t(adopted.name_key as CopyKey, {}, "era_1995") }, "era_1995"));
      expect(calls).toBe(0);
    } finally { await unmount(app as never); target.remove(); outside.remove(); }
  });
}

for (const key of ["{Enter}", " "]) for (const removed of [false, true]) {
  it.skipIf(!browser)(`adoption completion preserves a newer ${removed ? "removed" : "retained"} focus choice (${JSON.stringify(key)})`, async () => {
    const { userEvent } = await import("vitest/browser");
    const target = host();
    const outside = document.createElement("button"); outside.tabIndex = 0; target.after(outside);
    let finish!: (id: string | undefined) => void;
    const completion = new Promise<string | undefined>((resolve) => { finish = resolve; });
    const app = mount(PetAdoptionHarness, { target, props: { initial: { availability, adopted: undefined,
      era: "era_1995", pending: false, controlsEnabled: true, rejection: null, reducedMotion: true,
      onAdopt: () => completion,
    } } }) as unknown as { update(patch: Record<string, unknown>): void };
    try {
      await settle();
      const adopt = button(target, "action.adopt"); adopt.focus();
      await userEvent.keyboard(key); app.update({ pending: true }); await settle();
      await userEvent.keyboard("{Tab}{Tab}");
      expect(document.activeElement, "native Tab selected a newer outside control").toBe(outside);
      if (removed) outside.remove();
      const chosenFocus = document.activeElement;
      app.update({ adopted, availability: { ...availability, count: 1 }, pending: false });
      finish(adopted.pet_id); await settle();
      expect(document.activeElement, "completion cannot reclaim a superseded focus choice").toBe(chosenFocus);
      expect(target.querySelector("[aria-live=polite]")?.textContent).toBe(t("pet.adoption.announce.adopted", { pet_name: t(adopted.name_key as CopyKey, {}, "era_1995") }, "era_1995"));
    } finally { finish(undefined); await unmount(app as never); target.remove(); outside.remove(); }
  });
}

it.skipIf(!browser)("adopts by keyboard, focuses the welcome, announces once, and shows no price text", async () => {
  const { userEvent } = await import("vitest/browser");
  const adoptions: [string, string][] = [];
  let finish!: (id: string | undefined) => void;
  const initial = { availability, adopted: undefined, era: "era_1995" as const, pending: false, controlsEnabled: true,
    rejection: null, reducedMotion: false, onAdopt: (species: string, name: string) => {
      adoptions.push([species, name]);
      return new Promise<string | undefined>((resolve) => { finish = resolve; });
    } };
  const target = host();
  const sentinel = document.createElement("button");
  target.before(sentinel);
  const app = mount(PetAdoptionHarness, { target, props: { initial } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    await assertAxe(target, "card");
    const radios = [...target.querySelectorAll<HTMLInputElement>("input[type=radio]")];
    expect(radios.map((radio) => radio.checked)).toEqual([true, false, false]);
    sentinel.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement, "Tab reaches the preselected name").toBe(radios[0]);
    await userEvent.keyboard("{ArrowRight}");
    await settle();
    expect(document.activeElement).toBe(radios[1]);
    expect(radios.map((radio) => radio.checked)).toEqual([false, true, false]);
    await userEvent.keyboard("{Tab}");
    const adopt = button(target, "action.adopt");
    expect(document.activeElement, "Tab reaches Adopt after the radiogroup").toBe(adopt);
    await userEvent.keyboard("{Enter}");
    expect(adoptions).toEqual([["pet_species.server_room_cat", "pet.name.server_room_cat.n01"]]);
    app.update({ pending: true });
    await settle();
    expect(document.activeElement, "pending keeps the adopting player's focus").toBe(adopt);
    expect(adopt.getAttribute("aria-disabled")).toBe("true");
    const pendingStatus = target.querySelector<HTMLElement>("[role=status]")!;
    expect(pendingStatus.textContent).toBe(t("common.pending", {}, "era_1995"));
    expect(getComputedStyle(pendingStatus).position, "pending feedback is visible, not clipped announcement text").toBe("static");
    expect(getComputedStyle(pendingStatus).clipPath).toBe("none");
    await userEvent.keyboard("{Enter} ");
    expect(adoptions).toHaveLength(1);
    app.update({ pending: false, rejection: "pet.adoption.reject.adoption_cap_reached" });
    finish(undefined);
    await settle();
    expect(document.activeElement, "a refusal leaves focus at Adopt").toBe(adopt);
    expect(adopt.getAttribute("aria-describedby")).toBe("pet-adoption-rejection");
    for (const token of ["$", "0.00", "Free", "free", "limited"]) expect(target.textContent, token).not.toContain(token);
    await userEvent.keyboard("{Enter}");
    expect(adoptions).toHaveLength(2);
    app.update({ rejection: null });
    app.update({ adopted, availability: { ...availability, count: 1 } });
    await settle();
    const live = target.querySelector("[aria-live=polite]")!;
    expect(live.textContent, "a projection arriving before its applied receipt is not enough").toBe("");
    let announcements = 0;
    const observer = new MutationObserver(() => { if (live.textContent?.includes("announce adopted")) announcements += 1; });
    observer.observe(live, { childList: true, characterData: true, subtree: true });
    finish(adopted.pet_id);
    await settle();
    expect(document.activeElement?.id).toBe("pet-adoption-heading");
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
    expect(announcements, "one actual live-region update, including resync").toBe(1);
    observer.disconnect();
    expect(target.querySelectorAll("[aria-live]")).toHaveLength(1);
  } finally { unmount(app as never); target.remove(); sentinel.remove(); }
});

it.skipIf(!browser)("waits for the applied receipt's matching pet, not an unrelated projection", async () => {
  const { userEvent } = await import("vitest/browser");
  const target = host();
  const app = mount(PetAdoptionHarness, { target, props: { initial: { availability, adopted: undefined,
    era: "era_1995", pending: false, controlsEnabled: true, rejection: null, reducedMotion: true,
    onAdopt: async () => adopted.pet_id,
  } } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    const adopt = button(target, "action.adopt"); adopt.focus();
    await userEvent.keyboard("{Enter}"); await settle();
    expect(target.querySelector("[aria-live=polite]")?.textContent).toBe("");
    expect(document.activeElement).toBe(adopt);
    app.update({ adopted: { ...adopted, pet_id: "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb" } }); await settle();
    expect(target.querySelector("[aria-live=polite]")?.textContent).toBe("");
    expect(document.activeElement?.id).not.toBe("pet-adoption-heading");
    app.update({ adopted }); await settle();
    expect(target.querySelector("[aria-live=polite]")?.textContent).toBe(t("pet.adoption.announce.adopted", { pet_name: t(adopted.name_key as CopyKey, {}, "era_1995") }, "era_1995"));
    expect(document.activeElement?.id).toBe("pet-adoption-heading");
  } finally { await unmount(app as never); target.remove(); }
});

it.skipIf(!browser)("a late adoption completion cannot affect an unmounted card", async () => {
  const { userEvent } = await import("vitest/browser");
  const target = host();
  const outside = document.createElement("button"); target.after(outside);
  let finish!: (id: string | undefined) => void;
  const completion = new Promise<string | undefined>((resolve) => { finish = resolve; });
  const app = mount(AdoptionCard, { target, props: { availability, adopted: undefined, era: "era_1995",
    pending: false, controlsEnabled: true, rejection: null, reducedMotion: true, onAdopt: () => completion } });
  let mounted = true;
  try {
    await settle(); button(target, "action.adopt").focus(); await userEvent.keyboard("{Enter}");
    await unmount(app); mounted = false; outside.focus();
    finish(adopted.pet_id); await settle();
    expect(document.activeElement).toBe(outside);
    expect(target.querySelector("[aria-live]")).toBeNull();
  } finally { finish(undefined); if (mounted) await unmount(app); target.remove(); outside.remove(); }
});

for (const reopening of [false, true]) for (const boundary of ["newer-focus", "unmount"] as const) {
  it.skipIf(!browser)(`adoption ${reopening ? "reopen" : "collapse"} focus handoff respects ${boundary}`, async () => {
    const target = host();
    const outside = document.createElement("button"); target.after(outside);
    const app = mount(AdoptionCard, { target, props: { availability, adopted: undefined, era: "era_1995", pending: false,
      controlsEnabled: true, rejection: null, reducedMotion: true, onAdopt: async () => undefined } });
    let mounted = true;
    try {
      await settle();
      if (reopening) { button(target, "action.later").click(); await settle(); }
      const trigger = button(target, reopening ? "entry point" : "action.later");
      trigger.focus(); trigger.click();
      // The real handler schedules its post-render handoff; replace focus or
      // unmount synchronously before tick. This is a lifecycle control, not
      // a claim of keyboard traversal (the two journeys above use native input).
      if (boundary === "unmount") {
        const disposal = unmount(app); mounted = false;
        outside.focus(); await disposal;
      } else outside.focus();
      await settle();
      expect(document.activeElement).toBe(outside);
      if (mounted) expect(target.querySelector("fieldset") !== null).toBe(reopening);
    } finally { if (mounted) await unmount(app); target.remove(); outside.remove(); }
  });
}

it.skipIf(!browser)("renders a static pose under reduced motion", async () => {
  const target = host();
  const app = mount(AdoptionCard, { target, props: { availability: { ...availability, count: 1 }, adopted, era: "era_1995", pending: false, controlsEnabled: true, rejection: null, reducedMotion: true, onAdopt: async () => undefined } });
  try {
    await settle();
    const sprite = target.querySelector<HTMLElement>("[data-family=cat]")!;
    expect(sprite.dataset.animate).toBe("false");
    for (const part of sprite.querySelectorAll("div")) expect(getComputedStyle(part).animationName).toBe("none");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("collapses on Not now to a persistent entry point and shows rejections inline", async () => {
  const { userEvent } = await import("vitest/browser");
  const target = host(320);
  const sentinel = document.createElement("button"); target.before(sentinel);
  const app = mount(AdoptionCard, { target, props: { availability, adopted: undefined, era: "era_1995", pending: false, controlsEnabled: true,
    rejection: "pet.adoption.reject.adoption_cap_reached", reducedMotion: false, onAdopt: async () => undefined } });
  try {
    await settle();
    expect(target.scrollWidth).toBeLessThanOrEqual(target.clientWidth);
    const adopt = button(target, "action.adopt");
    expect(adopt.getAttribute("aria-describedby")).toBe("pet-adoption-rejection");
    expect(target.querySelector("#pet-adoption-rejection")?.textContent).toContain("adoption cap reached");
    await assertAxe(target, "rejection");
    sentinel.focus();
    await userEvent.keyboard("{Tab}{Tab}{Tab}");
    expect(document.activeElement, "Not now follows name choice and Adopt").toBe(button(target, "action.later"));
    await userEvent.keyboard(" ");
    await settle();
    const entry = button(target, "entry point");
    expect(target.querySelector("fieldset")).toBeNull();
    expect(document.activeElement, "the removed Not now control hands focus to its persistent entry point").toBe(entry);
    await userEvent.keyboard("{Enter}");
    await settle();
    expect(target.querySelector("fieldset")).not.toBeNull();
    expect(document.activeElement, "reopening returns to the selected name").toBe(target.querySelector("input:checked"));
  } finally { unmount(app); target.remove(); sentinel.remove(); }
});
