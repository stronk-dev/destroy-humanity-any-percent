import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const read = (file) => fs.readFileSync(path.join(root, file), "utf8");
const json = (file) => JSON.parse(read(file));
const clean = (value) => String(value ?? "-").replaceAll("\t", " ").replaceAll("\n", " ");
const lineOf = (file, needle) => {
  const lines = read(file).split("\n");
  const index = lines.findIndex((line) => line.includes(needle));
  if (index < 0) throw new Error(`missing ${needle} in ${file}`);
  return `${file}:${index + 1}`;
};
const join = (values) => values.length ? [...new Set(values)].sort().join(";") : "-";

const sourceFiles = fs.readdirSync(path.join(root, "copy/catalog")).filter((name) => name.endsWith(".json")).sort();
const sourceByKey = new Map();
for (const name of sourceFiles) {
  const file = `copy/catalog/${name}`;
  for (const entry of json(file).entries) {
    if (sourceByKey.has(entry.key)) throw new Error(`duplicate source key ${entry.key}`);
    sourceByKey.set(entry.key, { file, label: name.replace(/\.json$/, ""), entry });
  }
}
const generated = json("client/src/copy/generated/catalog.json").entries;
if (generated.length !== 208 || sourceByKey.size !== 208) throw new Error(`catalog denominator mismatch generated=${generated.length} source=${sourceByKey.size}`);
for (const entry of generated) if (!sourceByKey.has(entry.key)) throw new Error(`generated-only key ${entry.key}`);
const orphan = new Set(json("copy/generated/orphans.v1.json").keys);
if (orphan.size !== 161) throw new Error(`orphan denominator ${orphan.size}`);

const clientInventory = read("planning/platform-alignment/client-source-inventory.tsv").trimEnd().split("\n").slice(1).map((line) => line.split("\t"));
const reachableClientFiles = clientInventory.filter((row) => row[2].includes("main_") || row[2] === "worker_runtime_graph" || row[2] === "worker_url_graph").map((row) => `client/src/${row[0]}`).filter((file) => fs.existsSync(path.join(root, file)) && /\.(?:ts|svelte)$/.test(file));
const testToolFiles = [
  "client/src/ui/FixtureHost.svelte",
  ...walk("client/test"), ...walk("server"), ...walk("tools"), ...walk("balance/testdata")
].filter((file) => /(?:_test\.go|\.test\.ts|\.spec\.ts|FixtureHost\.svelte|balance\/testdata|tools\/)/.test(file));

function walk(relative) {
  const absolute = path.join(root, relative);
  if (!fs.existsSync(absolute)) return [];
  const result = [];
  for (const item of fs.readdirSync(absolute, { withFileTypes: true })) {
    const child = `${relative}/${item.name}`;
    if (item.isDirectory()) result.push(...walk(child));
    else if (/\.(?:go|ts|svelte|json|mjs|js)$/.test(item.name)) result.push(child);
  }
  return result;
}
function sites(files, key, predicate = () => true) {
  const result = [];
  for (const file of files) {
    const lines = read(file).split("\n");
    for (let index = 0; index < lines.length; index++) if (lines[index].includes(`"${key}"`) && predicate(lines[index])) result.push(`${file}:${index + 1}`);
  }
  return result;
}

const epoch = json("balance/epochs/phase0.json");
if (epoch.artifacts.length !== 19) throw new Error(`epoch artifact denominator ${epoch.artifacts.length}`);
const artifacts = new Map(epoch.artifacts.map((row) => [row.name, row.path]));
const artifactSites = new Map(generated.map(({ key }) => [key, []]));
for (const { key } of generated) for (const [name, file] of artifacts) {
  const found = sites([file], key);
  if (found.length) artifactSites.get(key).push(...found.map((site) => `${name}@${site}`));
}

// The checked-in registry's seven closed JSON paths. Upgrade rows are prefixes and expand to title/description.
const registered = new Map(generated.map(({ key }) => [key, []]));
const addRegistered = (key, artifact, needle = `"${key}"`) => {
  if (!registered.has(key)) throw new Error(`registered key absent from catalog: ${key}`);
  registered.get(key).push(`${artifact}@${lineOf(artifacts.get(artifact), needle)}`);
};
const achievements = json(artifacts.get("achievements")).achievements;
for (const row of achievements) {
  addRegistered(row.copy_key, "achievements");
  if (row.proof.justification_copy_key) addRegistered(row.proof.justification_copy_key, "achievements");
}
for (const row of json(artifacts.get("categories")).categories) addRegistered(row.name_key, "categories");
const economy = json(artifacts.get("economy"));
for (const row of economy.resources) if (row.hardcap) addRegistered(row.hardcap.reason_key, "economy");
for (const row of economy.generator_classes) if (row.provisioned_hardcap) addRegistered(row.provisioned_hardcap.reason_key, "economy");
for (const row of economy.upgrades) for (const suffix of ["title", "description"]) addRegistered(`${row.copy_key}.${suffix}`, "economy", `"copy_key": "${row.copy_key}"`);
for (const row of json(artifacts.get("factions")).factions) addRegistered(row.incorporation_copy_key, "factions");

