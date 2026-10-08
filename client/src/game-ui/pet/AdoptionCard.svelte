<script lang="ts">
  import { onDestroy, tick } from "svelte";

  import { t, type CopyEra, type CopyKey } from "../../copy";
  import PetSprite from "./PetSprite.svelte";
  import { petVisualSpec, type PetStatusBand } from "./visual";

  // Pet Adoption v1 PA8: the sincere, inline, non-modal adoption card. It
  // shows no price, urgency, streak, or parody chrome; "Not now" never nags.
  type AdoptionAvailability = Readonly<{ cap: number; count: number; name_keys: readonly string[]; starter_species_id: string }>;
  type AdoptedPet = Readonly<{ pet_id: string; name_key: string; palette_id: string; species_id: string; status_band: PetStatusBand; temperament: string }>;

  let { availability, adopted, era, pending, controlsEnabled, rejection, reducedMotion, onAdopt }: {
    availability: AdoptionAvailability;
    adopted: AdoptedPet | undefined;
    era: CopyEra;
    pending: boolean;
    controlsEnabled: boolean;
    rejection: CopyKey | null;
    reducedMotion: boolean;
    onAdopt(speciesID: string, nameKey: string): Promise<string | undefined>;
  } = $props();

  const speciesKey = $derived(`pet.species.${availability.starter_species_id.slice("pet_species.".length)}.name` as CopyKey);
  const descriptionKey = $derived(`pet.species.${availability.starter_species_id.slice("pet_species.".length)}.description` as CopyKey);
  let chosen = $state<string | undefined>();
  const selected = $derived(chosen ?? availability.name_keys[0]!);
  let collapsed = $state(false);
  let announced = $state<string | undefined>();
  let welcome = $state<HTMLElement | undefined>();
  let root: HTMLElement | undefined;
  let announcedPetID: string | undefined;
  let appliedPetID = $state<string | undefined>();
  let focusAfterReceipt = $state.raw<{ origin: HTMLButtonElement; stopObserving(): void }>();

  function clearFocusRequest(): void {
    focusAfterReceipt?.stopObserving();
    focusAfterReceipt = undefined;
  }
  onDestroy(clearFocusRequest);

  async function adopt(origin: HTMLButtonElement): Promise<void> {
    if (pending || !controlsEnabled) return;
    clearFocusRequest();
    appliedPetID = undefined;
    if (document.activeElement === origin) {
      const observeFocus = (event: FocusEvent) => {
        // A newer choice wins permanently, even if that control disappears
        // before the authoritative pet arrives and focus returns to body.
        if (event.target !== origin) clearFocusRequest();
      };
      document.addEventListener("focusin", observeFocus);
      focusAfterReceipt = { origin, stopObserving: () => document.removeEventListener("focusin", observeFocus) };
    }
    try {
      const petID = await onAdopt(availability.starter_species_id, selected);
      if (!root?.isConnected) return;
      appliedPetID = petID;
      if (!petID) clearFocusRequest();
    } catch (error) {
      clearFocusRequest();
      console.error("game UI invariant: adoption completion", error);
    }
  }

  async function setCollapsed(next: boolean, trigger: EventTarget | null): Promise<void> {
    const focused = trigger instanceof HTMLButtonElement && document.activeElement === trigger ? trigger : undefined;
    collapsed = next;
    await tick();
    // Only replace focus lost with the removed control. A newer player choice
    // outside this card wins; an unmounted card cannot take focus back.
    if (!focused || focused.isConnected || !root?.isConnected ||
        (document.activeElement !== focused && document.activeElement !== document.body)) return;
    const replacement = next ? root.querySelector<HTMLButtonElement>("button[data-adoption-entry]")
      : root.querySelector<HTMLInputElement>('input[name="pet-name"]:checked');
    replacement?.focus();
  }

  // PA8.3: celebrate an applied local receipt only after its matching pet
  // appears. Saved/remote reads and Desk remounts are not new adoptions.
  $effect(() => {
    if (!adopted || appliedPetID !== adopted.pet_id || announcedPetID === adopted.pet_id) return;
    announcedPetID = adopted.pet_id;
    announced = t("pet.adoption.announce.adopted", { pet_name: t(adopted.name_key as CopyKey, {}, era) }, era);
    const request = focusAfterReceipt;
    if (!request) return;
    void tick().then(() => {
      if (focusAfterReceipt !== request || !root?.isConnected) return;
      clearFocusRequest();
      if (!request.origin.isConnected && (document.activeElement === request.origin || document.activeElement === document.body)) welcome?.focus();
    });
  });
