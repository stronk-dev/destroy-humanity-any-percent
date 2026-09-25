<script lang="ts">
  import { onMount } from "svelte";

  import type { GameUISnapshot } from "../api/generated/types";
  import { applicationCopyCatalog, t, type CopyEra, type CopyKey } from "../copy";
  import { canonicalString } from "../numeric";
  import Amount from "../ui/Amount.svelte";
  import type { ShellView } from "../shell/controller";
  import { installTheme, UI_THEMES } from "../ui/themes";
  import { eraForSnapshot, type ParsedGameUISnapshot } from "./contracts";
  import type { ExitOfferSpawnedEvent, RunEndedEvent } from "./events";
  import { GameUINavigation } from "./navigation";
  import { GAME_UI_PRESENTATION, requirePresentation, requirePresentationConstant } from "./presentation";
  import { renderPrestigeTermRows } from "./prestige-terms";
  import { createBrowserGameUIRuntime, newIntentID, type GameUIRuntime, type GameUIRuntimeMessage } from "./runtime";
  import { defaultSurface, type GameUISurfaceID } from "./surface-catalog";
  import { priorPersonalBest, readLocalTiming, RTATimer, writeLocalRunTiming, type LocalTimingStorage } from "./timing";
  import { formatAmount } from "../ui/amount-format";
  import RunEndSurface from "./RunEndSurface.svelte";
  import ReputationTreeSurface from "./ReputationTreeSurface.svelte";
  import ReputationPlanPanel from "./ReputationPlanPanel.svelte";
  import AchievementsSurface from "./AchievementsSurface.svelte";
  import FiscalSurface from "./FiscalSurface.svelte";
  import MetersSurface from "./MetersSurface.svelte";
  import MinigameSessionSurface from "./minigame/MinigameSessionSurface.svelte";
  import SoulRecoverySurface from "./soul/SoulRecoverySurface.svelte";
  import { loadSoulRecoveryContent } from "./soul/recovery-surface";
  import { GameUIShell } from "./shell-bridge";
  import { noticeForError, noticeForOutcome, type SurfaceRejections } from "./intent-outcome";
  import { FEATURES_PRESENTATION } from "./features-presentation";
  import type { GameUIAnnouncementEvent } from "./events";

  let { runtime = createBrowserGameUIRuntime(), timingStorage }: { runtime?: GameUIRuntime; timingStorage?: LocalTimingStorage } = $props();
  function localTimingStorage(): LocalTimingStorage { return timingStorage ?? localStorage; }
  let root: HTMLElement;
  function startupSurface(): GameUISurfaceID { return runtime.hasCredentials() ? "desk" : "vision_slide"; }
  const initialSurface = startupSurface();
  let snapshot = $state<ParsedGameUISnapshot | undefined>();
  let surface = $state<GameUISurfaceID>(initialSurface);
  let navigation = new GameUINavigation(initialSurface);
  let actionPending = $state(false);
  let refreshPending = $state(false);
  const pending = $derived(actionPending || refreshPending);
  let offline = $state(false);
  let draining = $state(false);
  let resyncing = $state(false);
  let transportReady = $state(false);
  let visitorCount = $state<number | undefined>();
  let founderRevision = $state<number | undefined>();
  let personalBestMS = $state<number | undefined>();
  let splits = $state<Readonly<{ gate_id: string; rta_ms: number }>[]>([]);
  let offer = $state<ExitOfferSpawnedEvent | undefined>();
  let ended = $state<RunEndedEvent | undefined>();
  let orderPlaced = $state(false);
  let intentNotice = $state<CopyKey | null>(null);
  // GS0.6: one polite chrome region for cross-surface announcements, deduped
  // by stream cursor so a replay after reconnect announces nothing.
  let announcement = $state("");
  let metersChanged = $state(false);
  const announcedCursors = new Set<string>();
  let monotonicMS = $state(0);
  let snapshotMonotonicMS = $state(0);
  let subscribedFounderID: string | undefined;
  let runIdentity: string | undefined;
  let timer = $state<RTATimer | undefined>();
  const shell = new GameUIShell(() => { void refresh(); });
  let shellView = $state<ShellView>(shell.view());
  let unsubscribeShell = () => {};
  let unsubscribe = () => {};
  let tick: ReturnType<typeof setInterval> | undefined;
  let refreshTask: Promise<void> | undefined;
  let actionTask: Promise<void> | undefined;
  let activeActionKind: string | undefined;

  const era = $derived<CopyEra>(snapshot ? eraForSnapshot(snapshot) : "era_1995");

  $effect(() => { if (root) installTheme(root, UI_THEMES[era], matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false); });

  onMount(() => {
    unsubscribeShell = shell.subscribe((value) => { shellView = value; });
    monotonicMS = performance.now();
    tick = setInterval(() => { monotonicMS = performance.now(); }, 100);
    if (runtime.hasCredentials()) startShell();
    return () => { if (tick) clearInterval(tick); unsubscribe(); unsubscribeShell(); shell.dispose(); };
  });

  function startShell(): void {
    shell.start();
  }

  function show(next: GameUISurfaceID): void {
    if (next === "meters") metersChanged = false;
    navigation.select(next);
    surface = navigation.active;
  }

  function bindSnapshot(value: ParsedGameUISnapshot): void {
    const sampledMonotonicMs = performance.now();
    if (snapshot === undefined) {
      const authoritativeDefault = defaultSurface(Object.fromEntries(value.facts.map((fact) => [fact.fact_id, fact.value])));
      navigation = new GameUINavigation(authoritativeDefault);
      surface = authoritativeDefault;
    }
    timer ??= new RTATimer({ serverNowMs: value.server_now_ms, runStartedAtMs: value.run.run_started_at_ms, sampledMonotonicMs });
    timer.resample({ serverNowMs: value.server_now_ms, runStartedAtMs: value.run.run_started_at_ms, sampledMonotonicMs });
    monotonicMS = sampledMonotonicMs;
    snapshotMonotonicMS = sampledMonotonicMs;
    snapshot = value;
    founderRevision = "founder_revision" in value ? value.founder_revision : undefined;
    shell.publish(value);
    const nextRunIdentity = `${value.run.founder_id}\0${value.run.run_seq}\0${value.run.category}`;
    if (runIdentity !== nextRunIdentity) {
      runIdentity = nextRunIdentity;
      splits = [];
      personalBestMS = priorPersonalBest(readLocalTiming(localTimingStorage(), value.run.founder_id), value.run.founder_id, value.run.run_seq, value.run.category);
    }
    if (subscribedFounderID !== value.run.founder_id) {
      unsubscribe();
      subscribedFounderID = value.run.founder_id;
      transportReady = false;
      unsubscribe = runtime.subscribe(value.run.founder_id, consumePublication);
    }
    offline = false;
  }

  function refresh(): Promise<void> {
    if (refreshTask) return refreshTask;
    refreshPending = true;
    refreshTask = (async () => {
      try { bindSnapshot(await runtime.snapshot()); }
      catch { offline = true; }
      finally { refreshPending = false; refreshTask = undefined; }
    })();
    return refreshTask;
  }

  async function beginAttempt(): Promise<void> {
    actionPending = true; offline = false;
    try { const value = await runtime.bootstrap(); startShell(); bindSnapshot(value); show("desk"); }
    catch { offline = true; }
    finally { actionPending = false; }
  }

  // GS0.2: `scope` binds expected_revision to the Company or Founder stream;
  // a rejected outcome renders its reason instead of looking like offline.
  async function act(body: Record<string, unknown>, options: Readonly<{ scope?: "company" | "founder"; rejections?: SurfaceRejections; applied?: (receipt: Readonly<Record<string, unknown>>) => CopyKey | null }> = {}): Promise<void> {
    if (!snapshot) return;
    if (options.scope === "founder" && founderRevision === undefined) return;
    const kind = typeof body.kind === "string" ? body.kind : "";
    if (actionTask) {
      if (activeActionKind === kind) return;
      await actionTask;
    }
    // A click that raced an ordered event/receipt refresh must not disappear.
    // Finish that authoritative refresh, then bind the intent to its revision.
    if (refreshTask) await refreshTask;
    if (!snapshot || actionTask) return;
    actionPending = true;
    activeActionKind = kind;
    const task = (async () => {
      try {
        intentNotice = null;
        const expected = options.scope === "founder" ? founderRevision! : snapshot!.revision;
        const outcome = await runtime.intent({ intent_id: newIntentID(), expected_revision: expected, ...body });
        const notice = noticeForOutcome(outcome, options.rejections);
        intentNotice = outcome.outcome === "applied" && options.applied ? options.applied(outcome.receipt) : notice.notice;
        if (notice.effect === "refresh") void refresh();
        else if (outcome.outcome === "applied" && (kind === "cross_gate" || kind === "decline_exit_offer")) {
          bindSnapshot(await runtime.snapshot());
        }
      } catch (error) {
        const notice = noticeForError(error);
        intentNotice = notice.notice;
        if (notice.effect === "offline") offline = true;
        else if (notice.effect === "refresh") void refresh();
      }
      finally {
        actionTask = undefined;
        activeActionKind = undefined;
        actionPending = false;
      }
    })();
    actionTask = task;
    return task;
  }

  // R6: the advisory Exit plan; the key is omitted when empty so existing
  // requests stay byte-identical.
  let exitPlan = $state<string[]>([]);
  function withPlan(body: Record<string, unknown>): Record<string, unknown> {
    return exitPlan.length === 0 ? body : { ...body, reputation_plan: [...exitPlan] };
  }
  function acceptOffer(): void {
    if (!offer || founderRevision === undefined) return;
    void act(withPlan({ kind: "accept_exit_offer", expected_founder_revision: founderRevision, offer_id: offer.payload.offer_id }));
  }

  async function continueRun(): Promise<void> {
    if (!ended || pending) return;
    const expectedRunSeq = ended.payload.run_id.run_seq + 1;
    actionPending = true;
    try {
      const value = await runtime.snapshot();
      if (value.run.run_seq !== expectedRunSeq) throw new RangeError("next Company snapshot did not advance exactly one run");
      bindSnapshot(value);
      ended = undefined;
      offer = undefined;
      exitPlan = [];
      show("desk");
    } catch { offline = true; }
    finally { actionPending = false; }
  }

  function consumePublication(message: GameUIRuntimeMessage): void {
    if (message.kind === "transport_closed") { offline = true; transportReady = false; subscribedFounderID = undefined; unsubscribe(); return; }
    if (message.kind === "transport_recovered") { offline = false; transportReady = true; draining = false; resyncing = false; return; }
    if (message.kind === "snapshot") { bindSnapshot(message.value); draining = false; resyncing = false; return; }
    if (message.kind === "historical_event") return;
    // A terminal command persists run_ended and run_started in one ordered
    // transaction. Refreshing its trailing receipt would bind the already
    // created next run and let its lifecycle preempt the run-end screen before
    // the player explicitly continues.
    if (message.kind === "receipt") { if (!ended) void refresh(); return; }
    if (message.kind === "presence") { visitorCount = message.count; return; }
    if (message.kind === "announcement") { announce(message.scope, message.value); return; }
    if (message.kind === "system") {
      transportReady = false;
      if (message.value.kind === "server_restarting") draining = true;
      else resyncing = true;
      return;
    }
    if (message.kind !== "event") return;
    if (message.scope === "founder") founderRevision = Math.max(founderRevision ?? 0, message.revision);
    const value = message.value;
    if (message.scope === "company" && value.kind === "gate_crossed" && snapshot && value.payload.founder_id === snapshot.run.founder_id && value.payload.run_id.run_seq === snapshot.run.run_seq) {
      const split = { gate_id: value.payload.gate_id, rta_ms: Math.max(0, value.occurred_at_ms - snapshot.run.run_started_at_ms) };
      splits = [...splits.filter((row) => row.gate_id !== split.gate_id), split].sort((left, right) => left.gate_id < right.gate_id ? -1 : left.gate_id > right.gate_id ? 1 : 0);
      // Gate eligibility changes immediately. Refresh from the ordered event so
      // the next control does not depend on a separate receipt publication.
      void refresh();
    } else if (value.kind === "exit_offer_spawned" && !ended) {
      offer = value; navigation.lifecycle({ cursor: value.cursor, surface: "offer_sheet" }); surface = navigation.active;
    } else if (value.kind === "run_ended") {
      ended = value;
      timer?.terminal(value.payload.rta_ms);
      if (snapshot) writeLocalRunTiming(localTimingStorage(), { category: snapshot.run.category, founder_id: value.payload.founder_id, pb_rta_ms: value.payload.rta_ms, run_seq: value.payload.run_id.run_seq, splits });
      navigation.lifecycle({ cursor: value.cursor, surface: "run_end" }); surface = navigation.active;
    } else if (value.kind === "exit_offer_resolved" && value.payload.offer_id === offer?.payload.offer_id) offer = undefined;
  }

  function estimatedServerNowMS(): number {
    if (!snapshot) return 0;
    return snapshot.server_now_ms + Math.max(0, monotonicMS - snapshotMonotonicMS);
  }

  function visibleManualTokensMilli(): number {
    if (!snapshot) return 0;
    const elapsed = Math.max(0, estimatedServerNowMS() - snapshot.manual_action.refilled_at_ms);
    return Math.min(snapshot.manual_action.bucket_cap_milli, snapshot.manual_action.tokens_milli + Math.floor(elapsed) * snapshot.manual_action.refill_milli_per_ms);
  }

  function evaluationDay(): number {
    if (!snapshot) return 1;
    return Math.min(40, Math.max(1, Math.floor((estimatedServerNowMS() - snapshot.run.run_started_at_ms) / 86_400_000) + 1));
  }

  function duration(ms: number): string {
    const seconds = Math.max(0, Math.floor(ms / 1000));
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor(seconds % 3600 / 60);
    const remainder = seconds % 60;
    return `${hours}:${minutes.toString().padStart(2, "0")}:${remainder.toString().padStart(2, "0")}`;
  }

  function titleForCategory(category: string): string {
    const keys = { any_percent: "category.any_percent", ethical_percent: "category.ethical_percent", hundred_percent: "category.hundred_percent", low_percent: "category.low_percent", valuation: "category.valuation" } as const;
    const key = keys[category as keyof typeof keys];
    if (!key) throw new RangeError(`missing category presentation for ${category}`);
    return t(key, {}, era);
  }

  function meterLabel(meterID: string): string | undefined {
    if (meterID === FEATURES_PRESENTATION.doomMeter.meter_id) return t(FEATURES_PRESENTATION.doomMeter.title_key, {}, era);
    const row = FEATURES_PRESENTATION.trustMeters.get(meterID);
    return row ? t("meters.row_frame", { constituency: t(row.constituency_key, {}, era), axis: t(row.axis_key, {}, era) }, era) : undefined;
  }

  function announce(scope: "company" | "founder", value: GameUIAnnouncementEvent): void {
    const key = `${scope}\0${value.cursor}`;
    if (announcedCursors.has(key)) return;
    announcedCursors.add(key);
    if (value.kind === "achievement_earned") {
      const row = liveFeatures?.achievements?.rows.find((candidate) => candidate.achievement_id === value.payload.achievement_id);
      if (!row || !applicationCopyCatalog.byKey.has(row.copy_key)) { console.error(`game UI invariant: unannounceable achievement ${value.payload.achievement_id}`); return; }
      announcement = t("achievements.earned_announcement", { achievement: t(row.copy_key as CopyKey, {}, era) }, era);
      return;
    }
    const meter = meterLabel(value.payload.meter_id), band = FEATURES_PRESENTATION.meterBands.get(value.payload.to_band);
    if (!meter || !band) { console.error(`game UI invariant: unannounceable meter change ${value.payload.meter_id}`); return; }
    if (surface === "meters") announcement = t("meters.band_changed_announcement", { meter, band: t(band, {}, era) }, era);
    else metersChanged = true;
  }

  function factTrue(id: string): boolean { return snapshot?.facts.some((fact) => fact.fact_id === id && fact.value === true) ?? false; }
  const liveFeatures = $derived(snapshot && "features" in snapshot ? snapshot.features : undefined);
  const pitchAvailability = $derived(liveFeatures?.minigames?.rows.find((row) => row.minigame_id === "pitch"));
  const FISCAL_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/period_not_ripe", "fiscal.rejection.period_not_ripe"],
    ["unaffordable/fiscal_credit", "fiscal.rejection.unaffordable"],
    ["not_eligible/already_unlocked", "fiscal.rejection.already_unlocked"],
  ]);
  const REPUTATION_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/*", "reputation_tree.error.not_eligible"],
    ["unknown_id/*", "reputation_tree.error.unknown_id"],
    ["unaffordable/reputation", "reputation_tree.error.unaffordable"],
    ["invalid/*", "reputation_tree.error.invalid"],
  ]);
  function reputationApplied(): CopyKey { void refresh(); return "reputation_tree.result.applied"; }
  const HARVEST_OUTCOMES: Readonly<Record<string, CopyKey>> = {
    consumed_by_auto: "fiscal.outcome.consumed_by_auto", early_failed: "fiscal.outcome.early_failed",
    early_succeeded: "fiscal.outcome.early_succeeded", guaranteed: "fiscal.outcome.guaranteed",
  };
  function harvestNotice(receipt: Readonly<Record<string, unknown>>): CopyKey | null {
    const key = typeof receipt.harvest_outcome === "string" ? HARVEST_OUTCOMES[receipt.harvest_outcome] : undefined;
    if (!key) { console.error("game UI invariant: unknown Fiscal harvest outcome"); return "intent.rejection.unknown"; }
    return key;
  }
  // GS0.5: an arm going null while its surface is mounted returns to the Desk.
  $effect(() => {
    const armless = surface === "achievements" && !liveFeatures?.achievements || surface === "fiscal" && !liveFeatures?.fiscal ||
      surface === "meters" && !liveFeatures?.meters || surface === "reputation_tree" && !liveFeatures?.reputation;
    if (snapshot && armless) show("desk");
  });
  const founderControls = $derived(founderRevision !== undefined && !offline && !resyncing);

  function exitTitle(exitType: string): string { return t(requirePresentation(GAME_UI_PRESENTATION.exitTypes, exitType).title_key, {}, era); }

  function capFor(cap: GameUISnapshot["resources"][number]["cap"]): { amount: string; reason_key: CopyKey } | undefined {
    if (cap === null) return undefined;
    // F10: a missing reason key must not take the Desk down. The cap is
    // withheld and one invariant is reported loudly (tests assert it).
    if (!applicationCopyCatalog.byKey.has(cap.reason_key)) { console.error(`game UI invariant: missing cap copy ${cap.reason_key}`); return undefined; }
    return { amount: cap.amount, reason_key: cap.reason_key as CopyKey };
  }

  function visibleResourceAmount(resource: GameUISnapshot["resources"][number]): string {
    const value = shellView.resources[resource.resource_id]?.value;
    return value === undefined ? resource.amount : canonicalString(value);
  }

  export function fixtureSnapshot(value: ParsedGameUISnapshot): void { bindSnapshot(value); }
  export function fixtureSurface(value: GameUISurfaceID): void { show(value); }
  export function fixtureOffer(value: ExitOfferSpawnedEvent): void { offer = value; show("offer_sheet"); }
  export function fixtureRunEnd(value: RunEndedEvent): void { ended = value; show("run_end"); }
  export function fixtureSystem(value: "drain" | "resync"): void { if (value === "drain") draining = true; else resyncing = true; }
  export function fixtureMonotonicElapsed(value: number): void { monotonicMS = snapshotMonotonicMS + value; }
</script>

<main bind:this={root} class="game-ui" data-surface={surface} aria-busy={pending}>
  {#if snapshot}
    <header class="chrome cc-window">
      <div class="cc-titlebar">
        <strong>{t("chrome.run_title.company_fallback", {}, era)}</strong>
        <span>{titleForCategory(snapshot.run.category)}</span>
        <span>{t("screen.vision_slide.timer_frame", { rta: duration(timer?.elapsed(monotonicMS) ?? 0), pb: personalBestMS === undefined ? t("chrome.run_title.pb_empty", {}, era) : duration(personalBestMS) }, era)}</span>
        <span>{t("chrome.run_title.tier_frame", { tier: snapshot.run.tier }, era)}</span>
      </div>
      <nav aria-label={t("surface.desk.title", {}, era)}>
        <button type="button" aria-current={surface === "desk" ? "page" : undefined} onclick={() => show("desk")}>{t("surface.desk.title", {}, era)}</button>
        {#if factTrue("feature.achievements")}<button type="button" aria-current={surface === "achievements" ? "page" : undefined} onclick={() => show("achievements")}>{t("surface.achievements.title", {}, era)}</button>{/if}
        {#if factTrue("feature.fiscal")}<button type="button" aria-current={surface === "fiscal" ? "page" : undefined} onclick={() => show("fiscal")}>{t("surface.fiscal.title", {}, era)}</button>{/if}
        {#if factTrue("feature.meters")}<button type="button" aria-current={surface === "meters" ? "page" : undefined} onclick={() => show("meters")}>{t("surface.meters.title", {}, era)}{#if metersChanged} {t("meters.nav_changed_badge", {}, era)}{/if}</button>{/if}
        {#if runtime.minigame && factTrue("feature.minigame.pitch")}<button type="button" aria-current={surface === "minigame_session" ? "page" : undefined} onclick={() => show("minigame_session")}>{t("minigame.pitch.title", {}, era)}</button>{/if}
        {#if runtime.soulRecovery}<button type="button" aria-current={surface === "soul_recovery" ? "page" : undefined} onclick={() => show("soul_recovery")}>{t("soul.recovery_surface.title", {}, era)}</button>{/if}
        {#if factTrue("feature.reputation_tree")}<button type="button" aria-current={surface === "reputation_tree" ? "page" : undefined} onclick={() => show("reputation_tree")}>{t("reputation_tree.title", {}, era)}</button>{/if}
        <button type="button" aria-current={surface === "settings" ? "page" : undefined} onclick={() => show("settings")}>{t("surface.settings.title", {}, era)}</button>
      </nav>
      {#if snapshot.run.run_seq === 1 && visitorCount !== undefined}<span class="visitor" title={t("chrome.visitor_counter.tooltip", {}, era)}>{t("chrome.visitor_counter.frame", { count: visitorCount }, era)}</span>{/if}
    </header>
  {/if}

  {#if snapshot}<p class="announcement" role="status">{announcement}</p>{/if}
  {#if snapshot}<p class="intent-notice" role="status">{intentNotice ? t(intentNotice, {}, era) : ""}</p>{/if}
  {#if draining}
    <aside class="notice" role="status"><strong>{t("system.drain_notice.title", {}, era)}</strong><span>{t("system.drain_notice.body", {}, era)}</span></aside>
  {/if}
  {#if resyncing}
    <aside class="notice" role="alert"><strong>{t("system.resync.title", {}, era)}</strong><span>{t("system.resync.body", {}, era)}</span><button type="button" onclick={() => { resyncing = false; void refresh(); }}>{t("system.resync.continue", {}, era)}</button></aside>
  {/if}

  {#if surface === "vision_slide"}
    <section class="surface vision" aria-labelledby="vision-heading">
      <article class="slide cc-window">
        <h1 id="vision-heading">{t("screen.vision_slide.slide_heading", {}, era)}</h1>
        <p>{t("screen.vision_slide.slide_body", {}, era)}</p>
        <small>{t("screen.vision_slide.slide_footnote", {}, era)}</small>
      </article>
      <article class="contract cc-window">
        <h2>{t("screen.vision_slide.contract_title", {}, era)}</h2>
        <p>{t("screen.vision_slide.contract_category", {}, era)}</p>
        <p>{t("screen.vision_slide.timer_frame", { rta: duration(0), pb: t("chrome.run_title.pb_empty", {}, era) }, era)}</p>
        <button type="button" disabled={pending} onclick={beginAttempt}>{pending ? t("screen.vision_slide.connecting", {}, era) : t("screen.vision_slide.begin_attempt", {}, era)}</button>
        {#if offline}<p role="alert">{t("screen.vision_slide.offline_fallback", {}, era)}</p><button type="button" onclick={beginAttempt}>{t("screen.vision_slide.retry", {}, era)}</button>{/if}
        <small>{t("screen.vision_slide.small_print", {}, era)}</small>
      </article>
    </section>
  {:else if snapshot && surface === "desk"}
    <section class="surface desk" aria-labelledby="desk-heading">
      <h1 id="desk-heading">{t("surface.desk.title", {}, era)}</h1>
      <section class="manual cc-window">
        <h2>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).title_key, {}, era)}</h2>
        <p>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).description_key, {}, era)}</p>
        <label>{t("desk.manual.meter_label", {}, era)} <meter min="0" max={snapshot.manual_action.bucket_cap_milli} value={visibleManualTokensMilli()}></meter></label>
        <output>{t("desk.manual.meter_frame", { current: Math.floor(visibleManualTokensMilli() / 1000), cap: Math.floor(snapshot.manual_action.bucket_cap_milli / 1000) }, era)}</output>
        <button type="button" disabled={pending} title={t("desk.manual.meter_tooltip", {}, era)} onclick={() => act({ kind: "perform_manual_batch", action_id: snapshot!.manual_action.action_id, count: 1, window_ms: 1 })}>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).title_key, {}, era)}</button>
      </section>

      <section aria-labelledby="resources-heading">
        <h2 id="resources-heading">{t("desk.capped_label", {}, era)}</h2>
        <div class="cards">
          {#each snapshot.resources as resource (resource.resource_id)}
            <article class="card"><Amount value={visibleResourceAmount(resource)} cap={capFor(resource.cap)} {era} /><span>{t("desk.rate_frame", { rate: formatAmount(resource.rate_per_second) }, era)}</span></article>
          {/each}
        </div>
      </section>

      <section aria-labelledby="generators-heading">
        <h2 id="generators-heading">{t("desk.generators_label", {}, era)}</h2>
        <div class="cards">
          {#each snapshot.generators as generator (generator.generator_id)}
            {@const presentation = requirePresentation(GAME_UI_PRESENTATION.generators, generator.generator_id)}
            <article class="card">
              <h3>{t(presentation.title_key, {}, era)}</h3><p>{t(presentation.description_key, {}, era)}</p>
              <span>{t("desk.owned_frame", { count: generator.owned }, era)}</span><span>{t("desk.rate_frame", { rate: formatAmount(generator.rate_contribution) }, era)}</span>
              <Amount value={generator.next_cost} era={era} />
              {#if generator.provisioned > 0}<span>{t("desk.provisioned_frame", { count: generator.provisioned }, era)}</span>{/if}
              {#if "provision_cap" in generator && generator.provision_cap !== null && generator.provisioned >= generator.provision_cap.amount}<span>{t(generator.provision_cap.reason_key as CopyKey, {}, era)}</span>{/if}
              <div><button type="button" disabled={pending || generator.max_affordable < 1} onclick={() => act({ kind: "buy_generator", generator_id: generator.generator_id, count: { mode: "exact", value: 1 } })}>{t("desk.buy_one", {}, era)}</button><button type="button" disabled={pending || generator.max_affordable < 1} onclick={() => act({ kind: "buy_generator", generator_id: generator.generator_id, count: { mode: "max" } })}>{t("desk.buy_max", {}, era)}</button></div>
            </article>
          {/each}
        </div>
      </section>

      <section aria-labelledby="upgrades-heading">
        <h2 id="upgrades-heading">{t("desk.upgrades_label", {}, era)}</h2>
        <div class="cards">
          {#each snapshot.upgrades as upgrade (upgrade.upgrade_id)}
            {@const presentation = requirePresentation(GAME_UI_PRESENTATION.upgrades, upgrade.upgrade_id)}
            <article class="card"><h3>{t(presentation.title_key, {}, era)}</h3><p>{t(presentation.description_key, {}, era)}</p><Amount value={upgrade.cost_amount} era={era} />{#if upgrade.owned}<strong>{t("desk.upgrade.owned", {}, era)}</strong>{/if}<button type="button" disabled={pending || !upgrade.eligible || upgrade.owned} onclick={() => act({ kind: "buy_upgrade", upgrade_id: upgrade.upgrade_id })}>{t("desk.buy_one", {}, era)}</button></article>
          {/each}
        </div>
      </section>

      {#if splits.length}
        <details><summary>{t("chrome.splits.label", {}, era)}</summary>{#each splits as split}<p>{t(requirePresentation(GAME_UI_PRESENTATION.gates, split.gate_id).title_key, {}, era)} {duration(split.rta_ms)}</p>{/each}{#if personalBestMS === undefined}<p>{t("chrome.splits.first_attempt_note", {}, era)}</p>{/if}</details>
      {/if}

      <section class="card"><h2>{t("cosmetic.horse_armor_free.title", {}, era)}</h2><p>{t("cosmetic.horse_armor_free.description", { price: requirePresentationConstant("constant.price_zero") }, era)}</p><small>{t("cosmetic.horse_armor_free.disclosure", {}, era)}</small></section>
      {#if snapshot.schema_version >= 3 && "transitions" in snapshot}
        {@const transitions = snapshot.transitions}
        <section class="card">
          {#if transitions.cross_gate}
            <button type="button" disabled={pending || !transitions.cross_gate.eligible} onclick={() => act({ kind: "cross_gate", gate_id: transitions.cross_gate!.gate_id, route_id: null })}>{t("desk.cross_gate", {}, era)}</button>
          {/if}
          <button type="button" disabled={pending || !transportReady || !transitions.wind_down.eligible || founderRevision === undefined} onclick={() => act(withPlan({ kind: "wind_down", expected_founder_revision: founderRevision }))}>{t("desk.wind_down", {}, era)}</button>
          {#if liveFeatures?.reputation && transitions.wind_down.eligible}<ReputationPlanPanel arm={liveFeatures.reputation} {era} previewDelta={0} onChange={(plan) => { exitPlan = [...plan]; }} />{/if}
        </section>
      {/if}
      {#if era === "era_1995"}
        <section class="card" title={t("satire.unregistered.tooltip", {}, era)}><h2>{t("satire.unregistered.titlebar_frame", { day: evaluationDay() }, era)}</h2></section>
        <section class="card" aria-labelledby="order-heading">
          <h2 id="order-heading">{t("satire.order_form.window_title", {}, era)}</h2>
          <h3>{t("satire.order_form.heading", {}, era)}</h3>
          <p>{t("satire.order_form.item_full_version", { price: requirePresentationConstant("constant.price_zero") }, era)}</p>
          <p>{t("satire.order_form.item_shipping", { price: requirePresentationConstant("constant.price_zero") }, era)}</p>
          <p>{t("satire.order_form.item_site_license", { price: requirePresentationConstant("constant.price_zero") }, era)}</p>
          <strong>{t("satire.order_form.total", { price: requirePresentationConstant("constant.price_zero") }, era)}</strong>
          <button type="button" onclick={() => { orderPlaced = true; }}>{t("satire.order_form.place_order", {}, era)}</button>
          {#if orderPlaced}<p role="status">{t("satire.order_form.confirmation", { founder: requirePresentationConstant("constant.founder_fallback"), price: requirePresentationConstant("constant.price_zero") }, era)}</p>{/if}
          <small>{t("satire.order_form.small_print", {}, era)}</small>
        </section>
      {/if}
      <section class="readme"><h2>{t("satire.readme.window_title", {}, era)}</h2><pre>{t("satire.readme.body", {}, era)}</pre></section>
    </section>
  {:else if surface === "offer_sheet" && offer}
    <section class="surface" aria-labelledby="offer-heading">
      <h1 id="offer-heading">{t("screen.offer_sheet.heading", {}, era)}</h1><p>{t("screen.offer_sheet.preamble", {}, era)}</p><h2>{t("screen.offer_sheet.terms_label", {}, era)}</h2>
      <p>{exitTitle(offer.payload.exit_type)}</p>
      {#each renderPrestigeTermRows(offer.payload.payout_preview, era) as row}<p>{row}</p>{/each}
      <p title={t("screen.offer_sheet.countdown_tooltip", {}, era)}>{t("screen.offer_sheet.countdown_frame", { remaining: duration(offer.payload.expires_at_ms - estimatedServerNowMS()) }, era)}</p>
      {#if liveFeatures?.reputation}<ReputationPlanPanel arm={liveFeatures.reputation} {era} previewDelta={offer.payload.payout_preview.reputation_delta} onChange={(plan) => { exitPlan = [...plan]; }} />{/if}
      <button type="button" disabled={pending || !transportReady || founderRevision === undefined} onclick={acceptOffer}>{t("screen.offer_sheet.accept", {}, era)}</button>
      <button type="button" disabled={pending} onclick={() => act({ kind: "decline_exit_offer", offer_id: offer!.payload.offer_id })}>{t("screen.offer_sheet.decline", {}, era)}</button>
    </section>
  {:else if surface === "run_end" && ended}
    <RunEndSurface {ended} />
    <button type="button" disabled={pending} onclick={continueRun}>{t("screen.run_end.continue", {}, era)}</button>
    {#if offline}<p role="alert">{t("settings.save_status.offline", {}, era)}</p>{/if}
  {:else if surface === "achievements" && liveFeatures?.achievements}
    <AchievementsSurface arm={liveFeatures.achievements} {era} />
  {:else if surface === "meters" && liveFeatures?.meters}
    <MetersSurface arm={liveFeatures.meters} {era} />
  {:else if surface === "reputation_tree" && liveFeatures?.reputation}
    <ReputationTreeSurface arm={liveFeatures.reputation} {era} {pending} controlsEnabled={founderControls}
      onPurchase={(nodeID) => act({ kind: "purchase_reputation_node", node_id: nodeID }, { scope: "founder", rejections: REPUTATION_REJECTIONS, applied: reputationApplied })} />
  {:else if surface === "fiscal" && liveFeatures?.fiscal}
    <FiscalSurface arm={liveFeatures.fiscal} {era} serverNowMs={estimatedServerNowMS()} {pending} controlsEnabled={founderControls}
      onHarvest={() => act({ kind: "harvest_fiscal_period" }, { scope: "founder", rejections: FISCAL_REJECTIONS, applied: harvestNotice })}
      onSpendLevel={(generatorID) => act({ kind: "spend_fiscal_credit", target: { kind: "generator_level", generator_id: generatorID, levels: 1 } }, { scope: "founder", rejections: FISCAL_REJECTIONS })}
      onSpendUnlock={(unlockID) => act({ kind: "spend_fiscal_credit", target: { kind: "unlock", unlock_id: unlockID } }, { scope: "founder", rejections: FISCAL_REJECTIONS })} />
  {:else if snapshot && surface === "minigame_session" && runtime.minigame}
    {#if pitchAvailability && !pitchAvailability.unlocked}<p class="intent-notice" role="note">{t("minigame.availability.fiscal_locked", {}, era)}</p>{/if}
    {#if pitchAvailability?.human_content_locked}<p class="intent-notice" role="note">{t("minigame.availability.soul_locked", {}, era)}</p>{/if}
    <MinigameSessionSurface port={runtime.minigame} minigameID="pitch" {era} newCommandID={() => newIntentID()} onExitToHost={() => show("desk")} onTerminal={() => { void refresh(); }} />
  {:else if snapshot && surface === "soul_recovery" && runtime.soulRecovery}
    <SoulRecoverySurface port={runtime.soulRecovery} content={loadSoulRecoveryContent()} {era} onExitToHost={() => show("desk")} onTerminal={() => { void refresh(); }} />
  {:else if snapshot && surface === "settings"}
    <section class="surface" aria-labelledby="settings-heading"><h1 id="settings-heading">{t("surface.settings.title", {}, era)}</h1><p>{offline ? t("settings.save_status.offline", {}, era) : pending ? t("settings.save_status.saving", {}, era) : t("settings.save_status.saved_frame", { ago: duration(Math.max(0, monotonicMS - snapshotMonotonicMS)) }, era)}</p><p>{t("settings.account_note", {}, era)}</p></section>
  {/if}
</main>

<style>
  .game-ui { min-height: 100vh; padding: var(--cc-space-lg); color: var(--cc-color-text); background: var(--cc-color-bg); font-family: var(--cc-type-font_ui); font-size: var(--cc-type-size_base); line-height: var(--cc-type-line_height); }
  .cc-window, .surface, .card, .notice, .readme { border: var(--cc-border-width) var(--cc-border-style) var(--cc-chrome-window_border); border-radius: var(--cc-border-radius); background: var(--cc-chrome-window_bg); }
  .chrome { display: grid; gap: var(--cc-space-sm); margin-block-end: var(--cc-space-lg); }
  .cc-titlebar { display: flex; flex-wrap: wrap; gap: var(--cc-space-md); padding: var(--cc-space-sm) var(--cc-space-md); color: var(--cc-chrome-titlebar_text); background: var(--cc-chrome-titlebar_bg); font-family: var(--cc-type-font_display); font-weight: var(--cc-type-weight_bold); }
  nav, .visitor { display: flex; gap: var(--cc-space-sm); padding: 0 var(--cc-space-md) var(--cc-space-sm); }
  .surface { display: grid; gap: var(--cc-space-lg); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); }
  .vision { grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr)); }
  .slide, .contract, .card, .notice, .readme { padding: var(--cc-space-lg); }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr)); gap: var(--cc-space-md); }
  .card { display: grid; gap: var(--cc-space-sm); }
  .manual { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-lg); }
  .notice { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); margin-block-end: var(--cc-space-md); color: var(--cc-color-text); background: var(--cc-color-surface); }
  button { padding: var(--cc-space-sm) var(--cc-space-md); color: var(--cc-color-text); background: var(--cc-chrome-button_face); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); font-family: var(--cc-type-font_ui); font-size: inherit; cursor: pointer; }
  button:focus-visible, summary:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  button[disabled] { color: var(--cc-color-text_muted); cursor: default; }
  h1, h2, h3, p { margin: 0; }
  h1, h2, h3 { font-family: var(--cc-type-font_display); }
  pre { overflow: auto; color: var(--cc-color-text); background: var(--cc-color-surface); font-family: var(--cc-type-font_mono); white-space: pre-wrap; }
  meter { inline-size: 100%; accent-color: var(--cc-color-accent); }
</style>
