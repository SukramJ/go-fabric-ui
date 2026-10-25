// The shell's toasts: a host embedding the pages passes its own.
// Mirrors openccu-loom assets/ui/src/lib/stores/toast.svelte.ts:16-59 at
// b40b289a: info and success leave after 4 s, warnings and errors stay.
import type { Toast } from "../lib/context";

export type ToastSeverity = "info" | "success" | "error";
export type ToastEntry = { id: number; severity: ToastSeverity; message: string };

export function createToasts() {
  let entries = $state<ToastEntry[]>([]);
  let next = 1;
  function push(severity: ToastSeverity, message: string) {
    const id = next++;
    entries = [...entries, { id, severity, message }];
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
