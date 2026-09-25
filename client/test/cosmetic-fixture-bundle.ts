import cosmetics from "../../balance/testdata/cosmetics/fixture-v1.json?raw";
import type { ReplayArtifacts } from "../src/replay";
import { petSpeciesArtifacts } from "./pet-fixture-bundle";

// Cosmetic Shop v1 fixture chain: the Pet Adoption chain plus the cosmetics
// fixture, the smallest chain that can pin Founder v24.
export const cosmeticsArtifacts: ReplayArtifacts = { ...petSpeciesArtifacts, cosmetics };
export { constantsHashArtifacts, liveArtifacts, petSpeciesArtifacts } from "./pet-fixture-bundle";
