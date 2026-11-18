# Phase 2 — X4 "the browser" (Fable)

Public copy. HEAD read `1056db2`. go-fabric `v0.22.2` (`gf:`), go-fabric-server
`d61a545` (`srv:`), go-ha2fabric `ec7deea` (`ha2f:`). Roles R3 (malicious
page or origin) and R4 (hostile node and host data); threats T8–T13, T18.
M = read at the line or run; I = inferred. Parts of this unit's report that
concern held rows are not in this copy (decisions 5, 14, 18); they are
published with the pull request that fixes those rows.

Everything was run in a scratch worktree detached at `1056db2`, with
Playwright 1.65.0 (Chromium 158, Firefox 159).

| ID | sev | one line |
| --- | --- | --- |
| UR-P2-ui-90 | low | a commissioned node that repeats an id (an event number, a device type, an attribute, command or event id) makes the page's keyed list throw: the tab shows nothing |
| UR-P2-web-91 | low | the index is served for every path, with relative asset references: any path with a directory part (`/node/status`, `/a/`) is a blank page |
| UR-P2-test-95 | note | the properties that hold today (no cross-origin read, no framing by another origin, no forged frame) have no guard in the repository |

## Findings

### UR-P2-ui-90 — a node that repeats an id breaks the list the page keys by it

Severity **low**. A new path of the class of UR-P1-ui-130, -131 and -139
(fixed for the Diagnostics findings, the node's endpoint events and the
graph's edges): here the repeated key is chosen by a **commissioned node**
(R4). M: the pages by vitest and Playwright, the handler by a Go test, the
adapter and go-fabric by reading; not run against a device that sends such
a list.

| # | Page (keyed each) | What the node controls | Path to the page |
| --- | --- | --- | --- |
| a | `ui/src/lib/kit/EventList.svelte:28` (`e.number`), fed by `ui/src/lib/pages/controller/node/Events.svelte:38` | the event number of every event it reports | `gf:controller/client/network.go:621` records every delivered event; `gf:controller/client/event_history.go:52-66` appends without looking at the number, in delivery order; `:81-97` answers "numbered since or later"; `fabricui/gofabric/node.go:420-441`; `fabricui/handlers_controller.go:542-551` |
| b | `ui/src/lib/pages/controller/node/Endpoints.svelte:36` (`dt.id`) | its Descriptor `DeviceTypeList` | `fabricui/gofabric/node.go:171-183` appends the list "in its order"; `fabricui/handlers_controller.go:408-414` |
| c | `ui/src/lib/kit/AttributeTable.svelte:31` (`a.id`), `ui/src/lib/kit/ClusterView.svelte:47` (`e.id`, the accepted, generated and event lists) | `AttributeList`, `AcceptedCommandList`, `GeneratedCommandList`, `EventList` of a cluster | `gf:controller/client/cluster.go:192-204` `idList` sorts and does not remove a repeat; `:118-134`; `fabricui/gofabric/node.go:363-397`, `:400-411`; `fabricui/handlers_controller.go:473-497` |

Claim. Svelte 5 throws `each_key_duplicate` for a repeated key in
production. (a) Two events with one number: the Events tab never shows its
list. The same happens without a repeat when numbers arrive out of order —
a device whose event counter started again below the old numbers: the tab
keeps `events.at(-1).number` as its cursor (`Events.svelte:34`), asks for
`since = last + 1`, gets the higher-numbered older events again and appends
them (`:38`, no merge; the node's own endpoint page merges by number,
`ui/src/lib/pages/node/Endpoint.svelte:58-61`, the fix of UR-P1-ui-131).
(b) A `DeviceTypeList` that names a type twice: the Endpoints tab shows
nothing. (c) A repeated id in one of the four global lists: the cluster's
view in the Inventory tab throws. In the browser the rest of the shell
keeps working (other tabs and pages render; measured below); the tab of
that node stays unusable while the node reports the list.

Not traced to the end (I): `ui/src/lib/pages/controller/node/Overview.svelte:98`
(the ping's addresses keyed by `a.ip`, which is address and port,
`fabricui/gofabric/node.go:753`) and
`ui/src/lib/pages/controller/Commission.svelte:300` (discovered devices
keyed by `instance_name`) — whether go-fabric's record and discovery can
hold the same address or instance twice was not read.

Fix: key these lists by index, as UR-P1-ui-130's fix did (the Inventory tab
already does for device types, `Inventory.svelte:106`), or remove repeats
where the list is built (`deviceTypes`, `clusterDetail`); the Events tab
merges by number and takes its cursor from the highest number, as
`Endpoint.svelte:58-61`. Contract: no.

