import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const client = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const shell = path.join(client, "src", "shell");
const files = (await fs.readdir(shell)).filter((name) => name.endsWith(".ts") || name.endsWith(".svelte"));
for (const name of files) {
  const source = await fs.readFile(path.join(shell, name), "utf8");
  if (/balance[/\\](?:mutation|writer)|mutateBalance|writeBalance/.test(source)) throw new Error(`${name}: shell may not import balance mutation paths`);
  if (/addEventListener\s*\(\s*["']unload["']/.test(source)) throw new Error(`${name}: unload lifecycle handler is forbidden`);
}
const intents = await fs.readFile(path.join(shell, "intents.ts"), "utf8");
if (/from\s+["'][^"']*(?:prediction|display|controller)/.test(intents)) throw new Error("intent dispatcher may not import predicted state");
console.log(`shell boundaries ok: ${files.length} source files`);
