// refHref builds links only from ids of the contract's form (UR-P1-web-115).
import { describe, expect, it } from "vitest";
import { defaultRoutes, refHref } from "./context";

describe("UR-P1-web-115: refHref", () => {
  it("links a controller node only by a hex id", () => {
    expect(refHref({ page: "controller_node", id: "0000000000000011" }, defaultRoutes)).toBe("#/controller/node/0000000000000011");
    expect(refHref({ page: "controller_node", id: "../x" }, defaultRoutes)).toBeUndefined();
    expect(refHref({ page: "controller_node", id: "" }, defaultRoutes)).toBeUndefined();
  });
});
