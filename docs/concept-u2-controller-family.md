# Konzept U2 — die Controller-Familie von go-fabric-ui

Stand 2026-10-26. Fortsetzung von go-fabric Konzept v9 (Welle U2, dort nur
als Routen-Platzhalter angelegt: `/controller/commission`, `/controller/discovery`,
`/controller/nodes`, `/controller/nodes/{id}/…`, `/controller/fabric/label`;
`Options.Controller ControllerSource`). Faktenbasis: die Lesekarte
`read-u2-controller-family.md` (go-fabric `v0.13.3`, go-fabric-ui `v0.1.3`,
go-fabric-server `v0.3.2`). Sprache: Deutsch (Phasenkonzept); Briefs,
Issues und Doku auf Englisch.

## A. Ausgangslage

- go-fabric-ui `v0.1.3` hat die Node-Familie vollständig (17 Routen, sechs
  Seiten, SSE), die fünf Controller-Pfade stehen im Vertrag als
  `x-status: planned` mit 501 und ohne Schemas (`contract/openapi.yaml:292-332`).
  Es gibt kein `Options.Controller`, keinen `ControllerSource`, kein
  `ui/src/lib/pages/controller/`; die Shell kennt nur die sechs Node-Seiten.
- go-fabric-server `v0.3.2` spricht die Controller-Hälfte über das
  matterjs-server-WebSocket-Protokoll (Schema 14, rund 41 Kommandos, alle über
  `controller.Controller`, `client.Registry`, `client.Node`). Die Web-UI bekommt
  dort nur den Bridge-Node (`internal/daemon/web.go:25-34`); `ctrl` wird nie
  übergeben.
- go-fabric `v0.13.3` hat die Fassade, die eine UI braucht: `Controller.Pair`,
  `Discover`, `Registry().Commissioned/Decommission/SetFabricLabel`,
  `client.Node.Read/Write/Invoke/Subscribe/OpenCommissioningWindow/ICD/Properties/
  Record/ConnectionState/Endpoints`, `SoftwareUpdates()` (OTA-Konsens und
  -Warteschlange), `controller/thread` (Dataset-Decoder; kein OTBR-Client —
  Meilenstein 6). Für ACL und Bindings gibt es keine eigene Fassadenmethode;
  der Server schreibt sie über `Node.Write` (`cmd_acl.go:53`).

## B. Ziel

Jede Einbettung von go-fabrics Controller (go-fabric-server, später ein
Referenz-Controller oder loom) bekommt dieselben Controller-Seiten, so wie
jede Einbettung des Nodes heute die Node-Seiten bekommt: ein Go-Interface,
Handler und Vertrag im Modul, die Svelte-Seiten im npm-Paket, eine Shell, die
beide Familien nach Vorhandensein anzeigt.

## C. Form

### C.1 Go-Hälfte: `ControllerSource`

```go
type ControllerSource interface {
    Fabric() ControllerFabric                      // Fabric-ID, Compressed-Fabric-ID, Label, Vendor
    Nodes() []ControllerNode                       // Registry().Commissioned()
    Node(id uint64) (ControllerNode, bool)
    Pair(ctx, code string, opts PairOptions) (ControllerNode, error)
    Discover(ctx, filter DiscoverFilter, timeout) ([]Discovered, error)
    CommissionDiscovered(ctx, d Discovered, pin uint32) (ControllerNode, error)
    Decommission(ctx, id uint64) error
    SetFabricLabel(ctx, label string) error
    Credentials() CredentialStore                  // Wi-Fi/Thread: Namen lesen, write-only setzen/löschen (nil = keine)
    SoftwareUpdates() SoftwareUpdateSource         // nil = kein OTA
    Observe(fn func(ControllerEvent)) (unsubscribe func())
}
type ControllerNode interface {
    ID() uint64; Record() NodeRecord; Properties() NodeProperties
    ConnectionState() string; SubscriptionActive() bool; Disabled() bool
    Endpoints() []NodeEndpoint                     // Endpoints, Clusters (IDs+Namen), Device-Types
    Read(ctx, paths []AttributePath, fabricFiltered bool) ([]AttributeValue, error)
    Write(ctx, path AttributePath, value any, timed time.Duration) (WriteStatus, error)
    Invoke(ctx, path CommandPath, payload any, timed time.Duration) (any, error)
    OpenCommissioningWindow(ctx, timeout time.Duration) (WindowInfo, error)
    RevokeCommissioningWindow(ctx) error
    Interview(ctx) error; Enable(ctx) error; Disable(ctx) error; SetAutoSubscribe(ctx, bool) error
    ICD() (ICDState, bool); RegisterICD(ctx, multiAdmin bool) error; UnregisterICD(ctx, force bool) error
}
```

