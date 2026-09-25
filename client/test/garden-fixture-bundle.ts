import fiscalFixture from "../../balance/testdata/server-garden/fiscal-fixture-v1.json?raw";
import gardenFixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import type { ReplayArtifacts } from "../src/replay";
import { cosmeticsArtifacts } from "./cosmetic-fixture-bundle";

// Server Garden fixture chain: the cosmetics chain, the fixture Fiscal artifact
// (adds the minigame.server_garden unlock row), and the fixture server_garden.
// The smallest chain that can pin Founder v25.
export const gardenArtifacts: ReplayArtifacts = { ...cosmeticsArtifacts, fiscal: fiscalFixture, server_garden: gardenFixture };
export { constantsHashArtifacts, cosmeticsArtifacts } from "./cosmetic-fixture-bundle";
