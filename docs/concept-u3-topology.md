# Konzept U3 — die Topologie-Sicht von go-fabric-ui

Stand 2026-11-03. Phase U3 von go-fabrics Meilenstein 9 („Sicht“, Konzept
v13 §C.2; Owner-Entscheidung 2026-11-03: M9 läuft parallel zu M8, weil es
technisch von M8/M10 nicht abhängt). Faktenbasis: die Lesekarten
`read-u3-ui-server.md` und `read-u3-facade.md` (go-fabric `main` a6ce6272
= `v0.16.0` plus Phase 29, go-fabric-ui `v0.3.0`, go-fabric-server
`v0.4.9`); jede tragende Zeile unten ist am Stand nachgelesen. Sprache:
Deutsch (Phasenkonzept); Brief, Vertrag, Doku und Bericht auf Englisch.

## A. Ausgangslage

- **Die Daten sind da, die Seiten zeigen sie nicht.** Die Node-Familie
  zeigt Endpoints als flache Tabelle mit `parent_endpoint_id`
  (`contract/openapi.yaml:1012-1029`, `fabricui/handlers_diagnostics.go:167`
  `flattenTopology`), Sessions als Tabelle mit einem Subscription-Zähler je
  Session (`openapi.yaml:1037-1052`), Fabrics als Liste (`:959`). Wer mit wem
  verbunden ist — Fabric → Session → Node, Aggregator → Gerät → Teile — setzt
  der Betreiber im Kopf zusammen.
- **go-fabric liefert alles für die Node-Familie:** `Topology()` mit
  `BridgedDevice{Endpoint, DeviceType, DeviceTypeRevision, Clusters, Name,
  Reachable, Parts}` (`node/lifecycle.go:89-132`), `FabricInfos`
  (`node/node.go:1057`), `SessionDiagnostics{Sessions []SessionInfo,
  Occupancy}` mit Fabric-Index, Peer-Node-Id, PASE/CASE, Subscription-Zahl
  und Leerlaufzeiten (`node/inspect.go:61-108`), dazu die Observer
  `OnSessionOpened/OnSessionClosed/OnSubscriptionsChanged`
  (`node/observers.go:91-100`), die der Vertrag heute nicht weiterreicht
  (`/node/stream` kennt nur fabric/window/endpoints/exposure/surface,
  `openapi.yaml:297-312`). `NodeSource` (`fabricui/options.go:27-51`)
  hat `Topology()`, `SessionDiagnostics()` und `Observe(node.Observers)`
  bereits.
- **Die Controller-Familie hat den Netzgraphen nur im Server-WS:**
  `get_network_topology` (`controllerapi/cmd_topology.go:106`) baut aus
  NetworkCommissioning-, GeneralDiagnostics-, Thread- und Wi-Fi-Diagnostik der
  verbundenen Knoten einen Graphen `{nodes[], connections[]}` mit Rolle,
  RLOC16, Ext-Address, SSID/BSSID, LQI/RSSI (`cmd_topology.go:52-91`,
  `buildTopology :166`, nach matter.js
  `NetworkTopologyService.ts:304-430`), und drückt
  `network_topology_updated` (`:559`). Weder der Vertrag noch
  `ControllerSource` (`fabricui/controller.go:22-56`) kennen ihn. go-fabric
  hat die typisierten Zustände dafür (`cluster/spec/threadnetworkdiagnostics`
  `State.RoutingRole/NeighborTable/Rloc16`, `server_gen.go:22,28,85`;
  `cluster/spec/wifinetworkdiagnostics` `State.Bssid/Rssi`, `:21,25`).
- **Kein Graph, kein SVG, kein e2e in der UI.** `<svg` gibt es nur für den
  QR-Code (`ui/src/lib/pages/node/Pair.svelte`), Tests sind vitest auf
  happy-dom mit den Vertrags-Fixtures, `make e2e` ist der Go-Test gegen
  `fabrictest`. Playwright ist U5.
- **Versionsstand.** go-fabric-ui ist `v0.3.0` (go-fabric `v0.16.0`,
  2026-11-02, nur Abhängigkeit); die Roadmap-Zeile „go-fabric-ui 0.3.0,
  Vertrag 0.3.0“ ist damit überholt: M9 wird **go-fabric-ui 0.4.0 mit
  Vertrag 0.3.0**. go-fabric-server hängt noch an go-fabric-ui `v0.2.9` und
  go-fabric `v0.15.6` (`go.mod:6-7`).

