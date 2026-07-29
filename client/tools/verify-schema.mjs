import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import Ajv2020 from "ajv/dist/2020.js";
import Decimal from "break_infinity.js";

const clientDirectory = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryDirectory = path.resolve(clientDirectory, "..");
const balanceDirectory = path.join(repositoryDirectory, "balance");
const schemaPath = path.join(balanceDirectory, "economy.schema.json");
const minimumResourceLogTarget = new Decimal("5e-15");

async function jsonFiles(directory) {
  let entries;
  try {
    entries = await readdir(directory, { withFileTypes: true });
  } catch (error) {
    if (error?.code === "ENOENT") return [];
    throw error;
  }

  const files = [];
  for (const entry of entries.sort((left, right) => left.name.localeCompare(right.name))) {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await jsonFiles(entryPath)));
    else if (entry.isFile() && entry.name.endsWith(".json")) files.push(entryPath);
  }
  return files;
}

async function readJSON(filename) {
  try {
    return JSON.parse(await readFile(filename, "utf8"));
  } catch (error) {
    throw new Error(`${path.relative(repositoryDirectory, filename)}: ${error.message}`);
  }
}

function validationErrors(validate) {
  return (validate.errors ?? [])
    .map((error) => `${error.instancePath || "/"} ${error.message}`)
    .join("; ");
}

function resourceLogSemanticErrors(catalog) {
  const errors = [];
  for (const [coordinateIndex, coordinate] of (catalog.progress_coordinates ?? []).entries()) {
    const terms = coordinate.kind === "composite" ? coordinate.terms : [coordinate];
    for (const [termIndex, term] of (terms ?? []).entries()) {
      if (term.kind !== "resource_log") continue;
      const path =
        coordinate.kind === "composite"
          ? `/progress_coordinates/${coordinateIndex}/terms/${termIndex}/target`
          : `/progress_coordinates/${coordinateIndex}/target`;
      const target = new Decimal(term.target);
      const denominator = new Decimal(Decimal.add(1, target).log10());
      if (
        target.lt(minimumResourceLogTarget) ||
        !Number.isFinite(denominator.mantissa) ||
        !Number.isFinite(denominator.exponent) ||
        !denominator.gt(0)
      ) {
        errors.push(`${path} must be at least 5e-15 with a finite positive logarithm`);
      }
    }
  }
  return errors;
}

async function verifyResourceLogSource() {
  const source = await readFile(path.join(clientDirectory, "src", "economy-kernel.ts"), "utf8");
  const match = source.match(/function resourceLogProgress[\s\S]*?\n}\n\nfunction countFractionProgress/);
  if (
    !match ||
    !match[0].includes("numerator.div(denominator)") ||
    /log10\(\)\s*\//.test(match[0])
  ) {
    throw new Error(
      "client resourceLogProgress must divide Decimal logarithms with numerator.div(denominator)",
    );
  }
}

async function main() {
  const schema = await readJSON(schemaPath);
  const ajv = new Ajv2020({ allErrors: true, strict: true });
  const validate = ajv.compile(schema);

  const production = await jsonFiles(path.join(balanceDirectory, "catalogs"));
  const positive = await jsonFiles(path.join(balanceDirectory, "testdata", "valid"));
  const negative = await jsonFiles(path.join(balanceDirectory, "testdata", "invalid"));

  if (positive.length === 0 || negative.length === 0) {
    throw new Error("schema verification requires at least one positive and one negative fixture");
  }

  for (const filename of [...production, ...positive]) {
    const data = await readJSON(filename);
    if (!validate(data)) {
      throw new Error(
        `${path.relative(repositoryDirectory, filename)}: expected valid catalog: ${validationErrors(validate)}`,
      );
    }
    const semanticErrors = resourceLogSemanticErrors(data);
    if (semanticErrors.length > 0) {
      throw new Error(
        `${path.relative(repositoryDirectory, filename)}: expected valid catalog: ${semanticErrors.join("; ")}`,
      );
    }
  }

  for (const filename of negative) {
    const data = await readJSON(filename);
    const shapeValid = validate(data);
    const semanticErrors = shapeValid ? resourceLogSemanticErrors(data) : [];
    if (shapeValid && semanticErrors.length === 0) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: expected schema rejection`);
    }
  }

  await verifyResourceLogSource();

  console.log(
    `schema ok: 1 shape schema + resource_log semantics/source, ${production.length} production catalog(s), ${positive.length} positive fixture(s), ${negative.length} negative fixture(s)`,
  );
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
