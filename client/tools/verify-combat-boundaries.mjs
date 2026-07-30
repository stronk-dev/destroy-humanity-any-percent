import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "src", "combat");

function strips(source) {
  return source
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/\/\/.*$/gm, "")
    .replace(/"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`(?:\\.|[^`\\])*`/g, "");
}

export function hasDirectDivision(source) {
  return /(^|[^/])\/(?![/*])/m.test(strips(source));
}

if (!hasDirectDivision("const invalid = left / right;")) throw new Error("combat division guard does not reject its seeded violation");
for (const entry of (await readdir(root, { withFileTypes: true })).sort((left, right) => left.name.localeCompare(right.name))) {
  if (!entry.isFile() || !entry.name.endsWith(".ts")) continue;
  const source = await readFile(path.join(root, entry.name), "utf8");
  if (hasDirectDivision(source)) throw new Error(`src/combat/${entry.name}: native division is forbidden; use idiv`);
}
