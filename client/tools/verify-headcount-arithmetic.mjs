import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import ts from "typescript";

// S2 forbids forming an unsafe intermediate product, even when a subsequent
// clamp would return the correct integer. Observe operators in the actual
// self-contained source; do not maintain a second copy of its algorithm.
const source = readFileSync(new URL("../src/headcount/allocation.ts", import.meta.url), "utf8");
const fixture = JSON.parse(readFileSync(new URL("../../testdata/headcount-allocation.json", import.meta.url), "utf8"));
const observerName = "__observeHeadcountProduct";

function instrument(input, multiply) {
  const parsed = ts.createSourceFile("allocation.ts", input, ts.ScriptTarget.ES2022, true);
  if (parsed.parseDiagnostics.length) throw new Error("headcount witness cannot parse source");
  let operators = 0;
  const output = ts.transpileModule(input, {
    compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
    transformers: { before: [(context) => {
      const visit = (node) => {
        if (ts.isImportDeclaration(node) || ts.isImportEqualsDeclaration(node)) {
          throw new Error("headcount witness requires self-contained arithmetic source");
        }
        if (ts.isIdentifier(node) && node.text === observerName) throw new Error("headcount witness observer name collision");
        if (ts.isBinaryExpression(node)) {
          if ([ts.SyntaxKind.AsteriskEqualsToken, ts.SyntaxKind.AsteriskAsteriskToken,
            ts.SyntaxKind.AsteriskAsteriskEqualsToken].includes(node.operatorToken.kind)) {
            throw new Error("headcount witness requires explicit coverage of this arithmetic operator");
          }
          if (node.operatorToken.kind === ts.SyntaxKind.AsteriskToken) {
            operators++;
            return ts.factory.createCallExpression(ts.factory.createIdentifier(observerName), undefined,
              [ts.visitNode(node.left, visit), ts.visitNode(node.right, visit)]);
          }
        }
        return ts.visitEachChild(node, visit, context);
      };
      return (file) => ts.visitNode(file, visit);
    }] },
  });
  assert.ok(operators > 0, "headcount witness observed no source multiplication operators");
  const exports = {};
  new Function("exports", observerName, output.outputText)(exports, multiply);
  assert.equal(typeof exports.project, "function");
  return exports.project;
}

let products = 0;
function checkedMultiply(left, right) {
  products++;
  assert.ok(Number.isSafeInteger(left) && Number.isSafeInteger(right), "headcount multiplication operands must be safe integers");
  const exact = BigInt(left) * BigInt(right);
  assert.ok(exact >= -BigInt(Number.MAX_SAFE_INTEGER) && exact <= BigInt(Number.MAX_SAFE_INTEGER),
    `unsafe headcount multiplication: ${left} * ${right}`);
  return left * right;
}

function invoke(project, row) {
  const roles = fixture.roles.map((role) => role.kind === "convert" ? { ...role, heads_per_converter: row.k } : { ...role });
  return project(fixture.hardcap, row.seats, roles, row.assigned);
}

const actual = instrument(source, checkedMultiply);
for (const row of fixture.effective_heads) {
  assert.deepEqual(JSON.parse(JSON.stringify(invoke(actual, row))), row.expected, row.name);
}
assert.ok(products > 0, "headcount witness executed no multiplication");
const actualProducts = products;

// This admitted shared vector demonstrates the algorithmic restriction, not an
// impossible JS overflow-output fault. The unobserved mutation still returns
// the right result; observing its multiplication must reject the same execution.
const edge = fixture.effective_heads.find((row) => row.k === 3 && row.assigned["role.converter"] === 3_100_000_000_000_000);
assert.ok(edge, "headcount unsafe-product control lost its admitted shared vector");
const guarded = "converters > Math.floor(heads / k) ? heads : converters * k";
assert.equal(source.split(guarded).length, 2, "headcount control anchor changed; update the mutation against actual source");
const naive = source.replace(guarded, "Math.min(converters * k, heads)");
assert.deepEqual(JSON.parse(JSON.stringify(invoke(instrument(naive, (left, right) => left * right), edge))), edge.expected);
assert.throws(() => invoke(instrument(naive, checkedMultiply), edge), /unsafe headcount multiplication/);

// Unsupported/new forms must not silently bypass observation.
assert.throws(() => instrument("export function project( {", checkedMultiply), /cannot parse source/);
assert.throws(() => instrument(`${source}\nlet value = 1; value *= 2;`, checkedMultiply), /explicit coverage/);
assert.throws(() => instrument(`${source}\nconst value = 2 ** 3;`, checkedMultiply), /explicit coverage/);
assert.throws(() => instrument(`import { helper } from './other';\n${source}`, checkedMultiply), /self-contained/);
assert.throws(() => instrument(source.replaceAll("converters * k", "converters"), checkedMultiply), /no source multiplication/);
console.log(`headcount intermediate-product witness: ${fixture.effective_heads.length} actual vectors, ${actualProducts} safe products; unsafe-product and five refusal controls passed`);
