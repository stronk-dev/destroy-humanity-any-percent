import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import ts from "../../client/node_modules/typescript/lib/typescript.js";

const repositoryRoot = new URL("../../", import.meta.url);
const read = (relative) => readFileSync(new URL(relative, repositoryRoot), "utf8");
const digest = (value) => createHash("sha256").update(value).digest("hex");
const count = (value, pattern) => [...value.matchAll(pattern)].length;
const clean = (value) => value.replaceAll("\t", " ").replaceAll("\n", " ").replace(/\s+/g, " ").trim();
const lineAt = (source, offset) => source.slice(0, offset).split("\n").length;

const serverInventory = read("planning/platform-alignment/server-test-file-inventory.tsv").trimEnd().split("\n").slice(1)
  .map((line) => {
    const [file, packageName, testFunctions, fuzzFunctions, dependencySignal, structuralKind] = line.split("\t");
    return { file, packageName, testFunctions: Number(testFunctions), fuzzFunctions: Number(fuzzFunctions), dependencySignal, structuralKind };
  });
if (serverInventory.length !== 151 || serverInventory.reduce((sum, row) => sum + row.testFunctions + row.fuzzFunctions, 0) !== 592) {
  throw new Error("server oracle source population drift");
}
const clientInventory = read("planning/platform-alignment/client-test-artifact-inventory.tsv").trimEnd().split("\n").slice(1)
  .map((line) => line.split("\t"))
  .filter((row) => row[1] === "yes")
  .map(([file, , kind, subject, productionRelationship, evidenceLimit]) => ({ file, kind, subject, productionRelationship, evidenceLimit }));
if (clientInventory.length !== 43 || new Set(clientInventory.map(({ file }) => file)).size !== 43) throw new Error("client oracle source population drift");

function matchingBrace(source, open) {
  let depth = 0;
  let state = "code";
  let escaped = false;
  for (let index = open; index < source.length; index += 1) {
    const character = source[index];
    const next = source[index + 1];
    if (state === "line") {
      if (character === "\n") state = "code";
      continue;
    }
    if (state === "block") {
      if (character === "*" && next === "/") { state = "code"; index += 1; }
      continue;
    }
    if (state !== "code") {
      if (state !== "raw" && escaped) { escaped = false; continue; }
      if (state !== "raw" && character === "\\") { escaped = true; continue; }
      if ((state === "double" && character === '"') || (state === "single" && character === "'") || (state === "raw" && character === "`")) state = "code";
      continue;
    }
    if (character === "/" && next === "/") { state = "line"; index += 1; continue; }
    if (character === "/" && next === "*") { state = "block"; index += 1; continue; }
    if (character === '"') { state = "double"; continue; }
    if (character === "'") { state = "single"; continue; }
    if (character === "`") { state = "raw"; continue; }
    if (character === "{") depth += 1;
    else if (character === "}" && --depth === 0) return index;
  }
  throw new Error("unclosed Go function body");
}