## B. Ziel

Eine Seite je Familie, die die Zusammenhänge zeichnet statt sie zu listen:

- **Node-Familie, Seite „Topology“:** links der Node mit Aggregator →
  Bridged Devices → Parts (Erreichbarkeit, Gerätetyp, Name); rechts die
  Fabrics mit ihren Sessions (PASE/CASE, Peer-Node, Subscriptions, Leerlauf)
  als zweite Ebene; Sessions hängen am Node. Live: Session auf/zu und
  Subscription-Wechsel kommen über SSE, nicht per Reload.
- **Controller-Familie, Tab „Network“ in der Knotenliste:** der Netzgraph
  (Thread-Router/-Endgeräte mit Rolle und RLOC16, LQI/RSSI je Kante; Wi-Fi-AP
  mit SSID/BSSID und RSSI je Knoten; Ethernet), aus demselben Vertrag, in
  jeder Einbettung, nicht nur im Daemon.
- **Ein Graph-Schema, ein Renderer.** Beide Seiten liefern denselben
  Vertragstyp `TopologyGraph` und zeichnen ihn mit einer Kit-Komponente in
  eigenem SVG (Owner-Entscheidung v13 E.4: keine Graph-Bibliothek).

## C. Form

### C.1 Vertrag 0.3.0 (additiv)

| Route | Methode | Quelle |
| --- | --- | --- |
| `/node/topology` | GET | `Topology()`, `FabricInfos`, `SessionDiagnostics` → `TopologyGraph` |
| `/controller/topology` | GET, `?refresh=true` (Admin) | `ControllerSource.NetworkTopology(ctx, refresh)` → `TopologyGraph` |
| `/node/stream` | + Events `session_opened`, `session_closed`, `subscriptions_changed` (`SessionEvent`) | `node.Observers` |
| `/controller/stream` | + Event `topology_changed` (`{}`) | `ControllerSource.Observe` |

`TopologyGraph{as_of, family: node|controller, nodes[], edges[]}`.
`TopologyNode{id, kind, label, state, ref?, details[]}` — `kind` aus
`node | aggregator | device | part | fabric | session | peer | thread_router
| thread_end_device | wifi_ap | ethernet`; `state` aus `ok | unreachable |
idle | unknown`; `ref` ein Deep-Link `{page, id}` auf die bestehende Seite
(Endpoint-Tabelle, Fabric, Session, Controller-Knoten); `details` eine
geordnete Liste `{key, value}` (Gerätetyp mit Revision, RLOC16, SSID …), damit
der Renderer kein Wissen je `kind` braucht und neue Angaben ohne
Schemaänderung ankommen. Je `kind` zusätzlich die typisierten Felder, die eine
Seite verlinkt oder färbt (`endpoint`, `fabric_index`, `session_id`,
`node_id_hex`, `reachable`, `is_pase`, `subscriptions`, `role`, `rloc16`,
`ssid`, `bssid`). `TopologyEdge{source, target, kind: contains | session |
link, strength?: strong|medium|weak|none|unknown, lqi?, rssi?, label?}`
(`none` ab Vertrag 0.3.1, U3b). 64-Bit-Ids
nur als Hex-Strings; Knoten-Ids im Graphen sind Strings (`ep:12`, `fabric:1`,
`session:42`, `peer:000000000001B669`).

Der Vertrag bleibt 0.2.0-kompatibel (nichts entfernt, `contract/previous`
= 0.2.0, Guard grün); Fixtures `node-topology.200`, `node-topology.503`,
`controller-topology.200`, `controller-topology.403`, aufgezeichnet und
gelesen. `make gen-types`, `make generate` (Controller-Stubs).

### C.2 Go-Hälfte

- **Node:** `handlers_topology.go` baut den Graphen aus den drei
  `NodeSource`-Aufrufen; kein neuer Port. Sessions werden über
  `SessionInfo.FabricIndex` an `FabricInfo.FabricIndex` gehängt; ein Peer
  ohne Fabric (PASE) hängt am Node. `events.go` bekommt die drei
  Session-Events; der Stream-Hub speist sie aus `Observe(node.Observers{
  OnSessionOpened, OnSessionClosed, OnSubscriptionsChanged})`.
