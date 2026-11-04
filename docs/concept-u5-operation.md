# Konzept U5 — Betrieb (Milestone 9, Phase U5)

Stand 2026-11-04. Roadmap-Zeile U5 (go-fabric `docs/roadmap.md:378`,
Langform `docs/concept-v13-roadmap-extension.md:99`): Live-Status statt
Reload; eine Diagnoseseite mit Log-Level, Log-Stream und `ping_node`; der
Kamera-QR-Scan im Kommissionierungsfluss; Ökosystem-Hinweise im Pairing;
der Server zeigt die Node-Seiten für jeden Node, nicht nur die Bridge; die
Selection-Seite bekommt im Server einen Grund oder einen Inhalt;
Playwright-e2e je Seite. Inventur: `~/.cache/go-fabric-scratch/read-u5.md`
(Sonnet, Karte); tragende Zeilen von Fable geprüft (unten zitiert).
Vergleichsziel: matterjs-servers Dashboard bei 0bd06f22; Hilfsreferenz
home-assistant-matter-hub (Log-API-Form).

## A. Ist

- **Live-Status:** Der Node-Stream (`fabricui/stream.go`, `events.go:14-31`)
  trägt Fabrics, Fenster, Endpoints, Surface, Sessions, Subscriptions und
  seit U4a `attribute_changed`; der Controller-Stream Verbindung und
  Subscription je Node. Health wird je Statusladen einmal abgefragt;
  `Node.OnHealthSample` (go-fabric `node/health.go:93`) und
  `Node.OnDiagnosticEvent` (`node/diagnostics.go:40`) sind nicht in
  `NodeSource` (`fabricui/options.go:25`) und nicht im Stream; die
  Diagnoseseite lädt einmal mit Refresh-Knopf.
- **Logs und Ping:** Nur go-fabric-server kennt `get_loglevel`/`set_loglevel`
  (`controllerapi/cmd_server.go:76-102`, `slog.LevelVar` in
  `internal/daemon/run.go:184`) und `ping_node` (`cmd_node.go:167`: ein
  Matter-Read von DataModelRevision je Adresse, kein ICMP — distroless,
  server by_design B1) — alles nur auf `/ws`. Einen Log-Stream hat niemand:
  der Server loggt nach stderr; matterjs-servers Dashboard hat einen
  Log-Level-Dialog (`log-level-section.ts`) und ruft `pingNode` nie.
  go-fabric selbst hat weder LevelVar noch Ring (`node.Config.Logger` ist
  ein `*slog.Logger`, `node/config.go:217-220`).
- **QR-Scan:** nirgends (go-fabric-ui, matterjs-server, hub); der Vertrag
  braucht keine Änderung — ein Scan füllt `CommissionRequest.code`
  (`contract/openapi.yaml`, `Commission.svelte`, Kopfkommentar „camera not
  used“). Kamera nur im sicheren Kontext (HTTPS/localhost), wie die
  Zwischenablage auf `Pair.svelte`.
- **Ökosystem-Hinweise:** Daten vorhanden (`eligibility.EcosystemForVendor`,
  `/node/compatibility`, i18n `node.diag.ecosystem.*`), nur die Diagnose
  rendert sie; matter.js `docs/ECOSYSTEMS.md:85-131` ist die Textquelle.
- **Node-Seiten:** Die Node-Familie ist eine `NodeSource` (`options.go:62-65`),
  der Server hostet genau einen Geräte-Node (`run.go:370`); kommissionierte
  Controller-Nodes haben Seiten, importierte Test-Nodes nicht.
- **e2e:** vitest + goldene Fixtures, `fabricui/e2e_test.go`,
  `memory_controller.go`; kein Playwright.

## B. Ziel

Betrieb ohne Reload und ohne zweites Werkzeug: Status und Diagnose
aktualisieren sich vom Stream; Log-Level und Log-Stream, Ping, Scan und
Hinweise liegen in der UI; der Server hostet mehrere Geräte-Nodes und die
UI wählt den Node; jede Seite hat einen Playwright-Smoke.

## C. Entwurf

### C.1 Vertrag 0.5.0 (additiv gegen 0.4.1)

