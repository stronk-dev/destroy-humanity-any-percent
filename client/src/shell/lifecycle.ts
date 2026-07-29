import type { SnapshotStream } from "./contracts";
import type { ClientShellPolicy } from "./policy";

export interface ReturnStory { readonly gapMs: number; readonly fastForwardMs: number; readonly gainsInHeader: true; readonly showOptionalModal: boolean; readonly badgesOnly: true }
export function selectReturnStory(gapMs: number, optionalModalReady: boolean, policy: ClientShellPolicy): ReturnStory | undefined {
  if (!Number.isSafeInteger(gapMs) || gapMs < 0) throw new RangeError("invalid reconnect gap");
  return gapMs > policy.reconnectStoryThresholdMs ? { gapMs, fastForwardMs: Math.min(policy.catchupCeilingMs, gapMs), gainsInHeader: true, showOptionalModal: optionalModalReady, badgesOnly: true } : undefined;
}

export function bindLifecycle(documentTarget: Document, windowTarget: Window, stream: SnapshotStream, disposeWorker: () => void, visibleAfterGap: (gapMs: number) => void = () => {}, clock: () => number = () => performance.now()): () => void {
  let hiddenAt = documentTarget.visibilityState === "hidden" ? clock() : undefined;
  const visibility = () => {
    if (documentTarget.visibilityState === "hidden") { hiddenAt = clock(); return; }
    if (hiddenAt !== undefined) { visibleAfterGap(Math.max(0, Math.floor(clock() - hiddenAt))); hiddenAt = undefined; }
    stream.requestSnapshot();
  };
  const pagehide = () => stream.flush();
  const freeze = () => { disposeWorker(); stream.dispose(); };
  documentTarget.addEventListener("visibilitychange", visibility); windowTarget.addEventListener("pagehide", pagehide); documentTarget.addEventListener("freeze", freeze);
  return () => { documentTarget.removeEventListener("visibilitychange", visibility); windowTarget.removeEventListener("pagehide", pagehide); documentTarget.removeEventListener("freeze", freeze); };
}
