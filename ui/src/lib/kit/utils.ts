// Class composition and text matching for the kit. Ported from
// openccu-loom assets/ui/src/lib/utils.ts:1-65 at b40b289a.
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { stringifyJson } from "../api/json";

/** Tailwind class composition: conditional classes, conflicts resolved. */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/**
 * A case-insensitive predicate over a free-text search term: `|` separates
 * alternatives, each matched as a literal substring, so `lamp|light` finds
 * both and a pasted name with parentheses finds itself. The term is never
 * run as a regular expression: a backtracking pattern typed one character
 * at a time against host-supplied text froze the tab (UR-P1-ui-142).
 * Diverges from the port's source, which also tried the term as a regex.
 */
export function makeTextMatcher(term: string): (text: string) => boolean {
  const alternatives = term
    .split("|")
    .map((a) => a.trim().toLowerCase())
    .filter((a) => a !== "");
  if (alternatives.length === 0) return () => true;
  return (text: string) => {
    const lower = text.toLowerCase();
    return alternatives.some((a) => lower.includes(a));
  };
}

/**
 * A generic value (an attribute's, an event's data) as indented JSON, with
 * a 64-bit integer kept as its source text printed as written
 * (stringifyJson; UR-P1-ui-193). stringifyJson refuses an unsafe JS
 * number, since a request must not carry a rounded one; a display shows
 * it as JSON.stringify does rather than fail the page.
 */
export function formatJson(value: unknown): string {
  try {
    return stringifyJson(value, 2);
  } catch {
    return JSON.stringify(value, null, 2);
  }
}

/** Reads a string from localStorage, `fallback` when absent or unavailable. */
export function loadLS(key: string, fallback = ""): string {
  try {
    return localStorage.getItem(key) ?? fallback;
  } catch {
    return fallback;
  }
}

/** Best-effort write to localStorage. */
export function saveLS(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    // storage unavailable — the value simply does not persist.
  }
}
