import assert from "node:assert/strict";

export const n5Faults = Object.freeze(["checkout", "payment-request"]);
export const n5RejectionPrefix = "COSMETIC_N5_REJECTED ";
export const n5CleanupMarker = "COSMETIC_N5_CLEANUP_COMPLETE";
const marker = "cosmetic-n5-fixture";
const buyAnchor = "function buy(id: string): void {";

// Build-time fixture only: change the actual shelf's native Buy callback in
// an isolated Vite output, never repository source or a released build.
export function cosmeticN5Fixture(fault) {
  assert.ok(n5Faults.includes(fault), `unknown cosmetic N5 fixture: ${fault}`);
  let transformed = 0;
  const attempt = fault === "checkout"
    ? 'void fetch("https://example.invalid/checkout").catch(() => {});'
    : "new window.PaymentRequest();";
  return {
    name: `cosmetic-n5-${fault}-fixture`,
    enforce: "pre",
    transform(source, id) {
      if (!id.replaceAll("\\", "/").endsWith("/src/game-ui/cosmetics/CosmeticShelf.svelte")) return null;
      assert.equal(source.split(buyAnchor).length, 2, "N5 fixture requires exactly one actual shelf Buy callback");
      transformed += 1;
      assert.equal(transformed, 1, "N5 fixture transformed the shelf more than once");
      return { code: source.replace(buyAnchor, `${buyAnchor}
        window.dispatchEvent(new CustomEvent(${JSON.stringify(marker)}, { detail: ${JSON.stringify(fault)} }));
        try { ${attempt} } catch { /* keep the real Buy journey running */ }
      `), map: null };
    },
    generateBundle(_options, bundle) {
      assert.equal(transformed, 1, "N5 fixture did not transform the actual shelf");
      assert.ok(Object.values(bundle).some((file) => file.type === "chunk" && file.code.includes(marker) &&
        file.code.includes(fault === "checkout" ? "example.invalid/checkout" : "PaymentRequest")),
      "N5 fixture attempt is absent from the emitted production bundle");
    },
  };
}

// Expected failure must be THIS gate rejecting THIS built/native Buy attempt.
// A green child, unrelated exception, missing injection or secondary egress
// abort must never turn the negative population green.
export function assertCosmeticN5Rejection(fault, { code, signal, stdout, stderr }) {
  assert.ok(n5Faults.includes(fault), `unknown cosmetic N5 fixture: ${fault}`);
  assert.equal(code, 1, "N5 negative build must exit 1");
  assert.equal(signal, null, "N5 negative build must exit normally, not be killed");
  const lines = `${stdout}\n${stderr}`.split(/\r?\n/u);
  assert.equal(lines.filter((line) => line === n5CleanupMarker).length, 1, "N5 negative build must complete fixture/server cleanup");
  const reports = lines.filter((line) => line.startsWith(n5RejectionPrefix));
  assert.equal(reports.length, 1, "N5 negative build must report exactly one gate rejection");
  const report = JSON.parse(reports[0].slice(n5RejectionPrefix.length));
  assert.equal(report.phase, "after-native-buy");
  assert.equal(report.fixture, fault);
  assert.deepEqual(report.fixtureExecutions, [fault], "the actual built Buy callback must execute once");
  assert.deepEqual(report.directViolations, [fault === "checkout" ? "fetch https://example.invalid/checkout" : "PaymentRequest"]);
  assert.deepEqual(report.violations, [], "forbidden attempts must be trapped before creating a browser request");
  assert.deepEqual(report.safetyBlockedRequests, [], "secondary egress abort is not primary trap evidence");
  assert.deepEqual(report.pageErrors, [], "an unrelated page exception is not N5 evidence");
  return report;
}
