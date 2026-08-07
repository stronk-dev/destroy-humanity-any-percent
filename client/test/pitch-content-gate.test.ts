import { describe, expect, it } from "vitest";
import content from "../../balance/testdata/pitch-v1.json?raw";
import corpusSource from "../../testdata/pitch/content-gate-v1.json";
import { applyPitch, createPitch, pitchBestExponent, type PitchResult, type PitchSnapshot } from "../src/pitch/engine";

interface Corpus {
  readonly version: number;
  readonly pitch_content_hash: string;
  readonly transition_budget: number;
  readonly scenarios: readonly { readonly name: string; readonly seed: number; readonly commands: readonly unknown[];
    readonly expected_terminal: PitchSnapshot; readonly expected_result: PitchResult; readonly covers_cards: readonly string[];
    readonly covers_hacks: readonly string[]; readonly assertions: readonly string[] }[];
  readonly exponent_boundaries: readonly { readonly valuation: string; readonly expected: number }[];
}

const corpus = corpusSource as unknown as Corpus;
const scaling = { "minigame.pitch": 1 } as const;

describe("Pitch shared content gate", () => {
  it("byte-replays every declared terminal scenario", async () => {
    let transitions = 0;
    for (const scenario of corpus.scenarios) {
      const identity = { content, content_hash: corpus.pitch_content_hash, content_schema_version: 1 } as const;
      let snapshot = await createPitch({ ...identity, seed: BigInt(scenario.seed), mode: "solo", scaling_inputs: scaling });
      let terminal: PitchResult | null = null;
      for (let index = 0; index < scenario.commands.length; index++) {
        const output = await applyPitch({ ...identity, seed: BigInt(scenario.seed), mode: "solo", scaling_inputs: scaling,
          revision: index + 1, snapshot, command: scenario.commands[index] });
        snapshot = output.snapshot; terminal = output.result; transitions++;
      }
      expect(JSON.stringify(snapshot), scenario.name).toBe(JSON.stringify(scenario.expected_terminal));
      expect(JSON.stringify(terminal), scenario.name).toBe(JSON.stringify(scenario.expected_result));
    }
    expect(transitions).toBe(corpus.transition_budget);
  });

  it("covers every launch row and the ruled exponent boundaries", () => {
    expect([...new Set(corpus.scenarios.flatMap((row) => row.covers_cards))].sort()).toHaveLength(12);
    expect([...new Set(corpus.scenarios.flatMap((row) => row.covers_hacks))].sort()).toHaveLength(8);
    for (const vector of corpus.exponent_boundaries) expect(pitchBestExponent(vector.valuation)).toBe(vector.expected);
  });
});
