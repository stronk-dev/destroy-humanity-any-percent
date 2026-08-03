import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

export const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const paramName = /^[a-z][a-z0-9_]*$/;
const canonicalDecimal = /^(?:0|-?[1-9](?:\.\d{0,10}[1-9])?e(?:0|-?[1-9]\d*))$/;
const tones = new Set(["achievement", "corporate", "diegetic", "lore_card"]);
const paramTypes = new Set(["canonical_decimal", "integer", "string"]);
const eras = new Set(["era_1995", "era_2000"]);
const statuses = new Set(["model", "plausible", "verified"]);
const fieldKinds = new Set(["copy_key", "name_key", "reason_key"]);
const artifactSchemas = Object.freeze({
  categories: "balance/leaderboards.schema.json",
  economy: "balance/economy.schema.json",
  factions: "balance/factions.schema.json",
});

const sourceCatalogDirectory = path.join(repositoryRoot, "copy/catalog");
const generatedCatalogPath = path.join(repositoryRoot, "client/src/copy/generated/catalog.json");
const generatedTypesPath = path.join(repositoryRoot, "client/src/copy/generated/types.ts");
const generatedHashPath = path.join(repositoryRoot, "client/src/copy/generated/copy-hash.txt");
const generatedOrphansPath = path.join(repositoryRoot, "copy/generated/orphans.v1.json");
const generatedCodeReferencesPath = path.join(repositoryRoot, "copy/generated/code-references.v1.json");

export const generatedPaths = Object.freeze([
  generatedCatalogPath,
  generatedTypesPath,
  generatedHashPath,
  generatedOrphansPath,
  generatedCodeReferencesPath,
]);

function fail(label, message) {
  throw new SyntaxError(`${label}: ${message}`);
}

let cachedConfig;

function copyConfig() {
  if (cachedConfig) return cachedConfig;
  const root = exactObject(readJSON(path.join(repositoryRoot, "copy/config.v1.json")), ["limits", "schema_version", "statistic_detector"], "copy/config.v1.json");
  const limits = exactObject(root.limits, ["max_text_lines", "max_text_utf8_bytes"], "copy/config.v1.json.limits");
  const detector = exactObject(root.statistic_detector, ["currency_tokens", "historical_year_max", "historical_year_min", "unit_tokens"], "copy/config.v1.json.statistic_detector");
  if (root.schema_version !== 1 || !Number.isSafeInteger(limits.max_text_utf8_bytes) || limits.max_text_utf8_bytes < 1 || !Number.isSafeInteger(limits.max_text_lines) || limits.max_text_lines < 1) fail("copy/config.v1.json", "has invalid limits");
  sortedUnique(detector.currency_tokens, "copy/config.v1.json.statistic_detector.currency_tokens");
  sortedUnique(detector.unit_tokens, "copy/config.v1.json.statistic_detector.unit_tokens");
  if (!Number.isSafeInteger(detector.historical_year_min) || !Number.isSafeInteger(detector.historical_year_max) || detector.historical_year_min < 1000 || detector.historical_year_max > 9999 || detector.historical_year_min > detector.historical_year_max) fail("copy/config.v1.json", "has an invalid historical-year range");
  cachedConfig = Object.freeze({ limits: Object.freeze({ maxTextBytes: limits.max_text_utf8_bytes, maxTextLines: limits.max_text_lines }), detector: Object.freeze({ currencyTokens: Object.freeze([...detector.currency_tokens]), unitTokens: Object.freeze([...detector.unit_tokens]), historicalYearMin: detector.historical_year_min, historicalYearMax: detector.historical_year_max }) });
  return cachedConfig;
}

function exactObject(value, keys, label) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) fail(label, "must be an object");
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) {
    fail(label, `expected exact keys ${expected.join(", ")}; got ${actual.join(", ")}`);
  }
  return value;
}

