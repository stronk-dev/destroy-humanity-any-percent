import { readFileSync } from "node:fs";

const root = new URL("../../", import.meta.url);
const readTSV = (relative) => {
  const lines = readFileSync(new URL(relative, root), "utf8").trimEnd().split("\n");
  return { header: lines[0].split("\t"), rows: lines.slice(1).map((line) => line.split("\t")) };
};

const structural = readTSV("planning/platform-alignment/capability-atomization-inventory.tsv");
const outcomes = readTSV("planning/platform-alignment/capability-outcome-ledger.tsv");
const expectedHeader = [
  "capability_id", "parent_id", "design_ref", "user_outcome", "actor", "producer", "consumer",
  "current_data", "default_workflow", "executable_witness", "failure_or_refusal", "verdict",
  "authority_route", "evidence_limit",
];
const verdicts = new Set([
  "proven_integration", "proven_bounded_primitive", "partial_integration", "backend_or_data_only",
  "client_or_fixture_only", "claimed_only", "absent", "blocked",
]);

function validate(rows) {
  if (outcomes.header.length !== expectedHeader.length || outcomes.header.some((value, index) => value !== expectedHeader[index])) {
    throw new Error("outcome header drift");
  }
  if (structural.rows.length !== 433 || rows.length !== structural.rows.length) throw new Error("outcome denominator drift");
  if (new Set(rows.map((row) => row[0])).size !== rows.length) throw new Error("duplicate outcome identity");
  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index];
    if (row.length !== expectedHeader.length || row.some((value) => value.length === 0)) throw new Error(`row shape/empty field: ${row[0] ?? index}`);
    if (row.slice(0, 4).some((value, column) => value !== structural.rows[index][column])) throw new Error(`structural identity drift: ${row[0]}`);
    if (!verdicts.has(row[11])) throw new Error(`unknown verdict: ${row[0]} ${row[11]}`);
    if (row[11].startsWith("proven") && [row[5], row[6], row[7], row[9]].some((value) => value === "none")) {
      throw new Error(`vacuous proven row: ${row[0]}`);
    }
    if (!row[11].startsWith("proven") && row[12] === "none") throw new Error(`missing authority route: ${row[0]}`);
  }
}

validate(outcomes.rows);
const seededFailures = [];
const mustReject = (label, mutate) => {
  try {
    validate(mutate(outcomes.rows.map((row) => [...row])));
  } catch {
    seededFailures.push(label);
    return;
  }
  throw new Error(`seeded failure accepted: ${label}`);
};
mustReject("dropped-outcome", (rows) => rows.slice(1));
mustReject("duplicate-outcome", (rows) => [rows[0], ...rows.slice(0, -1)]);
mustReject("vacuous-promotion", (rows) => {
  const index = rows.findIndex((row) => row[0] === "M-002.01");
  rows[index][5] = "none";
  rows[index][11] = "proven_integration";
  return rows;
});
mustReject("missing-route", (rows) => {
  const index = rows.findIndex((row) => row[0] === "M-002.01");
  rows[index][12] = "";
  return rows;
});

const counts = Object.fromEntries([...verdicts].map((verdict) => [verdict, 0]));
for (const row of outcomes.rows) counts[row[11]] += 1;
console.log(JSON.stringify({ rows: outcomes.rows.length, verdicts: counts, seededFailures }));
