<script lang="ts">
  import type { GameUIMetersArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import { FEATURES_PRESENTATION } from "./features-presentation";

  // GS3: read-only. Every cell carries numeric and band text beside its
  // native <meter>; values are the committed ones, never extrapolated.
  let { arm, era }: { arm: GameUIMetersArm; era: CopyEra } = $props();

  type Row = GameUIMetersArm["meters"][number];
  const byID = $derived(new Map(arm.meters.map((row) => [row.meter_id, row])));
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

{#snippet cell(row: Row | undefined, label: string)}
  {#if row}
    <td data-label={label}>
      <meter min={row.min} max={row.max} value={row.value}></meter>
      <span>{t("meters.value_frame", { value: row.value, max: row.max }, era)}</span>
      <span>{band(row)}</span>
    </td>
  {:else}
    <td data-label={label}></td>
  {/if}
{/snippet}

<section class="surface meters" aria-labelledby="meters-heading">
  <h1 id="meters-heading" tabindex="-1">{t("surface.meters.title", {}, era)}</h1>
  <p>{t("meters.curtain", {}, era)}</p>
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
  {#if doom}
    <section class="doom" aria-labelledby="doom-heading">
      <h2 id="doom-heading" title={t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}>{t(FEATURES_PRESENTATION.doomMeter.title_key, {}, era)}</h2>
      <p>{t(FEATURES_PRESENTATION.doomMeter.tooltip_key, {}, era)}</p>
      <meter min={doom.min} max={doom.max} value={doom.value}></meter>
      <span>{t("meters.value_frame", { value: doom.value, max: doom.max }, era)}</span>
      <span>{band(doom)}</span>
    </section>
  {/if}
  <small>{t("meters.as_of_note", {}, era)}</small>
</section>

<style>
  .meters { display: grid; gap: var(--cc-space-md); }
  table { inline-size: 100%; border-collapse: collapse; }
  th, td { padding: var(--cc-space-xs) var(--cc-space-sm); text-align: start; border-block-end: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  td { display: grid; gap: var(--cc-space-xs); }
  meter { inline-size: 100%; accent-color: var(--cc-color-accent); }
  .doom { display: grid; gap: var(--cc-space-xs); }
  h1, h2, p { margin: 0; }
  @media (max-width: 30rem) {
    thead { position: absolute; inline-size: 1px; block-size: 1px; overflow: hidden; clip-path: inset(50%); }
    table, tbody, tr, th, td { display: block; }
    td::before { content: attr(data-label); font-weight: var(--cc-type-weight_bold); }
  }
</style>
