// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { testUI } from "../../../test/support";
import Status from "./Status.svelte";

afterEach(cleanup);

describe("Status", () => {
  it("shows the snapshot: listening, healthy, counts, advertising, the open window", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    const card = await screen.findByTestId("node-status");
    for (const text of ["node.status.listening", "node.status.healthy", 'node.status.fabrics:{"count":2}', 'node.status.exposed:{"count":1}', "node.status.advertising"]) {
      expect(card).toHaveTextContent(text);
    }
    expect(card).toHaveTextContent("(300s)");
  });

  it("says so when the host serves no node", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200-disabled"] }), page: Status });
    expect(await screen.findByText("node.status.disabled")).toBeInTheDocument();
  });
});
