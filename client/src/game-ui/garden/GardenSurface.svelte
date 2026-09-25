<script lang="ts">
  import { onDestroy, onMount, tick, untrack } from "svelte";

  import type { GardenCurrentResponse } from "../../api/generated/types";
  import { COPY_KEYS, t, type CopyEra, type CopyKey } from "../../copy";
  import { tickDelta, type GardenPort, type GardenViewState } from "./garden-port";

  // Server Garden SG10: a DOM-first grid of native buttons over the advisory
  // read (no client simulation). The surface re-reads at next_tick_wall_ms
  // while visible and whenever the host bumps refreshKey after a receipt.
  let { port, era, pending, refreshKey, rejection, onPlant, onUproot, onHarvest, onSetSubstrate, visible = () => document.visibilityState === "visible", now = () => Date.now() }: {
    port: GardenPort;
    era: CopyEra;
    pending: boolean;
    refreshKey: number;
    rejection: CopyKey | null;
    onPlant(row: number, col: number, speciesID: string): void;
    onUproot(row: number, col: number): void;
    onHarvest(plots: readonly { row: number; col: number }[]): void;
    onSetSubstrate(substrateID: string): void;
    visible?: () => boolean;
    now?: () => number;
  } = $props();

  type ActiveView = Extract<GardenCurrentResponse, { kind: "active" }>;
  let viewState = $state<GardenViewState>({ kind: "loading" });
  let announcement = $state("");
  let focusRow = $state(0);
  let focusCol = $state(0);
  let menu = $state<{ row: number; col: number } | null>(null);
  let timer: ReturnType<typeof setTimeout> | undefined;
  let destroyed = false;
  const cells = new Map<string, HTMLButtonElement>();

  // Substrates resolve from the copy registry (no second hardcoded catalog):
  // every garden.substrate.<id>.name key is one selectable substrate.
  const SUBSTRATES = COPY_KEYS.filter((key) => /^garden\.substrate\.[a-z][a-z0-9_]*\.name$/u.test(key)).map((key) => key.split(".")[2]!).sort();
  const active = $derived(viewState.kind === "active" || viewState.kind === "stale" ? viewState.view : null);

  onMount(() => {
    const listener = () => { if (visible()) void load(); };
    document.addEventListener("visibilitychange", listener);
    return () => document.removeEventListener("visibilitychange", listener);
  });
  onDestroy(() => { destroyed = true; if (timer !== undefined) clearTimeout(timer); });

  // Re-read on every refreshKey bump; load() itself must not become a dependency.
  $effect(() => { void refreshKey; untrack(() => { void load(); }); });

  async function load(): Promise<void> {
    if (viewState.kind === "active") viewState = { kind: "stale", view: viewState.view };
    try {
      const next = await port.current();
      if (destroyed) return;
      if (next.kind === "inactive") { viewState = { kind: "error" }; return; }
      if (next.kind === "locked") { viewState = { kind: "locked" }; return; }
      const delta = tickDelta(viewState, next);
      if (delta.matured > 0 || delta.spawned > 0) announcement = t("garden.announce.tick", { matured: delta.matured, spawned: delta.spawned }, era);
      viewState = { kind: "active", view: next };
      schedule(next);
    } catch {
      if (!destroyed) viewState = active ? { kind: "stale", view: active } : { kind: "error" };
    }
  }

  function schedule(view: ActiveView): void {
    if (timer !== undefined) clearTimeout(timer);
    const next = view.garden.next_tick_wall_ms;
    if (next === null) return;
    const wait = Math.max(1_000, next - view.server_ms);
    timer = setTimeout(() => { timer = undefined; if (visible()) void load(); }, wait);
  }

  function plotAt(view: ActiveView, row: number, col: number) { return view.garden.plots.find((plot) => plot.row === row && plot.col === col); }
  function speciesName(id: string): string { return t(`garden.species.${id}.name` as CopyKey, {}, era); }
  function stageOf(view: ActiveView, row: number, col: number): "empty" | "dormant" | "growing" | "mature" {
    const plot = plotAt(view, row, col);
    // The server's dormant flag wins: a planted plot outside the active grid
    // neither grows nor counts, so it reads as dormant, never "growing".
    if (plot) return plot.dormant ? "dormant" : plot.stage;
    return row >= view.garden.height || col >= view.garden.width ? "dormant" : "empty";
  }
  function stageText(stage: "empty" | "dormant" | "growing" | "mature"): string {
    return stage === "empty" ? t("garden.plot.empty", {}, era) : stage === "dormant" ? t("garden.plot.dormant", {}, era) : t(`garden.stage.${stage}`, {}, era);
  }
  function label(view: ActiveView, row: number, col: number): string {
    const plot = plotAt(view, row, col);
    return t("garden.plot.label", { row: row + 1, col: col + 1, species: plot ? speciesName(plot.species_id) : t("garden.plot.empty", {}, era), stage: stageText(stageOf(view, row, col)) }, era);
  }
  function mature(view: ActiveView) { return view.garden.plots.filter((plot) => plot.stage === "mature").map((plot) => ({ row: plot.row, col: plot.col })); }

  async function move(event: KeyboardEvent, view: ActiveView): Promise<void> {
    const deltas: Record<string, [number, number]> = { ArrowUp: [-1, 0], ArrowDown: [1, 0], ArrowLeft: [0, -1], ArrowRight: [0, 1] };
    const delta = deltas[event.key];
    if (!delta) return;
    event.preventDefault();
    focusRow = Math.min(view.garden.max_height - 1, Math.max(0, focusRow + delta[0]));
    focusCol = Math.min(view.garden.max_width - 1, Math.max(0, focusCol + delta[1]));
    await tick();
    cells.get(`${focusRow},${focusCol}`)?.focus();
  }
  function register(node: HTMLButtonElement, key: string) { cells.set(key, node); return { destroy: () => cells.delete(key) }; }
  function open(row: number, col: number): void { focusRow = row; focusCol = col; menu = { row, col }; }
  async function closeMenu(): Promise<void> { const at = menu; menu = null; await tick(); if (at) cells.get(`${at.row},${at.col}`)?.focus(); }
  function act(run: () => void): void { run(); void closeMenu(); }
