// The tab descriptor of Tabs.svelte. Ported from openccu-loom
// assets/ui/src/lib/components/ui/tabs.ts:1-19 at b40b289a.
import type { IconName } from "./icons";

export type TabItem = {
  /** Compared against the strip's `active`. */
  key: string;
  label: string;
  /** A link for route-driven tabs; without it the tab is a button reporting through onSelect. */
  href?: string;
  icon?: IconName;
  badge?: string | number;
};
