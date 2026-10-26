import { describe, expect, it } from "vitest";
import {
  distinctValues,
  fieldMatches,
  itemValues,
  listValues,
  type ListSource,
} from "./fields";

const f = (v: string | undefined): ListSource => ({
  fields: v === undefined ? {} : { l: v },
});

describe("list-valued fields", () => {
  it("splits on commas, trims and drops empty values", () => {
    expect(listValues(" Kitchen, Evening ,, ")).toEqual(["Kitchen", "Evening"]);
    expect(listValues("")).toEqual([]);
    expect(listValues(undefined)).toEqual([]);
  });

  it("reads lists first: a value with a comma stays one value", () => {
    expect(
      itemValues(
        { fields: {}, lists: { l: ["Living, dining", "", "Evening"] } },
        "l",
      ),
    ).toEqual(["Living, dining", "Evening"]);
    expect(
      itemValues(
        { fields: { l: "a, b" }, lists: { l: ["Living, dining"] } },
        "l",
      ),
    ).toEqual(["Living, dining"]);
    expect(
      itemValues({ fields: { l: "a, b" }, lists: { l: [] } }, "l"),
    ).toEqual([]);
  });

  it("falls back to the comma split when lists has no such key", () => {
    expect(itemValues({ fields: { l: "Kitchen, Evening" } }, "l")).toEqual([
      "Kitchen",
      "Evening",
    ]);
    expect(
      itemValues(
        { fields: { l: "Kitchen, Evening" }, lists: { m: ["x"] } },
        "l",
      ),
    ).toEqual(["Kitchen", "Evening"]);
  });

  it("matches by membership for a list, by equality otherwise", () => {
    expect(fieldMatches(f("Kitchen, Evening"), "l", "Evening", true)).toBe(
      true,
    );
    expect(
      fieldMatches(f("Kitchen, Evening"), "l", "Kitchen, Evening", true),
    ).toBe(false);
    expect(
      fieldMatches(f("Kitchen, Evening"), "l", "Kitchen, Evening", false),
    ).toBe(true);
    expect(fieldMatches(f("Kitchen, Evening"), "l", "Evening", false)).toBe(
      false,
    );
    const listed = {
      fields: { l: "Living, dining" },
      lists: { l: ["Living, dining"] },
    };
    expect(fieldMatches(listed, "l", "Living, dining", true)).toBe(true);
    expect(fieldMatches(listed, "l", "Living", true)).toBe(false);
  });

  it("never matches an empty value or an empty item", () => {
    expect(fieldMatches(f(""), "l", "", true)).toBe(false);
    expect(fieldMatches(f("a, , b"), "l", " ", true)).toBe(false);
    expect(fieldMatches(f(undefined), "l", "a", true)).toBe(false);
    expect(
      fieldMatches({ fields: {}, lists: { l: [""] } }, "l", "", true),
    ).toBe(false);
  });

  it("builds the options from the individual values, not the combinations", () => {
    expect(
      distinctValues(
        [f("Kitchen, Evening"), f("Evening"), f(""), f(undefined)],
        "l",
        true,
      ),
    ).toEqual(["Evening", "Kitchen"]);
    expect(
      distinctValues([f("Kitchen, Evening"), f("Evening"), f("")], "l", false),
    ).toEqual(["Evening", "Kitchen, Evening"]);
    expect(
      distinctValues(
        [
          f("Kitchen, Evening"),
          { fields: { l: "x" }, lists: { l: ["Living, dining"] } },
        ],
        "l",
        true,
      ),
    ).toEqual(["Evening", "Kitchen", "Living, dining"]);
  });
});
