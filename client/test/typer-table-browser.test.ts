import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import content from "../../balance/testdata/typer-v1.json?raw";
import { t } from "../src/copy";
import TyperTable from "../src/game-ui/minigame/TyperTable.svelte";
import { typerContentHash } from "../src/typer/catalog";
import { applyTyper, createTyper, type TyperCommand, type TyperResult, type TyperSnapshot } from "../src/typer/engine";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import TyperTableHarness from "./TyperTableHarness.svelte";

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

function render(value: TyperSnapshot, commands: TyperCommand[], width?: string, exitToHost: () => void = () => {}) {
  const target = document.createElement("main");
  if (width) target.style.width = width;
  document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  const app = mount(TyperTable, { target, props: { snapshot: value, serverTimeSample: value.last_server_ms ?? 1_000, pending: false, era: "era_2000",
    dispatch: (command: TyperCommand) => commands.push(command), exitToHost } });
  return { target, dispose: () => { unmount(app); target.remove(); } };
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}: ${target.textContent}`);
  return found;
}

interface ControlledTyper {
  advance(next: TyperSnapshot): void;
  setPending(value: boolean): void;
}

function controlled(value: TyperSnapshot, dispatch: (command: TyperCommand) => void, exitToHost: () => void = () => {}) {
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  const component = mount(TyperTableHarness, { target, props: { initial: value, dispatch, exitToHost } });
  return { target, app: component as unknown as ControlledTyper, dispose: () => { unmount(component); target.remove(); } };
}

for (const action of ["Timed run", "Untimed run", "Enter", "End run"] as const) for (const key of ["{Enter}", " "] as const) {
  it.skipIf(!browser)(`Typer native ${action} ${key} keeps pending focus and refuses duplicates`, async () => {
    const { userEvent } = await import("vitest/browser");
    const commands: TyperCommand[] = [];
    const value = action === "Timed run" || action === "Untimed run" ? snapshot() : typing();
    const view = controlled(value, (command) => {
      commands.push(command);
      flushSync(() => view.app.setPending(true));
    });
    try {
      await settle();
      const trigger = button(view.target, action);
      trigger.focus();
      await userEvent.keyboard(key);
      await settle();
      expect(commands).toHaveLength(1);
      expect(document.activeElement, "pending must retain native action focus").toBe(trigger);
      expect(trigger.disabled, "pending must remain keyboard reachable").toBe(false);
      expect(trigger.getAttribute("aria-disabled")).toBe("true");
      expect(view.target.querySelector("section")?.getAttribute("aria-busy")).toBe("true");
      await userEvent.keyboard("{Enter} ");
      trigger.click(); // A focusable pending control also needs a handler guard.
      expect(commands).toHaveLength(1);
      flushSync(() => view.app.setPending(false));
      await settle();
      await userEvent.keyboard(key);
      expect(commands).toHaveLength(2);
    } finally { view.dispose(); }
  });
}

for (const boundary of ["before-response", "before-focus-tick"] as const) {
  it.skipIf(!browser)(`Typer prompt advance respects newer focus ${boundary}`, async () => {
    const view = controlled(typing(), () => {});
    const outside = document.createElement("button");
    view.target.after(outside);
    try {
      await settle();
      const chosen = boundary === "before-response" ? button(view.target, "Leave the table") : outside;
      if (boundary === "before-response") chosen.focus();
      flushSync(() => view.app.advance(typing({ current_prompt_id: "cd_www", current_prompt_text: "cd /var/www", prompt_index: 1, revision: 3 })));
      if (boundary === "before-focus-tick") chosen.focus();
      await settle();
      expect(document.activeElement).toBe(chosen);
      expect(view.target.querySelector<HTMLInputElement>("input")?.value).toBe("");
      expect(view.target.querySelector(".prompt-announcement")?.textContent).toBe("cd /var/www");
    } finally { view.dispose(); outside.remove(); }
  });
}

for (const preserveNewerFocus of [false, true]) {
  it.skipIf(!browser)(`Typer terminal handoff ${preserveNewerFocus ? "preserves newer focus" : "reaches Leave"}`, async () => {
    const view = controlled(typing(), () => {});
    const outside = document.createElement("button");
    view.target.after(outside);
    try {
      await settle();
      expect(document.activeElement).toBe(view.target.querySelector("input"));
      flushSync(() => view.app.advance(snapshot({ phase: "terminal", revision: 3 })));
      if (preserveNewerFocus) outside.focus();
      await settle();
      expect(document.activeElement).toBe(preserveNewerFocus ? outside : button(view.target, "Leave the table"));
    } finally { view.dispose(); outside.remove(); }
  });
}

for (const phase of ["ready", "typing"] as const) {
  it.skipIf(!browser)(`Typer ${phase} native Tab and Shift-Tab reach every action without a trap`, async () => {
    const { userEvent } = await import("vitest/browser");
    const commands: TyperCommand[] = [];
    const view = controlled(phase === "ready" ? snapshot() : typing(), (command) => commands.push(command));
    const before = document.createElement("button"), after = document.createElement("button");
    before.tabIndex = 0; after.tabIndex = 0;
    view.target.before(before); view.target.after(after);
    try {
      await settle();
      const stops = phase === "ready"
        ? ["Timed run", "Untimed run", "End run", "Leave the table"]
        : ["Enter", "End run", "Leave the table"];
      if (phase === "ready") before.focus();
      else expect(document.activeElement).toBe(view.target.querySelector("input"));
      for (const label of stops) {
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement, `forward ${label}`).toBe(button(view.target, label));
      }
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement, "Tab can leave the child").toBe(after);
      for (const label of [...stops].reverse()) {
        await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
        expect(document.activeElement, `backward ${label}`).toBe(button(view.target, label));
      }
      await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
      expect(document.activeElement).toBe(phase === "ready" ? before : view.target.querySelector("input"));
      expect(commands).toEqual([]);
    } finally { view.dispose(); before.remove(); after.remove(); }
  });
}

it.skipIf(!browser)("Typer native keyboard untimed run completes through the real pure engine and leaves the table", async () => {
  const { userEvent } = await import("vitest/browser");
  const identity = { content, content_hash: await typerContentHash(content), content_schema_version: 1,
    seed: 42n, mode: "solo" as const, scaling_inputs: { "typer.era_tier": 1 } };
  let encoded = await createTyper(identity);
  let state = JSON.parse(encoded) as TyperSnapshot;
  let result: TyperResult | null = null;
  let operation = Promise.resolve();
  let serverTime = 1_000;
  let exits = 0;
  const commands: TyperCommand[] = [];
  const view = controlled(state, (command) => {
    commands.push(command);
    flushSync(() => view.app.setPending(true));
    operation = applyTyper({ ...identity, snapshot: encoded, revision: state.revision,
      command: JSON.stringify(command), server_time_ms: serverTime++ }).then((output) => {
      encoded = output.snapshot;
      state = JSON.parse(encoded) as TyperSnapshot;
      result = output.result;
      flushSync(() => { view.app.advance(state); view.app.setPending(false); });
    });
  }, () => { exits++; });
  const sentinel = document.createElement("button");
  view.target.before(sentinel);
  try {
    await settle();
    sentinel.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement, "native Tab reaches timed choice").toBe(button(view.target, "Timed run"));
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement, "native Tab reaches untimed choice").toBe(button(view.target, "Untimed run"));
    await userEvent.keyboard("{Enter}");
    await operation; await settle();
    const input = view.target.querySelector<HTMLInputElement>("input")!;
    expect(document.activeElement).toBe(input);
    await userEvent.keyboard("WRONG{Enter}");
    await operation; await settle();
    expect(state.misses).toBe(1);
    expect(input.value).toBe("WRONG");
    expect(view.target.querySelector(".feedback")?.textContent).toBe(t("typer.feedback.miss", { index: 1 }, "era_2000"));
    await userEvent.keyboard("{Backspace}".repeat(5));
    const total = state.prompts_total;
    for (let index = 0; index < total; index++) {
      const prompt = view.target.querySelector(".prompt")?.textContent;
      expect(prompt, `current prompt ${index}`).toBe(state.current_prompt_text);
      expect(document.activeElement, `prompt ${index} input focus`).toBe(input);
      expect(input.value, `prompt ${index} input cleared`).toBe("");
      await userEvent.keyboard(`${prompt}{Enter}`);
      await operation; await settle();
      expect(state.lines_cleared).toBe(index + 1);
    }
    expect(state.phase).toBe("terminal");
    expect(result).toMatchObject({ outcome: "completed", rating_delta: null, score_facts: [
      { kind: "typer.assisted", value: 1 }, { kind: "typer.clean_lines", value: total - 1 },
      { kind: "typer.elapsed_ms", value: total + 1 }, { kind: "typer.lines_cleared", value: total }, { kind: "typer.misses", value: 1 },
    ] });
    expect(commands).toHaveLength(total + 2);
    expect(document.activeElement, "completion retains a reachable host exit").toBe(button(view.target, "Leave the table"));
    await assertAxe(view.target, "terminal");
    await userEvent.keyboard(" ");
    expect(exits).toBe(1);
  } finally { await operation; view.dispose(); sentinel.remove(); }
});

it.skipIf(!browser)("samples an increasing monotonic clock without reactive feedback across server revisions", async () => {
  const target = document.createElement("main"); document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  const commands: TyperCommand[] = [];
  let clock = 0;
  const app = mount(TyperTableHarness, { target, props: {
    initial: snapshot(), dispatch: (command: TyperCommand) => commands.push(command),
    // Every read changes. The ordinary browser clock can accidentally mask a
    // feedback dependency when successive reads return the same sample.
    monotonicNow: () => ++clock,
  } }) as unknown as { advance(next: TyperSnapshot): void };
  try {
    await settle();
    expect(clock).toBeGreaterThan(0);
    expect(button(target, "Untimed run").disabled).toBe(false);
    flushSync(() => app.advance(typing({ assist_level: "timed", deadline_server_ms: 121_000 })));
    await settle();
    expect(target.querySelector(".time")?.textContent).toBe("120 seconds left");
    flushSync(() => app.advance(typing({ assist_level: "timed", deadline_server_ms: 121_000, last_server_ms: 21_000, revision: 3 })));
    await settle();
    expect(target.querySelector(".time")?.textContent).toBe("100 seconds left");
    expect(commands).toEqual([]);
  } finally { unmount(app as never); target.remove(); }
});

it.skipIf(!browser)("offers timed and untimed without a default and begins by keyboard", async () => {
  const { userEvent } = await import("vitest/browser");
  const commands: TyperCommand[] = [];
  const { target, dispose } = render(snapshot(), commands);
  try {
    await settle();
    await assertAxe(target, "ready");
    expect(target.querySelector("[aria-pressed=true], [checked]")).toBeNull();
    const untimed = button(target, "Untimed run");
    expect(untimed.getAttribute("aria-describedby")).toBe("typer-untimed-note");
    untimed.focus();
    await userEvent.keyboard("{Enter}");
    expect(commands).toEqual([{ kind: "begin", assist_level: "untimed" }]);
  } finally { dispose(); }
});

it.skipIf(!browser)("submits the typed line on Enter, never during composition, and never blocks paste", async () => {
  const { userEvent } = await import("vitest/browser");
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
    input.focus();
    await userEvent.keyboard("{Enter}");
    expect(commands).toEqual([{ kind: "submit_line", text: "ls -la" }]);
  } finally { dispose(); }
});

it.skipIf(!browser)("ends the run and exits to the host using native keyboard activation", async () => {
  const { userEvent } = await import("vitest/browser");
  const commands: TyperCommand[] = [];
  let exits = 0;
  const { target, dispose } = render(typing(), commands, undefined, () => { exits++; });
  try {
    await settle();
    button(target, "End run").focus();
    await userEvent.keyboard("{Enter}");
    expect(commands).toEqual([{ kind: "end_run" }]);
    button(target, "Leave the table").focus();
    await userEvent.keyboard(" ");
    expect(exits).toBe(1);
  } finally { dispose(); }
});

it.skipIf(!browser)("announces each new prompt once while focus remains in the input", async () => {
  const target = document.createElement("main"); document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  const app = mount(TyperTableHarness, { target, props: { initial: typing(), dispatch: () => {} } }) as unknown as { advance(next: TyperSnapshot): void };
  try {
    await settle();
    const input = target.querySelector<HTMLInputElement>("#typer-line")!;
    expect(document.activeElement).toBe(input);
    const promptStatus = target.querySelectorAll(".prompt-announcement[aria-live=polite]");
    expect(promptStatus).toHaveLength(1);
    expect(promptStatus[0]?.textContent).toContain("ls -la");
    flushSync(() => app.advance(typing({ current_prompt_id: "cd_www", current_prompt_text: "cd /var/www", prompt_index: 1, revision: 3 })));
    await settle();
    expect(document.activeElement).toBe(input);
    expect(target.querySelectorAll(".prompt-announcement[aria-live=polite]")).toHaveLength(1);
    expect(promptStatus[0]?.textContent).toContain("cd /var/www");
    expect(promptStatus[0]?.textContent).not.toContain("ls -la");
  } finally { unmount(app as never); target.remove(); }
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
