# Review phase 1a — inventory (script, read-only map)

Generated at `49cc081` by grep over the non-test Go files of `fabricui/` and `fabricui/gofabric/`. A map, never a citation: every row is a grep hit, not a reading.

## 0. Files and sizes

```
   943 fabricui/controller.go
    58 fabricui/devicetype_names.go
    23 fabricui/doc.go
   345 fabricui/events.go
   107 fabricui/gofabric/diagnostics.go
   626 fabricui/gofabric/gofabric.go
   261 fabricui/gofabric/hooks.go
   855 fabricui/gofabric/node.go
   362 fabricui/gofabric/topology.go
    96 fabricui/gofabric/updates.go
   619 fabricui/handler.go
  1876 fabricui/handlers_controller.go
   385 fabricui/handlers_diagnostics.go
   221 fabricui/handlers_exposure.go
   580 fabricui/handlers_node.go
   276 fabricui/handlers_pairing.go
    31 fabricui/handlers_surface.go
   301 fabricui/handlers_topology.go
  1167 fabricui/memory_controller.go
   328 fabricui/node_detail.go
   183 fabricui/nodeswitch.go
   415 fabricui/operations.go
   347 fabricui/options.go
   188 fabricui/ports.go
   140 fabricui/problem.go
   165 fabricui/provider.go
   331 fabricui/registry.go
   207 fabricui/spa.go
    99 fabricui/stream.go
 11535 total
```

## 1. Goroutines

Pattern: `(^|[^a-zA-Z_])go (func|[a-zA-Z_.]+\()`

```
fabricui/gofabric/gofabric.go:573: go func() { //nolint:gosec,contextcheck // G118: the push outlives the request by design
```

## 2. Locks and atomics

Pattern: `\.(Lock|Unlock|RLock|RUnlock)\(\)|sync\.(Mutex|RWMutex|Once|WaitGroup|Map)|atomic\.`

```
fabricui/gofabric/updates.go:67: s.topoMu.Lock()
fabricui/gofabric/updates.go:69: s.topoMu.Unlock()
fabricui/events.go:199: mu sync.Mutex
fabricui/events.go:209: e.mu.Lock()
fabricui/events.go:213: e.mu.Unlock()
fabricui/events.go:215: e.mu.Lock()
fabricui/events.go:217: e.mu.Unlock()
fabricui/events.go:223: e.mu.Lock()
fabricui/events.go:228: e.mu.Unlock()
fabricui/events.go:264: var mu sync.Mutex
fabricui/events.go:282: mu.Lock()
fabricui/events.go:285: mu.Unlock()
fabricui/gofabric/gofabric.go:126: mu sync.Mutex
fabricui/gofabric/gofabric.go:414: s.mu.Lock()
fabricui/gofabric/gofabric.go:418: s.mu.Unlock()
fabricui/gofabric/gofabric.go:481: s.mu.Lock()
fabricui/gofabric/gofabric.go:483: s.mu.Unlock()
fabricui/gofabric/gofabric.go:526: s.mu.Lock()
fabricui/gofabric/gofabric.go:528: s.mu.Unlock()
fabricui/handler.go:79: config atomic.Pointer[NodeConfigView]
fabricui/handler.go:82: regMu sync.Mutex
fabricui/handler.go:87: mu sync.Mutex
fabricui/gofabric/topology.go:64: s.topoMu.Lock()
fabricui/gofabric/topology.go:66: s.topoMu.Unlock()
fabricui/gofabric/topology.go:73: s.topoMu.Lock()
fabricui/gofabric/topology.go:76: s.topoMu.Unlock()
fabricui/gofabric/topology.go:89: topoMu sync.Mutex
fabricui/gofabric/hooks.go:32: mu sync.Mutex
fabricui/gofabric/hooks.go:53: h.mu.Lock()
fabricui/gofabric/hooks.go:55: h.mu.Unlock()
fabricui/gofabric/hooks.go:60: h.mu.Unlock()
fabricui/gofabric/hooks.go:73: h.mu.Lock()
fabricui/gofabric/hooks.go:81: h.mu.Unlock()
fabricui/gofabric/hooks.go:89: h.mu.Lock()
fabricui/gofabric/hooks.go:90: defer h.mu.Unlock()
fabricui/gofabric/hooks.go:105: h.mu.Lock()
fabricui/gofabric/hooks.go:107: h.mu.Unlock()
fabricui/gofabric/hooks.go:114: h.mu.Lock()
fabricui/gofabric/hooks.go:116: h.mu.Unlock()
fabricui/gofabric/hooks.go:138: h.mu.Lock()
fabricui/gofabric/hooks.go:139: defer h.mu.Unlock()
fabricui/gofabric/hooks.go:145: h.mu.Lock()
fabricui/gofabric/hooks.go:150: h.mu.Unlock()
fabricui/operations.go:300: mu sync.Mutex
fabricui/operations.go:325: m.mu.Lock()
fabricui/operations.go:327: m.mu.Unlock()
fabricui/operations.go:340: m.mu.Unlock()
fabricui/operations.go:357: m.mu.Lock()
fabricui/operations.go:358: defer m.mu.Unlock()
fabricui/operations.go:364: m.mu.Lock()
fabricui/operations.go:366: m.mu.Unlock()
fabricui/operations.go:372: m.mu.Lock()
fabricui/operations.go:373: defer m.mu.Unlock()
fabricui/operations.go:390: m.mu.Lock()
fabricui/operations.go:394: m.mu.Unlock()
fabricui/operations.go:396: m.mu.Lock()
fabricui/operations.go:398: m.mu.Unlock()
fabricui/operations.go:408: m.mu.Lock()
fabricui/operations.go:410: m.mu.Unlock()
fabricui/registry.go:153: h.regMu.Lock()
fabricui/registry.go:159: h.regMu.Unlock()
fabricui/registry.go:167: h.regMu.Lock()
fabricui/registry.go:168: defer h.regMu.Unlock()
fabricui/registry.go:318: h.regMu.Lock()
fabricui/registry.go:320: h.regMu.Unlock()
fabricui/handlers_pairing.go:98: h.mu.Lock()
fabricui/handlers_pairing.go:99: defer h.mu.Unlock()
fabricui/handlers_pairing.go:233: h.mu.Lock()
fabricui/handlers_pairing.go:235: h.mu.Unlock()
fabricui/handlers_pairing.go:273: h.mu.Lock()
fabricui/handlers_pairing.go:275: h.mu.Unlock()
fabricui/nodeswitch.go:39: // cur points at the interface value, not an atomic.Value: a Store of
fabricui/nodeswitch.go:42: cur atomic.Pointer[NodeSource]
fabricui/gofabric/node.go:737: var wg sync.WaitGroup
fabricui/memory_controller.go:22: mu sync.Mutex
fabricui/memory_controller.go:49: m.mu.Lock()
fabricui/memory_controller.go:50: defer m.mu.Unlock()
fabricui/memory_controller.go:56: m.mu.Lock()
fabricui/memory_controller.go:57: defer m.mu.Unlock()
fabricui/memory_controller.go:73: m.mu.Lock()
fabricui/memory_controller.go:74: defer m.mu.Unlock()
fabricui/memory_controller.go:92: m.mu.Lock()
fabricui/memory_controller.go:93: defer m.mu.Unlock()
fabricui/memory_controller.go:100: m.mu.Lock()
fabricui/memory_controller.go:101: defer m.mu.Unlock()
fabricui/memory_controller.go:119: m.mu.Lock()
fabricui/memory_controller.go:120: defer m.mu.Unlock()
fabricui/memory_controller.go:127: m.mu.Lock()
fabricui/memory_controller.go:128: defer m.mu.Unlock()
fabricui/memory_controller.go:181: mu sync.Mutex
fabricui/memory_controller.go:230: m.mu.Lock()
fabricui/memory_controller.go:231: defer m.mu.Unlock()
fabricui/memory_controller.go:242: m.mu.Lock()
fabricui/memory_controller.go:243: defer m.mu.Unlock()
fabricui/memory_controller.go:253: m.mu.Lock()
fabricui/memory_controller.go:254: defer m.mu.Unlock()
fabricui/memory_controller.go:312: m.mu.Lock()
fabricui/memory_controller.go:314: m.mu.Unlock()
fabricui/memory_controller.go:342: m.mu.Lock()
fabricui/memory_controller.go:344: m.mu.Unlock()
fabricui/memory_controller.go:348: m.mu.Unlock()
fabricui/memory_controller.go:358: m.mu.Unlock()
fabricui/memory_controller.go:370: m.mu.Lock()
fabricui/memory_controller.go:372: m.mu.Unlock()
fabricui/memory_controller.go:422: m.mu.Lock()
fabricui/memory_controller.go:424: m.mu.Unlock()
fabricui/memory_controller.go:428: m.mu.Unlock()
fabricui/memory_controller.go:435: m.mu.Lock()
fabricui/memory_controller.go:437: m.mu.Unlock()
fabricui/memory_controller.go:473: m.mu.Lock()
fabricui/memory_controller.go:479: m.mu.Unlock()
fabricui/memory_controller.go:515: m.mu.Lock()
fabricui/memory_controller.go:517: m.mu.Unlock()
fabricui/memory_controller.go:567: u.m.mu.Lock()
fabricui/memory_controller.go:568: defer u.m.mu.Unlock()
fabricui/memory_controller.go:589: u.m.mu.Lock()
fabricui/memory_controller.go:590: defer u.m.mu.Unlock()
fabricui/memory_controller.go:642: mu sync.Mutex
fabricui/memory_controller.go:768: n.mu.Lock()
fabricui/memory_controller.go:769: defer n.mu.Unlock()
fabricui/memory_controller.go:774: n.mu.Lock()
fabricui/memory_controller.go:775: defer n.mu.Unlock()
fabricui/memory_controller.go:780: n.mu.Lock()
fabricui/memory_controller.go:781: defer n.mu.Unlock()
fabricui/memory_controller.go:786: n.mu.Lock()
fabricui/memory_controller.go:787: defer n.mu.Unlock()
fabricui/memory_controller.go:795: n.mu.Lock()
fabricui/memory_controller.go:796: defer n.mu.Unlock()
fabricui/memory_controller.go:801: n.mu.Lock()
fabricui/memory_controller.go:802: defer n.mu.Unlock()
fabricui/memory_controller.go:807: n.mu.Lock()
fabricui/memory_controller.go:808: defer n.mu.Unlock()
fabricui/memory_controller.go:823: n.mu.Lock()
fabricui/memory_controller.go:824: defer n.mu.Unlock()
fabricui/memory_controller.go:855: n.mu.Lock()
fabricui/memory_controller.go:856: defer n.mu.Unlock()
fabricui/memory_controller.go:872: n.mu.Lock()
fabricui/memory_controller.go:873: defer n.mu.Unlock()
fabricui/memory_controller.go:923: n.mu.Lock()
fabricui/memory_controller.go:924: defer n.mu.Unlock()
fabricui/memory_controller.go:951: n.mu.Lock()
fabricui/memory_controller.go:956: n.mu.Unlock()
fabricui/memory_controller.go:1023: n.mu.Lock()
fabricui/memory_controller.go:1037: n.mu.Unlock()
fabricui/memory_controller.go:1040: n.mu.Lock()
fabricui/memory_controller.go:1046: n.mu.Unlock()
fabricui/memory_controller.go:1123: n.mu.Lock()
fabricui/memory_controller.go:1125: n.mu.Unlock()
fabricui/memory_controller.go:1130: n.mu.Lock()
fabricui/memory_controller.go:1132: n.mu.Unlock()
fabricui/memory_controller.go:1138: n.mu.Lock()
fabricui/memory_controller.go:1139: defer n.mu.Unlock()
fabricui/memory_controller.go:1147: n.mu.Lock()
fabricui/memory_controller.go:1148: defer n.mu.Unlock()
fabricui/memory_controller.go:1160: n.mu.Lock()
fabricui/memory_controller.go:1161: defer n.mu.Unlock()
```

