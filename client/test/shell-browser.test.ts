import axe from "axe-core";
import { flushSync, mount, unmount } from "svelte";
import { expect, it } from "vitest";

import FixtureHost from "../src/ui/FixtureHost.svelte";
import { amountRenderScheduler } from "../src/ui/render-scheduler";

interface FixtureExports {
  switchEra(): void;
  setAmount(value: string): void;
  setCap(value: { readonly amount: string; readonly reason_key: "resource.company_cash.cap.phase0" } | undefined): void;
  setReducedMotion(value: boolean): void;
}

function structure(node: Element): unknown {
  return {
    tag: node.tagName.toLowerCase(),
    attrs: [...node.attributes].map((attribute) => attribute.name).filter((name) => name !== "style" && name !== "data-era" && name !== "data-reduced-motion").sort(),
    children: [...node.children].map(structure),
  };
}

it.skipIf(typeof document === "undefined")("switches the full primitive fixture between eras without changing DOM structure", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const fixture = mount(FixtureHost, { target }) as unknown as FixtureExports;
  flushSync();
  const root = target.querySelector("main")!;
  const before = structure(root);
  expect(root.dataset.era).toBe("era_1995");
  expect(root.style.getPropertyValue("--cc-motion-duration_base")).toBe("0ms");

  fixture.switchEra(); flushSync();
  expect(root.dataset.era).toBe("era_2000");
  expect(root.style.getPropertyValue("--cc-color-accent")).toBe("#3366cc");
  expect(structure(root)).toEqual(before);

  fixture.setReducedMotion(true); flushSync();
  expect(root.dataset.reducedMotion).toBe("true");
  expect(root.style.getPropertyValue("--cc-motion-duration_base")).toBe("0ms");
  await unmount(fixture); target.remove();
});

it.skipIf(typeof document === "undefined")("renders Amount synchronously for cap changes and coalesces a 20 Hz feed", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const fixture = mount(FixtureHost, { target }) as unknown as FixtureExports;
  flushSync();
  const output = target.querySelector("output")!;
  expect(output.textContent).toBe("1.23 K");

  fixture.setCap({ amount: "1e3", reason_key: "resource.company_cash.cap.phase0" }); flushSync();
  expect(target.textContent).toContain("Cash is capped. The cap is a number, the number is visible, and nothing will ever sell you the difference.");
  expect(target.textContent).not.toContain("resource.company_cash.cap.phase0");

  let outputMutations = 0;
  const observer = new MutationObserver(() => { outputMutations += 1; });
  observer.observe(output, { childList: true, characterData: true, subtree: true });
  fixture.setCap(undefined); flushSync();
  outputMutations = 0;
  for (let index = 0; index < 20; index += 1) fixture.setAmount(`1.${index + 10}1e3`);
  flushSync();
  expect(outputMutations).toBe(0);
  await new Promise((resolve) => setTimeout(resolve, 120));
  expect(outputMutations).toBeLessThanOrEqual(1);
  observer.disconnect();

  fixture.setAmount("2e3"); flushSync();
  expect(amountRenderScheduler.pendingCount).toBe(1);
  await unmount(fixture);
  expect(amountRenderScheduler.pendingCount).toBe(0);
  target.remove();
});

it.skipIf(typeof document === "undefined")("meets the primitive accessibility, focus, and naming baseline", async () => {
  const { userEvent } = await import("vitest/browser");
  const target = document.createElement("div"); document.body.append(target);
  const fixture = mount(FixtureHost, { target }) as unknown as FixtureExports;
  flushSync();
  const button = target.querySelector("button")!;
  button.focus();
  expect(document.activeElement).toBe(button);
  await userEvent.keyboard("{Enter}");
  flushSync();
  expect(target.querySelector("main")?.dataset.era).toBe("era_2000");
  expect(getComputedStyle(button).outlineStyle).not.toBe("none");
  expect(button.textContent?.trim().length).toBeGreaterThan(0);
  expect(target.querySelector("output")?.getAttribute("aria-label")).toBeNull();

  for (const era of ["era_1995", "era_2000"] as const) {
    if (era === "era_1995") {
      fixture.switchEra();
    } else {
      fixture.switchEra();
      fixture.setCap({ amount: "1e3", reason_key: "resource.company_cash.cap.phase0" });
      flushSync();
      expect(target.querySelector("output")?.getAttribute("aria-label")).toBe("1.23 K. Cash is capped. The cap is a number, the number is visible, and nothing will ever sell you the difference.");
    }
    const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
    expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), era).toEqual([]);
  }
  await unmount(fixture); target.remove();
});
