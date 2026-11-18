# Review phase 1b — inventory (script, read-only map)

Generated at `007ca961e84c0628fd30a0387d0ba88fbef0ede5` by grep over the hand-written web source
(`ui/src`, without tests, `*.gen.ts`, the catalogues, `test-setup.ts`).
A map, never a citation: every row is a grep hit, not a reading. The
per-site list of strings from nodes, hosts and logs is
`read-review-p0-inventory.md` section 3.3 (at `25f6e7b4`; lines may have
moved by a few — re-grep before citing).

## 0. Files and sizes

```
  9437 total
   456 ui/src/lib/api/client.ts
     3 ui/src/lib/api/index.ts
   151 ui/src/lib/api/stream.ts
   167 ui/src/lib/context.ts
   249 ui/src/lib/controller-store.svelte.ts
    14 ui/src/lib/countdown.ts
    50 ui/src/lib/i18n/index.ts
    11 ui/src/lib/index.ts
    53 ui/src/lib/kit/AttributeTable.svelte
    26 ui/src/lib/kit/Badge.svelte
    50 ui/src/lib/kit/Button.svelte
    16 ui/src/lib/kit/Card.svelte
    53 ui/src/lib/kit/ClusterView.svelte
   239 ui/src/lib/kit/DataTable.svelte
    24 ui/src/lib/kit/EmptyState.svelte
    22 ui/src/lib/kit/ErrorState.svelte
    45 ui/src/lib/kit/EventList.svelte
   272 ui/src/lib/kit/Graph.svelte
    12 ui/src/lib/kit/Icon.svelte
    22 ui/src/lib/kit/Input.svelte
    15 ui/src/lib/kit/LoadingState.svelte
    32 ui/src/lib/kit/PageHeader.svelte
    16 ui/src/lib/kit/PageShell.svelte
    49 ui/src/lib/kit/Select.svelte
    59 ui/src/lib/kit/Spinner.svelte
    66 ui/src/lib/kit/Tabs.svelte
    27 ui/src/lib/kit/TagChips.svelte
    24 ui/src/lib/kit/data-table.ts
   194 ui/src/lib/kit/graph-layout.ts
    41 ui/src/lib/kit/icons.ts
    22 ui/src/lib/kit/im-status.ts
    28 ui/src/lib/kit/index.ts
    13 ui/src/lib/kit/tabs.ts
    49 ui/src/lib/kit/utils.ts
    62 ui/src/lib/nodes-store.svelte.ts
   324 ui/src/lib/pages/controller/Commission.svelte
   265 ui/src/lib/pages/controller/Diagnostics.svelte
   112 ui/src/lib/pages/controller/Network.svelte
   105 ui/src/lib/pages/controller/Node.svelte
   181 ui/src/lib/pages/controller/Nodes.svelte
    67 ui/src/lib/pages/controller/QrScan.svelte
   160 ui/src/lib/pages/controller/Settings.svelte
    49 ui/src/lib/pages/controller/format.ts
   151 ui/src/lib/pages/controller/node/Acl.svelte
   100 ui/src/lib/pages/controller/node/Attributes.svelte
   114 ui/src/lib/pages/controller/node/Bindings.svelte
    61 ui/src/lib/pages/controller/node/Endpoints.svelte
    68 ui/src/lib/pages/controller/node/Events.svelte
    83 ui/src/lib/pages/controller/node/Fabrics.svelte
    92 ui/src/lib/pages/controller/node/Icd.svelte
   184 ui/src/lib/pages/controller/node/Inventory.svelte
   120 ui/src/lib/pages/controller/node/Overview.svelte
    75 ui/src/lib/pages/controller/node/Updates.svelte
    80 ui/src/lib/pages/controller/node/Window.svelte
    27 ui/src/lib/pages/index.ts
   281 ui/src/lib/pages/node/Diagnostics.svelte
    62 ui/src/lib/pages/node/EcosystemHints.svelte
   228 ui/src/lib/pages/node/Endpoint.svelte
   138 ui/src/lib/pages/node/Endpoints.svelte
   241 ui/src/lib/pages/node/Fabrics.svelte
    31 ui/src/lib/pages/node/HostCounters.svelte
    41 ui/src/lib/pages/node/HostWarnings.svelte
    47 ui/src/lib/pages/node/NodeConfig.svelte
   111 ui/src/lib/pages/node/NodeManage.svelte
   191 ui/src/lib/pages/node/Pair.svelte
   440 ui/src/lib/pages/node/Selection.svelte
   140 ui/src/lib/pages/node/Status.svelte
   290 ui/src/lib/pages/node/Surface.svelte
    71 ui/src/lib/pages/node/Topology.svelte
    35 ui/src/lib/pages/node/ecosystems.ts
    67 ui/src/lib/pages/node/fields.ts
    29 ui/src/lib/pages/node/surface.ts
   169 ui/src/lib/pages/operation/Operation.svelte
    45 ui/src/lib/pages/operation/log.ts
    58 ui/src/lib/qr.ts
    81 ui/src/lib/scan/index.ts
   421 ui/src/lib/store.svelte.ts
   159 ui/src/lib/theme/ha-bridge.ts
     1 ui/src/lib/theme/index.ts
   316 ui/src/lib/theme/tokens.css
   276 ui/src/shell/App.svelte
    33 ui/src/shell/ConfirmDialog.svelte
    58 ui/src/shell/NodeCreateDialog.svelte
    35 ui/src/shell/NodeFamily.svelte
    38 ui/src/shell/NodeScope.svelte
    20 ui/src/shell/Toaster.svelte
    24 ui/src/shell/app.css
    23 ui/src/shell/confirm.svelte.ts
    95 ui/src/shell/index.html
     7 ui/src/shell/main.ts
    53 ui/src/shell/route.ts
    32 ui/src/shell/toast.svelte.ts
```

## 1. Render sinks

### {@html}

```
ui/src/lib/pages/controller/node/Window.svelte:66: <div class="w-48">{@html qrSvg}</div>
ui/src/lib/pages/node/Pair.svelte:150: <div class="h-auto max-w-full">{@html qrSvg}</div>
```

### DOM string sinks, createElement

```
ui/src/lib/scan/index.ts:54: const canvas = document.createElement("canvas");
```

### dynamic href

```
ui/src/lib/kit/Tabs.svelte:53: <a href={item.href} role="tab" aria-selected={isActive} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
ui/src/lib/pages/controller/Commission.svelte:262: <a class="ml-2 font-medium text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + result.node.node_id_hex}>{nodeName(result.node)}</a>
ui/src/lib/pages/controller/Diagnostics.svelte:132: <a class="font-mono text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + s.peer_node_id_hex}>0x{s.peer_node_id_hex}</a>
ui/src/lib/pages/controller/Diagnostics.svelte:148: <a class="text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + p.node_id_hex} data-testid="peer-link">{peerName(p)}</a>
ui/src/lib/pages/controller/Node.svelte:65: <a href={routes.controllerNodes} class="text-sm text-[var(--ha-primary-color)] hover:underline">← {t("controller.node.back")}</a>
ui/src/lib/pages/controller/Nodes.svelte:119: <a href={routes.controllerCommission} class="ml-auto text-sm font-medium text-[var(--ha-primary-color)] hover:underline">{t("controller.nodes.commission")}</a>
ui/src/lib/pages/controller/Nodes.svelte:161: <a class="inline-flex items-center px-2 text-sm text-[var(--ha-primary-color)] hover:underline" href={`${routes.controllerNode}${n.node_id_hex}?tab=window`}>{t("controller.no
ui/src/lib/pages/node/EcosystemHints.svelte:48: <a class="text-[var(--ha-primary-color)] underline" href={link.href} target="_blank" rel="noopener noreferrer">{t(`node.pair.hints.${hint.ecosystem}.${link.key}`)}</a>
ui/src/lib/pages/node/Endpoint.svelte:128: <a href={routes.endpoints} class="text-sm text-[var(--ha-primary-color)] hover:underline">← {t("node.endpoint.back")}</a>
ui/src/lib/pages/node/Endpoint.svelte:153: <span>{t("node.endpoint.parent")} <a class="text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + d.parent_endpoint_id}>#{d.parent_endpoint_id}</a></span>
ui/src/lib/pages/node/Endpoint.svelte:159: <a class="ml-1 text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + p}>#{p}</a>
ui/src/lib/pages/node/Endpoints.svelte:96: <a class="rounded border border-[var(--ha-divider-color)] px-3 py-1 text-sm text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + s.number}>
ui/src/lib/pages/node/Endpoints.svelte:111: <a class="font-medium text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + ep.endpoint_id}>{ep.name}</a>
ui/src/lib/pages/node/Fabrics.svelte:237: <a href={routes.pair} data-testid="share-link" class="mt-3 inline-flex items-center text-sm font-medium text-[var(--ha-primary-color)] hover:underline">
ui/src/lib/pages/node/NodeManage.svelte:106: <a class="mt-1 inline-block text-sm text-[var(--ha-primary-color)] underline" href={HINT_SOURCE} target="_blank" rel="noopener noreferrer">
```

### navigation

```
ui/src/lib/kit/Graph.svelte:29: let { graph, hrefOf = () => undefined, follow = (href) => window.location.assign(href), label, class: className }: Props = $props();
ui/src/lib/pages/controller/Nodes.svelte:74: location.hash = routes.controllerNode + n.node_id_hex;
ui/src/lib/pages/controller/node/Window.svelte:62: <Button disabled={busy} onclick={() => void open()}>{t("controller.window.open")}</Button>
ui/src/shell/App.svelte:71: let route = $state<Route>(routeOf(location.hash));
ui/src/shell/App.svelte:72: const onHash = () => (route = routeOf(location.hash));
ui/src/shell/App.svelte:126: location.hash = key === defaultKey ? ui.routes.status : routesForNode(ui.routes, key).status;
ui/src/shell/App.svelte:131: location.hash = routesForNode(ui.routes, created.key).selection;
ui/src/shell/App.svelte:138: location.hash = ui.routes.status;
```

### target / rel

```
ui/src/lib/pages/node/EcosystemHints.svelte:48: <a class="text-[var(--ha-primary-color)] underline" href={link.href} target="_blank" rel="noopener noreferrer">{t(`node.pair.hints.${hint.ecosystem}.${link.key}`)}</a>
ui/src/lib/pages/node/NodeManage.svelte:106: <a class="mt-1 inline-block text-sm text-[var(--ha-primary-color)] underline" href={HINT_SOURCE} target="_blank" rel="noopener noreferrer">
ui/src/shell/index.html:8: <link rel="icon" href="./favicon.svg" type="image/svg+xml" />
```

### style from values