## 3. Channels, select, close

Pattern: `make\(chan|close\(|select \{|<-ctx\.Done|<-[a-z]+\.Done`

```
fabricui/gofabric/node.go:736: sem := make(chan struct{}, pingParallel)
fabricui/stream.go:33: ch := make(chan Event, streamBuffer)
fabricui/stream.go:35: select {
fabricui/stream.go:57: select {
fabricui/stream.go:58: case <-ctx.Done():
```

## 4. Subscriptions and observers

Pattern: `(Observe|Subscribe|subscribe|unsubscribe|Unsubscribe|cancel\(\)|Stop\(\))`

```
fabricui/gofabric/diagnostics.go:99:func (s *source) OnDiagnosticEvent(fn func(fabricui.DiagnosticEvent)) (unsubscribe func()) {
fabricui/events.go:22: // go-fabric's node.Observers.OnFabricUpdated, an UpdateNOC,
fabricui/events.go:26: // index alone: node.Observers.OnCommissioningComplete, a successful
fabricui/events.go:30: // /node/endpoints/{ep}/events: node.Observers.OnEvent, every event a
fabricui/events.go:39: // carry a [SessionEvent]: go-fabric's node.Observers OnSessionOpened
fabricui/events.go:45: // node.Observers.OnAttributeChanged, ids only.
fabricui/events.go:85:// Observers.OnSurfaceChanged reports it (one endpoint per event) — the key
fabricui/events.go:93:// Observers.OnAttributeChanged reports them (go-fabric node/observers.go
fabricui/events.go:138:// ObserveNode registers push for every transition of n the node stream
fabricui/events.go:139:// carries — [NodeObservers] through n.Observe, and the health probe's
fabricui/events.go:142:// which node.Observers does not hold — and returns the function that
fabricui/events.go:146:// unobserve := fabricui.ObserveNode(n, push)
fabricui/events.go:147:func ObserveNode(n NodeSource, push func(Event)) (unsubscribe func()) {
fabricui/events.go:151:// observeNode is ObserveNode on the handler's port.
fabricui/events.go:152:func observeNode(p nodePort, push func(Event)) (unsubscribe func()) {
fabricui/events.go:154: p.observe(NodeObservers(push, p.window().Open)),
fabricui/events.go:188:// own: Subscribe registers fn for every event from now on and returns the
fabricui/events.go:189:// unsubscribe function. fn must not block.
fabricui/events.go:191: Subscribe(fn func(Event)) (unsubscribe func())
fabricui/events.go:197:// Node.Observe and need no wiring. It is safe for concurrent use.
fabricui/events.go:207:// Subscribe implements [EventSource].
fabricui/events.go:208:func (e *Events) Subscribe(fn func(Event)) (unsubscribe func()) {
fabricui/events.go:221:// Publish hands ev to every subscriber, on the caller's goroutine.
fabricui/events.go:234:// NodeObservers turns the node's fabric, window, surface and session
fabricui/events.go:242:// three. open is the window state the subscriber last saw, so a repeated
fabricui/events.go:253:// The handler registers it with node.Observe for each open stream, and the
fabricui/events.go:257:// health samples and the trace entries too, with [ObserveNode]:
fabricui/events.go:259:// unobserve := n.Observe(fabricui.NodeObservers(push, n.Window().Open))
fabricui/events.go:263:func NodeObservers(push func(Event), open bool) node.Observers {
fabricui/events.go:265: return node.Observers{
fabricui/gofabric/hooks.go:16:// subscribed node reports every change, and a page reloads the node on
fabricui/gofabric/hooks.go:78: t.Stop()
fabricui/gofabric/node.go:712: defer cancel()
fabricui/gofabric/node.go:799:func (w *node) SetAutoSubscribe(ctx context.Context, on bool) error {
fabricui/gofabric/node.go:800: return nodeError(w.n.SetAutoSubscribe(ctx, on))
fabricui/gofabric/updates.go:58:// watchUpdates subscribes the source to the update manager's events
fabricui/gofabric/updates.go:59:// (softwareupdate.Manager.Observe, controller/softwareupdate/
fabricui/gofabric/updates.go:66: unobserve := m.Observe(func(e softwareupdate.UpdateEvent) { s.publish(updateEvent(&e)) })
fabricui/gofabric/topology.go:56:// watchTopology subscribes the source to the service's changes: every
fabricui/gofabric/topology.go:61: unsubscribe := svc.OnChanged(func(topology.Topology) {
fabricui/gofabric/topology.go:65: s.unwatch = unsubscribe
fabricui/gofabric/topology.go:74: unsubscribe, unobserve := s.unwatch, s.unobserve
fabricui/gofabric/topology.go:77: if unsubscribe != nil {
fabricui/gofabric/topology.go:78: unsubscribe()
fabricui/controller.go:54: // Observe registers fn for every controller event from now on and
fabricui/controller.go:56: Observe(fn func(ControllerEvent)) (unsubscribe func())
fabricui/controller.go:78: OnDiagnosticEvent(fn func(DiagnosticEvent)) (unsubscribe func())
fabricui/controller.go:275: SetAutoSubscribe(ctx context.Context, on bool) error
fabricui/controller.go:740:// name verbatim, plus go-fabric's Discovery 2.1, PASE 2.2 and Subscribe
fabricui/options.go:40: Observe(o node.Observers) (unsubscribe func())
fabricui/options.go:70: OnHealthSample(fn func(node.HealthSample)) (unsubscribe func())
fabricui/options.go:71: OnDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func())
fabricui/options.go:247: // stream through Node.Observe, registered for each stream while it is
fabricui/options.go:264: // with 503 and Retry-After: a controller that subscribes before the
fabricui/registry.go:47: // keeps its state (a window's credentials, its stream's subscribers)
fabricui/gofabric/gofabric.go:518: AutoSubscribe: po.AutoSubscribe,
fabricui/gofabric/gofabric.go:575: defer cancel()
fabricui/gofabric/gofabric.go:602:func (s *source) Observe(fn func(fabricui.ControllerEvent)) (unsubscribe func()) {
fabricui/gofabric/gofabric.go:603: return s.events.Subscribe(func(e fabricui.Event) {
fabricui/handler.go:50: // node.Observe, per stream (subscribe).
fabricui/handler.go:589:// publish hands ev to the stream's subscribers.
fabricui/handler.go:592:// subscribe registers fn for every event from now on — the handler's own,
fabricui/handler.go:593:// the host's source and the node's transitions (node.Observe, the health
fabricui/handler.go:598:func (h *handler) subscribe(fn func(Event)) (unsubscribe func()) {
fabricui/handler.go:599: unsubs := []func(){h.hub.Subscribe(fn)}
fabricui/handler.go:601: unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
fabricui/handler.go:604: unsubs = append(unsubs, h.extra.Subscribe(fn))
fabricui/stream.go:20:// own routes, the node's transitions (node.Observe) and the host's
fabricui/stream.go:25: h.serveSSE(r.Context(), w, h.subscribe)
fabricui/stream.go:28:// serveSSE serves server-sent events from subscribe until ctx ends, with
fabricui/stream.go:31:func (h *handler) serveSSE(ctx context.Context, w http.ResponseWriter, subscribe func(func(Event)) func()) {
fabricui/stream.go:41: // open (node.Observe) and removed when it closes.
fabricui/stream.go:42: defer subscribe(push)()
fabricui/stream.go:55: defer tick.Stop()
fabricui/handlers_controller.go:56: "POST /controller/nodes/{id}/auto-subscribe": true,
fabricui/handlers_controller.go:632:func (s *controllerServer) SetNodeAutoSubscribe(ctx context.Context, req api.SetNodeAutoSubscribeRequestObject) (api.SetNodeAutoSubscr
fabricui/handlers_controller.go:640: if err := n.SetAutoSubscribe(ctx, req.Body.Enabled); err != nil {
fabricui/handlers_controller.go:643: return api.SetNodeAutoSubscribe204Response{}, nil
fabricui/handlers_controller.go:1284: defer cancel()
fabricui/handlers_controller.go:1632: c.h.serveSSE(c.ctx, w, c.h.subscribeController)
fabricui/handlers_controller.go:1640:// subscribeController registers fn for the controller's events and the
fabricui/handlers_controller.go:1643:func (h *handler) subscribeController(fn func(Event)) (unsubscribe func()) {
fabricui/handlers_controller.go:1644: unsubs := []func(){h.ctrlHub.Subscribe(fn)}
fabricui/handlers_controller.go:1646: unsubs = append(unsubs, h.ctrl.Observe(func(e ControllerEvent) { fn(controllerEvent(e)) }))
fabricui/memory_controller.go:153:// commissioning reports the steps Discovery 2.1, PASE 2.2 and Subscribe
fabricui/memory_controller.go:284:// StepSubscribe; controller/commissioner/progress.go:78-80 at v0.18.0).
fabricui/memory_controller.go:288: memoryStepSubscribe = PairingStep{Number: 21, Sub: 1, Name: "Subscribe"}
fabricui/memory_controller.go:340:// and Subscribe.
fabricui/memory_controller.go:362: report(memoryStepSubscribe, "started")
fabricui/memory_controller.go:363: report(memoryStepSubscribe, "succeeded")
fabricui/memory_controller.go:447:// Observe implements [ControllerSource].
fabricui/memory_controller.go:448:func (m *MemoryController) Observe(fn func(ControllerEvent)) (unsubscribe func()) {
fabricui/memory_controller.go:449: return m.subs.Subscribe(func(e Event) {
fabricui/memory_controller.go:553:func (m *MemoryController) OnDiagnosticEvent(func(DiagnosticEvent)) (unsubscribe func()) {
fabricui/memory_controller.go:644: // disabled and subscribed are the node's management state.
fabricui/memory_controller.go:645: disabled, subscribed bool
fabricui/memory_controller.go:657: return &memoryNode{m: m, id: id, spec: s, subscribed: s.conn == ConnectionConnected}
fabricui/memory_controller.go:797: return n.subscribed && !n.disabled && n.spec.conn == ConnectionConnected
fabricui/memory_controller.go:1129:func (n *memoryNode) SetAutoSubscribe(_ context.Context, on bool) error {
fabricui/memory_controller.go:1131: n.subscribed = on
fabricui/nodeswitch.go:127:// Observe implements [NodeSource] on the current target.
fabricui/nodeswitch.go:128:func (s *NodeSwitch) Observe(o node.Observers) (unsubscribe func()) { return s.Current().Observe(o) }
fabricui/nodeswitch.go:162:func (s *NodeSwitch) OnHealthSample(fn func(node.HealthSample)) (unsubscribe func()) {
fabricui/nodeswitch.go:167:func (s *NodeSwitch) OnDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func()) {
fabricui/ports.go:38: observe(o node.Observers) (unsubscribe func())
fabricui/ports.go:52: onHealthSample(fn func(node.HealthSample)) (unsubscribe func())
fabricui/ports.go:53: onDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func())
fabricui/ports.go:113:func (a nodeAdapter) observe(o node.Observers) (unsubscribe func()) { return a.n.Observe(o) }
fabricui/ports.go:165:func (a nodeAdapter) onHealthSample(fn func(node.HealthSample)) (unsubscribe func()) {
fabricui/ports.go:169:func (a nodeAdapter) onDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func()) {
fabricui/handlers_pairing.go:155: // A controller that subscribes before the first endpoint is in the
fabricui/operations.go:35: // ObserveLog registers fn for every record the ring takes from now
fabricui/operations.go:38: ObserveLog(fn func(LogRecord)) (unsubscribe func())
fabricui/operations.go:280: return h.ops.ObserveLog(func(rec LogRecord) {
fabricui/operations.go:388:// ObserveLog implements [Operations].
fabricui/operations.go:389:func (m *MemoryOperations) ObserveLog(fn func(LogRecord)) (unsubscribe func()) {
```

