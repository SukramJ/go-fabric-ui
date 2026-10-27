// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/lib/theme/ha-bridge.test.ts and
// skin-remap.test.ts at b40b289a (the shell-frame cases are loom's own).
import fs from "node:fs";
import path from "node:path";
import { afterEach, describe, expect, it, vi } from "vitest";
import { HA_THEME_VARS, applySkin, isEmbedded, parseRgb, relativeLuminance, resolveSkin, startHaBridge } from "./ha-bridge";

// Tests run from ui/; under happy-dom import.meta.url is not a file URL.
const css = fs.readFileSync(path.resolve(process.cwd(), "src/lib/theme/tokens.css"), "utf8");
const indexHtml = fs.readFileSync(path.resolve(process.cwd(), "src/shell/index.html"), "utf8");
// The shell's pre-paint script: the first inline (non-module) <script>.
const prepaint = indexHtml.match(/<script>([\s\S]*?)<\/script>/)![1];
const runPrepaint = () => new Function(prepaint)();

function frame(opts: { path?: string; parent?: unknown }) {
  vi.spyOn(window, "top", "get").mockReturnValue({} as Window);
  if (opts.path) window.history.replaceState(null, "", opts.path);
  vi.spyOn(window, "parent", "get").mockReturnValue(opts.parent as Window);
}

function haParent(vars: Record<string, string>, root: HTMLElement = document.createElement("html")) {
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
  it("drops a variable the parent no longer defines and follows a live switch", async () => {
    const vars: Record<string, string> = { "--primary-color": "#ff6d00", "--primary-background-color": "#2a0a3d" };
    const parentRoot = document.createElement("html");
    frame({ parent: haParent(vars, parentRoot) });
    const stop = startHaBridge();
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    delete vars["--primary-color"];
    vars["--primary-background-color"] = "#fafafa";
    parentRoot.setAttribute("style", "--primary-background-color:#fafafa");
    await new Promise((r) => setTimeout(r, 0));
    expect(document.documentElement.style.getPropertyValue("--primary-color")).toBe("");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.style.colorScheme).toBe("light");
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

describe("the pre-paint script (index.html)", () => {
  function prefersDark(dark: boolean) {
    vi.spyOn(window, "matchMedia").mockReturnValue({ matches: dark } as MediaQueryList);
  }
  it("copies the same variables the bridge copies", () => {
    const listed = [...prepaint.slice(prepaint.indexOf("var vars = [")).split("];")[0].matchAll(/"(--[a-z0-9-]+)"/g)].map((m) => m[1]);
    expect(listed).toEqual([...HA_THEME_VARS]);
  });
  it("sets the HA skin and follows prefers-color-scheme standalone", () => {
    prefersDark(true);
    runPrepaint();
    expect(document.documentElement.dataset.skin).toBe("ha");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    prefersDark(false);
    runPrepaint();
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.getAttribute("style")).toBeNull();
  });
  it("takes Home Assistant's theme and mode inside the Ingress frame, over the OS preference", () => {
    prefersDark(true);
    frame({ parent: haParent({ "--primary-color": "#ff6d00", "--primary-background-color": "#fafafa" }) });
    runPrepaint();
    expect(document.documentElement.style.getPropertyValue("--primary-color")).toBe("#ff6d00");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.style.colorScheme).toBe("light");
  });
  it("decides dark as the bridge does", () => {
    for (const bg of ["#111111", "#fafafa", "#2a0a3d", "#f3e8ff", "#777", "rgb(120, 120, 120)", "rgba(200, 200, 200, 0.5)", "#bcbcbc"]) {
      vi.restoreAllMocks();
      document.documentElement.className = "";
      prefersDark(false);
      frame({ parent: haParent({ "--primary-background-color": bg }) });
      runPrepaint();
      const want = relativeLuminance(parseRgb(bg)!) < 0.5;
      expect(document.documentElement.classList.contains("dark"), bg).toBe(want);
    }
  });
  it("keeps the OS preference for an unreadable parent", () => {
    prefersDark(true);
    window.history.replaceState(null, "", "/api/hassio_ingress/abc/");
    vi.spyOn(window, "top", "get").mockReturnValue({} as Window);
    vi.spyOn(window, "parent", "get").mockReturnValue({
      get document(): never {
        throw new Error("cross-origin");
      },
    } as unknown as Window);
    expect(() => runPrepaint()).not.toThrow();
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    expect(document.documentElement.dataset.skin).toBe("ha");
  });
});

describe("tokens.css", () => {
  it("gives the label on a primary fill a token: white standalone, Home Assistant's text-primary-color in the skin", () => {
    expect(css).toMatch(/:root\s*\{[^}]*--ha-on-primary-color: #ffffff;/);
    expect(css).toContain("--ha-on-primary-color: var(--text-primary-color, #ffffff);");
    expect(css).toMatch(/html\[data-skin="ha"\]\s*\{[^}]*accent-color: var\(--ha-primary-color\);/);
    expect(css).toMatch(/html\[data-skin="ha"\]\.dark\s*\{[^}]*color-scheme: dark;/);
  });
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