- **Node-Wähler.** Neuer Pfadparameter `{node}` (stabiler Schlüssel des
  Geräte-Nodes, `GET /nodes` → `NodeSummary[]{key, label, kind, enabled}`)
  vor jeder Node-Route: `/nodes/{node}/status`, `/nodes/{node}/pair`, …,
  `/nodes/{node}/stream`. Die bestehenden `/node/*`-Routen bleiben und
  meinen den **Standard-Node** (`Options.Node`), so bleibt der Vertrag
  additiv und loom/go-ha2fabric laufen unverändert; `Options.Nodes
  map[string]NodeSource` (+ `NodeLabels`) ist die Erweiterung, `Options.Node`
  der Standard. Shell: Node-Wähler in der Kopfzeile, Route `#/node/<key>/…`.
- **Betrieb.** `GET /operation/log/level` → `{level, levels[]}`,
  `PUT /operation/log/level` (Admin), `GET /operation/log` →
  `LogRecord[]{time, level, msg, attrs}` (Ring, neueste zuletzt, `since`,
  `limit`), `GET /operation/log/stream` (SSE `log_record`, Heartbeat wie
  die anderen Streams), `GET /operation/info` → `{version, go_fabric,
  uptime_s, started_at}`. Familienneutral (auch ein reiner Controller-Daemon
  hat die Seite).
- **Ping.** `POST /controller/nodes/{id}/ping` (Admin) →
  `PingResult{addresses[]{ip, reachable, rtt_ms?}, method: "matter-read"}`;
  die UI nennt die Semantik („erreichbar per Matter-Read“).
- **Stream.** Node-Stream: `health_sample {healthy, note, sampled_at}` und
  `diagnostic_event` (Form von `DiagnosticEvent`), aus
  `Observers`-Erweiterung der NodeSource (C.2).
- Vertragsregeln wie bisher: nur Ergänzungen, Fixtures auf beiden Seiten,
  `contract/previous` = 0.4.1.

### C.2 Go-Hälfte

- `NodeSource` wird **erweitert (breaking, Entscheidung F.3):**
  `OnHealthSample(func(node.HealthSample)) func()` und
  `OnDiagnosticEvent(func(node.DiagnosticEvent)) func()` — `*node.Node`
  erfüllt beides; `fabricui.NodeObservers` reicht sie als Stream-Ereignisse.
  Migrationszeile für loom, go-ha2fabric, Server in `changelog.md`
  (`### Changed`, Breaking).
- Neuer Port `fabricui.Operations` (Options-Feld): `LogLevel() slog.Level`,
  `SetLogLevel(slog.Level) error`, `Log(since, limit) []LogRecord`,
  `ObserveLog(func(LogRecord)) func()`, `Info() OperationInfo`. Ein Host
  erfüllt ihn mit dem go-fabric-Paket aus C.4; `MemoryOperations` als
  Testdouble.
- `ControllerNode.Ping(ctx) (PingResult, error)` auf `Read` (ein Read von
  BasicInformation DataModelRevision je bekannter Adresse; Adapter
  `fabricui/gofabric/node.go`); `MemoryController` folgt.
- Mehrere Nodes: Handler-Registrierung je Schlüssel, Standard-Node unter
  `/node/*`; Stream je Node.

### C.3 Svelte-Hälfte

- Seite **Betrieb** (beide Familien): Log-Level-Wahl, Log-Stream (Ring +
  SSE, Pause, Filter nach Level/Text, Kopieren), Info (Version, Uptime).
- Status- und Diagnoseseite laden auf `health_sample`/`diagnostic_event`
  nach (keine Reload-Knöpfe mehr nötig; Refresh bleibt).
- Controller-Node-Seite: Ping-Knopf (Admin) mit Ergebnis je Adresse.
- Kommissionierung: Knopf „Scannen“ → lazy geladener Chunk (Decoder
  MIT/Apache, Notice + Lizenztext, `notices-check`); native
  `BarcodeDetector` wenn vorhanden, sonst Bibliothek; auf plain HTTP eine
  Erklärung statt des Knopfs; das Ergebnis füllt `code`.
- Pairing: Ökosystem-Hinweise (Apple, Google, Alexa, SmartThings, Home
  Assistant) als statische Texte in `en.json`/`de.json` aus
  ECOSYSTEMS.md:85-131, mit Links statt Versionsnummern; die vorhandene
  Verträglichkeit (`/node/compatibility`) daneben.
- Selection-Seite: ohne `Exposure` zeigt sie den Grund des Hosts
  (`status.exposure_reason`, neu, optional) statt zu fehlen.
