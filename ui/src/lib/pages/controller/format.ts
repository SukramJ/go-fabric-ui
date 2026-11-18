// Helpers the controller pages share: a node's display name, the badge of
// a connection state, ids as the contract carries them.
import type { ConnectionState, ControllerNodeSummary } from "../../api/client";
import type { Translate } from "../../i18n";

/** A node's name: its label, else its product, else its id. */
export function nodeName(n: ControllerNodeSummary): string {
  return n.node_label || n.product_name || `0x${n.node_id_hex}`;
}

/** The vendor of a node or a device: the name the server sends, else the id. */
export function vendorLabel(name: string | undefined, id: number | undefined): string {
  if (name) return name;
  if (id === undefined) return "";
  return `0x${id.toString(16).toUpperCase().padStart(4, "0")}`;
}

export type BadgeVariant = "default" | "success" | "warning" | "danger" | "muted";

/** The badge colour of a connection state. */
export function connectionVariant(s: ConnectionState): BadgeVariant {
  switch (s) {
    case "connected":
      return "success";
    case "reconnecting":
      return "warning";
    case "waiting_for_discovery":
      return "muted";
    default:
      return "danger";
  }
}

/** A connection state's label. */
export function connectionLabel(t: Translate, s: ConnectionState): string {
  return t(`controller.connection.${s}`);
}

/** A 32-bit id as "0x0006", the contract's hex form. */
export function hexId(v: number): string {
  return `0x${v.toString(16).toUpperCase().padStart(4, "0")}`;
}

/**
 * Parses a decimal number of up to five digits within 0..max; null for
 * anything else ("abc", "1.5", "-1", "1e3", a value above max). Endpoint
 * numbers are 16-bit with 0xFFFF reserved (max 65534), group ids 16-bit
 * (max 65535).
 */
export function parseUint16(s: string, max: number): number | null {
  const t = s.trim();
  if (!/^\d{1,5}$/.test(t)) return null;
  const n = Number(t);
  return n <= max ? n : null;
}

/** The largest endpoint number; 0xFFFF is reserved. */
export const MAX_ENDPOINT = 0xfffe;
/** The largest group id. */
export const MAX_GROUP = 0xffff;

/** Parses "0x0006", "6" or "0006" as hex; null when it is not hex. */
export function parseHexId(s: string): number | null {
  const t = s.trim().replace(/^0x/i, "");
  if (!/^[0-9a-f]{1,8}$/i.test(t)) return null;
  return parseInt(t, 16);
}
