# Konzept U4 — Gerätedetails (Milestone 9 „View“, Phase U4)

Stand 2026-11-03. Pins: go-fabric-ui main 2d326bdc (v0.4.1, Vertrag 0.3.1),
go-fabric main 53b0d93d, go-fabric-server main 78a3ebb, matterjs-server
0bd06f22. Faktenkarte `~/.cache/go-fabric-scratch/read-u4.md` (Sonnet); die
tragenden Zeilen sind geprüft und hier zitiert. Roadmap-Zeile: go-fabric
`docs/roadmap.md:377` — „device details: per endpoint the device type with
revision, clusters with revision and feature map, the attribute inventory
with live values, events, parts, tags, reachability, the host's surface
decision; on the controller side an inventory tab, an event log, decoded
feature maps“.

## A. Ausgangslage

- **Node-Familie.** `Endpoints.svelte` listet Endpoints mit Gerätetyp (ohne
  Revision, `:92`), Reachable-Badge (`:86-88`), „part of“ (`:93-95`) und
  Cluster als Namens-Chips (`:97-105`): keine Revision, keine FeatureMap,
  keine Attribute, keine Events, keine Tags/Labels/Scope. Root- und
  Aggregator-Endpoint fehlen (`fabricui/handlers_diagnostics.go:172-198`
  `flattenTopology`). Die einzige Stelle mit FeatureMap und Attribut-Ids
  ist der Surface-Report (`SurfaceCluster{id, revision, featureMap,
  attributes[], accepted[], generated[], events[]}`, `contract/openapi.yaml:1454`),
  den `Surface.svelte` als Hex-Listen ohne Namen zeigt (`:152-160`).
  Der Stream kennt kein Attribut-Ereignis (`fabricui/events.go:16-28`);
  `node.Observers.OnAttributeChanged` (`node/observers.go:116`) ist nicht
  abgebildet. `BridgedDevice.Scope/Tags/Labels` (`node/lifecycle.go:128-137`)
  werden in `fabricui/` nirgends benutzt.
- **Controller-Familie.** `Node.svelte` hat Tabs (`:28`); `node/Endpoints.svelte`
  zeigt Gerätetypen mit Revision (`:35-37`) und Cluster-Badges, deren
  Revision und Features nur im Tooltip stehen (`:42`). `node/Attributes.svelte`
  ist ein gezielter Leser über die Leitung (`client.readAttributes`, `:45`),
  nach Entscheidung 0.4 aus U2 (`concept-u2-controller-family.md:170`,
  `contract.md:255-258` „no wildcard, no inventory“). Dabei liegt das
  Inventar bereits im Image des Controllers: `Cluster.Revision() :235`,
  `FeatureMap() :250`, `Features() :260`, `AttributeList() :270`,
  `AcceptedCommands :286`, `GeneratedCommands :294`, `Version() :364`,
  `Value() :374`, `Decoded() :398` (`controller/client/cluster.go`);
  `Endpoint.Conformance()` (`controller/client/conformance.go:93`).
  Matter-Events erreichen die Seiten nicht: `client.Hooks.Event`
  (`controller/client/config.go:123`) wird in der Hook-Kette der Seiten
  nicht verkettet (`fabricui/gofabric/hooks.go:199` `Event: a.Event`;
  go-fabric-server `internal/daemon/run.go:245-247`), und es gibt keinen
  Speicher dafür — weder im Client noch im Adapter; go-fabric-server hält
  eine globale Historie der letzten 25 (`controllerapi/events.go:15`).
- **Importregel.** Produktions-Go in go-fabric-ui importiert aus go-fabric nur
  `node`, `eligibility`, `contract`, `clock`, `surface`, `schema`
  (`internal/arch/arch_test.go:29-42`); der Adapter `fabricui/gofabric`
  zusätzlich `controller*`, `cluster/spec*` (`:44-70`). Ein Node-Handler
  kann also weder Werte über `Bridge()` lesen noch Namen aus `cluster/spec`
  holen. `schema` hat `ClusterName` (`schema/lookup.go:20`),
  `DeviceTypeName` (`:53`) und `ClusterFeatureNames(clusterID, featureMap)`
  (`schema/devicetype_requirements.go:338`), aber keine Attribut-,
  Command- oder Event-Namen.
