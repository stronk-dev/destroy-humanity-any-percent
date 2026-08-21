import { readFileSync } from "node:fs";

const repositoryRoot = new URL("../../", import.meta.url);
const readTSV = (relative) => {
  const lines = readFileSync(new URL(relative, repositoryRoot), "utf8").trimEnd().split("\n");
  return { header: lines[0].split("\t"), rows: lines.slice(1).map((line) => line.split("\t")) };
};
const structure = readTSV("planning/platform-alignment/test-oracle-row-structure.tsv");
const ledger = readTSV("planning/platform-alignment/test-oracle-row-ledger.tsv");
const expectedHeader = [
  "oracle_id", "runtime", "file", "declaration", "source_line", "body_sha256", "subject",
  "execution_lane", "fixture_or_data", "dependency_skip_guard", "assertion_oracle",
  "negative_control", "verdict", "authority_route", "evidence_limit",
];
const verdicts = new Set([
  "integrated_discriminating", "bounded_discriminating", "positive_only",
  "fixture_or_mock_only", "dependency_conditional", "non_discriminating",
  "invalid_or_guarded", "helper_not_oracle", "review_unresolved",
]);
const expectedCounts = {
  integrated_discriminating: 0,
  bounded_discriminating: 171,
  positive_only: 533,
  fixture_or_mock_only: 43,
  dependency_conditional: 51,
  non_discriminating: 1,
  invalid_or_guarded: 1,
  helper_not_oracle: 2,
  review_unresolved: 0,
};

function validate(rows) {
  if (ledger.header.length !== expectedHeader.length || ledger.header.some((value, index) => value !== expectedHeader[index])) {
    throw new Error("oracle verdict header drift");
  }
  if (structure.rows.length !== 802 || rows.length !== structure.rows.length) throw new Error("oracle verdict denominator drift");
  if (new Set(rows.map((row) => row[0])).size !== rows.length) throw new Error("duplicate oracle verdict identity");
  const counts = Object.fromEntries([...verdicts].map((verdict) => [verdict, 0]));
  for (let index = 0; index < rows.length; index += 1) {
    const row = rows[index];
    if (row.length !== expectedHeader.length || row.some((value) => value.length === 0)) throw new Error(`oracle verdict shape/empty field: ${row[0] ?? index}`);
    if (row.slice(0, 6).some((value, column) => value !== structure.rows[index][column])) throw new Error(`oracle structural identity drift: ${row[0]}`);
    if (!verdicts.has(row[12])) throw new Error(`unknown oracle verdict: ${row[0]} ${row[12]}`);
    counts[row[12]] += 1;
    if (row[12] === "integrated_discriminating" && (!row[8].includes("real") || row[11].includes("none") || row[9].includes("skip"))) {
      throw new Error(`vacuous integration promotion: ${row[0]}`);
    }
    if (row[12] === "dependency_conditional" && !row[9].includes("skip")) throw new Error(`conditional row hides skip: ${row[0]}`);
    if (row[12] === "helper_not_oracle" && !row[8].includes("helper")) throw new Error(`helper relabeled: ${row[0]}`);
  }
  for (const [verdict, expected] of Object.entries(expectedCounts)) {
    if (counts[verdict] !== expected) throw new Error(`oracle verdict count drift: ${verdict} ${counts[verdict]} != ${expected}`);
  }
  const axe = rows.find((row) => row[2] === "client/test/game-ui-screens-browser.test.ts" && row[3].includes("passes the C11 axe gate"));
  if (!axe || axe[12] !== "non_discriminating" || !axe[11].includes("stayed green")) throw new Error("fired Game UI oracle control disappeared");
  const api = rows.find((row) => row[2] === "server/publicapi/generate_test.go" && row[3] === "TestGenerateOpenAPIAndTypeScriptFromImmutableRegistry");
  if (!api || api[12] !== "bounded_discriminating" || !api[14].includes("hand-mounted")) throw new Error("API registry boundary was overpromoted");
  return counts;
}

const counts = validate(ledger.rows);
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
mustReject("dropped-oracle", (rows) => rows.slice(1));
mustReject("duplicate-oracle", (rows) => [rows[0], ...rows.slice(0, -1)]);
mustReject("body-hash-drift", (rows) => {
  rows[0][5] = "0".repeat(64);
  return rows;
});
mustReject("conditional-as-integrated", (rows) => {
  const row = rows.find((candidate) => candidate[12] === "dependency_conditional");
  row[12] = "integrated_discriminating";
  return rows;
});
mustReject("helper-as-positive", (rows) => {
  const row = rows.find((candidate) => candidate[12] === "helper_not_oracle");
  row[12] = "positive_only";
  return rows;
});
mustReject("blind-oracle-promotion", (rows) => {
  const row = rows.find((candidate) => candidate[12] === "non_discriminating");
  row[12] = "bounded_discriminating";
  return rows;
});
mustReject("missing-route", (rows) => {
  rows[0][13] = "";
  return rows;
});

console.log(JSON.stringify({ rows: ledger.rows.length, verdicts: counts, seededFailures }));