function sortedUnique(values, label) {
  if (!Array.isArray(values)) fail(label, "must be an array");
  for (let index = 0; index < values.length; index += 1) {
    if (typeof values[index] !== "string" || values[index] === "") fail(`${label}[${index}]`, "must be a non-empty string");
    if (index > 0 && values[index - 1] >= values[index]) fail(label, "must be byte-sorted and unique");
  }
  return values;
}

export function validatePlainCopyText(value, label = "copy text") {
  const { maxTextBytes: maxBytes, maxTextLines: maxLines } = copyConfig().limits;
  if (typeof value !== "string") fail(label, "must be a string");
  if (value !== value.normalize("NFC")) fail(label, "must be NFC-normalized");
  if (Buffer.byteLength(value, "utf8") > maxBytes) fail(label, `exceeds ${maxBytes} UTF-8 bytes`);
  if (value.split("\n").length > maxLines) fail(label, `exceeds ${maxLines} lines`);
  if (/[^\P{Cc}\n]/u.test(value)) fail(label, "contains a control character");
  if (/[<`*_~\[\]|\\]|-->|^ {4}|^\s{0,3}(?:#{1,6}(?:\s|$)|>|[-+]\s|\d+[.)](?:\s|$))|^\s{0,3}(?:-\s*){3,}$|^\s{0,3}===+\s*$/mu.test(value)) {
    fail(label, "must be plain text, not HTML or Markdown");
  }
  return value;
}

export function placeholders(text, label = "copy text") {
  const found = [];
  for (let index = 0; index < text.length; ) {
    if (text.startsWith("{{", index) || text.startsWith("}}", index)) {
      index += 2;
      continue;
    }
    if (text[index] === "{") {
      const end = text.indexOf("}", index + 1);
      if (end === -1) fail(label, "has an unclosed placeholder");
      const name = text.slice(index + 1, end);
      if (!paramName.test(name)) fail(label, `has invalid placeholder {${name}}`);
      found.push(name);
      index = end + 1;
      continue;
    }
    if (text[index] === "}") fail(label, "has an unmatched closing brace");
    index += 1;
  }
  return [...new Set(found)].sort();
}

function validateTextParams(text, expected, label) {
  const actual = placeholders(text, label);
  if (actual.length !== expected.length || actual.some((name, index) => name !== expected[index])) {
    fail(label, `placeholders ${actual.join(", ")} do not match params ${expected.join(", ")}`);
  }
}

function validateEntry(value, label) {
  const row = exactObject(value, ["era_variants", "key", "params", "provenance", "text", "tone"], label);
  if (typeof row.key !== "string" || !mechanicalID.test(row.key)) fail(`${label}.key`, "must be a mechanical ID");
  if (!tones.has(row.tone)) fail(`${label}.tone`, "is outside the closed tone union");
  if (!Array.isArray(row.params)) fail(`${label}.params`, "must be an array");
  const names = [];
  for (let index = 0; index < row.params.length; index += 1) {
    const item = exactObject(row.params[index], ["name", "type"], `${label}.params[${index}]`);
    if (typeof item.name !== "string" || !paramName.test(item.name)) fail(`${label}.params[${index}].name`, "is invalid");
    if (!paramTypes.has(item.type)) fail(`${label}.params[${index}].type`, "is outside the closed param union");
    if (index > 0 && names[index - 1] >= item.name) fail(`${label}.params`, "must be byte-sorted and unique by name");
    names.push(item.name);
  }
  row.text = validatePlainCopyText(row.text, `${label}.text`);
  validateTextParams(row.text, names, `${label}.text`);
  if (row.era_variants !== null) {
    if (typeof row.era_variants !== "object" || Array.isArray(row.era_variants)) fail(`${label}.era_variants`, "must be null or an object");
    const variantKeys = Object.keys(row.era_variants);
    if (variantKeys.length === 0 || variantKeys.some((era) => !eras.has(era))) fail(`${label}.era_variants`, "contains an unknown or empty era set");
    if (variantKeys.some((era, index) => index > 0 && variantKeys[index - 1] >= era)) fail(`${label}.era_variants`, "keys must be byte-sorted");
    for (const era of variantKeys) {
      row.era_variants[era] = validatePlainCopyText(row.era_variants[era], `${label}.era_variants.${era}`);
      validateTextParams(row.era_variants[era], names, `${label}.era_variants.${era}`);
    }
  }
  sortedUnique(row.provenance, `${label}.provenance`);
  return row;
}

function catalogFiles() {
  const walk = (directory) => readdirSync(directory, { withFileTypes: true })
    .flatMap((entry) => entry.isDirectory() ? walk(path.join(directory, entry.name)) : [path.join(directory, entry.name)])
    .filter((filename) => filename.endsWith(".json"));
  return walk(sourceCatalogDirectory).sort((left, right) => Buffer.from(path.relative(repositoryRoot, left)).compare(Buffer.from(path.relative(repositoryRoot, right))));
}

function readJSON(filename, label = path.relative(repositoryRoot, filename)) {
  const bytes = readFileSync(filename, "utf8");
  try {
    return JSON.parse(bytes);
  } catch (error) {
    fail(label, `invalid JSON (${error.message})`);
  }
}

export function markdownAnchors(markdown) {
  const anchors = [];
  for (const line of markdown.split("\n")) {
    const match = line.match(/^#{1,6}\s+(.+?)\s*#*$/u);
    if (!match) continue;
    const anchor = match[1]
      .normalize("NFC")
      .toLowerCase()
      .replace(/[`*_~]/gu, "")
      .replace(/[^\p{L}\p{N}\s-]/gu, "")
      .trim()
      .replace(/\s+/gu, "-")
      .replace(/-+/gu, "-");
    anchors.push(anchor);
  }
  return anchors;
}

