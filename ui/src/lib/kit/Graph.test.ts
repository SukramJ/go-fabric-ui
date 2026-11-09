// @vitest-environment happy-dom
// The topology graph component on the node-topology.200 fixture's graph:
// one node per graph node with its state, selection into the side panel,
// the keyboard (focus selects, Enter follows the ref, Escape clears) and
// pan and zoom on the view box.
import { afterEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { cleanup, fireEvent, render, screen, within } from "@testing-library/svelte";
import { body } from "../../test/support";
import type { TopologyGraph, TopologyNode } from "../api/client";
import Graph from "./Graph.svelte";

afterEach(cleanup);

const graph = () => body<TopologyGraph>("node-topology.200");
const hrefOf = (n: TopologyNode) => (n.ref ? `#/node/${n.ref.page}` : undefined);
const node = (id: string) => screen.getAllByTestId("graph-node").find((g) => g.getAttribute("data-id") === id)!;
// happy-dom's PointerEvent and WheelEvent drop clientX/clientY from their
// init; a mouse event carrying a pointerId (or a deltaY) is what the
// handlers read.
async function pointer(el: Element, type: string, pointerId: number, clientX = 0, clientY = 0) {
  const e = new MouseEvent(type, { bubbles: true, clientX, clientY });
  Object.defineProperty(e, "pointerId", { value: pointerId });
  el.dispatchEvent(e);
  await tick();
}
const viewBox = () => screen.getByTestId("graph").getAttribute("viewBox")!.split(" ").map(Number);

describe("Graph", () => {
  it("draws every node with its state and an accessible name, and every edge", () => {
    render(Graph, { graph: graph(), hrefOf, label: "Topology" });
    expect(screen.getByRole("group", { name: "Topology" })).toBeInTheDocument();
    const nodes = screen.getAllByTestId("graph-node");
    expect(nodes.map((n) => n.getAttribute("data-id"))).toEqual(graph().nodes.map((n) => n.id));
    expect(node("ep:3")).toHaveAttribute("data-state", "unreachable");
    expect(node("session:19")).toHaveAttribute("data-state", "idle");
    expect(node("ep:2")).toHaveAttribute("aria-label", "Device Kitchen lamp, OK");
    expect(node("ep:2").querySelector("rect")).toHaveAttribute("stroke", "var(--ha-primary-color)");
    expect(node("ep:3").querySelector("rect")).toHaveAttribute("stroke", "var(--ha-error-color)");
    expect(node("ep:4").querySelector("rect")).toHaveAttribute("stroke-dasharray", "4 3");
    expect(document.querySelectorAll("path[data-edge-kind]")).toHaveLength(graph().edges.length);
    expect(document.querySelectorAll('path[data-edge-kind="session"]')).toHaveLength(3);
  });

  it("draws a link by its strength — thicker is stronger, a dead one dotted — with its quality as the title", () => {
    render(Graph, { graph: body<TopologyGraph>("controller-topology.200"), label: "Network" });
    const link = (s: string) => document.querySelector(`path[data-strength="${s}"]`)!;
    expect([...document.querySelectorAll('path[data-edge-kind="link"]')].map((p) => p.getAttribute("data-strength"))).toEqual([
      "strong",
      "medium",
      "weak",
      "none",
      "unknown",
      "strong",
    ]);
    expect(["strong", "medium", "weak", "none", "unknown"].map((s) => link(s).getAttribute("stroke-width"))).toEqual(["3", "2", "1", "1", "1"]);
    expect(link("none")).toHaveAttribute("stroke-dasharray", "1 4");
    expect(link("none")).toHaveAttribute("stroke", "var(--ha-divider-color)");
    expect(link("unknown")).not.toHaveAttribute("stroke-dasharray");
    expect(link("strong")).toHaveAttribute("stroke", "var(--ha-secondary-text-color)");
    expect(link("strong").querySelector("title")).toHaveTextContent("Strong · LQI 3 · RSSI -45 dBm · LQI 3/2");
    expect(link("none").querySelector("title")).toHaveTextContent("No link · LQI 0");
    const session = document.querySelector('path[data-edge-kind="session"]')!;
    expect(session).toHaveAttribute("stroke-dasharray", "5 4");
    expect(session).not.toHaveAttribute("data-strength");
    expect(session.querySelector("title")).toBeNull();
  });

  it("shows the selected node's details and link in the side panel and clears it with Escape", async () => {
    render(Graph, { graph: graph(), hrefOf, label: "Topology" });
    const panel = screen.getByTestId("graph-panel");
    expect(panel).toHaveTextContent("Select a node to see its details.");
    await fireEvent.click(node("fabric:1"));
    expect(node("fabric:1")).toHaveAttribute("aria-pressed", "true");
    expect(within(panel).getByRole("heading")).toHaveTextContent("Home");
    expect(panel).toHaveTextContent("Fabric");
    expect(panel).toHaveTextContent("Vendor");
    expect(panel).toHaveTextContent("Apple Home");
    expect(within(panel).getByRole("link", { name: "Open the page" })).toHaveAttribute("href", "#/node/fabrics");
    await fireEvent.keyDown(node("fabric:1"), { key: "Escape" });
    expect(panel).toHaveTextContent("Select a node to see its details.");
  });

  it("selects the focused node, follows its ref on Enter and selects without following on Space", async () => {
    const follow = vi.fn();
    render(Graph, { graph: graph(), hrefOf, follow, label: "Topology" });
    await fireEvent.focus(node("session:17"));
    expect(screen.getByTestId("graph-panel")).toHaveTextContent("Peer node id");
    await fireEvent.keyDown(node("ep:2"), { key: " " });
    expect(follow).not.toHaveBeenCalled();
    expect(node("ep:2")).toHaveAttribute("aria-pressed", "true");
    await fireEvent.keyDown(node("ep:2"), { key: "Enter" });
    expect(follow).toHaveBeenCalledWith("#/node/endpoints");
    expect(node("ep:2")).toHaveAttribute("tabindex", "0");
  });

  it("zooms with the buttons and the wheel, pans by dragging and fits again", async () => {
    render(Graph, { graph: graph(), hrefOf, label: "Topology" });
    expect(viewBox()).toEqual([0, 0, 1053, 184]);
    await fireEvent.click(screen.getByRole("button", { name: "Zoom in" }));
    const [, , w1] = viewBox();
    expect(w1).toBeLessThan(1053);
    await fireEvent.click(screen.getByRole("button", { name: "Zoom out" }));
    expect(viewBox()[2]).toBeCloseTo(1053);
    const wheel = new MouseEvent("wheel", { bubbles: true, cancelable: true, clientX: 0, clientY: 0 });
    Object.defineProperty(wheel, "deltaY", { value: 100 });
    screen.getByTestId("graph").dispatchEvent(wheel);
    await tick();
    expect(wheel.defaultPrevented).toBe(true);
    expect(viewBox()[2]).toBeGreaterThan(1053);
    const svg = screen.getByTestId("graph");
    const before = viewBox();
    await pointer(svg, "pointerdown", 1, 100, 100);
    await pointer(svg, "pointermove", 1, 60, 90);
    await pointer(svg, "pointerup", 1);
    const after = viewBox();
    expect(after[0]).toBeGreaterThan(before[0]);
    expect(after[1]).toBeGreaterThan(before[1]);
    await fireEvent.click(screen.getByRole("button", { name: "Fit" }));
    expect(viewBox()).toEqual([0, 0, 1053, 184]);
  });

  it("does not pan from a pointer on a node", async () => {
    render(Graph, { graph: graph(), hrefOf, label: "Topology" });
    const svg = screen.getByTestId("graph");
    await pointer(node("node").querySelector("rect")!, "pointerdown", 2, 100, 100);
    await pointer(svg, "pointermove", 2, 10, 10);
    expect(viewBox()).toEqual([0, 0, 1053, 184]);
  });
});
