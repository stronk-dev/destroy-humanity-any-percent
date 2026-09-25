<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  import { t, type CopyEra, type CopyKey } from "../../copy";
  import type { ArcadeSnakeContent } from "../../arcade/catalog";
  import { snakeOpposite, snakeStep, type SnakeCommand, type SnakeDirection, type SnakeMutable, type SnakeSnapshot, type SnakeTurn } from "../../arcade/snake";

  // AR6.3 tenant child for (snake, 1.0.0). The client runs the shared engine
  // locally at presentation pace and flushes `advance` batches; the server
  // snapshot is the only truth (divergence resyncs). Transport stays with the
  // host: `submit` and `current` are its callbacks.
  let { initial, content, era, submit, current, tickMS = content.presentation_tick_ms, flushEvery = 16 }: {
    initial: SnakeSnapshot;
    content: ArcadeSnakeContent;
    era: CopyEra;
    submit(command: SnakeCommand): Promise<SnakeSnapshot>;
    current(): Promise<SnakeSnapshot>;
    tickMS?: number;
    flushEvery?: number;
  } = $props();

  const PACES = [["arcade.snake.pace.full", 1], ["arcade.snake.pace.three_quarter", 0.75], ["arcade.snake.pace.half", 0.5]] as const;

  // The host mounts one child per session: `initial` seeds local state once
  // and every later truth arrives through submit/current.
  // svelte-ignore state_referenced_locally
  let server = $state<SnakeSnapshot>(initial);
  // svelte-ignore state_referenced_locally
  let local = $state<SnakeMutable>(clone(initial));
  let turns: SnakeTurn[] = [];
  let desired: SnakeDirection | undefined;
  let inFlight = false;
  let paused = $state(true);
  let pace = $state(1);
  let notice = $state<CopyKey | null>(null);
  let announcement = $state("");
  let announcedScore = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let destroyed = false;
  let root: HTMLElement | undefined;

  const cells = $derived(local.width * local.height);
  const body = $derived(new Set(local.body));
  const terminal = $derived(local.phase === "terminal" || server.phase === "terminal");

  function clone(value: SnakeSnapshot): SnakeMutable { return { ...value, body: [...value.body] }; }

  onMount(() => {
    const hidden = () => { if (document.visibilityState === "hidden") pause(); };
    document.addEventListener("visibilitychange", hidden);
    window.addEventListener("blur", pause);
    return () => { document.removeEventListener("visibilitychange", hidden); window.removeEventListener("blur", pause); };
  });
  onDestroy(() => { destroyed = true; clearTimeout(timer); });

  function schedule(): void {
    clearTimeout(timer);
    if (paused || terminal || destroyed) return;
    timer = setTimeout(step, Math.round(tickMS / pace));
  }

  function step(): void {
    if (paused || terminal || destroyed) return;
    // Freeze while the unacknowledged local lead is at the advance window.
    if (local.tick - server.tick >= content.max_ticks_per_advance) { void flush(); schedule(); return; }
    const tick = local.tick + 1;
    if (desired && desired !== local.direction && desired !== snakeOpposite(local.direction)) {
      turns.push({ tick, direction: desired });
      local.direction = desired;
    }
    desired = undefined;
    const outcome = snakeStep(local, tick, content.growth_per_food, BigInt(local.food_seed));
    if (outcome) { local.phase = "terminal"; announcement = t_outcome(outcome); }
    else if (local.score > announcedScore) { announcedScore = local.score; announcement = t("arcade.snake.score_frame", { count: local.score }, era); }
    if (outcome || local.tick - server.tick >= flushEvery) void flush();
    schedule();
  }

  function t_outcome(outcome: "cleared" | "crashed"): string { return t(outcome === "cleared" ? "arcade.outcome.cleared" : "arcade.outcome.crashed", {}, era); }

  async function flush(): Promise<void> {
    if (inFlight || local.tick <= server.tick) return;
    inFlight = true;
    const through = local.tick;
    const batch = turns.filter((turn) => turn.tick > server.tick && turn.tick <= through);
    try {
      const next = await submit({ kind: "advance", through_tick: through, turns: batch });
      if (destroyed) return;
      if (next.tick !== through) { await resync(); return; }
      server = next;
      turns = turns.filter((turn) => turn.tick > through);
      if (next.phase === "terminal" && local.tick === through) local = clone(next);
    } catch {
      await resync();
    } finally { inFlight = false; }
  }

  async function resync(): Promise<void> {
    try {
      const truth = await current();
      if (destroyed) return;
      server = truth;
      local = clone(truth);
      turns = [];
      notice = "arcade.resync.notice";
    } catch { notice = "arcade.error.rejected"; }
  }

  function steer(direction: SnakeDirection): void {
    if (terminal) return;
    desired = direction;
    if (paused) resume();
  }

  function pause(): void { if (!paused) { paused = true; clearTimeout(timer); } }
  function resume(): void { if (terminal) return; paused = false; notice = null; schedule(); }

  async function quit(): Promise<void> {
    pause();
    try {
      await flush();
      const next = await submit({ kind: "quit" });
      server = next; local = clone(next); announcement = t("arcade.outcome.quit", {}, era);
    } catch { await resync(); }
  }

  function keydown(event: KeyboardEvent): void {
    const map: Record<string, SnakeDirection> = { ArrowUp: "up", ArrowDown: "down", ArrowLeft: "left", ArrowRight: "right", w: "up", s: "down", a: "left", d: "right", W: "up", S: "down", A: "left", D: "right" };
    if (event.key in map) { event.preventDefault(); steer(map[event.key]!); return; }
    if (event.key === "p" || event.key === "P" || event.key === "Escape") { event.preventDefault(); if (paused) resume(); else pause(); }
  }

  function focusout(event: FocusEvent): void {
    if (!root?.contains(event.relatedTarget as Node | null)) pause();
  }
