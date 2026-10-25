import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    // Svelte 5 ships browser and server builds under separate export
    // conditions; without "browser" `render()` gets the server build.
    conditions: ["browser"],
  },
  test: {
    include: ["src/**/*.test.ts", "scripts/**/*.test.mjs"],
    environment: "node",
    globals: false,
    // The first render of a page compiles it; a cold run takes seconds.
    testTimeout: 20000,
    setupFiles: ["src/test-setup.ts"],
  },
});
