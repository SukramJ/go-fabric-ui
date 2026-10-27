// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { fixture, testUI } from "../../../test/support";
import Settings from "./Settings.svelte";

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
