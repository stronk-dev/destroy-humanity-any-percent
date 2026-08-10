<script lang="ts">
  import { t } from "../copy";
  import Amount from "./Amount.svelte";
  import type { AmountCap } from "./amount-types";
  import { UI_THEMES, installTheme, type UIEra } from "./themes";

  let root: HTMLElement;
  let activeEra = $state<UIEra>("era_1995");
  let reducedMotion = $state(false);
  let amount = $state("1.2345e3");
  let cap = $state<AmountCap | undefined>();

  $effect(() => {
    if (root) installTheme(root, UI_THEMES[activeEra], reducedMotion);
  });

  export function switchEra(): void { activeEra = activeEra === "era_1995" ? "era_2000" : "era_1995"; }
  export function setAmount(value: string): void { amount = value; }
  export function setCap(value: AmountCap | undefined): void { cap = value; }
  export function setReducedMotion(value: boolean): void { reducedMotion = value; }
</script>

<main bind:this={root} class="cc-fixture">
  <header class="cc-window">
    <div class="cc-titlebar">{t("category.valuation", {}, activeEra)}</div>
    <button type="button" onclick={switchEra}>{t("category.any_percent", {}, activeEra)}</button>
  </header>
  <section class="cc-panel" aria-labelledby="ui-fixture-heading">
    <h1 id="ui-fixture-heading">{t("achievement.generators_purchased_1", {}, activeEra)}</h1>
    <Amount value={amount} {cap} era={activeEra} />
    <a href="#ui-fixture-heading">{t("category.hundred_percent", {}, activeEra)}</a>
  </section>
</main>

<style>
  .cc-fixture { min-height: 100vh; padding: var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-color-bg); font-family: var(--cc-type-font_ui); font-size: var(--cc-type-size_base); line-height: var(--cc-type-line_height); }
  .cc-window { max-width: 48rem; margin: auto; border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  .cc-titlebar { padding: var(--cc-space-sm) var(--cc-space-md); color: var(--cc-chrome-titlebar_text); background: var(--cc-chrome-titlebar_bg); font-family: var(--cc-type-font_display); font-size: var(--cc-type-size_large); font-weight: var(--cc-type-weight_bold); }
  button { margin: var(--cc-space-md); padding: var(--cc-space-sm) var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-chrome-button_face); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); font-family: var(--cc-type-font_ui); font-size: var(--cc-type-size_base); cursor: pointer; }
  button:focus-visible, a:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  .cc-panel { display: grid; gap: var(--cc-space-md); padding: var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-color-surface); }
  h1 { margin: 0; font-family: var(--cc-type-font_display); font-size: var(--cc-type-size_display); }
  a { color: var(--cc-color-link); }
</style>
