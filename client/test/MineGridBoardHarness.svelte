<script lang="ts">
  import MineGridBoard from "../src/game-ui/minigame/MineGridBoard.svelte";
  import type { ArcadePreset } from "../src/arcade/catalog";
  import type { MineGridCommand, MineGridSnapshot } from "../src/arcade/mine-grid";

  let { initial, presets, onCommand }: {
    initial: MineGridSnapshot;
    presets: readonly ArcadePreset[];
    onCommand(command: MineGridCommand): void;
  } = $props();
  const initialSnapshot = () => initial;
  let snapshot = $state(initialSnapshot());
  export function advance(next: MineGridSnapshot): void { snapshot = next; }
</script>

<MineGridBoard {snapshot} {presets} era="era_1995" pending={false} {onCommand} />
