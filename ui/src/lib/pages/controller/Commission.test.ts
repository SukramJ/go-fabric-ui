// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { fixture, testUI } from "../../../test/support";
import Commission from "./Commission.svelte";

afterEach(cleanup);

describe("Controller Commission", () => {
  it("commissions by code and links the new node; the code field is cleared", async () => {
    const ui = testUI({ fixtures: ["controller-commission.200", "controller-nodes.200"] });
    render(Harness, { ui, page: Commission });
    const input = screen.getByLabelText("controller.commission.code") as HTMLInputElement;
    await fireEvent.input(input, { target: { value: "34970112332" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    expect(await screen.findByText("controller.commission.succeeded")).toBeInTheDocument();
    expect(screen.getByText("Plug").closest("a")?.getAttribute("href")).toBe("#/controller/node/0000000000000013");
    await waitFor(() => expect(input.value).toBe(""));
    expect(ui.sent.find((s) => s.method === "POST")?.body).toEqual(fixture("controller-commission.200").request.body);
  });

  it("shows the server's problem when no device answers", async () => {
    render(Harness, { ui: testUI({ fixtures: ["controller-commission.404"] }), page: Commission });
    await fireEvent.input(screen.getByLabelText("controller.commission.code"), { target: { value: "12345678901" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    expect(await screen.findByText("controller.commission.failed")).toBeInTheDocument();
    expect(screen.getByTestId("commission-progress")).toHaveTextContent("fabricui: no device answered");
  });

  it("discovers devices and pairs one with its passcode", async () => {
    const ui = testUI({
      fixtures: ["controller-commission.200-discovered", "controller-nodes.200"],
      answers: { "GET /controller/discovery?timeout=5": fixture("controller-discovery.200").response },
    });
    render(Harness, { ui, page: Commission });
    await fireEvent.click(screen.getByText("controller.commission.search"));
    expect(await screen.findByText("Plug")).toBeInTheDocument();
    await fireEvent.input(screen.getByLabelText("controller.commission.passcode"), { target: { value: "20202021" } });
    await fireEvent.click(screen.getByText("controller.commission.pair"));
    await waitFor(() => expect(screen.getByText("controller.commission.succeeded")).toBeInTheDocument());
    expect(ui.sent.find((s) => s.method === "POST")?.body).toEqual(fixture("controller-commission.200-discovered").request.body);
  });

  it("refuses a passcode out of range before any request", async () => {
    const ui = testUI({ answers: { "GET /controller/discovery?timeout=5": fixture("controller-discovery.200").response } });
    render(Harness, { ui, page: Commission });
    await fireEvent.click(screen.getByText("controller.commission.search"));
    await screen.findByText("Plug");
    await fireEvent.click(screen.getByText("controller.commission.pair"));
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.commission.passcode_invalid");
    expect(ui.sent.some((s) => s.method === "POST")).toBe(false);
  });

  it("follows the stream's progress", async () => {
    const ui = testUI();
    render(Harness, { ui, page: Commission });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "failed", detail: "timeout" } });
    expect(await screen.findByText("controller.commission.failed")).toBeInTheDocument();
  });
});
