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

function maxRoutesPerRun(catalog) {
  const valuesBySlot = new Map();
  const routes = [];
  for (const gate of catalog.gates ?? []) {
    for (const route of gate.routes ?? []) {
      routes.push(route);
      const values = valuesBySlot.get(route.exclusion_slot) ?? new Set();
      values.add(route.exclusion_value);
      valuesBySlot.set(route.exclusion_slot, values);
    }
  }
  const slots = [...valuesBySlot.keys()].sort();
  let maximum = 0;
  const assignment = new Map();
  const search = (index) => {
    if (index === slots.length) {
      maximum = Math.max(maximum, routes.filter((route) => assignment.get(route.exclusion_slot) === route.exclusion_value).length);
      return;
    }
    const slot = slots[index];
    for (const value of [...valuesBySlot.get(slot)].sort()) {
      assignment.set(slot, value);
      search(index + 1);
    }
  };
  search(0);
  return maximum;
}

function routeSemanticErrors(catalog) {
  const errors = [];
  for (const [gateIndex, gate] of (catalog.gates ?? []).entries()) {
    for (const [routeIndex, route] of (gate.routes ?? []).entries()) {
      if (route.active && route.requires_context_version > catalog.context_version) {
        errors.push(`/gates/${gateIndex}/routes/${routeIndex} active route requires unavailable context`);
      }
      if (route.effect?.kind === "discount") {
        const fraction = new Decimal(route.effect.fraction);
        if (!fraction.gt(0) || !fraction.lt(1)) errors.push(`/gates/${gateIndex}/routes/${routeIndex}/effect/fraction must be in (0,1)`);
      }
    }
  }
  if (maxRoutesPerRun(catalog) >= catalog.depletion_distinct_routes_required) {
    errors.push("depletion is reachable in one run");
  }
  return errors;
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

  const routesSchema = await readJSON(path.join(balanceDirectory, "routes.schema.json"));
  const validateRoutes = ajv.compile(routesSchema);
  const routeCatalogs = await jsonFiles(path.join(balanceDirectory, "routes"));
  const validRoutes = await jsonFiles(path.join(balanceDirectory, "routes-testdata", "valid"));
  const invalidRoutes = await jsonFiles(path.join(balanceDirectory, "routes-testdata", "invalid"));
  if (routeCatalogs.length === 0 || validRoutes.length === 0 || invalidRoutes.length === 0) {
    throw new Error("routes schema verification requires production, positive, and negative catalogs");
  }
  for (const filename of [...routeCatalogs, ...validRoutes]) {
    const data = await readJSON(filename);
    if (!validateRoutes(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateRoutes)}`);
    const errors = routeSemanticErrors(data);
    if (errors.length > 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${errors.join("; ")}`);
  }
  for (const filename of invalidRoutes) {
    const data = await readJSON(filename);
    const shapeValid = validateRoutes(data);
    const errors = shapeValid ? routeSemanticErrors(data) : [];
    if (shapeValid && errors.length === 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: expected routes rejection`);
  }

  const harnessDirectory = path.join(repositoryDirectory, "testdata", "harness");
  const scenarioSchema = await readJSON(path.join(harnessDirectory, "scenario.schema.json"));
  const reportSchema = await readJSON(path.join(harnessDirectory, "report.schema.json"));
  const validateScenario = ajv.compile(scenarioSchema);
  const validateReport = ajv.compile(reportSchema);
  const scenarios = await jsonFiles(path.join(harnessDirectory, "scenarios"));
  const invalidScenarios = await jsonFiles(path.join(harnessDirectory, "invalid"));
  if (scenarios.length === 0 || invalidScenarios.length === 0) {
    throw new Error("harness schema verification requires positive and negative scenarios");
  }
  for (const filename of scenarios) {
    if (!validateScenario(await readJSON(filename))) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateScenario)}`);
    }
  }
  for (const filename of invalidScenarios) {
    if (validateScenario(await readJSON(filename))) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: expected harness scenario rejection`);
    }
  }
  const baseline = path.join(harnessDirectory, "pacing-baseline.json");
  if (!validateReport(await readJSON(baseline))) {
    throw new Error(`${path.relative(repositoryDirectory, baseline)}: ${validationErrors(validateReport)}`);
  }

  console.log(
    `schema ok: economy + routes + harness, ${production.length} economy catalog(s), ${routeCatalogs.length} routes catalog(s), ${scenarios.length} scenario(s)`,
  );
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
