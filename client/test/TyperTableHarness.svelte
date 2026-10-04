<script lang="ts">
  import TyperTable from "../src/game-ui/minigame/TyperTable.svelte";
  import type { TyperCommand, TyperSnapshot } from "../src/typer/engine";

  let { initial, dispatch }: { initial: TyperSnapshot; dispatch(command: TyperCommand): void } = $props();
  const initialSnapshot = () => initial;
  let snapshot = $state(initialSnapshot());
  export function advance(next: TyperSnapshot): void { snapshot = next; }
</script>

<TyperTable {snapshot} serverTimeSample={snapshot.last_server_ms ?? 1_000} pending={false}
  era="era_2000" {dispatch} exitToHost={() => {}} />