- **Vergleich matterjs-server** (`packages/dashboard/src/pages/`): Drill-down
  Node → Endpoint → Cluster (`matter-dashboard-app.ts:208-229`), Rohwerte
  live über `attribute_updated`, decodierte Feature-Chips
  (`matter-cluster-view.ts:432-460`), Tag-Chips (`matter-endpoint-view.ts:89,138`);
  Namen aus einer generierten `descriptions.ts` im Bundle. Kein
  Ereignislog, keine Cluster-Revision, keine DataVersion, keine
  Konformität, keine Surface-Entscheidungen.

## B. Ziel

Zwei Detailseiten, eine je Familie, aus einem Vertragstyp: pro Endpoint die
Gerätetypen mit Revision, die Cluster mit Revision, FeatureMap (hex) und
decodierten Feature-Bits, das Attributinventar mit Namen und — wo die
Fassade es hergibt — Werten und DataVersion, Accepted/Generated Commands
und Event-Ids mit Namen, Parts, Tags, Labels, Scope, Erreichbarkeit; auf
der Node-Seite daneben die Surface-Entscheidungen dieses Endpoints, auf der
Controller-Seite die Konformitätsverdikte. Dazu ein Ereignislog je Node
(beide Familien), sobald go-fabric den Ring hält (D.2).

## C. Form

### C.1 Vertrag 0.4.0 (additiv über 0.3.1)

Ein gemeinsamer Typ für beide Familien (Präzedenz U3,
`concept-u3-topology.md:177-182`):

- `ClusterDetail{id, id_hex, name?, revision, feature_map_hex, features[]
  {code, title?, bit}, data_version?, attributes[]{id, id_hex, name?,
  value?, status?}, accepted[]{id, id_hex, name?}, generated[]{…},
  events[]{id, id_hex, name?}}` — `value` als JSON nach Feldnamen wie
  `AttributeValue.value` heute (`contract.md:262-264`); `name` optional,
  weil 16 Cluster keine generierte Definition haben und die Node-Seite die
  Namen erst nach G-N2 kennt.
- `EndpointDetail{endpoint_id, parent_endpoint_id?, name, reachable,
  device_types[]{id, id_hex, name?, revision}, scope?, tags[]{namespace,
  tag, mfg_code?, label?}, labels[]{label, value, user}, parts[] (Nummern),
  clusters[] ClusterDetail, decisions[] SurfaceDecision (Node-Seite),
  conformance? {status, violations[], unjudged[]} (Controller-Seite)}`.
- Node-Routen: `GET /node/endpoints/{n}` → `EndpointDetail` (auch `0` und
  der Aggregator, aus dem Surface-Report gejoint); Stream-Ereignis
  `attribute_changed {endpoint, cluster, attributes[]}` aus
  `OnAttributeChanged` (Ids, keine Werte). `GET /node/endpoints/{n}/events`
  → `EventRecord[]` folgt mit G-N3 (U4a).
- Controller-Routen: `GET /controller/nodes/{id}/endpoints/{ep}` →
  `EndpointDetail` ohne Attributwerte (Struktur), `GET
  /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}` →
  `ClusterDetail` mit Werten **aus dem Cache** (Entscheidung F.0.1: kein
  Wire-Read, kein Admin; die gezielten Reads nach 0.4 bleiben wie sie
  sind), `GET /controller/nodes/{id}/events` → `EventRecord[]` (U4a).
- `EventRecord{number, priority, endpoint, cluster_hex, event_hex, name?,
  epoch_ms?, data?}` — eine Form für `im.EventRecord` (`im/eventlog.go:19-45`)
  und `client.Hooks.Event`.
- Vertragsregeln wie bisher: nur Ergänzungen, keine
  `additionalProperties:false` auf decodierten Antworten, Fixtures auf
  beiden Seiten (`fixtures.test.ts:1-29`, `fabricui/misc_test.go:308-311`).
  `contract/previous/openapi.yaml` wird mit 0.4.0 nachgezogen.

### C.2 Go-Hälfte

- Node-Handler (`fabricui/handlers_*.go`, nur Fassade): `EndpointDetail` aus
  `Node.Topology()` (`BridgedDevice` mit `Scope/Tags/Labels/Parts`), dem
  Surface-Report (Revision, FeatureMap, Attribut-/Command-/Event-Ids auch
  für Root und Aggregator; Entscheidungen dieses Endpoints) und `schema`
  (`ClusterName`, `DeviceTypeName`, `ClusterFeatureNames`). Werte und Namen
  der Elemente kommen erst mit der Modul-Welle (G-N1/G-N2) — bis dahin
  `value` und `name` leer, die Seite zeigt Ids in Hex.
