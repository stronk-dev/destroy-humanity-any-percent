// GS1 display phase: derived from the host's server-time estimate against the
// pinned windows. Display only; the harvest receipt decides.
export type FiscalPhase = "ripening" | "early" | "guaranteed";

export function fiscalPhase(elapsedMS: number, period: Readonly<{ early_ms: number; guaranteed_ms: number }>): FiscalPhase {
  if (elapsedMS < period.early_ms) return "ripening";
  return elapsedMS < period.guaranteed_ms ? "early" : "guaranteed";
}
