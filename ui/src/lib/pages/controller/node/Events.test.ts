// @vitest-environment happy-dom
// The controller's Events tab against a node that repeats an event number
// or reports numbers out of order (UR-P2-ui-90).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import NodeHarness from "../../../../test/NodeHarness.svelte";
import { body, fixtureFetch, testUI, uncaught } from "../../../../test/support";
import { createClient, type EventRecord } from "../../../api/client";

afterEach(cleanup);

const L = "0000000000000011";

async function shown(run: () => Promise<void>) {
  const u = uncaught();
  let thrown = "";
  try {
    await run();
  } catch (err) {
    thrown = String((err as Error).message).split("\n")[0];
  }
  u.stop();
  return { thrown, uncaught: u.errors, rows: screen.queryAllByTestId(/^event-\d+$/).length };
}

describe("Events tab — event numbers a node repeats (UR-P2-ui-90)", () => {
  it("lists two events with the same number", async () => {
    const events = body<{ events: EventRecord[] }>("controller-node-events.200");
    events.events = [events.events[0], { ...events.events[1], number: events.events[0].number }];
    const got = await shown(async () => {
      const ui = testUI({ fixtures: ["controller-node.200"], answers: { [`GET /controller/nodes/${L}/events`]: { status: 200, body: events } } });
      render(NodeHarness, { ui, nodeId: L, tab: "events" });
      await screen.findByTestId("event-list");
    });
    expect(got).toEqual({ thrown: "", uncaught: [], rows: 2 });
  });

  it("lists events delivered out of number order once across the next reload", async () => {
    // The controller's history is in delivery order and answers the events
    // numbered since or later.
    const ring = body<{ events: EventRecord[] }>("controller-node-events.200").events.slice().reverse();
    const ff = fixtureFetch(["controller-node.200"]);
    const base = createClient("/api/v1", ff.fetch);
    const client = { ...base, controllerNodeEvents: async (_id: string, q: { since?: number } = {}) => ({ events: ring.filter((e) => e.number >= (q.since ?? 0)) }) };
    const got = await shown(async () => {
      const ui = testUI({ client });
      render(NodeHarness, { ui, nodeId: L, tab: "events" });
      await screen.findByTestId("event-list");
      ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } } as never);
      await new Promise((r) => setTimeout(r, 700));
    });
    expect(got).toEqual({ thrown: "", uncaught: [], rows: 2 });
  });
});

describe("UR-P4-ui-302: an event's epoch outside the Date range", () => {
  it("does not take the events tab down, and shows the number as it came", async () => {
    const ev = body<{ events: Record<string, unknown>[] }>("controller-node-events.200");
    ev.events[0].epoch_ms = 9_000_000_000_000_000; // beyond Date's ±8.64e15 ms
    const ui = testUI({ fixtures: ["controller-node.200"], answers: { [`GET /controller/nodes/${L}/events`]: { status: 200, body: ev } } });
    const u = uncaught();
    render(NodeHarness, { ui, nodeId: L, tab: "events" });
    await screen.findByTestId("node-title");
    await new Promise((r) => setTimeout(r, 200));
    u.stop();
    expect(u.errors).toEqual([]);
    expect(screen.getAllByTestId(/^event-\d/).length).toBe(ev.events.length);
    expect(screen.getByText("9000000000000000")).toBeInTheDocument();
  });
});
