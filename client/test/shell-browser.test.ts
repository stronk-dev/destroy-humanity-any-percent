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
  await controller.beginAttempt();
  controller.applyAuthoritative({ revision: 7, evaluatedThroughMs: 1, constantsHash: `sha256:${"a".repeat(64)}`, resources: { "company.cash": { amount: "1e3", ratePerSecond: "1e0", cap: { amount: "1e3", reasonKey: "cap.phase0_cash" } } }, discrete: {}, progress: [] }, undefined, 1);
  await Promise.resolve();
  expect(target.textContent).toContain("company.cash"); expect(target.textContent).toContain("cap.phase0_cash");
  await unmount(app); target.remove();
});
