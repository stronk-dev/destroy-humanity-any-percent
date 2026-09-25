<script lang="ts">
  import type { GameUIFiscalArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import { FEATURES_PRESENTATION } from "./features-presentation";
  import { fiscalPhase } from "./fiscal-phase";
  import { GAME_UI_PRESENTATION, requirePresentation } from "./presentation";

  // GS1. The harvest phase is display-only, derived from the host's server-time
  // estimate against the pinned windows; the receipt decides (a stale early
  // estimate produces an honest rejection).
  let { arm, era, serverNowMs, pending, controlsEnabled, onHarvest, onSpendLevel, onSpendUnlock }: {
    arm: GameUIFiscalArm;
    era: CopyEra;
    serverNowMs: number;
    pending: boolean;
    controlsEnabled: boolean;
    onHarvest(): void;
    onSpendLevel(generatorID: string): void;
    onSpendUnlock(unlockID: string): void;
  } = $props();

  const elapsed = $derived(Math.max(0, serverNowMs - arm.period.opened_wall_ms));
  const phase = $derived(fiscalPhase(elapsed, arm.period));
  const spendable = $derived(arm.sweep_preview.credit_after);
  const unlocks = $derived(arm.unlocks.filter((row) => FEATURES_PRESENTATION.fiscalUnlocks.has(row.unlock_id)));

  function remaining(ms: number): string {
    const seconds = Math.max(0, Math.ceil(ms / 1000));
    const hours = Math.floor(seconds / 3600), minutes = Math.floor(seconds % 3600 / 60);
    return `${hours}:${minutes.toString().padStart(2, "0")}:${(seconds % 60).toString().padStart(2, "0")}`;
  }
  function percent(ppm: number): string {
    const hundredths = Math.floor(ppm / 100);
    return `${Math.floor(hundredths / 100)}.${(hundredths % 100).toString().padStart(2, "0")}`;
  }
  function generatorTitle(id: string): string { return t(requirePresentation(GAME_UI_PRESENTATION.generators, id).title_key, {}, era); }
</script>

<section class="surface fiscal" aria-labelledby="fiscal-heading" data-phase={phase}>
  <h1 id="fiscal-heading" tabindex="-1">{t("surface.fiscal.title", {}, era)}</h1>

  <section class="card" aria-labelledby="fiscal-credit-heading">
    <h2 id="fiscal-credit-heading">{t("fiscal.credit_label", {}, era)}</h2>
    <p>{t("fiscal.credit_frame", { credit: arm.credit, cap: arm.credit_cap.amount }, era)}</p>
    {#if arm.credit >= arm.credit_cap.amount || arm.sweep_preview.saturated}<p>{t(arm.credit_cap.reason_key as CopyKey, {}, era)}</p>{/if}
    {#if arm.sweep_preview.periods > 0}<p>{t("fiscal.sweep_preview_frame", { periods: arm.sweep_preview.periods, credited: arm.sweep_preview.credited }, era)}</p>{/if}
    <p>{t("fiscal.hoard_frame", { percent: percent(arm.hoard.preview_ppm), cap: arm.hoard.cap_credits }, era)}</p>
    <p>{t("fiscal.next_run_note", {}, era)}</p>
  </section>

  <section class="card" aria-labelledby="fiscal-harvest-heading">
    <h2 id="fiscal-harvest-heading">{t("fiscal.harvest", {}, era)}</h2>
    {#if phase === "ripening"}
      <p class="phase">{t("fiscal.period.ripening_frame", { remaining: remaining(arm.period.early_ms - elapsed) }, era)}</p>
    {:else if phase === "early"}
      <p class="phase">{t("fiscal.period.early_frame", { success_percent: Math.floor(arm.period.early_success_ppm / 10_000) }, era)}</p>
    {:else}
      <p class="phase">{t("fiscal.period.guaranteed", {}, era)}</p>
    {/if}
    <p>{t("fiscal.period.auto_note", { remaining: remaining(arm.period.auto_ms - elapsed) }, era)}</p>
    <button type="button" aria-describedby="fiscal-harvest-curtain" disabled={pending || !controlsEnabled || phase === "ripening"} onclick={onHarvest}>{t("fiscal.harvest", {}, era)}</button>
    <small id="fiscal-harvest-curtain">{t("fiscal.harvest_tooltip", {}, era)}</small>
  </section>

  <section class="card" aria-labelledby="fiscal-levels-heading">
    <h2 id="fiscal-levels-heading">{t("fiscal.levels_label", {}, era)}</h2>
    <ul>
      {#each arm.generator_levels as row (row.generator_id)}
        <li>
          <h3>{generatorTitle(row.generator_id)}</h3>
          <span>{t("fiscal.level_frame", { level: row.level, cap: row.level_cap.amount }, era)}</span>
          {#if row.next_level_cost === null}
            <span>{t(row.level_cap.reason_key as CopyKey, {}, era)}</span>
          {:else}
            <button type="button" disabled={pending || !controlsEnabled || row.next_level_cost > spendable} onclick={() => onSpendLevel(row.generator_id)}>{t("fiscal.level_buy", { cost: row.next_level_cost }, era)}</button>
          {/if}
        </li>
      {/each}
    </ul>
  </section>

  {#if unlocks.length}
    <section class="card" aria-labelledby="fiscal-unlocks-heading">
      <h2 id="fiscal-unlocks-heading">{t("fiscal.unlocks_label", {}, era)}</h2>
      <ul>
        {#each unlocks as row (row.unlock_id)}
          {@const presentation = FEATURES_PRESENTATION.fiscalUnlocks.get(row.unlock_id)!}
          <li>
            <h3>{t(presentation.title_key, {}, era)}</h3>
            <p>{t(presentation.description_key, {}, era)}</p>
            {#if row.owned}
              <strong>{t("fiscal.unlock_owned", {}, era)}</strong>
            {:else}
              <button type="button" disabled={pending || !controlsEnabled || row.cost > spendable} onclick={() => onSpendUnlock(row.unlock_id)}>{t("fiscal.unlock_buy", { cost: row.cost }, era)}</button>
            {/if}
          </li>
        {/each}
      </ul>
    </section>
  {/if}
</section>

<style>
  .fiscal { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(18rem, 100%), 1fr)); gap: var(--cc-space-md); }
  .fiscal > h1 { grid-column: 1 / -1; }
  .card { display: grid; gap: var(--cc-space-sm); align-content: start; }
  ul { display: grid; gap: var(--cc-space-sm); margin: 0; padding: 0; list-style: none; }
  li { display: grid; gap: var(--cc-space-xs); }
  h1, h2, h3, p { margin: 0; }
  h3 { font-size: inherit; }
  button { justify-self: start; }
</style>
