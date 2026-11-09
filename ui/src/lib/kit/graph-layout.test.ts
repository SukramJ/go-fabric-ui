// The layered layout as a golden test: the positions of the
// node-topology.200 fixture's graph (a bridge with two devices, one with a
// part, two fabrics, three sessions incl. one PASE) are fixed here, so a
// change to the layout shows as a diff of numbers.
import { describe, expect, it } from "vitest";
import { body } from "../../test/support";
import type { TopologyGraph } from "../api/client";
import { columnsOf, fitText, LAYOUT, layoutGraph, nodeWidth } from "./graph-layout";

describe("layoutGraph", () => {
  it("lays out the node-topology.200 graph: controllers left of the node, the bridge right of it", () => {
    const g = body<TopologyGraph>("node-topology.200");
    const l = layoutGraph(g.nodes, g.edges);
    expect({ width: l.width, height: l.height }).toEqual({ width: 1053, height: 184 });
    expect(l.nodes.map((n) => [n.id, n.column, n.row, n.x, n.y, n.width, n.text])).toEqual([
      ["node", 2, 0, 396, 72, 96, "Node"],
      ["ep:1", 3, 0, 564, 72, 98, "Aggregator"],
      ["ep:2", 4, 0, 734, 44, 112, "Kitchen lamp"],
      ["ep:3", 4, 1, 738, 100, 105, "Hall sensor"],
      ["ep:4", 5, 0, 918, 72, 119, "Hall humidity"],
      ["fabric:1", 0, 0, 16, 44, 96, "Home"],
      ["fabric:2", 0, 1, 16, 100, 96, "Google"],
      ["session:17", 1, 0, 184, 16, 140, "000000000001B669"],
      ["session:19", 1, 1, 184, 72, 140, "0000000000000077"],
      ["session:18", 1, 2, 206, 128, 96, "PASE"],
    ]);
    expect(l.edges.map((e) => [e.source, e.target, e.path])).toEqual([
      ["node", "ep:1", "M492 92 C528 92 528 92 564 92"],
      ["ep:1", "ep:2", "M662 92 C698 92 698 64 734 64"],
      ["ep:1", "ep:3", "M662 92 C700 92 700 120 738 120"],
      ["ep:3", "ep:4", "M843 120 C880.5 120 880.5 92 918 92"],
      ["fabric:1", "session:17", "M112 64 C148 64 148 36 184 36"],
      ["session:17", "node", "M324 36 C360 36 360 92 396 92"],
      ["fabric:2", "session:19", "M112 120 C148 120 148 92 184 92"],
      ["session:19", "node", "M324 92 C360 92 360 92 396 92"],
      ["session:18", "node", "M302 148 C349 148 349 92 396 92"],
    ]);
  });

  it("lays out the controller-topology.200 graph: the controller, its nodes, then what they report", () => {
    const g = body<TopologyGraph>("controller-topology.200");
    const l = layoutGraph(g.nodes, g.edges);
    expect({ width: l.width, height: l.height }).toEqual({ width: 568, height: 240 });
    expect(l.nodes.map((n) => [n.id, n.column, n.row, n.x, n.y, n.width, n.text])).toEqual([
      ["node", 0, 0, 16, 100, 98, "Controller"],
      ["peer:0000000000000011", 1, 0, 215, 16, 96, "Light"],
      ["peer:0000000000000012", 1, 1, 215, 72, 96, "Sensor"],
      ["peer:0000000000000013", 1, 2, 186, 128, 154, "0x0000000000000013"],
      ["peer:0000000000000014", 1, 3, 215, 184, 96, "Bridge"],
      ["unknown_0011223344556677", 2, 0, 412, 16, 140, "0011223344556677"],
      ["unknown_8899AABBCCDDEEFF", 2, 1, 412, 72, 140, "8899AABBCCDDEEFF"],
      ["ap_AABBCCDDEEFF", 2, 2, 434, 128, 96, "home-wifi"],
      ["br_CAFE0000CAFE0001", 2, 3, 426, 184, 112, "otbr-kitchen"],
    ]);
    expect(l.edges.map((e) => [e.source, e.target, e.path])).toEqual([
      ["node", "peer:0000000000000011", "M114 120 C164.5 120 164.5 36 215 36"],
      ["node", "peer:0000000000000012", "M114 120 C164.5 120 164.5 92 215 92"],
      ["node", "peer:0000000000000013", "M114 120 C150 120 150 148 186 148"],
      ["node", "peer:0000000000000014", "M114 120 C164.5 120 164.5 204 215 204"],
      ["peer:0000000000000011", "peer:0000000000000012", "M263 56 C263 64 263 64 263 72"],
      ["peer:0000000000000011", "unknown_0011223344556677", "M311 36 C361.5 36 361.5 36 412 36"],
      ["peer:0000000000000011", "unknown_8899AABBCCDDEEFF", "M311 36 C361.5 36 361.5 92 412 92"],
      ["peer:0000000000000012", "unknown_0011223344556677", "M311 92 C361.5 92 361.5 36 412 36"],
      ["peer:0000000000000013", "ap_AABBCCDDEEFF", "M340 148 C387 148 387 148 434 148"],
      ["peer:0000000000000011", "br_CAFE0000CAFE0001", "M311 36 C368.5 36 368.5 204 426 204"],
    ]);
  });

  it("puts a node no edge reaches in the column of its kind, else left of every other", () => {
    const nodes = [
      { id: "node", kind: "node", label: "n" },
      { id: "fabric:1", kind: "fabric", label: "f1" },
      { id: "session:1", kind: "session", label: "s" },
      { id: "fabric:2", kind: "fabric", label: "lonely" },
      { id: "x", kind: "wifi_ap", label: "ap" },
    ] as const;
    const edges = [
      { source: "fabric:1", target: "session:1", kind: "contains" },
      { source: "session:1", target: "node", kind: "session" },
    ] as const;
    const cols = columnsOf([...nodes], [...edges]);
    expect(Object.fromEntries(cols)).toEqual({ node: 0, "session:1": -1, "fabric:1": -2, "fabric:2": -2, x: -3 });
  });

  it("starts at the first node without a root, ignores an edge to an unknown node and joins one column bottom to top", () => {
    const l = layoutGraph(
      [
        { id: "a", kind: "peer", label: "a" },
        { id: "b", kind: "peer", label: "b" },
        { id: "c", kind: "peer", label: "c" },
      ],
      [
        { source: "a", target: "b", kind: "link" },
        { source: "a", target: "c", kind: "link" },
        { source: "b", target: "c", kind: "link" },
        { source: "a", target: "ghost", kind: "link" },
      ],
    );
    expect(l.nodes.map((n) => n.column)).toEqual([0, 1, 1]);
    expect(l.edges).toHaveLength(3);
    expect(l.edges[2].path).toBe(`M${16 + 72 + 96 + 48} ${16 + 40} C${16 + 72 + 96 + 48} ${16 + 48} ${16 + 72 + 96 + 48} ${16 + 48} ${16 + 72 + 96 + 48} ${16 + 56}`);
  });

  it("lays out an empty graph as its margins", () => {
    expect(layoutGraph([], [])).toEqual({ width: 2 * LAYOUT.margin, height: 2 * LAYOUT.margin, nodes: [], edges: [] });
  });
});

describe("node sizes", () => {
  it("grow with the label between the minimum and the maximum, cutting a long label with an ellipsis", () => {
    expect(nodeWidth("PASE")).toBe(LAYOUT.minWidth);
    expect(nodeWidth("Kitchen lamp")).toBe(12 * LAYOUT.charWidth + 2 * LAYOUT.padX);
    const long = "A bridged device with a name far too long for one box";
    const cut = fitText(long);
    expect(cut.endsWith("…")).toBe(true);
    expect(nodeWidth(cut)).toBeLessThanOrEqual(LAYOUT.maxWidth);
    expect(fitText("short")).toBe("short");
  });
});
