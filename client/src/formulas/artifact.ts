import { FORMULA_SCHEMA_NAME, FORMULA_SCHEMAS } from "../api/generated/formula-schema";
import { parseCanonical } from "../numeric";

class NumberToken { constructor(readonly source: string) {} }
type Value = null | boolean | string | NumberToken | readonly Value[] | { readonly [key: string]: Value };
interface Descriptor {
  readonly kind: "object" | "array" | "string" | "integer" | "boolean" | "null" | "ref" | "oneOf";
  readonly fields?: readonly { readonly name: string; readonly required: boolean; readonly schema: Descriptor }[];
  readonly items?: Descriptor;
  readonly enum?: readonly string[];
  readonly minimum?: string;
  readonly maximum?: string;
  readonly format?: string;
  readonly ref?: string;
  readonly alternates?: readonly Descriptor[];
}
const definitions: ReadonlyMap<string, Descriptor> = new Map(FORMULA_SCHEMAS.map((row) => [row.name, row.schema]));
const invalid = (): never => { throw new SyntaxError("invalid production formula artifact"); };

// Validate only: keep the original artifact text for hashing/replay. Metadata is
// generated from the Go owner's descriptor, not a second handwritten schema.
export function validateFormulaArtifact(source: string): void {
  if (typeof source !== "string") return invalid();
  const value = parseFormulaJSON(source);
  if (!matches(definitions.get(FORMULA_SCHEMA_NAME)!, value)) return invalid();
}

function isObject(value: Value): value is { readonly [key: string]: Value } {
  return value !== null && typeof value === "object" && !Array.isArray(value) && !(value instanceof NumberToken);
}
function matches(schema: Descriptor, value: Value): boolean {
  switch (schema.kind) {
    case "object": {
      if (!isObject(value)) return false;
      const fields = schema.fields ?? [];
      if (Object.keys(value).some((key) => !fields.some((field) => field.name === key))) return false;
      return fields.every((field) => Object.hasOwn(value, field.name)
        ? matches(field.schema, value[field.name]!) : !field.required);
    }
    case "array": return Array.isArray(value) && schema.items !== undefined && value.every((item: Value) => matches(schema.items!, item));
    case "string": {
      if (typeof value !== "string" || schema.enum !== undefined && !schema.enum.includes(value)) return false;
      switch (schema.format ?? "") {
        case "": return true;
        case "sha256": return value.length === 64 && /^[0-9a-f]{64}$/u.test(value);
        case "mechanical-id": return value === value.trim() && /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u.test(value);
        case "canonical-decimal": try { parseCanonical(value); return true; } catch { return false; }
        default: return false;
      }
    }
    case "integer": {
      if (!(value instanceof NumberToken) || !/^-?(?:0|[1-9][0-9]*)$/u.test(value.source)) return false;
      const integer = BigInt(value.source);
      return (schema.minimum === undefined || integer >= BigInt(schema.minimum)) &&
        (schema.maximum === undefined || integer <= BigInt(schema.maximum));
    }
    case "boolean": return typeof value === "boolean";
    case "null": return value === null;
    case "ref": { const target = definitions.get(schema.ref ?? ""); return target !== undefined && matches(target, value); }
    case "oneOf": return (schema.alternates ?? []).filter((alternate) => matches(alternate, value)).length === 1;
  }
}

// Follow the existing raw snapshot scanner's JSON grammar, but retain integer
// tokens rather than coercing them through a JavaScript number.
// Duplicate keys retain JSON's last-value behavior, matching the Go decoder.
// The iterative stack also handles overwritten deeply nested values without
// imposing the valid schema's depth on JSON syntax. Its 10,000-container limit
// matches encoding/json's maxNestingDepth, not a new artifact operating budget.
function parseFormulaJSON(source: string): Value {
  // Raw lone UTF-16 surrogates cannot represent the owner's valid UTF-8 bytes.
  // Escaped JSON surrogate tokens remain ASCII and follow JSON string decoding.
  for (let i = 0; i < source.length; i++) {
    const unit = source.charCodeAt(i);
    if (unit >= 0xd800 && unit <= 0xdbff) {
      const next = source.charCodeAt(++i);
      if (!(next >= 0xdc00 && next <= 0xdfff)) return invalid();
    } else if (unit >= 0xdc00 && unit <= 0xdfff) return invalid();
  }
  let index = 0;
  const space = () => { while (index < source.length && " \t\n\r".includes(source[index]!)) index++; };
  const string = (): string => {
    if (source[index] !== '"') return invalid();
    const start = index++;
    while (index < source.length && source[index] !== '"') index += source[index] === "\\" ? 2 : 1;
    if (index >= source.length) return invalid();
    index++;
    return JSON.parse(source.slice(start, index)) as string;
  };
  type Frame = { kind: "object"; result: { [key: string]: Value }; next: "key" | "after"; first: boolean }
    | { kind: "array"; result: Value[]; next: "value" | "after"; first: boolean };
  const stack: Frame[] = [];
  const value = (): Value => {
    space();
    const char = source[index];
    if ((char === "{" || char === "[") && stack.length >= 10000) return invalid();
    if (char === "{") {
      index++;
      const result: { [key: string]: Value } = Object.create(null) as { [key: string]: Value };
      stack.push({kind: "object", result, next: "key", first: true});
      return result;
    }
    if (char === "[") {
      index++;
      const result: Value[] = [];
      stack.push({kind: "array", result, next: "value", first: true});
      return result;
    }
    if (char === '"') return string();
    for (const [token, decoded] of [["true", true], ["false", false], ["null", null]] as const) {
      if (source.startsWith(token,index)) { index += token.length; return decoded; }
    }
    const number = /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?/u.exec(source.slice(index));
    if (number === null) return invalid();
    index += number[0].length;
    return new NumberToken(number[0]);
  };
  const result = value();
  while (stack.length !== 0) {
    const frame = stack[stack.length-1]!; space();
    const close = frame.kind === "object" ? "}" : "]";
    if (frame.next === "after") {
      if (source[index] === close) { index++; stack.pop(); continue; }
      if (source[index++] !== ",") return invalid();
      frame.next = frame.kind === "object" ? "key" : "value";
      continue;
    }
    if (frame.first && source[index] === close) { index++; stack.pop(); continue; }
    frame.first = false;
    frame.next = "after";
    if (frame.kind === "object") {
      const key = string(); space();
      if (source[index++] !== ":") return invalid();
      frame.result[key] = value();
    } else frame.result.push(value());
  }
  space();
  if (index !== source.length) return invalid();
  return result;
}
