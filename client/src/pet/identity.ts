import { MAX_EXACT_INTEGER } from "../numeric";
import type { PetCatalog } from "./catalog";
import type { PetStatID } from "./grammar";
import { PET_TEMPERAMENT_ORDER, petSpeciesRow, type PetSpeciesCatalog, type PetTemperament } from "./species";
import type { PetCareState } from "./state";

// Pet Adoption v1 PA3: the immutable identity of an adopted pet, keyed like
// the care map. Byte twin of server/pet/identity.go.
export interface PetIdentity {
  readonly species_id: string;
  readonly temperament: PetTemperament;
  readonly palette_id: string;
  readonly name_key: string;
  readonly adopted_at_ms: number;
  readonly adopted_at_attended_ms: number;
}

const identityKeys = ["adopted_at_attended_ms", "adopted_at_ms", "name_key", "palette_id", "species_id", "temperament"] as const;
const petIDV7 = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u;
const speciesID = /^pet_species\.[a-z][a-z0-9_]*$/u;
const paletteID = /^pet_palette\.[a-z0-9_]+$/u;
const nameKey = /^pet\.name\.[a-z0-9_]+\.[a-z0-9_]+$/u;

export function parsePetIdentities(source: unknown, care: Readonly<Record<string, PetCareState>>, species: PetSpeciesCatalog): Record<string, PetIdentity> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError("pet_identities must be an object");
  const result: Record<string, PetIdentity> = {};
  for (const [id, raw] of Object.entries(source as Record<string, unknown>)) {
    if (raw === null || typeof raw !== "object" || Array.isArray(raw)) throw new SyntaxError("pet identity must be an object");
    const keys = Object.keys(raw).sort();
    if (keys.length !== identityKeys.length || keys.some((key, index) => key !== identityKeys[index])) throw new SyntaxError("pet identity keys are not exact");
    const row = raw as Record<string, unknown>;
    result[id] = Object.freeze({ species_id: row.species_id as string, temperament: row.temperament as PetTemperament, palette_id: row.palette_id as string,
      name_key: row.name_key as string, adopted_at_ms: row.adopted_at_ms as number, adopted_at_attended_ms: row.adopted_at_attended_ms as number });
  }
  validatePetIdentities(result, care, species);
  return result;
}

export function validatePetIdentities(identities: Readonly<Record<string, PetIdentity>>, care: Readonly<Record<string, PetCareState>>, species: PetSpeciesCatalog): void {
  const ids = Object.keys(identities);
  if (ids.length !== Object.keys(care).length) throw new SyntaxError("pet identity and care key sets differ");
  if (ids.length > species.max_pets_per_founder) throw new SyntaxError("pet identities exceed the cap");
  for (const id of ids) {
    const identity = identities[id]!, record = care[id];
    if (!record || !petIDV7.test(id)) throw new SyntaxError("pet identity has no care record or is not UUIDv7");
    if (typeof identity.species_id !== "string" || !speciesID.test(identity.species_id) || !PET_TEMPERAMENT_ORDER.includes(identity.temperament) ||
      typeof identity.palette_id !== "string" || !paletteID.test(identity.palette_id) || typeof identity.name_key !== "string" || !nameKey.test(identity.name_key)) {
      throw new SyntaxError("pet identity is malformed");
    }
    if (!Number.isSafeInteger(identity.adopted_at_ms) || identity.adopted_at_ms < 0 || identity.adopted_at_ms > MAX_EXACT_INTEGER ||
      !Number.isSafeInteger(identity.adopted_at_attended_ms) || identity.adopted_at_attended_ms < 0 || identity.adopted_at_attended_ms > record.evaluated_through_attended_ms) {
      throw new SyntaxError("pet identity has an invalid adoption coordinate");
    }
    const row = petSpeciesRow(species, identity.species_id);
    if (!row || !row.allowed_temperaments.includes(identity.temperament) || !row.palette_ids.includes(identity.palette_id) || !row.name_keys.includes(identity.name_key)) {
      throw new SyntaxError("pet identity does not resolve under the pinned pet_species");
    }
  }
}

export function encodePetIdentities(identities: Readonly<Record<string, PetIdentity>>): Record<string, PetIdentity> {
  return Object.fromEntries(Object.keys(identities).sort().map((id) => {
    const identity = identities[id]!;
    return [id, { adopted_at_attended_ms: identity.adopted_at_attended_ms, adopted_at_ms: identity.adopted_at_ms, name_key: identity.name_key,
      palette_id: identity.palette_id, species_id: identity.species_id, temperament: identity.temperament }];
  }));
}

// PA4.6.2's canonical initial care record, watermarked at the frozen
// attendance total so pre-adoption attendance never decays the pet.
export function initialPetCareState(catalog: PetCatalog, attendedMs: number): PetCareState {
  if (!Number.isSafeInteger(attendedMs) || attendedMs < 0 || attendedMs > MAX_EXACT_INTEGER) throw new RangeError("invalid initial care attendance");
  const stats = Object.fromEntries(catalog.stat_policy.stats.map((row) => [row.stat_id, row.initial_ppm])) as Record<PetStatID, number>;
  const remainders = Object.fromEntries(catalog.stat_policy.stats.map((row) => [row.stat_id, 0])) as Record<PetStatID, number>;
  return { stats_ppm: stats, stat_decay_remainders_ppm: remainders, cooldown_until_attended_ms: {}, trust_ppm: catalog.trust_policy.initial_ppm,
    trust_decay_remainder_ppm: 0, evaluated_through_attended_ms: attendedMs, behavior_state: "idle", behavior_entered_at_attended_ms: attendedMs,
    behavior_queue: [], behavior_prng_cursor: 0 };
}
