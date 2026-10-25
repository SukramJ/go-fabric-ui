<!-- The standalone shell the Go module serves at "/": hash routing over the
     node pages, the header with their tabs, toasts and the confirm dialog,
     the theme bridge. The API is "./api/v1" relative to the page, so the
     shell works under any base path (an Ingress among them). -->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { createClient } from "../lib/api/client";
  import { createFabricUI, provideFabricUI } from "../lib/context";
  import { createTranslator, detectLocale } from "../lib/i18n";
  import { applySkin, isEmbedded, startHaBridge } from "../lib/theme/ha-bridge";
  import PageHeader from "../lib/kit/PageHeader.svelte";
  import PageShell from "../lib/kit/PageShell.svelte";
  import Select from "../lib/kit/Select.svelte";
  import Tabs from "../lib/kit/Tabs.svelte";
  import NodeStatus from "../lib/pages/node/Status.svelte";
  import NodePair from "../lib/pages/node/Pair.svelte";
  import NodeFabrics from "../lib/pages/node/Fabrics.svelte";
  import NodeSelection from "../lib/pages/node/Selection.svelte";
  import NodeEndpoints from "../lib/pages/node/Endpoints.svelte";
  import NodeDiagnostics from "../lib/pages/node/Diagnostics.svelte";
  import ConfirmDialog from "./ConfirmDialog.svelte";
  import Toaster from "./Toaster.svelte";
  import { createConfirm } from "./confirm.svelte";
  import { createToasts } from "./toast.svelte";

  let locale = $state(detectLocale());
  const t = createTranslator(() => locale);
  const toasts = createToasts();
  const confirm = createConfirm();
  const ui = provideFabricUI(createFabricUI({ client: createClient("./api/v1"), t, toast: toasts.toast, confirm: confirm.ask }));
  const node = ui.node;

  const pages = { status: NodeStatus, pair: NodePair, fabrics: NodeFabrics, selection: NodeSelection, endpoints: NodeEndpoints, diagnostics: NodeDiagnostics };
  type Page = keyof typeof pages;

  function pageOf(hash: string): Page {
    const key = hash.replace(/^#\/node\//, "").split(/[/?]/)[0];
    return key in pages ? (key as Page) : "status";
  }

  let page = $state<Page>(pageOf(location.hash));
  const onHash = () => (page = pageOf(location.hash));

  const tabs = $derived(
    (Object.keys(pages) as Page[])
      .filter((p) => p !== "selection" || node.status?.exposure)
      .map((p) => ({ key: p, label: t(`node.tab.${p}`), href: ui.routes[p] })),
  );

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
    void node.loadStatus();
    node.connect();
  });

  onDestroy(() => {
    window.removeEventListener("hashchange", onHash);
    media?.removeEventListener("change", onScheme);
    stopBridge();
    node.close();
  });

  const Current = $derived(pages[page]);
</script>

<PageShell>
  <PageHeader title={t("node.title")} subtitle={t("node.subtitle")}>
    {#snippet actions()}
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
  <Tabs items={tabs} active={page} />
  <div class="mt-4">
    <Current />
  </div>
</PageShell>

<Toaster entries={toasts.entries} dismiss={toasts.dismiss} closeLabel={t("common.close")} />
<ConfirmDialog current={confirm.current} answer={confirm.answer} cancelLabel={t("common.cancel")} confirmLabel={t("common.confirm")} />
