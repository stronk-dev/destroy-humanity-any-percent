import { expect, it } from "vitest";

// Not a *.test.ts file: materialize at a temporary test path, run only that
// file, then remove it. Deliberately fails and holds before the aggregate.
const browser = typeof document !== "undefined";
it.skipIf(!browser)("ordinary reporter negative control", () => {
  expect("observed", "controlled ordinary failure before a later guard").toBe("expected");
});
it.skipIf(!browser).fails("expected failure reporter control", () => { expect(1).toBe(2); });
it.skipIf(!browser)("healthy reporter control", () => { expect(1).toBe(1); });
it.skipIf(!browser).fails("unexpected pass reporter negative control", () => { expect(1).toBe(1); });
it.skipIf(!browser)("held reporter control", async () => {
  const { server } = await import("vitest/browser");
  console.info("browser-failure-probe-hold", server.browser);
  await new Promise<void>(() => {});
}, 90_000);
