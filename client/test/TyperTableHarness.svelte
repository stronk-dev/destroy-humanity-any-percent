<script lang="ts">
  import TyperTable from "../src/game-ui/minigame/TyperTable.svelte";
  import type { TyperCommand, TyperSnapshot } from "../src/typer/engine";

  let { initial, dispatch, monotonicNow, exitToHost = () => {} }: {
    initial: TyperSnapshot; dispatch(command: TyperCommand): void; monotonicNow?: () => number; exitToHost?: () => void;
  } = $props();
  const initialSnapshot = () => initial;
  let snapshot = $state(initialSnapshot());
  let pending = $state(false);
  export function advance(next: TyperSnapshot): void { snapshot = next; }
  export function setPending(value: boolean): void { pending = value; }
</script>

<TyperTable {snapshot} serverTimeSample={snapshot.last_server_ms ?? 1_000} {pending}
  era="era_2000" {dispatch} {monotonicNow} {exitToHost} />
