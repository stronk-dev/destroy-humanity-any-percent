<script lang="ts">
  import { t, type CopyEra } from "../copy";
  import type { RunEndSurfaceProps } from "./events";
  import { GAME_UI_PRESENTATION, requirePresentation } from "./presentation";
  import { renderPrestigeTermRows } from "./prestige-terms";

  let { ended }: RunEndSurfaceProps = $props();
  const era = $derived<CopyEra>(ended.payload.tier === 0 ? "era_1995" : ended.payload.tier === 1 ? "era_2000" : (() => { throw new RangeError(`tier ${ended.payload.tier} has no shipped UI era`); })());

  function duration(ms: number): string {
    const seconds = Math.max(0, Math.floor(ms / 1000));
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor(seconds % 3600 / 60);
    return `${hours}:${minutes.toString().padStart(2, "0")}:${(seconds % 60).toString().padStart(2, "0")}`;
  }

  function exitTitle(exitType: string): string {
    return t(requirePresentation(GAME_UI_PRESENTATION.exitTypes, exitType).title_key, {}, era);
  }
</script>

<section class="surface" aria-labelledby="run-end-heading">
  <h1 id="run-end-heading">{ended.payload.exit_type === "scripted_first" ? t("curriculum.scripted_first_failure.title", {}, era) : t("screen.run_end.standard.title", {}, era)}</h1>
  <p>{t("screen.run_end.exit_frame", { exit_type: exitTitle(ended.payload.exit_type), tier: ended.payload.tier }, era)}</p>
  <p>{t("screen.run_end.attended_frame", { attended: duration(ended.payload.attended_ms) }, era)}</p>
  {#if ended.payload.exit_type === "scripted_first"}
    <p>{t("curriculum.scripted_first_failure.body", {}, era)}</p>
    <p>{t("curriculum.scripted_first_failure.next_run", {}, era)}</p>
  {:else}
    <p>{t("screen.run_end.founder_note", {}, era)}</p>
  {/if}
  <h2>{t("screen.run_end.delta_heading", { run_seq: ended.payload.run_id.run_seq + 1 }, era)}</h2>
  {#each renderPrestigeTermRows(ended.payload.payout, era) as row}<p>{row}</p>{/each}
</section>

<style>
  .surface { display: grid; gap: var(--cc-space-lg); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-chrome-window_bg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); }
  h1, h2, p { margin: 0; }
  h1, h2 { font-family: var(--cc-type-font_display); }
</style>
