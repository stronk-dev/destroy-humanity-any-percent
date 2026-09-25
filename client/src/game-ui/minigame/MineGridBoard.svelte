<script lang="ts">
  import { t, type CopyEra, type CopyKey } from "../../copy";
  import type { ArcadePreset } from "../../arcade/catalog";
  import type { MineGridCommand, MineGridSnapshot } from "../../arcade/mine-grid";

  // AR6.2 tenant child for (mine_grid, 1.0.0). Presentation only: it renders
  // the server snapshot and raises commands; the host owns transport. No
  // timer is shown — the certified result has none.
  let { snapshot, presets, era, pending, onCommand }: {
    snapshot: MineGridSnapshot;
    presets: readonly ArcadePreset[];
    era: CopyEra;
    pending: boolean;
    onCommand(command: MineGridCommand): void;
  } = $props();

  let mode = $state<"reveal" | "flag">("reveal");
  let focused = $state(0);
  let confirmingQuit = $state(false);
  let cells: HTMLButtonElement[] = $state([]);

  const cellCount = $derived(snapshot.width * snapshot.height);
  const revealed = $derived(new Map(snapshot.revealed.map((row) => [row.cell, row.adjacent])));
  const flags = $derived(new Set(snapshot.flags));
  const mines = $derived(new Set(snapshot.mine_cells));
  const outcome = $derived<CopyKey | null>(snapshot.phase !== "terminal" ? null
    : snapshot.exploded_cell !== -1 ? "arcade.outcome.detonated"
    : snapshot.preset_id !== null && snapshot.revealed.length === cellCount - snapshot.mines ? "arcade.outcome.cleared" : "arcade.outcome.quit");
  const announcement = $derived(outcome ? t(outcome, {}, era) : snapshot.phase === "playing" ? t("arcade.mine_grid.cells_open_frame", { count: snapshot.revealed.length }, era) : "");

  function label(cell: number): string {
    const row = Math.floor(cell / snapshot.width) + 1, column = cell % snapshot.width + 1;
    const adjacent = revealed.get(cell);
    if (adjacent !== undefined) return t("arcade.mine_grid.cell.revealed", { adjacent, column, row }, era);
    if (flags.has(cell)) return t("arcade.mine_grid.cell.flagged", { column, row }, era);
    return t("arcade.mine_grid.cell.hidden", { column, row }, era);
  }

  function act(cell: number, kind: "reveal" | "toggle_flag" | "chord"): void {
    if (pending || snapshot.phase !== "playing") return;
    onCommand({ kind, cell });
  }

  function primary(cell: number): void {
    if (revealed.has(cell)) act(cell, "chord");
    else act(cell, mode === "flag" ? "toggle_flag" : "reveal");
  }

  function move(next: number): void {
    focused = Math.max(0, Math.min(cellCount - 1, next));
    cells[focused]?.focus();
  }

  function keydown(event: KeyboardEvent, cell: number): void {
    const width = snapshot.width;
    const moves: Record<string, number> = { ArrowLeft: cell % width === 0 ? cell : cell - 1, ArrowRight: cell % width === width - 1 ? cell : cell + 1,
      ArrowUp: cell - width >= 0 ? cell - width : cell, ArrowDown: cell + width < cellCount ? cell + width : cell };
    if (event.key in moves) { event.preventDefault(); move(moves[event.key]!); return; }
    if (event.key === "f" || event.key === "F") { event.preventDefault(); act(cell, "toggle_flag"); return; }
    if (event.key === "c" || event.key === "C") { event.preventDefault(); act(cell, "chord"); }
  }
</script>

