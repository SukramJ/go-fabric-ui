# Konzept U6 — Mehrere Nodes, in der UI angelegt (Phase U6)

Stand 2026-11-05. Entscheidungen des Owners vom selben Tag (Abschnitt F).
Englisch sind Issues, PRs und die Doku; dieses Konzept ist deutsch, solange es
gilt, und wandert nach `docs/history/`, wenn U6 abgeschlossen ist.

## A. Ist

- go-fabric-ui 0.5.0 (Contract 0.5.0): `GET /nodes` listet `Options.Node`
  (Key `default`) und `Options.Nodes`, jede Route auch unter
  `/nodes/{key}/…`; die Shell zeigt den Selector ab zwei Nodes
  (`docs/contract.md` §„The node selector“). `Options.Nodes` ist ein
  statischer Slice; Exposure, Reassemble und Events gehören dem
  Default-Node, jeder weitere Node antwortet auf den Exposure-Routen 503
  (`fabricui/options.go:103-109`).
- go-fabric-server 0.5.0 hostet mehrere Device-Nodes deklarativ über
  `--bridge-node key,listen=…` (`internal/daemon/nodes.go:21-34,:122`),
  jeder mit Port, Diskriminator, Passcode, Store `bridge-<key>.db`,
  `/bridge/<key>`.
- go-ha2fabric 0.5.2 hostet genau einen Node (`cmd/ha2fabric/main.go:87`),
  Selection in `selection.json` (`internal/selection/selection.go:51-74`),
  ein HA-Client mit einem `Handler` (`internal/ha/client.go:85,:149`).
- Vergleichspartner home-assistant-matter-hub: Bridges werden in der UI
  angelegt, „each … completely independent … uses its own port“, der Port
  wird automatisch ab 5540 vergeben
  (`docs-site/docs/getting-started/bridge-configuration.md:3-4,:16`;
  `packages/backend/src/services/bridges/bridge-service.ts:80`
  `getNextAvailablePort`; `packages/common/src/bridge-data.ts:219-222`
  `CreateBridgeRequest` mit optionalem Port). Ökosystem-Eigenheit, die HAMH
  dokumentiert: Alexa koppelt zuverlässig nur auf Port 5540 und verweigert
  Bridges über etwa 80–100 Geräte (`bridge-configuration.md:28-33`).

## B. Ziel

Ein Host zeigt verschiedenen Ökosystemen verschiedene Geräteumfänge: ein
Node für Apple Home, einer für Alexa. Matter erlaubt einem Node mehrere
Fabrics, aber ein Endpoint-Baum ist nicht fabric-spezifisch; ein anderer
Umfang je Ökosystem ist ein anderer Node. Der Operator legt den Node in der
UI an, wählt seine Geräte, liest den Pairing-Code, koppelt — ohne Neustart,
ohne Port, Diskriminator oder Passcode von Hand.

Regel: **der erste Node ist Konfiguration** (die Identität des Daemons:
`--bridge-*`, die Add-on-Optionen, wie heute); **jeder weitere Node ist
Registry-Zustand** im Store des Hosts, angelegt, umbenannt und gelöscht
über die UI. Keine zweite Quelle: go-fabric-server's `--bridge-node`
entfällt wieder.

## C. Entwurf

### C.1 Contract 0.6.0 (additiv über 0.5.0)

| Route | Methode | Admin | Antwort | Liest |
| --- | --- | --- | --- | --- |
| `/nodes` | POST | ja | 201 `NodeSummary`, 400, 403, 409 `node_limit`, 503 | `NodeRegistry.Create(ctx, NodeSpec{Label})` |
| `/nodes/{key}/label` | PUT | ja | 204, 400, 403, 404, 503 | `NodeRegistry.Rename(ctx, key, label)` |
| `/nodes/{key}` | DELETE | ja | 204, 403, 404, 409 `node_is_default`, 503 | `NodeRegistry.Delete(ctx, key)` |

- `GET /nodes` bleibt; mit Registry liest es `NodeRegistry.List()` statt
  `Options.Nodes`, und `NodeSummary` bekommt `managed: bool` (anlegbar,
  löschbar) und `exposure: bool`. `GET /node/status` meldet
  `node_registry: bool`, damit die Shell den Anlegen-Knopf zeigt.
- Ereignis `nodes_changed` auf dem Event-Stream (`{keys: []string}`), nach
  Create, Rename, Delete.
- Port in `fabricui`:

  ```go
  type NodeSpec struct{ Label string }
  type NodeRegistry interface {
      List() []NodeEntry
      Create(ctx context.Context, spec NodeSpec) (NodeEntry, error)
      Rename(ctx context.Context, key, label string) error
      Delete(ctx context.Context, key string) error
  }
  ```

  `Options.Registry NodeRegistry`: gesetzt, liefert sie die Liste neben
  `Options.Node` (dem Default-Node) und ersetzt `Options.Nodes`; beides
  gesetzt ist ein Konfigurationsfehler (`New` lehnt ab). Der Host vergibt
  Key (Slug aus dem Label, eindeutig), Port, Diskriminator, Passcode,
  Store; er startet den Node in `Create` und stoppt ihn in `Delete`.
  Fehler des Hosts: `fabricui.ErrNodeLimit` → 409 `node_limit`,
  `ErrDefaultNode` → 409 `node_is_default`, sonst 500.
