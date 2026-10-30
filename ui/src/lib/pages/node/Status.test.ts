// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { Status as StatusT } from "../../api/client";
import Status from "./Status.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

describe("Status", () => {
  it("shows the snapshot: listening, healthy, counts, advertising, the open window", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    const card = await screen.findByTestId("node-status");
    for (const text of ["node.status.listening", "node.status.healthy", 'node.status.fabrics:{"count":2}', 'node.status.exposed:{"count":1}', "node.status.advertising"]) {
      expect(card).toHaveTextContent(text);
    }
  });

  it("counts the open window down from its deadline", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(new Date("2026-10-25T12:01:00Z"));
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    const left = await screen.findByTestId("window-left");
    expect(left).toHaveTextContent("(240s)");
    await vi.advanceTimersByTimeAsync(5000);
    expect(left).toHaveTextContent(/\((23[45])s\)/);
  });

  it("shows the duration when the server reports no deadline", async () => {
    const s = body<StatusT>("node-status.200");
    delete s.window_deadline;
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: s } } }), page: Status });
    expect(await screen.findByTestId("window-left")).toHaveTextContent("(300s)");
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

  it("shows the surface row: level with its origin, profiles, extensions and the non-zero counts", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    const row = await screen.findByTestId("node-surface");
    for (const text of [
      "node.status.surface",
      'node.status.surface_level:{"level":"1.5.1","origin":"node.status.level_origin_explicit"}',
      'node.status.surface_profiles:{"profiles":"apple-home, home-assistant"}',
      'node.status.surface_extensions:{"extensions":"BD-chip-Thermostat-Events, BD-chip-Thermostat-Schedules"}',
      'node.status.surface_withheld:{"count":1}',
      'node.status.surface_warnings:{"count":1}',
      'node.status.surface_changed:{"count":1}',
    ]) {
      expect(row).toHaveTextContent(text);
    }
    expect(row.compareDocumentPosition(within(screen.getByTestId("node-status")).getByText("node.status.enabled")) & Node.DOCUMENT_POSITION_PRECEDING).toBeTruthy();
  });

  it("names the pin or the profiles as the level's origin and says when nothing is selected; zero counts get no chip", async () => {
    const s = body<StatusT>("node-status.200");
    s.surface = { ...s.surface!, level_origin: "pin", profiles: [], extensions: [], withheld_count: 0, warning_count: 0, changed_count: 0 };
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: s } } }), page: Status });
    const row = await screen.findByTestId("node-surface");
    expect(row).toHaveTextContent('node.status.surface_level:{"level":"1.5.1","origin":"node.status.level_origin_pin"}');
    expect(row).toHaveTextContent("node.status.surface_no_profile");
    expect(row).toHaveTextContent("node.status.surface_no_extension");
    for (const key of ["surface_withheld", "surface_warnings", "surface_changed"]) {
      expect(row).not.toHaveTextContent(`node.status.${key}`);
    }
    cleanup();
    s.surface = { ...s.surface!, level_origin: "profile apple-home, home-assistant" };
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: s } } }), page: Status });
    expect(await screen.findByTestId("node-surface")).toHaveTextContent('node.status.surface_level:{"level":"1.5.1","origin":"apple-home, home-assistant"}');
  });

  it("shows no surface row without a surface", async () => {
    const s = body<StatusT>("node-status.200");
    delete s.surface;
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: s } } }), page: Status });
    await screen.findByTestId("node-status");
    expect(screen.queryByTestId("node-surface")).toBeNull();
  });

  it("says so when the host serves no node", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200-disabled"] }), page: Status });
    expect(await screen.findByText("node.status.disabled")).toBeInTheDocument();
    expect(screen.queryByTestId("node-surface")).toBeNull();
  });
});
