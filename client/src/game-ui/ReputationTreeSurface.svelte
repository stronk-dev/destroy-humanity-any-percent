<script lang="ts">
  import { tick } from "svelte";

  import type { GameUIReputationArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";

  // Reputation Tree v1 R9. Every state is server-derived (arm.nodes[].state);
  // the client never recomputes eligibility and the receipt is the authority.
  // Buying is irreversible, so Buy only reveals an inline Confirm/Cancel pair.
  let { arm, era, pending, controlsEnabled, onPurchase }: {
    arm: GameUIReputationArm;
    era: CopyEra;
    pending: boolean;
    controlsEnabled: boolean;
    onPurchase(nodeID: string): void;
  } = $props();

  let confirming = $state<string | null>(null);
  const confirmButtons = new Map<string, HTMLButtonElement>();
  const buyButtons = new Map<string, HTMLButtonElement>();

  const titles = $derived(new Map(arm.nodes.map((node) => [node.node_id, t(node.title_key as CopyKey, {}, era)])));
  const STATE_KEYS: Readonly<Record<string, CopyKey>> = {
    available: "reputation_tree.state.available", locked: "reputation_tree.state.locked",
    owned: "reputation_tree.state.owned", unaffordable: "reputation_tree.state.unaffordable",
  };

  function register(map: Map<string, HTMLButtonElement>, id: string) {
    return (element: HTMLButtonElement) => { map.set(id, element); return () => { map.delete(id); }; };
  }
  async function openConfirm(id: string): Promise<void> {
    confirming = id;
    await tick();
    confirmButtons.get(id)?.focus();
  }
  async function cancel(id: string): Promise<void> {
    confirming = null;
    await tick();
    buyButtons.get(id)?.focus();
  }
  function confirm(id: string): void {
    confirming = null;
    onPurchase(id);
    buyButtons.get(id)?.focus();
  }
  function keydown(event: KeyboardEvent, id: string): void {
    if (event.key === "Escape" && confirming === id) { event.preventDefault(); void cancel(id); }
  }
</script>

<section class="surface reputation" aria-labelledby="reputation-heading">
  <h1 id="reputation-heading" tabindex="-1">{t("reputation_tree.title", {}, era)}</h1>
  <p>{t("reputation_tree.intro", {}, era)}</p>
  <section class="card" aria-label={t("reputation_tree.title", {}, era)}>
    <p>{t("reputation_tree.balance.available", { amount: arm.available }, era)}</p>
    <p>{t("reputation_tree.balance.level", { amount: arm.level }, era)}</p>
    <p>{t("reputation_tree.balance.spent", { amount: arm.spent }, era)}</p>
    <p>{arm.bonus_factor_this_run === null ? t("reputation_tree.bonus.this_run_none", {}, era) : t("reputation_tree.bonus.this_run", { factor: arm.bonus_factor_this_run }, era)}</p>
    <p>{t("reputation_tree.bonus.next_run", { factor: arm.bonus_factor_next_run }, era)}</p>
    <p>{t("reputation_tree.bonus.formula", { perlevel: arm.per_level_ppm, unlock: arm.unlock_ppm }, era)}</p>
    <p>{t("reputation_tree.applies_next_run", {}, era)}</p>
  </section>
  <ol class="nodes">
    {#each arm.nodes as node (node.node_id)}
      <li class="card" data-state={node.state}>
        <h2>{titles.get(node.node_id)}</h2>
        <p>{t(node.body_key as CopyKey, {}, era)}</p>
        {#if node.requires.length}<p>{t("reputation_tree.requires", { list: node.requires.map((id) => titles.get(id) ?? "").join(", ") }, era)}</p>{/if}
        <p class="state">{t(STATE_KEYS[node.state]!, {}, era)}</p>
        {#if node.state === "available"}
          {#if confirming === node.node_id}
            <span class="confirm" role="group" aria-label={titles.get(node.node_id)}>
              <button type="button" {@attach register(confirmButtons, node.node_id)} disabled={pending || !controlsEnabled}
                onclick={() => confirm(node.node_id)} onkeydown={(event) => keydown(event, node.node_id)}>{t("reputation_tree.action.confirm", {}, era)}</button>
              <button type="button" onclick={() => cancel(node.node_id)} onkeydown={(event) => keydown(event, node.node_id)}>{t("reputation_tree.action.cancel", {}, era)}</button>
            </span>
          {:else}
            <button type="button" {@attach register(buyButtons, node.node_id)} disabled={pending || !controlsEnabled}
              onclick={() => openConfirm(node.node_id)}>{t("reputation_tree.action.buy", { cost: node.cost }, era)}</button>
          {/if}
        {/if}
      </li>
    {/each}
  </ol>
</section>

<style>
  .reputation { display: grid; gap: var(--cc-space-md); }
  .card { display: grid; gap: var(--cc-space-xs); padding: var(--cc-space-md); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  .nodes { display: grid; gap: var(--cc-space-sm); padding: 0; list-style: none; }
  .card[data-state="owned"] { border-color: var(--cc-color-accent); }
  .confirm { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); }
  h1, h2, p { margin: 0; }
</style>
