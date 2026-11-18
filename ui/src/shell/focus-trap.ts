// A modal's keyboard handling for the shell's dialogs (UR-P1-ui-117): on
// mount the focus moves into the dialog (to the element marked
// data-autofocus, else the first control), Tab and Shift+Tab stay inside,
// Escape calls onEscape, and on destroy the focus returns to where it was.
const FOCUSABLE = 'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export function trapFocus(node: HTMLElement, onEscape?: () => void) {
  const before = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  let escape = onEscape;
  const focusables = () => [...node.querySelectorAll<HTMLElement>(FOCUSABLE)];
  const first = node.querySelector<HTMLElement>("[data-autofocus]") ?? focusables()[0];
  first?.focus();

  function onKey(e: KeyboardEvent) {
    if (e.key === "Escape") {
      e.preventDefault();
      escape?.();
      return;
    }
    if (e.key !== "Tab") return;
    const list = focusables();
    if (list.length === 0) return;
    const i = list.indexOf(document.activeElement as HTMLElement);
    const next = e.shiftKey ? (i <= 0 ? list.length - 1 : i - 1) : i === list.length - 1 || i < 0 ? 0 : i + 1;
    e.preventDefault();
    list[next].focus();
  }
  node.addEventListener("keydown", onKey);
  return {
    update(next?: () => void) {
      escape = next;
    },
    destroy() {
      node.removeEventListener("keydown", onKey);
      if (before?.isConnected) before.focus();
    },
  };
}
