<script lang="ts">
  import { PET_SWATCHES, type PetVisualSpec } from "./visual";

  // CSS-only nested-div sprite (PA8.5): no image assets. It is decorative and
  // aria-hidden; the owner surface carries the text alternative. Under
  // reduced motion `animate` is false and the pose is static.
  let { spec }: { spec: PetVisualSpec } = $props();
  const swatch = $derived(PET_SWATCHES.get(spec.palette_id)!);
</script>

<div class="sprite" aria-hidden="true" data-family={spec.family} data-pose={spec.pose} data-animate={spec.animate}
  style:--cc-pet-fur={swatch.fur} style:--cc-pet-outline={swatch.outline}>
  <div class="tail"></div>
  <div class="body"></div>
  <div class="head"><div class="ear left"></div><div class="ear right"></div></div>
</div>

<style>
  .sprite { position: relative; inline-size: calc(var(--cc-space-xl) * 3); block-size: calc(var(--cc-space-xl) * 2); flex: none; }
  .body, .head, .ear, .tail { position: absolute; background: var(--cc-pet-fur); border: var(--cc-border-width) var(--cc-border-style) var(--cc-pet-outline); }
  .body { inset-block-end: 0; inset-inline-start: 15%; inline-size: 60%; block-size: 55%; border-radius: var(--cc-space-xl); }
  .head { inset-block-end: 35%; inset-inline-end: 5%; inline-size: 38%; block-size: 50%; border-radius: var(--cc-space-lg); }
  .ear { inset-block-start: calc(var(--cc-space-sm) * -1); inline-size: 30%; block-size: 35%; }
  .ear.left { inset-inline-start: 8%; }
  .ear.right { inset-inline-end: 8%; }
  .tail { inset-block-end: 20%; inset-inline-start: 0; inline-size: 25%; block-size: 12%; border-radius: var(--cc-space-sm); }
  [data-pose="low"] .head { inset-block-end: 20%; }
  [data-pose="withdrawn"] .head { inset-block-end: 5%; inset-inline-end: 25%; }
  [data-pose="withdrawn"] .tail { inset-inline-start: 50%; }
  [data-animate="true"][data-pose="content"] .tail { animation-name: sway; animation-duration: var(--cc-motion-duration_slow); animation-timing-function: var(--cc-motion-easing); animation-iteration-count: infinite; animation-direction: alternate; }
  @keyframes sway { to { inset-block-end: 30%; } }
  @media (prefers-reduced-motion: reduce) { .tail { animation-name: none; } }
</style>
