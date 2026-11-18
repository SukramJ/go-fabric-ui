// @vitest-environment happy-dom
// TagChips under a page's translator: the chip's title comes from the
// catalogue, not hard-coded English (UR-P1-ui-141).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import Harness from "../../test/Harness.svelte";
import { testUI } from "../../test/support";
import TagChips from "./TagChips.svelte";

afterEach(cleanup);

describe("TagChips under the page's translator (UR-P1-ui-141)", () => {
  it("takes the chip's title from the catalogue", () => {
    const ui = testUI();
    render(Harness, { ui, page: TagChips, pageProps: { tags: [{ namespace: 7, tag: 1, label: "One" }, { namespace: 8, tag: 2, mfg_code: 0xfff1 }] } });
    expect(screen.getByText("One")).toHaveAttribute("title", 'kit.tag_title:{"namespace":7,"tag":1}');
    expect(screen.getByText("8:2")).toHaveAttribute("title", 'kit.tag_title_mfg:{"namespace":8,"tag":2,"mfg":"0xFFF1"}');
  });
});
