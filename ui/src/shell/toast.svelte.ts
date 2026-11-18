// The shell's toasts: a host embedding the pages passes its own.
// Mirrors openccu-loom assets/ui/src/lib/stores/toast.svelte.ts:16-59 at
// b40b289a: info and success leave after 4 s, warnings and errors stay;
// at most MAX_TOASTS at once, a repeated message once, a long one cut.
import type { Toast } from "../lib/context";

export type ToastSeverity = "info" | "success" | "error";
export type ToastEntry = { id: number; severity: ToastSeverity; message: string };

/** The most toasts shown at once; the oldest leave first (UR-P1-bound-111). */
export const MAX_TOASTS = 5;
/** The longest message a toast shows; a longer one (an update's error text) is cut. */
export const MAX_TOAST_CHARS = 300;

export function createToasts() {
  let entries = $state<ToastEntry[]>([]);
  let next = 1;
  function push(severity: ToastSeverity, text: string) {
    const message = text.length > MAX_TOAST_CHARS ? text.slice(0, MAX_TOAST_CHARS - 1) + "…" : text;
    // A message already shown is not shown twice (a stream event repeating a failure).
    if (entries.some((e) => e.severity === severity && e.message === message)) return;
    const id = next++;
    entries = [...entries, { id, severity, message }].slice(-MAX_TOASTS);
    if (severity !== "error") setTimeout(() => dismiss(id), 4000);
  }
  function dismiss(id: number) {
    entries = entries.filter((e) => e.id !== id);
  }
  const toast: Toast = {
    info: (m) => push("info", m),
    success: (m) => push("success", m),
    error: (m) => push("error", m),
  };
  return {
    get entries() {
      return entries;
    },
    dismiss,
    toast,
  };
}
