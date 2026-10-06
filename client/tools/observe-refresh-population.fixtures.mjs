// Standalone Node controls, not a Vitest-discovered client test suite.
import assert from "node:assert/strict";
import test from "node:test";
import { leaves, parents, refreshPopulationObserver } from "./observe-refresh-population.mjs";

const event = (Action, Test) => ({ Action, Package: "cloud-clicker/server/account", ...(Test ? { Test } : {}) });
function completeFixture() {
  const rows = [event("start")];
  for (const parent of parents) {
    rows.push(event("run", parent));
    for (const child of leaves.filter((name) => name.startsWith(`${parent}/`))) rows.push(event("run", child), event("pass", child));
    rows.push(event("pass", parent));
  }
  rows.push(event("pass"));
  return rows;
}
function validate(rows, exit = 0, stable = true) {
  const observer = refreshPopulationObserver();
  for (const row of rows) observer.consumeLine(typeof row === "string" ? row : JSON.stringify(row));
  return observer.finish(exit, stable);
}

test("synthetic complete seven-leaf/four-parent control is valid (not DB evidence)", () => {
  const result = validate(completeFixture());
  assert.equal(result.valid, true); assert.equal(result.completed_leaves, 7); assert.equal(result.completed_parents, 4);
});

const defects = {
  empty: () => [],
  "package-only-pass": () => [event("start"), event("pass")],
  "parent-only-pass": () => [event("start"), ...parents.flatMap((p) => [event("run", p), event("pass", p)]), event("pass")],
  "missing-leaf": () => completeFixture().filter((r) => r.Test !== leaves[0]),
  "skipped-leaf": () => completeFixture().map((r) => r.Test === leaves[0] && r.Action === "pass" ? { ...r, Action: "skip" } : r),
  "failed-leaf": () => completeFixture().map((r) => r.Test === leaves[0] && r.Action === "pass" ? { ...r, Action: "fail" } : r),
  "duplicate-leaf": () => { const r = completeFixture(); r.splice(2, 0, event("run", leaves[0])); return r; },
  "unexpected-test": () => { const r = completeFixture(); r.splice(1, 0, event("run", "TestOther")); return r; },
  "malformed-json": () => ["{bad-json", ...completeFixture()],
  "invalid-event": () => [{}, ...completeFixture()],
  "invalid-test-type": () => [{ ...event("run"), Test: 7 }, ...completeFixture()],
  "invalid-output-type": () => [{ ...event("output"), Output: {} }, ...completeFixture()],
  "wrong-package": () => completeFixture().map((r) => ({ ...r, Package: "elsewhere" })),
  "pass-without-run": () => completeFixture().filter((r) => !(r.Test === leaves[0] && r.Action === "run")),
  "package-before-tests": () => [event("start"), event("pass"), ...completeFixture().slice(1)],
  "package-failure": () => [...completeFixture().slice(0, -1), event("fail")],
  "missing-package-start": () => completeFixture().slice(1),
};
for (const [name, fixture] of Object.entries(defects)) test(`rejects ${name} despite child exit zero`, () => assert.equal(validate(fixture()).valid, false));
test("rejects complete fixture with nonzero child, signal or changed source", () => {
  assert.equal(validate(completeFixture(), 1).valid, false);
  assert.equal(validate(completeFixture(), null).valid, false);
  assert.equal(validate(completeFixture(), 0, false).valid, false);
});
test("private output is never retained and harmless Make status cannot fill missing cases", () => {
  const privateRow = { ...event("output", parents[0]), Output: "secret-do-not-retain" };
  const report = validate([...completeFixture().slice(0, -1), privateRow, event("pass")]);
  assert.equal(report.valid, true); assert.equal(JSON.stringify(report).includes("secret-do-not-retain"), false);
  assert.equal(validate(["make status", event("start"), event("pass")]).valid, false);
});
