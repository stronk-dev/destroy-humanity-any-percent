import { mkdtemp, mkdir, readFile, readdir, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "src", "combat");

function scriptKind(fileName) {
  if (fileName.endsWith(".tsx")) return ts.ScriptKind.TSX;
  if (fileName.endsWith(".jsx")) return ts.ScriptKind.JSX;
  if (fileName.endsWith(".js")) return ts.ScriptKind.JS;
  return ts.ScriptKind.TS;
}

export function hasDirectDivision(source, fileName = "combat.ts") {
  const sourceFile = ts.createSourceFile(fileName, source, ts.ScriptTarget.Latest, true, scriptKind(fileName));
  if (sourceFile.parseDiagnostics.length > 0) {
    const diagnostic = sourceFile.parseDiagnostics[0];
    throw new SyntaxError(
      `combat division guard could not parse ${fileName}: ${ts.flattenDiagnosticMessageText(diagnostic.messageText, "\n")}`,
    );
  }

  let found = false;
  function visit(node) {
    if (
      ts.isBinaryExpression(node) &&
      (node.operatorToken.kind === ts.SyntaxKind.SlashToken || node.operatorToken.kind === ts.SyntaxKind.SlashEqualsToken)
    ) {
      found = true;
      return;
    }
    ts.forEachChild(node, visit);
  }
  visit(sourceFile);
  return found;
}

async function typescriptFiles(directory) {
  const files = [];
  for (const entry of (await readdir(directory, { withFileTypes: true })).sort((left, right) => left.name.localeCompare(right.name))) {
    const absolute = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await typescriptFiles(absolute)));
    else if (entry.isFile() && /\.(?:js|jsx|ts|tsx|mts|cts)$/.test(entry.name)) files.push(absolute);
  }
  return files;
}

async function assertSelfTests() {
  const cases = [
    ["const invalid = left / right;", true],
    ["let invalid = left; invalid /= right;", true],
    ['const url = "https://example.test"; const invalid = left / right;', true],
    ["const invalid = `value=${left / right}`;", true],
    ["const label = `value=${left}`; const invalid = top / bottom;", true],
    ["const invalid = <output>{left / right}</output>;", true, "combat.tsx"],
    ['const safe = "left / right"; // left / right\n/* left / right */', false],
    ["const safe = /left \\/ right/g; const label = `value=${left}`;", false],
  ];
  for (const [source, expected, fileName] of cases) {
    if (hasDirectDivision(source, fileName) !== expected) throw new Error(`combat division AST self-test failed: ${source}`);
  }

  const fixture = await mkdtemp(path.join(tmpdir(), "cloud-clicker-combat-boundary-"));
  try {
    const nested = path.join(fixture, "duel", "engine");
    await mkdir(nested, { recursive: true });
    await writeFile(path.join(nested, "violation.mts"), "export const invalid = left / right;\n");
    const files = await typescriptFiles(fixture);
    if (files.length !== 1 || !hasDirectDivision(await readFile(files[0], "utf8"), files[0])) {
      throw new Error("combat division guard does not reject its recursive seeded violation");
    }
  } finally {
    await rm(fixture, { recursive: true, force: true });
  }
}

await assertSelfTests();
for (const file of await typescriptFiles(root)) {
  const source = await readFile(file, "utf8");
  if (hasDirectDivision(source, file)) {
    throw new Error(`${path.relative(path.dirname(root), file)}: native division is forbidden; use idiv`);
  }
}
