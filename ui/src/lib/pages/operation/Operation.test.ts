// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, echoT, testUI } from "../../../test/support";
import type { LogRecord } from "../../api/client";
import Operation from "./Operation.svelte";
import { formatUptime, mergeRecords, recordLine, visibleRecords } from "./log";

afterEach(cleanup);

const ALL = ["operation-info.200", "operation-log-level.200", "operation-log.200", "operation-log-level-set.200"];

function rec(seq: number, level: LogRecord["level"], msg: string, attrs: Record<string, string> = {}): LogRecord {
  return { seq, level, msg, attrs, time: "2026-10-25T12:00:00Z" };
}

function renderOperation(fixtures = ALL) {
  let push: (r: LogRecord) => void = () => {};
  const ui = testUI({
    fixtures,
    // The page reads the ring with the route's maximum; the fixture's request reads since=2.
    answers: fixtures.includes("operation-log.200") ? { "GET /operation/log?limit=2000": { status: 200, body: body("operation-log.200") } } : {},
    subscribeLog: (fn) => {
      push = fn;
      return () => {};
    },
  });
  render(Harness, { ui, page: Operation });
  return { ui, push: (r: LogRecord) => push(r) };
}

describe("Operation", () => {
  it("shows the version, the go-fabric version, the uptime and the ring", async () => {
    const { ui } = renderOperation();
    const info = await screen.findByTestId("operation-info");
    expect(info).toHaveTextContent('operation.version:{"version":"0.5.0"}');
    expect(info).toHaveTextContent("v0.16.1-0.20261104133901-894adbcd4cd3");
    expect(info.textContent).toContain(`operation.uptime_hours:{\\"h\\":1,\\"m\\":30}`);
    expect(screen.getAllByTestId("log-record")).toHaveLength(3);
    expect(screen.getByText("subscription failed")).toBeInTheDocument();
    expect(screen.getByText("peer.node=000000000001B669")).toBeInTheDocument();
    expect(ui.sent.map((s) => s.path)).toContain("/api/v1/operation/log?limit=2000");
  });

  it("appends what the stream delivers, once each, and holds it while paused", async () => {
    const { push } = renderOperation();
    await screen.findAllByTestId("log-record");
    push(rec(4, "error", "subscription failed", { error: "timeout" }));
    push(rec(5, "info", "window opened"));
    await waitFor(() => expect(screen.getAllByTestId("log-record")).toHaveLength(4));
    await fireEvent.click(screen.getByText("operation.pause"));
    push(rec(6, "info", "held back"));
    expect(await screen.findByTestId("paused")).toHaveTextContent('operation.paused:{"count":1}');
    expect(screen.queryByText("held back")).toBeNull();
    await fireEvent.click(screen.getByText("operation.resume"));
    expect(await screen.findByText("held back")).toBeInTheDocument();
  });

  it("filters by level and text", async () => {
    renderOperation();
    await screen.findAllByTestId("log-record");
    await fireEvent.change(screen.getByLabelText("operation.filter_level"), { target: { value: "warn" } });
    await waitFor(() => expect(screen.getAllByTestId("log-record")).toHaveLength(2));
    await fireEvent.input(screen.getByLabelText("operation.filter_text"), { target: { value: "TIMEOUT" } });
    await waitFor(() => expect(screen.getAllByTestId("log-record")).toHaveLength(1));
  });

  it("sets the level through the admin route", async () => {
    const { ui } = renderOperation();
    await screen.findByTestId("operation-info");
    await fireEvent.change(screen.getByLabelText("operation.level_label"), { target: { value: "debug" } });
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith('operation.level_set:{"level":"debug"}'));
    const put = ui.sent.find((s) => s.method === "PUT");
    expect(put?.body).toEqual({ level: "debug" });
    expect(put?.headers["X-Fabric-UI-Token"]).toBe("test-token");
  });

  it("keeps the level and says why when the server refuses it", async () => {
    const { ui } = renderOperation(["operation-info.200", "operation-log-level.200", "operation-log.200", "operation-log-level-set.403"]);
    await screen.findByTestId("operation-info");
    await fireEvent.change(screen.getByLabelText("operation.level_label"), { target: { value: "debug" } });
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalled());
    expect((screen.getByLabelText("operation.level_label") as HTMLSelectElement).value).toBe("info");
  });

  it("says the host serves no operation on 503", async () => {
    renderOperation(["operation-info.503"]);
    expect(await screen.findByText("operation.unavailable")).toBeInTheDocument();
  });
});

describe("operation log helpers", () => {
  it("merges by sequence number and keeps the newest", () => {
    expect(mergeRecords([rec(2, "info", "b"), rec(1, "info", "a")], [rec(2, "info", "b"), rec(3, "info", "c")]).map((r) => r.seq)).toEqual([1, 2, 3]);
    const many = Array.from({ length: 2005 }, (_, i) => rec(i + 1, "info", "x"));
    const kept = mergeRecords(many, []);
    expect(kept).toHaveLength(2000);
    expect(kept[0].seq).toBe(6);
  });

  it("filters, renders a line and formats the uptime", () => {
    const rs = [rec(1, "debug", "a"), rec(2, "warn", "b", { k: "V" })];
    expect(visibleRecords(rs, "info", "").map((r) => r.seq)).toEqual([2]);
    expect(visibleRecords(rs, "debug", "k=v").map((r) => r.seq)).toEqual([2]);
    expect(recordLine(rs[1])).toBe("2026-10-25T12:00:00Z WARN b k=V");
    expect(formatUptime(echoT, 90061)).toBe('operation.uptime_days:{"d":1,"h":1}');
    expect(formatUptime(echoT, 59)).toBe('operation.uptime_minutes:{"m":0}');
  });
});