- Controller-Adapter (`fabricui/gofabric/node.go`): `ClusterDetail` aus dem
  Image (`Cluster.Definition()` für Namen und `Feature.Title`
  (`cluster/spec/model.go:332`), `Value()` + `client.DecodeJSON(…,
  client.NameBased)` wie `Read` heute `:158`, `Version()`), DeviceTypeList
  und TagList aus Descriptor (`Cluster(0x1D).Decoded(0/4)`), Reachable aus
  BridgedDeviceBasicInformation (0x39/0x11), `Endpoint.Conformance()`.
  Port `ControllerNode` bekommt `EndpointDetail(ep)` und
  `ClusterDetail(ep, cluster)`; `MemoryController` wächst mit.
- Ereignislog: beide Familien lesen einen Ring in go-fabric (F.0.3):
  Controller `controller/client` hält je Node einen begrenzten Ring aus
  `Hooks.Event`, Node-Seite eine Fassadenmethode über `EventLog().Query`
  (G-N3). Bis die Modul-Welle landet, gibt es keinen Ereignis-Tab.
- Live-Werte (F.0.4): kein neuer Hook. Controller-Seite lädt
  `ClusterDetail` bei `node_updated` des Nodes neu (Cache, keine Leitung;
  Store `RELOAD_DELAY` 250 ms, Hook 500 ms); Node-Seite lädt bei
  `attribute_changed` ihres Endpoints den betroffenen Cluster neu.

### C.3 Svelte-Hälfte

- Node-Seite: neue Seite `#/node/endpoint/<n>` (Routes-Key `endpoint`,
  `refHref` für Topology-Refs `{page:"endpoints", id}` → Detail,
  `context.ts:58-69`); `Endpoints.svelte` bekommt Revision je Gerätetyp und
  den Link zur Detailseite, dazu die Zeilen für Endpoint 0 und Aggregator.
- Controller-Seite: `Node.svelte` erhält die Tabs **Inventar** (Endpoint →
  Cluster lazy, ein `ClusterDetail` je Aufruf) und **Ereignisse** (U4a);
  der Tab **Attributes** bleibt als roher gezielter Leser mit Admin-Write
  (F.1.6).
- Gemeinsame Komponenten in `ui/src/lib/kit/`: `ClusterView` (Kopf mit
  Revision, FeatureMap hex + Feature-Chips `code · title`), `AttributeTable`
  (Id hex, Name, Wert als formatiertes JSON, DataVersion), `TagChips`,
  `EventList`. i18n-Schlüssel in `en.json` und `de.json`, Coverage-Zeilen
  für neue Server-Enums (`coverage.test.ts:46-61`).
- Tests: vitest + goldene Fixtures (`node-endpoint.200`, `node-endpoint.404`,
  `controller-node-endpoint.200`, `controller-node-cluster.200`,
  `controller-node-cluster.404`, später `*-events.200`); Playwright bleibt
  U5.

## D. Wellen

- **D.1 U4b — Controller-Seite zuerst** (Opus, go-fabric-ui, Adapter und
  Seiten): Vertrag 0.4.0 mit `EndpointDetail`/`ClusterDetail`, Inventar-Tab
  aus dem Cache, decodierte FeatureMaps, DeviceTypeList, Tags, Reachable je
  Bridged Endpoint, Konformität; ohne Ereignis-Tab. go-fabric-ui 0.5.0
  (ungetaggt bis zur Freigabe).
- **D.2 Modul-Welle „M9 U4-module“** (Opus in go-fabric, nach Phase 31,
  go-fabric-Issue aus diesem Konzept): G-N1 Lesen servierter Attributwerte
  über die Fassade, G-N2 Attribut-/Command-/Event-Namen in `schema`
  (generiert aus dem Snapshot), G-N3 Fassadenliste emittierter Events und
  `Observers.OnEvent`, der Ereignis-Ring je Node im Controller-Client; dazu
  die Korrektur der Roadmap-Versionen (`docs/roadmap.md:105,377`: Vertrag
  0.3.0 ist vergeben, U4 = 0.4.0, go-fabric-ui 0.5.0).
