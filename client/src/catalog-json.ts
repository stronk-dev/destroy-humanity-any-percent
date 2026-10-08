/** Raw integer-catalog grammar shared by Garden SG1 and Pet Adoption PA2.
 * Check before JSON.parse erases duplicate keys and numeric token spelling.
 * Shape/domain validation belongs to each catalog loader. Every scanner loop
 * advances within the input; final JSON.parse rejects other malformed syntax. */
export function parseIntegerCatalogJSON(text: string, label: string): unknown {
  const scopes: { kind: "{" | "["; keys: Set<string>; expectKey: boolean }[] = [];
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index]!, top = scopes.at(-1);
    if (char === "\"") {
      let end = index + 1;
      while (end < text.length && text[end] !== "\"") end += text[end] === "\\" ? 2 : 1;
      if (end >= text.length) throw new SyntaxError(`${label} has an unfinished string`);
      if (top?.kind === "{" && top.expectKey) {
        const key = JSON.parse(text.slice(index, end + 1)) as string;
        if (top.keys.has(key)) throw new SyntaxError(`${label} has duplicate keys`);
        top.keys.add(key); top.expectKey = false;
      }
      index = end; continue;
    }
    if (char === "-" || char >= "0" && char <= "9") {
      let end = index + 1;
      while (end < text.length && /[0-9.eE+-]/u.test(text[end]!)) end += 1;
      const token = text.slice(index, end);
      if (!/^-?(?:0|[1-9][0-9]*)$/u.test(token) || !Number.isSafeInteger(Number(token))) throw new SyntaxError(`${label} requires exact safe integer tokens`);
      index = end - 1; continue;
    }
    if (char === "{" || char === "[") scopes.push({ kind: char, keys: new Set(), expectKey: char === "{" });
    else if (char === "}" || char === "]") scopes.pop();
    else if (char === "," && top?.kind === "{") top.expectKey = true;
  }
  return JSON.parse(text);
}
