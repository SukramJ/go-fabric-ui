// List-valued item fields (Column.multi, Filter.multi): the provider
// carries the values in `lists[key]`, one entry per value, so a value may
// contain a comma. The comma-separated `fields[key]` is the deprecated
// fallback (until v0.2.0), read only when `lists` has no such key.

/** The part of an item the list helpers read. */
export type ListSource = {
  fields: Record<string, string>;
  lists?: Record<string, string[]>;
};

/**
 * The values of a comma-separated field, trimmed, without empty ones.
 *
 * @deprecated The comma form is accepted until v0.2.0; read
 * {@link itemValues}, which prefers `lists`.
 */
export function listValues(field: string | undefined): string[] {
  if (!field) return [];
  return field
    .split(",")
    .map((v) => v.trim())
    .filter((v) => v !== "");
}

/**
 * The values of an item's list-valued field: `lists[key]` when the key is
 * present there (empty entries dropped, the rest as given), otherwise the
 * comma split of `fields[key]`.
 */
export function itemValues(item: ListSource, key: string): string[] {
  const list = item.lists?.[key];
  if (list) return list.filter((v) => v !== "");
  return listValues(item.fields[key]);
}

/**
 * Whether an item matches a filter value on a field: by equality, or, for
 * a list-valued field, by membership in its list. An empty value never
 * matches a list.
 */
export function fieldMatches(
  item: ListSource,
  key: string,
  value: string,
  multi: boolean,
): boolean {
  if (!multi) return item.fields[key] === value;
  const v = value.trim();
  return v !== "" && itemValues(item, key).includes(v);
}

/** The distinct filter values of a field across items, sorted for display. */
export function distinctValues(
  items: ListSource[],
  key: string,
  multi: boolean,
): string[] {
  const values = new Set<string>();
  for (const item of items) {
    if (multi) for (const v of itemValues(item, key)) values.add(v);
    else if (item.fields[key]) values.add(item.fields[key]);
  }
  return [...values].sort((a, b) =>
    a.localeCompare(b, undefined, { sensitivity: "base", numeric: true }),
  );
}
