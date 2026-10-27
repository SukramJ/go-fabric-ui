<!-- The node's access control entries of our fabric: read, edit, write
     back as one list. Saving a list without an Administer entry for this
     controller locks it out; the page asks first. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { AclEntry } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import Input from "../../../kit/Input.svelte";
  import LoadingState from "../../../kit/LoadingState.svelte";
  import Select from "../../../kit/Select.svelte";
  import { hexId, parseHexId } from "../format";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, confirm, client, controller } = useFabricUI();

  let entries = $state<AclEntry[] | null>(null);
  let error = $state<string | null>(null);
  let dirty = $state(false);
  let busy = $state(false);

  let privilege = $state<AclEntry["privilege"]>("operate");
  let authMode = $state<AclEntry["auth_mode"]>("case");
  let subjects = $state("");
  let targetCluster = $state("");
  let targetEndpoint = $state("");

  const PRIVILEGES = ["view", "proxy_view", "operate", "manage", "administer"] as const;
  const MODES = ["pase", "case", "group"] as const;

  async function load() {
    error = null;
    try {
      entries = (await client.nodeAcl(nodeId)).entries;
      dirty = false;
    } catch (err) {
      error = errorText(err);
    }
  }
  onMount(() => void load());

  function add() {
    const subs = subjects.split(/[\s,]+/).filter(Boolean);
    if (subs.some((s) => !/^(0x)?[0-9a-f]{1,16}$/i.test(s))) {
      toast.error(t("controller.acl.invalid_subject"));
      return;
    }
    const target: AclEntry["targets"][number] = {};
    if (targetCluster.trim()) {
      const c = parseHexId(targetCluster);
      if (c === null) {
        toast.error(t("controller.acl.invalid_target"));
        return;
      }
      target.cluster_hex = hexId(c);
    }
    if (targetEndpoint.trim()) target.endpoint = Number(targetEndpoint);
    const e: AclEntry = {
      privilege,
      auth_mode: authMode,
      subjects_hex: subs.map((s) => s.replace(/^0x/i, "").toUpperCase().padStart(16, "0")),
      targets: Object.keys(target).length ? [target] : [],
    };
    entries = [...(entries ?? []), e];
    subjects = targetCluster = targetEndpoint = "";
    dirty = true;
  }

  function remove(i: number) {
    entries = (entries ?? []).filter((_, j) => j !== i);
    dirty = true;
  }

  // Does the list keep this controller's Administer access?
  function keepsUs(list: AclEntry[]): boolean {
    const own = controller.status?.node_id_hex;
    return list.some((e) => e.privilege === "administer" && e.auth_mode === "case" && (!own || e.subjects_hex.length === 0 || e.subjects_hex.includes(own)));
  }

  async function save() {
    if (!entries) return;
    if (!keepsUs(entries)) {
      const ok = await confirm({ title: t("controller.acl.lockout"), body: t("controller.acl.lockout_body"), destructive: true, confirmLabel: t("common.save") });
      if (!ok) return;
    }
    busy = true;
    try {
      const res = await client.setNodeAcl(nodeId, entries);
      if (res.status === 0) toast.success(t("controller.acl.saved"));
      else toast.error(t("controller.attributes.write_status", { status: hexId(res.status) }));
      await load();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  const privilegeOptions = $derived(PRIVILEGES.map((p) => ({ value: p, label: t(`controller.acl.privilege.${p}`) })));
  const modeOptions = $derived(MODES.map((m) => ({ value: m, label: t(`controller.acl.mode.${m}`) })));
</script>

{#if error}
  <ErrorState message={error} onRetry={() => void load()} />
{:else if !entries}
  <LoadingState message={t("common.loading")} />
{:else}
  <div class="space-y-4">
    <Card class="p-4">
      {#if entries.length === 0}
        <EmptyState message={t("controller.acl.empty")} />
      {:else}
        <ul class="divide-y divide-[var(--ha-divider-color)]" data-testid="acl-entries">
          {#each entries as e, i (i)}
            <li class="flex flex-wrap items-center gap-2 py-2 text-sm">
              <Badge variant={e.privilege === "administer" ? "warning" : "default"}>{t(`controller.acl.privilege.${e.privilege}`)}</Badge>
              <Badge variant="muted">{t(`controller.acl.mode.${e.auth_mode}`)}</Badge>
              <span class="font-mono text-xs">{e.subjects_hex.length ? e.subjects_hex.map((s) => "0x" + s).join(", ") : t("controller.acl.any_subject")}</span>
              <span class="text-xs text-[var(--ha-secondary-text-color)]">
                {e.targets.length
                  ? e.targets.map((x) => [x.endpoint !== undefined ? t("controller.endpoints.endpoint", { number: x.endpoint }) : "", x.cluster_hex ?? "", x.device_type_hex ?? ""].filter(Boolean).join(" ")).join("; ")
                  : t("controller.acl.whole_node")}
              </span>
              <Button size="sm" variant="ghost" class="ml-auto text-[var(--ha-error-color)]" onclick={() => remove(i)}>{t("common.remove")}</Button>
            </li>
          {/each}
        </ul>
      {/if}
    </Card>
    <Card class="p-4">
      <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.acl.add")}</h3>
      <div class="flex flex-wrap items-center gap-2">
        <Select class="w-auto" ariaLabel={t("controller.acl.privilege_label")} value={privilege} onValueChange={(v) => (privilege = v as typeof privilege)} options={privilegeOptions} />
        <Select class="w-auto" ariaLabel={t("controller.acl.mode_label")} value={authMode} onValueChange={(v) => (authMode = v as typeof authMode)} options={modeOptions} />
        <Input class="w-64 font-mono" bind:value={subjects} placeholder={t("controller.acl.subjects")} aria-label={t("controller.acl.subjects")} />
        <Input class="w-28 font-mono" bind:value={targetCluster} placeholder={t("controller.acl.target_cluster")} aria-label={t("controller.acl.target_cluster")} />
        <Input class="w-24" bind:value={targetEndpoint} placeholder={t("controller.acl.target_endpoint")} aria-label={t("controller.acl.target_endpoint")} inputmode="numeric" />
        <Button variant="outline" onclick={add}>{t("controller.acl.add_entry")}</Button>
      </div>
      <div class="mt-3 flex gap-2">
        <Button disabled={!dirty || busy} onclick={() => void save()}>{t("common.save")}</Button>
        <Button variant="ghost" disabled={!dirty || busy} onclick={() => void load()}>{t("controller.acl.discard")}</Button>
      </div>
    </Card>
  </div>
{/if}
