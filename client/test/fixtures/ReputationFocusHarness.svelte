<script lang="ts">
  import { untrack } from "svelte";
  import type { GameUIReputationArm } from "../../src/api/generated/types";
  import ReputationTreeSurface from "../../src/game-ui/ReputationTreeSurface.svelte";
  import type { CopyEra } from "../../src/copy";

  let { initialArm, era, onPurchase }: {
    initialArm: GameUIReputationArm;
    era: CopyEra;
    onPurchase(id: string): void;
  } = $props();
  let arm = $state(untrack(() => initialArm));
  let pending = $state(false);

  // Controlled props, not a receipt/transport implementation or product seam.
  export function setPending(value: boolean): void { pending = value; }
  export function deliverArm(value: GameUIReputationArm): void { arm = value; pending = false; }
</script>

<ReputationTreeSurface {arm} {era} {pending} controlsEnabled={true} {onPurchase} />
