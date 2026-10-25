// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { Status as StatusT } from "../../api/client";
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

  it("shows the host's warnings in a card above the status, code as a chip", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    const card = await screen.findByTestId("host-warnings");
    expect(card).toHaveTextContent("node.warnings.title");
    expect(within(card).getByText("label_unresolved")).toBeInTheDocument();
    expect(within(card).getByText('The label "Basement" names no item.')).toBeInTheDocument();
    expect(within(card).getAllByRole("listitem")).toHaveLength(2);
    expect(within(card).getAllByText(/^node\.warnings\.since/)).toHaveLength(1);
    expect(card.compareDocumentPosition(screen.getByTestId("node-status")) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("shows no warnings card when the host reports none", async () => {
    const s = body<StatusT>("node-status.200");
    delete s.warnings;
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: s } } }), page: Status });
    await screen.findByTestId("node-status");
    expect(screen.queryByTestId("host-warnings")).toBeNull();
  });

  it("says so when the host serves no node", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200-disabled"] }), page: Status });
    expect(await screen.findByText("node.status.disabled")).toBeInTheDocument();
  });
});
