// Cosmetic Shop v1 §5 / N3 / N8: the cosmetic packages are mechanically
// isolated. server/cosmetic may depend on the Go standard library only (no
// ledger, multiplier, production, faction, fiscal, achievement, Soul, pet
// care, persistence, or metrics/telemetry package, transitively), and
// client/src/cosmetic may import only within its own package. Each rule ships with
// fixtures that must be rejected.
import { execFileSync } from "node:child_process";
import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const clientDirectory = path.join(root, "client/src/cosmetic");

export function assertGoDependencies(dependencies, label) {
  const forbidden = dependencies.filter((value) => value !== "cloud-clicker/server/cosmetic" && !value.startsWith("vendor/") && (value.includes(".") || value.startsWith("cloud-clicker/")));
  if (forbidden.length > 0) throw new Error(`${label}: cosmetic may depend on the standard library only, not ${forbidden.join(", ")}`);
}

export function assertTypeScriptImports(source, label, fileName = path.join(clientDirectory, "fixture.ts")) {
  const kind = /\.[jt]sx$/u.test(fileName) ? ts.ScriptKind.TSX : ts.ScriptKind.TS;
  const parsed = ts.createSourceFile(fileName, source, ts.ScriptTarget.Latest, true, kind);
  if (parsed.parseDiagnostics.length > 0) {
    throw new SyntaxError(`${label}: cannot parse isolation subject: ${ts.flattenDiagnosticMessageText(parsed.parseDiagnostics[0].messageText, "\n")}`);
  }
  const check = (specifier) => {
    if (!specifier || !ts.isStringLiteralLike(specifier)) throw new Error(`${label}: computed cosmetic imports are forbidden`);
    const value = specifier.text;
    const resolved = path.resolve(path.dirname(fileName), value);
    const relative = path.relative(clientDirectory, resolved);
    if ((!value.startsWith("./") && !value.startsWith("../")) || value.includes("\\") ||
      relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) {
      throw new Error(`${label}: cosmetic may import only its own package, not ${value}`);
    }
  };
  const visit = (node) => {
    if (ts.isImportDeclaration(node)) check(node.moduleSpecifier);
    else if (ts.isExportDeclaration(node) && node.moduleSpecifier) check(node.moduleSpecifier);
    else if (ts.isImportEqualsDeclaration(node) && ts.isExternalModuleReference(node.moduleReference)) check(node.moduleReference.expression);
    else if (ts.isImportTypeNode(node)) check(ts.isLiteralTypeNode(node.argument) ? node.argument.literal : undefined);
    else if (ts.isCallExpression(node) && (node.expression.kind === ts.SyntaxKind.ImportKeyword ||
      (ts.isIdentifier(node.expression) && node.expression.text === "require") ||
      (ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === "require"))) check(node.arguments[0]);
    ts.forEachChild(node, visit);
  };
  visit(parsed);
}

function scriptFiles(directory) {
  const result = [];
  for (const entry of readdirSync(directory, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
    const absolute = path.join(directory, entry.name);
    if (entry.isSymbolicLink()) throw new Error(`${absolute}: isolation subjects must not be symlinks`);
    if (entry.isDirectory()) result.push(...scriptFiles(absolute));
    else if (entry.isFile() && /\.(?:[cm]?[jt]s|[jt]sx)$/u.test(entry.name)) result.push(absolute);
  }
  return result;
}

function expectRejected(label, operation) {
  try { operation(); } catch { return; }
  throw new Error(`${label} fixture was not rejected`);
}

for (const fixture of [["cloud-clicker/server/economy"], ["cloud-clicker/server/production"], ["cloud-clicker/server/save"],
  ["github.com/prometheus/client_golang/prometheus"], ["cloud-clicker/server/operations"]]) {
  expectRejected(`Go dependency ${fixture[0]}`, () => assertGoDependencies(["fmt", "cloud-clicker/server/cosmetic", ...fixture], "fixture"));
}
const rejectedImports = ['import { applyLogged } from "../replay";', 'import { parseCatalog } from "../economy-kernel";', 'import { metric } from "../game-ui/runtime";', 'import telemetry from "some-telemetry";',
  'import { parseCatalog } from "./../economy-kernel";', 'export function leak() { return import("../economy-kernel"); }',
  'export * from "../replay";', 'import replay = require("../replay");', 'const replay = require("../replay");',
  'const replay = module.require("../replay");', 'type Replay = import("../replay").Replay;',
  'const replay = import(target);', 'const replay = import(`../${target}`);',
  'import {', 'import "./nested/../../replay";'];
for (const fixture of rejectedImports) {
  expectRejected(`TS import ${fixture}`, () => assertTypeScriptImports(fixture, "fixture"));
}
assertTypeScriptImports('import { cosmeticItem } from "./catalog";', "neutral fixture");
assertTypeScriptImports('export * from "./catalog"; const local = import("./catalog");', "local exports and dynamic imports");
assertTypeScriptImports('const text = "import(../replay)"; // import("../replay")\n/* import("../replay") */', "non-code import text");
assertTypeScriptImports('export * from "../catalog";', "nested package-local import", path.join(clientDirectory, "nested/module.ts"));
expectRejected("nested package escape", () => assertTypeScriptImports('export * from "../../replay";', "nested fixture", path.join(clientDirectory, "nested/module.ts")));

const fixtureDirectory = path.join(root, "client/tools/testdata/cosmetic-boundary");
const nestedFixtures = scriptFiles(fixtureDirectory);
if (nestedFixtures.length !== 2) throw new Error("cosmetic recursive isolation fixtures are missing");
for (const file of nestedFixtures) {
  const relative = path.relative(fixtureDirectory, file);
  const check = () => assertTypeScriptImports(readFileSync(file, "utf8"), `recursive fixture ${relative}`, path.join(clientDirectory, relative));
  if (path.basename(file) === "reject.mts") expectRejected(relative, check);
  else check();
}

const goEnvironment = { ...process.env, GOCACHE: process.env.GOCACHE ?? path.join(root, ".cache", "go-build") };
const dependencies = execFileSync("go", ["list", "-deps", "./cosmetic"], { cwd: path.join(root, "server"), encoding: "utf8", env: goEnvironment }).trim().split("\n");
assertGoDependencies(dependencies, "server/cosmetic");
const files = scriptFiles(clientDirectory);
if (files.length === 0) throw new Error("cosmetic isolation population is empty");
for (const file of files) {
  assertTypeScriptImports(readFileSync(file, "utf8"), path.relative(root, file), file);
}
console.log(`cosmetic package boundary ok: server/cosmetic is stdlib-only (${dependencies.length} deps); ${files.length} client files import within their package; ${5 + rejectedImports.length + 2} negative fixtures rejected`);
