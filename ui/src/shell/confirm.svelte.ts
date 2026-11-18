// The shell's confirm dialog state: one question at a time, answered by
// ConfirmDialog.svelte. A confirmation within GUARD_MS of a question
// being shown is ignored: the second click of a double click would
// otherwise answer the second step of a two-step confirmation, shown
// where the first one's button was (UR-P1-ui-104).
import type { Confirm, ConfirmOptions } from "../lib/context";

/** How long after a question is shown its confirmation is ignored. */
export const GUARD_MS = 500;

export function createConfirm(now: () => number = () => Date.now()) {
  let current = $state<(ConfirmOptions & { resolve: (ok: boolean) => void }) | null>(null);
  let shownAt = 0;
  const ask: Confirm = (opts) =>
    new Promise<boolean>((resolve) => {
      current?.resolve(false);
      shownAt = now();
      current = { ...opts, resolve };
    });
  function answer(ok: boolean) {
    if (ok && now() - shownAt < GUARD_MS) return;
    current?.resolve(ok);
    current = null;
  }
  return {
    get current() {
      return current;
    },
    ask,
    answer,
  };
}
