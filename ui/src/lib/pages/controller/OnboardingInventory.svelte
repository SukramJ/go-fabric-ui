<!-- The controller's onboarding inventory (contract 0.13.0, issue #68):
     one row per commissioned node — name, vendor, product, serial number,
     how its onboarding code came to be stored — the code shown on request,
     labels printed, and a code attached to a node (typed, scanned with the
     camera, or read from a photograph in the browser). The codes are
     secrets: the entries live in this component only, fetched with
     cache "no-store", and never reach storage, the location, a toast, the
     console or an error text. The page reloads on node_added and
     node_removed and after its own writes. A host that does not serve the
     inventory (capabilities.onboarding_codes false, 501) shows an empty
     state; a refusal (403) says that admin rights are needed. -->
<script lang="ts">
  import { flushSync, onMount, tick } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import { ApiError, type OnboardingInventoryEntry } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Input from "../../kit/Input.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import { formatDate } from "../../kit/dates";
  import { qrSvgOrNull } from "../../qr";
  import { nodeName, vendorLabel } from "./format";
  import QrScan from "./QrScan.svelte";

  type Entry = OnboardingInventoryEntry;
  const { t, toast, confirm, client, controller, routes } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let notServed = $state(false);
  let denied = $state(false);
  let entries = $state<Entry[]>([]);
  // The rows whose code is shown, by node id.
  let shown = $state<Set<string>>(new Set());
  let busy = $state<Set<string>>(new Set());
  // The print block exists only while the browser prints this page
  // (beforeprint to afterprint, U12-R-02).
  let printing = $state(false);
  let attachCard = $state<HTMLElement | null>(null);

  // The attach dialog: the node, the text field, a refusal under it.
  let attachFor = $state<Entry | null>(null);
  let code = $state("");
  let sending = $state(false);
  let refusal = $state<{ title: string; detail?: string } | null>(null);
  let photoReading = $state(false);
  let photoNone = $state(false);

  // The status says so before the route is asked.
  const served = $derived(controller.status?.capabilities.onboarding_codes !== false);
  const notAdmin = $derived(controller.status !== null && !controller.isAdmin);

  let seq = 0;
  async function load(quiet = false) {
    const mine = ++seq;
    if (!served || notAdmin) {
      loading = false;
      return;
    }
    if (!quiet) loading = true;
    error = null;
    try {
      const inv = await client.onboardingInventory();
      if (mine !== seq) return;
      entries = inv.entries;
      notServed = denied = false;
    } catch (err) {
      if (mine !== seq) return;
      if (err instanceof ApiError && err.status === 501) notServed = true;
      else if (err instanceof ApiError && err.status === 403) denied = true;
      else error = errorText(err);
    } finally {
      if (mine === seq) loading = false;
    }
  }

  // Only the Print button prints labels (owner, 2026-11-22): it arms the
  // print, and beforeprint draws them (synchronously, before the browser
  // lays out the sheet) and scopes the print rule to this print only when
  // armed and there is something to print. A print from the browser's
  // menu prints the page as it is, without codes. afterprint (and the
  // page's end) disarms and takes labels and class away.
  let armed = false;
  function beforePrint() {
    if (!armed || printable.length === 0) return;
    flushSync(() => (printing = true));
    document.documentElement.classList.add("fui-print-labels");
  }
  function afterPrint() {
    armed = false;
    document.documentElement.classList.remove("fui-print-labels");
    printing = false;
  }

  onMount(() => {
    void load();
    window.addEventListener("beforeprint", beforePrint);
    window.addEventListener("afterprint", afterPrint);
    return () => {
      window.removeEventListener("beforeprint", beforePrint);
      window.removeEventListener("afterprint", afterPrint);
      afterPrint();
    };
  });

  // Moves the label block to <body>, where the print rule finds it
  // whatever the host's layout around the page.
  function toBody(node: HTMLElement) {
    document.body.appendChild(node);
    return { destroy: () => node.remove() };
  }

  // A node was commissioned or removed: reload in place. The first run
  // only records the epoch.
  let seenEpoch = controller.nodeSetEpoch;
  $effect(() => {
    const epoch = controller.nodeSetEpoch;
    if (epoch === seenEpoch) return;
    seenEpoch = epoch;
    void load(true);
  });

  const name = (e: Entry) => nodeName(e);
  const time = (iso: string | undefined) => (iso ? formatDate(t, iso) : "");
  const hex4 = (v: number) => `0x${v.toString(16).toUpperCase().padStart(4, "0")}`;

  function toggle(id: string) {
    const next = new Set(shown);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    shown = next;
  }

  // The clipboard API is unavailable outside a secure context: the
  // refusal surfaces, as on the pairing page.
  async function copy(value: string) {
    try {
      await navigator.clipboard.writeText(value);
      toast.success(t("node.pair.copied"));
    } catch {
      toast.error(t("node.pair.copy_failed"));
    }
  }

  // The dialog opens above the table: the focus moves to its field
  // (U12-R-08).
  async function openAttach(e: Entry) {
    attachFor = e;
    code = "";
    refusal = null;
    photoNone = false;
    await tick();
    const input = attachCard?.querySelector("input[data-testid='attach-code']") as HTMLInputElement | null;
    input?.focus();
    attachCard?.scrollIntoView?.({ block: "nearest" });
  }

  function closeAttach() {
    attachFor = null;
    code = "";
    refusal = null;
    photoNone = false;
  }

  // The photograph is read in the browser by the scan chunk and uploaded
  // nowhere; the input is cleared so the same file can be chosen again.
  async function readPhoto(ev: Event) {
    const input = ev.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    photoReading = true;
    photoNone = false;
    try {
      const scan = await import("../../scan");
      const text = await scan.decodeImage(file);
      if (text) code = text;
      else photoNone = true;
    } catch {
      photoNone = true;
    } finally {
      photoReading = false;
    }
  }

  // A refusal stays in the dialog: the title by problem code and, for a
  // mismatch, the sentence by field in the page's language; for an
  // invalid code the server's reason (go-fabric's, English).
  function refusalOf(err: unknown): { title: string; detail?: string } {
    const p = err instanceof ApiError ? err.problem : undefined;
    if (p?.code === "onboarding_code_invalid") return { title: t("controller.inventory.refused_invalid"), detail: p.detail };
    if (p?.code === "onboarding_code_mismatch" && p.field && p.code_value !== undefined && p.node_value !== undefined) {
      const fmt = p.field === "discriminator" ? String : hex4;
      return {
        title: t("controller.inventory.refused_mismatch"),
        detail: t(`controller.inventory.mismatch_${p.field}`, { code: fmt(p.code_value), node: fmt(p.node_value) }),
      };
    }
    return { title: errorText(err) };
  }

  async function attach() {
    const e = attachFor;
    const text = code.trim();
    if (!e || !text) return;
    sending = true;
    refusal = null;
    try {
      const stored = await client.setOnboardingCode(e.node_id_hex, text);
      entries = entries.map((x) => (x.node_id_hex === stored.node_id_hex ? stored : x));
      closeAttach();
      toast.success(t("controller.inventory.attached"));
    } catch (err) {
      refusal = refusalOf(err);
    } finally {
      sending = false;
    }
  }

  async function remove(e: Entry) {
    const ok = await confirm({
      title: t("controller.inventory.remove"),
      body: t("controller.inventory.remove_confirm", { name: name(e) }),
      destructive: true,
      confirmLabel: t("controller.inventory.remove"),
    });
    if (!ok) return;
    busy = new Set([...busy, e.node_id_hex]);
    try {
      await client.clearOnboardingCode(e.node_id_hex);
      toast.success(t("controller.inventory.removed"));
      const next = new Set(shown);
      next.delete(e.node_id_hex);
      shown = next;
      await load(true);
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      const next = new Set(busy);
      next.delete(e.node_id_hex);
      busy = next;
    }
  }

  const printable = $derived(entries.filter((e) => e.has_code));

  // The button arms the print and starts the browser's; beforeprint draws
  // the labels.
  function print() {
    armed = true;
    window.print();
  }

  const columns: DataColumn<Entry>[] = $derived([
    { key: "name", label: t("controller.inventory.col_name"), sortable: true, title: true, align: "left", get: (e) => name(e) },
    { key: "vendor", label: t("controller.inventory.col_vendor"), sortable: true, get: (e) => vendorLabel(e.vendor_name, e.vendor_id) },
    { key: "product", label: t("controller.inventory.col_product"), sortable: true, get: (e) => e.product_name ?? "" },
    { key: "serial", label: t("controller.inventory.col_serial"), get: (e) => e.serial_number ?? "" },
    { key: "code", label: t("controller.inventory.col_code"), sortable: true, get: (e) => (e.has_code ? (e.source ?? "") : "") },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);
</script>

{#snippet cell(e: Entry, col: DataColumn<Entry>)}
  {#if col.key === "name"}
    <a class="text-[var(--fui-primary-text)] hover:underline" href={routes.controllerNode + e.node_id_hex}>{name(e)}</a>
  {:else if col.key === "vendor"}
    {vendorLabel(e.vendor_name, e.vendor_id) || "—"}
  {:else if col.key === "product"}
    {e.product_name ?? (e.product_id !== undefined ? hex4(e.product_id) : "—")}
  {:else if col.key === "serial"}
    <span class="font-mono text-xs">{e.serial_number ?? "—"}</span>
  {:else if col.key === "code"}
    <div class="space-y-1" data-testid={`code-${e.node_id_hex}`}>
      {#if !e.has_code}
        <Badge variant="muted">{t("controller.inventory.no_code")}</Badge>
      {:else}
        <div class="flex flex-wrap items-center gap-1">
          {#if e.source === "paired"}
            <Badge variant="success">{t("controller.inventory.source_paired")}</Badge>
          {:else}
            <Badge variant="warning" title={t("controller.inventory.source_attached_hint")}>{t("controller.inventory.source_attached")}</Badge>
            {#if e.checks && (e.checks.vendor || e.checks.product || e.checks.discriminator)}
              {#if e.checks.vendor}<Badge variant="muted">{t("controller.inventory.check_vendor")}</Badge>{/if}
              {#if e.checks.product}<Badge variant="muted">{t("controller.inventory.check_product")}</Badge>{/if}
              {#if e.checks.discriminator}<Badge variant="muted">{t("controller.inventory.check_discriminator")}</Badge>{/if}
            {:else}
              <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.checks_none")}</span>
            {/if}
          {/if}
        </div>
        {#if e.stored_at}
          <div class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.stored_at", { when: time(e.stored_at) })}</div>
        {/if}
        <Button type="button" variant="ghost" size="sm" onclick={() => toggle(e.node_id_hex)} data-testid={`show-${e.node_id_hex}`}>
          {shown.has(e.node_id_hex) ? t("controller.inventory.hide") : t("controller.inventory.show")}
        </Button>
        {#if shown.has(e.node_id_hex)}
          <div class="space-y-2 pt-1" data-testid={`shown-${e.node_id_hex}`}>
            {#if e.qr_code}
              {@const svg = qrSvgOrNull(e.qr_code, { size: 160 })}
              {#if svg}
                <!-- The SVG is drawn from the QR matrix alone (qr.ts). -->
                <div class="inline-block" data-testid="qr-image">{@html svg}</div>
              {/if}
            {:else}
              <p class="max-w-xs text-xs text-[var(--ha-secondary-text-color)]" data-testid="no-qr">{t("controller.inventory.no_qr")}</p>
            {/if}
            {#if e.manual_code}
              <div class="flex items-center gap-2">
                <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.manual_code")}</span>
                <span class="font-mono text-sm" data-testid="manual-code">{e.manual_code}</span>
                <Button type="button" variant="outline" size="sm" onclick={() => void copy(e.manual_code ?? "")}>{t("common.copy")}</Button>
              </div>
            {/if}
          </div>
        {/if}
      {/if}
    </div>
  {:else if col.key === "action"}
    <div class="flex flex-wrap justify-end gap-1">
      {#if e.has_code}
        <Button type="button" variant="outline" size="sm" disabled={busy.has(e.node_id_hex)} onclick={() => void openAttach(e)}>{t("controller.inventory.replace")}</Button>
        <Button type="button" variant="outline-destructive" size="sm" disabled={busy.has(e.node_id_hex)} onclick={() => void remove(e)}>{t("controller.inventory.remove")}</Button>
      {:else}
        <Button type="button" variant="outline" size="sm" disabled={busy.has(e.node_id_hex)} onclick={() => void openAttach(e)}>{t("controller.inventory.attach")}</Button>
      {/if}
    </div>
  {/if}
{/snippet}

<div class="space-y-4">
  <div class="flex flex-wrap items-center justify-between gap-2">
    <h2 class="text-lg font-semibold text-[var(--ha-primary-text-color)]">{t("controller.inventory.title")}</h2>
    {#if served && !notServed && !notAdmin && !denied && !loading && !error}
      <div class="flex items-center gap-2">
        {#if printable.length > 0}
          <Button variant="outline" onclick={print}>{t("controller.inventory.print")}</Button>
        {:else}
          <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.print_empty")}</span>
        {/if}
        <Button variant="outline" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
      </div>
    {/if}
  </div>

  {#if !served || notServed}
    <EmptyState message={t("controller.inventory.not_served")} />
  {:else if notAdmin || denied}
    <div data-testid="denied">
      <EmptyState icon="mdi:cancel" message={t("controller.inventory.admin_required")} />
    </div>
  {:else if loading}
    <LoadingState />
  {:else if error}
    <ErrorState message={error} onRetry={() => void load()} />
  {:else}
    <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("controller.inventory.intro")}</p>

    {#if attachFor}
      <Card class="space-y-3 p-4">
        <div role="group" aria-label={t("controller.inventory.attach_title", { name: name(attachFor) })} class="space-y-3" data-testid="attach-dialog" bind:this={attachCard}>
          <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.inventory.attach_title", { name: name(attachFor) })}</h3>
          <p class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.attach_hint")}</p>
          {#if attachFor.has_code && attachFor.source === "paired"}
            <!-- U12-F-02: the device accepted this code; its replacement is unproven. -->
            <p class="text-xs text-[var(--fui-warning-text)]" data-testid="replace-paired">{t("controller.inventory.replace_paired_hint")}</p>
          {/if}
          <form
            class="space-y-3"
            onsubmit={(ev) => {
              ev.preventDefault();
              void attach();
            }}
          >
            <label class="block space-y-1">
              <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.code_label")}</span>
              <Input bind:value={code} autocomplete="off" spellcheck={false} autocapitalize="off" placeholder="MT:… / 3497-011-2332" data-testid="attach-code" />
            </label>
            <div class="flex flex-wrap items-center gap-2">
              <QrScan onCode={(payload) => (code = payload)} disabled={sending} insecureMessage={t("controller.inventory.camera_note")} />
              <label class="inline-flex cursor-pointer items-center gap-2 text-sm text-[var(--fui-primary-text)] underline-offset-2 hover:underline">
                <input type="file" accept="image/*" class="sr-only" onchange={(ev) => void readPhoto(ev)} disabled={sending || photoReading} data-testid="attach-photo" />
                {t("controller.inventory.photo")}
              </label>
            </div>
            {#if photoNone}
              <p class="text-xs text-[var(--fui-warning-text)]" data-testid="photo-none">{t("controller.inventory.photo_none")}</p>
            {/if}
            {#if refusal}
              <div class="text-sm" role="alert" data-testid="attach-refusal">
                <p class="font-medium text-[var(--fui-error-text)]">{refusal.title}</p>
                {#if refusal.detail}<p class="text-xs text-[var(--ha-secondary-text-color)]">{refusal.detail}</p>{/if}
              </div>
            {/if}
            <div class="flex gap-2">
              <Button type="submit" disabled={sending || code.trim() === ""}>{t("controller.inventory.attach")}</Button>
              <Button type="button" variant="outline" onclick={closeAttach}>{t("common.cancel")}</Button>
            </div>
          </form>
        </div>
      </Card>
    {/if}

    <Card class="p-4">
      <DataTable
        rows={entries}
        {columns}
        rowKey={(e) => e.node_id_hex}
        {cell}
        emptyMessage={t("controller.inventory.empty")}
      />
    </Card>
  {/if}
</div>

{#if printing}
  <div class="print-labels" data-testid="print-labels" use:toBody>
    <!-- Keyed by position: a server may list a node id twice. -->
    {#each printable as e, i (i)}
      {@const svg = e.qr_code ? qrSvgOrNull(e.qr_code, { size: 140 }) : null}
      <div class="print-label">
        <div class="font-semibold">{name(e)}</div>
        {#if svg}<div>{@html svg}</div>{/if}
        {#if e.manual_code}<div class="font-mono">{e.manual_code}</div>{/if}
      </div>
    {/each}
  </div>
{/if}