## 5. Context and timeouts

Pattern: `context\.(WithTimeout|WithCancel|WithDeadline|Background|TODO|WithoutCancel)|time\.(After|NewTimer|NewTicker|AfterFunc)`

```
fabricui/gofabric/hooks.go:144: t = time.AfterFunc(h.delay, func() {
fabricui/gofabric/node.go:711: rctx, cancel := context.WithTimeout(ctx, pingTimeout)
fabricui/gofabric/gofabric.go:574: ctx, cancel := context.WithTimeout(context.Background(), labelPushTimeout)
fabricui/handler.go:387: req, err := http.NewRequestWithContext(context.Background(), method, path, http.NoBody)
fabricui/memory_controller.go:211: _ = m.creds.SetWiFi(context.Background(), "home", "HomeNet", "secret-passphrase")
fabricui/memory_controller.go:212: _ = m.creds.SetThread(context.Background(), "mesh", []byte{0x03, 0x04, 'M', 'e', 's', 'h'})
fabricui/stream.go:54: tick := time.NewTicker(h.heartbeat)
fabricui/handlers_controller.go:1283: ctx, cancel = context.WithTimeout(ctx, time.Duration(*b.TimeoutSeconds)*time.Second)
```

## 6. Parse and decode sites

Pattern: `json\.(Unmarshal|NewDecoder)|\.Decode\(|strconv\.|ParseUint|ParseInt|hex\.Decode|base64|url\.|Sscanf|readJSON|decodeBody`

```
fabricui/handlers_node.go:275: idx, err := strconv.ParseUint(r.PathValue("index"), 10, 8)
fabricui/handlers_node.go:326: _ = json.NewDecoder(r.Body).Decode(&body)
fabricui/handlers_node.go:377: _ = json.NewDecoder(r.Body).Decode(&body)
fabricui/handlers_node.go:447: v, err := strconv.Atoi(s)
fabricui/handlers_node.go:456: v, err := strconv.ParseUint(r.PathValue("ep"), 10, 16)
fabricui/handlers_node.go:540: if err := json.Unmarshal(b, &rep); err != nil {
fabricui/handlers_topology.go:137: topologyDetail{"endpoints", strconv.Itoa(countDevices(topo.Devices))},
fabricui/handlers_topology.go:138: topologyDetail{"fabrics", strconv.Itoa(len(fabrics))},
fabricui/handlers_topology.go:139: topologyDetail{"sessions", strconv.Itoa(len(diag.Sessions))},
fabricui/handlers_topology.go:146: Ref: &topologyRef{Page: "endpoints", ID: strconv.Itoa(int(topo.Aggregator))},
fabricui/handlers_topology.go:147: Details: []topologyDetail{{"endpoint", strconv.Itoa(int(topo.Aggregator))}, {"devices", strconv.Itoa(len(topo.Devices))}},
fabricui/handlers_topology.go:213:func endpointNodeID(ep uint16) string { return "ep:" + strconv.Itoa(int(ep)) }
fabricui/handlers_topology.go:214:func fabricNodeID(idx uint8) string { return "fabric:" + strconv.Itoa(int(idx)) }
fabricui/handlers_topology.go:215:func sessionNodeID(id uint16) string { return "session:" + strconv.Itoa(int(id)) }
fabricui/handlers_topology.go:231: {"endpoint", strconv.Itoa(int(d.Endpoint))},
fabricui/handlers_topology.go:235: details = append(details, topologyDetail{"device_type_revision", strconv.Itoa(int(d.DeviceTypeRevision))})
fabricui/handlers_topology.go:237: details = append(details, topologyDetail{"clusters", strconv.Itoa(len(d.Clusters))})
fabricui/handlers_topology.go:243: Ref: &topologyRef{Page: "endpoints", ID: strconv.Itoa(int(d.Endpoint))},
fabricui/handlers_topology.go:270: Ref: &topologyRef{Page: "fabrics", ID: strconv.Itoa(int(f.FabricIndex))},
fabricui/handlers_topology.go:291: Ref: &topologyRef{Page: "diagnostics", ID: strconv.Itoa(int(s.ID))},
fabricui/handlers_topology.go:295: {"subscriptions", strconv.Itoa(s.Subscriptions)},
fabricui/handlers_topology.go:296: {"peer_idle_seconds", strconv.FormatInt(int64(s.PeerIdle.Seconds()), 10)},
fabricui/gofabric/node.go:119: return json.Unmarshal(raw, v) == nil
fabricui/gofabric/node.go:383: _ = json.Unmarshal(raw, &events)
fabricui/gofabric/node.go:494: if v.Status == 0 && json.Unmarshal(v.Value, &idx) == nil && idx != 0 {
fabricui/operations.go:215: "level "+strconv.Quote(body.Level)+" is not one of debug, info, warn, error")
fabricui/node_detail.go:295: return err == nil && json.Unmarshal(raw, v) == nil
fabricui/problem.go:127: dec := json.NewDecoder(r.Body)
fabricui/problem.go:129: return dec.Decode(v)
fabricui/spa.go:9: "encoding/base64"
fabricui/spa.go:148: return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
fabricui/handlers_controller.go:131: if on, err := strconv.ParseBool(r.URL.Query().Get(q)); err == nil && on {
fabricui/handlers_controller.go:213: v, err := strconv.ParseUint(t, 16, bits)
fabricui/handlers_controller.go:829: return nil, fmt.Errorf("%w: status %#x for %s/%s/%s", ErrNodeFailed, v.Status, strconv.Itoa(int(path.Endpoint)), hexID(path.Cluster),
fabricui/handlers_controller.go:833: return nil, fmt.Errorf("%w: no value for %s/%s/%s", ErrNodeFailed, strconv.Itoa(int(path.Endpoint)), hexID(path.Cluster), hexID(path.
fabricui/handlers_controller.go:880: if err := json.Unmarshal(raw, &entries); err != nil {
fabricui/handlers_controller.go:1017: if err := json.Unmarshal(v.Value, &targets); err != nil {
fabricui/handlers_controller.go:1173: if err := json.Unmarshal(raw, &idx); err != nil {
fabricui/handlers_controller.go:1193: if err := json.Unmarshal(raw, &fabrics); err != nil {
fabricui/handlers_controller.go:1222: payload := json.RawMessage(`{"fabricIndex":` + strconv.Itoa(req.Index) + `}`)
fabricui/handlers_controller.go:1233: if res.HasResponse && json.Unmarshal(res.Response, &noc) == nil && noc.StatusCode != 0 {
fabricui/handlers_controller.go:1438: ds, herr := hex.DecodeString(*b.DatasetHex)
fabricui/memory_controller.go:1020: if err := json.Unmarshal(payload, &req); err != nil {
fabricui/memory_controller.go:1026: _ = json.Unmarshal(n.spec.values[path], &fabrics)
```

## 7. Log calls

Pattern: `slog\.|\.(Info|Warn|Error|Debug)(Context)?\(|log\.Print`

