/** AR3.4 raw snapshot grammar: JSON.parse erases duplicate keys and number-token spelling.
 * Values and exact field shapes are still checked by the engine's snapshot decoder. */
export function parseSnapshotJSON(source: string): unknown {
  let index = 0;
  const invalid = (): never => { throw new SyntaxError("invalid arcade snapshot JSON"); };
  const space = () => { while (index < source.length && " \t\n\r".includes(source[index]!)) index++; };
  const string = (): string => {
    if (source[index] !== '"') return invalid();
    const start = index++;
    while (index < source.length && source[index] !== '"') index += source[index] === "\\" ? 2 : 1;
    if (index >= source.length) return invalid();
    index++;
    return JSON.parse(source.slice(start, index)) as string;
  };
  const value = (depth = 0): void => {
    space();
    const char = source[index];
    // AR3.4: root → revealed array → row is the deepest legal compound shape.
    if ((char === "{" || char === "[") && depth >= 3) return invalid();
    if (char === "{") {
      index++;
      const seen = new Set<string>();
      space();
      if (source[index] === "}") { index++; return; }
      for (;;) {
        space();
        const key = string();
        if (seen.has(key)) return invalid();
        seen.add(key);
        space();
        if (source[index++] !== ":") return invalid();
        value(depth + 1);
        space();
        if (source[index] === ",") { index++; continue; }
        if (source[index] === "}") { index++; return; }
        return invalid();
      }
    }
    if (char === "[") {
      index++;
      space();
      if (source[index] === "]") { index++; return; }
      for (;;) {
        value(depth + 1);
        space();
        if (source[index] === ",") { index++; continue; }
        if (source[index] === "]") { index++; return; }
        return invalid();
      }
    }
    if (char === '"') { string(); return; }
    const rest = source.slice(index);
    const literal = /^(?:true|false|null)/u.exec(rest);
    if (literal) { index += literal[0].length; return; }
    const number = /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?/u.exec(rest);
    if (!number || /[.eE]/u.test(number[0]) || !Number.isSafeInteger(Number(number[0]))) return invalid();
    index += number[0].length;
  };
  value();
  space();
  if (index !== source.length) return invalid();
  return JSON.parse(source);
}