- **D.3 U4a — Node-Seite und Ereignislog beider Familien** (Opus,
  go-fabric-ui, nach D.2): `GET /node/endpoints/{n}` mit Werten und Namen,
  `attribute_changed`, Ereignis-Routen beider Familien, Ereignis-Tab.
  Vertrag 0.4.x additiv.

## E. Besser als andere, besser als vorher, Schnitt richtig?

- **Besser als andere.** matterjs-servers Dashboard zeigt Rohwerte und
  Feature-Chips; hier kommen Cluster-Revision, DataVersion, Namen der
  Commands und Events, Erreichbarkeit je Bridged Endpoint, die
  Surface-Entscheidung des Hosts und die Konformität des Controllers hinzu,
  und ein Ereignislog, das matterjs-server nicht hat. Die Namen liegen
  serverseitig in Go (`schema`, `cluster/spec`), nicht als generierte Tabelle
  im Bundle.
- **Besser als vorher.** Die Endpoint-Liste zeigte weder Revisionen noch
  Features; der Attribut-Leser las über die Leitung, was der Controller
  längst im Cache hielt; Scope, Tags, Labels und Attributänderungen der
  Fassade lagen brach.
- **Schnitt.** Ein Vertragstyp für beide Familien, gebaut in Go aus
  Fassadenaufrufen; der Controller liest nur den Cache (0.4 bleibt für die
  Leitung); Werte und Namen der Node-Seite kommen aus der Fassade, nicht
  über einen Seam — die drei Lücken sind go-fabric-Issues. Der Ereignis-Ring
  liegt in go-fabric, damit UI und go-fabric-server dieselbe Quelle lesen.
- **Nicht gemacht:** keine Stream-Werte (Neuladen aus dem Cache reicht),
  kein Write aus dem Inventar (Admin-Write bleibt beim Attributes-Tab), kein
  Playwright (U5), keine Tag-Namenstabelle (Namespace/Tag als Zahl und
  Label aus dem Attribut; eine Namenstabelle ist ein späteres
  go-fabric-Thema), Tags der Node-Seite bleiben leer, solange der
  Assembler TagList nicht einschaltet (`node/lifecycle.go:129-133`; die
  Seite sagt das).

## F. Entscheidungen

Owner, per Wizard 2026-11-03:

- **0.1 Inventar aus dem Cache:** 0.4 wird ergänzt — ein Read-only-Inventar
  aus dem Image des Controllers ist erlaubt (kein Wire-Read, kein Admin);
  Wire-Reads bleiben gezielt.
- **0.2 Reihenfolge:** U4b (Controller) zuerst; die Modul-Welle für
  G-N1–G-N3 parallel nach Phase 31; dann U4a.
- **0.3 Ereignislog:** begrenzter Ring je Node in go-fabric
  (Controller-Client), Node-Seite über eine Fassadenmethode; UI und
  go-fabric-server lesen dieselbe Quelle.
- **0.4 Live-Werte:** Neuladen aus dem Cache bei `node_updated`; Node-Seite
  `attribute_changed` mit Ids; keine Stream-Werte.

Konzept (Fable), aus den Owner-Entscheidungen abgeleitet:

- **1.1** Node-Seite als eigene Detailseite `#/node/endpoint/<n>`, nicht als
  aufklappbare Liste („the two detail pages“).
- **1.2** Werte roh als JSON nach Feldnamen (wie `AttributeValue.value`);
  Decodierung von Enums/Bitmaps zu Namen ist nicht Teil von U4.
- **1.3** Feature-Bits als `code · title · bit` plus FeatureMap hex, eine
  Komponente für beide Familien.
- **1.4** Ein Vertragstyp `ClusterDetail`/`EndpointDetail` für beide
  Familien; die bestehenden `Cluster`/`NodeCluster` bleiben unverändert.
- **1.5** Versionen: Vertrag 0.4.0, go-fabric-ui 0.5.0; Roadmap-Korrektur in
  go-fabric.
- **1.6** Attributes-Tab bleibt (roher Leser + Admin-Write); kein Write aus
  dem Inventar.
- **1.7** Leere Tags auf der Node-Seite werden als leerer Zustand mit Grund
  gezeigt, Scope und Labels immer.
