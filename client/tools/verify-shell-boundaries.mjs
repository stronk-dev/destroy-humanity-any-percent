import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "svelte/compiler";
import { execFileSync } from "node:child_process";

const client = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const shell = path.join(client, "src", "shell");
const ui = path.join(client, "src", "ui");
const gameUI = path.join(client, "src", "game-ui");
const root = path.resolve(client, "..");
const cosmeticAuthority = JSON.parse(execFileSync("go", ["run", path.join(client, "tools/cosmetic-intent-kinds.go"), path.join(root, "server/production")], {
  cwd: root, encoding: "utf8", env: { ...process.env, GOCACHE: process.env.GOCACHE ?? path.join(root, ".cache/go-build") },
}));
if (!Array.isArray(cosmeticAuthority.kinds) || cosmeticAuthority.kinds.length === 0 ||
  cosmeticAuthority.kinds.some((kind, index, rows) => typeof kind !== "string" || kind.length === 0 || index > 0 && rows[index - 1] >= kind) ||
  cosmeticAuthority.negative_fixtures !== 10 || cosmeticAuthority.positive_fixtures !== 2) throw new Error("invalid cosmetic production-source observation");
const cosmeticKinds = new Set(cosmeticAuthority.kinds);
// Player surface components mounted by the Game UI obey the same boundary.
const surfaceDirectories = ["game-ui", "minigame", "soul"];
async function sourceFiles(directory, prefix = "") {
  const found = [];
  for (const entry of (await fs.readdir(directory, { withFileTypes: true })).sort((left, right) => left.name.localeCompare(right.name))) {
    const relative = path.join(prefix, entry.name);
    if (entry.isDirectory()) found.push(...await sourceFiles(path.join(directory, entry.name), relative));
    else if (entry.name.endsWith(".ts") || entry.name.endsWith(".svelte")) found.push(relative);
  }
  return found;
}

