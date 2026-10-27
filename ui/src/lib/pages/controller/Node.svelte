<!-- One node: its record and properties, endpoints, a targeted attribute
     reader, ACL, bindings, a commissioning window for another
     administrator, ICD, the fabrics on it and its pending updates — one
     tab each. The page reloads the node when the stream names it. -->
<script lang="ts">
  import { untrack } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { ControllerNodeDetail } from "../../api/client";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import Tabs from "../../kit/Tabs.svelte";
  import { nodeName } from "./format";
  import Overview from "./node/Overview.svelte";
  import Endpoints from "./node/Endpoints.svelte";
  import Attributes from "./node/Attributes.svelte";
  import Acl from "./node/Acl.svelte";
  import Bindings from "./node/Bindings.svelte";
  import Window from "./node/Window.svelte";
  import Icd from "./node/Icd.svelte";
  import Fabrics from "./node/Fabrics.svelte";
  import Updates from "./node/Updates.svelte";

  let { nodeId, tab = "overview" }: { nodeId: string; tab?: string } = $props();

  const { t, client, controller, routes } = useFabricUI();

  const TABS = ["overview", "endpoints", "attributes", "acl", "bindings", "window", "icd", "fabrics", "updates"] as const;
  type Tab = (typeof TABS)[number];
  const isTab = (v: string): v is Tab => (TABS as readonly string[]).includes(v);

  // svelte-ignore state_referenced_locally
  let active = $state<Tab>(isTab(tab) ? tab : "overview");
  let detail = $state<ControllerNodeDetail | null>(null);
  let error = $state<string | null>(null);

  async function load() {
    error = null;
    try {
      detail = await client.controllerNode(nodeId);
    } catch (err) {
      error = errorText(err);
    }
  }

  $effect(() => {
    void nodeId;
    untrack(() => void load());
  });

  // An event about this node reloads it.
  $effect(() => {
    void controller.revision;
    if (untrack(() => controller.changed) === nodeId && untrack(() => detail) !== null) untrack(() => void load());
  });

  const items = $derived(TABS.map((k) => ({ key: k, label: t(`controller.node.tab.${k}`) })));
</script>

<div class="space-y-4">
  <a href={routes.controllerNodes} class="text-sm text-[var(--ha-primary-color)] hover:underline">← {t("controller.node.back")}</a>
  {#if error}
    <ErrorState message={error} onRetry={() => void load()} />
  {:else if !detail}
    <LoadingState message={t("common.loading")} />
  {:else}
    <h2 class="text-lg font-medium text-[var(--ha-primary-text-color)]" data-testid="node-title">{nodeName(detail.node)}</h2>
    <Tabs {items} {active} onSelect={(k) => (active = isTab(k) ? k : "overview")} ariaLabel={t("controller.node.tabs")} />
    <div class="mt-2">
      {#if active === "overview"}
        <Overview {detail} />
      {:else if active === "endpoints"}
        <Endpoints {detail} />
      {:else if active === "attributes"}
        <Attributes {nodeId} {detail} />
      {:else if active === "acl"}
        <Acl {nodeId} />
      {:else if active === "bindings"}
        <Bindings {nodeId} {detail} />
      {:else if active === "window"}
        <Window {nodeId} />
      {:else if active === "icd"}
        <Icd {nodeId} />
      {:else if active === "fabrics"}
        <Fabrics {nodeId} />
      {:else if active === "updates"}
        <Updates {nodeId} />
      {/if}
    </div>
  {/if}
</div>
