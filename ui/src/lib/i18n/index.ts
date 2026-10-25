// The catalogues and the translator. A host passes its own `t` through the
// fabric-ui context (it may wrap `createTranslator` with catalogues of its
// own); the standalone shell provides the default. Mirrors openccu-loom
// assets/ui/src/lib/i18n.ts:36, :8688-8719 at b40b289a (format, the
// fallback chain active locale → English → raw key, catalogKeys).
import en from "./en.json";
import de from "./de.json";

export type Catalog = Record<string, string>;
export type Vars = Record<string, string | number>;
export type Translate = (key: string, vars?: Vars) => string;

/** The catalogues the package ships. */
export const catalogs: Record<string, Catalog> = { en, de };

/** Replaces `{name}` placeholders; an unknown one stays as written. */
export function format(s: string, vars?: Vars): string {
  if (!vars) return s;
  return s.replace(/\{(\w+)\}/g, (_, key: string) => {
    const v = vars[key];
    return v == null ? `{${key}}` : String(v);
  });
}

/**
 * A translator over the shipped catalogues plus `extra` (a host's own
 * keys, or overrides of ours, by locale). `locale` is read on every call,
 * so a translator whose locale getter reads reactive state re-renders the
 * text that calls it when the language changes.
 */
export function createTranslator(locale: () => string, extra: Record<string, Catalog> = {}): Translate {
  return (key, vars) => {
    const loc = locale();
    const value = extra[loc]?.[key] ?? catalogs[loc]?.[key] ?? extra.en?.[key] ?? en[key as keyof typeof en] ?? key;
    return format(value, vars);
  };
}

/** The English translator the kit falls back to outside a fabric-ui context. */
export const defaultT: Translate = createTranslator(() => "en");

/** The keys one shipped catalogue defines, without the English fallback. */
export function catalogKeys(locale: string): readonly string[] {
  return Object.keys(catalogs[locale] ?? {});
}

/** "de" when the browser's first language is German, "en" otherwise. */
export function detectLocale(languages: readonly string[] = typeof navigator === "undefined" ? [] : navigator.languages): "en" | "de" {
  return (languages[0] ?? "").toLowerCase().startsWith("de") ? "de" : "en";
}