</script>

<section class="garden cc-window" aria-labelledby="garden-heading" aria-busy={pending || viewState.kind === "loading"}>
  <h1 id="garden-heading">{t("garden.title", {}, era)}</h1>
  <p class="hint">{t("garden.hint.first", {}, era)}</p>
  <small title={t("garden.why", {}, era)}>{t("garden.why", {}, era)}</small>
  <p class="live" role="status" aria-live="polite">{announcement}</p>
  {#if rejection}<p role="alert">{t(rejection, {}, era)}</p>{/if}

  {#if viewState.kind === "loading"}
    <p>{t("garden.state.loading", {}, era)}</p>
  {:else if viewState.kind === "locked"}
    <p>{t("garden.state.locked", {}, era)}</p>
  {:else if viewState.kind === "error"}
    <p role="alert">{t("garden.state.error", {}, era)}</p>
    <button type="button" onclick={() => load()}>{t("garden.action.refresh", {}, era)}</button>
  {:else if active}
    {#if viewState.kind === "stale"}<p role="status">{t("garden.state.stale", {}, era)}</p>{/if}
    <p>{t("garden.collection.progress", { count: active.garden.seed_collection.length, total: active.garden.species_total }, era)}</p>
    <div class="grid" role="grid" aria-labelledby="garden-heading" aria-rowcount={active.garden.max_height} aria-colcount={active.garden.max_width}>
      {#each { length: active.garden.max_height } as _, row (row)}
        <div class="row" role="row">
          {#each { length: active.garden.max_width } as _, col (col)}
            {@const stage = stageOf(active, row, col)}
            <div role="gridcell">
              <button type="button" class="cell" data-stage={stage} tabindex={row === focusRow && col === focusCol ? 0 : -1}
                aria-label={label(active, row, col)} disabled={pending} use:register={`${row},${col}`}
                onkeydown={(event) => move(event, active)} onclick={() => open(row, col)}>
                <span aria-hidden="true">{t(`garden.glyph.${stage}`, {}, era)}</span>
              </button>
            </div>
          {/each}
        </div>
      {/each}
    </div>
    {#if menu}
      {@const plot = plotAt(active, menu.row, menu.col)}
      {@const at = menu}
      <div class="menu" role="group" aria-label={label(active, at.row, at.col)}>
        {#if !plot && stageOf(active, at.row, at.col) === "empty"}
          <p>{t("garden.plot.choose_seed", {}, era)}</p>
          {#each active.garden.seed_collection as species (species)}
            <button type="button" disabled={pending} onclick={() => act(() => onPlant(at.row, at.col, species))}>{t("garden.action.plant_frame", { species: speciesName(species) }, era)}</button>
          {/each}
        {/if}
        {#if plot?.stage === "mature"}
          <button type="button" disabled={pending} onclick={() => act(() => onHarvest([{ row: at.row, col: at.col }]))}>{t("garden.action.harvest", {}, era)}</button>
        {/if}
        {#if plot}
          <button type="button" disabled={pending} onclick={() => act(() => onUproot(at.row, at.col))}>{t("garden.action.uproot", {}, era)}</button>
        {/if}
        <button type="button" onclick={() => closeMenu()}>{t("garden.action.close", {}, era)}</button>
      </div>
    {/if}
    {#if mature(active).length > 0}
      <button type="button" disabled={pending} onclick={() => onHarvest(mature(active))}>{t("garden.action.harvest_all", {}, era)}</button>
    {/if}
    <fieldset class="substrates" disabled={pending || active.garden.substrate_lockout_until_ms !== null}>
      <legend>{t("garden.substrate.label", {}, era)}</legend>
      {#each SUBSTRATES as substrate (substrate)}
        <button type="button" aria-pressed={active.garden.substrate_id === substrate} aria-describedby={`garden-substrate-${substrate}`}
          onclick={() => { if (active.garden.substrate_id !== substrate) onSetSubstrate(substrate); }}>{t(`garden.substrate.${substrate}.name` as CopyKey, {}, era)}</button>
        <small id={`garden-substrate-${substrate}`}>{t(`garden.substrate.${substrate}.tooltip` as CopyKey, {}, era)}</small>
      {/each}
    </fieldset>
    {#if active.garden.substrate_lockout_until_ms !== null}<p>{t("cap.garden_substrate_lockout", {}, era)}</p>{/if}
    {#if active.garden.pending_catchup_forfeited_ms > 0}<p>{t("cap.garden_catchup", {}, era)}</p>{/if}
    {#if active.garden.width === active.garden.max_width && active.garden.height === active.garden.max_height}<p>{t("cap.garden_grid", {}, era)}</p>{/if}
    {#if active.garden.next_tick_wall_ms !== null}
      <p>{t("garden.next_tick", { remaining: `${Math.max(0, Math.ceil((active.garden.next_tick_wall_ms - active.server_ms) / 1000))}s` }, era)}</p>
    {/if}
  {/if}
</section>

<style>
  .garden { display: grid; gap: var(--cc-space-md); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); color: var(--cc-color-text); }
  h1 { margin: 0; font-family: var(--cc-type-font_display); }
  p { margin: 0; }
  .live { min-block-size: 1em; }
  .grid { display: grid; gap: var(--cc-space-xs); max-inline-size: 100%; }
  .row { display: grid; grid-template-columns: repeat(6, minmax(1.5rem, 1fr)); gap: var(--cc-space-xs); }
  .cell { inline-size: 100%; min-block-size: 2.75rem; padding: 0; border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); background: var(--cc-color-surface); color: var(--cc-color-text); font-family: var(--cc-type-font_ui); }
  .cell[data-stage="mature"] { border-style: double; }
  .cell[data-stage="dormant"] { border-style: dotted; }
  .cell:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
  .menu, .substrates { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); align-items: center; }
</style>
