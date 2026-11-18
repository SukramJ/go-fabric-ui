// @vitest-environment happy-dom
// UR-P1-bound-134 with the cap lowered to 20: the page's paused backlog is
// bounded by KEEP_RECORDS. Rendering the real 2000 rows took over a minute
// under a full parallel run.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { LogRecord } from "../../api/client";
import Operation from "./Operation.svelte";

vi.mock("./log", async (orig) => ({ ...(await orig<typeof import("./log")>()), KEEP_RECORDS: 20 }));

afterEach(cleanup);

function rec(seq: number, msg: string): LogRecord {
  return { seq, level: "info", msg, attrs: {}, time: "2026-10-25T12:00:00Z" };
}

describe("Operation — paused backlog (UR-P1-bound-134)", () => {
  it("keeps at most the newest KEEP_RECORDS while paused and says how many it dropped", async () => {
    let push: (r: LogRecord) => void = () => {};
    const ui = testUI({
      fixtures: ["operation-info.200", "operation-log-level.200", "operation-log.200", "node-status.200"],
      answers: { "GET /operation/log?limit=20": { status: 200, body: body("operation-log.200") } },
      subscribeLog: (fn) => {
        push = fn;
        return () => {};
      },
    });
    render(Harness, { ui, page: Operation });
    const before = (await screen.findAllByTestId("log-record")).length;
    await fireEvent.click(screen.getByText("operation.pause"));
    for (let seq = 5; seq < 5 + 50; seq++) push(rec(seq, `r${seq}`));
    expect(await screen.findByTestId("paused")).toHaveTextContent('operation.paused_dropped:{"count":50,"kept":20}');
    await fireEvent.click(screen.getByText("operation.resume"));
    await waitFor(() => expect(screen.getAllByTestId("log-record")).toHaveLength(before + 20));
    expect(screen.getByText("r54")).toBeInTheDocument();
    expect(screen.getByText("r35")).toBeInTheDocument();
    expect(screen.queryByText("r34")).toBeNull();
  });
});
