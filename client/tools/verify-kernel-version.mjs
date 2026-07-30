import { readFileSync } from "node:fs";

const source = readFileSync(new URL("../../kernel/VERSION", import.meta.url), "utf8").trim();
const go = readFileSync(new URL("../../server/kernel/version.go", import.meta.url), "utf8");
const client = readFileSync(new URL("../src/kernel/version.ts", import.meta.url), "utf8");
const goVersion = go.match(/const Version = "([^"]+)"/)?.[1];
const clientVersion = client.match(/KERNEL_VERSION = "([^"]+)"/)?.[1];

if (!source || source !== goVersion || source !== clientVersion) {
  throw new Error(`kernel version drift: source=${source} go=${goVersion} client=${clientVersion}`);
}

console.log(`kernel version parity ok: ${source}`);
