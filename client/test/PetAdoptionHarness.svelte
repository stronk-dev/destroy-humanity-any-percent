<script lang="ts">
  import type { ComponentProps } from "svelte";

  import AdoptionCard from "../src/game-ui/pet/AdoptionCard.svelte";

  type CardProps = ComponentProps<typeof AdoptionCard>;
  // Test-only harness: holds the card's props in state so a test can replay
  // a resync by re-delivering values to the same mounted instance.
  let { initial }: { initial: CardProps } = $props();
  let overrides = $state<Partial<CardProps>>({});
  const cardProps = $derived<CardProps>({ ...initial, ...overrides });
  export function update(patch: Partial<CardProps>): void { overrides = { ...overrides, ...patch }; }
</script>

<AdoptionCard {...cardProps} />
