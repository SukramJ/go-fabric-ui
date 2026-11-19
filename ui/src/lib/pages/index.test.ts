// The pages a host mounts: every page the shell routes is exported, from
// ./pages and from the package root.
import { describe, expect, it } from "vitest";
import * as pages from "./index";
import * as root from "../index";
import ControllerDiagnostics from "./controller/Diagnostics.svelte";

describe("page exports", () => {
  // The name ControllerDiagnostics is the API type of the same name, which
  // the package root re-exports too (UR-P1-api-171).
  it("export the controller's Diagnostics page as ControllerDiagnosticsPage", () => {
    expect(pages.ControllerDiagnosticsPage).toBe(ControllerDiagnostics);
    expect(root.ControllerDiagnosticsPage).toBe(ControllerDiagnostics);
  });
});
