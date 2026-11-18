// @vitest-environment happy-dom
// The controller's Endpoints tab against a node that repeats an id in its
// Descriptor lists (UR-P2-ui-90).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import NodeHarness from "../../../../test/NodeHarness.svelte";
import { body, testUI, uncaught } from "../../../../test/support";
import type { ControllerNodeDetail } from "../../../api/client";

afterEach(cleanup);

const L = "0000000000000011";

async function renderDetail(detail: ControllerNodeDetail) {
  const u = uncaught();
  let thrown = "";
  try {
    const ui = testUI({ answers: { [`GET /controller/nodes/${L}`]: { status: 200, body: detail } } });
    render(NodeHarness, { ui, nodeId: L, tab: "endpoints" });
    await screen.findByTestId("endpoint-1");
  } catch (err) {
    thrown = String((err as Error).message).split("\n")[0];
  }
  u.stop();
  return { thrown, uncaught: u.errors, shown: screen.queryByTestId("endpoint-1") !== null };
}

describe("Endpoints tab — ids a node repeats (UR-P2-ui-90)", () => {
  it("shows an endpoint whose DeviceTypeList names a device type twice", async () => {
    const detail = body<ControllerNodeDetail>("controller-node.200");
    const ep1 = detail.endpoints.find((e) => e.endpoint === 1)!;
    ep1.device_types = [ep1.device_types[0], { ...ep1.device_types[0] }];
    expect(await renderDetail(detail)).toEqual({ thrown: "", uncaught: [], shown: true });
  });

  it("shows an endpoint whose server list names a cluster twice", async () => {
    const detail = body<ControllerNodeDetail>("controller-node.200");
    const ep1 = detail.endpoints.find((e) => e.endpoint === 1)!;
    ep1.clusters = [ep1.clusters[0], { ...ep1.clusters[0] }];
    expect(await renderDetail(detail)).toEqual({ thrown: "", uncaught: [], shown: true });
  });
});
