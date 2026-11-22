// @vitest-environment happy-dom
// The controller's onboarding inventory (contract 0.13.0) on the
// controller-inventory and -node-onboarding-code fixtures: the rows and
// their sources, the code shown on request (no QR image for a manual
// code), the attach dialog's refusals in the page's language, the print
// labels, the reload on node_added, the states of a host that does not
// serve it or refuses it — and no code in storage, a toast or the console.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import { createTranslator } from "../../i18n";
import OnboardingInventory from "./OnboardingInventory.svelte";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  localStorage.clear();
  sessionStorage.clear();
});

const QR = "MT:Y.K9042C00KA0648G00";
const MANUAL = "34970112332";
const invCalls = (ui: ReturnType<typeof testUI>) => ui.sent.filter((x) => x.path === "/api/v1/controller/inventory");

describe("Controller onboarding inventory", () => {
  it("lists every node with its source and shows a code only on request", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200"] });
    render(Harness, { ui, page: OnboardingInventory });
    expect(await screen.findByText("Kitchen")).toBeInTheDocument();
    expect(within(screen.getByTestId("code-0000000000000011")).getByText("controller.inventory.no_code")).toBeInTheDocument();
    const attached = screen.getByTestId("code-0000000000000012");
    expect(within(attached).getByText("controller.inventory.source_attached")).toBeInTheDocument();
    expect(within(attached).getByText("controller.inventory.check_discriminator")).toBeInTheDocument();
    expect(within(attached).queryByText("controller.inventory.check_vendor")).toBeNull();
    expect(within(screen.getByTestId("code-0000000000000013")).getByText("controller.inventory.source_paired")).toBeInTheDocument();
    // Nothing is drawn for a closed row.
    expect(document.body.innerHTML).not.toContain(MANUAL);
    expect(document.body.innerHTML).not.toContain(QR);
    expect(screen.queryByTestId("qr-image")).toBeNull();

    await fireEvent.click(screen.getByTestId("show-0000000000000013"));
    const paired = screen.getByTestId("shown-0000000000000013");
    expect(within(paired).getByTestId("qr-image").querySelector("svg")).not.toBeNull();
    expect(within(paired).getByTestId("manual-code")).toHaveTextContent(MANUAL);

    // An entry without qr_code: the manual code and the sentence, no image.
    await fireEvent.click(screen.getByTestId("show-0000000000000012"));
    const manual = screen.getByTestId("shown-0000000000000012");
    expect(within(manual).queryByTestId("qr-image")).toBeNull();
    expect(within(manual).getByTestId("no-qr")).toHaveTextContent("controller.inventory.no_qr");
    expect(within(manual).getByTestId("manual-code")).toHaveTextContent(MANUAL);

    await fireEvent.click(screen.getByTestId("show-0000000000000013"));
    expect(screen.queryByTestId("shown-0000000000000013")).toBeNull();
  });

  it("fetches the inventory with cache no-store and keeps no code in storage, a toast or the console", async () => {
    const log = vi.spyOn(console, "log");
    const warn = vi.spyOn(console, "warn");
    const err = vi.spyOn(console, "error");
    const setItem = vi.spyOn(Storage.prototype, "setItem");
    const ui = testUI({ fixtures: ["controller-inventory.200", "controller-node-onboarding-code-set.200"] });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    await fireEvent.click(screen.getByTestId("show-0000000000000013"));
    // The sensor (0x12) has an attached code: its row offers Replace.
    await fireEvent.click(screen.getAllByText("controller.inventory.replace")[0]);
    await fireEvent.input(screen.getByTestId("attach-code"), { target: { value: "3497-011-2332" } });
    await fireEvent.submit(screen.getByTestId("attach-code").closest("form")!);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.inventory.attached"));
    expect(setItem).not.toHaveBeenCalled();
    expect(localStorage.length + sessionStorage.length).toBe(0);
    expect(location.hash).not.toContain(MANUAL);
    const said = JSON.stringify([log.mock.calls, warn.mock.calls, err.mock.calls, ui.toastSpy.success.mock.calls, ui.toastSpy.error.mock.calls]);
    for (const secret of [QR, MANUAL, "3497-011-2332"]) expect(said).not.toContain(secret);
  });

  it("asks the client's three routes with cache no-store", async () => {
    const inits: RequestInit[] = [];
    const { createClient } = await import("../../api/client");
    const c = createClient("/api/v1", async (_i, init) => {
      inits.push(init ?? {});
      return new Response(init?.method === "DELETE" ? null : JSON.stringify(init?.method === "PUT" ? body("controller-node-onboarding-code-set.200") : body("controller-inventory.200")), {
        status: init?.method === "DELETE" ? 204 : 200,
      });
    });
    await c.onboardingInventory();
    await c.setOnboardingCode("0000000000000012", MANUAL);
    await c.clearOnboardingCode("0000000000000012");
    expect(inits.map((i) => i.cache)).toEqual(["no-store", "no-store", "no-store"]);
  });

  it("attaches a code: the row is replaced, the dialog closes, the toast carries no code", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200", "controller-node-onboarding-code-set.200"] });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    // The sensor (0x12) has a code: its row offers Replace.
    await fireEvent.click(screen.getAllByText("controller.inventory.replace")[0]);
    const dialog = screen.getByTestId("attach-dialog");
    expect(within(dialog).getByText(/controller.inventory.attach_title/)).toBeInTheDocument();
    await fireEvent.input(screen.getByTestId("attach-code"), { target: { value: "3497-011-2332" } });
    await fireEvent.submit(screen.getByTestId("attach-code").closest("form")!);
    await waitFor(() => expect(screen.queryByTestId("attach-dialog")).toBeNull());
    // Nothing is sent before Attach; then one PUT with the typed text.
    const put = ui.sent.filter((s) => s.method === "PUT");
    expect(put).toHaveLength(1);
    expect(put[0].body).toEqual({ code: "3497-011-2332" });
  });

  it("words a mismatch from field, code_value and node_value in the page's language", async () => {
    const t = createTranslator(() => "de");
    const ui = testUI({ fixtures: ["controller-inventory.200", "controller-node-onboarding-code-set.409-mismatch"], t });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    await fireEvent.click(screen.getAllByText("Code hinterlegen")[0]);
    await fireEvent.input(screen.getByTestId("attach-code"), { target: { value: QR } });
    await fireEvent.submit(screen.getByTestId("attach-code").closest("form")!);
    const refusal = await screen.findByTestId("attach-refusal");
    expect(refusal).toHaveTextContent("Der Code gehört zu einem anderen Gerät; es wurde nichts gespeichert.");
    expect(refusal).toHaveTextContent("Die Produkt-ID des Codes ist 0x8000, die des Geräts ist 0x8001.");
    expect(ui.toastSpy.error).not.toHaveBeenCalled();
    // The dialog stays with the text.
    expect(screen.getByTestId("attach-code")).toHaveValue(QR);
  });

  it("states a discriminator mismatch in decimal", async () => {
    const ui = testUI({
      fixtures: ["controller-inventory.200"],
      answers: {
        "PUT /controller/nodes/0000000000000011/onboarding-code": {
          status: 409,
          body: { type: "x", title: "The code names another device", status: 409, code: "onboarding_code_mismatch", field: "discriminator", code_value: 3, node_value: 15 },
        },
      },
    });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    await fireEvent.click(screen.getAllByText("controller.inventory.attach")[0]);
    await fireEvent.input(screen.getByTestId("attach-code"), { target: { value: MANUAL } });
    await fireEvent.submit(screen.getByTestId("attach-code").closest("form")!);
    expect(await screen.findByTestId("attach-refusal")).toHaveTextContent('controller.inventory.mismatch_discriminator:{"code":"3","node":"15"}');
  });

  it("shows an invalid code's reason as the server words it", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200"] });
    ui.client.setOnboardingCode = async () => {
      const { ApiError } = await import("../../api/client");
      const p = body<{ detail: string }>("controller-node-onboarding-code-set.400-invalid");
      throw new ApiError(400, p.detail, p as never);
    };
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    await fireEvent.click(screen.getAllByText("controller.inventory.attach")[0]);
    await fireEvent.input(screen.getByTestId("attach-code"), { target: { value: "not a code" } });
    await fireEvent.submit(screen.getByTestId("attach-code").closest("form")!);
    const refusal = await screen.findByTestId("attach-refusal");
    expect(refusal).toHaveTextContent("controller.inventory.refused_invalid");
    expect(refusal).toHaveTextContent("Invalid pairing code");
  });

  it("says the live camera needs a secure page outside one", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200"] });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    await fireEvent.click(screen.getAllByText("controller.inventory.attach")[0]);
    // happy-dom is no secure context.
    expect(screen.getByTestId("scan-insecure")).toHaveTextContent("controller.inventory.camera_note");
    expect(screen.queryByText("controller.commission.scan")).toBeNull();
    expect(screen.getByTestId("attach-photo")).toHaveAttribute("accept", "image/*");
    expect(screen.getByTestId("attach-photo")).not.toHaveAttribute("capture");
  });

  it("removes a code after the confirmation and reloads", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200", "controller-node-onboarding-code-clear.204"] });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    // The rows in the server's order: 0x12 is the first with a code.
    await fireEvent.click(screen.getAllByText("controller.inventory.remove")[0]);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.inventory.removed"));
    expect(ui.confirmSpy).toHaveBeenCalled();
    expect(ui.sent.some((s) => s.method === "DELETE" && s.path.endsWith("/controller/nodes/0000000000000012/onboarding-code"))).toBe(true);
    expect(invCalls(ui)).toHaveLength(2);
  });

  it("prints one label per node with a code, drawn only for the print", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200"] });
    let labels: string[] = [];
    let images = 0;
    vi.stubGlobal("print", () => {
      const block = screen.getByTestId("print-labels");
      labels = [...block.querySelectorAll(".print-label")].map((l) => l.textContent ?? "");
      images = block.querySelectorAll("svg").length;
    });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    expect(screen.queryByTestId("print-labels")).toBeNull();
    await fireEvent.click(screen.getByText("controller.inventory.print"));
    await waitFor(() => expect(labels).toHaveLength(2));
    expect(labels[0]).toContain(MANUAL);
    expect(labels[1]).toContain("Plug");
    expect(images).toBe(1);
    await waitFor(() => expect(screen.queryByTestId("print-labels")).toBeNull());
  });

  it("reloads on node_added", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.200"] });
    render(Harness, { ui, page: OnboardingInventory });
    await screen.findByText("Kitchen");
    ui.controller.applyEvent({ type: "node_added", data: { node_id_hex: "0000000000000014" } } as never);
    await waitFor(() => expect(invCalls(ui)).toHaveLength(2));
  });

  it("shows the not-served state on a 501 and an admin state on a 403, without a toast", async () => {
    const ui = testUI({ fixtures: ["controller-inventory.501"] });
    render(Harness, { ui, page: OnboardingInventory });
    expect(await screen.findByText("controller.inventory.not_served")).toBeInTheDocument();
    cleanup();
    const denied = testUI({ fixtures: ["controller-inventory.403"] });
    render(Harness, { ui: denied, page: OnboardingInventory });
    expect(await screen.findByText("controller.inventory.admin_required")).toBeInTheDocument();
    expect(ui.toastSpy.error).not.toHaveBeenCalled();
    expect(denied.toastSpy.error).not.toHaveBeenCalled();
  });

  it("asks nothing when the status says the host does not serve the inventory", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-inventory.200"] });
    await ui.controller.loadStatus();
    render(Harness, { ui, page: OnboardingInventory });
    expect(await screen.findByText("controller.inventory.not_served")).toBeInTheDocument();
    expect(invCalls(ui)).toHaveLength(0);
  });
});
