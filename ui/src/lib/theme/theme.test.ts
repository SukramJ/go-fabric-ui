// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/lib/theme/ha-bridge.test.ts and
// skin-remap.test.ts at b40b289a (the shell-frame cases are loom's own).
import fs from "node:fs";
import path from "node:path";
import { afterEach, describe, expect, it, vi } from "vitest";
import { applySkin, isEmbedded, parseRgb, relativeLuminance, resolveSkin, startHaBridge } from "./ha-bridge";

// Tests run from ui/; under happy-dom import.meta.url is not a file URL.
const css = fs.readFileSync(path.resolve(process.cwd(), "src/lib/theme/tokens.css"), "utf8");

function frame(opts: { path?: string; parent?: unknown }) {
  vi.spyOn(window, "top", "get").mockReturnValue({} as Window);
  if (opts.path) window.history.replaceState(null, "", opts.path);
  vi.spyOn(window, "parent", "get").mockReturnValue(opts.parent as Window);
}

function haParent(vars: Record<string, string>) {
  const root = document.createElement("html");
  const doc = { documentElement: root, querySelector: (s: string) => (s === "home-assistant" ? {} : null) };
  return {
    document: doc,
    getComputedStyle: () => ({ getPropertyValue: (n: string) => vars[n] ?? "", backgroundColor: "" }),
  };
}

afterEach(() => {
  vi.restoreAllMocks();
  window.history.replaceState(null, "", "/");
  document.documentElement.removeAttribute("style");
  document.documentElement.className = "";
  delete document.documentElement.dataset.skin;
});

describe("isEmbedded", () => {
  it("is false when window.self === window.top (standalone)", () => {
    expect(isEmbedded()).toBe(false);
  });
  it("is true framed under the Ingress path", () => {
    frame({ path: "/api/hassio_ingress/abc/", parent: { document: { querySelector: () => null } } });
    expect(isEmbedded()).toBe(true);
  });
  it("is true framed by a same-origin parent that is the Home Assistant frontend", () => {
    frame({ parent: haParent({}) });
    expect(isEmbedded()).toBe(true);
  });
  it("is false in an unknown frame and in a cross-origin frame off the Ingress path", () => {
    frame({ parent: { document: { querySelector: () => null } } });
    expect(isEmbedded()).toBe(false);
    frame({
      parent: {
        get document(): never {
          throw new Error("cross-origin");
        },
      },
    });
    expect(isEmbedded()).toBe(false);
  });
});

describe("resolveSkin and applySkin", () => {
  it("returns the stored skin standalone and forces ha when embedded", () => {
    expect(resolveSkin("fabric")).toBe("fabric");
    expect(applySkin("fabric")).toBe("fabric");
    expect(document.documentElement.dataset.skin).toBeUndefined();
    frame({ parent: haParent({}) });
    expect(resolveSkin("fabric")).toBe("ha");
    applySkin("fabric");
    expect(document.documentElement.dataset.skin).toBe("ha");
  });
});

describe("startHaBridge", () => {
  it("is a no-op standalone", () => {
    const stop = startHaBridge();
    expect(document.documentElement.getAttribute("style")).toBeNull();
    stop();
  });
  it("copies the theme variables of a same-origin parent and follows its dark background", () => {
    frame({ parent: haParent({ "--primary-color": "#ff0000", "--primary-background-color": "#111111" }) });
    const stop = startHaBridge();
    expect(document.documentElement.style.getPropertyValue("--primary-color")).toBe("#ff0000");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    stop();
  });
  it("swallows a throwing parent and returns a no-op cleanup", () => {
    vi.spyOn(window, "top", "get").mockReturnValue({} as Window);
    window.history.replaceState(null, "", "/api/hassio_ingress/abc/");
    vi.spyOn(window, "parent", "get").mockReturnValue({
      get document(): never {
        throw new Error("cross-origin");
      },
    } as unknown as Window);
    expect(() => startHaBridge()()).not.toThrow();
  });
  it("parses colours and their luminance", () => {
    expect(parseRgb("#fff")).toEqual([255, 255, 255]);
    expect(parseRgb("rgb(1, 2, 3)")).toEqual([1, 2, 3]);
    expect(parseRgb("nope")).toBeNull();
    expect(relativeLuminance([0, 0, 0])).toBe(0);
    expect(relativeLuminance([255, 255, 255])).toBeCloseTo(1);
  });
});

describe("tokens.css", () => {
  it("defines the --ha-* tokens, the dark block and the HA skin with Home Assistant fallbacks", () => {
    expect(css).toMatch(/:root\s*\{[^}]*--ha-primary-color:/);
    expect(css).toMatch(/html\.dark\s*\{[^}]*--ha-card-background-color:/);
    expect(css).toContain("--ha-primary-color: var(--primary-color, #009ac7);");
    expect(css).toContain("--ha-elevation-card: var(--ha-card-box-shadow, none);");
  });
  it("remaps the palette under the HA skin without a .dark variant and leaves white and black alone", () => {
    const remap = css.slice(css.indexOf("--color-slate-50"));
    expect(remap).toContain("--color-slate-500: #727272;");
    expect(remap).toContain("--color-brand-500: #009ac7;");
    expect(css).not.toContain("--color-white");
    expect(css).not.toContain("--color-black");
  });
  it("carries the phone reflow of data tables", () => {
    expect(css).toContain(".table-reflow td::before");
  });
});
