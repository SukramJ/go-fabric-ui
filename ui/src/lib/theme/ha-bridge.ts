// The Home Assistant theme bridge. Inside Home Assistant's Ingress frame
// (same origin) it mirrors the live theme: it copies Home Assistant's CSS
// variables off the parent document onto our :root and follows its
// light/dark choice by the parent's background luminance. Anywhere else it
// is inert and the static defaults of tokens.css [data-skin="ha"] apply.
// Ported from openccu-loom assets/ui/src/lib/theme/ha-bridge.ts:1-184 at
// b40b289a, without the shell-frame detection that is loom's own.

// Variables copied from the parent document. The --ha-* tokens read them
// through var(--…, fallback), so a live value overrides the default.
export const HA_THEME_VARS = [
  "--primary-color",
  "--accent-color",
  "--primary-background-color",
  "--secondary-background-color",
  "--card-background-color",
  "--primary-text-color",
  "--secondary-text-color",
  "--disabled-text-color",
  "--divider-color",
  "--error-color",
  "--warning-color",
  "--success-color",
  "--info-color",
  "--ha-card-border-radius",
  "--ha-card-box-shadow",
  "--app-header-background-color",
  "--app-header-text-color",
  "--sidebar-background-color",
  "--rgb-primary-color",
  "--text-primary-color",
  "--ha-font-family-body",
  "--header-height",
  "--rgb-accent-color",
  "--rgb-primary-text-color",
  "--rgb-error-color",
  "--rgb-success-color",
  "--rgb-info-color",
  "--shadow-color",
  "--ha-color-text-primary",
  "--ha-color-text-secondary",
  "--ha-color-surface-default",
] as const;

// The path segment the Supervisor serves an add-on's Ingress under.
const HA_INGRESS_SEGMENT = "/api/hassio_ingress/";

export type Skin = "fabric" | "ha";

// True when the page runs inside Home Assistant, detected positively: a
// frame alone is not enough. Either the page is served under the Ingress
// path, or a same-origin parent document is Home Assistant's frontend
// (`<home-assistant>`).
export function isEmbedded(): boolean {
  if (typeof window === "undefined") return false;
  try {
    if (window.self === window.top) return false;
  } catch {
    // top inaccessible — still framed
  }
  if (window.location.pathname.includes(HA_INGRESS_SEGMENT)) return true;
  try {
    return window.parent.document.querySelector("home-assistant") !== null;
  } catch {
    // Cross-origin or unreadable parent — not provably Home Assistant.
    return false;
  }
}

// The effective skin: inside Home Assistant always "ha", otherwise the
// stored choice.
export function resolveSkin(stored: Skin): Skin {
  return isEmbedded() ? "ha" : stored;
}

// A CSS colour as [r,g,b] (0–255), or null: #rgb, #rrggbb, rgb(), rgba().
export function parseRgb(value: string): [number, number, number] | null {
  const v = value.trim();
  if (!v) return null;
  if (v[0] === "#") {
    let hex = v.slice(1);
    if (hex.length === 3) hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
    if (hex.length >= 6) {
      const r = parseInt(hex.slice(0, 2), 16);
      const g = parseInt(hex.slice(2, 4), 16);
      const b = parseInt(hex.slice(4, 6), 16);
      if (!Number.isNaN(r) && !Number.isNaN(g) && !Number.isNaN(b)) return [r, g, b];
    }
    return null;
  }
  const m = v.match(/rgba?\(\s*([\d.]+)[\s,]+([\d.]+)[\s,]+([\d.]+)/i);
  return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : null;
}

// WCAG relative luminance (0 black, 1 white) of an sRGB triple.
export function relativeLuminance([r, g, b]: [number, number, number]): number {
  const lin = (c: number) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
}

// Starts mirroring the live theme and returns the cleanup. A no-op outside
// Home Assistant and for a parent that is or becomes inaccessible.
export function startHaBridge(): () => void {
  if (!isEmbedded()) return () => {};
  try {
    const parentWin = window.parent;
    const parentRoot = parentWin.document.documentElement;
    const ourRoot = document.documentElement;

    const apply = () => {
      try {
        const ps = parentWin.getComputedStyle(parentRoot);
        for (const name of HA_THEME_VARS) {
          const value = ps.getPropertyValue(name).trim();
          if (value) ourRoot.style.setProperty(name, value);
        }
        // Follow Home Assistant's light/dark by its background luminance;
        // this overrides the page's own choice while embedded.
        const bg = ps.getPropertyValue("--primary-background-color").trim() || ps.backgroundColor;
        const rgb = parseRgb(bg);
        if (rgb) {
          const dark = relativeLuminance(rgb) < 0.5;
          ourRoot.classList.toggle("dark", dark);
          ourRoot.style.colorScheme = dark ? "dark" : "light";
        }
      } catch {
        // The parent became inaccessible — the last good values stay.
      }
    };

    apply();
    const observer = new MutationObserver(() => apply());
    observer.observe(parentRoot, { attributes: true, attributeFilter: ["style", "class"] });
    const onFocus = () => apply();
    window.addEventListener("focus", onFocus);
    document.addEventListener("visibilitychange", onFocus);
    return () => {
      observer.disconnect();
      window.removeEventListener("focus", onFocus);
      document.removeEventListener("visibilitychange", onFocus);
    };
  } catch {
    return () => {};
  }
}

/** Sets data-skin on <html> to the effective skin and returns it. */
export function applySkin(stored: Skin = "ha", root: HTMLElement = document.documentElement): Skin {
  const skin = resolveSkin(stored);
  if (skin === "ha") root.dataset.skin = "ha";
  else delete root.dataset.skin;
  return skin;
}
