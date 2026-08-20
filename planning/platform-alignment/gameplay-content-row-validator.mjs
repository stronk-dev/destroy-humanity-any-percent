import { readFileSync } from "node:fs";

const repositoryRoot = new URL("../../", import.meta.url);
const readTSV = (relative) => {
  const lines = readFileSync(new URL(relative, repositoryRoot), "utf8").trimEnd().split("\n");
  return { header: lines[0].split("\t"), rows: lines.slice(1).map((line) => line.split("\t")) };
};
const structure = readTSV("planning/platform-alignment/gameplay-content-row-structure.tsv");
const ledger = readTSV("planning/platform-alignment/gameplay-content-row-ledger.tsv");
const expectedHeader = [
  "unit_id", "artifact", "family", "json_pointer", "unit_kind", "authored_identity", "loader",
  "runtime_producer", "player_consumer", "current_reachability", "executable_witness",
  "failure_or_limit", "verdict", "authority_route",
];
const verdicts = new Set([
  "proven_mounted_effect", "proven_mounted_presentation", "partial_mounted", "backend_active",
  "backend_registered_dormant", "measurement_only", "uncomposed", "zero_or_empty_placeholder",
  "contradicted", "blocked",
]);

function validate(rows) {
  if (ledger.header.length !== expectedHeader.length || ledger.header.some((value, index) => value !== expectedHeader[index])) {
    throw new Error("content verdict header drift");
  }
  if (structure.rows.length !== 579 || rows.length !== structure.rows.length) throw new Error("content verdict denominator drift");
  if (new Set(rows.map((row) => row[0])).size !== rows.length) throw new Error("duplicate content verdict identity");
  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index];
    if (row.length !== expectedHeader.length || row.some((value) => value.length === 0)) throw new Error(`content verdict shape/empty field: ${row[0] ?? index}`);
    if (row.slice(0, 6).some((value, column) => value !== structure.rows[index][column])) throw new Error(`content structural identity drift: ${row[0]}`);
    if (!verdicts.has(row[12])) throw new Error(`unknown content verdict: ${row[0]} ${row[12]}`);
    if (row[12].startsWith("proven_mounted") && (row[8].includes("none") || !row[9].includes("mounted") || row[10].includes("none"))) {
      throw new Error(`vacuous mounted promotion: ${row[0]}`);
    }
    if (row[12] === "partial_mounted" && (row[8].includes("none") || !row[9].includes("mounted"))) throw new Error(`partial row lacks mounted path: ${row[0]}`);
    if (row[12] === "backend_active" && (row[9].includes("explicit deploy-current empty") || row[9].includes("value is exactly zero"))) {
      throw new Error(`empty/zero row relabeled active: ${row[0]}`);
    }
    if (row[12] === "backend_registered_dormant" && !row[9].includes("registered")) throw new Error(`dormant row lacks reachability state: ${row[0]}`);
    if (row[12] === "zero_or_empty_placeholder" && !row[9].includes("empty") && !row[9].includes("zero")) throw new Error(`zero/empty row lacks condition: ${row[0]}`);
  }
}

validate(ledger.rows);
const seededFailures = [];
const mustReject = (label, mutate) => {
  try {
    validate(mutate(ledger.rows.map((row) => [...row])));
  } catch {
    seededFailures.push(label);
    return;
  }
  throw new Error(`seeded failure accepted: ${label}`);
};
mustReject("dropped-unit", (rows) => rows.slice(1));
mustReject("duplicate-unit", (rows) => [rows[0], ...rows.slice(0, -1)]);
mustReject("backend-as-mounted", (rows) => {
  const row = rows.find((candidate) => candidate[12] === "backend_active");
  row[12] = "proven_mounted_effect";
  return rows;
});
mustReject("empty-as-active", (rows) => {
  const row = rows.find((candidate) => candidate[12] === "zero_or_empty_placeholder");
  row[12] = "backend_active";
  return rows;
});
mustReject("missing-route", (rows) => {
  rows[0][13] = "";
  return rows;
});

const counts = Object.fromEntries([...verdicts].map((verdict) => [verdict, 0]));
for (const row of ledger.rows) counts[row[12]] += 1;
console.log(JSON.stringify({ rows: ledger.rows.length, verdicts: counts, seededFailures }));
