// The shell's confirm dialog state: one question at a time, answered by
// ConfirmDialog.svelte.
import type { Confirm, ConfirmOptions } from "../lib/context";

export function createConfirm() {
  let current = $state<(ConfirmOptions & { resolve: (ok: boolean) => void }) | null>(null);
  const ask: Confirm = (opts) =>
    new Promise<boolean>((resolve) => {
      current?.resolve(false);
      current = { ...opts, resolve };
    });
  function answer(ok: boolean) {
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
