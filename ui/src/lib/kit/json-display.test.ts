// @vitest-environment happy-dom
// The kit prints generic values with the client's serialiser, so a 64-bit
// integer kept as its source text shows as written — a bare number, not a
// quoted string (UR-P1-ui-193).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import { JsonText } from "../api/json";
import AttributeTable from "./AttributeTable.svelte";
import EventList from "./EventList.svelte";

afterEach(cleanup);

const BIG = "1311768467463790321";

describe("generic values in the kit (UR-P1-ui-193)", () => {
  it("prints an attribute's 64-bit value as written", () => {
    render(AttributeTable, { attributes: [{ id: 0, id_hex: "0x0000", value: [{ "1": 5, "3": [new JsonText(BIG)] }] as never }] });
    const pre = screen.getByTestId("attribute-0x0000").querySelector("pre")!;
    expect(pre.textContent).toContain(`${BIG}`);
    expect(pre.textContent).not.toContain(`"${BIG}"`);
  });

  it("prints an event's 64-bit data as written", () => {
    render(EventList, {
      props: { empty: "none", events: [{ number: 1, priority: "info", endpoint: 0, cluster_hex: "0x001F", event_hex: "0x0000", data: { subject: new JsonText(BIG) } as never }] },
    });
    const pre = screen.getByTestId("event-1").querySelector("pre")!;
    expect(pre.textContent).toContain(BIG);
    expect(pre.textContent).not.toContain(`"${BIG}"`);
  });
});
