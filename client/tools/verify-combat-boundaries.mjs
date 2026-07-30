import { mkdtemp, mkdir, readFile, readdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "src", "combat");

export function hasDirectDivision(source) {
  const scanner = ts.createScanner(ts.ScriptTarget.Latest, true, ts.LanguageVariant.Standard, source);
  for (let token = scanner.scan(); token !== ts.SyntaxKind.EndOfFileToken; token = scanner.scan()) {
    if (token === ts.SyntaxKind.SlashToken || token === ts.SyntaxKind.SlashEqualsToken) return true;
  }
  return false;
}

async function typescriptFiles(directory) {
  const files = [];
  for (const entry of (await readdir(directory, { withFileTypes: true })).sort((left, right) => left.name.localeCompare(right.name))) {
    const absolute = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await typescriptFiles(absolute)));
    else if (entry.isFile() && entry.name.endsWith(".ts")) files.push(absolute);
  }
  return files;
}

async function assertSelfTests() {
  const cases = [
    ["const invalid = left / right;", true],
    ['const url = "https://example.test"; const invalid = left / right;', true],
    ["const invalid = `value=${left / right}`;", true],
    ['const safe = "left / right"; // left / right\n/* left / right */', false],
  ];
  for (const [source, expected] of cases) {
    if (hasDirectDivision(source) !== expected) throw new Error(`combat division tokenizer self-test failed: ${source}`);
  }

  const fixture = await mkdtemp(path.join(tmpdir(), "cloud-clicker-combat-boundary-"));
  try {
    const nested = path.join(fixture, "duel", "engine");
    await mkdir(nested, { recursive: true });
    await writeFile(path.join(nested, "violation.ts"), "export const invalid = left / right;\n");
    const files = await typescriptFiles(fixture);
    if (files.length !== 1 || !hasDirectDivision(await readFile(files[0], "utf8"))) {
      throw new Error("combat division guard does not reject its recursive seeded violation");
    }
  } finally {
    await rm(fixture, { recursive: true, force: true });
  }
}

await assertSelfTests();
for (const file of await typescriptFiles(root)) {
  const source = await readFile(file, "utf8");
  if (hasDirectDivision(source)) {
    throw new Error(`${path.relative(path.dirname(root), file)}: native division is forbidden; use idiv`);
  }
}
