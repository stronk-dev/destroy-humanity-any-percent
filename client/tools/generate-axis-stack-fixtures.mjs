// Generates the Clout v1 / PR Interns (rfc/clout-v1-and-pr-interns.md) fixture
// economy v5 catalog and the shared Go/TS loader corpus. The fixture is the
// pinned epoch-8 economy plus exactly the CV8 proposal rows that exist today
// (pr_intern_1, pr_intern_2; pr_intern_3 waits on Tier 2 content, OD-5).
// Numbers are PROPOSED (CV8), ratified later by owner SHA (OD-5).
import { readFileSync, writeFileSync, mkdirSync } from "node:fs";

const root = new URL("../../", import.meta.url);
const read = (path) => JSON.parse(readFileSync(new URL(path, root), "utf8"));
const write = (path, value) => {
  mkdirSync(new URL(path.slice(0, path.lastIndexOf("/") + 1), root), { recursive: true });
  writeFileSync(new URL(path, root), `${JSON.stringify(value, null, 2)}\n`);
};

const economy = read("balance/catalogs/phase0.json");
if (economy.schema_version !== 4 || "axis_stack" in economy) throw new Error("base economy must be the schema-4 epoch-8 catalog");

const intern = (id, cost, minimum, factorPpm) => ({
  id: `upgrade.${id}`,
  cost: { resource: "company.cash", amount: cost },
  window: { from_gate: "gate.t0_to_t1", to_gate: null },
  requires: [{ kind: "axis_at_least", minimum }],
  effects: [{ source_id: `upgrade.${id}.axis`, slot: "axis_stack", target: "all", factor_ppm: factorPpm }],
  roles: ["synergy_feed"],
  copy_key: `upgrade.${id}`,
});

const fixture = {
  ...economy,
  schema_version: 5,
  upgrades: [...economy.upgrades, intern("pr_intern_1", "1e6", 6, 25_000), intern("pr_intern_2", "5e7", 10, 20_000)],
  multiplier_sources: [
    ...economy.multiplier_sources,
    { id: "upgrade.pr_intern_1.axis", slot: "axis_stack", target: "all", provider: "axis_stack" },
    { id: "upgrade.pr_intern_2.axis", slot: "axis_stack", target: "all", provider: "axis_stack" },
  ],
  axis_stack: { input: "achievement_attainment_run", input_cap: 44, cap_reason_key: "cap.axis_stack_input" },
};
write("balance/testdata/axis-stack/economy-v5-fixture.json", fixture);

