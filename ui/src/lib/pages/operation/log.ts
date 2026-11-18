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

/**
 * records with one streamed record added. The stream delivers in order, so
 * a newer number is appended without a sort. A number at or below the
 * newest one is a record already held (dropped), a gap filled late
 * (merged), or the first record of a restarted daemon, which numbers from
 * the start again: then the previous process's records go, so two
 * sequences never interleave.
 */
export function appendRecord(records: LogRecord[], rec: LogRecord): LogRecord[] {
  const last = records.at(-1);
  if (!last || rec.seq > last.seq) {
    const out = [...records, rec];
    return out.length > KEEP_RECORDS ? out.slice(out.length - KEEP_RECORDS) : out;
  }
  const same = records.find((r) => r.seq === rec.seq);
  if (same) return same.msg === rec.msg && same.level === rec.level ? records : [rec];
  if (rec.seq > records[0].seq && rec.time <= last.time) return mergeRecords(records, [rec]);
  return [rec];
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
