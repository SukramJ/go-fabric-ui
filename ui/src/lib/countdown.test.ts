import { describe, expect, it } from "vitest";
import { secondsUntil } from "./countdown";

describe("secondsUntil", () => {
  const now = Date.parse("2026-10-25T12:00:00Z");

  it("counts whole seconds to the deadline, rounding a started second up", () => {
    expect(secondsUntil("2026-10-25T12:05:00Z", now)).toBe(300);
    expect(secondsUntil("2026-10-25T12:05:00Z", now + 1500)).toBe(299);
  });

  it("is 0 once the deadline has passed", () => {
    expect(secondsUntil("2026-10-25T11:59:00Z", now)).toBe(0);
  });

  it("is null without a deadline or for one that does not parse", () => {
    expect(secondsUntil(undefined, now)).toBeNull();
    expect(secondsUntil("", now)).toBeNull();
    expect(secondsUntil("soon", now)).toBeNull();
  });
});
