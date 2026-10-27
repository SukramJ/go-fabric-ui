# @sukramj/fabric-ui

The web interface of [go-fabric](https://github.com/SukramJ/go-fabric)
nodes as a Svelte 5 library: the node pages, the UI kit they are built
from, design tokens named after Home Assistant's CSS variables with a theme
bridge for an Ingress panel, the API client generated from the contract,
and English and German catalogues. The Go module `fabricui` of the same
repository serves the API; its standalone shell is built from `src/shell/`.

```sh
npm install @sukramj/fabric-ui
```

The package ships sources (`.svelte`, `.ts`), compiled by the host's Vite
build with `@sveltejs/vite-plugin-svelte`; `svelte` is a peer dependency.

## Embedding the pages

Provide the context once, above the pages, with the host's translator,
toast and confirm:

```svelte
<script lang="ts">
  import { createClient, createFabricUI, createTranslator, provideFabricUI, NodePair, NodeFabrics } from "@sukramj/fabric-ui";

  const ui = provideFabricUI(
    createFabricUI({
      client: createClient("/api/v1", hostFetch), // the host's fetch: its auth, its base path
      t: createTranslator(() => prefs.locale, { en: hostEn, de: hostDe }), // or the host's own t
      toast: { info: toastStore.info, success: toastStore.success, error: toastStore.error },
      confirm: (o) => confirmStore.ask(o),
      routes: { pair: "#/node/pair" }, // where the pages link each other
    }),
  );
  ui.node.connect(); // the event stream; ui.node.close() on teardown
</script>

<NodePair />
```

Pages: `NodeStatus`, `NodePair`, `NodeFabrics` (with maintenance),
`NodeSelection`, `NodeEndpoints`, `NodeDiagnostics`. Each reads only the
context; none names a host concept. The selection page is driven by the
provider's schema (see the repository's `docs/provider.md`).

Controller pages (v0.2.0): `ControllerNodes`, `ControllerCommission`,
`ControllerSettings`, and `ControllerNode` with the props `nodeId` (the
16-digit hex `node_id_hex`) and an optional first `tab` (`overview`,
`endpoints`, `attributes`, `acl`, `bindings`, `window`, `icd`, `fabrics`,
`updates`). They read `ui.controller` (`createControllerStore`); the
routes `controllerNodes`, `controllerCommission`, `controllerSettings` and
the prefix `controllerNode` link them. `ui.controller.loadStatus()` sets
`present` to false on a 404: the host serves no controller.

### Translator

`t(key, vars)` is the host's. `createTranslator(locale, extra)` builds one
over the package's catalogues (`catalogs.en`, `catalogs.de`, keys
`node.*`, `controller.*`, `shell.*`, `common.*`) plus the host's own; a `locale` getter that reads
reactive state re-renders the pages on a language switch.

### Stylesheet

The pages use Tailwind CSS 4 utilities on the `--ha-*` tokens. In the
host's stylesheet, after Tailwind:

```css
@import "tailwindcss";
@import "@sukramj/fabric-ui/theme/tokens.css";
@source "../node_modules/@sukramj/fabric-ui/src/lib";
@custom-variant dark (&:where(.dark, .dark *));
```

`.dark` on `<html>` switches the dark tokens; `data-skin="ha"` on `<html>`
resolves the tokens from Home Assistant's variables. `applySkin("ha")` and
`startHaBridge()` (from `@sukramj/fabric-ui/theme`) do both inside an
Ingress frame and are inert elsewhere. A host on a skin of its own (any
other `data-skin`, or none) keeps the standalone tokens; the HA skin's
extras (`accent-color` and `color-scheme` on native controls,
`--ha-on-primary-color` from Home Assistant's `--text-primary-color`) apply
only under `data-skin="ha"`. The standalone shell sets the HA skin and the
mode in a pre-paint script (`src/shell/index.html`) so the first frame
already wears them; a host that wants the same does it in its own
`index.html`.

### Event stream

`ui.node.connect()` opens `GET /node/stream` with an `EventSource`. A host
with a hub of its own feeds it instead:

```ts
import { subscribeStream } from "@sukramj/fabric-ui/client";

createFabricUI({ client, subscribe: (onEvent) => hub.on("fabric-ui", onEvent) });
// or bridge the stream into the hub:
const close = subscribeStream(client.streamUrl(), (ev) => hub.publish(ev.type, ev.data));
```

`ui.controller.connect()` opens `GET /controller/stream` the same way;
`subscribeControllerStream` and the `subscribeController` option of
`createFabricUI` are its twins.

## Developing

```sh
npm ci
npm run dev       # the shell on :5174, /api proxied to FABRIC_UI_API (default http://localhost:8080)
npm run check     # svelte-check
npm test          # vitest: pages, kit, client against the contract fixtures, the contract guard
npm run gen:types # regenerate src/lib/api/schema.gen.ts from ../contract/openapi.yaml
npm run build     # the shell into ../fabricui/dist (committed; make dist-check guards it)
```