function signals(body, runtime) {
  const expansion = runtime === "go"
    ? `subtests=${count(body, /\bt\.Run\s*\(/g)};fuzz_seeds=${count(body, /\bf\.Add\s*\(/g)}`
    : `each=${/\b(?:it|test)\.each\s*\(/.test(body) ? 1 : 0};loops=${count(body, /\b(?:for|while)\s*\(/g)}`;
  const dependencies = [
    /TEST_DATABASE_URL/.test(body) && "postgres_env",
    /\.(?:Skip|skip)\s*\(/.test(body) && "explicit_skip",
    /runtime\.GOARCH|process\.arch/.test(body) && "architecture",
    /testing\.Short|process\.env/.test(body) && "environment",
    /context\.WithTimeout|setTimeout|deadline|Deadline/.test(body) && "deadline",
    /guard|Guard|maxIterations|max_steps|maxSteps/.test(body) && "guard",
  ].filter(Boolean).join(",") || "none";
  const assertions = runtime === "go"
    ? `fatal=${count(body, /\.(?:Fatal|Fatalf)\s*\(/g)};error=${count(body, /\.(?:Error|Errorf)\s*\(/g)};panic=${count(body, /\bpanic\s*\(/g)};compare=${count(body, /(?:cmp\.|reflect\.DeepEqual|slices\.Equal)/g)}`
    : `expect=${count(body, /\bexpect\s*\(/g)};assert=${count(body, /\bassert\w*\s*\(/g)};throw=${count(body, /\bthrow\s+new\b/g)}`;
  const negativeTerms = [...new Set((body.match(/\b(?:reject|invalid|malformed|unknown|duplicate|missing|fail|error|panic|refus|forbid|stale|overflow|mismatch|mutation|negative|unauthori)\w*/gi) ?? []).map((value) => value.toLowerCase()))];
  return { expansion, dependencies, assertions, negative: negativeTerms.length === 0 ? "none" : negativeTerms.slice(0, 12).join(",") };
}

const units = [];
function emit({ runtime, file, declaration, offset, body, unitKind, subject, fileClass }) {
  const found = signals(body, runtime);
  units.push([
    `${runtime}:${file}:${declaration}:${offset}`,
    runtime,
    file,
    clean(declaration),
    String(lineAt(read(file), offset)),
    digest(body),
    unitKind,
    found.expansion,
    found.dependencies,
    found.assertions,
    found.negative,
    clean(subject),
    clean(fileClass),
  ]);
}

for (const inventory of serverInventory) {
  const source = read(inventory.file);
  const functions = [...source.matchAll(/^func\s+((?:Test|Fuzz)[A-Za-z0-9_]*)\s*\(/gm)];
  if (functions.length !== inventory.testFunctions + inventory.fuzzFunctions) throw new Error(`Go declaration count drift: ${inventory.file}`);
  for (const match of functions) {
    const open = source.indexOf("{", match.index);
    const close = matchingBrace(source, open);
    emit({ runtime: "go", file: inventory.file, declaration: match[1], offset: match.index, body: source.slice(open, close + 1), unitKind: match[1].startsWith("Fuzz") ? "fuzz" : "test", subject: inventory.packageName, fileClass: `${inventory.structuralKind}/${inventory.dependencySignal}` });
  }
}

function clientCallKind(expression) {
  if (ts.isIdentifier(expression) && (expression.text === "it" || expression.text === "test")) return "normal";
  if (ts.isPropertyAccessExpression(expression) && ts.isIdentifier(expression.expression) && (expression.expression.text === "it" || expression.expression.text === "test")) {
    return ["skip", "only", "fails", "todo", "concurrent"].includes(expression.name.text) ? expression.name.text : null;
  }
  if (ts.isCallExpression(expression) && ts.isPropertyAccessExpression(expression.expression)
      && ts.isIdentifier(expression.expression.expression) && (expression.expression.expression.text === "it" || expression.expression.expression.text === "test")
      && ["each", "skipIf", "runIf"].includes(expression.expression.name.text)) return expression.expression.name.text;
  return null;
}

for (const inventory of clientInventory) {
  const source = read(inventory.file);
  const sourceFile = ts.createSourceFile(inventory.file, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS);
  let declarations = 0;
  const visit = (node) => {
    if (ts.isCallExpression(node)) {
      const kind = clientCallKind(node.expression);
      if (kind !== null) {
        declarations += 1;
        const first = node.arguments[0];
        const name = first && (ts.isStringLiteral(first) || ts.isNoSubstitutionTemplateLiteral(first)) ? first.text : `<dynamic@${sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile)).line + 1}>`;
        const callback = [...node.arguments].reverse().find((argument) => ts.isArrowFunction(argument) || ts.isFunctionExpression(argument));
        const body = callback ? callback.getText(sourceFile) : node.getText(sourceFile);
        emit({ runtime: "client", file: inventory.file, declaration: `${kind}:${name}`, offset: node.getStart(sourceFile), body, unitKind: kind === "each" ? "parameterized_test" : kind === "normal" ? "test" : `test_${kind}`, subject: inventory.subject, fileClass: `${inventory.kind}/${inventory.productionRelationship}` });
      }
    }
    ts.forEachChild(node, visit);
  };
  visit(sourceFile);
  if (declarations === 0) emit({ runtime: "client", file: inventory.file, declaration: "<helper>", offset: 0, body: source, unitKind: "helper_not_oracle", subject: inventory.subject, fileClass: `${inventory.kind}/${inventory.productionRelationship}` });
}

const header = ["oracle_id", "runtime", "file", "declaration", "source_line", "body_sha256", "unit_kind", "expansion_signals", "dependency_signals", "assertion_signals", "negative_signals", "subject", "file_class"];
function validate(rows) {
  if (rows.length !== units.length || new Set(rows.map((row) => row[0])).size !== rows.length) throw new Error("oracle-unit denominator/identity drift");
  for (let index = 0; index < rows.length; index += 1) {
    if (rows[index].length !== header.length || rows[index].some((value, column) => value !== units[index][column])) throw new Error(`oracle-unit body/order drift: ${rows[index]?.[0] ?? index}`);
  }
  if (!rows.some((row) => row[6] === "helper_not_oracle")) throw new Error("zero-declaration helpers disappeared");
}
validate(units);
const seededFailures = [];
const mustReject = (label, rows) => { try { validate(rows); } catch { seededFailures.push(label); return; } throw new Error(`seeded failure accepted: ${label}`); };
mustReject("dropped-declaration", units.slice(1));
mustReject("duplicate-declaration", [units[0], ...units.slice(0, -1)]);
mustReject("helper-omission", units.filter((row, index) => index !== units.findIndex((candidate) => candidate[6] === "helper_not_oracle")));
mustReject("body-hash-drift", units.map((row, index) => index === 0 ? [...row.slice(0, 5), "0".repeat(64), ...row.slice(6)] : row));
console.log(header.join("\t"));
for (const row of units) console.log(row.join("\t"));
const summary = {};
for (const row of units) { const key = `${row[1]}:${row[6]}`; summary[key] = (summary[key] ?? 0) + 1; }
console.error(JSON.stringify({ serverFiles: serverInventory.length, serverFunctions: units.filter((row) => row[1] === "go").length, clientFiles: clientInventory.length, clientUnits: units.filter((row) => row[1] === "client").length, summary, seededFailures }));
