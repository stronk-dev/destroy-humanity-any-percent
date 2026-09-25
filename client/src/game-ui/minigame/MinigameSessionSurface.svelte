<script lang="ts">
  import { onDestroy, onMount, tick } from "svelte";

  import type { MinigameResolutionReceipt, MinigameTenantCommand } from "../../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../../copy";
  import Amount from "../../ui/Amount.svelte";
  import PitchTable from "./PitchTable.svelte";
  import { PitchContentMismatch, pitchContentFor, type PitchContent } from "./pitch-content";
  import type { MinigameSessionPort } from "./session-port";
  import { commandRequest, rejectionFor, stateFromCurrent, stateFromSession, type MinigameHostState, type PendingCommand } from "./session-surface";
  import { MINIGAME_TENANT_SURFACES, tenantArm } from "./tenant-registry";

  let { port, minigameID, era, newCommandID, onExitToHost, onTerminal, contentFor = pitchContentFor, retryDelayMS = 1_000 }: {
    port: MinigameSessionPort;
    minigameID: string;
    era: CopyEra;
    newCommandID(): string;
    onExitToHost(): void;
    onTerminal(receipt: MinigameResolutionReceipt): void;
    contentFor?: (hash: string) => Promise<PitchContent>;
    retryDelayMS?: number;
  } = $props();

  const MAX_AUTOMATIC_RETRIES = 5;
  let view = $state<MinigameHostState>({ kind: "loading" });
  let notice = $state<CopyKey | null>(null);
  let content = $state<PitchContent | undefined>();
  let selectionEpoch = $state(0);
  let busy = $state(false);
  let createKey: string | undefined;
  let retryTimer: ReturnType<typeof setTimeout> | undefined;
  let automaticRetries = 0;
  let destroyed = false;
  let heading: HTMLElement | undefined;

  const pending = $derived(busy || view.kind === "active" && view.inFlight !== null);

  onMount(() => { void refetch(); });
  onDestroy(() => { destroyed = true; clearRetry(); });

  function clearRetry(): void { if (retryTimer !== undefined) clearTimeout(retryTimer); retryTimer = undefined; }

  async function bind(next: MinigameHostState): Promise<void> {
    if (destroyed) return;
    if ((next.kind === "active" || next.kind === "required_terminal") && !checkTenant(next.kind === "active" ? next.session : next.response)) return;
    if (next.kind === "active") {
      try { content = await contentFor(next.snapshot.pitch_content_hash); }
      catch (error) { view = { kind: "error", message: error instanceof PitchContentMismatch ? "minigame.error.content_mismatch" : "minigame.error.generic" }; return; }
      automaticRetries = 0;
    }
    const wasTerminal = view.kind === "required_terminal";
    view = next;
    if (next.kind === "paused_reconnect") scheduleRetry();
    if (next.kind === "required_terminal" && !wasTerminal) {
      onTerminal(next.response.resolution_receipt);
      await tick();
      heading?.focus();
    }
  }

  function checkTenant(session: { engine_ref: string; engine_version: string; minigame_id: string }): boolean {
    const row = MINIGAME_TENANT_SURFACES.get(tenantArm(session.engine_ref, session.engine_version));
    if (row && row.minigame_id === session.minigame_id) return true;
    view = { kind: "error", message: "minigame.error.generic" };
    return false;
  }

  function scheduleRetry(): void {
    clearRetry();
    if (automaticRetries >= MAX_AUTOMATIC_RETRIES) return;
    automaticRetries += 1;
    retryTimer = setTimeout(() => { retryTimer = undefined; void retry(); }, retryDelayMS);
  }

  async function refetch(): Promise<void> {
    busy = true;
    try { await bind(stateFromCurrent(await port.current())); }
    catch (error) { await reject(error, null); }
    finally { busy = false; }
  }

  async function reject(error: unknown, failed: PendingCommand | null): Promise<void> {
    const effect = rejectionFor(error);
    notice = effect.notice;
    if (effect.clearSelection) selectionEpoch += 1;
    const last = view.kind === "active" ? { session: view.session, snapshot: view.snapshot } : view.kind === "paused_reconnect" ? view.last : null;
    switch (effect.effect) {
      case "stay":
        if (view.kind === "active") view = { ...view, inFlight: null };
        else if (view.kind === "paused_reconnect" && last) view = { kind: "active", ...last, inFlight: null };
        else if (view.kind === "loading") view = { kind: "launcher", notice: effect.notice };
        return;
      case "refetch": await refetch(); return;
      case "launcher": view = { kind: "launcher", notice: effect.notice }; return;
      case "pause": await bind({ kind: "paused_reconnect", last, retry: failed }); return;
      case "error": view = { kind: "error", message: effect.notice ?? "minigame.error.generic" }; return;
    }
  }

  async function start(): Promise<void> {
    if (pending) return;
    notice = null;
    // The key is held until success so a retried create is idempotent.
    createKey ??= newCommandID();
    busy = true;
    try { await bind(stateFromSession(await port.create(minigameID, createKey))); createKey = undefined; }
    catch (error) { await reject(error, null); }
    finally { busy = false; }
  }

  async function send(pendingCommand: PendingCommand): Promise<void> {
    if (view.kind === "active") view = { ...view, inFlight: pendingCommand };
    notice = null;
    try {
      const response = await port.command(pendingCommand.sessionID, pendingCommand.request);
      await bind(stateFromSession(response));
    } catch (error) { await reject(error, pendingCommand); }
  }

  function dispatch(command: MinigameTenantCommand): void {
    if (view.kind !== "active" || pending) return;
    void send(commandRequest(view, newCommandID(), command));
  }

  async function retry(): Promise<void> {
    clearRetry();
    if (view.kind !== "paused_reconnect") return;
    // Resend the SAME command_id and body; the server's receipt makes it
    // idempotent (MA-C13). With nothing pending, re-read the session.
    const pendingCommand = view.retry;
    if (pendingCommand && view.last) {
      view = { kind: "active", ...view.last, inFlight: pendingCommand };
      await send(pendingCommand);
    } else await refetch();
  }

  function manualRetry(): void { automaticRetries = 0; void retry(); }
