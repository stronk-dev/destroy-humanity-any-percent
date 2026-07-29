<script lang="ts">
  import { onMount } from "svelte";
  import type { ShellController, ShellTab, ShellView } from "./controller";

  let { controller }: { controller: ShellController } = $props();
  let view: ShellView = $state({ screen: "contract", activeTab: "company", revision: 0, resources: {}, discrete: {}, progress: [], returnFastForwardComplete: false, offlineGains: {}, attemptElapsedMs: 0 });
  const tabs: readonly ShellTab[] = ["company", "world", "pet", "minigame"];
  const formatElapsed = (milliseconds: number) => {
    const centiseconds = Math.floor(milliseconds / 10);
    const hours = Math.floor(centiseconds / 360000);
    const minutes = Math.floor(centiseconds / 6000) % 60;
    const seconds = Math.floor(centiseconds / 100) % 60;
    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}.${String(centiseconds % 100).padStart(2, "0")}`;
  };
  onMount(() => {
    const unsubscribe = controller.subscribe((next) => { view = next; });
    const timer = window.setInterval(() => { view = controller.view(); }, 100);
    return () => { window.clearInterval(timer); unsubscribe(); };
  });
</script>

{#if view.screen === "contract"}
  <main class="contract" aria-labelledby="contract-title">
    <p class="eyebrow">HUMANITY</p>
    <h1 id="contract-title">Destroy Any%</h1>
    <p class="timer">00:00:00.00</p>
    <p>PB/WR comparison unlocks after the first Exit.</p>
    <button type="button" onclick={() => controller.beginAttempt()}>BEGIN ATTEMPT</button>
    <small>Free forever. No purchases. No ads. The only thing this game harvests is the fictional planet.</small>
  </main>
{:else if view.screen === "main"}
  <main class="shell">
    <header>
      <strong>Cloud Clicker</strong><span class="timer">RTA {formatElapsed(view.attemptElapsedMs)}</span><span>Revision {view.revision}</span>
      {#if Object.keys(view.offlineGains).length > 0}
        <aside class="gains" aria-label="Return gains">
          {#each Object.entries(view.offlineGains) as [id, gain]}<span>{id} +{gain}</span>{/each}
        </aside>
      {/if}
    </header>
    <nav aria-label="Game panels">
      {#each tabs as tab}<button type="button" class:active={view.activeTab === tab} aria-pressed={view.activeTab === tab} onclick={() => controller.selectTab(tab)}>{tab}</button>{/each}
    </nav>
    {#if view.returnStory}
      <section class="return-story" aria-live="polite" aria-label="Offline return">
        {#if !view.returnFastForwardComplete}
          <h1>Fast-forwarding your return</h1>
          <p>{view.returnStory.gapMs} ms elapsed. The authoritative result is already loaded behind this recap.</p>
          <button type="button" onclick={() => controller.completeReturnFastForward()}>SKIP</button>
        {:else}
          <h1>Return complete</h1>
          {#if view.returnStory.showOptionalModal}<p>An optional event is ready.</p>{/if}
          <button type="button" onclick={() => controller.dismissReturnStory()}>CONTINUE</button>
        {/if}
      </section>
    {/if}
    <section aria-label={`${view.activeTab} panel`}>
      <h1>{view.activeTab}</h1>
      {#if Object.keys(view.resources).length === 0}<p>No authoritative snapshot yet.</p>{/if}
      {#each Object.entries(view.resources) as [id, resource]}
        <article class:pulse={resource.pulse}>
          <h2>{id}</h2><output>{resource.value}</output>
          {#if resource.activityPpm > 0}<progress max="1000000" value={resource.activityPpm} aria-label={`${id} sub-unit activity`}></progress>{/if}
          {#if resource.capReasonKey}<p class="reason">Cap: {resource.capReasonKey}</p>{/if}
          {#if resource.explanation}<p class="receipt">{resource.explanation}</p>{/if}
        </article>
      {/each}
      {#each Object.entries(view.discrete) as [id, fact]}
        <article class="discrete" data-state-id={id}>
          <h2>{id}</h2><output>{String(fact)}</output>
        </article>
      {/each}
      {#each view.progress as item}
        <article class="coordinate">
          <h2>{item.stageId}</h2>
          <div class="progress-track" role="progressbar" aria-label={item.stageId} aria-valuetext={`${item.current} / ${item.target}`}>
            <span style={`width: ${item.fillPpm / 10000}%`}></span>
          </div>
          <data data-current={item.current} data-target={item.target}>{item.current} / {item.target}</data>
        </article>
      {/each}
    </section>
  </main>
{:else}
  <main class="contract" aria-labelledby="run-end-title">
    <p class="eyebrow">ATTEMPT COMPLETE</p>
    <h1 id="run-end-title">Attempt complete</h1>
    <p class="timer">RTA {formatElapsed(view.attemptElapsedMs)}</p>
    <p>The record is final.</p>
    <button type="button" onclick={() => controller.returnToContract()}>RETURN TO CONTRACT</button>
  </main>
{/if}
