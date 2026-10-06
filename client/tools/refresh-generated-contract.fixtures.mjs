// Manual compiler census; synthetic counterfactuals are never product contracts.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const root = fileURLToPath(new URL("../../", import.meta.url));
const generated = resolve(root, "client/src/api/generated/types.ts");
const virtual = resolve(root, "client/src/api/generated/__refresh_census_virtual__.ts");
const sourcePaths = ["client/src/api/generated/types.ts", "docs/generated/api.json",
  "docs/generated/api-compat-v1.json", "server/account/api.go", "server/publicapi/generate.go",
  "client/src/game-ui/runtime.ts", "client/src/game-ui/minigame/session-port.ts",
  "client/package.json", "client/pnpm-lock.yaml", "client/tools/refresh-generated-contract.fixtures.mjs"];
function identity() {
  return { head: execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim(),
    sha256: Object.fromEntries(sourcePaths.map((path) => [path,
      createHash("sha256").update(readFileSync(resolve(root, path))).digest("hex")])) };
}

function compileCaller(generatedSource, caller) {
  const options = { strict: true, noEmit: true, types: [], target: ts.ScriptTarget.ES2022,
    module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler };
  const host = ts.createCompilerHost(options);
  const originalGetSourceFile = host.getSourceFile.bind(host), originalExists = host.fileExists.bind(host);
  host.fileExists = (path) => resolve(path) === virtual || originalExists(path);
  host.getSourceFile = (path, language, onError, shouldCreate) => {
    if (resolve(path) === virtual) return ts.createSourceFile(path,
      `import { operations, type APIError, type BootstrapSession } from "./types";\n${caller}`, language, true);
    if (resolve(path) === generated) return ts.createSourceFile(path, generatedSource, language, true);
    return originalGetSourceFile(path, language, onError, shouldCreate);
  };
  const program = ts.createProgram([virtual], options, host);
  return ts.getPreEmitDiagnostics(program).map((diagnostic) => ({ code: diagnostic.code,
    caller: diagnostic.file && resolve(diagnostic.file.fileName) === virtual,
    message: ts.flattenDiagnosticMessageText(diagnostic.messageText, " ") }));
}

const callers = {
  registered_path: 'const path: typeof operations[keyof typeof operations]["path"] = "/api/v1/bootstrap";',
  registered_error: 'const error: APIError = { category: "invalid", detail: "body" };',
  bootstrap_pair_shape: 'const pair: BootstrapSession = { access_token: "synthetic-access", refresh_token: "synthetic-refresh" };',
  refresh_path: 'const path: typeof operations[keyof typeof operations]["path"] = "/api/v1/session/refresh";',
  refresh_token_error: 'const error: APIError = { category: "unauthorized", detail: "refresh_token" };',
  refresh_reuse_error: 'const error: APIError = { category: "refresh_reused", detail: "session_family_revoked" };',
};
function rejected(diagnostics) {
  assert.ok(diagnostics.length > 0, "expected current generated contract refusal");
  assert.ok(diagnostics.every((row) => row.caller && row.code === 2322), "refusal must be caller assignability, not incidental compilation failure");
}

const started = Date.now(), before = identity();
const source = readFileSync(generated, "utf8");
const baseline = Object.fromEntries(Object.entries(callers).map(([name, caller]) => [name, compileCaller(source, caller)]));
for (const name of ["registered_path", "registered_error", "bootstrap_pair_shape"]) assert.deepEqual(baseline[name], []);
for (const name of ["refresh_path", "refresh_token_error", "refresh_reuse_error"]) rejected(baseline[name]);

// Deliberately artificial, in-memory only. No operation ID, descriptor or policy adoption.
const pathCounterfactual = source.replace("export const operations = {",
  'export const operations = {\n  census_only_synthetic: { auth: "none", method: "POST", path: "/api/v1/session/refresh", pathParameters: [] },');
assert.notEqual(pathCounterfactual, source);
const errorCounterfactual = source.replace('category: "conflict" |', 'category: "refresh_reused" | "conflict" |')
  .replace('detail: "access_token" |', 'detail: "refresh_token" | "session_family_revoked" | "access_token" |');