```
ui/src/lib/kit/Graph.svelte:176: style:height="{Math.min(640, Math.max(240, layout.height))}px"
ui/src/lib/kit/Spinner.svelte:11:<span class={cn("fui-loader text-[var(--ha-secondary-text-color)]", className)} style="--fui-size:{size}px" aria-hidden="true">
ui/src/lib/pages/controller/node/Endpoints.svelte:32: <Card class="p-4" style={`margin-left: ${depth * 1.5}rem`} data-testid={`endpoint-${ep.endpoint}`}>
ui/src/lib/pages/node/Pair.svelte:141: style:transition="stroke-dashoffset 1s linear"
ui/src/lib/theme/ha-bridge.ts:120: if (value) ourRoot.style.setProperty(name, value);
ui/src/lib/theme/ha-bridge.ts:121: else ourRoot.style.removeProperty(name);
ui/src/lib/theme/ha-bridge.ts:130: ourRoot.style.colorScheme = dark ? "dark" : "light";
ui/src/shell/index.html:62: if (value) root.style.setProperty(vars[i], value);
ui/src/shell/index.html:80: root.style.colorScheme = dark ? "dark" : "light";
```

### attributes from values (title, aria, data-)

```
ui/src/lib/kit/AttributeTable.svelte:31: <tr class="border-t border-[var(--ha-divider-color)] align-top" data-testid={`attribute-${a.id_hex}`}>
ui/src/lib/kit/AttributeTable.svelte:41: <td class="py-1 text-xs" data-testid={a.status !== undefined ? "attribute-status" : undefined}>
ui/src/lib/kit/AttributeTable.svelte:44: <Badge variant="warning" title={`${imStatusName(a.status)} (${statusHex(a.status)})`}>{t(`inventory.status.${IM_STATUS[a.status]}`)}</Badge>
ui/src/lib/kit/ClusterView.svelte:23:<div class={cn("space-y-3", className)} data-testid={`cluster-view-${cluster.id_hex}`}>
ui/src/lib/kit/ClusterView.svelte:28: <Badge variant="muted" title={t("inventory.feature_map")}>{cluster.feature_map_hex}</Badge>
ui/src/lib/kit/ClusterView.svelte:36: <li><Badge title={t("inventory.feature_bit", { bit: f.bit })}>{f.title ? `${f.code} · ${f.title}` : f.code}</Badge></li>
ui/src/lib/kit/ClusterView.svelte:48: <Badge variant="muted" class="mr-1" title={e.id_hex}>{e.name ?? e.id_hex}</Badge>
ui/src/lib/kit/DataTable.svelte:209: data-label={col.label}
ui/src/lib/kit/EventList.svelte:28: <li class="py-2 text-sm" data-testid={`event-${e.number}`}>
ui/src/lib/kit/EventList.svelte:32: <span class="font-medium text-[var(--ha-primary-text-color)]" title={`${e.cluster_hex} · ${e.event_hex}`}>{e.name ?? e.event_hex}</span>
ui/src/lib/kit/Graph.svelte:166: <Button variant="outline" size="sm" onclick={() => zoom(1 / 1.2)} aria-label={t("graph.zoom_in")} title={t("graph.zoom_in")}>+</Button>
ui/src/lib/kit/Graph.svelte:167: <Button variant="outline" size="sm" onclick={() => zoom(1.2)} aria-label={t("graph.zoom_out")} title={t("graph.zoom_out")}>−</Button>
ui/src/lib/kit/Graph.svelte:174: aria-label={label}
ui/src/lib/kit/Graph.svelte:191: data-edge-kind={e.kind}
ui/src/lib/kit/Graph.svelte:192: data-strength={edge?.kind === "link" ? (edge.strength ?? "unknown") : undefined}
ui/src/lib/kit/Graph.svelte:207: data-id={n.id}
ui/src/lib/kit/Graph.svelte:208: data-state={n.state}
ui/src/lib/kit/Graph.svelte:211: aria-label={nodeAria(n)}
ui/src/lib/kit/Graph.svelte:251: <button type="button" class="text-[var(--ha-secondary-text-color)]" onclick={() => (selectedId = null)} aria-label={t("common.close")}>×</button>
ui/src/lib/kit/Icon.svelte:12:<Cmp {size} class={className} aria-label={ariaLabel} aria-hidden={ariaLabel ? undefined : "true"} {title} />
ui/src/lib/kit/Select.svelte:34: aria-label={ariaLabel}
ui/src/lib/kit/Tabs.svelte:49:<div class={listClass} role="tablist" aria-label={ariaLabel}>
ui/src/lib/kit/TagChips.svelte:22: <li><Badge variant="muted" title={title(tag)}>{tag.label || `${tag.namespace}:${tag.tag}`}</Badge></li>
ui/src/lib/pages/controller/Commission.svelte:214: <Input class="max-w-sm font-mono" bind:value={code} placeholder={t("controller.commission.code_placeholder")} aria-label={t("controller.commission.code")} autocomplete="
ui/src/lib/pages/controller/Commission.svelte:312: aria-label={t("controller.commission.passcode")}
ui/src/lib/pages/controller/Network.svelte:100: <ul class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-[var(--ha-secondary-text-color)]" data-testid="network-legend" aria-label={t("controller.network.le
ui/src/lib/pages/controller/Network.svelte:102: <li class="inline-flex items-center gap-1.5" data-kind={k}>
ui/src/lib/pages/controller/Settings.svelte:101: <Input class="max-w-xs" bind:value={label} maxlength={32} aria-label={t("controller.settings.label")} />
ui/src/lib/pages/controller/Settings.svelte:131: <Input class="w-32" bind:value={id} placeholder={t("controller.settings.col_id")} aria-label={t("controller.settings.col_id")} />
ui/src/lib/pages/controller/Settings.svelte:133: <Input class="w-40" bind:value={ssid} placeholder={t("controller.settings.ssid")} aria-label={t("controller.settings.ssid")} />
ui/src/lib/pages/controller/Settings.svelte:134: <Input class="w-48" type="password" autocomplete="new-password" bind:value={secret} placeholder={t("controller.settings.passphrase")} aria-label={t("controller.settings.pa
ui/src/lib/pages/controller/Settings.svelte:136: <Input class="w-72 font-mono" type="password" autocomplete="off" bind:value={secret} placeholder={t("controller.settings.dataset")} aria-label={t("controller.settings.data
ui/src/lib/pages/controller/node/Acl.svelte:140: <Input class="w-64 font-mono" bind:value={subjects} placeholder={t("controller.acl.subjects")} aria-label={t("controller.acl.subjects")} />
ui/src/lib/pages/controller/node/Acl.svelte:141: <Input class="w-28 font-mono" bind:value={targetCluster} placeholder={t("controller.acl.target_cluster")} aria-label={t("controller.acl.target_cluster")} />
ui/src/lib/pages/controller/node/Acl.svelte:142: <Input class="w-24" bind:value={targetEndpoint} placeholder={t("controller.acl.target_endpoint")} aria-label={t("controller.acl.target_endpoint")} inputmode="numeric" />
ui/src/lib/pages/controller/node/Attributes.svelte:84: <Input class="w-28 font-mono" bind:value={attribute} aria-label={t("controller.attributes.attribute")} placeholder="0x0000" />
ui/src/lib/pages/controller/node/Attributes.svelte:94: <Input class="max-w-md font-mono" bind:value={edit} aria-label={t("controller.attributes.value")} />
ui/src/lib/pages/controller/node/Bindings.svelte:91: <Card class="p-4" data-testid={`bindings-${eb.endpoint}`}>
ui/src/lib/pages/controller/node/Bindings.svelte:106: <Input class="w-56 font-mono" placeholder={t("controller.bindings.target")} aria-label={t("controller.bindings.target")} value={drafts[eb.endpoint]?.target ?? ""} oni
ui/src/lib/pages/controller/node/Bindings.svelte:107: <Input class="w-24" placeholder={t("controller.bindings.endpoint")} aria-label={t("controller.bindings.endpoint")} inputmode="numeric" value={drafts[eb.endpoint]?.end
ui/src/lib/pages/controller/node/Bindings.svelte:108: <Input class="w-28 font-mono" placeholder={t("controller.bindings.cluster")} aria-label={t("controller.bindings.cluster")} value={drafts[eb.endpoint]?.cluster ?? ""} 
ui/src/lib/pages/controller/node/Endpoints.svelte:32: <Card class="p-4" style={`margin-left: ${depth * 1.5}rem`} data-testid={`endpoint-${ep.endpoint}`}>
ui/src/lib/pages/controller/node/Endpoints.svelte:37: <Badge title={dt.id_hex}>{dt.name || dt.id_hex}{dt.revision ? ` r${dt.revision}` : ""}</Badge>
ui/src/lib/pages/controller/node/Endpoints.svelte:46: data-testid={`inventory-link-${ep.endpoint}`}
ui/src/lib/pages/controller/node/Endpoints.svelte:54: <li><Badge variant="muted" title={`${c.id_hex} r${c.revision}${c.features.length ? " · " + c.features.join(", ") : ""}`}>{c.name || c.id_hex}</Badge></li>
ui/src/lib/pages/controller/node/Inventory.svelte:100: <Card class={["p-4", focus === ne.endpoint ? "ring-2 ring-[var(--ha-primary-color)]" : ""].join(" ")} data-testid={`inventory-endpoint-${ne.endpoint}`}>
ui/src/lib/pages/controller/node/Inventory.svelte:107: <Badge title={dt.id_hex}>{dt.name ?? dt.id_hex} r{dt.revision}</Badge>
ui/src/lib/pages/controller/node/Inventory.svelte:113: <Badge variant={conformanceVariant(ep.conformance.status)} title={ep.conformance.unjudged.length ? t("controller.inventory.unjudged", { clusters: ep.conformance.unju
ui/src/lib/pages/controller/node/Inventory.svelte:129: <li><Badge variant="muted" title={t(l.user ? "controller.inventory.user_label" : "controller.inventory.fixed_label")}>{l.label}: {l.value}</Badge></li>
ui/src/lib/pages/controller/node/Inventory.svelte:137: <ul class="mt-2 space-y-1 text-xs text-[var(--ha-error-color)]" data-testid={`violations-${ne.endpoint}`}>
ui/src/lib/pages/controller/node/Inventory.svelte:155: data-testid={`cluster-${ne.endpoint}-${c.id_hex}`}
ui/src/lib/pages/controller/node/Inventory.svelte:162: <Badge variant="muted" title={t("inventory.feature_bit", { bit: f.bit })}>{f.code}</Badge>
ui/src/lib/pages/controller/node/Overview.svelte:75: {#if r.operational_host}<dt class="text-[var(--ha-secondary-text-color)]" title={t("controller.overview.operational_host_hint")}>{t("controller.overview.operational_ho
ui/src/lib/pages/node/Endpoint.svelte:142: <Badge title={dt.id_hex}>{dt.name ?? dt.id_hex} r{dt.revision}</Badge>
ui/src/lib/pages/node/Endpoint.svelte:148: <Badge variant="muted" title={t("node.endpoint.scope_hint")}>{t("node.endpoint.scope", { scope: d.scope })}</Badge>
ui/src/lib/pages/node/Endpoint.svelte:168: <li><Badge variant="muted" title={t(l.user ? "controller.inventory.user_label" : "controller.inventory.fixed_label")}>{l.label}: {l.value}</Badge></li>
ui/src/lib/pages/node/Endpoint.svelte:180: <details open={d.clusters.length <= OPEN_CLUSTERS} data-testid={`endpoint-cluster-${c.id_hex}`}>
ui/src/lib/pages/node/Fabrics.svelte:200: <Badge variant="warning" title={t("node.fabrics.node_id_rounded_hint")}>{t("node.fabrics.node_id_rounded")}</Badge>
ui/src/lib/pages/node/Pair.svelte:148: <div class="max-w-full rounded border border-[var(--ha-divider-color)] p-2" aria-label={t("node.pair.qr_caption")} data-testid="pair-qr">
ui/src/lib/pages/node/Pair.svelte:159: <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_manual_code")} onclick={() => void copy(win.manual_code)}>
ui/src/lib/pages/node/Pair.svelte:168: <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_qr_payload")} onclick={() => void copy(win.qr_code)}>
ui/src/lib/pages/node/Selection.svelte:220: aria-label={t("node.selection.group_count", { count: rows.length })}>{rows.length}</span
ui/src/lib/pages/node/Selection.svelte:232: title={bulkable(item) ? undefined : t("node.selection.unmappable_checkbox_title")}
ui/src/lib/pages/node/Selection.svelte:238: aria-label={t("node.selection.select_row")}
ui/src/lib/pages/node/Selection.svelte:243: <span class="inline-flex justify-center"><Icon name={si.icon} size={18} class={si.color} aria-label={si.label} title={si.label} /></span>
ui/src/lib/pages/node/Selection.svelte:385: aria-label={t("node.selection.drawer_aria")}
ui/src/lib/pages/node/Selection.svelte:389: <button type="button" class="rounded-md p-2 hover:bg-[var(--ha-secondary-background-color)]" onclick={() => (drawer = null)} aria-label={t("common.close")}>✕</button>
ui/src/lib/pages/node/Selection.svelte:402: title={item.verdict.state === "unmappable" ? t("node.selection.unmappable_checkbox_title") : undefined}
ui/src/lib/pages/node/Status.svelte:82: <Badge variant={s.healthy ? "success" : "danger"} title={s.health_note}>
ui/src/lib/pages/node/Surface.svelte:132: <details class="rounded border border-[var(--ha-divider-color)] p-3" data-testid={testid}>
ui/src/lib/pages/node/Surface.svelte:203: <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]" title={r.policies}>{t("node.surface.policies", { digest: shortDigest(r.policies) })}</span>
ui/src/lib/pages/node/Surface.svelte:269: <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]" title={`${c.before} → ${c.after}`}>
ui/src/lib/pages/operation/Operation.svelte:128: <span class="text-[var(--ha-secondary-text-color)]" title={info.started_at}>{t("operation.uptime", { uptime: formatUptime(t, info.uptime_s) })}</span>
ui/src/lib/pages/operation/Operation.svelte:144: <Input class="w-auto min-w-[12rem] flex-1" placeholder={t("operation.filter_text")} aria-label={t("operation.filter_text")} bind:value={text} />
ui/src/shell/App.svelte:212: title={family === "controller" ? t("controller.title") : family === "operation" ? t("operation.title") : t("node.title")}
ui/src/shell/App.svelte:213: subtitle={family === "controller" ? t("controller.subtitle") : family === "operation" ? t("operation.subtitle") : t("node.subtitle")}
ui/src/shell/Toaster.svelte:17: <button type="button" class="text-[var(--ha-secondary-text-color)]" aria-label={closeLabel} onclick={() => dismiss(e.id)}>✕</button>
```

