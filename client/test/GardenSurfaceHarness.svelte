<script lang="ts">
  import type { ComponentProps } from "svelte";
  import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";

  type GardenProps = ComponentProps<typeof GardenSurface>;
  // Test-only receipt/props delivery into one actual mounted component.
  let { initial }: { initial: GardenProps } = $props();
  let overrides = $state<Partial<GardenProps>>({});
  const gardenProps = $derived<GardenProps>({ ...initial, ...overrides });
  export function update(patch: Partial<GardenProps>): void { overrides = { ...overrides, ...patch }; }
</script>

<GardenSurface {...gardenProps} />
