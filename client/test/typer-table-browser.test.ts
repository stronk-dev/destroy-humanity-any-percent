import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import TyperTable from "../src/game-ui/minigame/TyperTable.svelte";
import type { TyperCommand, TyperSnapshot } from "../src/typer/engine";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const HASH = `sha256:${"a".repeat(64)}`;

function snapshot(overrides: Partial<TyperSnapshot> = {}): TyperSnapshot {
  return { assist_level: null, clean_lines: 0, current_prompt_id: null, current_prompt_misses: 0, current_prompt_text: null, deadline_server_ms: null,
    era_tier: 1, last_server_ms: null, last_submission: null, lines_cleared: 0, misses: 0, phase: "ready", prompt_index: 0, prompts_total: 8,
    revision: 1, started_server_ms: null, typer_content_hash: HASH, typer_schema_version: 1, ...overrides };
}
const typing = (overrides: Partial<TyperSnapshot> = {}) => snapshot({ phase: "typing", assist_level: "untimed", started_server_ms: 1_000, last_server_ms: 1_000,
  current_prompt_id: "ls_la", current_prompt_text: "ls -la", revision: 2, ...overrides });

async function settle(): Promise<void> { for (let index = 0; index < 4; index++) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function render(value: TyperSnapshot, commands: TyperCommand[], width?: string) {
  const target = document.createElement("main");
  if (width) target.style.width = width;
  document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  const app = mount(TyperTable, { target, props: { snapshot: value, serverTimeSample: value.last_server_ms ?? 1_000, pending: false, era: "era_2000",
    dispatch: (command: TyperCommand) => commands.push(command), exitToHost: () => {} } });
  return { target, dispose: () => { unmount(app); target.remove(); } };
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}: ${target.textContent}`);
  return found;
}

it.skipIf(!browser)("offers timed and untimed without a default and begins by keyboard", async () => {
  const commands: TyperCommand[] = [];
  const { target, dispose } = render(snapshot(), commands);
  try {
    await settle();
    await assertAxe(target, "ready");
    expect(target.querySelector("[aria-pressed=true], [checked]")).toBeNull();
    const untimed = button(target, "Untimed run");
    expect(untimed.getAttribute("aria-describedby")).toBe("typer-untimed-note");
    untimed.focus();
    untimed.click();
    expect(commands).toEqual([{ kind: "begin", assist_level: "untimed" }]);
  } finally { dispose(); }
});

it.skipIf(!browser)("submits the typed line on Enter, never during composition, and never blocks paste", async () => {
  const commands: TyperCommand[] = [];
  const { target, dispose } = render(typing(), commands);
  try {
    await settle();
    await assertAxe(target, "typing");
    const input = target.querySelector<HTMLInputElement>("#typer-line")!;
    expect(document.activeElement).toBe(input);
    for (const [attribute, value] of [["autocomplete", "off"], ["autocapitalize", "off"], ["spellcheck", "false"]] as const) expect(input.getAttribute(attribute)).toBe(value);
    const paste = new Event("paste", { bubbles: true, cancelable: true });
    input.dispatchEvent(paste);
    expect(paste.defaultPrevented).toBe(false);
    input.dispatchEvent(new CompositionEvent("compositionstart", { bubbles: true }));
    const composingEnter = new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true });
    input.dispatchEvent(composingEnter);
    expect(composingEnter.defaultPrevented).toBe(true);
    target.querySelector("form")!.requestSubmit();
    expect(commands).toEqual([]);
    input.dispatchEvent(new CompositionEvent("compositionend", { bubbles: true }));
    input.value = "ls -la";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await settle();
    target.querySelector("form")!.requestSubmit();
    expect(commands).toEqual([{ kind: "submit_line", text: "ls -la" }]);
  } finally { dispose(); }
});

it.skipIf(!browser)("announces a miss as text with a 1-based position, not by colour", async () => {
  const { target, dispose } = render(typing({ last_submission: { outcome: "miss", first_mismatch_index: 4 }, misses: 1, current_prompt_misses: 1 }), []);
  try {
    await settle();
    const status = target.querySelector(".feedback")!;
    expect(status.getAttribute("role")).toBe("status");
    expect(status.textContent).toContain("First difference at character 5.");
    await assertAxe(target, "miss");
  } finally { dispose(); }
});

it.skipIf(!browser)("shows remaining time as text and the expired state with end_run offered", async () => {
  const commands: TyperCommand[] = [];
  const running = render(typing({ assist_level: "timed", deadline_server_ms: 121_000, last_server_ms: 1_000 }), commands);
  try {
    await settle();
    expect(running.target.textContent).toContain("120 seconds left");
  } finally { running.dispose(); }
  const expired = render(typing({ assist_level: "timed", deadline_server_ms: 121_000, last_server_ms: 130_000 }), commands);
  try {
    await settle();
    expect(expired.target.textContent).toContain("Time is up.");
    await assertAxe(expired.target, "expired");
    button(expired.target, "End run").click();
    expect(commands).toEqual([{ kind: "end_run" }]);
  } finally { expired.dispose(); }
});

it.skipIf(!browser)("reflows the longest permitted prompt at 320 px without horizontal overflow", async () => {
  const longest = "x".repeat(256);
  const { target, dispose } = render(typing({ current_prompt_text: longest }), [], "320px");
  try {
    await settle();
    const prompt = target.querySelector<HTMLElement>(".prompt")!;
    expect(prompt.scrollWidth).toBeLessThanOrEqual(prompt.clientWidth + 1);
    expect(target.scrollWidth).toBeLessThanOrEqual(320 + 1);
  } finally { dispose(); }
});