function resolveResearchAnchor(sourceFile, sourceAnchor, label) {
  if (typeof sourceFile !== "string" || !/^design\/research\/[a-z0-9][a-z0-9-]*\.md$/u.test(sourceFile)) {
    fail(`${label}.source_file`, "must name a Markdown file under design/research");
  }
  if (typeof sourceAnchor !== "string" || !/^[a-z0-9]+(?:-[a-z0-9]+)*$/u.test(sourceAnchor)) {
    fail(`${label}.source_anchor`, "must be a normalized heading anchor");
  }
  const filename = path.resolve(repositoryRoot, sourceFile);
  if (!filename.startsWith(`${path.join(repositoryRoot, "design/research")}${path.sep}`) || !existsSync(filename)) {
    fail(`${label}.source_file`, "does not resolve under design/research");
  }
  try {
    execFileSync("git", ["ls-files", "--error-unmatch", "--", sourceFile], { cwd: repositoryRoot, stdio: "ignore" });
  } catch {
    fail(`${label}.source_file`, "must be tracked in Git (untracked research cannot satisfy a shipping gate)");
  }
  const count = markdownAnchors(readFileSync(filename, "utf8")).filter((anchor) => anchor === sourceAnchor).length;
  if (count !== 1) fail(`${label}.source_anchor`, `must resolve exactly once (found ${count})`);
}

export function validateProvenanceRegistry(value, label = "copy/provenance.v1.json") {
  const root = exactObject(value, ["claims", "schema_version"], label);
  if (root.schema_version !== 1 || !Array.isArray(root.claims)) fail(label, "must be provenance schema version 1");
  let previous = "";
  const claims = new Map();
  for (let index = 0; index < root.claims.length; index += 1) {
    const claimLabel = `${label}.claims[${index}]`;
    const claim = exactObject(root.claims[index], ["claim_id", "source_anchor", "source_file", "source_urls", "status"], claimLabel);
    if (typeof claim.claim_id !== "string" || !mechanicalID.test(claim.claim_id) || claim.claim_id <= previous) fail(`${claimLabel}.claim_id`, "must be a byte-sorted unique mechanical ID");
    previous = claim.claim_id;
    if (!statuses.has(claim.status)) fail(`${claimLabel}.status`, "is outside the closed status union");
    sortedUnique(claim.source_urls, `${claimLabel}.source_urls`);
    if (claim.source_urls.some((url) => { try { return new URL(url).protocol !== "https:"; } catch { return true; } })) fail(`${claimLabel}.source_urls`, "must contain only HTTPS URLs");
    if (claim.status === "verified" && claim.source_urls.length === 0) fail(claimLabel, "verified claims require an HTTPS source URL");
    resolveResearchAnchor(claim.source_file, claim.source_anchor, claimLabel);
    claims.set(claim.claim_id, claim);
  }
  return claims;
}

