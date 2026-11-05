<!-- The node's configuration on its status page (Status.config, contract
     0.6.0): one read-only block for every node — where the values come
     from, the label, the address, the discriminator, vendor and product
     and the store; never the passcode, which stays on the Pair page. A
     configured node says its values are set in the host's configuration
     (with the host's note, "add-on options"); a registry node says the host
     allocated them and only its name is editable (NodeManage below). -->
<script lang="ts">
  import type { NodeConfig } from "../../api/client";
  import { useFabricUI } from "../../context";
  import Card from "../../kit/Card.svelte";

  let { config }: { config: NodeConfig } = $props();

  const { t } = useFabricUI();

  const hex = (v: number) => "0x" + v.toString(16).toUpperCase().padStart(4, "0");
  const named = (name: string | undefined, id: number) => (name ? `${name} (${hex(id)})` : hex(id));
  const rows = $derived(
    [
      config.label ? [t("node.config.label"), config.label] : null,
      [t("node.config.listen"), config.listen],
      [t("node.config.discriminator"), String(config.discriminator)],
      [t("node.config.vendor"), named(config.vendor_name, config.vendor_id)],
      [t("node.config.product"), named(config.product_name, config.product_id)],
      config.store ? [t("node.config.store"), config.store] : null,
    ].filter((r): r is string[] => r !== null),
  );
  const note = $derived(
    config.source === "registry"
      ? t("node.config.from_registry")
      : config.source_note
        ? t("node.config.from_configuration_note", { note: config.source_note })
        : t("node.config.from_configuration"),
  );
</script>

<Card class="mt-4 p-4" data-testid="node-config">
  <h2 class="mb-1 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.config.title")}</h2>
  <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]" data-testid="node-config-source">{note}</p>
  <dl class="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm">
    {#each rows as [name, value] (name)}
      <dt class="text-[var(--ha-secondary-text-color)]">{name}</dt>
      <dd class="break-all text-[var(--ha-primary-text-color)]">{value}</dd>
    {/each}
  </dl>
</Card>