```
fabricui/gofabric/gofabric.go:72: // Logger receives the adapter's logs; nil is slog.Default(). Nothing
fabricui/gofabric/gofabric.go:74: Logger *slog.Logger
fabricui/gofabric/gofabric.go:103: s.log = slog.Default()
fabricui/gofabric/gofabric.go:122: log *slog.Logger
fabricui/gofabric/gofabric.go:307: s.log.Info("fabricui.pair.ble_skipped", slog.String("reason", string(d.Reason)))
fabricui/gofabric/gofabric.go:581: s.log.Warn("fabricui.gofabric.label_push_failed", slog.String("error", err.Error()))
fabricui/handler.go:54: log *slog.Logger
fabricui/handler.go:258: h.log = slog.Default()
fabricui/gofabric/node.go:755: pa.Error = o.err.Error()
fabricui/gofabric/node.go:782: res.Error = err.Error()
fabricui/gofabric/updates.go:83: out.UpdateErr = e.Err.Error()
fabricui/handlers_exposure.go:150: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, title, err.Error())
fabricui/handlers_exposure.go:214: h.log.WarnContext(r.Context(), "fabricui.exposure.reassemble_failed", slog.String("error", err.Error()))
fabricui/handlers_exposure.go:219: h.log.InfoContext(r.Context(), "fabricui.exposure.applied", slog.Int("applied", applied))
fabricui/handlers_pairing.go:79: h.log.WarnContext(r.Context(), "fabricui.pairing.codes_unavailable", slog.String("error", err.Error()))
fabricui/handlers_pairing.go:180: h.log.WarnContext(r.Context(), "fabricui.window.blocked", slog.String("error", err.Error()))
fabricui/handlers_pairing.go:197: h.log.InfoContext(r.Context(), "fabricui.window.opened",
fabricui/handlers_pairing.go:198: slog.Int("discriminator", int(*res.Discriminator)), slog.Int("duration_seconds", int(res.DurationSeconds)),
fabricui/handlers_pairing.go:199: slog.Bool("configured", h.configured))
fabricui/handlers_diagnostics.go:178: h.log.InfoContext(r.Context(), "fabricui.mdns.reannounced")
fabricui/handlers_node.go:175: res.AttestationError = err.Error()
fabricui/handlers_node.go:285: h.log.InfoContext(r.Context(), "fabricui.fabric.removed", slog.Int("fabric_index", int(idx)))
fabricui/handlers_node.go:343: h.log.WarnContext(r.Context(), "fabricui.factory_reset.partial",
fabricui/handlers_node.go:344: slog.Int("removed", len(removed)), slog.Int("of", len(infos)), slog.Any("fabric_indexes", removed))
fabricui/handlers_node.go:350: h.log.InfoContext(r.Context(), "fabricui.factory_reset.done", slog.Int("removed", len(removed)))
fabricui/handlers_node.go:388: h.log.InfoContext(r.Context(), "fabricui.erase.done")
fabricui/handlers_node.go:471: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such endpoint", err.Error())
fabricui/handlers_node.go:473: unready(w, r, "Node not started", err.Error())
fabricui/handlers_controller.go:109: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request parameter", err.Error())
fabricui/handlers_controller.go:174: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", err.Error())
fabricui/handlers_controller.go:176: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such endpoint", err.Error())
fabricui/handlers_controller.go:178: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such cluster", err.Error())
fabricui/handlers_controller.go:180: writeProblem(w, r, http.StatusServiceUnavailable, codeServiceUnready, "Not read yet", err.Error())
fabricui/handlers_controller.go:182: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No device answered", err.Error())
fabricui/handlers_controller.go:184: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such credential", err.Error())
fabricui/handlers_controller.go:186: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request", err.Error())
fabricui/handlers_controller.go:188: writeProblem(w, r, http.StatusConflict, codeConflict, "Refused by the node", err.Error())
fabricui/handlers_controller.go:192: h.log.WarnContext(r.Context(), "fabricui.controller.node_failed", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
fabricui/handlers_controller.go:193: writeProblem(w, r, http.StatusBadGateway, codeNodeError, "The node failed", err.Error())
fabricui/operations.go:23:// (go-fabric issue #393): a slog.Handler with a LevelVar and a ring in
fabricui/operations.go:29: LogLevel() slog.Level
fabricui/operations.go:31: SetLogLevel(l slog.Level) error
fabricui/operations.go:48: Level slog.Level
fabricui/operations.go:50: // Attrs are the record's attributes as text (slog.Value.String), a
fabricui/operations.go:90: level slog.Level
fabricui/operations.go:92: {"debug", slog.LevelDebug},
fabricui/operations.go:93: {"info", slog.LevelInfo},
fabricui/operations.go:94: {"warn", slog.LevelWarn},
fabricui/operations.go:95: {"error", slog.LevelError},
fabricui/operations.go:100:func levelName(l slog.Level) string {
fabricui/operations.go:111:func parseLevel(s string) (slog.Level, bool) {
fabricui/operations.go:185: in := h.ops.Info()
fabricui/operations.go:222: h.log.InfoContext(r.Context(), "fabricui.operation.log_level", "level", levelName(l))
fabricui/operations.go:301: level slog.Level
fabricui/operations.go:319: return &MemoryOperations{level: slog.LevelInfo, capacity: capacity, info: info, obs: map[int]func(LogRecord){}}
fabricui/operations.go:324:func (m *MemoryOperations) Record(t time.Time, l slog.Level, msg string, attrs map[string]string) (LogRecord, bool) {
fabricui/operations.go:356:func (m *MemoryOperations) LogLevel() slog.Level {
fabricui/operations.go:363:func (m *MemoryOperations) SetLogLevel(l slog.Level) error {
fabricui/options.go:253: // Logger receives the handlers' logs; nil is slog.Default(). Nothing
fabricui/options.go:255: Logger *slog.Logger
fabricui/problem.go:92: problem: newProblem(r, http.StatusConflict, codeTermsRequired, "Terms and conditions required", e.Error()),
fabricui/problem.go:104: h.log.ErrorContext(r.Context(), title, slog.String("error", err.Error()),
fabricui/problem.go:105: slog.String("method", r.Method), slog.String("path", r.URL.Path))
fabricui/problem.go:136: writeProblem(w, r, http.StatusRequestEntityTooLarge, codeTooLarge, "Request body too large", err.Error())
fabricui/problem.go:139: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request body", err.Error())
fabricui/stream.go:70: h.log.WarnContext(ctx, "fabricui.stream.invalid_type", "type", ev.Type)
fabricui/stream.go:75: h.log.WarnContext(ctx, "fabricui.stream.encode_failed", "type", ev.Type, "error", err.Error())
fabricui/registry.go:147: h.log.Warn("fabricui.registry.entry_skipped", slog.String("error", err.Error()))
fabricui/registry.go:234: writeProblem(w, r, http.StatusConflict, codeNodeLimit, "Node limit reached", err.Error())
fabricui/registry.go:236: writeProblem(w, r, http.StatusConflict, codeNodeIsDefault, "Node is configuration", err.Error())
fabricui/registry.go:252: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid label", err.Error())
fabricui/registry.go:282: h.log.InfoContext(r.Context(), "fabricui.registry.created", slog.String("key", e.Key))
fabricui/registry.go:321: h.log.InfoContext(r.Context(), "fabricui.registry.deleted", slog.String("key", key))
```

## 8. Error mapping call sites

Pattern: `writeProblem|controllerError|writeAPIError|badRequest\(|serviceUnready\(|apiProblem\{|errors\.(Is|As)\(`