const codeRegistry = json("copy/code-reference-sites.v1.json").references;
const goRefs = new Map(generated.map(({ key }) => [key, []]));
for (const ref of codeRegistry) {
  const site = lineOf(ref.source_file, `copykeys.${ref.key === "compact.recruitment.mid_t3" ? "CompactRecruitmentMidT3" : "RouteRegistryFirstReordered"}`);
  goRefs.get(ref.key).push(`${ref.go_function}@${site}`);
}

const literal = new Map(generated.map(({ key }) => [key, sites(reachableClientFiles, key, (line) => /\b(?:t|resolveCopy)\s*\(/.test(line))]));
const dynamic = new Map(generated.map(({ key }) => [key, []]));
const addDynamic = (key, value) => {
  if (!dynamic.has(key)) throw new Error(`dynamic key absent: ${key}`);
  dynamic.get(key).push(value);
};

const presentation = json("client/src/game-ui/presentation.generated.json");
for (const row of presentation.generators) {
  addDynamic(row.title_key, `Desk generator binding@client/src/game-ui/GameUIApp.svelte:273`);
  addDynamic(row.description_key, `Desk generator binding@client/src/game-ui/GameUIApp.svelte:273`);
}
for (const row of presentation.upgrades) {
  addDynamic(row.title_key, `Desk upgrade binding@client/src/game-ui/GameUIApp.svelte:287`);
  addDynamic(row.description_key, `Desk upgrade binding@client/src/game-ui/GameUIApp.svelte:287`);
}
for (const row of presentation.manual_actions) {
  addDynamic(row.title_key, `Desk manual binding@client/src/game-ui/GameUIApp.svelte:251`);
  addDynamic(row.description_key, `Desk manual binding@client/src/game-ui/GameUIApp.svelte:252`);
}
for (const row of presentation.gates) addDynamic(row.title_key, `Desk split binding@client/src/game-ui/GameUIApp.svelte:293`);
for (const row of presentation.exit_types) addDynamic(row.title_key, `Offer/Run End exit binding@client/src/game-ui/GameUIApp.svelte:186`);
for (const row of presentation.clout_reach_notes) addDynamic(row.text_key, `Offer/Run End terms binding@client/src/game-ui/prestige-terms.ts:13`);

const curriculum = json(artifacts.get("curriculum"));
for (const branch of curriculum.first_failure.branches) {
  addDynamic(branch.title_key, `Run End bounded branch map@${lineOf("client/src/game-ui/RunEndSurface.svelte", branch.title_key)}`);
  addDynamic(branch.body_key, `Run End bounded branch map@${lineOf("client/src/game-ui/RunEndSurface.svelte", branch.body_key)}`);
}
for (const key of ["curriculum.scripted_first_failure.title", "curriculum.scripted_first_failure.body"]) addDynamic(key, `Run End default branch@client/src/game-ui/RunEndSurface.svelte:35`);

const categoryKeys = ["category.any_percent", "category.ethical_percent", "category.hundred_percent", "category.low_percent", "category.valuation"];
for (const key of categoryKeys) addDynamic(key, `bounded category map@${lineOf("client/src/game-ui/GameUIApp.svelte", key)}`);
for (const row of economy.resources) if (row.scope === "company" && row.hardcap) addDynamic(row.hardcap.reason_key, `Desk Amount resource-cap DTO@client/src/game-ui/GameUIApp.svelte:188;client/src/ui/Amount.svelte:31`);

// Every nonliteral mounted resolver call is explicitly understood. This is the fail-loud
// unbounded-dynamic control: a new computed grammar stops the audit instead of silently widening.
const computedCallSites = [];
for (const file of reachableClientFiles.filter((file) => file !== "client/src/copy/index.ts")) {
  const lines = read(file).split("\n");
  for (let index = 0; index < lines.length; index++) {
    const line = lines[index];
    if (/\bt\(\s*(?!["'])/.test(line) || /\bresolveCopy\(/.test(line)) computedCallSites.push(`${file}:${index + 1}`);
  }
}
const expectedComputedCallSites = [
  "client/src/game-ui/GameUIApp.svelte:183", "client/src/game-ui/GameUIApp.svelte:186",
  "client/src/game-ui/GameUIApp.svelte:251", "client/src/game-ui/GameUIApp.svelte:252",
  "client/src/game-ui/GameUIApp.svelte:255", "client/src/game-ui/GameUIApp.svelte:273", "client/src/game-ui/GameUIApp.svelte:287",
  "client/src/game-ui/GameUIApp.svelte:293", "client/src/game-ui/RunEndSurface.svelte:18",
  "client/src/game-ui/RunEndSurface.svelte:47", "client/src/game-ui/RunEndSurface.svelte:51",
  "client/src/game-ui/prestige-terms.ts:13", "client/src/game-ui/prestige-terms.ts:16",
  "client/src/ui/Amount.svelte:31"
];
if (join(computedCallSites) !== join(expectedComputedCallSites)) throw new Error(`unbounded/changed computed resolver sites: ${join(computedCallSites)}`);

const otherRefs = new Map(generated.map(({ key }) => [key, []]));
for (const { key } of generated) {
  for (const site of artifactSites.get(key)) if (!registered.get(key).some((registeredSite) => registeredSite.split("@")[0] === site.split("@")[0])) otherRefs.get(key).push(`unregistered-current:${site}`);
  const presentationSites = sites(["client/src/game-ui/presentation.generated.json"], key);
  if (presentationSites.length) otherRefs.get(key).push(...presentationSites.map((site) => `presentation:${site}`));
}
const testRefs = new Map(generated.map(({ key }) => [key, sites(testToolFiles, key)]));

function workflowFor(key, lit, dyn, deploy, go, other, test) {
  if (lit.some((site) => site.includes("RunEndSurface")) || dyn.some((site) => site.includes("Run End"))) return "Run End";
  if (lit.some((site) => site.includes("prestige-terms"))) return "Offer/Run End terms";
  if (lit.some((site) => site.includes("GameUIApp")) || dyn.some((site) => site.includes("Desk") || site.includes("Offer"))) {
    if (key.startsWith("screen.vision") || key.startsWith("satire.")) return "Vision/Desk satire";
    if (key.startsWith("screen.offer")) return "Offer Sheet";
    if (key.startsWith("settings.") || key === "surface.settings.title") return "Settings";
    return "Game UI/Desk";
  }
  if (dyn.some((site) => site.includes("category"))) return "category title binding";
  if (go.length) return "backend event/receipt only";
  if (deploy.length || other.some((site) => site.startsWith("unregistered-current:"))) return `deploy-current ${[...deploy, ...other].find((site) => site.includes("@"))?.split("@")[0].replace("unregistered-current:", "") ?? "artifact"}`;
  if (test.length) return "test/fixture/tool";
  return "none discovered";
}

function classify(key, lit, dyn, deploy, go, other, test) {
  if (dyn.some((site) => site === "unbounded-runtime-key")) return "ambiguous_dynamic";
  const mountedLiteral = lit.length > 0 && key !== "terms.network_slot_unlock.frame";
  const categoryBlocked = categoryKeys.includes(key) && key !== "category.any_percent";
  const mountedDynamic = dyn.length > 0 && !categoryBlocked;
  if (mountedLiteral || mountedDynamic) return "mounted_player_copy";
  if (deploy.length || go.length || other.some((site) => site.startsWith("unregistered-current:"))) return "shipped_backend_or_data_only";
  if (lit.length || dyn.length || other.length) return "shipped_unmounted_surface_copy";
  if (test.length) return "fixture_or_tool_only";
  return "unreferenced_candidate";
}

function evidenceLimit(key, verdict, lit, dyn, deploy, go, other) {
  if (key === "terms.network_slot_unlock.frame") return "reachable literal is guarded by an empty presentation network-slot map and current terms produce no slots";
  if (categoryKeys.includes(key) && key !== "category.any_percent") return "bounded client map exists, but current Game UI projector hardcodes any_percent; catalog/fixture presence is not a live selector";
  if (verdict === "mounted_player_copy") return "static bounded path at 190a4fa; not a browser interaction-frequency claim";
  if (verdict === "shipped_backend_or_data_only") return "current data/producer presence does not prove a mounted reader or visible rendering";
  if (verdict === "shipped_unmounted_surface_copy") return "binding/source ships, but no current executable producer-to-mounted-consumer path was found";
  if (verdict === "fixture_or_tool_only") return "test/tool selection is not production reachability";
  if (verdict === "ambiguous_dynamic") return "computed selection grammar is unbounded; neither live nor unused can be claimed";
  return "no bounded exact reference found; deletion still requires successor diff review and dynamic-grammar guard";
}

const header = ["key", "source_catalog", "source_label", "params", "tone", "current_orphan_report", "mounted_literal_sites", "mounted_dynamic_sites", "deploy_artifact_refs", "go_producer_refs", "other_catalog_refs", "test_tool_refs", "workflow", "verdict", "evidence_limit"];
const rows = [];
const counts = new Map();
for (const entry of generated) {
  const key = entry.key, source = sourceByKey.get(key);
  const lit = literal.get(key), dyn = dynamic.get(key), deploy = registered.get(key), go = goRefs.get(key), other = otherRefs.get(key), test = testRefs.get(key);
  const verdict = classify(key, lit, dyn, deploy, go, other, test);
  counts.set(verdict, (counts.get(verdict) ?? 0) + 1);
  rows.push([key, source.file, source.label, entry.params.length ? entry.params.map((row) => `${row.name}:${row.type}`).join(",") : "-", entry.tone, orphan.has(key) ? "orphan" : "referenced", join(lit), join(dyn), join(deploy), join(go), join(other), join(test), workflowFor(key, lit, dyn, deploy, go, other, test), verdict, evidenceLimit(key, verdict, lit, dyn, deploy, go, other)]);
}
function validateRows(values) {
  if (values.length !== 208 || new Set(values.map((row) => row[0])).size !== 208) throw new Error("output denominator/uniqueness failure");
  for (const required of ["desk.buy_one", "chrome.run_title.company_fallback", "screen.run_end.founder_note"]) if (values.find((row) => row[0] === required)?.[13] !== "mounted_player_copy") throw new Error(`live positive control failed: ${required}`);
  for (const required of ["achievement.first_gate", "pitch.card.api_call", "soul.recovery.defrag.title"]) if (values.find((row) => row[0] === required)?.[13] !== "shipped_backend_or_data_only") throw new Error(`backend-only positive control failed: ${required}`);
}
validateRows(rows);
if (classify("audit.fake.absent", [], [], [], [], [], []) !== "unreferenced_candidate") throw new Error("in-memory absent negative control failed");
if (classify("audit.dynamic.unbounded", [], ["unbounded-runtime-key"], [], [], [], []) !== "ambiguous_dynamic") throw new Error("unbounded dynamic control failed");
const seededFailures = [];
function mustReject(label, operation) {
  try { operation(); } catch { seededFailures.push(label); return; }
  throw new Error(`seeded failure was accepted: ${label}`);
}
mustReject("dropped-row", () => validateRows(rows.slice(1)));
mustReject("live-as-unused", () => validateRows(rows.map((row) => row[0] === "desk.buy_one" ? [...row.slice(0, 13), "unreferenced_candidate", row[14]] : row)));
mustReject("backend-as-mounted", () => validateRows(rows.map((row) => row[0] === "achievement.first_gate" ? [...row.slice(0, 13), "mounted_player_copy", row[14]] : row)));

const missingCurrent = [];
function collectCopyFields(value, pointer, artifact, file) {
  if (Array.isArray(value)) return value.forEach((row, index) => collectCopyFields(row, `${pointer}/${index}`, artifact, file));
  if (value === null || typeof value !== "object") return;
  for (const [field, child] of Object.entries(value)) {
    const childPointer = `${pointer}/${field}`;
    const copyLike = typeof child === "string" && /(?:copy_key|name_key|reason_key|title_key|body_key|next_run_key|curtain_copy_key)$/.test(field);
    const prefix = artifact === "economy" && /^\/upgrades\/\d+\/copy_key$/.test(childPointer);
    if (copyLike && !prefix && !sourceByKey.has(child)) missingCurrent.push({ artifact, file, pointer: childPointer, key: child });
    collectCopyFields(child, childPointer, artifact, file);
  }
}
for (const [artifact, file] of artifacts) collectCopyFields(json(file), "", artifact, file);
const absentCurrentCopyKeys = [...new Set(missingCurrent.map((row) => row.key))].sort();
if (absentCurrentCopyKeys.join("\0") !== ["cap.active_combo", "cap.cash", "cap.fiscal_credit", "cap.fiscal_level.beige_tower"].join("\0")) throw new Error(`unexpected current-artifact/catalog mismatch: ${absentCurrentCopyKeys.join(",")}`);
process.stderr.write(JSON.stringify({ rows: rows.length, orphan: orphan.size, verdicts: Object.fromEntries([...counts].sort()), mountedMarkedOrphan: rows.filter((row) => row[13] === "mounted_player_copy" && row[5] === "orphan").length, referencedButNoMountedConsumer: rows.filter((row) => row[5] === "referenced" && row[13] !== "mounted_player_copy").length, computedCallSites: computedCallSites.length, absentCurrentCopyKeys, seededFailures }) + "\n");
process.stdout.write([header, ...rows].map((row) => row.map(clean).join("\t")).join("\n") + "\n");
