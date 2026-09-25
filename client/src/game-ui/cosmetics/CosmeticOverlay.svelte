<script lang="ts">
  import type { WearerReaction } from "./presentation";

  // Cosmetic Shop v1 §7.4: a presentation-only CSS layer inside the CSS-only
  // sprite system. `annoyed` is ordinary cat annoyance (ears back, a tail
  // flick, a slow blink away) with no speech, text, or caption (I5); it never
  // reaches pet stats, Trust, mood, or behavior (I2). Static under reduced
  // motion. It renders no text node by construction.
  let { renderKey, reaction, animate }: { renderKey: string; reaction: WearerReaction; animate: boolean } = $props();
</script>

<div class="overlay" aria-hidden="true" data-render={renderKey} data-reaction={reaction} data-animate={animate}>
  <div class="plate chest"></div>
  <div class="plate crest"></div>
  <div class="ears-back"></div>
  <div class="flick"></div>
</div>

<style>
  .overlay { position: absolute; inset: 0; pointer-events: none; }
  .plate { position: absolute; background: var(--cc-color-surface); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  [data-render="horse_armor"] .chest { inset-block-end: 5%; inset-inline-start: 18%; inline-size: 54%; block-size: 35%; border-radius: var(--cc-space-sm); }
  [data-render="horse_armor"] .crest { inset-block-end: 55%; inset-inline-end: 8%; inline-size: 30%; block-size: 20%; border-radius: var(--cc-space-xs); }
  .ears-back, .flick { position: absolute; display: none; }
  [data-reaction="annoyed"] .ears-back { display: block; inset-block-start: 5%; inset-inline-end: 12%; inline-size: 26%; block-size: 8%; background: var(--cc-color-border); }
  [data-reaction="annoyed"] .flick { display: block; inset-block-end: 22%; inset-inline-start: 0; inline-size: 18%; block-size: 8%; background: var(--cc-color-border); border-radius: var(--cc-space-xs); }
  [data-reaction="annoyed"][data-animate="true"] .flick { animation-name: flick; animation-duration: var(--cc-motion-duration_base); animation-timing-function: var(--cc-motion-easing); animation-iteration-count: 2; animation-direction: alternate; }
  @keyframes flick { to { inset-block-end: 32%; } }
  @media (prefers-reduced-motion: reduce) { .flick { animation-name: none; } }
</style>
