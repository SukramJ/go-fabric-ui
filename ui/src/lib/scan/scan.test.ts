import { describe, expect, it } from "vitest";
import { isPayload } from "./index";

describe("isPayload", () => {
  it("accepts an onboarding payload and nothing else", () => {
    expect(isPayload("MT:Y.K9042C00KA0648G00")).toBe(true);
    expect(isPayload(" MT:Y.K9042C00KA0648G00\n")).toBe(true);
    expect(isPayload("34970112332")).toBe(false);
    expect(isPayload("https://example.com")).toBe(false);
    expect(isPayload("MT:")).toBe(false);
  });
});
