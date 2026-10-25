import { describe, expect, it } from "vitest";
import { distinctValues, fieldMatches, listValues } from "./fields";

describe("list-valued fields", () => {
  it("splits on commas, trims and drops empty values", () => {
    expect(listValues(" Kitchen, Evening ,, ")).toEqual(["Kitchen", "Evening"]);
    expect(listValues("")).toEqual([]);
    expect(listValues(undefined)).toEqual([]);
  });

  it("matches by membership for a list, by equality otherwise", () => {
    expect(fieldMatches("Kitchen, Evening", "Evening", true)).toBe(true);
    expect(fieldMatches("Kitchen, Evening", "Kitchen, Evening", true)).toBe(false);
    expect(fieldMatches("Kitchen, Evening", "Kitchen, Evening", false)).toBe(true);
    expect(fieldMatches("Kitchen, Evening", "Evening", false)).toBe(false);
  });

  it("never matches an empty value or an empty item", () => {
    expect(fieldMatches("", "", true)).toBe(false);
    expect(fieldMatches("a, , b", " ", true)).toBe(false);
    expect(fieldMatches(undefined, "a", true)).toBe(false);
  });

  it("builds the options from the individual values, not the combinations", () => {
    expect(distinctValues(["Kitchen, Evening", "Evening", "", undefined], true)).toEqual(["Evening", "Kitchen"]);
    expect(distinctValues(["Kitchen, Evening", "Evening", ""], false)).toEqual(["Evening", "Kitchen, Evening"]);
  });
});
