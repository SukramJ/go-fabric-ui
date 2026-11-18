// @vitest-environment happy-dom
// The attribute reader and writer (review phase 1b).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../../test/Harness.svelte";
import { body, testUI } from "../../../../test/support";
import { createClient, type ControllerNodeDetail } from "../../../api/client";
import Attributes from "./Attributes.svelte";

afterEach(cleanup);

const L = "0000000000000011";
const SUBJECT = "18364758544493064721";

/** A client whose read answers `answer` as raw text; it records every request body as sent. */
function rawClient(answer: (path: { endpoint: number; cluster_hex: string; attribute_hex: string }) => string) {
  const sent: { method: string; body: string }[] = [];
  const client = createClient("/api/v1", async (url, init) => {
    const text = String(init?.body ?? "");
    sent.push({ method: init?.method ?? "GET", body: text });
    if (init?.method === "PUT") return new Response('{"status":0}', { status: 200 });
    const p = (JSON.parse(text) as { paths: { endpoint: number; cluster_hex: string; attribute_hex: string }[] }).paths[0];
    return new Response(answer(p), { status: 200 });
  });
  return { client, sent };
}

function renderAttributes(client: ReturnType<typeof createClient>) {
  const ui = testUI({ client });
  render(Harness, { ui, page: Attributes, pageProps: { nodeId: L, detail: body<ControllerNodeDetail>("controller-node.200") } });
  return ui;
}

async function readPath(cluster: string, attribute: string) {
  await fireEvent.change(screen.getByLabelText("controller.attributes.cluster"), { target: { value: cluster } });
  await fireEvent.input(screen.getByLabelText("controller.attributes.attribute"), { target: { value: attribute } });
  await fireEvent.click(screen.getByText("controller.attributes.read"));
}

describe("UR-P1-ctr-101: a 64-bit value read and written back", () => {
  it("shows, seeds and writes the ACL subject exactly", async () => {
    const { client, sent } = rawClient(
      () => `{"values":[{"endpoint":0,"cluster_hex":"0x001F","attribute_hex":"0x0000","value":[{"1":5,"2":2,"3":[${SUBJECT}],"4":null,"254":1}]}]}`,
    );
    const ui = renderAttributes(client);
    await readPath("0x001F", "0");
    expect(await screen.findByTestId("attribute-result")).toHaveTextContent(SUBJECT);
    expect((screen.getByLabelText("controller.attributes.value") as HTMLInputElement).value).toContain(SUBJECT);
    await fireEvent.click(screen.getByText("controller.attributes.write"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.attributes.written"));
    const put = sent.find((s) => s.method === "PUT");
    expect(put?.body).toContain(`"3":[${SUBJECT}]`);
  });

  it("sends a hand-typed id beyond 2^53 as typed", async () => {
    const { client, sent } = rawClient(() => `{"values":[{"endpoint":0,"cluster_hex":"0x0028","attribute_hex":"0x0005","value":"Kitchen"}]}`);
    const ui = renderAttributes(client);
    await readPath("0x0028", "5");
    await screen.findByTestId("attribute-result");
    await fireEvent.input(screen.getByLabelText("controller.attributes.value"), { target: { value: `[${SUBJECT}]` } });
    await fireEvent.click(screen.getByText("controller.attributes.write"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalled());
    expect(sent.find((s) => s.method === "PUT")?.body).toContain(`"value":[${SUBJECT}]`);
  });
});
