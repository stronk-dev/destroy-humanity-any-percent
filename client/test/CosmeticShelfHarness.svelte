<script lang="ts">
  import type { ComponentProps } from "svelte";

  import CosmeticShelf from "../src/game-ui/cosmetics/CosmeticShelf.svelte";

  type ShelfProps = ComponentProps<typeof CosmeticShelf>;
  // Test-only harness: re-delivers props to one mounted shelf so a test can
  // step unowned → pending → owned → equipped like authoritative snapshots.
  let { initial }: { initial: ShelfProps } = $props();
  let overrides = $state<Partial<ShelfProps>>({});
  const shelfProps = $derived<ShelfProps>({ ...initial, ...overrides });
  export function update(patch: Partial<ShelfProps>): void { overrides = { ...overrides, ...patch }; }
</script>

<CosmeticShelf {...shelfProps} />
