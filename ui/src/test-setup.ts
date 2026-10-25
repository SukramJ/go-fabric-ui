// Jest-dom matchers for every test running under happy-dom, and a
// localStorage where Node's own global is an empty getter. Ported from
// openccu-loom assets/ui/src/test-setup.ts at b40b289a.
import "@testing-library/jest-dom/vitest";
import { Storage } from "happy-dom";

if (typeof window !== "undefined") {
  const global = globalThis as Record<string, unknown>;
  for (const key of ["localStorage", "sessionStorage"] as const) {
    if (global[key] === undefined) {
      Object.defineProperty(globalThis, key, { value: new Storage(), configurable: true, writable: true });
    }
  }
}
