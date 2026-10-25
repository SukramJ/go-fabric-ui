// The commissioning window's countdown from the deadline the status
// reports (window_deadline, on the node's clock), so a reload or a page
// opened late counts down what is left instead of the full duration.

/**
 * Whole seconds left until deadline at now (milliseconds since the epoch),
 * never negative; null without a deadline or for one that does not parse.
 */
export function secondsUntil(deadline: string | undefined, now: number = Date.now()): number | null {
  if (!deadline) return null;
  const t = Date.parse(deadline);
  if (Number.isNaN(t)) return null;
  return Math.max(0, Math.ceil((t - now) / 1000));
}
