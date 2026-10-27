import { describe, expect, it } from "vitest";
import { routeOf } from "./route";

describe("routeOf", () => {
  it("reads the node family's pages, defaulting to status", () => {
    expect(routeOf("#/node/pair")).toEqual({ family: "node", page: "pair" });
    expect(routeOf("")).toEqual({ family: "node", page: "status" });
    expect(routeOf("#/elsewhere")).toEqual({ family: "node", page: "status" });
  });

  it("reads the controller family's pages and a node's page with its tab", () => {
    expect(routeOf("#/controller/commission")).toEqual({ family: "controller", page: "commission" });
    expect(routeOf("#/controller")).toEqual({ family: "controller", page: "nodes" });
    expect(routeOf("#/controller/node/11?tab=acl")).toEqual({ family: "controller", page: "node", nodeId: "0000000000000011", tab: "acl" });
    expect(routeOf("#/controller/node/not-hex")).toEqual({ family: "controller", page: "node" });
  });
});