Test (vitest, `ui/src/lib/pages/controller/x4-keys.test.ts`):

```ts
// @vitest-environment happy-dom
// Review phase 2, unit X4 (R4): lists a commissioned node chooses, keyed
// by an id the node may repeat. Each test fails today (each_key_duplicate).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/svelte";
import NodeHarness from "../../../test/NodeHarness.svelte";
import { body, fixtureFetch, testUI } from "../../../test/support";
import { createClient, type ClusterDetail, type ControllerNodeDetail, type EventRecord } from "../../api/client";
import ClusterView from "../../kit/ClusterView.svelte";

afterEach(cleanup);

const L = "0000000000000011";

// uncaught collects what Svelte throws while it renders or updates.
function uncaught(): { errors: string[]; stop: () => void } {
  const errors: string[] = [];
  const onErr = (e: ErrorEvent) => {
    errors.push(String(e.error?.message ?? e.message).split("\n")[0]);
    e.preventDefault();
  };
  const onRej = (e: PromiseRejectionEvent) => {
    errors.push(String((e.reason as Error)?.message ?? e.reason).split("\n")[0]);
    e.preventDefault();
  };
  window.addEventListener("error", onErr);
  window.addEventListener("unhandledrejection", onRej);
  const onProc = (e: unknown) => errors.push(String((e as Error)?.message ?? e).split("\n")[0]);
  process.on("uncaughtException", onProc);
  process.on("unhandledRejection", onProc);
  return {
    errors,
    stop: () => {
      window.removeEventListener("error", onErr);
      window.removeEventListener("unhandledrejection", onRej);
      process.off("uncaughtException", onProc);
      process.off("unhandledRejection", onProc);
    },
  };
}

describe("X4 R4: a node repeats an id the page keys a list by", () => {
  it("Endpoints tab: a DeviceTypeList that names a device type twice", async () => {
    const detail = body<ControllerNodeDetail>("controller-node.200");
    const ep1 = detail.endpoints.find((e) => e.endpoint === 1)!;
    ep1.device_types = [ep1.device_types[0], { ...ep1.device_types[0] }];
    const u = uncaught();
    let thrown = "";
    try {
      const ui = testUI({ answers: { [`GET /controller/nodes/${L}`]: { status: 200, body: detail } } });
      render(NodeHarness, { ui, nodeId: L, tab: "endpoints" });
      await screen.findByTestId("endpoint-1");
    } catch (err) {
      thrown = String((err as Error).message).split("\n")[0];
    }
    u.stop();
    expect({ thrown, uncaught: u.errors, shown: screen.queryByTestId("endpoint-1") !== null }).toEqual({ thrown: "", uncaught: [], shown: true });
  });

  it("Events tab: two events with the same number", async () => {
    const events = body<{ events: EventRecord[] }>("controller-node-events.200");
    events.events = [events.events[0], { ...events.events[1], number: events.events[0].number }];
    const u = uncaught();
    let thrown = "";
    try {
      const ui = testUI({ fixtures: ["controller-node.200"], answers: { [`GET /controller/nodes/${L}/events`]: { status: 200, body: events } } });
      render(NodeHarness, { ui, nodeId: L, tab: "events" });
      await screen.findByTestId("event-list");
    } catch (err) {
      thrown = String((err as Error).message).split("\n")[0];
    }
    u.stop();
    expect({ thrown, uncaught: u.errors, rows: screen.queryAllByTestId(/^event-\d+$/).length }).toEqual({ thrown: "", uncaught: [], rows: 2 });
  });

  it("Events tab: numbers out of order (42 delivered before 41), then the node's next reload", async () => {
    // The controller's history is in delivery order and answers the events
    // numbered since or later (gf:controller/client/event_history.go:81-97).
    const ring = body<{ events: EventRecord[] }>("controller-node-events.200").events.slice().reverse();
    const ff = fixtureFetch(["controller-node.200"]);
    const base = createClient("/api/v1", ff.fetch);
    const client = { ...base, controllerNodeEvents: async (_id: string, q: { since?: number } = {}) => ({ events: ring.filter((e) => e.number >= (q.since ?? 0)) }) };
    const u = uncaught();
    let thrown = "";
    let ui: ReturnType<typeof testUI> | undefined;
    try {
      ui = testUI({ client });
      render(NodeHarness, { ui, nodeId: L, tab: "events" });
      await screen.findByTestId("event-list");
      ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } } as never);
      await new Promise((r) => setTimeout(r, 700));
    } catch (err) {
      thrown = String((err as Error).message).split("\n")[0];
    }
    u.stop();
    expect({ thrown, uncaught: u.errors, rows: screen.queryAllByTestId(/^event-\d+$/).length }).toEqual({ thrown: "", uncaught: [], rows: 2 });
  });

  for (const list of ["attributes", "accepted", "generated", "events"] as const) {
    it(`ClusterView: a ${list} list that names an id twice`, async () => {
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
```


