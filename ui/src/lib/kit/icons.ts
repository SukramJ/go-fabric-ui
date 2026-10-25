// The icon registry: Home Assistant's mdi-style names mapped onto Lucide
// components, so a template names an icon once. Ported from openccu-loom
// assets/ui/src/lib/icons.ts:1-20, :205-300 and resolveIcon at b40b289a,
// reduced to the icons the node pages use.
import type { Component } from "svelte";
import Ban from "@lucide/svelte/icons/ban";
import Circle from "@lucide/svelte/icons/circle";
import CircleAlert from "@lucide/svelte/icons/circle-alert";
import CircleCheck from "@lucide/svelte/icons/circle-check";
import Info from "@lucide/svelte/icons/info";
import LinkIcon from "@lucide/svelte/icons/link";
import List from "@lucide/svelte/icons/list";
import ListChecks from "@lucide/svelte/icons/list-checks";

export type IconName =
  | "mdi:alert-circle"
  | "mdi:cancel"
  | "mdi:check-circle"
  | "mdi:circle-outline"
  | "mdi:format-list-bulleted"
  | "mdi:information-outline"
  | "mdi:link"
  | "mdi:list-checks"
  | "mdi:text-box-search-outline";

const REGISTRY: Record<IconName, Component> = {
  "mdi:alert-circle": CircleAlert as unknown as Component,
  "mdi:cancel": Ban as unknown as Component,
  "mdi:check-circle": CircleCheck as unknown as Component,
  "mdi:circle-outline": Circle as unknown as Component,
  "mdi:format-list-bulleted": List as unknown as Component,
  "mdi:information-outline": Info as unknown as Component,
  "mdi:link": LinkIcon as unknown as Component,
  "mdi:list-checks": ListChecks as unknown as Component,
  "mdi:text-box-search-outline": List as unknown as Component,
};

/** The component for an icon name; an unknown name renders the info icon. */
export function resolveIcon(name: IconName): Component {
  return REGISTRY[name] ?? (Info as unknown as Component);
}
