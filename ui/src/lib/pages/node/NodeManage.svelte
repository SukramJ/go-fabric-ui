<!-- A managed node's own card on its status page (contract 0.6.0): rename
     it (PUT /nodes/{key}/label) and delete it (DELETE /nodes/{key}) after
     two confirmations that name its label and its fabric count, since a
     deletion is final and every controller that paired the node loses it.
     Shown for a node the node list reports managed; the controls need
     admin rights (the token of token mode or Status.admin). Below them the
     ecosystem hint for several nodes: a static copy of
     home-assistant-matter-hub's bridge configuration warning,
     docs-site/docs/getting-started/bridge-configuration.md:28-33 at
     8bcefc2c (Alexa pairs reliably only on port 5540 and may refuse a
     bridge above about 80–100 devices), with its link. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import Input from "../../kit/Input.svelte";

  const { t, client, node, nodes, toast, confirm } = useFabricUI();

  const HINT_SOURCE =
    "https://github.com/RiDDiX/home-assistant-matter-hub/blob/8bcefc2c1b8dae6578389be6338471762684f923/docs-site/docs/getting-started/bridge-configuration.md";

  const key = client.nodeKey;
  const entry = $derived(nodes.find(key));
  const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);

  let label = $state("");
  let busy = $state(false);
  // The field follows the node's label until the operator edits it.
  let edited = $state(false);
  $effect(() => {
    if (!edited && entry) label = entry.label;
  });
  const trimmed = $derived(label.trim());
  const canRename = $derived(!!entry && trimmed.length > 0 && trimmed.length <= 64 && trimmed !== entry.label && !busy);

  onMount(() => {
    if (!nodes.loaded) void nodes.load();
  });

  async function rename(e: SubmitEvent) {
    e.preventDefault();
    if (!key || !canRename) return;
    busy = true;
    try {
      await nodes.rename(key, trimmed);
      edited = false;
      toast.success(t("node.manage.renamed", { label: trimmed }));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function remove() {
    if (!key || !entry) return;
    const name = entry.label;
    const count = node.status?.fabric_count ?? 0;
    const first = await confirm({
      title: t("node.manage.delete_title", { label: name }),
      body: t("node.manage.delete_body", { label: name, count }),
      destructive: true,
      confirmLabel: t("node.manage.delete"),
    });
    if (!first) return;
    const second = await confirm({
      title: t("node.manage.delete_again_title", { label: name }),
      body: t("node.manage.delete_again_body", { count }),
      destructive: true,
      confirmLabel: t("node.manage.delete_confirm"),
    });
    if (!second) return;
    busy = true;
    try {
      await nodes.remove(key);
      toast.success(t("node.manage.deleted", { label: name }));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }
</script>

{#if entry?.managed}
  <Card class="mt-4 p-4" data-testid="node-manage">
    <h2 class="mb-3 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.manage.title")}</h2>
    {#if isAdmin}
      <form class="flex flex-wrap items-end gap-2" onsubmit={rename}>
        <label class="min-w-[12rem] flex-1 text-sm text-[var(--ha-secondary-text-color)]">
          {t("node.manage.label")}
          <Input class="mt-1" maxlength={64} bind:value={label} oninput={() => (edited = true)} data-testid="node-manage-label" />
        </label>
        <Button type="submit" variant="outline" disabled={!canRename}>{t("node.manage.rename")}</Button>
        <Button type="button" variant="outline-destructive" disabled={busy} onclick={() => void remove()} data-testid="node-manage-delete">
          {t("node.manage.delete")}
        </Button>
      </form>
    {/if}
    <div class="mt-4 rounded border border-[var(--ha-divider-color)] px-3 py-2" data-testid="node-manage-hint">
      <p class="text-sm font-medium text-[var(--ha-primary-text-color)]">{t("node.manage.hint.title")}</p>
      <p class="mt-1 text-sm text-[var(--ha-secondary-text-color)]">{t("node.manage.hint.alexa")}</p>
      <a class="mt-1 inline-block text-sm text-[var(--ha-primary-color)] underline" href={HINT_SOURCE} target="_blank" rel="noopener noreferrer">
        {t("node.manage.hint.source")}
      </a>
    </div>
  </Card>
{/if}
