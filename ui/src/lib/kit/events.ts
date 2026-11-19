// The event lists' merge: the routes answer "numbered since or later",
// and a node's numbers need not arrive in order or once (UR-P1-ui-131,
// UR-P2-ui-90).
import type { EventRecord } from "../api/client";

/**
 * An event's number, exact: its number_str (0.10.0) read as a bigint; the
 * deprecated number on a server before it.
 */
export function eventNumber(e: EventRecord): bigint {
  if (e.number_str !== undefined && /^\d+$/.test(e.number_str)) return BigInt(e.number_str);
  return BigInt(Math.trunc(e.number));
}

/** An event's number as decimal text, exact (number_str when sent). */
export function eventNumberText(e: EventRecord): string {
  return eventNumber(e).toString();
}

/**
 * have with the events of more whose number it does not hold, in number
 * order (a stable sort: a number repeated within one answer stays, both
 * events in their order).
 */
export function mergeEvents(have: EventRecord[], more: EventRecord[]): EventRecord[] {
  const seen = new Set(have.map(eventNumberText));
  return [...have, ...more.filter((e) => !seen.has(eventNumberText(e)))].sort((a, b) => {
    const x = eventNumber(a);
    const y = eventNumber(b);
    return x < y ? -1 : x > y ? 1 : 0;
  });
}

/** The highest number of events, the cursor of the next incremental read; undefined for none. */
export function lastNumber(events: EventRecord[]): bigint | undefined {
  let last: bigint | undefined;
  for (const e of events) {
    const n = eventNumber(e);
    if (last === undefined || n > last) last = n;
  }
  return last;
}

/** The window of the read after events: since the highest number + 1, as since_str (exact); everything without events. */
export function sinceAfter(events: EventRecord[]): { since?: string } {
  const last = lastNumber(events);
  return last === undefined ? {} : { since: (last + 1n).toString() };
}