const files = await sourceFiles(shell);
for (const name of files) {
  const source = await fs.readFile(path.join(shell, name), "utf8");
  if (/balance[/\\](?:mutation|writer)|mutateBalance|writeBalance/.test(source)) throw new Error(`${name}: shell may not import balance mutation paths`);
  if (/addEventListener\s*\(\s*["']unload["']/.test(source)) throw new Error(`${name}: unload lifecycle handler is forbidden`);
}
const intents = await fs.readFile(path.join(shell, "intents.ts"), "utf8");
if (/from\s+["'][^"']*(?:prediction|display|controller)/.test(intents)) throw new Error("intent dispatcher may not import predicted state");

const governedStyle = /^(?:color|background(?:-color)?|border(?:-(?:top|right|bottom|left))?(?:-color)?|outline(?:-color)?|fill|stroke|font(?:-family)?|box-shadow|text-shadow|border-radius|animation(?:-duration)?|transition(?:-duration)?)$/;
const literalStyle = /(?:#[0-9a-f]{3,8}\b|\brgba?\(|\bhsla?\(|\b(?:repeating-)?(?:linear|radial|conic)-gradient\(|\b(?:0|[1-9]\d*)(?:\.\d+)?m?s\b|\b(?:Arial|Courier|Geneva|Helvetica|Tahoma|Verdana)\b|\b(?:inset\s+)?-?(?:0|[1-9]\d*)px\s+-?(?:0|[1-9]\d*)px)/i;
const forbiddenImports = /from\s+["'][^"']*(?:\/transport|\/replay|\/production|\/economy(?:-kernel)?|\/shell\/runtime|balance\/)[^"']*["']/;
const forbiddenNetwork = /\b(?:fetch\s*\(|new\s+WebSocket\s*\()/;
const playerFacingAttributes = new Set(["alt", "aria-description", "aria-label", "placeholder", "title", "value"]);

function visit(node, callback, seen = new Set()) {
  if (node === null || typeof node !== "object" || seen.has(node)) return;
  seen.add(node);
  callback(node);
  for (const value of Object.values(node)) {
    if (Array.isArray(value)) for (const child of value) visit(child, callback, seen);
    else visit(value, callback, seen);
  }
}

function assertGovernedValue(property, value, label) {
  if (!governedStyle.test(property)) return;
  if (!value.includes("var(--cc-") || literalStyle.test(value)) {
    throw new Error(`${label}: ${property} must use governed --cc-* tokens`);
  }
}

function propertyName(property) {
  if (property.type !== "Property" || property.computed) return undefined;
  return property.key.type === "Identifier" ? property.key.name : property.key.type === "Literal" ? property.key.value : undefined;
}

// GS6-A2: production owns kinds; syntax that cannot prove its envelope fails.
// This is a source gate, not proof of server eligibility or real receipt flow.
function verifyCosmeticCommands(ast, label) {
  const host = label === "game-ui/GameUIApp.svelte";
  const functions = new Map();
  visit(ast, (node) => {
    if (node.type === "FunctionDeclaration" && ["act", "withPlan", "gardenAct", "actTransition"].includes(node.id?.name)) {
      if (functions.has(node.id.name)) throw new Error(`${label}: duplicate command wrapper`);
      functions.set(node.id.name, node);
    }
  });
  if (host && !functions.has("act")) throw new Error(`${label}: host intent dispatcher missing`);
  const isBody = (node) => node?.type === "Identifier" && node.name === "body";
  const withPlan = functions.get("withPlan");
  if (withPlan) {
    const returned = withPlan.body.body.length === 1 && withPlan.body.body[0].type === "ReturnStatement" ? withPlan.body.body[0].argument : undefined;
    const properties = returned?.alternate?.type === "ObjectExpression" ? returned.alternate.properties : [];
    const test = returned?.test;
    if (label !== "game-ui/GameUIApp.svelte" || !isBody(withPlan.params[0]) || returned?.type !== "ConditionalExpression" ||
      test?.type !== "BinaryExpression" || test.operator !== "===" || test.right.type !== "Literal" || test.right.value !== 0 ||
      test.left.type !== "MemberExpression" || test.left.computed || test.left.object.type !== "Identifier" ||
      test.left.object.name !== "exitPlan" || test.left.property.name !== "length" || !isBody(returned.consequent) || properties.length !== 2 ||
      properties[0].type !== "SpreadElement" || !isBody(properties[0].argument) || propertyName(properties[1]) !== "reputation_plan" ||
      properties[1].value.type !== "ArrayExpression" || properties[1].value.elements.length !== 1 ||
      properties[1].value.elements[0].type !== "SpreadElement" || properties[1].value.elements[0].argument.type !== "Identifier" ||
      properties[1].value.elements[0].argument.name !== "exitPlan") throw new Error(`${label}: unsupported Exit-plan command transformation`);
  }
  const forwarders = ["gardenAct", "actTransition"].map((name) => functions.get(name)).filter(Boolean);
  for (const forwarder of forwarders) {
    let bodyReferences = 0, forwardingCalls = 0;
    const memberNames = new Set();
    visit(forwarder.body, (node) => {
      if (node.type === "MemberExpression" && !node.computed) memberNames.add(node.property);
    });
    visit(forwarder.body, (node) => {
      if (isBody(node) && !memberNames.has(node)) bodyReferences++;
      if (node.type === "CallExpression" && node.callee.type === "Identifier" && node.callee.name === "act" && isBody(node.arguments[0])) forwardingCalls++;
    });
    if (label !== "game-ui/GameUIApp.svelte" || !isBody(forwarder.params[0]) || bodyReferences !== 1 || forwardingCalls !== 1) {
      throw new Error(`${label}: unsupported ${forwarder.id.name} command forwarding (body references=${bodyReferences}, act calls=${forwardingCalls})`);
    }
  }
  const seen = new Set();
  let hostRuntimeForwards = 0;
  const walk = (node, ancestors = []) => {
    if (node === null || typeof node !== "object" || seen.has(node)) return;
    seen.add(node);
    if (node.type === "CallExpression") {
      if (host && node.callee.type === "Identifier" && ["act", "withPlan", "gardenAct", "actTransition"].includes(node.callee.name)) {
        const argument = node.arguments[0];
        const verifiedForward = node.callee.name === "act" && forwarders.some((forwarder) => ancestors.includes(forwarder)) && isBody(argument);
        const planWrapper = argument?.type === "CallExpression" && argument.callee.type === "Identifier" && argument.callee.name === "withPlan" && withPlan;
        const envelope = planWrapper ? argument.arguments[0] : argument;
        if (!verifiedForward && (envelope?.type !== "ObjectExpression" || envelope.properties.some((property) => property.type !== "Property" || property.computed))) {
          throw new Error(`${label}: opaque intent envelope cannot prove cosmetic source authority`);
        }
        const kinds = verifiedForward ? [] : envelope.properties.filter((property) => propertyName(property) === "kind");
        if (!verifiedForward && (kinds.length !== 1 || kinds[0].value.type !== "Literal" || typeof kinds[0].value.value !== "string")) {
          throw new Error(`${label}: dynamic intent kind cannot prove cosmetic source authority`);
        }
      }
      if (node.callee.type === "MemberExpression" && !node.callee.computed && node.callee.property.name === "intent") {
        const envelope = node.arguments[0];
        const owner = ancestors.findLast((ancestor) => ancestor.type === "FunctionDeclaration" && ancestor.id?.name === "act");
        const properties = envelope?.type === "ObjectExpression" ? envelope.properties : [];
        const forwarding = label === "game-ui/GameUIApp.svelte" && owner?.params[0]?.type === "Identifier" && owner.params[0].name === "body" &&
          properties.length === 3 && propertyName(properties[0]) === "intent_id" && propertyName(properties[1]) === "expected_revision" &&
          properties[2].type === "SpreadElement" && properties[2].argument.type === "Identifier" && properties[2].argument.name === "body";
        if (forwarding) hostRuntimeForwards++;
        if (!forwarding && (envelope?.type !== "ObjectExpression" || envelope.properties.some((property) => property.type !== "Property" || property.computed))) {
          throw new Error(`${label}: opaque runtime intent bypasses verified source envelopes`);
        }
      }
    }
    if (node.type === "ObjectExpression") {
      const names = node.properties.map(propertyName);
      const kindProperties = node.properties.filter((property) => propertyName(property) === "kind");
      const kind = kindProperties.length === 1 && kindProperties[0].value.type === "Literal" ? kindProperties[0].value.value : undefined;
      if (names.includes("cosmetic_id") || cosmeticKinds.has(kind)) {
        if (node.properties.some((property) => property.type !== "Property" || property.computed) || kindProperties.length !== 1 ||
          new Set(names).size !== names.length || typeof kind !== "string") throw new Error(`${label}: ambiguous cosmetic command envelope`);
        if (!cosmeticKinds.has(kind)) throw new Error(`${label}: cosmetic command ${kind} is not in production source authority`);
      }
    }
    for (const value of Object.values(node)) {
      if (Array.isArray(value)) for (const child of value) walk(child, [...ancestors, node]);
      else walk(value, [...ancestors, node]);
    }
  };
  walk(ast);
  if (host && hostRuntimeForwards !== 1) throw new Error(`${label}: host must have one verified runtime forwarding path`);
}

function verifySvelteSource(source, label) {
  const ast = parse(source, { modern: true });
  verifyCosmeticCommands(ast, label);
  const attributeText = new Set();
  visit(ast.fragment, (node) => {
    if (node.type === "Attribute" && Array.isArray(node.value)) for (const part of node.value) attributeText.add(part);
  });
  visit(ast.css, (node) => {
    if (node.type === "Declaration") assertGovernedValue(node.property, node.value, label);
  });
  visit(ast.fragment, (node) => {
    if (node.type === "Text" && !attributeText.has(node) && node.data.trim().length > 0) throw new Error(`${label}: player-facing text must resolve through the Copy pipeline`);
    if (node.type === "Attribute" && playerFacingAttributes.has(node.name) && Array.isArray(node.value) &&
      node.value.some((part) => part.type === "Text" && part.data.trim().length > 0)) {
      throw new Error(`${label}: player-facing attribute ${node.name} must resolve through the Copy pipeline`);
    }
    if (node.type === "Attribute" && node.name === "style") {
      if (!Array.isArray(node.value) || node.value.some((part) => part.type !== "Text")) throw new Error(`${label}: dynamic inline style attributes are forbidden`);
      const inline = node.value.map((part) => part.data).join("");
      const wrapper = parse(`<style>.fixture{${inline}}</style>`, { modern: true });
      visit(wrapper.css, (declaration) => {
        if (declaration.type === "Declaration") assertGovernedValue(declaration.property, declaration.value, label);
      });
    }
    if (node.type === "StyleDirective" && governedStyle.test(node.name)) throw new Error(`${label}: governed style directives are forbidden`);
  });
}

const uiFiles = await sourceFiles(ui);
for (const name of uiFiles) {
  const source = await fs.readFile(path.join(ui, name), "utf8");
  if (forbiddenImports.test(source)) throw new Error(`${name}: UI boundary may not import authoritative or transport internals`);
  if (forbiddenNetwork.test(source)) throw new Error(`${name}: UI boundary may not open raw fetch/WebSocket connections`);
  if (name.endsWith(".svelte")) verifySvelteSource(source, name);
}

const gameUIFiles = [];
for (const directory of surfaceDirectories) {
  for (const name of (await sourceFiles(path.join(client, "src", directory))).filter((file) => file.endsWith(".svelte"))) gameUIFiles.push(path.join(directory, name));
}
for (const name of gameUIFiles) {
  const source = await fs.readFile(path.join(client, "src", name), "utf8");
  if (forbiddenImports.test(source)) throw new Error(`${name}: Game UI component may not import authoritative or transport internals`);
  if (forbiddenNetwork.test(source)) throw new Error(`${name}: Game UI component may not open raw fetch/WebSocket connections`);
  verifySvelteSource(source, name);
}

verifySvelteSource(`<p style="width: 3px"><span class="ok">{value}</span></p><style>.ok{color:var(--cc-color-text);width:3px}</style>`, "seeded pass");
for (const seeded of [
  `<style>.bad{color:#fff}</style>`,
  `<style>.bad{background:linear-gradient(var(--cc-color-bg),var(--cc-color-surface))}</style>`,
  `<p style="border-radius: 4px">{value}</p>`,
  `<p style:color={value}>{value}</p>`,
  `<p>Unregistered player copy</p>`,
  `<button aria-label="Unregistered label">{value}</button>`,
  `<input placeholder="Unregistered placeholder" />`,
]) {
  let rejected = false;
  try { verifySvelteSource(seeded, "seeded violation"); } catch { rejected = true; }
  if (!rejected) throw new Error("UI literal-style lint did not reject its seeded violation");
}

for (const seeded of ["fetch('/api')", "new WebSocket('wss://example.invalid')"]) {
  if (!forbiddenNetwork.test(seeded)) throw new Error("UI raw-network lint did not reject its seeded violation");
}

const cosmeticSeed = (body) => parse(`<script lang="ts">${body}</script>`, { modern: true });
for (const kind of cosmeticKinds) verifyCosmeticCommands(cosmeticSeed(`const body = {kind:${JSON.stringify(kind)},cosmetic_id:"horse_armor"};`), "registered cosmetic fixture");
verifyCosmeticCommands(cosmeticSeed(`const text = '({kind:"buy_horse_armor",cosmetic_id:"horse_armor"})'; // act({kind:"bad",cosmetic_id:"x"})`), "non-code cosmetic lookalikes");
const forwardingFixture = 'function act(body) { runtime.intent({intent_id:identity(),expected_revision:1,...body}); }';
const planFixture = 'function withPlan(body) { return exitPlan.length === 0 ? body : {...body,reputation_plan:[...exitPlan]}; }';
const gardenFixture = 'function gardenAct(body) { void act(body,{scope:"founder"}).then(() => { refresh++; }); }';
const transitionFixture = 'async function actTransition(body, origin) { await act(body); if (document.activeElement === document.body) origin.focus(); }';
verifyCosmeticCommands(cosmeticSeed(forwardingFixture + 'act({kind:"buy_generator",generator_id:"generator.beige_tower"});'), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed(forwardingFixture), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed(forwardingFixture + planFixture + 'act(withPlan({kind:"wind_down"}));'), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed(forwardingFixture + gardenFixture + 'gardenAct({kind:"garden_plant"});'), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed(forwardingFixture + transitionFixture + 'actTransition({kind:"cross_gate"}, origin);'), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed(forwardingFixture + transitionFixture + planFixture + 'actTransition(withPlan({kind:"wind_down"}), origin);'), "game-ui/GameUIApp.svelte");
verifyCosmeticCommands(cosmeticSeed('function act(run) { run(); } act(() => onPlant());'), "garden local callback, not host dispatcher");
const rejectedCosmetic = [
  'act({kind:"buy_horse_armor",cosmetic_id:"horse_armor"});',
  'const body = {kind:"buy_horse_armor",cosmetic_id:"horse_armor"};',
  'act({kind:dynamic,cosmetic_id:"horse_armor"});',
  'act({kind:"acquire_cosmetic",...body,cosmetic_id:"horse_armor"});',
  'act(body);',
  'runtime.intent(body);',
  'runtime.intent({...body});',
  'act({["kind"]:"acquire_cosmetic",cosmetic_id:"horse_armor"});',
  'act({kind:"acquire_cosmetic",kind:"buy_horse_armor",cosmetic_id:"horse_armor"});',
  planFixture.replace('reputation_plan:[...exitPlan]', 'kind:"buy_horse_armor"') + 'act(withPlan({kind:"wind_down"}));',
  gardenFixture.replace('void act(body', 'body.kind = "buy_horse_armor"; void act(body') + 'gardenAct({kind:"garden_plant"});',
  transitionFixture.replace('await act(body)', 'body.kind = "buy_horse_armor"; await act(body)') + 'actTransition({kind:"cross_gate"}, origin);',
  transitionFixture + 'actTransition(body, origin);',
  transitionFixture + 'actTransition({kind:dynamic}, origin);',
  transitionFixture + 'actTransition({kind:"buy_horse_armor",cosmetic_id:"horse_armor"}, origin);',
];
for (const fixture of rejectedCosmetic) {
  let rejected = false;
  try { verifyCosmeticCommands(cosmeticSeed(forwardingFixture + fixture), "game-ui/GameUIApp.svelte"); } catch { rejected = true; }
  if (!rejected) throw new Error(`cosmetic source guard did not reject ${fixture}`);
}
console.log(`shell/UI boundaries ok: ${files.length} shell, ${uiFiles.length} UI, and ${gameUIFiles.length} Game UI component files; ` +
  `cosmetic authority ${JSON.stringify([...cosmeticKinds])}; ${cosmeticAuthority.negative_fixtures} Go and ${rejectedCosmetic.length} Svelte cosmetic source negatives rejected`);
