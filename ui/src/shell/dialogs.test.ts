// @vitest-environment happy-dom
// The shell's confirm and create dialogs (review phase 1b).
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import Harness from "../test/Harness.svelte";
import { testUI } from "../test/support";
import ConfirmDialog from "./ConfirmDialog.svelte";
import NodeCreateDialog from "./NodeCreateDialog.svelte";
import { createConfirm, GUARD_MS } from "./confirm.svelte";

afterEach(cleanup);

describe("UR-P1-ui-104: a double click does not pass both steps", () => {
  it("ignores a confirmation that comes within the guard of the question being shown", async () => {
    let now = 1000;
    const c = createConfirm(() => now);
    const first = c.ask({ title: "Erase?" });
    now += GUARD_MS + 1;
    // The first click answers the first question; the caller asks the
    // second at once, and the double click's second click lands on it.
    c.answer(true);
    expect(await first).toBe(true);
    const second = c.ask({ title: "Really erase?" });
    now += 80;
    c.answer(true);
    expect(c.current?.title).toBe("Really erase?");
    now += GUARD_MS;
    c.answer(true);
    expect(await second).toBe(true);
  });

  it("lets Cancel through at once", async () => {
    const c = createConfirm(() => 0);
    const q = c.ask({ title: "Erase?" });
    c.answer(false);
    expect(await q).toBe(false);
  });
});

describe("UR-P1-ui-117: the shell's dialogs take the focus", () => {
  it("confirm: the focus moves to Cancel, Tab stays inside, Escape answers no, and the focus returns", async () => {
    const trigger = document.createElement("button");
    document.body.append(trigger);
    trigger.focus();
    const answer = vi.fn();
    const { rerender } = render(ConfirmDialog, { current: { title: "Erase?", destructive: true }, answer, cancelLabel: "Cancel", confirmLabel: "OK" });
    const dialog = screen.getByRole("alertdialog");
    expect(document.activeElement).toBe(screen.getByText("Cancel").closest("button"));
    // Tab and Shift+Tab stay inside (the confirm button is not armed yet).
    await fireEvent.keyDown(document.activeElement as Element, { key: "Tab" });
    expect(dialog.contains(document.activeElement)).toBe(true);
    await fireEvent.keyDown(document.activeElement as Element, { key: "Tab", shiftKey: true });
    expect(dialog.contains(document.activeElement)).toBe(true);
    await fireEvent.keyDown(dialog, { key: "Escape" });
    expect(answer).toHaveBeenCalledWith(false);
    await rerender({ current: null, answer, cancelLabel: "Cancel", confirmLabel: "OK" });
    expect(document.activeElement).toBe(trigger);
    trigger.remove();
  });

  it("create node: the focus moves to the label, Escape closes", async () => {
    const onClose = vi.fn();
    render(Harness, { ui: testUI(), page: NodeCreateDialog, pageProps: { onCreate: async () => {}, onClose } });
    expect(document.activeElement).toBe(screen.getByLabelText("shell.node_create.label"));
    await fireEvent.keyDown(screen.getByTestId("node-create-dialog"), { key: "Escape" });
    expect(onClose).toHaveBeenCalled();
  });
});

describe("UR-P1-ui-104: the confirm button is armed after the guard", () => {
  it("is disabled while the question is new", async () => {
    vi.useFakeTimers();
    try {
      render(ConfirmDialog, { current: { title: "Erase?" }, answer: () => {}, cancelLabel: "Cancel", confirmLabel: "OK" });
      expect(screen.getByText("OK").closest("button")).toBeDisabled();
      await vi.advanceTimersByTimeAsync(GUARD_MS);
      expect(screen.getByText("OK").closest("button")).not.toBeDisabled();
    } finally {
      vi.useRealTimers();
    }
  });
});
