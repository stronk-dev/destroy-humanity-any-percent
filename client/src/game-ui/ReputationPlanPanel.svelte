<script lang="ts">
  import type { GameUIReputationArm } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";

  // R6/R9 plan panel. Advisory only: it proposes a purchase order (tree array
  // order, so in-plan prerequisites come first) and the server re-validates
  // the whole plan inside the Exit. Default is empty, so the one-action Exit
  // flow is unchanged.
  let { arm, era, previewDelta, onChange }: {
    arm: GameUIReputationArm;
    era: CopyEra;
    previewDelta: number;
    onChange(plan: readonly string[]): void;
  } = $props();
  let selected = $state<string[]>([]);

  const owned = $derived(new Set(arm.nodes.filter((node) => node.state === "owned").map((node) => node.node_id)));
  const cost = $derived(arm.nodes.filter((node) => selected.includes(node.node_id)).reduce((sum, node) => sum + node.cost, 0));
  const projected = $derived(arm.available + previewDelta - cost);

  function selectable(id: string): boolean {
    const node = arm.nodes.find((row) => row.node_id === id)!;
    if (owned.has(id)) return false;
    if (selected.includes(id)) return true;
    return node.requires.every((requirement) => owned.has(requirement) || selected.includes(requirement)) && node.cost <= projected;
  }
  function toggle(id: string, checked: boolean): void {
    const next = checked ? [...selected, id] : selected.filter((value) => value !== id);
    // Dropping a node also drops any selection that required it.
    let kept = next;
    for (let changed = true; changed;) {
      changed = false;
      kept = kept.filter((value) => {
        const node = arm.nodes.find((row) => row.node_id === value)!;
        const ok = node.requires.every((requirement) => owned.has(requirement) || kept.includes(requirement));
        if (!ok) changed = true;
        return ok;
      });
    }
    selected = arm.nodes.map((node) => node.node_id).filter((value) => kept.includes(value));
    onChange(selected);
  }
</script>

<details class="plan">
  <summary>{t("reputation_tree.plan.heading", {}, era)}</summary>
  <p role="status">{t("reputation_tree.plan.projected", { amount: projected }, era)}</p>
  <fieldset>
    <legend>{t("reputation_tree.plan.heading", {}, era)}</legend>
    {#each arm.nodes.filter((node) => !owned.has(node.node_id)) as node (node.node_id)}
      <label>
        <input type="checkbox" checked={selected.includes(node.node_id)} disabled={!selectable(node.node_id)}
          onchange={(event) => toggle(node.node_id, event.currentTarget.checked)} />
        {t(node.title_key as CopyKey, {}, era)} {t("reputation_tree.action.buy", { cost: node.cost }, era)}
      </label>
    {/each}
  </fieldset>
  <button type="button" disabled={selected.length === 0} onclick={() => { selected = []; onChange(selected); }}>{t("reputation_tree.plan.clear", {}, era)}</button>
</details>

<style>
  .plan { display: grid; gap: var(--cc-space-sm); }
  fieldset { display: grid; gap: var(--cc-space-xs); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  p { margin: 0; }
</style>
