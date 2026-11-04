// The shell's hash routes: #/node/<page>, #/node/endpoint/<number>, the
// same under a node's key (#/node/<key>/<page>, #/node/<key>/endpoint/<n>,
// contract 0.5.0), #/controller/<page>[?tab=<tab>],
// #/controller/node/<id>[?tab=<tab>] and #/operation. A segment after
// #/node/ that is no page name is a node's key; the Go side refuses a key
// that is one (fabricui reservedNodeKeys). Anything else is the default
// node's first page, as before the controller family.
import { NODE_ROUTES } from "../lib/context";

export type Route = {
  family: "node" | "controller" | "operation";
  page: string;
  nodeId?: string;
  tab?: string;
  endpoint?: number;
  /** The node's key; absent for the default node. */
  nodeKey?: string;
};

const NODE_PAGES: readonly string[] = NODE_ROUTES;

export function routeOf(hash: string): Route {
  const [path, query = ""] = hash.replace(/^#\/?/, "").split("?");
  const parts = path.split("/").filter(Boolean);
  if (parts[0] === "operation") return { family: "operation", page: "operation" };
  if (parts[0] === "controller") {
    const tab = new URLSearchParams(query).get("tab") ?? undefined;
    if (parts[1] === "node" && parts[2] && /^[0-9a-f]{1,16}$/i.test(parts[2])) {
      return { family: "controller", page: "node", nodeId: parts[2].toUpperCase().padStart(16, "0"), tab };
    }
    return tab ? { family: "controller", page: parts[1] ?? "nodes", tab } : { family: "controller", page: parts[1] ?? "nodes" };
  }
  if (parts[0] !== "node") return { family: "node", page: "status" };
  let rest = parts.slice(1);
  let keyed: { nodeKey?: string } = {};
  if (rest[0] !== undefined && !NODE_PAGES.includes(rest[0])) {
    keyed = { nodeKey: decodeURIComponent(rest[0]) };
    rest = rest.slice(1);
  }
  if (rest[0] === "endpoint") {
    const n = Number(rest[1]);
    return /^\d+$/.test(rest[1] ?? "") && n <= 0xffff ? { family: "node", page: "endpoint", endpoint: n, ...keyed } : { family: "node", page: "endpoints", ...keyed };
  }
  return { family: "node", page: rest[0] ?? "status", ...keyed };
}
