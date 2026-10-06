<script lang="ts">
  import { untrack } from "svelte";
  import type { GameUIReputationArm } from "../../src/api/generated/types";
  import type { CopyEra } from "../../src/copy";
  import ReputationPlanPanel from "../../src/game-ui/ReputationPlanPanel.svelte";

  let { initialArm, initialPreview, era, onChange }: {
    initialArm: GameUIReputationArm;
    initialPreview: number;
    era: CopyEra;
    onChange(plan: readonly string[]): void;
  } = $props();
  let arm = $state(untrack(() => initialArm));
  let previewDelta = $state(untrack(() => initialPreview));

  // Component-prop diagnostic only. Never mutate the panel's selected state.
  export function deliverArm(value: GameUIReputationArm): void { arm = value; }
  export function deliverPreview(value: number): void { previewDelta = value; }
</script>

<ReputationPlanPanel {arm} {era} {previewDelta} {onChange} />
