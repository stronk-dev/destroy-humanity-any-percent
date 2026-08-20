import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";

const repositoryRoot = new URL("../../", import.meta.url);
const inventoryLines = readFileSync(new URL("planning/platform-alignment/balance-file-inventory.tsv", repositoryRoot), "utf8")
  .trimEnd().split("\n");
const artifacts = inventoryLines.slice(1).map((line) => line.split("\t"))
  .filter((row) => row[1] === "epoch_artifact")
  .map(([file, , family]) => ({ file, family }));
if (artifacts.length !== 19 || new Set(artifacts.map(({ file }) => file)).size !== 19) {
  throw new Error("epoch-artifact population drift");
}

const escapePointer = (value) => value.replaceAll("~", "~0").replaceAll("/", "~1");
const pointer = (parts) => `/${parts.map((part) => escapePointer(String(part))).join("/")}`;
const hash = (value) => createHash("sha256").update(JSON.stringify(value)).digest("hex");
const scalar = (value) => value === null || typeof value !== "object";
const authoredIdentity = (value, path, kind) => {
  if (kind === "empty_collection") return "<empty>";
  if (scalar(value)) return JSON.stringify(value);
  if (typeof value.id === "string") return value.id;
  const identityKey = Object.keys(value).find((key) => key.endsWith("_id") && typeof value[key] === "string");
  return identityKey === undefined ? `object:${pointer(path)}` : `${identityKey}:${value[identityKey]}`;
};

const units = [];
function emit(artifact, path, kind, value) {
  const jsonPointer = pointer(path);
  units.push([
    `${artifact.file}#${jsonPointer}`,
    artifact.file,
    artifact.family,
    jsonPointer,
    kind,
    authoredIdentity(value, path, kind),
    hash(value),
  ]);
}

function walkArray(artifact, value, path) {
  if (value.length === 0) {
    emit(artifact, path, "empty_collection", value);
    return;
  }
  for (let index = 0; index < value.length; index += 1) {
    const child = value[index];
    const childPath = [...path, index];
    if (Array.isArray(child)) walkArray(artifact, child, childPath);
    else if (scalar(child)) emit(artifact, childPath, "primitive_edge", child);
    else walkObject(artifact, child, childPath, "object_record");
  }
}

function walkObject(artifact, value, path, kind) {
  emit(artifact, path, kind, value);
  for (const [key, child] of Object.entries(value)) {
    const childPath = [...path, key];
    if (Array.isArray(child)) walkArray(artifact, child, childPath);
    else if (!scalar(child)) walkObject(artifact, child, childPath, "nested_object");
  }
}

for (const artifact of artifacts) {
  const root = JSON.parse(readFileSync(new URL(artifact.file, repositoryRoot), "utf8"));
  for (const [key, value] of Object.entries(root)) {
    if (key === "schema_version") continue;
    const path = [key];
    if (Array.isArray(value)) walkArray(artifact, value, path);
    else if (scalar(value)) emit(artifact, path, "root_policy_field", value);
    else walkObject(artifact, value, path, "singleton_policy");
  }
}

const header = ["unit_id", "artifact", "family", "json_pointer", "unit_kind", "authored_identity", "payload_sha256"];
function validate(rows) {
  if (rows.length !== units.length) throw new Error("content-unit denominator drift");
  if (new Set(rows.map((row) => row[0])).size !== rows.length) throw new Error("duplicate content-unit identity");
  for (let index = 0; index < rows.length; index += 1) {
    if (rows[index].length !== header.length || rows[index].some((value, column) => value !== units[index][column])) {
      throw new Error(`content-unit order/payload drift: ${rows[index]?.[0] ?? index}`);
    }
  }
  if (!rows.some((row) => row[4] === "empty_collection")) throw new Error("empty collections disappeared");
  if (!rows.some((row) => row[4] === "root_policy_field")) throw new Error("root policies disappeared");
}

validate(units);
const seededFailures = [];
const mustReject = (label, rows) => {
  try {
    validate(rows);
  } catch {
    seededFailures.push(label);
    return;
  }
  throw new Error(`seeded failure accepted: ${label}`);
};
mustReject("dropped-unit", units.slice(1));
mustReject("duplicate-unit", [units[0], ...units.slice(0, -1)]);
mustReject("empty-collection-omission", units.filter((row, index) => index !== units.findIndex((candidate) => candidate[4] === "empty_collection")));
mustReject("root-policy-omission", units.filter((row, index) => index !== units.findIndex((candidate) => candidate[4] === "root_policy_field")));

console.log(header.join("\t"));
for (const row of units) console.log(row.join("\t"));
const kinds = {};
const families = {};
for (const row of units) {
  kinds[row[4]] = (kinds[row[4]] ?? 0) + 1;
  families[row[2]] = (families[row[2]] ?? 0) + 1;
}
console.error(JSON.stringify({ artifacts: artifacts.length, units: units.length, kinds, families, seededFailures }));
