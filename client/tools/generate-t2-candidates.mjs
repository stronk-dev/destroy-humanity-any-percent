// Derives the Tier 2 content candidates (rfc/tier2-content.md §A1, §A2, §B1–§B3) from the
// epoch-8 active bytes by insertion only, so the owner-ratified candidate is a plain copy at
// mint. Headcount (§H) and the `headcount_seats` roles are held on OD-1 (T2-DG-1).
// Usage: node client/tools/generate-t2-candidates.mjs [--check]
import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const read = (relative) => readFileSync(path.join(root, relative), "utf8");
const check = process.argv.includes("--check");

function insertOnce(source, anchor, insertion, label) {
  const index = source.indexOf(anchor);
  if (index < 0 || source.indexOf(anchor, index + 1) >= 0) throw new Error(`${label}: anchor must occur exactly once`);
  return source.slice(0, index + anchor.length) + insertion + source.slice(index + anchor.length);
}

const ladder = (a, b, c) => `      "ladder": [
        { "purchased_at": ${a}, "multiplier_ppm": 2000000 },
        { "purchased_at": ${b}, "multiplier_ppm": 2000000 },
        { "purchased_at": ${c}, "multiplier_ppm": 2000000 }
      ],`;

const generators = `
    {
      "id": "generator.open_plan_floor",
      "tier": 2,
      "category": "category.office",
      "price": { "resource_id": "company.cash", "base": "1.2e7", "curve": { "kind": "geometric", "ratio": "1.13e0" } },
      "production": { "resource_id": "company.cash", "base_rate": "2e4" },
${ladder(25, 50, 100)}
      "roles": [{ "kind": "synergy_feed", "pool_id": "pool.org_chart" }]
    },
    {
      "id": "generator.managed_services_contract",
      "tier": 2,
      "category": "category.software",
      "price": { "resource_id": "company.cash", "base": "6e7", "curve": { "kind": "geometric", "ratio": "1.11e0" } },
      "production": { "resource_id": "company.cash", "base_rate": "1.3e5" },
      "provisions": { "generator_id": "generator.garage_rack", "rate_ppm": 100000 },
${ladder(20, 50, 100)}
      "roles": [
        { "kind": "provision", "generator_id": "generator.garage_rack" },
        { "kind": "synergy_feed", "pool_id": "pool.org_chart" }
      ]
    },
    {
      "id": "generator.hot_desk_program",
      "tier": 2,
      "category": "category.office",
      "price": { "resource_id": "company.cash", "base": "1.44e8", "curve": { "kind": "geometric", "ratio": "1.12e0" } },
      "production": { "resource_id": "company.cash", "base_rate": "2.5e5" },
${ladder(25, 55, 100)}
      "roles": [{ "kind": "synergy_feed", "pool_id": "pool.org_chart" }]
    },`;

const upgrade = (id, amount, target) => `,
    {
      "id": "${id}",
      "cost": { "resource": "company.cash", "amount": "${amount}" },
      "window": { "from_gate": "gate.t1_to_t2", "to_gate": "gate.t2_to_t3" },
      "requires": [{ "kind": "resource_at_least", "resource_id": "company.cash", "value": "${amount}" }],
      "effects": [{ "source_id": "${id}.factor", "slot": "upgrades", "target": "${target}", "factor": "2e0" }],
      "roles": ["synergy_feed"],
      "copy_key": "${id}"
    }`;

const pool = `,
    {
      "id": "pool.org_chart",
      "sources": [
        { "kind": "generator", "id_or_class": "generator.open_plan_floor", "per_count_ppm": 3000 },
        { "kind": "generator", "id_or_class": "generator.managed_services_contract", "per_count_ppm": 3000 },
        { "kind": "generator", "id_or_class": "generator.hot_desk_program", "per_count_ppm": 4000 },
        { "kind": "upgrade", "id_or_class": "upgrade.ping_pong_table", "per_count_ppm": 30000 },
        { "kind": "upgrade", "id_or_class": "upgrade.move_fast_break_things", "per_count_ppm": 30000 },
        { "kind": "upgrade", "id_or_class": "upgrade.nap_pod", "per_count_ppm": 30000 }
      ],
      "slot": "upgrades",
      "curve": "log"
    }`;

