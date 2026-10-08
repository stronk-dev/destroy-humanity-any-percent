import { spawn } from "node:child_process";
import { fileURLToPath } from "node:url";
import { assertCosmeticN5Rejection, n5Faults } from "./cosmetic-n5-fixture.mjs";

const root = fileURLToPath(new URL("../../", import.meta.url));
for (const fault of n5Faults) {
  const result = await new Promise((resolve, reject) => {
    const child = spawn(process.execPath, ["client/tools/test-cosmetic-composed.mjs", `--n5-fault=${fault}`],
      { cwd: root, stdio: ["ignore", "pipe", "pipe"] });
    let stdout = "", stderr = "";
    child.stdout.on("data", (data) => { stdout += data; });
    child.stderr.on("data", (data) => { stderr += data; });
    child.on("error", reject);
    child.on("close", (code, signal) => resolve({ code, signal, stdout, stderr }));
  });
  try {
    const report = assertCosmeticN5Rejection(fault, result);
    console.log(`composed Cosmetic AC13 N5 negative: ${JSON.stringify({ childExit: result.code, ...report })}: PASS`);
  } catch (error) {
    process.stdout.write(result.stdout);
    process.stderr.write(result.stderr);
    throw error;
  }
}
