import { readFileSync } from "node:fs";
import path from "node:path";

import {
  assertDenylistExtension,
  assertDenylistRecordsStable,
  assertVerifiedClaimsStable,
  buildCopyArtifact,
  containsStatistic,
  deniedTerm,
  generatedOutputs,
  historyPathEverExisted,
  repositoryRoot,
  validateDenylist,
  validateCopySafety,
  validatePlainCopyText,
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
const citationFixture = { term: "habbo", source_file: "design/research/social-spaces.md", source_anchor: "6-legal-safe-vs-fictionalize" };
expectFailure("denylist citation-retarget history fixture", () => assertDenylistRecordsStable([citationFixture], [{ ...citationFixture, source_anchor: "1-the-canon-deconstructed" }], "fixture"), /retargets protected copy denylist citation/);
// A1-F1 closure: an absent NON-research source must never receive the unpublication escape.
const nonResearchFixture = { term: "habbo", source_file: "docs/withdrawn-legal-notes.md", source_anchor: "6-legal-safe-vs-fictionalize" };
expectFailure("denylist non-research retarget escape fixture", () => assertDenylistRecordsStable([nonResearchFixture], [{ ...nonResearchFixture, source_anchor: "1-legal-elsewhere" }], "fixture", (record) => record.source_file.startsWith("design/research/") && true), /retargets protected copy denylist citation/);
if (!historyPathEverExisted("HEAD", "design/00-vision.md")) throw new Error("tracked citation history fixture was not found");
if (historyPathEverExisted("HEAD", "design/research/fixture-never-committed.md")) throw new Error("never-committed citation fixture unexpectedly exists in history");
const verifiedFixture = { claim_id: "fixture.verified", source_file: "design/research/speedrun-governance.md", source_anchor: "31-the-taxonomy-and-how-it-proliferates", status: "verified", source_urls: ["https://example.invalid/source"] };
expectFailure("verified provenance mutation history fixture", () => assertVerifiedClaimsStable([verifiedFixture], [{ ...verifiedFixture, source_anchor: "changed" }], "fixture"), /mutates or removes verified provenance claim/);
assertVerifiedClaimsStable([verifiedFixture], [{ ...verifiedFixture, source_file: "design/research/provenance-extracts.md", source_anchor: "changed" }], "fixture", () => true);
expectFailure("verified provenance status-mutation-under-escape fixture", () => assertVerifiedClaimsStable([verifiedFixture], [{ ...verifiedFixture, source_anchor: "changed", status: "attributed" }], "fixture", () => true), /mutates or removes verified provenance claim/);

for (const fixture of ["100%", "$12", "12$", "12€", "12 EUR", "12 ppm", "12 days", "1995"]) {
  if (!containsStatistic(fixture)) throw new Error(`statistic detector missed ${fixture}`);
}
for (const fixture of ["{amount}%", "version twelve", "company.cash", "NOTUSD12", "12USDsuffix"]) {
  if (containsStatistic(fixture)) throw new Error(`statistic detector overmatched ${fixture}`);
}
const safetyEntry = { key: "fixture.statistic", text: "Adoption reached 12%", params: [], era_variants: null, provenance: [], tone: "corporate" };
expectFailure("missing statistic provenance fixture", () => validateCopySafety([safetyEntry], claims, terms), /requires verified provenance/);
expectFailure("known-name copy fixture", () => validateCopySafety([{ ...safetyEntry, text: "Habbo", provenance: [] }], claims, terms), /known red-list term habbo/);
for (const fixture of ["     indented code", ">quoted", "#", "<!DOCTYPE html>", "<?xml version=\"1.0\"?>", "line  \nnext", "Heading\n=", "Heading\n--"]) {
  expectFailure("plain-text build fixture", () => validatePlainCopyText(fixture), /plain text/);
}
validatePlainCopyText("Offer declined. Run {run_seq} continues.", "underscore placeholder fixture");

const references = JSON.parse(readFileSync(path.join(repositoryRoot, "copy/references.v1.json"), "utf8"));
const codeReferences = built.codeReferences;
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
