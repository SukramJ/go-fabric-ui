// The shell's hash routes: #/node/<page>, #/node/endpoint/<number>,
// #/controller/<page>[?tab=<tab>] and #/controller/node/<id>[?tab=<tab>].
// Anything else is the node family's first page, as before the controller
// family.
export type Route = { family: "node" | "controller"; page: string; nodeId?: string; tab?: string; endpoint?: number };

export function routeOf(hash: string): Route {
  const [path, query = ""] = hash.replace(/^#\/?/, "").split("?");
  const parts = path.split("/").filter(Boolean);
  if (parts[0] === "controller") {
    const tab = new URLSearchParams(query).get("tab") ?? undefined;
    if (parts[1] === "node" && parts[2] && /^[0-9a-f]{1,16}$/i.test(parts[2])) {
      return { family: "controller", page: "node", nodeId: parts[2].toUpperCase().padStart(16, "0"), tab };
    }
    return tab ? { family: "controller", page: parts[1] ?? "nodes", tab } : { family: "controller", page: parts[1] ?? "nodes" };
  }
  if (parts[0] === "node" && parts[1] === "endpoint") {
    const n = Number(parts[2]);
    return /^\d+$/.test(parts[2] ?? "") && n <= 0xffff ? { family: "node", page: "endpoint", endpoint: n } : { family: "node", page: "endpoints" };
  }
  return { family: "node", page: parts[0] === "node" ? (parts[1] ?? "status") : "status" };
}
