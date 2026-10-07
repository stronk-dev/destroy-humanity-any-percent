<script lang="ts">
  import { onMount } from "svelte";
  import type { GameUIMetersArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import { FEATURES_PRESENTATION } from "./features-presentation";

  // GS3: read-only. Every cell carries numeric and band text beside its
  // native <meter>; values are the committed ones, never extrapolated.
  let { arm, era }: { arm: GameUIMetersArm; era: CopyEra } = $props();

  type Row = GameUIMetersArm["meters"][number];
  let narrow = $state(false);
  onMount(() => {
    const media = window.matchMedia("(width < 30rem)");
    const update = () => { narrow = media.matches; };
    update();
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  });
  const presentationUnavailable = $derived(arm.meters.some((row) => !FEATURES_PRESENTATION.meterBands.has(row.band_id)));
  let presentationErrorReported = false;
  $effect(() => {
    if (!presentationUnavailable) { presentationErrorReported = false; return; }
    if (!presentationErrorReported) {
      console.error("game UI invariant: meters surface presentation unavailable");
      presentationErrorReported = true;
    }
  });
  const byID = $derived(new Map(arm.meters.map((row) => [row.meter_id, row])));
  const trustRows = $derived([...FEATURES_PRESENTATION.trustMeters.values()].flatMap((presentation) => {
    const row = byID.get(presentation.meter_id);
    return row ? [{ row, presentation }] : [];
  }));
  const constituencies = $derived.by(() => {
    const groups = new Map<CopyKey, { standing?: Row; grievance?: Row }>();
    for (const [meterID, presentation] of FEATURES_PRESENTATION.trustMeters) {
      const row = byID.get(meterID);
      if (!row) continue;
      const group = groups.get(presentation.constituency_key) ?? {};
      if (presentation.axis_key === "meters.axis.standing") group.standing = row; else group.grievance = row;
      groups.set(presentation.constituency_key, group);
    }
    return [...groups.entries()];
  });
  const doom = $derived(byID.get(FEATURES_PRESENTATION.doomMeter.meter_id));
  function band(row: Row): string {
    const key = FEATURES_PRESENTATION.meterBands.get(row.band_id);
    if (!key) throw new RangeError(`meter band ${row.band_id} has no presentation`);
    return t(key, {}, era);
  }
</script>

{#snippet value(row: Row, labelledBy?: string)}
  <meter min={row.min} max={row.max} value={row.value} aria-labelledby={labelledBy}></meter>
  <span>{t("meters.value_frame", { value: row.value, max: row.max }, era)}</span>
  <span>{band(row)}</span>
{/snippet}

{#snippet cell(row: Row | undefined, label: string)}
  {#if row}
    <td data-label={label}>
      {@render value(row)}
    </td>
  {:else}
    <td data-label={label}></td>
  {/if}
{/snippet}

<section class="surface meters" aria-labelledby="meters-heading">
  <h1 id="meters-heading" tabindex="-1">{t("surface.meters.title", {}, era)}</h1>
  {#if presentationUnavailable}
    <p role="alert">{t("common.surface_error", {}, era)}</p>
  {:else}
  <p>{t("meters.curtain", {}, era)}</p>
  {#if narrow}
    <dl class="trust-list">
      {#each trustRows as { row, presentation } (row.meter_id)}
        <div data-meter-id={row.meter_id}>
          <dt id={`meter-term-${row.meter_id}`}>{t(presentation.constituency_key, {}, era)} {t(presentation.axis_key, {}, era)}</dt>
          <dd>{@render value(row, `meter-term-${row.meter_id}`)}</dd>
        </div>
      {/each}
    </dl>
  {:else}
  <table>
    <thead>
      <tr>
        <th scope="col">{t("meters.constituency_label", {}, era)}</th>
        <th scope="col">{t("meters.axis.standing", {}, era)}</th>
        <th scope="col">{t("meters.axis.grievance", {}, era)}</th>
      </tr>
    </thead>
    <tbody>
      {#each constituencies as [constituency, group] (constituency)}
        <tr>
          <th scope="row">{t(constituency, {}, era)}</th>
          {@render cell(group.standing, t("meters.axis.standing", {}, era))}
          {@render cell(group.grievance, t("meters.axis.grievance", {}, era))}
        </tr>
      {/each}
    </tbody>
  </table>
  {/if}
  {#if doom}
    <section class="doom" aria-labelledby={narrow ? `meter-term-${doom.meter_id}` : "doom-heading"}>
      {#if narrow}
        <dl>
          <div data-meter-id={doom.meter_id}>
            <dt id={`meter-term-${doom.meter_id}`} title={t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}>{t(FEATURES_PRESENTATION.doomMeter.title_key, {}, era)}</dt>
            <dd>
              <p>{t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}</p>
              {@render value(doom, `meter-term-${doom.meter_id}`)}
            </dd>
          </div>
        </dl>
      {:else}
      <h2 id="doom-heading" title={t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}>{t(FEATURES_PRESENTATION.doomMeter.title_key, {}, era)}</h2>
      <p>{t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}</p>
      {@render value(doom)}
      {/if}
    </section>
  {/if}
  <small>{t("meters.as_of_note", {}, era)}</small>
  {/if}
</section>

<style>
  .meters { display: grid; gap: var(--cc-space-md); }
  table { inline-size: 100%; border-collapse: collapse; }
  th, td { padding: var(--cc-space-xs) var(--cc-space-sm); text-align: start; border-block-end: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  td { display: grid; gap: var(--cc-space-xs); }
  meter { inline-size: 100%; accent-color: var(--cc-color-accent); }
  .doom { display: grid; gap: var(--cc-space-xs); }
  h1, h2, p { margin: 0; }
  dl { margin: 0; }
  dl > div, dd { display: grid; gap: var(--cc-space-xs); }
  dl > div { padding: var(--cc-space-xs) var(--cc-space-sm); border-block-end: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  dt { font-weight: var(--cc-type-weight_bold); }
  dd { margin: 0; }
</style>
