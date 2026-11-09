// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { fixture, fixtureFetch, testUI } from "../../../test/support";
import { createClient } from "../../api/client";
import Commission from "./Commission.svelte";

// The scan chunk stands in for the camera: it reads the fixture's QR payload at once.
vi.mock("../../scan", () => ({
  startScan: async (_video: HTMLVideoElement, onCode: (payload: string) => void) => {
    onCode("MT:Y.K9042C00KA0648G00");
    return () => {};
  },
}));

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

  it("explains that the camera needs a secure context instead of offering the scan", () => {
    vi.stubGlobal("isSecureContext", false);
    render(Harness, { ui: testUI({}), page: Commission });
    expect(screen.getByTestId("scan-insecure")).toHaveTextContent("controller.commission.scan_insecure");
    expect(screen.queryByText("controller.commission.scan")).toBeNull();
    vi.unstubAllGlobals();
  });

  it("fills the code from a scanned QR payload", async () => {
    vi.stubGlobal("isSecureContext", true);
    Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia: vi.fn() } });
    const ui = testUI({});
    render(Harness, { ui, page: Commission });
    await fireEvent.click(screen.getByText("controller.commission.scan"));
    const input = screen.getByLabelText("controller.commission.code") as HTMLInputElement;
    await waitFor(() => expect(input.value).toBe("MT:Y.K9042C00KA0648G00"));
    expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.commission.scan_read");
    // Nothing is sent before the operator starts the commissioning.
    expect(ui.sent).toEqual([]);
    vi.unstubAllGlobals();
    Reflect.deleteProperty(navigator, "mediaDevices");
  });

  it("sends the terms acknowledgement when both values are set, and refuses one out of range (0.7.0)", async () => {
    const ui = testUI({ fixtures: ["controller-commission.200", "controller-nodes.200"] });
    render(Harness, { ui, page: Commission });
    await fireEvent.input(screen.getByLabelText("controller.commission.code"), { target: { value: "34970112332" } });
    await fireEvent.input(screen.getByTestId("commission-terms-version"), { target: { value: "70000" } });
    await fireEvent.input(screen.getByTestId("commission-terms-acknowledgements"), { target: { value: "1" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.commission.terms_invalid");
    expect(ui.sent.some((s) => s.method === "POST")).toBe(false);
    await fireEvent.input(screen.getByTestId("commission-terms-version"), { target: { value: "1" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    await waitFor(() => expect(ui.sent.some((s) => s.method === "POST")).toBe(true));
    expect(ui.sent.find((s) => s.method === "POST")?.body).toEqual({ code: "34970112332", terms_and_conditions: { version: 1, acknowledgements: 1 } });
  });

  it("opens the terms block on terms_required and sends the same commissioning with the acknowledgement (0.8.0)", async () => {
    const ff = fixtureFetch(["controller-commission.409-terms-required", "controller-nodes.200"]);
    const client = createClient("/api/v1", ff.fetch);
    const ui = testUI({ client });
    render(Harness, { ui, page: Commission });
    await fireEvent.input(screen.getByLabelText("controller.commission.code"), { target: { value: "34970112332" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    const retry = await screen.findByTestId("commission-terms-retry");
    expect((screen.getByTestId("commission-terms") as HTMLDetailsElement).open).toBe(true);
    expect((screen.getByTestId("commission-terms-version") as HTMLInputElement).value).toBe("1");
    await waitFor(() => expect(document.activeElement).toBe(screen.getByTestId("commission-terms-acknowledgements")));
    expect(screen.getByTestId("commission-progress")).toHaveTextContent('controller.commission.terms_required:{"version":1}');
    const date = new Date((946684800 + 852076800) * 1000).toLocaleDateString();
    expect(screen.getByTestId("commission-terms-deadline")).toHaveTextContent(`controller.commission.terms_deadline:${JSON.stringify({ date })}`);
    // Without an acknowledgement the retry is refused before any request.
    await fireEvent.click(retry);
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.commission.terms_invalid");
    expect(ff.sent.filter((s) => s.method === "POST")).toHaveLength(1);
    ff.routes.set("POST /api/v1/controller/commission", fixture("controller-commission.200"));
    await fireEvent.input(screen.getByTestId("commission-terms-acknowledgements"), { target: { value: "1" } });
    await fireEvent.click(retry);
    expect(await screen.findByText("controller.commission.succeeded")).toBeInTheDocument();
    const posts = ff.sent.filter((s) => s.method === "POST");
    expect(posts.map((s) => s.body)).toEqual([{ code: "34970112332" }, { code: "34970112332", terms_and_conditions: { version: 1, acknowledgements: 1 } }]);
    expect(screen.queryByTestId("commission-terms-retry")).toBeNull();
  });

  it("preselects the only credential and sends it (0.7.1)", async () => {
    const ui = testUI({
      fixtures: ["controller-status.200", "controller-commission.200", "controller-nodes.200"],
      answers: { "GET /controller/credentials": { status: 200, body: { credentials: [{ kind: "wifi", id: "home", name: "HomeNet" }] } } },
    });
    await ui.controller.loadStatus();
    render(Harness, { ui, page: Commission });
    const select = screen.getByLabelText("controller.commission.network") as HTMLSelectElement;
    await waitFor(() => expect(select.value).toBe("wifi:home"));
    await fireEvent.input(screen.getByLabelText("controller.commission.code"), { target: { value: "34970112332" } });
    await fireEvent.click(screen.getByText("controller.commission.start"));
    await waitFor(() => expect(ui.sent.some((s) => s.method === "POST")).toBe(true));
    expect(ui.sent.find((s) => s.method === "POST")?.body).toEqual({ code: "34970112332", wifi_id: "home" });
  });

  it("preselects the default among several and names it (0.7.1)", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200"] });
    await ui.controller.loadStatus();
    render(Harness, { ui, page: Commission });
    const select = screen.getByLabelText("controller.commission.network") as HTMLSelectElement;
    await waitFor(() => expect(select.value).toBe("thread:mesh"));
    expect(screen.getByText("controller.credential.kind_thread · mesh (controller.credential.default)")).toBeInTheDocument();
    expect(screen.queryByTestId("commission-ble-hint")).toBeNull();
  });

  it("shows the hint while no network is picked on a host with a central (0.7.1)", async () => {
    const ui = testUI({ fixtures: ["controller-status.200-ble", "controller-credentials.200"] });
    await ui.controller.loadStatus();
    render(Harness, { ui, page: Commission });
    const select = screen.getByLabelText("controller.commission.network") as HTMLSelectElement;
    await waitFor(() => expect(select.value).toBe("thread:mesh"));
    expect(screen.queryByTestId("commission-ble-hint")).toBeNull();
    await fireEvent.change(select, { target: { value: "" } });
    expect(await screen.findByTestId("commission-ble-hint")).toHaveTextContent("controller.commission.ble_hint");
    await fireEvent.click(screen.getByText("controller.commission.network_only"));
    await waitFor(() => expect(screen.queryByTestId("commission-ble-hint")).toBeNull());
  });

  it("says why Bluetooth LE was not used, and that it was (0.7.1)", async () => {
    const ui = testUI();
    render(Harness, { ui, page: Commission });
    // The wire bytes of fabricui's TestCommission_BLEDecision.
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", ble: { used: false, reason: "no_network" } } });
    expect(await screen.findByTestId("commission-ble")).toHaveTextContent("controller.commission.ble_skipped.no_network");
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", ble: { used: true } } });
    await waitFor(() => expect(screen.getByTestId("commission-ble")).toHaveTextContent("controller.commission.ble_used"));
  });

  it("lists the steps the stream reports under the progress (0.7.0)", async () => {
    const ui = testUI();
    render(Harness, { ui, page: Commission });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", step: { number: 2, sub: 1, name: "Discovery", phase: "started" } } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", step: { number: 2, sub: 1, name: "Discovery", phase: "succeeded" } } });
    const steps = await screen.findByTestId("commission-steps");
    expect(steps).toHaveTextContent("2.1");
    expect(steps).toHaveTextContent("Discovery");
    expect(steps).toHaveTextContent("controller.commission.step.succeeded");
    expect(ui.controller.progress.phase).toBe("started");
  });
});
