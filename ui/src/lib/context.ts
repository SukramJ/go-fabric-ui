// The fabric-ui context: what every page reads — the client, the host's
// translator, toast and confirm, the node store and the links between the
// pages. A host provides it once above the pages (provideFabricUI); the
// standalone shell provides its own.
import { getContext, hasContext, setContext } from "svelte";
import type { FabricClient } from "./api/client";
import { defaultT, type Translate } from "./i18n";
import { createNodeStore, type NodeStore, type NodeStoreDeps, type Toast } from "./store.svelte";

export type { Toast } from "./store.svelte";

export interface ConfirmOptions {
  title: string;
  body?: string;
  destructive?: boolean;
  confirmLabel?: string;
}

export type Confirm = (opts: ConfirmOptions) => Promise<boolean>;

/** The hrefs the pages link each other with; a host maps them onto its router. */
export interface Routes {
  status: string;
  pair: string;
  fabrics: string;
  selection: string;
  endpoints: string;
  diagnostics: string;
}

export const defaultRoutes: Routes = {
  status: "#/node/status",
  pair: "#/node/pair",
  fabrics: "#/node/fabrics",
  selection: "#/node/selection",
  endpoints: "#/node/endpoints",
  diagnostics: "#/node/diagnostics",
};

export interface FabricUI {
  client: FabricClient;
  t: Translate;
  toast: Toast;
  confirm: Confirm;
  node: NodeStore;
  routes: Routes;
}

export interface FabricUIOptions {
  client: FabricClient;
  t?: Translate;
  toast?: Toast;
  confirm?: Confirm;
  routes?: Partial<Routes>;
  subscribe?: NodeStoreDeps["subscribe"];
}

const KEY = Symbol.for("@sukramj/fabric-ui");

const consoleToast: Toast = {
  info: (m) => console.info(m),
  success: (m) => console.info(m),
  error: (m) => console.error(m),
};

/** Builds a context; the node store is created on the client. */
export function createFabricUI(opts: FabricUIOptions): FabricUI {
  const t = opts.t ?? defaultT;
  const toast = opts.toast ?? consoleToast;
  return {
    client: opts.client,
    t,
    toast,
    confirm: opts.confirm ?? (async (o) => window.confirm(o.body ? `${o.title}\n\n${o.body}` : o.title)),
    node: createNodeStore(opts.client, { t, toast, subscribe: opts.subscribe }),
    routes: { ...defaultRoutes, ...opts.routes },
  };
}

/** Puts ui into the component context; call it in the component above the pages. */
export function provideFabricUI(ui: FabricUI): FabricUI {
  setContext(KEY, ui);
  return ui;
}

/** The context the pages read; it throws outside provideFabricUI. */
export function useFabricUI(): FabricUI {
  const ui = getContext<FabricUI | undefined>(KEY);
  if (!ui) throw new Error("fabric-ui: no context — call provideFabricUI above the pages");
  return ui;
}

/** The translator of the context, or English outside one (the kit uses it). */
export function useT(): Translate {
  return hasContext(KEY) ? getContext<FabricUI>(KEY).t : defaultT;
}
