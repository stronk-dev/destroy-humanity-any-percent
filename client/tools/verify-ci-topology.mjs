import { readFileSync } from "node:fs";
import { verifyCITopology } from "./ci-topology.mjs";

const root = new URL("../../", import.meta.url);
const ci = readFileSync(new URL(".github/workflows/ci.yml", root), "utf8");
const maintenance = readFileSync(new URL(".github/workflows/maintenance.yml", root), "utf8");
verifyCITopology(ci, maintenance);
console.log("CI fast/maintenance topology ok");
