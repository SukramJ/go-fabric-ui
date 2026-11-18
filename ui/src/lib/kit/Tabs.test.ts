// @vitest-environment happy-dom
// Tabs: route tabs are a navigation with aria-current, button tabs follow
// the tab pattern with a roving tabindex and arrow keys (UR-P1-ui-138).
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import Tabs from "./Tabs.svelte";

afterEach(cleanup);

describe("Tabs (UR-P1-ui-138)", () => {
  it("renders route tabs as links in a navigation, the active one aria-current, no tab role", () => {
    render(Tabs, {
      items: [
        { key: "a", label: "A", href: "#/a" },
        { key: "b", label: "B", href: "#/b" },
      ],
      active: "b",
      ariaLabel: "Pages",
    });
    expect(screen.queryAllByRole("tab")).toHaveLength(0);
    const nav = screen.getByRole("navigation", { name: "Pages" });
    const links = [...nav.querySelectorAll("a")];
    expect(links.map((a) => a.getAttribute("aria-current"))).toEqual([null, "page"]);
  });

  it("gives button tabs a roving tabindex and moves focus with the arrow keys", async () => {
    const onSelect = vi.fn();
    render(Tabs, {
      items: [
        { key: "a", label: "A" },
        { key: "b", label: "B" },
        { key: "c", label: "C" },
      ],
      active: "a",
      onSelect,
    });
    const tabs = screen.getAllByRole("tab");
    expect(tabs.map((t) => t.getAttribute("tabindex"))).toEqual(["0", "-1", "-1"]);
    tabs[0].focus();
    await fireEvent.keyDown(tabs[0], { key: "ArrowRight" });
    expect(document.activeElement).toBe(tabs[1]);
    await fireEvent.keyDown(tabs[1], { key: "End" });
    expect(document.activeElement).toBe(tabs[2]);
    await fireEvent.keyDown(tabs[2], { key: "ArrowRight" });
    expect(document.activeElement).toBe(tabs[0]);
    await fireEvent.keyDown(tabs[0], { key: "ArrowLeft" });
    expect(document.activeElement).toBe(tabs[2]);
  });
});
