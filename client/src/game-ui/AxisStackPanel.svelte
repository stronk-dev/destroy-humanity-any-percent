<script lang="ts">
  import type { GameUIAxisStackArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import Amount from "../ui/Amount.svelte";
  import { upgradePresentation } from "./axis-presentation";

  // Clout v1 CV9: the one axis-stack readout (product, formula, visible cap)
  // and each PR Intern's attainment progress and factor at the current input.
  // Every number is server-derived; nothing is computed here.
  let { arm, era }: { arm: GameUIAxisStackArm; era: CopyEra } = $props();
</script>

<section class="axis card" aria-labelledby="axis-heading">
  <h2 id="axis-heading">{t("axis_stack.panel.title", {}, era)}</h2>
  <p>{t("axis_stack.formula.caption", {}, era)}</p>
  <p id="axis-why"><small>{t("axis_stack.tooltip.why", {}, era)}</small></p>
  <p>{t("axis_stack.input_frame", { value: Math.min(arm.input_value, arm.input_cap), cap: arm.input_cap }, era)}</p>
  {#if arm.saturated}<p role="status">{t("axis_stack.saturated", {}, era)} <small>{t(arm.cap_reason_key as CopyKey, {}, era)}</small></p>{/if}
  <p>{t("axis_stack.product_label", {}, era)} <Amount value={arm.product} {era} /></p>
  <ul class="interns">
    {#each arm.interns as intern (intern.upgrade_id)}
      <li data-upgrade={intern.upgrade_id}>
        <strong>{t(upgradePresentation(intern.upgrade_id).title_key, {}, era)}</strong>
        {#if !intern.owned}<span>{t("axis_stack.progress_frame", { minimum: intern.minimum, value: Math.min(arm.input_value, arm.input_cap) }, era)}</span>
          <progress max={intern.minimum} value={Math.min(arm.input_value, intern.minimum)} aria-describedby="axis-why"></progress>{/if}
        <span>{t("axis_stack.factor_label", {}, era)} <Amount value={intern.factor} {era} /></span>
        {#if intern.owned}<span>{t("desk.upgrade.owned", {}, era)}</span>{/if}
      </li>
    {/each}
  </ul>
</section>

<style>
  .axis { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  h2, p { margin: 0; }
  h2 { font-family: var(--cc-type-font_display); }
  .interns { display: grid; gap: var(--cc-space-sm); margin: 0; padding: 0; list-style: none; }
  .interns li { display: grid; gap: var(--cc-space-xs); }
  progress { inline-size: 100%; accent-color: var(--cc-color-accent); }
</style>