Ein Adapter `fabricui.ControllerOf(*controller.Controller, ControllerOptions)`
im Modul erfüllt das Interface über die Fassade (der Weg, den go-fabric-server
heute pro WS-Kommando geht, einmal zentral); go-fabric-server übergibt `ctrl`
in `Options.Controller`. ACL und Bindings sind keine Interface-Methoden: die
Handler bauen die Writes aus den generierten Definitionen (`accesscontroldef`,
`bindingdef`) und gehen über `Write`, wie der Server (`cmd_acl.go`).

### C.2 Vertrag und Routen (`/api/v1/controller/…`)

| Route | Methode | Admin | Inhalt |
| --- | --- | --- | --- |
| `/controller/status` | GET | – | Fabric (hex), Label, Vendor, Node-Zahl, Fähigkeiten (OTA, Thread, BLE), `csrf_token` |
| `/controller/nodes` | GET | – | Liste: `node_id_hex`, Vendor/Produkt/Name, Verbindung, Subscription, ICD, Adressen, Disabled |
| `/controller/nodes/{id}` | GET | – | Record, Properties, Endpoints mit Clustern und Device-Types |
| `/controller/nodes/{id}` | DELETE | ja | Decommission |
| `/controller/nodes/{id}/interview` · `/enable` · `/disable` · `/auto-subscribe` | POST | ja | Verwaltung |
| `/controller/nodes/{id}/window` · `/window/close` | POST | ja | Multi-Admin-Fenster (Codes, QR, Timeout) |
| `/controller/nodes/{id}/attributes` | POST | – | gezieltes Lesen (Pfade im Body), roh: `{endpoint, cluster_hex, attribute_hex, value}` |
| `/controller/nodes/{id}/attributes` | PUT | ja | ein Write (typisierte Seiten nutzen ihn; generische Writes nur Admin) |
| `/controller/nodes/{id}/commands` | POST | ja | ein Invoke |
| `/controller/nodes/{id}/acl` | GET / PUT | – / ja | ACL-Einträge der eigenen Fabric |
| `/controller/nodes/{id}/bindings` | GET / PUT | – / ja | Bindings je Endpoint |
| `/controller/nodes/{id}/icd` | GET / POST register / POST unregister | – / ja | ICD |
| `/controller/nodes/{id}/fabrics` | GET / DELETE `{index}` | – / ja | Fabrics auf dem Node |
| `/controller/commission` | POST | ja | `{code}` oder `{discovered, pin}`; Antwort: Node oder Problem |
| `/controller/discovery` | GET | ja | `?filter=…&timeout=…`: commissionable Geräte |
| `/controller/credentials` | GET / PUT / DELETE | – / ja | Wi-Fi/Thread-Zugangsdaten: Namen lesen, write-only setzen |
| `/controller/fabric/label` | PUT | ja | Label |
| `/controller/updates` | GET / POST consent / POST check | – / ja | OTA-Warteschlange, Konsens, Prüfung |
| `/controller/stream` | GET | – | SSE: `node_added/updated/removed`, `connection`, `subscription`, `commissioning_progress`, `update` |

Regeln wie in der Node-Familie: 64-Bit-IDs als `*_hex`, Problem+JSON, Admin
über `Options.Admin` oder `X-Fabric-UI-Token`, Geheimnisse (PIN, Zugangsdaten)
nur in Admin-Routen und nie zurückgelesen. Vertragsversion 0.2.0; `contract/
previous/` bleibt der 0.1.x-Vergleich; `Events.Observers` fällt in 0.2.0 weg
(U1b Frage 1; beide Hosts haben die Zeile schon entfernt).

### C.3 Svelte-Seiten (`ui/src/lib/pages/controller/`)

| Seite | Inhalt |
| --- | --- |
| Nodes | Tabelle (DataTable) mit Filtern (Verbindung, Vendor, ICD); Aktionen: Interview, Fenster öffnen, Deaktivieren, Entfernen |
| Commission | Code/QR-Eingabe (Kamera nicht in U2), Discovery-Liste mit „Koppeln“ + PIN, Zugangsdaten-Auswahl, Fortschritt über SSE, Ergebnis |
| Node | Tabs: Übersicht (Record, Properties, Verbindung), Endpoints (Baum, Cluster, Device-Types), Attribute (Pfad-Leser), ACL, Bindings, Fenster, ICD, Fabrics, Updates |
| Settings | Fabric-Label, Zugangsdaten (Namen; Anlegen write-only), Controller-Info |