```
fabricui/gofabric/gofabric.go:364: if errors.As(err, &terms) {
fabricui/gofabric/gofabric.go:368: case errors.Is(err, controller.ErrPairingCode), errors.Is(err, controller.ErrConfiguration):
fabricui/gofabric/gofabric.go:370: case errors.Is(err, controller.ErrNoDevice):
fabricui/gofabric/gofabric.go:372: case errors.Is(err, context.Canceled):
fabricui/gofabric/gofabric.go:567: if errors.Is(err, commissioner.ErrLabel) {
fabricui/gofabric/gofabric.go:617: case errors.Is(err, client.ErrJSONValue):
fabricui/gofabric/gofabric.go:619: case errors.Is(err, context.Canceled):
fabricui/gofabric/node.go:750: if errors.Is(o.err, client.ErrNoSession) || errors.Is(o.err, client.ErrNotCommissioned) || errors.Is(o.err, client.ErrClosed) {
fabricui/handler.go:433: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+r.PathValue("node"))
fabricui/handler.go:455: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
fabricui/handler.go:574: writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required", detail)
fabricui/handlers_exposure.go:139: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, fmt.Sprintf("Invalid item %d", i), "key is required")
fabricui/handlers_exposure.go:149: if errors.Is(err, ErrUnknownItem) || errors.Is(err, ErrInvalidUpdate) {
fabricui/handlers_exposure.go:150: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, title, err.Error())
fabricui/handlers_pairing.go:149: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "duration_seconds out of range",
fabricui/handlers_pairing.go:163: writeProblem(w, r, http.StatusConflict, codeConflict, title, "a commissioning window is already open")
fabricui/handlers_pairing.go:176: case errors.Is(err, node.ErrCommissioningWindowAlreadyOpen), errors.Is(err, node.ErrAdministratorWindowOpen):
fabricui/handlers_pairing.go:177: writeProblem(w, r, http.StatusConflict, codeConflict, title, "a commissioning window is already open")
fabricui/handlers_pairing.go:178: case errors.Is(err, node.ErrCommissioningBlocked):
fabricui/handlers_pairing.go:181: writeProblem(w, r, http.StatusConflict, codeConflict, title,
fabricui/handlers_pairing.go:183: case errors.Is(err, node.ErrCommissioningWindowNotConfigured):
fabricui/handlers_node.go:277: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid fabric index", "fabric_index must be 1..254")
fabricui/handlers_node.go:328: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Factory reset not confirmed",
fabricui/handlers_node.go:370: writeProblem(w, r, http.StatusNotImplemented, codeNotImplemented, "Erase not available", "the host wires no Erase")
fabricui/handlers_node.go:379: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Erase not confirmed",
fabricui/handlers_node.go:407: return 0, 0, badRequest("since: %d is not an event number", *sinceQ)
fabricui/handlers_node.go:414: return 0, 0, badRequest("limit: %d is not within 1..%d", *limitQ, maxEventLimit)
fabricui/handlers_node.go:449: return nil, badRequest("%s: %q is not an integer", name, s)
fabricui/handlers_node.go:458: return 0, badRequest("ep: %q is not an endpoint number", r.PathValue("ep"))
fabricui/handlers_node.go:463:// writeAPIError answers a node route's error: an apiProblem as it says,
fabricui/handlers_node.go:465:func (h *handler) writeAPIError(w http.ResponseWriter, r *http.Request, title string, err error) {
fabricui/handlers_node.go:468: case errors.As(err, &p):
fabricui/handlers_node.go:469: writeProblem(w, r, p.status, p.code, p.title, p.detail)
fabricui/handlers_node.go:470: case errors.Is(err, ErrEndpointNotFound):
fabricui/handlers_node.go:471: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such endpoint", err.Error())
fabricui/handlers_node.go:472: case errors.Is(err, node.ErrNotStarted):
fabricui/handlers_node.go:491: h.writeAPIError(w, r, "", err)
fabricui/handlers_node.go:496: h.writeAPIError(w, r, "Failed to read the endpoint", err)
fabricui/handlers_node.go:561: h.writeAPIError(w, r, "", err)
fabricui/handlers_node.go:566: h.writeAPIError(w, r, "", err)
fabricui/handlers_node.go:571: h.writeAPIError(w, r, "", err)
fabricui/handlers_node.go:576: h.writeAPIError(w, r, "", err)
fabricui/node_detail.go:62: if errors.As(err, &se) && se.Status == statusUnsupportedEndpoint {
fabricui/problem.go:47:// writeProblem writes a problem with the code, title and detail; the
fabricui/problem.go:49:func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
fabricui/problem.go:50: writeProblemBody(w, status, code, newProblem(r, status, code, title, detail))
fabricui/problem.go:53:// newProblem is the problem writeProblem writes, for a problem with
fabricui/problem.go:69:// writeProblemBody writes body, a problem or one embedding it, with the
fabricui/problem.go:71:func writeProblemBody(w http.ResponseWriter, status int, code string, body any) {
fabricui/problem.go:96: writeProblemBody(w, http.StatusConflict, codeTermsRequired, p)
fabricui/problem.go:106: writeProblem(w, r, http.StatusInternalServerError, codeInternal, title, "")
fabricui/problem.go:112: writeProblem(w, r, http.StatusServiceUnavailable, codeServiceUnready, title, detail)
fabricui/problem.go:136: writeProblem(w, r, http.StatusRequestEntityTooLarge, codeTooLarge, "Request body too large", err.Error())
fabricui/problem.go:139: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request body", err.Error())
fabricui/operations.go:214: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid log level",
fabricui/operations.go:234: h.writeAPIError(w, r, "Reading the log failed", err)
fabricui/operations.go:257: return 0, 0, badRequest("since: %d is not a sequence number", *sinceQ)
fabricui/operations.go:264: return 0, 0, badRequest("limit: %d is not within 1..%d", *limitQ, maxLogLimit)
fabricui/handlers_controller.go:102: ResponseErrorHandlerFunc: h.controllerError,
fabricui/handlers_controller.go:109: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request parameter", err.Error())
fabricui/handlers_controller.go:113: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
fabricui/handlers_controller.go:126: writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required",
fabricui/handlers_controller.go:132: writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required",
fabricui/handlers_controller.go:153:func badRequest(format string, args ...any) error {
fabricui/handlers_controller.go:154: return &apiProblem{http.StatusBadRequest, codeBadRequest, "Invalid request", fmt.Sprintf(format, args...)}
fabricui/handlers_controller.go:157:func serviceUnready(title, detail string) error {
fabricui/handlers_controller.go:158: return &apiProblem{http.StatusServiceUnavailable, codeServiceUnready, title, detail}
fabricui/handlers_controller.go:161:// controllerError writes the problem of a handler body's error: its own
fabricui/handlers_controller.go:164:func (h *handler) controllerError(w http.ResponseWriter, r *http.Request, err error) {
fabricui/handlers_controller.go:166: writeProblem(w, r, p.status, p.code, p.title, p.detail)
fabricui/handlers_controller.go:171: case errors.As(err, &terms):
fabricui/handlers_controller.go:173: case errors.Is(err, ErrNodeNotFound):
fabricui/handlers_controller.go:174: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", err.Error())
fabricui/handlers_controller.go:175: case errors.Is(err, ErrEndpointNotFound):
fabricui/handlers_controller.go:176: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such endpoint", err.Error())
fabricui/handlers_controller.go:177: case errors.Is(err, ErrClusterNotFound):
fabricui/handlers_controller.go:178: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such cluster", err.Error())
fabricui/handlers_controller.go:179: case errors.Is(err, ErrNotRead):
fabricui/handlers_controller.go:180: writeProblem(w, r, http.StatusServiceUnavailable, codeServiceUnready, "Not read yet", err.Error())
fabricui/handlers_controller.go:181: case errors.Is(err, ErrNoDevice):
fabricui/handlers_controller.go:182: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No device answered", err.Error())
fabricui/handlers_controller.go:183: case errors.Is(err, ErrUnknownCredential):
fabricui/handlers_controller.go:184: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such credential", err.Error())
fabricui/handlers_controller.go:185: case errors.Is(err, ErrInvalidArgument), errors.Is(err, ErrInvalidCredential):
fabricui/handlers_controller.go:186: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request", err.Error())
fabricui/handlers_controller.go:187: case errors.Is(err, ErrConflict):
fabricui/handlers_controller.go:188: writeProblem(w, r, http.StatusConflict, codeConflict, "Refused by the node", err.Error())
fabricui/handlers_controller.go:189: case errors.Is(err, ErrNotImplemented):
fabricui/handlers_controller.go:190: writeProblem(w, r, http.StatusNotImplemented, codeNotImplemented, "Not implemented", "the host's controller does not serve this route
fabricui/handlers_controller.go:191: case errors.Is(err, ErrNodeFailed), errors.Is(err, context.DeadlineExceeded):
fabricui/handlers_controller.go:193: writeProblem(w, r, http.StatusBadGateway, codeNodeError, "The node failed", err.Error())
fabricui/handlers_controller.go:223: return 0, badRequest("%s: %v", field, err)
fabricui/handlers_controller.go:231: return 0, badRequest("%s: %v", field, err)
fabricui/handlers_controller.go:238: return 0, badRequest("%s: %d is not an endpoint number", field, v)
fabricui/handlers_controller.go:249: return 0, badRequest("timed_ms: %d is not 0 to 65535", *ms)
fabricui/handlers_controller.go:638: return nil, badRequest("the body {enabled} is required")
fabricui/handlers_controller.go:657: return nil, badRequest("timeout_seconds: %d is not 0 to 65535", t)
fabricui/handlers_controller.go:708: return nil, badRequest("paths: at least one path is required")
fabricui/handlers_controller.go:711: return nil, badRequest("paths: %d paths, at most %d", len(req.Body.Paths), maxReadPaths)
fabricui/handlers_controller.go:754: return nil, badRequest("value is required")
fabricui/handlers_controller.go:778: return nil, badRequest("the body {endpoint, cluster_hex, command_hex} is required")
fabricui/handlers_controller.go:921: return nil, badRequest("the body {entries} is required")
fabricui/handlers_controller.go:927: return nil, badRequest("entries[%d].privilege: %q is unknown", i, e.Privilege)
fabricui/handlers_controller.go:931: return nil, badRequest("entries[%d].auth_mode: %q is unknown", i, e.AuthMode)
fabricui/handlers_controller.go:1045: return nil, badRequest("the body {endpoint, bindings} is required")
fabricui/handlers_controller.go:1063: return nil, badRequest("bindings[%d].group: %d is not a group id", i, *b.Group)
fabricui/handlers_controller.go:1084: return nil, badRequest("bindings[%d]: a group target names no node and no endpoint", i)
fabricui/handlers_controller.go:1086: return nil, badRequest("bindings[%d]: a unicast target names node_id_hex and endpoint", i)
fabricui/handlers_controller.go:1132: return nil, badRequest("the node has no ICD Management")
fabricui/handlers_controller.go:1147: return nil, badRequest("the node has no ICD Management")
fabricui/handlers_controller.go:1213: return nil, badRequest("index: %d is not a fabric index (1 to 254)", req.Index)
fabricui/handlers_controller.go:1220: return nil, badRequest("fabric %d is this controller's; decommission the node instead", req.Index)
fabricui/handlers_controller.go:1262: return nil, badRequest("the body {code} or {discovered, passcode} is required")
fabricui/handlers_controller.go:1267: return nil, badRequest("code and discovered name two devices; send one")
fabricui/handlers_controller.go:1269: return nil, badRequest("code or discovered is required")
fabricui/handlers_controller.go:1271: return nil, badRequest("discovered needs passcode")
fabricui/handlers_controller.go:1273: return nil, badRequest("passcode: not 1 to 99999998")
fabricui/handlers_controller.go:1276: return nil, badRequest("terms_and_conditions: version and acknowledgements are 0 to 65535")
fabricui/handlers_controller.go:1280: return nil, badRequest("timeout_seconds: %d is not 1 to 600", *b.TimeoutSeconds)
fabricui/handlers_controller.go:1323: case errors.As(err, &terms):
fabricui/handlers_controller.go:1325: case errors.Is(err, ErrInvalidArgument):
fabricui/handlers_controller.go:1327: case errors.Is(err, ErrNoDevice):
fabricui/handlers_controller.go:1329: case errors.Is(err, context.DeadlineExceeded):
fabricui/handlers_controller.go:1331: case errors.Is(err, ErrNodeFailed):
fabricui/handlers_controller.go:1343: return nil, badRequest("filter: %q is unknown", f.Kind)
fabricui/handlers_controller.go:1348: return nil, badRequest("value: filter %s needs a value", f.Kind)
fabricui/handlers_controller.go:1355: return nil, badRequest("timeout: %d is not 1 to 60", *req.Params.Timeout)
fabricui/handlers_controller.go:1385: return nil, serviceUnready("No credential store", "the host stores no Wi-Fi or Thread credentials")
fabricui/handlers_controller.go:1419: return nil, badRequest("the body {kind, id, …} is required")
fabricui/handlers_controller.go:1422: return nil, badRequest("id: 1 to 64 characters")
fabricui/handlers_controller.go:1427: return nil, badRequest("ssid is required for a Wi-Fi credential")
fabricui/handlers_controller.go:1436: return nil, badRequest("dataset_hex is required for a Thread dataset")
fabricui/handlers_controller.go:1440: return nil, badRequest("dataset_hex is not hex")
fabricui/handlers_controller.go:1444: return nil, badRequest("kind: %q is not wifi or thread", b.Kind)
fabricui/handlers_controller.go:1458: return nil, badRequest("kind: %q is not wifi or thread", req.Params.Kind)
fabricui/handlers_controller.go:1475: return nil, badRequest("the body {label} is required")
fabricui/handlers_controller.go:1478: return nil, badRequest("label: longer than %d characters", maxFabricLabel)
fabricui/handlers_controller.go:1514: return nil, serviceUnready("No OTA provider", "the controller runs no OTA provider")
fabricui/handlers_controller.go:1526: return nil, badRequest("the body is required")
fabricui/handlers_controller.go:1533: return nil, badRequest("vendor_id, product_id or target_software_version out of range")
fabricui/registry.go:219: writeProblem(w, r, http.StatusConflict, codeNodeIsDefault, "Node is configuration",
fabricui/registry.go:224: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+key)
fabricui/registry.go:233: case errors.Is(err, ErrNodeLimit):
fabricui/registry.go:234: writeProblem(w, r, http.StatusConflict, codeNodeLimit, "Node limit reached", err.Error())
fabricui/registry.go:235: case errors.Is(err, ErrDefaultNode):
fabricui/registry.go:236: writeProblem(w, r, http.StatusConflict, codeNodeIsDefault, "Node is configuration", err.Error())
fabricui/registry.go:252: writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid label", err.Error())
```

## 9. Panics and recover

Pattern: `panic\(|recover\(\)`

```
fabricui/handler.go:330: panic(err)
fabricui/handler.go:350: if r := recover(); r != nil {
fabricui/options.go:334: panic(err)
fabricui/nodeswitch.go:53: panic("fabricui: NewNodeSwitch: nil source")
fabricui/nodeswitch.go:65: panic("fabricui: NodeSwitch.Set: nil source")
fabricui/spa.go:34: panic(err)
fabricui/registry.go:119: if recover() != nil {
fabricui/memory_controller.go:967: panic(err)
```

## 10. Response headers and status codes

Pattern: `Header\(\)\.(Set|Add)|WriteHeader\(`

```
fabricui/handlers_pairing.go:158: w.Header().Set("Retry-After", "30")
fabricui/handlers_pairing.go:255: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_diagnostics.go:179: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_surface.go:28: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/handlers_surface.go:29: w.WriteHeader(http.StatusOK)
fabricui/stream.go:48: w.WriteHeader(http.StatusOK)
fabricui/problem.go:72: w.Header().Set("Content-Type", "application/problem+json")
fabricui/problem.go:73: w.Header().Set("X-Problem-Code", code)
fabricui/problem.go:74: w.WriteHeader(status)
fabricui/problem.go:118: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/problem.go:119: w.WriteHeader(http.StatusOK)
fabricui/spa.go:60: w.Header().Set("Cache-Control", cacheControlFor(path))
fabricui/spa.go:99: w.WriteHeader(i.status)
fabricui/handlers_node.go:286: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_node.go:305: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_node.go:351: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_node.go:389: w.WriteHeader(http.StatusNoContent)
fabricui/handlers_exposure.go:199: w.WriteHeader(http.StatusNoContent)
fabricui/registry.go:284: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/registry.go:285: w.WriteHeader(http.StatusCreated)
fabricui/registry.go:304: w.WriteHeader(http.StatusNoContent)
fabricui/registry.go:323: w.WriteHeader(http.StatusNoContent)
```

