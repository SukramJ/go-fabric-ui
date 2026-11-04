<!-- The standalone shell the Go module serves at "/": the page families the
     host serves, found by presence — GET /node/status and GET
     /controller/status; a 404 hides a family, as does a disabled node next
     to a controller —, hash routing (#/node/<page>, #/controller/<page>,
     #/controller/node/<id>), the header with the family's tabs, toasts and
     the confirm dialog, the theme bridge. Contract 0.5.0 adds the node
     selector (GET /nodes; shown when the host lists more than one node,
     #/node/<key>/… for any but the default) and the Operation page
     (#/operation; its tab shows when GET /operation/info answers). The API is "./api/v1" relative
     to the page, so the shell works under any base path (an Ingress among
     them). -->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { ApiError, createClient, type NodeSummary } from "../lib/api/client";
  import { createFabricUI, provideFabricUI, routesForNode } from "../lib/context";
  import { createTranslator, detectLocale } from "../lib/i18n";
  import { applySkin, isEmbedded, startHaBridge } from "../lib/theme/ha-bridge";
  import PageHeader from "../lib/kit/PageHeader.svelte";
  import PageShell from "../lib/kit/PageShell.svelte";
  import LoadingState from "../lib/kit/LoadingState.svelte";
  import EmptyState from "../lib/kit/EmptyState.svelte";
  import Select from "../lib/kit/Select.svelte";
  import Tabs from "../lib/kit/Tabs.svelte";
  import Operation from "../lib/pages/operation/Operation.svelte";
  import NodeFamily from "./NodeFamily.svelte";
  import NodeScope from "./NodeScope.svelte";
  import NodeStatus from "../lib/pages/node/Status.svelte";
  import NodePair from "../lib/pages/node/Pair.svelte";
  import NodeFabrics from "../lib/pages/node/Fabrics.svelte";
  import NodeSelection from "../lib/pages/node/Selection.svelte";
  import NodeEndpoints from "../lib/pages/node/Endpoints.svelte";
  import NodeSurface from "../lib/pages/node/Surface.svelte";
  import NodeDiagnostics from "../lib/pages/node/Diagnostics.svelte";
  import NodeTopology from "../lib/pages/node/Topology.svelte";
  import ControllerNodes from "../lib/pages/controller/Nodes.svelte";
  import ControllerCommission from "../lib/pages/controller/Commission.svelte";
  import ControllerNode from "../lib/pages/controller/Node.svelte";
  import ControllerSettings from "../lib/pages/controller/Settings.svelte";
  import ConfirmDialog from "./ConfirmDialog.svelte";
  import Toaster from "./Toaster.svelte";
  import { createConfirm } from "./confirm.svelte";
  import { createToasts } from "./toast.svelte";
  import { routeOf, type Route } from "./route";

  let locale = $state(detectLocale());
  const t = createTranslator(() => locale);
  const toasts = createToasts();
  const confirm = createConfirm();
  const ui = provideFabricUI(createFabricUI({ client: createClient("./api/v1"), t, toast: toasts.toast, confirm: confirm.ask }));
  const node = ui.node;
  const controller = ui.controller;

  const nodePages = { status: NodeStatus, pair: NodePair, fabrics: NodeFabrics, selection: NodeSelection, endpoints: NodeEndpoints, surface: NodeSurface, diagnostics: NodeDiagnostics, topology: NodeTopology };
  const controllerPages = { nodes: ControllerNodes, commission: ControllerCommission, settings: ControllerSettings };

  // null until probed: whether GET /node/status answered (not 404).
  let nodeServed = $state<boolean | null>(null);
  // The nodes GET /nodes lists, the default first; empty until loaded or
  // from a server before 0.5.0.
  let nodeList = $state<NodeSummary[]>([]);
  // Whether GET /operation/info answered (not 404 or 503).
  let operationServed = $state(false);
  let route = $state<Route>(routeOf(location.hash));
  const onHash = () => (route = routeOf(location.hash));

  // The node family shows when the node route answers and either serves a
  // node or no controller is there to show instead.
  const nodeFamily = $derived(nodeServed === true && (node.status?.enabled !== false || controller.present !== true));
  const controllerFamily = $derived(controller.present === true);
  const probed = $derived(nodeServed !== null && controller.present !== null);
  const family = $derived<"node" | "controller" | "operation" | null>(
    route.family === "operation" && operationServed
      ? "operation"
      : route.family === "controller" && controllerFamily
        ? "controller"
        : nodeFamily
          ? "node"
          : controllerFamily
            ? "controller"
            : operationServed
              ? "operation"
              : null,
  );

  const familyTabs = $derived([
    ...(nodeFamily ? [{ key: "node", label: t("shell.family.node"), href: ui.routes.status }] : []),
    ...(controllerFamily ? [{ key: "controller", label: t("shell.family.controller"), href: ui.routes.controllerNodes }] : []),
    ...(operationServed ? [{ key: "operation", label: t("shell.family.operation"), href: ui.routes.operation }] : []),
  ]);

  // The node the node family shows: the route's key when GET /nodes lists
  // it and it is not the default; undefined is the default node.
  const defaultKey = $derived(nodeList[0]?.key);
  const nodeKey = $derived(
    route.family === "node" && route.nodeKey && route.nodeKey !== defaultKey && nodeList.some((n) => n.key === route.nodeKey) ? route.nodeKey : undefined,
  );
  const nodeOptions = $derived(nodeList.map((n) => ({ value: n.key, label: n.enabled ? n.label : t("shell.node_disabled", { label: n.label }) })));
  function selectNode(key: string) {
    location.hash = key === defaultKey ? ui.routes.status : routesForNode(ui.routes, key).status;
  }
  const controllerTabs = $derived([
    { key: "nodes", label: t("controller.tab.nodes"), href: ui.routes.controllerNodes },
    { key: "commission", label: t("controller.tab.commission"), href: ui.routes.controllerCommission },
    { key: "settings", label: t("controller.tab.settings"), href: ui.routes.controllerSettings },
  ]);

  const controllerPage = $derived(route.family === "controller" && route.page in controllerPages ? (route.page as keyof typeof controllerPages) : "nodes");

  async function probeNode() {
    try {
      await ui.client.status();
      nodeServed = true;
    } catch (err) {
      nodeServed = !(err instanceof ApiError && err.status === 404);
    }
    if (nodeServed) void node.loadStatus();
    try {
      nodeList = (await ui.client.nodes()).nodes;
    } catch {
      nodeList = [];
    }
  }

  async function probeOperation() {
    try {
      await ui.client.operationInfo();
      operationServed = true;
    } catch (err) {
      operationServed = !(err instanceof ApiError && (err.status === 404 || err.status === 503));
    }
  }

  let stopBridge = () => {};
  let media: MediaQueryList | null = null;
  const onScheme = (e: MediaQueryListEvent) => document.documentElement.classList.toggle("dark", e.matches);

  onMount(() => {
    window.addEventListener("hashchange", onHash);
    applySkin("ha");
    stopBridge = startHaBridge();
    if (!isEmbedded() && window.matchMedia) {
      media = window.matchMedia("(prefers-color-scheme: dark)");
      media.addEventListener("change", onScheme);
    }
    void probeNode();
    void probeOperation();
    void controller.loadStatus();
  });

  // Each family's stream opens once the family is known to be there.
  $effect(() => {
    if (nodeFamily) node.connect();
  });
  $effect(() => {
    if (controllerFamily) controller.connect();
  });

  onDestroy(() => {
    window.removeEventListener("hashchange", onHash);
    media?.removeEventListener("change", onScheme);
    stopBridge();
    node.close();
    controller.close();
  });

  const ControllerCurrent = $derived(controllerPages[controllerPage]);
</script>

<PageShell>
  <PageHeader
    title={family === "controller" ? t("controller.title") : family === "operation" ? t("operation.title") : t("node.title")}
    subtitle={family === "controller" ? t("controller.subtitle") : family === "operation" ? t("operation.subtitle") : t("node.subtitle")}
  >
    {#snippet actions()}
      {#if family === "node" && nodeList.length > 1}
        <Select
          class="w-auto"
          ariaLabel={t("shell.node")}
          value={nodeKey ?? defaultKey}
          onValueChange={selectNode}
          options={nodeOptions}
        />
      {/if}
      <Select
        class="w-auto"
        ariaLabel={t("shell.language")}
        value={locale}
        onValueChange={(v) => (locale = v === "de" ? "de" : "en")}
        options={[
          { value: "en", label: "English" },
          { value: "de", label: "Deutsch" },
        ]}
      />
    {/snippet}
  </PageHeader>
  {#if !probed && family === null}
    <LoadingState message={t("common.loading")} />
  {:else if family === null}
    <EmptyState message={t("shell.no_family")} />
  {:else}
    {#if familyTabs.length > 1}
      <Tabs items={familyTabs} active={family} variant="segmented" class="mb-3" ariaLabel={t("shell.families")} />
    {/if}
    {#if family === "node"}
      {#key nodeKey ?? ""}
        <NodeScope {nodeKey}>
          <NodeFamily route={route.family === "node" ? route : { family: "node", page: "status" }} pages={nodePages} />
        </NodeScope>
      {/key}
    {:else if family === "operation"}
      <Operation />
    {:else}
      <Tabs items={controllerTabs} active={route.nodeId ? "nodes" : controllerPage} />
      <div class="mt-4">
        {#if route.nodeId}
          {#key route.nodeId}
            <ControllerNode nodeId={route.nodeId} tab={route.tab} />
          {/key}
        {:else if controllerPage === "nodes"}
          <ControllerNodes tab={route.tab} />
        {:else}
          <ControllerCurrent />
        {/if}
      </div>
    {/if}
  {/if}
</PageShell>

<Toaster entries={toasts.entries} dismiss={toasts.dismiss} closeLabel={t("common.close")} />
<ConfirmDialog current={confirm.current} answer={confirm.answer} cancelLabel={t("common.cancel")} confirmLabel={t("common.confirm")} />
