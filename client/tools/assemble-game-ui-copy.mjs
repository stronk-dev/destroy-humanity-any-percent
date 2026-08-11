import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

import { containsStatistic, repositoryRoot, validateLongformCopyText, validatePlainCopyText, validateProvenanceRegistry } from "./copy-pipeline.mjs";

const check = process.argv.slice(2).includes("--check");
const rulingPath = path.join(repositoryRoot, "planning/t0-t1-content/screen-copy-ruling-v1.md");
const candidatePath = path.join(repositoryRoot, "copy/catalog/game-ui-candidate.json");
const phase0Path = path.join(repositoryRoot, "copy/catalog/phase0.json");
const presentationV1Path = path.join(repositoryRoot, "balance/testdata/t0-t1/presentation-v1.json");
const presentationV2Path = path.join(repositoryRoot, "balance/testdata/t0-t1/presentation-v2.json");
const presentationClientPath = path.join(repositoryRoot, "client/src/game-ui/presentation.generated.json");
const eventCopyV1Path = path.join(repositoryRoot, "balance/testdata/t0-t1/event-copy-v1.json");
const eventCopyV2Path = path.join(repositoryRoot, "balance/testdata/t0-t1/event-copy-v2.json");

const cashReplacement = Object.freeze({
  key: "resource.company_cash.cap.phase0",
  text: "Cash is capped. The cap is a number, the number is visible, and nothing will ever sell you the difference.",
  params: [],
  era_variants: null,
  provenance: [],
  tone: "diegetic",
});

const existingPermits = Object.freeze({
  key: "resource.company_permits.cap.phase0",
  text: "The county issues at most 24 concurrent permits. The county does not care about your roadmap.",
  params: [],
  era_variants: null,
  provenance: [],
  tone: "corporate",
});

const additions = Object.freeze([
  ["event.exit_offer_resolved", "Offer accepted. The rest is paperwork.", "corporate"],
  ["exit_type.acquihire.title", "Acqui-hired", "corporate"],
  ["exit_type.acquisition.title", "Acquired", "corporate"],
  ["exit_type.collapse.title", "Collapsed", "corporate"],
  ["exit_type.ipo.title", "Went Public", "corporate"],
  ["exit_type.scripted_first.title", "First Failure", "corporate"],
  ["gate.t0_to_t1.title", "Garage", "diegetic"],
]);

const paramTypes = Object.freeze({
  ago: "string",
  amount: "string",
  attended: "string",
  attended_ms: "integer",
  cap: "integer",
  cap_hours: "integer",
  cost: "canonical_decimal",
  count: "integer",
  current: "integer",
  day: "integer",
  exit_type: "string",
  expires_at_ms: "integer",
  founder: "string",
  gate_id: "string",
  generator_id: "string",
  pb: "string",
  price: "string",
  rate: "string",
  rate_percent: "integer",
  remaining: "string",
  rta: "string",
  run_seq: "integer",
  tier: "integer",
  upgrade_id: "string",
});

function fail(message) {
  throw new Error(`game-ui copy assembly: ${message}`);
}

function bytes(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function sha256(value) {
  return `sha256:${createHash("sha256").update(value).digest("hex")}`;
}

function output(filename, value) {
  if (check) {
    let actual;
    try { actual = readFileSync(filename, "utf8"); } catch { fail(`missing ${path.relative(repositoryRoot, filename)} (run make game-ui-copy-candidate)`); }
    if (actual !== value) fail(`${path.relative(repositoryRoot, filename)} drift (run make game-ui-copy-candidate)`);
    return;
  }
  writeFileSync(filename, value, "utf8");
}

function blockText(block, key) {
  const fenced = block.match(/Text:\s*\n  ```\n([\s\S]*?)\n  ```/u);
  if (fenced) return fenced[1].split("\n").map((line) => line.startsWith("  ") ? line.slice(2) : line).join("\n");
  const inline = block.match(/Text: "([^"\n]*)"/u);
  if (!inline) fail(`missing Text for ${key}`);
  return inline[1];
}

