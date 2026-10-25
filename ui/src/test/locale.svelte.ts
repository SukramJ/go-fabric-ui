// A reactive locale for tests that switch the language under a page.
export function reactiveLocale(initial: string) {
  let locale = $state(initial);
  return {
    get: () => locale,
    set: (l: string) => {
      locale = l;
    },
  };
}
