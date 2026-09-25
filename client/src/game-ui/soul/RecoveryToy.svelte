<script lang="ts">
  import type { SoulToyKind } from "../../soul/catalog";
  import { toyCellOrder } from "./recovery-surface";

  // Presentation only (SR-C3): the toy receives a local decorative seed and
  // the server-reported progress fraction. It is hidden from assistive tech;
  // the surface's <progress> element carries the accessible value. Nothing
  // animates, so reduced motion needs no special case.
  let { toyKind, seed, fraction }: { toyKind: SoulToyKind; seed: number; fraction: number } = $props();

  const CELLS = 48;
  const order = $derived(toyCellOrder(seed, CELLS));
  const filled = $derived(Math.max(0, Math.min(CELLS, Math.floor(fraction * CELLS))));
  const settled = $derived(new Set(order.slice(0, filled)));
</script>

<div class="toy" data-toy={toyKind} aria-hidden="true">
  {#each { length: CELLS } as _, index (index)}
    <span class="cell" data-settled={settled.has(index)}></span>
  {/each}
</div>

<style>
  .toy { display: grid; grid-template-columns: repeat(12, 1fr); gap: var(--cc-space-xs); max-width: 24rem; }
  .cell { aspect-ratio: 1; border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); background: var(--cc-color-surface); }
  .cell[data-settled="true"] { background: var(--cc-color-accent); }
  .toy[data-toy="repot"] .cell { border-radius: var(--cc-border-radius); }
  .toy[data-toy="server_room"] { grid-template-columns: repeat(8, 1fr); }
</style>