Output (`npx vitest run src/lib/pages/controller/x4-keys.test.ts`):

```
     × Endpoints tab: a DeviceTypeList that names a device type twice 1175ms
     × Events tab: two events with the same number 1037ms
     × Events tab: numbers out of order (42 delivered before 41), then the node's next reload 752ms
     × ClusterView: a attributes list that names an id twice 35ms
     × ClusterView: a accepted list that names an id twice 18ms
     × ClusterView: a generated list that names an id twice 32ms
     × ClusterView: a events list that names an id twice 29ms
⎯⎯⎯⎯⎯⎯⎯ Failed Tests 7 ⎯⎯⎯⎯⎯⎯⎯
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > Endpoints tab: a DeviceTypeList that names a device type twice
AssertionError: expected { …(3) } to deeply equal { thrown: '', uncaught: [], …(1) }
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > Events tab: two events with the same number
AssertionError: expected { …(3) } to deeply equal { thrown: '', uncaught: [], rows: 2 }
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > Events tab: numbers out of order (42 delivered before 41), then the node's next reload
AssertionError: expected { thrown: '', …(2) } to deeply equal { thrown: '', uncaught: [], rows: 2 }
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > ClusterView: a attributes list that names an id twice
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > ClusterView: a accepted list that names an id twice
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > ClusterView: a generated list that names an id twice
 FAIL  src/lib/pages/controller/x4-keys.test.ts > X4 R4: a node repeats an id the page keys a list by > ClusterView: a events list that names an id twice
AssertionError: expected { thrown: 'each_key_duplicate', …(1) } to deeply equal { thrown: '', uncaught: [] }
 Test Files  1 failed (1)
      Tests  7 failed (7)
```


The handler removes nothing (Go):

```go
// Support for UR-P2-ui-90: the handler passes on what the port gives; it
// removes no repeated event number and no repeated device type.
func TestX4_Ui90_TheHandlerPassesRepeatedIdsOn(t *testing.T) {
	ctrl := NewMemoryController()
	n := ctrl.nodes[0x11]
	n.mu.Lock()
	n.spec.events = append(n.spec.events, n.spec.events[len(n.spec.events)-1])
	n.spec.endpoints[1].DeviceTypes = append(n.spec.endpoints[1].DeviceTypes, n.spec.endpoints[1].DeviceTypes[0])
	n.mu.Unlock()
	_, mux := newTestHandler(t, testOptions{controller: ctrl})
	var ev struct {
		Events []struct {
			Number uint64 `json:"number"`
		} `json:"events"`
	}
	rec := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/events", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &ev)
	seen := map[uint64]bool{}
	for _, e := range ev.Events {
		if seen[e.Number] {
			t.Errorf("GET …/events answers the event number %d twice", e.Number)
		}
		seen[e.Number] = true
	}
	var d struct {
		Endpoints []struct {
			Endpoint    int `json:"endpoint"`
			DeviceTypes []struct {
				ID uint32 `json:"id"`
			} `json:"device_types"`
		} `json:"endpoints"`
	}
	rec = do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &d)
	for _, e := range d.Endpoints {
		ids := map[uint32]bool{}
		for _, dt := range e.DeviceTypes {
			if ids[dt.ID] {
				t.Errorf("GET …/nodes/{id} answers device type %#x twice on endpoint %d", dt.ID, e.Endpoint)
			}
			ids[dt.ID] = true
		}
	}
}
```


```
--- FAIL: TestX4_Ui90_TheHandlerPassesRepeatedIdsOn (0.00s)
    x4_new_test.go:115: GET …/events answers the event number 42 twice
    x4_new_test.go:133: GET …/nodes/{id} answers device type 0x100 twice on endpoint 1
```


