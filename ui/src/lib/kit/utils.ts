// Class composition and text matching for the kit. Ported from
// openccu-loom assets/ui/src/lib/utils.ts:1-65 at b40b289a.
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** Tailwind class composition: conditional classes, conflicts resolved. */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/**
 * A case-insensitive predicate over a free-text search term. The literal
 * substring always matches; when the term is also a valid regular
 * expression it matches by that too, so `lamp|light` finds both while a
 * pasted name with parentheses still finds itself.
 */
export function makeTextMatcher(term: string): (text: string) => boolean {
  const q = term.trim();
  if (!q) return () => true;
  let re: RegExp | null = null;
  try {
    re = new RegExp(q, "i");
  } catch {
    re = null;
  }
  const lower = q.toLowerCase();
  return (text: string) => {
    if (text.toLowerCase().includes(lower)) return true;
    return re ? re.test(text) : false;
  };
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
