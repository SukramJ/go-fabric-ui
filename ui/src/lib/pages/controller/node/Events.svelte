<!-- The events the controller received from the node (contract 0.4.1,
     GET /controller/nodes/{id}/events): its default subscription's, the
     subscription's bootstrap read's and the read after commissioning,
     from the controller's memory — nothing is read from the node. The tab
     fetches the events after the last one it shows whenever the page's
     node detail is reloaded (the stream's node_updated, which an event the
     node delivers raises too). -->
<script lang="ts">
  import { untrack } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { ControllerNodeDetail, EventRecord } from "../../../api/client";
  import Button from "../../../kit/Button.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import EventList from "../../../kit/EventList.svelte";
  import { mergeEvents, sinceAfter } from "../../../kit/events";
  import LoadingState from "../../../kit/LoadingState.svelte";

  let { nodeId, detail }: { nodeId: string; detail: ControllerNodeDetail } = $props();
  const { t, client } = useFabricUI();

  // How many events the tab keeps; the controller's history holds 100 by default.
  const MAX_EVENTS = 500;

  let events = $state<EventRecord[]>([]);
  let loaded = $state(false);
  let error = $state<string | null>(null);
  // A newer load wins: an older answer that arrives last is dropped.
  let seq = 0;

  // fresh reads the whole history; otherwise only the events after the
  // highest number shown (since = it + 1). The history is in delivery
  // order and a node's numbers need not rise: the answer is merged by
  // number, so an event read again is listed once (UR-P2-ui-90).
  async function load(fresh: boolean) {
    const mine = ++seq;
    // The cursor travels as since_str, exact beyond 2^53 (0.10.0).
    const w = fresh ? {} : sinceAfter(events);
    try {
      const l = await client.controllerNodeEvents(nodeId, w);
      if (mine !== seq) return;
      events = mergeEvents(w.since === undefined ? [] : events, l.events).slice(-MAX_EVENTS);
      error = null;
      loaded = true;
    } catch (err) {
      if (mine === seq) error = errorText(err);
    }
  }

  // The first detail loads the history; each reload of the node fetches
  // the newer events.
  let first = true;
  $effect(() => {
    void detail;
    untrack(() => {
      void load(first);
      first = false;
    });
  });
</script>

<div class="mb-3 flex items-center justify-between gap-2">
  <p class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.events.hint")}</p>
  <Button variant="outline" size="sm" onclick={() => void load(true)}>{t("common.refresh")}</Button>
</div>
{#if error}
  <ErrorState message={error} onRetry={() => void load(true)} />
{:else if !loaded}
  <LoadingState label={t("common.loading")} />
{:else}
  <EventList {events} empty={t("controller.events.empty")} />
{/if}