- **Controller:** `ControllerSource` bekommt `NetworkTopology(ctx, refresh
  bool) (NetworkTopology, error)` mit eigenen kleinen Typen (wie heute: kein
  go-fabric-Typ im Port) und `ControllerEvent{Type: topology_changed}`.
  **Der Builder liegt in go-fabric** (`controller/topology`, Owner-
  Entscheidung F.2): er liest je verbundenem Knoten die Thread-/Wi-Fi-
  Diagnostik und NetworkCommissioning über `StateFrom(ep.Cluster(id))`,
  ordnet Nachbarn Knoten zu (Ext-Address, RLOC16) und stuft die Stärke, nach
  matter.js `NetworkTopologyService.ts:304-430, 485-500` — das Matter-Wissen
  bleibt im Modul, mit Zitat und Fable-Review. Der Adapter
  `fabricui/gofabric` ruft ihn und mappt in die Port-Typen; der Server
  ersetzt in seiner Folgewelle `buildTopology` (`cmd_topology.go:166`) durch
  denselben Aufruf. Bis die go-fabric-Welle gelandet ist, steht
  `/controller/topology` im Vertrag mit Schema und `x-status: planned`, 501
  (das U2-Muster für ungefüllte Routen).
- **Import-Grenze** wie bisher: `fabricui` nur Facade; `fabricui/gofabric`
  darf `controller`, `controller/client` und die generierten
  Cluster-Definitionen (die beiden Diagnostik-Pakete kommen dazu, Guard
  `internal/arch`).

### C.3 Svelte-Hälfte

- `ui/src/lib/kit/Graph.svelte` + `graph-layout.ts`: geschichtetes Layout
  (Spalten nach Tiefe, Zeilen nach Reihenfolge, Kanten als kubische Pfade),
  reine Funktion → Golden-Test der Positionen; Pan/Zoom über `viewBox`;
  Farben und Schrift aus `theme/tokens.css` (`--ha-primary-color`,
  `--ha-error-color`, `--ha-divider-color`), Knotenform je `kind`, Zustand
  als Rand; Auswahl zeigt `details[]` in einer Seitenleiste mit dem
  Deep-Link `ref`; Tastatur erreichbar (Tab über Knoten, Enter folgt `ref`).
- `pages/node/Topology.svelte`: lädt `/node/topology`, hört auf die neuen
  SSE-Events plus `endpoints_changed`/`fabric_*` und lädt neu (entprellt wie
  `store.svelte.ts`); Leer-, Lade- und Fehlerzustand aus dem Kit.
- `pages/controller/Nodes.svelte` bekommt den Tab „Network“ mit
  `/controller/topology`, Admin-Knopf „Refresh“, SSE `topology_changed`.
- Shell: Node-Familie bekommt den Menüpunkt „Topology“, Routen
  `#/node/topology` und `#/controller/nodes?tab=network`. i18n en/de.
- Tests: vitest happy-dom je Seite auf den Fixtures, Layout-Golden,
  Graph-Komponente mit einem Fixture-Graphen; Go-Handler-Tests mit denselben
  Fixtures; `make e2e` (fabrictest-Site) prüft `/node/topology` mit einer
  kommissionierten Session.

## D. Wellen

- **U3a → go-fabric-ui `v0.4.0`** (jetzt, UI-Repo): Vertrag 0.3.0 vollständig
  (beide Routen mit Schema, `/controller/topology` noch `x-status: planned`
  501, die vier SSE-Events), `/node/topology` mit Handler und Session-Events,
  Port-Methode `NetworkTopology` und `topology_changed` im Port, Kit-Graph,
  Seite „Topology“, Doku. Unabhängig von go-fabric.
- **M9-Modulwelle in go-fabric** (eine kleine Opus-Welle in der laufenden
  Pipeline, nach dem nächsten Merge): `controller/topology` nach
  `NetworkTopologyService.ts` mit chip/matter.js-Zitaten, plus die
  U4-Fassadenlücken aus F.4 (ein Issue, eine Welle).
