import type { SnapshotStream } from "./contracts";
import type { ClientShellPolicy } from "./policy";

export interface ReturnStory { readonly fastForwardMs: number; readonly gainsInHeader: true; readonly showRipeQuarter: boolean; readonly badgesOnly: true }
export function selectReturnStory(gapMs: number, ripeQuarter: boolean, policy: ClientShellPolicy): ReturnStory | undefined {
  if (!Number.isSafeInteger(gapMs) || gapMs < 0) throw new RangeError("invalid reconnect gap");
  return gapMs > policy.reconnectStoryThresholdMs ? { fastForwardMs: Math.min(policy.catchupCeilingMs, gapMs), gainsInHeader: true, showRipeQuarter: ripeQuarter, badgesOnly: true } : undefined;
}

export function bindLifecycle(documentTarget: Document, windowTarget: Window, stream: SnapshotStream, disposeWorker: () => void): () => void {
  const visibility = () => { if (documentTarget.visibilityState === "visible") stream.requestSnapshot(); };
  const pagehide = () => stream.flush();
  const freeze = () => { disposeWorker(); stream.dispose(); };
  documentTarget.addEventListener("visibilitychange", visibility); windowTarget.addEventListener("pagehide", pagehide); documentTarget.addEventListener("freeze", freeze);
  return () => { documentTarget.removeEventListener("visibilitychange", visibility); windowTarget.removeEventListener("pagehide", pagehide); documentTarget.removeEventListener("freeze", freeze); };
}
