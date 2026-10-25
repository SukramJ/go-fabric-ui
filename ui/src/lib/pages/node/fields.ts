// List-valued item fields (Column.multi, Filter.multi): the provider
// carries several values as one comma-separated string.

/** The values of a comma-separated field, trimmed, without empty ones. */
export function listValues(field: string | undefined): string[] {
  if (!field) return [];
  return field
    .split(",")
    .map((v) => v.trim())
    .filter((v) => v !== "");
}

/**
 * Whether an item's field matches a filter value: by equality, or, for a
 * list-valued field, by membership in its list. An empty value never
 * matches a list.
 */
export function fieldMatches(field: string | undefined, value: string, multi: boolean): boolean {
  if (!multi) return field === value;
  const v = value.trim();
  return v !== "" && listValues(field).includes(v);
}

/** The distinct filter values of a field across items, sorted for display. */
export function distinctValues(fields: (string | undefined)[], multi: boolean): string[] {
  const values = new Set<string>();
  for (const f of fields) {
    if (multi) for (const v of listValues(f)) values.add(v);
    else if (f) values.add(f);
  }
  return [...values].sort((a, b) => a.localeCompare(b, undefined, { sensitivity: "base", numeric: true }));
}
