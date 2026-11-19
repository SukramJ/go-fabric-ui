// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { fixture, testUI } from "../../../test/support";
import Settings from "./Settings.svelte";
import { createTranslator } from "../../i18n";

afterEach(cleanup);

describe("Controller Settings", () => {
  it("shows the label, the credentials by name and the controller's identity", async () => {
    render(Harness, { ui: testUI({ fixtures: ["controller-status.200", "controller-credentials.200"] }), page: Settings });
    expect(await screen.findByText("HomeNet")).toBeInTheDocument();
    expect(screen.getByTestId("credentials")).not.toHaveTextContent("secret-passphrase");
    const info = screen.getByTestId("controller-info");
    expect(info).toHaveTextContent("0x0000000000000001");
    expect(info).toHaveTextContent("controller.settings.cap.ota");
    expect((screen.getByLabelText("controller.settings.label") as HTMLInputElement).value).toBe("fabric-ui");
  });

  it("sets the fabric label", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200", "controller-fabric-label.204"] });
    render(Harness, { ui, page: Settings });
    await screen.findByText("HomeNet");
    await fireEvent.input(screen.getByLabelText("controller.settings.label"), { target: { value: "Home" } });
    await fireEvent.click(screen.getAllByText("common.save")[0]);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.settings.label_saved"));
    expect(ui.sent.find((s) => s.method === "PUT")?.body).toEqual(fixture("controller-fabric-label.204").request.body);
  });

  it("stores a Wi-Fi credential write-only and clears the secret", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200", "controller-credentials-set.204"] });
    render(Harness, { ui, page: Settings });
    await screen.findByText("HomeNet");
    await fireEvent.input(screen.getByLabelText("controller.settings.col_id"), { target: { value: "office" } });
    await fireEvent.input(screen.getByLabelText("controller.settings.ssid"), { target: { value: "Office" } });
    const pass = screen.getByLabelText("controller.settings.passphrase") as HTMLInputElement;
    expect(pass.type).toBe("password");
    await fireEvent.input(pass, { target: { value: "office-passphrase" } });
    await fireEvent.click(screen.getByText("controller.settings.credential_add"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.settings.credential_saved"));
    expect(ui.sent.find((s) => s.method === "PUT")?.body).toEqual(fixture("controller-credentials-set.204").request.body);
    await waitFor(() => expect(pass.value).toBe(""));
  });

  it("removes a credential after the confirmation", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200", "controller-credentials-remove.204"] });
    render(Harness, { ui, page: Settings });
    await screen.findAllByText("mesh");
    await fireEvent.click(screen.getAllByText("common.remove")[1]);
    await waitFor(() => expect(ui.sent.some((s) => s.method === "DELETE" && s.path.endsWith("kind=thread&id=mesh"))).toBe(true));
  });
});

describe("Controller Settings — capabilities and the credential form", () => {
  it("shows the diagnostics capability (UR-P1-ui-174)", async () => {
    render(Harness, { ui: testUI({ fixtures: ["controller-status.200", "controller-credentials.200"] }), page: Settings });
    const info = await screen.findByTestId("controller-info");
    expect(info).toHaveTextContent("controller.settings.cap.diagnostics");
  });

  it("clears the typed secret when the kind changes, and refuses a dataset that is not hex bytes (UR-P1-ui-175)", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200", "controller-credentials-set.204"] });
    render(Harness, { ui, page: Settings });
    await screen.findByText("HomeNet");
    await fireEvent.input(screen.getByLabelText("controller.settings.col_id"), { target: { value: "mesh2" } });
    await fireEvent.input(screen.getByLabelText("controller.settings.passphrase"), { target: { value: "office-passphrase" } });
    await fireEvent.change(screen.getByLabelText("controller.settings.col_kind"), { target: { value: "thread" } });
    const dataset = (await screen.findByLabelText("controller.settings.dataset")) as HTMLInputElement;
    expect(dataset.value).toBe("");
    // Not hex, and an odd number of digits.
    for (const bad of ["zz00", "0e08000"]) {
      await fireEvent.input(dataset, { target: { value: bad } });
      await fireEvent.click(screen.getByText("controller.settings.credential_add"));
    }
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.settings.dataset_invalid");
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
  });
});

describe("Controller Settings — the label's limit (UR-P4-ui-64)", () => {
  // The rule is 32 bytes of UTF-8 (decision 52): the hint names bytes and
  // the limit, as the refusal does.
  it("states the limit in bytes, in both catalogues", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200"] });
    render(Harness, { ui, page: Settings });
    expect(await screen.findByText('controller.settings.label_hint:{"max":32}')).toBeInTheDocument();
    const en = createTranslator(() => "en");
    const de = createTranslator(() => "de");
    expect(en("controller.settings.label_hint", { max: 32 })).toMatch(/32 bytes/);
    expect(de("controller.settings.label_hint", { max: 32 })).toMatch(/32 Bytes/);
  });
});
