import { verifyCIKernelHistory } from "./ci-kernel-history.mjs";

const workflow = (checkout) => `jobs:\n  client:\n    steps:\n      - uses: actions/checkout@v6\n${checkout}      - run: make verify-client\n  browser:\n    steps: []\n`;
verifyCIKernelHistory(workflow("        with:\n          fetch-depth: 0\n"));

for (const invalid of [
  workflow(""),
  workflow("        with:\n          fetch-depth: 1\n"),
  `jobs:\n  client:\n    steps:\n      - uses: actions/checkout@v6\n        with:\n          fetch-depth: 0\n`,
]) {
  let rejected = false;
  try { verifyCIKernelHistory(invalid); } catch { rejected = true; }
  if (!rejected) throw new Error("CI history contract accepted an unsafe workflow fixture");
}
console.log("CI kernel-history adversarial fixtures ok");
