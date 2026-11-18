<!-- A targeted attribute reader: one concrete path at a time (no
     inventory), the value as JSON; an administrator can write the value
     back. -->
<script lang="ts">
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { AttributeValue, ControllerNodeDetail } from "../../../api/client";
  import { jsonText, stringifyJson, type JsonText } from "../../../api/json";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import Input from "../../../kit/Input.svelte";
  import Select from "../../../kit/Select.svelte";
  import { hexId, parseHexId } from "../format";

  let { nodeId, detail }: { nodeId: string; detail: ControllerNodeDetail } = $props();
  const { t, toast, client } = useFabricUI();

  // svelte-ignore state_referenced_locally
  let endpoint = $state(String(detail.endpoints[0]?.endpoint ?? 0));
  let cluster = $state("");
  let attribute = $state("0x0000");
  let result = $state<AttributeValue | null>(null);
  // Whether the last read answered: a read the node answered with neither
  // value nor status says so instead of showing nothing.
  let answered = $state(false);
  let edit = $state("");
  let busy = $state(false);
  // Counts the reads and the path's changes: a read answered after the
  // path changed is dropped.
  let readSeq = 0;
  const hasValue = $derived(result !== null && result.status === undefined && result.value !== undefined);

  // A value read from one path is never offered for writing to another:
  // a change of the path drops the result (UR-P1-ui-161).
  function clearResult() {
    readSeq++;
    result = null;
    answered = false;
    edit = "";
  }

  function setEndpoint(v: string) {
    endpoint = v;
    // The cluster of one endpoint may not exist on another.
    cluster = "";
    clearResult();
  }

  function setCluster(v: string) {
    cluster = v;
    clearResult();
  }

  const endpointOptions = $derived(detail.endpoints.map((e) => ({ value: String(e.endpoint), label: t("controller.endpoints.endpoint", { number: e.endpoint }) })));
  const clusterOptions = $derived(
    (detail.endpoints.find((e) => String(e.endpoint) === endpoint)?.clusters ?? []).map((c) => ({ value: c.id_hex, label: `${c.name || c.id_hex} (${c.id_hex})` })),
  );

  function path() {
    const c = parseHexId(cluster);
    const a = parseHexId(attribute);
    if (c === null || a === null) return null;
    return { endpoint: Number(endpoint), cluster_hex: hexId(c), attribute_hex: hexId(a) };
  }

  async function read() {
    const p = path();
    if (!p) {
      toast.error(t("controller.attributes.invalid_path"));
      return;
    }
    busy = true;
    const seq = ++readSeq;
    try {
      const res = await client.readAttributes(nodeId, [p]);
      if (seq !== readSeq) return;
      result = res.values[0] ?? null;
      answered = true;
      // The value's integers beyond 2^53 are kept exact (a node id in an
      // ACL subject): shown and seeded as the node sent them.
      edit = result && result.status === undefined && result.value !== undefined ? stringifyJson(result.value) : "";
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function write() {
    const p = path();
    if (!p) return;
    // The text is sent as typed, once it is JSON: parsing it to a JS
    // value would round an integer beyond 2^53.
    let value: JsonText;
    try {
      value = jsonText(edit);
    } catch {
      toast.error(t("controller.attributes.invalid_json"));
      return;
    }
    busy = true;
    try {
      const res = await client.writeAttribute(nodeId, { ...p, value });
      if (res.status === 0) toast.success(t("controller.attributes.written"));
      else toast.error(t("controller.attributes.write_status", { status: hexId(res.status) }));
      await read();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }
</script>

<Card class="p-4">
  <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.attributes.hint")}</p>
  <form class="flex flex-wrap items-center gap-2" onsubmit={(e) => { e.preventDefault(); void read(); }}>
    <Select class="w-auto" ariaLabel={t("controller.attributes.endpoint")} value={endpoint} onValueChange={setEndpoint} options={endpointOptions} />
    <Select class="w-auto" ariaLabel={t("controller.attributes.cluster")} value={cluster} onValueChange={setCluster} options={clusterOptions} />
    <Input class="w-28 font-mono" bind:value={attribute} oninput={clearResult} aria-label={t("controller.attributes.attribute")} placeholder="0x0000" />
    <Button type="submit" disabled={busy || !cluster}>{t("controller.attributes.read")}</Button>
  </form>
  {#if answered}
    <div class="mt-4" data-testid="attribute-result">
      {#if result?.status !== undefined}
        <p class="text-sm text-[var(--ha-error-color)]">{t("controller.attributes.status", { status: hexId(result.status) })}</p>
      {:else if !result || !hasValue}
        <p class="text-sm text-[var(--ha-secondary-text-color)]" data-testid="attribute-no-answer">{t("controller.attributes.no_answer")}</p>
      {:else}
        <pre class="overflow-x-auto rounded bg-[var(--ha-secondary-background-color)] p-2 font-mono text-xs">{stringifyJson(result.value, 2)}</pre>
        <div class="mt-2 flex flex-wrap items-center gap-2">
          <Input class="max-w-md font-mono" bind:value={edit} aria-label={t("controller.attributes.value")} />
          <Button variant="outline" disabled={busy} onclick={() => void write()}>{t("controller.attributes.write")}</Button>
        </div>
      {/if}
    </div>
  {/if}
</Card>
