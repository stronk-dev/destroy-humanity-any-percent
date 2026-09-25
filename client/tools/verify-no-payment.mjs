// Cosmetic Shop v1 §10 N4: no real money, by construction. The client's
// runtime dependencies must equal the checked-in allowlist (the gate); a
// denylist scan of the lockfile and Go module files rejects known payment,
// in-app-purchase and ads SDK names (the secondary net). Both halves ship
// with failing fixtures that must be rejected on every run.
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "..");
const read = (relative) => readFileSync(path.join(root, relative), "utf8");

const DENYLIST = ["stripe", "paypal", "braintree", "adyen", "square", "paddle", "chargebee", "recurly", "revenuecat",
  "in-app-purchase", "inapp-purchase", "google-pay", "googlepay", "apple-pay", "applepay", "adsense", "admob", "doubleclick", "googleads"];

export function checkRuntimeDependencies(packageJSON, allowlist) {
  const actual = Object.keys(JSON.parse(packageJSON).dependencies ?? {}).sort();
  const expected = [...JSON.parse(allowlist).dependencies].sort();
  if (actual.length !== expected.length || actual.some((name, index) => name !== expected[index])) {
    throw new Error(`client runtime dependencies ${JSON.stringify(actual)} differ from the allowlist ${JSON.stringify(expected)}`);
  }
}

export function deniedPackage(text) {
  const lower = text.toLowerCase();
  for (const term of DENYLIST) {
    const escaped = term.replace(/[-]/gu, "[-_]?");
    if (new RegExp(`(?:^|[^a-z0-9])${escaped}(?:[^a-z0-9]|$)`, "mu").test(lower)) return term;
  }
  return null;
}

function expectFailure(label, operation) {
  try { operation(); } catch { return; }
  throw new Error(`${label} was not rejected`);
}

const packageJSON = read("client/package.json");
const allowlist = read("client/tools/runtime-dependency-allowlist.json");
checkRuntimeDependencies(packageJSON, allowlist);
for (const [label, source] of [["client/pnpm-lock.yaml", read("client/pnpm-lock.yaml")], ["server/go.mod", read("server/go.mod")], ["server/go.sum", read("server/go.sum")]]) {
  const term = deniedPackage(source);
  if (term) throw new Error(`${label} contains a denylisted payment/ads package: ${term}`);
}

// Failing fixtures: a runtime dependency added outside the allowlist, and a
// lockfile / go.mod carrying payment SDKs, must each be rejected.
const withExtra = JSON.stringify({ ...JSON.parse(packageJSON), dependencies: { ...JSON.parse(packageJSON).dependencies, "left-pad": "1.3.0" } });
expectFailure("extra runtime dependency fixture", () => checkRuntimeDependencies(withExtra, allowlist));
for (const fixture of ["  /stripe@12.0.0:\n    resolution: {integrity: sha512-x}", "  '@stripe/stripe-js@2.1.0':", "require github.com/stripe/stripe-go/v76 v76.0.0", "  /react-native-in-app-purchase@1.0.0:", "  /@paypal/paypal-js@8.0.0:"]) {
  if (!deniedPackage(fixture)) throw new Error(`payment denylist missed fixture ${fixture}`);
}
for (const nearMiss of ["  /squared-distance@1.0.0:", "  /stripes-css@1.0.0:"]) {
  if (deniedPackage(nearMiss)) throw new Error(`payment denylist overmatched ${nearMiss}`);
}
console.log("no-payment boundary ok: runtime dependencies equal the allowlist; lockfile and Go modules carry no payment/ads SDK; 6 negative fixtures rejected, 2 near-misses accepted");