- Playwright: `ui/e2e/` mit einem Smoke je Seite (lädt, zeigt das Fixture,
  keine Konsolenfehler), nur Chromium, gegen den Go-e2e-Server
  (`fabricui/e2e_test.go`-Site + `MemoryController`, als `make e2e-web`);
  eigener CI-Job.

### C.4 go-fabric-Modulteil (Issue in go-fabric)

Paket `logring` (Entscheidung F.2): ein `slog.Handler`-Wrapper mit
`LevelVar`, begrenztem Ring (`N` Einträge, Standard 1000), `Records(since,
limit)` und `Observe(func(Record)) func()`; jeder Host schaltet ihn vor
seinen Logger (`node.Config.Logger`); keine Passcodes im Log (go-fabric
loggt sie heute nicht; der Ring redigiert nichts weiter). Außerdem die
Prüfung, dass `OnHealthSample`/`OnDiagnosticEvent` die Formen tragen, die
der Stream braucht (sie tun es: `node/health.go:93`, `node/diagnostics.go:40`).

### C.5 go-fabric-server (eigene Welle)

Mehrere Geräte-Nodes je Daemon (Konfiguration je Node: Name, Port,
Discriminator, Store-Pfad; `startBridge` je Eintrag, `run.go:370`),
`Options.Nodes` an die UI, `logring` vor den Logger, `Operations`- und
`Ping`-Port verdrahtet (`/ws`-Befehle bleiben, lesen dieselben Quellen),
Selection-Grund „der Server hat kein Host-Modell“.

## D. Wellen

- **D.1 U5a (go-fabric-ui, Opus):** Vertrag 0.5.0, Ports, Node-Wähler,
  Betriebsseite, Ping, Stream-Ereignisse, Scan, Hinweise, Playwright-Smoke;
  Testdoubles für alles; go-fabric-ui bleibt 0.5.0 unreleased.
- **D.2 logring (go-fabric, Opus, nach Phase 33):** Issue aus C.4.
- **D.3 S-U5 (go-fabric-server, Opus, nach D.1 und D.2):** C.5.

## E. Besser als andere, besser als vorher, Schnitt richtig?

- **Besser als andere.** matterjs-servers Dashboard hat einen
  Log-Level-Dialog, keinen Log-Stream, keinen Ping-Aufruf, keinen Scan und
  keine Ökosystem-Hinweise; hier liegen alle fünf in der UI, der Log-Ring
  ist eine Quelle für UI und `/ws`, und der Ping sagt ehrlich, was er misst.
- **Besser als vorher.** Status und Diagnose aktualisieren sich vom Stream;
  Betrieb braucht kein zweites Werkzeug; der Server trägt mehrere Nodes;
  jede Seite hat einen Browser-Smoke statt nur Komponententests.
- **Schnitt.** Der Ring und das Level liegen im Modul (jeder Host, eine
  Form); die UI kennt nur Ports; der Vertrag bleibt additiv (Standard-Node
  unter `/node/*`); das Breaking in `NodeSource` ist bewusst (F.3): ein
  optionales Interface hätte zwei Verhaltensweisen erzeugt.
- **Nicht gemacht:** keine Log-Datei, kein ICMP, keine Stream-Werte, keine
  Node-Erstellung aus der UI, Playwright nur Smoke (keine vollen Flüsse).

## F. Entscheidungen

Owner, per Wizard 2026-11-04:

- **F.1 Mehrere Geräte-Nodes im Server** mit Node-Wähler in der UI
  (`{node}`, `Options.Nodes`; Breaking für Hosts, Vertrag additiv).
- **F.2 Log-Ring und Level-Helfer im go-fabric-Modul** (Paket für alle
  Hosts); die UI liest über den Port.
- **F.3 `NodeSource` erweitert (Breaking):** OnHealthSample und
  OnDiagnosticEvent Pflicht; Ping als Matter-Read-Route.
- **F.4 Lazy-Chunk-Decoder + Playwright-Smoke je Seite**, nur Chromium,
  eigener CI-Job im UI-Repo.
- Eigene Festlegungen: Ökosystem-Hinweise als statische Texte aus
  ECOSYSTEMS.md mit Links; Vertrag 0.5.0; go-fabric-ui-Tag durch den Owner.