let economy = read("balance/catalogs/phase0.json");
economy = insertOnce(economy,
  `"production": { "resource_id": "company.cash", "base_rate": "1.7850625e3" },`,
  `\n      "provisioned_hardcap": { "count": 9007199254740991, "reason_key": "generator.garage_rack.provisioned_cap" },`,
  "garage_rack hardcap");
economy = insertOnce(economy, `      "roles": [{ "kind": "provision", "generator_id": "generator.beige_tower" }]
    },`, generators, "T2 generators");
economy = insertOnce(economy, `      "copy_key": "upgrade.institutional_memory"
    }`, upgrade("upgrade.ping_pong_table", "3e7", "generator.open_plan_floor") +
  upgrade("upgrade.move_fast_break_things", "1.5e8", "generator.managed_services_contract") +
  upgrade("upgrade.nap_pod", "3.6e8", "generator.hot_desk_program"), "T2 upgrades");
economy = insertOnce(economy, `      "slot": "upgrades",
      "curve": "log"
    }`, pool, "T2 pool");
economy = insertOnce(economy,
  `{ "id": "pool.operational_excellence", "slot": "upgrades", "target": "all", "provider": "pool.operational_excellence" }`,
  `,\n    { "id": "pool.org_chart", "slot": "upgrades", "target": "all", "provider": "pool.org_chart" }`,
  "T2 multiplier source");

let routes = read("balance/routes/phase0.json");
routes = insertOnce(routes, `      "requirement": [{ "resource_id": "company.cash", "amount": "1e5" }],
      "routes": []
    },`, `
    {
      "gate_id": "gate.t1_to_t2",
      "requirement": [{ "resource_id": "company.cash", "amount": "1e7" }],
      "routes": []
    },`, "T1→T2 gate");

let categories = read("balance/categories/phase0.json");
categories = insertOnce(categories, `"full_gate_set": ["gate.t0_to_t1", `, `"gate.t1_to_t2", `, "categories gate set");

// Presentation schema v3 carry with the Tier-2 bindings (§E4), so minting cannot leave a
// rendered generator, upgrade, or previewed gate without its copy binding.
let presentation = read("balance/testdata/t0-t1/presentation-v3.json");
presentation = insertOnce(presentation,
  `"id": "generator.garage_rack", "title_key": "generator.garage_rack.title", "description_key": "generator.garage_rack.description", "cap_reason_key": null`,
  "", "garage_rack presentation anchor");
presentation = presentation.replace(
  `"id": "generator.garage_rack", "title_key": "generator.garage_rack.title", "description_key": "generator.garage_rack.description", "cap_reason_key": null`,
  `"id": "generator.garage_rack", "title_key": "generator.garage_rack.title", "description_key": "generator.garage_rack.description", "cap_reason_key": "generator.garage_rack.provisioned_cap"`);
const generatorBinding = (id) => `    { "id": "${id}", "title_key": "${id}.title", "description_key": "${id}.description", "cap_reason_key": null },\n`;
const upgradeBinding = (id) => `    { "id": "${id}", "title_key": "${id}.title", "description_key": "${id}.description" },\n`;
presentation = insertOnce(presentation, `"cap_reason_key": "generator.garage_rack.provisioned_cap" },\n`,
  generatorBinding("generator.hot_desk_program"), "hot desk presentation");
presentation = insertOnce(presentation, `"id": "generator.legal_dept", "title_key": "generator.legal_dept.title", "description_key": "generator.legal_dept.description", "cap_reason_key": null },\n`,
  generatorBinding("generator.managed_services_contract"), "managed services presentation");
