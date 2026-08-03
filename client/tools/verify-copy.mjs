import { readFileSync } from "node:fs";
import path from "node:path";

import {
  assertDenylistExtension,
  assertVerifiedClaimsStable,
  buildCopyArtifact,
  containsStatistic,
  deniedTerm,
  generatedOutputs,
  repositoryRoot,
  validateDenylist,
  validateCopySafety,
  validateProvenanceRegistry,
  validateReferences,
  verifyAppendOnlyHistory,
} from "./copy-pipeline.mjs";

function expectFailure(label, operation, pattern) {
  try {
    operation();
  } catch (error) {
    if (pattern.test(String(error))) return;
    throw new Error(`${label} failed for the wrong reason: ${error}`);
  }
  throw new Error(`${label} unexpectedly passed`);
}

const built = buildCopyArtifact();
const terms = validateDenylist(readFileSync(path.join(repositoryRoot, "moderation/copy-denylist.txt"), "utf8"));
const claims = validateProvenanceRegistry(JSON.parse(readFileSync(path.join(repositoryRoot, "copy/provenance.v1.json"), "utf8")));
if (claims.get("speedrun.category_taxonomy")?.status !== "verified") throw new Error("verified provenance fixture did not resolve");

if (deniedTerm("H.a-b_b/o", terms) !== "habbo" || deniedTerm("Club—Penguin", terms) !== "club penguin") {
  throw new Error("separator-bypass denylist fixtures were not rejected");
}
if (deniedTerm("habitable office", terms) !== null) throw new Error("denylist word-boundary near-miss was rejected");
expectFailure("denylist removal history fixture", () => assertDenylistExtension(["habbo", "neopets"], ["habbo"], "fixture"), /removes copy denylist terms: neopets/);
const verifiedFixture = { claim_id: "fixture.verified", source_file: "design/research/speedrun-governance.md", source_anchor: "31-the-taxonomy-and-how-it-proliferates", status: "verified", source_urls: ["https://example.invalid/source"] };
expectFailure("verified provenance mutation history fixture", () => assertVerifiedClaimsStable([verifiedFixture], [{ ...verifiedFixture, source_anchor: "changed" }], "fixture"), /mutates or removes verified provenance claim/);

for (const fixture of ["100%", "$12", "12 EUR", "12 ppm", "12 days", "1995"]) {
  if (!containsStatistic(fixture)) throw new Error(`statistic detector missed ${fixture}`);
}
for (const fixture of ["{amount}%", "version twelve", "company.cash"]) {
  if (containsStatistic(fixture)) throw new Error(`statistic detector overmatched ${fixture}`);
}
const safetyEntry = { key: "fixture.statistic", text: "Adoption reached 12%", params: [], era_variants: null, provenance: [], tone: "corporate" };
expectFailure("missing statistic provenance fixture", () => validateCopySafety([safetyEntry], claims, terms), /requires verified provenance/);
expectFailure("known-name copy fixture", () => validateCopySafety([{ ...safetyEntry, text: "Habbo", provenance: [] }], claims, terms), /known red-list term habbo/);

const references = JSON.parse(readFileSync(path.join(repositoryRoot, "copy/references.v1.json"), "utf8"));
const codeReferences = JSON.parse(readFileSync(path.join(repositoryRoot, "copy/code-references.v1.json"), "utf8")).keys;
const keys = new Set(built.artifact.entries.map((entry) => entry.key));
const missing = new Set(keys);
missing.delete("category.any_percent");
expectFailure("missing-copy completeness fixture", () => validateReferences(references, missing, codeReferences), /categories:\/categories\/0\/name_key:category\.any_percent/);

for (const [filename, expected] of generatedOutputs()) {
  let actual;
  try {
    actual = readFileSync(filename, "utf8");
  } catch {
    throw new Error(`generated copy artifact missing: ${path.relative(repositoryRoot, filename)}`);
  }
  if (actual !== expected) throw new Error(`generated copy artifact drift: ${path.relative(repositoryRoot, filename)} (run make copy-generate)`);
}

verifyAppendOnlyHistory();
console.log(`copy pipeline ok: ${built.artifact.entries.length} keys, ${built.copyHash}, ${built.orphans.length} orphan warnings`);
