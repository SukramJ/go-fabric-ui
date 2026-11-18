// @vitest-environment happy-dom
// A cluster's view against a node that repeats an id in one of its global
// lists (UR-P2-ui-90).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/svelte";
import { body, uncaught } from "../../test/support";
import type { ClusterDetail } from "../api/client";
import ClusterView from "./ClusterView.svelte";

afterEach(cleanup);

describe("ClusterView — ids a node repeats (UR-P2-ui-90)", () => {
  for (const list of ["attributes", "accepted", "generated", "events"] as const) {
    it(`shows a ${list} list that names an id twice`, async () => {
      const cluster = body<{ clusters: ClusterDetail[] }>("controller-node-endpoint.200").clusters.find((c) => c.id === 6)!;
      const el = { id: 0, id_hex: "0x0000", name: "Dup" };
      (cluster[list] as unknown[]) = [el, { ...el }];
      const u = uncaught();
      let thrown = "";
      try {
        render(ClusterView, { props: { cluster }, context: new Map() });
        await waitFor(() => expect(screen.getByTestId("cluster-view-0x0006")).toBeInTheDocument());
      } catch (err) {
        thrown = String((err as Error).message).split("\n")[0];
      }
      u.stop();
      expect({ thrown, uncaught: u.errors }).toEqual({ thrown: "", uncaught: [] });
    });
  }
});
