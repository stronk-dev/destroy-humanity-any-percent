import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "svelte/compiler";

const client = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const shell = path.join(client, "src", "shell");
const ui = path.join(client, "src", "ui");
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

function verifySvelteSource(source, label) {
  const ast = parse(source, { modern: true });
  const attributeText = new Set();
  visit(ast.fragment, (node) => {
    if (node.type === "Attribute" && Array.isArray(node.value)) for (const part of node.value) attributeText.add(part);
  });
  visit(ast.css, (node) => {
    if (node.type === "Declaration") assertGovernedValue(node.property, node.value, label);
  });
  visit(ast.fragment, (node) => {
    if (node.type === "Text" && !attributeText.has(node) && node.data.trim().length > 0) throw new Error(`${label}: player-facing text must resolve through the Copy pipeline`);
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

verifySvelteSource(`<p style="width: 3px"><span class="ok">{value}</span></p><style>.ok{color:var(--cc-color-text);width:3px}</style>`, "seeded pass");
for (const seeded of [
  `<style>.bad{color:#fff}</style>`,
  `<style>.bad{background:linear-gradient(var(--cc-color-bg),var(--cc-color-surface))}</style>`,
  `<p style="border-radius: 4px">{value}</p>`,
  `<p style:color={value}>{value}</p>`,
  `<p>Unregistered player copy</p>`,
]) {
  let rejected = false;
  try { verifySvelteSource(seeded, "seeded violation"); } catch { rejected = true; }
  if (!rejected) throw new Error("UI literal-style lint did not reject its seeded violation");
}

for (const seeded of ["fetch('/api')", "new WebSocket('wss://example.invalid')"]) {
  if (!forbiddenNetwork.test(seeded)) throw new Error("UI raw-network lint did not reject its seeded violation");
}

console.log(`shell/UI boundaries ok: ${files.length} shell and ${uiFiles.length} UI source files`);
