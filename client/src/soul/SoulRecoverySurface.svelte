<script lang="ts">
  import { onDestroy, tick } from "svelte";

  import type { SoulRecoveryProgressResponse, SoulRecoveryStartResponse, SoulRecoveryTerminalResponse } from "../api/generated/types";
  import { t, type CopyEra, type CopyKey } from "../copy";
  import type { SoulRecoveryActivity } from "./catalog";
  import { RecoveryScheduler, type RecoveryPauseReason, type RecoveryVisibility } from "./recovery-scheduler";
  import { localToySeed, recoveryActivity, recoveryRejectionFor, type SoulRecoveryContent, type SoulRecoveryPort } from "./recovery-surface";
  import RecoveryToy from "./RecoveryToy.svelte";

  let { port, content, era, onExitToHost, onTerminal, visibility = documentVisibility(), now = () => Math.floor(performance.now()), toySeed = localToySeed() }: {
    port: SoulRecoveryPort;
    content: SoulRecoveryContent;
    era: CopyEra;
    onExitToHost(): void;
    onTerminal(response: SoulRecoveryTerminalResponse): void;
    visibility?: RecoveryVisibility;
    now?: () => number;
    toySeed?: number;
  } = $props();

  type Session = Readonly<{ activity: SoulRecoveryActivity; sessionID: string; attended: number; required: number; eligible: boolean }>;
  type View =
    | Readonly<{ kind: "picker" }>
    | Readonly<{ kind: "active"; session: Session; paused: RecoveryPauseReason | null }>
    | Readonly<{ kind: "terminal"; response: SoulRecoveryTerminalResponse }>
    | Readonly<{ kind: "ended"; message: CopyKey }>
    | Readonly<{ kind: "error"; message: CopyKey }>;

  let view = $state<View>({ kind: "picker" });
  let notice = $state<CopyKey | null>(null);
  let busy = $state(false);
  let scheduler: RecoveryScheduler | undefined;
  let lastProgressError: unknown;
  let heading: HTMLElement | undefined;

  onDestroy(() => { scheduler?.stop(); scheduler = undefined; });

  function documentVisibility(): RecoveryVisibility {
    return { subscribe(callback) {
      const listener = () => callback(document.visibilityState === "visible");
      document.addEventListener("visibilitychange", listener);
      return () => document.removeEventListener("visibilitychange", listener);
    } };
  }

  function sessionFrom(activity: SoulRecoveryActivity, response: SoulRecoveryStartResponse | SoulRecoveryProgressResponse, eligible: boolean): Session {
    return { activity, sessionID: response.session_id, attended: response.attended_progress_ms, required: response.required_duration_attended_ms, eligible };
  }

  function bindScheduler(session: Session, token: string): void {
    scheduler?.stop();
    lastProgressError = undefined;
    scheduler = new RecoveryScheduler({
      session_id: session.sessionID, progress_token: token, beat_interval_ms: content.beatIntervalMS, now, visibility,
      transport: { progress: async (sessionID, progressToken) => {
        try { return await port.progress(sessionID, progressToken); }
        catch (error) { lastProgressError = error; throw error; }
      } },
    }, {
      on_progress: (response) => {
        if (view.kind !== "active") return;
        view = { kind: "active", session: sessionFrom(view.session.activity, response, response.eligible), paused: view.paused };
      },
      on_pause: (reason) => {
        if (view.kind !== "active") return;
        // A session the server already ended (watchdog/terminal) must not be
        // silently replaced by a reconnect-start.
        if (reason === "network" && lastProgressError !== undefined && recoveryRejectionFor(lastProgressError).effect === "gone") { endSession("soul.recovery_surface.gone"); return; }
        // A required reconnect outranks a background pause: the scheduler
        // will not resume until reconnect rotates the token.
        view = { ...view, paused: view.paused === "network" ? "network" : reason };
      },
      on_resume: () => { if (view.kind === "active" && view.paused === "hidden") view = { ...view, paused: null }; },
      on_token_rotated: () => {},
      on_terminal: () => {},
    });
    scheduler.start();
  }

  function endSession(message: CopyKey): void {
    scheduler?.stop();
    scheduler = undefined;
    view = { kind: "ended", message };
  }

  async function begin(activity: SoulRecoveryActivity): Promise<void> {
    if (busy) return;
    busy = true; notice = null;
    try {
      const response = await port.start(activity.activity_id);
      const session = sessionFrom(recoveryActivity(content, response.activity_id), response, response.attended_progress_ms >= response.required_duration_attended_ms);
      view = { kind: "active", session, paused: null };
      bindScheduler(session, response.progress_token);
    } catch (error) { fail(error); }
    finally { busy = false; }
  }

  // Reconnect is start() again: the server rotates the progress token for the
  // same session (SR-C6). A different session ID means the old one ended.
  async function reconnect(): Promise<void> {
    if (view.kind !== "active" || busy) return;
    const current = view.session;
    busy = true; notice = null;
    try {
      const response = await port.start(current.activity.activity_id);
      const session = sessionFrom(current.activity, response, response.attended_progress_ms >= response.required_duration_attended_ms);
      view = { kind: "active", session, paused: null };
      if (response.session_id === current.sessionID && scheduler) { lastProgressError = undefined; scheduler.reconnect(response.progress_token); }
      else bindScheduler(session, response.progress_token);
    } catch (error) { fail(error); }
    finally { busy = false; }
  }

  async function finish(action: "resolve" | "cancel"): Promise<void> {
    if (view.kind !== "active" || busy) return;
    busy = true; notice = null;
    try {
      const response = action === "resolve" ? await port.resolve(view.session.sessionID) : await port.cancel(view.session.sessionID);
      scheduler?.stop(action === "resolve" ? "resolved" : response.cancelled_by === "watchdog" ? "watchdog" : "cancelled");
      scheduler = undefined;
      view = { kind: "terminal", response };
      onTerminal(response);
      await tick();
      heading?.focus();
    } catch (error) { fail(error); }
    finally { busy = false; }
  }

  function fail(error: unknown): void {
    const rejection = recoveryRejectionFor(error);
    if (rejection.effect === "notice") { notice = rejection.notice; return; }
    if (rejection.effect === "gone") { endSession("soul.recovery_surface.gone"); return; }
    if (rejection.effect === "reconnect") { if (view.kind === "active") view = { ...view, paused: "network" }; else notice = "minigame.state.paused_reconnect"; return; }
    scheduler?.stop(); scheduler = undefined;
    view = { kind: "error", message: rejection.notice };
  }

  function terminalMessage(response: SoulRecoveryTerminalResponse): string {
    if (response.action === "resolve") return t("soul.recovery_surface.resolved_frame", { after: response.soul_after, before: response.soul_before }, era);
    return t(response.cancelled_by === "watchdog" ? "soul.recovery_surface.watchdog" : "soul.recovery_surface.cancelled", {}, era);
  }