### class from values

```
ui/src/lib/kit/AttributeTable.svelte:13: let { attributes, class: className }: Props = $props();
ui/src/lib/kit/AttributeTable.svelte:19:<div class={cn("overflow-x-auto", className)}>
ui/src/lib/kit/Badge.svelte:11: let { variant = "default", class: className, title, children }: Props = $props();
ui/src/lib/kit/Badge.svelte:24:<span class={cn("inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium", variants[variant], className)} {title}>
ui/src/lib/kit/Button.svelte:14: let { variant = "default", size = "md", class: className, children, ...rest }: Props = $props();
ui/src/lib/kit/Button.svelte:41: class={cn(
ui/src/lib/kit/Card.svelte:8: let { class: className, children, ...rest }: Props = $props();
ui/src/lib/kit/Card.svelte:12: class={cn("rounded-[var(--ha-radius-card)] border border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)] shadow-[var(--ha-elevation-card)]", className)}
ui/src/lib/kit/ClusterView.svelte:13: let { cluster, class: className }: Props = $props();
ui/src/lib/kit/ClusterView.svelte:23:<div class={cn("space-y-3", className)} data-testid={`cluster-view-${cluster.id_hex}`}>
ui/src/lib/kit/EmptyState.svelte:10: let { message, description, icon = "mdi:text-box-search-outline", class: className, action }: Props = $props();
ui/src/lib/kit/EmptyState.svelte:13:<div class={cn("flex flex-col items-center justify-center gap-3 py-12 text-center", className)}>
ui/src/lib/kit/ErrorState.svelte:10: let { message, onRetry, class: className }: Props = $props();
ui/src/lib/kit/ErrorState.svelte:14:<Card class={cn("flex items-start gap-3 p-3", className)}>
ui/src/lib/kit/EventList.svelte:14: let { events, empty, class: className }: Props = $props();
ui/src/lib/kit/EventList.svelte:24: <p class={cn("text-xs text-[var(--ha-secondary-text-color)]", className)} data-testid="event-list-empty">{empty}</p>
ui/src/lib/kit/EventList.svelte:26: <ul class={cn("divide-y divide-[var(--ha-divider-color)]", className)} data-testid="event-list">
ui/src/lib/kit/Graph.svelte:29: let { graph, hrefOf = () => undefined, follow = (href) => window.location.assign(href), label, class: className }: Props = $props();
ui/src/lib/kit/Graph.svelte:163:<div class={cn("flex flex-col gap-3 md:flex-row", className)} onkeydown={onKey}>
ui/src/lib/kit/Icon.svelte:7: let { name, size = 18, class: className = "", "aria-label": ariaLabel, title }: Props = $props();
ui/src/lib/kit/Icon.svelte:12:<Cmp {size} class={className} aria-label={ariaLabel} aria-hidden={ariaLabel ? undefined : "true"} {title} />
ui/src/lib/kit/Input.svelte:8: let { class: className, value = $bindable(), ...rest }: Props = $props();
ui/src/lib/kit/Input.svelte:13: class={cn(
ui/src/lib/kit/LoadingState.svelte:8: let { label, message, class: className }: Props = $props();
ui/src/lib/kit/LoadingState.svelte:12:<div class={cn("flex items-center justify-center gap-2 py-8 text-sm text-[var(--ha-secondary-text-color)]", className)} role="status" aria-live="polite">
ui/src/lib/kit/PageHeader.svelte:16: let { title, subtitle, actions, children, class: cls = "" }: Props = $props();
ui/src/lib/kit/PageHeader.svelte:19:<header class={cn("mb-4 flex flex-wrap items-start justify-between gap-3", cls)}>
ui/src/lib/kit/PageShell.svelte:9: let { width = "default", class: cls = "", children }: Props = $props();
ui/src/lib/kit/PageShell.svelte:14:<section class={cn("mx-auto w-full px-4 py-6 sm:px-6", widths[width], cls)}>
ui/src/lib/kit/Select.svelte:27: class: className,
ui/src/lib/kit/Select.svelte:36: class={cn(
ui/src/lib/kit/Spinner.svelte:8: let { size = 18, class: className }: Props = $props();
ui/src/lib/kit/Spinner.svelte:11:<span class={cn("fui-loader text-[var(--ha-secondary-text-color)]", className)} style="--fui-size:{size}px" aria-hidden="true">
ui/src/lib/kit/Tabs.svelte:19: let { items, active, variant = "underline", fill = false, ariaLabel, onSelect, class: cls = "" }: Props = $props();
ui/src/lib/kit/Tabs.svelte:49:<div class={listClass} role="tablist" aria-label={ariaLabel}>
ui/src/lib/kit/Tabs.svelte:53: <a href={item.href} role="tab" aria-selected={isActive} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
ui/src/lib/kit/Tabs.svelte:59: <button type="button" role="tab" aria-selected={isActive} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
ui/src/lib/kit/TagChips.svelte:12: let { tags, empty, class: className }: Props = $props();
ui/src/lib/kit/TagChips.svelte:20: <ul class={cn("flex flex-wrap gap-1", className)} data-testid="tag-chips">
ui/src/lib/kit/TagChips.svelte:26: <p class={cn("text-xs text-[var(--ha-secondary-text-color)]", className)}>{empty}</p>
ui/src/lib/pages/controller/node/Inventory.svelte:100: <Card class={["p-4", focus === ne.endpoint ? "ring-2 ring-[var(--ha-primary-color)]" : ""].join(" ")} data-testid={`inventory-endpoint-${ne.endpoint}`}>
ui/src/lib/pages/node/Fabrics.svelte:191: <span class={fabric.label ? "text-[var(--ha-primary-text-color)]" : "text-[var(--ha-secondary-text-color)]"}>
ui/src/lib/pages/node/Selection.svelte:243: <span class="inline-flex justify-center"><Icon name={si.icon} size={18} class={si.color} aria-label={si.label} title={si.label} /></span>
ui/src/lib/pages/node/Selection.svelte:352: <span class="inline-flex items-center gap-1"><Icon name={li.icon} size={14} class={li.color} /><span>{li.label}</span></span>
ui/src/lib/theme/ha-bridge.ts:129: ourRoot.classList.toggle("dark", dark);
ui/src/shell/App.svelte:176: const onScheme = (e: MediaQueryListEvent) => document.documentElement.classList.toggle("dark", e.matches);
ui/src/shell/index.html:87: root.classList.toggle("dark", dark);
```

### clipboard

```
ui/src/lib/pages/controller/QrScan.svelte:5: it explains that instead of offering the button, as the clipboard
ui/src/lib/pages/controller/node/Window.svelte:49: await navigator.clipboard.writeText(v);
ui/src/lib/pages/node/Pair.svelte:55: // The clipboard API is unavailable on a page served over plain HTTP (an
ui/src/lib/pages/node/Pair.svelte:59: await navigator.clipboard.writeText(value);
ui/src/lib/pages/operation/Operation.svelte:102: await navigator.clipboard.writeText(shown.map(recordLine).join("\n"));
ui/src/lib/pages/operation/log.ts:2:// filtering, and the text a copy puts on the clipboard.
```

### native dialogs

