<script lang="ts">
  import { onMount, tick as afterDOMUpdate } from "svelte";

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
  import AdoptionCard from "./pet/AdoptionCard.svelte";
  import PetCareSurface from "./pet/PetCareSurface.svelte";
  import CosmeticShelf from "./cosmetics/CosmeticShelf.svelte";
  import { COSMETIC_SHOP_PRESENTATION } from "./cosmetics/presentation";
  import ReputationPlanPanel from "./ReputationPlanPanel.svelte";
  import AchievementsSurface from "./AchievementsSurface.svelte";
  import FiscalSurface from "./FiscalSurface.svelte";
  import MetersSurface from "./MetersSurface.svelte";
  import MinigameSessionSurface from "./minigame/MinigameSessionSurface.svelte";
  import SoulRecoverySurface from "./soul/SoulRecoverySurface.svelte";
  import GardenSurface from "./garden/GardenSurface.svelte";
  import type { GardenPort } from "./garden/garden-port";
  import { loadSoulRecoveryContent } from "./soul/recovery-surface";
  import { GameUIShell } from "./shell-bridge";
  import { GameUIRequestError, noticeForError, noticeForOutcome, type IntentOutcome, type SurfaceRejections } from "./intent-outcome";
  import { FEATURES_PRESENTATION } from "./features-presentation";
  import { upgradePresentation } from "./axis-presentation";
  import AxisStackPanel from "./AxisStackPanel.svelte";
  import OpportunityRegion from "./OpportunityRegion.svelte";
  import { lastClaimFromReceipt, OPPORTUNITY_REJECTIONS, type LastClaim } from "./opportunity-claim";
  import type { GameUIAnnouncementEvent } from "./events";

  let { runtime = createBrowserGameUIRuntime(), timingStorage }: { runtime?: GameUIRuntime; timingStorage?: LocalTimingStorage } = $props();
  function localTimingStorage(): LocalTimingStorage { return timingStorage ?? localStorage; }
  let root: HTMLElement;
  const initialReducedMotion = typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches;
  let prefersReducedMotion = $state(initialReducedMotion);
  function startupSurface(): GameUISurfaceID { return runtime.hasCredentials() ? "desk" : "vision_slide"; }
  const initialSurface = startupSurface();
  let snapshot = $state<ParsedGameUISnapshot | undefined>();
  let surface = $state<GameUISurfaceID>(initialSurface);
  let navigation = new GameUINavigation(initialSurface);
  let selectionGeneration = 0;
  let actionPending = $state(false);
  let refreshPending = $state(false);
  const pending = $derived(actionPending || refreshPending);
  let offline = $state(false);
  // Socket readiness cannot prove that a failed HTTP read/command left our
  // revisions current. Only a successfully bound authoritative snapshot can.
  let needsAuthoritativeSnapshot = false;
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
  let energyRefilled = $state(false);
  let intentNotice = $state<CopyKey | null>(null);
  let intentNoticeOwner = $state<GameUISurfaceID | undefined>();
  // GS0.6: one polite chrome region for cross-surface announcements, deduped
  // by event identity at its stream cursor, not revision alone: one commit
  // can contain distinct Fiscal and care/achievement events.
  let announcement = $state("");
  let metersChanged = $state(false);
  // OD-3: a Fiscal harvest while elsewhere badges the Fiscal nav (no modal).
  let fiscalHarvested = $state(false);
  const announcedCursors = new Set<string>();
  let monotonicMS = $state(0);
  let snapshotMonotonicMS = $state(0);
  let subscribedFounderID: string | undefined;
  let runIdentity: string | undefined;
  let timer = $state<RTATimer | undefined>();
  const shell = new GameUIShell(() => { void refresh(); }, initialReducedMotion);
  let shellView = $state<ShellView>(shell.view());
  let unsubscribeShell = () => {};
  let unsubscribe = () => {};
  let tick: ReturnType<typeof setInterval> | undefined;
  // HTTP work can finish after component cleanup. It must not recreate the
  // socket/Worker or submit a command that was still waiting in this host.
  let disposed = false;
  const focusObservers = new Set<(event: FocusEvent) => void>();
  let refreshTask: Promise<boolean> | undefined;
  // Bounded to one local command. Unknown/older receipts remain conservative.
  let lastRefreshedIntentID: string | undefined;
  let actionTask: Promise<void> | undefined;
  let activeActionKind: string | undefined;
  const generatorPending = $derived(pending && activeActionKind === "buy_generator");
  const upgradePending = $derived(pending && activeActionKind === "buy_upgrade");
  const purchasePending = $derived(generatorPending || upgradePending);
  const gatePending = $derived(pending && activeActionKind === "cross_gate");
  const incorporatePending = $derived(pending && activeActionKind === "incorporate");
  const windDownPending = $derived(pending && activeActionKind === "wind_down");
  const deskPending = $derived(purchasePending || gatePending || incorporatePending || windDownPending);

  const era = $derived<CopyEra>(snapshot ? eraForSnapshot(snapshot) : "era_1995");

  $effect(() => { if (root) installTheme(root, UI_THEMES[era], prefersReducedMotion); });

  onMount(() => {
    const motion = typeof matchMedia === "function" ? matchMedia("(prefers-reduced-motion: reduce)") : undefined;
    const updateMotion = () => {
      prefersReducedMotion = motion?.matches ?? false;
      shell.setReducedMotion(prefersReducedMotion);
    };
    updateMotion();
    motion?.addEventListener("change", updateMotion);
    unsubscribeShell = shell.subscribe((value) => { shellView = value; });
    monotonicMS = performance.now();
    tick = setInterval(() => { monotonicMS = performance.now(); }, 100);
    if (runtime.hasCredentials()) startShell();
    return () => {
      disposed = true;
      for (const observer of focusObservers) document.removeEventListener("focusin", observer);
      focusObservers.clear();
      motion?.removeEventListener("change", updateMotion);
      if (tick) clearInterval(tick);
      unsubscribe(); unsubscribeShell(); shell.dispose();
    };
  });

  function observeDocumentFocus(observer: (event: FocusEvent) => void): () => void {
    focusObservers.add(observer);
    document.addEventListener("focusin", observer);
    return () => { focusObservers.delete(observer); document.removeEventListener("focusin", observer); };
  }

  function startShell(): void {
    if (disposed) return;
    shell.start();
  }

  function show(next: GameUISurfaceID): void {
    if (disposed) return;
    selectionGeneration += 1;
    if (next === "meters") metersChanged = false;
    if (next === "fiscal") fiscalHarvested = false;
    navigation.select(next);
    surface = navigation.active;
  }

  function bindSnapshot(value: ParsedGameUISnapshot): void {
    if (disposed) return;
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
    const nextGardenContext = `${value.run.founder_id}\0${value.constants_hash}`;
    if (gardenContext !== nextGardenContext) {
      const origin = document.activeElement;
      const ownedFocus = origin instanceof HTMLElement && surface === "garden" &&
        (root?.querySelector(".garden")?.contains(origin) || origin.matches('nav button[aria-current="page"]'));
      const selected = selectionGeneration;
      gardenContext = nextGardenContext;
      gardenPresence = "unknown";
      void probeGarden();
      if (ownedFocus) void afterDOMUpdate().then(() => {
        if (disposed || gardenContext !== nextGardenContext || surface !== "garden" ||
            selectionGeneration !== selected || origin.isConnected ||
            document.activeElement !== origin && document.activeElement !== document.body) return;
        root?.querySelector<HTMLElement>("#garden-heading")?.focus();
      });
    }
    founderRevision = "founder_revision" in value ? value.founder_revision : undefined;
    shell.publish(value);
    const nextRunIdentity = `${value.run.founder_id}\0${value.run.run_seq}\0${value.run.category}`;
    if (runIdentity !== nextRunIdentity) {
      runIdentity = nextRunIdentity;
      lastRefreshedIntentID = undefined;
      splits = [];
      personalBestMS = priorPersonalBest(readLocalTiming(localTimingStorage(), value.run.founder_id), value.run.founder_id, value.run.run_seq, value.run.category);
    }
    if (subscribedFounderID !== value.run.founder_id) {
      unsubscribe();
      subscribedFounderID = value.run.founder_id;
      transportReady = false;
      unsubscribe = runtime.subscribe(value.run.founder_id, consumePublication);
    }
    needsAuthoritativeSnapshot = false;
    offline = false;
  }

  function refresh(): Promise<boolean> {
    if (disposed) return Promise.resolve(false);
    if (refreshTask) return refreshTask;
    refreshPending = true;
    refreshTask = (async () => {
      try {
        const value = await runtime.snapshot();
        if (disposed) return false;
        bindSnapshot(value); return true;
      }
      catch { if (!disposed) { needsAuthoritativeSnapshot = true; offline = true; } return false; }
      finally { refreshPending = false; refreshTask = undefined; }
    })();
    return refreshTask;
  }

  async function beginAttempt(): Promise<void> {
    if (disposed || pending) return;
    let origin = document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
    let latestFocus: EventTarget | null | undefined = origin;
    const observeFocus = (event: FocusEvent) => { latestFocus = event.target; };
    const stopObservingFocus = observeDocumentFocus(observeFocus);
    const replaceRemovedFocus = async (destination: "vision_slide" | "desk", generation: number) => {
      await afterDOMUpdate();
      if (disposed || !root?.isConnected || surface !== destination || selectionGeneration !== generation ||
          !origin || origin.isConnected || latestFocus !== origin ||
          (document.activeElement !== origin && document.activeElement !== document.body)) return;
      const replacement = root.querySelector<HTMLElement>(destination === "desk" ? "#desk-heading" : "#vision-begin");
      if (replacement) { origin = replacement; replacement.focus(); }
    };
    actionPending = true; offline = false;
    // Retry disappears when its old error clears. Keep pending focus in the
    // same region; after success the removed entry control hands off to Desk.
    const pendingFocus = replaceRemovedFocus("vision_slide", selectionGeneration);
    try {
      const value = await runtime.bootstrap();
      if (disposed) return;
      startShell(); bindSnapshot(value); show("desk");
      await pendingFocus;
      await replaceRemovedFocus("desk", selectionGeneration);
    }
    catch { if (!disposed) offline = true; }
    finally { await pendingFocus; actionPending = false; stopObservingFocus(); }
  }

  // GS0.2: `scope` binds expected_revision to the Company or Founder stream;
  // a rejected outcome renders its reason instead of looking like offline.
  async function act(body: Record<string, unknown>, options: Readonly<{
    scope?: "company" | "founder";
    rejections?: SurfaceRejections;
    applied?: (receipt: Readonly<Record<string, unknown>>) => CopyKey | null;
    observed?: (outcome: IntentOutcome) => void;
    failed?: (error: unknown) => void;
  }> = {}): Promise<void> {
    // Capture before any queue/read await: completion belongs to the submitter,
    // not whichever panel the player selects while the response is in flight.
    const noticeOwner = surface;
    if (disposed || !snapshot || !commandControls) return;
    if (options.scope === "founder" && founderRevision === undefined) return;
    const kind = typeof body.kind === "string" ? body.kind : "";
    if (actionTask) {
      if (activeActionKind === kind) return;
      await actionTask;
    }
    if (disposed) return;
    // A click that raced an ordered event/receipt refresh must not disappear.
    // Finish that authoritative refresh, then bind the intent to its revision.
    if (refreshTask) await refreshTask;
    if (disposed || !snapshot || actionTask || !commandControls) return;
    if ((options.scope === "founder" || kind === "wind_down") && founderRevision === undefined) return;
    actionPending = true;
    activeActionKind = kind;
    const task = (async () => {
      try {
        intentNotice = null;
        intentNoticeOwner = noticeOwner;
        const expected = options.scope === "founder" ? founderRevision! : snapshot!.revision;
        const intentID = newIntentID();
        // Wind Down consumes both streams. Bind its Founder coordinate after
        // any queue/read wait, just like the Company's expected revision.
        if (kind === "wind_down") body = { ...body, expected_founder_revision: founderRevision! };
        const outcome = await runtime.intent({ intent_id: intentID, expected_revision: expected, ...body });
        if (disposed) return;
        const notice = noticeForOutcome(outcome, options.rejections);
        if (notice.invariant) console.error("game UI invariant: intent rejection");
        intentNotice = outcome.outcome === "applied" && options.applied ? options.applied(outcome.receipt) : notice.notice;
        options.observed?.(outcome);
        // Keep the single-flight guard through conflict recovery too. A repeat
        // of this kind during the read must drop, not await it as fresh consent.
        if (notice.effect === "refresh") await refresh();
        else if (outcome.outcome === "applied") {
          // GS0.2: keep controls pending until the next intent can bind to the
          // authoritative revision, even when its stream receipt arrives late.
          let postResponseRead = refreshTask === undefined;
          let refreshed = await refresh();
          if (disposed) return;
          // Gate/Decline used to bypass coalescing. Reuse an in-flight read,
          // but not its result if it sampled before this command committed.
          const appliedRevision = outcome.receipt.new_revision;
          if (!offline && (kind === "cross_gate" || kind === "decline_exit_offer") &&
              typeof appliedRevision === "number" && Number.isSafeInteger(appliedRevision) &&
              snapshot && snapshot.revision < appliedRevision) {
            postResponseRead = true;
            refreshed = await refresh();
          }
          // The stream may deliver this same persisted receipt after HTTP and
          // its read have finished. Only a successful read STARTED after the
          // response covers it; a reused pre-commit read or a failure does not.
          if (refreshed && postResponseRead && outcome.receipt.intent_id === intentID) lastRefreshedIntentID = intentID;
        }
      } catch (error) {
        if (disposed) return;
        const notice = noticeForError(error);
        if (notice.invariant) console.error("game UI invariant: invalid intent response");
        intentNotice = notice.notice;
        options.failed?.(error);
        if (notice.effect === "offline") { needsAuthoritativeSnapshot = true; offline = true; }
        else if (notice.effect === "refresh") await refresh();
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
  async function actTransition(body: Record<string, unknown>, origin: HTMLButtonElement): Promise<void> {
    if (disposed) return;
    const region = origin.closest("section");
    const controls = [...(region?.querySelectorAll<HTMLButtonElement>("button") ?? [])];
    const originIndex = controls.indexOf(origin);
    const submittedSelection = selectionGeneration;
    let latestFocus: EventTarget | null = document.activeElement;
    const observeFocus = (event: FocusEvent) => { latestFocus = event.target; };
    const stopObservingFocus = observeDocumentFocus(observeFocus);
    try {
      await act(body);
      await afterDOMUpdate();
      // GS0.6: a disappearing transition returns focus within its own region.
      // A newer native focus choice or lifecycle navigation always takes precedence.
      if (disposed || surface !== "desk" || selectionGeneration !== submittedSelection || origin.isConnected
          || latestFocus !== origin || (document.activeElement !== origin && document.activeElement !== document.body)) return;
      const nearest = controls.map((control, index) => ({ control, index, distance: Math.abs(index - originIndex) }))
        .filter(({ control }) => control.isConnected && !control.disabled)
        .sort((left, right) => left.distance - right.distance || right.index - left.index)[0]?.control;
      (nearest ?? root?.querySelector<HTMLElement>("#desk-heading"))?.focus();
    } finally { stopObservingFocus(); }
  }
  function acceptOffer(): void {
    if (pending || !offer || founderRevision === undefined) return;
    void act(withPlan({ kind: "accept_exit_offer", expected_founder_revision: founderRevision, offer_id: offer.payload.offer_id }));
  }
  function declineOffer(): void {
    if (pending || !offer) return;
    void act({ kind: "decline_exit_offer", offer_id: offer.payload.offer_id });
  }

  async function continueRun(): Promise<void> {
    if (disposed || !ended || pending) return;
    const expectedFounderID = ended.payload.founder_id;
    const expectedRunSeq = ended.payload.run_id.run_seq + 1;
    const focusedOrigin = document.activeElement;
    actionPending = true;
    try {
      const value = await runtime.snapshot();
      if (disposed) return;
      if (value.run.founder_id !== expectedFounderID) throw new RangeError("next Company snapshot belongs to another Founder");
      if (value.run.run_seq !== expectedRunSeq) throw new RangeError("next Company snapshot did not advance exactly one run");
      bindSnapshot(value);
      ended = undefined;
      offer = undefined;
      exitPlan = [];
      show("desk");
      const continuationSelection = selectionGeneration;
      await afterDOMUpdate();
      if (!disposed && selectionGeneration === continuationSelection && surface === "desk"
        && focusedOrigin instanceof HTMLElement && !focusedOrigin.isConnected
        && (document.activeElement === focusedOrigin || document.activeElement === document.body)) {
        root?.querySelector<HTMLElement>("#desk-heading")?.focus();
      }
    } catch { if (!disposed) { needsAuthoritativeSnapshot = true; offline = true; } }
    finally { actionPending = false; }
  }

  function preemptLifecycle(cursor: number, destination: "offer_sheet" | "run_end"): void {
    const previous = navigation.active;
    navigation.lifecycle({ cursor, surface: destination });
    surface = navigation.active;
    if (surface === previous || surface !== destination) return;
    const forcedSelection = selectionGeneration;
    void afterDOMUpdate().then(() => {
      if (disposed || selectionGeneration !== forcedSelection || surface !== destination) return;
      const headingID = destination === "offer_sheet" ? "offer-heading" : "run-end-heading";
      root?.querySelector<HTMLElement>(`#${headingID}`)?.focus();
    });
  }

  function consumePublication(message: GameUIRuntimeMessage): void {
    if (disposed) return;
    if (message.kind === "transport_closed") { offline = true; transportReady = false; subscribedFounderID = undefined; unsubscribe(); return; }
    if (message.kind === "transport_recovering") { offline = true; transportReady = false; return; }
    if (message.kind === "transport_recovered") { offline = needsAuthoritativeSnapshot; transportReady = true; draining = false; resyncing = false; return; }
    if (message.kind === "snapshot") { bindSnapshot(message.value); draining = false; resyncing = false; return; }
    if (message.kind === "historical_event") return;
    // A terminal command persists run_ended and run_started in one ordered
    // transaction. Refreshing its trailing receipt would bind the already
    // created next run and let its lifecycle preempt the run-end screen before
    // the player explicitly continues.
    if (message.kind === "receipt") {
      if (!ended) {
        if (offline || message.intentID === undefined || message.intentID !== lastRefreshedIntentID) void refresh();
        // SG10: the main snapshot does not carry Garden's advisory DTO.
        // Receipt publications must invalidate the mounted Garden read too.
        gardenRefresh += 1;
      }
      return;
    }
    if (message.kind === "presence") { visitorCount = message.count; return; }
    if (message.kind === "announcement") { announce(message.scope, message.value, message.eventID); return; }
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
      offer = value; preemptLifecycle(value.cursor, "offer_sheet");
    } else if (value.kind === "run_ended") {
      ended = value;
      timer?.terminal(value.payload.rta_ms);
      if (snapshot) writeLocalRunTiming(localTimingStorage(), { category: snapshot.run.category, founder_id: value.payload.founder_id, pb_rta_ms: value.payload.rta_ms, run_seq: value.payload.run_id.run_seq, splits });
      preemptLifecycle(value.cursor, "run_end");
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

  function announce(scope: "company" | "founder", value: GameUIAnnouncementEvent, eventID?: string): void {
    // Production always supplies the validated transport event ID. Older
    // fixture runtimes omit it; kind keeps distinct event families separate.
    const key = `${scope}\0${value.cursor}\0${eventID ?? value.kind}`;
    if (announcedCursors.has(key)) return;
    announcedCursors.add(key);
    if (value.kind === "achievement_earned") {
      const row = liveFeatures?.achievements?.rows.find((candidate) => candidate.achievement_id === value.payload.achievement_id);
      if (!row || !applicationCopyCatalog.byKey.has(row.copy_key)) { console.error(`game UI invariant: unannounceable achievement ${value.payload.achievement_id}`); return; }
      announcement = t("achievements.earned_announcement", { achievement: t(row.copy_key as CopyKey, {}, era) }, era);
      return;
    }
    if (value.kind === "fiscal_period_harvested") {
      if (surface !== "fiscal") fiscalHarvested = true;
      return;
    }
    if (value.kind === "pet_status_changed") {
      if (scope !== "founder" || surface !== "pet") return;
      const pet = liveFeatures?.pet_adoption?.pets.find((row) => row.pet_id === value.payload.pet_id);
      const band = FEATURES_PRESENTATION.petBands.get(value.payload.to_status_band);
      if (!pet || !band) { console.error(`game UI invariant: unannounceable pet status ${value.payload.pet_id}`); return; }
      announcement = t("pet.status_changed_announcement", { band: t(band, {}, era) }, era);
      return;
    }
    if (value.kind === "buff_started") {
      const row = FEATURES_PRESENTATION.opportunityEffects.get(value.payload.effect_row_id);
      if (!row) { console.error(`game UI invariant: unannounceable buff ${value.payload.effect_row_id}`); return; }
      if (surface === "desk") announcement = t("desk.buff.started_announcement", { effect: t(row.title_key, {}, era) }, era);
      return;
    }
    const meter = meterLabel(value.payload.meter_id), band = FEATURES_PRESENTATION.meterBands.get(value.payload.to_band);
    if (!meter || !band) { console.error(`game UI invariant: unannounceable meter change ${value.payload.meter_id}`); return; }
    if (surface === "meters") announcement = t("meters.band_changed_announcement", { meter, band: t(band, {}, era) }, era);
    else metersChanged = true;
  }

  function factTrue(id: string): boolean { return snapshot?.facts.some((fact) => fact.fact_id === id && fact.value === true) ?? false; }
  function upgradeReasonKey(upgrade: GameUISnapshot["upgrades"][number]): CopyKey | undefined {
    const reason = upgrade.ineligible_reason;
    if (reason === undefined || reason === null) return undefined;
    return ({ owned: "desk.upgrade.owned", window: "desk.rejection.not_in_window", requirement: "desk.rejection.requires", unaffordable: "desk.rejection.unaffordable" } as const)[reason];
  }
  const liveFeatures = $derived(snapshot && "features" in snapshot ? snapshot.features : undefined);

  // GS5: the last applied claim (receipt evidence) and a once-per-opportunity
  // polite spawn announcement while the Desk is mounted. Focus never moves.
  let lastClaim = $state<LastClaim | undefined>();
  const announcedOpportunities = new Set<string>();
  $effect(() => {
    const offer = liveFeatures?.opportunity?.pending;
    if (!offer || surface !== "desk" || announcedOpportunities.has(offer.opportunity_id)) return;
    announcedOpportunities.add(offer.opportunity_id);
    const row = FEATURES_PRESENTATION.opportunityEffects.get(offer.effect_row_id);
    if (row) announcement = t("desk.opportunity.spawned_announcement", { effect: t(row.title_key, {}, era) }, era);
  });
  function claimApplied(receipt: Readonly<Record<string, unknown>>): CopyKey | null {
    try { lastClaim = lastClaimFromReceipt(receipt); }
    catch { console.error("game UI invariant: applied claim without opportunity evidence"); return "intent.rejection.unknown"; }
    return lastClaim.saturated && lastClaim.credited !== null ? "desk.opportunity.lucky_capped" : null;
  }
  const pitchAvailability = $derived(liveFeatures?.minigames?.rows.find((row) => row.minigame_id === "pitch"));
  const FISCAL_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/period_not_ripe", "fiscal.rejection.period_not_ripe"],
    ["unaffordable/fiscal_credit", "fiscal.rejection.unaffordable"],
    ["not_eligible/already_unlocked", "fiscal.rejection.already_unlocked"],
  ]);
  const fiscalRejections: SurfaceRejections = $derived(new Map([
    ...FISCAL_REJECTIONS,
    ...(liveFeatures?.fiscal?.generator_levels ?? []).map((row): [string, CopyKey] =>
      [`cap_exceeded/${row.generator_id}`, row.level_cap.reason_key as CopyKey]),
  ]));
  const REPUTATION_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/*", "reputation_tree.error.not_eligible"],
    ["unknown_id/*", "reputation_tree.error.unknown_id"],
    ["unaffordable/reputation", "reputation_tree.error.unaffordable"],
    ["invalid/*", "reputation_tree.error.invalid"],
  ]);
  function reputationApplied(): CopyKey { void refresh(); return "reputation_tree.result.applied"; }
  let reputationFeedback = $state<{ nodeID: string; key: CopyKey } | null>(null);
  function purchaseReputation(nodeID: string): Promise<void> {
    reputationFeedback = null;
    return act({ kind: "purchase_reputation_node", node_id: nodeID }, {
      scope: "founder", rejections: REPUTATION_REJECTIONS, applied: reputationApplied,
      observed: (outcome) => {
        if (outcome.outcome !== "rejected") return;
        // Presentation only: do not put conflicts in SurfaceRejections, whose
        // early match would remove the shared authoritative refresh effect.
        const key = outcome.category === "revision_conflict" ? "reputation_tree.error.revision_conflict" :
          REPUTATION_REJECTIONS.get(`${outcome.category}/${outcome.detail}`) ?? REPUTATION_REJECTIONS.get(`${outcome.category}/*`);
        if (key) reputationFeedback = { nodeID, key };
      },
      failed: (error) => {
        if (error instanceof GameUIRequestError && error.status === 409) {
          reputationFeedback = { nodeID, key: "reputation_tree.error.revision_conflict" };
        }
      },
    });
  }
  // Pet Adoption v1 PA8: adoption rejections render inline on the card.
  // GS4: exact care rejection pairs (server/pet grammar + founder_replay).
  const CARE_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/cooldown", "pet.care.rejection.cooldown"], ["not_eligible/ineligible", "pet.care.rejection.ineligible"],
    ["not_eligible/saturated", "pet.care.rejection.saturated"], ["not_eligible/human_content_locked", "pet.care.rejection.soul_locked"],
  ]);
  const ADOPTION_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/adoption_inactive", "pet.adoption.reject.adoption_inactive"],
    ["not_eligible/species_locked", "pet.adoption.reject.species_locked"],
    ["not_eligible/adoption_cap_reached", "pet.adoption.reject.adoption_cap_reached"],
  ]);
  function adoptionApplied(): null { void refresh(); return null; }
  // Cosmetic Shop v1 §7.3: inline rejections and the session-local parody
  // receipt (not re-announced after a reload; ownership comes from the snapshot).
  // Server Garden SG5: every closed detail maps to its error.garden.* key.
  const GARDEN_REJECTIONS: SurfaceRejections = new Map<string, CopyKey>([
    ...["garden_inactive", "fiscal_unlock_required", "human_content_locked", "plot_dormant", "plot_occupied", "seed_not_collected", "plant_not_mature",
      "substrate_unchanged", "substrate_lockout"].map((detail): [string, CopyKey] => [`not_eligible/${detail}`, `error.garden.${detail}` as CopyKey]),
    ...["garden_species", "garden_plot", "garden_substrate"].map((detail): [string, CopyKey] => [`unknown_id/${detail}`, `error.garden.${detail}` as CopyKey]),
  ]);
  // SG2/SG9/SG10: presence and the mounted read belong to one Founder and
  // pinned catalog. Startup and surface reads share this ordering guard.
  let gardenContext = $state("");
  let gardenPresence = $state<"unknown" | "inactive" | "present">("unknown");
  let gardenReadGeneration = 0;
  let gardenRefresh = $state(0);
  const gardenPort: GardenPort = {
    async current() {
      const context = gardenContext, generation = ++gardenReadGeneration;
      const view = await runtime.garden!.current();
      if (!disposed && context === gardenContext && generation === gardenReadGeneration) {
        gardenPresence = view.kind === "inactive" ? "inactive" : "present";
      }
      return view;
    },
  };
  async function probeGarden(): Promise<void> {
    if (!runtime.garden) return;
    // Failed reads are not evidence of absence. The mounted surface owns its
    // error/stale display; an unconfirmed startup tab stays hidden.
    try { await gardenPort.current(); } catch { /* retain unknown/last known presence */ }
  }
  function gardenAct(body: Record<string, unknown>): void {
    const context = gardenContext;
    void act(body, { scope: "founder", rejections: GARDEN_REJECTIONS }).then(() => {
      if (!disposed && context === gardenContext) gardenRefresh += 1;
    });
  }
  const COSMETIC_REJECTIONS: SurfaceRejections = new Map([
    ["not_eligible/inactive", "shop.cosmetics.reject.inactive"], ["not_eligible/locked", "shop.cosmetics.reject.locked"],
    ["not_eligible/owned", "shop.cosmetics.reject.owned"], ["not_eligible/not_owned", "shop.cosmetics.reject.not_owned"],
    ["not_eligible/already_equipped", "shop.cosmetics.reject.already_equipped"], ["not_eligible/nothing_equipped", "shop.cosmetics.reject.nothing_equipped"],
    ["unknown_id/*", "shop.cosmetics.reject.unknown"],
  ]);
  let cosmeticReceipt = $state<Readonly<{ cosmeticId: string; orderNumber: number }> | null>(null);
  function cosmeticApplied(receipt: Readonly<Record<string, unknown>>): null {
    const event = receipt.event as { kind?: unknown; payload?: { cosmetic_id?: unknown; order_number?: unknown } } | undefined;
    if (receipt.kind === "acquire_cosmetic" && typeof event?.payload?.cosmetic_id === "string" && Number.isSafeInteger(event.payload.order_number)) {
      cosmeticReceipt = { cosmeticId: event.payload.cosmetic_id, orderNumber: event.payload.order_number as number };
    }
    void refresh();
    return null;
  }
  function cosmeticPetName(petID: string): string {
    const row = liveFeatures?.pet_adoption?.pets.find((pet) => pet.pet_id === petID);
    return row ? t(row.name_key as CopyKey, {}, era) : "";
  }
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
      surface === "meters" && !liveFeatures?.meters || surface === "reputation_tree" && !liveFeatures?.reputation ||
      surface === "garden" && gardenPresence === "inactive";
    if (snapshot && armless) {
      const forcedSnapshot = snapshot;
      show("desk");
      const forcedSelection = selectionGeneration;
      void afterDOMUpdate().then(() => {
        if (selectionGeneration !== forcedSelection || snapshot !== forcedSnapshot || surface !== "desk") return;
        root?.querySelector<HTMLElement>("#desk-heading")?.focus();
      });
    }
  });
  const commandControls = $derived(transportReady && !offline && !resyncing);
  const founderControls = $derived(founderRevision !== undefined && commandControls);

  function exitTitle(exitType: string): string { return t(requirePresentation(GAME_UI_PRESENTATION.exitTypes, exitType).title_key, {}, era); }

  function declaredCopyKey(key: string): CopyKey {
    if (!applicationCopyCatalog.byKey.has(key)) throw new RangeError(`missing copy ${key}`);
    return key as CopyKey;
  }

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
        <button type="button" tabindex="0" aria-current={surface === "desk" ? "page" : undefined} onclick={() => show("desk")}>{t("surface.desk.title", {}, era)}</button>
        {#if factTrue("feature.achievements")}<button type="button" tabindex="0" aria-current={surface === "achievements" ? "page" : undefined} onclick={() => show("achievements")}>{t("surface.achievements.title", {}, era)}</button>{/if}
        {#if factTrue("feature.fiscal")}<button type="button" tabindex="0" aria-current={surface === "fiscal" ? "page" : undefined} onclick={() => show("fiscal")}>{t("surface.fiscal.title", {}, era)}{#if fiscalHarvested}<span class="nav-badge">{t("fiscal.nav.harvest_badge", {}, era)}</span>{/if}</button>{/if}
        {#if factTrue("feature.pets")}<button type="button" tabindex="0" aria-current={surface === "pet" ? "page" : undefined} onclick={() => show("pet")}>{t("pet.care.panel.title", {}, era)}</button>{/if}
        {#if factTrue("feature.meters")}<button type="button" tabindex="0" aria-current={surface === "meters" ? "page" : undefined} onclick={() => show("meters")}>{t("surface.meters.title", {}, era)}{#if metersChanged} {t("meters.nav_changed_badge", {}, era)}{/if}</button>{/if}
        {#if runtime.minigame && factTrue("feature.minigame.pitch")}<button type="button" tabindex="0" aria-current={surface === "minigame_session" ? "page" : undefined} onclick={() => show("minigame_session")}>{t("minigame.pitch.title", {}, era)}</button>{/if}
        {#if runtime.soulRecovery}<button type="button" tabindex="0" aria-current={surface === "soul_recovery" ? "page" : undefined} onclick={() => show("soul_recovery")}>{t("soul.recovery_surface.title", {}, era)}</button>{/if}
        {#if factTrue("feature.reputation_tree")}<button type="button" tabindex="0" aria-current={surface === "reputation_tree" ? "page" : undefined} onclick={() => show("reputation_tree")}>{t("reputation_tree.title", {}, era)}</button>{/if}
        {#if runtime.garden && gardenPresence === "present"}<button type="button" tabindex="0" aria-current={surface === "garden" ? "page" : undefined} onclick={() => show("garden")}>{t("garden.title", {}, era)}</button>{/if}
        <button type="button" tabindex="0" aria-current={surface === "settings" ? "page" : undefined} onclick={() => show("settings")}>{t("surface.settings.title", {}, era)}</button>
      </nav>
      {#if snapshot.run.run_seq === 1 && visitorCount !== undefined}<span class="visitor" title={t("chrome.visitor_counter.tooltip", {}, era)}>{t("chrome.visitor_counter.frame", { count: visitorCount }, era)}</span>{/if}
    </header>
  {/if}

  {#if snapshot}<p class="announcement" role="status">{announcement}</p>{/if}
  {#if snapshot && (surface === "desk" || surface === "offer_sheet" || surface === "achievements" || surface === "meters") && !commandControls}
    <p class="snapshot-stale">{t("common.stale_note", {}, era)}</p>
  {/if}
  {#if snapshot && surface !== "fiscal" && surface !== "pet"}
    <p class="intent-notice" role="status">{#if surface === "offer_sheet" && pending}<span id="offer-pending">{t("common.pending", {}, era)}</span> {/if}{#if surface === "desk" && deskPending}<span id="desk-pending">{t("common.pending", {}, era)}</span> {/if}{intentNoticeOwner === surface && intentNotice && !(surface === "reputation_tree" && reputationFeedback) ? t(intentNotice, {}, era) : ""}</p>
  {/if}
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
        <button id="vision-begin" type="button" tabindex="0" aria-disabled={pending || undefined} onclick={beginAttempt}>{pending ? t("screen.vision_slide.connecting", {}, era) : t("screen.vision_slide.begin_attempt", {}, era)}</button>
        {#if offline}<p role="alert">{t("screen.vision_slide.offline_fallback", {}, era)}</p><button type="button" tabindex="0" aria-disabled={pending || undefined} onclick={beginAttempt}>{t("screen.vision_slide.retry", {}, era)}</button>{/if}
        <small>{t("screen.vision_slide.small_print", {}, era)}</small>
      </article>
    </section>
  {:else if !snapshot && surface === "desk"}
    <section class="surface loading" aria-labelledby="loading-heading">
      <h1 id="loading-heading" tabindex="-1">{t("surface.desk.title", {}, era)}</h1>
      <p role="status">{t("common.loading", {}, era)}</p>
    </section>
  {:else if snapshot && surface === "desk"}
    <section class="surface desk" aria-labelledby="desk-heading">
      <h1 id="desk-heading" tabindex="-1">{t("surface.desk.title", {}, era)}</h1>
      {#if liveFeatures?.pet_adoption}
        {@const adoption = liveFeatures.pet_adoption}
        <AdoptionCard availability={adoption.pet_adoption} adopted={adoption.pets[0]} {era} {pending} controlsEnabled={founderControls}
          rejection={intentNotice?.startsWith("pet.adoption.reject.") ? intentNotice : null} reducedMotion={prefersReducedMotion}
          onAdopt={(speciesID, nameKey) => act({ kind: "adopt_pet", species_id: speciesID, name_key: nameKey }, { scope: "founder", rejections: ADOPTION_REJECTIONS, applied: adoptionApplied })} />
      {/if}
      <section class="manual cc-window">
        <h2>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).title_key, {}, era)}</h2>
        <p>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).description_key, {}, era)}</p>
        <label>{t("desk.manual.meter_label", {}, era)} <meter min="0" max={snapshot.manual_action.bucket_cap_milli} value={visibleManualTokensMilli()}></meter></label>
        <output>{t("desk.manual.meter_frame", { current: Math.floor(visibleManualTokensMilli() / 1000), cap: Math.floor(snapshot.manual_action.bucket_cap_milli / 1000) }, era)}</output>
        <button type="button" disabled={pending || !commandControls} title={t("desk.manual.meter_tooltip", {}, era)} onclick={() => act({ kind: "perform_manual_batch", action_id: snapshot!.manual_action.action_id, count: 1, window_ms: 1 })}>{t(requirePresentation(GAME_UI_PRESENTATION.manualActions, snapshot.manual_action.action_id).title_key, {}, era)}</button>
      </section>
      <OpportunityRegion arm={liveFeatures?.opportunity ?? null} {era} {pending} controlsEnabled={commandControls} {lastClaim}
        onClaim={(opportunityID) => act({ kind: "claim_opportunity", opportunity_id: opportunityID }, { rejections: OPPORTUNITY_REJECTIONS, applied: claimApplied })} />

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
              <div><button type="button" tabindex="0" disabled={!commandControls || generator.max_affordable < 1} aria-disabled={generatorPending || undefined} aria-describedby={generatorPending ? "desk-pending" : undefined} onclick={() => act({ kind: "buy_generator", generator_id: generator.generator_id, count: { mode: "exact", value: 1 } })}>{t("desk.buy_one", {}, era)}</button><button type="button" tabindex="0" disabled={!commandControls || generator.max_affordable < 1} aria-disabled={generatorPending || undefined} aria-describedby={generatorPending ? "desk-pending" : undefined} onclick={() => act({ kind: "buy_generator", generator_id: generator.generator_id, count: { mode: "max" } })}>{t("desk.buy_max", {}, era)}</button></div>
            </article>
          {/each}
        </div>
      </section>

      {#if liveFeatures?.axis_stack}<AxisStackPanel arm={liveFeatures.axis_stack} {era} />{/if}
      <section aria-labelledby="upgrades-heading">
        <h2 id="upgrades-heading">{t("desk.upgrades_label", {}, era)}</h2>
        <div class="cards">
          {#each snapshot.upgrades as upgrade (upgrade.upgrade_id)}
            {@const presentation = upgradePresentation(upgrade.upgrade_id)}
            {@const reasonKey = upgradeReasonKey(upgrade)}
            {@const reasonID = `upgrade-reason-${upgrade.upgrade_id}`}
            <article class="card">
              <h3>{t(presentation.title_key, {}, era)}</h3>
              <p>{t(presentation.description_key, {}, era)}</p>
              <Amount value={upgrade.cost_amount} era={era} />
              {#if upgrade.owned}<strong id={reasonID}>{t("desk.upgrade.owned", {}, era)}</strong>
              {:else if reasonKey}<p id={reasonID}>{t(reasonKey, {}, era)}</p>{/if}
              <button type="button" tabindex="0" disabled={!commandControls || !upgrade.eligible || upgrade.owned} aria-disabled={upgradePending || undefined}
                aria-describedby={[...(upgrade.owned || reasonKey ? [reasonID] : []), ...(upgradePending ? ["desk-pending"] : [])].join(" ") || undefined}
                onclick={() => act({ kind: "buy_upgrade", upgrade_id: upgrade.upgrade_id })}>{t("desk.buy_one", {}, era)}</button>
            </article>
          {/each}
        </div>
      </section>

      {#if splits.length}
        <details><summary>{t("chrome.splits.label", {}, era)}</summary>{#each splits as split}<p>{t(requirePresentation(GAME_UI_PRESENTATION.gates, split.gate_id).title_key, {}, era)} {duration(split.rta_ms)}</p>{/each}{#if personalBestMS === undefined}<p>{t("chrome.splits.first_attempt_note", {}, era)}</p>{/if}</details>
      {/if}

      {#if liveFeatures?.cosmetics?.active}
        {@const shop = liveFeatures.cosmetics}
        {#if shop.items.some((item) => item.acquirable || item.owned)}
          <CosmeticShelf arm={shop} presentation={COSMETIC_SHOP_PRESENTATION} price={requirePresentationConstant("constant.price_zero")} {era} {pending}
            controlsEnabled={founderControls} petName={cosmeticPetName} receipt={cosmeticReceipt}
            rejection={intentNotice?.startsWith("shop.cosmetics.reject.") ? intentNotice : null}
            onAcquire={(id) => act({ kind: "acquire_cosmetic", cosmetic_id: id }, { scope: "founder", rejections: COSMETIC_REJECTIONS, applied: cosmeticApplied })}
            onEquip={(id, petID) => act({ kind: "equip_cosmetic", cosmetic_id: id, pet_id: petID }, { scope: "founder", rejections: COSMETIC_REJECTIONS, applied: cosmeticApplied })}
            onUnequip={(petID) => act({ kind: "unequip_cosmetic", pet_id: petID }, { scope: "founder", rejections: COSMETIC_REJECTIONS, applied: cosmeticApplied })} />
        {/if}
      {:else}
        <!-- §6 pre-activation: the static card, unchanged (fail closed). -->
      <section class="card"><h2>{t("cosmetic.horse_armor_free.title", {}, era)}</h2><p>{t("cosmetic.horse_armor_free.description", { price: requirePresentationConstant("constant.price_zero") }, era)}</p><small>{t("cosmetic.horse_armor_free.disclosure", {}, era)}</small></section>
      {/if}
      {#if snapshot.schema_version >= 3 && "transitions" in snapshot}
        {@const transitions = snapshot.transitions}
        <section class="card">
          {#if transitions.cross_gate}
            <button type="button" tabindex="0" disabled={!commandControls || !transitions.cross_gate.eligible} aria-disabled={gatePending || undefined} aria-describedby={gatePending ? "desk-pending" : undefined} onclick={(event) => actTransition({ kind: "cross_gate", gate_id: transitions.cross_gate!.gate_id, route_id: null }, event.currentTarget)}>{t("desk.cross_gate", {}, era)}</button>
          {/if}
          {#if transitions.incorporate}
            <fieldset class="incorporate">
              <legend>{t("incorporate.panel.title", {}, era)}</legend>
              <p>{t("incorporate.panel.hint", {}, era)}</p>
              {#each transitions.incorporate.factions as row (row.faction_id)}
                <button type="button" tabindex="0" disabled={!commandControls} aria-disabled={incorporatePending || undefined} aria-describedby={incorporatePending ? "desk-pending" : undefined} onclick={(event) => actTransition({ kind: "incorporate", faction_id: row.faction_id }, event.currentTarget)}>{t(declaredCopyKey(row.copy_key), {}, era)}</button>
              {/each}
            </fieldset>
          {/if}
          <button type="button" tabindex="0" disabled={!founderControls || !transitions.wind_down.eligible} aria-disabled={windDownPending || undefined} aria-describedby={windDownPending ? "desk-pending" : undefined} onclick={(event) => actTransition(withPlan({ kind: "wind_down", expected_founder_revision: founderRevision }), event.currentTarget)}>{t("desk.wind_down", {}, era)}</button>
          {#if liveFeatures?.reputation && transitions.wind_down.eligible}<ReputationPlanPanel arm={liveFeatures.reputation} {era} previewDelta={0} onChange={(plan) => { exitPlan = [...plan]; }} />{/if}
        </section>
      {/if}
      {#if era === "era_2010"}
        <!-- FarmVille-era energy bar: presentation only. It limits nothing, holds no state, and its refill emits no intent (§E3). -->
        <section class="card energy" aria-labelledby="energy-heading" title={t("era_2010.energy_bar.curtain", {}, era)}>
          <h2 id="energy-heading">{t("era_2010.energy_bar.label", {}, era)}</h2>
          <meter min={0} max={1} value={1} aria-labelledby="energy-heading" aria-describedby="energy-curtain"></meter>
          <button type="button" onclick={() => { energyRefilled = true; }}>{t("era_2010.energy_bar.refill", {}, era)}</button>
          {#if energyRefilled}<p role="status">{t("era_2010.energy_bar.refilled", {}, era)}</p>{/if}
          <small id="energy-curtain">{t("era_2010.energy_bar.curtain", {}, era)}</small>
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
      <h1 id="offer-heading" tabindex="-1">{t("screen.offer_sheet.heading", {}, era)}</h1><p>{t("screen.offer_sheet.preamble", {}, era)}</p><h2>{t("screen.offer_sheet.terms_label", {}, era)}</h2>
      <p>{exitTitle(offer.payload.exit_type)}</p>
      {#each renderPrestigeTermRows(offer.payload.payout_preview, era) as row}<p>{row}</p>{/each}
      <p title={t("screen.offer_sheet.countdown_tooltip", {}, era)}>{t("screen.offer_sheet.countdown_frame", { remaining: duration(offer.payload.expires_at_ms - estimatedServerNowMS()) }, era)}</p>
      {#if liveFeatures?.reputation}<ReputationPlanPanel arm={liveFeatures.reputation} {era} previewDelta={offer.payload.payout_preview.reputation_delta} onChange={(plan) => { exitPlan = [...plan]; }} />{/if}
      <button type="button" tabindex="0" disabled={!founderControls} aria-disabled={pending || undefined} aria-describedby={pending ? "offer-pending" : undefined} onclick={acceptOffer}>{t("screen.offer_sheet.accept", {}, era)}</button>
      <button type="button" tabindex="0" disabled={!commandControls} aria-disabled={pending || undefined} aria-describedby={pending ? "offer-pending" : undefined} onclick={declineOffer}>{t("screen.offer_sheet.decline", {}, era)}</button>
    </section>
  {:else if surface === "run_end" && ended}
    <RunEndSurface {ended} />
    <button type="button" tabindex="0" aria-disabled={pending || undefined} onclick={continueRun}>{t("screen.run_end.continue", {}, era)}</button>
    {#if offline}<p role="alert">{t("settings.save_status.offline", {}, era)}</p>{/if}
  {:else if surface === "achievements" && liveFeatures?.achievements}
    <AchievementsSurface arm={liveFeatures.achievements} {era} />
  {:else if surface === "meters" && liveFeatures?.meters}
    <MetersSurface arm={liveFeatures.meters} {era} />
  {:else if surface === "reputation_tree" && liveFeatures?.reputation}
    <ReputationTreeSurface arm={liveFeatures.reputation} {era} {pending} controlsEnabled={founderControls && transportReady}
      offline={offline || !transportReady}
      feedback={reputationFeedback} onPurchase={purchaseReputation} />
  {:else if surface === "fiscal" && liveFeatures?.fiscal}
    <FiscalSurface arm={liveFeatures.fiscal} {era} serverNowMs={estimatedServerNowMS()} {pending} controlsEnabled={founderControls && transportReady} notice={intentNoticeOwner === "fiscal" ? intentNotice : null}
      onHarvest={() => act({ kind: "harvest_fiscal_period" }, { scope: "founder", rejections: fiscalRejections, applied: harvestNotice })}
      onSpendLevel={(generatorID) => act({ kind: "spend_fiscal_credit", target: { kind: "generator_level", generator_id: generatorID, levels: 1 } }, { scope: "founder", rejections: fiscalRejections })}
      onSpendUnlock={(unlockID) => act({ kind: "spend_fiscal_credit", target: { kind: "unlock", unlock_id: unlockID } }, { scope: "founder", rejections: fiscalRejections })} />
  {:else if snapshot && surface === "minigame_session" && runtime.minigame}
    {#if pitchAvailability && !pitchAvailability.unlocked}<p class="intent-notice" role="note">{t("minigame.availability.fiscal_locked", {}, era)}</p>{/if}
    {#if pitchAvailability?.human_content_locked}<p class="intent-notice" role="note">{t("minigame.availability.soul_locked", {}, era)}</p>{/if}
    <MinigameSessionSurface port={runtime.minigame} minigameID="pitch" {era} newCommandID={() => newIntentID()} onExitToHost={() => show("desk")} onTerminal={() => { void refresh(); }} />
  {:else if snapshot && surface === "pet" && liveFeatures?.pet_adoption && liveFeatures.pet_adoption.pets.length > 0}
    <PetCareSurface pets={liveFeatures.pet_adoption.pets} cosmetics={liveFeatures.cosmetics ?? null} {era} {pending} controlsEnabled={founderControls && transportReady} reducedMotion={prefersReducedMotion} notice={intentNoticeOwner === "pet" ? intentNotice : null}
      onCare={(petID, actionID) => act({ kind: "care_action", pet_id: petID, action_id: actionID }, { scope: "founder", rejections: CARE_REJECTIONS, applied: () => "pet.care.applied" })} />
  {:else if snapshot && surface === "garden" && runtime.garden}
    {#key gardenContext}
    <GardenSurface port={gardenPort} {era} {pending} controlsEnabled={founderControls} refreshKey={gardenRefresh}
      rejection={intentNotice?.startsWith("error.garden.") ? intentNotice : null}
      onPlant={(row, col, species) => gardenAct({ kind: "garden_plant", row, col, species_id: species })}
      onUproot={(row, col) => gardenAct({ kind: "garden_uproot", row, col })}
      onHarvest={(plots) => gardenAct({ kind: "garden_harvest", plots: plots.map((plot) => ({ row: plot.row, col: plot.col })) })}
      onSetSubstrate={(substrate) => gardenAct({ kind: "garden_set_substrate", substrate_id: substrate })} />
    {/key}
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
  nav, .visitor { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); padding: 0 var(--cc-space-md) var(--cc-space-sm); }
  .cc-titlebar > *, nav > button { min-inline-size: 0; overflow-wrap: anywhere; }
  .surface { display: grid; gap: var(--cc-space-lg); max-width: 72rem; margin: auto; padding: var(--cc-space-lg); }
  .vision { grid-template-columns: repeat(auto-fit, minmax(min(18rem, 100%), 1fr)); }
  .slide, .contract, .card, .notice, .readme { padding: var(--cc-space-lg); }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(15rem, 100%), 1fr)); gap: var(--cc-space-md); }
  .surface > *, .surface h1 { min-inline-size: 0; overflow-wrap: anywhere; }
  .card { display: grid; gap: var(--cc-space-sm); }
  .manual { display: grid; gap: var(--cc-space-sm); padding: var(--cc-space-lg); }
  .notice { display: flex; flex-wrap: wrap; gap: var(--cc-space-sm); margin-block-end: var(--cc-space-md); color: var(--cc-color-text); background: var(--cc-color-surface); }
  button { padding: var(--cc-space-sm) var(--cc-space-md); color: var(--cc-color-text); background: var(--cc-chrome-button_face); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); font-family: var(--cc-type-font_ui); font-size: inherit; cursor: pointer; }
  button:focus-visible, summary:focus-visible { outline: var(--cc-border-width) var(--cc-border-style) var(--cc-color-accent); outline-offset: var(--cc-space-xs); }
  button[disabled] { color: var(--cc-color-text_muted); cursor: default; }
  h1, h2, h3, p { margin: 0; }
  h1, h2, h3 { font-family: var(--cc-type-font_display); }
  pre { overflow: auto; color: var(--cc-color-text); background: var(--cc-color-surface); font-family: var(--cc-type-font_mono); white-space: pre-wrap; }
  .nav-badge { margin-inline-start: var(--cc-space-xs); }
  meter { inline-size: 100%; accent-color: var(--cc-color-accent); }
</style>