</script>

<!-- focusout here implements AR6.3 auto-pause when focus leaves the toy. -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<section class="snake" aria-labelledby="snake-heading" bind:this={root} onfocusout={focusout}>
  <h2 id="snake-heading">{t("arcade.toy.snake.title", {}, era)}</h2>
  <p class="live" role="status" aria-live="polite">{announcement}</p>
  {#if notice}<p role="status">{t(notice, {}, era)}</p>{/if}
  <output>{t("arcade.snake.score_frame", { count: local.score }, era)}</output>
  <!-- The board is a focusable real-time control surface (AR6.3); keys steer it. -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex a11y_no_noninteractive_element_interactions -->
  <div class="board" role="application" tabindex="0" aria-label={t("arcade.snake.board_label", {}, era)} aria-describedby="snake-heading" onkeydown={keydown}
    style:--snake-columns={local.width}>
    {#each { length: cells } as _, cell (cell)}
      <span class="cell" aria-hidden="true" data-state={cell === local.body[0] ? "head" : body.has(cell) ? "body" : cell === local.food_cell ? "food" : "empty"}></span>
    {/each}
  </div>
  {#if paused && !terminal}<p>{t("arcade.snake.paused", {}, era)}</p>{/if}
  <div class="controls">
    {#if !terminal}
      <button type="button" onclick={() => (paused ? resume() : pause())}>{paused ? t(local.tick === 0 ? "arcade.action.play" : "arcade.action.resume", {}, era) : t("arcade.action.pause", {}, era)}</button>
      <div class="dpad" role="group" aria-label={t("arcade.snake.board_label", {}, era)}>
        <button type="button" class="up" onclick={() => steer("up")}>{t("arcade.snake.dpad.up", {}, era)}</button>
        <button type="button" class="left" onclick={() => steer("left")}>{t("arcade.snake.dpad.left", {}, era)}</button>
        <button type="button" class="right" onclick={() => steer("right")}>{t("arcade.snake.dpad.right", {}, era)}</button>
        <button type="button" class="down" onclick={() => steer("down")}>{t("arcade.snake.dpad.down", {}, era)}</button>
      </div>
      <label>{t("arcade.snake.pace_label", {}, era)}
        <select bind:value={pace} onchange={schedule}>
          {#each PACES as [key, value] (key)}<option value={value}>{t(key, {}, era)}</option>{/each}
        </select>
      </label>
      <button type="button" onclick={quit}>{t("arcade.action.quit", {}, era)}</button>
    {/if}
  </div>
</section>

<style>
  .snake { display: grid; gap: var(--cc-space-md); }
  h2, p { margin: 0; }
  .board { display: grid; grid-template-columns: repeat(var(--snake-columns), 1fr); gap: var(--cc-space-xs); max-inline-size: 30rem; border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); padding: var(--cc-space-xs); background: var(--cc-color-bg); }
  .board:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
  .cell { aspect-ratio: 1; background: var(--cc-color-surface); }
  .cell[data-state="body"] { background: var(--cc-color-accent); }
  .cell[data-state="head"] { background: var(--cc-color-text); }
  .cell[data-state="food"] { background: var(--cc-color-border); border-radius: var(--cc-border-radius); }
  .controls { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); align-items: center; }
  .dpad { display: grid; grid-template-areas: ". up ." "left . right" ". down ."; gap: var(--cc-space-xs); }
  .dpad button { min-inline-size: 2.75rem; min-block-size: 2.75rem; }
  .up { grid-area: up; } .down { grid-area: down; } .left { grid-area: left; } .right { grid-area: right; }
</style>
