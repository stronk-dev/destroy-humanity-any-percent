<script lang="ts">
  import type { GameUIOpportunityArm } from "../api/generated/types";
  import { applicationCopyCatalog, t, type CopyEra, type CopyKey } from "../copy";
  import Amount from "../ui/Amount.svelte";
  import { FEATURES_PRESENTATION } from "./features-presentation";
  import type { LastClaim } from "./opportunity-claim";

  // GS5: an always-present Desk region (stable DOM order; focus is never moved
  // here). Remaining time is attended seconds as of the last snapshot, not a
  // wall-clock countdown (OD-12). The region never causes a spawn: it only
  // renders the projected arm and raises the claim callback.
  let { arm, era, pending, controlsEnabled, lastClaim, onClaim }: {
    arm: GameUIOpportunityArm | null;
    era: CopyEra;
    pending: boolean;
    controlsEnabled: boolean;
    lastClaim: LastClaim | undefined;
    onClaim(opportunityID: string): void;
  } = $props();

  function seconds(expires: number): number { return Math.max(0, Math.ceil((expires - (arm?.attended_now_ms ?? 0)) / 1000)); }
  function effectRow(id: string) { return FEATURES_PRESENTATION.opportunityEffects.get(id); }
  function capText(key: string | null): string | null { return key !== null && applicationCopyCatalog.byKey.has(key) ? t(key as CopyKey, {}, era) : null; }
  const offer = $derived(arm?.pending && effectRow(arm.pending.effect_row_id) ? arm.pending : null);
  $effect(() => { if (arm?.pending && !effectRow(arm.pending.effect_row_id)) console.error(`game UI invariant: unknown opportunity effect ${arm.pending.effect_row_id}`); });
</script>

<section class="opportunity cc-window" data-region="desk.region.opportunity" aria-label={t("desk.opportunity.region_label", {}, era)}>
  {#if offer}
    {@const row = effectRow(offer.effect_row_id)!}
    <h3>{t(row.title_key, {}, era)}</h3>
    <p>{t(row.description_key, {}, era)}</p>
    <p>{t("desk.opportunity.remaining_frame", { seconds: seconds(offer.expires_attended_ms) }, era)} <small>{t("desk.opportunity.attended_note", {}, era)}</small></p>
    {#if offer.effect_row_id === "active.lucky"}<small>{t("desk.opportunity.lucky_tooltip", {}, era)}</small>{/if}
    <button type="button" disabled={pending || !controlsEnabled} onclick={() => onClaim(offer.opportunity_id)}>{t("desk.opportunity.claim", {}, era)}</button>
  {/if}
  {#if lastClaim && lastClaim.credited !== null}
    <p>{t("desk.opportunity.lucky_frame", { amount: lastClaim.credited }, era)} <Amount value={lastClaim.credited} {era} /></p>
    {#if lastClaim.saturated}<p>{t("desk.opportunity.lucky_capped", {}, era)} {capText(lastClaim.capReasonKey) ?? ""}</p>{/if}
  {/if}
  {#if arm && arm.buffs.length}
    <h3>{t("desk.buffs_label", {}, era)}</h3>
    <ul class="buffs">
      {#each arm.buffs as buff (buff.buff_instance_id)}
        {@const row = effectRow(buff.effect_row_id)}
        {#if row}<li><span>{t(row.title_key, {}, era)}</span> <span>{t("desk.buff.remaining_frame", { seconds: seconds(buff.expires_attended_ms) }, era)}</span></li>{/if}
      {/each}
    </ul>
    {#if arm.combo.saturated}<p>{t("desk.buff.combo_capped", {}, era)} {capText(arm.combo.reason_key) ?? ""}</p>{/if}
  {/if}
</section>

<style>
  .opportunity { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-md); }
  h3, p { margin: 0; }
  .buffs { display: grid; gap: var(--cc-space-xs); margin: 0; padding-inline-start: var(--cc-space-lg); }
  button { justify-self: start; }
</style>