In the browser (the events answer rewritten on its way by `page.route`):

```ts
test("RED UR-P2-ui-90: a node that repeats an event number leaves the Events tab without its list", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message.split("\n")[0]));
  await page.route(`**/api/v1/controller/nodes/${L}/events*`, async (route) => {
    const res = await route.fetch();
    const j = await res.json();
    j.events = [j.events[0], { ...j.events[1], number: j.events[0].number }];
    await route.fulfill({ response: res, json: j });
  });
  await page.goto(`/#/controller/node/${L}?tab=overview`);
  await expect(page.getByTestId("node-title")).toBeVisible();
  await page.getByRole("tab", { name: "Events" }).or(page.getByRole("button", { name: "Events" })).first().click();
  await page.waitForTimeout(1000);
  const eventsShown = await page.getByTestId("event-list").count();
  // Leave the tab and the page: does the shell still work?
  await page.evaluate(() => (location.hash = "#/controller/settings"));
  await page.waitForTimeout(1000);
  const settingsShown = await page.getByTestId("controller-info").count();
  await page.evaluate(() => (location.hash = "#/node/status"));
  await page.waitForTimeout(1000);
  const statusShown = await page.getByTestId("node-config").count();
  const bodyText = (await page.locator("body").innerText()).replace(/\s+/g, " ").slice(0, 160);
  await page.reload();
  const afterReload = await page.getByTestId("node-config").count();
  console.log(JSON.stringify({ errors: [...new Set(errors)], eventsShown, settingsShown, statusShown, bodyText, afterReload }));
  expect({ errors: [...new Set(errors)], eventsShown, settingsShown, statusShown }).toEqual({ errors: [], eventsShown: 1, settingsShown: 1, statusShown: 1 });
});
```


```
{"errors":["https://svelte.dev/e/each_key_duplicate"],"eventsShown":0,"settingsShown":1,"statusShown":1,"bodyText":"Node Pairing, fabrics, exposed items and diagnostics of this node. default Second bridge Create node… English Deutsch Node Controller Operation Status Pair Fabr","afterReload":1}
  ✘  1 [chromium] › e2e-x4/r4.spec.ts:7:1 › RED UR-P2-ui-90: a node that repeats an event number leaves the Events tab without its list (4.4s)
