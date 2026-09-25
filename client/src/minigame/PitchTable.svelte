<script lang="ts">
  import type { PitchSnapshot } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import Amount from "../ui/Amount.svelte";
  import { pitchCardInstanceBase, type PitchContent } from "./pitch-content";

  // Tenant child for (pitch, 1.0.0). Presentation only: it renders the server
  // snapshot and raises callbacks; the host surface owns transport, errors and
  // lifecycle (MA3/MA-C9). No client simulation.
  let { snapshot, content, era, pending, selectionEpoch, onPlayHand, onBuyHack, onEndShop }: {
    snapshot: PitchSnapshot;
    content: PitchContent;
    era: CopyEra;
    pending: boolean;
    selectionEpoch: number;
    onPlayHand(cardIDs: readonly string[]): void;
    onBuyHack(offerID: string): void;
    onEndShop(): void;
  } = $props();

  let selected = $state<readonly string[]>([]);
  // A new snapshot revision or a host-requested clear drops the selection.
  $effect(() => { void snapshot.revision; void selectionEpoch; selected = []; });

  const playSize = $derived(content.catalog.policy.play_size);
  const full = $derived(selected.length >= playSize);

  function card(instance: string) {
    const base = pitchCardInstanceBase(instance);
    const row = content.catalog.metric_cards.find((candidate) => candidate.card_id === base);
    if (!row) throw new RangeError(`Pitch card ${base} is not in the pinned catalog`);
    return row;
  }
  function hackKey(hackID: string): CopyKey {
    const row = content.catalog.growth_hacks.find((candidate) => candidate.hack_id === hackID);
    if (!row) throw new RangeError(`Pitch hack ${hackID} is not in the pinned catalog`);
    return row.copy_key as CopyKey;
  }
  function toggle(instance: string, checked: boolean): void {
    if (checked && !selected.includes(instance) && !full) selected = [...selected, instance];
    else if (!checked) selected = selected.filter((value) => value !== instance);
  }
  function play(): void {
    if (pending || selected.length === 0) return;
    // The wire requires unique, byte-sorted instance IDs.
    onPlayHand([...selected].sort((left, right) => left < right ? -1 : left > right ? 1 : 0));
  }
</script>

<section class="pitch" aria-labelledby="pitch-status-heading">
  <h2 id="pitch-status-heading">{t("pitch.table.round_frame", { round: snapshot.round }, era)}</h2>
  <dl class="facts">
    <div><dt>{t("pitch.table.target_label", {}, era)}</dt><dd><Amount value={snapshot.funding_target} {era} /></dd></div>
    <div><dt>{t("pitch.table.best_label", {}, era)}</dt><dd><Amount value={snapshot.round_best_valuation} {era} /></dd></div>
    <div><dt>{t("pitch.table.hands_frame", { count: snapshot.hands_remaining }, era)}</dt></div>
    <div><dt>{t("pitch.table.currency_frame", { count: snapshot.run_currency }, era)}</dt></div>
    <div><dt>{t("pitch.table.deck_frame", { count: snapshot.deck_count }, era)}</dt></div>
  </dl>

  {#if snapshot.slotted_hacks.length}
    <section aria-labelledby="pitch-slotted-heading">
      <h3 id="pitch-slotted-heading">{t("pitch.table.slotted_heading", {}, era)}</h3>
      <ul>{#each snapshot.slotted_hacks as hackID (hackID)}<li>{t(hackKey(hackID), {}, era)}</li>{/each}</ul>
    </section>
  {/if}

  {#if snapshot.phase === "playing"}
    <fieldset class="hand" disabled={pending}>
      <legend>{t("pitch.table.hand_legend", { max: playSize }, era)}</legend>
      {#each snapshot.hand as instance (instance)}
        {@const row = card(instance)}
        {@const checked = selected.includes(instance)}
        <label class="card" data-selected={checked}>
          <input type="checkbox" value={instance} {checked} aria-disabled={!checked && full ? "true" : undefined}
            onchange={(event) => { if (!checked && full) { event.currentTarget.checked = false; return; } toggle(instance, event.currentTarget.checked); }} />
          <span>{t(row.copy_key as CopyKey, {}, era)}</span>
          <small>{t("pitch.table.card_metric_label", {}, era)} <Amount value={row.base_metric} {era} /></small>
        </label>
      {/each}
      {#if full}<p class="cap" role="status">{t(content.catalog.policy.play_size_reason_key as CopyKey, {}, era)}</p>{/if}
    </fieldset>
    <button type="button" disabled={pending || selected.length === 0} onclick={play}>{t("pitch.table.play", {}, era)}</button>
  {:else if snapshot.phase === "shop"}
    <section aria-labelledby="pitch-shop-heading">
      <h3 id="pitch-shop-heading">{t("pitch.table.shop_heading", {}, era)}</h3>
      <ul class="offers">
        {#each snapshot.shop_offers as offer (offer.offer_id)}
          <li><span>{t(hackKey(offer.hack_id), {}, era)}</span>
            <button type="button" disabled={pending || offer.price > snapshot.run_currency} onclick={() => onBuyHack(offer.offer_id)}>{t("pitch.table.offer_frame", { price: offer.price }, era)}</button></li>
        {/each}
      </ul>
      <button type="button" disabled={pending} onclick={onEndShop}>{t("pitch.table.end_shop", {}, era)}</button>
    </section>
  {:else}
    <p role="status">{t("pitch.table.terminal", {}, era)}</p>
  {/if}
</section>

<style>
  .pitch { display: grid; gap: var(--cc-space-md); }
  .facts { display: grid; grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr)); gap: var(--cc-space-sm); margin: 0; }
  .facts div { display: grid; gap: var(--cc-space-xs); }
  dd { margin: 0; }
  .hand { display: grid; grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr)); gap: var(--cc-space-sm); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); padding: var(--cc-space-md); }
  .card { display: grid; gap: var(--cc-space-xs); padding: var(--cc-space-sm); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-color-surface); cursor: pointer; }
  .card[data-selected="true"] { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
  .card input:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  .cap { grid-column: 1 / -1; margin: 0; }
  .offers { display: grid; gap: var(--cc-space-sm); padding: 0; list-style: none; }
  .offers li { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); align-items: center; justify-content: space-between; }
  ul { margin: 0; }
</style>