</script>

<section class="recovery cc-window" aria-labelledby="recovery-heading" aria-busy={busy}>
  <h1 id="recovery-heading" tabindex="-1" bind:this={heading}>{t("soul.recovery_surface.heading", {}, era)}</h1>
  {#if notice}<p class="notice" role="status">{t(notice, {}, era)}</p>{/if}

  {#if view.kind === "picker"}
    <p>{t("soul.recovery_surface.intro", {}, era)}</p>
    <ul class="activities">
      {#each content.catalog.recovery_activities as activity (activity.activity_id)}
        <li>
          <h2>{t(activity.title_copy_key as CopyKey, {}, era)}</h2>
          <p>{t(activity.description_copy_key as CopyKey, {}, era)}</p>
          <p>{t("soul.recovery_surface.duration_frame", { minutes: Math.ceil(activity.duration_attended_ms / 60_000) }, era)}</p>
          <small id={`disclosure-${activity.activity_id}`}>{t(activity.disclosure_copy_key as CopyKey, {}, era)}</small>
          <button type="button" disabled={busy} aria-describedby={`disclosure-${activity.activity_id}`} onclick={() => begin(activity)}>{t("soul.recovery_surface.begin", {}, era)}</button>
        </li>
      {/each}
    </ul>
    <button type="button" onclick={onExitToHost}>{t("soul.recovery_surface.back", {}, era)}</button>
  {:else if view.kind === "active"}
    {@const session = view.session}
    <h2>{t(session.activity.title_copy_key as CopyKey, {}, era)}</h2>
    <RecoveryToy toyKind={session.activity.toy_kind} seed={toySeed} fraction={session.required > 0 ? session.attended / session.required : 0} />
    <label>{t("soul.recovery_surface.progress_label", {}, era)} <progress max={session.required} value={Math.min(session.attended, session.required)}></progress></label>
    <small>{t(session.activity.disclosure_copy_key as CopyKey, {}, era)}</small>
    {#if view.paused === "hidden"}<p role="status">{t("soul.recovery_surface.paused_hidden", {}, era)}</p>{/if}
    {#if view.paused === "network"}
      <p role="alert">{t("soul.recovery_surface.paused_network", {}, era)}</p>
      <button type="button" disabled={busy} onclick={reconnect}>{t("soul.recovery_surface.reconnect", {}, era)}</button>
    {/if}
    {#if session.eligible}<button type="button" disabled={busy} onclick={() => finish("resolve")}>{t("soul.recovery_surface.finish", {}, era)}</button>{/if}
    <button type="button" disabled={busy} aria-describedby="recovery-cancel-note" onclick={() => finish("cancel")}>{t("soul.recovery_surface.cancel", {}, era)}</button>
    <small id="recovery-cancel-note">{t("soul.recovery_surface.cancel_note", {}, era)}</small>
  {:else if view.kind === "terminal"}
    <p role="status">{terminalMessage(view.response)}</p>
    <button type="button" onclick={onExitToHost}>{t("soul.recovery_surface.back", {}, era)}</button>
  {:else}
    <p role="alert">{t(view.message, {}, era)}</p>
    <button type="button" onclick={onExitToHost}>{t("soul.recovery_surface.back", {}, era)}</button>
  {/if}
</section>

<style>
  .recovery { display: grid; gap: var(--cc-space-md); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); color: var(--cc-color-text); }
  h1, h2 { margin: 0; font-family: var(--cc-type-font_display); }
  h1:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); }
  p { margin: 0; }
  .notice { padding: var(--cc-space-sm); background: var(--cc-color-surface); }
  .activities { display: grid; grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr)); gap: var(--cc-space-md); margin: 0; padding: 0; list-style: none; }
  .activities li { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-md); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); background: var(--cc-color-surface); }
  progress { inline-size: 100%; accent-color: var(--cc-color-accent); }
  button { justify-self: start; }
</style>