export function normalizedTokens(value) {
  return value.normalize("NFC").toLowerCase().match(/[\p{L}\p{N}]+/gu) ?? [];
}

export function validateDenylistRecords(bytes, label = "moderation/copy-denylist.txt", { resolveSources = true } = {}) {
  const records = [];
  let source = null;
  for (const [offset, raw] of bytes.split("\n").entries()) {
    const line = raw.trim();
    if (line === "") continue;
    if (line.startsWith("#")) {
      const match = line.match(/^# source: (design\/research\/[a-z0-9][a-z0-9-]*\.md)#([a-z0-9]+(?:-[a-z0-9]+)*)$/u);
      source = match ? { file: match[1], anchor: match[2] } : null;
      continue;
    }
    const lineLabel = `${label}:${offset + 1}`;
    const normalized = normalizedTokens(line).join(" ");
    if (line !== normalized || normalized === "") fail(lineLabel, "term must be case-folded NFC tokens separated by one space");
    if (!source) fail(lineLabel, "requires an immediately preceding research source comment");
    if (!source.anchor.includes("legal")) fail(lineLabel, "source anchor must name a legal matrix or legal-safety section");
    if (resolveSources) resolveResearchAnchor(source.file, source.anchor, lineLabel);
    if (records.length > 0 && records.at(-1).term >= normalized) fail(label, "terms must be byte-sorted and unique");
    records.push(Object.freeze({ term: normalized, source_file: source.file, source_anchor: source.anchor }));
    source = null;
  }
  if (records.length === 0) fail(label, "must contain at least one term");
  return records;
}

export function validateDenylist(bytes, label = "moderation/copy-denylist.txt") {
  return validateDenylistRecords(bytes, label).map((record) => record.term);
}

export function deniedTerm(text, terms) {
  const tokens = normalizedTokens(text);
  const collapsed = tokens.join(" ");
  const hasCompactedSequence = (needle) => {
    for (let start = 0; start < tokens.length; start += 1) {
      let value = "";
      for (let end = start; end < tokens.length && value.length < needle.length; end += 1) {
        value += tokens[end];
        if (value === needle) return true;
      }
    }
    return false;
  };
  for (const term of terms) {
    const termTokens = term.split(" ");
    const compactTerm = termTokens.join("");
    if (termTokens.length === 1
      ? tokens.includes(term) || hasCompactedSequence(compactTerm)
      : (` ${collapsed} `).includes(` ${term} `) || hasCompactedSequence(compactTerm)) return term;
  }
  return null;
}

function withoutPlaceholders(text) {
  let result = "";
  for (let index = 0; index < text.length; ) {
    if (text.startsWith("{{", index)) { result += " "; index += 2; continue; }
    if (text.startsWith("}}", index)) { result += " "; index += 2; continue; }
    if (text[index] === "{") {
      const end = text.indexOf("}", index + 1);
      result += " ";
      index = end + 1;
      continue;
    }
    result += text[index];
    index += 1;
  }
  return result;
}

export function containsStatistic(text) {
  const literal = withoutPlaceholders(text);
  const { currencyTokens, unitTokens, historicalYearMin, historicalYearMax } = copyConfig().detector;
  const escaped = (value) => value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
  const currencies = currencyTokens.map(escaped).join("|");
  const units = unitTokens.map(escaped).join("|");
  if (new RegExp(`(?:\\d+(?:\\.\\d+)?\\s*%|(?<![\\p{L}\\p{N}_])(?:${currencies})\\s*\\d|\\d+(?:\\.\\d+)?\\s*(?:${currencies}|${units})(?![\\p{L}\\p{N}_]))`, "u").test(literal)) return true;
  return (literal.match(/\b\d{4}\b/gu) ?? []).some((value) => Number(value) >= historicalYearMin && Number(value) <= historicalYearMax);
}

export function validateCopySafety(entries, claims, denylist) {
  for (const entry of entries) {
    const texts = [entry.text, ...Object.values(entry.era_variants ?? {})];
    const requiresProvenance = entry.tone === "lore_card" || texts.some(containsStatistic) || entry.provenance.length > 0;
    if (requiresProvenance && entry.provenance.length === 0) fail(entry.key, "requires verified provenance");
    for (const claimID of entry.provenance) {
      const claim = claims.get(claimID);
      if (!claim) fail(entry.key, `references unknown provenance claim ${claimID}`);
      if (claim.status !== "verified") fail(entry.key, `cannot ship ${claim.status} provenance claim ${claimID}`);
    }
    for (const text of texts) {
      const denied = deniedTerm(text, denylist);
      if (denied) fail(entry.key, `contains known red-list term ${denied}`);
    }
  }
}

function schemaCandidates(schema, root) {
  if (schema === true || schema === false || schema === null || typeof schema !== "object") return [schema];
  const resolved = typeof schema.$ref === "string" && schema.$ref.startsWith("#/")
    ? schema.$ref.slice(2).split("/").reduce((value, segment) => value?.[segment.replaceAll("~1", "/").replaceAll("~0", "~")], root)
    : schema;
  const branches = [resolved];
  for (const key of ["allOf", "anyOf", "oneOf"]) if (Array.isArray(resolved?.[key])) branches.push(...resolved[key]);
  for (const key of ["then", "else"]) if (resolved?.[key]) branches.push(resolved[key]);
  return branches;
}

function schemaSupportsPattern(schema, segments, root, seen = new Set()) {
  if (segments.length === 0) return true;
  for (const candidate of schemaCandidates(schema, root)) {
    if (candidate === true) return true;
    if (!candidate || candidate === false || typeof candidate !== "object") continue;
    const identity = `${segments.join("/")}::${JSON.stringify(candidate)}`;
    if (seen.has(identity)) continue;
    seen.add(identity);
    if (typeof candidate.$ref === "string" && schemaSupportsPattern(candidate, segments, root, seen)) return true;
    const [segment, ...rest] = segments;
    if (segment === "*" && candidate.items && schemaSupportsPattern(candidate.items, rest, root, seen)) return true;
    if (segment !== "*" && candidate.properties?.[segment] !== undefined && schemaSupportsPattern(candidate.properties[segment], rest, root, seen)) return true;
  }
  return false;
}

function pointerSegments(pattern, label) {
  if (typeof pattern !== "string" || !pattern.startsWith("/") || pattern.endsWith("/")) fail(label, "must be an absolute JSON-pointer pattern");
  const segments = pattern.slice(1).split("/").map((segment) => segment.replaceAll("~1", "/").replaceAll("~0", "~"));
  if (segments.some((segment) => segment === "" || segment !== "*" && !/^[a-z][a-z0-9_]*$/u.test(segment))) fail(label, "contains an invalid segment");
  return segments;
}

function walkPointer(value, segments, rendered = "") {
  if (segments.length === 0) return [{ path: rendered || "/", value }];
  const [segment, ...rest] = segments;
  if (segment === "*") {
    if (!Array.isArray(value)) return [];
    return value.flatMap((item, index) => walkPointer(item, rest, `${rendered}/${index}`));
  }
  if (value === null || typeof value !== "object" || Array.isArray(value) || !(segment in value)) return [];
  return walkPointer(value[segment], rest, `${rendered}/${segment}`);
}

function epochArtifacts() {
  const seed = exactObject(readJSON(path.join(repositoryRoot, "balance/epochs/phase0.json")), ["artifacts", "current_epoch_id", "epochs", "schema_version"], "balance/epochs/phase0.json");
  const result = new Map();
  for (const artifact of seed.artifacts) {
    const row = exactObject(artifact, ["name", "path"], `epoch artifact ${result.size}`);
    if (result.has(row.name)) fail("epoch artifacts", `duplicate ${row.name}`);
    result.set(row.name, row.path);
  }
  return result;
}

export function validateReferences(value, copyKeys, codeReferences) {
  const root = exactObject(value, ["references", "schema_version"], "copy/references.v1.json");
  if (root.schema_version !== 1 || !Array.isArray(root.references)) fail("copy/references.v1.json", "must be schema version 1");
  const artifacts = epochArtifacts();
  const referenced = new Set(codeReferences);
  let previous = "";
  for (let index = 0; index < root.references.length; index += 1) {
    const label = `copy/references.v1.json.references[${index}]`;
    const row = exactObject(root.references[index], ["artifact_name", "field_kind", "json_pointer_pattern"], label);
    const orderKey = `${row.artifact_name}\0${row.json_pointer_pattern}`;
    if (typeof row.artifact_name !== "string" || !mechanicalID.test(row.artifact_name) || orderKey <= previous) fail(label, "rows must be byte-sorted and unique");
    previous = orderKey;
    if (!fieldKinds.has(row.field_kind)) fail(`${label}.field_kind`, "is outside the closed field-kind union");
    const artifactPath = artifacts.get(row.artifact_name);
    if (!artifactPath) fail(label, `references unknown epoch artifact ${row.artifact_name}`);
    const schemaPath = artifactSchemas[row.artifact_name];
    if (!schemaPath) fail(label, `has no registered schema authority for ${row.artifact_name}`);
    const segments = pointerSegments(row.json_pointer_pattern, `${label}.json_pointer_pattern`);
    const schema = readJSON(path.join(repositoryRoot, schemaPath));
    if (!schemaSupportsPattern(schema, segments, schema)) fail(label, `pattern matches no field in ${schemaPath}`);
    const artifact = readJSON(path.join(repositoryRoot, artifactPath));
    for (const match of walkPointer(artifact, segments)) {
      if (typeof match.value !== "string" || !mechanicalID.test(match.value)) fail(`${row.artifact_name}:${match.path}`, "copy reference must be a mechanical ID");
      if (!copyKeys.has(match.value)) fail(`${row.artifact_name}:${match.path}:${match.value}`, "missing copy key");
      referenced.add(match.value);
    }
  }
  for (const key of codeReferences) if (!copyKeys.has(key)) fail(`code:${key}`, "missing copy key");
  return referenced;
}

function codeReferencesFromSites() {
  const label = "copy/code-reference-sites.v1.json";
  const root = exactObject(readJSON(path.join(repositoryRoot, label)), ["references", "schema_version"], label);
  if (root.schema_version !== 1 || !Array.isArray(root.references)) fail(label, "must be schema version 1");
  const keys = [];
  let previous = "";
  for (let index = 0; index < root.references.length; index += 1) {
    const rowLabel = `${label}.references[${index}]`;
    const row = exactObject(root.references[index], ["go_function", "json_field", "key", "source_file"], rowLabel);
    if (typeof row.key !== "string" || !mechanicalID.test(row.key) || row.key <= previous) fail(rowLabel, "keys must be byte-sorted unique mechanical IDs");
    if (typeof row.source_file !== "string" || !/^server\/[a-z0-9_/-]+\.go$/u.test(row.source_file)) fail(`${rowLabel}.source_file`, "must name an explicit Go producer site");
    if (typeof row.go_function !== "string" || !/^[A-Za-z][A-Za-z0-9]*$/u.test(row.go_function) || typeof row.json_field !== "string" || !paramName.test(row.json_field)) fail(rowLabel, "has an invalid Go function or JSON field binding");
    const filename = path.resolve(repositoryRoot, row.source_file);
    if (!filename.startsWith(`${path.join(repositoryRoot, "server")}${path.sep}`) || !existsSync(filename)) fail(`${rowLabel}.source_file`, "does not resolve under server");
    previous = row.key;
    keys.push(row.key);
  }
  return keys;
}

export function buildCopyArtifact() {
  const entries = [];
  const seen = new Set();
  for (const filename of catalogFiles()) {
    const relative = path.relative(repositoryRoot, filename);
    const root = exactObject(readJSON(filename), ["entries", "schema_version"], relative);
    if (root.schema_version !== 1 || !Array.isArray(root.entries) || root.entries.length === 0) fail(relative, "must be a non-empty schema-version-1 catalog");
    let previous = "";
    for (let index = 0; index < root.entries.length; index += 1) {
      const entry = validateEntry(root.entries[index], `${relative}.entries[${index}]`);
      if (entry.key <= previous) fail(relative, "entries must be byte-sorted and unique");
      if (seen.has(entry.key)) fail(relative, `duplicate global copy key ${entry.key}`);
      previous = entry.key;
      seen.add(entry.key);
      entries.push(entry);
    }
  }
  entries.sort((left, right) => Buffer.from(left.key).compare(Buffer.from(right.key)));
  const claims = validateProvenanceRegistry(readJSON(path.join(repositoryRoot, "copy/provenance.v1.json")));
  const denylist = validateDenylist(readFileSync(path.join(repositoryRoot, "moderation/copy-denylist.txt"), "utf8"));
  validateCopySafety(entries, claims, denylist);
  const copyKeys = new Set(entries.map((entry) => entry.key));
  const codeReferences = codeReferencesFromSites();
  const referenced = validateReferences(readJSON(path.join(repositoryRoot, "copy/references.v1.json")), copyKeys, codeReferences);
  const artifact = { schema_version: 1, entries };
  const bytes = `${JSON.stringify(artifact, null, 2)}\n`;
  const copyHash = `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
  const orphans = entries.map((entry) => entry.key).filter((key) => !referenced.has(key));
  return { artifact, bytes, copyHash, orphans, codeReferences };
}

function tsType(type) {
  if (type === "integer") return "number";
  return "string";
}

export function generatedTypes(entries, copyHash) {
  const config = copyConfig();
  const lines = [
    "// Generated by `make copy-generate`; do not edit.",
    `export const COPY_HASH = ${JSON.stringify(copyHash)} as const;`,
    `export const COPY_MAX_TEXT_UTF8_BYTES = ${config.limits.maxTextBytes} as const;`,
    `export const COPY_MAX_TEXT_LINES = ${config.limits.maxTextLines} as const;`,
    "",
    "export interface CopyParamsByKey {",
  ];
  for (const entry of entries) {
    const value = entry.params.length === 0
      ? "Readonly<Record<string, never>>"
      : `{ ${entry.params.map((param) => `readonly ${JSON.stringify(param.name)}: ${tsType(param.type)}`).join("; ")} }`;
    lines.push(`  readonly ${JSON.stringify(entry.key)}: ${value};`);
  }
  lines.push("}", "", "export type CopyKey = keyof CopyParamsByKey;", "", `export const COPY_KEYS = ${JSON.stringify(entries.map((entry) => entry.key), null, 2)} as const;`, "");
  return lines.join("\n");
}

export function generatedOutputs() {
  const built = buildCopyArtifact();
  return new Map([
    [generatedCatalogPath, built.bytes],
    [generatedTypesPath, generatedTypes(built.artifact.entries, built.copyHash)],
    [generatedHashPath, `${built.copyHash}\n`],
    [generatedOrphansPath, `${JSON.stringify({ schema_version: 1, keys: built.orphans }, null, 2)}\n`],
    [generatedCodeReferencesPath, `${JSON.stringify({ schema_version: 1, keys: built.codeReferences }, null, 2)}\n`],
  ]);
}

function git(...args) {
  return execFileSync("git", args, { cwd: repositoryRoot, encoding: "utf8" }).trim();
}

export function assertDenylistExtension(before, after, label) {
  const removed = before.filter((term) => !after.includes(term));
  if (removed.length > 0) throw new Error(`${label} removes copy denylist terms: ${removed.join(", ")}`);
}

export function assertDenylistRecordsStable(before, afterRows, label, canCorrectInvalidSource = () => false) {
  const after = new Map(afterRows.map((record) => [record.term, record]));
  for (const record of before) {
    const next = after.get(record.term);
    if (!next) throw new Error(`${label} removes copy denylist terms: ${record.term}`);
    if ((next.source_file !== record.source_file || next.source_anchor !== record.source_anchor) && !canCorrectInvalidSource(record)) {
      throw new Error(`${label} retargets protected copy denylist citation for ${record.term}`);
    }
  }
}

export function assertVerifiedClaimsStable(before, afterRows, label) {
  const after = new Map(afterRows.map((claim) => [claim.claim_id, claim]));
  for (const claim of before.filter((row) => row.status === "verified")) {
    if (JSON.stringify(after.get(claim.claim_id)) !== JSON.stringify(claim)) throw new Error(`${label} mutates or removes verified provenance claim ${claim.claim_id}`);
  }
}

function parseHistoricalJSON(commit, filename) {
  return JSON.parse(git("show", `${commit}:${filename}`));
}

function historyFileExists(commit, filename) {
  try {
    execFileSync("git", ["cat-file", "-e", `${commit}:${filename}`], { cwd: repositoryRoot, stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

export function historyPathEverExisted(commit, filename) {
  return git("log", "--format=%H", commit, "--", filename) !== "";
}

export function verifyAppendOnlyHistory() {
  if (git("rev-parse", "--is-shallow-repository") === "true") throw new Error("copy governance history guard requires complete Git history");
  const protectedFiles = ["moderation/copy-denylist.txt", "copy/provenance.v1.json"];
  for (const filename of protectedFiles) {
    const introduction = git("log", "--diff-filter=A", "--format=%H", "--", filename).split("\n").filter(Boolean).at(-1);
    if (!introduction) continue;
    const commits = [introduction, ...git("rev-list", "--reverse", `${introduction}..HEAD`).split("\n").filter(Boolean)];
    for (const commit of commits) {
      if (!historyFileExists(commit, filename)) throw new Error(`commit ${commit} removes protected copy file ${filename}`);
      const parents = git("rev-list", "--parents", "-n", "1", commit).split(" ").slice(1);
      for (const parent of parents) {
        if (!historyFileExists(parent, filename)) continue;
        if (filename.endsWith("copy-denylist.txt")) {
          const before = validateDenylistRecords(git("show", `${parent}:${filename}`), `${parent}:${filename}`, { resolveSources: false });
          const after = validateDenylistRecords(git("show", `${commit}:${filename}`), `${commit}:${filename}`, { resolveSources: false });
          assertDenylistRecordsStable(before, after, `commit ${commit}`, (record) => !historyPathEverExisted(parent, record.source_file));
        } else {
          const before = parseHistoricalJSON(parent, filename).claims;
          assertVerifiedClaimsStable(before, parseHistoricalJSON(commit, filename).claims, `commit ${commit}`);
        }
      }
    }
  }
  if (historyFileExists("HEAD", "moderation/copy-denylist.txt")) {
    const before = validateDenylistRecords(git("show", "HEAD:moderation/copy-denylist.txt"), "HEAD denylist", { resolveSources: false });
    const after = validateDenylistRecords(readFileSync(path.join(repositoryRoot, "moderation/copy-denylist.txt"), "utf8"), "worktree denylist");
    assertDenylistRecordsStable(before, after, "worktree", (record) => !historyPathEverExisted("HEAD", record.source_file));
  }
  if (historyFileExists("HEAD", "copy/provenance.v1.json")) {
    const before = parseHistoricalJSON("HEAD", "copy/provenance.v1.json").claims;
    assertVerifiedClaimsStable(before, readJSON(path.join(repositoryRoot, "copy/provenance.v1.json")).claims, "worktree");
  }
}

export function validateCanonicalDecimal(value) {
  return typeof value === "string" && canonicalDecimal.test(value);
}
