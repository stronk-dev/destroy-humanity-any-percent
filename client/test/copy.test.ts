import { describe, expect, it, vi } from "vitest";

import generatedCatalog from "../src/copy/generated/catalog.json";
import { COPY_HASH, loadCopyCatalog, resolveCopy, verifyCopyCatalogHash } from "../src/copy";

function catalog(entry: Record<string, unknown>): string {
  return `${JSON.stringify({ schema_version: 1, entries: [entry] })}\n`;
}

const base = {
  key: "test.message",
  text: "Balance: {amount} {{audited}} after {count} days",
  params: [
    { name: "amount", type: "canonical_decimal" },
    { name: "count", type: "integer" },
  ],
  era_variants: { era_1995: "Ledger: {amount} {{audited}} after {count} days" },
  provenance: [],
  tone: "corporate",
};

describe("copy catalog", () => {
  it("binds the generated artifact to its independent copy hash", async () => {
    const bytes = `${JSON.stringify(generatedCatalog, null, 2)}\n`;
    await expect(verifyCopyCatalogHash(bytes, COPY_HASH)).resolves.toBe(true);
    await expect(verifyCopyCatalogHash(bytes.replace("Any%", "Other%"), COPY_HASH)).resolves.toBe(false);
    const unrelated = loadCopyCatalog(catalog(base));
    expect("copyHash" in unrelated).toBe(false);
  });

  it("round-trips typed params, literal braces, era overrides, and fallback", () => {
    const loaded = loadCopyCatalog(catalog(base));
    expect(resolveCopy(loaded, "test.message", { amount: "1.25e6", count: 2 })).toBe("Balance: 1.25e6 {audited} after 2 days");
    expect(resolveCopy(loaded, "test.message", { amount: "1.25e6", count: 2 }, "era_1995")).toBe("Ledger: 1.25e6 {audited} after 2 days");
    expect(resolveCopy(loaded, "test.message", { amount: "1.25e6", count: 2 }, "era_2000")).toBe("Balance: 1.25e6 {audited} after 2 days");
  });

  it("rejects missing, extra, and mistyped params", () => {
    const loaded = loadCopyCatalog(catalog(base));
    expect(() => resolveCopy(loaded, "test.message", { amount: "1e1" })).toThrow(/exact params/);
    expect(() => resolveCopy(loaded, "test.message", { amount: "1e1", count: 2, extra: "x" })).toThrow(/exact params/);
    expect(() => resolveCopy(loaded, "test.message", { amount: "01", count: 2 })).toThrow(/canonical Decimal/);
    expect(() => resolveCopy(loaded, "test.message", { amount: "1e1", count: 2.5 })).toThrow(/safe integer/);
  });

  it("fails loudly in development and reports one loud-key fallback in production", () => {
    const loaded = loadCopyCatalog(catalog(base));
    expect(() => resolveCopy(loaded, "missing.key", {})).toThrow(/unknown copy key/);
    const report = vi.fn();
    expect(resolveCopy(loaded, "missing.key", {}, undefined, { mode: "production", reportInvariant: report })).toBe("missing.key");
    expect(report).toHaveBeenCalledOnce();
    expect(() => resolveCopy(loaded, "missing.key", {}, undefined, { mode: "production" } as never)).toThrow(/requires an invariant reporter/);
  });

  it("rejects grammar drift before resolution", () => {
    expect(() => loadCopyCatalog(catalog({ ...base, extra: true }))).toThrow(/exact keys/);
    expect(() => loadCopyCatalog(catalog({ ...base, text: "Wrong {field}" }))).toThrow(/placeholders differ/);
    expect(() => loadCopyCatalog(catalog({ ...base, text: "<b>{amount}</b> {count}" }))).toThrow(/plain text/);
    for (const text of ["**{amount}** {count}", "`{amount}` {count}", "#", ">quoted", "# {amount} {count}", "<!-- {amount} --> {count}", "1) {amount} {count}", "     {amount} {count}", "[{amount}][source] {count}", "| {amount} | {count} |", "<!DOCTYPE html>", "<?xml version=\"1.0\"?>", "line  \nnext"]) {
      expect(() => loadCopyCatalog(catalog({ ...base, text }))).toThrow(/plain text/);
    }
  });
});
