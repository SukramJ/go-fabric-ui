<!-- A topology graph (the contract's TopologyGraph) in hand-written SVG:
     the layered layout of graph-layout.ts, a shape per kind and a border
     colour per state from the theme tokens, pan and zoom on the viewBox
     (wheel, drag, buttons), and a side panel with the selected node's
     details and the link its ref names. Keyboard: Tab moves across the
     nodes and selects the focused one, Enter follows its link, Escape
     clears the selection. -->
<script lang="ts">
  import type { TopologyGraph, TopologyNode } from "../api/client";
  import { useT } from "../context";
  import Badge from "./Badge.svelte";
  import Button from "./Button.svelte";
  import { layoutGraph } from "./graph-layout";
  import { cn } from "./utils";

  type Props = {
    graph: TopologyGraph;
    /** The href of a node's page (its ref); undefined draws no link. */
    hrefOf?: (node: TopologyNode) => string | undefined;
    /** Follows a node's href on Enter; the default navigates the window. */
    follow?: (href: string) => void;
    /** The accessible name of the drawing. */
    label: string;
    class?: string;
  };

  let { graph, hrefOf = () => undefined, follow = (href) => window.location.assign(href), label, class: className }: Props = $props();

  const t = useT();

  const layout = $derived(layoutGraph(graph.nodes, graph.edges));
  const byId = $derived(new Map(graph.nodes.map((n) => [n.id, n])));

  // The view box: the whole graph until the viewer pans or zooms; a
  // reload keeps a view the viewer chose.
  let touched = $state(false);
  let chosen = $state({ x: 0, y: 0, w: 1, h: 1 });
  const view = $derived(touched ? chosen : { x: 0, y: 0, w: layout.width, h: layout.height });

  let selectedId = $state<string | null>(null);
  const selected = $derived(selectedId ? byId.get(selectedId) : undefined);

  let svg: SVGSVGElement | undefined = $state();

  // Rounded corners and a dash per kind; the state is the border colour.
  const SHAPE: Record<string, { rx: number; dash?: string }> = {
    node: { rx: 20 },
    aggregator: { rx: 4 },
    device: { rx: 8 },
    part: { rx: 8, dash: "4 3" },
    fabric: { rx: 0 },
    session: { rx: 20 },
    peer: { rx: 20 },
    thread_router: { rx: 4 },
    thread_end_device: { rx: 12 },
    wifi_ap: { rx: 0 },
    ethernet: { rx: 0 },
  };
  const STROKE: Record<string, string> = {
    ok: "var(--ha-primary-color)",
    unreachable: "var(--ha-error-color)",
    idle: "var(--ha-warning-color)",
    unknown: "var(--ha-divider-color)",
  };
  const STATE_BADGE = { ok: "success", unreachable: "danger", idle: "warning", unknown: "muted" } as const;

  const kindName = (k: string) => t(`graph.kind.${k}`);
  const stateName = (s: string) => t(`graph.state.${s}`);
  // A detail key the catalogue does not name shows as written.
  function detailName(key: string): string {
    const k = `graph.detail.${key}`;
    const s = t(k);
    return s === k ? key : s;
  }

  function setView(v: { x: number; y: number; w: number; h: number }) {
    chosen = v;
    touched = true;
  }

  /** Zooms by factor (below 1 zooms in) around a point of the drawing, within 1/8 and 4 times the graph. */
  function zoom(factor: number, cx = view.x + view.w / 2, cy = view.y + view.h / 2) {
    const w = Math.min(layout.width * 4, Math.max(layout.width / 8, view.w * factor));
    const f = w / view.w;
    setView({ x: cx - (cx - view.x) * f, y: cy - (cy - view.y) * f, w, h: view.h * f });
  }

  function fit() {
    touched = false;
  }

  /** The drawing's units per screen pixel and the box on screen. */
  function scale() {
    const rect = svg?.getBoundingClientRect();
    const width = rect?.width || view.w;
    return { rect, k: view.w / width };
  }

  function onWheel(e: WheelEvent) {
    e.preventDefault();
    const { rect, k } = scale();
    const cx = view.x + (e.clientX - (rect?.left ?? 0)) * k;
    const cy = view.y + (e.clientY - (rect?.top ?? 0)) * k;
    zoom(e.deltaY > 0 ? 1.2 : 1 / 1.2, cx, cy);
  }

  let drag: { x: number; y: number; vx: number; vy: number; id: number } | null = null;

  function onPointerDown(e: PointerEvent) {
    if ((e.target as Element | null)?.closest?.("[data-graph-node]")) return;
    drag = { x: e.clientX, y: e.clientY, vx: view.x, vy: view.y, id: e.pointerId };
    svg?.setPointerCapture?.(e.pointerId);
  }

  function onPointerMove(e: PointerEvent) {
    if (!drag || e.pointerId !== drag.id) return;
    const { k } = scale();
    setView({ x: drag.vx - (e.clientX - drag.x) * k, y: drag.vy - (e.clientY - drag.y) * k, w: view.w, h: view.h });
  }

  function onPointerUp(e: PointerEvent) {
    if (drag && e.pointerId === drag.id) {
      svg?.releasePointerCapture?.(e.pointerId);
      drag = null;
    }
  }

  function onNodeKey(e: KeyboardEvent, node: TopologyNode) {
    if (e.key === "Enter") {
      e.preventDefault();
      selectedId = node.id;
      const href = hrefOf(node);
      if (href) follow(href);
    } else if (e.key === " ") {
      e.preventDefault();
      selectedId = node.id;
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === "Escape" && selectedId) {
      e.preventDefault();
      selectedId = null;
    }
  }

  const nodeAria = (n: TopologyNode) => t("graph.node_aria", { kind: kindName(n.kind), label: n.label, state: stateName(n.state) });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class={cn("flex flex-col gap-3 md:flex-row", className)} onkeydown={onKey}>
  <div class="relative min-w-0 flex-1 rounded-[var(--ha-radius-control)] border border-[var(--ha-divider-color)] bg-[var(--ha-secondary-background-color)]">
    <div class="absolute right-2 top-2 z-10 flex gap-1">
      <Button variant="outline" size="sm" onclick={() => zoom(1 / 1.2)} aria-label={t("graph.zoom_in")} title={t("graph.zoom_in")}>+</Button>
      <Button variant="outline" size="sm" onclick={() => zoom(1.2)} aria-label={t("graph.zoom_out")} title={t("graph.zoom_out")}>−</Button>
      <Button variant="outline" size="sm" onclick={fit}>{t("graph.fit")}</Button>
    </div>
    <svg
      bind:this={svg}
      data-testid="graph"
      role="group"
      aria-label={label}
      class="block w-full touch-none select-none"
      style:height="{Math.min(640, Math.max(240, layout.height))}px"
      viewBox="{view.x} {view.y} {view.w} {view.h}"
      onwheel={onWheel}
      onpointerdown={onPointerDown}
      onpointermove={onPointerMove}
      onpointerup={onPointerUp}
      onpointercancel={onPointerUp}
    >
      <g fill="none" stroke="var(--ha-divider-color)" stroke-width="1.5">
        {#each layout.edges as e (e.source + ">" + e.target)}
          <path d={e.path} data-edge-kind={e.kind} stroke-dasharray={e.kind === "session" ? "5 4" : undefined} />
        {/each}
      </g>
      {#each layout.nodes as ln (ln.id)}
        {@const n = byId.get(ln.id)!}
        {@const shape = SHAPE[n.kind] ?? { rx: 6 }}
        <g
          data-graph-node
          data-testid="graph-node"
          data-id={n.id}
          data-state={n.state}
          role="button"
          tabindex="0"
          aria-label={nodeAria(n)}
          aria-pressed={selectedId === n.id}
          class="cursor-pointer outline-none [&:focus-visible>rect]:stroke-[3]"
          onclick={() => (selectedId = n.id)}
          onfocus={() => (selectedId = n.id)}
          onkeydown={(e) => onNodeKey(e, n)}
        >
          <rect
            x={ln.x}
            y={ln.y}
            width={ln.width}
            height={ln.height}
            rx={Math.min(shape.rx, ln.height / 2)}
            fill="var(--ha-card-background-color)"
            stroke={STROKE[n.state] ?? STROKE.unknown}
            stroke-width={selectedId === n.id ? 3 : 1.5}
            stroke-dasharray={shape.dash}
          />
          <text
            x={ln.x + ln.width / 2}
            y={ln.y + ln.height / 2}
            text-anchor="middle"
            dominant-baseline="central"
            font-size="12"
            fill="var(--ha-primary-text-color)">{ln.text}</text
          >
          <title>{n.label}</title>
        </g>
      {/each}
    </svg>
  </div>
  <aside
    class="w-full shrink-0 rounded-[var(--ha-radius-control)] border border-[var(--ha-divider-color)] p-3 text-sm md:w-64"
    data-testid="graph-panel"
    aria-live="polite"
  >
    {#if selected}
      {@const href = hrefOf(selected)}
      <div class="flex items-start justify-between gap-2">
        <h4 class="break-all font-medium text-[var(--ha-primary-text-color)]">{selected.label}</h4>
        <button type="button" class="text-[var(--ha-secondary-text-color)]" onclick={() => (selectedId = null)} aria-label={t("common.close")}>×</button>
      </div>
      <div class="mt-1 flex flex-wrap gap-1">
        <Badge variant="muted">{kindName(selected.kind)}</Badge>
        <Badge variant={STATE_BADGE[selected.state] ?? "muted"}>{stateName(selected.state)}</Badge>
      </div>
      {#if selected.details.length > 0}
        <dl class="mt-3 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
          {#each selected.details as d, i (i)}
            <dt class="text-[var(--ha-secondary-text-color)]">{detailName(d.key)}</dt>
            <dd class="break-all font-mono text-xs leading-5 text-[var(--ha-primary-text-color)]">{d.value}</dd>
          {/each}
        </dl>
      {/if}
      {#if href}
        <a class="mt-3 inline-block text-[var(--ha-primary-color)] underline" {href}>{t("graph.open")}</a>
      {/if}
    {:else}
      <p class="text-[var(--ha-secondary-text-color)]">{t("graph.select_hint")}</p>
    {/if}
  </aside>
</div>
