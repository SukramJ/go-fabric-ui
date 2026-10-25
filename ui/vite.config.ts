// The standalone shell's build. The output is committed under
// ../fabricui/dist and embedded by the Go module; `make dist-check` rebuilds
// it and fails on any byte of difference, so the build is reproducible:
// pinned versions, `npm ci`, no sourcemaps, content-hashed names.
// Mirrors openccu-loom assets/ui/vite.config.ts at b40b289a (base './', the
// vendor chunk), without its dev-server path prefix.
import fs from "node:fs";
import { fileURLToPath } from "node:url";
import { defineConfig, type Plugin } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";

const here = (p: string) => fileURLToPath(new URL(p, import.meta.url));

// With FABRIC_UI_VENDOR_LIST set, writes the npm packages that land in the
// vendor chunk (name, version, licence) to that file — the list
// THIRD-PARTY-NOTICES.md is checked against (scripts/check-notices.mjs).
// Nothing is added to the bundle.
function vendorList(): Plugin {
  return {
    name: "fabric-ui-vendor-list",
    apply: "build",
    generateBundle(_, bundle) {
      const out = process.env.FABRIC_UI_VENDOR_LIST;
      if (!out) return;
      const pkgs = new Map<string, { name: string; version: string; license: string }>();
      for (const chunk of Object.values(bundle)) {
        if (chunk.type !== "chunk") continue;
        for (const id of Object.keys(chunk.modules)) {
          const m = id.match(/node_modules\/((?:@[^/]+\/)?[^/]+)\//g);
          if (!m) continue;
          const name = m[m.length - 1].slice("node_modules/".length, -1);
          if (pkgs.has(name)) continue;
          const meta = JSON.parse(fs.readFileSync(here(`./node_modules/${name}/package.json`), "utf8"));
          pkgs.set(name, { name, version: meta.version, license: meta.license });
        }
      }
      const list = [...pkgs.values()].sort((a, b) => a.name.localeCompare(b.name));
      fs.writeFileSync(out, JSON.stringify(list, null, 2) + "\n");
    },
  };
}

export default defineConfig({
  root: here("./src/shell"),
  base: "./",
  plugins: [svelte({ configFile: here("./svelte.config.js") }), tailwindcss(), vendorList()],
  build: {
    outDir: process.env.FABRIC_UI_OUT_DIR ?? here("../fabricui/dist"),
    emptyOutDir: true,
    sourcemap: false,
    chunkSizeWarningLimit: 500,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          return id.includes("node_modules") ? "vendor" : undefined;
        },
      },
    },
  },
  server: {
    port: 5174,
    strictPort: true,
    proxy: {
      "/api": { target: process.env.FABRIC_UI_API ?? "http://localhost:8080", changeOrigin: true },
    },
  },
});
