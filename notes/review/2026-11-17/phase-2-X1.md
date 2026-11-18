# Phase 2 — X1, secret flows

Reviewer X1 (Fable), HEAD `1056db2`,
go-fabric `v0.22.2` (`gf:`), hosts `srv:` `d61a545`, `ha2f:` `ec7deea`.
Scratch worktree `scratchpad/p2wt/x1` (detached at `1056db2`); every test
below lives there only, as an untracked file, and fails today. **M** = run
or read at the line, **I** = inferred. Threats T5, T14.

**No held candidate.** No finding is high: no secret reaches a non-admin, a log, a
cache, an error text or another node by a path that the registers do not
already hold.

Counts: 0 high, 1 medium, 4 low, 4 note.

| ID | sev | one line |
| --- | --- | --- |
| UR-P2-conc-04 | medium | A stream opened before an erase stays on the erased node; nothing on the page reconnects, so the Pair flow keeps the erased node's codes and never hears of the new node |
| UR-P2-conc-01 | low | A window replaced by another (open → open) raises no stream event; the page keeps the QR code and countdown of a window that is gone |
| UR-P2-err-02 | low | `GET /node/pairing` forgets the handler's own fresh window when the node could not publish its record, and answers the configured codes for it |
| UR-P2-adpt-03 | low | A commissioning that happened is reported as failed when the context ends, or the node's start fails, after the commissioner is done |
| UR-P2-ui-05 | low | The page records its own fresh window as kind `own`; its own `window_opened` or any resync drops the codes and re-reads them, and loses them if that read fails |
| UR-P2-ui-06 | note | The countdown subtracts the browser's clock from a deadline on the node's |
| UR-P2-conc-07 | note | Two concurrent opens both pass the 409 check; the node replaces the first window, both callers get 200 |
| UR-P2-leak-08 | note | Secrets held in memory past their use: the ended window's credentials, the code kept for a terms retry, the window in phase `success` |
| UR-P2-err-09 | note | Upstream: a commissioning cut at CommissioningComplete leaves the device's fail-safe armed; the retry answers 502 until it expires |

## UR-P2-conc-04 — the stream stays on the erased node, and the page does not reconnect

- **Severity:** medium. Deterministic after every erase on both pinned
  hosts (`srv:internal/daemon/nodes.go:296`, `ha2f:cmd/ha2fabric/nodes.go:136`
  call `NodeSwitch.Set`); nothing is lost, a reload or hiding and showing
  the tab heals it. Lower to low if "documented in a code comment" is taken
  as accepted.
