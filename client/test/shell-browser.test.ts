import { mount, unmount } from "svelte";
import { expect, it } from "vitest";
import policySource from "../../balance/client-shell/phase0.json";
import App from "../src/shell/App.svelte";
import { ShellController } from "../src/shell/controller";
import { parseClientShellPolicy } from "../src/shell/policy";

it.skipIf(typeof document === "undefined")("renders the contract, enters the generic shell, and shows cap receipts", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const controller = new ShellController(parseClientShellPolicy(policySource));
  const app = mount(App, { target, props: { controller } });
  expect(target.textContent).toContain("BEGIN ATTEMPT");
  expect(target.textContent).toContain("comparison unlocks after the first Exit");
  await controller.beginAttempt();
  controller.applyAuthoritative({ revision: 7, evaluatedThroughMs: 1, constantsHash: `sha256:${"a".repeat(64)}`, resources: { "company.cash": { amount: "1e3", ratePerSecond: "1e0", cap: { amount: "1e3", reasonKey: "cap.phase0_cash" } } }, discrete: {}, progress: [{ stageId: "tier.bootstrap", current: "1e0", target: "1e1" }] }, undefined, 1);
  await Promise.resolve();
  expect(target.textContent).toContain("company.cash"); expect(target.textContent).toContain("cap.phase0_cash"); expect(target.textContent).toContain("tier.bootstrap");
  controller.showRunEnd(); await Promise.resolve(); expect(target.textContent).toContain("Attempt complete");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("snaps discrete facts while a continuous value bends and renders a rejected buy", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const controller = new ShellController(parseClientShellPolicy(policySource));
  const app = mount(App, { target, props: { controller } }); await controller.beginAttempt();
  const base = { evaluatedThroughMs: 1, constantsHash: `sha256:${"a".repeat(64)}`, progress: [] };
  controller.applyAuthoritative({ ...base, revision: 1, resources: { "company.cash": { amount: "1e2", ratePerSecond: "1e0" } }, discrete: { "unlock.shop": false } }, undefined, 0);
  controller.applyPrediction({ revision: 1, atMonotonicMs: 100, resources: { "company.cash": { mantissa: 1.005, exponent: 2 } } }, 100);
  controller.applyAuthoritative({ ...base, revision: 2, resources: { "company.cash": { amount: "1e2", ratePerSecond: "1e0" } }, discrete: { "unlock.shop": true } }, undefined, 100);
  await Promise.resolve();
  expect(Number(controller.view(100).resources["company.cash"].value)).toBeCloseTo(100.5, 12);
  expect(target.querySelector('[data-state-id="unlock.shop"]')?.textContent).toContain("true");
  controller.applyAuthoritative({ ...base, revision: 3, resources: { "company.cash": { amount: "1e2", ratePerSecond: "1e0" } }, discrete: { "unlock.shop": true } }, { revision: 3, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected", rejectionCode: "unaffordable" }, 100);
  await Promise.resolve(); expect(target.textContent).toContain("unaffordable");
  await unmount(app); target.remove(); controller.dispose();
});

it.skipIf(typeof document === "undefined")("covers a ten-minute return with authoritative gains behind the recap", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const controller = new ShellController(parseClientShellPolicy(policySource));
  const app = mount(App, { target, props: { controller } });
  await controller.beginAttempt();
  const base = { evaluatedThroughMs: 1, constantsHash: `sha256:${"a".repeat(64)}`, discrete: {}, progress: [] };
  controller.applyAuthoritative({ ...base, revision: 1, resources: { "company.cash": { amount: "1e2", ratePerSecond: "1e0" } } }, undefined, 1);
  controller.beginReturnStory(600_000);
  controller.applyAuthoritative({ ...base, revision: 2, evaluatedThroughMs: 600_001, resources: { "company.cash": { amount: "7e2", ratePerSecond: "1e0" } } }, undefined, 2);
  await Promise.resolve();
  expect(target.textContent).toContain("Fast-forwarding your return");
  expect(target.textContent).toContain("company.cash +600");
  expect(target.querySelector("article.pulse")).toBeNull();
  controller.completeReturnFastForward(); await Promise.resolve();
  expect(target.textContent).toContain("Return complete");
  controller.dismissReturnStory();
  await unmount(app); target.remove(); controller.dispose();
});