```


### UR-P2-web-91 — a path with a directory part is a blank page

Severity **low** (T18: a crafted or mistyped link breaks the page; nothing
runs and nothing leaks). M.

| | |
| --- | --- |
| Where | `fabricui/spa.go:42-59` (any path that is no file → the index, 200); `fabricui/dist/index.html:105-108` and `ui/vite.config.ts:46` (`base: "./"`: the four asset references are relative); the comment `fabricui/spa.go:26-28` ("a deep path is a reload of the shell") |
| Claim | The index is the answer for `/node/status`, `/a/`, `/a/b`, `/assets/`. The browser resolves `./assets/…` against that path, asks for `/node/assets/index-….js`, and gets the index again (200, `text/html`), which `nosniff` and the module MIME check refuse: no script, no style, an empty page with no word about why. A single segment (`/x`) works. `/node/status` is what a reader of the hash routes types without the `#`. The API base `./api/v1` (`ui/src/shell/App.svelte:62`) would resolve wrongly in the same way if the script loaded |
| Fix | answer the index only for a path without a directory part (`/`, `/index.html`, `/x`); for any other path that is no file a 404, or a relative redirect to the root (`Location: ../` once per directory level — relative, so it holds under an Ingress prefix the handler does not know). A `<base>` is not an option: the policy says `base-uri 'none'` (`spa.go:139`). `docs/embedding.md:374` names "every deep path" |
| Contract | no (the bundle's paths are not in `contract/openapi.yaml`); a docs line |

```go
// UR-P2-web-91: the index is the answer for every path that is no file,
// but its asset references are relative ("./assets/…"), so under a path
// with a directory part the browser asks for them below that directory
// and gets the index again.
func TestX4_Web91_TheIndexUnderADeepPathCanLoadItsAssets(t *testing.T) {
	spa := spaHandler(nil)
	ref := regexp.MustCompile(`(?:src|href)="(\./assets/[^"]+)"`)
	for _, p := range []string{"/", "/x", "/a/", "/a/b", "/assets/", "/node/status"} {
		rec := httptest.NewRecorder()
		spa.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, p, http.NoBody))
		if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/html") {
			continue // not the index: a 404 or a redirect is a sound answer
		}
		refs := ref.FindAllStringSubmatch(rec.Body.String(), -1)
		if len(refs) == 0 {
			t.Fatalf("%s: the index names no asset", p)
		}
		base, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://h"+p, http.NoBody)
		for _, m := range refs {
			u, err := base.URL.Parse(m[1])
			if err != nil {
				t.Fatal(err)
			}
			got := httptest.NewRecorder()
			spa.ServeHTTP(got, httptest.NewRequestWithContext(context.Background(), http.MethodGet, u.Path, http.NoBody))
			if ct := got.Header().Get("Content-Type"); strings.HasPrefix(ct, "text/html") {
				t.Errorf("GET %s → the index, whose %s resolves to %s → %d %s (the index again, not the asset)", p, m[1], u.Path, got.Code, ct)
			}
		}
	}
}
```


```
--- FAIL: TestX4_Web91_TheIndexUnderADeepPathCanLoadItsAssets (0.01s)
    x4_new_test.go:47: GET /a/ → the index, whose ./assets/index-xVQtlniF.js resolves to /a/assets/index-xVQtlniF.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/ → the index, whose ./assets/rolldown-runtime-WkwZ5CU-.js resolves to /a/assets/rolldown-runtime-WkwZ5CU-.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/ → the index, whose ./assets/vendor-vMc_eSiI.js resolves to /a/assets/vendor-vMc_eSiI.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/ → the index, whose ./assets/index-B-JMWn6y.css resolves to /a/assets/index-B-JMWn6y.css → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/b → the index, whose ./assets/index-xVQtlniF.js resolves to /a/assets/index-xVQtlniF.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/b → the index, whose ./assets/rolldown-runtime-WkwZ5CU-.js resolves to /a/assets/rolldown-runtime-WkwZ5CU-.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/b → the index, whose ./assets/vendor-vMc_eSiI.js resolves to /a/assets/vendor-vMc_eSiI.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /a/b → the index, whose ./assets/index-B-JMWn6y.css resolves to /a/assets/index-B-JMWn6y.css → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /assets/ → the index, whose ./assets/index-xVQtlniF.js resolves to /assets/assets/index-xVQtlniF.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /assets/ → the index, whose ./assets/rolldown-runtime-WkwZ5CU-.js resolves to /assets/assets/rolldown-runtime-WkwZ5CU-.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /assets/ → the index, whose ./assets/vendor-vMc_eSiI.js resolves to /assets/assets/vendor-vMc_eSiI.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /assets/ → the index, whose ./assets/index-B-JMWn6y.css resolves to /assets/assets/index-B-JMWn6y.css → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /node/status → the index, whose ./assets/index-xVQtlniF.js resolves to /node/assets/index-xVQtlniF.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /node/status → the index, whose ./assets/rolldown-runtime-WkwZ5CU-.js resolves to /node/assets/rolldown-runtime-WkwZ5CU-.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /node/status → the index, whose ./assets/vendor-vMc_eSiI.js resolves to /node/assets/vendor-vMc_eSiI.js → 200 text/html; charset=utf-8 (the index again, not the asset)
    x4_new_test.go:47: GET /node/status → the index, whose ./assets/index-B-JMWn6y.css resolves to /node/assets/index-B-JMWn6y.css → 200 text/html; charset=utf-8 (the index again, not the asset)