const intern1 = { id: "upgrade.pr_intern_1" };
const effect = [ "upgrades", intern1, "effects", 0 ];
const declaration1 = [ "multiplier_sources", { id: "upgrade.pr_intern_1.axis" } ];
const cases = [
  { name: "fixture_accepts", mutations: [], expect: "accept" },
  { name: "schema_4_forbids_axis_stack", mutations: [{ op: "set", path: ["schema_version"], value: 4 }], expect: "reject" },
  { name: "axis_effect_in_upgrades_slot", mutations: [{ op: "set", path: [...effect, "slot"], value: "upgrades" }], expect: "reject" },
  { name: "axis_effect_target_not_all", mutations: [{ op: "set", path: [...effect, "target"], value: "generator.beige_tower" }], expect: "reject" },
  { name: "factor_ppm_zero", mutations: [{ op: "set", path: [...effect, "factor_ppm"], value: 0 }], expect: "reject" },
  { name: "factor_ppm_above_one_million", mutations: [{ op: "set", path: [...effect, "factor_ppm"], value: 1_000_001 }], expect: "reject" },
  { name: "factor_ppm_at_one_million", mutations: [{ op: "set", path: [...effect, "factor_ppm"], value: 1_000_000 }], expect: "accept" },
  { name: "axis_effect_with_static_factor", mutations: [{ op: "set", path: [...effect, "factor"], value: "2e0" }], expect: "reject" },
  { name: "unknown_input", mutations: [{ op: "set", path: ["axis_stack", "input"], value: "achievements_everything" }], expect: "reject" },
  { name: "clout_run_without_ledger", mutations: [{ op: "set", path: ["axis_stack", "input"], value: "clout_run" }], expect: "reject" },
  { name: "score_run_contrast_arm_loads", mutations: [{ op: "set", path: ["axis_stack", "input"], value: "achievement_score_run" }], expect: "accept" },
  { name: "input_cap_zero", mutations: [{ op: "set", path: ["axis_stack", "input_cap"], value: 0 }], expect: "reject" },
  { name: "mixed_arm_upgrade", mutations: [{ op: "append", path: ["upgrades", intern1, "effects"], value: { source_id: "upgrade.pr_intern_1.static", slot: "upgrades", target: "generator.beige_tower", factor: "2e0" } }], expect: "reject" },
  { name: "axis_upgrade_without_multiplier_source", mutations: [{ op: "delete", path: declaration1 }], expect: "reject" },
  { name: "axis_source_wrong_provider", mutations: [{ op: "set", path: [...declaration1, "provider"], value: "upgrade.pr_intern_1" }], expect: "reject" },
  { name: "axis_provider_in_wrong_slot", mutations: [{ op: "set", path: [...declaration1, "slot"], value: "upgrades" }], expect: "reject" },
  { name: "orphan_axis_source", mutations: [{ op: "append", path: ["multiplier_sources"], value: { id: "upgrade.pr_intern_9.axis", slot: "axis_stack", target: "all", provider: "axis_stack" } }], expect: "reject" },
  { name: "axis_stack_without_axis_effects", mutations: [
    { op: "delete", path: ["upgrades", { id: "upgrade.pr_intern_2" }] },
    { op: "delete", path: ["multiplier_sources", { id: "upgrade.pr_intern_2.axis" }] },
    { op: "delete", path: ["upgrades", intern1] },
    { op: "delete", path: declaration1 },
  ], expect: "reject" },
  { name: "axis_at_least_without_axis_stack", mutations: [
    { op: "delete", path: ["axis_stack"] },
    { op: "delete", path: ["upgrades", { id: "upgrade.pr_intern_2" }] },
    { op: "delete", path: ["multiplier_sources", { id: "upgrade.pr_intern_2.axis" }] },
    { op: "delete", path: declaration1 },
    { op: "set", path: [...effect], value: { source_id: "upgrade.pr_intern_1.static", slot: "upgrades", target: "generator.beige_tower", factor: "2e0" } },
  ], expect: "reject" },
  { name: "axis_at_least_minimum_zero", mutations: [{ op: "set", path: ["upgrades", intern1, "requires", 0, "minimum"], value: 0 }], expect: "reject" },
  { name: "axis_at_least_twice", mutations: [{ op: "append", path: ["upgrades", intern1, "requires"], value: { kind: "axis_at_least", minimum: 7 } }], expect: "reject" },
  { name: "axis_at_least_beside_route_condition", mutations: [{ op: "append", path: ["upgrades", intern1, "requires"], value: { kind: "resource_at_least", resource_id: "company.cash", value: "1e6" } }], expect: "accept" },
];
// Cross-artifact rows (CV1): the economy loads; the bundle check rejects.
const cross = [
  { name: "input_cap_below_reachable_attainment", mutations: [{ op: "set", path: ["axis_stack", "input_cap"], value: 43 }], achievements_pinned: true, maximum_attainment_run: 44, maximum_score_run: 70, expect: "reject" },
  { name: "input_cap_equal_reachable_attainment", mutations: [], achievements_pinned: true, maximum_attainment_run: 44, maximum_score_run: 70, expect: "accept" },
  { name: "score_run_cap_below_all_grants", mutations: [{ op: "set", path: ["axis_stack", "input"], value: "achievement_score_run" }], achievements_pinned: true, maximum_attainment_run: 44, maximum_score_run: 70, expect: "reject" },
  { name: "achievement_input_without_achievements", mutations: [], achievements_pinned: false, maximum_attainment_run: 0, maximum_score_run: 0, expect: "reject" },
];
// A route predicate never accepts the upgrade-only kind (routes loader).
const routes = read("balance/routes/phase0.json");
const gateIndex = routes.gates.findIndex((gate) => gate.routes.some((route) => route.predicate.length > 0));
if (gateIndex < 0) throw new Error("no route predicate to seed");
const routeCases = [
  { name: "route_predicate_rejects_axis_at_least", mutations: [{ op: "append", path: ["gates", gateIndex, "routes", 0, "predicate"], value: { kind: "axis_at_least", minimum: 1 } }], expect: "reject" },
  { name: "route_base_accepts", mutations: [], expect: "accept" },
];
write("testdata/axis-stack/loader-corpus-v1.json", { schema_version: 1, economy_fixture: "balance/testdata/axis-stack/economy-v5-fixture.json", routes_base: "balance/routes/phase0.json", economy_cases: cases, cross_cases: cross, route_cases: routeCases });
console.log(`axis-stack fixtures: ${cases.length} economy, ${cross.length} cross, ${routeCases.length} route cases`);
