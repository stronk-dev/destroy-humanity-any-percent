<script lang="ts">
  import type { GameUIAchievementsArm } from "../api/generated/types";
  import { applicationCopyCatalog, t, type CopyEra, type CopyKey } from "../copy";

  // GS2: read-only. State is always text, never color alone. Score is a
  // score, not Clout (OD-6).
  let { arm, era }: { arm: GameUIAchievementsArm; era: CopyEra } = $props();

  const earned = $derived(arm.rows.filter((row) => row.earned !== null).length);
  function title(row: GameUIAchievementsArm["rows"][number]): string {
    if (!applicationCopyCatalog.byKey.has(row.copy_key)) throw new RangeError(`achievement copy ${row.copy_key} is missing`);
    return t(row.copy_key as CopyKey, {}, era);
  }
  function state(row: GameUIAchievementsArm["rows"][number]): CopyKey {
    return row.earned === "run" ? "achievements.state.earned_run" : row.earned === "lifetime" ? "achievements.state.earned_lifetime" : "achievements.state.locked";
  }
</script>

<section class="surface achievements" aria-labelledby="achievements-heading">
  <h1 id="achievements-heading" tabindex="-1">{t("surface.achievements.title", {}, era)}</h1>
  <p>{t("achievements.score_frame", { run: arm.score.run, lifetime: arm.score.lifetime }, era)}</p>
  {#if arm.rows.length === 0}
    <p>{t("achievements.empty", {}, era)}</p>
  {:else}
    <p>{t("achievements.progress_frame", { earned, total: arm.rows.length }, era)}</p>
    <ul>
      {#each arm.rows as row (row.achievement_id)}
        <li data-earned={row.earned ?? "none"}>
          <h2>{title(row)}</h2>
          <span>{t(row.condition_scope === "run" ? "achievements.scope.run" : "achievements.scope.career", {}, era)}</span>
          <strong>{t(state(row), {}, era)}</strong>
          <span>{t("achievements.grant_frame", { score: row.score_grant }, era)}</span>
          {#if row.proof_kind === "possession"}<small>{t("achievement.possession_warning", {}, era)}</small>{/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .achievements { display: grid; gap: var(--cc-space-md); }
  ul { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(15rem, 100%), 1fr)); gap: var(--cc-space-sm); margin: 0; padding: 0; list-style: none; }
  li { display: grid; gap: var(--cc-space-xs); padding: var(--cc-space-sm); border: var(--cc-border-width) var(--cc-border-style) var(--cc-color-border); border-radius: var(--cc-border-radius); background: var(--cc-color-surface); }
  h1, h2, p { margin: 0; }
  h2 { font-size: inherit; }
</style>
