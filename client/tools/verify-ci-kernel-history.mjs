import { readFileSync } from "node:fs";
import { verifyCIKernelHistory } from "./ci-kernel-history.mjs";

verifyCIKernelHistory(readFileSync(new URL("../../.github/workflows/ci.yml", import.meta.url), "utf8"));
console.log("CI kernel-history checkout contract ok");
