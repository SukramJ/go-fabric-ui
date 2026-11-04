// The Operation page's log helpers: merging the ring with the stream,
// filtering, and the text a copy puts on the clipboard.
import type { LogLevelName, LogRecord } from "../../api/client";
import type { Translate } from "../../i18n";

/** The contract's levels, lowest first (LogLevel.levels). */
export const LEVELS: readonly LogLevelName[] = ["debug", "info", "warn", "error"];

/** How many records the page keeps: the route's maximum limit. */
export const KEEP_RECORDS = 2000;

/** a and b merged by sequence number, each record once, oldest first, the newest KEEP_RECORDS. */
export function mergeRecords(a: LogRecord[], b: LogRecord[]): LogRecord[] {
  const bySeq = new Map<number, LogRecord>();
  for (const r of [...a, ...b]) bySeq.set(r.seq, r);
  const out = [...bySeq.values()].sort((x, y) => x.seq - y.seq);
  return out.length > KEEP_RECORDS ? out.slice(out.length - KEEP_RECORDS) : out;
}

/** The records at minLevel or above whose message or attributes contain text (case-insensitive). */
export function visibleRecords(records: LogRecord[], minLevel: LogLevelName, text: string): LogRecord[] {
  const floor = LEVELS.indexOf(minLevel);
  const needle = text.trim().toLowerCase();
  return records.filter((r) => {
    if (LEVELS.indexOf(r.level) < floor) return false;
    if (!needle) return true;
    return recordLine(r).toLowerCase().includes(needle);
  });
}

/** One record as a line of text: time, level, message, attributes. */
export function recordLine(r: LogRecord): string {
  const attrs = Object.entries(r.attrs).map(([k, v]) => `${k}=${v}`);
  return [r.time, r.level.toUpperCase(), r.msg, ...attrs].join(" ");
}

/** seconds as days, hours and minutes, the largest two units. */
export function formatUptime(t: Translate, seconds: number): string {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return t("operation.uptime_days", { d, h });
  if (h > 0) return t("operation.uptime_hours", { h, m });
  return t("operation.uptime_minutes", { m });
}
