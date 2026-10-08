import assert from "node:assert/strict";
import test from "node:test";
import { assertCosmeticN5Rejection, cosmeticN5Fixture, n5CleanupMarker, n5Faults, n5RejectionPrefix } from "./cosmetic-n5-fixture.mjs";

const shelf = "/fixture/client/src/game-ui/cosmetics/CosmeticShelf.svelte";
const source = "function buy(id: string): void { onAcquire(id); }";
function failure(fault) {
  return { code: 1, signal: null, stdout: `${n5CleanupMarker}\n`, stderr: `${n5RejectionPrefix}${JSON.stringify({
    phase: "after-native-buy", fixture: fault, fixtureExecutions: [fault],
    directViolations: [fault === "checkout" ? "fetch https://example.invalid/checkout" : "PaymentRequest"],
    violations: [], safetyBlockedRequests: [], pageErrors: [],
  })}\n` };
}

for (const fault of n5Faults) {
  test(`${fault}: injects the actual Buy callback and requires emitted fixture bytes`, () => {
    const plugin = cosmeticN5Fixture(fault);
    assert.equal(plugin.transform(source, "/fixture/client/src/main.ts"), null);
    const result = plugin.transform(source, shelf);
    assert.match(result.code, /onAcquire\(id\)/u);
    assert.ok(result.code.includes(`detail: "${fault}"`));
    assert.ok(result.code.includes(fault === "checkout" ? 'fetch("https://example.invalid/checkout")' : "new window.PaymentRequest()"));
    assert.throws(() => plugin.generateBundle({}, {}), /absent/u);
    plugin.generateBundle({}, { "entry.js": { type: "chunk", code: result.code } });
    assert.throws(() => plugin.transform(source, shelf), /more than once/u);
  });
  test(`${fault}: only accepts the exact pre-request rejection after native Buy`, () => {
    assert.equal(assertCosmeticN5Rejection(fault, failure(fault)).fixture, fault);
    for (const invalid of [
      { ...failure(fault), code: 0 }, { ...failure(fault), code: 2 },
      { ...failure(fault), signal: "SIGTERM" }, { ...failure(fault), stderr: "unrelated failure" },
      { ...failure(fault), stdout: "" }, { ...failure(fault), stdout: `${n5CleanupMarker}\n`.repeat(2) },
      { ...failure(fault), stderr: failure(fault).stderr.repeat(2) },
    ]) assert.throws(() => assertCosmeticN5Rejection(fault, invalid));
    for (const changed of [
      { phase: "startup" }, { fixture: "other" }, { fixtureExecutions: [] }, { fixtureExecutions: [fault, fault] },
      { directViolations: [] }, { directViolations: ["wrong violation"] },
      { violations: ["off-origin request"] }, { safetyBlockedRequests: ["https://example.invalid/checkout"] },
      { pageErrors: ["unrelated exception"] },
    ]) {
      const valid = failure(fault);
      const report = JSON.parse(valid.stderr.slice(n5RejectionPrefix.length));
      valid.stderr = `${n5RejectionPrefix}${JSON.stringify({ ...report, ...changed })}\n`;
      assert.throws(() => assertCosmeticN5Rejection(fault, valid));
    }
  });
}
test("refuses unknown fixtures, a missing shelf and changed/duplicate Buy anchors", () => {
  assert.throws(() => cosmeticN5Fixture("other"), /unknown/u);
  assert.throws(() => cosmeticN5Fixture("checkout").generateBundle({}, {}), /did not transform/u);
  assert.throws(() => cosmeticN5Fixture("checkout").transform("changed Buy callback", shelf), /exactly one/u);
  assert.throws(() => cosmeticN5Fixture("checkout").transform(source.repeat(2), shelf), /exactly one/u);
});
