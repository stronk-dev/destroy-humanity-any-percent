import pinnedPitchBytes from "../../../../balance/pitch.json?raw";

import { COPY_KEYS } from "../../copy";
import { parsePitchCatalog, pitchContentHash, type PitchCatalog } from "../../pitch/catalog";

// The Pitch snapshot carries only instance IDs and pitch_content_hash. The
// table resolves presentation from the pinned catalog bytes bundled with the
// client, and only when those bytes hash to the snapshot's identity: a
// mismatch is a loud content_mismatch error, never a deploy-current fallback.
export interface PitchContent { readonly hash: string; readonly catalog: PitchCatalog }

export class PitchContentMismatch extends Error {}

let pinned: Promise<PitchContent> | undefined;

export function loadPitchContent(bytes: string = pinnedPitchBytes): Promise<PitchContent> {
  const load = async (): Promise<PitchContent> => ({ hash: await pitchContentHash(bytes), catalog: parsePitchCatalog(JSON.parse(bytes), new Set(COPY_KEYS)) });
  if (bytes !== pinnedPitchBytes) return load();
  pinned ??= load();
  return pinned;
}

export async function pitchContentFor(snapshotHash: string, source: Promise<PitchContent> = loadPitchContent()): Promise<PitchContent> {
  const content = await source;
  if (content.hash !== snapshotHash) throw new PitchContentMismatch(`Pitch content ${snapshotHash} is not the bundled ${content.hash}`);
  return content;
}

export function pitchCardInstanceBase(instance: string): string {
  const match = /^([a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*)#([1-9][0-9]*)$/u.exec(instance);
  if (!match) throw new SyntaxError("invalid Pitch card instance");
  return match[1]!;
}