```


In the browser (`ui/e2e-x4/r4.spec.ts`, "T18", `text` is the length of the
page's visible text):

```
/index.html#/node/status | 200 text/html; charset=utf-8 | xss=null text=1026 hash=#/node/status | []
/a/#/node/status | 200 text/html; charset=utf-8 | xss=null text=0 hash=#/node/status | ["console: Refused to apply style from 'http://127.0.0.1:18190/a/assets/index-B-JMWn6y.css' because its MIME type ('text/html') is not a supported styl","console: Failed to load module script: Expected a JavaScript-or-Wasm module script but th
/a/b/c#/node/status | 200 text/html; charset=utf-8 | xss=null text=0 hash=#/node/status | ["console: Refused to apply style from 'http://127.0.0.1:18190/a/b/assets/index-B-JMWn6y.css' because its MIME type ('text/html') is not a supported st","console: Failed to load module script: Expected a JavaScript-or-Wasm module script but
/assets/#/node/status | 200 text/html; charset=utf-8 | xss=null text=0 hash=#/node/status | ["console: Refused to apply style from 'http://127.0.0.1:18190/assets/assets/index-B-JMWn6y.css' because its MIME type ('text/html') is not a supported","console: Failed to load module script: Expected a JavaScript-or-Wasm module script b
/assets/nope.js | 200 text/html; charset=utf-8 | xss=null text=0 hash= | ["console: Refused to apply style from 'http://127.0.0.1:18190/assets/assets/index-B-JMWn6y.css' because its MIME type ('text/html') is not a supported","console: Failed to load module script: Expected a JavaScript-or-Wasm module script but the server respo
/api#/node/status | 200 text/html; charset=utf-8 | xss=null text=1026 hash=#/node/status | []
```


### UR-P2-test-95 — the browser properties that hold have no guard

Severity **note**. M (grep).

`ui/e2e/smoke.spec.ts:67` asserts the policy's text
(`toContain("frame-ancestors 'self'")`), not that a frame of another
origin is refused; nothing in `fabricui/*_test.go` or `ui/e2e` asserts
that the API sends no `Access-Control-*` header, that a preflight is not
answered, or that event data cannot leave its line. Decision 10 makes a
Playwright test "a guard once its fix lands"; decision 21's fix has landed.
The guards written here pass and can be taken as they are: "guard T10" and
"guard T8/T9" in a scratch Playwright spec (they need a second origin, the
spec starts one), and `TestX4_Guard_StreamDataStaysOnOneLine`. Contract: no.

```ts
// ---- T10: framing (UR-P0-web-23, fixed in phase 0), end to end -----------

test("guard T10: a page of another origin cannot frame the shell; a same-origin parent can", async ({ page }) => {
  const con = logConsole(page);
  pages.set("/frame.html", `<!doctype html><iframe id=f src="${TOKEN}/#/node/fabrics" width=800 height=600></iframe><iframe id=g src="${TOKEN}/deep/path" width=800 height=600></iframe>`);
  await page.goto(OTHER + "/frame.html");
  await page.waitForTimeout(1500);
  const kids = page.frames().filter((f) => f !== page.mainFrame());
  console.log("cross-origin frames:", kids.map((f) => f.url()).join(", "), "| console:", con.join(" | ").slice(0, 400));
  expect(kids.length).toBe(2);
  for (const k of kids) expect(k.url().startsWith(TOKEN)).toBe(false);
  expect(con.join("\n")).toMatch(/frame-ancestors/);

  // The API's answers and the bundle's other files carry no policy.
  const page2 = await page.context().newPage();
  pages.set("/frame2.html", `<!doctype html><iframe id=a src="${TOKEN}/api/v1/node/status"></iframe><iframe id=b src="${TOKEN}/favicon.svg"></iframe>`);
  await page2.goto(OTHER + "/frame2.html");
  await page2.waitForTimeout(1000);
  console.log("framed API answer + asset:", page2.frames().slice(1).map((f) => f.url()).join(", "));

  // Same origin: admitted.
  const page3 = await page.context().newPage();
  await page3.route(TOKEN + "/parent.html", (route) => route.fulfill({ contentType: "text/html", body: `<!doctype html><iframe src="/#/node/fabrics" width=800 height=600></iframe>` }));
  await page3.goto(TOKEN + "/parent.html");
  await expect(page3.frameLocator("iframe").getByText("Home", { exact: true }).first()).toBeVisible();
});
```


```
cross-origin frames: chrome-error://chromewebdata/, chrome-error://chromewebdata/ | console: [error] Framing 'http://127.0.0.1:18190/' violates the following Content Security Policy directive: "frame-ancestors 'self'". The request has been blocked.
framed API answer + asset: http://127.0.0.1:18190/api/v1/node/status, http://127.0.0.1:18190/favicon.svg
  ✓  7 [chromium] › e2e-x4/guards.spec.ts:185:1 › guard T10: a page of another origin cannot frame the shell; a same-origin parent can (3.6s)
cross-origin frames: ,  | console: [error] [JavaScript Error: "Content-Security-Policy: The page’s settings blocked the loading of a resource (frame-ancestors) at <unknown> because it violates the following directive: “frame-ancestors 'self'”" {file: "chrome://juggler/content/NetworkObserver.js" line: 422}] | [error] [JavaScript Error: "Content-Security-Policy: The page’s settings blocked the loading of a resource (f
framed API answer + asset: http://127.0.0.1:18190/api/v1/node/status, http://127.0.0.1:18190/favicon.svg
  ✓  7 [firefox] › e2e-x4/guards.spec.ts:185:1 › guard T10: a page of another origin cannot frame the shell; a same-origin parent can (6.7s)
```


```go
// Guard (green): a host's event data cannot break the stream's framing —
// a string with line breaks is escaped, raw JSON with line breaks is
// compacted, and invalid raw JSON drops the event.
func TestX4_Guard_StreamDataStaysOnOneLine(t *testing.T) {
	src := &recordingSource{}
	h, mux := newTestHandler(t, testOptions{events: src})
	h.heartbeat = time.Hour
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/node/stream", http.NoBody)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	waitFor(t, func() bool { return src.subscribers() == 1 })
	forged := "x\n\nevent: fabric_removed\ndata: {\"fabric_index\":1}\n\n\r\revent: y "
	src.publish(Event{Type: EventEndpointsChanged, Data: map[string]string{"k": forged}})
	src.publish(Event{Type: EventEndpointsChanged, Data: json.RawMessage("{\n\"a\":\n1\n}")})
	src.publish(Event{Type: EventEndpointsChanged, Data: json.RawMessage("{}\n\nevent: fabric_removed\ndata: {}")})
	src.publish(Event{Type: EventExposureChanged, Data: struct{}{}})
	var types []string
	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "event: ") {
			types = append(types, strings.TrimPrefix(line, "event: "))
		}
		if strings.HasPrefix(line, "data: ") && strings.ContainsAny(strings.TrimPrefix(line, "data: "), "\r") {
			t.Errorf("a data line holds a carriage return: %q", line)
		}
		if line == "event: exposure_changed" {
			break
		}
	}
	want := []string{EventEndpointsChanged, EventEndpointsChanged, EventExposureChanged}
	if strings.Join(types, ",") != strings.Join(want, ",") {
		t.Errorf("event lines %v, want %v", types, want)
	}
}
```


---

## Covered without finding

**R3, forged writes (T8).** Token mode from another origin: a request with
`X-Fabric-UI-Token` is stopped at the preflight, which the handler answers
with a problem and no `Access-Control-*` header (Go control, and guard
T8/T9 in Chromium 158 and Firefox 159); without the header an admin route
is 403.

**R3, cross-origin reads (T9).** Status, token, pairing, the node stream
(`EventSource`), a `no-cors` fetch (opaque), a `<script src>` of a JSON
answer (`nosniff`): nothing readable, both browsers. No `Access-Control-*`
header is set anywhere (`grep -n "Access-Control" fabricui/*.go`: none
outside tests).

**R3, framing (T10; UR-P0-web-23, decision 21, end to end).** A page of
another origin cannot frame `/` or a deep path (Chromium: "Framing … violates
… frame-ancestors 'self'", the frame is `chrome-error://`; Firefox: the
same refusal); a same-origin parent can, and the framed shell renders. The
API's answers and `favicon.svg` carry no policy and can be framed: a JSON
document without script that the parent cannot read — no finding. The 503
page carries the policy (`fabricui/spa.go:84-91`, `:94-101`).

**R3, a hostile parent or sibling frame under an Ingress (T13).** Read:
`ui/src/lib/theme/ha-bridge.ts:54-68`, `:133-178`,
`ui/src/shell/index.html:17-101`. Only a same-origin parent is readable
(`try`/`catch` around every access); the 31 names are constants; the
values go to custom properties through `style.setProperty`, never into
markup; a `url(...)` in one is held by `default-src`/`img-src 'self'`.
A cross-origin parent with `/api/hassio_ingress/` in the frame's own path
changes the skin only. No `postMessage`, no `message` listener, no cookie,
no `window.name` (grep over `ui/src`). `localStorage`: one use, the table
sort (`ui/src/lib/kit/DataTable.svelte:61-78`, `:85-93`), read with a shape check
(UR-P1-web-114) and never written into markup; nothing else is stored.

**R3, crafted links (T18).** 26 links in Chromium (`ui/e2e-x4/r4.spec.ts`):
malformed `%`, `__proto__` and `constructor` as page, key, id and tab, an
out-of-range endpoint, markup in the tab and the key, a 100 000-character
key, dot segments, `#javascript:`, markup in the query. No script ran, no
page error; each rendered a page, except the paths of UR-P2-web-91. No
page writes or makes node traffic on mount from the link alone, beyond the
reads its tab does anyway (`grep` over the pages' `onMount` and `$effect`).

**R4, Go's JSON.** Every answer through `encoding/json`
(`fabricui/problem.go:117-121`, `:71-76`; the generated responses). A
NaN or infinite float of a node is `null` before it reaches the adapter
(`gf:internal/valuejson/valuejson.go:186-188`). A reflected path in a 404
detail is JSON in `application/problem+json` with `nosniff`.

**R4, SSE framing (T12).** One writer (`fabricui/stream.go:114`); the type
is checked (`:105-108`, `:125-135`, UR-P0-web-33); the data is
`json.Marshal` output. `TestX4_Guard_StreamDataStaysOnOneLine` (green): a
string with `\n\nevent: …` and U+2028 stays one line, raw JSON with line
breaks is compacted, raw text that is not JSON drops the event. All three
streams go through `serveSSE` (`stream.go:31`, `operations.go:282`,
`handlers_controller.go:1680`).

**R4, the client's parse.** `ui/src/lib/api/json.ts:58-150`: `__proto__`
as an own property (`:99-100`), no `eval`; `client.ts:326-348`: an answer that
is not JSON keeps its status, a problem is shape-checked before its text
is shown; `stream.ts:186-192`: data that is not JSON is skipped. An event
whose data is `null` throws inside the listener of that event only
(`store.svelte.ts:360-441`, `controller-store.svelte.ts:167-219`); the
stream stays open. Lists fed by streams are bounded (`Endpoint.svelte:31`,
`Events.svelte:22`, `Operation.svelte:47-48`), reloads are collapsed by
timers (`controller-store.svelte.ts:138-165`).

**R4, every page against hostile strings.** `TestServeX4Hostile` rewrites
every string of every JSON answer (streams excepted; the log stream
carries markup of its own every second). 28 pages × 6 payloads in
Chromium, looking for a script that ran, an injected `img`/`svg`/`script`,
a link that is not `#/…` or `https://`, a `style` attribute with the
payload, a dialog, a policy violation, a page error, a page that does not
answer within 20 s:

```
mode | pages | script ran | injected element | link not #/ or https | hostile style | each_key_duplicate | other error or dialog | slowest page
html, free text only (ids kept) | 28 | 0 | 0 | 0 | 0 | 2 | 0 | 1306 ms
html, every string | 28 | 0 | 0 | 0 | 0 | 18 | 0 | 1210 ms
javascript: url, every string | 28 | 0 | 0 | 0 | 0 | 18 | 0 | 1150 ms
__proto__, every string | 28 | 0 | 0 | 0 | 0 | 18 | 0 | 1158 ms
constructor, every string | 28 | 0 | 0 | 0 | 0 | 18 | 0 | 1152 ms
200 kB string, free text only | 28 | 0 | 0 | 0 | 0 | 2 | 0 | 4419 ms
```


No script, no injected element, no hostile link or style, no dialog, no
policy violation, in any mode; the slowest page with 200 kB in every free
text took 4.4 s. The only error is `each_key_duplicate`, where the rewrite
made two keys equal: with every string rewritten that is every list keyed
by an id or an enum (an artefact of the rewrite); with the ids kept, the
Selection page (the host's schema fields, `Selection.svelte:329`) and the
Operation page (the level names of `Select`) — both the handler's or the
host's own enumerations, not a node's. The node-chosen keys are
UR-P2-ui-90. Sinks re-read at the lines: the two `{@html}` take the QR
matrix only (`ui/src/lib/qr.ts:46-57`); every `href` from data is a
constant `#/` prefix plus an id (`context.ts:92-97`,
`Commission.svelte:266`, `Diagnostics.svelte:133`, `:149`,
`Nodes.svelte:166`), the two external ones are constants
(`EcosystemHints.svelte:48`, `NodeManage.svelte:106`); the two `style`
values from data are numbers the page computes
(`controller/node/Endpoints.svelte:32`, `kit/Graph.svelte:175`).

**Not settled.** Safari was not run. Two keys of UR-P2-ui-90 were not
traced into go-fabric (ping addresses, discovered instances).