</script>

<section bind:this={root} class="adoption" aria-labelledby="pet-adoption-heading">
  <p class="live" class:pending role="status" aria-live="polite">{pending ? t("common.pending", {}, era) : announced ?? ""}</p>
  {#if adopted}
    {@const petName = t(adopted.name_key as CopyKey, {}, era)}
    <h2 id="pet-adoption-heading" tabindex="-1" bind:this={welcome}>{t("pet.adoption.welcome.title", { pet_name: petName }, era)}</h2>
    <div class="pet">
      <PetSprite spec={petVisualSpec(adopted, adopted.status_band, reducedMotion)} />
      <p>{t("pet.adoption.sprite.alt", { pet_name: petName, species: t(speciesKey, {}, era), temperament: t(`pet.temperament.${adopted.temperament}.label` as CopyKey, {}, era) }, era)}</p>
    </div>
    <p>{t("pet.adoption.welcome.body", { pet_name: petName }, era)}</p>
  {:else if collapsed}
    <h2 id="pet-adoption-heading" class="visually-hidden">{t("pet.adoption.card.title", {}, era)}</h2>
    <button type="button" tabindex="0" data-adoption-entry onclick={(event) => { void setCollapsed(false, event.currentTarget); }}>{t("pet.adoption.entry_point", {}, era)}</button>
  {:else}
    <h2 id="pet-adoption-heading">{t("pet.adoption.card.title", {}, era)}</h2>
    <p>{t(descriptionKey, {}, era)}</p>
    <p>{t("pet.adoption.card.body", {}, era)}</p>
    <fieldset>
      <legend>{t("pet.adoption.name_choice.label", {}, era)}</legend>
      <div class="names">
        {#each availability.name_keys as key (key)}
          <label><input type="radio" name="pet-name" value={key} checked={selected === key} tabindex={selected === key ? 0 : -1} onchange={() => { chosen = key; }} /> {t(key as CopyKey, {}, era)}</label>
        {/each}
      </div>
    </fieldset>
    <p class="cap">{t("pet.adoption.cap.label", { count: availability.count, cap: availability.cap }, era)}</p>
    <div class="actions">
      <button type="button" tabindex="0" aria-describedby={rejection ? "pet-adoption-rejection" : undefined} disabled={!controlsEnabled} aria-disabled={pending || undefined}
        onclick={(event) => { void adopt(event.currentTarget); }}>{t("pet.adoption.action.adopt", {}, era)}</button>
      <button type="button" tabindex="0" onclick={(event) => { void setCollapsed(true, event.currentTarget); }}>{t("pet.adoption.action.later", {}, era)}</button>
    </div>
    {#if rejection}<p id="pet-adoption-rejection">{t(rejection, {}, era)}</p>{/if}
  {/if}
</section>

<style>
  .adoption { display: grid; gap: var(--cc-space-md); padding: var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-color-surface);
    border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); }
  h2, p { margin: 0; }
  h2 { font-family: var(--cc-type-font_ui); }
  h2:focus-visible, button:focus-visible, input:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  fieldset { display: grid; gap: var(--cc-space-sm); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); padding: var(--cc-space-md); min-inline-size: 0; }
  .names { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm) var(--cc-space-md); }
  label { display: inline-flex; gap: var(--cc-space-xs); align-items: center; min-block-size: var(--cc-space-xl); }
  .pet { display: flex; flex-wrap: wrap; gap: var(--cc-space-md); align-items: center; }
  .actions { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); }
  button { min-block-size: var(--cc-space-xl); min-inline-size: var(--cc-space-xl); }
  .live:not(.pending), .visually-hidden { position: absolute; inline-size: 1px; block-size: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
</style>