</script>

<section class="minigame-session cc-window" aria-labelledby="minigame-heading" aria-busy={pending}>
  <h1 id="minigame-heading" tabindex="-1" bind:this={heading}>{t("minigame.pitch.title", {}, era)}</h1>
  {#if notice}<p class="notice" role="status">{t(notice, {}, era)}</p>{/if}

  {#if view.kind === "loading"}
    <p role="status">{t("minigame.state.loading", {}, era)}</p>
  {:else if view.kind === "launcher"}
    <p>{t("minigame.state.none", {}, era)}</p>
    <button type="button" disabled={pending} onclick={start}>{t("minigame.start", {}, era)}</button>
  {:else if view.kind === "active" && content}
    <PitchTable snapshot={view.snapshot} {content} {era} {pending} {selectionEpoch}
      onPlayHand={(cardIDs) => dispatch({ kind: "play_hand", card_ids: [...cardIDs] })}
      onBuyHack={(offerID) => dispatch({ kind: "buy_hack", offer_id: offerID })}
      onEndShop={() => dispatch({ kind: "end_shop" })} />
  {:else if view.kind === "paused_reconnect"}
    <p role="alert">{t("minigame.state.paused_reconnect", {}, era)}</p>
    <button type="button" disabled={busy} onclick={manualRetry}>{t("minigame.retry", {}, era)}</button>
  {:else if view.kind === "required_terminal"}
    {@const receipt = view.response.resolution_receipt}
    <p role="status">{t("pitch.table.terminal", {}, era)}</p>
    <p>{t("minigame.receipt.credited_label", {}, era)} <Amount value={receipt.credited_delta} {era} /></p>
    {#if receipt.configured_cap_forfeit_units > 0}<p>{t("minigame.receipt.forfeit_frame", { units: receipt.configured_cap_forfeit_units }, era)}</p>{/if}
    <button type="button" onclick={onExitToHost}>{t("minigame.return", {}, era)}</button>
  {:else if view.kind === "error"}
    <p role="alert">{t(view.message, {}, era)}</p>
    <button type="button" onclick={onExitToHost}>{t("minigame.return", {}, era)}</button>
  {/if}

  {#if view.kind === "active" || view.kind === "launcher" || view.kind === "paused_reconnect"}
    <button type="button" class="leave" onclick={onExitToHost}>{t("minigame.leave_table", {}, era)}</button>
  {/if}
</section>

<style>
  .minigame-session { display: grid; gap: var(--cc-space-md); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); color: var(--cc-color-text); }
  h1 { margin: 0; font-family: var(--cc-type-font_display); }
  h1:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
  p { margin: 0; }
  .notice { padding: var(--cc-space-sm); background: var(--cc-color-surface); }
  .leave { justify-self: start; }
</style>
