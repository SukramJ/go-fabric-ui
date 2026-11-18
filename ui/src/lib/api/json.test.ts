// UR-P1-ctr-101 (with -162): 64-bit integers in the generic values travel
// exactly — parsed as their source text, written back verbatim.
import { afterEach, describe, expect, it } from "vitest";
import { createClient } from "./client";
import { isRawJson, jsonFeatures, jsonText, JsonText, parseJson, stringifyJson } from "./json";

const BIG = "18446744073709551615";
const SUBJECT = "18364758544493064721";

afterEach(() => {
  jsonFeatures.rawJSON = typeof (JSON as unknown as { rawJSON?: unknown }).rawJSON === "function";
});

describe("parseJson", () => {
  for (const mode of ["native rawJSON", "fallback JsonText"] as const) {
    it(`keeps an integer beyond 2^53 under a kept member exactly (${mode})`, () => {
      jsonFeatures.rawJSON = mode === "native rawJSON";
      const text = `{"values":[{"value":{"subjects":[${BIG},5]},"n":${BIG}}]}`;
      const v = parseJson(text, ["value"]) as { values: { value: { subjects: unknown[] }; n: number }[] };
      const subject = v.values[0].value.subjects[0];
      expect(isRawJson(subject)).toBe(true);
      expect(mode === "fallback JsonText" ? subject instanceof JsonText : !(subject instanceof JsonText)).toBe(true);
      expect(v.values[0].value.subjects[1]).toBe(5);
      // Outside a kept member a number stays a number, as JSON.parse makes it.
      expect(typeof v.values[0].n).toBe("number");
      expect(stringifyJson(v.values[0].value)).toBe(`{"subjects":[${BIG},5]}`);
    });
  }

  it("prints exactly through JSON.stringify where the engine has JSON.rawJSON", () => {
    const v = parseJson(`{"value":[${SUBJECT}]}`, ["value"]);
    expect(JSON.stringify(v)).toBe(`{"value":[${SUBJECT}]}`);
  });

  it("parses what JSON.parse parses, and refuses what it refuses", () => {
    const text = `{"a":[1,-2.5,1e3,true,false,null,"s\\u00e9\\n"],"__proto__":{"x":1},"b":{},"c":[], "d" : ${BIG} }`;
    const mine = parseJson(text, true) as Record<string, unknown>;
    const theirs = JSON.parse(text) as Record<string, unknown>;
    expect(mine.a).toEqual(theirs.a);
    expect(Object.hasOwn(mine, "__proto__")).toBe(true);
    expect(stringifyJson(mine.d)).toBe(BIG);
    for (const bad of ["{", "[1,]", '{"a":1,}', "01", "1 2", "'s'", '"\u0001"', `[${BIG}`]) {
      expect(() => parseJson(bad, true), bad).toThrow(SyntaxError);
    }
  });
});

describe("stringifyJson", () => {
  it("refuses an integer number outside the safe range, and writes a bigint as its digits", () => {
    expect(() => stringifyJson({ value: 2 ** 60 })).toThrow(RangeError);
    expect(stringifyJson({ value: 2n ** 64n - 1n })).toBe(`{"value":${BIG}}`);
    expect(stringifyJson({ a: undefined, b: [undefined], c: 1.5 })).toBe('{"b":[null],"c":1.5}');
    expect(stringifyJson({ a: [1, { b: 2 }] }, 2)).toBe(JSON.stringify({ a: [1, { b: 2 }] }, null, 2));
  });

  it("writes JSON text verbatim, after checking that it is JSON", () => {
    expect(stringifyJson({ value: jsonText(` [${BIG}] `) })).toBe(`{"value":[${BIG}]}`);
    expect(() => jsonText("[1,")).toThrow(SyntaxError);
  });
});

describe("the client's generic values", () => {
  function client() {
    const bodies: string[] = [];
    const c = createClient("/api/v1", async (_url, init) => {
      if (init?.body) bodies.push(String(init.body));
      if (init?.method === "PUT") return new Response('{"status":0}', { status: 200 });
      return new Response(
        `{"values":[{"endpoint":0,"cluster_hex":"0x001F","attribute_hex":"0x0000","value":[{"1":5,"2":2,"3":[${SUBJECT}],"4":null,"254":1}]}]}`,
        { status: 200 },
      );
    });
    return { c, bodies };
  }

  it("reads an ACL subject beyond 2^53 exactly and writes it back unchanged", async () => {
    const { c, bodies } = client();
    const res = await c.readAttributes("0000000000000011", [{ endpoint: 0, cluster_hex: "0x001F", attribute_hex: "0x0000" }]);
    const value = res.values[0].value;
    expect(JSON.stringify(value)).toContain(SUBJECT);
    await c.writeAttribute("0000000000000011", { endpoint: 0, cluster_hex: "0x001F", attribute_hex: "0x0000", value });
    expect(bodies.at(-1)).toContain(`[${SUBJECT}]`);
  });

  it("refuses to send a rounded integer on the invoke path", async () => {
    const { c, bodies } = client();
    const n = bodies.length;
    await expect(c.invokeCommand("0000000000000011", { endpoint: 1, cluster_hex: "0x0006", command_hex: "0x00", payload: { id: 2 ** 60 } })).rejects.toThrow(
      RangeError,
    );
    expect(bodies.length).toBe(n);
  });
});

describe("the stream's cluster_event", () => {
  it("keeps an integer beyond 2^53 in the event's data exact", async () => {
    const { subscribeStream } = await import("./stream");
    const listeners = new Map<string, (ev: MessageEvent) => void>();
    const got: unknown[] = [];
    subscribeStream("/s", (ev) => got.push(ev), () => ({ addEventListener: (t, fn) => listeners.set(t, fn), close: () => {} }));
    listeners.get("cluster_event")?.({ data: `{"number":3,"data":{"node":${BIG}}}` } as MessageEvent);
    expect(stringifyJson((got[0] as { data: { data: unknown } }).data.data)).toBe(`{"node":${BIG}}`);
  });
});
