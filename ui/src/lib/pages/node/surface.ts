// What the status row and the Surface page share about go-fabric's
// surface report: how the level's origin and the ids read.
import type { Translate } from "../../i18n";

/**
 * The level's origin as the pages say it: go-fabric's
 * surface.Report.LevelOrigin is "explicit", "pin" or
 * "profile <name>[, <name>…]"; anything else is shown as it comes.
 */
export function levelOrigin(t: Translate, origin: string): string {
  if (origin === "explicit") return t("node.status.level_origin_explicit");
  if (origin === "pin") return t("node.status.level_origin_pin");
  if (origin.startsWith("profile ")) return origin.slice("profile ".length);
  return origin;
}

/**
 * An id of the report ("0x" and lowercase hex, as wide as its kind) as the
 * pages print ids: uppercase, at least four digits. A value that is not
 * hex is returned as it is.
 */
export function hexId(id: string, width = 4): string {
  const n = Number(id);
  if (!id.startsWith("0x") || !Number.isFinite(n)) return id;
  return `0x${n.toString(16).toUpperCase().padStart(width, "0")}`;
}

/** The first twelve hex digits of a digest; the page puts the full value in a title. */
export const shortDigest = (digest: string) => digest.slice(0, 12);