- **Where:** `fabricui/handler.go:646-647` (`observeNode(h.node, fn)` once
  per stream), `fabricui/nodeswitch.go:127-128` (`Observe` on the target of
  that moment), `:25-29` (the comment that says so and leaves it to "when
  it reconnects"); `ui/src/lib/pages/node/Fabrics.svelte:104-115`,
  `:138-141` (the erase flow: poll, `loadStatus`, `loadFabrics`, no
  reconnect, no `resync`); `ui/src/lib/store.svelte.ts:133-148`
  (`loadStatus` does not judge the pairing flow; only `resync` and
  `window_opened` do, `:401`, `:446`).
- **Claim (M, run):** after `POST …/maintenance/erase` the open stream
  carries neither a `window_closed` for the erased node's window nor
  anything of the new node: not its window, not `fabric_added` when a
  controller commissions it. The routes read the new node at once
  (`GET /node/pairing` answers its configured codes). On the page: a Pair
  flow that was open on a fresh window keeps that window's QR code, manual
  code and countdown — codes of a node that no longer exists, while the new
  node's own window is open on the configured credentials — until the
  countdown ends (up to 15 min); and whatever the flow showed, the pairing
  of the new node, which is what follows an erase, never reaches
  `success`, the fabric list and the status do not move.
- **Not known:** the registers hold nothing on it; `nodeswitch.go:25-29`
  describes the server half as intended.
- **Failing test 1 (Go, `fabricui/x1_erase_stream_test.go`):**

```go
// X1: a page that holds its stream across an erase (NodeSwitch.Set) — the
// bundled page does, nothing in its erase flow reconnects — is told
// nothing about the new node: not the window the handler opens on it, not
// the controller that commissions it. GET /node/pairing meanwhile answers
// for the new node.
func TestX1_StreamFollowsTheNodeAcrossAnErase(t *testing.T) {
	s := newSiteWith(t, false, true)
	// A fresh window on the first node, as a Pair page left open shows it.
	var w1 struct {
		Passcode uint32 `json:"passcode"`
	}
	if err := json.Unmarshal(s.reopen("basic"), &w1); err != nil {
		t.Fatal(err)
	}
	s.expect(http.StatusNoContent, http.MethodPost, "/node/maintenance/erase", `{"confirm":"erase-node"}`)
	for len(s.events) > 0 {
		t.Logf("event around the erase: %s", <-s.events)
	}
	st := s.status()
	var p struct {
		Passcode uint32 `json:"passcode"`
		Source   string `json:"source"`
	}
	s.getJSON("/node/pairing", &p)
	t.Logf("after the erase: status %+v; pairing source %q passcode %d (the page still shows %d)", st, p.Source, p.Passcode, w1.Passcode)

	c := s.site.AddController(fabrictest.ControllerOptions{Label: "Controller"})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if _, err := c.Commission(ctx, s.device, fabrictest.CommissionOptions{}); err != nil {
		t.Fatal(err)
	}
	var fl fabricList
	s.getJSON("/node/fabrics", &fl)
	if len(fl.Fabrics) != 1 {
		t.Fatalf("fabrics after commissioning the new node %+v", fl)
	}
	deadline := time.After(3 * time.Second)
	var got []string
	for {
		select {
		case ev := <-s.events:
			got = append(got, ev)
			if ev == "fabric_added" {
				return
			}
		case <-deadline:
			t.Fatalf("the new node was commissioned (GET /node/fabrics lists the fabric) and the stream opened before the erase carried %v, no fabric_added", got)
		}
	}
}
```

```
--- FAIL: TestX1_StreamFollowsTheNodeAcrossAnErase (3.15s)
    x1_erase_stream_test.go:29: event around the erase: cluster_event
    x1_erase_stream_test.go:37: after the erase: status {WindowOpen:true WindowKind:own WindowDeadline:2026-01-03T00:00:00Z WindowDuration:65535}; pairing source "configured" passcode 20202021 (the page still shows 75500059)
    x1_erase_stream_test.go:60: the new node was commissioned (GET /node/fabrics lists the fabric) and the stream opened before the erase carried [], no fabric_added
```

- **Failing test 2 (vitest, `ui/src/lib/x1-pairing.test.ts`; the harness
  is shared by -05 and -06):**

```ts
function harness(opts: { pairing: () => Response }) {
  let st: Status = { ...body<Status>("node-status.200"), commissioning_window_open: false, window_kind: "none" };
  const ff = fixtureFetch(["node-pairing-window.200", "node-fabrics.200"]);
  const sent: string[] = [];
  const client = createClient("/api/v1", async (url, init) => {
    sent.push(`${init?.method ?? "GET"} ${url.replace("/api/v1", "")}`);
    if (url.endsWith("/node/status")) return new Response(JSON.stringify(st), { status: 200 });
    if (url.endsWith("/node/pairing")) return opts.pairing();
    return ff.fetch(url, init);
  });
  const store = createNodeStore(client, { t, toast, subscribe: () => () => {} });
  return { store, sent, set: (patch: Partial<Status>) => (st = { ...st, ...patch }) };
}

const freshCodes = body<{ qr_code: string; manual_code: string }>("node-pairing-window.200");
```

```ts
describe("X1: an erase under an open Pair flow", () => {
  // Fabrics.svelte:104-115: after the erase's 204 the page polls the
  // status and then calls node.loadStatus() and node.loadFabrics(). The
  // stream stays on the erased node (nodeswitch.go:25-29), so no window
  // event arrives.
  it("the erased node's window codes leave the page once the status names the new node's window", async () => {
    const h = harness({
      pairing: () => new Response(JSON.stringify({ ...body<Record<string, unknown>>("node-pairing.200"), source: "configured", window_open: true }), { status: 200 }),
    });
    await h.store.loadStatus();
    await h.store.openWindow(300);
    const shown = h.store.commissioning.window?.qr_code;
    // The new node: no fabric, its own window on the configured credentials for 48 h.
    h.set({ commissioning_window_open: true, window_kind: "own", fabric_count: 0, window_deadline: new Date(Date.now() + 48 * 3600_000).toISOString() });
    await h.store.loadStatus();
    await h.store.loadFabrics();
    await vi.advanceTimersByTimeAsync(50);
    expect(shown).toBe(freshCodes.qr_code);
    expect(h.store.commissioning.window?.qr_code ?? null).not.toBe(shown);
  });
});
```

```
 FAIL  src/lib/x1-pairing.test.ts > X1: an erase under an open Pair flow > the erased node's window codes leave the page once the status names the new node's window
AssertionError: expected 'MT:FRESH1234' not to be 'MT:FRESH1234' // Object.is equality
```

- **Fix:** server: a stream over a `NodeSwitch` follows it — `Set` ends
  the open streams of its handlers (as `stop` does for a deleted registry
  node, `handler.go:87-91`) so `EventSource` reconnects and the client's
  resync runs, or `observeNode` re-registers on the new target and pushes a
  `window_closed`/`window_opened` pair. Page: the erase flow calls
  `node.resync()` after the node answers (`Fabrics.svelte:113-115`) and
  resets an open pairing flow at the erase's 204.
- **Contract:** no (a stream that ends and is reopened is within
  `docs/contract.md`; a new event would be yes).

## UR-P2-conc-01 — a replaced window raises no stream event

- **Severity:** low. Adds a path to UR-P1-err-198 (fixed for a resync and
  for a `window_opened`): the event that fix waits for never comes.
- **Where:** `fabricui/events.go:281-293` (`changed := open != w.Open`: a
  snapshot that is open after an open one is dropped); `gf:bridge/commissioning_window.go:619-628`,
  `:664-666` (an administrator's window replaces the node's own: the old
  window's close hook runs after the new state is set, so both hooks
  snapshot `Open: true`), `:456-460` (`OpenOwnWindow` on an open own or
  basic window: "replaced without closing", one hook); `gf:node/commissioning.go:109-121`,
  `:356-359`.
- **Claim (M, run):** a fresh window opened through the handler, then an
  administrator's `OpenCommissioningWindow` sent without a
  `RevokeCommissioning` before it: the status says `window_kind: enhanced`
  with a new deadline, `GET /node/pairing` forgets the fresh credentials
  (right) and answers `configured` (UR-P1-err-04), and the stream carries
  **nothing**. A Pair page keeps the fresh window's QR code and its
  countdown for a window those codes no longer open; decision 28's notice
  never shows, because nothing makes the page read again. The same holds
  for a host's `OpenOwnWindow` or `OpenBasicWindow` under an open window
  (I, read at `:456-460`).
- **Precondition:** an administrator that opens without revoking. go-fabric's
  own client and matter.js revoke first (`gf:controller/client/window.go:114`),
  which closes the fresh window and does raise `window_closed`; that
  connectedhomeip-based administrators do not is **I** (not in a pin).
- **Failing test (`fabricui/x1_window_replaced_test.go`, package
  `fabricui_test`, imports `admincomm "…/cluster/spec/administratorcommissioning"`,
  `…/controller/client`, `…/fabrictest`, `…/secure/spake2`):**

```go
// X1: the handler's fresh window is replaced by an administrator's
// enhanced window (a controller on the node's fabric sends
// OpenCommissioningWindow). The node fires its transition hooks, the
// stream must tell the page: the codes it shows no longer open the window.
func TestX1_ReplacedWindowReachesTheStream(t *testing.T) {
	s := newSite(t, false)
	c := s.site.AddController(fabrictest.ControllerOptions{Label: "Controller"})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pair, err := c.Commission(ctx, s.device, fabrictest.CommissionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s.waitEvent("commissioning_complete")
	if st := s.status(); st.WindowOpen {
		// the own window closed with the commissioning
		s.waitEvent("window_closed")
	}
	var w struct {
		Passcode uint32 `json:"passcode"`
	}
	if err := json.Unmarshal(s.expect(http.StatusOK, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`), &w); err != nil {
		t.Fatal(err)
	}
	s.waitEvent("window_opened")
	if st := s.status(); st.WindowKind != "basic" {
		t.Fatalf("status after the fresh open %+v", st)
	}
	// drain what is queued
	for len(s.events) > 0 {
		<-s.events
	}
	n, err := pair.Node()
	if err != nil {
		t.Fatal(err)
	}
	// As a chip-based administrator sends it: OpenCommissioningWindow
	// without a RevokeCommissioning before (go-fabric's own client revokes
	// first, client/window.go:114, which closes the fresh window).
	const adminPasscode, iterations = 34567890, 1000
	salt := make([]byte, 32)
	vc, err := spake2.NewVerifierContext(adminPasscode, salt, iterations)
	if err != nil {
		t.Fatal(err)
	}
	l, err := vc.L.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	verifier := make([]byte, spake2.VerifierW0Size, spake2.VerifierW0Size+spake2.VerifierLSize)
	vc.W0.FillBytes(verifier)
	verifier = append(verifier, l...)
	if err := s.site.Run(ctx, func(ctx context.Context) error {
		_, err := n.Invoke(ctx, client.CommandAt(0, admincomm.ClusterID, admincomm.CmdOpenCommissioningWindow),
			admincomm.OpenCommissioningWindowRequest{
				CommissioningTimeout: 300, PakePasscodeVerifier: verifier,
				Discriminator: 3000, Iterations: iterations, Salt: salt,
			}, client.InvokeOptions{Timed: true})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	st := s.status()
	var p struct {
		Source   string `json:"source"`
		Passcode uint32 `json:"passcode"`
	}
	s.getJSON("/node/pairing", &p)
	t.Logf("after the administrator's window: status %+v; pairing source %q; fresh passcode %d, served %d", st, p.Source, w.Passcode, p.Passcode)
	if st.WindowKind != "enhanced" {
		t.Fatalf("the administrator's window did not replace the fresh one: %+v", st)
	}
	deadline := time.After(3 * time.Second)
	var got []string
	for {
		select {
		case ev := <-s.events:
			got = append(got, ev)
			if ev == "window_opened" || ev == "window_closed" {
				return
			}
		case <-deadline:
			t.Fatalf("no window event on the stream after an administrator's window replaced the fresh one (events: %v): a page keeps the fresh window's QR code and countdown", got)
		}
	}
}
```

```
--- FAIL: TestX1_ReplacedWindowReachesTheStream (3.11s)
    x1_window_replaced_test.go:84: after the administrator's window: status {WindowOpen:true WindowKind:enhanced WindowDeadline:2026-01-01T00:05:00Z WindowDuration:300}; pairing source "configured"; fresh passcode 32215086, served 20202021
    x1_window_replaced_test.go:98: no window event on the stream after an administrator's window replaced the fresh one (events: []): a page keeps the fresh window's QR code and countdown
```

- **Fix:** `NodeObservers` remembers the window it last reported (kind and
  `OpenedAt`, or kind and discriminator) and pushes `window_opened` when an
  open snapshot names another one; a snapshot of the same window (the
  AdminFabricIndex cleared, `events.go:242-244`) stays silent. The page
  already re-reads on `window_opened` under an open flow
  (`store.svelte.ts:400-401`).
- **Contract:** no new event or field; `docs/contract.md` gets the sentence
  that `window_opened` may follow `window_opened`.

## UR-P2-err-02 — the handler forgets its own fresh window when the announce failed

- **Severity:** low. A new path to UR-P1-err-04 (decision 28): there the
  window is someone else's, here it is the handler's own.
- **Where:** `fabricui/handlers_pairing.go:104` (`win.Discriminator !=
  eph.creds.Discriminator` forgets); `gf:node/commissioning.go:278-284`
  (`announce` records the discriminator only "once published": on a
  publish error it returns before `setAnnounced`), `:373-375` (the
  snapshot's discriminator is the announced one), `gf:bridge/bridge.go:2352-2357`
  (`Publish` failing, or running into `AdvertiseTimeout`, `:2333`).
- **Claim (M, run):** with an advertiser that refuses the publish,
  `POST /node/pairing/window` answers 200 with the fresh codes, the window
  is open and accepts only the fresh passcode (`gf:node/commissioning.go:250-272`),
  `node.Window()` says `open, basic, discriminator 0` — and the next
  `GET /node/pairing` drops the credentials and answers `source:
  configured, window_open: true` with the configured passcode. The fresh
  passcode cannot be read again; the page (which re-reads, UR-P2-ui-05)
  replaces the right codes by the configured ones under decision 28's
  "opened elsewhere" notice. No secret leaves the admin route. A node
  without an advertiser is not affected (`gf:bridge/bridge.go:2330-2332` returns nil,
  the discriminator is recorded).
- **Failing test (`fabricui/x1_announce_test.go`):**

```go
// flakyAdvertiser refuses to publish while fail is set: a responder whose
// socket is gone, a publish that ran into Bridge's AdvertiseTimeout.
type flakyAdvertiser struct {
	mdns.Advertiser
	fail atomic.Bool
}

func (a *flakyAdvertiser) Publish(ctx context.Context, svc mdns.Service) error {
	if a.fail.Load() {
		return errors.New("mdns: publish failed")
	}
	return a.Advertiser.Publish(ctx, svc)
}

// X1: the handler opens a window on fresh credentials while the node
// cannot publish its commissionable record. The window is open and PASE
// runs against the fresh passcode (node/commissioning.go:250-272), but
// node.Window still reports the discriminator of the last record that was
// published (:278-284), so GET /node/pairing forgets the credentials it
// just drew and answers the configured codes for the open window.
func TestX1_PairingKeepsTheFreshWindowWhenTheAnnounceFails(t *testing.T) {
	adv := &flakyAdvertiser{Advertiser: mdns.NewNoop()}
	fs := fabrictest.NewSite(t, fabrictest.Options{})
	plug, err := device.NewOnOffPlugInUnit(device.Options{StableKey: device.StringKey("plug:1"), Name: "Plug"})
	if err != nil {
		t.Fatal(err)
	}
	dev := fs.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{plug}, Configure: func(c *node.Config) { c.Advertiser = adv }})
	s := &site{t: t, site: fs, device: dev}
	s.srv = httptest.NewServer(fabricui.Handler(fabricui.Options{Node: dev.Node()}))
	t.Cleanup(s.srv.Close)
	var st struct {
		CSRFToken string `json:"csrf_token"`
	}
	s.getJSON("/node/status", &st)
	s.token = st.CSRFToken

	s.expect(http.StatusNoContent, http.MethodPost, "/node/pairing/window/close", "")
	adv.fail.Store(true)
	var w struct {
		Passcode      uint32 `json:"passcode"`
		Discriminator uint16 `json:"discriminator"`
		QRCode        string `json:"qr_code"`
	}
	if err := json.Unmarshal(s.expect(http.StatusOK, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`), &w); err != nil {
		t.Fatal(err)
	}
	win := dev.Node().Window()
	t.Logf("node.Window: open %v kind %v discriminator %d; the fresh window's discriminator %d", win.Open, win.Kind, win.Discriminator, w.Discriminator)
	var p struct {
		Passcode   uint32 `json:"passcode"`
		Source     string `json:"source"`
		WindowOpen bool   `json:"window_open"`
		QRCode     string `json:"qr_code"`
	}
	s.getJSON("/node/pairing", &p)
	if p.Source != "window" || p.Passcode != w.Passcode || p.QRCode != w.QRCode {
		t.Fatalf("GET /node/pairing right after the open: source %q window_open %v passcode %d (configured %d), want the fresh window's %d",
			p.Source, p.WindowOpen, p.Passcode, dev.Passcode(), w.Passcode)
	}
}
```

```
--- FAIL: TestX1_PairingKeepsTheFreshWindowWhenTheAnnounceFails (0.04s)
    x1_announce_test.go:68: node.Window: open true kind basic discriminator 0; the fresh window's discriminator 108
    x1_announce_test.go:77: GET /node/pairing right after the open: source "configured" window_open true passcode 20202021 (configured 20202021), want the fresh window's 3939410
```

- **Fix:** identify the handler's window by what the node says of the
  window itself, not of its record: keep the snapshot's `OpenedAt` (read
  at `handlers_pairing.go:190`) in `ephemeralWindow` and compare `Kind`
  and `OpenedAt`; the discriminator only when the snapshot has one.
- **Contract:** no.

## UR-P2-adpt-03 — a commissioning that happened is answered as failed

- **Severity:** low.
- **Where:** `fabricui/gofabric/gofabric.go:351-354` (`Pair`: the node
  go-fabric returns beside the error is dropped), `:591-593`
  (`CommissionDiscovered`: likewise, and `:594` `forget` is skipped);
  `gf:controller/pair_race.go:71-73` and `gf:controller/controller.go:829-831`
  (the node is returned with the error), `gf:controller/client/fabricops.go:156-157`,
  `:229-235` ("an error of its initial read is returned and the node stays
  commissioned"); `fabricui/handlers_controller.go:1358-1361` (phase
  `failed` on the ungated stream, then the error), `:191-193` (502) or
  `problem.go:104-110` (500).
- **Claim (M, run for the adapter; read for the handler):** when the
  request's context ends — the page or a proxy left, `timeout_seconds` ran
  out (`handlers_controller.go:1324-1331`) — or the node's start fails
  after CommissioningComplete (step Subscribe 21.1), the device is on our
  fabric and the source lists the node, but `Pair` answers an error: the
  stream says `commissioning_progress failed`, the route 500 or 502 "The
  node failed", the Commission page "failed" (`Commission.svelte:127`)
  while the node list gains a node. A second attempt with the same code
  then fails for another reason ("node is already commissioned", probe
  below). So the abort is not clean in this one window; before it, it is
  (see "Covered").
- **Failing test (`fabricui/gofabric/x1_commission_test.go`):**

```go
// X1: the request's context ends (the page left, timeout_seconds ran out)
// after the commissioner finished and while the registry starts the node
// (step Subscribe 21.1, client/fabricops.go:240-246: "an error of its
// initial read is returned and the node stays commissioned"). go-fabric's
// Pair returns the node and the error; the adapter must not report a
// commissioning that happened as one that failed.
func TestX1_PairThatCommissionedIsNotReportedFailed(t *testing.T) {
	s := newCtrlSite(t)
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var steps []string
	var n fabricui.ControllerNode
	var perr error
	_ = s.site.Run(ctx, func(ctx context.Context) error {
		ctx, abort := context.WithCancel(ctx)
		defer abort()
		n, perr = s.src.Pair(ctx, sp.Manual, fabricui.PairOptions{Progress: func(st fabricui.PairingStep) {
			steps = append(steps, st.Name+" "+st.Phase)
			if st.Number == 21 && st.Sub == 1 && st.Phase == "started" {
				abort() // the request ends here
			}
		}})
		return nil
	})
	nodes := s.src.Nodes()
	var fl []fabricInfo
	if infos, err := s.dev.Node().FabricInfos(context.Background()); err == nil {
		for _, f := range infos {
			fl = append(fl, fabricInfo{f.FabricIndex, f.Label})
		}
	}
	t.Logf("last steps %v", steps[max(0, len(steps)-3):])
	t.Logf("Pair: node %v, err %v; the source lists %d node(s); the device holds fabrics %+v", n != nil, perr, len(nodes), fl)
	if perr != nil && len(nodes) != 0 {
		t.Fatalf("Pair answered an error (%v) for a commissioning that happened: the device is on our fabric and the source lists it", perr)
	}
}

type fabricInfo struct {
	Index uint8
	Label string
}
```

```
--- FAIL: TestX1_PairThatCommissionedIsNotReportedFailed (0.11s)
    x1_commission_test.go:46: last steps [OperationalCredentials.UpdateFabricLabel succeeded Subscribe started Subscribe failed]
    x1_commission_test.go:47: Pair: node false, err context canceled; the source lists 1 node(s); the device holds fabrics [{Index:1 Label:e2e}]
    x1_commission_test.go:49: Pair answered an error (context canceled) for a commissioning that happened: the device is on our fabric and the source lists it
```

- **Fix:** in both adapter paths, an error with a commissioned node
  (`n != nil && n.IsCommissioned()`) returns the node: the commissioning
  succeeded, the failed start is the step's (already reported as `Subscribe
  failed`) and at most a log line; `CommissionDiscovered` also forgets the
  cached device. The handler then answers 200 and `succeeded`.
- **Contract:** no (a 200 where a 500/502 was wrong).

## UR-P2-ui-05 — the page takes its own fresh window for another one

- **Severity:** low. Introduced with UR-P1-err-198's fix (`24e3626`).
- **Where:** `ui/src/lib/store.svelte.ts:281` (`shown = { kind: "own", … }`
  after every `POST …/pairing/window`), `:293-294`, `:400-401`, `:446`,
  `:462-468`; `fabricui/handlers_pairing.go:169-173` (fresh credentials are
  the default; both pinned hosts leave `ConfiguredWindow` unset, grep over
  `srv:` and `ha2f:`), `fabricui/handlers_node.go:192` (`window_kind:
  basic` for it).
- **Claim (M, run):** in the default mode the status names the page's own
  window `basic`, the store remembers `own`. The window's own
  `window_opened`, when it is handled after the POST's answer, and every
  later resync (each time the tab is shown again, decision 29) make
  `otherWindow` true: the flow is set idle, the codes are dropped and
  `GET /node/pairing` is read again. If that read fails, the page shows
  the empty form over an open window whose codes it just had; with
  UR-P2-err-02 it comes back with the configured codes and the "opened
  elsewhere" notice. In the good case it costs one more read of the secret
  per resync.
- **Failing test (`ui/src/lib/x1-pairing.test.ts`, harness above):**

```ts
const freshCodes = body<{ qr_code: string; manual_code: string }>("node-pairing-window.200");

describe("X1: the page's own window on fresh credentials", () => {
  // The handler's default: POST /node/pairing/window opens a basic window
  // (handlers_pairing.go:221-238), the status names it window_kind "basic"
  // (handlers_node.go:191), and its window_opened follows on the stream.
  it("its own window_opened does not drop the codes the POST answered", async () => {
    const h = harness({ pairing: () => new Response(JSON.stringify({ type: "about:blank", title: "x", status: 500, code: "internal" }), { status: 500 }) });
    await h.store.loadStatus();
    await h.store.openWindow(300);
    h.set({ commissioning_window_open: true, window_kind: "basic", window_deadline: new Date(Date.now() + 300_000).toISOString(), commissioning_window_duration_seconds: 300 });
    h.store.applyEvent({ type: "window_opened", data: { duration_seconds: 300 } });
    await vi.advanceTimersByTimeAsync(50);
    expect({ phase: h.store.commissioning.phase, reads: h.sent.filter((s) => s === "GET /node/pairing").length }).toEqual({ phase: "open", reads: 0 });
    expect(h.store.commissioning.window?.qr_code).toBe(freshCodes.qr_code);
    expect(h.sent.filter((s) => s === "GET /node/pairing")).toEqual([]);
  });
});
```

```
 FAIL  src/lib/x1-pairing.test.ts > X1: the page's own window on fresh credentials > its own window_opened does not drop the codes the POST answered
AssertionError: expected { phase: 'idle', reads: 1 } to deeply equal { phase: 'open', reads: +0 }
```

- **Fix:** remember the window by what identifies it in both modes — the
  deadline alone after a POST (`shown.kind` undefined until a status named
  it), or take the kind from the status read that follows; keep the shown
  codes when a re-read fails and the status still names the same window.
- **Contract:** no. A `window_kind` in the window answer would be yes
  (with decision 28's field, phase 3).

## UR-P2-ui-06 — the countdown mixes two clocks (note)

- **Where:** `ui/src/lib/store.svelte.ts:327`, `:343`; `ui/src/lib/countdown.ts:9-13`
  (`Date.now()`); `fabricui/handlers_node.go:194-197` (`window_deadline` on
  the node's clock, no "now" beside it).
- **Claim (M, run):** a browser ahead of the host by S seconds counts S
  seconds too few and ends the flow early; one behind shows the codes S
  seconds past the window unless `window_closed` arrives; one ahead by more
  than what is left gets `left = 0`, no countdown at all, and the codes
  stay until an event. With the stream up the effect is cosmetic.
- **Test (same file):**

```ts
describe("X1: the countdown against the node's clock", () => {
  // window_deadline is on the node's clock (handlers_node.go:194,
  // countdown.ts:1-3); the page subtracts the browser's.
  it("a browser ten minutes ahead of the node still counts the open window down", async () => {
    const h = harness({
      pairing: () => new Response(JSON.stringify({ ...body<Record<string, unknown>>("node-pairing.200"), source: "configured", window_open: true, duration_seconds: 900 }), { status: 200 }),
    });
    const nodeNow = Date.now() - 600_000;
    h.set({ commissioning_window_open: true, window_kind: "own", commissioning_window_duration_seconds: 900, window_deadline: new Date(nodeNow + 900_000).toISOString() });
    await h.store.hydrateCommissioning();
    expect(h.store.commissioning.phase).toBe("open");
    // 300 s of the window are left on the browser's clock, 900 on the node's.
    expect(h.store.commissioning.remaining).toBe(900);
  });
  it("a browser twenty minutes ahead shows the codes with no countdown at all", async () => {
    const h = harness({
      pairing: () => new Response(JSON.stringify({ ...body<Record<string, unknown>>("node-pairing.200"), source: "configured", window_open: true, duration_seconds: 900 }), { status: 200 }),
    });
    const nodeNow = Date.now() - 1_200_000;
    h.set({ commissioning_window_open: true, window_kind: "own", commissioning_window_duration_seconds: 900, window_deadline: new Date(nodeNow + 900_000).toISOString() });
    await h.store.hydrateCommissioning();
    await vi.advanceTimersByTimeAsync(3_600_000);
    // An hour later, with no event: the codes are still on the page.
    expect(h.store.commissioning).toMatchObject({ phase: "idle" });
  });
});
```

```
AssertionError: expected 300 to be 900 // Object.is equality
AssertionError: expected { phase: 'open', …(4) } to match object { phase: 'idle' }
```

- **Fix:** the seconds left as a field of the status and the pairing
  answer (`window_remaining_seconds`), or the answer's `Date` header as the
  node's now. **Contract:** yes for the field.

## UR-P2-conc-07 — two concurrent opens both succeed (note)

- **Where:** `fabricui/handlers_pairing.go:162-165` (the 409 check) against
  `:230-235` (open, then store), no lock across them;
  `gf:node/commissioning.go:539-540`, `:258-267` (a basic window opened
  under a basic one replaces it).
- **Claim (I, read):** two admin requests inside the gap both answer 200
  with different codes; the node serves the second window only; the first
  caller's codes are dead, and its page is not told (UR-P2-conc-01: open →
  open). `h.ephemeral` is whichever stored last; if it is the loser's,
  `GET …/pairing` answers `configured` (UR-P1-err-04). UR-P1-conc-05 fixed
  the close/open pair of this family, not the open/open one.
- **Fix:** hold `h.mu` (or an open-in-progress flag) from the check to the
  store. **Contract:** no.

## UR-P2-leak-08 — secrets held past their use, in memory only (note)

- `fabricui/handlers_pairing.go:97-110`, `:233-235`: the fresh window's
  passcode, salt and codes stay in `h.ephemeral` after the window ended
  until the next `GET …/pairing` or close through this handler — for the
  life of the process if no one asks. Never served: `activeEphemeral`
  checks the node's window and deadline first (M, read). Fix: forget on the
  node's close transition (an `Observe` of the handler's own).
- `ui/src/lib/store.svelte.ts:370`: `fabric_added` keeps
  `commissioning.window` (passcode, codes) in the store through phase
  `success` until the operator clicks Close; not rendered (`Pair.svelte:202-210`).
- `ui/src/lib/pages/controller/Commission.svelte:49`, `:139`: the code
  or passcode refused with `terms_required` stays in `retry` until the next
  commissioning or the page is left; documented at `:18-19`.
- **Contract:** no.

## UR-P2-err-09 — upstream: a cut at CommissioningComplete (note)

- **Claim (M, probe `fabricui/gofabric/x1_abort_probe_test.go`, run):** the
  context ended at step 20.1 `GeneralCommissioning.Complete started`: `Pair`
  answers `context canceled`, the source lists nothing, the device holds
  our fabric under its fail-safe; an immediate second attempt answers 502
  with `commissioner: Commission error for "armFailSafe": BusyWithOtherAdmin
  (4), fail-safe owned by fabric 1; requesting fabric 0 rejected` until the
  fail-safe expires. go-fabric's commissioner does not disarm on a
  cancelled context; that a cut after the device processed the command and
  before its answer leaves a fabric the controller does not know is **I**.
  Not this module's: for go-fabric's review (a disarm on a detached
  context). The text carries no secret.

## Covered without finding

**Pairing**

- Gate: `GET …/pairing`, the window open and close are admin routes for
  `/node` and `/nodes/{node}` (`handler.go:289-291`, loop `:331-345`); the
  gate runs before the body. Answers are `no-store` (`handler.go:373-379`).
- Fresh credentials: drawn per open (`handler.go:266`, `gf:node/credentials.go:63-69`),
  in the POST's answer and `h.ephemeral` only; the log line has the
  discriminator, duration and mode (`handlers_pairing.go:197-199`), the
  stream's `window_opened` the discriminator and duration (`events.go:334-336`)
  — both known (T5, threat model 3.2).
- Served only for the window it belongs to: closed, timed out, commissioned,
  replaced by the own or an administrator's window, deadline passed —
  forgotten on the next read (`handlers_pairing.go:104-108`; run:
  `TestX1_ReplacedWindowReachesTheStream` logs `source "configured"` after
  the replacement, `TestX1_StreamFollowsTheNodeAcrossAnErase` after the
  erase). A stale ephemeral is served again only if a later basic window
  has the same 12-bit discriminator before the old deadline: not reachable
  by a role.
- A second node: each node has its own handler and `ephemeral`
  (`handler.go:501-520`; `sibling` copies none); the page builds a store per
  node key (`shell/App.svelte:307-308`, `NodeScope.svelte:21-25`), so one
  node's codes are not shown under another. A registry handler rebuilt
  under an open window loses its ephemeral: UR-P1-err-04's known path.
- `NodeSwitch` swap and erase, server side: every pairing route reads the
  current target; after an erase the old node's ephemeral is not served
  (the new node's window is `own`). The stream is UR-P2-conc-04.
- Factory reset under an open fresh window: the window stays, its codes
  still apply (`gf:node/commissioning.go:397-401`: nothing reopens while a
  window is open); the page's flow is not touched (`Fabrics.svelte:72-90`).
- Expiry on a node without a fabric: `window_closed`, then the own window's
  `window_opened` (`gf:node/commissioning.go:112-120`); the page re-reads
  and shows the configured codes, which apply (`store.svelte.ts:392-401`).
- Configured mode: `openConfigured` answers the setup payload and closes
  the window again if it cannot (`handlers_pairing.go:205-216`).
- Error texts: `setup: invalid passcode %d` (`gf:secure/setup/setup.go:129`,
  `:250`) would reach the log at `handlers_pairing.go:79` and `:186`, but a
  `*node.Node` cannot hold or draw such a passcode
  (`gf:node/config.go:654-656`, same rule `:612`); the 503 and 409 details
  are fixed texts (`:80`, `:163`, `:181-184`).
- Page: codes are rendered as text and as a locally built SVG
  (`Pair.svelte:174`, `:181`, `:191`); a viewer without admin rights reads
  no codes (`:109-117`, `store.svelte.ts:317`); nothing of the flow goes to
  `localStorage` or the console (grep).

**Commissioning**

- Gate: `POST /controller/commission`, `GET /controller/discovery`, the
  credential writes (`handlers_controller.go:70-73`), before decoding
  (`:125-129`).
- The code and the passcode: in no log line, event or error text —
  validation answers fixed texts (`:1308-1320`), the adapter replaces the
  parser's error (`gofabric.go:329-332`; the second parse inside
  `ctrl.Pair` cannot fail after it), the stream carries the kind only
  (`:1366-1384`). A body the decoder refuses is answered to its sender, not
  logged (`:105`, `problem.go:143-149`); Go's type error names a number
  literal, never a string.
- Credential store: the list answers kind, id, name and default
  (`:1438-1456`); passphrase and dataset have no response field and reach
  `commissioner.Options` only (`gofabric.go:277-288`); an unknown id is
  answered by its id (`:279`, `:286`). go-fabric sends the Wi-Fi credentials
  only to a device with a Wi-Fi interface that is not connected
  (`gf:controller/commissioner/network.go:221-234`), after attestation. The
  remaining commissioner errors (T14, finding 31's "X1 reads the rest"):
  grep over `Errorf`, `newError`, `errors.New`, `slog.` in
  `gf:controller/commissioner`, `gf:controller`, `gf:controller/client/fabricops.go`
  for passcode, credential, dataset, passphrase, verifier, salt — only the
  SSID sites already in UR-P0-leak-31 (`network.go:271`, `:296`) and the
  scan warning (`:254`).
- BLE decision: four host-side inputs (`gofabric.go:305-318`), reported as
  a reason word (`handlers_controller.go:1341-1343`), logged as the reason
  (`gofabric.go:343-345`); the template's central is not used when the
  decision says no (`:339-343`).
- Terms retry: 409 with version and deadline (`problem.go:91-98`,
  `gofabric.go:399-402`); the page repeats the same request with the
  acknowledgement once (`Commission.svelte:117-156`); the request's
  acknowledgement wins over the template's, the host's `Edit` runs last
  (`gofabric.go:387-392`).
- Progress events: step number, name and phase from go-fabric's constants,
  no device text (`gofabric.go:368-378`); `failed` carries a kind. That two
  commissionings share one stream without a correlation is UR-P1-api-18.
- Abort and context end before CommissioningComplete (M, probe run): cut
  at PASE (2.2) and at 11.1 — `Pair` answers the context's error, the
  source lists nothing, the device has no fabric and its own window is
  open, a second attempt commissions at once. Cut at 20.1: UR-P2-err-09.
  After the commissioner is done: UR-P2-adpt-03.
- The wrong device: `CommissionDiscovered` takes the cached device by
  instance name, else rediscovers by long discriminator and instance name
  (`gofabric.go:551-563`); PASE on the request's passcode decides. Long
  discriminator 0 is UR-P1-adpt-65.
- Page: the code field is cleared after the call (`Commission.svelte:165-166`),
  a discovered device's passcode after success (`:192-195`); no abort
  control exists, leaving the page leaves the request running, closing the
  tab ends it (the request's context is the commissioning's).
