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

function routeSemanticErrors(catalog, resourceIDs) {
  const errors = [];
  for (const [gateIndex, gate] of (catalog.gates ?? []).entries()) {
    for (const [requirementIndex, requirement] of (gate.requirement ?? []).entries()) {
      if (!resourceIDs.has(requirement.resource_id)) errors.push(`/gates/${gateIndex}/requirement/${requirementIndex} references unknown company resource`);
    }
    for (const [routeIndex, route] of (gate.routes ?? []).entries()) {
      const doctrineTransitions = (route.predicate ?? [])
        .filter((condition) => condition.kind === "doctrine_is" || condition.kind === "doctrine_is_not")
        .map((condition) => condition.transition);
      if (doctrineTransitions.length > 0) {
        const gateTier = adjacentBoundaryStart(gate.gate_id, "gate");
        if (gateTier === undefined) errors.push(`/gates/${gateIndex} doctrine-bearing route requires a canonical adjacent tier gate`);
        for (const transition of doctrineTransitions) {
          const transitionTier = adjacentBoundaryStart(transition, "transition");
          if (transitionTier === undefined || gateTier === undefined || gateTier < transitionTier) errors.push(`/gates/${gateIndex}/routes/${routeIndex} doctrine transition occurs after gate`);
        }
      }
      if (route.active && route.requires_context_version > catalog.context_version) {
        errors.push(`/gates/${gateIndex}/routes/${routeIndex} active route requires unavailable context`);
      }
      if (route.effect?.kind === "discount") {
        const fraction = new Decimal(route.effect.fraction);
        if (!fraction.gt(0) || !fraction.lt(1)) errors.push(`/gates/${gateIndex}/routes/${routeIndex}/effect/fraction must be in (0,1)`);
      }
      if ((route.predicate ?? []).some((condition) => condition.kind === "meter_band" || condition.kind === "region_trait") && route.requires_context_version < 2) {
        errors.push(`/gates/${gateIndex}/routes/${routeIndex} meter/region condition requires context v2`);
      }
      const exclusionMatched = route.exclusion_slot === "structure"
        ? (route.predicate ?? []).some((condition) => condition.kind === "structure_is" && condition.structure_id === route.exclusion_value)
        : (route.predicate ?? []).some((condition) => condition.kind === "doctrine_is" && `doctrine:${condition.transition}` === route.exclusion_slot && condition.doctrine_id === route.exclusion_value);
      if (!exclusionMatched) errors.push(`/gates/${gateIndex}/routes/${routeIndex} exclusion is not an executable predicate`);
      for (const [conditionIndex, condition] of (route.predicate ?? []).entries()) {
        if ((condition.kind === "resource_at_least" || condition.kind === "resource_at_most") && !resourceIDs.has(condition.resource_id)) {
          errors.push(`/gates/${gateIndex}/routes/${routeIndex}/predicate/${conditionIndex} references unknown company resource`);
        }
      }
    }
  }
  if (maxRoutesPerRun(catalog) >= catalog.depletion_distinct_routes_required) {
    errors.push("depletion is reachable in one run");
  }
  return errors;
}

function adjacentBoundaryStart(value, prefix) {
  const match = new RegExp(`^${prefix}\\.t([0-9]+)_to_t([0-9]+)$`).exec(value);
  if (!match) return undefined;
  const from = Number(match[1]);
  const to = Number(match[2]);
  return Number.isSafeInteger(from) && Number.isSafeInteger(to) && to === from + 1 ? from : undefined;
}

