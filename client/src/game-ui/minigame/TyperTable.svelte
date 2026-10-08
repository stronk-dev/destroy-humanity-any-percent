<script lang="ts">
  import { onDestroy, onMount, tick } from "svelte";

  import { t, type CopyEra } from "../../copy";
  import type { TyperCommand, TyperSnapshot } from "../../typer/engine";

  // Tenant child for (typer, 1.0.0), TT9. Presentation only: it renders the
  // server snapshot and raises commands; local keystroke echo is ordinary
  // text-field editing and nothing here is authoritative. The host owns
  // transport, chrome, errors, and lifecycle.
  let { snapshot, serverTimeSample, pending, era, dispatch, exitToHost, monotonicNow = () => performance.now() }: {
    snapshot: TyperSnapshot;
    serverTimeSample: number;
    pending: boolean;
    era: CopyEra;
    dispatch(command: TyperCommand): void;
    exitToHost(): void;
    monotonicNow?: () => number;
  } = $props();

  let line = $state("");
  let composing = false;
  let input = $state<HTMLInputElement | undefined>();
  let root: HTMLElement | undefined;
  let leave: HTMLButtonElement | undefined;
  let sampledAt = $state(0);
  let now = $state(0);
  let timer: ReturnType<typeof setInterval> | undefined;
  let lastPromptID: string | null = null;
  let lastPhase: TyperSnapshot["phase"] | undefined;
  let promptAnnouncement = $state("");

  onMount(() => {
    now = monotonicNow();
    timer = setInterval(() => { now = monotonicNow(); }, 1000);
  });
  onDestroy(() => { if (timer !== undefined) clearInterval(timer); });

  // A new server sample restarts local elapsed time (the Game UI RTA pattern).
  $effect(() => {
    void serverTimeSample; void snapshot.revision;
    // Reading sampledAt here would make the effect depend on its own write.
    const sample = monotonicNow();
    sampledAt = sample; now = sample;
  });

  // Capture focus before a response removes its control. Keep typed text on a
  // miss, and never undo a newer player focus choice while awaiting render.
  $effect.pre(() => {
    const promptID = snapshot.current_prompt_id;
    const phase = snapshot.phase;
    if (promptID === lastPromptID && phase === lastPhase) return;
    const previousPrompt = lastPromptID;
    const focused = document.activeElement;
    const ownedFocus = focused !== null && root?.contains(focused);
    const enterInput = promptID !== null && (focused === input ||
      previousPrompt === null && (ownedFocus || focused === document.body));
    const terminalHandoff = phase === "terminal" && ownedFocus && focused !== leave;
    lastPromptID = promptID;
    lastPhase = phase;
    promptAnnouncement = snapshot.current_prompt_text ?? "";
    line = "";
    if (!enterInput && !terminalHandoff) return;
    void tick().then(() => {
      if (!root?.isConnected || snapshot.current_prompt_id !== promptID || snapshot.phase !== phase) return;
      if (document.activeElement !== focused && !(focused && !focused.isConnected && document.activeElement === document.body)) return;
      if (terminalHandoff) leave?.focus();
      else input?.focus();
    });
  });

  const remainingSeconds = $derived.by(() => {
    if (snapshot.deadline_server_ms === null) return null;
    const serverNow = Math.max(serverTimeSample, snapshot.last_server_ms ?? serverTimeSample) + Math.max(0, now - sampledAt);
    return Math.max(0, Math.ceil((snapshot.deadline_server_ms - serverNow) / 1000));
  });
  const expired = $derived(remainingSeconds === 0);

  function submit(event: SubmitEvent): void {
    event.preventDefault();
    if (pending || composing || snapshot.phase !== "typing") return;
    dispatch({ kind: "submit_line", text: line });
  }

  function begin(assist_level: "timed" | "untimed"): void {
    if (!pending && snapshot.phase === "ready") dispatch({ kind: "begin", assist_level });
  }

  function endRun(): void {
    if (!pending && snapshot.phase !== "terminal") dispatch({ kind: "end_run" });
  }

  // Enter during IME composition confirms the composition, never the line.
  function keydown(event: KeyboardEvent): void {
    if (event.key === "Enter" && (event.isComposing || composing)) event.preventDefault();
  }
