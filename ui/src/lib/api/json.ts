// Lossless JSON for the generic values of the contract (an attribute's or
// command's value, an event's data): go-fabric writes a 64-bit integer in
// full as a plain decimal, and JSON.parse would round one above 2^53 — a
// node or fabric id in an ACL subject or a Binding among them. Such an
// integer is kept as its source text: as JSON.rawJSON where the engine has
// it (JSON.stringify then prints it exactly), else as a JsonText. The
// client's own serialiser writes either back verbatim, and refuses a JS
// number that is an integer outside the safe range, since it may already
// have been rounded.

type RawJSONApi = { rawJSON?: (text: string) => object; isRawJSON?: (v: unknown) => boolean };
const native = JSON as unknown as RawJSONApi;

/** Which engine features the parser uses; a test turns them off to run the fallback. */
export const jsonFeatures = { rawJSON: typeof native.rawJSON === "function" && typeof native.isRawJSON === "function" };

/** A JSON value held as its source text, written back verbatim by stringifyJson. */
export class JsonText {
  readonly rawJSON: string;
  constructor(text: string) {
    this.rawJSON = text;
  }
  /** JSON.stringify cannot print raw text: a JsonText prints as a string there. */
  toJSON(): string {
    return this.rawJSON;
  }
  toString(): string {
    return this.rawJSON;
  }
}

/** Whether v is a value kept as its source text (JsonText or JSON.rawJSON). */
export function isRawJson(v: unknown): v is { readonly rawJSON: string } {
  if (v instanceof JsonText) return true;
  return typeof native.isRawJSON === "function" && native.isRawJSON(v);
}

function raw(text: string): unknown {
  return jsonFeatures.rawJSON && native.rawJSON ? native.rawJSON(text) : new JsonText(text);
}

/** Whether a parsed number would not be its source: an integer beyond 2^53, or out of range. */
function inexact(n: number): boolean {
  return !Number.isFinite(n) || (Number.isInteger(n) && !Number.isSafeInteger(n));
}

// An integer that may be unsafe has 16 digits or more, or an exponent.
const MAYBE_UNSAFE = /\d{16}|\d[eE][+-]?\d/;

const WS = /[ \t\n\r]*/y;
const NUM = /-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/y;
// eslint-disable-next-line no-control-regex
const STR = /"(?:[^"\\\u0000-\u001f]|\\(?:["\\/bfnrt]|u[0-9a-fA-F]{4}))*"/y;
const LIT = /true|false|null/y;

/**
 * Parses text as JSON. Inside a member named in `keep` (at any depth
 * below it), an integer JSON.parse would round is kept as its source text;
 * with keep === true, everywhere. Without keep, JSON.parse.
 */
export function parseJson(text: string, keep?: readonly string[] | true): unknown {
  if (!keep || !MAYBE_UNSAFE.test(text)) return JSON.parse(text);
  const keys = keep === true ? null : new Set(keep);
  let pos = 0;
  const fail = (): never => {
    throw new SyntaxError(`JSON: unexpected input at position ${pos}`);
  };
  const ws = () => {
    WS.lastIndex = pos;
    WS.exec(text);
    pos = WS.lastIndex;
  };
  const token = (re: RegExp): string | null => {
    re.lastIndex = pos;
    const m = re.exec(text);
    if (!m) return null;
    pos = re.lastIndex;
    return m[0];
  };
  const value = (scoped: boolean): unknown => {
    ws();
    const c = text[pos];
    if (c === "{") {
      pos++;
      const out: Record<string, unknown> = {};
      ws();
      if (text[pos] === "}") {
        pos++;
        return out;
      }
      for (;;) {
        ws();
        const k = token(STR);
        if (k === null) fail();
        const key = JSON.parse(k as string) as string;
        ws();
        if (text[pos] !== ":") fail();
        pos++;
        const v = value(scoped || (keys !== null && keys.has(key)));
        // A "__proto__" member is an own property, as JSON.parse makes it.
        Object.defineProperty(out, key, { value: v, enumerable: true, writable: true, configurable: true });
        ws();
        if (text[pos] === ",") {
          pos++;
          continue;
        }
        if (text[pos] === "}") {
          pos++;
          return out;
        }
        fail();
      }
    }
    if (c === "[") {
      pos++;
      const out: unknown[] = [];
      ws();
      if (text[pos] === "]") {
        pos++;
        return out;
      }
      for (;;) {
        out.push(value(scoped));
        ws();
        if (text[pos] === ",") {
          pos++;
          continue;
        }
        if (text[pos] === "]") {
          pos++;
          return out;
        }
        fail();
      }
    }
    if (c === '"') {
      const s = token(STR);
      return s === null ? fail() : JSON.parse(s);
    }
    const n = token(NUM);
    if (n !== null) {
      const v = Number(n);
      return scoped && inexact(v) ? raw(n) : v;
    }
    const l = token(LIT);
    if (l !== null) return l === "true" ? true : l === "false" ? false : null;
    return fail();
  };
  const out = value(keys === null);
  ws();
  if (pos !== text.length) fail();
  return out;
}

/**
 * The value text as written, after checking that it is one JSON value: what
 * a page sends where the operator typed the JSON (an attribute's value).
 */
export function jsonText(text: string): JsonText {
  const t = text.trim();
  parseJson(t, true);
  return new JsonText(t);
}

/**
 * JSON.stringify for a request body, with a value kept as source text
 * written verbatim and a bigint as its digits. A JS number that is an
 * integer outside the safe range is refused (RangeError): it may already
 * have been rounded, and the node would get another value than meant.
 * `space` indents as JSON.stringify's does.
 */
export function stringifyJson(value: unknown, space = 0): string {
  const out = ser(value, "", space > 0 ? " ".repeat(space) : "");
  return out === undefined ? "null" : out;
}

function ser(v: unknown, indent: string, step: string): string | undefined {
  if (v === null) return "null";
  if (isRawJson(v)) return v.rawJSON;
  switch (typeof v) {
    case "string":
      return JSON.stringify(v);
    case "number":
      if (Number.isInteger(v) && !Number.isSafeInteger(v)) {
        throw new RangeError(`the integer ${v} is outside the safe range and may have been rounded; send it as JSON text`);
      }
      return JSON.stringify(v);
    case "boolean":
      return v ? "true" : "false";
    case "bigint":
      return v.toString();
    case "object":
      break;
    default:
      return undefined;
  }
  const o = v as { toJSON?: () => unknown };
  if (typeof o.toJSON === "function") return ser(o.toJSON(), indent, step);
  const inner = indent + step;
  const sep = step ? ",\n" + inner : ",";
  const open = step ? "\n" + inner : "";
  const close = step ? "\n" + indent : "";
  if (Array.isArray(v)) {
    if (v.length === 0) return "[]";
    return "[" + open + v.map((x) => ser(x, inner, step) ?? "null").join(sep) + close + "]";
  }
  const parts: string[] = [];
  for (const [k, x] of Object.entries(v as Record<string, unknown>)) {
    const s = ser(x, inner, step);
    if (s !== undefined) parts.push(JSON.stringify(k) + (step ? ": " : ":") + s);
  }
  if (parts.length === 0) return "{}";
  return "{" + open + parts.join(sep) + close + "}";
}