<section class="mine-grid" aria-labelledby="mine-grid-heading">
  <h2 id="mine-grid-heading">{t("arcade.toy.mine_grid.title", {}, era)}</h2>
  <p class="live" role="status" aria-live="polite">{announcement}</p>

  {#if snapshot.phase === "setup"}
    <p>{t("arcade.toy.mine_grid.description", {}, era)}</p>
    <div class="presets">
      {#each presets as preset (preset.preset_id)}
        <button type="button" disabled={pending} onclick={() => onCommand({ kind: "choose_board", preset_id: preset.preset_id })}>{t(preset.copy_key as CopyKey, {}, era)}</button>
      {/each}
    </div>
  {:else}
    {#if snapshot.phase === "playing"}
      <div class="toolbar">
        <div role="group" aria-label={t("arcade.toy.mine_grid.title", {}, era)}>
          <button type="button" aria-pressed={mode === "reveal"} onclick={() => { mode = "reveal"; }}>{t("arcade.mine_grid.mode.reveal", {}, era)}</button>
          <button type="button" aria-pressed={mode === "flag"} onclick={() => { mode = "flag"; }}>{t("arcade.mine_grid.mode.flag", {}, era)}</button>
        </div>
        <output>{t("arcade.mine_grid.flags_remaining", { count: snapshot.mines - snapshot.flags.length }, era)}</output>
      </div>
    {/if}
    <div class="scroller" role="region" aria-label={t("arcade.mine_grid.board_label", {}, era)} tabindex="-1">
      <div class="grid" role="grid" aria-labelledby="mine-grid-heading" style:--mine-grid-columns={snapshot.width}>
        {#each { length: snapshot.height } as _, row (row)}
          <div role="row" class="row">
            {#each { length: snapshot.width } as _, column (column)}
              {@const cell = row * snapshot.width + column}
              {@const adjacent = revealed.get(cell)}
              <button type="button" role="gridcell" bind:this={cells[cell]} tabindex={cell === focused ? 0 : -1}
                class="cell" data-state={adjacent !== undefined ? "revealed" : flags.has(cell) ? "flagged" : mines.has(cell) ? "mine" : "hidden"}
                aria-label={label(cell)} aria-disabled={pending || snapshot.phase !== "playing" ? "true" : undefined}
                onfocus={() => { focused = cell; }} onclick={() => primary(cell)} onkeydown={(event) => keydown(event, cell)}>
                {#if adjacent !== undefined && adjacent > 0}{adjacent}{:else if flags.has(cell)}{t("arcade.mine_grid.glyph.flag", {}, era)}{:else if mines.has(cell)}{t("arcade.mine_grid.glyph.mine", {}, era)}{/if}
              </button>
            {/each}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if snapshot.phase !== "terminal"}
    {#if confirmingQuit}
      <p>{t("arcade.action.quit_confirm", {}, era)}</p>
      <button type="button" disabled={pending} onclick={() => { confirmingQuit = false; onCommand({ kind: "quit" }); }}>{t("arcade.action.quit", {}, era)}</button>
      <button type="button" onclick={() => { confirmingQuit = false; }}>{t("arcade.action.resume", {}, era)}</button>
    {:else}
      <button type="button" disabled={pending} onclick={() => { confirmingQuit = true; }}>{t("arcade.action.quit", {}, era)}</button>
    {/if}
  {/if}
</section>

<style>
  .mine-grid { display: grid; gap: var(--cc-space-md); }
  h2, p { margin: 0; }
  .presets, .toolbar { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); align-items: center; }
  .scroller { max-inline-size: 100%; max-block-size: 70vh; overflow: auto; }
  .grid { display: grid; gap: var(--cc-space-xs); inline-size: max-content; }
  .row { display: grid; grid-template-columns: repeat(var(--mine-grid-columns), 1.75rem); gap: var(--cc-space-xs); }
  .cell { min-inline-size: 1.75rem; min-block-size: 1.75rem; padding: 0; border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); background: var(--cc-chrome-button_face); color: var(--cc-color-text); font-family: var(--cc-type-font_mono); }
  .cell[data-state="revealed"] { background: var(--cc-color-surface); }
  .cell[data-state="mine"] { background: var(--cc-color-surface); font-weight: var(--cc-type-weight_bold); }
  .cell:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  button[aria-pressed="true"] { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
</style>
