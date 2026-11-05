<!-- The shell's dialog of the node selector's "Create node…" (contract
     0.6.0): the label the host creates the node with (POST /nodes); the
     host allocates everything else. The shell switches to the new node's
     selection page when onCreate resolves; a failure stays in the dialog. -->
<script lang="ts">
  import { useFabricUI } from "../lib/context";
  import { errorText } from "../lib/store.svelte";
  import Button from "../lib/kit/Button.svelte";
  import Input from "../lib/kit/Input.svelte";

  let { onCreate, onClose }: { onCreate: (label: string) => Promise<void>; onClose: () => void } = $props();

  const { t } = useFabricUI();

  let label = $state("");
  let busy = $state(false);
  let error = $state<string | null>(null);
  const valid = $derived(label.trim().length > 0 && label.trim().length <= 64);

  async function submit(e: SubmitEvent) {
    e.preventDefault();
    if (!valid || busy) return;
    busy = true;
    error = null;
    try {
      await onCreate(label.trim());
    } catch (err) {
      error = errorText(err);
    } finally {
      busy = false;
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" role="presentation">
  <div
    class="w-full max-w-md rounded-[var(--ha-radius-card)] bg-[var(--ha-card-background-color)] p-6 shadow-[var(--ha-elevation-modal)]"
    role="dialog"
    aria-modal="true"
    aria-labelledby="fui-node-create-title"
    data-testid="node-create-dialog"
  >
  <form onsubmit={submit}>
    <h2 id="fui-node-create-title" class="text-base font-semibold text-[var(--ha-primary-text-color)]">{t("shell.node_create.title")}</h2>
    <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t("shell.node_create.intro")}</p>
    <label class="mt-4 block text-sm font-medium text-[var(--ha-primary-text-color)]" for="fui-node-create-label">{t("shell.node_create.label")}</label>
    <!-- svelte-ignore a11y_autofocus -->
    <Input id="fui-node-create-label" class="mt-1" maxlength={64} autofocus placeholder={t("shell.node_create.placeholder")} bind:value={label} />
    {#if error}
      <p class="mt-2 text-sm text-[var(--ha-error-color)]" role="alert">{error}</p>
    {/if}
    <div class="mt-6 flex justify-end gap-2">
      <Button type="button" variant="outline" onclick={onClose}>{t("common.cancel")}</Button>
      <Button type="submit" disabled={!valid || busy}>{t("shell.node_create.submit")}</Button>
    </div>
  </form>
  </div>
</div>