assert.notEqual(errorCounterfactual, source);
const counterfactuals = {};
for (const [arm, text] of [["synthetic_path_only", pathCounterfactual], ["synthetic_error_only", errorCounterfactual]]) {
  counterfactuals[arm] = Object.fromEntries(["refresh_path", "refresh_token_error", "refresh_reuse_error"]
    .map((name) => [name, compileCaller(text, callers[name])]));
}
assert.deepEqual(counterfactuals.synthetic_path_only.refresh_path, []);
rejected(counterfactuals.synthetic_path_only.refresh_token_error);
rejected(counterfactuals.synthetic_path_only.refresh_reuse_error);
rejected(counterfactuals.synthetic_error_only.refresh_path);
assert.deepEqual(counterfactuals.synthetic_error_only.refresh_token_error, []);
assert.deepEqual(counterfactuals.synthetic_error_only.refresh_reuse_error, []);

const parsed = ts.createSourceFile(generated, source, ts.ScriptTarget.ES2022, true);
const declaration = parsed.statements.filter(ts.isVariableStatement).flatMap((row) => row.declarationList.declarations)
  .find((row) => row.name.getText(parsed) === "operations");
assert.ok(declaration && ts.isAsExpression(declaration.initializer) && ts.isObjectLiteralExpression(declaration.initializer.expression));
const operations = declaration.initializer.expression.properties.map((property) => {
  assert.ok(ts.isPropertyAssignment(property) && ts.isObjectLiteralExpression(property.initializer));
  const fields = Object.fromEntries(property.initializer.properties.filter(ts.isPropertyAssignment)
    .filter((row) => ts.isStringLiteral(row.initializer)).map((row) => [row.name.getText(parsed), row.initializer.text]));
  return { id: property.name.getText(parsed), method: fields.method, path: fields.path, auth: fields.auth };
});
const api = JSON.parse(readFileSync(resolve(root, "docs/generated/api.json"), "utf8"));
const openapi = Object.entries(api.paths).flatMap(([path, methods]) => Object.entries(methods)
  .filter(([method]) => ["get", "post", "delete", "put", "patch", "options", "head"].includes(method))
  .map(([method, row]) => ({ id: row.operationId, method: method.toUpperCase(), path })));
const normalize = (rows) => rows.map(({ id, method, path }) => ({ id, method, path })).sort((a, b) => a.id.localeCompare(b.id));
assert.deepEqual(normalize(operations), normalize(openapi));
assert.equal(operations.some((row) => row.path === "/api/v1/session/refresh"), false);

const callSites = [];
for (const path of ["client/src/game-ui/runtime.ts", "client/src/game-ui/minigame/session-port.ts"]) {
  const tree = ts.createSourceFile(path, readFileSync(resolve(root, path), "utf8"), ts.ScriptTarget.ES2022, true);
  function visit(node) {
    if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && ["fetch", "fetcher"].includes(node.expression.text)) {
      callSites.push({ path, line: tree.getLineAndCharacterOfPosition(node.getStart(tree)).line + 1,
        callee: node.expression.text, first_argument: node.arguments[0]?.getText(tree) });
    }
    ts.forEachChild(node, visit);
  }
  visit(tree);
}
assert.ok(callSites.some((row) => row.first_argument === '"/api/v1/intents"'));
const after = identity();
assert.deepEqual(before, after, "listed sources/HEAD changed during observation");
console.info(JSON.stringify({ schema_version: 1, started_at: new Date(started).toISOString(), elapsed_ms: Date.now() - started,
  node: process.version, typescript: ts.version, observation_valid: true, capability_complete: false,
  source_before: before, source_after: after, operations, compiler_baseline: baseline, synthetic_counterfactuals: counterfactuals,
  scoped_fetcher_call_sites: callSites, limits: ["No HTTP, database, successful rotation or browser run.",
    "Counterfactual strings exist in compiler memory only, not accepted operation/schema bytes.",
    "Call-site census covers only the two named TS files; not a complete AC4 lint.",
    "Source identity covers listed files/HEAD, not environment or arbitrary ignored files."] }, null, 2));