- **U3b → go-fabric-ui `v0.4.1`**: `/controller/topology` gefüllt
  (`ControllerOf` ruft `controller/topology`), Tab „Network“, Fixture 200.
  *Gelandet* (Brief U3b, `docs/reports/U3b.md`): Vertrag 0.3.1,
  `ControllerOptions.Topology`, go-fabric `c91e5b7f`.
- **S-Welle im Server** danach: go.mod auf go-fabric und go-fabric-ui
  `v0.4.1`, `controllerSource.NetworkTopology`, `get_network_topology` als
  Aufrufer von `controller/topology`, `web.go:21` Kommentar.
- U4 (Gerätedetails) und U5 (Betrieb, Playwright) bekommen je ein eigenes
  Konzept.

## E. Besser als andere, besser als vorher, Schnitt richtig?

- **Besser als andere.** Home Assistants Matter-Integration und
  python-matter-server zeigen keine Fabric-/Session-Sicht und keinen
  Netzgraphen; matter.js' `ws-controller` liefert den Netzgraphen als
  Daten, ohne Oberfläche. Hier: ein Vertragstyp für beide Familien, Fixtures
  auf beiden Seiten, eine Renderer-Komponente ohne Bibliothek (kein
  d3/dagre/elk im Bundle, keine Lizenz- und Größenfrage).
- **Besser als vorher.** Die Diagnose-Seite listet Sessions, die
  Endpoint-Seite listet Endpoints; wer an welcher Fabric hängt und was ein
  Gerät umfasst, war nicht sichtbar. Der Netzgraph des Servers wird vom
  WS-Nebenprodukt zum Vertragsbestandteil jeder Einbettung.
- **Schnitt.** Der Graph wird serverseitig gebaut (Go, aus Facade-Aufrufen),
  nicht clientseitig aus drei Listen: so ist die Form fixiert und getestet,
  Partition/Scope oder Tags kommen später als weitere `details` ohne
  Seitenänderung, und der Server kann denselben Graphen über WS geben. Der
  Builder der Controller-Seite liegt in go-fabric (`controller/topology`):
  die Zuordnung von Nachbartabellen zu Knoten und die Stärkestufen sind
  Matter-Wissen nach matter.js, das nach der Modellregel ins Modul gehört und
  dort unter Fable-Review zitiert wird; UI-Adapter und Server sind zwei
  Aufrufer einer Implementierung. Subscriptions bleiben in U3 Zähler je Session,
  weil die Fassade keine Aufzählung hat (`im/subscription/manager.go:846`
  `snapshot()` unexportiert, `CountBySession :864`); die Aufzählung mit
  Pfaden ist ein go-fabric-Issue für U4, kein Umweg über einen Seam.
- **Nicht gemacht:** keine Graph-Bibliothek (E.4), kein Attribut-Browser
  (U4), kein Playwright (U5), keine Live-Attributwerte (U4), keine
  Thread-Border-Router-Registry und keine Routentabellen im Netzgraphen (der
  Server lässt sie heute auch aus, `cmd_topology.go:28-31`).

## F. Entscheidungen (Owner, 2026-11-03, per Wizard)

- **0.1 Graph-Form:** ein Vertragstyp `TopologyGraph` für beide Familien,
  serverseitig gebaut (C.1), nicht clientseitig aus den bestehenden Listen.
- **0.2 Builder der Controller-Topologie:** in go-fabric
  `controller/topology` (nicht im Adapter): Matter-Wissen gehört ins Modul;
  eine kleine go-fabric-Welle in der laufenden Pipeline; U3 wird in U3a
  (Node, jetzt) und U3b (Controller, nach der Modulwelle) geteilt.
- **0.3 Version:** U3 = go-fabric-ui `v0.4.0`, Vertrag 0.3.0; Roadmap-Zeile
  M9 wird angepasst.
- **0.4 go-fabric-Issue jetzt:** Subscription-Aufzählung mit Pfaden,
  Partition/Scope und TagList am `BridgedDevice`, node-weiter
  Attribut-Observer — ein Issue mit Zitaten, umgesetzt in der M9-Modulwelle
  zusammen mit `controller/topology`, vor dem Konzept U4.
