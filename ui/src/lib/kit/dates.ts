// Dates and times in the UI's language, not the browser's or the
// runtime's (UR-P1-ui-182). The language comes from the active
// catalogue's `common.locale`, a BCP 47 tag; a host that adds a catalogue
// for another language names its tag there too.
import type { Translate } from "../i18n";

/** The UI's locale for Intl; undefined (the runtime's) when the translator names none. */
export function uiLocale(t: Translate): string | undefined {
  const l = t("common.locale");
  return l && l !== "common.locale" ? l : undefined;
}

/** value (an ISO string, epoch milliseconds or a Date) as date and time, date or time in the UI's locale. */
export function formatDate(t: Translate, value: string | number | Date, style: "datetime" | "date" | "time" = "datetime"): string {
  const d = new Date(value);
  const as = (loc: string | undefined) => (style === "date" ? d.toLocaleDateString(loc) : style === "time" ? d.toLocaleTimeString(loc) : d.toLocaleString(loc));
  try {
    return as(uiLocale(t));
  } catch {
    // A tag Intl does not take (RangeError): the runtime's locale.
    return as(undefined);
  }
}