```
ui/src/lib/context.ts:142: confirm: opts.confirm ?? (async (o) => window.confirm(o.body ? `${o.title}\n\n${o.body}` : o.title)),
ui/src/lib/pages/controller/Nodes.svelte:91: const ok = await confirm({
ui/src/lib/pages/controller/Settings.svelte:69: const ok = await confirm({ title: t("controller.settings.credential_remove"), body: c.name, destructive: true, confirmLabel: t("common.remove") });
ui/src/lib/pages/controller/node/Acl.svelte:88: const ok = await confirm({ title: t("controller.acl.lockout"), body: t("controller.acl.lockout_body"), destructive: true, confirmLabel: t("common.save") });
ui/src/lib/pages/controller/node/Fabrics.svelte:34: const ok = await confirm({
ui/src/lib/pages/controller/node/Icd.svelte:41: const ok = await confirm({ title: t("controller.icd.multi_admin"), body: errorText(err), confirmLabel: t("controller.icd.register") });
ui/src/lib/pages/node/Fabrics.svelte:67: const ok = await confirm({
ui/src/lib/pages/node/Fabrics.svelte:116: const first = await confirm({
ui/src/lib/pages/node/Fabrics.svelte:123: const second = await confirm({
ui/src/lib/pages/node/Fabrics.svelte:144: const ok = await confirm({
ui/src/lib/pages/node/NodeManage.svelte:62: const first = await confirm({
ui/src/lib/pages/node/NodeManage.svelte:69: const second = await confirm({
```

## 2. API, token and admin

### token

```
ui/src/lib/api/client.ts:144:/** The header the admin token travels in (token mode). */
ui/src/lib/api/client.ts:145:export const TOKEN_HEADER = "X-Fabric-UI-Token";
ui/src/lib/api/client.ts:168: /** Sets the admin token sent on admin routes; status() sets it from csrf_token. */
ui/src/lib/api/client.ts:169: setToken(token: string | undefined): void;
ui/src/lib/api/client.ts:205: * admin token is shared.
ui/src/lib/api/client.ts:229: /** The controller's status; it sets the admin token from csrf_token, as status() does. 404 when the host serves no controller. */
ui/src/lib/api/client.ts:290: // One token for every node's client: the handler issues one per process.
ui/src/lib/api/client.ts:291: const auth: { token?: string } = {};
ui/src/lib/api/client.ts:296: if (opts.admin && auth.token) headers[TOKEN_HEADER] = auth.token;
ui/src/lib/api/client.ts:332: auth.token = t || undefined;
ui/src/lib/api/client.ts:336: if (s.csrf_token) auth.token = s.csrf_token;
ui/src/lib/api/client.ts:378: if (s.csrf_token) auth.token = s.csrf_token;
ui/src/lib/index.ts:3:// ./i18n, ./theme, ./theme/tokens.css.
ui/src/lib/kit/Badge.svelte:13: // A 15% tint of the matching token as the fill and the token as the text
ui/src/lib/kit/Button.svelte:2: tokens instead of the slate ramp, and the label on the primary fill uses --ha-on-primary-color (Home Assistant's
ui/src/lib/kit/DataTable.svelte:4: tokens.css), groupBy sections and activatable rows. The column filter
ui/src/lib/kit/Graph.svelte:3: colour per state from the theme tokens, pan and zoom on the viewBox
ui/src/lib/kit/Tabs.svelte:2: weights; the active marker uses the primary token so the skin applies. -->
ui/src/lib/pages/node/Diagnostics.svelte:26: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/NodeManage.svelte:6: admin rights (the token of token mode or Status.admin). Below them the
ui/src/lib/pages/node/NodeManage.svelte:27: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/Pair.svelte:5: exposed yet" hint, and the colours are the theme tokens. Below the
ui/src/lib/theme/ha-bridge.ts:5:// is inert and the static defaults of tokens.css [data-skin="ha"] apply.
ui/src/lib/theme/ha-bridge.ts:9:// Variables copied from the parent document. The --ha-* tokens read them
ui/src/lib/theme/ha-bridge.ts:119: // to the default) is dropped, so the token falls back again.
ui/src/lib/theme/tokens.css:2: * The design tokens of go-fabric-ui. Components read only the --ha-*
ui/src/lib/theme/tokens.css:3: * consumption tokens, named after Home Assistant's CSS variables, so a page
ui/src/lib/theme/tokens.css:8: * - html[data-skin="ha"] — the tokens resolve from Home Assistant's own
ui/src/lib/theme/tokens.css:26: * (tokens, dark block), :106-142 (HA skin, HA dark), :144-152 (base
ui/src/lib/theme/tokens.css:32:/* HA-style semantic CSS variables. Components prefer these tokens
ui/src/lib/theme/tokens.css:86:/* Home-Assistant skin. When data-skin="ha" the --ha-* consumption tokens
ui/src/lib/theme/tokens.css:132:/* Base document colours track the theme tokens so the boot paint and any
ui/src/shell/App.svelte:106: // admin rights — the token of token mode, or the host's admin verdict.
ui/src/shell/App.svelte:107: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/shell/app.css:2: * `.dark` class, the package's tokens, and the sources Tailwind scans. A
ui/src/shell/app.css:8:@import "../lib/theme/tokens.css";
```

### admin marks and checks