function harnessSemanticErrors(scenario) {
  const errors = [];
  const policies = new Set((scenario.runs ?? []).map((run) => run.policy_id));
  const milestones = new Set((scenario.milestones ?? []).map((milestone) => milestone.id));
  const seen = new Set();
  for (const [index, envelope] of (scenario.envelopes ?? []).entries()) {
    if (!policies.has(envelope.policy_id)) errors.push(`/envelopes/${index} references unknown policy`);
    if (!milestones.has(envelope.milestone_id)) errors.push(`/envelopes/${index} references unknown milestone`);
    const key = `${envelope.policy_id}\0${envelope.milestone_id}\0${envelope.statistic}`;
    if (seen.has(key)) errors.push(`/envelopes/${index} duplicates an observation tuple`);
    seen.add(key);
  }
  for (const policy of policies) {
    for (const milestone of milestones) {
      for (const statistic of ["p50", "p95"]) {
        const key = `${policy}\0${milestone}\0${statistic}`;
        if (!seen.has(key)) errors.push(`missing pacing observation ${policy}/${milestone}/${statistic}`);
      }
    }
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

  const companyResourceIDs = new Set();
  for (const filename of production) {
    const catalog = await readJSON(filename);
    for (const resource of catalog.resources ?? []) if (resource.scope === "company") companyResourceIDs.add(resource.id);
  }

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
    const errors = routeSemanticErrors(data, companyResourceIDs);
    if (errors.length > 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${errors.join("; ")}`);
  }
  for (const filename of invalidRoutes) {
    const data = await readJSON(filename);
    const shapeValid = validateRoutes(data);
    const errors = shapeValid ? routeSemanticErrors(data, companyResourceIDs) : [];
    if (shapeValid && errors.length === 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: expected routes rejection`);
  }

  const commonsSchema = await readJSON(path.join(balanceDirectory, "commons.schema.json"));
  const validateCommons = ajv.compile(commonsSchema);
  const commonsCatalogs = await jsonFiles(path.join(balanceDirectory, "commons"));
  if (commonsCatalogs.length === 0) throw new Error("commons schema verification requires a production catalog");
  for (const filename of commonsCatalogs) {
    const data = await readJSON(filename);
    if (!validateCommons(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateCommons)}`);
    const catalog = parseCommonsPolicy(data);
    const economySources = new Map();
    for (const economyFile of production) for (const source of (await readJSON(economyFile)).multiplier_sources ?? []) economySources.set(source.id, source);
    for (const weight of catalog.source_weights) {
      const source = economySources.get(weight.source_id);
      if (!source || source.slot !== weight.slot || source.slot === "commons") throw new Error(`${path.relative(repositoryDirectory, filename)}: source weight does not match economy declaration`);
    }
    const commonsSource = economySources.get("commons.member");
    if (!commonsSource || commonsSource.slot !== "commons" || commonsSource.target !== "all" || commonsSource.provider !== "commons") throw new Error("economy catalog must declare the single commons.member provider");
  }

  const factionSchema = await readJSON(path.join(balanceDirectory, "factions.schema.json"));
  const validateFaction = ajv.compile(factionSchema);
  const factionCatalogs = await jsonFiles(path.join(balanceDirectory, "factions"));
  if (factionCatalogs.length === 0) throw new Error("faction schema verification requires a production catalog");
  const commonsPolicy = parseCommonsPolicy(await readJSON(commonsCatalogs[0]));
  for (const filename of factionCatalogs) {
    const data = await readJSON(filename);
    if (!validateFaction(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateFaction)}`);
    const errors = factionSemanticErrors(data, commonsPolicy);
    if (errors.length > 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${errors.join("; ")}`);
  }

  const shellSchema = await readJSON(path.join(balanceDirectory, "client-shell.schema.json"));
  const validateShell = ajv.compile(shellSchema);
  const shellCatalogs = await jsonFiles(path.join(balanceDirectory, "client-shell"));
  if (shellCatalogs.length === 0) throw new Error("client shell schema verification requires a production catalog");
  for (const filename of shellCatalogs) {
    const data = await readJSON(filename);
    if (!validateShell(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateShell)}`);
  }

  const prestigeSchema = await readJSON(path.join(balanceDirectory, "prestige.schema.json"));
  const validatePrestige = ajv.compile(prestigeSchema);
  const prestigeCatalogs = await jsonFiles(path.join(balanceDirectory, "prestige"));
  if (prestigeCatalogs.length === 0) throw new Error("prestige schema verification requires a production catalog");
  for (const filename of prestigeCatalogs) {
    const data = await readJSON(filename);
    if (!validatePrestige(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validatePrestige)}`);
    const threshold = new Decimal(data.threshold);
    if (!threshold.gt(0) || !Number.isFinite(threshold.mantissa) || !Number.isSafeInteger(threshold.exponent)) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: threshold must be a positive state Decimal`);
    }
  }

  const transportSchema = await readJSON(path.join(balanceDirectory, "transport.schema.json"));
  const validateTransport = ajv.compile(transportSchema);
  const transportCatalogs = await jsonFiles(path.join(balanceDirectory, "transport"));
  if (transportCatalogs.length === 0) throw new Error("transport schema verification requires a production policy");
  for (const filename of transportCatalogs) {
    const data = await readJSON(filename);
    if (!validateTransport(data)) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateTransport)}`);
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
    const data = await readJSON(filename);
    if (!validateScenario(data)) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: ${validationErrors(validateScenario)}`);
    }
    const errors = harnessSemanticErrors(data);
    if (errors.length > 0) throw new Error(`${path.relative(repositoryDirectory, filename)}: ${errors.join("; ")}`);
  }
  for (const filename of invalidScenarios) {
    const data = await readJSON(filename);
    const shapeValid = validateScenario(data);
    const errors = shapeValid ? harnessSemanticErrors(data) : [];
    if (shapeValid && errors.length === 0) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: expected harness scenario rejection`);
    }
  }
  const baseline = path.join(harnessDirectory, "pacing-baseline.json");
  if (!validateReport(await readJSON(baseline))) {
    throw new Error(`${path.relative(repositoryDirectory, baseline)}: ${validationErrors(validateReport)}`);
  }

  console.log(
    `schema ok: economy + routes + commons + factions + client-shell + prestige + transport + harness, ${production.length} economy catalog(s), ${routeCatalogs.length} routes catalog(s), ${commonsCatalogs.length} commons catalog(s), ${factionCatalogs.length} faction catalog(s), ${shellCatalogs.length} client-shell catalog(s), ${prestigeCatalogs.length} prestige catalog(s), ${transportCatalogs.length} transport policy(s), ${scenarios.length} scenario(s)`,
  );
}

function parseCommonsPolicy(data) {
  const ppm = (value) => Number.isSafeInteger(value) && value >= 0 && value <= 1_000_000;
  if (data.minimum_tithe_ppm > data.default_tithe_ppm || data.default_tithe_ppm > data.maximum_tithe_ppm ||
      data.guild_health_weight_ppm + data.cohort_health_weight_ppm + data.server_health_weight_ppm !== 1_000_000 ||
      data.healthy_health_ppm <= data.collapse_health_ppm ||
      data.health_recovery_ppm_per_hour <= data.health_decay_ppm_per_hour ||
      data.cohort_merge_floor > data.cohort_target_size || data.npc_population_floor < data.cohort_merge_floor ||
      !ppm(data.collective_weight_ppm) || !Number.isSafeInteger(data.collective_exponent_ppm) ||
      data.collective_exponent_ppm < 1_000_000 || data.collective_exponent_ppm > 10_000_000) throw new Error("invalid commons policy relationship");
  const ids = new Set();
  for (const source of data.source_weights) { if (ids.has(source.source_id)) throw new Error("duplicate commons source weight"); ids.add(source.source_id); }
  return data;
}

function factionSemanticErrors(catalog, commonsPolicy) {
  const errors = [];
  const expectedIDs = ["bootstrapper", "enterprise", "open_source", "vc_funded"];
  const expectedResources = ["compliance", "hype", "libraries", "revenue"];
  const byID = new Map();
  const produced = new Set();
  const consumed = new Set();
  for (const faction of catalog.factions ?? []) {
    if (byID.has(faction.id)) errors.push(`duplicate faction ${faction.id}`);
    if (produced.has(faction.produces)) errors.push(`duplicate produced resource ${faction.produces}`);
    if (consumed.has(faction.consumes)) errors.push(`duplicate consumed resource ${faction.consumes}`);
    if (faction.produces === faction.consumes) errors.push(`${faction.id} consumes its own stock`);
    if (faction.incorporation_copy_key !== `incorporate.${faction.id}`) errors.push(`${faction.id} copy key mismatch`);
    if ((faction.modifier_slots ?? []).length !== 0) errors.push(`${faction.id} has undeclared Phase-0 modifiers`);
    byID.set(faction.id, faction); produced.add(faction.produces); consumed.add(faction.consumes);
  }
  if (JSON.stringify([...byID.keys()].sort()) !== JSON.stringify(expectedIDs)) errors.push("faction id set mismatch");
  if (JSON.stringify([...produced].sort()) !== JSON.stringify(expectedResources) || JSON.stringify([...consumed].sort()) !== JSON.stringify(expectedResources)) errors.push("stock resource set mismatch");
  const openSource = byID.get("open_source");
  if (!openSource?.compact?.auto_sign || openSource.compact.tithe_ppm <= commonsPolicy.default_tithe_ppm ||
      openSource.compact.tithe_ppm < commonsPolicy.minimum_tithe_ppm || openSource.compact.tithe_ppm > commonsPolicy.maximum_tithe_ppm) errors.push("Open Source tithe is outside the Commons contract");
  for (const [id, faction] of byID) if (id !== "open_source" && faction.compact !== null) errors.push(`${id} compact must be null`);
  const consumer = new Map([...byID.values()].map((faction) => [faction.consumes, faction.id]));
  let current = expectedIDs[0]; const visited = new Set();
  for (let index = 0; index < expectedIDs.length; index++) {
    if (visited.has(current)) break;
    visited.add(current); current = consumer.get(byID.get(current)?.produces);
  }
  if (visited.size !== expectedIDs.length || current !== expectedIDs[0]) errors.push("faction stock graph is not one Hamiltonian cycle");
  return errors;
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
