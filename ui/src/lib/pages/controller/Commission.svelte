<!-- Commissioning: a QR payload or a manual pairing code, or a device a
     discovery found and its passcode; the stored credential that puts a
     Wi-Fi or Thread device on its network; the progress from the stream
     and the result. A camera scan fills the code (QrScan, contract 0.5.0;
     the decoder is a chunk loaded on demand). The code and the passcode go
     to the admin route only and are cleared after the call. The steps the
     controller reports come under the progress as they start and end, and
     a collapsed block sends a Terms and Conditions acknowledgement — the
     two values chip-tool's --tc-acknowledgements-version and
     --tc-acknowledgements send (contract 0.7.0). The network is an
     explicit input: the host's default credential, or the only one, is
     preselected; while none is picked on a host with a Bluetooth LE
     central a hint says BLE needs one, and under the progress the
     commissioning says whether BLE was used and why not (contract
     0.7.1). A device that requires terms answers 409 terms_required: the
     block opens with the device's version filled in, the acknowledgements
     field focused and the deadline as a date, and one click sends the
     same commissioning again with the acknowledgement (contract 0.8.0);
     the code or passcode is held for that retry only. -->
<script lang="ts">
  import { onMount, tick } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import { ApiError, type CommissionRequest, type ControllerNodeDetail, type Discovered } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import Input from "../../kit/Input.svelte";
  import Select from "../../kit/Select.svelte";
  import Spinner from "../../kit/Spinner.svelte";
  import { nodeName, vendorLabel } from "./format";
  import QrScan from "./QrScan.svelte";

  const { t, toast, client, controller, routes } = useFabricUI();

  let code = $state("");
  let network = $state("");
  let networkOnly = $state(false);
  // The Terms and Conditions acknowledgement, sent when both are set.
  let tcVersion = $state("");
  let tcAcknowledgements = $state("");
  let termsOpen = $state(false);
  // A device's terms requirement from a 409 terms_required, and the
  // request it refused, held for the one retry with the acknowledgement.
  let termsRequired = $state<{ version: number; deadline: number | null } | null>(null);
  let retry = $state<CommissionRequest | null>(null);
  let running = $state(false);
  let result = $state<ControllerNodeDetail | null>(null);
  let failure = $state<string | null>(null);

  let devices = $state<Discovered[]>([]);
  let searching = $state(false);
  let searched = $state(false);
  let passcodes = $state<Record<string, string>>({});

  onMount(() => {
    controller.resetProgress();
    if (controller.status?.capabilities.credentials) void controller.loadCredentials().then(preselect);
  });

  // The host's default entry, else the only one, unless the operator has
  // picked already; "No network credential" stays a choice.
  function preselect() {
    if (network !== "") return;
    const list = controller.credentials;
    const pick = list.find((c) => c.default) ?? (list.length === 1 ? list[0] : undefined);
    if (pick) network = `${pick.kind}:${pick.id}`;
  }

  const networkOptions = $derived([
    { value: "", label: t("controller.commission.network_none") },
    ...controller.credentials.map((c) => ({
      value: `${c.kind}:${c.id}`,
      label: `${t(`controller.credential.kind_${c.kind}`)} · ${c.name}${c.default ? ` (${t("controller.credential.default")})` : ""}`,
    })),
  ]);
  const bleHint = $derived(!!controller.status?.capabilities.ble && network === "" && !networkOnly && controller.credentials.length > 0);

  function withNetwork(req: CommissionRequest): CommissionRequest {
    const [kind, ...rest] = network.split(":");
    const id = rest.join(":");
    if (kind === "wifi") req.wifi_id = id;
    if (kind === "thread") req.thread_id = id;
    if (networkOnly) req.network_only = true;
    const tc = terms();
    if (tc) req.terms_and_conditions = tc;
    return req;
  }

  // The acknowledgement when both fields hold a 16-bit integer; undefined
  // otherwise, and then nothing is sent.
  function terms(): { version: number; acknowledgements: number } | undefined {
    const v = tcVersion.trim();
    const a = tcAcknowledgements.trim();
    if (v === "" || a === "") return undefined;
    const version = Number(v);
    const acknowledgements = Number(a);
    const ok = (n: number) => Number.isInteger(n) && n >= 0 && n <= 0xffff;
    return ok(version) && ok(acknowledgements) ? { version, acknowledgements } : undefined;
  }
  const termsInvalid = $derived((tcVersion.trim() !== "" || tcAcknowledgements.trim() !== "") && !terms());

  // The Matter epoch (2000-01-01T00:00:00Z) in Unix seconds: a terms
  // deadline counts from it.
  const MATTER_EPOCH = 946684800;
  const deadlineText = (s: number) => new Date((MATTER_EPOCH + s) * 1000).toLocaleDateString();

  async function commission(req: CommissionRequest) {
    running = true;
    result = null;
    failure = null;
    retry = null;
    termsRequired = null;
    try {
      // A copy: the request held for a terms retry stays without the
      // network and terms of this attempt, which the retry sets anew.
      result = await client.commission(withNetwork({ ...req }));
      toast.success(t("controller.commission.done", { name: nodeName(result.node) }));
      void controller.loadNodes();
    } catch (err) {
      if (err instanceof ApiError && err.problem?.code === "terms_required") {
        await requireTerms(req, err.problem.version ?? 0, err.problem.deadline ?? null);
      } else {
        failure = errorText(err);
      }
    } finally {
      running = false;
    }
  }

  // The device requires terms: the block opens with its version, the
  // acknowledgements field takes the focus, and the request waits for
  // the retry.
  async function requireTerms(req: CommissionRequest, version: number, deadline: number | null) {
    termsRequired = { version, deadline };
    retry = req;
    tcVersion = String(version);
    termsOpen = true;
    failure = t("controller.commission.terms_required", { version });
    await tick();
    document.getElementById("commission-terms-acknowledgements")?.focus();
  }

  async function sendTerms() {
    const req = retry;
    if (!req) return;
    if (!terms()) {
      toast.error(t("controller.commission.terms_invalid"));
      return;
    }
    await commission(req);
  }

  async function byCode() {
    const c = code.trim();
    if (!c) return;
    if (termsInvalid) {
      toast.error(t("controller.commission.terms_invalid"));
      return;
    }
    await commission({ code: c });
    code = "";
  }

  async function search() {
    searching = true;
    try {
      devices = (await client.discover({ timeout: 5 })).devices;
      searched = true;
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      searching = false;
    }
  }

  async function pair(d: Discovered) {
    const pin = Number((passcodes[d.instance_name] ?? "").trim());
    if (!Number.isInteger(pin) || pin < 1 || pin > 99999998) {
      toast.error(t("controller.commission.passcode_invalid"));
      return;
    }
    if (termsInvalid) {
      toast.error(t("controller.commission.terms_invalid"));
      return;
    }
    await commission({ discovered: { instance_name: d.instance_name, long_discriminator: d.long_discriminator }, passcode: pin });
    passcodes = { ...passcodes, [d.instance_name]: "" };
    devices = devices.filter((x) => x.instance_name !== d.instance_name);
  }

  const columns: DataColumn<Discovered>[] = $derived([
    { key: "device", label: t("controller.commission.col_device"), title: true, get: (d) => d.device_name || d.instance_name },
    { key: "vendor", label: t("controller.commission.col_vendor"), get: (d) => vendorLabel(d.vendor_name, d.vendor_id) },
    { key: "discriminator", label: t("controller.commission.col_discriminator"), numeric: true, get: (d) => d.long_discriminator },
    { key: "addresses", label: t("controller.commission.col_addresses"), get: (d) => d.addresses.join(" ") },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);

  const phase = $derived(controller.progress.phase);
  const steps = $derived(controller.progress.steps);
  const ble = $derived(controller.progress.ble);
  const stepVariant = (p: string): "success" | "danger" | "muted" | "default" =>
    p === "succeeded" ? "success" : p === "failed" ? "danger" : p === "skipped" ? "muted" : "default";
</script>

<div class="space-y-6">
  <Card class="p-4">
    <h3 class="mb-1 font-medium text-[var(--ha-primary-text-color)]">{t("controller.commission.by_code")}</h3>
    <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.commission.by_code_hint")}</p>
    <form class="flex flex-wrap items-center gap-3" onsubmit={(e) => { e.preventDefault(); void byCode(); }}>
      <Input class="max-w-sm font-mono" bind:value={code} placeholder={t("controller.commission.code_placeholder")} aria-label={t("controller.commission.code")} autocomplete="off" />
      <Button type="submit" disabled={running || !code.trim()}>{t("controller.commission.start")}</Button>
      <QrScan disabled={running} onCode={(payload) => (code = payload)} />
    </form>
    <div class="mt-3 flex flex-wrap items-center gap-3">
      <Select class="w-auto" ariaLabel={t("controller.commission.network")} value={network} onValueChange={(v) => (network = v as typeof network)} options={networkOptions} />
      <label class="inline-flex items-center gap-2 text-sm text-[var(--ha-secondary-text-color)]">
        <input type="checkbox" bind:checked={networkOnly} />
        {t("controller.commission.network_only")}
      </label>
    </div>
    {#if bleHint}
      <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]" data-testid="commission-ble-hint">{t("controller.commission.ble_hint")}</p>
    {/if}
    <details class="mt-3 text-sm" data-testid="commission-terms" bind:open={termsOpen}>
      <summary class="cursor-pointer text-[var(--ha-secondary-text-color)]">{t("controller.commission.terms")}</summary>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.commission.terms_hint")}</p>
      <div class="mt-2 flex flex-wrap items-center gap-3">
        <label class="text-xs text-[var(--ha-secondary-text-color)]">
          {t("controller.commission.terms_version")}
          <Input class="mt-1 w-28 font-mono" inputmode="numeric" autocomplete="off" bind:value={tcVersion} data-testid="commission-terms-version" />
        </label>
        <label class="text-xs text-[var(--ha-secondary-text-color)]">
          {t("controller.commission.terms_acknowledgements")}
          <Input class="mt-1 w-28 font-mono" inputmode="numeric" autocomplete="off" bind:value={tcAcknowledgements} id="commission-terms-acknowledgements" data-testid="commission-terms-acknowledgements" />
        </label>
        {#if retry}
          <Button size="sm" class="self-end" disabled={running} onclick={() => void sendTerms()} data-testid="commission-terms-retry">{t("controller.commission.terms_retry")}</Button>
        {/if}
      </div>
      {#if termsRequired && termsRequired.deadline !== null}
        <p class="mt-2 text-xs text-[var(--ha-primary-text-color)]" data-testid="commission-terms-deadline">
          {t("controller.commission.terms_deadline", { date: deadlineText(termsRequired.deadline) })}
        </p>
      {/if}
      {#if termsInvalid}
        <p class="mt-1 text-xs text-[var(--ha-error-color)]">{t("controller.commission.terms_invalid")}</p>
      {/if}
    </details>
  </Card>

  {#if running || phase !== "idle" || result || failure}
    <Card class="p-4" data-testid="commission-progress">
      {#if running}
        <span class="inline-flex items-center gap-2 text-sm text-[var(--ha-secondary-text-color)]"><Spinner size={16} />{t("controller.commission.running")}</span>
      {:else if result}
        <p class="text-sm text-[var(--ha-primary-text-color)]">
          <Badge variant="success">{t("controller.commission.succeeded")}</Badge>
          <a class="ml-2 font-medium text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + result.node.node_id_hex}>{nodeName(result.node)}</a>
        </p>
      {:else if failure}
        <p class="text-sm text-[var(--ha-error-color)]"><Badge variant="danger">{t("controller.commission.failed")}</Badge> {failure}</p>
      {:else if phase === "failed"}
        <p class="text-sm text-[var(--ha-error-color)]">{t("controller.commission.failed")}</p>
      {/if}
      {#if ble}
        <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]" data-testid="commission-ble">
          {ble.used ? t("controller.commission.ble_used") : t(`controller.commission.ble_skipped.${ble.reason ?? "no_central"}`)}
        </p>
      {/if}
      {#if steps.length > 0}
        <ol class="mt-3 space-y-1 text-xs" data-testid="commission-steps">
          {#each steps as step, i (i)}
            <li class="flex flex-wrap items-center gap-2">
              <span class="font-mono text-[var(--ha-secondary-text-color)]">{step.number}.{step.sub}</span>
              <span class="text-[var(--ha-primary-text-color)]">{step.name}</span>
              <Badge variant={stepVariant(step.phase)}>{t(`controller.commission.step.${step.phase}`)}</Badge>
            </li>
          {/each}
        </ol>
      {/if}
    </Card>
  {/if}

  <Card class="p-4">
    <div class="mb-3 flex flex-wrap items-center gap-3">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.commission.discovery")}</h3>
      <Button variant="outline" size="sm" disabled={searching} onclick={() => void search()}>
        {searching ? t("controller.commission.searching") : t("controller.commission.search")}
      </Button>
    </div>
    {#if searched}
      <DataTable rows={devices} {columns} rowKey={(d) => d.instance_name} emptyMessage={t("controller.commission.none_found")}>
        {#snippet cell(d, col)}
          {#if col.key === "device"}
            <span class="text-[var(--ha-primary-text-color)]">{d.device_name || d.instance_name}</span>
          {:else if col.key === "vendor"}
            <span class="text-[var(--ha-secondary-text-color)]">{vendorLabel(d.vendor_name, d.vendor_id)}</span>
          {:else if col.key === "discriminator"}
            <span class="font-mono text-xs">{d.long_discriminator}</span>
          {:else if col.key === "addresses"}
            <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{d.addresses.join(", ")}</span>
          {:else if col.key === "action"}
            <span class="inline-flex items-center gap-2">
              <Input
                class="w-32 font-mono"
                inputmode="numeric"
                autocomplete="off"
                aria-label={t("controller.commission.passcode")}
                placeholder={t("controller.commission.passcode")}
                value={passcodes[d.instance_name] ?? ""}
                oninput={(e) => (passcodes = { ...passcodes, [d.instance_name]: (e.currentTarget as HTMLInputElement).value })}
              />
              <Button size="sm" disabled={running} onclick={() => void pair(d)}>{t("controller.commission.pair")}</Button>
            </span>
          {/if}
        {/snippet}
      </DataTable>
    {/if}
  </Card>
</div>