```
ui/src/lib/api/client.ts:144:/** The header the admin token travels in (token mode). */
ui/src/lib/api/client.ts:168: /** Sets the admin token sent on admin routes; status() sets it from csrf_token. */
ui/src/lib/api/client.ts:187: /** Asks the node to announce its DNS-SD records now (admin; 0.7.0). */
ui/src/lib/api/client.ts:198: /** Erases the node — its fabrics, store and endpoint records — and ends it (admin; 0.7.0): 501 when the host wires no erase. */
ui/src/lib/api/client.ts:205: * admin token is shared.
ui/src/lib/api/client.ts:212: /** Creates a node in the host's node registry (admin; 0.6.0): 409 node_limit when the host serves no further node. */
ui/src/lib/api/client.ts:214: /** Renames a managed node (admin; 0.6.0): 409 node_is_default for the default node. */
ui/src/lib/api/client.ts:216: /** Deletes a managed node with its fabrics, store and selection (admin; 0.6.0): 409 node_is_default for the default node. */
ui/src/lib/api/client.ts:229: /** The controller's status; it sets the admin token from csrf_token, as status() does. 404 when the host serves no controller. */
ui/src/lib/api/client.ts:241: /** Whether the node answers a Matter read of DataModelRevision, and how fast (admin; 0.5.0). */
ui/src/lib/api/client.ts:269: /** The controller's network as a graph; refresh re-reads the nodes' diagnostics first (admin). 501 when the host serves no graph. */
ui/src/lib/api/client.ts:293: async function call<T>(method: string, path: string, opts: { body?: unknown; admin?: boolean } = {}): Promise<T> {
ui/src/lib/api/client.ts:296: if (opts.admin && auth.token) headers[TOKEN_HEADER] = auth.token;
ui/src/lib/api/client.ts:340: removeFabric: (i) => call("DELETE", `${np}/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
ui/src/lib/api/client.ts:341: pairing: () => call("GET", np + "/pairing", { admin: true }),
ui/src/lib/api/client.ts:342: openWindow: (d) => call("POST", np + "/pairing/window", { admin: true, body: { duration_seconds: d } }),
ui/src/lib/api/client.ts:343: closeWindow: () => call("POST", np + "/pairing/window/close", { admin: true }),
ui/src/lib/api/client.ts:351: reannounce: () => call("POST", np + "/mdns/reannounce", { admin: true }),
ui/src/lib/api/client.ts:357: const res = await call<Applied | undefined>("PUT", np + "/exposure", { admin: true, body: { items } });
ui/src/lib/api/client.ts:360: previewExposure: (items) => call("POST", np + "/exposure/preview", { admin: true, body: { items } }),
ui/src/lib/api/client.ts:361: reassemble: () => call("POST", np + "/maintenance/reassemble", { admin: true }),
ui/src/lib/api/client.ts:363: call("POST", np + "/maintenance/factory-reset", { admin: true, body: { confirm: FACTORY_RESET_CONFIRMATION } }),
ui/src/lib/api/client.ts:364: erase: () => call("POST", np + "/maintenance/erase", { admin: true, body: { confirm: ERASE_CONFIRMATION } }),
ui/src/lib/api/client.ts:367: createNode: (label) => call("POST", "/nodes", { admin: true, body: { label } }),
ui/src/lib/api/client.ts:368: renameNode: (key, label) => call("PUT", `/nodes/${encodeURIComponent(key)}/label`, { admin: true, body: { label } }),
ui/src/lib/api/client.ts:369: deleteNode: (key) => call("DELETE", `/nodes/${encodeURIComponent(key)}`, { admin: true }),
ui/src/lib/api/client.ts:372: setLogLevel: (level) => call("PUT", "/operation/log/level", { admin: true, body: { level } }),
ui/src/lib/api/client.ts:387: decommissionNode: (id) => call("DELETE", node(id), { admin: true }),
ui/src/lib/api/client.ts:388: interviewNode: (id) => call("POST", node(id) + "/interview", { admin: true }),
ui/src/lib/api/client.ts:389: pingNode: (id) => call("POST", node(id) + "/ping", { admin: true }),
ui/src/lib/api/client.ts:390: enableNode: (id) => call("POST", node(id) + "/enable", { admin: true }),
ui/src/lib/api/client.ts:391: disableNode: (id) => call("POST", node(id) + "/disable", { admin: true }),
ui/src/lib/api/client.ts:392: setNodeAutoSubscribe: (id, enabled) => call("POST", node(id) + "/auto-subscribe", { admin: true, body: { enabled } }),
ui/src/lib/api/client.ts:394: call("POST", node(id) + "/window", { admin: true, body: timeoutSeconds === undefined ? {} : { timeout_seconds: timeoutSeconds } }),
ui/src/lib/api/client.ts:395: closeNodeWindow: (id) => call("POST", node(id) + "/window/close", { admin: true }),
ui/src/lib/api/client.ts:398: writeAttribute: (id, req) => call("PUT", node(id) + "/attributes", { admin: true, body: req }),
ui/src/lib/api/client.ts:399: invokeCommand: (id, req) => call("POST", node(id) + "/commands", { admin: true, body: req }),
ui/src/lib/api/client.ts:401: setNodeAcl: (id, entries) => call("PUT", node(id) + "/acl", { admin: true, body: { entries } }),
ui/src/lib/api/client.ts:403: setNodeBindings: (id, endpoint, bindings) => call("PUT", node(id) + "/bindings", { admin: true, body: { endpoint, bindings } }),
ui/src/lib/api/client.ts:406: call("POST", node(id) + "/icd/register", { admin: true, body: allowMultiAdmin === undefined ? {} : { allow_multi_admin: allowMultiAdmin } }),
ui/src/lib/api/client.ts:408: call("POST", node(id) + "/icd/unregister", { admin: true, body: force === undefined ? {} : { force } }),
ui/src/lib/api/client.ts:410: removeNodeFabric: (id, i) => call("DELETE", `${node(id)}/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
ui/src/lib/api/client.ts:411: commission: (req) => call("POST", "/controller/commission", { admin: true, body: req }),
ui/src/lib/api/client.ts:418: return call("GET", "/controller/discovery" + (qs ? "?" + qs : ""), { admin: true });
ui/src/lib/api/client.ts:421: setCredential: (req) => call("PUT", "/controller/credentials", { admin: true, body: req }),
ui/src/lib/api/client.ts:423: call("DELETE", `/controller/credentials?${new URLSearchParams({ kind, id }).toString()}`, { admin: true }),
ui/src/lib/api/client.ts:424: setFabricLabel: (label) => call("PUT", "/controller/fabric/label", { admin: true, body: { label } }),
ui/src/lib/api/client.ts:426: consentUpdate: (req) => call("POST", "/controller/updates/consent", { admin: true, body: req }),
ui/src/lib/api/client.ts:427: checkUpdates: () => call("POST", "/controller/updates/check", { admin: true }),
ui/src/lib/api/client.ts:429: refresh ? call("GET", "/controller/topology?refresh=true", { admin: true }) : call("GET", "/controller/topology"),
ui/src/lib/pages/controller/Commission.svelte:6: to the admin route only and are cleared after the call. The steps the
ui/src/lib/pages/controller/Network.svelte:8: diagnostics first (refresh=true, an admin route). A host without a
ui/src/lib/pages/controller/Network.svelte:53: // admin rights, 502 when the read fails) keeps the drawn graph.
ui/src/lib/pages/controller/Node.svelte:3: administrator, ICD, the fabrics on it and its pending updates — one
ui/src/lib/pages/controller/Nodes.svelte:2: connection, vendor and ICD, and the actions an administrator takes on
ui/src/lib/pages/controller/Nodes.svelte:3: a node — interview, open a window for another administrator (on the
ui/src/lib/pages/controller/node/Acl.svelte:33: const PRIVILEGES = ["view", "proxy_view", "operate", "manage", "administer"] as const;
ui/src/lib/pages/controller/node/Acl.svelte:82: return list.some((e) => e.privilege === "administer" && e.auth_mode === "case" && (!own || e.subjects_hex.length === 0 || e.subjects_hex.includes(own)));
ui/src/lib/pages/controller/node/Acl.svelte:121: <Badge variant={e.privilege === "administer" ? "warning" : "default"}>{t(`controller.acl.privilege.${e.privilege}`)}</Badge>
ui/src/lib/pages/controller/node/Attributes.svelte:2: inventory), the value as JSON; an administrator can write the value
ui/src/lib/pages/controller/node/Fabrics.svelte:2: administrator's. Ours goes with decommissioning. -->
ui/src/lib/pages/controller/node/Icd.svelte:38: // Other administrators' vendors may not support LIT: ask, then retry.
ui/src/lib/pages/controller/node/Icd.svelte:41: const ok = await confirm({ title: t("controller.icd.multi_admin"), body: errorText(err), confirmLabel: t("controller.icd.register") });
ui/src/lib/pages/controller/node/Overview.svelte:1:<!-- A node's record, properties and connection, and the ping (admin,
ui/src/lib/pages/controller/node/Window.svelte:1:<!-- A commissioning window on the node for another administrator: the
ui/src/lib/pages/node/Diagnostics.svelte:7: Refresh button stays. An admin asks the node to announce its DNS-SD
ui/src/lib/pages/node/Diagnostics.svelte:26: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/NodeManage.svelte:6: admin rights (the token of token mode or Status.admin). Below them the
ui/src/lib/pages/node/NodeManage.svelte:27: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/operation/Operation.svelte:2: what it is (version, go-fabric, uptime), the level it logs at (admin
ui/src/shell/App.svelte:11: admin rights the selector shows (one node is enough) and ends in
ui/src/shell/App.svelte:106: // admin rights — the token of token mode, or the host's admin verdict.
ui/src/shell/App.svelte:107: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
```

### fetch and EventSource

```
ui/src/lib/api/client.ts:288:export function createClient(baseUrl: string, fetchImpl: Fetch = (i, init) => fetch(i, init)): FabricClient {
ui/src/lib/api/client.ts:301: credentials: "same-origin",
ui/src/lib/api/client.ts:420: credentials: () => call("GET", "/controller/credentials"),
ui/src/lib/api/stream.ts:54:/** The part of EventSource the helper uses, so a test can hand in a fake. */
ui/src/lib/api/stream.ts:55:export interface EventSourceLike {
ui/src/lib/api/stream.ts:60:export type EventSourceFactory = (url: string) => EventSourceLike;
ui/src/lib/api/stream.ts:64: * an event whose data is not JSON is skipped. The browser's EventSource
ui/src/lib/api/stream.ts:70: open: EventSourceFactory = (u) => new EventSource(u, { withCredentials: true }),
ui/src/lib/api/stream.ts:114: open: EventSourceFactory = (u) => new EventSource(u, { withCredentials: true }),
ui/src/lib/api/stream.ts:138: open: EventSourceFactory = (u) => new EventSource(u, { withCredentials: true }),
ui/src/lib/context.ts:121: /** Subscribes to the log stream; the default opens an EventSource on the client's log stream URL. */
ui/src/lib/controller-store.svelte.ts:36: /** Subscribes to the controller stream; the default opens an EventSource on the client's URL. */
ui/src/lib/store.svelte.ts:41: /** Subscribes to the event stream; the default opens an EventSource on the client's stream URL. */
```

### client calls from stores, pages and shell

```
ui/src/lib/context.ts:147: subscribeLog: opts.subscribeLog ?? ((fn) => subscribeLogStream(opts.client.logStreamUrl(), fn)),
ui/src/lib/controller-store.svelte.ts:78: status = await client.controllerStatus();
ui/src/lib/controller-store.svelte.ts:94: nodes = (await client.controllerNodes()).nodes;
ui/src/lib/controller-store.svelte.ts:105: credentials = (await client.credentials()).credentials;
ui/src/lib/controller-store.svelte.ts:113: updates = await client.updates();
ui/src/lib/controller-store.svelte.ts:208: const sub = deps.subscribe ?? ((fn) => subscribeControllerStream(client.controllerStreamUrl(), fn));
ui/src/lib/nodes-store.svelte.ts:18: next = (await client.nodes()).nodes;
ui/src/lib/nodes-store.svelte.ts:29: const node = await client.createNode(label);
ui/src/lib/nodes-store.svelte.ts:35: await client.renameNode(key, label);
ui/src/lib/nodes-store.svelte.ts:40: await client.deleteNode(key);
ui/src/lib/pages/controller/Commission.svelte:119: result = await client.commission(withNetwork({ ...req }));
ui/src/lib/pages/controller/Commission.svelte:170: devices = (await client.discover({ timeout: 5 })).devices;
ui/src/lib/pages/controller/Diagnostics.svelte:54: const [d, m, ev] = await Promise.all([client.controllerDiagnostics(), client.controllerMdns(), client.controllerEvents()]);
ui/src/lib/pages/controller/Network.svelte:42: graph = await client.controllerTopology();
ui/src/lib/pages/controller/Network.svelte:57: graph = await client.controllerTopology(true);
ui/src/lib/pages/controller/Node.svelte:44: detail = await client.controllerNode(nodeId);
ui/src/lib/pages/controller/Nodes.svelte:97: if (ok) await act(n, () => client.decommissionNode(n.node_id_hex), "controller.nodes.removed");
ui/src/lib/pages/controller/Nodes.svelte:158: <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.interviewNode(n.node_id_hex), "controller.nodes.interviewed")}>
ui/src/lib/pages/controller/Nodes.svelte:163: <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.enableNode(n.node_id_hex), "controller.nodes.enabled")}>
ui/src/lib/pages/controller/Nodes.svelte:167: <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.disableNode(n.node_id_hex), "controller.nodes.disabled_done")}>
ui/src/lib/pages/controller/Settings.svelte:37: await client.setFabricLabel(label.trim());
ui/src/lib/pages/controller/Settings.svelte:57: await client.setCredential(req);
ui/src/lib/pages/controller/Settings.svelte:72: await client.removeCredential(c.kind, c.id);
ui/src/lib/pages/controller/node/Acl.svelte:39: entries = (await client.nodeAcl(nodeId)).entries;
ui/src/lib/pages/controller/node/Acl.svelte:93: const res = await client.setNodeAcl(nodeId, entries);
ui/src/lib/pages/controller/node/Attributes.svelte:45: const res = await client.readAttributes(nodeId, [p]);
ui/src/lib/pages/controller/node/Attributes.svelte:67: const res = await client.writeAttribute(nodeId, { ...p, value });
ui/src/lib/pages/controller/node/Bindings.svelte:29: lists = (await client.nodeBindings(nodeId)).endpoints;
ui/src/lib/pages/controller/node/Bindings.svelte:43: const res = await client.setNodeBindings(nodeId, eb.endpoint, bindings.map(({ fabric_index: _f, ...b }) => b));
ui/src/lib/pages/controller/node/Events.svelte:36: const l = await client.controllerNodeEvents(nodeId, last === undefined ? {} : { since: last + 1 });
ui/src/lib/pages/controller/node/Fabrics.svelte:26: fabrics = (await client.nodeFabrics(nodeId)).fabrics;
ui/src/lib/pages/controller/node/Fabrics.svelte:42: await client.removeNodeFabric(nodeId, f.fabric_index);
ui/src/lib/pages/controller/node/Icd.svelte:25: icd = await client.nodeIcd(nodeId);
ui/src/lib/pages/controller/node/Icd.svelte:35: icd = await client.registerNodeIcd(nodeId, allowMultiAdmin);
ui/src/lib/pages/controller/node/Icd.svelte:54: icd = await client.unregisterNodeIcd(nodeId, force);
ui/src/lib/pages/controller/node/Inventory.svelte:42: const value = await client.controllerEndpointDetail(nodeId, ep);
ui/src/lib/pages/controller/node/Inventory.svelte:53: const value = await client.controllerClusterDetail(nodeId, ep, hex);
ui/src/lib/pages/controller/node/Overview.svelte:39: ping = await client.pingNode(n.node_id_hex);
ui/src/lib/pages/controller/node/Overview.svelte:51: await client.setNodeAutoSubscribe(n.node_id_hex, on);
ui/src/lib/pages/controller/node/Updates.svelte:26: available = (await client.checkUpdates()).updates.filter((u) => u.node_id_hex === nodeId);
ui/src/lib/pages/controller/node/Updates.svelte:38: await client.consentUpdate({ node_id_hex: u.node_id_hex, vendor_id: u.vendor_id, product_id: u.product_id, target_software_version: u.software_version });
ui/src/lib/pages/controller/node/Window.svelte:26: win = await client.openNodeWindow(nodeId, Number(timeout));
ui/src/lib/pages/controller/node/Window.svelte:37: await client.closeNodeWindow(nodeId);
ui/src/lib/pages/node/Diagnostics.svelte:48: const [s, m, c, ev] = await Promise.all([client.sessions(), client.mdns(), client.compatibility(), client.events()]);
ui/src/lib/pages/node/Diagnostics.svelte:68: await client.reannounce();
ui/src/lib/pages/node/EcosystemHints.svelte:21: compat = await client.compatibility();
ui/src/lib/pages/node/Endpoint.svelte:47: const d = await client.nodeEndpointDetail(endpoint);
ui/src/lib/pages/node/Endpoint.svelte:62: const l = await client.nodeEndpointEvents(endpoint, last === undefined ? {} : { since: last + 1 });
ui/src/lib/pages/node/Endpoints.svelte:36: const [eps, rep] = await Promise.all([client.endpoints(), client.surface().catch(() => null)]);
ui/src/lib/pages/node/Fabrics.svelte:56: await client.reassemble();
ui/src/lib/pages/node/Fabrics.svelte:76: await client.factoryReset();
ui/src/lib/pages/node/Fabrics.svelte:102: await client.status();
ui/src/lib/pages/node/Fabrics.svelte:132: await client.erase();
ui/src/lib/pages/node/Fabrics.svelte:152: await client.removeFabric(f.fabric_index);
ui/src/lib/pages/node/Surface.svelte:36: client.surface(),
ui/src/lib/pages/node/Surface.svelte:37: client.endpoints().then(
ui/src/lib/pages/node/Topology.svelte:34: graph = await client.topology();
ui/src/lib/pages/operation/Operation.svelte:53: const [i, l, log] = await Promise.all([client.operationInfo(), client.logLevel(), client.logRecords({ limit: KEEP_RECORDS })]);
ui/src/lib/pages/operation/Operation.svelte:88: level = (await client.setLogLevel(next as LogLevelName)).level;
ui/src/lib/store.svelte.ts:113: status = await client.status();
ui/src/lib/store.svelte.ts:126: fabrics = (await client.fabrics()).fabrics;
ui/src/lib/store.svelte.ts:138: const [s, list] = await Promise.all([client.exposureSchema(), client.exposure()]);
ui/src/lib/store.svelte.ts:212: return (await client.previewExposure(updates)).items;
ui/src/lib/store.svelte.ts:218: const applied = await client.applyExposure(updates);
ui/src/lib/store.svelte.ts:246: const win = await client.openWindow(durationSeconds);
ui/src/lib/store.svelte.ts:258: st = await client.status();
ui/src/lib/store.svelte.ts:262: const p = await client.pairing();
ui/src/lib/store.svelte.ts:283: await client.closeWindow();
ui/src/lib/store.svelte.ts:364: const sub = deps.subscribe ?? ((fn) => subscribeStream(client.streamUrl(), fn));
ui/src/shell/App.svelte:156: await ui.client.status();
ui/src/shell/App.svelte:167: await ui.client.operationInfo();
ui/src/shell/NodeScope.svelte:21: const client = parent.client.forNode(key);
```

## 3. Parse sites

### parsing and conversion

```
ui/src/lib/api/client.ts:300: body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
ui/src/lib/api/client.ts:305: const data: unknown = text ? JSON.parse(text) : undefined;
ui/src/lib/api/client.ts:326: const np = nodeKey === undefined ? "/node" : `/nodes/${encodeURIComponent(nodeKey)}`;
ui/src/lib/api/client.ts:340: removeFabric: (i) => call("DELETE", `${np}/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
ui/src/lib/api/client.ts:368: renameNode: (key, label) => call("PUT", `/nodes/${encodeURIComponent(key)}/label`, { admin: true, body: { label } }),
ui/src/lib/api/client.ts:369: deleteNode: (key) => call("DELETE", `/nodes/${encodeURIComponent(key)}`, { admin: true }),
ui/src/lib/api/client.ts:385: call("GET", `${node(id)}/endpoints/${endpoint}/clusters/${encodeURIComponent(clusterHex)}`),
ui/src/lib/api/client.ts:410: removeNodeFabric: (id, i) => call("DELETE", `${node(id)}/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
ui/src/lib/api/client.ts:451: return `/controller/nodes/${encodeURIComponent(id)}`;
ui/src/lib/api/stream.ts:77: data = JSON.parse(String(msg.data));
ui/src/lib/api/stream.ts:121: data = JSON.parse(String(msg.data));
ui/src/lib/api/stream.ts:144: data = JSON.parse(String(msg.data));
ui/src/lib/context.ts:79: for (const r of NODE_ROUTES) out[r] = routes[r].replace(/^#\/node\//, `#/node/${encodeURIComponent(key)}/`);
ui/src/lib/kit/AttributeTable.svelte:16: const format = (v: unknown) => JSON.stringify(v, null, 2);
ui/src/lib/kit/DataTable.svelte:63: if (raw) base = { ...base, ...(JSON.parse(raw) as Partial<Persisted>) };
ui/src/lib/kit/DataTable.svelte:80: localStorage.setItem("fabric-ui.table:" + persistKey, JSON.stringify(snapshot));
ui/src/lib/kit/EventList.svelte:20: const format = (v: unknown) => JSON.stringify(v, null, 2);
ui/src/lib/kit/utils.ts:22: re = new RegExp(q, "i");
ui/src/lib/pages/controller/Commission.svelte:82: const [kind, ...rest] = network.split(":");
ui/src/lib/pages/controller/Commission.svelte:98: const version = Number(v);
ui/src/lib/pages/controller/Commission.svelte:99: const acknowledgements = Number(a);
ui/src/lib/pages/controller/Commission.svelte:180: const pin = Number((passcodes[d.instance_name] ?? "").trim());
ui/src/lib/pages/controller/Diagnostics.svelte:106: { key: "flags", label: t("controller.diag.col_flags"), get: (s) => Number(!!s.resumed) + 2 * Number(!!s.peer_lost) },
ui/src/lib/pages/controller/Diagnostics.svelte:112: { key: "reachable", label: t("controller.diag.col_reachable"), sortable: true, get: (p) => Number(p.reachable) },
ui/src/lib/pages/controller/format.ts:48: return parseInt(t, 16);
ui/src/lib/pages/controller/node/Acl.svelte:48: const subs = subjects.split(/[\s,]+/).filter(Boolean);
ui/src/lib/pages/controller/node/Acl.svelte:62: if (targetEndpoint.trim()) target.endpoint = Number(targetEndpoint);
ui/src/lib/pages/controller/node/Attributes.svelte:34: return { endpoint: Number(endpoint), cluster_hex: hexId(c), attribute_hex: hexId(a) };
ui/src/lib/pages/controller/node/Attributes.svelte:47: edit = result && result.status === undefined ? JSON.stringify(result.value) : "";
ui/src/lib/pages/controller/node/Attributes.svelte:60: value = JSON.parse(edit);
ui/src/lib/pages/controller/node/Attributes.svelte:92: <pre class="overflow-x-auto rounded bg-[var(--ha-secondary-background-color)] p-2 font-mono text-xs">{JSON.stringify(result.value, null, 2)}</pre>
ui/src/lib/pages/controller/node/Bindings.svelte:56: if (/^g:\d+$/i.test(target)) b.group = Number(target.slice(2));
ui/src/lib/pages/controller/node/Bindings.svelte:59: b.endpoint = Number(d.endpoint);
ui/src/lib/pages/controller/node/Inventory.svelte:77: const [ep, hex] = key.split("/");
ui/src/lib/pages/controller/node/Inventory.svelte:78: void loadCluster(Number(ep), hex);
ui/src/lib/pages/controller/node/Window.svelte:20: const options = $derived(["180", "300", "600", "900"].map((s) => ({ value: s, label: t("controller.window.minutes", { minutes: Number(s) / 60 }) })));
ui/src/lib/pages/controller/node/Window.svelte:26: win = await client.openNodeWindow(nodeId, Number(timeout));
ui/src/lib/pages/node/Endpoint.svelte:116: const name = detail?.clusters.find((c) => c.id === Number(id))?.name;
ui/src/lib/pages/node/Endpoint.svelte:119: const deviceTypeLabel = (id: string) => detail?.device_types.find((d) => d.id === Number(id))?.name ?? hexId(id);
ui/src/lib/pages/node/Pair.svelte:112: onValueChange={(v) => (selectedDuration = Number(v))}
ui/src/lib/pages/node/Surface.svelte:63: const deviceTypeLabel = (id: string) => deviceTypeNames.get(Number(id)) ?? hexId(id);
ui/src/lib/pages/node/Surface.svelte:65: const name = clusterNames.get(Number(id));
ui/src/lib/pages/node/Surface.svelte:99: { key: "id", label: t("node.surface.col_cluster"), title: true, sortable: true, align: "left", get: (c) => Number(c.id) },
ui/src/lib/pages/node/fields.ts:21: .split(",")
ui/src/lib/pages/node/surface.ts:23: const n = Number(id);
ui/src/lib/theme/ha-bridge.ts:84: const r = parseInt(hex.slice(0, 2), 16);
ui/src/lib/theme/ha-bridge.ts:85: const g = parseInt(hex.slice(2, 4), 16);
ui/src/lib/theme/ha-bridge.ts:86: const b = parseInt(hex.slice(4, 6), 16);
ui/src/lib/theme/ha-bridge.ts:91: const m = v.match(/rgba?\(\s*([\d.]+)[\s,]+([\d.]+)[\s,]+([\d.]+)/i);
ui/src/lib/theme/ha-bridge.ts:92: return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : null;
ui/src/shell/index.html:69: if (hex.length >= 6) rgb = [parseInt(hex.slice(0, 2), 16), parseInt(hex.slice(2, 4), 16), parseInt(hex.slice(4, 6), 16)];
ui/src/shell/index.html:71: var m = bg.match(/rgba?\(\s*([\d.]+)[\s,]+([\d.]+)[\s,]+([\d.]+)/i);
ui/src/shell/index.html:72: if (m) rgb = [Number(m[1]), Number(m[2]), Number(m[3])];
ui/src/shell/route.ts:23: const [path, query = ""] = hash.replace(/^#\/?/, "").split("?");
ui/src/shell/route.ts:24: const parts = path.split("/").filter(Boolean);
ui/src/shell/route.ts:41: key = decodeURIComponent(rest[0]);
ui/src/shell/route.ts:49: const n = Number(rest[1]);
```

## 4. Life cycle

### effects and mount hooks

```
ui/src/lib/kit/DataTable.svelte:76: $effect(() => {
ui/src/lib/kit/DataTable.svelte:128: const processed = $derived.by(() => {
ui/src/lib/kit/DataTable.svelte:150: const groups = $derived.by(() => {
ui/src/lib/pages/controller/Commission.svelte:21: import { onMount, tick } from "svelte";
ui/src/lib/pages/controller/Commission.svelte:58: onMount(() => {
ui/src/lib/pages/controller/Diagnostics.svelte:12: import { onMount } from "svelte";
ui/src/lib/pages/controller/Diagnostics.svelte:71: onMount(() => void load());
ui/src/lib/pages/controller/Diagnostics.svelte:75: $effect(() => {
ui/src/lib/pages/controller/Network.svelte:14: import { onMount, untrack } from "svelte";
ui/src/lib/pages/controller/Network.svelte:67: onMount(load);
ui/src/lib/pages/controller/Network.svelte:69: const startEpoch = untrack(() => controller.graphEpoch);
ui/src/lib/pages/controller/Network.svelte:70: $effect(() => {
ui/src/lib/pages/controller/Network.svelte:71: if (controller.graphEpoch !== startEpoch) untrack(() => void load());
ui/src/lib/pages/controller/Node.svelte:6: import { untrack } from "svelte";
ui/src/lib/pages/controller/Node.svelte:50: $effect(() => {
ui/src/lib/pages/controller/Node.svelte:52: untrack(() => void load());
ui/src/lib/pages/controller/Node.svelte:56: $effect(() => {
ui/src/lib/pages/controller/Node.svelte:58: if (untrack(() => controller.changed) === nodeId && untrack(() => detail) !== null) untrack(() => void load());
ui/src/lib/pages/controller/Nodes.svelte:8: import { onMount } from "svelte";
ui/src/lib/pages/controller/Nodes.svelte:35: $effect(() => {
ui/src/lib/pages/controller/Nodes.svelte:42: onMount(() => {
ui/src/lib/pages/controller/QrScan.svelte:8: import { onDestroy, tick } from "svelte";
ui/src/lib/pages/controller/QrScan.svelte:51: onDestroy(cancel);
ui/src/lib/pages/controller/Settings.svelte:5: import { onMount } from "svelte";
ui/src/lib/pages/controller/Settings.svelte:28: onMount(async () => {
ui/src/lib/pages/controller/node/Acl.svelte:5: import { onMount } from "svelte";
ui/src/lib/pages/controller/node/Acl.svelte:45: onMount(() => void load());
ui/src/lib/pages/controller/node/Bindings.svelte:5: import { onMount } from "svelte";
ui/src/lib/pages/controller/node/Bindings.svelte:34: onMount(() => void load());
ui/src/lib/pages/controller/node/Endpoints.svelte:15: const ordered = $derived.by(() => {
ui/src/lib/pages/controller/node/Events.svelte:9: import { untrack } from "svelte";
ui/src/lib/pages/controller/node/Events.svelte:49: $effect(() => {
ui/src/lib/pages/controller/node/Events.svelte:51: untrack(() => {
ui/src/lib/pages/controller/node/Fabrics.svelte:4: import { onMount } from "svelte";
ui/src/lib/pages/controller/node/Fabrics.svelte:31: onMount(() => void load());
ui/src/lib/pages/controller/node/Icd.svelte:4: import { onMount } from "svelte";
ui/src/lib/pages/controller/node/Icd.svelte:30: onMount(() => void load());
ui/src/lib/pages/controller/node/Inventory.svelte:8: import { untrack } from "svelte";
ui/src/lib/pages/controller/node/Inventory.svelte:72: $effect(() => {
ui/src/lib/pages/controller/node/Inventory.svelte:74: untrack(() => {
ui/src/lib/pages/controller/node/Inventory.svelte:84: $effect(() => {
ui/src/lib/pages/controller/node/Inventory.svelte:86: untrack(() => document.querySelector(`[data-testid="inventory-endpoint-${focus}"]`)?.scrollIntoView?.({ block: "start" }));
ui/src/lib/pages/controller/node/Updates.svelte:4: import { onMount } from "svelte";
ui/src/lib/pages/controller/node/Updates.svelte:19: onMount(() => void controller.loadUpdates());
ui/src/lib/pages/node/Diagnostics.svelte:12: import { onMount } from "svelte";
ui/src/lib/pages/node/Diagnostics.svelte:63: onMount(() => void load());
ui/src/lib/pages/node/Diagnostics.svelte:99: $effect(() => {
ui/src/lib/pages/node/EcosystemHints.svelte:8: import { onMount } from "svelte";
ui/src/lib/pages/node/EcosystemHints.svelte:19: onMount(async () => {
ui/src/lib/pages/node/Endpoint.svelte:11: import { untrack } from "svelte";
ui/src/lib/pages/node/Endpoint.svelte:76: $effect(() => {
ui/src/lib/pages/node/Endpoint.svelte:78: untrack(refresh);
ui/src/lib/pages/node/Endpoint.svelte:83: const startAttr = untrack(() => node.attributeEpoch(endpoint));
ui/src/lib/pages/node/Endpoint.svelte:84: const startTopo = untrack(() => node.topologyEpoch);
ui/src/lib/pages/node/Endpoint.svelte:85: $effect(() => {
ui/src/lib/pages/node/Endpoint.svelte:88: untrack(() => {
ui/src/lib/pages/node/Endpoint.svelte:97: const startEvent = untrack(() => node.clusterEvent);
ui/src/lib/pages/node/Endpoint.svelte:98: $effect(() => {
ui/src/lib/pages/node/Endpoint.svelte:101: untrack(() => {
ui/src/lib/pages/node/Endpoints.svelte:11: import { onMount, untrack } from "svelte";
ui/src/lib/pages/node/Endpoints.svelte:47: onMount(load);
ui/src/lib/pages/node/Endpoints.svelte:51: const startEpoch = untrack(() => node.topologyEpoch);
ui/src/lib/pages/node/Endpoints.svelte:52: $effect(() => {
ui/src/lib/pages/node/Endpoints.svelte:53: if (node.topologyEpoch !== startEpoch) untrack(() => void load());
ui/src/lib/pages/node/Fabrics.svelte:11: import { onDestroy, onMount } from "svelte";
ui/src/lib/pages/node/Fabrics.svelte:28: onMount(() => {
ui/src/lib/pages/node/Fabrics.svelte:90: onDestroy(() => {
ui/src/lib/pages/node/NodeManage.svelte:13: import { onMount } from "svelte";
ui/src/lib/pages/node/NodeManage.svelte:33: $effect(() => {
ui/src/lib/pages/node/NodeManage.svelte:39: onMount(() => {
ui/src/lib/pages/node/Pair.svelte:8: import { onMount } from "svelte";
ui/src/lib/pages/node/Pair.svelte:27: onMount(async () => {
ui/src/lib/pages/node/Selection.svelte:15: import { onMount } from "svelte";
ui/src/lib/pages/node/Selection.svelte:41: onMount(() => {
ui/src/lib/pages/node/Selection.svelte:51: $effect(() => {
ui/src/lib/pages/node/Selection.svelte:76: const valueFilters = $derived.by<ValueFilter[]>(() => {
ui/src/lib/pages/node/Selection.svelte:99: const filtered = $derived.by(() => {
ui/src/lib/pages/node/Status.svelte:20: import { onMount } from "svelte";
ui/src/lib/pages/node/Status.svelte:37: onMount(() => {
ui/src/lib/pages/node/Surface.svelte:8: import { onMount, untrack } from "svelte";
ui/src/lib/pages/node/Surface.svelte:51: onMount(load);
ui/src/lib/pages/node/Surface.svelte:54: const startEpoch = untrack(() => node.topologyEpoch);
ui/src/lib/pages/node/Surface.svelte:55: $effect(() => {
ui/src/lib/pages/node/Surface.svelte:56: if (node.topologyEpoch !== startEpoch) untrack(() => void load());
ui/src/lib/pages/node/Topology.svelte:9: import { onMount, untrack } from "svelte";
ui/src/lib/pages/node/Topology.svelte:42: onMount(load);
ui/src/lib/pages/node/Topology.svelte:44: const startEpoch = untrack(() => node.graphEpoch);
ui/src/lib/pages/node/Topology.svelte:45: $effect(() => {
ui/src/lib/pages/node/Topology.svelte:46: if (node.graphEpoch !== startEpoch) untrack(() => void load());
ui/src/lib/pages/operation/Operation.svelte:9: import { onMount } from "svelte";
ui/src/lib/pages/operation/Operation.svelte:66: onMount(() => {
ui/src/shell/App.svelte:19: import { onDestroy, onMount } from "svelte";
ui/src/shell/App.svelte:136: $effect(() => {
ui/src/shell/App.svelte:178: onMount(() => {
ui/src/shell/App.svelte:192: $effect(() => {
ui/src/shell/App.svelte:195: $effect(() => {
ui/src/shell/App.svelte:199: onDestroy(() => {
ui/src/shell/NodeScope.svelte:9: import { onDestroy, onMount, type Snippet } from "svelte";
ui/src/shell/NodeScope.svelte:30: onMount(() => {
ui/src/shell/NodeScope.svelte:35: onDestroy(() => scoped?.node.close());
```

### listeners, observers, aborts, closes

```
ui/src/lib/api/stream.ts:56: addEventListener(type: string, fn: (ev: MessageEvent) => void): void;
ui/src/lib/api/stream.ts:74: es.addEventListener(type, (msg: MessageEvent) => {
ui/src/lib/api/stream.ts:84: return () => es.close();
ui/src/lib/api/stream.ts:118: es.addEventListener(type, (msg: MessageEvent) => {
ui/src/lib/api/stream.ts:128: return () => es.close();
ui/src/lib/api/stream.ts:141: es.addEventListener("log_record", (msg: MessageEvent) => {
ui/src/lib/api/stream.ts:150: return () => es.close();
ui/src/lib/theme/ha-bridge.ts:138: const observer = new MutationObserver(() => apply());
ui/src/lib/theme/ha-bridge.ts:141: window.addEventListener("focus", onFocus);
ui/src/lib/theme/ha-bridge.ts:142: document.addEventListener("visibilitychange", onFocus);
ui/src/lib/theme/ha-bridge.ts:144: observer.disconnect();
ui/src/lib/theme/ha-bridge.ts:145: window.removeEventListener("focus", onFocus);
ui/src/lib/theme/ha-bridge.ts:146: document.removeEventListener("visibilitychange", onFocus);
ui/src/shell/App.svelte:179: window.addEventListener("hashchange", onHash);
ui/src/shell/App.svelte:184: media.addEventListener("change", onScheme);
ui/src/shell/App.svelte:200: window.removeEventListener("hashchange", onHash);
ui/src/shell/App.svelte:201: media?.removeEventListener("change", onScheme);
ui/src/shell/App.svelte:203: node.close();
ui/src/shell/App.svelte:204: controller.close();
ui/src/shell/NodeScope.svelte:35: onDestroy(() => scoped?.node.close());
```

### timers

```
ui/src/lib/controller-store.svelte.ts:68: let reload: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/controller-store.svelte.ts:69: let graphTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/controller-store.svelte.ts:70: let diagnosticTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/controller-store.svelte.ts:121: reload = setTimeout(() => {
ui/src/lib/controller-store.svelte.ts:131: if (graphTimer) clearTimeout(graphTimer);
ui/src/lib/controller-store.svelte.ts:132: graphTimer = setTimeout(() => {
ui/src/lib/controller-store.svelte.ts:141: if (diagnosticTimer) clearTimeout(diagnosticTimer);
ui/src/lib/controller-store.svelte.ts:142: diagnosticTimer = setTimeout(() => {
ui/src/lib/controller-store.svelte.ts:213: for (const timer of [reload, graphTimer, diagnosticTimer]) if (timer) clearTimeout(timer);
ui/src/lib/pages/node/Fabrics.svelte:89: let poll: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/pages/node/Fabrics.svelte:91: if (poll) clearTimeout(poll);
ui/src/lib/pages/node/Fabrics.svelte:99: poll = setTimeout(async () => {
ui/src/lib/pages/node/Status.svelte:39: const tick = setInterval(() => (now = Date.now()), 1000);
ui/src/lib/pages/node/Status.svelte:40: return () => clearInterval(tick);
ui/src/lib/scan/index.ts:56: let timer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/scan/index.ts:59: if (timer) clearTimeout(timer);
ui/src/lib/scan/index.ts:77: timer = setTimeout(() => void tick(), 200);
ui/src/lib/store.svelte.ts:99: let countdown: ReturnType<typeof setInterval> | null = null;
ui/src/lib/store.svelte.ts:100: let exposureReload: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:101: let closedTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:102: let graphTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:103: let attributeTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:104: let healthTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:105: let diagnosticTimer: ReturnType<typeof setTimeout> | null = null;
ui/src/lib/store.svelte.ts:151: if (exposureReload) clearTimeout(exposureReload);
ui/src/lib/store.svelte.ts:152: exposureReload = setTimeout(() => {
ui/src/lib/store.svelte.ts:161: if (graphTimer) clearTimeout(graphTimer);
ui/src/lib/store.svelte.ts:162: graphTimer = setTimeout(() => {
ui/src/lib/store.svelte.ts:171: attributeTimer = setTimeout(() => {
ui/src/lib/store.svelte.ts:181: healthTimer = setTimeout(() => {
ui/src/lib/store.svelte.ts:189: diagnosticTimer = setTimeout(() => {
ui/src/lib/store.svelte.ts:226: clearInterval(countdown);
ui/src/lib/store.svelte.ts:235: countdown = setInterval(() => {
ui/src/lib/store.svelte.ts:298: clearTimeout(closedTimer);
ui/src/lib/store.svelte.ts:334: closedTimer = setTimeout(() => {
ui/src/lib/store.svelte.ts:370: for (const timer of [exposureReload, closedTimer, graphTimer, attributeTimer, healthTimer, diagnosticTimer]) if (timer) clearTimeout(timer);
ui/src/shell/toast.svelte.ts:15: if (severity !== "error") setTimeout(() => dismiss(id), 4000);
```

### subscriptions and media

```
ui/src/lib/api/client.ts:392: setNodeAutoSubscribe: (id, enabled) => call("POST", node(id) + "/auto-subscribe", { admin: true, body: { enabled } }),
ui/src/lib/api/stream.ts:67:export function subscribeStream(
ui/src/lib/api/stream.ts:108: * Subscribes to the controller stream at `url`, as subscribeStream does
ui/src/lib/api/stream.ts:111:export function subscribeControllerStream(
ui/src/lib/api/stream.ts:135:export function subscribeLogStream(
ui/src/lib/context.ts:7:import { subscribeLogStream } from "./api/stream";
ui/src/lib/context.ts:110: subscribeLog: (onRecord: (rec: LogRecord) => void) => () => void;
ui/src/lib/context.ts:119: subscribe?: NodeStoreDeps["subscribe"];
ui/src/lib/context.ts:120: subscribeController?: ControllerStoreDeps["subscribe"];
ui/src/lib/context.ts:122: subscribeLog?: FabricUI["subscribeLog"];
ui/src/lib/context.ts:143: node: createNodeStore(opts.client, { t, toast, subscribe: opts.subscribe, onNodesChanged: () => void nodes.load() }),
ui/src/lib/context.ts:145: controller: createControllerStore(opts.client, { t, toast, subscribe: opts.subscribeController }),
ui/src/lib/context.ts:147: subscribeLog: opts.subscribeLog ?? ((fn) => subscribeLogStream(opts.client.logStreamUrl(), fn)),
ui/src/lib/controller-store.svelte.ts:15:import { subscribeControllerStream, type ControllerStreamEvent } from "./api/stream";
ui/src/lib/controller-store.svelte.ts:37: subscribe?: (onEvent: (ev: ControllerStreamEvent) => void) => () => void;
ui/src/lib/controller-store.svelte.ts:208: const sub = deps.subscribe ?? ((fn) => subscribeControllerStream(client.controllerStreamUrl(), fn));
ui/src/lib/pages/controller/Diagnostics.svelte:158: {#if p.subscription_active}<Badge variant="muted">{t("controller.diag.subscribed")}</Badge>{/if}
ui/src/lib/pages/controller/Nodes.svelte:152: {#if n.subscription_active}<Badge variant="default">{t("controller.nodes.subscribed")}</Badge>{/if}
ui/src/lib/pages/controller/QrScan.svelte:16: const canScan = typeof window !== "undefined" && window.isSecureContext === true && !!navigator.mediaDevices?.getUserMedia;
ui/src/lib/pages/controller/node/Overview.svelte:72: {#if n.subscription_active}<Badge>{t("controller.nodes.subscribed")}</Badge>{/if}
ui/src/lib/pages/controller/node/Overview.svelte:81: {n.subscription_active ? t("controller.overview.unsubscribe") : t("controller.overview.subscribe")}
ui/src/lib/pages/operation/Operation.svelte:23: const { t, toast, client, subscribeLog } = useFabricUI();
ui/src/lib/pages/operation/Operation.svelte:69: const close = subscribeLog(append);
ui/src/lib/scan/index.ts:49: const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" }, audio: false });
ui/src/lib/scan/index.ts:60: for (const track of stream.getTracks()) track.stop();
ui/src/lib/store.svelte.ts:19:import { subscribeStream, type StreamEvent } from "./api/stream";
ui/src/lib/store.svelte.ts:42: subscribe?: (onEvent: (ev: StreamEvent) => void) => () => void;
ui/src/lib/store.svelte.ts:364: const sub = deps.subscribe ?? ((fn) => subscribeStream(client.streamUrl(), fn));
```

## 5. Storage and globals

### storage

```
ui/src/lib/kit/DataTable.svelte:39: /** When set, sort and search persist under this localStorage key. */
ui/src/lib/kit/DataTable.svelte:62: const raw = localStorage.getItem("fabric-ui.table:" + persistKey);
ui/src/lib/kit/DataTable.svelte:80: localStorage.setItem("fabric-ui.table:" + persistKey, JSON.stringify(snapshot));
ui/src/lib/kit/utils.ts:33:/** Reads a string from localStorage, `fallback` when absent or unavailable. */
ui/src/lib/kit/utils.ts:36: return localStorage.getItem(key) ?? fallback;
ui/src/lib/kit/utils.ts:42:/** Best-effort write to localStorage. */
ui/src/lib/kit/utils.ts:45: localStorage.setItem(key, value);
```

### parent frame

```
ui/src/lib/theme/ha-bridge.ts:57: if (window.self === window.top) return false;
ui/src/lib/theme/ha-bridge.ts:63: return window.parent.document.querySelector("home-assistant") !== null;
ui/src/lib/theme/ha-bridge.ts:109: const parentWin = window.parent;
ui/src/shell/index.html:55: if (window.self !== window.top) {
ui/src/shell/index.html:56: var parentWin = window.parent;
```

### console

```
ui/src/lib/context.ts:128: info: (m) => console.info(m),
ui/src/lib/context.ts:129: success: (m) => console.info(m),
ui/src/lib/context.ts:130: error: (m) => console.error(m),
```

## 6. i18n keys

    ui/src/lib/i18n/en.json	750
    ui/src/lib/i18n/de.json	750

### Literal t("…") keys missing from en.json

```
```

### en keys missing from de.json / de keys not in en.json

```
--
```

### Template keys (dynamic; checked by prefix)

```
controller.acl.mode.${e.auth_mode}
controller.acl.mode.${m}
controller.acl.privilege.${e.privilege}
controller.acl.privilege.${p}
controller.commission.ble_skipped.${ble.reason ?? "no_central"}
controller.commission.step.${step.phase}
controller.connection.${s}
controller.credential.kind_${c.kind}
controller.icd.mode.${icd.operating_mode}
controller.inventory.conformance.${ep.conformance.status}
controller.node.tab.${k}
controller.nodes.tab.${k}
controller.overview.flag.${f}
controller.settings.cap.${k}
controller.updates.status.${u.status}
events.priority.${e.priority}
graph.kind.${k}
graph.state.${s}
graph.strength.${e.strength ?? "unknown"}
inventory.status.${IM_STATUS[a.status]}
node.diag.ecosystem.${eco.ecosystem}
node.diag.ecosystem.${finding.ecosystem}
node.diag.ecosystem.${hint.ecosystem}
node.diag.kind_${event.kind}
node.diag.severity.${finding.severity}
node.pair.hints.${hint.ecosystem}
node.pair.hints.${hint.ecosystem}.${link.key}
node.status.boot_reason.${s.boot_reason}
node.surface.${label}
node.surface.change_${changeKind(ch.before, ch.after)}
node.surface.kind.${d.kind}
node.surface.kind.${e.kind}
node.surface.verdict.${d.verdict}
node.surface.verdict.${dec.verdict}
node.surface.verdict.${v}
node.tab.${p}
operation.level.${l}
```

### en keys never seen as a literal string in the source (candidates for unused or dynamic-only)

```
graph.detail.address
graph.detail.addresses
graph.detail.bridge
graph.detail.bssid
graph.detail.clusters
graph.detail.collected_at
graph.detail.device_type
graph.detail.device_type_revision
graph.detail.devices
graph.detail.endpoint
graph.detail.endpoints
graph.detail.ext_address
graph.detail.ext_pan_id
graph.detail.fabric_id
graph.detail.fabrics
graph.detail.host_name
graph.detail.key
graph.detail.kind
graph.detail.label
graph.detail.last_seen
graph.detail.model_name
graph.detail.network
graph.detail.network_name
graph.detail.node_id
graph.detail.nodes
graph.detail.peer_idle_seconds
graph.detail.peer_node_id
graph.detail.rloc16
graph.detail.role
graph.detail.session_type
graph.detail.sessions
graph.detail.ssid
graph.detail.subscriptions
graph.detail.vendor
graph.detail.vendor_id
graph.detail.vendor_name
node.status.load_error
```