## 11. Unbounded growth candidates (append in loops, maps)

Pattern: `make\(map|= append\(`

```
fabricui/gofabric/diagnostics.go:35: out.Sessions = append(out.Sessions, fabricui.ControllerSession{
fabricui/gofabric/diagnostics.go:44: out.Peers = append(out.Peers, fabricui.ControllerPeer{
fabricui/gofabric/diagnostics.go:80: out = append(out, fabricui.MDNSFinding{Severity: string(f.Severity), Code: f.Code, Message: f.Message, Service: f.Service})
fabricui/gofabric/diagnostics.go:91: out = append(out, diagnosticEventOf(&events[i]))
fabricui/events.go:226: fns = append(fns, fn)
fabricui/events.go:323: ids = append(ids, hexID(a))
fabricui/gofabric/updates.go:38: out = append(out, e)
fabricui/gofabric/updates.go:50: out = append(out, fabricui.AvailableUpdate{
fabricui/gofabric/gofabric.go:187: nodes = append(nodes, w)
fabricui/gofabric/gofabric.go:201: out = append(out, n)
fabricui/gofabric/gofabric.go:421: out = append(out, discoveredOf(&found[i]))
fabricui/gofabric/gofabric.go:432: out.Addresses = append(out.Addresses, a.IP.String())
fabricui/gofabric/topology.go:104: ids := make(map[string]string, len(t.Nodes))
fabricui/gofabric/topology.go:118: out.Nodes = append(out.Nodes, fabricui.NetworkNode{
fabricui/gofabric/topology.go:129: out.Nodes = append(out.Nodes, peerOf(n, ids[n.ID], names))
fabricui/gofabric/topology.go:130: out.Connections = append(out.Connections, fabricui.NetworkConnection{Source: controllerID, Target: ids[n.ID], Kind: "session"})
fabricui/gofabric/topology.go:132: out.Nodes = append(out.Nodes, threadUnknownOf(n))
fabricui/gofabric/topology.go:134: out.Nodes = append(out.Nodes, borderRouterOf(n, t.CollectedAt))
fabricui/gofabric/topology.go:136: out.Nodes = append(out.Nodes, accessPointOf(n))
fabricui/gofabric/topology.go:138: out.Nodes = append(out.Nodes, otherOf(n))
fabricui/gofabric/topology.go:142: out.Connections = append(out.Connections, linkOf(&t.Connections[i], ids))
fabricui/gofabric/topology.go:347: *d = append(*d, fabricui.TopologyDetail{Key: key, Value: value})
fabricui/gofabric/node.go:71: out.Addresses = append(out.Addresses, netip.AddrPortFrom(a.IP, a.Port).String())
fabricui/gofabric/node.go:151: e.Clusters = append(e.Clusters, fabricui.NodeCluster{ID: c.ID(), Name: c.Name(), Revision: c.Revision(), Features: append([]string{}, c.Fea
fabricui/gofabric/node.go:153: out = append(out, e)
fabricui/gofabric/node.go:179: out = append(out, fabricui.NodeDeviceType{ID: dt.DeviceType, Name: deviceTypeName(dt.DeviceType), Revision: dt.Revision})
fabricui/gofabric/node.go:191: out = append(out, t)
fabricui/gofabric/node.go:245: d.Parts = append(d.Parts, p.Number())
fabricui/gofabric/node.go:248: d.Clusters = append(d.Clusters, clusterDetail(c, false))
fabricui/gofabric/node.go:292: out = append(out, st)
fabricui/gofabric/node.go:312: out = append(out, fabricui.EndpointLabel{Label: l.Label, Value: l.Value, User: user})
fabricui/gofabric/node.go:323: out.Violations = append(out.Violations, fabricui.ConformanceViolation{
fabricui/gofabric/node.go:328: out.Unjudged = append(out.Unjudged, fabricui.ClusterElement{ID: id, Name: clusterName(id, nil)})
fabricui/gofabric/node.go:361: d.Features = append(d.Features, f)
fabricui/gofabric/node.go:378: d.Attributes = append(d.Attributes, a)
fabricui/gofabric/node.go:393: d.Events = append(d.Events, e)
fabricui/gofabric/node.go:409: out = append(out, e)
fabricui/gofabric/node.go:440: out = append(out, r)
fabricui/gofabric/node.go:451: req.Attributes = append(req.Attributes, client.AttributeAt(p.Endpoint, p.Cluster, p.Attribute))
fabricui/gofabric/node.go:475: out = append(out, v)
fabricui/gofabric/node.go:518: ts = append(ts, accesscontrol.AccessControlTargetStruct{
fabricui/gofabric/node.go:524: out = append(out, s)
fabricui/gofabric/node.go:542: out = append(out, binding.TargetStruct{Node: t.Node, Group: t.Group, Endpoint: t.Endpoint, Cluster: t.Cluster, FabricIndex: idx})
fabricui/gofabric/node.go:709: addrs = append(addrs, netip.AddrPortFrom(a.IP, a.Port))
fabricui/gofabric/node.go:762: res.Addresses = append(res.Addresses, pa)
fabricui/gofabric/node.go:791: res.Addresses = append(res.Addresses, pa)
fabricui/handler.go:173: in = append([]NodeEntry{{Key: DefaultNodeKey, Source: o.Node}}, in...)
fabricui/handler.go:182: out = append(out, entryOf(&in[i], false))
fabricui/handler.go:419: patterns = append(patterns, pattern)
fabricui/handler.go:474: h.keyed = make(map[string]*handler, len(served))
fabricui/handler.go:555: out.Nodes = append(out.Nodes, summaryOf(e))
fabricui/handler.go:558: out.Nodes = append(out.Nodes, summaryOf(entryOf(&reg[i], true)))
fabricui/handler.go:601: unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
fabricui/handler.go:604: unsubs = append(unsubs, h.extra.Subscribe(fn))
fabricui/handler.go:607: unsubs = append(unsubs, observeNode(h.node, fn))
fabricui/handlers_controller.go:328: out.Nodes = append(out.Nodes, nodeSummary(n))
fabricui/handlers_controller.go:401: d.Properties.DeviceTypes = append(d.Properties.DeviceTypes, int(t))
fabricui/handlers_controller.go:409: e.DeviceTypes = append(e.DeviceTypes, api.NodeDeviceType{ID: int(dt.ID), IDHex: hexID(dt.ID), Name: dt.Name, Revision: nonZero(dt.Rev
fabricui/handlers_controller.go:412: e.Clusters = append(e.Clusters, api.NodeCluster{ID: int(c.ID), IDHex: hexID(c.ID), Name: c.Name, Revision: int(c.Revision), Features:
fabricui/handlers_controller.go:414: d.Endpoints = append(d.Endpoints, e)
fabricui/handlers_controller.go:438: out.DeviceTypes = append(out.DeviceTypes, api.EndpointDeviceType{ID: int(dt.ID), IDHex: hexID(dt.ID), Name: optString(dt.Name), Revis
fabricui/handlers_controller.go:441: out.Tags = append(out.Tags, api.SemanticTag{Namespace: int(t.Namespace), Tag: int(t.Tag), MfgCode: optUint(t.MfgCode), Label: optStri
fabricui/handlers_controller.go:444: out.Labels = append(out.Labels, api.EndpointLabel{Label: l.Label, Value: l.Value, User: l.User})
fabricui/handlers_controller.go:447: out.Parts = append(out.Parts, int(p))
fabricui/handlers_controller.go:450: out.Clusters = append(out.Clusters, clusterDetail(d.Clusters[i]))
fabricui/handlers_controller.go:455: v.Violations = append(v.Violations, api.ConformanceViolation{
fabricui/handlers_controller.go:475: out.Features = append(out.Features, api.ClusterFeature{Code: f.Code, Title: optString(f.Title), Bit: int(f.Bit)})
fabricui/handlers_controller.go:482: out.Attributes = append(out.Attributes, e)
fabricui/handlers_controller.go:490: out = append(out, api.ClusterElement{ID: int(e.ID), IDHex: hexID(e.ID), Name: optString(e.Name)})
fabricui/handlers_controller.go:591: out.Addresses = append(out.Addresses, pa)
fabricui/handlers_controller.go:719: paths = append(paths, ap)
fabricui/handlers_controller.go:734: out.Values = append(out.Values, e)
fabricui/handlers_controller.go:890: a.SubjectsHex = append(a.SubjectsHex, hex16(sub))
fabricui/handlers_controller.go:903: a.Targets = append(a.Targets, at)
fabricui/handlers_controller.go:905: out.Entries = append(out.Entries, a)
fabricui/handlers_controller.go:940: entry.Subjects = append(entry.Subjects, sub)
fabricui/handlers_controller.go:969: entry.Targets = append(entry.Targets, at)
fabricui/handlers_controller.go:971: entries = append(entries, entry)
fabricui/handlers_controller.go:1000: paths = append(paths, AttributePath{Endpoint: ep.Number, Cluster: clusterBinding, Attribute: attributeBinding})
fabricui/handlers_controller.go:1032: eb.Bindings = append(eb.Bindings, b)
fabricui/handlers_controller.go:1034: out.Endpoints = append(out.Endpoints, eb)
fabricui/handlers_controller.go:1088: targets = append(targets, t)
fabricui/handlers_controller.go:1198: out.Fabrics = append(out.Fabrics, api.NodeFabric{
fabricui/handlers_controller.go:1375: out.Devices = append(out.Devices, e)
fabricui/handlers_controller.go:1405: out.Credentials = append(out.Credentials, c)
fabricui/handlers_controller.go:1506: out.Queue = append(out.Queue, e)
fabricui/handlers_controller.go:1556: out.Updates = append(out.Updates, api.AvailableUpdate{
fabricui/handlers_controller.go:1604: out.Details = append(out.Details, api.TopologyDetail{Key: d.Key, Value: d.Value})
fabricui/handlers_controller.go:1606: g.Nodes = append(g.Nodes, out)
fabricui/handlers_controller.go:1619: g.Edges = append(g.Edges, e)
fabricui/handlers_controller.go:1646: unsubs = append(unsubs, h.ctrl.Observe(func(e ControllerEvent) { fn(controllerEvent(e)) }))
fabricui/handlers_controller.go:1649: unsubs = append(unsubs, u)
fabricui/handlers_controller.go:1747: out.Sessions = append(out.Sessions, controllerSessionOf(&d.Sessions[i]))
fabricui/handlers_controller.go:1750: out.Peers = append(out.Peers, controllerPeerOf(&d.Peers[i]))
fabricui/handlers_controller.go:1819: out.SessionIds = append(out.SessionIds, int(id))
fabricui/handlers_controller.go:1833: out.Findings = append(out.Findings, api.MdnsFinding{
fabricui/handlers_controller.go:1858: out.Events = append(out.Events, ev)
fabricui/handlers_diagnostics.go:121: out = append(out, sessionOf(&diag.Sessions[i]))
fabricui/handlers_diagnostics.go:255: e.DeviceTypes = append(e.DeviceTypes, api.EndpointDeviceType{ID: int(dt.ID), IDHex: hexID(dt.ID), Name: optString(dt.Name), Revision
fabricui/handlers_diagnostics.go:260: out = append(out, e)
fabricui/handlers_diagnostics.go:287: out = append(out, ep)
fabricui/handlers_diagnostics.go:302: out = append(out, clusterResponse{ID: c.ID, Name: c.Name, Revision: c.Revision})
fabricui/handlers_diagnostics.go:344: vendorIDs = append(vendorIDs, f.VendorID)
fabricui/handlers_diagnostics.go:345: out.Ecosystems = append(out.Ecosystems, ecosystemResponse{
fabricui/handlers_diagnostics.go:359: out.Findings = append(out.Findings, compatFinding{
fabricui/handlers_diagnostics.go:382: out = append(out, traceEventOf(&events[i]))
fabricui/handlers_exposure.go:53: out := make(map[string][]string, len(in))
fabricui/handlers_exposure.go:55: out[k] = append([]string{}, v...)
fabricui/handlers_exposure.go:68: clusters := append([]uint32{}, it.Verdict.Clusters...)
fabricui/handlers_exposure.go:69: out = append(out, itemResponse{
fabricui/handlers_exposure.go:209: keys = append(keys, updates[i].Key)
fabricui/handlers_topology.go:126: g.Edges = append(g.Edges, topologyEdge{Source: src, Target: dst, Kind: kind})
fabricui/handlers_topology.go:134: root.Details = append(root.Details, topologyDetail{"address", addr})
fabricui/handlers_topology.go:136: root.Details = append(root.Details,
fabricui/handlers_topology.go:141: g.Nodes = append(g.Nodes, root)
fabricui/handlers_topology.go:144: g.Nodes = append(g.Nodes, topologyNode{
fabricui/handlers_topology.go:157: g.Nodes = append(g.Nodes, deviceNode(d, id, kind))
fabricui/handlers_topology.go:164: listed := make(map[uint8]bool, len(fabrics))
fabricui/handlers_topology.go:168: g.Nodes = append(g.Nodes, fabricNode(f))
fabricui/handlers_topology.go:176: g.Nodes = append(g.Nodes, sessionNode(s))
fabricui/handlers_topology.go:185: g.Nodes = append(g.Nodes, sessionNode(s))
fabricui/handlers_topology.go:235: details = append(details, topologyDetail{"device_type_revision", strconv.Itoa(int(d.DeviceTypeRevision))})
fabricui/handlers_topology.go:237: details = append(details, topologyDetail{"clusters", strconv.Itoa(len(d.Clusters))})
fabricui/handlers_topology.go:239: details = append(details, topologyDetail{"key", d.Key.String()})
fabricui/handlers_topology.go:260: details = append(details, topologyDetail{"label", f.Label})
fabricui/handlers_topology.go:262: details = append(details,
fabricui/handlers_node.go:250: out = append(out, fabricResponse{
fabricui/handlers_node.go:348: removed = append(removed, f.FabricIndex)
fabricui/handlers_node.go:436: out.Events = append(out.Events, r)
fabricui/handlers_node.go:545: out = append(out, rep.Decisions[i])
fabricui/operations.go:169: out.Levels = append(out.Levels, n.name)
fabricui/operations.go:240: out.Records = append(out.Records, logRecordOf(&recs[i]))
fabricui/operations.go:332: m.ring = append(m.ring, rec)
fabricui/operations.go:334: m.ring = append([]LogRecord(nil), m.ring[len(m.ring)-m.capacity:]...)
fabricui/operations.go:338: fns = append(fns, fn)
fabricui/operations.go:348: out := make(map[string]string, len(in))
fabricui/operations.go:379: out = append(out, r)
fabricui/provider.go:153: out = append(out, v)
fabricui/provider.go:161: out = append(out, v)
fabricui/registry.go:140: seen := make(map[string]bool, len(list)+len(h.keyed))
fabricui/registry.go:151: out = append(out, e)
fabricui/registry.go:274: seen := make(map[string]bool, len(h.keyed))
fabricui/node_detail.go:96: d.Clusters = append(d.Clusters, servedCluster(id, vals))
fabricui/node_detail.go:99: d.Labels = append(d.Labels, labelsOf(vals, false)...)
fabricui/node_detail.go:101: d.Labels = append(d.Labels, labelsOf(vals, true)...)
fabricui/node_detail.go:130: out = append(out, infos[i].Endpoint)
fabricui/node_detail.go:140: out = append(out, NodeDeviceType{ID: dt.ID, Name: schemaDeviceTypeName(dt.ID), Revision: dt.Revision})
fabricui/node_detail.go:194: c.Features = append(c.Features, ClusterFeature{Code: f.Name, Title: f.Title, Bit: f.Bit})
fabricui/node_detail.go:204: attrs = append(attrs, v.Attribute)
fabricui/node_detail.go:218: c.Attributes = append(c.Attributes, e)
fabricui/node_detail.go:234: out = append(out, e)
fabricui/node_detail.go:246: out = append(out, ClusterElement{ID: id, Name: name})
fabricui/node_detail.go:266: out = append(out, st)
fabricui/node_detail.go:280: out = append(out, EndpointLabel{Label: l.Label, Value: l.Value, User: user})
fabricui/ports.go:126: addrs = append(addrs, ip.String())
fabricui/ports.go:128: txt := make(map[string]string, len(svc.TXT))
fabricui/ports.go:132: subtypes := append([]string{}, svc.Subtypes...)
fabricui/ports.go:133: out.Services = append(out.Services, mdnsService{
fabricui/ports.go:147: out.Findings = append(out.Findings, mdnsFinding{
fabricui/ports.go:185: out = append(out, eventRecordOf(&evs[i]))
fabricui/spa.go:126: scripts = append(scripts, hashSource(m[2]))
fabricui/spa.go:129: styles := append(make([]string, 0, 2), "'self'")
fabricui/spa.go:131: styles = append(styles, hashSource(m[1]))
fabricui/memory_controller.go:60: out = append(out, Credential{Kind: CredentialWiFi, ID: id, Name: m.wifi[id].ssid, Default: m.defaults[Credential{Kind: CredentialWiFi, I
fabricui/memory_controller.go:63: out = append(out, Credential{Kind: CredentialThread, ID: id, Name: m.thread[id].name, Default: m.defaults[Credential{Kind: CredentialThr
fabricui/memory_controller.go:246: out = append(out, m.nodes[id])
fabricui/memory_controller.go:476: nodes = append(nodes, m.nodes[id])
fabricui/memory_controller.go:495: out.Sessions = append(out.Sessions, ControllerSession{
fabricui/memory_controller.go:503: out.Peers = append(out.Peers, p)
fabricui/memory_controller.go:597: u.queue = append(u.queue, PendingUpdate{NodeID: node, VendorID: vendorID, ProductID: productID, TargetSoftwareVersion: version, Status:
fabricui/memory_controller.go:840: d.Parts = append(d.Parts, o.Number)
fabricui/memory_controller.go:844: d.Clusters = append(d.Clusters, n.cluster(ep, c, false))
fabricui/memory_controller.go:877: out = append(out, e)
fabricui/memory_controller.go:900: ids = append(ids, p.Attribute)
fabricui/memory_controller.go:909: d.Attributes = append(d.Attributes, a)
fabricui/memory_controller.go:931: out = append(out, AttributeValue{Path: p, Value: v})
fabricui/memory_controller.go:933: out = append(out, AttributeValue{Path: p, Status: statusUnsupportedAttribute})
fabricui/memory_controller.go:984: ts = append(ts, map[string]any{"cluster": t.Cluster, "endpoint": t.Endpoint, "deviceType": t.DeviceType})
fabricui/memory_controller.go:988: out = append(out, m)
fabricui/memory_controller.go:1009: out = append(out, m)
fabricui/memory_controller.go:1034: kept = append(kept, f)
fabricui/memory_controller.go:1093: res.Addresses = append(res.Addresses, PingAddress{Address: a, Reachable: res.Reachable, RTT: res.RTT, Error: res.Error})
fabricui/memory_controller.go:1100: res.Addresses = append(res.Addresses, PingAddress{Address: a, Error: "client: address unreachable: " + a})
fabricui/memory_controller.go:1104: res.Addresses = append(res.Addresses, PingAddress{Address: a, Reachable: true, RTT: memoryPingRTT})
```

