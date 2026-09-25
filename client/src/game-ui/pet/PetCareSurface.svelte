<script lang="ts">
  import type { GameUICosmeticsArm, GameUIPetRow } from "../../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../../copy";
  import CosmeticOverlay from "../cosmetics/CosmeticOverlay.svelte";
  import { COSMETIC_SHOP_PRESENTATION } from "../cosmetics/presentation";
  import { FEATURES_PRESENTATION } from "../features-presentation";
  import PetSprite from "./PetSprite.svelte";
  import { petVisualSpec } from "./visual";

  // GS4 over the PA7 projection: identity, status band, and the server's
  // eligible actions only. Every catalog action gets a button in catalog
  // order; one the server does not list as eligible is disabled AND says so in
  // text (never greying alone). Eligibility is decided by the server; this
  // surface pre-evaluates nothing. Sincere tone; no price, urgency, or parody.
  let { pets, cosmetics, era, pending, controlsEnabled, reducedMotion, onCare }: {
    pets: readonly GameUIPetRow[];
    cosmetics: GameUICosmeticsArm | null;
    era: CopyEra;
    pending: boolean;
    controlsEnabled: boolean;
    reducedMotion: boolean;
    onCare(petID: string, actionID: string): void;
  } = $props();

  function worn(petID: string) {
    const id = cosmetics?.wearers.find((row) => row.pet_id === petID)?.worn;
    return id ? COSMETIC_SHOP_PRESENTATION.cosmetics.get(id) : undefined;
  }
</script>

<section class="surface pet-care" aria-labelledby="pet-care-heading">
  <h1 id="pet-care-heading">{t("pet.care.panel.title", {}, era)}</h1>
  {#each pets as pet (pet.pet_id)}
    {@const name = t(pet.name_key as CopyKey, {}, era)}
    {@const band = FEATURES_PRESENTATION.petBands.get(pet.status_band)}
    {@const wearing = worn(pet.pet_id)}
    <article class="pet cc-window" aria-label={name}>
      <div class="portrait">
        <PetSprite spec={petVisualSpec(pet, pet.status_band, reducedMotion)} />
        {#if wearing}<CosmeticOverlay renderKey={wearing.render_key} reaction={wearing.wearer_reaction} animate={!reducedMotion} />{/if}
      </div>
      <h2>{name}</h2>
      {#if band}<p>{t(band, {}, era)}</p>{/if}
      <ul class="actions">
        {#each [...FEATURES_PRESENTATION.petActions] as [actionID, titleKey] (actionID)}
          {@const eligible = pet.eligible_action_ids.includes(actionID)}
          <li>
            <button type="button" disabled={pending || !controlsEnabled || !eligible} onclick={() => onCare(pet.pet_id, actionID)}>{t(titleKey, {}, era)}</button>
            {#if !eligible}<small>{t("pet.care.action.unavailable", {}, era)}</small>{/if}
          </li>
        {/each}
      </ul>
    </article>
  {/each}
</section>

<style>
  .pet-care { display: grid; gap: var(--cc-space-lg); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); }
  .pet { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-md); }
  .portrait { position: relative; inline-size: fit-content; }
  .actions { display: grid; grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr)); gap: var(--cc-space-sm); margin: 0; padding: 0; list-style: none; }
  .actions li { display: grid; gap: var(--cc-space-xs); }
  h1, h2, p { margin: 0; }
</style>
