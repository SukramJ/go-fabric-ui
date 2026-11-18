// The shell's toasts stay bounded (UR-P1-bound-111).
import { describe, expect, it } from "vitest";
import { createToasts, MAX_TOASTS, MAX_TOAST_CHARS } from "./toast.svelte";

describe("UR-P1-bound-111: error toasts are bounded", () => {
  it("keeps at most the cap, the newest, and collapses a repeated message", () => {
    const t = createToasts();
    for (let i = 0; i < 50; i++) t.toast.error(`update failed ${i}`);
    expect(t.entries).toHaveLength(MAX_TOASTS);
    expect(t.entries.at(-1)?.message).toBe("update failed 49");
    const before = t.entries.length;
    t.toast.error("update failed 49");
    expect(t.entries).toHaveLength(before);
  });

  it("clamps a long message", () => {
    const t = createToasts();
    t.toast.error("x".repeat(10_000));
    expect(t.entries[0].message.length).toBeLessThanOrEqual(MAX_TOAST_CHARS);
    expect(t.entries[0].message.endsWith("…")).toBe(true);
  });
});
