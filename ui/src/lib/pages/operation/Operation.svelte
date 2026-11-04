<!-- The daemon's operation, beside both page families (contract 0.5.0):
     what it is (version, go-fabric, uptime), the level it logs at (admin
     sets it), and its log — the ring of GET /operation/log, then every
     record GET /operation/log/stream appends, with pause, a level and text
     filter and copy. A host without an operations port answers 503: the
     page says so instead. Better than matterjs-server's dashboard, which
     has a log-level dialog and no log view (concept U5 E). -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { ApiError, type LogLevelName, type LogRecord, type OperationInfo } from "../../api/client";
  import { errorText } from "../../store.svelte";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Input from "../../kit/Input.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import Select from "../../kit/Select.svelte";
  import { formatUptime, KEEP_RECORDS, LEVELS, mergeRecords, recordLine, visibleRecords } from "./log";

  const { t, toast, client, subscribeLog } = useFabricUI();

  let loading = $state(true);
  let unavailable = $state(false);
  let error = $state<string | null>(null);
  let info = $state<OperationInfo | null>(null);
  let level = $state<LogLevelName>("info");
  let levels = $state<LogLevelName[]>([...LEVELS]);
  let records = $state<LogRecord[]>([]);
  let paused = $state(false);
  // held are the records the stream delivered while paused.
  let held: LogRecord[] = [];
  let heldCount = $state(0);
  let minLevel = $state<LogLevelName>("debug");
  let text = $state("");
  let saving = $state(false);

  function append(rec: LogRecord) {
    if (paused) {
      held.push(rec);
      heldCount = held.length;
      return;
    }
    records = mergeRecords(records, [rec]);
  }

  async function load() {
    loading = true;
    error = null;
    try {
      const [i, l, log] = await Promise.all([client.operationInfo(), client.logLevel(), client.logRecords({ limit: KEEP_RECORDS })]);
      info = i;
      level = l.level;
      levels = l.levels;
      records = mergeRecords(log.records, records);
    } catch (err) {
      if (err instanceof ApiError && (err.status === 503 || err.status === 404)) unavailable = true;
      else error = errorText(err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    // The stream first: a record logged while the ring is read is kept,
    // and mergeRecords drops the one both deliver.
    const close = subscribeLog(append);
    void load();
    return close;
  });

  function togglePause() {
    paused = !paused;
    if (!paused && held.length > 0) {
      records = mergeRecords(records, held);
      held = [];
      heldCount = 0;
    }
  }

  async function changeLevel(next: string) {
    const prev = level;
    level = next as LogLevelName;
    saving = true;
    try {
      level = (await client.setLogLevel(next as LogLevelName)).level;
      toast.success(t("operation.level_set", { level: level }));
    } catch (err) {
      level = prev;
      toast.error(errorText(err));
    } finally {
      saving = false;
    }
  }

  const shown = $derived(visibleRecords(records, minLevel, text));

  async function copy() {
    try {
      await navigator.clipboard.writeText(shown.map(recordLine).join("\n"));
      toast.success(t("node.pair.copied"));
    } catch {
      toast.error(t("node.pair.copy_failed"));
    }
  }

  const levelVariant = (l: LogLevelName) => (l === "error" ? "danger" : l === "warn" ? "warning" : l === "info" ? "default" : "muted");
  const levelOptions = $derived(levels.map((l) => ({ value: l, label: t(`operation.level.${l}`) })));
</script>

{#if loading && !info && !unavailable}
  <LoadingState message={t("common.loading")} />
{:else if unavailable}
  <EmptyState message={t("operation.unavailable")} />
{:else if error}
  <ErrorState message={error} onRetry={() => void load()} />
{:else}
  <div class="space-y-4">
    {#if info}
      <Card class="p-4" data-testid="operation-info">
        <div class="flex flex-wrap items-center gap-4 text-sm">
          <span class="font-medium text-[var(--ha-primary-text-color)]">{t("operation.version", { version: info.version })}</span>
          {#if info.go_fabric}
            <span class="text-[var(--ha-secondary-text-color)]">{t("operation.go_fabric", { version: info.go_fabric })}</span>
          {/if}
          <span class="text-[var(--ha-secondary-text-color)]" title={info.started_at}>{t("operation.uptime", { uptime: formatUptime(t, info.uptime_s) })}</span>
        </div>
      </Card>
    {/if}

    <Card class="p-4">
      <div class="flex flex-wrap items-center gap-3">
        <h2 class="mr-auto text-base font-semibold text-[var(--ha-primary-text-color)]">{t("operation.log")}</h2>
        <label class="flex items-center gap-2 text-sm text-[var(--ha-secondary-text-color)]">
          {t("operation.level_label")}
          <Select class="w-auto" value={level} disabled={saving} ariaLabel={t("operation.level_label")} options={levelOptions} onValueChange={(v) => void changeLevel(v)} />
        </label>
      </div>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("operation.log_hint")}</p>
      <div class="mt-3 flex flex-wrap items-center gap-2">
        <Select class="w-auto" value={minLevel} ariaLabel={t("operation.filter_level")} options={levelOptions} onValueChange={(v) => (minLevel = v as LogLevelName)} />
        <Input class="w-auto min-w-[12rem] flex-1" placeholder={t("operation.filter_text")} aria-label={t("operation.filter_text")} bind:value={text} />
        <Button variant="outline" size="sm" onclick={togglePause} aria-pressed={paused}>{paused ? t("operation.resume") : t("operation.pause")}</Button>
        <Button variant="outline" size="sm" onclick={() => void copy()} disabled={shown.length === 0}>{t("common.copy")}</Button>
      </div>
      {#if paused}
        <p class="mt-2 text-xs text-[var(--ha-warning-color)]" data-testid="paused">{t("operation.paused", { count: heldCount })}</p>
      {/if}
      {#if shown.length === 0}
        <EmptyState class="mt-3" message={records.length === 0 ? t("operation.log_empty") : t("operation.log_filtered")} />
      {:else}
        <ol class="mt-3 max-h-[32rem] space-y-1 overflow-y-auto font-mono text-xs" data-testid="log">
          {#each shown as rec (rec.seq)}
            <li class="flex flex-wrap items-baseline gap-2" data-testid="log-record">
              <span class="shrink-0 text-[var(--ha-secondary-text-color)]">{new Date(rec.time).toLocaleTimeString()}</span>
              <Badge variant={levelVariant(rec.level)}>{rec.level}</Badge>
              <span class="text-[var(--ha-primary-text-color)]">{rec.msg}</span>
              {#each Object.entries(rec.attrs) as [k, v] (k)}
                <span class="text-[var(--ha-secondary-text-color)]">{k}={v}</span>
              {/each}
            </li>
          {/each}
        </ol>
      {/if}
    </Card>
  </div>
{/if}
