// makeTextMatcher: a search term is never run as a regular expression, so
// a backtracking pattern typed into the search box cannot freeze the tab;
// `|` still separates alternatives (UR-P1-ui-142).
import { describe, expect, it } from "vitest";
import { makeTextMatcher } from "./utils";

describe("makeTextMatcher (UR-P1-ui-142)", () => {
  it("answers a catastrophic-backtracking term at once", () => {
    const text = "a".repeat(25) + "b";
    const start = performance.now();
    expect(makeTextMatcher("(a+)+$")(text)).toBe(false);
    // The regex form took seconds (8.9 s at the red run); a full parallel run
    // slows even the literal match, so the bound is generous.
    expect(performance.now() - start).toBeLessThan(1000);
  });

  it("treats | as OR between literal alternatives, case-insensitive", () => {
    const m = makeTextMatcher("lamp | LIGHT");
    expect(m("Ceiling light")).toBe(true);
    expect(m("Kitchen Lamp")).toBe(true);
    expect(m("Hall sensor")).toBe(false);
    expect(makeTextMatcher("a.c")("abc")).toBe(false);
    expect(makeTextMatcher("|")("anything")).toBe(true);
  });
});
