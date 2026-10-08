// Diagnostic transcription of accepted TL4, not a production ticker selector.
// Default: assert TL-AC4 and exit 1 on its counterexample.
// --self-test: verify the diagnostic controls and the retained contradiction.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";

assert.ok(process.argv.slice(2).every((arg) => arg === "--self-test"), "unknown argument");
const selfTest = process.argv.includes("--self-test");
const era = "era_1995";
const founder = "01986666-0000-7000-8000-000000000001";
const run = 1;
const start = 64;
const length = 64;
const noRepeat = 12;
const makeLine = (suffix, category, npc = null) => ({
  line_id: `fixture.ticker.${suffix}`, copy_key: `fixture.ticker.${suffix}`,
  category, eras: [era], weight: 10, conditions: [], params: {}, npc_id: npc, claim_ids: [],
});
const byteCompare = (left, right) => Buffer.compare(Buffer.from(left), Buffer.from(right));
const corpus = {
  schema_version: 1, corpus_id: "launch", slot_ms: 30_000, window_slots: length,
  no_repeat_window: noRepeat, refetch_min_interval_ms: 120_000,
  world: { interval_ms: 300_000, ttl_ms: 900_000, npc_population_floor: 10 },
  presentation: { scroll_px_per_second: 60, max_template_chars: 140 },
  eras: [era], tier_era: [{ tier: 0, era }],
  npc_companies: [{ npc_id: "npc_a", name_copy_key: "fixture.ticker.npc_a_name" }],
  lines: [
    ...Array.from({ length: 14 }, (_, index) => makeLine(`ambient_${String(index).padStart(2, "0")}`, "era_ambient")),
    makeLine("npc_a", "npc_company", "npc_a"),
    ...Array.from({ length: 3 }, (_, index) => makeLine(`offline_${index}`, "offline_fallback")),
  ].sort((left, right) => byteCompare(left.line_id, right.line_id)),
};

// Same recursive byte-sorted object-key grammar as canonicalJSONString/CanonicalJSON.
function canonical(value) {
  if (value === null || typeof value !== "object") return JSON.stringify(value);
  if (Array.isArray(value)) return `[${value.map(canonical).join(",")}]`;
  return `{${Object.keys(value).sort(byteCompare).map((key) => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}`;
}
const corpusHash = `sha256:${createHash("sha256").update(canonical(corpus)).digest("hex")}`;
const eligible = corpus.lines.filter((line) => line.category === "era_ambient");

function randomUint64(parts) {
  return createHash("sha256").update(`cloud-clicker/ticker/v1\0${parts.join("\0")}`).digest().readBigUInt64BE();
}

function pick(candidates, random) {
  const sorted = [...candidates].sort((left, right) => byteCompare(left.line_id, right.line_id));
  let remaining = random % BigInt(sorted.reduce((sum, line) => sum + line.weight, 0));
  for (const line of sorted) {
    if (remaining < BigInt(line.weight)) return line;
    remaining -= BigInt(line.weight);
  }
  throw new Error("empty or invalid pick population");
}

function personalWindow(s0, lines = eligible) {
  const recent = [];
  const out = [];
  for (let slot = s0 - noRepeat; slot < s0 + length; slot += 1) {
    let candidates = lines.filter((line) => !recent.includes(line.line_id));
    if (candidates.length === 0) candidates = lines.filter((line) => line.line_id !== recent.at(-1));
    if (candidates.length === 0) candidates = lines;
    const line = pick(candidates, randomUint64([corpusHash, "personal", founder, String(run), String(slot)]));
    recent.push(line.line_id);
    if (recent.length > noRepeat) recent.shift();
    if (slot >= s0) out.push({ slot, line_id: line.line_id });
  }
  return out;
}

function overlapDifferences(first, second) {
  const indexed = new Map(first.map((row) => [row.slot, row.line_id]));
  return second.filter((row) => indexed.has(row.slot) && indexed.get(row.slot) !== row.line_id)
    .map((row) => ({ slot: row.slot, first: indexed.get(row.slot), second: row.line_id }));
}

function repeatsWithinWindow(rows) {
  return rows.filter((row, index) => rows.slice(Math.max(0, index - noRepeat), index)
    .some((previous) => previous.line_id === row.line_id));
}

// Controls: identical request repeats exactly; input order cannot change byte-sorted picks;
// the same-slot comparator accepts equal records and detects one changed record.
const first = personalWindow(start);
const second = personalWindow(start + 1);
assert.deepEqual(first, personalWindow(start));
assert.deepEqual(first, personalWindow(start, [...eligible].reverse()));
assert.deepEqual(overlapDifferences(first, first), []);
assert.equal(overlapDifferences(first, first.map((row, index) => index === 0
  ? { ...row, line_id: "fixture.ticker.changed_control" } : row)).length, 1);
assert.equal(first.length, length);
assert.equal(second.length, length);
assert.equal(eligible.length, 14);
assert.equal(repeatsWithinWindow(first).length, 0);
assert.equal(repeatsWithinWindow(second).length, 0);
const differences = overlapDifferences(first, second);
const result = {
  subject: "RFC ticker-launch-corpus TL4 / TL-AC4", mode: selfTest ? "diagnostic-self-test" : "assert-contract",
  corpus_hash: corpusHash, founder, run, no_repeat_window: noRepeat, window_slots: length,
  starts: [start, start + 1], overlapping_slots: length - 1,
  mismatching_slots: differences.length, first_mismatches: differences.slice(0, 5),
  repeats_within_each_window: 0, controls_passed: true,
  boundary: "literal selector transcription; synthetic unadopted corpus, not a production loader or player journey",
};
console.log(JSON.stringify(result, null, 2));
if (selfTest) {
  assert.equal(differences.length, 55, "retained contradiction changed; inspect the transcription and RFC");
} else {
  assert.equal(differences.length, 0, "TL4 warm-up does not guarantee identical overlapping slots");
}
