// The event lists' merge: the routes answer "numbered since or later",
// and a node's numbers need not arrive in order or once (UR-P1-ui-131,
// UR-P2-ui-90).
import type { EventRecord } from "../api/client";

/**
 * have with the events of more whose number it does not hold, in number
 * order (a stable sort: a number repeated within one answer stays, both
 * events in their order).
 */
export function mergeEvents(have: EventRecord[], more: EventRecord[]): EventRecord[] {
  const seen = new Set(have.map((e) => e.number));
  return [...have, ...more.filter((e) => !seen.has(e.number))].sort((a, b) => a.number - b.number);
}

/** The highest number of events, the cursor of the next incremental read; undefined for none. */
export function lastNumber(events: EventRecord[]): number | undefined {
  let last: number | undefined;
  for (const e of events) if (last === undefined || e.number > last) last = e.number;
  return last;
}
