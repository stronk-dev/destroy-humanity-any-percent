import { MAX_EXACT_INTEGER } from "../numeric";
import type { PetCareState } from "./state";

export interface PetCombatInputs {
  readonly pet_trust_ppm: number;
  readonly soul: number;
}

export function petCombatInputs(petTrustPpm: number, soul: number): PetCombatInputs {
  if (!Number.isSafeInteger(petTrustPpm) || petTrustPpm < 0 || petTrustPpm > 1_000_000 ||
    !Number.isSafeInteger(soul) || soul < -MAX_EXACT_INTEGER || soul > MAX_EXACT_INTEGER) {
    throw new RangeError("invalid pet combat inputs");
  }
  return { pet_trust_ppm: petTrustPpm, soul };
}

export function petCombatInputsForState(state: PetCareState, soul: number): PetCombatInputs {
  return petCombatInputs(state.trust_ppm, soul);
}