- `NodeEntry` bekommt je Node `Exposure ExposureProvider`,
  `Reassemble func(ctx) error` und `Events *Events` (nil wie heute: 503
  und `exposure: false`). Die Exposure-Routen und die Selection-Seite
  arbeiten damit unter `/nodes/{key}/…`; `Options.Exposure` bleibt für den
  Default-Node.
- Delete ist endgültig: der Host entfernt Fabrics (die Controller
  verlieren den Node), Store und Selection. Die Shell fragt zweimal, mit
  dem Label und der Fabric-Zahl.

### C.2 Shell

- Selector-Header: Eintrag „Node anlegen…“ (nur mit `node_registry` und
  Admin); Dialog mit Label → `POST /nodes` → Wechsel auf
  `#/node/<key>/selection` → Hinweis „Pairing-Code unter Pair“.
- Statusseite eines verwalteten Nodes: Label ändern, Node löschen
  (Doppelbestätigung), Ökosystem-Hinweis aus HAMHs Tabelle: „Alexa:
  Node auf Port 5540, unter 80 Geräten“ (statische Kopie mit Link wie die
  Ecosystem-Hinweise der U5).
- Playwright-Smoke: anlegen → Selector zeigt zwei → löschen.

### C.3 Hosts

- **go-ha2fabric**: Registry `nodes.json` neben dem Store
  (`{nodes: [{key, label, listen, discriminator, passcode, created_at}]}`),
  je Node `ha2fabric-<key>.db` und `selection-<key>.json`, ein
  `bridge.Bridge` je Node am selben HA-Client — der Client fächert seinen
  `Handler` auf (ein `Handler`, der an N Bridges verteilt); Port = nächster
  freier ab Default+1, Diskriminator zufällig 12 Bit und eindeutig,
  Passcode zufällig und `setup.IsValidPasscode`; Start in `Create`, Stop
  und Löschen in `Delete`; Start aller Registry-Nodes beim Daemon-Start.
  Add-on: `host_network: true` (`addon/config.yaml:37`), neue Ports
  brauchen keine Freigabe.
- **go-fabric-server**: dieselbe Registry im Store des Daemons,
  `/bridge/<key>` entsteht zur Laufzeit, `server_info.node` wie heute;
  `--bridge-node` entfällt (CHANGELOG Removed; 0.5.0 ist einen Tag alt).
  Die Mehr-Node-Tests (`internal/daemon/multinode_test.go`) laufen über die
  Registry.

## D. Wellen

1. **U6a** go-fabric-ui: Contract 0.6.0, Port, Routen, Shell, Smoke,
   `docs/contract.md`, Changelog (dabei die Überschrift `[0.5.0] -
   unreleased` auf das Tag-Datum). Owner taggt 0.6.0.
2. **U6b** go-ha2fabric und **S-U6** go-fabric-server, parallel, je ein
   Opus, auf go-fabric-ui 0.6.0.
3. Ökosystem-Profil je Node (Device-Type-Wahl für Alexa/Google, Warnung
   bei nicht unterstützten Typen wie HAMHs controller-compatibility
   warnings): go-hamapping-Issue, Umsetzung nach dem Alexa-/Google-Lauf des
   Owners.

## E. Besser als andere, besser als vorher, Schnitt richtig?

- **Besser als HAMH**: derselbe Fluss (anlegen, filtern, koppeln, Port
  automatisch), aber als Port im UI-Contract, den zwei Hosts tragen, mit
  Selection je Node auf der bestehenden Seite und Löschen, das Fabrics und
  Store mitnimmt. HAMH kennt nur sich selbst als Host.
- **Besser als vorher**: ein Besitzer für die Exposure (Node und Selection
  sind beide UI-Zustand), kein Neustart, kein Port von Hand, kein
  `--bridge-node` neben der UI.
- **Schnitt**: Modul (go-fabric) unverändert — es trägt N Nodes je Prozess
  schon; UI = Contract, Routen, Fluss; Host = Registry, Allokation,
  Lebenszyklus, Store. Nicht deklarativ, weil zwei Quellen für einen
  Begriff die schlechtere Architektur sind (Entscheidung F.1).

## F. Entscheidungen (Owner, 2026-11-05)

1. Nodes werden in der UI angelegt, nicht deklarativ; der erste Node bleibt
   Konfiguration. `--bridge-node` entfällt im Server.
2. Drei Wellen jetzt als Nebenläufe neben der go-fabric-Roadmap:
   U6a → Tag → U6b ∥ S-U6.
3. Ökosystem-Profil nur als go-hamapping-Issue bis zum Alexa-/Google-Lauf.