function extractRows(markdown) {
  const section = markdown.match(/# Phase-A screen-copy set[\s\S]*?\n---\n\n## Summary/u)?.[0];
  if (!section) fail("could not locate the ruled copy section");
  const matches = [...section.matchAll(/^- `([^`]+)` —([^\n]*)(?:\n(?!- `)[\s\S]*?)?(?=\n- `|\n## |\n### |\n---)/gmu)];
  if (matches.length !== 128) fail(`expected 128 authored rows; found ${matches.length}`);
  const seen = new Set();
  return matches.map((match) => {
    const [block, key, heading] = match;
    if (seen.has(key)) fail(`duplicate authored key ${key}`);
    seen.add(key);
    const tone = heading.match(/tone: (corporate|diegetic)/u)?.[1];
    if (!tone) fail(`missing tone for ${key}`);
    const paramsLabel = heading.match(/params: ([^—\n]+?)(?: —|$)/u)?.[1].trim();
    const names = paramsLabel === undefined || paramsLabel === "(none)"
      ? []
      : paramsLabel.split(",").map((name) => name.trim()).filter(Boolean).sort();
    for (const name of names) if (!paramTypes[name]) fail(`unruled param type ${key}.${name}`);
    const variants = {};
    for (const variant of block.matchAll(/era_(1995|2000): "([^"\n]*)"/gu)) variants[`era_${variant[1]}`] = variant[2];
    return {
      key,
      text: blockText(block, key),
      params: names.map((name) => ({ name, type: paramTypes[name] })),
      era_variants: Object.keys(variants).length === 0 ? null : variants,
      provenance: [],
      tone,
    };
  });
}

function applyRulings(row) {
  const next = structuredClone(row);
  const text = {
    "cosmetic.horse_armor_free.description": "Decorative armor for a horse. You do not have a horse. Price: {price}, forever.",
    "satire.order_form.confirmation": "Thank you, {founder}!! Your {price} has been processed. Nothing will ship. You already had everything.",
    "satire.order_form.item_full_version": "Full version ............ {price}",
    "satire.order_form.item_shipping": "Shipping & handling ..... {price}",
    "satire.order_form.item_site_license": "Site license ............ {price}",
    "satire.order_form.total": "TOTAL ................... {price}",
    "system.offline_progress.tooltip": "While you were away the machine kept working at {rate_percent}%, up to {cap_hours} hours. BBS door games did this first, because everyone shared one phone line. They called it fairness. So do we.",
  }[next.key];
  if (text !== undefined) next.text = text;
  if (["cosmetic.horse_armor_free.description", "satire.order_form.item_full_version", "satire.order_form.item_shipping", "satire.order_form.item_site_license", "satire.order_form.total"].includes(next.key)) {
    next.params = [{ name: "price", type: "string" }];
  }
  if (next.key === "satire.order_form.confirmation") {
    next.params = [{ name: "founder", type: "string" }, { name: "price", type: "string" }];
  }
  if (next.key === "system.offline_progress.tooltip") {
    next.params = [{ name: "cap_hours", type: "integer" }, { name: "rate_percent", type: "integer" }];
  }
  if (next.key === "cosmetic.horse_armor_free.disclosure") next.provenance = ["gaming.horse_armor_2006"];
  if (next.key === "satire.unregistered.tooltip") next.provenance = ["shareware.mail_registration"];
  if (next.key === "satire.readme.body") {
    next.text = next.text
      .replace("  equipment keeps working while you are away -- at 90%, for up to 24\n  hours, the way the door games on the BBS shared one phone line.", "  equipment keeps working while you are away -- at ninety percent,\n  for up to a day, the way the door games on the BBS shared one phone line.")
      .replace("mail us $0.00", "mail us zero dollars")
      .replace("-- the authors, 1995", "-- the authors, nineteen ninety-five");
    next.text_kind = "longform";
  }
  return next;
}

const provenance = validateProvenanceRegistry(JSON.parse(readFileSync(path.join(repositoryRoot, "copy/provenance.v1.json"), "utf8")));
for (const claim of ["gaming.horse_armor_2006", "shareware.mail_registration"]) {
  if (provenance.get(claim)?.status !== "verified") fail(`missing verified claim ${claim}`);
}

const ruled = extractRows(readFileSync(rulingPath, "utf8")).map(applyRulings);
const omitted = new Map(ruled.filter((row) => row.key.startsWith("resource.company_")).map((row) => [row.key, row]));
if (omitted.size !== 2 || !omitted.has(cashReplacement.key) || !omitted.has(existingPermits.key)) fail("the two existing cap rows were not both present in the ruling");

const entries = ruled
  .filter((row) => !omitted.has(row.key))
  .concat(additions.map(([key, text, tone]) => ({ key, text, params: [], era_variants: null, provenance: [], tone })))
  .sort((left, right) => Buffer.from(left.key).compare(Buffer.from(right.key)));
if (entries.length !== 133 || new Set(entries.map((row) => row.key)).size !== 133) fail(`expected 133 new unique rows; found ${entries.length}`);

for (const entry of entries) {
  if (entry.text_kind === "longform") validateLongformCopyText(entry.text, entry.key);
  else validatePlainCopyText(entry.text, entry.key);
  if (containsStatistic(entry.text) && entry.provenance.length === 0) fail(`${entry.key} contains an unproven statistic`);
}

let phase0Bytes = readFileSync(phase0Path, "utf8");
const phase0 = JSON.parse(phase0Bytes);
const cashIndex = phase0.entries.findIndex((row) => row.key === cashReplacement.key);
if (cashIndex === -1) fail("phase0 cash cap row is missing");
if (JSON.stringify(phase0.entries[cashIndex]) !== JSON.stringify(cashReplacement)) {
  if (check) fail("copy/catalog/phase0.json cash cap row drift (run make game-ui-copy-candidate)");
  const rendered = JSON.stringify(cashReplacement, null, 2).split("\n").map((line) => `    ${line}`).join("\n");
  const pattern = /    \{\n      "key": "resource\.company_cash\.cap\.phase0",[\s\S]*?\n    \}(?=,?\n)/u;
  if (!pattern.test(phase0Bytes)) fail("could not replace the phase0 cash cap row");
  phase0Bytes = phase0Bytes.replace(pattern, rendered);
}
const permitsCatalog = JSON.parse(readFileSync(path.join(repositoryRoot, "copy/catalog/permits-candidate.json"), "utf8"));
if (JSON.stringify(permitsCatalog.entries) !== JSON.stringify([existingPermits])) fail("reviewed permits cap row drifted");

const presentationV1 = readFileSync(presentationV1Path, "utf8");
const presentation = JSON.parse(presentationV1);
if (presentation.schema_version !== 1) fail("presentation v1 identity drifted");
const presentationV2 = presentationV1
  .replace('"schema_version": 1', '"schema_version": 2')
  .replace(/\n\}\n$/u, `,\n  "gates": [\n    { "id": "gate.t0_to_t1", "title_key": "gate.t0_to_t1.title" }\n  ],\n  "exit_types": [\n    { "id": "acquihire", "title_key": "exit_type.acquihire.title" },\n    { "id": "acquisition", "title_key": "exit_type.acquisition.title" },\n    { "id": "collapse", "title_key": "exit_type.collapse.title" },\n    { "id": "ipo", "title_key": "exit_type.ipo.title" },\n    { "id": "scripted_first", "title_key": "exit_type.scripted_first.title" }\n  ]\n}\n`);

const eventCopyV1 = readFileSync(eventCopyV1Path, "utf8");
const eventCopy = JSON.parse(eventCopyV1);
if (eventCopy.schema_version !== 1 || eventCopy.bindings.some((row) => row.event_kind === "exit_offer_resolved")) fail("event-copy v1 identity drifted");
const eventCopyV2 = eventCopyV1.replace(
  '    { "event_kind": "exit_offer_expired", "copy_key": "event.exit_offer_expired", "parameters": [] },\n',
  '    { "event_kind": "exit_offer_expired", "copy_key": "event.exit_offer_expired", "parameters": [] },\n    { "event_kind": "exit_offer_resolved", "copy_key": "event.exit_offer_resolved", "parameters": [] },\n',
);

const outputs = new Map([
  [candidatePath, bytes({ schema_version: 1, entries })],
  [presentationV2Path, presentationV2],
  [presentationClientPath, presentationV2],
  [eventCopyV2Path, eventCopyV2],
]);
for (const [filename, value] of outputs) output(filename, value);
if (!check) writeFileSync(phase0Path, phase0Bytes, "utf8");

for (const [filename, value] of new Map([[phase0Path, phase0Bytes], ...outputs])) {
  console.log(`${check ? "verified" : "assembled"} ${path.relative(repositoryRoot, filename)} ${sha256(value)}`);
}
