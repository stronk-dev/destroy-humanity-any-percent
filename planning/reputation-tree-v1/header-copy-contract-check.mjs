// R9 explicit header parameter census. Metadata only: never prose, units,
// owner adoption or full formula acceptance. Default checks the actual repo.
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../", import.meta.url));
const rfc = readFileSync(`${root}rfc/reputation-tree-v1.md`, "utf8");
const catalog = JSON.parse(readFileSync(`${root}copy/catalog/reputation-candidate.json`, "utf8"));
const suffixes = ["balance.available", "balance.level", "balance.spent", "bonus.this_run", "bonus.next_run", "bonus.formula"];
const start = rfc.indexOf("**Copy keys**");
const end = rfc.indexOf("### R10", start);
if (start < 0 || end <= start || catalog.schema_version !== 1 || !Array.isArray(catalog.entries)) throw new Error("invalid census authority/input");
const block = rfc.slice(start, end).replace(/\s+/g, " ");
const expected = suffixes.map((suffix) => {
  const key = `reputation_tree.${suffix}`;
  const match = block.match(new RegExp(`\\.${suffix.replaceAll(".", "\\.")}\\s*\\{([^}]*)\\}`));
  if (!match) throw new Error(`missing authoritative definition: ${key}`);
  const params = match[1].split(",").map((value) => {
    const [name, type, extra] = value.trim().split(":");
    if (!name || !["integer", "canonical_decimal", "string"].includes(type) || extra !== undefined) throw new Error(`invalid parameter authority: ${key}`);
    return { name, type };
  });
  return { key, params };
});
function census(entries) {
  return expected.map(({ key, params }) => {
    const matches = entries.filter((entry) => entry.key === key);
    const actual = matches.length === 1 ? matches[0].params : null;
    return { key, expected: params, actual, pass: JSON.stringify(params) === JSON.stringify(actual) };
  });
}
if (process.argv.length > 3 || process.argv[2] !== undefined && process.argv[2] !== "--self-test") throw new Error("only --self-test is permitted");
if (process.argv[2] === "--self-test") {
  // Synthetic metadata alignment, explicitly NOT the current copy catalog.
  const aligned = structuredClone(expected.map(({ key, params }) => ({ key, params })));
  if (census(aligned).some((row) => !row.pass)) throw new Error("synthetic aligned control failed");
  const key = "reputation_tree.bonus.formula";
  for (const fault of ["wrong-name", "wrong-type", "missing-key", "duplicate-key"]) {
    let entries = structuredClone(aligned);
    const formula = entries.find((entry) => entry.key === key);
    if (fault === "wrong-name") formula.params[0].name = "perlevel";
    if (fault === "wrong-type") formula.params[0].type = "string";
    if (fault === "missing-key") entries = entries.filter((entry) => entry.key !== key);
    if (fault === "duplicate-key") entries.push(structuredClone(formula));
    const failures = census(entries).filter((row) => !row.pass);
    if (failures.length !== 1 || failures[0].key !== key) throw new Error(`fault did not discriminate: ${fault}`);
  }
  console.log(JSON.stringify({ scope: "SELF TEST ONLY: synthetic metadata, not actual catalog or prose/unit approval", aligned: 6, rejected: 4 }));
} else {
  const rows = census(catalog.entries);
  console.log(JSON.stringify({ scope: "R9 explicit header parameter metadata only; not prose, units, adoption or full formula acceptance", rows,
    passed: rows.filter((row) => row.pass).length, failed: rows.filter((row) => !row.pass).length }, null, 2));
  process.exitCode = rows.some((row) => !row.pass) ? 1 : 0;
}
