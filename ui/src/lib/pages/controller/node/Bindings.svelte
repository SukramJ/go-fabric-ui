<!-- The binding targets of every endpoint with a Binding cluster: add a
     unicast (node and endpoint) or group target, remove one, write the
     endpoint's list back. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { BindingTarget, ControllerNodeDetail, EndpointBindings } from "../../../api/client";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import Input from "../../../kit/Input.svelte";
  import LoadingState from "../../../kit/LoadingState.svelte";
  import { hexId, parseHexId } from "../format";

  let { nodeId, detail }: { nodeId: string; detail: ControllerNodeDetail } = $props();
  const { t, toast, client } = useFabricUI();

  let lists = $state<EndpointBindings[] | null>(null);
  let error = $state<string | null>(null);
  let drafts = $state<Record<number, { target: string; endpoint: string; cluster: string }>>({});

  const hasBinding = $derived(detail.endpoints.some((e) => e.clusters.some((c) => c.id === 0x001e)));

  async function load() {
    error = null;
    try {
      lists = (await client.nodeBindings(nodeId)).endpoints;
    } catch (err) {
      error = errorText(err);
    }
  }
  onMount(() => void load());

  function describe(b: BindingTarget): string {
    const where = b.group !== undefined ? t("controller.bindings.group", { group: b.group }) : `0x${b.node_id_hex ?? ""} · ${t("controller.endpoints.endpoint", { number: b.endpoint ?? 0 })}`;
    return b.cluster_hex ? `${where} · ${b.cluster_hex}` : where;
  }

  async function write(eb: EndpointBindings, bindings: BindingTarget[]) {
    try {
      const res = await client.setNodeBindings(nodeId, eb.endpoint, bindings.map(({ fabric_index: _f, ...b }) => b));
      if (res.status === 0) toast.success(t("controller.bindings.saved"));
      else toast.error(t("controller.attributes.write_status", { status: hexId(res.status) }));
      await load();
    } catch (err) {
      toast.error(errorText(err));
    }
  }

  async function add(eb: EndpointBindings) {
    const d = drafts[eb.endpoint] ?? { target: "", endpoint: "", cluster: "" };
    const b: BindingTarget = {};
    const target = d.target.trim();
    if (/^g:\d+$/i.test(target)) b.group = Number(target.slice(2));
    else if (/^(0x)?[0-9a-f]{1,16}$/i.test(target) && d.endpoint.trim()) {
      b.node_id_hex = target.replace(/^0x/i, "").toUpperCase().padStart(16, "0");
      b.endpoint = Number(d.endpoint);
    } else {
      toast.error(t("controller.bindings.invalid"));
      return;
    }
    if (d.cluster.trim()) {
      const c = parseHexId(d.cluster);
      if (c === null) {
        toast.error(t("controller.bindings.invalid"));
        return;
      }
      b.cluster_hex = hexId(c);
    }
    drafts = { ...drafts, [eb.endpoint]: { target: "", endpoint: "", cluster: "" } };
    await write(eb, [...eb.bindings, b]);
  }

  function draft(ep: number, key: "target" | "endpoint" | "cluster", value: string) {
    const d = drafts[ep] ?? { target: "", endpoint: "", cluster: "" };
    drafts = { ...drafts, [ep]: { ...d, [key]: value } };
  }
</script>

{#if !hasBinding}
  <EmptyState message={t("controller.bindings.none")} />
{:else if error}
  <ErrorState message={error} onRetry={() => void load()} />
{:else if !lists}
  <LoadingState message={t("common.loading")} />
{:else}
  <div class="space-y-4">
    {#each lists as eb (eb.endpoint)}
      <Card class="p-4" data-testid={`bindings-${eb.endpoint}`}>
        <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.endpoints.endpoint", { number: eb.endpoint })}</h3>
        {#if eb.bindings.length === 0}
          <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("controller.bindings.empty")}</p>
        {:else}
          <ul class="space-y-1">
            {#each eb.bindings as b, i (i)}
              <li class="flex items-center gap-2 text-sm">
                <span class="font-mono text-xs">{describe(b)}</span>
                <Button size="sm" variant="ghost" class="ml-auto text-[var(--ha-error-color)]" onclick={() => void write(eb, eb.bindings.filter((_, j) => j !== i))}>{t("common.remove")}</Button>
              </li>
            {/each}
          </ul>
        {/if}
        <div class="mt-3 flex flex-wrap items-center gap-2">
          <Input class="w-56 font-mono" placeholder={t("controller.bindings.target")} aria-label={t("controller.bindings.target")} value={drafts[eb.endpoint]?.target ?? ""} oninput={(e) => draft(eb.endpoint, "target", (e.currentTarget as HTMLInputElement).value)} />
          <Input class="w-24" placeholder={t("controller.bindings.endpoint")} aria-label={t("controller.bindings.endpoint")} inputmode="numeric" value={drafts[eb.endpoint]?.endpoint ?? ""} oninput={(e) => draft(eb.endpoint, "endpoint", (e.currentTarget as HTMLInputElement).value)} />
          <Input class="w-28 font-mono" placeholder={t("controller.bindings.cluster")} aria-label={t("controller.bindings.cluster")} value={drafts[eb.endpoint]?.cluster ?? ""} oninput={(e) => draft(eb.endpoint, "cluster", (e.currentTarget as HTMLInputElement).value)} />
          <Button variant="outline" onclick={() => void add(eb)}>{t("controller.bindings.add")}</Button>
        </div>
      </Card>
    {/each}
  </div>
{/if}
