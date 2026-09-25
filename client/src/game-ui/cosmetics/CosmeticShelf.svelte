<script lang="ts">
  import { tick } from "svelte";

  import type { GameUICosmeticsArm } from "../../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../../copy";
  import type { CosmeticShopPresentation } from "./presentation";

  // Cosmetic Shop v1 §7.3: the $0.00 shelf. Every bound curtain renders as
  // persistent small print in every item state and is the controls'
  // aria-describedby (§8, OD-9). There is no cart, confirm dialog, quantity,
  // timer, countdown, badge, urgency toast, or upsell. Ownership renders only
  // from the authoritative snapshot, never before an applied receipt (I3).
  let { arm, presentation, price, era, pending, controlsEnabled, petName, receipt, rejection, onAcquire, onEquip, onUnequip }: {
    arm: GameUICosmeticsArm;
    presentation: CosmeticShopPresentation;
    price: string;
    era: CopyEra;
    pending: boolean;
    controlsEnabled: boolean;
    petName(petID: string): string;
    receipt: Readonly<{ cosmeticId: string; orderNumber: number }> | null;
    rejection: CopyKey | null;
    onAcquire(cosmeticID: string): void;
    onEquip(cosmeticID: string, petID: string): void;
    onUnequip(petID: string): void;
  } = $props();

  const rows = $derived(arm.items.flatMap((item) => {
    const row = presentation.cosmetics.get(item.cosmetic_id);
    // A catalog id without a presentation row is withheld, never rendered mechanically.
    return row ? [{ item, row }] : [];
  }));
  const ownedHeadings = new Map<string, HTMLElement>();
  let focusAfterReceipt: string | undefined;

  function curtainID(id: string, pattern: string): string { return `curtain-${id}-${pattern}`; }
  function describedBy(id: string): string {
    const row = presentation.cosmetics.get(id)!;
    return [...row.curtains.map((curtain) => curtainID(id, curtain.pattern)), curtainID("shop", "checkout_flow")].join(" ");
  }
  function buy(id: string): void { focusAfterReceipt = id; onAcquire(id); }
  // Keyboard focus follows the replaced Buy button to the owned-state heading.
  $effect(() => {
    for (const { item } of rows) {
      if (item.owned && focusAfterReceipt === item.cosmetic_id) {
        focusAfterReceipt = undefined;
        void tick().then(() => ownedHeadings.get(item.cosmetic_id)?.focus());
      }
    }
  });
  function register(node: HTMLElement, id: string) { ownedHeadings.set(id, node); return { destroy: () => ownedHeadings.delete(id) }; }
</script>

<section class="shelf cc-window" aria-labelledby="shelf-heading" data-testid="cosmetic-shelf">
  <h2 id="shelf-heading">{t(presentation.shop.heading_key, {}, era)}</h2>
  <small id={curtainID("shop", "checkout_flow")} class="curtain">{t(presentation.shop.curtains[0]!.key, {}, era)}</small>
  {#each rows as { item, row } (item.cosmetic_id)}
    <article class="item" data-cosmetic={item.cosmetic_id} data-state={item.owned ? "owned" : pending ? "pending" : "unowned"}>
      <h3>{t(row.title_key, {}, era)}</h3>
      <p>{t(row.description_key as "cosmetic.horse_armor_free.description", { price }, era)}</p>
      {#if row.anchor_key}<p class="anchor" aria-describedby={curtainID(item.cosmetic_id, "reference_price_anchor")}><s>{t(row.anchor_key, {}, era)}</s></p>{/if}
      {#if item.owned}
        <p class="owned" tabindex="-1" use:register={item.cosmetic_id}>{t("shop.cosmetics.owned", {}, era)}</p>
        {#if arm.wearers.length === 0}
          <p>{t("shop.cosmetics.no_wearer", {}, era)}</p>
        {:else}
          <ul class="wearers">
            {#each arm.wearers as wearer (wearer.pet_id)}
              <li>
                {#if wearer.worn === item.cosmetic_id}
                  <button type="button" disabled={pending || !controlsEnabled} onclick={() => onUnequip(wearer.pet_id)}>{t("shop.cosmetics.unequip", { pet: petName(wearer.pet_id) }, era)}</button>
                {:else}
                  <button type="button" disabled={pending || !controlsEnabled} onclick={() => onEquip(item.cosmetic_id, wearer.pet_id)}>{t("shop.cosmetics.equip", { pet: petName(wearer.pet_id) }, era)}</button>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      {:else}
        {#if item.lock}<p>{t("shop.cosmetics.locked", { tier: item.lock.tier }, era)}</p>{/if}
        <button type="button" disabled={pending || !controlsEnabled || !item.acquirable} aria-describedby={describedBy(item.cosmetic_id)}
          onclick={() => buy(item.cosmetic_id)}>{t("shop.cosmetics.buy", { price }, era)}</button>
      {/if}
      {#if receipt && receipt.cosmeticId === item.cosmetic_id}
        <p role="status" class="receipt">{t("shop.receipt.line", { order_number: receipt.orderNumber, item: t(row.title_key, {}, era), price }, era)} {t("shop.receipt.payment_method", {}, era)}</p>
      {/if}
      {#if rejection}<p role="status">{t(rejection, {}, era)}</p>{/if}
      {#each row.curtains as curtain (curtain.pattern)}
        <small id={curtainID(item.cosmetic_id, curtain.pattern)} class="curtain" data-pattern={curtain.pattern}>{t(curtain.key, {}, era)}</small>
      {/each}
    </article>
  {/each}
</section>

<style>
  .shelf { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  .item { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-md); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); background: var(--cc-color-surface); }
  .curtain { display: block; color: var(--cc-color-text); }
  .wearers { display: grid; gap: var(--cc-space-xs); margin: 0; padding: 0; list-style: none; }
  h2, h3, p { margin: 0; }
  .owned:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
</style>
