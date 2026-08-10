<script lang="ts">
  import { onDestroy } from "svelte";

  import { applicationCopyCatalog, resolveCopy, type CopyEra } from "../copy";
  import { parseCanonical } from "../numeric";
  import { formatAmount } from "./amount-format";
  import type { AmountCap } from "./amount-types";
  import { amountRenderScheduler } from "./render-scheduler";

  let { value, cap, era }: { value: string; cap?: AmountCap; era?: CopyEra } = $props();
  const owner = {};
  let rendered = $state("");
  let priorValue: string | undefined;
  let priorCap: string | undefined;

  $effect(() => {
    const nextValue = value;
    const nextCap = cap ? `${cap.amount}\0${cap.reason_key}\0${era ?? ""}` : "";
    if (priorValue === undefined || nextCap !== priorCap) {
      amountRenderScheduler.cancel(owner);
      rendered = formatAmount(nextValue);
    } else if (nextValue !== priorValue) {
      amountRenderScheduler.schedule(owner, () => { rendered = formatAmount(nextValue); });
    }
    priorValue = nextValue;
    priorCap = nextCap;
  });

  const capped = $derived(cap !== undefined && parseCanonical(value).gte(parseCanonical(cap.amount)));
  const capReason = $derived(capped && cap
    ? resolveCopy(applicationCopyCatalog, cap.reason_key, {}, era)
    : undefined);

  onDestroy(() => amountRenderScheduler.cancel(owner));
</script>

<span class="cc-amount" data-capped={capped || undefined}>
  <output aria-label={capReason}>{rendered}</output>
  {#if capReason}<small class="cc-amount__reason">{capReason}</small>{/if}
</span>

<style>
  .cc-amount { display: inline-flex; flex-direction: column; gap: var(--cc-space-xs); color: var(--cc-color-text); font-family: var(--cc-type-font_mono); font-variant-numeric: tabular-nums; }
  .cc-amount[data-capped] output { color: var(--cc-color-danger); font-weight: var(--cc-type-weight_bold); }
  .cc-amount__reason { color: var(--cc-color-text_muted); font-family: var(--cc-type-font_ui); font-size: var(--cc-type-size_small); }
</style>
