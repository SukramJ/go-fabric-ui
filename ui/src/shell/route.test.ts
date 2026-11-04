import { describe, expect, it } from "vitest";
import { routeOf } from "./route";

describe("routeOf", () => {
  it("reads the node family's pages, defaulting to status", () => {
    expect(routeOf("#/node/pair")).toEqual({ family: "node", page: "pair" });
    expect(routeOf("#/node/topology")).toEqual({ family: "node", page: "topology" });
    expect(routeOf("")).toEqual({ family: "node", page: "status" });
    expect(routeOf("#/elsewhere")).toEqual({ family: "node", page: "status" });
  });

  it("reads an endpoint's page with its number, else the endpoint list", () => {
    expect(routeOf("#/node/endpoint/0")).toEqual({ family: "node", page: "endpoint", endpoint: 0 });
    expect(routeOf("#/node/endpoint/12")).toEqual({ family: "node", page: "endpoint", endpoint: 12 });
    expect(routeOf("#/node/endpoint/x")).toEqual({ family: "node", page: "endpoints" });
    expect(routeOf("#/node/endpoint/70000")).toEqual({ family: "node", page: "endpoints" });
    expect(routeOf("#/node/endpoint")).toEqual({ family: "node", page: "endpoints" });
  });

  it("reads the controller family's pages and a node's page with its tab", () => {
    expect(routeOf("#/controller/commission")).toEqual({ family: "controller", page: "commission" });
    expect(routeOf("#/controller")).toEqual({ family: "controller", page: "nodes" });
    expect(routeOf("#/controller/nodes?tab=network")).toEqual({ family: "controller", page: "nodes", tab: "network" });
    expect(routeOf("#/controller/node/11?tab=acl")).toEqual({ family: "controller", page: "node", nodeId: "0000000000000011", tab: "acl" });
    expect(routeOf("#/controller/node/not-hex")).toEqual({ family: "controller", page: "node" });
  });
});
