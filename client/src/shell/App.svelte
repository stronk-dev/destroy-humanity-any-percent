<script lang="ts">
  import { onMount } from "svelte";
  import type { ShellController, ShellTab, ShellView } from "./controller";

  let { controller }: { controller: ShellController } = $props();
  let view: ShellView = $state({ screen: "contract", activeTab: "company", revision: 0, resources: {}, discrete: {} });
  const tabs: readonly ShellTab[] = ["company", "world", "pet", "minigame"];
  onMount(() => controller.subscribe((next) => { view = next; }));
</script>

{#if view.screen === "contract"}
  <main class="contract" aria-labelledby="contract-title">
    <p class="eyebrow">HUMANITY</p>
    <h1 id="contract-title">Destroy Any%</h1>
    <p class="timer">00:00:00.00</p>
    <p>PB: — · WR: 03:58:44</p>
    <button onclick={() => controller.beginAttempt()}>BEGIN ATTEMPT</button>
    <small>Free forever. No purchases. No ads. The only thing this game harvests is the fictional planet.</small>
  </main>
{:else}
  <main class="shell">
    <header><strong>Cloud Clicker</strong><span>Revision {view.revision}</span></header>
    <nav aria-label="Game panels">
      {#each tabs as tab}<button class:active={view.activeTab === tab} aria-pressed={view.activeTab === tab} onclick={() => controller.selectTab(tab)}>{tab}</button>{/each}
    </nav>
    <section aria-live="polite" aria-label={`${view.activeTab} panel`}>
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
    </section>
  </main>
{/if}