Shell: Familien nach Vorhandensein (`/node/status` 200 → Node-Tabs,
`/controller/status` 200 → Controller-Tabs; 404 → Familie ausgeblendet),
Hash-Routen `#/controller/<seite>`; i18n-Schlüssel `controller.*` en/de.

## D. Welle

Eine Welle (Owner-Entscheidung 0.2): **U2 → go-fabric-ui `v0.2.0`** — Vertrag
0.2.0 mit allen Routen und Schemas, `ControllerSource` + `ControllerOf`, die
Handler (Stubs aus dem Vertrag generiert, siehe E), die vier Seiten, die Shell
mit Familienerkennung, `Events.Observers` entfernt, Doku. Danach **S4** in
go-fabric-server (`Options.Controller = fabricui.ControllerOf(ctrl, …)`,
Ingress zeigt beide Familien, `v0.4.0`); die Add-on-Bluetooth-Option
(`host_dbus`, Adapter) folgt nach dem manuellen BLE-Test im selben Bündel.

## E. Besser als andere, besser als vorher, Schnitt richtig?

- **Besser als andere.** matter.js und chip bieten keine Controller-Verwaltungs-
  oberfläche; Home Assistants Matter-Integration zeigt nur die Gerätekarten,
  keine Fabrics, Fenster, ACL, ICD oder OTA-Warteschlange; python-matter-server
  hat ein Debug-Dashboard mit Roh-Kommandos. Hier: eine Familie von Seiten,
  die jede go-fabric-Einbettung bekommt, mit Vertrag, Fixtures und Tests auf
  beiden Seiten.
- **Besser als vorher.** Bis U1 gab es die Controller-Hälfte nur über das
  WS-Protokoll für HA; wer den Daemon betreibt, sieht Knoten nicht. Nach U2
  zeigt Ingress beide Familien; die Fassadenaufrufe, die der Server pro
  WS-Kommando macht, stehen einmal in `ControllerOf`.
- **Schnitt.** Die Seiten reden REST gegen Handler im Modul (nicht direkt
  gegen `/ws`): so läuft die Familie in jeder Einbettung, nicht nur im
  Daemon, und Auth/Problem/Hex-Regeln bleiben eine. Das WS-Protokoll bleibt
  der HA-Vertrag und ändert sich nicht. `ControllerSource` ist ein
  Port-Interface mit eigenen kleinen Typen (kein go-fabric-Typ im Vertrag,
  wie `NodeSource`), damit das Paket ohne go-fabric-Versionskopplung lebt;
  der Adapter `ControllerOf` ist die einzige Stelle, die go-fabrics Fassade
  kennt. **Generiert statt handgeschrieben:** der Vertrag ist die Quelle —
  die Go-Handler-Stubs (Router, Request-/Response-Typen, Validierung) kommen
  per `oapi-codegen` aus `contract/openapi.yaml`, der TS-Client wie bisher per
  `openapi-typescript`; Vertrag, Handler und Client können nicht
  auseinanderlaufen, handgeschrieben bleibt nur die Handler-Logik (Port-
  Aufruf, Problem-Antwort). Das ist go-fabrics ADR-0013-Muster auf der
  UI-Schicht. Die Node-Familie bleibt in U2 handgeschrieben; ihre Umstellung
  auf die generierten Stubs ist eine eigene spätere Welle. Kein Attribut-Browser über alle Pfade (das ist chip-tool); gelesen
  wird gezielt, geschrieben nur typisiert oder als Admin.

## F. Entscheidungen (Owner, 2026-10-27, per Wizard)

- **0.1 Form:** A — REST-Handler im Modul über den Port `ControllerSource`
  mit dem Adapter `ControllerOf`, **die Handler-Stubs aus dem Vertrag
  generiert** (`oapi-codegen`), der TS-Client aus demselben Vertrag. Nicht
  die Browser-Seite am WS-Protokoll (B: fremder, HA-geformter Vertrag, nur
  im Daemon lauffähig); nicht ein eigener Kommando-Layer mit zwei
  Transporten (C: die transportneutrale Logik ist schon go-fabrics Fassade).
- **0.2 Schnitt:** eine Welle, `v0.2.0`; dann S4.
- **0.3 Zugangsdaten:** Wi-Fi/Thread-Zugangsdaten in der UI anlegbar und
  löschbar (Admin, write-only; Namen lesbar, Geheimnisse nie zurückgelesen).
- **0.4 Roh-Zugriff:** gezieltes Lesen beliebiger Attributpfade ohne Admin;
  generische Writes und Invokes nur als Admin; keine Vollinventur.