</script>

<section bind:this={root} class="typer" aria-labelledby="typer-heading" aria-busy={pending}>
  <h2 id="typer-heading">{t("typer.title", {}, era)}</h2>
  <p class="host">{t("typer.host", {}, era)}</p>
  <p class="prompt-announcement" aria-live="polite" aria-atomic="true">{promptAnnouncement}</p>

  {#if snapshot.phase === "ready"}
    <div class="modes">
      <button type="button" tabindex="0" aria-disabled={pending || undefined} onclick={() => begin("timed")}>{t("typer.mode.timed", {}, era)}</button>
      <button type="button" tabindex="0" aria-disabled={pending || undefined} aria-describedby="typer-untimed-note" onclick={() => begin("untimed")}>{t("typer.mode.untimed", {}, era)}</button>
    </div>
    <p id="typer-untimed-note" class="note">{t("typer.mode.untimed.note", {}, era)}</p>
  {:else if snapshot.phase === "typing"}
    <p>{t("typer.progress_frame", { current: snapshot.prompt_index + 1, total: snapshot.prompts_total }, era)}</p>
    {#if remainingSeconds !== null}
      <p class="time">{expired ? t("typer.expired", {}, era) : t("typer.time_remaining", { seconds: remainingSeconds }, era)}</p>
    {/if}
    <p id="typer-prompt-label">{t("typer.prompt_label", {}, era)}</p>
    <code class="prompt" aria-labelledby="typer-prompt-label">{snapshot.current_prompt_text}</code>
    <form onsubmit={submit}>
      <label for="typer-line">{t("typer.input_label", {}, era)}</label>
      <input id="typer-line" bind:this={input} bind:value={line} type="text" autocomplete="off" autocapitalize="off" spellcheck="false"
        autocorrect="off" onkeydown={keydown} oncompositionstart={() => { composing = true; }} oncompositionend={() => { composing = false; }} />
      <button type="submit" tabindex="0" aria-disabled={pending || undefined}>{t("typer.submit", {}, era)}</button>
    </form>
    <p class="feedback" role="status" aria-live="polite">
      {#if snapshot.last_submission?.outcome === "miss"}{t("typer.feedback.miss", { index: (snapshot.last_submission.first_mismatch_index ?? 0) + 1 }, era)}{:else if snapshot.last_submission?.outcome === "cleared"}{t("typer.feedback.cleared", {}, era)}{/if}
    </p>
  {/if}

  {#if snapshot.phase !== "terminal"}
    <button type="button" tabindex="0" aria-disabled={pending || undefined} onclick={endRun}>{t("typer.end_run", {}, era)}</button>
  {/if}
  <button bind:this={leave} type="button" tabindex="0" class="leave" onclick={exitToHost}>{t("minigame.leave_table", {}, era)}</button>
</section>

<style>
  .typer { display: grid; gap: var(--cc-space-md); min-width: 0; }
  h2 { margin: 0; font-family: var(--cc-type-font_display); }
  p { margin: 0; }
  .modes { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); }
  .prompt { display: block; padding: var(--cc-space-sm); overflow-wrap: anywhere; white-space: pre-wrap; font-family: var(--cc-type-font_mono);
    background: var(--cc-color-surface); color: var(--cc-color-text); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); }
  form { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); align-items: center; }
  input { flex: 1 1 12rem; min-width: 0; font-family: var(--cc-type-font_mono); font-size: inherit; color: var(--cc-color-text);
    background: var(--cc-color-bg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); padding: var(--cc-space-xs); }
  input:focus-visible, button:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  .leave { justify-self: start; }
  .prompt-announcement { position: absolute; inline-size: 1px; block-size: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
</style>
