<!-- The controller's settings: its fabric label, the stored Wi-Fi and
     Thread credentials (names only; a secret is entered once and never
     shown again) and the controller's identity. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Credential, CredentialRequest } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Input from "../../kit/Input.svelte";
  import Select from "../../kit/Select.svelte";
  import { vendorLabel } from "./format";

  const { t, toast, confirm, client, controller } = useFabricUI();

  let label = $state("");
  let kind = $state<Credential["kind"]>("wifi");
  let id = $state("");
  let ssid = $state("");
  let secret = $state("");
  let busy = $state(false);

  onMount(async () => {
    await controller.loadStatus();
    label = controller.status?.label ?? "";
    if (controller.status?.capabilities.credentials) void controller.loadCredentials();
  });

  async function saveLabel() {
    busy = true;
    try {
      await client.setFabricLabel(label.trim());
      toast.success(t("controller.settings.label_saved"));
      await controller.loadStatus();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function addCredential() {
    const req: CredentialRequest = { kind, id: id.trim() };
    if (kind === "wifi") {
      req.ssid = ssid.trim();
      req.passphrase = secret;
    } else {
      req.dataset_hex = secret.replace(/\s+/g, "");
    }
    busy = true;
    try {
      await client.setCredential(req);
      toast.success(t("controller.settings.credential_saved"));
      id = ssid = secret = "";
      await controller.loadCredentials();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function removeCredential(c: Credential) {
    const ok = await confirm({ title: t("controller.settings.credential_remove"), body: c.name, destructive: true, confirmLabel: t("common.remove") });
    if (!ok) return;
    try {
      await client.removeCredential(c.kind, c.id);
      await controller.loadCredentials();
    } catch (err) {
      toast.error(errorText(err));
    }
  }

  const kindOptions = $derived([
    { value: "wifi", label: t("controller.credential.kind_wifi") },
    { value: "thread", label: t("controller.credential.kind_thread") },
  ]);

  const columns: DataColumn<Credential>[] = $derived([
    { key: "name", label: t("controller.settings.col_name"), title: true, sortable: true, get: (c) => c.name },
    { key: "kind", label: t("controller.settings.col_kind"), sortable: true, get: (c) => c.kind },
    { key: "id", label: t("controller.settings.col_id"), sortable: true, get: (c) => c.id },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);

  const s = $derived(controller.status);
</script>

<div class="space-y-6">
  {#if controller.statusError}
    <ErrorState message={controller.statusError} onRetry={() => void controller.loadStatus()} />
  {/if}
  <Card class="p-4">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.settings.label")}</h3>
    <form class="flex flex-wrap items-center gap-2" onsubmit={(e) => { e.preventDefault(); void saveLabel(); }}>
      <Input class="max-w-xs" bind:value={label} maxlength={32} aria-label={t("controller.settings.label")} />
      <Button type="submit" disabled={busy}>{t("common.save")}</Button>
    </form>
    <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.settings.label_hint")}</p>
  </Card>

  <Card class="p-4" data-testid="credentials">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.settings.credentials")}</h3>
    {#if s && !s.capabilities.credentials}
      <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("controller.settings.no_store")}</p>
    {:else}
      {#if controller.credentialsError}
        <ErrorState message={controller.credentialsError} onRetry={() => void controller.loadCredentials()} />
      {:else}
        <DataTable rows={controller.credentials} {columns} rowKey={(c) => `${c.kind}:${c.id}`} emptyMessage={t("controller.settings.credentials_empty")}>
          {#snippet cell(c, col)}
            {#if col.key === "name"}
              {c.name}
            {:else if col.key === "kind"}
              <Badge variant="muted">{t(`controller.credential.kind_${c.kind}`)}</Badge>
            {:else if col.key === "id"}
              <span class="font-mono text-xs">{c.id}</span>
            {:else if col.key === "action"}
              <Button size="sm" variant="ghost" class="text-[var(--ha-error-color)]" onclick={() => void removeCredential(c)}>{t("common.remove")}</Button>
            {/if}
          {/snippet}
        </DataTable>
      {/if}
      <form class="mt-4 flex flex-wrap items-center gap-2" onsubmit={(e) => { e.preventDefault(); void addCredential(); }}>
        <Select class="w-auto" ariaLabel={t("controller.settings.col_kind")} value={kind} onValueChange={(v) => (kind = v as typeof kind)} options={kindOptions} />
        <Input class="w-32" bind:value={id} placeholder={t("controller.settings.col_id")} aria-label={t("controller.settings.col_id")} />
        {#if kind === "wifi"}
          <Input class="w-40" bind:value={ssid} placeholder={t("controller.settings.ssid")} aria-label={t("controller.settings.ssid")} />
          <Input class="w-48" type="password" autocomplete="new-password" bind:value={secret} placeholder={t("controller.settings.passphrase")} aria-label={t("controller.settings.passphrase")} />
        {:else}
          <Input class="w-72 font-mono" type="password" autocomplete="off" bind:value={secret} placeholder={t("controller.settings.dataset")} aria-label={t("controller.settings.dataset")} />
        {/if}
        <Button type="submit" disabled={busy || !id.trim()}>{t("controller.settings.credential_add")}</Button>
      </form>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.settings.write_only")}</p>
    {/if}
  </Card>

  {#if s}
    <Card class="p-4" data-testid="controller-info">
      <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.settings.info")}</h3>
      <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.fabric_id")}</dt><dd class="font-mono">0x{s.fabric_id_hex}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.compressed_fabric_id")}</dt><dd class="font-mono">{s.compressed_fabric_id_hex}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.node_id")}</dt><dd class="font-mono">0x{s.node_id_hex}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.vendor")}</dt><dd>{vendorLabel(s.vendor_name, s.vendor_id)}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.nodes")}</dt><dd>{s.node_count}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.settings.capabilities")}</dt>
        <dd class="flex flex-wrap gap-1">
          {#each (["ota", "credentials", "ble"] as const).filter((k) => s.capabilities[k]) as k (k)}<Badge variant="muted">{t(`controller.settings.cap.${k}`)}</Badge>{/each}
        </dd>
      </dl>
    </Card>
  {/if}
</div>
