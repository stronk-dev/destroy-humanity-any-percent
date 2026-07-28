import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import Ajv2020 from "ajv/dist/2020.js";

const clientDirectory = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryDirectory = path.resolve(clientDirectory, "..");
const balanceDirectory = path.join(repositoryDirectory, "balance");
const schemaPath = path.join(balanceDirectory, "economy.schema.json");

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
  }

  for (const filename of negative) {
    const data = await readJSON(filename);
    if (validate(data)) {
      throw new Error(`${path.relative(repositoryDirectory, filename)}: expected schema rejection`);
    }
  }

  console.log(
    `schema ok: 1 schema, ${production.length} production catalog(s), ${positive.length} positive fixture(s), ${negative.length} negative fixture(s)`,
  );
}

main().catch((error) => {
  console.error(error.message);
  process.exitCode = 1;
});
