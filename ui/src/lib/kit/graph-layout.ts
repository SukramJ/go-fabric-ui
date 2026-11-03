// The layered layout of a topology graph (GET /node/topology,
// GET /controller/topology): a pure function, so its positions are a
// golden test. No graph library (concept v13 E.4): columns by signed depth
// from the root, rows in input order, cubic edge paths, node sizes from
// the label's length.
import type { TopologyEdge, TopologyNode } from "../api/client";

/** Every node kind of the contract, for the translations a page needs. */
export const TOPOLOGY_KINDS = [
  "node",
  "aggregator",
  "device",
  "part",
  "fabric",
  "session",
  "peer",
  "thread_router",
  "thread_end_device",
  "wifi_ap",
  "ethernet",
] as const satisfies readonly TopologyNode["kind"][];

/** Every node state of the contract. */
export const TOPOLOGY_STATES = ["ok", "unreachable", "idle", "unknown"] as const satisfies readonly TopologyNode["state"][];

/** Every link strength of the contract, strongest first. */
export const TOPOLOGY_STRENGTHS = ["strong", "medium", "weak", "none", "unknown"] as const satisfies readonly NonNullable<TopologyEdge["strength"]>[];

/** The measures the layout works in, in SVG user units. */
export const LAYOUT = {
  charWidth: 7,
  padX: 14,
  minWidth: 96,
  maxWidth: 200,
  nodeHeight: 40,
  columnGap: 72,
  rowGap: 16,
  margin: 16,
} as const;

export type LayoutInputNode = Pick<TopologyNode, "id" | "kind" | "label">;
export type LayoutInputEdge = Pick<TopologyEdge, "source" | "target" | "kind">;

export interface LaidNode {
  id: string;
  /** The label as drawn, cut with an ellipsis to fit maxWidth. */
  text: string;
  column: number;
  row: number;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface LaidEdge {
  source: string;
  target: string;
  kind: string;
  /** An SVG path: a cubic curve between the two nodes' facing sides. */
  path: string;
}

export interface GraphLayout {
  width: number;
  height: number;
  nodes: LaidNode[];
  edges: LaidEdge[];
}

/** The label cut to the characters that fit maxWidth. */
export function fitText(label: string): string {
  const max = Math.floor((LAYOUT.maxWidth - 2 * LAYOUT.padX) / LAYOUT.charWidth);
  return label.length <= max ? label : label.slice(0, max - 1) + "…";
}

/** A node's width from its drawn text, within minWidth and maxWidth. */
export function nodeWidth(text: string): number {
  return Math.min(LAYOUT.maxWidth, Math.max(LAYOUT.minWidth, text.length * LAYOUT.charWidth + 2 * LAYOUT.padX));
}

/**
 * Columns: the root is 0; following an edge from its source moves one
 * column right, from its target one column left — so what a node contains
 * lies right of it and what reaches it (a session, the fabric it belongs
 * to) left. The walk is breadth-first in input order. A node the walk does
 * not reach takes the column of the first placed node of its kind, else
 * the column left of every other.
 */
export function columnsOf(nodes: LayoutInputNode[], edges: LayoutInputEdge[], root = "node"): Map<string, number> {
  const ids = new Set(nodes.map((n) => n.id));
  const col = new Map<string, number>();
  const start = ids.has(root) ? root : nodes[0]?.id;
  if (start === undefined) return col;
  col.set(start, 0);
  const queue = [start];
  while (queue.length > 0) {
    const u = queue.shift()!;
    const c = col.get(u)!;
    for (const e of edges) {
      if (!ids.has(e.source) || !ids.has(e.target)) continue;
      if (e.source === u && !col.has(e.target)) {
        col.set(e.target, c + 1);
        queue.push(e.target);
      } else if (e.target === u && !col.has(e.source)) {
        col.set(e.source, c - 1);
        queue.push(e.source);
      }
    }
  }
  const byKind = new Map<string, number>();
  for (const n of nodes) if (col.has(n.id) && !byKind.has(n.kind)) byKind.set(n.kind, col.get(n.id)!);
  for (const n of nodes) {
    if (col.has(n.id)) continue;
    const c = byKind.get(n.kind) ?? Math.min(...col.values()) - 1;
    col.set(n.id, c);
    if (!byKind.has(n.kind)) byKind.set(n.kind, c);
  }
  return col;
}

/** Lays the graph out; nodes keep their input order within a column. */
export function layoutGraph(nodes: LayoutInputNode[], edges: LayoutInputEdge[], root = "node"): GraphLayout {
  const { nodeHeight: h, rowGap, columnGap, margin } = LAYOUT;
  const col = columnsOf(nodes, edges, root);
  const columns = [...new Set(col.values())].sort((a, b) => a - b);
  const members = new Map<number, LayoutInputNode[]>(columns.map((c) => [c, []]));
  for (const n of nodes) members.get(col.get(n.id)!)!.push(n);

  const texts = new Map(nodes.map((n) => [n.id, fitText(n.label)]));
  const colWidth = new Map(columns.map((c) => [c, Math.max(...members.get(c)!.map((n) => nodeWidth(texts.get(n.id)!)))]));
  const colHeight = (c: number) => members.get(c)!.length * (h + rowGap) - rowGap;
  const inner = columns.length ? Math.max(...columns.map(colHeight)) : 0;

  const laid = new Map<string, LaidNode>();
  let x = margin;
  columns.forEach((c, ci) => {
    const cw = colWidth.get(c)!;
    const y0 = margin + Math.round((inner - colHeight(c)) / 2);
    members.get(c)!.forEach((n, row) => {
      const text = texts.get(n.id)!;
      const width = nodeWidth(text);
      laid.set(n.id, { id: n.id, text, column: ci, row, x: x + Math.round((cw - width) / 2), y: y0 + row * (h + rowGap), width, height: h });
    });
    x += cw + columnGap;
  });
  const width = columns.length ? x - columnGap + margin : 2 * margin;

  const out: LaidEdge[] = [];
  for (const e of edges) {
    const a = laid.get(e.source);
    const b = laid.get(e.target);
    if (a && b) out.push({ source: e.source, target: e.target, kind: e.kind, path: edgePath(a, b) });
  }
  return { width, height: inner + 2 * margin, nodes: nodes.map((n) => laid.get(n.id)!), edges: out };
}

/** A cubic curve from a's side facing b to b's side facing a; between two nodes of one column, bottom to top. */
function edgePath(a: LaidNode, b: LaidNode): string {
  if (a.column === b.column) {
    const [top, bottom] = a.y <= b.y ? [a, b] : [b, a];
    const x1 = top.x + top.width / 2;
    const y1 = top.y + top.height;
    const x2 = bottom.x + bottom.width / 2;
    const y2 = bottom.y;
    const my = (y1 + y2) / 2;
    return `M${x1} ${y1} C${x1} ${my} ${x2} ${my} ${x2} ${y2}`;
  }
  const [left, right] = a.column < b.column ? [a, b] : [b, a];
  const x1 = left.x + left.width;
  const y1 = left.y + left.height / 2;
  const x2 = right.x;
  const y2 = right.y + right.height / 2;
  const mx = (x1 + x2) / 2;
  return `M${x1} ${y1} C${mx} ${y1} ${mx} ${y2} ${x2} ${y2}`;
}