presentation = insertOnce(presentation, `"id": "generator.nephew_intern", "title_key": "generator.nephew_intern.title", "description_key": "generator.nephew_intern.description", "cap_reason_key": null }`,
  `,\n${generatorBinding("generator.open_plan_floor").trimEnd().replace(/,$/, "")}`, "open plan presentation");
presentation = insertOnce(presentation, `"id": "upgrade.institutional_memory", "title_key": "upgrade.institutional_memory.title", "description_key": "upgrade.institutional_memory.description" },\n`,
  upgradeBinding("upgrade.move_fast_break_things") + upgradeBinding("upgrade.nap_pod"), "move fast / nap pod presentation");
presentation = insertOnce(presentation, `"id": "upgrade.nephew_business_cards", "title_key": "upgrade.nephew_business_cards.title", "description_key": "upgrade.nephew_business_cards.description" },\n`,
  upgradeBinding("upgrade.ping_pong_table"), "ping pong presentation");
presentation = insertOnce(presentation, `{ "id": "gate.t0_to_t1", "title_key": "gate.t0_to_t1.title" }`,
  `,\n    { "id": "gate.t1_to_t2", "title_key": "gate.t1_to_t2.title" }`, "T1→T2 gate presentation");

// §P3 relevance: the six Tier-2 rows (window gate.t1_to_t2 → gate.t2_to_t3, epsilon 1000 ms,
// no trap exemptions) inserted into the epoch-8 policy in raw-byte order.
let relevance = read("balance/relevance/t0-t1.json");
const relevanceRow = (id) => `    { "purchasable_id": "${id}", "availability_window": { "from_gate": "gate.t1_to_t2", "to_gate": "gate.t2_to_t3" }, "epsilon_ms": 1000, "trap_exempt": false, "justification_key": null, "group_ids": [] },\n`;
const insertRelevanceBefore = (nextID, rows) => {
  const anchor = `    { "purchasable_id": "${nextID}",`;
  const index = relevance.indexOf(anchor);
  if (index < 0 || relevance.indexOf(anchor, index + 1) >= 0) throw new Error(`relevance anchor ${nextID} must occur exactly once`);
  relevance = relevance.slice(0, index) + rows.map(relevanceRow).join("") + relevance.slice(index);
};
insertRelevanceBefore("generator.nephew_intern", ["generator.hot_desk_program", "generator.managed_services_contract"]);
insertRelevanceBefore("upgrade.beige_tower_cache", ["generator.open_plan_floor"]);
insertRelevanceBefore("upgrade.nephew_business_cards", ["upgrade.move_fast_break_things", "upgrade.nap_pod"]);
insertRelevanceBefore("upgrade.rack_rail_standardization", ["upgrade.ping_pong_table"]);

const outputs = {
  "balance/testdata/t2/economy-candidate-v1.json": economy,
  "balance/testdata/t2/routes-candidate-v1.json": routes,
  "balance/testdata/t2/categories-candidate-v1.json": categories,
  "balance/testdata/t2/presentation-candidate-v3.json": presentation,
  "balance/testdata/t2/relevance-candidate-v1.json": relevance,
};
const sha = Object.entries(outputs).map(([file, bytes]) => `${createHash("sha256").update(bytes).digest("hex")}  ${file}`).join("\n") + "\n";
outputs["balance/testdata/t2/candidates.sha256"] = sha;

let drift = [];
for (const [file, bytes] of Object.entries(outputs)) {
  const target = path.join(root, file);
  if (check) {
    let current = "";
    try { current = readFileSync(target, "utf8"); } catch { /* missing */ }
    if (current !== bytes) drift.push(file);
  } else writeFileSync(target, bytes);
}
if (drift.length) { console.error(`T2 candidate drift: ${drift.join(", ")}`); process.exit(1); }
console.log(check ? "T2 candidates ok" : "generated T2 candidates");
