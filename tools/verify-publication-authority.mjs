#!/usr/bin/env node

import { existsSync, mkdtempSync, readFileSync, rmSync } from "node:fs";
import { dirname, isAbsolute, join, normalize, relative, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";

const defaultManifest = "planning/platform-alignment/publication-authority-manifest.json";

function fail(message) {
  throw new Error(message);
}

function parseArgs(argv) {
  const options = { root: process.cwd(), manifest: defaultManifest, selfTest: false, freshClone: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--root") options.root = resolve(argv[++index]);
    else if (arg === "--manifest") options.manifest = argv[++index];
    else if (arg === "--self-test") options.selfTest = true;
    else if (arg === "--fresh-clone") options.freshClone = true;
    else fail(`unknown argument: ${arg}`);
  }
  return options;
}

function git(root, args, { allowFailure = false } = {}) {
  const result = spawnSync("git", args, { cwd: root, encoding: "utf8" });
  if (!allowFailure && result.status !== 0) {
    fail(`git ${args.join(" ")} failed: ${(result.stderr || result.stdout).trim()}`);
  }
  return result;
}

function trackedFiles(root) {
  return new Set(git(root, ["ls-files", "-z"]).stdout.split("\0").filter(Boolean));
}

function ignoredPaths(root, paths) {
  const unique = [...new Set(paths)];
  if (unique.length === 0) return new Set();
  const result = spawnSync("git", ["check-ignore", "--no-index", "-z", "--stdin"], {
    cwd: root,
    encoding: "utf8",
    input: `${unique.join("\0")}\0`,
  });
  if (result.status !== 0 && result.status !== 1) {
    fail(`git check-ignore failed: ${(result.stderr || result.stdout).trim()}`);
  }
  return new Set(result.stdout.split("\0").filter(Boolean));
}

function uniqueStrings(values, label) {
  if (!Array.isArray(values) || values.some((value) => typeof value !== "string" || value.length === 0)) {
    fail(`${label} must be a non-empty string array`);
  }
  if (new Set(values).size !== values.length) fail(`${label} contains duplicates`);
}

function requireTracked(root, tracked, path, label) {
  if (!tracked.has(path)) fail(`${label} is not tracked: ${path}`);
  if (!existsSync(join(root, path))) fail(`${label} is absent: ${path}`);
}

function requirePrivate(tracked, ignored, path, label) {
  if (tracked.has(path)) fail(`${label} must remain untracked: ${path}`);
  if (!ignored.has(path)) fail(`${label} lacks an ignore contract: ${path}`);
}

function classifiedByBatch(root, batchFiles, path) {
  const fullNeedle = `\`${path}\``;
  const basenameNeedle = `\`${path.split("/").at(-1)}\``;
  return batchFiles.find((batch) => {
    const body = readFileSync(join(root, batch), "utf8");
    return body.includes(fullNeedle) || body.includes(basenameNeedle);
  });
}

function normalizedRepoPath(root, source, rawTarget) {
  const target = rawTarget.trim().replace(/^<|>$/g, "").split("#", 1)[0].split("?", 1)[0];
  if (!target || /^(?:https?:|mailto:|#)/.test(target) || isAbsolute(target)) return null;
  const absolute = resolve(root, dirname(source), target);
  const repoPath = normalize(relative(root, absolute)).replaceAll("\\", "/");
  return repoPath.startsWith("../") ? null : repoPath;
}

function absentAuthorityReferences(root, manifest, tracked) {
  const references = [];
  const explicitPath = /`((?:design|planning|rfc|docs|balance|client|server|testdata)\/[A-Za-z0-9_@./+-]+\.(?:md|json|tsv|yaml|yml))`/g;
  const markdownLink = /\]\(([^)]+)\)/g;

  for (const source of manifest.authorityFiles) {
    const body = readFileSync(join(root, source), "utf8");
    for (const match of body.matchAll(explicitPath)) {
      const target = normalize(match[1]).replaceAll("\\", "/");
      if (!tracked.has(target)) references.push({ source, target });
    }
    for (const match of body.matchAll(markdownLink)) {
      const target = normalizedRepoPath(root, source, match[1]);
      if (target && !tracked.has(target)) references.push({ source, target });
    }
  }
  return references;
}

function validate(root, manifest) {
  const tracked = trackedFiles(root);
  const counts = manifest.expectedCounts ?? {};
  uniqueStrings(manifest.authorityFiles, "authorityFiles");
  uniqueStrings(manifest.publicSupportingResearch, "publicSupportingResearch");
  uniqueStrings(manifest.publicClassC, "publicClassC");
  uniqueStrings(manifest.privateClassC, "privateClassC");

  if (manifest.version !== 1) fail(`unsupported manifest version: ${manifest.version}`);
  if (manifest.publicClassC.length !== counts.publicClassC) fail("public Class-C count drift");
  if (manifest.privateClassC.length !== counts.privateClassC) fail("private Class-C count drift");
  if (manifest.publicClassC.length + manifest.privateClassC.length !== counts.classC) {
    fail("Class-C denominator drift");
  }
  if (manifest.duplicateDrafts.length !== counts.duplicateDrafts) fail("duplicate-draft count drift");
  if (manifest.diagnostics.length !== counts.diagnostics) fail("diagnostic count drift");

  const overlap = manifest.publicClassC.filter((path) => manifest.privateClassC.includes(path));
  if (overlap.length > 0) fail(`public/private Class-C overlap: ${overlap.join(", ")}`);

  const authorityReferences = absentAuthorityReferences(root, manifest, tracked);
  const privatePaths = [
    ...manifest.privateClassC,
    ...manifest.privateClassB.map((item) => item.path),
    ...manifest.duplicateDrafts.map((item) => item.path),
    ...manifest.historicalMoves.map((item) => item.path),
    ...manifest.diagnostics.map((item) => item.path),
  ];
  const ignored = ignoredPaths(root, [...privatePaths, ...authorityReferences.map(({ target }) => target)]);

  for (const path of manifest.authorityFiles) requireTracked(root, tracked, path, "authority file");
  for (const path of manifest.publicSupportingResearch) {
    requireTracked(root, tracked, path, "public supporting research");
  }

  const batchFiles = [...tracked].filter((path) =>
    /^planning\/platform-alignment\/publication-rights-batch-[0-9]+\.md$/.test(path),
  );
  if (batchFiles.length === 0) fail("no tracked publication-rights batch records");

  for (const path of manifest.publicClassC) {
    requireTracked(root, tracked, path, "public Class-C dossier");
    if (!classifiedByBatch(root, batchFiles, path)) fail(`public dossier lacks review contract: ${path}`);
  }
  for (const path of manifest.privateClassC) {
    requirePrivate(tracked, ignored, path, "private Class-C source");
    if (!classifiedByBatch(root, batchFiles, path)) fail(`private source lacks review contract: ${path}`);
  }

  for (const item of manifest.privateClassB) {
    requirePrivate(tracked, ignored, item.path, "private Class-B source");
    requireTracked(root, tracked, item.publicDerivative, "Class-B public derivative");
    requireTracked(root, tracked, item.contract, "Class-B source contract");
    if (!readFileSync(join(root, item.contract), "utf8").includes(`\`${item.path}\``)) {
      fail(`Class-B contract does not name source: ${item.path}`);
    }
  }

  for (const item of manifest.duplicateDrafts) {
    requirePrivate(tracked, ignored, item.path, "duplicate draft");
    requireTracked(root, tracked, item.canonical, "duplicate canonical artifact");
  }
  for (const item of manifest.historicalMoves) {
    requirePrivate(tracked, ignored, item.path, "historical pre-move path");
    requireTracked(root, tracked, item.canonical, "historical moved record");
  }
  for (const item of manifest.diagnostics) {
    requirePrivate(tracked, ignored, item.path, "generated diagnostic");
    requireTracked(root, tracked, item.canonicalRecord, "diagnostic canonical record");
  }

  const expectedResearch = new Set([
    "design/research/README.md",
    ...manifest.publicSupportingResearch,
    ...manifest.publicClassC,
  ]);
  const trackedResearch = [...tracked].filter((path) => /^design\/research\/[^/]+\.md$/.test(path));
  const unexpectedResearch = trackedResearch.filter((path) => !expectedResearch.has(path));
  const absentResearch = [...expectedResearch].filter((path) => !tracked.has(path));
  if (unexpectedResearch.length > 0) fail(`unclassified tracked research: ${unexpectedResearch.join(", ")}`);
  if (absentResearch.length > 0) fail(`manifest public research absent: ${absentResearch.join(", ")}`);

  const registeredAbsent = new Set(privatePaths);
  const unknownReferences = authorityReferences.filter(
    ({ target }) => ignored.has(target) && !registeredAbsent.has(target),
  );
  if (unknownReferences.length > 0) {
    fail(`authority references uncontracted ignored artifacts: ${unknownReferences
      .map(({ source, target }) => `${source} -> ${target}`)
      .join(", ")}`);
  }

  return {
    publicClassC: manifest.publicClassC.length,
    privateClassC: manifest.privateClassC.length,
    duplicates: manifest.duplicateDrafts.length,
    diagnostics: manifest.diagnostics.length,
  };
}

function expectRejected(root, manifest, mutate, label, pattern) {
  const forged = structuredClone(manifest);
  mutate(forged);
  try {
    validate(root, forged);
  } catch (error) {
    if (!pattern.test(error.message)) fail(`${label} rejected for the wrong reason: ${error.message}`);
    process.stdout.write(`negative control rejected (${label}): ${error.message}\n`);
    return;
  }
  fail(`negative control was accepted: ${label}`);
}

function runSelfTests(root, manifest) {
  expectRejected(
    root,
    manifest,
    (forged) => { forged.publicClassC[0] = "design/research/forged-missing-public.md"; },
    "missing public authority",
    /public Class-C dossier (?:is not tracked|is absent)/,
  );
  expectRejected(
    root,
    manifest,
    (forged) => { forged.diagnostics[0].canonicalRecord = forged.diagnostics[0].path; },
    "diagnostic promoted without canonical record",
    /diagnostic canonical record is not tracked/,
  );
  expectRejected(
    root,
    manifest,
    (forged) => { forged.privateClassC.pop(); },
    "private-source denominator truncation",
    /private Class-C count drift/,
  );
}

function runFreshClone(root) {
  const temporary = mkdtempSync(join(tmpdir(), "cloud-clicker-authority-"));
  const clone = join(temporary, "clone");
  try {
    const result = spawnSync("git", ["clone", "--quiet", "--no-hardlinks", root, clone], {
      cwd: root,
      encoding: "utf8",
    });
    if (result.status !== 0) fail(`fresh clone failed: ${(result.stderr || result.stdout).trim()}`);
    const check = spawnSync(
      process.execPath,
      [join(clone, "tools/verify-publication-authority.mjs"), "--root", clone, "--self-test"],
      { cwd: clone, encoding: "utf8" },
    );
    process.stdout.write(check.stdout);
    process.stderr.write(check.stderr);
    if (check.status !== 0) fail(`fresh-clone authority check exited ${check.status}`);
  } finally {
    rmSync(temporary, { recursive: true, force: true });
  }
}

const options = parseArgs(process.argv.slice(2));
if (options.freshClone) {
  runFreshClone(options.root);
  process.stdout.write("fresh-clone publication authority: PASS\n");
} else {
  const manifestPath = isAbsolute(options.manifest)
    ? options.manifest
    : join(options.root, options.manifest);
  const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
  if (options.selfTest) runSelfTests(options.root, manifest);
  const result = validate(options.root, manifest);
  process.stdout.write(`publication authority: PASS (${result.publicClassC} public Class-C, ${result.privateClassC} private Class-C, ${result.duplicates} duplicates, ${result.diagnostics} diagnostics)\n`);
}
