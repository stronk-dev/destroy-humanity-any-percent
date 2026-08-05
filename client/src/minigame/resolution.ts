import { integrateFixedGrid } from "../fixed-grid";
import { MAX_EXACT_INTEGER } from "../numeric";
import { offlineGradeForScore, type OfflineQualityPolicy, type OfflineQualityState } from "./offline-quality";
import type { MinigameRatingPolicy } from "./catalog";

export interface MinigameRatingState { readonly elo: number; readonly season_member: string; readonly games_counted: number }
export interface CertifiedScoreFact { readonly kind: string; readonly value: number }
export interface CertifiedMinigameResult { readonly outcome: string; readonly score_facts: readonly CertifiedScoreFact[]; readonly rating_delta: number | null }
export interface FounderResolutionTransition {
  readonly rating_before: MinigameRatingState;
  readonly rating_after: MinigameRatingState;
  readonly quality_before: OfflineQualityState;
  readonly quality_after: OfflineQualityState;
}

export function applyFounderMinigameResolution(rating: MinigameRatingState, quality: OfflineQualityState,
  result: CertifiedMinigameResult, ratingPolicy: MinigameRatingPolicy, qualityPolicy: OfflineQualityPolicy,
  founderAttendedMs: number): FounderResolutionTransition {
  safe(founderAttendedMs, 0, MAX_EXACT_INTEGER);
  safe(rating.elo, ratingPolicy.elo_floor, ratingPolicy.elo_ceiling);
  safe(rating.games_counted, 0, MAX_EXACT_INTEGER);
  if (rating.season_member !== ratingPolicy.season_member || founderAttendedMs < quality.last_founder_attended_ms) throw new RangeError("invalid resolution state");
  const score = uniqueScore(result, qualityPolicy.score_fact);
  const elapsed = founderAttendedMs - quality.last_founder_attended_ms;
  const decay = integrateFixedGrid(elapsed, qualityPolicy.decay_ppm_per_grid, quality.decay_remainder_ppm, qualityPolicy.decay_grid_ms);
  const distance = quality.grade_ppm - qualityPolicy.neutral_floor_ppm;
  if (distance < 0) throw new RangeError("quality below neutral floor");
  // Evaluate decay before replacement so live and replay reject the same stale
  // or out-of-domain carried state even though the certified grade becomes the
  // new visible value at this sample.
  if (decay.whole < BigInt(distance) && decay.remainder >= 1_000_000) throw new RangeError("quality remainder outside wire domain");
  const qualityAfter = Object.freeze({ grade_ppm: offlineGradeForScore(qualityPolicy, score),
    last_founder_attended_ms: founderAttendedMs, decay_remainder_ppm: 0 });
  let ratingAfter = rating;
  if (result.rating_delta !== null) {
    safe(result.rating_delta, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER);
    if (rating.games_counted === MAX_EXACT_INTEGER) throw new RangeError("games counted overflow");
    const sum = BigInt(rating.elo) + BigInt(result.rating_delta);
    const clamped = sum < BigInt(ratingPolicy.elo_floor) ? BigInt(ratingPolicy.elo_floor) :
      sum > BigInt(ratingPolicy.elo_ceiling) ? BigInt(ratingPolicy.elo_ceiling) : sum;
    ratingAfter = Object.freeze({ elo: Number(clamped), season_member: rating.season_member, games_counted: rating.games_counted + 1 });
  }
  return { rating_before: rating, rating_after: ratingAfter, quality_before: quality, quality_after: qualityAfter };
}

function uniqueScore(result: CertifiedMinigameResult, kind: string): number {
  if (typeof result.outcome !== "string" || result.outcome.length === 0 || !Array.isArray(result.score_facts)) throw new RangeError("invalid certified result");
  const matching = result.score_facts.filter((fact) => fact.kind === kind);
  if (matching.length !== 1) throw new RangeError("missing or duplicate quality score");
  return safe(matching[0]!.value, 0, MAX_EXACT_INTEGER);
}

function safe(value: number, minimum: number, maximum: number): number {
  if (!Number.isSafeInteger(value) || value < minimum || value > maximum) throw new RangeError("integer outside resolution domain");
  return value;
}
