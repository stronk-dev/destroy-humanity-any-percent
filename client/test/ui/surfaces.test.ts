import { describe, expect, it } from "vitest";

import { SurfaceHost, parseSurfaceRegistry, surfaceUnlocked, type SurfaceFactory } from "../../src/ui/surfaces";

describe("UI surface registry", () => {
  const known = new Set(["gate.t1", "guild.member"]);
  const source = [
    { surface_id: "company", mount_id: "mount.company", unlock: { kind: "always" } },
    { surface_id: "world", mount_id: "mount.world", unlock: { kind: "fact_equals", fact_id: "gate.t1", value: true } },
  ];

  it("loads byte-sorted rows and evaluates only committed shell facts", () => {
    const rows = parseSurfaceRegistry(source, known);
    expect(surfaceUnlocked(rows[0], {})).toBe(true);
    expect(surfaceUnlocked(rows[1], { "gate.t1": false })).toBe(false);
    expect(surfaceUnlocked(rows[1], { "gate.t1": true })).toBe(true);

    expect(() => parseSurfaceRegistry([...source].reverse(), known)).toThrow(/byte-sorted/);
    expect(() => parseSurfaceRegistry([{ ...source[0], unlock: { kind: "fact_equals", fact_id: "invented.client_fact", value: true } }], known)).toThrow(/unknown fact/);
    expect(() => parseSurfaceRegistry([source[0], { ...source[1], mount_id: "mount.company" }], known)).toThrow(/mount IDs/);
  });

  it("unsubscribes and unmounts before activating the next surface", () => {
    const rows = parseSurfaceRegistry(source, known);
    const events: string[] = [];
    const listeners = new Map<string, (value: string) => void>();
    const factory = (id: string): SurfaceFactory<string> => () => ({
      subscribe(listener) { events.push(`subscribe:${id}`); listeners.set(id, listener); return () => { events.push(`unsubscribe:${id}`); listeners.delete(id); }; },
      unmount() { events.push(`unmount:${id}`); },
    });
    const host = new SurfaceHost(rows, new Map([
      ["mount.company", {} as HTMLElement],
      ["mount.world", {} as HTMLElement],
    ]), new Map([
      ["company", factory("company")],
      ["world", factory("world")],
    ]));
    const received: string[] = [];
    host.activate("company", {}, (value) => received.push(value));
    listeners.get("company")?.("company:1");
    host.activate("world", { "gate.t1": true }, (value) => received.push(value));
    listeners.get("company")?.("company:2");
    listeners.get("world")?.("world:1");
    host.dispose();

    expect(events).toEqual([
      "subscribe:company",
      "unsubscribe:company",
      "unmount:company",
      "subscribe:world",
      "unsubscribe:world",
      "unmount:world",
    ]);
    expect(received).toEqual(["company:1", "world:1"]);
  });
});
