<script lang="ts">
  import type { GameUIAxisStackArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import { formatAxisFactor } from "./axis-factor-format";

  // Clout v1 CV9: the one axis-stack readout (product, formula, visible cap)
  // Upgrade cards own each PR Intern's progress and current factor, adjacent
  // to its purchase control; this panel owns only the shared stack readout.
  // Every number is server-derived; nothing is computed here.
  let { arm, era }: { arm: GameUIAxisStackArm; era: CopyEra } = $props();
</script>

<section class="axis card" aria-labelledby="axis-heading">
  <h2 id="axis-heading">{t("axis_stack.panel.title", {}, era)}</h2>
  <details class="axis-help">
    <summary tabindex="0" aria-label={t("axis_stack.help.label", {}, era)}>{t("axis_stack.help.symbol", {}, era)}</summary>
    <p class="codex">{t("codex.axis_stack", {}, era)}</p>
  </details>
  <p>{t("axis_stack.formula.caption", {}, era)}</p>
  <p id="axis-why"><small>{t("axis_stack.tooltip.why", {}, era)}</small></p>
  <p>{t("axis_stack.input_frame", { value: Math.min(arm.input_value, arm.input_cap), cap: arm.input_cap }, era)}</p>
  {#if arm.saturated}<p role="status">{t("axis_stack.saturated", {}, era)} <small>{t(arm.cap_reason_key as CopyKey, {}, era)}</small></p>{/if}
  <!-- Factors are not resource amounts: Amount intentionally rounds values
       below 1000 to integers, which would conceal the effect of a PR Intern.
       Display the supplied factor without computing or rounding its value. -->
  <p>{t("axis_stack.product_label", {}, era)} <output>{formatAxisFactor(arm.product)}</output></p>
</section>

<style>
  .axis { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  h2, p { margin: 0; }
  h2 { font-family: var(--cc-type-font_display); }
  summary { display: grid; place-items: center; inline-size: max-content; min-inline-size: 44px; min-block-size: 44px; cursor: pointer; }
  summary:focus-visible { outline: 2px solid var(--cc-color-accent); outline-offset: 2px; }
  .codex { overflow-wrap: anywhere; }
</style>