## 12. Function index of the two split files

```
fabricui/handlers_controller.go:98:func (h *handler) controllerRoutes() http.Handler {
fabricui/handlers_controller.go:120:func (h *handler) controllerGate(sub *http.ServeMux) http.Handler {
fabricui/handlers_controller.go:151:func (p *apiProblem) Error() string { return p.title + ": " + p.detail }
fabricui/handlers_controller.go:153:func badRequest(format string, args ...any) error {
fabricui/handlers_controller.go:157:func serviceUnready(title, detail string) error {
fabricui/handlers_controller.go:164:func (h *handler) controllerError(w http.ResponseWriter, r *http.Request, err error) {
fabricui/handlers_controller.go:202:func hex16(v uint64) string { return fmt.Sprintf("%016X", v) }
fabricui/handlers_controller.go:206:func hexID(v uint32) string { return fmt.Sprintf("0x%04X", v) }
fabricui/handlers_controller.go:208:func parseHex(s string, bits int) (uint64, error) {
fabricui/handlers_controller.go:220:func parseHex32(field, s string) (uint32, error) {
fabricui/handlers_controller.go:228:func parseNodeID(field, s string) (uint64, error) {
fabricui/handlers_controller.go:236:func endpointOf(field string, v int) (uint16, error) {
fabricui/handlers_controller.go:244:func timedOf(ms *int) (time.Duration, error) {
fabricui/handlers_controller.go:254:func (s *controllerServer) node(id string) (ControllerNode, error) {
fabricui/handlers_controller.go:266:func ptr[T any](v T) *T { return &v }
fabricui/handlers_controller.go:269:func optString(s string) *string {
fabricui/handlers_controller.go:276:func optTime(t time.Time) *time.Time {
fabricui/handlers_controller.go:284:func optUint[T ~uint8 | ~uint16 | ~uint32](v *T) *int {
fabricui/handlers_controller.go:291:func nonZero[T ~uint8 | ~uint16 | ~uint32](v T) *int {
fabricui/handlers_controller.go:300:func (s *controllerServer) GetControllerStatus(context.Context, api.GetControllerStatusRequestObject) (a
fabricui/handlers_controller.go:325:func (s *controllerServer) ListControllerNodes(context.Context, api.ListControllerNodesRequestObject) (a
fabricui/handlers_controller.go:333:func nodeSummary(n ControllerNode) api.ControllerNodeSummary {
fabricui/handlers_controller.go:352:func nonNil[T any](s []T) []T {
fabricui/handlers_controller.go:359:func nodeDetail(n ControllerNode) api.ControllerNodeDetail {
fabricui/handlers_controller.go:419:func (s *controllerServer) GetControllerNode(_ context.Context, req api.GetControllerNodeRequestObject) 
fabricui/handlers_controller.go:428:func endpointDetail(d EndpointDetail) api.EndpointDetail {
fabricui/handlers_controller.go:465:func clusterDetail(c ClusterDetail) api.ClusterDetail {
fabricui/handlers_controller.go:487:func elements(in []ClusterElement) []api.ClusterElement {
fabricui/handlers_controller.go:497:func (s *controllerServer) GetNodeEndpoint(_ context.Context, req api.GetNodeEndpointRequestObject) (api
fabricui/handlers_controller.go:515:func (s *controllerServer) GetNodeCluster(_ context.Context, req api.GetNodeClusterRequestObject) (api.G
fabricui/handlers_controller.go:538:func (s *controllerServer) ListNodeEvents(_ context.Context, req api.ListNodeEventsRequestObject) (api.L
fabricui/handlers_controller.go:550:func (s *controllerServer) DecommissionNode(ctx context.Context, req api.DecommissionNodeRequestObject) 
fabricui/handlers_controller.go:565:func (s *controllerServer) PingNode(ctx context.Context, req api.PingNodeRequestObject) (api.PingNodeRes
fabricui/handlers_controller.go:578:func pingResult(r *PingResult) api.PingResult {
fabricui/handlers_controller.go:597:func rttMS(d time.Duration) float32 { return float32(d.Microseconds()) / 1000 }
fabricui/handlers_controller.go:599:func (s *controllerServer) InterviewNode(ctx context.Context, req api.InterviewNodeRequestObject) (api.I
fabricui/handlers_controller.go:610:func (s *controllerServer) EnableNode(ctx context.Context, req api.EnableNodeRequestObject) (api.EnableN
fabricui/handlers_controller.go:621:func (s *controllerServer) DisableNode(ctx context.Context, req api.DisableNodeRequestObject) (api.Disab
fabricui/handlers_controller.go:632:func (s *controllerServer) SetNodeAutoSubscribe(ctx context.Context, req api.SetNodeAutoSubscribeRequest
fabricui/handlers_controller.go:648:func (s *controllerServer) OpenNodeWindow(ctx context.Context, req api.OpenNodeWindowRequestObject) (api
fabricui/handlers_controller.go:673:func (s *controllerServer) CloseNodeWindow(ctx context.Context, req api.CloseNodeWindowRequestObject) (a
fabricui/handlers_controller.go:686:func attributePathOf(field string, endpoint int, cluster, attribute string) (AttributePath, error) {
fabricui/handlers_controller.go:702:func (s *controllerServer) ReadNodeAttributes(ctx context.Context, req api.ReadNodeAttributesRequestObje
fabricui/handlers_controller.go:739:func writeResult(st WriteStatus) api.WriteResult {
fabricui/handlers_controller.go:747:func (s *controllerServer) WriteNodeAttribute(ctx context.Context, req api.WriteNodeAttributeRequestObje
fabricui/handlers_controller.go:771:func (s *controllerServer) InvokeNodeCommand(ctx context.Context, req api.InvokeNodeCommandRequestObject
fabricui/handlers_controller.go:819:func readOne(ctx context.Context, n ControllerNode, path AttributePath, filtered bool) (json.RawMessage,
fabricui/handlers_controller.go:846:func keyOf[K, V comparable](m map[K]V, v V) (K, bool) {
fabricui/handlers_controller.go:870:func (s *controllerServer) GetNodeACL(ctx context.Context, req api.GetNodeACLRequestObject) (api.GetNode
fabricui/handlers_controller.go:915:func (s *controllerServer) SetNodeACL(ctx context.Context, req api.SetNodeACLRequestObject) (api.SetNode
fabricui/handlers_controller.go:991:func (s *controllerServer) GetNodeBindings(ctx context.Context, req api.GetNodeBindingsRequestObject) (a
fabricui/handlers_controller.go:1039:func (s *controllerServer) SetNodeBindings(ctx context.Context, req api.SetNodeBindingsRequestObject) (
fabricui/handlers_controller.go:1099:func icdState(n ControllerNode) api.IcdState {
fabricui/handlers_controller.go:1118:func (s *controllerServer) GetNodeIcd(_ context.Context, req api.GetNodeIcdRequestObject) (api.GetNodeI
fabricui/handlers_controller.go:1126:func (s *controllerServer) RegisterNodeIcd(ctx context.Context, req api.RegisterNodeIcdRequestObject) (
fabricui/handlers_controller.go:1141:func (s *controllerServer) UnregisterNodeIcd(ctx context.Context, req api.UnregisterNodeIcdRequestObjec
fabricui/handlers_controller.go:1167:func (s *controllerServer) ownIndex(ctx context.Context, n ControllerNode) (uint8, error) {
fabricui/handlers_controller.go:1179:func (s *controllerServer) ListNodeFabrics(ctx context.Context, req api.ListNodeFabricsRequestObject) (
fabricui/handlers_controller.go:1207:func (s *controllerServer) RemoveNodeFabric(ctx context.Context, req api.RemoveNodeFabricRequestObject)
fabricui/handlers_controller.go:1241:func pairOptionsOf(b *api.CommissionRequest) PairOptions {
fabricui/handlers_controller.go:1259:func (s *controllerServer) CommissionNode(ctx context.Context, req api.CommissionNodeRequestObject) (ap
fabricui/handlers_controller.go:1320:func progressDetail(err error) string {
fabricui/handlers_controller.go:1338:func (s *controllerServer) DiscoverDevices(ctx context.Context, req api.DiscoverDevicesRequestObject) (
fabricui/handlers_controller.go:1382:func (s *controllerServer) credentials() (CredentialStore, error) {
fabricui/handlers_controller.go:1390:func (s *controllerServer) ListCredentials(ctx context.Context, _ api.ListCredentialsRequestObject) (ap
fabricui/handlers_controller.go:1412:func (s *controllerServer) SetCredential(ctx context.Context, req api.SetCredentialRequestObject) (api.
fabricui/handlers_controller.go:1452:func (s *controllerServer) RemoveCredential(ctx context.Context, req api.RemoveCredentialRequestObject)
fabricui/handlers_controller.go:1473:func (s *controllerServer) SetFabricLabel(ctx context.Context, req api.SetFabricLabelRequestObject) (ap
fabricui/handlers_controller.go:1488:func (s *controllerServer) ListUpdates(context.Context, api.ListUpdatesRequestObject) (api.ListUpdatesR
fabricui/handlers_controller.go:1511:func (s *controllerServer) updates() (SoftwareUpdateSource, error) {
fabricui/handlers_controller.go:1519:func (s *controllerServer) ConsentUpdate(ctx context.Context, req api.ConsentUpdateRequestObject) (api.
fabricui/handlers_controller.go:1545:func (s *controllerServer) CheckUpdates(ctx context.Context, _ api.CheckUpdatesRequestObject) (api.Chec
fabricui/handlers_controller.go:1570:func (s *controllerServer) GetControllerTopology(ctx context.Context, req api.GetControllerTopologyRequ
fabricui/handlers_controller.go:1582:func networkGraph(topo NetworkTopology) api.TopologyGraph {
fabricui/handlers_controller.go:1631:func (c controllerStream) VisitControllerStreamResponse(w http.ResponseWriter) error {
fabricui/handlers_controller.go:1636:func (s *controllerServer) ControllerStream(ctx context.Context, _ api.ControllerStreamRequestObject) (
fabricui/handlers_controller.go:1643:func (h *handler) subscribeController(fn func(Event)) (unsubscribe func()) {
fabricui/handlers_controller.go:1693:func pairingStepOf(s PairingStep) *pairingStep {
fabricui/handlers_controller.go:1698:func controllerEvent(e ControllerEvent) Event {
fabricui/handlers_controller.go:1729:func (s *controllerServer) GetControllerDiagnostics(context.Context, api.GetControllerDiagnosticsReques
fabricui/handlers_controller.go:1757:func wireTime(t time.Time) time.Time { return t.UTC().Truncate(time.Second) }
fabricui/handlers_controller.go:1760:func optWireTime(t time.Time) *time.Time {
fabricui/handlers_controller.go:1770:func controllerSessionOf(c *ControllerSession) api.Session {
fabricui/handlers_controller.go:1795:func controllerPeerOf(p *ControllerPeer) api.ControllerPeer {
fabricui/handlers_controller.go:1826:func (s *controllerServer) GetControllerMdns(context.Context, api.GetControllerMdnsRequestObject) (api.
fabricui/handlers_controller.go:1843:func (s *controllerServer) ListControllerDiagnosticEvents(context.Context, api.ListControllerDiagnostic
fabricui/handlers_controller.go:1866:func controllerTraceEvent(e DiagnosticEvent) Event {
fabricui/handlers_controller.go:1874:func (s *controllerServer) diagnosticsServed() bool {
fabricui/handlers_node.go:116:func summariseSurface(r *surface.Report) *surfaceSummary {
fabricui/handlers_node.go:137:func (h *handler) status(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:238:func (h *handler) fabrics(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:270:func (h *handler) removeFabric(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:293:func (h *handler) maintenanceReassemble(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:317:func (h *handler) factoryReset(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:364:func (h *handler) eraseNode(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:394:func hex64(v uint64) string { return fmt.Sprintf("%016X", v) }
fabricui/handlers_node.go:404:func eventWindow(sinceQ, limitQ *int) (since uint64, limit int, err error) {
fabricui/handlers_node.go:422:func eventRecords(in []EventRecord) api.EventRecordList {
fabricui/handlers_node.go:442:func queryInt(r *http.Request, name string) (*int, error) {
fabricui/handlers_node.go:455:func pathEndpoint(r *http.Request) (uint16, error) {
fabricui/handlers_node.go:465:func (h *handler) writeAPIError(w http.ResponseWriter, r *http.Request, title string, err error) {
fabricui/handlers_node.go:484:func (h *handler) nodeEndpoint(w http.ResponseWriter, r *http.Request) {
fabricui/handlers_node.go:512:func surfaceKeyOf(topo node.BridgeTopology, ep uint16) (key string, ok bool) {
fabricui/handlers_node.go:528:func (h *handler) decisionsOf(key string, ok bool) ([]api.SurfaceDecision, error) {
fabricui/handlers_node.go:554:func (h *handler) nodeEndpointEvents(w http.ResponseWriter, r *http.Request) {
```
